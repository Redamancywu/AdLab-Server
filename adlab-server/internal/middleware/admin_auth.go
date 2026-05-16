package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuth 管理 API 简单鉴权中间件
// 开发阶段：若配置了 admin_token，则校验 Authorization: Bearer <token>
// 若 adminToken 为空字符串，则跳过鉴权（开发模式）
func AdminAuth(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 未配置 token 时跳过鉴权（开发模式）
		if adminToken == "" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    9002,
				"message": "缺少 Authorization 请求头",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    9002,
				"message": "Authorization 格式错误，应为 Bearer <token>",
			})
			return
		}

		if parts[1] != adminToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    9002,
				"message": "token 无效",
			})
			return
		}

		c.Next()
	}
}
