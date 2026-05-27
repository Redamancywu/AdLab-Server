package service

import (
	"context"
	"fmt"
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/pkg/utils"
)

// InitComplete 处理 SDK 初始化完成上报
func (s *SDKService) InitComplete(ctx context.Context, req *SDKInitCompleteRequest) (*SDKInitCompleteResponse, error) {
	if req.AppID == "" {
		return nil, errors.New(errors.CodeValidationFailed, "app_id 不能为空")
	}

	for _, net := range req.Networks {
		_ = s.trackingRepo.Create(&model.TrackingEventLog{
			RequestID:  fmt.Sprintf("init_%s_%d", req.AppID, time.Now().UnixMilli()),
			MaterialID: net.NetworkType,
			EventType:  fmt.Sprintf("sdk_init_%s", net.Status),
			Timestamp:  time.Now().UnixMilli(),
		})
	}

	failedNetworks := make(map[string]bool)
	for _, net := range req.Networks {
		if net.Status != "success" {
			failedNetworks[net.NetworkType] = true
		}
	}

	if len(failedNetworks) == 0 {
		return &SDKInitCompleteResponse{Message: "all networks initialized successfully"}, nil
	}

	initResp, err := s.Init(ctx, &SDKInitRequest{AppID: req.AppID, Platform: req.Platform})
	if err != nil {
		return &SDKInitCompleteResponse{Message: "ok"}, nil
	}

	var adjusted []SDKPlacementConfig
	for _, p := range initResp.Placements {
		var filteredWaterfall []SDKWaterfallItem
		for _, item := range p.Waterfall {
			if !failedNetworks[item.NetworkType] {
				filteredWaterfall = append(filteredWaterfall, item)
			}
		}
		p.Waterfall = filteredWaterfall
		adjusted = append(adjusted, p)
	}

	return &SDKInitCompleteResponse{
		AdjustedPlacements: adjusted,
		Message:            fmt.Sprintf("%d network(s) failed, waterfall adjusted", len(failedNetworks)),
	}, nil
}

// Heartbeat 处理 SDK 心跳
func (s *SDKService) Heartbeat(ctx context.Context, req *SDKHeartbeatRequest) (*SDKHeartbeatResponse, error) {
	_ = s.trackingRepo.Create(&model.TrackingEventLog{
		RequestID:  utils.NewID(),
		MaterialID: req.AppID,
		EventType:  "sdk_heartbeat",
		Timestamp:  time.Now().UnixMilli(),
	})

	currentVersion := s.GetConfigVersion()
	configUpdated := req.ConfigVersion > 0 && req.ConfigVersion != currentVersion
	configHash := fmt.Sprintf("v%d", currentVersion)
	refreshReason := ""
	if configUpdated {
		refreshReason = "config_version_changed"
	}

	return &SDKHeartbeatResponse{
		ConfigUpdated: configUpdated,
		ConfigVersion: currentVersion,
		ConfigHash:    configHash,
		RefreshReason: refreshReason,
		ServerTime:    time.Now().UnixMilli(),
	}, nil
}

// ReportECPM 处理 eCPM 上报
func (s *SDKService) ReportECPM(ctx context.Context, req *SDKECPMReportRequest) error {
	if req.PlacementID == "" || req.SourceID == "" {
		return errors.New(errors.CodeValidationFailed, "placement_id 和 source_id 不能为空")
	}
	if req.ECPM < 0 {
		return errors.New(errors.CodeValidationFailed, "ecpm 不能为负数")
	}

	if req.Displayed && req.ECPM > 0 {
		if err := s.sourceRepo.UpdateECPM(req.SourceID, req.ECPM); err != nil {
			_ = err
		}
	}

	_ = s.trackingRepo.Create(&model.TrackingEventLog{
		RequestID:  utils.NewID(),
		MaterialID: fmt.Sprintf("%s_%s", req.PlacementID, req.SourceID),
		EventType:  "ecpm_report",
		Timestamp:  time.Now().UnixMilli(),
	})

	return nil
}
