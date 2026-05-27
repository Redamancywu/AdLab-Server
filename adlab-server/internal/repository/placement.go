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

type BindSourceParams struct {
	PlacementID        string
	SourceID           string
	InstanceID         string
	InstanceName       string
	AdUnitID           string
	TimeoutMsOverride  int
	FloorPriceOverride float64
	LoadParamsJSON     string
	Status             string
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

// FindByAppID 查询指定 App 下的所有广告位（支持分页）
func (r *PlacementRepository) FindByAppID(appID string, page, pageSize int) ([]model.Placement, int64, error) {
	var placements []model.Placement
	var total int64

	query := r.db.Model(&model.Placement{}).Where("app_id = ?", appID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计应用广告位数量失败", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Find(&placements).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDatabaseError, "查询应用广告位列表失败", err)
	}

	for i := range placements {
		var count int64
		if err := r.db.Model(&model.PlacementSource{}).Where("placement_id = ?", placements[i].PlacementID).Count(&count).Error; err != nil {
			return nil, 0, errors.Wrap(errors.CodeDatabaseError, "统计应用广告位绑定数量失败", err)
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

// BindSourceDetailed 以实例级字段绑定广告源到广告位
func (r *PlacementRepository) BindSourceDetailed(params BindSourceParams) error {
	existing := model.PlacementSource{}
	tx := r.db.Where("placement_id = ? AND source_id = ?", params.PlacementID, params.SourceID).First(&existing)
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		return errors.Wrap(errors.CodeDatabaseError, "查询绑定广告源失败", tx.Error)
	}

	if tx.Error == gorm.ErrRecordNotFound {
		ps := model.PlacementSource{
			PlacementID:        params.PlacementID,
			SourceID:           params.SourceID,
			InstanceID:         params.InstanceID,
			InstanceName:       params.InstanceName,
			AdUnitID:           params.AdUnitID,
			TimeoutMsOverride:  params.TimeoutMsOverride,
			FloorPriceOverride: params.FloorPriceOverride,
			LoadParamsJSON:     params.LoadParamsJSON,
			Status:             params.Status,
		}
		if ps.InstanceID == "" {
			ps.InstanceID = params.PlacementID + "_" + params.SourceID
		}
		if ps.Status == "" {
			ps.Status = "active"
		}
		if err := r.db.Create(&ps).Error; err != nil {
			return errors.Wrap(errors.CodeDatabaseError, "绑定广告源失败", err)
		}
		return nil
	}

	if params.InstanceID != "" {
		existing.InstanceID = params.InstanceID
	}
	if existing.InstanceID == "" {
		existing.InstanceID = params.PlacementID + "_" + params.SourceID
	}
	existing.InstanceName = params.InstanceName
	existing.AdUnitID = params.AdUnitID
	existing.TimeoutMsOverride = params.TimeoutMsOverride
	existing.FloorPriceOverride = params.FloorPriceOverride
	existing.LoadParamsJSON = params.LoadParamsJSON
	if params.Status != "" {
		existing.Status = params.Status
	} else {
		existing.Status = "active"
	}

	if err := r.db.Save(&existing).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "绑定广告源失败", err)
	}
	return nil
}

// UpdateBindingByInstanceID 按实例 ID 更新绑定记录
func (r *PlacementRepository) UpdateBindingByInstanceID(instanceID string, params BindSourceParams) error {
	existing := model.PlacementSource{}
	if err := r.db.Where("instance_id = ?", instanceID).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.CodeEntityNotFound, "实例绑定不存在: "+instanceID)
		}
		return errors.Wrap(errors.CodeDatabaseError, "查询实例绑定失败", err)
	}

	if params.PlacementID != "" {
		existing.PlacementID = params.PlacementID
	}
	if params.SourceID != "" {
		existing.SourceID = params.SourceID
	}
	existing.InstanceName = params.InstanceName
	existing.AdUnitID = params.AdUnitID
	existing.TimeoutMsOverride = params.TimeoutMsOverride
	existing.FloorPriceOverride = params.FloorPriceOverride
	existing.LoadParamsJSON = params.LoadParamsJSON
	if params.Status != "" {
		existing.Status = params.Status
	}

	if err := r.db.Save(&existing).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新实例绑定失败", err)
	}
	return nil
}

// BindSource 绑定广告源到广告位
func (r *PlacementRepository) BindSource(placementID, sourceID string, adUnitID ...string) error {
	params := BindSourceParams{
		PlacementID: placementID,
		SourceID:    sourceID,
		Status:      "active",
	}
	if len(adUnitID) > 0 {
		params.AdUnitID = adUnitID[0]
	}
	return r.BindSourceDetailed(params)
}

// UnbindSource 解绑广告源
func (r *PlacementRepository) UnbindSource(placementID, sourceID string) error {
	result := r.db.Where("placement_id = ? AND source_id = ?", placementID, sourceID).Delete(&model.PlacementSource{})
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "解绑广告源失败", result.Error)
	}
	return nil
}
