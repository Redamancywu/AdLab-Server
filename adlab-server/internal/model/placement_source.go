package model

import "time"

// PlacementSource 广告位与广告源的多对多关联表
type PlacementSource struct {
	PlacementID string    `gorm:"primaryKey;size:100" json:"placement_id"`
	SourceID    string    `gorm:"primaryKey;size:100" json:"source_id"`
	AdUnitID    string    `gorm:"size:500" json:"ad_unit_id,omitempty"`                         // 第三方平台广告位 ID（绑定级）
	Status      string    `gorm:"not null;size:20;default:active" json:"status"`               // active / inactive
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
