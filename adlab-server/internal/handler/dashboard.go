package handler

import (
	"net/http"
	"time"

	"adlab-server/internal/service"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 首页概览仪表盘
type DashboardHandler struct {
	statsSvc *service.StatsService
}

// NewDashboardHandler 创建 DashboardHandler
func NewDashboardHandler(statsSvc *service.StatsService) *DashboardHandler {
	return &DashboardHandler{statsSvc: statsSvc}
}

// GetDashboard 获取今日概览统计
//
// GET /admin/dashboard
// 返回今日总请求量、填充率、预估收入、Top DSP 列表等核心指标
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := todayStart.AddDate(0, 0, -1)

	// 今日统计
	todayFilter := service.StatsFilter{
		StartTime: &todayStart,
		EndTime:   &now,
	}

	// 昨日统计（用于同比计算）
	yesterdayFilter := service.StatsFilter{
		StartTime: &yesterday,
		EndTime:   &todayStart,
	}

	todayStats, err := h.statsSvc.GetOverallStats(todayFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    5001,
			"message": "查询今日统计失败: " + err.Error(),
		})
		return
	}

	yesterdayStats, _ := h.statsSvc.GetOverallStats(yesterdayFilter)

	// DSP 维度统计
	dspStats, _ := h.statsSvc.GetDSPStats(todayFilter)

	// 汇总今日核心指标
	var totalReqs, totalFills int64
	var totalRevenue float64

	for _, s := range todayStats {
		totalReqs += s.TotalRequests
		totalFills += s.SuccessCount
		// 预估收入 = 填充数 × 平均出价 / 1000（CPM → 单次收入，USD）
		totalRevenue += float64(s.SuccessCount) * s.AvgBidPrice / 1000.0
	}

	var fillRate float64
	if totalReqs > 0 {
		fillRate = float64(totalFills) / float64(totalReqs) * 100
	}

	// 昨日汇总（用于同比）
	var yesterdayReqs, yesterdayFills int64
	var yesterdayRevenue float64
	for _, s := range yesterdayStats {
		yesterdayReqs += s.TotalRequests
		yesterdayFills += s.SuccessCount
		yesterdayRevenue += float64(s.SuccessCount) * s.AvgBidPrice / 1000.0
	}

	// 同比变化率（防止除零）
	reqChange := calcChangeRate(float64(totalReqs), float64(yesterdayReqs))
	revenueChange := calcChangeRate(totalRevenue, yesterdayRevenue)

	// Top 5 DSP（按请求量排序，前端可自行排序）
	topDSPs := dspStats
	if len(topDSPs) > 5 {
		topDSPs = topDSPs[:5]
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"period": gin.H{
				"start": todayStart.Format("2006-01-02 15:04:05"),
				"end":   now.Format("2006-01-02 15:04:05"),
			},
			"overview": gin.H{
				"total_requests":  totalReqs,
				"total_fills":     totalFills,
				"fill_rate":       fillRate,
				"estimated_revenue_usd": totalRevenue,
				"req_change_pct":  reqChange,
				"rev_change_pct":  revenueChange,
			},
			"top_dsps":        topDSPs,
			"placement_stats": todayStats,
		},
	})
}

// calcChangeRate 计算同比变化率（百分比）
// 若基准值为 0，返回 0（而非 Inf）
func calcChangeRate(current, base float64) float64 {
	if base == 0 {
		return 0
	}
	return (current - base) / base * 100
}
