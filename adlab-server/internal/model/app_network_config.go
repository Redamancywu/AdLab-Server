package model

import "time"

// AppNetworkConfig 应用级广告网络初始化配置
type AppNetworkConfig struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AppID          string    `gorm:"not null;size:100;index:idx_app_platform_network,priority:1" json:"app_id"`
	Platform       string    `gorm:"not null;size:20;index:idx_app_platform_network,priority:2" json:"platform"`
	NetworkType    string    `gorm:"not null;size:50;index:idx_app_platform_network,priority:3" json:"network_type"`
	AdapterClass   string    `gorm:"size:200" json:"adapter_class"`
	InitParamsJSON string    `gorm:"size:4000" json:"init_params_json"`
	MinSDKVersion  string    `gorm:"size:50" json:"min_sdk_version"`
	Status         string    `gorm:"not null;size:20;default:active" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
