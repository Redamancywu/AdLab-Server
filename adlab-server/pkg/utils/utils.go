// Package utils 提供共享工具函数
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

// fallback 计数器，crypto/rand 失败时使用
var idCounter uint64

// NewID 生成 16 位小写 hex 唯一 ID
// 优先使用 crypto/rand；若失败则 fallback 到时间戳+原子计数器，不 panic
func NewID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// fallback：时间戳(ns) XOR 原子计数器，保证唯一性
		n := uint64(time.Now().UnixNano()) ^ atomic.AddUint64(&idCounter, 1)
		return fmt.Sprintf("%016x", n)
	}
	return hex.EncodeToString(b)
}

// NewUUID 保留兼容性，内部调用 NewID
// Deprecated: 新代码请使用 NewID()
func NewUUID() string {
	return NewID()
}
