package service

import (
	"log/slog"
	"sync"
	"time"
)

// 断路器状态常量
const (
	cbStateClosed   = "closed"    // 正常，允许请求通过
	cbStateOpen     = "open"      // 熔断，拒绝所有请求
	cbStateHalfOpen = "half_open" // 半开，允许少量探测请求
)

// cbConfig 断路器配置参数
type cbConfig struct {
	windowDuration  time.Duration // 统计窗口时长（默认 5s）
	failureThreshold float64      // 失败率阈值（默认 50%）
	minRequests     int           // 窗口内最少请求数才触发熔断（默认 5）
	openDuration    time.Duration // 熔断持续时长（默认 30s）
	halfOpenTests   int           // 半开状态允许通过的探测请求数（默认 3）
}

// cbWindow 时间窗口内的请求统计
type cbWindow struct {
	success  int
	failure  int
	resetAt  time.Time
}

// cbEntry 单个 DSP 的断路器状态
type cbEntry struct {
	state      string
	window     cbWindow
	openedAt   time.Time // 进入熔断状态的时间
	halfTests  int       // 半开状态已通过的探测请求数
	mu         sync.Mutex
}

// CircuitBreaker 多 DSP 断路器管理器
type CircuitBreaker struct {
	entries map[string]*cbEntry
	mu      sync.RWMutex
	cfg     cbConfig
}

// NewCircuitBreaker 创建断路器管理器（使用默认配置）
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		entries: make(map[string]*cbEntry),
		cfg: cbConfig{
			windowDuration:   5 * time.Second,
			failureThreshold: 0.5, // 50% 失败率触发熔断
			minRequests:      5,
			openDuration:     30 * time.Second,
			halfOpenTests:    3,
		},
	}
}

// IsOpen 判断指定 DSP 的断路器是否处于熔断（Open）状态
// 返回 true 表示该 DSP 当前应被跳过
func (cb *CircuitBreaker) IsOpen(dspID string) bool {
	entry := cb.getOrCreate(dspID)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	switch entry.state {
	case cbStateClosed:
		return false
	case cbStateOpen:
		// 检查是否到达半开探测时间
		if time.Since(entry.openedAt) >= cb.cfg.openDuration {
			entry.state = cbStateHalfOpen
			entry.halfTests = 0
			slog.Info("断路器进入半开状态", "dsp_id", dspID)
			return false // 允许探测请求通过
		}
		return true // 仍在熔断期，拒绝
	case cbStateHalfOpen:
		if entry.halfTests < cb.cfg.halfOpenTests {
			return false // 允许探测请求
		}
		return true // 超出探测配额
	}
	return false
}

// RecordSuccess 记录一次成功的 DSP 请求
func (cb *CircuitBreaker) RecordSuccess(dspID string) {
	entry := cb.getOrCreate(dspID)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	cb.refreshWindow(entry)

	switch entry.state {
	case cbStateClosed:
		entry.window.success++
	case cbStateHalfOpen:
		entry.halfTests++
		entry.window.success++
		// 达到半开探测成功次数阈值，关闭断路器
		if entry.halfTests >= cb.cfg.halfOpenTests {
			entry.state = cbStateClosed
			entry.window = cbWindow{resetAt: time.Now().Add(cb.cfg.windowDuration)}
			slog.Info("断路器恢复关闭状态", "dsp_id", dspID)
		}
	}
}

// RecordFailure 记录一次失败的 DSP 请求（超时/502/无填充等）
func (cb *CircuitBreaker) RecordFailure(dspID string) {
	entry := cb.getOrCreate(dspID)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	cb.refreshWindow(entry)

	switch entry.state {
	case cbStateClosed:
		entry.window.failure++
		total := entry.window.success + entry.window.failure
		if total >= cb.cfg.minRequests {
			failRate := float64(entry.window.failure) / float64(total)
			if failRate >= cb.cfg.failureThreshold {
				// 触发熔断
				entry.state = cbStateOpen
				entry.openedAt = time.Now()
				slog.Warn("断路器触发熔断",
					"dsp_id", dspID,
					"fail_rate", failRate,
					"total_requests", total,
					"open_duration", cb.cfg.openDuration,
				)
			}
		}
	case cbStateHalfOpen:
		// 半开探测失败，重新打开断路器
		entry.state = cbStateOpen
		entry.openedAt = time.Now()
		slog.Warn("断路器半开探测失败，重新熔断", "dsp_id", dspID)
	}
}

// Status 返回所有 DSP 断路器的当前状态（用于监控/健康检查）
func (cb *CircuitBreaker) Status() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	result := make(map[string]interface{}, len(cb.entries))
	for dspID, entry := range cb.entries {
		entry.mu.Lock()
		total := entry.window.success + entry.window.failure
		var failRate float64
		if total > 0 {
			failRate = float64(entry.window.failure) / float64(total)
		}
		result[dspID] = map[string]interface{}{
			"state":      entry.state,
			"success":    entry.window.success,
			"failure":    entry.window.failure,
			"fail_rate":  failRate,
			"opened_at":  entry.openedAt,
		}
		entry.mu.Unlock()
	}
	return result
}

// ─── 内部辅助方法 ──────────────────────────────────────────────────────────────

func (cb *CircuitBreaker) getOrCreate(dspID string) *cbEntry {
	cb.mu.RLock()
	entry, ok := cb.entries[dspID]
	cb.mu.RUnlock()
	if ok {
		return entry
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()
	// 双重检查
	if entry, ok = cb.entries[dspID]; ok {
		return entry
	}
	entry = &cbEntry{
		state: cbStateClosed,
		window: cbWindow{
			resetAt: time.Now().Add(cb.cfg.windowDuration),
		},
	}
	cb.entries[dspID] = entry
	return entry
}

// refreshWindow 检查并重置已过期的时间窗口
func (cb *CircuitBreaker) refreshWindow(entry *cbEntry) {
	if time.Now().After(entry.window.resetAt) {
		entry.window = cbWindow{
			resetAt: time.Now().Add(cb.cfg.windowDuration),
		}
	}
}
