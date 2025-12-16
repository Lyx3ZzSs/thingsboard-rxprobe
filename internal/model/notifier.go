package model

import (
	"time"
)

// NotifyChannelType 通知渠道类型
type NotifyChannelType string

const (
	NotifyChannelTypeWeCom NotifyChannelType = "wecom" // 企业微信
)

// NotifyChannel 通知渠道
type NotifyChannel struct {
	ID          uint64            `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"size:128;not null"`        // 渠道名称
	Type        NotifyChannelType `json:"type" gorm:"size:32;not null;index"`   // 渠道类型
	WebhookURL  string            `json:"webhook_url" gorm:"size:512;not null"` // Webhook URL
	MessageTpl  string            `json:"message_tpl" gorm:"type:text"`         // 消息模板
	MentionAll  bool              `json:"mention_all" gorm:"default:true"`      // 是否@所有人
	Enabled     bool              `json:"enabled" gorm:"default:true;index"`    // 是否启用
	Description string            `json:"description" gorm:"size:256"`          // 描述
	CreatedAt   time.Time         `json:"created_at" gorm:"autoCreateTime"`     // 创建时间
	UpdatedAt   time.Time         `json:"updated_at" gorm:"autoUpdateTime"`     // 更新时间
}

// TableName 表名
func (NotifyChannel) TableName() string {
	return "notify_channels"
}

// DefaultFiringMessageTemplate 默认告警触发消息模板
const DefaultFiringMessageTemplate = `🚨 告警通知

目标：{{.TargetName}}
原因：{{.Message}}
时间：{{.FiredAt}}`

// DefaultResolvedMessageTemplate 默认告警恢复消息模板
const DefaultResolvedMessageTemplate = `✅ 恢复通知

目标：{{.TargetName}}
恢复时间：{{.ResolvedAt}}
故障时长：{{.Duration}}`
