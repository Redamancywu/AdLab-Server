package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONRaw 是一个可以存储任意 JSON 数据的类型
type JSONRaw json.RawMessage

// Value 实现 driver.Valuer 接口，将 JSONRaw 序列化为字符串存储
func (j JSONRaw) Value() (driver.Value, error) {
	if j == nil {
		return "null", nil
	}
	return string(j), nil
}

// Scan 实现 sql.Scanner 接口，从数据库读取字符串并反序列化
func (j *JSONRaw) Scan(value interface{}) error {
	if value == nil {
		*j = JSONRaw("null")
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("JSONRaw: 不支持的类型 %T", value)
	}
	*j = JSONRaw(bytes)
	return nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSONRaw) UnmarshalJSON(data []byte) error {
	*j = JSONRaw(data)
	return nil
}

// Material 广告素材模型
type Material struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MaterialID      string    `gorm:"uniqueIndex;not null;size:100" json:"material_id"`
	Name            string    `gorm:"not null;size:200" json:"name"`
	Title           string    `gorm:"size:500" json:"title"`
	Description     string    `gorm:"size:1000" json:"description"`
	ClickThroughURL string    `gorm:"size:1000" json:"click_through_url"`
	MediaFiles      JSONRaw   `gorm:"type:text" json:"media_files"` // JSON 数组，存储媒体文件信息
	IconURL         string    `gorm:"size:1000" json:"icon_url"`
	DurationSec     int       `gorm:"default:30" json:"duration_sec"` // 视频时长（秒），默认 30s
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
