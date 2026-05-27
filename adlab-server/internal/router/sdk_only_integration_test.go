package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"adlab-server/internal/handler"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/internal/service"
)

func TestSetupSDKOnlyInitEndpointWithRuntime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("ADLAB_COMPONENT", "sdkapi")
	t.Setenv("ADLAB_VERSION", "test-version")
	t.Setenv("ADLAB_GIT_SHA", "test-sha")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	if err := model.AutoMigrate(db); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	appRepo := repository.NewAppRepository(db)
	placementRepo := repository.NewPlacementRepository(db)
	sourceRepo := repository.NewAdSourceRepository(db)
	appNetworkConfigRepo := repository.NewAppNetworkConfigRepository(db)
	trackingRepo := repository.NewTrackingEventLogRepository(db)

	app := &model.App{
		AppID:              "app_sdkapi_test",
		Name:               "SDKAPI Test App",
		Platform:           "ios",
		BundleID:           "com.example.sdkapi",
		Category:           "game",
		Status:             "active",
		EnableMockFallback: false,
	}
	if err := appRepo.Create(app); err != nil {
		t.Fatalf("create app failed: %v", err)
	}
	if err := appNetworkConfigRepo.Create(&model.AppNetworkConfig{
		AppID:          app.AppID,
		Platform:       "ios",
		NetworkType:    "admob",
		AdapterClass:   "AdLabAdMobAdapter",
		InitParamsJSON: `{"app_id":"ca-app-pub-test~123"}`,
		Status:         "active",
	}); err != nil {
		t.Fatalf("create app network config failed: %v", err)
	}

	sdkSvc := service.NewSDKService(appRepo, placementRepo, sourceRepo, nil, trackingRepo).
		WithAppNetworkConfigRepo(appNetworkConfigRepo)
	adRequestSvc := service.NewAdRequestService(placementRepo, sourceRepo, nil, nil, appRepo, nil, nil)

	r := gin.New()
	SetupSDKOnly(r, &Handlers{
		Docs:     handler.NewDocsHandler(),
		Strategy: &handler.StrategyHandler{},
		S2S:      &handler.S2SHandler{},
		Waterfall:&handler.WaterfallHandler{},
		C2S:      &handler.C2SHandler{},
		DSP:      &handler.DSPHandler{},
		VAST:     &handler.VASTHandler{},
		Track:    &handler.TrackHandler{},
		Log:      &handler.LogHandler{},
		Stats:    &handler.StatsHandler{},
		MockAd:   &handler.MockAdHandler{},
		Document: handler.NewDocumentHandler(db),
		SDK:      handler.NewSDKHandler(sdkSvc, adRequestSvc),
	}, func() error { return nil }, SDKOnlyOptions{
		EnableDocs:       false,
		EnableHealth:     true,
		EnableLab:        false,
		EnableMetrics:    true,
		RateLimitEnabled: false,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sdk/init", strings.NewReader(`{"app_id":"app_sdkapi_test","platform":"ios"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			AppID    string `json:"app_id"`
			Platform string `json:"platform"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if payload.Data.AppID != "app_sdkapi_test" {
		t.Fatalf("unexpected app_id %q", payload.Data.AppID)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsResp := httptest.NewRecorder()
	r.ServeHTTP(metricsResp, metricsReq)
	if metricsResp.Code != http.StatusOK {
		t.Fatalf("expected /metrics 200, got %d", metricsResp.Code)
	}
	if !strings.Contains(metricsResp.Body.String(), "sdkapi_requests_total") {
		t.Fatalf("expected metrics payload, got %q", metricsResp.Body.String())
	}
	if !strings.Contains(metricsResp.Body.String(), "sdkapi_sdk_init_requests_total") {
		t.Fatalf("expected init request metric, got %q", metricsResp.Body.String())
	}
	if !strings.Contains(metricsResp.Body.String(), "sdkapi_status_200_total") {
		t.Fatalf("expected status metric, got %q", metricsResp.Body.String())
	}

	versionReq := httptest.NewRequest(http.MethodGet, "/version", nil)
	versionResp := httptest.NewRecorder()
	r.ServeHTTP(versionResp, versionReq)
	if versionResp.Code != http.StatusOK {
		t.Fatalf("expected /version 200, got %d", versionResp.Code)
	}
	var versionPayload struct {
		Data struct {
			Version   string `json:"version"`
			GitSHA    string `json:"git_sha"`
			Component string `json:"component"`
		} `json:"data"`
	}
	if err := json.Unmarshal(versionResp.Body.Bytes(), &versionPayload); err != nil {
		t.Fatalf("decode version response failed: %v", err)
	}
	if versionPayload.Data.Component == "" {
		t.Fatalf("expected component in /version response, got %q", versionResp.Body.String())
	}
}
