package model

import (
	"time"
)

// 追踪事件类型常量
const (
	EventImpression    = "impression"
	EventClick         = "click"
	EventStart         = "start"
	EventFirstQuartile = "firstQuartile"
	EventMidpoint      = "midpoint"
	EventThirdQuartile = "thirdQuartile"
	EventComplete      = "complete"
	EventMute          = "mute"
	EventUnmute        = "unmute"
	EventPause         = "pause"
	EventResume        = "resume"
	EventSkip          = "skip"
)

// TrackingEventLog 追踪事件日志
type TrackingEventLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID  string    `gorm:"not null;size:100;index" json:"request_id"`
	MaterialID string    `gorm:"not null;size:100" json:"material_id"`
	EventType  string    `gorm:"not null;size:50" json:"event_type"` // impression / click / start / first_quartile / midpoint / third_quartile / complete / mute / unmute / pause / resume / skip
	Timestamp  int64     `gorm:"not null" json:"timestamp"`           // Unix 毫秒时间戳
	ClientIP   string    `gorm:"size:50" json:"client_ip"`
	UserAgent  string    `gorm:"size:500" json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}
