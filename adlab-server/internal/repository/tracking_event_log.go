package repository

import (
	"adlab-server/internal/errors"
	"adlab-server/internal/model"

	"gorm.io/gorm"
)

// TrackingEventLogRepository 追踪事件日志数据访问层
type TrackingEventLogRepository struct {
	db *gorm.DB
}

// NewTrackingEventLogRepository 创建 TrackingEventLogRepository
func NewTrackingEventLogRepository(db *gorm.DB) *TrackingEventLogRepository {
	return &TrackingEventLogRepository{db: db}
}

// Create 创建追踪事件日志
func (r *TrackingEventLogRepository) Create(log *model.TrackingEventLog) error {
	if err := r.db.Create(log).Error; err != nil {
		return errors.Wrap(errors.CodeDatabaseError, "创建追踪事件日志失败", err)
	}
	return nil
}

// FindByRequestID 根据 request_id 查询追踪事件日志（按时间戳升序）
func (r *TrackingEventLogRepository) FindByRequestID(requestID string) ([]model.TrackingEventLog, error) {
	var logs []model.TrackingEventLog
	result := r.db.Where("request_id = ?", requestID).Order("timestamp ASC").Find(&logs)
	if result.Error != nil {
		return nil, errors.Wrap(errors.CodeDatabaseError, "查询追踪事件日志失败", result.Error)
	}
	return logs, nil
}
