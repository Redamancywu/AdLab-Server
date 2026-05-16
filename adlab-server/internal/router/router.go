package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/handler"
	"adlab-server/internal/middleware"
)

// Handlers 所有处理器的集合，用于依赖注入
type Handlers struct {
	Strategy  *handler.StrategyHandler
	S2S       *handler.S2SHandler
	Waterfall *handler.WaterfallHandler
	C2S       *handler.C2SHandler
	DSP       *handler.DSPHandler
	VAST      *handler.VASTHandler
	Track     *handler.TrackHandler
	Log       *handler.LogHandler
	Admin     *handler.AdminHandler
	Stats     *handler.StatsHandler
	MockAd    *handler.MockAdHandler
	Docs      *handler.DocsHandler
	SDK       *handler.SDKHandler
}

// Setup 注册所有路由并应用中间件
// adminToken 为空时跳过管理 API 鉴权（开发模式）
// dbPing 为可选的数据库连通性检查函数，传 nil 则跳过 DB 检查
func Setup(r *gin.Engine, h *Handlers, adminToken string, dbPing func() error) {
	// ── 全局中间件 ──────────────────────────────────────
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	// ── API 文档（Swagger UI）────────────────────────────
	r.GET("/docs", h.Docs.ServeUI)
	r.GET("/docs/openapi.json", h.Docs.ServeSpec)

	// ── 健康检查 ────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {		dbStatus := "ok"
		if dbPing != nil {
			if err := dbPing(); err != nil {
				dbStatus = "error: " + err.Error()
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"code":    apperrors.CodeDatabaseError,
					"message": "数据库连接异常",
					"data":    gin.H{"status": "unhealthy", "db": dbStatus},
				})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    apperrors.CodeSuccess,
			"message": "ok",
			"data":    gin.H{"status": "healthy", "db": dbStatus},
		})
	})
	// ── SDK API：/api/v1/* ───────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// 策略投放（按广告位获取配置，兼容旧接口）
		v1.GET("/strategy/:placement_id", h.Strategy.GetStrategy)

		// ── SDK 专用接口 ──────────────────────────────────
		sdk := v1.Group("/sdk")
		{
			// 1. SDK 初始化：获取 App 级别网络参数 + 所有广告位 Waterfall 配置
			sdk.POST("/init", h.SDK.Init)
			// 2. 初始化完成上报：上报各网络初始化结果，服务端返回调整后的 Waterfall
			sdk.POST("/init_complete", h.SDK.InitComplete)
			// 3. 心跳上报：定期上报在线状态，检查配置是否有更新
			sdk.POST("/heartbeat", h.SDK.Heartbeat)
			// 4. eCPM 上报：C2S 竞价完成后上报实际 eCPM，优化 Waterfall 排序
			sdk.POST("/ecpm", h.SDK.ReportECPM)
		}

		// 统一广告请求：自动选择竞价模式，无填充时 Mock 兜底，返回完整 VAST XML
		v1.POST("/ad/request", h.SDK.RequestAd)

		// S2S 竞价
		v1.POST("/s2s/bid", h.S2S.Bid)

		// Waterfall 竞价（按优先级顺序，有填充即停）
		v1.POST("/waterfall/bid", h.Waterfall.Bid)

		// C2S 上报
		v1.POST("/c2s/result", h.C2S.Report)
		v1.POST("/c2s/display", h.C2S.DisplayConfirm) // 展示确认接口

		// VAST 生成
		v1.GET("/vast/generate", h.VAST.Generate)

		// 追踪事件（支持 GET 和 POST）
		v1.GET("/track", h.Track.Track)
		v1.POST("/track", h.Track.Track)

		// Mock 广告填充（无第三方 DSP 时使用）
		v1.POST("/mock/fill", h.MockAd.Fill)

		// 日志查询
		logs := v1.Group("/logs")
		{
			logs.GET("/requests", h.Log.QueryRequests)
			logs.GET("/requests/:request_id", h.Log.GetRequestDetail)
			logs.GET("/requests/:request_id/details", h.Log.GetBidDetails) // 独立 DSP 明细接口
			logs.GET("/tracking/:request_id", h.Log.GetTrackingChain)
			logs.GET("/export", h.Log.ExportLogs)
		}

		// 竞价统计报表
		stats := v1.Group("/stats")
		{
			stats.GET("/overview", h.Stats.GetOverallStats)
			stats.GET("/dsp", h.Stats.GetDSPStats)
			stats.GET("/timeseries", h.Stats.GetTimeSeriesStats)
		}
	}

	// ── 虚拟 DSP 模拟器：/lab/dsp/* ─────────────────────
	lab := r.Group("/lab/dsp")
	{
		lab.POST("/:dsp_id/bid", h.DSP.HandleBid)
		lab.POST("/:dsp_id/win", h.DSP.HandleWinNotice)
	}

	// ── 管理 API：/admin/* ──────────────────────────────
	admin := r.Group("/admin")
	admin.Use(middleware.AdminAuth(adminToken))
	{
		// 广告位 CRUD
		admin.GET("/placements", h.Admin.ListPlacements)
		admin.POST("/placements", h.Admin.CreatePlacement)
		admin.PUT("/placements/:id", h.Admin.UpdatePlacement)
		admin.DELETE("/placements/:id", h.Admin.DeletePlacement)
		admin.GET("/placements/:id/sources", h.Admin.GetPlacementWithSources)
		admin.POST("/placements/:id/test", h.Admin.TestPlacement) // 一键测试竞价

		// 应用 CRUD
		admin.GET("/apps", h.Admin.ListApps)
		admin.POST("/apps", h.Admin.CreateApp)
		admin.PUT("/apps/:id", h.Admin.UpdateApp)
		admin.DELETE("/apps/:id", h.Admin.DeleteApp)
		admin.GET("/apps/:id/placements", h.Admin.GetAppWithPlacements)

		// 广告源 CRUD
		admin.GET("/sources", h.Admin.ListSources)
		admin.POST("/sources", h.Admin.CreateSource)
		admin.PUT("/sources/:id", h.Admin.UpdateSource)
		admin.DELETE("/sources/:id", h.Admin.DeleteSource)

		// 广告位-广告源绑定/解绑
		admin.POST("/placement-sources", h.Admin.BindSource)
		admin.DELETE("/placement-sources", h.Admin.UnbindSource)

		// DSP 配置 CRUD
		admin.GET("/dsp-configs", h.Admin.ListDSPConfigs)
		admin.POST("/dsp-configs", h.Admin.CreateDSPConfig)
		admin.PUT("/dsp-configs/:id", h.Admin.UpdateDSPConfig)

		// 素材 CRUD
		admin.GET("/materials", h.Admin.ListMaterials)
		admin.POST("/materials", h.Admin.CreateMaterial)
		admin.PUT("/materials/:id", h.Admin.UpdateMaterial)
		admin.DELETE("/materials/:id", h.Admin.DeleteMaterial)

		// 场景切换
		admin.POST("/scenarios/switch", h.Admin.SwitchScenario)

		// 配置变更日志
		admin.GET("/change-logs", h.Admin.ListChangeLogs)

		// 配置导出/导入
		admin.GET("/export", h.Admin.ExportConfig)
		admin.POST("/import", h.Admin.ImportConfig)

		// 日志清理
		admin.DELETE("/logs/cleanup", h.Admin.CleanupLogs)

		// 演示数据初始化
		admin.POST("/seed", h.Admin.SeedData)

		// Mock 广告管理
		admin.GET("/mock-ads", h.MockAd.ListMockAds)
		admin.POST("/mock-ads", h.MockAd.CreateMockAd)
		admin.GET("/mock-ads/:id", h.MockAd.GetMockAd)
		admin.PUT("/mock-ads/:id", h.MockAd.UpdateMockAd)
		admin.DELETE("/mock-ads/:id", h.MockAd.DeleteMockAd)
	}
}
