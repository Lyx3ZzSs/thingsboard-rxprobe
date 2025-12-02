package prober

import (
	"context"
	"fmt"
	"net"
	"time"
)

// TCPProber TCP 探针
type TCPProber struct{}

// NewTCPProber 创建 TCP 探针
func NewTCPProber() *TCPProber {
	return &TCPProber{}
}

// Type 返回探针类型
func (p *TCPProber) Type() string {
	return "tcp"
}

// ConfigSchema 返回配置表单 schema
func (p *TCPProber) ConfigSchema() map[string]FieldSchema {
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
			DefaultValue: 80,
		},
	}
}

// Probe 执行探测
func (p *TCPProber) Probe(ctx context.Context, target Target) (*ProbeResult, error) {
	start := time.Now()

	host := getStringConfig(target.Config, "host", "localhost")
	port := getIntConfig(target.Config, "port", 80)
	addr := fmt.Sprintf("%s:%d", host, port)

	dialer := net.Dialer{
		Timeout: target.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	latency := time.Since(start)

	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   latency,
			Message:   fmt.Sprintf("连接失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}
	defer conn.Close()

	metrics := map[string]any{
		"remote_addr": conn.RemoteAddr().String(),
		"local_addr":  conn.LocalAddr().String(),
	}

	return &ProbeResult{
		Success:   true,
		Latency:   latency,
		Message:   fmt.Sprintf("TCP 连接成功: %s", addr),
		Metrics:   metrics,
		CheckedAt: time.Now(),
	}, nil
}

// Validate 验证目标配置
func (p *TCPProber) Validate(target Target) error {
	if _, ok := target.Config["host"]; !ok {
		return fmt.Errorf("缺少必填字段: host")
	}
	if _, ok := target.Config["port"]; !ok {
		return fmt.Errorf("缺少必填字段: port")
	}
	return nil
}
