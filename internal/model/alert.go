package model

import (
	"time"
)

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

// AlertRecord 告警记录
type AlertRecord struct {
	ID         uint64      `json:"id" gorm:"primaryKey"`
	TargetID   uint64      `json:"target_id" gorm:"index;not null"`
	TargetName string      `json:"target_name" gorm:"size:128"`
	TargetType string      `json:"target_type" gorm:"size:32"`
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
