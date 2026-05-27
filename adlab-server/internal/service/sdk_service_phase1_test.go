package service

import (
	"context"
	"testing"

	"adlab-server/internal/model"
	"adlab-server/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSDKServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	if err := model.AutoMigrate(db); err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

func TestSDKInitPrefersAppNetworkConfigAndEmitsInstances(t *testing.T) {
	db := setupSDKServiceTestDB(t)

	appRepo := repository.NewAppRepository(db)
	placementRepo := repository.NewPlacementRepository(db)
	sourceRepo := repository.NewAdSourceRepository(db)
	cfgRepo := repository.NewAppNetworkConfigRepository(db)
	trackingRepo := repository.NewTrackingEventLogRepository(db)

	app := &model.App{
		AppID:              "app_sdk_snapshot",
		Name:               "SDK Snapshot App",
		Platform:           "ios",
		BundleID:           "com.example.snapshot",
		Category:           "game",
		Status:             "active",
		EnableMockFallback: false,
	}
	if err := appRepo.Create(app); err != nil {
		t.Fatalf("create app failed: %v", err)
	}

	if err := cfgRepo.Create(&model.AppNetworkConfig{
		AppID:          app.AppID,
		Platform:       "ios",
		NetworkType:    "admob",
		AdapterClass:   "AdLabAdMobAdapter",
		InitParamsJSON: `{"app_id":"ca-app-pub-new~123"}`,
		Status:         "active",
	}); err != nil {
		t.Fatalf("create app network config failed: %v", err)
	}

	src := &model.AdSource{
		SourceID:        "src_sdk_snapshot",
		Name:            "Snapshot Source",
		BidMode:         "waterfall",
		Priority:        10,
		FloorPrice:      1.1,
		TimeoutMs:       900,
		Status:          "active",
		NetworkType:     "admob",
		AppID:           "legacy_should_not_win",
		HistoricalECPM:  2.2,
		ECPMSampleCount: 5,
	}
	if err := sourceRepo.Create(src); err != nil {
		t.Fatalf("create source failed: %v", err)
	}

	placement := &model.Placement{
		PlacementID: "plc_sdk_snapshot",
		AppID:       app.AppID,
		Name:        "SDK Snapshot Placement",
		AdType:      "rewarded_video",
		Status:      "active",
	}
	if err := placementRepo.Create(placement); err != nil {
		t.Fatalf("create placement failed: %v", err)
	}

	if err := placementRepo.BindSourceDetailed(repository.BindSourceParams{
		PlacementID:        placement.PlacementID,
		SourceID:           src.SourceID,
		InstanceID:         "ins_sdk_snapshot",
		InstanceName:       "Snapshot Instance",
		AdUnitID:           "ca-app-pub-xxx/yyy",
		TimeoutMsOverride:  1200,
		FloorPriceOverride: 1.4,
		LoadParamsJSON:     `{"orientation":"portrait"}`,
		Status:             "active",
	}); err != nil {
		t.Fatalf("bind detailed source failed: %v", err)
	}

	svc := NewSDKService(appRepo, placementRepo, sourceRepo, nil, trackingRepo).WithAppNetworkConfigRepo(cfgRepo)

	resp, err := svc.Init(context.Background(), &SDKInitRequest{
		AppID:    app.AppID,
		Platform: "ios",
	})
	if err != nil {
		t.Fatalf("sdk init failed: %v", err)
	}

	if len(resp.Networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(resp.Networks))
	}
	if resp.Networks[0].InitParams["app_id"] != "ca-app-pub-new~123" {
		t.Fatalf("expected app network config to win, got %#v", resp.Networks[0].InitParams)
	}
	if len(resp.Placements) != 1 {
		t.Fatalf("expected 1 placement, got %d", len(resp.Placements))
	}
	if len(resp.Placements[0].Instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(resp.Placements[0].Instances))
	}
	instance := resp.Placements[0].Instances[0]
	if instance.InstanceID != "ins_sdk_snapshot" {
		t.Fatalf("unexpected instance_id: %q", instance.InstanceID)
	}
	if instance.TimeoutMs != 1200 {
		t.Fatalf("expected timeout override to apply, got %d", instance.TimeoutMs)
	}
	if instance.FloorPrice != 1.4 {
		t.Fatalf("expected floor override to apply, got %f", instance.FloorPrice)
	}
	if instance.LoadParams["orientation"] != "portrait" {
		t.Fatalf("expected load params to be parsed, got %#v", instance.LoadParams)
	}
	if len(resp.Placements[0].Waterfall) != 1 {
		t.Fatalf("expected compatibility waterfall to be emitted")
	}
}
