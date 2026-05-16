package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/openrtb"
	"adlab-server/internal/repository"
	"adlab-server/pkg/utils"

	"golang.org/x/sync/errgroup"
)

// S2SBidRequest S2S 竞价请求
type S2SBidRequest struct {
	PlacementID string      `json:"placement_id"`
	App         *AppInfo    `json:"app,omitempty"`
	Device      *DeviceInfo `json:"device,omitempty"`
}

// AppInfo 应用信息
type AppInfo struct {
	Bundle  string `json:"bundle,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	UA    string `json:"ua,omitempty"`
	IP    string `json:"ip,omitempty"`
	OS    string `json:"os,omitempty"`
	Model string `json:"model,omitempty"`
	IFA   string `json:"ifa,omitempty"`
}

// S2SBidResponse S2S 竞价响应
type S2SBidResponse struct {
	RequestID    string  `json:"request_id"`
	WinnerDSPID  string  `json:"winner_dsp_id"`
	WinnerPrice  float64 `json:"winner_price"`
	AdMarkup     string  `json:"ad_markup,omitempty"`
	Status       string  `json:"status"` // success / no_fill / error
}

// dspBidResult 单个 DSP 的竞价结果（内部使用）
type dspBidResult struct {
	sourceID  string
	price     float64
	adMarkup  string
	latencyMs int
	status    string // win / lose / no_bid / timeout / error
	errMsg    string
}

// S2SBiddingService S2S 竞价引擎
type S2SBiddingService struct {
	placementRepo  *repository.PlacementRepository
	sourceRepo     *repository.AdSourceRepository
	bidRequestRepo *repository.BidRequestLogRepository
	bidDetailRepo  *repository.BidDetailLogRepository
	appRepo        *repository.AppRepository
	maxConcurrency int
	localBaseURL   string
}

// NewS2SBiddingService 创建 S2SBiddingService
func NewS2SBiddingService(
	placementRepo *repository.PlacementRepository,
	sourceRepo *repository.AdSourceRepository,
	bidRequestRepo *repository.BidRequestLogRepository,
	bidDetailRepo *repository.BidDetailLogRepository,
	appRepo *repository.AppRepository,
) *S2SBiddingService {
	return &S2SBiddingService{
		placementRepo:  placementRepo,
		sourceRepo:     sourceRepo,
		bidRequestRepo: bidRequestRepo,
		bidDetailRepo:  bidDetailRepo,
		appRepo:        appRepo,
		maxConcurrency: 10,
		localBaseURL:   "http://localhost:8080",
	}
}

// WithLocalBaseURL 设置本地服务基础 URL（用于测试或自定义端口）
func (s *S2SBiddingService) WithLocalBaseURL(baseURL string) *S2SBiddingService {
	s.localBaseURL = baseURL
	return s
}

// Bid 执行 S2S 竞价流程
func (s *S2SBiddingService) Bid(ctx context.Context, req *S2SBidRequest) (*S2SBidResponse, error) {
	startTime := time.Now()

	// 1. 生成 UUID request_id
	requestID := utils.NewUUID()

	// 2. 查询广告位
	placement, err := s.placementRepo.FindByPlacementID(req.PlacementID)
	if err != nil {
		return nil, err
	}
	if placement.Status != "active" {
		return nil, errors.New(errors.CodePlacementNotFound, "广告位未激活: "+req.PlacementID)
	}

	// 3. 查询 active S2S 广告源
	sources, err := s.sourceRepo.FindActiveS2SByPlacementID(req.PlacementID)
	if err != nil {
		return nil, errors.Wrap(errors.CodeS2SBidError, "查询 S2S 广告源失败", err)
	}
	if len(sources) == 0 {
		// 记录 no_fill 日志
		s.recordBidLog(requestID, req.PlacementID, 0, "", 0, int(time.Since(startTime).Milliseconds()), "no_fill")
		return nil, errors.New(errors.CodeNoValidBid, "无可用 S2S 广告源")
	}

	// 4. 构建 OpenRTB BidRequest
	ortbReq := s.buildOpenRTBRequest(requestID, req, sources, placement)

	// 5. 使用 errgroup 并发请求 DSP（最大并发 10）
	results := s.concurrentBid(ctx, ortbReq, sources)

	// 6. 过滤底价，选最高出价
	winner := s.selectWinner(results, sources)

	totalLatency := int(time.Since(startTime).Milliseconds())

	// 7. 记录 BidDetailLog（写入失败记录日志，不影响主流程）
	for _, result := range results {
		detail := &model.BidDetailLog{
			RequestID: requestID,
			DSPID:     result.sourceID,
			BidPrice:  result.price,
			LatencyMs: result.latencyMs,
			Status:    result.status,
			ErrorMsg:  result.errMsg,
		}
		if err := s.bidDetailRepo.Create(detail); err != nil {
			// 日志写入失败不影响竞价结果，但需要记录
			_ = err // TODO: 引入结构化日志后替换为 logger.Error
		}
	}

	// 8. 无有效出价
	if winner == nil {
		s.recordBidLog(requestID, req.PlacementID, len(sources), "", 0, totalLatency, "no_fill")
		return nil, errors.New(errors.CodeNoValidBid, "无有效出价")
	}

	// 9. 发送 Win Notice（异步，不阻塞主流程）
	winnerSource := s.findSource(sources, winner.sourceID)
	if winnerSource != nil && winnerSource.DSPURL != "" {
		go s.sendWinNotice(winnerSource, requestID, winner.price)
	}

	// 10. 记录 BidRequestLog
	s.recordBidLog(requestID, req.PlacementID, len(sources), winner.sourceID, winner.price, totalLatency, "success")

	return &S2SBidResponse{
		RequestID:   requestID,
		WinnerDSPID: winner.sourceID,
		WinnerPrice: winner.price,
		AdMarkup:    winner.adMarkup,
		Status:      "success",
	}, nil
}

// buildOpenRTBRequest 构建 OpenRTB BidRequest
// 优先使用请求中携带的 App/Device 信息，若广告位关联了 App 则自动补充 bundle/name
func (s *S2SBiddingService) buildOpenRTBRequest(requestID string, req *S2SBidRequest, sources []model.AdSource, placement *model.Placement) *openrtb.BidRequest {
	// 底价取：广告位全局底价 vs 各广告源底价的最小值，取较大者作为 Imp 底价
	var minSourceFloor float64
	if len(sources) > 0 {
		minSourceFloor = sources[0].FloorPrice
		for _, src := range sources[1:] {
			if src.FloorPrice < minSourceFloor {
				minSourceFloor = src.FloorPrice
			}
		}
	}
	impFloor := minSourceFloor
	if placement.FloorPrice > impFloor {
		impFloor = placement.FloorPrice // 广告位全局底价优先
	}

	imp := openrtb.Imp{
		ID:          "1",
		BidFloor:    impFloor,
		BidFloorCur: "USD",
	}

	// 根据广告位类型填充 Imp 的广告格式对象（让 DSP 知道广告类型）
	switch placement.AdType {
	case "rewarded_video", "interstitial":
		w, h := 1280, 720
		imp.Video = &openrtb.Video{
			MIMEs:       []string{"video/mp4", "video/webm"},
			MinDuration: 5,
			MaxDuration: 60,
			W:           &w,
			H:           &h,
		}
	case "banner":
		w, h := 320, 50
		imp.Banner = &openrtb.Banner{W: &w, H: &h}
	case "native":
		imp.Native = &openrtb.Native{Request: `{"ver":"1.2","layout":1}`, Ver: "1.2"}
	}

	ortbReq := &openrtb.BidRequest{
		ID:   requestID,
		Imp:  []openrtb.Imp{imp},
		AT:   1, // First Price Auction
		TMax: 200,
	}

	// 优先使用请求中的 App 信息
	if req.App != nil {
		ortbReq.App = &openrtb.App{
			Bundle: req.App.Bundle,
			Name:   req.App.Name,
			Ver:    req.App.Version,
		}
	}

	// 若广告位关联了 App，自动补充缺失字段
	if placement.AppID != "" {
		if app, err := s.appRepo.FindByAppID(placement.AppID); err == nil {
			if ortbReq.App == nil {
				ortbReq.App = &openrtb.App{}
			}
			if ortbReq.App.Bundle == "" {
				ortbReq.App.Bundle = app.BundleID
			}
			if ortbReq.App.Name == "" {
				ortbReq.App.Name = app.Name
			}
			if ortbReq.App.StoreURL == "" && app.AppStoreURL != "" {
				ortbReq.App.StoreURL = app.AppStoreURL
			}
		}
	}

	// 填充 Device 信息
	if req.Device != nil {
		ortbReq.Device = &openrtb.Device{
			UA:    req.Device.UA,
			IP:    req.Device.IP,
			OS:    req.Device.OS,
			Model: req.Device.Model,
			IFA:   req.Device.IFA,
		}
	}

	return ortbReq
}

// concurrentBid 并发向所有 DSP 发送竞价请求
func (s *S2SBiddingService) concurrentBid(ctx context.Context, ortbReq *openrtb.BidRequest, sources []model.AdSource) []dspBidResult {
	results := make([]dspBidResult, len(sources))
	var mu sync.Mutex

	// 使用 semaphore 限制最大并发数
	sem := make(chan struct{}, s.maxConcurrency)

	g, gCtx := errgroup.WithContext(ctx)

	for idx, src := range sources {
		idx, src := idx, src // capture loop variables
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.bidFromDSP(gCtx, src, ortbReq)
			mu.Lock()
			results[idx] = result
			mu.Unlock()
			return nil // 单个 DSP 失败不影响整体
		})
	}

	_ = g.Wait()
	return results
}

// bidFromDSP 向单个 DSP 发送竞价请求并解析响应
func (s *S2SBiddingService) bidFromDSP(ctx context.Context, src model.AdSource, ortbReq *openrtb.BidRequest) dspBidResult {
	result := dspBidResult{
		sourceID: src.SourceID,
		status:   "no_bid",
	}

	// DSP URL 为空时自动回退到本地内置模拟器
	dspURL := src.DSPURL
	if dspURL == "" {
		dspURL = fmt.Sprintf("%s/lab/dsp/%s/bid", s.localBaseURL, src.SourceID)
	}

	// 序列化请求
	body, err := json.Marshal(ortbReq)
	if err != nil {
		result.status = "error"
		result.errMsg = fmt.Sprintf("序列化 BidRequest 失败: %v", err)
		return result
	}

	// 创建带超时的 HTTP 客户端
	client := utils.NewHTTPClient(src.TimeoutMs)

	// 构建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, dspURL, bytes.NewReader(body))
	if err != nil {
		result.status = "error"
		result.errMsg = fmt.Sprintf("创建 HTTP 请求失败: %v", err)
		return result
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-openrtb-version", "2.6")

	startTime := time.Now()
	resp, err := client.Do(httpReq)
	result.latencyMs = int(time.Since(startTime).Milliseconds())

	if err != nil {
		// 判断是否超时
		if ctx.Err() != nil || isTimeoutError(err) {
			result.status = "timeout"
			result.errMsg = "请求超时"
		} else {
			result.status = "error"
			result.errMsg = fmt.Sprintf("HTTP 请求失败: %v", err)
		}
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.status = "error"
		result.errMsg = fmt.Sprintf("DSP 返回非 200 状态码: %d", resp.StatusCode)
		return result
	}

	// 解析响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.status = "error"
		result.errMsg = fmt.Sprintf("读取响应体失败: %v", err)
		return result
	}

	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(respBody, &bidResp); err != nil {
		result.status = "error"
		result.errMsg = fmt.Sprintf("解析 BidResponse 失败: %v", err)
		return result
	}

	// 提取出价
	if len(bidResp.SeatBid) == 0 || len(bidResp.SeatBid[0].Bid) == 0 {
		result.status = "no_bid"
		return result
	}

	bid := bidResp.SeatBid[0].Bid[0]
	result.price = bid.Price
	result.adMarkup = bid.AdM
	result.status = "lose" // 初始设为 lose，胜出后更新为 win
	return result
}

// selectWinner 从竞价结果中选出胜出者（最高有效出价）
func (s *S2SBiddingService) selectWinner(results []dspBidResult, sources []model.AdSource) *dspBidResult {
	// 构建 sourceID -> floorPrice 映射
	floorMap := make(map[string]float64, len(sources))
	for _, src := range sources {
		floorMap[src.SourceID] = src.FloorPrice
	}

	var winner *dspBidResult
	for i := range results {
		r := &results[i]
		if r.status != "lose" {
			// 只考虑有出价的结果（status == "lose" 表示有出价但未胜出）
			continue
		}
		floor, ok := floorMap[r.sourceID]
		if !ok {
			continue
		}
		// 出价必须高于底价
		if r.price <= floor {
			r.status = "no_bid"
			continue
		}
		if winner == nil || r.price > winner.price {
			winner = r
		}
	}

	if winner != nil {
		winner.status = "win"
	}
	return winner
}

// findSource 在广告源列表中查找指定 sourceID 的广告源
func (s *S2SBiddingService) findSource(sources []model.AdSource, sourceID string) *model.AdSource {
	for i := range sources {
		if sources[i].SourceID == sourceID {
			return &sources[i]
		}
	}
	return nil
}

// sendWinNotice 向胜出 DSP 发送赢标通知
func (s *S2SBiddingService) sendWinNotice(src *model.AdSource, requestID string, price float64) {
	// DSP URL 为空时回退到本地模拟器
	dspURL := src.DSPURL
	if dspURL == "" {
		dspURL = fmt.Sprintf("%s/lab/dsp/%s", s.localBaseURL, src.SourceID)
	}
	winURL := fmt.Sprintf("%s/win", dspURL)

	// 替换 OpenRTB 标准价格宏 ${AUCTION_PRICE}
	winURL = replacePriceMacro(winURL, price)

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

// replacePriceMacro 替换 OpenRTB 标准价格宏
// ${AUCTION_PRICE} → 实际成交价格（USD CPM，保留 4 位小数）
func replacePriceMacro(urlStr string, price float64) string {
	priceStr := fmt.Sprintf("%.4f", price)
	result := urlStr
	result = strings.ReplaceAll(result, "${AUCTION_PRICE}", priceStr)
	result = strings.ReplaceAll(result, "%24%7BAUCTION_PRICE%7D", priceStr) // URL 编码形式
	return result
}

// recordBidLog 记录竞价请求日志
func (s *S2SBiddingService) recordBidLog(requestID, placementID string, dspCount int, winnerDSPID string, winnerPrice float64, latencyMs int, status string) {
	log := &model.BidRequestLog{
		RequestID:      requestID,
		PlacementID:    placementID,
		BidMode:        "s2s",
		DSPCount:       dspCount,
		WinnerDSPID:    winnerDSPID,
		WinnerPrice:    winnerPrice,
		TotalLatencyMs: latencyMs,
		Status:         status,
	}
	_ = s.bidRequestRepo.Create(log)
}

// isTimeoutError 判断是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	type timeoutErr interface {
		Timeout() bool
	}
	if te, ok := err.(timeoutErr); ok {
		return te.Timeout()
	}
	return false
}
