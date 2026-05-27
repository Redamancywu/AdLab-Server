package model

import "time"

// Document 开发者文档实体，支持动态编辑与同步
type Document struct {
	Key       string    `gorm:"primaryKey;size:32" json:"key"` // e.g., ios, android, web, api
	Title     string    `gorm:"not null;size:64" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"` // Markdown 格式内容
	UpdatedAt time.Time `json:"updated_at"`
}
