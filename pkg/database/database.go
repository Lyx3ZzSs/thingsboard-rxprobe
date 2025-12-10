package database

import (
	"fmt"
	"time"

	"github.com/thingsboard-rxprobe/internal/model"
	"github.com/thingsboard-rxprobe/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB

// Config 数据库配置
type Config struct {
	Driver       string
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
}

// Init 初始化数据库
func Init(cfg Config) error {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
		)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.DBName)
	default:
		return fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}

	var err error
	db, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	logger.Info("数据库连接成功",
		zap.String("driver", cfg.Driver),
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
	)

	return nil
}

// autoMigrate 自动迁移数据库
func autoMigrate() error {
	return db.AutoMigrate(
		&model.ProbeTarget{},
		&model.ProbeResult{},
		&model.AlertRecord{},
	)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
