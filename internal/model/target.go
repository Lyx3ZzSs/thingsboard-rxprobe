package model

import (
	"time"

	"gorm.io/datatypes"
)

// ProbeTarget 状态常量
const (
	TargetStatusHealthy   = "healthy"   // 健康
	TargetStatusUnhealthy = "unhealthy" // 不健康
	TargetStatusUnknown   = "unknown"   // 未知（未探测或等待首次探测）
	TargetStatusDisabled  = "disabled"  // 已禁用
)

// ProbeTarget 探测目标
type ProbeTarget struct {
	ID               uint64         `json:"id" gorm:"primaryKey"`
	Name             string         `json:"name" gorm:"size:128;not null"`
	Type             string         `json:"type" gorm:"size:32;not null;index"`
	Config           datatypes.JSON `json:"config" gorm:"type:jsonb"`
	TimeoutSeconds   int            `json:"timeout_seconds" gorm:"default:5"`
	IntervalSeconds  int            `json:"interval_seconds" gorm:"default:30"`
	Enabled          bool           `json:"enabled" gorm:"default:true;index"`
	Status           string         `json:"status" gorm:"size:16;default:'unknown'"` // healthy, unhealthy, unknown, disabled
	LastCheckAt      *time.Time     `json:"last_check_at"`
	LastLatencyMs    int64          `json:"last_latency_ms"`
	LastMessage      string         `json:"last_message" gorm:"size:512"`
	Group            string         `json:"group" gorm:"size:64;index"`           // 分组
	NotifyChannelIDs datatypes.JSON `json:"notify_channel_ids" gorm:"type:jsonb"` // 通知渠道ID列表
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// TableName 表名
func (ProbeTarget) TableName() string {
	return "probe_targets"
}

// ProbeResult 探测结果
type ProbeResult struct {
	ID        uint64         `json:"id" gorm:"primaryKey"`
	TargetID  uint64         `json:"target_id" gorm:"index;not null"`
	Success   bool           `json:"success"`
	LatencyMs int64          `json:"latency_ms"`
	Message   string         `json:"message" gorm:"size:512"`
	Metrics   datatypes.JSON `json:"metrics" gorm:"type:jsonb"`
	Warnings  datatypes.JSON `json:"warnings" gorm:"type:jsonb"`
	CheckedAt time.Time      `json:"checked_at" gorm:"index"`
}

// TableName 表名
func (ProbeResult) TableName() string {
	return "probe_results"
}

// CreateTargetRequest 创建目标请求
type CreateTargetRequest struct {
	Name             string         `json:"name" binding:"required"`
	Type             string         `json:"type" binding:"required"`
	Config           map[string]any `json:"config" binding:"required"`
	TimeoutSeconds   int            `json:"timeout_seconds"`
	IntervalSeconds  int            `json:"interval_seconds"`
	Enabled          bool           `json:"enabled"`
	Group            string         `json:"group"`              // 分组
	NotifyChannelIDs []uint64       `json:"notify_channel_ids"` // 通知渠道ID列表
}

// UpdateTargetRequest 更新目标请求
type UpdateTargetRequest struct {
	Name             string         `json:"name"`
	Config           map[string]any `json:"config"`
	TimeoutSeconds   int            `json:"timeout_seconds"`
	IntervalSeconds  int            `json:"interval_seconds"`
	Enabled          *bool          `json:"enabled"`
	Group            *string        `json:"group"`              // 分组
	NotifyChannelIDs *[]uint64      `json:"notify_channel_ids"` // 通知渠道ID列表
}

// TestTargetRequest 测试目标请求
type TestTargetRequest struct {
	Type           string         `json:"type" binding:"required"`
	Config         map[string]any `json:"config" binding:"required"`
	TimeoutSeconds int            `json:"timeout_seconds"`
}
