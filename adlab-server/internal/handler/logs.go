package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/service"
)

// LogHandler 日志查询处理器
type LogHandler struct {
	logService *service.LogService
}

// NewLogHandler 创建 LogHandler
func NewLogHandler(logService *service.LogService) *LogHandler {
	return &LogHandler{logService: logService}
}

// QueryRequests 处理 GET /api/v1/logs/requests
// 支持过滤参数：placement_id、bid_mode、start_time（RFC3339）、end_time（RFC3339）、page、page_size
func (h *LogHandler) QueryRequests(c *gin.Context) {
	filter := service.LogQueryFilter{
		PlacementID: c.Query("placement_id"),
		BidMode:     c.Query("bid_mode"),
		Page:        parseIntQuery(c, "page", 1),
		PageSize:    parseIntQuery(c, "page_size", 20),
	}

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartTime = &t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndTime = &t
		}
	}

	logs, total, err := h.logService.QueryRequests(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeLogQueryError,
			Message: "查询竞价日志失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data: gin.H{
			"total":     total,
			"page":      filter.Page,
			"page_size": filter.PageSize,
			"items":     logs,
		},
	})
}

// GetRequestDetail 处理 GET /api/v1/logs/requests/:request_id
func (h *LogHandler) GetRequestDetail(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "request_id 不能为空",
		})
		return
	}

	detail, err := h.logService.GetRequestDetail(requestID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeLogQueryError,
			Message: "查询竞价详情失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    detail,
	})
}

// GetBidDetails 处理 GET /api/v1/logs/requests/:request_id/details
// 独立返回该请求下所有 DSP 的竞价明细
func (h *LogHandler) GetBidDetails(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "request_id 不能为空",
		})
		return
	}

	details, err := h.logService.GetBidDetails(requestID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeLogQueryError,
			Message: "查询竞价明细失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    details,
	})
}

// GetTrackingChain 处理 GET /api/v1/logs/tracking/:request_id
func (h *LogHandler) GetTrackingChain(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    apperrors.CodeValidationFailed,
			Message: "request_id 不能为空",
		})
		return
	}

	events, err := h.logService.GetTrackingChain(requestID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus(), ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeLogQueryError,
			Message: "查询追踪事件失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    events,
	})
}

// ExportLogs 处理 GET /api/v1/logs/export
func (h *LogHandler) ExportLogs(c *gin.Context) {
	filter := service.LogQueryFilter{
		PlacementID: c.Query("placement_id"),
		BidMode:     c.Query("bid_mode"),
	}

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartTime = &t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndTime = &t
		}
	}

	csvData, total, err := h.logService.ExportLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeLogQueryError,
			Message: "导出日志失败",
			Details: err.Error(),
		})
		return
	}

	filename := fmt.Sprintf("bid_logs_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	// 若总数超过导出上限，在响应头中提示
	if total > 5000 {
		c.Header("X-Total-Count", fmt.Sprintf("%d", total))
		c.Header("X-Export-Limit", "5000")
		c.Header("X-Export-Truncated", "true")
	}
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvData)
}

// parseIntQuery 解析整型 query 参数，失败时返回默认值
func parseIntQuery(c *gin.Context, key string, defaultVal int) int {
	if str := c.Query(key); str != "" {
		if v, err := strconv.Atoi(str); err == nil && v > 0 {
			return v
		}
	}
	return defaultVal
}
