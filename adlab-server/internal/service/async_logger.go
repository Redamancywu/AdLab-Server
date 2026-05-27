package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"adlab-server/internal/model"
	"adlab-server/internal/repository"
)

const (
	asyncLogChanSize = 10000 // 内存 Channel 缓冲区容量
	asyncBatchSize   = 500   // 每批次最多写入 500 条
	asyncFlushMs     = 100   // 每 100ms 强制刷写一次（即使未达到批次上限）
	asyncWorkers     = 4     // Worker goroutine 数量
)

// bidLogEntry 竞价请求日志条目
type bidLogEntry struct {
	log    *model.BidRequestLog
	detail []*model.BidDetailLog // 关联的 DSP 明细
}

// trackLogEntry 追踪事件日志条目
type trackLogEntry struct {
	log *model.TrackingEventLog
}

// AsyncLogger 异步日志管道
// 采用生产者-消费者模式：业务线程非阻塞写入 Channel，Worker goroutine 批量落盘
type AsyncLogger struct {
	bidCh   chan bidLogEntry
	trackCh chan trackLogEntry

	bidRepo    *repository.BidRequestLogRepository
	detailRepo *repository.BidDetailLogRepository
	trackRepo  *repository.TrackingEventLogRepository

	wg sync.WaitGroup
}

// NewAsyncLogger 创建 AsyncLogger
func NewAsyncLogger(
	bidRepo *repository.BidRequestLogRepository,
	detailRepo *repository.BidDetailLogRepository,
	trackRepo *repository.TrackingEventLogRepository,
) *AsyncLogger {
	return &AsyncLogger{
		bidCh:      make(chan bidLogEntry, asyncLogChanSize),
		trackCh:    make(chan trackLogEntry, asyncLogChanSize),
		bidRepo:    bidRepo,
		detailRepo: detailRepo,
		trackRepo:  trackRepo,
	}
}

// Start 启动异步 Worker，ctx 取消时 Worker 退出
func (l *AsyncLogger) Start(ctx context.Context) {
	// 竞价日志 Worker
	for i := 0; i < asyncWorkers/2; i++ {
		l.wg.Add(1)
		go l.bidWorker(ctx)
	}
	// 追踪事件 Worker
	for i := 0; i < asyncWorkers/2; i++ {
		l.wg.Add(1)
		go l.trackWorker(ctx)
	}
	slog.Info("异步日志管道已启动", "workers", asyncWorkers, "bid_chan_size", asyncLogChanSize)
}

// Shutdown 优雅关闭：等待所有 Channel 消息落盘后退出
func (l *AsyncLogger) Shutdown(ctx context.Context) {
	close(l.bidCh)
	close(l.trackCh)

	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("异步日志管道已安全关闭")
	case <-ctx.Done():
		slog.Warn("异步日志关闭超时，可能有少量日志未落盘")
	}
}

// LogBidRequest 非阻塞写入竞价请求日志（若 Channel 满则降级为同步写入）
func (l *AsyncLogger) LogBidRequest(log *model.BidRequestLog, details ...*model.BidDetailLog) {
	entry := bidLogEntry{log: log, detail: details}
	select {
	case l.bidCh <- entry:
		// 成功写入 Channel，由 Worker 异步处理
	default:
		// Channel 已满（流量过大），降级为同步写入，确保不丢数据
		slog.Warn("异步竞价日志 Channel 已满，降级同步写入", "request_id", log.RequestID)
		l.flushBidEntry(entry)
	}
}

// LogTrackingEvent 非阻塞写入追踪事件日志
func (l *AsyncLogger) LogTrackingEvent(log *model.TrackingEventLog) {
	entry := trackLogEntry{log: log}
	select {
	case l.trackCh <- entry:
	default:
		slog.Warn("异步追踪日志 Channel 已满，降级同步写入", "event", log.EventType)
		l.flushTrackEntry(entry)
	}
}

// ─── Worker goroutines ────────────────────────────────────────────────────────

func (l *AsyncLogger) bidWorker(ctx context.Context) {
	defer l.wg.Done()
	ticker := time.NewTicker(asyncFlushMs * time.Millisecond)
	defer ticker.Stop()

	batch := make([]bidLogEntry, 0, asyncBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		for _, entry := range batch {
			l.flushBidEntry(entry)
		}
		batch = batch[:0]
	}

	for {
		select {
		case entry, ok := <-l.bidCh:
			if !ok {
				// Channel 已关闭，处理剩余数据
				flush()
				return
			}
			batch = append(batch, entry)
			if len(batch) >= asyncBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}

func (l *AsyncLogger) trackWorker(ctx context.Context) {
	defer l.wg.Done()
	ticker := time.NewTicker(asyncFlushMs * time.Millisecond)
	defer ticker.Stop()

	batch := make([]trackLogEntry, 0, asyncBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		for _, entry := range batch {
			l.flushTrackEntry(entry)
		}
		batch = batch[:0]
	}

	for {
		select {
		case entry, ok := <-l.trackCh:
			if !ok {
				flush()
				return
			}
			batch = append(batch, entry)
			if len(batch) >= asyncBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}

// ─── 实际写入操作 ──────────────────────────────────────────────────────────────

func (l *AsyncLogger) flushBidEntry(entry bidLogEntry) {
	if err := l.bidRepo.Create(entry.log); err != nil {
		slog.Error("写入竞价日志失败", "request_id", entry.log.RequestID, "error", err)
	}
	for _, detail := range entry.detail {
		if err := l.detailRepo.Create(detail); err != nil {
			slog.Error("写入竞价明细失败", "request_id", detail.RequestID, "error", err)
		}
	}
}

func (l *AsyncLogger) flushTrackEntry(entry trackLogEntry) {
	if err := l.trackRepo.Create(entry.log); err != nil {
		slog.Error("写入追踪事件失败", "event", entry.log.EventType, "error", err)
	}
}

// Stats 返回当前 Channel 使用情况（用于监控告警）
func (l *AsyncLogger) Stats() map[string]interface{} {
	return map[string]interface{}{
		"bid_chan_len":   len(l.bidCh),
		"bid_chan_cap":   cap(l.bidCh),
		"track_chan_len": len(l.trackCh),
		"track_chan_cap": cap(l.trackCh),
		"workers":        asyncWorkers,
	}
}
