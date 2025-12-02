package prober

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

// CassandraProber Cassandra 探针
type CassandraProber struct{}

// NewCassandraProber 创建 Cassandra 探针
func NewCassandraProber() *CassandraProber {
	return &CassandraProber{}
}

// Type 返回探针类型
func (p *CassandraProber) Type() string {
	return "cassandra"
}

// ConfigSchema 返回配置表单 schema
func (p *CassandraProber) ConfigSchema() map[string]FieldSchema {
	return map[string]FieldSchema{
		"hosts": {
			Type:        "string",
			Label:       "集群节点",
			Required:    true,
			Placeholder: "192.168.1.1,192.168.1.2,192.168.1.3",
			Hint:        "多个节点用逗号分隔",
		},
		"port": {
			Type:         "number",
			Label:        "端口",
			Required:     true,
			DefaultValue: 9042,
		},
		"username": {
			Type:     "string",
			Label:    "用户名",
			Required: false,
		},
		"password": {
			Type:     "password",
			Label:    "密码",
			Required: false,
		},
		"keyspace": {
			Type:         "string",
			Label:        "Keyspace",
			Required:     true,
			DefaultValue: "thingsboard",
		},
		"consistency": {
			Type:         "select",
			Label:        "一致性级别",
			Required:     false,
			DefaultValue: "quorum",
			Options: []Option{
				{Value: "one", Label: "ONE"},
				{Value: "quorum", Label: "QUORUM"},
				{Value: "all", Label: "ALL"},
				{Value: "local_quorum", Label: "LOCAL_QUORUM"},
			},
		},
		"datacenter": {
			Type:        "string",
			Label:       "数据中心",
			Required:    false,
			Placeholder: "dc1",
		},
	}
}

// Probe 执行探测
func (p *CassandraProber) Probe(ctx context.Context, target Target) (*ProbeResult, error) {
	start := time.Now()

	// 解析主机列表
	hostsStr := getStringConfig(target.Config, "hosts", "localhost")
	hosts := strings.Split(hostsStr, ",")
	for i := range hosts {
		hosts[i] = strings.TrimSpace(hosts[i])
	}

	port := getIntConfig(target.Config, "port", 9042)
	keyspace := getStringConfig(target.Config, "keyspace", "thingsboard")

	cluster := gocql.NewCluster(hosts...)
	cluster.Port = port
	cluster.Keyspace = keyspace
	cluster.Timeout = target.Timeout
	cluster.ConnectTimeout = target.Timeout

	// 认证配置
	username := getStringConfig(target.Config, "username", "")
	if username != "" {
		password := getStringConfig(target.Config, "password", "")
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: username,
			Password: password,
		}
	}

	// 一致性级别
	consistency := getStringConfig(target.Config, "consistency", "quorum")
	switch consistency {
	case "one":
		cluster.Consistency = gocql.One
	case "quorum":
		cluster.Consistency = gocql.Quorum
	case "all":
		cluster.Consistency = gocql.All
	case "local_quorum":
		cluster.Consistency = gocql.LocalQuorum
	}

	// 数据中心感知
	dc := getStringConfig(target.Config, "datacenter", "")
	if dc != "" {
		cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy(dc)
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("连接失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}
	defer session.Close()

	metrics := make(map[string]any)
	var warnings []string

	// 1. 获取集群节点状态
	iter := session.Query("SELECT peer, data_center, rack, release_version FROM system.peers").Iter()
	var peer, peerDC, rack, version string
	nodeCount := 0
	nodes := []map[string]string{}

	for iter.Scan(&peer, &peerDC, &rack, &version) {
		nodeCount++
		nodes = append(nodes, map[string]string{
			"peer":       peer,
			"datacenter": peerDC,
			"rack":       rack,
			"version":    version,
		})
	}
	if err := iter.Close(); err != nil {
		warnings = append(warnings, fmt.Sprintf("获取节点信息失败: %v", err))
	}
	metrics["peer_nodes"] = nodes
	metrics["peer_node_count"] = nodeCount

	// 2. 检查本节点信息
	var localDC, localRack, localVersion string
	err = session.Query("SELECT data_center, rack, release_version FROM system.local").
		Scan(&localDC, &localRack, &localVersion)
	if err == nil {
		metrics["local_datacenter"] = localDC
		metrics["local_rack"] = localRack
		metrics["local_version"] = localVersion
	}

	// 3. 检查 keyspace 复制因子
	var replication map[string]string
	err = session.Query(`
		SELECT replication FROM system_schema.keyspaces WHERE keyspace_name = ?
	`, keyspace).Scan(&replication)
	if err == nil {
		metrics["keyspace_replication"] = replication
	}

	// 4. 简单读测试
	testStart := time.Now()
	var testResult int
	err = session.Query("SELECT COUNT(*) FROM system.local").Scan(&testResult)
	readLatency := time.Since(testStart)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("读测试失败: %v", err))
	} else {
		metrics["read_latency_ms"] = readLatency.Milliseconds()
	}

	latency := time.Since(start)
	message := fmt.Sprintf("Cassandra 集群正常，共 %d 个 peer 节点", nodeCount)

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

// Validate 验证目标配置
func (p *CassandraProber) Validate(target Target) error {
	if _, ok := target.Config["hosts"]; !ok {
		return fmt.Errorf("缺少必填字段: hosts")
	}
	if _, ok := target.Config["keyspace"]; !ok {
		return fmt.Errorf("缺少必填字段: keyspace")
	}
	return nil
}
