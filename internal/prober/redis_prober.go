package prober

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisProber Redis 探针
type RedisProber struct{}

// NewRedisProber 创建 Redis 探针
func NewRedisProber() *RedisProber {
	return &RedisProber{}
}

// Type 返回探针类型
func (p *RedisProber) Type() string {
	return "redis"
}

// ConfigSchema 返回配置表单 schema
func (p *RedisProber) ConfigSchema() map[string]FieldSchema {
	return map[string]FieldSchema{
		"mode": {
			Type:         "select",
			Label:        "部署模式",
			Required:     true,
			DefaultValue: "standalone",
			Options: []Option{
				{Value: "standalone", Label: "单机模式"},
				{Value: "sentinel", Label: "哨兵模式"},
				{Value: "cluster", Label: "集群模式"},
			},
		},
		"host": {
			Type:        "string",
			Label:       "主机地址",
			Required:    false,
			Placeholder: "localhost",
			ShowWhen:    map[string]any{"mode": "standalone"},
		},
		"port": {
			Type:         "number",
			Label:        "端口",
			Required:     false,
			DefaultValue: 6379,
			ShowWhen:     map[string]any{"mode": "standalone"},
		},
		"sentinel_addrs": {
			Type:        "string",
			Label:       "哨兵地址",
			Required:    false,
			Placeholder: "sentinel1:26379,sentinel2:26379,sentinel3:26379",
			Hint:        "多个地址用逗号分隔",
			ShowWhen:    map[string]any{"mode": "sentinel"},
		},
		"sentinel_master": {
			Type:        "string",
			Label:       "Master 名称",
			Required:    false,
			Placeholder: "mymaster",
			ShowWhen:    map[string]any{"mode": "sentinel"},
		},
		"cluster_addrs": {
			Type:        "string",
			Label:       "集群节点",
			Required:    false,
			Placeholder: "node1:6379,node2:6379,node3:6379",
			Hint:        "多个地址用逗号分隔",
			ShowWhen:    map[string]any{"mode": "cluster"},
		},
		"password": {
			Type:     "password",
			Label:    "密码",
			Required: false,
		},
		"database": {
			Type:         "number",
			Label:        "数据库编号",
			Required:     false,
			DefaultValue: 0,
		},
	}
}

// Probe 执行探测
func (p *RedisProber) Probe(ctx context.Context, target Target) (*ProbeResult, error) {
	start := time.Now()

	mode := getStringConfig(target.Config, "mode", "standalone")
	password := getStringConfig(target.Config, "password", "")
	database := getIntConfig(target.Config, "database", 0)

	var rdb redis.UniversalClient

	switch mode {
	case "standalone":
		host := getStringConfig(target.Config, "host", "localhost")
		port := getIntConfig(target.Config, "port", 6379)
		rdb = redis.NewClient(&redis.Options{
			Addr:        fmt.Sprintf("%s:%d", host, port),
			Password:    password,
			DB:          database,
			DialTimeout: target.Timeout,
		})
	case "sentinel":
		addrs := getStringSliceConfig(target.Config, "sentinel_addrs")
		masterName := getStringConfig(target.Config, "sentinel_master", "mymaster")
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    masterName,
			SentinelAddrs: addrs,
			Password:      password,
			DB:            database,
			DialTimeout:   target.Timeout,
		})
	case "cluster":
		addrs := getStringSliceConfig(target.Config, "cluster_addrs")
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:       addrs,
			Password:    password,
			DialTimeout: target.Timeout,
		})
	default:
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("不支持的部署模式: %s", mode),
			CheckedAt: time.Now(),
		}, nil
	}
	defer rdb.Close()

	// PING 测试
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("PING 失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	return &ProbeResult{
		Success:   true,
		Latency:   time.Since(start),
		Message:   fmt.Sprintf("Redis (%s模式) 服务可用", mode),
		CheckedAt: time.Now(),
	}, nil
}

// Validate 验证目标配置
func (p *RedisProber) Validate(target Target) error {
	mode := getStringConfig(target.Config, "mode", "standalone")

	switch mode {
	case "standalone":
		if _, ok := target.Config["host"]; !ok {
			return fmt.Errorf("单机模式缺少 host 配置")
		}
	case "sentinel":
		if _, ok := target.Config["sentinel_addrs"]; !ok {
			return fmt.Errorf("哨兵模式缺少 sentinel_addrs 配置")
		}
		if _, ok := target.Config["sentinel_master"]; !ok {
			return fmt.Errorf("哨兵模式缺少 sentinel_master 配置")
		}
	case "cluster":
		if _, ok := target.Config["cluster_addrs"]; !ok {
			return fmt.Errorf("集群模式缺少 cluster_addrs 配置")
		}
	default:
		return fmt.Errorf("不支持的部署模式: %s", mode)
	}
	return nil
}
