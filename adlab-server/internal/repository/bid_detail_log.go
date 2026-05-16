package repository

import (
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// BidDetailLogRepository 竞价明细日志数据访问层
type BidDetailLogRepository struct {
	db *gorm.DB
}

// NewBidDetailLogRepository 创建 BidDetailLogRepository
func NewBidDetailLogRepository(db *gorm.DB) *BidDetailLogRepository {
	return &BidDetailLogRepository{db: db}
}

// Create 创建竞价明细日志
func (r *BidDetailLogRepository) Create(log *model.BidDetailLog) error {
	if err := r.db.Create(log).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建竞价明细日志失败", err)
	}
	return nil
}

// FindByRequestID 根据 request_id 查询所有竞价明细日志
func (r *BidDetailLogRepository) FindByRequestID(requestID string) ([]model.BidDetailLog, error) {
	var logs []model.BidDetailLog
	result := r.db.Where("request_id = ?", requestID).Find(&logs)
	if result.Error != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询竞价明细日志失败", result.Error)
	}
	return logs, nil
}

// DSPStatsAggFilter DSP 统计聚合过滤条件
type DSPStatsAggFilter struct {
	DSPID     string
	StartTime *time.Time
	EndTime   *time.Time
}

// DSPStatsRow DSP 维度聚合结果行
type DSPStatsRow struct {
	DSPID         string
	TotalRequests int64
	BidCount      int64   // 有出价（win + lose）
	WinCount      int64
	NoBidCount    int64
	AvgBidPrice   float64
	MaxBidPrice   float64
	MinBidPrice   float64
	AvgLatencyMs  float64
}

// AggregateByDSP 按 DSP 聚合竞价明细统计（SQL GROUP BY，避免内存聚合）
func (r *BidDetailLogRepository) AggregateByDSP(filter DSPStatsAggFilter) ([]DSPStatsRow, error) {
	type rawRow struct {
		DSPID        string  `gorm:"column:dsp_id"`
		Total        int64   `gorm:"column:total"`
		BidCount     int64   `gorm:"column:bid_count"`
		WinCount     int64   `gorm:"column:win_count"`
		NoBidCount   int64   `gorm:"column:no_bid_count"`
		AvgPrice     float64 `gorm:"column:avg_price"`
		MaxPrice     float64 `gorm:"column:max_price"`
		MinPrice     float64 `gorm:"column:min_price"`
		AvgLatency   float64 `gorm:"column:avg_latency"`
	}

	query := r.db.Model(&model.BidDetailLog{}).
		Select(`
			dsp_id,
			COUNT(*) AS total,
			SUM(CASE WHEN status IN ('win','lose') THEN 1 ELSE 0 END) AS bid_count,
			SUM(CASE WHEN status = 'win' THEN 1 ELSE 0 END) AS win_count,
			SUM(CASE WHEN status = 'no_bid' THEN 1 ELSE 0 END) AS no_bid_count,
			AVG(CASE WHEN status IN ('win','lose') THEN bid_price ELSE NULL END) AS avg_price,
			MAX(CASE WHEN status IN ('win','lose') THEN bid_price ELSE NULL END) AS max_price,
			MIN(CASE WHEN status IN ('win','lose') THEN bid_price ELSE NULL END) AS min_price,
			AVG(latency_ms) AS avg_latency
		`).
		Group("dsp_id")

	if filter.DSPID != "" {
		query = query.Where("dsp_id = ?", filter.DSPID)
	}
	// 通过 bid_request_logs 关联时间过滤（JOIN）
	if filter.StartTime != nil || filter.EndTime != nil {
		query = query.
			Joins("JOIN bid_request_logs ON bid_request_logs.request_id = bid_detail_logs.request_id")
		if filter.StartTime != nil {
			query = query.Where("bid_request_logs.created_at >= ?", filter.StartTime)
		}
		if filter.EndTime != nil {
			query = query.Where("bid_request_logs.created_at <= ?", filter.EndTime)
		}
	}

	var rows []rawRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "DSP 统计聚合查询失败", err)
	}

	result := make([]DSPStatsRow, 0, len(rows))
	for _, r := range rows {
		result = append(result, DSPStatsRow{
			DSPID:         r.DSPID,
			TotalRequests: r.Total,
			BidCount:      r.BidCount,
			WinCount:      r.WinCount,
			NoBidCount:    r.NoBidCount,
			AvgBidPrice:   r.AvgPrice,
			MaxBidPrice:   r.MaxPrice,
			MinBidPrice:   r.MinPrice,
			AvgLatencyMs:  r.AvgLatency,
		})
	}
	return result, nil
}
