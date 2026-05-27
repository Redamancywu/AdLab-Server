package repository

import (
	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// AppNetworkConfigRepository 应用级网络配置数据访问层
type AppNetworkConfigRepository struct {
	db *gorm.DB
}

// NewAppNetworkConfigRepository 创建 AppNetworkConfigRepository
func NewAppNetworkConfigRepository(db *gorm.DB) *AppNetworkConfigRepository {
	return &AppNetworkConfigRepository{db: db}
}

// FindByAppAndPlatform 查询应用在指定平台下的激活网络配置
func (r *AppNetworkConfigRepository) FindByAppAndPlatform(appID, platform string) ([]model.AppNetworkConfig, error) {
	var rows []model.AppNetworkConfig
	query := r.db.Where("app_id = ? AND status = ?", appID, "active")
	if platform != "" {
		query = query.Where("platform = ? OR platform = ?", platform, "both")
	}
	if err := query.Order("network_type ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询应用网络配置失败", err)
	}
	return rows, nil
}

// FindByID 查询单条配置
func (r *AppNetworkConfigRepository) FindByID(id uint) (*model.AppNetworkConfig, error) {
	var row model.AppNetworkConfig
	if err := r.db.First(&row, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeEntityNotFound, "应用网络配置不存在")
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询应用网络配置失败", err)
	}
	return &row, nil
}

// Create 创建配置
func (r *AppNetworkConfigRepository) Create(cfg *model.AppNetworkConfig) error {
	if err := r.db.Create(cfg).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建应用网络配置失败", err)
	}
	return nil
}

// Update 更新配置
func (r *AppNetworkConfigRepository) Update(cfg *model.AppNetworkConfig) error {
	if err := r.db.Save(cfg).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新应用网络配置失败", err)
	}
	return nil
}

// Delete 删除配置
func (r *AppNetworkConfigRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.AppNetworkConfig{}, id).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "删除应用网络配置失败", err)
	}
	return nil
}
