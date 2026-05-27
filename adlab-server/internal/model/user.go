package model

import "time"

// Tenant 租户（开发者账户），每个 Tenant 代表一个接入平台的开发者/公司
type Tenant struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:100;not null;uniqueIndex" json:"name"` // 租户名称
	AppKey    string    `gorm:"size:64;not null;uniqueIndex" json:"app_key"`    // SDK 接入 Key（公开）
	AppSecret string    `gorm:"size:128;not null" json:"-"`                      // SDK 接入 Secret（保密）
	Status    string    `gorm:"size:20;default:'active'" json:"status"`          // active | disabled
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User 管理员用户，属于某个租户
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint      `gorm:"index;not null" json:"tenant_id"`         // 所属租户
	Username     string    `gorm:"size:64;not null;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"size:256;not null" json:"-"`              // bcrypt 哈希，禁止序列化输出
	Role         string    `gorm:"size:20;default:'admin'" json:"role"`     // superadmin | admin | viewer
	Status       string    `gorm:"size:20;default:'active'" json:"status"`  // active | disabled
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联
	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}
