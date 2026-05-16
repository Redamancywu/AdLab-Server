package model

import (
	"gorm.io/gorm"
)

// AutoMigrate 执行所有模型的自动迁移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&App{},
		&Placement{},
		&AdSource{},
		&PlacementSource{},
		&DSPConfig{},
		&Material{},
		&MockAd{},
		&BidRequestLog{},
		&BidDetailLog{},
		&TrackingEventLog{},
		&C2SReportLog{},
		&ConfigChangeLog{},
	)
}
