package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"adlab-server/internal/handler"
)

func TestSetupSDKOnlyRegistersCoreRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handlers := &Handlers{
		Docs:      handler.NewDocsHandler(),
		Strategy:  &handler.StrategyHandler{},
		S2S:       &handler.S2SHandler{},
		Waterfall: &handler.WaterfallHandler{},
		C2S:       &handler.C2SHandler{},
		DSP:       &handler.DSPHandler{},
		VAST:      &handler.VASTHandler{},
		Track:     &handler.TrackHandler{},
		Log:       &handler.LogHandler{},
		Stats:     &handler.StatsHandler{},
		MockAd:    &handler.MockAdHandler{},
		Document:  &handler.DocumentHandler{},
		SDK:       &handler.SDKHandler{},
	}

	SetupSDKOnly(r, handlers, func() error { return nil }, SDKOnlyOptions{
		EnableDocs:    true,
		EnableHealth:  true,
		EnableLab:     true,
		EnableMetrics: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected /health to be registered, got status %d", resp.Code)
	}

	if !hasRoute(r, http.MethodPost, "/api/v1/sdk/init") {
		t.Fatalf("expected /api/v1/sdk/init route to exist")
	}
	if !hasRoute(r, http.MethodPost, "/api/v1/ad/request") {
		t.Fatalf("expected /api/v1/ad/request route to exist")
	}
	if !hasRoute(r, http.MethodGet, "/metrics") {
		t.Fatalf("expected /metrics route to exist")
	}
	if !hasRoute(r, http.MethodGet, "/version") {
		t.Fatalf("expected /version route to exist")
	}
}

func hasRoute(r *gin.Engine, method, path string) bool {
	for _, route := range r.Routes() {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
