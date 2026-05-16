package database

import (
	"fmt"
	"log"

	"adlab-server/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 根据配置打开数据库连接。
func Open(cfg *config.Config) (*gorm.DB, error) {
	var gormLogger logger.Interface
	if cfg.Server.Mode == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	dialector, dsn, err := buildDialector(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	log.Printf("数据库连接成功: %s (%s)", cfg.Database.Type, dsn)
	return db, nil
}

func buildDialector(cfg *config.Config) (gorm.Dialector, string, error) {
	switch cfg.Database.Type {
	case "", "sqlite", "sqlite3":
		path := cfg.Database.Path
		if path == "" {
			path = "adlab.db"
		}
		return sqlite.Open(path), path, nil
	case "postgres", "postgresql":
		dsn := cfg.Database.DSN
		if dsn == "" {
			if cfg.Database.Host == "" || cfg.Database.User == "" || cfg.Database.DBName == "" {
				return nil, "", fmt.Errorf("postgres 配置不完整，至少需要 host/user/dbname 或完整 dsn")
			}
			dsn = fmt.Sprintf(
				"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
				cfg.Database.Host,
				cfg.Database.Port,
				cfg.Database.User,
				cfg.Database.Password,
				cfg.Database.DBName,
				cfg.Database.SSLMode,
				cfg.Database.Timezone,
			)
		}
		return postgres.Open(dsn), dsn, nil
	default:
		return nil, "", fmt.Errorf("不支持的数据库类型: %s", cfg.Database.Type)
	}
}
