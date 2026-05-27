package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"adlab-server/internal/buildinfo"
	apperrors "adlab-server/internal/errors"
)

func registerDocsRoutes(r *gin.Engine, h *Handlers) {
	r.GET("/docs", h.Docs.ServeUI)
	r.GET("/docs/openapi.json", h.Docs.ServeSpec)
}

func registerHealthRoutes(r *gin.Engine, dbPing func() error) {
	r.GET("/health", func(c *gin.Context) {
		dbStatus := "ok"
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

	r.GET("/ready", func(c *gin.Context) {
		if dbPing != nil {
			if err := dbPing(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"code":    apperrors.CodeDatabaseError,
					"message": "数据库未就绪",
					"data":    gin.H{"status": "not_ready", "db": err.Error()},
				})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    apperrors.CodeSuccess,
			"message": "ready",
			"data":    gin.H{"status": "ready"},
		})
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    apperrors.CodeSuccess,
			"message": "version",
			"data": gin.H{
				"version":   buildinfo.Version,
				"git_sha":   buildinfo.GitSHA,
				"component": buildinfo.Component,
			},
		})
	})
}

func registerAuthRoutes(r *gin.Engine, h *Handlers) {
	if h.Auth == nil {
		return
	}
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.RefreshToken)
		auth.GET("/me", h.Auth.Me)
	}
}

func registerLabRoutes(r *gin.Engine, h *Handlers) {
	lab := r.Group("/lab/dsp")
	{
		lab.POST("/:dsp_id/bid", h.DSP.HandleBid)
		lab.POST("/:dsp_id/win", h.DSP.HandleWinNotice)
	}
}
