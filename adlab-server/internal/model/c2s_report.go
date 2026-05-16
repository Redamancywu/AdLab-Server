package model

import (
	"time"
)

// C2SReportLog C2S 上报日志
type C2SReportLog struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID       string    `gorm:"uniqueIndex;not null;size:100" json:"request_id"`
	PlacementID     string    `gorm:"not null;size:100;index" json:"placement_id"`
	WinnerDSPID     string    `gorm:"size:100" json:"winner_dsp_id"`
	WinnerPrice     float64   `gorm:"default:0" json:"winner_price"`
	Displayed       bool      `gorm:"not null;default:false" json:"displayed"`
	BiddingDetails  JSONRaw   `gorm:"type:text" json:"bidding_details"` // JSON 数组，存储各 DSP 出价明细
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
