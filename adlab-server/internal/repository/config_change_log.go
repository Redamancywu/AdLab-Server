package repository

import (
	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// ConfigChangeLogRepository 配置变更日志数据访问层
type ConfigChangeLogRepository struct {
	db *gorm.DB
}

// NewConfigChangeLogRepository 创建 ConfigChangeLogRepository
func NewConfigChangeLogRepository(db *gorm.DB) *ConfigChangeLogRepository {
	return &ConfigChangeLogRepository{db: db}
}

// Create 创建配置变更日志
func (r *ConfigChangeLogRepository) Create(log *model.ConfigChangeLog) error {
	if err := r.db.Create(log).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建配置变更日志失败", err)
	}
	return nil
}

// FindAll 查询所有配置变更日志（支持分页 + 过滤）
func (r *ConfigChangeLogRepository) FindAll(page, pageSize int) ([]model.ConfigChangeLog, int64, error) {
	return r.FindWithFilter("", "", page, pageSize)
}

// ConfigChangeLogFilter 变更日志过滤条件
type ConfigChangeLogFilter struct {
	EntityType string // placement / ad_source / dsp_config / material / scenario 等
	Action     string // create / update / delete / bind / unbind / switch / import
}

// FindWithFilter 带过滤条件查询配置变更日志
func (r *ConfigChangeLogRepository) FindWithFilter(entityType, action string, page, pageSize int) ([]model.ConfigChangeLog, int64, error) {
	var logs []model.ConfigChangeLog
	var total int64

	query := r.db.Model(&model.ConfigChangeLog{})
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计配置变更日志数量失败", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "查询配置变更日志失败", err)
	}

	return logs, total, nil
}
