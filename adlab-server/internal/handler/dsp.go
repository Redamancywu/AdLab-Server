package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/openrtb"
	"adlab-server/internal/service"
)

// DSPHandler 虚拟 DSP 模拟器处理器
type DSPHandler struct {
	dspService      *service.DSPSimulatorService
	trackingService *service.TrackingService
}

// NewDSPHandler 创建 DSPHandler
func NewDSPHandler(dspService *service.DSPSimulatorService, trackingService *service.TrackingService) *DSPHandler {
	return &DSPHandler{dspService: dspService, trackingService: trackingService}
}

// HandleBid 处理 POST /lab/dsp/:dsp_id/bid
func (h *DSPHandler) HandleBid(c *gin.Context) {
	dspID := c.Param("dsp_id")
	if dspID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "dsp_id 不能为空",
		})
		return
	}

	var req openrtb.BidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "请求体格式错误",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.dspService.HandleBid(dspID, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			if appErr.Code == apperrors.CodeEntityNotFound {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Code:    appErr.Code,
					Message: appErr.Message,
				})
				return
			}
			// 模拟 DSP 错误响应
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
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

	c.JSON(http.StatusOK, resp)
}

// HandleWinNotice 处理 POST /lab/dsp/:dsp_id/win
func (h *DSPHandler) HandleWinNotice(c *gin.Context) {
	dspID := c.Param("dsp_id")
	if dspID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "dsp_id 不能为空",
		})
		return
	}

	var req service.WinNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "请求体格式错误",
			Details: err.Error(),
		})
		return
	}

	if err := h.dspService.HandleWinNotice(dspID, &req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			if appErr.Code == apperrors.CodeEntityNotFound {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Code:    appErr.Code,
					Message: appErr.Message,
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeInternalError,
			Message: "内部服务错误",
			Details: err.Error(),
		})
		return
	}

	// 持久化 Win Notice 接收事件到追踪日志
	if h.trackingService != nil && req.RequestID != "" {
		_ = h.trackingService.Track(&service.TrackRequest{
			Event:      "proxy_win",
			RequestID:  req.RequestID,
			MaterialID: dspID, // 用 dspID 作为 material_id 标识来源
			ClientIP:   c.ClientIP(),
		})
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "win notice received",
		Data:    gin.H{"dsp_id": dspID, "request_id": req.RequestID, "price": req.Price},
	})
}
