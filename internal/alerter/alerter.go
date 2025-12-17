package alerter

import (
	"context"

	"github.com/thingsboard-rxprobe/internal/model"
)

// Alerter 为历史兼容保留的告警发送接口。
// 当前项目主要通过 notifierRepo（数据库通知渠道）发送告警；
// 如果需要，也可以在启动时注入一个实现来复用旧的配置文件告警器逻辑。
type Alerter interface {
	Send(ctx context.Context, alert *model.Alert) error
}
