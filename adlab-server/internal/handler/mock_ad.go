package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
	"adlab-server/internal/service"
	"adlab-server/pkg/utils"
)

// MockAdHandler Mock 广告处理器
type MockAdHandler struct {
	mockAdService *service.MockAdService
	mockAdRepo    *repository.MockAdRepository
}

// NewMockAdHandler 创建 MockAdHandler
func NewMockAdHandler(mockAdService *service.MockAdService, mockAdRepo *repository.MockAdRepository) *MockAdHandler {
	return &MockAdHandler{mockAdService: mockAdService, mockAdRepo: mockAdRepo}
}

// ─────────────────────────────────────────────
// SDK 侧：广告填充接口
// ─────────────────────────────────────────────

// Fill 处理 POST /api/v1/mock/fill
// SDK 在没有第三方 DSP 时调用，获取本地 Mock 广告
func (h *MockAdHandler) Fill(c *gin.Context) {
	var req service.MockAdFillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "请求体格式错误",
			Details: err.Error(),
		})
		return
	}

	if req.PlacementID == "" && req.AdType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "placement_id 或 ad_type 至少填写一个",
		})
		return
	}

	// 构建 baseURL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	resp, err := h.mockAdService.Fill(&req, baseURL)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			if appErr.Code == apperrors.CodeNoValidBid {
				c.Status(http.StatusNoContent)
				return
			}
			c.JSON(appErr.HTTPStatus(), ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeInternalError,
			Message: "Mock 广告填充失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    resp,
	})
}

// ─────────────────────────────────────────────
// 管理侧：Mock 广告 CRUD
// ─────────────────────────────────────────────

// ListMockAds 处理 GET /admin/mock-ads
func (h *MockAdHandler) ListMockAds(c *gin.Context) {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 20)
	adType := c.Query("ad_type")

	ads, total, err := h.mockAdRepo.FindAll(page, pageSize, adType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeInternalError,
			Message: "查询 Mock 广告失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data: gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"items":     ads,
		},
	})
}

// CreateMockAd 处理 POST /admin/mock-ads
func (h *MockAdHandler) CreateMockAd(c *gin.Context) {
	var ad model.MockAd
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "请求体格式错误",
			Details: err.Error(),
		})
		return
	}

	if ad.Name == "" || ad.AdType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "name、ad_type 为必填字段",
		})
		return
	}
	// 若未提供 ID，自动生成
	if ad.MockAdID == "" {
		ad.MockAdID = utils.NewID()
	}
	if ad.Status == "" {
		ad.Status = "active"
	}
	if ad.Priority == 0 {
		ad.Priority = 100
	}

	if err := h.mockAdRepo.Create(&ad); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{Code: appErr.Code, Message: appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{Code: apperrors.CodeSuccess, Message: "created", Data: ad})
}

// UpdateMockAd 处理 PUT /admin/mock-ads/:id
func (h *MockAdHandler) UpdateMockAd(c *gin.Context) {
	mockAdID := c.Param("id")
	existing, err := h.mockAdRepo.FindByMockAdID(mockAdID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{Code: appErr.Code, Message: appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}

	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: apperrors.CodeValidationFailed, Message: err.Error()})
		return
	}
	existing.MockAdID = mockAdID

	if err := h.mockAdRepo.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "updated", Data: existing})
}

// DeleteMockAd 处理 DELETE /admin/mock-ads/:id
func (h *MockAdHandler) DeleteMockAd(c *gin.Context) {
	mockAdID := c.Param("id")
	if err := h.mockAdRepo.Delete(mockAdID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{Code: appErr.Code, Message: appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "deleted", Data: nil})
}

// GetMockAd 处理 GET /admin/mock-ads/:id
func (h *MockAdHandler) GetMockAd(c *gin.Context) {
	mockAdID := c.Param("id")
	ad, err := h.mockAdRepo.FindByMockAdID(mockAdID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{Code: appErr.Code, Message: appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "success", Data: ad})
}
