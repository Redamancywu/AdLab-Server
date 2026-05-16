package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/service"
)

// VASTHandler VAST 生成处理器
type VASTHandler struct {
	vastService *service.VASTGeneratorService
}

// NewVASTHandler 创建 VASTHandler
func NewVASTHandler(vastService *service.VASTGeneratorService) *VASTHandler {
	return &VASTHandler{vastService: vastService}
}

// Generate 处理 GET /api/v1/vast/generate
// 查询参数：material_id（必填）、request_id（必填）
func (h *VASTHandler) Generate(c *gin.Context) {
	materialID := c.Query("material_id")
	if materialID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeVASTError,
			Message: "material_id 不能为空",
		})
		return
	}

	requestID := c.Query("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeVASTError,
			Message: "request_id 不能为空",
		})
		return
	}

	// 构建 baseURL（从请求中提取 scheme + host）
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	xmlStr, err := h.vastService.Generate(materialID, requestID, baseURL)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeVASTError,
			Message: "生成 VAST XML 失败",
			Details: err.Error(),
		})
		return
	}

	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xmlStr))
}
