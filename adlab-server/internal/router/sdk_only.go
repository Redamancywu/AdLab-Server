package router

import (
	"github.com/gin-gonic/gin"

	"adlab-server/internal/middleware"
)

type SDKOnlyOptions struct {
	EnableDocs   bool
	EnableHealth bool
	EnableLab    bool
	EnableMetrics bool
	RateLimitEnabled bool
	RateLimitRPS     int
	RateLimitBurst   int
}

// SetupSDKOnly 只注册 SDK / public 相关路由，供独立 sdkapi 进程使用
func SetupSDKOnly(r *gin.Engine, h *Handlers, dbPing func() error, opts SDKOnlyOptions) {
	metrics := middleware.NewSDKAPIMetrics()
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())
	r.Use(middleware.AccessLog("sdkapi"))
	r.Use(metrics.Middleware())

	if opts.EnableDocs {
		registerDocsRoutes(r, h)
	}
	if opts.EnableHealth {
		registerHealthRoutes(r, dbPing)
	}
	if opts.RateLimitEnabled {
		registerSDKRoutesWithLimit(r, h, opts.RateLimitRPS, opts.RateLimitBurst)
	} else {
		registerSDKRoutes(r, h)
	}
	if opts.EnableLab {
		registerLabRoutes(r, h)
	}
	if opts.EnableMetrics {
		r.GET("/metrics", metrics.Handler())
	}
}
