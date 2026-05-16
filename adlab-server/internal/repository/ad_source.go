package repository

import (
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// AdSourceRepository 广告源数据访问层
type AdSourceRepository struct {
	db *gorm.DB
}

// NewAdSourceRepository 创建 AdSourceRepository
func NewAdSourceRepository(db *gorm.DB) *AdSourceRepository {
	return &AdSourceRepository{db: db}
}

// FindBySourceID 根据 source_id 查询广告源
func (r *AdSourceRepository) FindBySourceID(sourceID string) (*model.AdSource, error) {
	var source model.AdSource
	result := r.db.Where("source_id = ?", sourceID).First(&source)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeEntityNotFound, "广告源不存在: "+sourceID)
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询广告源失败", result.Error)
	}
	return &source, nil
}

// FindByPlacementID 查询广告位关联的所有广告源
func (r *AdSourceRepository) FindByPlacementID(placementID string) ([]model.AdSource, error) {
	var sources []model.AdSource
	result := r.db.
		Joins("JOIN placement_sources ON placement_sources.source_id = ad_sources.source_id").
		Where("placement_sources.placement_id = ?", placementID).
		Find(&sources)
	if result.Error != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询广告位关联广告源失败", result.Error)
	}
	return sources, nil
}

// FindActiveS2SByPlacementID 查询广告位关联的所有 active S2S 广告源
func (r *AdSourceRepository) FindActiveS2SByPlacementID(placementID string) ([]model.AdSource, error) {
	var sources []model.AdSource
	result := r.db.
		Joins("JOIN placement_sources ON placement_sources.source_id = ad_sources.source_id").
		Where("placement_sources.placement_id = ? AND ad_sources.status = ? AND ad_sources.bid_mode = ?",
			placementID, "active", "s2s").
		Order("ad_sources.priority ASC").
		Find(&sources)
	if result.Error != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询 S2S 广告源失败", result.Error)
	}
	return sources, nil
}

// FindActiveWaterfallByPlacementID 查询广告位关联的所有 active Waterfall 广告源（按优先级升序）
// Waterfall 模式：不按 bid_mode 过滤，所有 active 广告源都参与瀑布流
// 排序规则：有历史 eCPM 的按 eCPM 降序（收益优先），无历史数据的按 priority 升序（配置优先）
func (r *AdSourceRepository) FindActiveWaterfallByPlacementID(placementID string) ([]model.AdSource, error) {
	var sources []model.AdSource
	result := r.db.
		Joins("JOIN placement_sources ON placement_sources.source_id = ad_sources.source_id").
		Where("placement_sources.placement_id = ? AND ad_sources.status = ?", placementID, "active").
		// 有 eCPM 样本的按 eCPM 降序，无样本的按 priority 升序
		// SQLite 兼容写法：ecpm_sample_count > 0 的排在前面，按 historical_ecpm DESC；其余按 priority ASC
		Order("CASE WHEN ad_sources.ecpm_sample_count > 0 THEN 0 ELSE 1 END ASC, " +
			"CASE WHEN ad_sources.ecpm_sample_count > 0 THEN -ad_sources.historical_ecpm ELSE ad_sources.priority END ASC").
		Find(&sources)
	if result.Error != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询 Waterfall 广告源失败", result.Error)
	}
	return sources, nil
}

// FindAll 查询所有广告源
func (r *AdSourceRepository) FindAll(page, pageSize int) ([]model.AdSource, int64, error) {
	var sources []model.AdSource
	var total int64

	query := r.db.Model(&model.AdSource{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计广告源数量失败", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Find(&sources).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "查询广告源列表失败", err)
	}

	return sources, total, nil
}

// Create 创建广告源
func (r *AdSourceRepository) Create(source *model.AdSource) error {
	var count int64
	r.db.Model(&model.AdSource{}).Where("source_id = ?", source.SourceID).Count(&count)
	if count > 0 {
		return errors.New(errors.CodeEntityAlreadyExists, "广告源已存在: "+source.SourceID)
	}

	if err := r.db.Create(source).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建广告源失败", err)
	}
	return nil
}

// Update 更新广告源
func (r *AdSourceRepository) Update(source *model.AdSource) error {
	result := r.db.Save(source)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新广告源失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "广告源不存在: "+source.SourceID)
	}
	return nil
}

// Delete 删除广告源
func (r *AdSourceRepository) Delete(sourceID string) error {
	result := r.db.Where("source_id = ?", sourceID).Delete(&model.AdSource{})
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "删除广告源失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "广告源不存在: "+sourceID)
	}
	return nil
}

// UpdateECPM 用指数移动平均更新广告源的历史 eCPM
// newECPM：本次实际 eCPM（USD CPM）
// 公式：historical_ecpm = alpha * newECPM + (1-alpha) * historical_ecpm
// alpha = ECPMDecayFactor（首次样本直接赋值）
func (r *AdSourceRepository) UpdateECPM(sourceID string, newECPM float64) error {
	now := time.Now()
	// 先查当前值
	var src model.AdSource
	if err := r.db.Where("source_id = ?", sourceID).First(&src).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "查询广告源失败", err)
	}

	var updatedECPM float64
	if src.ECPMSampleCount == 0 {
		// 首次样本：直接赋值
		updatedECPM = newECPM
	} else {
		// 指数移动平均
		updatedECPM = model.ECPMDecayFactor*newECPM + (1-model.ECPMDecayFactor)*src.HistoricalECPM
	}

	result := r.db.Model(&model.AdSource{}).
		Where("source_id = ?", sourceID).
		Updates(map[string]interface{}{
			"historical_ecpm":    updatedECPM,
			"ecpm_sample_count":  src.ECPMSampleCount + 1,
			"ecpm_updated_at":    now,
		})
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新 eCPM 失败", result.Error)
	}
	return nil
}
