package prober

import (
	"context"
	"fmt"
	"time"

	"github.com/go-ping/ping"
)

// PingProber Ping 探针
type PingProber struct{}

// NewPingProber 创建 Ping 探针
func NewPingProber() *PingProber {
	return &PingProber{}
}

// Type 返回探针类型
func (p *PingProber) Type() string {
	return "ping"
}

// ConfigSchema 返回配置表单 schema
func (p *PingProber) ConfigSchema() map[string]FieldSchema {
	return map[string]FieldSchema{
		"host": {
			Type:        "string",
			Label:       "主机地址",
			Required:    true,
			Placeholder: "example.com 或 192.168.1.1",
			Hint:        "支持域名或 IP 地址",
		},
	}
}

// Probe 执行探测
func (p *PingProber) Probe(ctx context.Context, target Target) (*ProbeResult, error) {
	start := time.Now()

	host := getStringConfig(target.Config, "host", "")
	if host == "" {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   "主机地址不能为空",
			CheckedAt: time.Now(),
		}, nil
	}

	// 固定配置：ping 4 次，整个操作总超时 3 秒
	count := 4
	timeout := 3 * time.Second

	// 创建 ping 实例
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("创建 ping 失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	// 设置 ping 参数
	pinger.Count = count
	pinger.Timeout = timeout
	pinger.SetPrivileged(false) // 非特权模式（使用 UDP，跨平台兼容）

	// 使用 context 支持取消
	if ctx != nil {
		pinger.OnRecv = func(pkt *ping.Packet) {
			select {
			case <-ctx.Done():
				pinger.Stop()
			default:
			}
		}
	}

	// 执行 ping
	err = pinger.Run()
	latency := time.Since(start)

	// 检查 context 是否被取消
	if ctx != nil && ctx.Err() != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   latency,
			Message:   fmt.Sprintf("Ping 超时或被取消: %v", ctx.Err()),
			CheckedAt: time.Now(),
		}, nil
	}

	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   latency,
			Message:   fmt.Sprintf("Ping 失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	stats := pinger.Statistics()

	// 检查是否有成功的响应
	if stats.PacketsRecv == 0 {
		return &ProbeResult{
			Success:   false,
			Latency:   latency,
			Message:   fmt.Sprintf("Ping 无响应: 发送 %d 个包，接收 0 个", stats.PacketsSent),
			CheckedAt: time.Now(),
		}, nil
	}

	metrics := map[string]any{
		"packets_sent":     stats.PacketsSent,
		"packets_received": stats.PacketsRecv,
		"packet_loss":      stats.PacketLoss,
		"min_rtt_ms":       stats.MinRtt.Milliseconds(),
		"max_rtt_ms":       stats.MaxRtt.Milliseconds(),
		"avg_rtt_ms":       stats.AvgRtt.Milliseconds(),
		"std_dev_ms":       stats.StdDevRtt.Milliseconds(),
	}

	message := fmt.Sprintf("Ping 成功: %s (发送 %d/%d, 平均延迟 %.2fms)",
		host, stats.PacketsRecv, stats.PacketsSent, float64(stats.AvgRtt.Milliseconds()))

	return &ProbeResult{
		Success:   true,
		Latency:   latency,
		Message:   message,
		Metrics:   metrics,
		CheckedAt: time.Now(),
	}, nil
}

// Validate 验证目标配置
func (p *PingProber) Validate(target Target) error {
	if _, ok := target.Config["host"]; !ok {
		return fmt.Errorf("缺少必填字段: host")
	}
	return nil
}
