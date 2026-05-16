package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
)

// BidRequestDetail 竞价请求详情（含明细）
type BidRequestDetail struct {
	*model.BidRequestLog
	Details []model.BidDetailLog `json:"details"`
}

// LogQueryFilter 日志查询过滤条件
type LogQueryFilter struct {
	PlacementID string
	BidMode     string
	StartTime   *time.Time
	EndTime     *time.Time
	Page        int
	PageSize    int
}

// PagedResult 分页结果
type PagedResult struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Items    interface{} `json:"items"`
}

// LogService 日志查询服务
type LogService struct {
	bidRequestRepo  *repository.BidRequestLogRepository
	bidDetailRepo   *repository.BidDetailLogRepository
	trackingRepo    *repository.TrackingEventLogRepository
}

// NewLogService 创建 LogService
func NewLogService(
	bidRequestRepo *repository.BidRequestLogRepository,
	bidDetailRepo *repository.BidDetailLogRepository,
	trackingRepo *repository.TrackingEventLogRepository,
) *LogService {
	return &LogService{
		bidRequestRepo: bidRequestRepo,
		bidDetailRepo:  bidDetailRepo,
		trackingRepo:   trackingRepo,
	}
}

// QueryRequests 分页查询竞价请求日志
func (s *LogService) QueryRequests(filter LogQueryFilter) ([]model.BidRequestLog, int64, error) {
	repoFilter := repository.BidRequestLogFilter{
		PlacementID: filter.PlacementID,
		BidMode:     filter.BidMode,
		StartTime:   filter.StartTime,
		EndTime:     filter.EndTime,
		Page:        filter.Page,
		PageSize:    filter.PageSize,
	}
	return s.bidRequestRepo.QueryWithFilter(repoFilter)
}

// GetRequestDetail 获取单次竞价详情（含所有 DSP 明细）
func (s *LogService) GetRequestDetail(requestID string) (*BidRequestDetail, error) {
	log, err := s.bidRequestRepo.FindByRequestID(requestID)
	if err != nil {
		return nil, err
	}

	details, err := s.bidDetailRepo.FindByRequestID(requestID)
	if err != nil {
		return nil, errors.Wrap(errors.CodeLogQueryError, "查询竞价明细失败", err)
	}

	return &BidRequestDetail{
		BidRequestLog: log,
		Details:       details,
	}, nil
}

// GetTrackingChain 获取追踪事件链（按时间戳升序）
func (s *LogService) GetTrackingChain(requestID string) ([]model.TrackingEventLog, error) {
	logs, err := s.trackingRepo.FindByRequestID(requestID)
	if err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		// 不报错，返回空列表
		return []model.TrackingEventLog{}, nil
	}
	return logs, nil
}

// ExportLogs 导出竞价日志为 CSV 格式
// 最多导出 5000 条，超出时在响应头中提示
func (s *LogService) ExportLogs(filter LogQueryFilter) ([]byte, int64, error) {
	const maxExport = 5000
	filter.Page = 1
	filter.PageSize = maxExport

	logs, total, err := s.QueryRequests(filter)
	if err != nil {
		return nil, 0, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{
		"request_id", "placement_id", "bid_mode", "dsp_count",
		"winner_dsp_id", "winner_price", "total_latency_ms", "status", "created_at",
	}
	if err := w.Write(header); err != nil {
		return nil, 0, fmt.Errorf("写 CSV 表头失败: %w", err)
	}

	for _, log := range logs {
		row := []string{
			log.RequestID,
			log.PlacementID,
			log.BidMode,
			fmt.Sprintf("%d", log.DSPCount),
			log.WinnerDSPID,
			fmt.Sprintf("%.4f", log.WinnerPrice),
			fmt.Sprintf("%d", log.TotalLatencyMs),
			log.Status,
			log.CreatedAt.Format(time.RFC3339),
		}
		if err := w.Write(row); err != nil {
			return nil, 0, fmt.Errorf("写 CSV 数据行失败: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, 0, fmt.Errorf("刷新 CSV 缓冲区失败: %w", err)
	}

	return buf.Bytes(), total, nil
}

// GetBidDetails 获取单次竞价的 DSP 明细列表（独立接口）
func (s *LogService) GetBidDetails(requestID string) ([]model.BidDetailLog, error) {
	// 先确认请求存在
	if _, err := s.bidRequestRepo.FindByRequestID(requestID); err != nil {
		return nil, err
	}
	details, err := s.bidDetailRepo.FindByRequestID(requestID)
	if err != nil {
		return nil, errors.Wrap(errors.CodeLogQueryError, "查询竞价明细失败", err)
	}
	return details, nil
}
