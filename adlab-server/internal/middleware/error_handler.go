package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery 全局 panic 恢复中间件
// 捕获 panic，记录错误日志，返回 HTTP 500
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "error", r, "stack", string(debug.Stack()))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    9002,
					"message": "服务器内部错误",
					"details": "unexpected panic",
				})
			}
		}()
		c.Next()
	}
}
