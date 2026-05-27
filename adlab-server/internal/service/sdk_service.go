package service

import (
	"time"

	"adlab-server/internal/repository"
)

// WithAppNetworkConfigRepo 注入 AppNetworkConfigRepository
func (s *SDKService) WithAppNetworkConfigRepo(repo *repository.AppNetworkConfigRepository) *SDKService {
	s.appNetworkConfigRepo = repo
	return s
}

// NewSDKService 创建 SDKService
func NewSDKService(
	appRepo *repository.AppRepository,
	placementRepo *repository.PlacementRepository,
	sourceRepo *repository.AdSourceRepository,
	dspConfigRepo *repository.DSPConfigRepository,
	trackingRepo *repository.TrackingEventLogRepository,
) *SDKService {
	return &SDKService{
		appRepo:       appRepo,
		placementRepo: placementRepo,
		sourceRepo:    sourceRepo,
		dspConfigRepo: dspConfigRepo,
		trackingRepo:  trackingRepo,
		configVersion: time.Now().UnixMilli(),
	}
}

// BumpConfigVersion 递增配置版本号（管理操作后调用，触发 SDK 心跳检测到更新）
func (s *SDKService) BumpConfigVersion() {
	s.configMu.Lock()
	s.configVersion = time.Now().UnixMilli()
	s.configMu.Unlock()
}

// GetConfigVersion 获取当前配置版本号
func (s *SDKService) GetConfigVersion() int64 {
	s.configMu.RLock()
	defer s.configMu.RUnlock()
	return s.configVersion
}

// AdResponse 统一广告响应
type AdResponse struct {
	RequestID   string             `json:"request_id"`
	PlacementID string             `json:"placement_id"`
	AdType      string             `json:"ad_type"`
	BidMode     string             `json:"bid_mode"`
	WinnerDSPID string             `json:"winner_dsp_id,omitempty"`
	WinnerPrice float64            `json:"winner_price,omitempty"`
	IsMock      bool               `json:"is_mock"`
	Status      string             `json:"status"`
	VASTXML     string             `json:"vast_xml,omitempty"`
	ImageURL    string             `json:"image_url,omitempty"`
	SplashURL   string             `json:"splash_url,omitempty"`
	NativeAd    *MockNativeAd      `json:"native_ad,omitempty"`
	ClickURL    string             `json:"click_url,omitempty"`
	TrackURLs   AdTrackURLs        `json:"track_urls"`
	// C2S 广告源配置（当广告位有 C2S 广告源时附带，SDK 用于客户端竞价）
	C2SSources  []C2SSourceConfig  `json:"c2s_sources,omitempty"`
}

// AdTrackURLs 广告追踪 URL 集合
type AdTrackURLs struct {
	Impression    string `json:"impression"`
	Click         string `json:"click"`
	Start         string `json:"start,omitempty"`
	FirstQuartile string `json:"first_quartile,omitempty"`
	Midpoint      string `json:"midpoint,omitempty"`
	ThirdQuartile string `json:"third_quartile,omitempty"`
	Complete      string `json:"complete,omitempty"`
}
