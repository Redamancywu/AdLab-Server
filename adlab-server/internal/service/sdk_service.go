package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/pkg/utils"
)

// ── SDK 初始化 ────────────────────────────────────────────

// SDKInitRequest SDK 初始化请求
type SDKInitRequest struct {
	AppID      string         `json:"app_id"`              // AdLab 应用 ID（必填）
	SDKVersion string         `json:"sdk_version"`         // SDK 版本，如 "1.0.0"
	Platform   string         `json:"platform"`            // ios / android
	Device     *SDKDeviceInfo `json:"device,omitempty"`
}

// SDKDeviceInfo SDK 初始化时的设备信息
type SDKDeviceInfo struct {
	OSVersion   string `json:"os_version,omitempty"`
	DeviceModel string `json:"device_model,omitempty"`
	Language    string `json:"language,omitempty"`
	IFA         string `json:"ifa,omitempty"`   // IDFA / GAID
	IFAType     string `json:"ifa_type,omitempty"` // idfa / gaid / custom
}

// SDKInitResponse SDK 初始化响应（对标 TopOn/MAX 格式）
type SDKInitResponse struct {
	AppID    string `json:"app_id"`
	AppName  string `json:"app_name"`
	BundleID string `json:"bundle_id"`
	Platform string `json:"platform"`

	// 全局 SDK 配置
	Config SDKGlobalConfig `json:"config"`

	// 需要初始化的广告网络列表（App 级别，SDK 逐一初始化）
	// 对标 TopOn 的 channel_list / MAX 的 networks
	Networks []SDKNetworkInit `json:"networks"`

	// 广告位列表（含各广告位的 Waterfall 配置）
	// 对标 TopOn 的 placement_list
	Placements []SDKPlacementConfig `json:"placements"`

	// 服务端时间戳（用于 SDK 时钟校准）
	ServerTime int64 `json:"server_time"`
}

// SDKGlobalConfig SDK 全局配置
type SDKGlobalConfig struct {
	DefaultTimeoutMs   int    `json:"default_timeout_ms"`   // 默认竞价超时（ms）
	MaxRetries         int    `json:"max_retries"`          // 最大重试次数
	EnableMockFallback bool   `json:"enable_mock_fallback"` // 无填充时是否启用 Mock 兜底
	LogLevel           string `json:"log_level"`            // debug / info / warn / error
	HeartbeatIntervalS int    `json:"heartbeat_interval_s"` // 心跳上报间隔（秒，0=不上报）
}

// SDKNetworkInit App 级别的广告网络初始化参数
// SDK 收到后，逐一调用对应网络 SDK 的 initialize() 方法
type SDKNetworkInit struct {
	NetworkType  string            `json:"network_type"`            // admob/applovin/unity/pangle/mintegral/ironsource/custom
	AdapterClass string            `json:"adapter_class"`           // SDK 中对应的 Adapter 类名，如 "AdLabAdMobAdapter"
	AppID        string            `json:"app_id,omitempty"`        // 网络 App ID
	AppKey       string            `json:"app_key,omitempty"`       // 网络 App Key / SDK Key
	Extra        map[string]string `json:"extra,omitempty"`         // 网络特有参数（如 Unity 的 game_id）
}

// SDKPlacementConfig 广告位配置（含 Waterfall 排序）
type SDKPlacementConfig struct {
	PlacementID string              `json:"placement_id"`
	AdType      string              `json:"ad_type"`      // rewarded_video/interstitial/banner/splash/native
	FloorPrice  float64             `json:"floor_price"`  // 广告位全局底价
	Status      string              `json:"status"`
	// Waterfall 配置：按 priority 排序，SDK 按此顺序请求各网络
	Waterfall   []SDKWaterfallItem  `json:"waterfall"`
}

// SDKWaterfallItem Waterfall 中的单个广告源配置
type SDKWaterfallItem struct {
	SourceID    string  `json:"source_id"`
	NetworkType string  `json:"network_type"`
	BidMode     string  `json:"bid_mode"`     // s2s/c2s/waterfall
	Priority    int     `json:"priority"`     // 排序优先级（越小越优先）
	FloorPrice  float64 `json:"floor_price"`  // 该广告源底价
	TimeoutMs   int     `json:"timeout_ms"`
	// 广告单元 ID（C2S/Waterfall 模式下 SDK 直接用此 ID 向网络请求广告）
	AdUnitID    string  `json:"ad_unit_id,omitempty"`
	// 历史 eCPM（用于动态排序，初始为 0）
	HistoricalECPM float64 `json:"historical_ecpm"`
}

// ── 初始化完成上报 ────────────────────────────────────────

// SDKInitCompleteRequest SDK 初始化完成上报
type SDKInitCompleteRequest struct {
	AppID      string                    `json:"app_id"`
	SDKVersion string                    `json:"sdk_version"`
	Platform   string                    `json:"platform"`
	// 各网络初始化结果
	Networks   []SDKNetworkInitResult    `json:"networks"`
	// 初始化总耗时（ms）
	DurationMs int                       `json:"duration_ms"`
}

// SDKNetworkInitResult 单个网络的初始化结果
type SDKNetworkInitResult struct {
	NetworkType string `json:"network_type"`
	Status      string `json:"status"`       // success / failed / timeout
	ErrorMsg    string `json:"error_msg,omitempty"`
	DurationMs  int    `json:"duration_ms"`
	// 网络 SDK 版本（用于兼容性检查）
	AdapterVersion string `json:"adapter_version,omitempty"`
}

// SDKInitCompleteResponse 初始化完成上报响应
type SDKInitCompleteResponse struct {
	// 服务端根据初始化结果，返回调整后的 Waterfall 配置
	// 若某个网络初始化失败，从 Waterfall 中移除
	AdjustedPlacements []SDKPlacementConfig `json:"adjusted_placements,omitempty"`
	Message            string               `json:"message"`
}

// ── 心跳上报 ──────────────────────────────────────────────

// SDKHeartbeatRequest SDK 心跳请求
type SDKHeartbeatRequest struct {
	AppID      string `json:"app_id"`
	SDKVersion string `json:"sdk_version"`
	Platform   string `json:"platform"`
	// 设备标识（用于 DAU 统计）
	IFA        string `json:"ifa,omitempty"`
	// 当前活跃广告位（正在展示或已加载）
	ActivePlacements []string `json:"active_placements,omitempty"`
	// SDK 缓存的配置版本号（服务端比对，不一致则通知更新）
	ConfigVersion int64 `json:"config_version,omitempty"`
}

// SDKHeartbeatResponse 心跳响应
type SDKHeartbeatResponse struct {
	// 是否有配置更新（true 时 SDK 应重新调用 /sdk/init）
	ConfigUpdated  bool  `json:"config_updated"`
	ConfigVersion  int64 `json:"config_version"` // 当前配置版本号，SDK 缓存此值用于下次比对
	ServerTime     int64 `json:"server_time"`
}

// ── eCPM 上报 ─────────────────────────────────────────────

// SDKECPMReportRequest eCPM 上报请求（C2S 竞价完成后上报，用于优化 Waterfall 排序）
type SDKECPMReportRequest struct {
	AppID       string  `json:"app_id"`
	PlacementID string  `json:"placement_id"`
	SourceID    string  `json:"source_id"`
	NetworkType string  `json:"network_type"`
	ECPM        float64 `json:"ecpm"`         // 本次实际 eCPM（USD CPM）
	AdType      string  `json:"ad_type"`
	// 是否展示（false=加载成功但未展示，true=已展示）
	Displayed   bool    `json:"displayed"`
}

// ── SDKService ────────────────────────────────────────────

// SDKService SDK 服务
type SDKService struct {
	appRepo       *repository.AppRepository
	placementRepo *repository.PlacementRepository
	sourceRepo    *repository.AdSourceRepository
	dspConfigRepo *repository.DSPConfigRepository
	trackingRepo  *repository.TrackingEventLogRepository
	// configVersion 全局配置版本号，每次管理操作后递增，心跳用于检测是否需要重新拉取配置
	configVersion int64
	configMu      sync.RWMutex
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

// Init SDK 初始化
// 返回格式对标 TopOn/MAX：App 级别网络列表 + 广告位 Waterfall 配置
func (s *SDKService) Init(ctx context.Context, req *SDKInitRequest) (*SDKInitResponse, error) {
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

	// 查询该 App 下所有广告位
	allPlacements, _, err := s.placementRepo.FindAll(0, 0)
	if err != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询广告位失败", err)
	}

	// 收集所有需要初始化的网络（去重，App 级别）
	networkMap := make(map[string]*SDKNetworkInit)
	var sdkPlacements []SDKPlacementConfig

	for _, p := range allPlacements {
		if p.AppID != req.AppID || p.Status != "active" {
			continue
		}

		bindings, err := s.placementRepo.FindBindings(p.PlacementID)
		if err != nil {
			continue
		}

		var waterfall []SDKWaterfallItem
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

			// 收集 App 级别的网络初始化参数（去重）
			if src.NetworkType != "" && src.NetworkType != "custom" {
				if _, exists := networkMap[src.NetworkType]; !exists {
					extra := parseExtraParams(src.ExtraParams)
					networkMap[src.NetworkType] = &SDKNetworkInit{
						NetworkType:  src.NetworkType,
						AdapterClass: networkAdapterClass(src.NetworkType),
						AppID:        src.AppID,
						AppKey:       src.AppKey,
						Extra:        extra,
					}
				}
			}

			waterfall = append(waterfall, SDKWaterfallItem{
				SourceID:       src.SourceID,
				NetworkType:    src.NetworkType,
				BidMode:        src.BidMode,
				Priority:       src.Priority,
				FloorPrice:     src.FloorPrice,
				TimeoutMs:      src.TimeoutMs,
				AdUnitID:       binding.AdUnitID,
				HistoricalECPM: src.HistoricalECPM, // 真实历史 eCPM，驱动客户端排序
			})
		}

		sdkPlacements = append(sdkPlacements, SDKPlacementConfig{
			PlacementID: p.PlacementID,
			AdType:      p.AdType,
			FloorPrice:  p.FloorPrice,
			Status:      p.Status,
			Waterfall:   waterfall,
		})
	}

	// 转换 networkMap 为有序列表
	var networks []SDKNetworkInit
	for _, n := range networkMap {
		networks = append(networks, *n)
	}

	return &SDKInitResponse{
		AppID:    app.AppID,
		AppName:  app.Name,
		BundleID: app.BundleID,
		Platform: app.Platform,
		Config: SDKGlobalConfig{
			DefaultTimeoutMs:   200,
			MaxRetries:         1,
			EnableMockFallback: true,
			LogLevel:           "info",
			HeartbeatIntervalS: 300, // 5 分钟心跳
		},
		Networks:   networks,
		Placements: sdkPlacements,
		ServerTime: time.Now().UnixMilli(),
	}, nil
}

// InitComplete 处理 SDK 初始化完成上报
// SDK 初始化各网络 SDK 后调用，服务端根据结果调整 Waterfall
func (s *SDKService) InitComplete(ctx context.Context, req *SDKInitCompleteRequest) (*SDKInitCompleteResponse, error) {
	if req.AppID == "" {
		return nil, errors.New(errors.CodeValidationFailed, "app_id 不能为空")
	}

	// 记录初始化结果到追踪日志（用 proxy_win 事件类型复用）
	for _, net := range req.Networks {
		_ = s.trackingRepo.Create(&model.TrackingEventLog{
			RequestID:  fmt.Sprintf("init_%s_%d", req.AppID, time.Now().UnixMilli()),
			MaterialID: net.NetworkType,
			EventType:  fmt.Sprintf("sdk_init_%s", net.Status), // sdk_init_success / sdk_init_failed
			Timestamp:  time.Now().UnixMilli(),
		})
	}

	// 构建失败网络集合
	failedNetworks := make(map[string]bool)
	for _, net := range req.Networks {
		if net.Status != "success" {
			failedNetworks[net.NetworkType] = true
		}
	}

	// 若有网络初始化失败，返回调整后的 Waterfall（移除失败网络）
	if len(failedNetworks) == 0 {
		return &SDKInitCompleteResponse{Message: "all networks initialized successfully"}, nil
	}

	// 重新查询并过滤 Waterfall
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
	// 记录心跳（写入追踪日志）
	_ = s.trackingRepo.Create(&model.TrackingEventLog{
		RequestID:  utils.NewID(),
		MaterialID: req.AppID,
		EventType:  "sdk_heartbeat",
		Timestamp:  time.Now().UnixMilli(),
	})

	currentVersion := s.GetConfigVersion()
	// SDK 上报了缓存版本号，且与服务端不一致，通知更新
	configUpdated := req.ConfigVersion > 0 && req.ConfigVersion != currentVersion

	return &SDKHeartbeatResponse{
		ConfigUpdated: configUpdated,
		ConfigVersion: currentVersion,
		ServerTime:    time.Now().UnixMilli(),
	}, nil
}

// ReportECPM 处理 eCPM 上报（用于优化 Waterfall 排序）
// 使用指数移动平均（EMA）更新广告源的历史 eCPM，驱动 Waterfall 动态排序
func (s *SDKService) ReportECPM(ctx context.Context, req *SDKECPMReportRequest) error {
	if req.PlacementID == "" || req.SourceID == "" {
		return errors.New(errors.CodeValidationFailed, "placement_id 和 source_id 不能为空")
	}
	if req.ECPM < 0 {
		return errors.New(errors.CodeValidationFailed, "ecpm 不能为负数")
	}

	// 只有实际展示的广告才更新 eCPM（避免加载但未展示的数据污染）
	if req.Displayed && req.ECPM > 0 {
		if err := s.sourceRepo.UpdateECPM(req.SourceID, req.ECPM); err != nil {
			// eCPM 更新失败不影响主流程，记录错误后继续
			_ = err
		}
	}

	// 记录 eCPM 上报事件（用于审计和统计分析）
	_ = s.trackingRepo.Create(&model.TrackingEventLog{
		RequestID:  utils.NewID(),
		MaterialID: fmt.Sprintf("%s_%s", req.PlacementID, req.SourceID),
		EventType:  "ecpm_report",
		Timestamp:  time.Now().UnixMilli(),
	})

	return nil
}

// ── 辅助函数 ──────────────────────────────────────────────

// networkAdapterClass 返回各网络对应的 SDK Adapter 类名
// SDK 开发者根据此类名动态加载对应的 Adapter
func networkAdapterClass(networkType string) string {
	classMap := map[string]string{
		// 国际平台
		"admob":         "AdLabAdMobAdapter",
		"applovin":      "AdLabAppLovinAdapter",
		"unity":         "AdLabUnityAdsAdapter",
		"ironsource":    "AdLabIronSourceAdapter",
		"vungle":        "AdLabVungleAdapter",
		"chartboost":    "AdLabChartboostAdapter",
		"inmobi":        "AdLabInMobiAdapter",
		"facebook":      "AdLabFacebookAdapter",
		"digitalturbine": "AdLabDigitalTurbineAdapter",
		"ogury":         "AdLabOguryAdapter",
		"moloco":        "AdLabMolocoAdapter",
		"yandex":        "AdLabYandexAdapter",
		"monetag":       "AdLabMonetagAdapter",
		"adsterra":      "AdLabAdsterraAdapter",
		"propellerads":  "AdLabPropellerAdsAdapter",
		// 国内平台
		"pangle":    "AdLabPangleAdapter",
		"mintegral": "AdLabMintegralAdapter",
		"baidu":     "AdLabBaiduAdapter",
		"tencent":   "AdLabTencentAdapter",
		"kuaishou":  "AdLabKuaishouAdapter",
		"sigmob":    "AdLabSigmobAdapter",
		// 自定义
		"custom": "AdLabBuiltinAdapter",
	}
	if cls, ok := classMap[networkType]; ok {
		return cls
	}
	return "AdLabCustomAdapter"
}

// parseExtraParams 解析扩展参数 JSON 为 map
func parseExtraParams(extraJSON string) map[string]string {
	if extraJSON == "" {
		return nil
	}
	var result map[string]string
	if err := json.Unmarshal([]byte(extraJSON), &result); err != nil {
		return nil
	}
	return result
}

// ── 统一广告请求（保留在此文件，供 handler 使用）────────────

// AdRequest 统一广告请求（SDK 发起）
type AdRequest struct {
	PlacementID string          `json:"placement_id"`
	App         *AdRequestApp   `json:"app,omitempty"`
	Device      *AdRequestDevice `json:"device,omitempty"`
	BidMode     string          `json:"bid_mode,omitempty"`
	FloorPrice  float64         `json:"floor_price,omitempty"`
	TimeoutMs   int             `json:"timeout_ms,omitempty"`
}

// AdRequestApp 广告请求中的应用信息
type AdRequestApp struct {
	AppID    string `json:"app_id,omitempty"`
	BundleID string `json:"bundle_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
	StoreURL string `json:"store_url,omitempty"`
}

// AdRequestDevice 广告请求中的设备信息
type AdRequestDevice struct {
	Platform    string  `json:"platform,omitempty"`
	OSVersion   string  `json:"os_version,omitempty"`
	DeviceModel string  `json:"device_model,omitempty"`
	ScreenW     int     `json:"screen_w,omitempty"`
	ScreenH     int     `json:"screen_h,omitempty"`
	IFA         string  `json:"ifa,omitempty"`
	IFAType     string  `json:"ifa_type,omitempty"`
	IP          string  `json:"ip,omitempty"`
	UA          string  `json:"ua,omitempty"`
	Language    string  `json:"language,omitempty"`
	Carrier     string  `json:"carrier,omitempty"`
	ConnType    string  `json:"conn_type,omitempty"`
	Lat         float64 `json:"lat,omitempty"`
	Lon         float64 `json:"lon,omitempty"`
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
