package model

import (
	"time"
)

// ConfigChangeLog 配置变更日志
type ConfigChangeLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	EntityType string    `gorm:"not null;size:50;index" json:"entity_type"` // placement / ad_source / dsp_config / material
	EntityID   string    `gorm:"not null;size:100;index" json:"entity_id"`
	Action     string    `gorm:"not null;size:20" json:"action"` // create / update / delete
	OldValue   JSONRaw   `gorm:"type:text" json:"old_value"`     // 变更前的值（JSON）
	NewValue   JSONRaw   `gorm:"type:text" json:"new_value"`     // 变更后的值（JSON）
	Operator   string    `gorm:"size:100" json:"operator"`       // 操作者（可选）
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}
