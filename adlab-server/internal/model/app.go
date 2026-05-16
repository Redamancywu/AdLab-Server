package model

import "time"

// Platform 客户端平台
type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
	PlatformBoth    Platform = "both"
)

// AppCategory 应用分类
type AppCategory string

const (
	CategoryGame          AppCategory = "game"
	CategoryUtility       AppCategory = "utility"
	CategorySocial        AppCategory = "social"
	CategoryNews          AppCategory = "news"
	CategoryEntertainment AppCategory = "entertainment"
	CategoryShopping      AppCategory = "shopping"
	CategoryFinance       AppCategory = "finance"
	CategoryEducation     AppCategory = "education"
	CategoryOther         AppCategory = "other"
)

// App 应用模型
type App struct {
	ID          uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	AppID       string      `gorm:"uniqueIndex;not null;size:100" json:"app_id"`
	Name        string      `gorm:"not null;size:200" json:"name"`
	Platform    string      `gorm:"not null;size:20" json:"platform"`  // ios / android / both
	BundleID    string      `gorm:"not null;size:200" json:"bundle_id"` // com.example.app
	AppStoreURL string      `gorm:"size:500" json:"app_store_url"`
	Category    string      `gorm:"size:50;default:other" json:"category"`
	Description string      `gorm:"size:1000" json:"description"`
	IconURL     string      `gorm:"size:500" json:"icon_url"`
	Status      string      `gorm:"not null;size:20;default:active" json:"status"` // active / inactive

	// ── Mock 广告兜底控制 ──────────────────────────────────
	// 当所有真实广告源均无填充时，是否启用 Mock 广告兜底
	// 默认 true（开发/测试阶段保证始终有广告展示）
	// 生产环境接入真实广告网络后可设为 false
	EnableMockFallback bool `gorm:"not null;default:true" json:"enable_mock_fallback"`

	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`

	// 关联广告位（一对多）
	Placements []Placement `gorm:"foreignKey:AppID;references:AppID" json:"placements,omitempty"`
}
