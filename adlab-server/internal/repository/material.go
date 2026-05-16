package repository

import (
	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// MaterialRepository 素材数据访问层
type MaterialRepository struct {
	db *gorm.DB
}

// NewMaterialRepository 创建 MaterialRepository
func NewMaterialRepository(db *gorm.DB) *MaterialRepository {
	return &MaterialRepository{db: db}
}

// FindByMaterialID 根据 material_id 查询素材
func (r *MaterialRepository) FindByMaterialID(materialID string) (*model.Material, error) {
	var material model.Material
	result := r.db.Where("material_id = ?", materialID).First(&material)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeMaterialNotFound, "素材不存在: "+materialID)
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询素材失败", result.Error)
	}
	return &material, nil
}

// FindAll 查询所有素材（支持分页）
func (r *MaterialRepository) FindAll(page, pageSize int) ([]model.Material, int64, error) {
	var materials []model.Material
	var total int64

	query := r.db.Model(&model.Material{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计素材数量失败", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Find(&materials).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "查询素材列表失败", err)
	}

	return materials, total, nil
}

// FindRandom 随机查询一个素材
func (r *MaterialRepository) FindRandom() (*model.Material, error) {
	var material model.Material
	result := r.db.Order("RANDOM()").First(&material)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeMaterialNotFound, "没有可用素材")
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "随机查询素材失败", result.Error)
	}
	return &material, nil
}

// Create 创建素材
func (r *MaterialRepository) Create(material *model.Material) error {
	var count int64
	r.db.Model(&model.Material{}).Where("material_id = ?", material.MaterialID).Count(&count)
	if count > 0 {
		return errors.New(errors.CodeEntityAlreadyExists, "素材已存在: "+material.MaterialID)
	}

	if err := r.db.Create(material).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建素材失败", err)
	}
	return nil
}

// Delete 删除素材
func (r *MaterialRepository) Delete(materialID string) error {
	result := r.db.Where("material_id = ?", materialID).Delete(&model.Material{})
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "删除素材失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "素材不存在: "+materialID)
	}
	return nil
}

// Update 更新素材
func (r *MaterialRepository) Update(material *model.Material) error {
	result := r.db.Save(material)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新素材失败", result.Error)
	}
	return nil
}
