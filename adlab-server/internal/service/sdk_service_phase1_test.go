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

func TestSDKHeartbeatReusesInitConfigHash(t *testing.T) {
	db := setupSDKServiceTestDB(t)

	appRepo := repository.NewAppRepository(db)
	placementRepo := repository.NewPlacementRepository(db)
	sourceRepo := repository.NewAdSourceRepository(db)
	cfgRepo := repository.NewAppNetworkConfigRepository(db)
	trackingRepo := repository.NewTrackingEventLogRepository(db)

	app := &model.App{
		AppID:              "app_sdk_hash",
		Name:               "SDK Hash App",
		Platform:           "ios",
		BundleID:           "com.example.hash",
		Category:           "game",
		Status:             "active",
		EnableMockFallback: true,
	}
	if err := appRepo.Create(app); err != nil {
		t.Fatalf("create app failed: %v", err)
	}

	if err := cfgRepo.Create(&model.AppNetworkConfig{
		AppID:          app.AppID,
		Platform:       "ios",
		NetworkType:    "admob",
		AdapterClass:   "AdLabAdMobAdapter",
		InitParamsJSON: `{"app_id":"ca-app-pub-hash~123"}`,
		Status:         "active",
	}); err != nil {
		t.Fatalf("create app network config failed: %v", err)
	}

	placement := &model.Placement{
		PlacementID: "plc_sdk_hash",
		AppID:       app.AppID,
		Name:        "SDK Hash Placement",
		AdType:      "rewarded_video",
		Status:      "active",
	}
	if err := placementRepo.Create(placement); err != nil {
		t.Fatalf("create placement failed: %v", err)
	}

	svc := NewSDKService(appRepo, placementRepo, sourceRepo, nil, trackingRepo).WithAppNetworkConfigRepo(cfgRepo)

	initResp, err := svc.Init(context.Background(), &SDKInitRequest{
		AppID:    app.AppID,
		Platform: "ios",
	})
	if err != nil {
		t.Fatalf("sdk init failed: %v", err)
	}

	hbResp, err := svc.Heartbeat(context.Background(), &SDKHeartbeatRequest{
		AppID:         app.AppID,
		Platform:      "ios",
		ConfigVersion: initResp.ConfigVersion,
		ConfigHash:    initResp.ConfigHash,
	})
	if err != nil {
		t.Fatalf("sdk heartbeat failed: %v", err)
	}

	if hbResp.ConfigHash != initResp.ConfigHash {
		t.Fatalf("expected heartbeat config_hash %q to match init %q", hbResp.ConfigHash, initResp.ConfigHash)
	}
	if hbResp.ConfigUpdated {
		t.Fatalf("expected matching version/hash to report no config update")
	}
}

func TestSDKInitCompleteFiltersFailedNetworksFromInstancesAndWaterfall(t *testing.T) {
	db := setupSDKServiceTestDB(t)

	appRepo := repository.NewAppRepository(db)
	placementRepo := repository.NewPlacementRepository(db)
	sourceRepo := repository.NewAdSourceRepository(db)
	cfgRepo := repository.NewAppNetworkConfigRepository(db)
	trackingRepo := repository.NewTrackingEventLogRepository(db)

	app := &model.App{
		AppID:              "app_sdk_init_complete",
		Name:               "SDK Init Complete App",
		Platform:           "ios",
		BundleID:           "com.example.initcomplete",
		Category:           "game",
		Status:             "active",
		EnableMockFallback: true,
	}
	if err := appRepo.Create(app); err != nil {
		t.Fatalf("create app failed: %v", err)
	}

	placement := &model.Placement{
		PlacementID: "plc_sdk_init_complete",
		AppID:       app.AppID,
		Name:        "SDK Init Complete Placement",
		AdType:      "rewarded_video",
		Status:      "active",
	}
	if err := placementRepo.Create(placement); err != nil {
		t.Fatalf("create placement failed: %v", err)
	}

	sourceAdmob := &model.AdSource{
		SourceID:        "src_admob",
		Name:            "AdMob Source",
		BidMode:         "waterfall",
		Priority:        1,
		FloorPrice:      1.1,
		TimeoutMs:       900,
		Status:          "active",
		NetworkType:     "admob",
		HistoricalECPM:  1.8,
		ECPMSampleCount: 3,
	}
	if err := sourceRepo.Create(sourceAdmob); err != nil {
		t.Fatalf("create admob source failed: %v", err)
	}

	sourcePangle := &model.AdSource{
		SourceID:        "src_pangle",
		Name:            "Pangle Source",
		BidMode:         "waterfall",
		Priority:        2,
		FloorPrice:      0.9,
		TimeoutMs:       800,
		Status:          "active",
		NetworkType:     "pangle",
		HistoricalECPM:  1.2,
		ECPMSampleCount: 2,
	}
	if err := sourceRepo.Create(sourcePangle); err != nil {
		t.Fatalf("create pangle source failed: %v", err)
	}

	if err := placementRepo.BindSourceDetailed(repository.BindSourceParams{
		PlacementID:  placement.PlacementID,
		SourceID:     sourceAdmob.SourceID,
		InstanceID:   "ins_admob",
		InstanceName: "AdMob Instance",
		AdUnitID:     "admob-unit",
		Status:       "active",
	}); err != nil {
		t.Fatalf("bind admob source failed: %v", err)
	}

	if err := placementRepo.BindSourceDetailed(repository.BindSourceParams{
		PlacementID:  placement.PlacementID,
		SourceID:     sourcePangle.SourceID,
		InstanceID:   "ins_pangle",
		InstanceName: "Pangle Instance",
		AdUnitID:     "pangle-unit",
		Status:       "active",
	}); err != nil {
		t.Fatalf("bind pangle source failed: %v", err)
	}

	if err := cfgRepo.Create(&model.AppNetworkConfig{
		AppID:          app.AppID,
		Platform:       "ios",
		NetworkType:    "admob",
		AdapterClass:   "AdLabAdMobAdapter",
		InitParamsJSON: `{"app_id":"ca-app-pub-new~123"}`,
		Status:         "active",
	}); err != nil {
		t.Fatalf("create admob app config failed: %v", err)
	}
	if err := cfgRepo.Create(&model.AppNetworkConfig{
		AppID:          app.AppID,
		Platform:       "ios",
		NetworkType:    "pangle",
		AdapterClass:   "AdLabPangleAdapter",
		InitParamsJSON: `{"app_id":"pangle-app-id"}`,
		Status:         "active",
	}); err != nil {
		t.Fatalf("create pangle app config failed: %v", err)
	}

	svc := NewSDKService(appRepo, placementRepo, sourceRepo, nil, trackingRepo).WithAppNetworkConfigRepo(cfgRepo)

	resp, err := svc.InitComplete(context.Background(), &SDKInitCompleteRequest{
		AppID:    app.AppID,
		Platform: "ios",
		Networks: []SDKNetworkInitResult{
			{NetworkType: "admob", Status: "success"},
			{NetworkType: "pangle", Status: "error", ErrorMsg: "init failed"},
		},
	})
	if err != nil {
		t.Fatalf("sdk init_complete failed: %v", err)
	}

	if len(resp.AdjustedPlacements) != 1 {
		t.Fatalf("expected 1 adjusted placement, got %d", len(resp.AdjustedPlacements))
	}
	adjusted := resp.AdjustedPlacements[0]
	if len(adjusted.Waterfall) != 1 {
		t.Fatalf("expected failed network to be removed from waterfall, got %d items", len(adjusted.Waterfall))
	}
	if adjusted.Waterfall[0].NetworkType != "admob" {
		t.Fatalf("expected surviving waterfall network to be admob, got %q", adjusted.Waterfall[0].NetworkType)
	}
	if len(adjusted.Instances) != 1 {
		t.Fatalf("expected failed network to be removed from instances, got %d items", len(adjusted.Instances))
	}
	if adjusted.Instances[0].NetworkType != "admob" {
		t.Fatalf("expected surviving instance network to be admob, got %q", adjusted.Instances[0].NetworkType)
	}
}

func TestSDKInitCompleteReturnsErrorWhenRefreshInitFails(t *testing.T) {
	db := setupSDKServiceTestDB(t)

	appRepo := repository.NewAppRepository(db)
	placementRepo := repository.NewPlacementRepository(db)
	sourceRepo := repository.NewAdSourceRepository(db)
	trackingRepo := repository.NewTrackingEventLogRepository(db)

	svc := NewSDKService(appRepo, placementRepo, sourceRepo, nil, trackingRepo)

	_, err := svc.InitComplete(context.Background(), &SDKInitCompleteRequest{
		AppID:    "missing_app",
		Platform: "ios",
		Networks: []SDKNetworkInitResult{
			{NetworkType: "admob", Status: "error", ErrorMsg: "init failed"},
		},
	})
	if err == nil {
		t.Fatalf("expected init_complete to return error when refresh init fails")
	}
}
