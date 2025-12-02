package model

import (
	"time"

	"gorm.io/datatypes"
)

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

// AlertRule 告警规则
type AlertRule struct {
	ID             uint64         `json:"id" gorm:"primaryKey"`
	Name           string         `json:"name" gorm:"size:128;not null"`
	Threshold      int            `json:"threshold" gorm:"default:3"`        // 连续失败次数阈值
	SilenceMinutes int            `json:"silence_minutes" gorm:"default:30"` // 静默时间（分钟）
	Enabled        bool           `json:"enabled" gorm:"default:true"`
	NotifyConfig   datatypes.JSON `json:"notify_config" gorm:"type:jsonb"` // 通知配置
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// TableName 表名
func (AlertRule) TableName() string {
	return "alert_rules"
}

// AlertRecord 告警记录
type AlertRecord struct {
	ID         uint64      `json:"id" gorm:"primaryKey"`
	TargetID   uint64      `json:"target_id" gorm:"index;not null"`
	TargetName string      `json:"target_name" gorm:"size:128"`
	TargetType string      `json:"target_type" gorm:"size:32"`
	RuleID     uint64      `json:"rule_id" gorm:"index"`
	Status     AlertStatus `json:"status" gorm:"size:16;not null;index"`
	Message    string      `json:"message" gorm:"size:1024"`
	LatencyMs  int64       `json:"latency_ms"`
	FiredAt    time.Time   `json:"fired_at" gorm:"index"`
	ResolvedAt *time.Time  `json:"resolved_at"`
	Notified   bool        `json:"notified" gorm:"default:false"`
}

// TableName 表名
func (AlertRecord) TableName() string {
	return "alert_records"
}

// Alert 告警通知内容
type Alert struct {
	ID         uint64
	TargetID   uint64
	TargetName string
	TargetType string
	Endpoint   string
	Status     AlertStatus
	Message    string
	Latency    time.Duration
	FiredAt    time.Time
	ResolvedAt *time.Time
	DetailURL  string
	Metrics    map[string]any
}

// CreateAlertRuleRequest 创建告警规则请求
type CreateAlertRuleRequest struct {
	Name           string         `json:"name" binding:"required"`
	Threshold      int            `json:"threshold"`
	SilenceMinutes int            `json:"silence_minutes"`
	Enabled        bool           `json:"enabled"`
	NotifyConfig   map[string]any `json:"notify_config"`
}

// UpdateAlertRuleRequest 更新告警规则请求
type UpdateAlertRuleRequest struct {
	Name           string         `json:"name"`
	Threshold      int            `json:"threshold"`
	SilenceMinutes int            `json:"silence_minutes"`
	Enabled        *bool          `json:"enabled"`
	NotifyConfig   map[string]any `json:"notify_config"`
}
