package service

import (
	"time"

	"adlab-server/internal/errors"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"
)

// 合法的追踪事件类型集合
var validEventTypes = map[string]bool{
	model.EventImpression:    true,
	model.EventClick:         true,
	model.EventStart:         true,
	model.EventFirstQuartile: true,
	model.EventMidpoint:      true,
	model.EventThirdQuartile: true,
	model.EventComplete:      true,
	model.EventMute:          true,
	model.EventUnmute:        true,
	model.EventPause:         true,
	model.EventResume:        true,
	model.EventSkip:          true,
	"proxy_win":              true, // Win Notice 代理接收事件
}

// TrackRequest 追踪事件请求
type TrackRequest struct {
	Event      string `json:"event" form:"event"`
	RequestID  string `json:"request_id" form:"request_id"`
	MaterialID string `json:"material_id" form:"material_id"`
	Timestamp  int64  `json:"ts" form:"ts"` // Unix 毫秒时间戳，可选，默认取当前时间
	ClientIP   string `json:"-"`            // 从 HTTP 请求中提取
	UserAgent  string `json:"-"`            // 从 HTTP 请求中提取
}

// TrackingService 追踪事件服务
type TrackingService struct {
	trackingRepo *repository.TrackingEventLogRepository
}

// NewTrackingService 创建 TrackingService
func NewTrackingService(trackingRepo *repository.TrackingEventLogRepository) *TrackingService {
	return &TrackingService{
		trackingRepo: trackingRepo,
	}
}

// Track 记录追踪事件
func (s *TrackingService) Track(req *TrackRequest) error {
	// 1. 校验事件类型
	if req.Event == "" {
		return errors.New(errors.CodeTrackingError, "event 不能为空")
	}
	if !validEventTypes[req.Event] {
		return errors.Newf(errors.CodeTrackingError, "不支持的事件类型: %s", req.Event)
	}

	// 2. 校验必填字段
	if req.RequestID == "" {
		return errors.New(errors.CodeTrackingError, "request_id 不能为空")
	}
	if req.MaterialID == "" {
		return errors.New(errors.CodeTrackingError, "material_id 不能为空")
	}

	// 3. 时间戳默认取当前时间（毫秒）
	ts := req.Timestamp
	if ts <= 0 {
		ts = time.Now().UnixMilli()
	}

	// 4. 持久化追踪事件
	log := &model.TrackingEventLog{
		RequestID:  req.RequestID,
		MaterialID: req.MaterialID,
		EventType:  req.Event,
		Timestamp:  ts,
		ClientIP:   req.ClientIP,
		UserAgent:  req.UserAgent,
	}
	return s.trackingRepo.Create(log)
}
