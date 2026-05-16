package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/service"
)

// 1x1 透明 GIF 像素（43 字节）
var transparentPixel = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00,
	0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff,
	0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

// TrackHandler 追踪事件处理器
type TrackHandler struct {
	trackingService *service.TrackingService
}

// NewTrackHandler 创建 TrackHandler
func NewTrackHandler(trackingService *service.TrackingService) *TrackHandler {
	return &TrackHandler{trackingService: trackingService}
}

// Track 处理 GET /api/v1/track 和 POST /api/v1/track
func (h *TrackHandler) Track(c *gin.Context) {
	req := &service.TrackRequest{}

	if c.Request.Method == http.MethodPost {
		// POST：从 JSON body 读取
		if err := c.ShouldBindJSON(req); err != nil {
			// 降级：尝试从 query 参数读取
			h.bindFromQuery(c, req)
		}
	} else {
		// GET：从 query 参数读取
		h.bindFromQuery(c, req)
	}

	// 注入客户端信息
	req.ClientIP = c.ClientIP()
	req.UserAgent = c.Request.UserAgent()

	if err := h.trackingService.Track(req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeTrackingError,
			Message: "追踪事件记录失败",
			Details: err.Error(),
		})
		return
	}

	// 成功：GET 请求返回 1x1 透明像素，POST 请求返回 204
	if c.Request.Method == http.MethodGet {
		c.Data(http.StatusOK, "image/gif", transparentPixel)
	} else {
		c.Status(http.StatusNoContent)
	}
}

// bindFromQuery 从 query 参数绑定追踪请求
func (h *TrackHandler) bindFromQuery(c *gin.Context, req *service.TrackRequest) {
	req.Event = c.Query("event")
	req.RequestID = c.Query("request_id")
	req.MaterialID = c.Query("material_id")
	if tsStr := c.Query("ts"); tsStr != "" {
		if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
			req.Timestamp = ts
		}
	}
}
