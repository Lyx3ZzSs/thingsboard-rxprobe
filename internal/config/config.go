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
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Log       LogConfig       `mapstructure:"log"`
	Auth      AuthConfig      `mapstructure:"auth"`
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

// SchedulerConfig 调度配置
type SchedulerConfig struct {
	DefaultInterval     int `mapstructure:"default_interval"`      // 默认探测间隔（秒）
	DefaultTimeout      int `mapstructure:"default_timeout"`       // 默认超时时间（秒）
	ResultRetentionDays int `mapstructure:"result_retention_days"` // 探测结果保留天数
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // console, json
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"` // JWT密钥
	JWTExpiry string `mapstructure:"jwt_expiry"` // JWT过期时间，如：168h (7天)
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

	// Scheduler
	viper.SetDefault("scheduler.default_interval", 30)
	viper.SetDefault("scheduler.default_timeout", 5)
	viper.SetDefault("scheduler.result_retention_days", 30)

	// Log
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")

	// Auth
	viper.SetDefault("auth.jwt_secret", "your-secret-key-change-in-production")
	viper.SetDefault("auth.jwt_expiry", "168h") // 7天
}

// processEnvOverrides 处理环境变量覆盖
func processEnvOverrides() {
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
}

// Get 获取配置
func Get() *Config {
	return cfg
}
