package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/service"
)

// S2SHandler S2S 竞价处理器
type S2SHandler struct {
	s2sService *service.S2SBiddingService
}

// NewS2SHandler 创建 S2SHandler，注入 S2SBiddingService 依赖
func NewS2SHandler(s2sService *service.S2SBiddingService) *S2SHandler {
	return &S2SHandler{
		s2sService: s2sService,
	}
}

// Bid 处理 POST /api/v1/s2s/bid
// 绑定 JSON 请求体到 S2SBidRequest，调用 S2SBiddingService.Bid，返回统一 JSON 响应
func (h *S2SHandler) Bid(c *gin.Context) {
	var req service.S2SBidRequest
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

	resp, err := h.s2sService.Bid(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			// 无有效出价：返回 HTTP 204 No Content
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
		// 未知错误，返回内部错误
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
