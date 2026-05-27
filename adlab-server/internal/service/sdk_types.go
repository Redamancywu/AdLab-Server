package service

import (
	"sync"

	"adlab-server/internal/repository"
)

type SDKInitRequest struct {
	AppID      string         `json:"app_id"`
	SDKVersion string         `json:"sdk_version"`
	Platform   string         `json:"platform"`
	Device     *SDKDeviceInfo `json:"device,omitempty"`
}

type SDKDeviceInfo struct {
	OSVersion   string `json:"os_version,omitempty"`
	DeviceModel string `json:"device_model,omitempty"`
	Language    string `json:"language,omitempty"`
	IFA         string `json:"ifa,omitempty"`
	IFAType     string `json:"ifa_type,omitempty"`
}

type SDKInitResponse struct {
	AppID         string               `json:"app_id"`
	AppName       string               `json:"app_name"`
	BundleID      string               `json:"bundle_id"`
	Platform      string               `json:"platform"`
	ConfigVersion int64                `json:"config_version"`
	ConfigHash    string               `json:"config_hash"`
	Global        SDKGlobalConfig      `json:"global"`
	Networks      []SDKNetworkInit     `json:"networks"`
	Placements    []SDKPlacementConfig `json:"placements"`
	ServerTime    int64                `json:"server_time"`
}

type SDKGlobalConfig struct {
	DefaultTimeoutMs   int    `json:"default_timeout_ms"`
	MaxRetries         int    `json:"max_retries"`
	EnableMockFallback bool   `json:"enable_mock_fallback"`
	LogLevel           string `json:"log_level"`
	HeartbeatIntervalS int    `json:"heartbeat_interval_s"`
}

type SDKNetworkInit struct {
	NetworkType  string                 `json:"network_type"`
	AdapterClass string                 `json:"adapter_class"`
	InitParams   map[string]interface{} `json:"init_params,omitempty"`
	Status       string                 `json:"status"`
}

type SDKPlacementConfig struct {
	PlacementID     string                 `json:"placement_id"`
	AdType          string                 `json:"ad_type"`
	FloorPrice      float64                `json:"floor_price"`
	Status          string                 `json:"status"`
	PlacementParams SDKPlacementParams     `json:"placement_params"`
	Instances       []SDKPlacementInstance `json:"instances"`
	Waterfall       []SDKWaterfallItem     `json:"waterfall,omitempty"`
}

type SDKPlacementParams struct {
	DefaultTimeoutMs  int     `json:"default_timeout_ms"`
	DefaultFloorPrice float64 `json:"default_floor_price"`
	CacheTTLS         int     `json:"cache_ttl_s"`
}

type SDKPlacementInstance struct {
	InstanceID   string                 `json:"instance_id"`
	InstanceName string                 `json:"instance_name,omitempty"`
	SourceID     string                 `json:"source_id"`
	NetworkType  string                 `json:"network_type"`
	BidMode      string                 `json:"bid_mode"`
	AdUnitID     string                 `json:"ad_unit_id,omitempty"`
	Priority     int                    `json:"priority"`
	TimeoutMs    int                    `json:"timeout_ms"`
	FloorPrice   float64                `json:"floor_price"`
	LoadParams   map[string]interface{} `json:"load_params,omitempty"`
	Stats        SDKInstanceStats       `json:"stats"`
	Status       string                 `json:"status"`
}

type SDKInstanceStats struct {
	HistoricalECPM float64 `json:"historical_ecpm"`
	SampleCount    int     `json:"sample_count"`
}

type SDKWaterfallItem struct {
	SourceID       string  `json:"source_id"`
	NetworkType    string  `json:"network_type"`
	BidMode        string  `json:"bid_mode"`
	Priority       int     `json:"priority"`
	FloorPrice     float64 `json:"floor_price"`
	TimeoutMs      int     `json:"timeout_ms"`
	AdUnitID       string  `json:"ad_unit_id,omitempty"`
	HistoricalECPM float64 `json:"historical_ecpm"`
}

type SDKInitCompleteRequest struct {
	AppID      string                 `json:"app_id"`
	SDKVersion string                 `json:"sdk_version"`
	Platform   string                 `json:"platform"`
	Networks   []SDKNetworkInitResult `json:"networks"`
	DurationMs int                    `json:"duration_ms"`
}

type SDKNetworkInitResult struct {
	NetworkType    string `json:"network_type"`
	Status         string `json:"status"`
	ErrorMsg       string `json:"error_msg,omitempty"`
	DurationMs     int    `json:"duration_ms"`
	AdapterVersion string `json:"adapter_version,omitempty"`
}

type SDKInitCompleteResponse struct {
	AdjustedPlacements []SDKPlacementConfig `json:"adjusted_placements,omitempty"`
	Message            string               `json:"message"`
}

type SDKHeartbeatRequest struct {
	AppID            string   `json:"app_id"`
	SDKVersion       string   `json:"sdk_version"`
	Platform         string   `json:"platform"`
	IFA              string   `json:"ifa,omitempty"`
	ActivePlacements []string `json:"active_placements,omitempty"`
	ConfigVersion    int64    `json:"config_version,omitempty"`
	ConfigHash       string   `json:"config_hash,omitempty"`
}

type SDKHeartbeatResponse struct {
	ConfigUpdated bool   `json:"config_updated"`
	ConfigVersion int64  `json:"config_version"`
	ConfigHash    string `json:"config_hash"`
	RefreshReason string `json:"refresh_reason,omitempty"`
	ServerTime    int64  `json:"server_time"`
}

type SDKECPMReportRequest struct {
	AppID       string  `json:"app_id"`
	PlacementID string  `json:"placement_id"`
	SourceID    string  `json:"source_id"`
	NetworkType string  `json:"network_type"`
	ECPM        float64 `json:"ecpm"`
	AdType      string  `json:"ad_type"`
	Displayed   bool    `json:"displayed"`
}

type SDKService struct {
	appRepo              *repository.AppRepository
	placementRepo        *repository.PlacementRepository
	sourceRepo           *repository.AdSourceRepository
	appNetworkConfigRepo *repository.AppNetworkConfigRepository
	dspConfigRepo        *repository.DSPConfigRepository
	trackingRepo         *repository.TrackingEventLogRepository
	configVersion        int64
	configMu             sync.RWMutex
}
