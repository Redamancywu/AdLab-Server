// Package logger 提供基于 slog 的结构化日志
package logger

import (
	"log/slog"
	"os"
)

// LevelFromString 将字符串转换为 slog.Level
func LevelFromString(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Init 初始化全局 slog 默认 handler
func Init(level string) {
	lvl := LevelFromString(level)
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(h))
}
