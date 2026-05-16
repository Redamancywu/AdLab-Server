package repository

import (
	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// MockAdRepository Mock 广告数据访问层
type MockAdRepository struct {
	db *gorm.DB
}

// NewMockAdRepository 创建 MockAdRepository
func NewMockAdRepository(db *gorm.DB) *MockAdRepository {
	return &MockAdRepository{db: db}
}

// FindByMockAdID 根据 mock_ad_id 查询
func (r *MockAdRepository) FindByMockAdID(id string) (*model.MockAd, error) {
	var ad model.MockAd
	result := r.db.Where("mock_ad_id = ?", id).First(&ad)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeEntityNotFound, "Mock 广告不存在: "+id)
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询 Mock 广告失败", result.Error)
	}
	return &ad, nil
}

// FindAll 分页查询（支持按 ad_type 过滤）
func (r *MockAdRepository) FindAll(page, pageSize int, adType string) ([]model.MockAd, int64, error) {
	var ads []model.MockAd
	var total int64

	query := r.db.Model(&model.MockAd{})
	if adType != "" {
		query = query.Where("ad_type = ?", adType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计 Mock 广告数量失败", err)
	}

	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}

	if err := query.Order("priority ASC, created_at DESC").Find(&ads).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "查询 Mock 广告列表失败", err)
	}

	return ads, total, nil
}

// FindActiveByAdType 查询指定类型的所有 active Mock 广告（按优先级排序）
func (r *MockAdRepository) FindActiveByAdType(adType string) ([]model.MockAd, error) {
	var ads []model.MockAd
	result := r.db.Where("status = ? AND ad_type = ?", "active", adType).
		Order("priority ASC").Find(&ads)
	if result.Error != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询 Mock 广告失败", result.Error)
	}
	return ads, nil
}

// FindRandomActive 随机查询一个 active 的 Mock 广告（可按类型过滤）
func (r *MockAdRepository) FindRandomActive(adType string) (*model.MockAd, error) {
	var ad model.MockAd
	query := r.db.Where("status = ?", "active")
	if adType != "" {
		query = query.Where("ad_type = ?", adType)
	}
	result := query.Order("RANDOM()").First(&ad)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeEntityNotFound, "没有可用的 Mock 广告")
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "随机查询 Mock 广告失败", result.Error)
	}
	return &ad, nil
}

// Create 创建 Mock 广告
func (r *MockAdRepository) Create(ad *model.MockAd) error {
	var count int64
	r.db.Model(&model.MockAd{}).Where("mock_ad_id = ?", ad.MockAdID).Count(&count)
	if count > 0 {
		return errors.New(errors.CodeEntityAlreadyExists, "Mock 广告已存在: "+ad.MockAdID)
	}
	if err := r.db.Create(ad).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建 Mock 广告失败", err)
	}
	return nil
}

// Update 更新 Mock 广告
func (r *MockAdRepository) Update(ad *model.MockAd) error {
	result := r.db.Save(ad)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新 Mock 广告失败", result.Error)
	}
	return nil
}

// Delete 删除 Mock 广告
func (r *MockAdRepository) Delete(id string) error {
	result := r.db.Where("mock_ad_id = ?", id).Delete(&model.MockAd{})
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "删除 Mock 广告失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "Mock 广告不存在: "+id)
	}
	return nil
}
