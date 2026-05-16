package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/service"
)

// SuccessResponse 统一成功响应格式
type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse 统一错误响应格式
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// StrategyHandler 策略投放处理器
type StrategyHandler struct {
	strategyService *service.StrategyService
}

// NewStrategyHandler 创建 StrategyHandler，注入 StrategyService 依赖
func NewStrategyHandler(strategyService *service.StrategyService) *StrategyHandler {
	return &StrategyHandler{
		strategyService: strategyService,
	}
}

// GetStrategy 处理 GET /api/v1/strategy/:placement_id
// 提取路径参数 placement_id，调用 StrategyService.GetStrategy，返回统一 JSON 响应
func (h *StrategyHandler) GetStrategy(c *gin.Context) {
	placementID := c.Param("placement_id")
	if placementID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodePlacementNotFound,
			Message: "placement_id 不能为空",
		})
		return
	}

	resp, err := h.strategyService.GetStrategy(c.Request.Context(), placementID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
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
