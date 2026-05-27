package handler

import (
	"net/http"

	"adlab-server/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证相关接口处理器
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler 创建 AuthHandler
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Login 用户登录
//
// POST /auth/login
// Body: { "username": "admin", "password": "admin123" }
// 成功返回 JWT Token
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    4001,
			"message": "参数格式错误: " + err.Error(),
		})
		return
	}

	resp, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    9001,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登录成功",
		"data":    resp,
	})
}

// RefreshToken 刷新 JWT Token
//
// POST /auth/refresh
// Header: Authorization: Bearer <old_token>
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    4001,
			"message": "缺少 Authorization Token",
		})
		return
	}
	// 去掉 "Bearer " 前缀
	tokenString := authHeader[7:]

	resp, err := h.authSvc.RefreshToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    9001,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Token 刷新成功",
		"data":    resp,
	})
}

// Me 获取当前登录用户信息
//
// GET /auth/me
// Header: Authorization: Bearer <token>
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("jwt_user_id")
	username, _ := c.Get("jwt_username")
	role, _ := c.Get("jwt_role")
	tenantID, _ := c.Get("jwt_tenant_id")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"user_id":   userID,
			"username":  username,
			"role":      role,
			"tenant_id": tenantID,
		},
	})
}
