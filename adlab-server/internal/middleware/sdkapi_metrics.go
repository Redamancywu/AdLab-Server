package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type SDKAPIMetrics struct {
	RequestsTotal     atomic.Uint64
	Requests2xxTotal  atomic.Uint64
	Requests4xxTotal  atomic.Uint64
	Requests5xxTotal  atomic.Uint64
	InitRequestsTotal atomic.Uint64
	AdRequestsTotal   atomic.Uint64
	HealthChecksTotal atomic.Uint64
	ReadyChecksTotal  atomic.Uint64
	VersionRequestsTotal atomic.Uint64
	MetricsRequestsTotal atomic.Uint64
	Status200Total       atomic.Uint64
	Status429Total       atomic.Uint64
	Status500Total       atomic.Uint64
}

func NewSDKAPIMetrics() *SDKAPIMetrics {
	return &SDKAPIMetrics{}
}

func (m *SDKAPIMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		m.RequestsTotal.Add(1)
		switch normalizePath(c.FullPath()) {
		case "/api/v1/sdk/init":
			m.InitRequestsTotal.Add(1)
		case "/api/v1/ad/request":
			m.AdRequestsTotal.Add(1)
		case "/health":
			m.HealthChecksTotal.Add(1)
		case "/ready":
			m.ReadyChecksTotal.Add(1)
		case "/version":
			m.VersionRequestsTotal.Add(1)
		case "/metrics":
			m.MetricsRequestsTotal.Add(1)
		}
		status := c.Writer.Status()
		switch {
		case status >= 200 && status < 300:
			m.Requests2xxTotal.Add(1)
			if status == 200 {
				m.Status200Total.Add(1)
			}
		case status >= 400 && status < 500:
			m.Requests4xxTotal.Add(1)
			if status == 429 {
				m.Status429Total.Add(1)
			}
		case status >= 500:
			m.Requests5xxTotal.Add(1)
			if status == 500 {
				m.Status500Total.Add(1)
			}
		}
	}
}

func (m *SDKAPIMetrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, m.render())
	}
}

func (m *SDKAPIMetrics) render() string {
	return fmt.Sprintf(
		"sdkapi_requests_total %d\nsdkapi_requests_2xx_total %d\nsdkapi_requests_4xx_total %d\nsdkapi_requests_5xx_total %d\nsdkapi_sdk_init_requests_total %d\nsdkapi_ad_request_requests_total %d\nsdkapi_health_requests_total %d\nsdkapi_ready_requests_total %d\nsdkapi_version_requests_total %d\nsdkapi_metrics_requests_total %d\nsdkapi_status_200_total %d\nsdkapi_status_429_total %d\nsdkapi_status_500_total %d\n",
		m.RequestsTotal.Load(),
		m.Requests2xxTotal.Load(),
		m.Requests4xxTotal.Load(),
		m.Requests5xxTotal.Load(),
		m.InitRequestsTotal.Load(),
		m.AdRequestsTotal.Load(),
		m.HealthChecksTotal.Load(),
		m.ReadyChecksTotal.Load(),
		m.VersionRequestsTotal.Load(),
		m.MetricsRequestsTotal.Load(),
		m.Status200Total.Load(),
		m.Status429Total.Load(),
		m.Status500Total.Load(),
	)
}

func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	return strings.TrimSpace(path)
}
