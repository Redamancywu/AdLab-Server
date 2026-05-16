package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/service"
)

// SDKHandler SDK 专用处理器
type SDKHandler struct {
	sdkService       *service.SDKService
	adRequestService *service.AdRequestService
}

// NewSDKHandler 创建 SDKHandler
func NewSDKHandler(sdkService *service.SDKService, adRequestService *service.AdRequestService) *SDKHandler {
	return &SDKHandler{
		sdkService:       sdkService,
		adRequestService: adRequestService,
	}
}

// Init 处理 POST /api/v1/sdk/init
// SDK 启动时调用，返回 App 级别网络初始化参数 + 所有广告位 Waterfall 配置
func (h *SDKHandler) Init(c *gin.Context) {
	var req service.SDKInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "请求体格式错误",
			Details: err.Error(),
		})
		return
	}

	if req.AppID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "app_id 不能为空",
		})
		return
	}

	resp, err := h.sdkService.Init(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{Code: appErr.Code, Message: appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "success", Data: resp})
}

// InitComplete 处理 POST /api/v1/sdk/init_complete
// SDK 完成各网络 SDK 初始化后上报结果，服务端返回调整后的 Waterfall（移除初始化失败的网络）
func (h *SDKHandler) InitComplete(c *gin.Context) {
	var req service.SDKInitCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: apperrors.CodeValidationFailed, Message: err.Error()})
		return
	}

	resp, err := h.sdkService.InitComplete(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "success", Data: resp})
}

// Heartbeat 处理 POST /api/v1/sdk/heartbeat
// SDK 定期上报在线状态，服务端返回是否有配置更新
func (h *SDKHandler) Heartbeat(c *gin.Context) {
	var req service.SDKHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: apperrors.CodeValidationFailed, Message: err.Error()})
		return
	}

	resp, err := h.sdkService.Heartbeat(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "success", Data: resp})
}

// ReportECPM 处理 POST /api/v1/sdk/ecpm
// C2S 竞价完成后上报实际 eCPM，用于优化 Waterfall 排序
func (h *SDKHandler) ReportECPM(c *gin.Context) {
	var req service.SDKECPMReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: apperrors.CodeValidationFailed, Message: err.Error()})
		return
	}

	if err := h.sdkService.ReportECPM(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RequestAd 处理 POST /api/v1/ad/request
// 统一广告请求入口：自动选择竞价模式，无填充时 Mock 兜底，返回完整 VAST XML
func (h *SDKHandler) RequestAd(c *gin.Context) {
	var req service.AdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: apperrors.CodeValidationFailed, Message: err.Error()})
		return
	}

	if req.PlacementID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: apperrors.CodeValidationFailed, Message: "placement_id 不能为空"})
		return
	}

	// 从 HTTP 请求中自动补充设备信息
	if req.Device == nil {
		req.Device = &service.AdRequestDevice{}
	}
	if req.Device.IP == "" {
		req.Device.IP = c.ClientIP()
	}
	if req.Device.UA == "" {
		req.Device.UA = c.Request.UserAgent()
	}

	resp, err := h.adRequestService.Request(c.Request.Context(), &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			if appErr.Code == apperrors.CodeNoValidBid {
				c.Status(http.StatusNoContent)
				return
			}
			c.JSON(appErr.HTTPStatus(), ErrorResponse{Code: appErr.Code, Message: appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: apperrors.CodeInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Code: apperrors.CodeSuccess, Message: "success", Data: resp})
}
