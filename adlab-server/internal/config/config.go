package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 顶层配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	SDKAPI    SDKAPIConfig    `mapstructure:"sdkapi"`
	Database  DatabaseConfig  `mapstructure:"database"`
	DSP       DSPConfig       `mapstructure:"dsp"`
	Log       LogConfig       `mapstructure:"log"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// JWTConfig JWT 鉴权配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`      // 签名密钥（生产环境请使用强随机值）
	ExpireHour int    `mapstructure:"expire_hour"` // Token 有效时长（小时）
}

// RateLimitConfig API 限流配置
type RateLimitConfig struct {
	Enabled    bool `mapstructure:"enabled"`     // 是否开启限流
	RPS        int  `mapstructure:"rps"`         // 每 IP 每秒最大请求数
	Burst      int  `mapstructure:"burst"`       // 令牌桶突发容量
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `mapstructure:"level"` // debug | info | warn | error
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port int    `mapstructure:"port"` // 监听端口，默认 8080
	Mode string `mapstructure:"mode"` // gin 模式：debug / release / test
}

// SDKAPIConfig SDK API 独立入口配置
type SDKAPIConfig struct {
	Port             int  `mapstructure:"port"`               // 独立 sdkapi 端口，0 表示跟随 server.port
	EnableDocs       bool `mapstructure:"enable_docs"`        // 是否暴露 docs
	EnableLab        bool `mapstructure:"enable_lab"`         // 是否暴露 lab/dsp 模拟器
	EnableHealth     bool `mapstructure:"enable_health"`      // 是否暴露 health
	RateLimitEnabled bool `mapstructure:"rate_limit_enabled"` // 是否开启 sdkapi 独立限流
	RateLimitRPS     int  `mapstructure:"rate_limit_rps"`     // sdkapi 每 IP 每秒最大请求数
	RateLimitBurst   int  `mapstructure:"rate_limit_burst"`   // sdkapi 令牌桶突发容量
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `mapstructure:"type"`     // sqlite3 | postgres
	Path     string `mapstructure:"path"`     // SQLite 文件路径，如 "adlab.db"
	DSN      string `mapstructure:"dsn"`      // 可选：完整 DSN，优先级最高
	Host     string `mapstructure:"host"`     // PostgreSQL 主机
	Port     int    `mapstructure:"port"`     // PostgreSQL 端口
	User     string `mapstructure:"user"`     // PostgreSQL 用户
	Password string `mapstructure:"password"` // PostgreSQL 密码
	DBName   string `mapstructure:"dbname"`   // PostgreSQL 数据库名
	SSLMode  string `mapstructure:"sslmode"`  // PostgreSQL sslmode
	Timezone string `mapstructure:"timezone"` // PostgreSQL 时区
}

// DSPConfig DSP 默认参数
type DSPConfig struct {
	DefaultTimeoutMs int `mapstructure:"default_timeout_ms"` // 默认超时时间
	MaxConcurrency   int `mapstructure:"max_concurrency"`    // 最大并发请求数
}

// Load 加载配置文件，支持环境变量覆盖
// 环境变量前缀为 "ADLAB"，层级用 "_" 分隔
// 例如：ADLAB_SERVER_PORT=9090 会覆盖 server.port
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// 环境变量映射：ADLAB_SERVER_PORT → server.port
	v.SetEnvPrefix("ADLAB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 设置默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.SDKAPI.Port == 0 {
		cfg.SDKAPI.Port = cfg.Server.Port
	}
	if !v.IsSet("sdkapi.enable_docs") {
		cfg.SDKAPI.EnableDocs = true
	}
	if !v.IsSet("sdkapi.enable_lab") {
		cfg.SDKAPI.EnableLab = true
	}
	if !v.IsSet("sdkapi.enable_health") {
		cfg.SDKAPI.EnableHealth = true
	}
	if !v.IsSet("sdkapi.rate_limit_enabled") {
		cfg.SDKAPI.RateLimitEnabled = true
	}
	if cfg.SDKAPI.RateLimitRPS == 0 {
		if cfg.RateLimit.RPS > 0 {
			cfg.SDKAPI.RateLimitRPS = cfg.RateLimit.RPS
		} else {
			cfg.SDKAPI.RateLimitRPS = 100
		}
	}
	if cfg.SDKAPI.RateLimitBurst == 0 {
		if cfg.RateLimit.Burst > 0 {
			cfg.SDKAPI.RateLimitBurst = cfg.RateLimit.Burst
		} else {
			cfg.SDKAPI.RateLimitBurst = 200
		}
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "adlab.db"
	}
	if cfg.Database.Type == "" {
		cfg.Database.Type = "sqlite3"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Database.Timezone == "" {
		cfg.Database.Timezone = "Asia/Shanghai"
	}
	if cfg.DSP.DefaultTimeoutMs == 0 {
		cfg.DSP.DefaultTimeoutMs = 200
	}
	if cfg.DSP.MaxConcurrency == 0 {
		cfg.DSP.MaxConcurrency = 10
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	// JWT 默认值
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "adlab-dev-secret-change-in-production"
	}
	if cfg.JWT.ExpireHour == 0 {
		cfg.JWT.ExpireHour = 24
	}
	// 限流默认值
	if cfg.RateLimit.RPS == 0 {
		cfg.RateLimit.RPS = 100
	}
	if cfg.RateLimit.Burst == 0 {
		cfg.RateLimit.Burst = 200
	}

	return &cfg, nil
}
