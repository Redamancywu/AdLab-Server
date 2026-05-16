package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/service"
)

// C2SHandler C2S 上报处理器
type C2SHandler struct {
	c2sService *service.C2SReportingService
}

// NewC2SHandler 创建 C2SHandler
func NewC2SHandler(c2sService *service.C2SReportingService) *C2SHandler {
	return &C2SHandler{c2sService: c2sService}
}

// Report 处理 POST /api/v1/c2s/result
func (h *C2SHandler) Report(c *gin.Context) {
	var req service.C2SReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeC2SDuplicateReport,
			Message: "请求体格式错误",
			Details: err.Error(),
		})
		return
	}

	// 必填字段校验
	if req.RequestID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeC2SDuplicateReport,
			Message: "request_id 不能为空",
		})
		return
	}
	if req.PlacementID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeC2SDuplicateReport,
			Message: "placement_id 不能为空",
		})
		return
	}
	if len(req.BiddingDetails) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeC2SDuplicateReport,
			Message: "bidding_details 不能为空",
		})
		return
	}

	resp, err := h.c2sService.Report(c.Request.Context(), &req)
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
			Code:    apperrors.CodeInternalError,
			Message: "内部服务错误",
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

// DisplayConfirm 处理 POST /api/v1/c2s/display
// SDK 在广告实际展示后调用，触发 Win Notice 代理发送
func (h *C2SHandler) DisplayConfirm(c *gin.Context) {
	var req service.C2SDisplayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeC2SDuplicateReport,
			Message: "请求体格式错误",
			Details: err.Error(),
		})
		return
	}

	if req.RequestID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeC2SDuplicateReport,
			Message: "request_id 不能为空",
		})
		return
	}

	resp, err := h.c2sService.ConfirmDisplay(c.Request.Context(), &req)
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
			Code:    apperrors.CodeInternalError,
			Message: "内部服务错误",
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
