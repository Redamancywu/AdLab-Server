package utils

import (
	"net/http"
	"time"
)

// NewHTTPClient 创建一个带超时配置的自定义 HTTP 客户端
func NewHTTPClient(timeoutMs int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// DefaultHTTPClient 默认 HTTP 客户端（200ms 超时）
var DefaultHTTPClient = NewHTTPClient(200)
