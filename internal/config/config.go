package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Alerter   AlerterConfig   `mapstructure:"alerter"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Log       LogConfig       `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"` // postgres, sqlite
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// AlerterConfig 告警配置
type AlerterConfig struct {
	WeCom WeComConfig `mapstructure:"wecom"`
}

// WeComConfig 企业微信配置
type WeComConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	WebhookURL string `mapstructure:"webhook_url"`
}

// SchedulerConfig 调度配置
type SchedulerConfig struct {
	DefaultInterval     int `mapstructure:"default_interval"`      // 默认探测间隔（秒）
	DefaultTimeout      int `mapstructure:"default_timeout"`       // 默认超时时间（秒）
	AlertThreshold      int `mapstructure:"alert_threshold"`       // 告警阈值（连续失败次数）
	ResultRetentionDays int `mapstructure:"result_retention_days"` // 探测结果保留天数
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // console, json
}

var cfg *Config

// Load 加载配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 支持环境变量
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 处理环境变量覆盖
	processEnvOverrides()

	return cfg, nil
}

// setDefaults 设置默认值
func setDefaults() {
	// Server
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8088)
	viper.SetDefault("server.mode", "release")

	// Database
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "rxprobe")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "rxprobe.db")
	viper.SetDefault("database.max_open_conns", 20)
	viper.SetDefault("database.max_idle_conns", 10)

	// JWT
	viper.SetDefault("jwt.secret", "rxprobe-secret-key-change-me")
	viper.SetDefault("jwt.expire_hours", 24)

	// Alerter
	viper.SetDefault("alerter.wecom.enabled", false)
	viper.SetDefault("alerter.wecom.webhook_url", "")

	// Scheduler
	viper.SetDefault("scheduler.default_interval", 30)
	viper.SetDefault("scheduler.default_timeout", 5)
	viper.SetDefault("scheduler.alert_threshold", 3)
	viper.SetDefault("scheduler.result_retention_days", 30)

	// Log
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")
}

// processEnvOverrides 处理环境变量覆盖
func processEnvOverrides() {
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("WECOM_WEBHOOK_URL"); v != "" {
		cfg.Alerter.WeCom.WebhookURL = v
		cfg.Alerter.WeCom.Enabled = true
	}
}

// Get 获取配置
func Get() *Config {
	return cfg
}
