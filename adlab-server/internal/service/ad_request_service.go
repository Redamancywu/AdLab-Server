package service

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/pkg/utils"
)

// AdRequestService 统一广告请求服务
//
// 核心流程：
//  1. 查广告位 + 关联的所有 active 广告源
//  2. 按 bid_mode 分组：s2s 组并发竞价，waterfall 组顺序竞价，c2s 组返回配置
//  3. 取最高出价（s2s 胜出者 vs waterfall 胜出者 PK）
//  4. 若所有真实广告源均无填充，且 App 开启了 enable_mock_fallback，则用 Mock 广告兜底
type AdRequestService struct {
	placementRepo  *repository.PlacementRepository
	sourceRepo     *repository.AdSourceRepository
	bidRequestRepo *repository.BidRequestLogRepository
	bidDetailRepo  *repository.BidDetailLogRepository
	appRepo        *repository.AppRepository
	mockAdRepo     *repository.MockAdRepository
	materialRepo   *repository.MaterialRepository
	localBaseURL   string
}

// NewAdRequestService 创建 AdRequestService
func NewAdRequestService(
	placementRepo *repository.PlacementRepository,
	sourceRepo *repository.AdSourceRepository,
	bidRequestRepo *repository.BidRequestLogRepository,
	bidDetailRepo *repository.BidDetailLogRepository,
	appRepo *repository.AppRepository,
	mockAdRepo *repository.MockAdRepository,
	materialRepo *repository.MaterialRepository,
) *AdRequestService {
	return &AdRequestService{
		placementRepo:  placementRepo,
		sourceRepo:     sourceRepo,
		bidRequestRepo: bidRequestRepo,
		bidDetailRepo:  bidDetailRepo,
		appRepo:        appRepo,
		mockAdRepo:     mockAdRepo,
		materialRepo:   materialRepo,
		localBaseURL:   "http://localhost:8080",
	}
}

// WithLocalBaseURL 设置本地服务基础 URL
func (s *AdRequestService) WithLocalBaseURL(baseURL string) *AdRequestService {
	s.localBaseURL = baseURL
	return s
}

// Request 统一广告请求入口
//
// 设计原则：
//   - 一个广告位可以同时绑定 S2S、Waterfall、C2S 三种模式的广告源
//   - S2S 广告源：并发向所有 DSP 发请求，取最高出价
//   - Waterfall 广告源：按 priority（或历史 eCPM）顺序逐一请求，有填充即停
//   - C2S 广告源：服务端不竞价，只返回配置让 SDK 自行竞价
//   - S2S 和 Waterfall 的胜出者再 PK，取价格更高的
//   - 所有真实广告源无填充时，根据 App.EnableMockFallback 决定是否用 Mock 兜底
func (s *AdRequestService) Request(ctx context.Context, req *AdRequest, clientIP, userAgent string) (*AdResponse, error) {
	if req.PlacementID == "" {
		return nil, errors.New(errors.CodeValidationFailed, "placement_id 不能为空")
	}

	// 1. 查询广告位
	placement, err := s.placementRepo.FindByPlacementID(req.PlacementID)
	if err != nil {
		return nil, err
	}
	if placement.Status != "active" {
		return nil, errors.New(errors.CodePlacementNotFound, "广告位未激活: "+req.PlacementID)
	}

	// 2. 查询 App 配置（获取 enable_mock_fallback 开关）
	enableMock := true // 默认开启（无 App 关联时也兜底）
	if placement.AppID != "" {
		if app, err := s.appRepo.FindByAppID(placement.AppID); err == nil {
			enableMock = app.EnableMockFallback
		}
	}

	// 3. 查询广告位绑定及其关联广告源，按 bid_mode 分组
	bindings, err := s.placementRepo.FindBindings(req.PlacementID)
	if err != nil {
		return nil, errors.Wrap(errors.CodeS2SBidError, "查询广告位绑定失败", err)
	}

	var s2sSources, waterfallSources []model.AdSource
	var c2sBindings []C2SBindingConfig
	for _, binding := range bindings {
		src := binding.Source
		if src == nil {
			continue
		}
		if src.Status != "active" {
			continue
		}
		if binding.Status != "active" {
			continue
		}
		switch src.BidMode {
		case "s2s":
			s2sSources = append(s2sSources, *src)
		case "waterfall":
			waterfallSources = append(waterfallSources, *src)
		case "c2s":
			c2sBindings = append(c2sBindings, C2SBindingConfig{
				Source:   *src,
				AdUnitID: binding.AdUnitID,
			})
		}
	}

	// 4. 构建竞价请求
	s2sReq := s.buildS2SRequest(req)

	// 5. 并行执行 S2S 和 Waterfall 竞价（两者互不阻塞）
	type bidResult struct {
		resp    *AdResponse
		err     error
		mode    string
	}

	s2sCh := make(chan bidResult, 1)
	wfCh := make(chan bidResult, 1)

	// S2S：并发向所有 S2S 广告源发请求
	go func() {
		if len(s2sSources) == 0 {
			s2sCh <- bidResult{err: errors.New(errors.CodeNoValidBid, "无 S2S 广告源"), mode: "s2s"}
			return
		}
		resp, err := s.runS2SBid(ctx, s2sReq, placement)
		s2sCh <- bidResult{resp: resp, err: err, mode: "s2s"}
	}()

	// Waterfall：按优先级（或历史 eCPM）顺序请求
	go func() {
		if len(waterfallSources) == 0 {
			wfCh <- bidResult{err: errors.New(errors.CodeNoValidBid, "无 Waterfall 广告源"), mode: "waterfall"}
			return
		}
		resp, err := s.runWaterfallBid(ctx, s2sReq, placement)
		wfCh <- bidResult{resp: resp, err: err, mode: "waterfall"}
	}()

	s2sResult := <-s2sCh
	wfResult := <-wfCh

	// 6. 选出最终胜出者（S2S 和 Waterfall 取价格更高的）
	winner := s.pickWinner(s2sResult.resp, s2sResult.err, wfResult.resp, wfResult.err)

	// 7. 若有 C2S 广告源，在响应中附带 C2S 配置（SDK 可同时发起客户端竞价）
	if winner != nil && len(c2sBindings) > 0 {
		winner.C2SSources = s.buildC2SSourceList(c2sBindings)
	}

	// 8. 有真实广告填充，直接返回
	if winner != nil {
		return winner, nil
	}

	// 9. 纯 C2S 模式（只有 C2S 广告源，无 S2S/Waterfall）
	if len(c2sBindings) > 0 && len(s2sSources) == 0 && len(waterfallSources) == 0 {
		return s.buildC2SResponse(req, placement, c2sBindings), nil
	}

	// 10. 所有真实广告源无填充，根据 App 配置决定是否 Mock 兜底
	if enableMock {
		mockResp := s.tryMockFallback(req.PlacementID, placement.AdType, s.localBaseURL)
		if mockResp != nil {
			return mockResp, nil
		}
	}

	return nil, errors.New(errors.CodeNoValidBid, "无有效广告填充")
}

// pickWinner 从 S2S 和 Waterfall 结果中选出价格更高的胜出者
func (s *AdRequestService) pickWinner(s2sResp *AdResponse, s2sErr error, wfResp *AdResponse, wfErr error) *AdResponse {
	var candidates []*AdResponse
	if s2sErr == nil && s2sResp != nil {
		candidates = append(candidates, s2sResp)
	}
	if wfErr == nil && wfResp != nil {
		candidates = append(candidates, wfResp)
	}
	if len(candidates) == 0 {
		return nil
	}
	// 按 WinnerPrice 降序，取最高出价
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].WinnerPrice > candidates[j].WinnerPrice
	})
	return candidates[0]
}

// buildS2SRequest 将统一广告请求转换为 S2S 竞价请求
func (s *AdRequestService) buildS2SRequest(req *AdRequest) *S2SBidRequest {
	s2sReq := &S2SBidRequest{
		PlacementID: req.PlacementID,
	}
	if req.App != nil {
		s2sReq.App = &AppInfo{
			Bundle:  req.App.BundleID,
			Name:    req.App.Name,
			Version: req.App.Version,
		}
	}
	if req.Device != nil {
		s2sReq.Device = &DeviceInfo{
			UA:    req.Device.UA,
			IP:    req.Device.IP,
			OS:    req.Device.Platform,
			Model: req.Device.DeviceModel,
			IFA:   req.Device.IFA,
		}
	}
	return s2sReq
}

// runS2SBid 执行 S2S 竞价并转换为统一响应格式
func (s *AdRequestService) runS2SBid(ctx context.Context, req *S2SBidRequest, placement *model.Placement) (*AdResponse, error) {
	svc := NewS2SBiddingService(
		s.placementRepo, s.sourceRepo,
		s.bidRequestRepo, s.bidDetailRepo, s.appRepo,
	).WithLocalBaseURL(s.localBaseURL)

	result, err := svc.Bid(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.convertS2SResponse(result, placement, "s2s"), nil
}

// runWaterfallBid 执行 Waterfall 竞价并转换为统一响应格式
func (s *AdRequestService) runWaterfallBid(ctx context.Context, req *S2SBidRequest, placement *model.Placement) (*AdResponse, error) {
	svc := NewWaterfallBiddingService(
		s.placementRepo, s.sourceRepo,
		s.bidRequestRepo, s.bidDetailRepo,
	).WithLocalBaseURL(s.localBaseURL)

	result, err := svc.Bid(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.convertWaterfallResponse(result, placement), nil
}

// buildC2SResponse 纯 C2S 模式：返回所有 C2S 绑定配置，SDK 自行竞价
func (s *AdRequestService) buildC2SResponse(req *AdRequest, placement *model.Placement, c2sBindings []C2SBindingConfig) *AdResponse {
	requestID := utils.NewID()
	return &AdResponse{
		RequestID:   requestID,
		PlacementID: req.PlacementID,
		AdType:      placement.AdType,
		BidMode:     "c2s",
		Status:      "c2s_pending",
		C2SSources:  s.buildC2SSourceList(c2sBindings),
		TrackURLs:   s.buildTrackURLs(requestID, "", s.localBaseURL),
	}
}

// C2SSourceConfig C2S 广告源配置（下发给 SDK 用于客户端竞价）
type C2SSourceConfig struct {
	SourceID    string  `json:"source_id"`
	NetworkType string  `json:"network_type"`
	AdUnitID    string  `json:"ad_unit_id,omitempty"`
	AppID       string  `json:"app_id,omitempty"`
	AppKey      string  `json:"app_key,omitempty"`
	FloorPrice  float64 `json:"floor_price"`
	TimeoutMs   int     `json:"timeout_ms"`
	Priority    int     `json:"priority"`
}

type C2SBindingConfig struct {
	Source   model.AdSource
	AdUnitID string
}

// buildC2SSourceList 构建 C2S 广告源配置列表（按 priority 排序）
func (s *AdRequestService) buildC2SSourceList(bindings []C2SBindingConfig) []C2SSourceConfig {
	list := make([]C2SSourceConfig, 0, len(bindings))
	for _, binding := range bindings {
		src := binding.Source
		list = append(list, C2SSourceConfig{
			SourceID:    src.SourceID,
			NetworkType: src.NetworkType,
			AdUnitID:    binding.AdUnitID,
			AppID:       src.AppID,
			AppKey:      src.AppKey,
			FloorPrice:  src.FloorPrice,
			TimeoutMs:   src.TimeoutMs,
			Priority:    src.Priority,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Priority < list[j].Priority
	})
	return list
}

// tryMockFallback 尝试 Mock 广告兜底（仅在 App.EnableMockFallback = true 时调用）
func (s *AdRequestService) tryMockFallback(placementID, adType, baseURL string) *AdResponse {
	ad, err := s.mockAdRepo.FindRandomActive(adType)
	if err != nil || ad == nil {
		return nil
	}

	requestID := utils.NewID()
	resp := &AdResponse{
		RequestID:   requestID,
		PlacementID: placementID,
		AdType:      ad.AdType,
		BidMode:     "mock",
		WinnerDSPID: "mock",
		WinnerPrice: ad.CPMPrice,
		IsMock:      true,
		Status:      "success",
		ClickURL:    ad.ClickURL,
		TrackURLs:   s.buildTrackURLs(requestID, ad.MockAdID, baseURL),
	}

	switch ad.AdType {
	case "rewarded_video", "interstitial":
		mockSvc := &MockAdService{mockAdRepo: s.mockAdRepo, placementRepo: s.placementRepo}
		if vastXML, err := mockSvc.buildVAST(ad, requestID, baseURL); err == nil {
			resp.VASTXML = vastXML
		}
	case "banner":
		resp.ImageURL = ad.ImageURL
	case "splash":
		resp.SplashURL = ad.SplashURL
		if resp.SplashURL == "" {
			resp.SplashURL = ad.ImageURL
		}
	case "native":
		resp.NativeAd = &MockNativeAd{
			Title:        ad.NativeTitle,
			Description:  ad.NativeDescription,
			IconURL:      ad.NativeIconURL,
			ImageURL:     ad.ImageURL,
			CallToAction: ad.NativeCallToAction,
		}
	}

	return resp
}

// convertS2SResponse 将 S2S 竞价结果转换为统一广告响应
func (s *AdRequestService) convertS2SResponse(result *S2SBidResponse, placement *model.Placement, bidMode string) *AdResponse {
	resp := &AdResponse{
		RequestID:   result.RequestID,
		PlacementID: placement.PlacementID,
		AdType:      placement.AdType,
		BidMode:     bidMode,
		WinnerDSPID: result.WinnerDSPID,
		WinnerPrice: result.WinnerPrice,
		IsMock:      false,
		Status:      result.Status,
		TrackURLs:   s.buildTrackURLs(result.RequestID, "", s.localBaseURL),
	}
	switch placement.AdType {
	case "rewarded_video", "interstitial":
		resp.VASTXML = s.buildCompleteVAST(result.RequestID, result.AdMarkup, s.localBaseURL)
	case "banner":
		resp.ImageURL = result.AdMarkup
	default:
		resp.VASTXML = result.AdMarkup
	}
	return resp
}

// convertWaterfallResponse 将 Waterfall 竞价结果转换为统一广告响应
func (s *AdRequestService) convertWaterfallResponse(result *WaterfallBidResponse, placement *model.Placement) *AdResponse {
	resp := &AdResponse{
		RequestID:   result.RequestID,
		PlacementID: placement.PlacementID,
		AdType:      placement.AdType,
		BidMode:     "waterfall",
		WinnerDSPID: result.WinnerDSPID,
		WinnerPrice: result.WinnerPrice,
		IsMock:      false,
		Status:      result.Status,
		TrackURLs:   s.buildTrackURLs(result.RequestID, "", s.localBaseURL),
	}
	switch placement.AdType {
	case "rewarded_video", "interstitial":
		resp.VASTXML = s.buildCompleteVAST(result.RequestID, result.AdMarkup, s.localBaseURL)
	case "banner":
		resp.ImageURL = result.AdMarkup
	default:
		resp.VASTXML = result.AdMarkup
	}
	return resp
}

// buildCompleteVAST 生成完整的 VAST 4.2 XML
// 优先使用 DSP 返回的 ad_markup（若已含 MediaFiles 则直接用）
// 否则从素材库随机取一个，用 VASTGeneratorService 生成
func (s *AdRequestService) buildCompleteVAST(requestID, adMarkup, baseURL string) string {
	if adMarkup != "" && containsMediaFiles(adMarkup) {
		return adMarkup
	}
	if s.materialRepo != nil {
		if material, err := s.materialRepo.FindRandom(); err == nil && material != nil {
			vastSvc := NewVASTGeneratorService(s.materialRepo)
			if vastXML, err := vastSvc.Generate(material.MaterialID, requestID, baseURL); err == nil {
				return vastXML
			}
		}
	}
	return adMarkup
}

func containsMediaFiles(vastXML string) bool {
	return len(vastXML) > 100 &&
		(contains(vastXML, "<MediaFile") || contains(vastXML, "<MediaFiles"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// buildTrackURLs 构建完整的追踪 URL 集合
func (s *AdRequestService) buildTrackURLs(requestID, materialID, baseURL string) AdTrackURLs {
	build := func(event string) string {
		params := url.Values{}
		params.Set("event", event)
		params.Set("request_id", requestID)
		if materialID != "" {
			params.Set("material_id", materialID)
		}
		return fmt.Sprintf("%s/api/v1/track?%s", baseURL, params.Encode())
	}
	return AdTrackURLs{
		Impression:    build("impression"),
		Click:         build("click"),
		Start:         build("start"),
		FirstQuartile: build("firstQuartile"),
		Midpoint:      build("midpoint"),
		ThirdQuartile: build("thirdQuartile"),
		Complete:      build("complete"),
	}
}
