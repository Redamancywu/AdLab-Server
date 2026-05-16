package repository

import (
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// BidRequestLogFilter 竞价请求日志过滤条件
type BidRequestLogFilter struct {
	PlacementID string
	BidMode     string
	StartTime   *time.Time
	EndTime     *time.Time
	Page        int
	PageSize    int
}

// BidRequestLogRepository 竞价请求日志数据访问层
type BidRequestLogRepository struct {
	db *gorm.DB
}

// NewBidRequestLogRepository 创建 BidRequestLogRepository
func NewBidRequestLogRepository(db *gorm.DB) *BidRequestLogRepository {
	return &BidRequestLogRepository{db: db}
}

// Create 创建竞价请求日志
func (r *BidRequestLogRepository) Create(log *model.BidRequestLog) error {
	if err := r.db.Create(log).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建竞价请求日志失败", err)
	}
	return nil
}

// FindByRequestID 根据 request_id 查询竞价请求日志
func (r *BidRequestLogRepository) FindByRequestID(requestID string) (*model.BidRequestLog, error) {
	var log model.BidRequestLog
	result := r.db.Where("request_id = ?", requestID).First(&log)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeRequestNotFound, "竞价请求不存在: "+requestID)
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询竞价请求日志失败", result.Error)
	}
	return &log, nil
}

// QueryWithFilter 带过滤条件查询竞价请求日志
func (r *BidRequestLogRepository) QueryWithFilter(filter BidRequestLogFilter) ([]model.BidRequestLog, int64, error) {
	var logs []model.BidRequestLog
	var total int64

	query := r.db.Model(&model.BidRequestLog{})

	if filter.PlacementID != "" {
		query = query.Where("placement_id = ?", filter.PlacementID)
	}
	if filter.BidMode != "" {
		query = query.Where("bid_mode = ?", filter.BidMode)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计竞价请求日志数量失败", err)
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "查询竞价请求日志失败", err)
	}

	return logs, total, nil
}

// StatsAggFilter 统计聚合过滤条件
type StatsAggFilter struct {
	PlacementID string
	StartTime   *time.Time
	EndTime     *time.Time
}

// PlacementStatsRow 广告位维度聚合结果行
type PlacementStatsRow struct {
	PlacementID    string
	TotalRequests  int64
	SuccessCount   int64
	NoFillCount    int64
	AvgWinnerPrice float64
	MaxWinnerPrice float64
	MinWinnerPrice float64
	AvgLatencyMs   float64
}

// AggregateByPlacement 按广告位聚合竞价请求统计（SQL GROUP BY）
func (r *BidRequestLogRepository) AggregateByPlacement(filter StatsAggFilter) ([]PlacementStatsRow, error) {
	type rawRow struct {
		PlacementID string  `gorm:"column:placement_id"`
		Total       int64   `gorm:"column:total"`
		Success     int64   `gorm:"column:success_count"`
		NoFill      int64   `gorm:"column:no_fill_count"`
		AvgPrice    float64 `gorm:"column:avg_price"`
		MaxPrice    float64 `gorm:"column:max_price"`
		MinPrice    float64 `gorm:"column:min_price"`
		AvgLatency  float64 `gorm:"column:avg_latency"`
	}

	query := r.db.Model(&model.BidRequestLog{}).
		Select(`
			placement_id,
			COUNT(*) AS total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) AS success_count,
			SUM(CASE WHEN status = 'no_fill' THEN 1 ELSE 0 END) AS no_fill_count,
			AVG(CASE WHEN status = 'success' THEN winner_price ELSE NULL END) AS avg_price,
			MAX(CASE WHEN status = 'success' THEN winner_price ELSE NULL END) AS max_price,
			MIN(CASE WHEN status = 'success' THEN winner_price ELSE NULL END) AS min_price,
			AVG(total_latency_ms) AS avg_latency
		`).
		Group("placement_id")

	if filter.PlacementID != "" {
		query = query.Where("placement_id = ?", filter.PlacementID)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	var rows []rawRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "广告位统计聚合查询失败", err)
	}

	result := make([]PlacementStatsRow, 0, len(rows))
	for _, r := range rows {
		result = append(result, PlacementStatsRow{
			PlacementID:    r.PlacementID,
			TotalRequests:  r.Total,
			SuccessCount:   r.Success,
			NoFillCount:    r.NoFill,
			AvgWinnerPrice: r.AvgPrice,
			MaxWinnerPrice: r.MaxPrice,
			MinWinnerPrice: r.MinPrice,
			AvgLatencyMs:   r.AvgLatency,
		})
	}
	return result, nil
}

// TimeSeriesRow 时间序列聚合结果行
type TimeSeriesRow struct {
	Hour           string
	TotalRequests  int64
	SuccessCount   int64
	AvgWinnerPrice float64
	AvgLatencyMs   float64
}

// AggregateTimeSeries 按小时分桶聚合竞价统计（SQLite strftime）
func (r *BidRequestLogRepository) AggregateTimeSeries(filter StatsAggFilter) ([]TimeSeriesRow, error) {
	type rawRow struct {
		Hour       string  `gorm:"column:hour"`
		Total      int64   `gorm:"column:total"`
		Success    int64   `gorm:"column:success_count"`
		AvgPrice   float64 `gorm:"column:avg_price"`
		AvgLatency float64 `gorm:"column:avg_latency"`
	}

	timeBucketExpr := "strftime('%Y-%m-%d %H:00', created_at)"
	if r.db.Dialector.Name() == "postgres" {
		timeBucketExpr = "to_char(date_trunc('hour', created_at), 'YYYY-MM-DD HH24:00')"
	}

	query := r.db.Model(&model.BidRequestLog{}).
		Select(`
			` + timeBucketExpr + ` AS hour,
			COUNT(*) AS total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) AS success_count,
			AVG(CASE WHEN status = 'success' THEN winner_price ELSE NULL END) AS avg_price,
			AVG(total_latency_ms) AS avg_latency
		`).
		Group("hour").
		Order("hour ASC")

	if filter.PlacementID != "" {
		query = query.Where("placement_id = ?", filter.PlacementID)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	var rows []rawRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "时间序列统计聚合查询失败", err)
	}

	result := make([]TimeSeriesRow, 0, len(rows))
	for _, r := range rows {
		result = append(result, TimeSeriesRow{
			Hour:           r.Hour,
			TotalRequests:  r.Total,
			SuccessCount:   r.Success,
			AvgWinnerPrice: r.AvgPrice,
			AvgLatencyMs:   r.AvgLatency,
		})
	}
	return result, nil
}
