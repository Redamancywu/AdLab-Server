package repository

import (
	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// AppRepository 应用数据访问层
type AppRepository struct {
	db *gorm.DB
}

// NewAppRepository 创建 AppRepository
func NewAppRepository(db *gorm.DB) *AppRepository {
	return &AppRepository{db: db}
}

// FindByAppID 根据 app_id 查询应用
func (r *AppRepository) FindByAppID(appID string) (*model.App, error) {
	var app model.App
	result := r.db.Where("app_id = ?", appID).First(&app)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeEntityNotFound, "应用不存在: "+appID)
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询应用失败", result.Error)
	}
	return &app, nil
}

// FindAll 查询所有应用（支持分页）
func (r *AppRepository) FindAll(page, pageSize int) ([]model.App, int64, error) {
	var apps []model.App
	var total int64

	query := r.db.Model(&model.App{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计应用数量失败", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Order("created_at DESC").Find(&apps).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "查询应用列表失败", err)
	}

	return apps, total, nil
}

// FindWithPlacements 查询应用及其关联广告位
func (r *AppRepository) FindWithPlacements(appID string) (*model.App, error) {
	var app model.App
	result := r.db.Preload("Placements").Where("app_id = ?", appID).First(&app)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeEntityNotFound, "应用不存在: "+appID)
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询应用及广告位失败", result.Error)
	}
	return &app, nil
}

// Create 创建应用
func (r *AppRepository) Create(app *model.App) error {
	var count int64
	r.db.Model(&model.App{}).Where("app_id = ?", app.AppID).Count(&count)
	if count > 0 {
		return errors.New(errors.CodeEntityAlreadyExists, "应用已存在: "+app.AppID)
	}
	if err := r.db.Create(app).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建应用失败", err)
	}
	return nil
}

// Update 更新应用
func (r *AppRepository) Update(app *model.App) error {
	result := r.db.Save(app)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新应用失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "应用不存在: "+app.AppID)
	}
	return nil
}

// Delete 删除应用
func (r *AppRepository) Delete(appID string) error {
	result := r.db.Where("app_id = ?", appID).Delete(&model.App{})
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "删除应用失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "应用不存在: "+appID)
	}
	return nil
}
