package alerter

import (
	"context"

	"github.com/thingsboard-rxprobe/internal/model"
)

// Alerter 告警通道接口
type Alerter interface {
	// Type 返回告警通道类型
	Type() string

	// Send 发送告警
	Send(ctx context.Context, alert *model.Alert) error
}
