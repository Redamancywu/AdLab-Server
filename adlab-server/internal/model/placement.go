package model

import (
	"time"
)

// Placement 广告位模型
type Placement struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PlacementID string    `gorm:"uniqueIndex;not null;size:100" json:"placement_id"`
	AppID       string    `gorm:"size:100;index" json:"app_id"` // 关联 App.AppID，可为空（兼容旧数据）
	Name        string    `gorm:"not null;size:200" json:"name"`
	AdType      string    `gorm:"not null;size:50" json:"ad_type"` // rewarded_video / interstitial / banner / native
	FloorPrice  float64   `gorm:"default:0" json:"floor_price"`   // 广告位全局底价（USD CPM），0 表示不限制
	Status      string    `gorm:"not null;size:20;default:active" json:"status"` // active / inactive
	BindingCount int      `gorm:"-" json:"binding_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 多对多关联广告源
	Sources           []AdSource         `gorm:"many2many:placement_sources;foreignKey:PlacementID;joinForeignKey:PlacementID;References:SourceID;joinReferences:SourceID" json:"sources,omitempty"`
	PlacementSources  []PlacementSource `gorm:"foreignKey:PlacementID;references:PlacementID" json:"placement_sources,omitempty"`
}
