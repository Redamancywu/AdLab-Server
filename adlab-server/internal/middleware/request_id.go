package middleware

import (
	"github.com/gin-gonic/gin"

	"adlab-server/pkg/utils"
)

const requestIDKey = "X-Request-ID"

// RequestID 为每个请求生成唯一 ID 并注入上下文和响应头
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDKey)
		if id == "" {
			id = utils.NewID()
		}
		c.Set("request_id", id)
		c.Header(requestIDKey, id)
		c.Next()
	}
}
