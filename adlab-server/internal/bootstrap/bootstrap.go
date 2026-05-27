package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
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
	"adlab-server/pkg/logger"
)

type Runtime struct {
	Config      *config.Config
	DB          *gorm.DB
	Handlers    *router.Handlers
	AsyncLogger *service.AsyncLogger
	Shutdown    func()
}

func LoadConfig(configPath string) (*config.Config, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	logger.Init(cfg.Log.Level)
	gin.SetMode(cfg.Server.Mode)
	return cfg, nil
}

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.Open(cfg)
	if err != nil {
		return nil, err
	}
	if err := model.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}
	slog.Info("数据库迁移完成")
	return db, nil
}

func BuildRuntime(cfg *config.Config, db *gorm.DB) *Runtime {
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
	appNetworkConfigRepo := repository.NewAppNetworkConfigRepository(db)
	mockAdRepo := repository.NewMockAdRepository(db)
	userRepo := repository.NewUserRepository(db)
	tenantRepo := repository.NewTenantRepository(db)

	strategySvc := service.NewStrategyService(placementRepo, sourceRepo, dspConfigRepo, appRepo)
	authSvc := service.NewAuthService(userRepo, tenantRepo, cfg.JWT)
	authSvc.EnsureDefaultAdmin(tenantRepo)

	asyncLogger := service.NewAsyncLogger(bidRequestRepo, bidDetailRepo, trackingRepo)
	asyncCtx, asyncCancel := context.WithCancel(context.Background())
	asyncLogger.Start(asyncCtx)

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
	sdkSvc := service.NewSDKService(appRepo, placementRepo, sourceRepo, dspConfigRepo, trackingRepo).
		WithAppNetworkConfigRepo(appNetworkConfigRepo)
	adRequestSvc := service.NewAdRequestService(
		placementRepo, sourceRepo, bidRequestRepo, bidDetailRepo, appRepo, mockAdRepo, materialRepo,
	).WithLocalBaseURL(localBaseURL)

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
		Document:  handler.NewDocumentHandler(db),
		SDK:       handler.NewSDKHandler(sdkSvc, adRequestSvc),
		Auth:      handler.NewAuthHandler(authSvc),
		Dashboard: handler.NewDashboardHandler(statsSvc),
		Admin: handler.NewAdminHandler(
			placementRepo,
			sourceRepo,
			appNetworkConfigRepo,
			dspConfigRepo,
			materialRepo,
			changeLogRepo,
			appRepo,
			strategySvc,
			s2sBiddingSvc,
		).WithSDKService(sdkSvc).WithDB(db),
	}

	return &Runtime{
		Config:      cfg,
		DB:          db,
		Handlers:    handlers,
		AsyncLogger: asyncLogger,
		Shutdown: func() {
			asyncCancel()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			asyncLogger.Shutdown(shutdownCtx)
		},
	}
}
