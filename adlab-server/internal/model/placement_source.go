package model

import "time"

// PlacementSource 广告位与广告源的多对多关联表
type PlacementSource struct {
	PlacementID        string    `gorm:"primaryKey;size:100" json:"placement_id"`
	SourceID           string    `gorm:"primaryKey;size:100" json:"source_id"`
	InstanceID         string    `gorm:"size:120;index" json:"instance_id,omitempty"`                 // 实例 ID（绑定级）
	InstanceName       string    `gorm:"size:200" json:"instance_name,omitempty"`                     // 实例名称（绑定级）
	AdUnitID           string    `gorm:"size:500" json:"ad_unit_id,omitempty"`                        // 第三方平台广告位 ID（绑定级）
	TimeoutMsOverride  int       `json:"timeout_ms_override,omitempty"`                               // 超时覆盖（绑定级）
	FloorPriceOverride float64   `json:"floor_price_override,omitempty"`                              // 底价覆盖（绑定级）
	LoadParamsJSON     string    `gorm:"size:4000" json:"load_params_json,omitempty"`                 // 请求级扩展参数（绑定级）
	Status             string    `gorm:"not null;size:20;default:active" json:"status"`              // active / inactive
	CreatedAt          time.Time `json:"created_at,omitempty"`
	UpdatedAt          time.Time `json:"updated_at,omitempty"`
}
