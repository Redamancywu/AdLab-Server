package model

import (
	"time"
)

// DSPConfig 虚拟 DSP 配置模型
type DSPConfig struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceID       string    `gorm:"uniqueIndex;not null;size:100" json:"source_id"` // 关联 AdSource.SourceID
	BidMode        string    `gorm:"not null;size:20;default:fixed" json:"bid_mode"` // fixed / random / probabilistic
	BidValue       float64   `gorm:"default:1.0" json:"bid_value"`                   // fixed 模式出价
	BidMin         float64   `gorm:"default:0.5" json:"bid_min"`                     // random 模式最小出价
	BidMax         float64   `gorm:"default:2.0" json:"bid_max"`                     // random 模式最大出价
	BidProbWeights string    `gorm:"size:1000" json:"bid_prob_weights"`               // probabilistic 模式权重 JSON
	FillRate       float64   `gorm:"not null;default:100" json:"fill_rate"`           // 填充率 0~100
	LatencyMs      int       `gorm:"not null;default:50" json:"latency_ms"`           // 基础延迟 ms
	LatencyJitter  int       `gorm:"not null;default:10" json:"latency_jitter"`       // 延迟抖动 ms
	ErrorRate      float64   `gorm:"not null;default:0" json:"error_rate"`            // 错误率 0~100
	ErrorType      string    `gorm:"size:50" json:"error_type"`                       // http_500 / http_503 / timeout / invalid_json
	SupportWinNotice bool    `gorm:"not null;default:true" json:"support_win_notice"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
