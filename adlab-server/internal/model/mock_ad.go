package model

import "time"

// MockAdType Mock 广告类型
type MockAdType string

const (
	MockAdTypeRewardedVideo MockAdType = "rewarded_video" // 激励视频
	MockAdTypeInterstitial  MockAdType = "interstitial"   // 插屏视频/图片
	MockAdTypeBanner        MockAdType = "banner"         // Banner 图片
	MockAdTypeSplash        MockAdType = "splash"         // 开屏图片/视频
	MockAdTypeNative        MockAdType = "native"         // 原生广告
)

// MockAd Mock 广告素材（用于无第三方 DSP 时的本地广告填充）
type MockAd struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	MockAdID    string     `gorm:"uniqueIndex;not null;size:100" json:"mock_ad_id"`
	Name        string     `gorm:"not null;size:200" json:"name"`
	AdType      string     `gorm:"not null;size:50" json:"ad_type"` // rewarded_video / interstitial / banner / splash / native

	// 视频素材（rewarded_video / interstitial）
	VideoURL    string     `gorm:"size:1000" json:"video_url"`    // 视频文件 URL
	VideoWidth  int        `gorm:"default:0" json:"video_width"`  // 视频宽度（px）
	VideoHeight int        `gorm:"default:0" json:"video_height"` // 视频高度（px）
	DurationSec int        `gorm:"default:30" json:"duration_sec"` // 视频时长（秒）
	SkipAfterSec int       `gorm:"default:5" json:"skip_after_sec"` // 可跳过时间（秒，0=不可跳过）

	// 图片素材（banner / splash / native 封面）
	ImageURL    string     `gorm:"size:1000" json:"image_url"`    // 主图 URL
	ImageWidth  int        `gorm:"default:0" json:"image_width"`
	ImageHeight int        `gorm:"default:0" json:"image_height"`

	// 开屏广告（splash）
	SplashURL   string     `gorm:"size:1000" json:"splash_url"`   // 开屏图/视频 URL
	SplashDurationSec int  `gorm:"default:5" json:"splash_duration_sec"` // 展示时长（秒）

	// 原生广告（native）
	NativeTitle       string `gorm:"size:200" json:"native_title"`
	NativeDescription string `gorm:"size:500" json:"native_description"`
	NativeIconURL     string `gorm:"size:1000" json:"native_icon_url"`
	NativeCallToAction string `gorm:"size:100" json:"native_call_to_action"` // 如"立即下载"

	// 通用字段
	ClickURL    string     `gorm:"size:1000" json:"click_url"`    // 点击跳转 URL
	CPMPrice    float64    `gorm:"default:1.0" json:"cpm_price"`  // 模拟出价（USD CPM）
	Status      string     `gorm:"not null;size:20;default:active" json:"status"` // active / inactive
	Priority    int        `gorm:"default:100" json:"priority"`   // 优先级（数值越小越优先）
	Tags        string     `gorm:"size:500" json:"tags"`          // 标签，逗号分隔，如 "game,casual"

	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
