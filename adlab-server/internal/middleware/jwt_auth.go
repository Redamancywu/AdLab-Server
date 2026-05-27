package middleware

import (
	"net/http"
	"strings"

	"adlab-server/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID   = "jwt_user_id"
	ContextKeyTenantID = "jwt_tenant_id"
	ContextKeyUsername = "jwt_username"
	ContextKeyRole     = "jwt_role"
)

// JWTAuth JWT 鉴权中间件
// 从 Authorization: Bearer <token> 提取并验证 JWT，将 Claims 注入 Gin 上下文
func JWTAuth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    9001,
				"message": "缺少 Authorization 请求头",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    9001,
				"message": "Authorization 格式错误，应为 Bearer <token>",
			})
			return
		}

		claims, err := authSvc.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    9001,
				"message": "Token 无效或已过期: " + err.Error(),
			})
			return
		}

		// 注入用户信息到上下文
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyTenantID, claims.TenantID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)

		c.Next()
	}
}

// RequireRole 角色检查中间件（在 JWTAuth 之后使用）
// roles 为允许访问的角色列表，例如 RequireRole("superadmin", "admin")
func RequireRole(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    9003,
				"message": "无权限访问",
			})
			return
		}
		if !roleSet[role.(string)] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    9003,
				"message": "权限不足，需要角色: " + strings.Join(roles, " / "),
			})
			return
		}
		c.Next()
	}
}
