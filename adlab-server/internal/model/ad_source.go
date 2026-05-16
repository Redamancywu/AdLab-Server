package model

import (
	"time"
)

// ECPMDecayFactor 历史 eCPM 衰减因子（指数移动平均，新样本权重）
// 0.3 表示新样本占 30%，历史均值占 70%，平衡响应速度与稳定性
const ECPMDecayFactor = 0.3

// NetworkType 广告网络类型
type NetworkType string

const (
	// ── 国际主流平台（无需企业资质，个人开发者可申请）──────────────
	NetworkAdMob        NetworkType = "admob"        // Google AdMob — 全球最大移动广告平台
	NetworkAppLovin     NetworkType = "applovin"     // AppLovin MAX — 聚合 + 自有 DSP
	NetworkUnityAds     NetworkType = "unity"        // Unity Ads — 游戏广告首选
	NetworkIronSource   NetworkType = "ironsource"   // ironSource (Unity LevelPlay) — 激励视频强
	NetworkVungle       NetworkType = "vungle"       // Vungle / Liftoff — 视频广告
	NetworkChartboost   NetworkType = "chartboost"   // Chartboost — 游戏内广告
	NetworkInMobi       NetworkType = "inmobi"       // InMobi — 亚太区强势
	NetworkFacebook     NetworkType = "facebook"     // Meta Audience Network — 社交定向
	NetworkDigitalTurbine NetworkType = "digitalturbine" // Digital Turbine (Fyber) — 预装广告
	NetworkOgury        NetworkType = "ogury"        // Ogury — 隐私优先广告
	NetworkMoloco       NetworkType = "moloco"       // Moloco — 机器学习 DSP
	NetworkYandex       NetworkType = "yandex"       // Yandex Ads — 俄语区 / 东欧
	NetworkMonetag      NetworkType = "monetag"      // Monetag — 个人开发者友好，无需审核
	NetworkAdsterra     NetworkType = "adsterra"     // Adsterra — 个人站长友好，多格式
	NetworkPropellerAds NetworkType = "propellerads" // PropellerAds — 个人开发者，Push/Interstitial

	// ── 国内平台（个人开发者可申请）────────────────────────────────
	NetworkPangle      NetworkType = "pangle"      // 穿山甲（字节跳动）— 国内 eCPM 最高
	NetworkMintegral   NetworkType = "mintegral"   // Mintegral（汇量科技）— 出海首选
	NetworkBaidu       NetworkType = "baidu"       // 百度联盟 — 百度搜索流量变现
	NetworkTencent     NetworkType = "tencent"     // 腾讯优量汇 — 微信/QQ 流量
	NetworkKuaishou    NetworkType = "kuaishou"    // 快手广告联盟 — 短视频流量
	NetworkSigmob      NetworkType = "sigmob"      // Sigmob（移动魔方）— 国内激励视频

	// ── 自定义 ──────────────────────────────────────────────────────
	NetworkCustom NetworkType = "custom" // 内置模拟器 / 自定义 OpenRTB DSP
)

// AdSource 广告源模型
type AdSource struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceID   string    `gorm:"uniqueIndex;not null;size:100" json:"source_id"`
	Name       string    `gorm:"not null;size:200" json:"name"`
	BidMode    string    `gorm:"not null;size:20" json:"bid_mode"` // s2s / c2s / waterfall
	Priority   int       `gorm:"not null;default:100" json:"priority"`
	FloorPrice float64   `gorm:"not null;default:0" json:"floor_price"` // USD CPM
	TimeoutMs  int       `gorm:"not null;default:200" json:"timeout_ms"`
	Status     string    `gorm:"not null;size:20;default:active" json:"status"` // active / inactive
	DSPURL     string    `gorm:"size:500" json:"dsp_url"` // S2S 竞价端点 URL（OpenRTB 模式）

	// ── 第三方广告网络配置 ──────────────────────────────
	// SDK 请求策略时，服务端将这些配置下发给 SDK，SDK 用来初始化对应的广告网络 SDK
	NetworkType string `gorm:"size:50;default:custom" json:"network_type"` // 广告网络类型，见 NetworkType 常量
	AppID       string `gorm:"size:500" json:"app_id"`       // 广告网络的 App ID（如 AdMob: ca-app-pub-xxx）
	AppKey      string `gorm:"size:500" json:"app_key"`      // SDK Key / App Key（如 AppLovin SDK Key）
	ExtraParams string `gorm:"size:2000" json:"extra_params"` // 扩展参数 JSON（各网络特有参数）

	// ── eCPM 动态排序 ──────────────────────────────────
	// SDK 通过 /api/v1/sdk/ecpm 上报实际 eCPM，服务端更新此字段
	// Waterfall 排序优先使用 historical_ecpm（降序），ecpm_sample_count 为 0 时退化为 priority 排序
	HistoricalECPM   float64 `gorm:"default:0" json:"historical_ecpm"`    // 历史平均 eCPM（USD CPM）
	ECPMSampleCount  int     `gorm:"default:0" json:"ecpm_sample_count"`  // 样本数量（用于加权平均）
	ECPMUpdatedAt    *time.Time `gorm:"index" json:"ecpm_updated_at,omitempty"` // 最后更新时间

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 一对一关联 DSPConfig（内置模拟器配置）
	DSPConfig *DSPConfig `gorm:"foreignKey:SourceID;references:SourceID" json:"dsp_config,omitempty"`
}
