package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"adlab-server/internal/config"
	"adlab-server/internal/database"
	"adlab-server/internal/handler"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/internal/router"
	"adlab-server/internal/service"
)

func main() {
	// ── 1. 加载配置 ──────────────────────────────────────
	configPath := os.Getenv("ADLAB_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// ── 2. 设置 Gin 运行模式 ─────────────────────────────
	gin.SetMode(cfg.Server.Mode)

	// ── 3. 初始化数据库 ──────────────────────────────────
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// ── 4. 初始化 Repository 层 ──────────────────────────
	placementRepo := repository.NewPlacementRepository(db)
	sourceRepo := repository.NewAdSourceRepository(db)
	dspConfigRepo := repository.NewDSPConfigRepository(db)
	materialRepo := repository.NewMaterialRepository(db)
	bidRequestRepo := repository.NewBidRequestLogRepository(db)
	bidDetailRepo := repository.NewBidDetailLogRepository(db)
	trackingRepo := repository.NewTrackingEventLogRepository(db)
	c2sReportRepo := repository.NewC2SReportLogRepository(db)
	changeLogRepo := repository.NewConfigChangeLogRepository(db)
	appRepo := repository.NewAppRepository(db)
	mockAdRepo := repository.NewMockAdRepository(db)

	// ── 5. 初始化 Service 层 ─────────────────────────────
	strategySvc := service.NewStrategyService(placementRepo, sourceRepo, dspConfigRepo, appRepo)

	// 从配置读取本地服务地址，用于 S2S 回退到内置 DSP 模拟器
	localBaseURL := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	s2sBiddingSvc := service.NewS2SBiddingService(
		placementRepo, sourceRepo, bidRequestRepo, bidDetailRepo, appRepo,
	).WithLocalBaseURL(localBaseURL)

	waterfallSvc := service.NewWaterfallBiddingService(
		placementRepo, sourceRepo, bidRequestRepo, bidDetailRepo,
	).WithLocalBaseURL(localBaseURL)

	c2sReportingSvc := service.NewC2SReportingService(c2sReportRepo, dspConfigRepo, sourceRepo)
	dspSimulatorSvc := service.NewDSPSimulatorService(dspConfigRepo, materialRepo)
	vastGeneratorSvc := service.NewVASTGeneratorService(materialRepo)
	trackingSvc := service.NewTrackingService(trackingRepo)
	logSvc := service.NewLogService(bidRequestRepo, bidDetailRepo, trackingRepo)
	statsSvc := service.NewStatsService(bidRequestRepo, bidDetailRepo)
	mockAdSvc := service.NewMockAdService(mockAdRepo, placementRepo)
	sdkSvc := service.NewSDKService(appRepo, placementRepo, sourceRepo, dspConfigRepo, trackingRepo)
	adRequestSvc := service.NewAdRequestService(
		placementRepo, sourceRepo, bidRequestRepo, bidDetailRepo, appRepo, mockAdRepo, materialRepo,
	).WithLocalBaseURL(localBaseURL)

	// ── 6. 初始化 Handler 层 ─────────────────────────────
	handlers := &router.Handlers{
		Strategy:  handler.NewStrategyHandler(strategySvc),
		S2S:       handler.NewS2SHandler(s2sBiddingSvc),
		Waterfall: handler.NewWaterfallHandler(waterfallSvc),
		C2S:       handler.NewC2SHandler(c2sReportingSvc),
		DSP:       handler.NewDSPHandler(dspSimulatorSvc, trackingSvc),
		VAST:      handler.NewVASTHandler(vastGeneratorSvc),
		Track:     handler.NewTrackHandler(trackingSvc),
		Log:       handler.NewLogHandler(logSvc),
		Stats:     handler.NewStatsHandler(statsSvc),
		MockAd:    handler.NewMockAdHandler(mockAdSvc, mockAdRepo),
		Docs:      handler.NewDocsHandler(),
		SDK:       handler.NewSDKHandler(sdkSvc, adRequestSvc),
		Admin: handler.NewAdminHandler(
			placementRepo,
			sourceRepo,
			dspConfigRepo,
			materialRepo,
			changeLogRepo,
			appRepo,
			strategySvc,   // 注入 strategySvc 用于主动失效缓存
			s2sBiddingSvc, // 注入 s2sBiddingSvc 用于广告位测试
		).WithSDKService(sdkSvc).WithDB(db), // 注入 sdkSvc 用于配置版本号递增，db 用于日志清理
	}

	// ── 7. 初始化 Gin 引擎并注册路由 ─────────────────────
	r := gin.New()
	r.Use(gin.Logger())

	adminToken := os.Getenv("ADLAB_ADMIN_TOKEN") // 空字符串 = 开发模式，跳过鉴权
	router.Setup(r, handlers, adminToken, func() error {
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
		log.Printf("AdLab Server 启动，监听地址: %s，模式: %s", addr, cfg.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 监听系统信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("收到关闭信号，正在优雅关闭服务...")

	// 给进行中的请求最多 10 秒完成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务关闭超时: %v", err)
	} else {
		log.Println("服务已安全关闭")
	}
}

// initDB 初始化数据库连接并执行自动迁移
func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.Open(cfg)
	if err != nil {
		return nil, err
	}

	// 执行自动迁移，创建/更新所有表结构
	if err := model.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}
	log.Println("数据库迁移完成")

	return db, nil
}
