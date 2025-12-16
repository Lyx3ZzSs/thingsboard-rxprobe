package prober

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

// CPUProber CPU 占用率探针
type CPUProber struct{}

// NewCPUProber 创建 CPU 探针
func NewCPUProber() *CPUProber {
	return &CPUProber{}
}

// Type 返回探针类型
func (p *CPUProber) Type() string {
	return "cpu"
}

// ConfigSchema 返回配置表单 schema
func (p *CPUProber) ConfigSchema() map[string]FieldSchema {
	return map[string]FieldSchema{
		"threshold": {
			Type:         "number",
			Label:        "告警阈值（%）",
			Required:     true,
			DefaultValue: 80.0,
			Placeholder:  "80",
			Hint:         "当 CPU 占用率超过此值时触发告警，范围: 0-100",
		},
		"sample_duration": {
			Type:         "number",
			Label:        "采样时长（秒）",
			Required:     false,
			DefaultValue: 3,
			Placeholder:  "3",
			Hint:         "CPU 使用率采样时长，建议 1-10 秒",
		},
	}
}

// Probe 执行探测
func (p *CPUProber) Probe(ctx context.Context, target Target) (*ProbeResult, error) {
	start := time.Now()

	// 获取配置
	threshold := getFloatConfig(target.Config, "threshold", 80.0)
	sampleDuration := getIntConfig(target.Config, "sample_duration", 3)

	// 验证阈值范围
	if threshold < 0 || threshold > 100 {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   "告警阈值必须在 0-100 之间",
			CheckedAt: time.Now(),
		}, nil
	}

	// 验证采样时长
	if sampleDuration < 1 {
		sampleDuration = 1
	}
	if sampleDuration > 30 {
		sampleDuration = 30
	}

	// 获取 CPU 使用率（平均值，采样指定时长）
	duration := time.Duration(sampleDuration) * time.Second
	percentages, err := cpu.PercentWithContext(ctx, duration, false)
	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("获取 CPU 占用率失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	if len(percentages) == 0 {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   "未能获取 CPU 占用率数据",
			CheckedAt: time.Now(),
		}, nil
	}

	cpuPercent := percentages[0]

	// 获取 CPU 核心数
	cpuCounts, _ := cpu.CountsWithContext(ctx, true)

	// 构建 metrics
	metrics := make(map[string]any)
	metrics["cpu_percent"] = fmt.Sprintf("%.2f", cpuPercent)
	metrics["cpu_cores"] = cpuCounts
	metrics["threshold"] = threshold
	metrics["sample_duration"] = sampleDuration

	// 获取每个核心的使用率（用于详细信息）
	perCpuPercent, err := cpu.PercentWithContext(ctx, 0, true)
	if err == nil && len(perCpuPercent) > 0 {
		metrics["per_cpu_percent"] = perCpuPercent
	}

	// 判断是否超过告警阈值
	if cpuPercent > threshold {
		message := fmt.Sprintf("CPU 占用率 %.2f%% 超过告警阈值 %.1f%%", cpuPercent, threshold)
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   message,
			Metrics:   metrics,
			CheckedAt: time.Now(),
		}, nil
	}

	// 正常状态
	message := fmt.Sprintf("CPU 占用率: %.2f%% (阈值: %.1f%%)", cpuPercent, threshold)

	return &ProbeResult{
		Success:   true,
		Latency:   time.Since(start),
		Message:   message,
		Metrics:   metrics,
		CheckedAt: time.Now(),
	}, nil
}

// Validate 验证目标配置
func (p *CPUProber) Validate(target Target) error {
	threshold := getFloatConfig(target.Config, "threshold", -1)
	if threshold < 0 || threshold > 100 {
		return fmt.Errorf("告警阈值必须在 0-100 之间")
	}

	sampleDuration := getIntConfig(target.Config, "sample_duration", 3)
	if sampleDuration < 1 || sampleDuration > 30 {
		return fmt.Errorf("采样时长必须在 1-30 秒之间")
	}

	return nil
}
