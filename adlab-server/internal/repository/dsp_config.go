package repository

import (
	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// DSPConfigRepository DSP 配置数据访问层
type DSPConfigRepository struct {
	db *gorm.DB
}

// NewDSPConfigRepository 创建 DSPConfigRepository
func NewDSPConfigRepository(db *gorm.DB) *DSPConfigRepository {
	return &DSPConfigRepository{db: db}
}

// FindBySourceID 根据 source_id 查询 DSP 配置
func (r *DSPConfigRepository) FindBySourceID(sourceID string) (*model.DSPConfig, error) {
	var config model.DSPConfig
	result := r.db.Where("source_id = ?", sourceID).First(&config)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeEntityNotFound, "DSP 配置不存在: "+sourceID)
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询 DSP 配置失败", result.Error)
	}
	return &config, nil
}

// Create 创建 DSP 配置
func (r *DSPConfigRepository) Create(config *model.DSPConfig) error {
	var count int64
	r.db.Model(&model.DSPConfig{}).Where("source_id = ?", config.SourceID).Count(&count)
	if count > 0 {
		return errors.New(errors.CodeEntityAlreadyExists, "DSP 配置已存在: "+config.SourceID)
	}

	if err := r.db.Create(config).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建 DSP 配置失败", err)
	}
	return nil
}

// Update 更新 DSP 配置
func (r *DSPConfigRepository) Update(config *model.DSPConfig) error {
	result := r.db.Save(config)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新 DSP 配置失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "DSP 配置不存在: "+config.SourceID)
	}
	return nil
}

// Delete 删除 DSP 配置
func (r *DSPConfigRepository) Delete(sourceID string) error {
	result := r.db.Where("source_id = ?", sourceID).Delete(&model.DSPConfig{})
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "删除 DSP 配置失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "DSP 配置不存在: "+sourceID)
	}
	return nil
}

// FindAll 查询所有 DSP 配置（支持分页）
func (r *DSPConfigRepository) FindAll(page, pageSize int) ([]model.DSPConfig, int64, error) {
	var configs []model.DSPConfig
	var total int64

	query := r.db.Model(&model.DSPConfig{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计 DSP 配置数量失败", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Find(&configs).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "查询 DSP 配置列表失败", err)
	}

	return configs, total, nil
}
