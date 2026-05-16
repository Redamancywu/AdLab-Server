package database

import (
	"os"

	"adlab-server/internal/config"
)

// LoadConfig 加载数据库相关配置，主程序和脚本共用同一入口。
func LoadConfig() (*config.Config, error) {
	configPath := os.Getenv("ADLAB_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}
	return config.Load(configPath)
}
