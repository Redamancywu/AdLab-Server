package service

import (
	"context"
	"sort"
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
)

// Init SDK 初始化
// 返回格式对标 TopOn/MAX：App 级别网络列表 + 广告位 Waterfall 配置
func (s *SDKService) Init(ctx context.Context, req *SDKInitRequest) (*SDKInitResponse, error) {
	return s.buildInitResponse(ctx, req)
}

func (s *SDKService) buildInitResponse(ctx context.Context, req *SDKInitRequest) (*SDKInitResponse, error) {
	if req.AppID == "" {
		return nil, errors.New(errors.CodeValidationFailed, "app_id 不能为空")
	}

	app, err := s.appRepo.FindByAppID(req.AppID)
	if err != nil {
		return nil, err
	}
	if app.Status != "active" {
		return nil, errors.New(errors.CodeEntityNotFound, "应用未激活: "+req.AppID)
	}

	allPlacements, _, err := s.placementRepo.FindByAppID(req.AppID, 0, 0)
	if err != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询广告位失败", err)
	}

	networkMap := make(map[string]*SDKNetworkInit)
	if s.appNetworkConfigRepo != nil {
		configs, err := s.appNetworkConfigRepo.FindByAppAndPlatform(req.AppID, req.Platform)
		if err == nil {
			for _, cfg := range configs {
				networkMap[cfg.NetworkType] = &SDKNetworkInit{
					NetworkType:  cfg.NetworkType,
					AdapterClass: firstNonEmpty(cfg.AdapterClass, networkAdapterClass(cfg.NetworkType)),
					InitParams:   parseJSONObject(cfg.InitParamsJSON),
					Status:       cfg.Status,
				}
			}
		}
	}

	var sdkPlacements []SDKPlacementConfig
	for _, placement := range allPlacements {
		if placement.Status != "active" {
			continue
		}
		placementConfig, err := s.buildPlacementConfig(&placement, networkMap)
		if err != nil {
			continue
		}
		sdkPlacements = append(sdkPlacements, *placementConfig)
	}

	networks := make([]SDKNetworkInit, 0, len(networkMap))
	keys := make([]string, 0, len(networkMap))
	for key := range networkMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		networks = append(networks, *networkMap[key])
	}

	sort.Slice(sdkPlacements, func(i, j int) bool {
		return sdkPlacements[i].PlacementID < sdkPlacements[j].PlacementID
	})

	global := SDKGlobalConfig{
		DefaultTimeoutMs:   800,
		MaxRetries:         1,
		EnableMockFallback: app.EnableMockFallback,
		LogLevel:           "info",
		HeartbeatIntervalS: 300,
	}
	configVersion := s.GetConfigVersion()
	configHash := buildConfigHash(req.AppID, req.Platform, global, networks, sdkPlacements)

	return &SDKInitResponse{
		AppID:         app.AppID,
		AppName:       app.Name,
		BundleID:      app.BundleID,
		Platform:      app.Platform,
		ConfigVersion: configVersion,
		ConfigHash:    configHash,
		Global:        global,
		Networks:      networks,
		Placements:    sdkPlacements,
		ServerTime:    time.Now().UnixMilli(),
	}, nil
}

func (s *SDKService) buildPlacementConfig(placement *model.Placement, networkMap map[string]*SDKNetworkInit) (*SDKPlacementConfig, error) {
	bindings, err := s.placementRepo.FindBindings(placement.PlacementID)
	if err != nil {
		return nil, err
	}

	instances := make([]SDKPlacementInstance, 0, len(bindings))
	for _, binding := range bindings {
		src := binding.Source
		if src == nil || src.Status != "active" || binding.Status != "active" {
			continue
		}

		if src.NetworkType != "" && src.NetworkType != "custom" {
			if _, exists := networkMap[src.NetworkType]; !exists {
				initParams := make(map[string]interface{})
				if src.AppID != "" {
					initParams["app_id"] = src.AppID
				}
				if src.AppKey != "" {
					initParams["app_key"] = src.AppKey
				}
				for key, value := range parseJSONObject(src.ExtraParams) {
					initParams[key] = value
				}
				networkMap[src.NetworkType] = &SDKNetworkInit{
					NetworkType:  src.NetworkType,
					AdapterClass: networkAdapterClass(src.NetworkType),
					InitParams:   initParams,
					Status:       "active",
				}
			}
		}

		instance := SDKPlacementInstance{
			InstanceID:   firstNonEmpty(binding.InstanceID, placement.PlacementID+"_"+src.SourceID),
			InstanceName: binding.InstanceName,
			SourceID:     src.SourceID,
			NetworkType:  src.NetworkType,
			BidMode:      src.BidMode,
			Priority:     src.Priority,
			FloorPrice:   coalesceFloat(binding.FloorPriceOverride, src.FloorPrice),
			TimeoutMs:    coalesceInt(binding.TimeoutMsOverride, src.TimeoutMs),
			AdUnitID:     binding.AdUnitID,
			LoadParams:   parseJSONObject(binding.LoadParamsJSON),
			Stats: SDKInstanceStats{
				HistoricalECPM: src.HistoricalECPM,
				SampleCount:    src.ECPMSampleCount,
			},
			Status: binding.Status,
		}
		instances = append(instances, instance)
	}

	sort.Slice(instances, func(i, j int) bool {
		if instances[i].Stats.SampleCount > 0 && instances[j].Stats.SampleCount > 0 &&
			instances[i].Stats.HistoricalECPM != instances[j].Stats.HistoricalECPM {
			return instances[i].Stats.HistoricalECPM > instances[j].Stats.HistoricalECPM
		}
		if instances[i].Priority != instances[j].Priority {
			return instances[i].Priority < instances[j].Priority
		}
		if instances[i].NetworkType != instances[j].NetworkType {
			return instances[i].NetworkType < instances[j].NetworkType
		}
		return instances[i].SourceID < instances[j].SourceID
	})

	waterfall := make([]SDKWaterfallItem, 0, len(instances))
	for _, instance := range instances {
		waterfall = append(waterfall, SDKWaterfallItem{
			SourceID:       instance.SourceID,
			NetworkType:    instance.NetworkType,
			BidMode:        instance.BidMode,
			Priority:       instance.Priority,
			FloorPrice:     instance.FloorPrice,
			TimeoutMs:      instance.TimeoutMs,
			AdUnitID:       instance.AdUnitID,
			HistoricalECPM: instance.Stats.HistoricalECPM,
		})
	}

	return &SDKPlacementConfig{
		PlacementID: placement.PlacementID,
		AdType:      placement.AdType,
		FloorPrice:  placement.FloorPrice,
		Status:      placement.Status,
		PlacementParams: SDKPlacementParams{
			DefaultTimeoutMs:  800,
			DefaultFloorPrice: placement.FloorPrice,
			CacheTTLS:         3600,
		},
		Instances: instances,
		Waterfall: waterfall,
	}, nil
}
