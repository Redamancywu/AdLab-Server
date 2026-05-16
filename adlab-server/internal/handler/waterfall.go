package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/service"
)

// WaterfallHandler Waterfall 竞价处理器
type WaterfallHandler struct {
	waterfallService *service.WaterfallBiddingService
}

// NewWaterfallHandler 创建 WaterfallHandler
func NewWaterfallHandler(waterfallService *service.WaterfallBiddingService) *WaterfallHandler {
	return &WaterfallHandler{waterfallService: waterfallService}
}

// Bid 处理 POST /api/v1/waterfall/bid
func (h *WaterfallHandler) Bid(c *gin.Context) {
	var req service.S2SBidRequest // 复用 S2S 请求结构
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeS2SBidError,
			Message: "请求体格式错误",
			Details: err.Error(),
		})
		return
	}

	if req.PlacementID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeS2SBidError,
			Message: "placement_id 不能为空",
		})
		return
	}

	resp, err := h.waterfallService.Bid(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			if appErr.Code == apperrors.CodeNoValidBid {
				c.Status(http.StatusNoContent)
				return
			}
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
