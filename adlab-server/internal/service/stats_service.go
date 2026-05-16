package service

import (
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/repository"
)

// BidStats 竞价统计数据
type BidStats struct {
	PlacementID   string  `json:"placement_id,omitempty"`
	DSPID         string  `json:"dsp_id,omitempty"`
	TotalRequests int64   `json:"total_requests"`
	SuccessCount  int64   `json:"success_count"`
	NoFillCount   int64   `json:"no_fill_count"`
	FillRate      float64 `json:"fill_rate"`      // 填充率 %
	AvgBidPrice   float64 `json:"avg_bid_price"`  // 平均出价 USD CPM
	MaxBidPrice   float64 `json:"max_bid_price"`
	MinBidPrice   float64 `json:"min_bid_price"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	WinCount      int64   `json:"win_count"`
	WinRate       float64 `json:"win_rate"` // 胜出率 %（仅 DSP 维度有意义）
}

// StatsFilter 统计查询过滤条件
type StatsFilter struct {
	PlacementID string
	DSPID       string
	StartTime   *time.Time
	EndTime     *time.Time
}

// StatsService 竞价统计服务
type StatsService struct {
	bidRequestRepo *repository.BidRequestLogRepository
	bidDetailRepo  *repository.BidDetailLogRepository
}

// NewStatsService 创建 StatsService
func NewStatsService(
	bidRequestRepo *repository.BidRequestLogRepository,
	bidDetailRepo *repository.BidDetailLogRepository,
) *StatsService {
	return &StatsService{
		bidRequestRepo: bidRequestRepo,
		bidDetailRepo:  bidDetailRepo,
	}
}

// GetOverallStats 获取整体竞价统计（按广告位聚合，SQL GROUP BY）
func (s *StatsService) GetOverallStats(filter StatsFilter) ([]BidStats, error) {
	rows, err := s.bidRequestRepo.AggregateByPlacement(repository.StatsAggFilter{
		PlacementID: filter.PlacementID,
		StartTime:   filter.StartTime,
		EndTime:     filter.EndTime,
	})
	if err != nil {
		return nil, errors.Wrap(errors.CodeLogQueryError, "查询竞价统计失败", err)
	}

	result := make([]BidStats, 0, len(rows))
	for _, r := range rows {
		stat := BidStats{
			PlacementID:   r.PlacementID,
			TotalRequests: r.TotalRequests,
			SuccessCount:  r.SuccessCount,
			NoFillCount:   r.NoFillCount,
			AvgBidPrice:   r.AvgWinnerPrice,
			MaxBidPrice:   r.MaxWinnerPrice,
			MinBidPrice:   r.MinWinnerPrice,
			AvgLatencyMs:  r.AvgLatencyMs,
		}
		if r.TotalRequests > 0 {
			stat.FillRate = float64(r.SuccessCount) / float64(r.TotalRequests) * 100
		}
		result = append(result, stat)
	}
	return result, nil
}

// GetDSPStats 获取 DSP 维度的竞价统计（SQL GROUP BY）
func (s *StatsService) GetDSPStats(filter StatsFilter) ([]BidStats, error) {
	rows, err := s.bidDetailRepo.AggregateByDSP(repository.DSPStatsAggFilter{
		DSPID:     filter.DSPID,
		StartTime: filter.StartTime,
		EndTime:   filter.EndTime,
	})
	if err != nil {
		return nil, errors.Wrap(errors.CodeLogQueryError, "查询 DSP 统计失败", err)
	}

	result := make([]BidStats, 0, len(rows))
	for _, r := range rows {
		stat := BidStats{
			DSPID:         r.DSPID,
			TotalRequests: r.TotalRequests,
			SuccessCount:  r.BidCount,
			NoFillCount:   r.NoBidCount,
			WinCount:      r.WinCount,
			AvgBidPrice:   r.AvgBidPrice,
			MaxBidPrice:   r.MaxBidPrice,
			MinBidPrice:   r.MinBidPrice,
			AvgLatencyMs:  r.AvgLatencyMs,
		}
		if r.TotalRequests > 0 {
			stat.FillRate = float64(r.BidCount) / float64(r.TotalRequests) * 100
		}
		if r.BidCount > 0 {
			stat.WinRate = float64(r.WinCount) / float64(r.BidCount) * 100
		}
		result = append(result, stat)
	}
	return result, nil
}

// TimeSeriesBucket 时间序列统计桶（按小时）
type TimeSeriesBucket struct {
	Hour          string  `json:"hour"` // "2024-01-15 14:00"
	TotalRequests int64   `json:"total_requests"`
	SuccessCount  int64   `json:"success_count"`
	FillRate      float64 `json:"fill_rate"`
	AvgBidPrice   float64 `json:"avg_bid_price"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
}

// GetTimeSeriesStats 获取时间序列统计（SQL GROUP BY strftime）
func (s *StatsService) GetTimeSeriesStats(filter StatsFilter) ([]TimeSeriesBucket, error) {
	rows, err := s.bidRequestRepo.AggregateTimeSeries(repository.StatsAggFilter{
		PlacementID: filter.PlacementID,
		StartTime:   filter.StartTime,
		EndTime:     filter.EndTime,
	})
	if err != nil {
		return nil, errors.Wrap(errors.CodeLogQueryError, "查询时间序列统计失败", err)
	}

	result := make([]TimeSeriesBucket, 0, len(rows))
	for _, r := range rows {
		b := TimeSeriesBucket{
			Hour:          r.Hour,
			TotalRequests: r.TotalRequests,
			SuccessCount:  r.SuccessCount,
			AvgBidPrice:   r.AvgWinnerPrice,
			AvgLatencyMs:  r.AvgLatencyMs,
		}
		if r.TotalRequests > 0 {
			b.FillRate = float64(r.SuccessCount) / float64(r.TotalRequests) * 100
		}
		result = append(result, b)
	}
	return result, nil
}
