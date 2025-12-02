package alerter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/thingsboard-rxprobe/internal/model"
)

// WeComAlerter 企业微信告警器
type WeComAlerter struct {
	webhookURL string
	httpClient *http.Client
}

// WeComMessage 企业微信消息
type WeComMessage struct {
	MsgType  string   `json:"msgtype"`
	Markdown Markdown `json:"markdown"`
}

// Markdown 消息内容
type Markdown struct {
	Content string `json:"content"`
}

// WeComResponse 企业微信响应
type WeComResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// NewWeComAlerter 创建企业微信告警器
func NewWeComAlerter(webhookURL string) *WeComAlerter {
	return &WeComAlerter{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Type 返回告警通道类型
func (w *WeComAlerter) Type() string {
	return "wecom"
}

// Send 发送告警
func (w *WeComAlerter) Send(ctx context.Context, alert *model.Alert) error {
	if w.webhookURL == "" {
		return fmt.Errorf("企业微信 Webhook URL 未配置")
	}

	content := w.formatAlert(alert)

	msg := WeComMessage{
		MsgType: "markdown",
		Markdown: Markdown{
			Content: content,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.webhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("企业微信 API 错误: HTTP %d", resp.StatusCode)
	}

	var wecomResp WeComResponse
	if err := json.NewDecoder(resp.Body).Decode(&wecomResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if wecomResp.ErrCode != 0 {
		return fmt.Errorf("企业微信 API 错误: %d - %s", wecomResp.ErrCode, wecomResp.ErrMsg)
	}

	return nil
}

// formatAlert 格式化告警内容
func (w *WeComAlerter) formatAlert(alert *model.Alert) string {
	if alert.Status == model.AlertStatusFiring {
		// 告警触发
		return fmt.Sprintf(`🚨 <font color="warning">**Thingsboard 探针告警**</font>

**目标**：%s

**类型**：%s

**原因**：%s

**时间**：%s

<@all>`,
			alert.TargetName,
			getTypeLabel(alert.TargetType),
			alert.Message,
			alert.FiredAt.Format("2006-01-02 15:04:05"),
		)
	}

	// 告警恢复
	content := fmt.Sprintf(`✅ <font color="info">**Thingsboard 探针恢复**</font>

**目标**：%s

**类型**：%s

**时间**：%s`,
		alert.TargetName,
		getTypeLabel(alert.TargetType),
		alert.FiredAt.Format("2006-01-02 15:04:05"),
	)

	if alert.ResolvedAt != nil {
		// 计算故障时长
		duration := alert.ResolvedAt.Sub(alert.FiredAt)
		content += fmt.Sprintf("\n\n**恢复时间**：%s", alert.ResolvedAt.Format("2006-01-02 15:04:05"))
		content += fmt.Sprintf("\n\n**故障时长**：%s", formatDuration(duration))
	}

	// 恢复时也@所有人
	content += "\n\n<@all>"

	return content
}

// formatDuration 格式化时长
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d秒", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds > 0 {
			return fmt.Sprintf("%d分%d秒", minutes, seconds)
		}
		return fmt.Sprintf("%d分钟", minutes)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes > 0 {
		return fmt.Sprintf("%d小时%d分", hours, minutes)
	}
	return fmt.Sprintf("%d小时", hours)
}

// getTypeLabel 获取类型标签
func getTypeLabel(probeType string) string {
	labels := map[string]string{
		"postgresql": "PostgreSQL",
		"cassandra":  "Cassandra",
		"redis":      "Redis",
		"kafka":      "Kafka",
		"http":       "HTTP",
		"tcp":        "TCP",
	}
	if label, ok := labels[probeType]; ok {
		return label
	}
	return probeType
}
