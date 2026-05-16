package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/openrtb"
	"adlab-server/internal/repository"
	"adlab-server/pkg/utils"
)

// WaterfallBidResponse Waterfall 竞价响应
type WaterfallBidResponse struct {
	RequestID   string  `json:"request_id"`
	WinnerDSPID string  `json:"winner_dsp_id"`
	WinnerPrice float64 `json:"winner_price"`
	AdMarkup    string  `json:"ad_markup,omitempty"`
	Status      string  `json:"status"` // success / no_fill
	// 瀑布流特有：记录尝试了哪些 DSP 才找到填充
	TriedDSPs []string `json:"tried_dsps,omitempty"`
}

// WaterfallBiddingService Waterfall（瀑布流）竞价引擎
//
// Waterfall 逻辑：按广告源 priority 升序（数值越小优先级越高），
// 依次向每个 DSP 发送竞价请求，收到有效出价（高于底价）即停止，
// 不等待其他 DSP，直接返回第一个有效出价。
type WaterfallBiddingService struct {
	placementRepo  *repository.PlacementRepository
	sourceRepo     *repository.AdSourceRepository
	bidRequestRepo *repository.BidRequestLogRepository
	bidDetailRepo  *repository.BidDetailLogRepository
	localBaseURL   string
}

// NewWaterfallBiddingService 创建 WaterfallBiddingService
func NewWaterfallBiddingService(
	placementRepo *repository.PlacementRepository,
	sourceRepo *repository.AdSourceRepository,
	bidRequestRepo *repository.BidRequestLogRepository,
	bidDetailRepo *repository.BidDetailLogRepository,
) *WaterfallBiddingService {
	return &WaterfallBiddingService{
		placementRepo:  placementRepo,
		sourceRepo:     sourceRepo,
		bidRequestRepo: bidRequestRepo,
		bidDetailRepo:  bidDetailRepo,
		localBaseURL:   "http://localhost:8080",
	}
}

// WithLocalBaseURL 设置本地服务基础 URL
func (s *WaterfallBiddingService) WithLocalBaseURL(baseURL string) *WaterfallBiddingService {
	s.localBaseURL = baseURL
	return s
}

// Bid 执行 Waterfall 竞价流程
func (s *WaterfallBiddingService) Bid(ctx context.Context, req *S2SBidRequest) (*WaterfallBidResponse, error) {
	startTime := time.Now()
	requestID := utils.NewID()

	// 1. 查询广告位
	placement, err := s.placementRepo.FindByPlacementID(req.PlacementID)
	if err != nil {
		return nil, err
	}
	if placement.Status != "active" {
		return nil, errors.New(errors.CodePlacementNotFound, "广告位未激活: "+req.PlacementID)
	}

	// 2. 查询 active Waterfall 广告源（已按 priority ASC 排序）
	sources, err := s.sourceRepo.FindActiveWaterfallByPlacementID(req.PlacementID)
	if err != nil {
		return nil, errors.Wrap(errors.CodeS2SBidError, "查询 Waterfall 广告源失败", err)
	}
	if len(sources) == 0 {
		s.recordBidLog(requestID, req.PlacementID, 0, "", 0, int(time.Since(startTime).Milliseconds()), "no_fill")
		return nil, errors.New(errors.CodeNoValidBid, "无可用 Waterfall 广告源")
	}

	// 3. 构建 OpenRTB BidRequest（复用 S2S 的结构）
	ortbReq := &openrtb.BidRequest{
		ID: requestID,
		Imp: []openrtb.Imp{
			{ID: "1", BidFloorCur: "USD"},
		},
	}
	if req.App != nil {
		ortbReq.App = &openrtb.App{Bundle: req.App.Bundle, Name: req.App.Name, Ver: req.App.Version}
	}
	if req.Device != nil {
		ortbReq.Device = &openrtb.Device{UA: req.Device.UA, IP: req.Device.IP, OS: req.Device.OS, Model: req.Device.Model, IFA: req.Device.IFA}
	}

	// 4. 瀑布流：按优先级顺序逐一请求，有填充即停
	var triedDSPs []string
	var allDetails []model.BidDetailLog

	for _, src := range sources {
		triedDSPs = append(triedDSPs, src.SourceID)

		// 每个 DSP 使用自己的底价
		ortbReq.Imp[0].BidFloor = src.FloorPrice

		result := s.bidFromDSP(ctx, src, ortbReq)

		detail := model.BidDetailLog{
			RequestID: requestID,
			DSPID:     result.sourceID,
			BidPrice:  result.price,
			LatencyMs: result.latencyMs,
			Status:    result.status,
			ErrorMsg:  result.errMsg,
		}
		allDetails = append(allDetails, detail)

		// 有有效出价（status == "lose" 表示有出价，尚未判断胜负）
		if result.status == "lose" && result.price > src.FloorPrice {
			// 找到填充，标记为 win，停止瀑布流
			result.status = "win"
			detail.Status = "win"
			allDetails[len(allDetails)-1] = detail

			// 记录所有明细
			for _, d := range allDetails {
				d := d
				_ = s.bidDetailRepo.Create(&d)
			}

			totalLatency := int(time.Since(startTime).Milliseconds())
			s.recordBidLog(requestID, req.PlacementID, len(triedDSPs), src.SourceID, result.price, totalLatency, "success")

			// 异步发送 Win Notice
			go s.sendWinNotice(&src, requestID, result.price)

			return &WaterfallBidResponse{
				RequestID:   requestID,
				WinnerDSPID: src.SourceID,
				WinnerPrice: result.price,
				AdMarkup:    result.adMarkup,
				Status:      "success",
				TriedDSPs:   triedDSPs,
			}, nil
		}
		// 无填充或错误，继续下一个 DSP
	}

	// 所有 DSP 均无填充
	for _, d := range allDetails {
		d := d
		_ = s.bidDetailRepo.Create(&d)
	}
	totalLatency := int(time.Since(startTime).Milliseconds())
	s.recordBidLog(requestID, req.PlacementID, len(triedDSPs), "", 0, totalLatency, "no_fill")

	return nil, errors.New(errors.CodeNoValidBid, fmt.Sprintf("Waterfall 所有 %d 个 DSP 均无填充", len(sources)))
}

// bidFromDSP 向单个 DSP 发送竞价请求（与 S2S 逻辑相同，复用）
func (s *WaterfallBiddingService) bidFromDSP(ctx context.Context, src model.AdSource, ortbReq *openrtb.BidRequest) dspBidResult {
	result := dspBidResult{sourceID: src.SourceID, status: "no_bid"}

	dspURL := src.DSPURL
	if dspURL == "" {
		dspURL = fmt.Sprintf("%s/lab/dsp/%s/bid", s.localBaseURL, src.SourceID)
	}

	body, err := json.Marshal(ortbReq)
	if err != nil {
		result.status = "error"
		result.errMsg = "序列化 BidRequest 失败: " + err.Error()
		return result
	}

	client := utils.NewHTTPClient(src.TimeoutMs)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, dspURL, bytes.NewReader(body))
	if err != nil {
		result.status = "error"
		result.errMsg = "创建 HTTP 请求失败: " + err.Error()
		return result
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-openrtb-version", "2.6")

	t0 := time.Now()
	resp, err := client.Do(httpReq)
	result.latencyMs = int(time.Since(t0).Milliseconds())

	if err != nil {
		if ctx.Err() != nil || isTimeoutError(err) {
			result.status = "timeout"
			result.errMsg = "请求超时"
		} else {
			result.status = "error"
			result.errMsg = "HTTP 请求失败: " + err.Error()
		}
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.status = "error"
		result.errMsg = fmt.Sprintf("DSP 返回非 200 状态码: %d", resp.StatusCode)
		return result
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.status = "error"
		result.errMsg = "读取响应体失败: " + err.Error()
		return result
	}

	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(respBody, &bidResp); err != nil {
		result.status = "error"
		result.errMsg = "解析 BidResponse 失败: " + err.Error()
		return result
	}

	if len(bidResp.SeatBid) == 0 || len(bidResp.SeatBid[0].Bid) == 0 {
		result.status = "no_bid"
		return result
	}

	bid := bidResp.SeatBid[0].Bid[0]
	result.price = bid.Price
	result.adMarkup = bid.AdM
	result.status = "lose"
	return result
}

// sendWinNotice 向胜出 DSP 发送赢标通知
func (s *WaterfallBiddingService) sendWinNotice(src *model.AdSource, requestID string, price float64) {
	dspURL := src.DSPURL
	if dspURL == "" {
		dspURL = fmt.Sprintf("%s/lab/dsp/%s", s.localBaseURL, src.SourceID)
	}
	winURL := fmt.Sprintf("%s/win", dspURL)
	payload := map[string]interface{}{"request_id": requestID, "price": price}
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

// recordBidLog 记录竞价请求日志
func (s *WaterfallBiddingService) recordBidLog(requestID, placementID string, dspCount int, winnerDSPID string, winnerPrice float64, latencyMs int, status string) {
	log := &model.BidRequestLog{
		RequestID:      requestID,
		PlacementID:    placementID,
		BidMode:        "waterfall",
		DSPCount:       dspCount,
		WinnerDSPID:    winnerDSPID,
		WinnerPrice:    winnerPrice,
		TotalLatencyMs: latencyMs,
		Status:         status,
	}
	_ = s.bidRequestRepo.Create(log)
}
