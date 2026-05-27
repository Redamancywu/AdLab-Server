package repository

import (
	"errors"

	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户/租户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername 按用户名查找用户（包含关联租户）
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Preload("Tenant").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByID 按 ID 查找用户
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.Preload("Tenant").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		Update("last_login_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

// TenantRepository 租户数据访问层
type TenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository 创建 TenantRepository
func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// FindByAppKey 按 AppKey 查找租户（SDK 鉴权用）
func (r *TenantRepository) FindByAppKey(appKey string) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := r.db.Where("app_key = ? AND status = 'active'", appKey).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

// Create 创建租户
func (r *TenantRepository) Create(tenant *model.Tenant) error {
	return r.db.Create(tenant).Error
}

// List 列出所有租户
func (r *TenantRepository) List() ([]model.Tenant, error) {
	var tenants []model.Tenant
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}
