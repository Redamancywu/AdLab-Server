package model

import (
	"time"
)

// BidDetailLog 竞价明细日志（单个 DSP）
type BidDetailLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID  string    `gorm:"not null;size:100;index" json:"request_id"` // 关联 BidRequestLog.RequestID
	DSPID      string    `gorm:"not null;size:100" json:"dsp_id"`
	BidPrice   float64   `gorm:"default:0" json:"bid_price"`
	LatencyMs  int       `gorm:"default:0" json:"latency_ms"`
	Status     string    `gorm:"not null;size:20" json:"status"` // win / lose / no_bid / timeout / error
	ErrorMsg   string    `gorm:"size:500" json:"error_msg"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
