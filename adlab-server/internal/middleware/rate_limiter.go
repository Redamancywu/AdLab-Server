package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/gin-gonic/gin"
)

// ipLimiter 单个 IP 的限流器实体
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiterStore 全局 IP 限流器存储（懒加载 + 定期清理）
type RateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	rps      rate.Limit
	burst    int
}

// newRateLimiterStore 创建限流器存储并启动定期清理 goroutine
func newRateLimiterStore(rps, burst int) *RateLimiterStore {
	s := &RateLimiterStore{
		limiters: make(map[string]*ipLimiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
	// 每分钟清理 5 分钟内未活跃的 IP 限流器（防止内存泄漏）
	go s.cleanupLoop()
	return s
}

func (s *RateLimiterStore) getLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.limiters[ip]
	if !exists {
		entry = &ipLimiter{
			limiter:  rate.NewLimiter(s.rps, s.burst),
			lastSeen: time.Now(),
		}
		s.limiters[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (s *RateLimiterStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for ip, entry := range s.limiters {
			if time.Since(entry.lastSeen) > 5*time.Minute {
				delete(s.limiters, ip)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimit 按 IP 限流中间件（令牌桶算法）
// rps: 每秒允许通过的请求数
// burst: 令牌桶最大容量（允许短时间突发）
func RateLimit(rps, burst int) gin.HandlerFunc {
	store := newRateLimiterStore(rps, burst)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := store.getLimiter(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    4290,
				"message": "请求频率超过限制，请稍后重试",
			})
			return
		}
		c.Next()
	}
}
