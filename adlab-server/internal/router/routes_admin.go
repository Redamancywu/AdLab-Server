package router

import (
	"github.com/gin-gonic/gin"

	"adlab-server/internal/middleware"
)

func registerAdminRoutes(r *gin.Engine, h *Handlers, adminToken string) {
	admin := r.Group("/admin")
	admin.Use(middleware.AdminAuth(adminToken))
	{
		if h.Dashboard != nil {
			admin.GET("/dashboard", h.Dashboard.GetDashboard)
		}

		admin.GET("/placements", h.Admin.ListPlacements)
		admin.POST("/placements", h.Admin.CreatePlacement)
		admin.PUT("/placements/:id", h.Admin.UpdatePlacement)
		admin.DELETE("/placements/:id", h.Admin.DeletePlacement)
		admin.GET("/placements/:id/sources", h.Admin.GetPlacementWithSources)
		admin.POST("/placements/:id/test", h.Admin.TestPlacement)

		admin.GET("/apps", h.Admin.ListApps)
		admin.POST("/apps", h.Admin.CreateApp)
		admin.PUT("/apps/:id", h.Admin.UpdateApp)
		admin.DELETE("/apps/:id", h.Admin.DeleteApp)
		admin.GET("/apps/:id/placements", h.Admin.GetAppWithPlacements)
		admin.GET("/apps/:id/network-configs", h.Admin.ListAppNetworkConfigs)
		admin.POST("/apps/:id/network-configs", h.Admin.CreateAppNetworkConfig)
		admin.PUT("/apps/:id/network-configs/:config_id", h.Admin.UpdateAppNetworkConfig)
		admin.DELETE("/apps/:id/network-configs/:config_id", h.Admin.DeleteAppNetworkConfig)

		admin.GET("/sources", h.Admin.ListSources)
		admin.POST("/sources", h.Admin.CreateSource)
		admin.PUT("/sources/:id", h.Admin.UpdateSource)
		admin.DELETE("/sources/:id", h.Admin.DeleteSource)

		admin.POST("/placement-sources", h.Admin.BindSource)
		admin.PUT("/placement-sources/:instance_id", h.Admin.UpdateBinding)
		admin.DELETE("/placement-sources", h.Admin.UnbindSource)

		admin.GET("/dsp-configs", h.Admin.ListDSPConfigs)
		admin.POST("/dsp-configs", h.Admin.CreateDSPConfig)
		admin.PUT("/dsp-configs/:id", h.Admin.UpdateDSPConfig)

		admin.GET("/materials", h.Admin.ListMaterials)
		admin.POST("/materials", h.Admin.CreateMaterial)
		admin.PUT("/materials/:id", h.Admin.UpdateMaterial)
		admin.DELETE("/materials/:id", h.Admin.DeleteMaterial)

		admin.POST("/scenarios/switch", h.Admin.SwitchScenario)
		admin.GET("/change-logs", h.Admin.ListChangeLogs)
		admin.GET("/export", h.Admin.ExportConfig)
		admin.POST("/import", h.Admin.ImportConfig)
		admin.DELETE("/logs/cleanup", h.Admin.CleanupLogs)
		admin.POST("/seed", h.Admin.SeedData)

		admin.GET("/mock-ads", h.MockAd.ListMockAds)
		admin.POST("/mock-ads", h.MockAd.CreateMockAd)
		admin.GET("/mock-ads/:id", h.MockAd.GetMockAd)
		admin.PUT("/mock-ads/:id", h.MockAd.UpdateMockAd)
		admin.DELETE("/mock-ads/:id", h.MockAd.DeleteMockAd)

		admin.POST("/docs/:key", h.Document.SaveDoc)
	}
}
