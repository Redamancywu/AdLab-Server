package repository

import (
	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// C2SReportLogRepository C2S 上报日志数据访问层
type C2SReportLogRepository struct {
	db *gorm.DB
}

// NewC2SReportLogRepository 创建 C2SReportLogRepository
func NewC2SReportLogRepository(db *gorm.DB) *C2SReportLogRepository {
	return &C2SReportLogRepository{db: db}
}

// Create 创建 C2S 上报日志
func (r *C2SReportLogRepository) Create(log *model.C2SReportLog) error {
	if err := r.db.Create(log).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建 C2S 上报日志失败", err)
	}
	return nil
}

// FindByRequestID 根据 request_id 查询 C2S 上报日志
func (r *C2SReportLogRepository) FindByRequestID(requestID string) (*model.C2SReportLog, error) {
	var log model.C2SReportLog
	result := r.db.Where("request_id = ?", requestID).First(&log)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeRequestNotFound, "C2S 上报记录不存在: "+requestID)
		}
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询 C2S 上报日志失败", result.Error)
	}
	return &log, nil
}

// ExistsByRequestID 检查 request_id 是否已存在
func (r *C2SReportLogRepository) ExistsByRequestID(requestID string) (bool, error) {
	var count int64
	result := r.db.Model(&model.C2SReportLog{}).Where("request_id = ?", requestID).Count(&count)
	if result.Error != nil {
		return false, errors.Wrap(errors.CodeDatabaseError, "检查 C2S 上报记录失败", result.Error)
	}
	return count > 0, nil
}

// UpdateDisplayed 更新指定 request_id 的 displayed 状态
func (r *C2SReportLogRepository) UpdateDisplayed(requestID string, displayed bool) error {
	result := r.db.Model(&model.C2SReportLog{}).
		Where("request_id = ?", requestID).
		Update("displayed", displayed)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDatabaseError, "更新展示状态失败", result.Error)
	}
	return nil
}
