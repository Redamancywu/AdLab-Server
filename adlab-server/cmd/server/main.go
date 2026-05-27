package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"adlab-server/internal/bootstrap"
	"adlab-server/internal/router"
)

func main() {
	// ── 1. 加载配置 ──────────────────────────────────────
	configPath := os.Getenv("ADLAB_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := bootstrap.LoadConfig(configPath)
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	db, err := bootstrap.InitDB(cfg)
	if err != nil {
		slog.Error("初始化数据库失败", "error", err)
		os.Exit(1)
	}

	runtime := bootstrap.BuildRuntime(cfg, db)
	defer runtime.Shutdown()

	r := gin.New()
	r.Use(gin.Logger())

	adminToken := os.Getenv("ADLAB_ADMIN_TOKEN") // 空字符串 = 开发模式，跳过鉴权
	router.Setup(r, runtime.Handlers, adminToken, func() error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	})

	// ── 8. 启动 HTTP 服务（支持优雅关闭）────────────────
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 在独立 goroutine 中启动服务
	go func() {
		slog.Info("AdLab Server 启动", "addr", addr, "mode", cfg.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("服务启动失败", "error", err)
			os.Exit(1)
		}
	}()

	// 监听系统信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("收到关闭信号，正在优雅关闭服务...")

	// 给进行中的请求最多 10 秒完成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("服务关闭超时", "error", err)
	} else {
		slog.Info("服务已安全关闭")
	}
}
