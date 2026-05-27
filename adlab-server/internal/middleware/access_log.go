package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLog 记录最小访问日志，便于 sdkapi 独立部署时快速观测
func AccessLog(component string) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		requestID, _ := c.Get("request_id")
		slog.Info("http request",
			"component", component,
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency_ms", time.Since(started).Milliseconds(),
			"client_ip", c.ClientIP(),
			"request_id", requestID,
		)
	}
}
