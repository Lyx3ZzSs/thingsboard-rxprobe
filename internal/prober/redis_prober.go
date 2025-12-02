package prober

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
		"memory_threshold_percent": {
			Type:         "number",
			Label:        "内存使用告警阈值(%)",
			Required:     false,
			DefaultValue: 80,
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

	// 1. PING 测试
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("PING 失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	metrics := make(map[string]any)
	var warnings []string

	// 2. 获取 INFO 信息
	info, err := rdb.Info(ctx, "server", "memory", "clients", "stats", "replication").Result()
	if err == nil {
		infoMap := parseRedisInfo(info)

		// 服务器信息
		metrics["redis_version"] = infoMap["redis_version"]
		metrics["uptime_in_days"] = infoMap["uptime_in_days"]
		metrics["redis_mode"] = infoMap["redis_mode"]

		// 内存信息
		if usedMemory, ok := infoMap["used_memory"]; ok {
			usedMem, _ := strconv.ParseInt(usedMemory, 10, 64)
			metrics["used_memory_bytes"] = usedMem
			metrics["used_memory_human"] = infoMap["used_memory_human"]
		}
		if maxMemory, ok := infoMap["maxmemory"]; ok {
			maxMem, _ := strconv.ParseInt(maxMemory, 10, 64)
			metrics["max_memory_bytes"] = maxMem
			if maxMem > 0 {
				usedMem, _ := strconv.ParseInt(infoMap["used_memory"], 10, 64)
				usagePercent := float64(usedMem) / float64(maxMem) * 100
				metrics["memory_usage_percent"] = usagePercent

				threshold := getFloatConfig(target.Config, "memory_threshold_percent", 80)
				if usagePercent > threshold {
					warnings = append(warnings, fmt.Sprintf("内存使用率 %.1f%% 超过阈值 %.0f%%", usagePercent, threshold))
				}
			}
		}

		// 客户端连接数
		if connectedClients, ok := infoMap["connected_clients"]; ok {
			clients, _ := strconv.Atoi(connectedClients)
			metrics["connected_clients"] = clients
		}

		// 复制状态
		metrics["role"] = infoMap["role"]
		if infoMap["role"] == "slave" {
			metrics["master_host"] = infoMap["master_host"]
			metrics["master_port"] = infoMap["master_port"]
			metrics["master_link_status"] = infoMap["master_link_status"]

			if infoMap["master_link_status"] != "up" {
				warnings = append(warnings, "主从复制链接断开")
			}
		}
	}

	// 3. 集群模式额外检查
	if mode == "cluster" {
		if clusterClient, ok := rdb.(*redis.ClusterClient); ok {
			clusterInfo, err := clusterClient.ClusterInfo(ctx).Result()
			if err == nil {
				clusterMap := parseRedisInfo(clusterInfo)
				metrics["cluster_state"] = clusterMap["cluster_state"]
				metrics["cluster_slots_ok"] = clusterMap["cluster_slots_ok"]
				metrics["cluster_known_nodes"] = clusterMap["cluster_known_nodes"]

				if clusterMap["cluster_state"] != "ok" {
					warnings = append(warnings, fmt.Sprintf("集群状态异常: %s", clusterMap["cluster_state"]))
				}
			}
		}
	}

	latency := time.Since(start)
	message := fmt.Sprintf("Redis (%s模式) 运行正常", mode)

	if len(warnings) > 0 {
		message = fmt.Sprintf("存在告警: %v", warnings)
	}

	return &ProbeResult{
		Success:   true,
		Latency:   latency,
		Message:   message,
		Metrics:   metrics,
		CheckedAt: time.Now(),
		Warnings:  warnings,
	}, nil
}

// parseRedisInfo 解析 Redis INFO 命令输出
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
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
