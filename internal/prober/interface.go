// Package prober 探针接口定义
package prober

import (
	"context"
	"time"
)

// ProbeResult 探测结果
type ProbeResult struct {
	Success   bool           `json:"success"`
	Latency   time.Duration  `json:"latency"`
	Message   string         `json:"message"`
	Metrics   map[string]any `json:"metrics,omitempty"`
	Warnings  []string       `json:"warnings,omitempty"`
	CheckedAt time.Time      `json:"checked_at"`
}

// Target 探测目标
type Target struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Endpoint string         `json:"endpoint"`
	Timeout  time.Duration  `json:"timeout"`
	Interval time.Duration  `json:"interval"`
	Config   map[string]any `json:"config"`
}

// Prober 探针接口
type Prober interface {
	// Type 返回探针类型
	Type() string

	// Probe 执行探测
	Probe(ctx context.Context, target Target) (*ProbeResult, error)

	// Validate 验证目标配置
	Validate(target Target) error
}

// FieldSchema 表单字段 schema
type FieldSchema struct {
	Type         string         `json:"type"`          // string, number, password, boolean, select
	Label        string         `json:"label"`         // 显示标签
	Required     bool           `json:"required"`      // 是否必填
	Placeholder  string         `json:"placeholder"`   // 占位符
	Hint         string         `json:"hint"`          // 提示信息
	DefaultValue any            `json:"default_value"` // 默认值
	Options      []Option       `json:"options"`       // select 选项
	ShowWhen     map[string]any `json:"show_when"`     // 条件显示
}

// Option select 选项
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// SchemaProvider 探针 Schema 提供者接口
type SchemaProvider interface {
	ConfigSchema() map[string]FieldSchema
}

// Factory 探针工厂
type Factory struct {
	probers map[string]Prober
}

// NewFactory 创建探针工厂
func NewFactory() *Factory {
	f := &Factory{
		probers: make(map[string]Prober),
	}
	// 注册所有探针
	f.Register(NewPostgresProber())
	f.Register(NewCassandraProber())
	f.Register(NewRedisProber())
	f.Register(NewKafkaProber())
	f.Register(NewHTTPProber())
	f.Register(NewTCPProber())
	f.Register(NewPingProber())
	f.Register(NewCPUProber())
	return f
}

// Register 注册探针
func (f *Factory) Register(p Prober) {
	f.probers[p.Type()] = p
}

// Get 获取探针
func (f *Factory) Get(probeType string) (Prober, bool) {
	p, ok := f.probers[probeType]
	return p, ok
}

// GetAll 获取所有探针类型
func (f *Factory) GetAll() map[string]Prober {
	return f.probers
}

// GetTypes 获取所有支持的探针类型
func (f *Factory) GetTypes() []string {
	types := make([]string, 0, len(f.probers))
	for t := range f.probers {
		types = append(types, t)
	}
	return types
}
