package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"adlab-server/internal/bootstrap"
	"adlab-server/internal/buildinfo"
	"adlab-server/internal/config"
	"adlab-server/internal/router"
)

func main() {
	buildinfo.Component = "sdkapi"
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
	registerSDKAPIRoutes(r, runtime.Handlers, db, cfg)

	addr := fmt.Sprintf(":%d", resolveSDKAPIPort(cfg))
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		slog.Info(
			"AdLab SDK API 启动",
			"component", "sdkapi",
			"addr", addr,
			"mode", cfg.Server.Mode,
			"docs", cfg.SDKAPI.EnableDocs,
			"lab", cfg.SDKAPI.EnableLab,
			"health", cfg.SDKAPI.EnableHealth,
			"rate_limit_enabled", cfg.SDKAPI.RateLimitEnabled,
			"rate_limit_rps", cfg.SDKAPI.RateLimitRPS,
			"rate_limit_burst", cfg.SDKAPI.RateLimitBurst,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("SDK API 启动失败", "component", "sdkapi", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("收到关闭信号，正在优雅关闭 SDK API...", "component", "sdkapi")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("SDK API 关闭超时", "component", "sdkapi", "error", err)
	} else {
		slog.Info("SDK API 已安全关闭", "component", "sdkapi")
	}
}

func resolveSDKAPIPort(cfg *config.Config) int {
	if port := os.Getenv("ADLAB_SDKAPI_PORT"); port != "" {
		if parsed, err := strconv.Atoi(port); err == nil && parsed > 0 {
			return parsed
		}
	}
	if cfg.SDKAPI.Port > 0 {
		return cfg.SDKAPI.Port
	}
	if cfg.Server.Port == 0 {
		return 8080
	}
	return cfg.Server.Port
}

func registerSDKAPIRoutes(r *gin.Engine, handlers *router.Handlers, db *gorm.DB, cfg *config.Config) {
	router.SetupSDKOnly(r, handlers, func() error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	}, runtimeSDKOptions(cfg))
}

func runtimeSDKOptions(cfg *config.Config) router.SDKOnlyOptions {
	return router.SDKOnlyOptions{
		EnableDocs:       cfg.SDKAPI.EnableDocs,
		EnableHealth:     cfg.SDKAPI.EnableHealth,
		EnableLab:        cfg.SDKAPI.EnableLab,
		EnableMetrics:    true,
		RateLimitEnabled: cfg.SDKAPI.RateLimitEnabled,
		RateLimitRPS:     cfg.SDKAPI.RateLimitRPS,
		RateLimitBurst:   cfg.SDKAPI.RateLimitBurst,
	}
}
