package repository

import (
	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

type PlacementSourceBinding struct {
	model.PlacementSource
	Source *model.AdSource `json:"source,omitempty"`
}

// PlacementRepository 广告位数据访问层
type PlacementRepository struct {
	db *gorm.DB
}

// NewPlacementRepository 创建 PlacementRepository
func NewPlacementRepository(db *gorm.DB) *PlacementRepository {
	return &PlacementRepository{db: db}
}

// FindByPlacementID 根据 placement_id 查询广告位
func (r *PlacementRepository) FindByPlacementID(placementID string) (*model.Placement, error) {
	var placement model.Placement
	result := r.db.Where("placement_id = ?", placementID).First(&placement)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodePlacementNotFound, "广告位不存在: "+placementID)
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询广告位失败", result.Error)
	}
	return &placement, nil
}

// FindAll 查询所有广告位（支持分页）
func (r *PlacementRepository) FindAll(page, pageSize int) ([]model.Placement, int64, error) {
	var placements []model.Placement
	var total int64

	query := r.db.Model(&model.Placement{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计广告位数量失败", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Find(&placements).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "查询广告位列表失败", err)
	}

	for i := range placements {
		var count int64
		if err := r.db.Model(&model.PlacementSource{}).Where("placement_id = ?", placements[i].PlacementID).Count(&count).Error; err != nil {
			return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计广告位绑定数量失败", err)
		}
		placements[i].BindingCount = int(count)
	}

	return placements, total, nil
}

// Create 创建广告位
func (r *PlacementRepository) Create(placement *model.Placement) error {
	var count int64
	r.db.Model(&model.Placement{}).Where("placement_id = ?", placement.PlacementID).Count(&count)
	if count > 0 {
		return errors.New(errors.CodeEntityAlreadyExists, "广告位已存在: "+placement.PlacementID)
	}

	if err := r.db.Create(placement).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建广告位失败", err)
	}
	return nil
}

// Update 更新广告位
func (r *PlacementRepository) Update(placement *model.Placement) error {
	result := r.db.Save(placement)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新广告位失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "广告位不存在: "+placement.PlacementID)
	}
	return nil
}

// Delete 删除广告位
func (r *PlacementRepository) Delete(placementID string) error {
	result := r.db.Where("placement_id = ?", placementID).Delete(&model.Placement{})
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "删除广告位失败", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeEntityNotFound, "广告位不存在: "+placementID)
	}
	return nil
}

// FindBindings 查询广告位绑定及其关联的广告源详情
func (r *PlacementRepository) FindBindings(placementID string) ([]PlacementSourceBinding, error) {
	var rows []model.PlacementSource
	if err := r.db.Where("placement_id = ?", placementID).Find(&rows).Error; err != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询广告位绑定失败", err)
	}

	bindings := make([]PlacementSourceBinding, 0, len(rows))
	for _, row := range rows {
		source := &model.AdSource{}
		if err := r.db.
			Preload("DSPConfig").
			Where("source_id = ?", row.SourceID).
			First(source).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return nil, errors.Wrap(errors.CodeDatabaseError, "查询绑定广告源失败", err)
		}

		bindings = append(bindings, PlacementSourceBinding{
			PlacementSource: row,
			Source:          source,
		})
	}

	return bindings, nil
}

// FindWithSources 查询广告位及其绑定的广告源详情
func (r *PlacementRepository) FindWithSources(placementID string) (*model.Placement, error) {
	placement, err := r.FindByPlacementID(placementID)
	if err != nil {
		return nil, err
	}

	bindings, err := r.FindBindings(placementID)
	if err != nil {
		return nil, err
	}

	placement.PlacementSources = make([]model.PlacementSource, 0, len(bindings))
	placement.Sources = make([]model.AdSource, 0, len(bindings))

	for _, binding := range bindings {
		placement.PlacementSources = append(placement.PlacementSources, binding.PlacementSource)
		if binding.Source != nil {
			placement.Sources = append(placement.Sources, *binding.Source)
		}
	}

	return placement, nil
}

// BindSource 绑定广告源到广告位
func (r *PlacementRepository) BindSource(placementID, sourceID string, adUnitID ...string) error {
	ps := model.PlacementSource{
		PlacementID: placementID,
		SourceID:    sourceID,
		Status:      "active",
	}
	tx := r.db.Where("placement_id = ? AND source_id = ?", placementID, sourceID).First(&ps)
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		return errors.Wrap(errors.CodeDatabaseError, "查询绑定广告源失败", tx.Error)
	}

	if len(adUnitID) > 0 {
		ps.AdUnitID = adUnitID[0]
	}
	ps.Status = "active"

	if tx.Error == gorm.ErrRecordNotFound {
		if err := r.db.Create(&ps).Error; err != nil {
			return errors.Wrap(errors.CodeDatabaseError, "绑定广告源失败", err)
		}
		return nil
	}

	if err := r.db.Save(&ps).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "绑定广告源失败", err)
	}
	return nil
}

// UnbindSource 解绑广告源
func (r *PlacementRepository) UnbindSource(placementID, sourceID string) error {
	result := r.db.Where("placement_id = ? AND source_id = ?", placementID, sourceID).Delete(&model.PlacementSource{})
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "解绑广告源失败", result.Error)
	}
	return nil
}
