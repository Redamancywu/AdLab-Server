package router

import (
	"github.com/gin-gonic/gin"

	"adlab-server/internal/middleware"
)

func registerSDKRoutes(r *gin.Engine, h *Handlers) {
	registerSDKRoutesWithLimit(r, h, 100, 200)
}

func registerSDKRoutesWithLimit(r *gin.Engine, h *Handlers, rps, burst int) {
	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimit(rps, burst))
	{
		v1.GET("/strategy/:placement_id", h.Strategy.GetStrategy)

		sdk := v1.Group("/sdk")
		{
			sdk.POST("/init", h.SDK.Init)
			sdk.POST("/init_complete", h.SDK.InitComplete)
			sdk.POST("/heartbeat", h.SDK.Heartbeat)
			sdk.POST("/ecpm", h.SDK.ReportECPM)
		}

		v1.POST("/ad/request", h.SDK.RequestAd)
		v1.POST("/s2s/bid", h.S2S.Bid)
		v1.POST("/waterfall/bid", h.Waterfall.Bid)
		v1.POST("/c2s/result", h.C2S.Report)
		v1.POST("/c2s/display", h.C2S.DisplayConfirm)
		v1.GET("/vast/generate", h.VAST.Generate)
		v1.GET("/docs/:key", h.Document.GetDoc)
		v1.GET("/track", h.Track.Track)
		v1.POST("/track", h.Track.Track)
		v1.POST("/mock/fill", h.MockAd.Fill)

		logs := v1.Group("/logs")
		{
			logs.GET("/requests", h.Log.QueryRequests)
			logs.GET("/requests/:request_id", h.Log.GetRequestDetail)
			logs.GET("/requests/:request_id/details", h.Log.GetBidDetails)
			logs.GET("/tracking/:request_id", h.Log.GetTrackingChain)
			logs.GET("/export", h.Log.ExportLogs)
		}

		stats := v1.Group("/stats")
		{
			stats.GET("/overview", h.Stats.GetOverallStats)
			stats.GET("/dsp", h.Stats.GetDSPStats)
			stats.GET("/timeseries", h.Stats.GetTimeSeriesStats)
		}
	}
}
