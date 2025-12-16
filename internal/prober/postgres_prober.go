package prober

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresProber PostgreSQL 探针
type PostgresProber struct{}

// NewPostgresProber 创建 PostgreSQL 探针
func NewPostgresProber() *PostgresProber {
	return &PostgresProber{}
}

// Type 返回探针类型
func (p *PostgresProber) Type() string {
	return "postgresql"
}

// ConfigSchema 返回配置表单 schema
func (p *PostgresProber) ConfigSchema() map[string]FieldSchema {
	return map[string]FieldSchema{
		"host": {
			Type:        "string",
			Label:       "主机地址",
			Required:    true,
			Placeholder: "localhost",
		},
		"port": {
			Type:         "number",
			Label:        "端口",
			Required:     true,
			DefaultValue: 5432,
		},
		"username": {
			Type:     "string",
			Label:    "用户名",
			Required: true,
		},
		"password": {
			Type:     "password",
			Label:    "密码",
			Required: true,
		},
		"database": {
			Type:         "string",
			Label:        "数据库名",
			Required:     true,
			DefaultValue: "thingsboard",
		},
		"ssl_mode": {
			Type:         "select",
			Label:        "SSL 模式",
			Required:     false,
			DefaultValue: "disable",
			Options: []Option{
				{Value: "disable", Label: "禁用"},
				{Value: "require", Label: "要求"},
				{Value: "verify-ca", Label: "验证CA"},
				{Value: "verify-full", Label: "完全验证"},
			},
		},
	}
}

// Probe 执行探测
func (p *PostgresProber) Probe(ctx context.Context, target Target) (*ProbeResult, error) {
	start := time.Now()

	// 获取配置
	host := getStringConfig(target.Config, "host", "localhost")
	port := getIntConfig(target.Config, "port", 5432)
	username := getStringConfig(target.Config, "username", "")
	password := getStringConfig(target.Config, "password", "")
	database := getStringConfig(target.Config, "database", "thingsboard")
	sslMode := getStringConfig(target.Config, "ssl_mode", "disable")

	// 构建连接字符串
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=5",
		host, port, username, password, database, sslMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("连接失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}
	defer db.Close()

	// 连接检查
	if err := db.PingContext(ctx); err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("Ping 失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	return &ProbeResult{
		Success:   true,
		Latency:   time.Since(start),
		Message:   "PostgreSQL 服务可用",
		CheckedAt: time.Now(),
	}, nil
}

// Validate 验证目标配置
func (p *PostgresProber) Validate(target Target) error {
	required := []string{"host", "username", "password", "database"}
	for _, field := range required {
		if _, ok := target.Config[field]; !ok {
			return fmt.Errorf("缺少必填字段: %s", field)
		}
	}
	return nil
}
