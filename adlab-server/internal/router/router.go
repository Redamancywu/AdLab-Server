package router

import (
	"github.com/gin-gonic/gin"

	"adlab-server/internal/handler"
	"adlab-server/internal/middleware"
)

// Handlers 所有处理器的集合，用于依赖注入
type Handlers struct {
	Strategy   *handler.StrategyHandler
	S2S        *handler.S2SHandler
	Waterfall  *handler.WaterfallHandler
	C2S        *handler.C2SHandler
	DSP        *handler.DSPHandler
	VAST       *handler.VASTHandler
	Track      *handler.TrackHandler
	Log        *handler.LogHandler
	Admin      *handler.AdminHandler
	Stats      *handler.StatsHandler
	MockAd     *handler.MockAdHandler
	Docs       *handler.DocsHandler
	SDK        *handler.SDKHandler
	Auth       *handler.AuthHandler      // JWT 登录接口
	Dashboard  *handler.DashboardHandler // 仗表盘接口
	Document   *handler.DocumentHandler  // 新增：开发者文档 CMS 接口
}

// Setup 注册所有路由并应用中间件
// adminToken 为空时跳过管理 API 鉴权（开发模式）
// dbPing 为可选的数据库连通性检查函数，传 nil 则跳过 DB 检查
func Setup(r *gin.Engine, h *Handlers, adminToken string, dbPing func() error) {
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	registerDocsRoutes(r, h)
	registerHealthRoutes(r, dbPing)
	registerAuthRoutes(r, h)
	registerSDKRoutes(r, h)
	registerLabRoutes(r, h)
	registerAdminRoutes(r, h, adminToken)
}
