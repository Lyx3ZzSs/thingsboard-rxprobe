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
		"max_connections_threshold": {
			Type:         "number",
			Label:        "最大连接数告警阈值",
			Required:     false,
			DefaultValue: 100,
			Hint:         "当前连接数超过此值时告警",
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

	// 1. 基础连接检查
	if err := db.PingContext(ctx); err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("Ping 失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	metrics := make(map[string]any)
	var warnings []string

	// 2. 获取连接数统计
	var activeConns, maxConns int
	err = db.QueryRowContext(ctx, `
		SELECT 
			(SELECT count(*) FROM pg_stat_activity WHERE state = 'active') as active,
			(SELECT setting::int FROM pg_settings WHERE name = 'max_connections') as max
	`).Scan(&activeConns, &maxConns)
	if err == nil {
		metrics["active_connections"] = activeConns
		metrics["max_connections"] = maxConns
		metrics["connection_usage_percent"] = float64(activeConns) / float64(maxConns) * 100

		threshold := getIntConfig(target.Config, "max_connections_threshold", 100)
		if activeConns > threshold {
			warnings = append(warnings, fmt.Sprintf("连接数 %d 超过阈值 %d", activeConns, threshold))
		}
	}

	// 3. 获取数据库大小
	var dbSize string
	err = db.QueryRowContext(ctx, `
		SELECT pg_size_pretty(pg_database_size(current_database()))
	`).Scan(&dbSize)
	if err == nil {
		metrics["database_size"] = dbSize
	}

	// 4. 检查复制状态（如果是从库）
	var replicationLag sql.NullFloat64
	err = db.QueryRowContext(ctx, `
		SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))
	`).Scan(&replicationLag)
	if err == nil && replicationLag.Valid {
		metrics["replication_lag_seconds"] = replicationLag.Float64
		if replicationLag.Float64 > 60 {
			warnings = append(warnings, fmt.Sprintf("复制延迟 %.1f 秒", replicationLag.Float64))
		}
	}

	// 5. 检查慢查询数量
	var slowQueryCount int
	err = db.QueryRowContext(ctx, `
		SELECT count(*) FROM pg_stat_activity 
		WHERE state = 'active' 
		AND query_start < now() - interval '30 seconds'
	`).Scan(&slowQueryCount)
	if err == nil {
		metrics["slow_queries"] = slowQueryCount
		if slowQueryCount > 5 {
			warnings = append(warnings, fmt.Sprintf("存在 %d 个慢查询", slowQueryCount))
		}
	}

	// 6. 获取 PostgreSQL 版本
	var version string
	err = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err == nil {
		metrics["version"] = version
	}

	latency := time.Since(start)
	message := "PostgreSQL 运行正常"

	if len(warnings) > 0 {
		message = fmt.Sprintf("存在告警: %v", warnings)
	}

	// 设置 endpoint
	target.Endpoint = fmt.Sprintf("%s:%d/%s", host, port, database)

	return &ProbeResult{
		Success:   true,
		Latency:   latency,
		Message:   message,
		Metrics:   metrics,
		CheckedAt: time.Now(),
		Warnings:  warnings,
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
