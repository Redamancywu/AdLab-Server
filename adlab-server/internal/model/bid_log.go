package model

import (
	"time"
)

// BidRequestLog 竞价请求日志（汇总）
type BidRequestLog struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID      string    `gorm:"uniqueIndex;not null;size:100" json:"request_id"` // UUID
	PlacementID    string    `gorm:"not null;size:100;index" json:"placement_id"`
	BidMode        string    `gorm:"not null;size:20" json:"bid_mode"` // s2s / c2s
	DSPCount       int       `gorm:"not null;default:0" json:"dsp_count"`
	WinnerDSPID    string    `gorm:"size:100" json:"winner_dsp_id"`
	WinnerPrice    float64   `gorm:"default:0" json:"winner_price"`
	TotalLatencyMs int       `gorm:"default:0" json:"total_latency_ms"`
	Status         string    `gorm:"not null;size:20" json:"status"` // success / no_fill / error / timeout
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联明细
	Details []BidDetailLog `gorm:"foreignKey:RequestID;references:RequestID" json:"details,omitempty"`
}
