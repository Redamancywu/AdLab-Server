package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/pkg/utils"
)

// C2SBidDetail C2S 上报中单个 DSP 的出价明细
type C2SBidDetail struct {
	DSPID    string  `json:"dsp_id"`
	BidPrice float64 `json:"bid_price"`
	Status   string  `json:"status"` // win / lose / no_bid / timeout / error
}

// C2SReportRequest C2S 上报请求
type C2SReportRequest struct {
	RequestID      string         `json:"request_id"`
	PlacementID    string         `json:"placement_id"`
	WinnerDSPID    string         `json:"winner_dsp_id"`
	WinnerPrice    float64        `json:"winner_price"`
	Displayed      bool           `json:"displayed"`
	BiddingDetails []C2SBidDetail `json:"bidding_details"`
}

// C2SReportResponse C2S 上报响应
type C2SReportResponse struct {
	RequestID    string `json:"request_id"`
	ProxyWinSent bool   `json:"proxy_win_sent"`
}

// C2SReportingService C2S 上报服务
type C2SReportingService struct {
	c2sReportRepo *repository.C2SReportLogRepository
	dspConfigRepo *repository.DSPConfigRepository
	sourceRepo    *repository.AdSourceRepository
}

// NewC2SReportingService 创建 C2SReportingService
func NewC2SReportingService(
	c2sReportRepo *repository.C2SReportLogRepository,
	dspConfigRepo *repository.DSPConfigRepository,
	sourceRepo *repository.AdSourceRepository,
) *C2SReportingService {
	return &C2SReportingService{
		c2sReportRepo: c2sReportRepo,
		dspConfigRepo: dspConfigRepo,
		sourceRepo:    sourceRepo,
	}
}

// Report 处理 C2S 上报
func (s *C2SReportingService) Report(ctx context.Context, req *C2SReportRequest) (*C2SReportResponse, error) {
	// 1. 校验必填字段
	if req.PlacementID == "" {
		return nil, errors.New(errors.CodeC2SDuplicateReport, "placement_id 不能为空")
	}
	if req.WinnerDSPID == "" && req.Displayed {
		return nil, errors.New(errors.CodeC2SDuplicateReport, "displayed 为 true 时 winner_dsp_id 不能为空")
	}
	if len(req.BiddingDetails) == 0 {
		return nil, errors.New(errors.CodeC2SDuplicateReport, "bidding_details 不能为空")
	}

	// 2. 幂等检查：同一 request_id 不允许重复上报
	exists, err := s.c2sReportRepo.ExistsByRequestID(req.RequestID)
	if err != nil {
		return nil, errors.Wrap(errors.CodeC2SReportError, "检查重复上报失败", err)
	}
	if exists {
		return nil, errors.New(errors.CodeC2SDuplicateReport, "request_id 已上报: "+req.RequestID)
	}

	// 3. 序列化 bidding_details 为 JSON
	detailsJSON, err := json.Marshal(req.BiddingDetails)
	if err != nil {
		return nil, errors.Wrap(errors.CodeC2SReportError, "序列化 bidding_details 失败", err)
	}

	// 4. 持久化 C2SReportLog
	log := &model.C2SReportLog{
		RequestID:      req.RequestID,
		PlacementID:    req.PlacementID,
		WinnerDSPID:    req.WinnerDSPID,
		WinnerPrice:    req.WinnerPrice,
		Displayed:      req.Displayed,
		BiddingDetails: model.JSONRaw(detailsJSON),
	}
	if err := s.c2sReportRepo.Create(log); err != nil {
		return nil, err
	}

	// 5. 条件触发 Win Notice 代理
	proxyWinSent := false
	if req.Displayed && req.WinnerDSPID != "" {
		dspCfg, err := s.dspConfigRepo.FindBySourceID(req.WinnerDSPID)
		if err == nil && dspCfg.SupportWinNotice {
			// 查询广告源获取 DSP URL
			src, err := s.sourceRepo.FindBySourceID(req.WinnerDSPID)
			if err == nil && src.DSPURL != "" {
				go s.sendWinNotice(src.DSPURL, req.RequestID, req.WinnerPrice)
				proxyWinSent = true
			}
		}
	}

	return &C2SReportResponse{
		RequestID:    req.RequestID,
		ProxyWinSent: proxyWinSent,
	}, nil
}

// C2SDisplayRequest 展示确认请求
type C2SDisplayRequest struct {
	RequestID   string `json:"request_id"`
	PlacementID string `json:"placement_id"`
	Displayed   bool   `json:"displayed"`
}

// C2SDisplayResponse 展示确认响应
type C2SDisplayResponse struct {
	RequestID    string `json:"request_id"`
	ProxyWinSent bool   `json:"proxy_win_sent"`
}

// ConfirmDisplay 处理展示确认（更新已上报记录的 displayed 状态并触发 Win Notice）
func (s *C2SReportingService) ConfirmDisplay(ctx context.Context, req *C2SDisplayRequest) (*C2SDisplayResponse, error) {
	if req.RequestID == "" {
		return nil, errors.New(errors.CodeC2SDuplicateReport, "request_id 不能为空")
	}

	// 查询已上报记录
	log, err := s.c2sReportRepo.FindByRequestID(req.RequestID)
	if err != nil {
		return nil, err
	}

	// 已经确认过展示，幂等返回
	if log.Displayed {
		return &C2SDisplayResponse{RequestID: req.RequestID, ProxyWinSent: false}, nil
	}

	// 更新 displayed 状态
	log.Displayed = req.Displayed
	if err := s.c2sReportRepo.UpdateDisplayed(req.RequestID, req.Displayed); err != nil {
		return nil, errors.Wrap(errors.CodeC2SReportError, "更新展示状态失败", err)
	}

	// 条件触发 Win Notice 代理
	proxyWinSent := false
	if req.Displayed && log.WinnerDSPID != "" {
		dspCfg, err := s.dspConfigRepo.FindBySourceID(log.WinnerDSPID)
		if err == nil && dspCfg.SupportWinNotice {
			src, err := s.sourceRepo.FindBySourceID(log.WinnerDSPID)
			if err == nil && src.DSPURL != "" {
				go s.sendWinNotice(src.DSPURL, req.RequestID, log.WinnerPrice)
				proxyWinSent = true
			}
		}
	}

	return &C2SDisplayResponse{RequestID: req.RequestID, ProxyWinSent: proxyWinSent}, nil
}
func (s *C2SReportingService) sendWinNotice(dspURL, requestID string, price float64) {
	winURL := fmt.Sprintf("%s/win", dspURL)
	// 替换 OpenRTB 标准价格宏 ${AUCTION_PRICE}
	winURL = strings.ReplaceAll(winURL, "${AUCTION_PRICE}", fmt.Sprintf("%.4f", price))
	payload := map[string]interface{}{
		"request_id": requestID,
		"price":      price,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	client := utils.NewHTTPClient(500)
	req, err := http.NewRequest(http.MethodPost, winURL, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
