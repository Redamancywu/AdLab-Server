package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "adlab-server/internal/errors"
	"adlab-server/internal/service"
)

// StatsHandler 竞价统计处理器
type StatsHandler struct {
	statsService *service.StatsService
}

// NewStatsHandler 创建 StatsHandler
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// GetOverallStats 处理 GET /api/v1/stats/overview
// 按广告位聚合统计：填充率、平均出价、延迟
func (h *StatsHandler) GetOverallStats(c *gin.Context) {
	filter := h.parseFilter(c)
	stats, err := h.statsService.GetOverallStats(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeLogQueryError,
			Message: "查询统计数据失败",
			Details: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    stats,
	})
}

// GetDSPStats 处理 GET /api/v1/stats/dsp
// 按 DSP 聚合统计：出价率、胜出率、平均出价
func (h *StatsHandler) GetDSPStats(c *gin.Context) {
	filter := h.parseFilter(c)
	if dspID := c.Query("dsp_id"); dspID != "" {
		filter.DSPID = dspID
	}
	stats, err := h.statsService.GetDSPStats(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeLogQueryError,
			Message: "查询 DSP 统计数据失败",
			Details: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    stats,
	})
}

// GetTimeSeriesStats 处理 GET /api/v1/stats/timeseries
// 按小时分桶的时间序列统计
func (h *StatsHandler) GetTimeSeriesStats(c *gin.Context) {
	filter := h.parseFilter(c)
	buckets, err := h.statsService.GetTimeSeriesStats(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    apperrors.CodeLogQueryError,
			Message: "查询时间序列统计失败",
			Details: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    buckets,
	})
}

// parseFilter 从 query 参数解析统计过滤条件
func (h *StatsHandler) parseFilter(c *gin.Context) service.StatsFilter {
	filter := service.StatsFilter{
		PlacementID: c.Query("placement_id"),
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
	return filter
}
