package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/thingsboard-rxprobe/internal/alerter"
	"github.com/thingsboard-rxprobe/internal/model"
	"github.com/thingsboard-rxprobe/internal/prober"
	"github.com/thingsboard-rxprobe/internal/repository"
	"github.com/thingsboard-rxprobe/internal/scheduler"
	"github.com/thingsboard-rxprobe/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertService 告警服务
type AlertService struct {
	alertRepo    *repository.AlertRepository
	targetRepo   *repository.TargetRepository
	resultRepo   *repository.ResultRepository
	notifierRepo *repository.NotifierRepository
	alerter      alerter.Alerter
	scheduler    *scheduler.Scheduler
	silenceMap   sync.Map // map[uint64]time.Time 静默记录
	stopChan     chan struct{}
}

// NewAlertService 创建告警服务
func NewAlertService(
	alertRepo *repository.AlertRepository,
	targetRepo *repository.TargetRepository,
	resultRepo *repository.ResultRepository,
	notifierRepo *repository.NotifierRepository,
	alerter alerter.Alerter,
	sch *scheduler.Scheduler,
) *AlertService {
	return &AlertService{
		alertRepo:    alertRepo,
		targetRepo:   targetRepo,
		resultRepo:   resultRepo,
		notifierRepo: notifierRepo,
		alerter:      alerter,
		scheduler:    sch,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动告警服务
func (s *AlertService) Start(ctx context.Context) {
	go s.processAlerts(ctx)
	go s.processResults(ctx)
	logger.Info("告警服务已启动")
}

// Stop 停止告警服务
func (s *AlertService) Stop() {
	close(s.stopChan)
	logger.Info("告警服务已停止")
}

// processAlerts 处理告警事件
func (s *AlertService) processAlerts(ctx context.Context) {
	alertChan := s.scheduler.GetAlertChan()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case event := <-alertChan:
			s.handleAlert(ctx, event)
		}
	}
}

// processResults 处理探测结果
func (s *AlertService) processResults(ctx context.Context) {
	resultChan := s.scheduler.GetResultChan()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case event := <-resultChan:
			s.saveResult(ctx, event)
		}
	}
}

// handleAlert 处理告警
func (s *AlertService) handleAlert(ctx context.Context, event *scheduler.AlertEvent) {
	// 静默期只影响“发送通知”，不影响告警记录/目标状态更新
	silenced := s.isSilenced(event.Target.ID)

	firedAt := event.Result.CheckedAt
	if firedAt.IsZero() {
		firedAt = time.Now()
	}

	var configMap map[string]any
	if err := json.Unmarshal(event.Target.Config, &configMap); err != nil {
		logger.Error("解析配置失败", zap.Error(err))
	}

	if event.Status == model.AlertStatusFiring {
		// “每次失败都告警”会频繁触发 firing 事件：这里做记录层去重
		// - 如果已有未恢复告警记录：更新其 message/latency（保持 fired_at 作为故障开始时间）
		// - 如果没有：创建新的未恢复告警记录
		record, err := s.alertRepo.GetLastFiringRecord(ctx, event.Target.ID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("查询未恢复告警记录失败", zap.Error(err))
			}
			record = nil
		}

		isNewRecord := false
		if record == nil {
			isNewRecord = true
			record = &model.AlertRecord{
				TargetID:   event.Target.ID,
				TargetName: event.Target.Name,
				TargetType: event.Target.Type,
				Status:     model.AlertStatusFiring,
				Message:    event.Result.Message,
				LatencyMs:  event.Result.Latency.Milliseconds(),
				FiredAt:    firedAt,
			}
			if err := s.alertRepo.CreateRecord(ctx, record); err != nil {
				logger.Error("创建告警记录失败", zap.Error(err))
			}
		} else {
			// 更新为最新失败原因（但不改变 FiredAt）
			record.TargetName = event.Target.Name
			record.TargetType = event.Target.Type
			record.Message = event.Result.Message
			record.LatencyMs = event.Result.Latency.Milliseconds()
		}

		// 发送告警通知
		alert := &model.Alert{
			ID:         record.ID,
			TargetID:   event.Target.ID,
			TargetName: event.Target.Name,
			TargetType: event.Target.Type,
			Status:     model.AlertStatusFiring,
			Message:    event.Result.Message,
			Latency:    event.Result.Latency,
			// 通知里的时间使用本次失败发生时间，避免每次重复告警都显示“首次失败时间”
			FiredAt: firedAt,
		}

		if silenced {
			logger.Debug("告警处于静默期，跳过发送通知",
				zap.Uint64("target_id", event.Target.ID),
			)
		} else {
			// 从数据库获取启用的通知渠道并发送
			if err := s.sendToAllChannels(ctx, alert); err != nil {
				logger.Error("发送告警失败", zap.Error(err))
			} else {
				record.Notified = true
				logger.Info("告警发送成功",
					zap.Uint64("target_id", event.Target.ID),
					zap.String("target_name", event.Target.Name),
				)
			}
		}

		// 持久化记录更新（避免“每次失败都告警”模式下无限新增 firing 记录）
		if isNewRecord {
			// 新记录已 Create；仅在通知成功且 Notified 变更时需要 Update
			if record.Notified {
				if err := s.alertRepo.UpdateRecord(ctx, record); err != nil {
					logger.Error("更新告警记录失败", zap.Error(err))
				}
			}
		} else {
			// 老记录 message/latency 可能变化；Notified 也可能变化
			if err := s.alertRepo.UpdateRecord(ctx, record); err != nil {
				logger.Error("更新告警记录失败", zap.Error(err))
			}
		}

		// 更新目标状态
		s.targetRepo.UpdateStatus(ctx, event.Target.ID, model.TargetStatusUnhealthy, event.Result.Latency.Milliseconds(), event.Result.Message)

	} else if event.Status == model.AlertStatusResolved {
		// 查找并恢复告警记录（仅更新数据库，不发送通知）
		record, err := s.alertRepo.GetLastFiringRecord(ctx, event.Target.ID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("查询未恢复告警记录失败", zap.Error(err))
			}
		} else if record != nil {
			s.alertRepo.ResolveRecord(ctx, record.ID)
			logger.Info("告警已恢复（不发送通知）",
				zap.Uint64("target_id", event.Target.ID),
				zap.String("target_name", event.Target.Name),
			)
		}

		// 更新目标状态
		s.targetRepo.UpdateStatus(ctx, event.Target.ID, model.TargetStatusHealthy, event.Result.Latency.Milliseconds(), event.Result.Message)
	}
}

// saveResult 保存探测结果
func (s *AlertService) saveResult(ctx context.Context, event *scheduler.ProbeResultEvent) {
	// 转换指标和警告为 JSON
	metricsJSON, _ := json.Marshal(event.Result.Metrics)
	warningsJSON, _ := json.Marshal(event.Result.Warnings)

	result := &model.ProbeResult{
		TargetID:  event.TargetID,
		Success:   event.Result.Success,
		LatencyMs: event.Result.Latency.Milliseconds(),
		Message:   event.Result.Message,
		Metrics:   metricsJSON,
		Warnings:  warningsJSON,
		CheckedAt: event.Result.CheckedAt,
	}

	if err := s.resultRepo.Create(ctx, result); err != nil {
		logger.Error("保存探测结果失败",
			zap.Uint64("target_id", event.TargetID),
			zap.Error(err),
		)
	}

	// 更新目标状态
	status := model.TargetStatusHealthy
	if !event.Result.Success {
		status = model.TargetStatusUnhealthy
	}
	s.targetRepo.UpdateStatus(ctx, event.TargetID, status, event.Result.Latency.Milliseconds(), event.Result.Message)
}

// isSilenced 检查是否在静默期
func (s *AlertService) isSilenced(targetID uint64) bool {
	if v, ok := s.silenceMap.Load(targetID); ok {
		silenceUntil := v.(time.Time)
		if time.Now().Before(silenceUntil) {
			return true
		}
		s.silenceMap.Delete(targetID)
	}
	return false
}

// SilenceAlert 静默告警
func (s *AlertService) SilenceAlert(targetID uint64, duration time.Duration) {
	silenceUntil := time.Now().Add(duration)
	s.silenceMap.Store(targetID, silenceUntil)
	logger.Info("告警已静默",
		zap.Uint64("target_id", targetID),
		zap.Duration("duration", duration),
	)
}

// TriggerAlert 手动触发告警（用于测试）
func (s *AlertService) TriggerAlert(ctx context.Context, target *model.ProbeTarget, result *prober.ProbeResult) {
	event := &scheduler.AlertEvent{
		Target:    target,
		Result:    result,
		Status:    model.AlertStatusFiring,
		FailCount: 3,
	}
	s.handleAlert(ctx, event)
}

// ListRecords 获取告警记录列表
func (s *AlertService) ListRecords(ctx context.Context, query repository.AlertRecordQuery) ([]*model.AlertRecord, int64, error) {
	return s.alertRepo.ListRecords(ctx, query)
}

// GetRecordByID 获取告警记录详情
func (s *AlertService) GetRecordByID(ctx context.Context, id uint64) (*model.AlertRecord, error) {
	return s.alertRepo.GetRecordByID(ctx, id)
}

// sendToAllChannels 发送告警到目标配置的通知渠道
func (s *AlertService) sendToAllChannels(ctx context.Context, alert *model.Alert) error {
	var alerterErr error
	var hasAnyChannel bool // 标记是否有任何可用的通知渠道
	var successCount int   // 成功发送的数量

	// 首先尝试使用配置文件中的告警器（向后兼容）
	if s.alerter != nil {
		hasAnyChannel = true
		if err := s.alerter.Send(ctx, alert); err != nil {
			logger.Error("配置文件告警器发送失败", zap.Error(err))
			alerterErr = err
		} else {
			successCount++
			logger.Debug("配置文件告警器发送成功")
		}
	}

	// 从数据库获取启用的通知渠道
	if s.notifierRepo == nil {
		// 如果没有通知渠道仓库，只依赖 alerter 的结果
		if hasAnyChannel {
			return alerterErr
		}
		return nil
	}

	// 获取目标信息，检查其配置的通知渠道
	target, err := s.targetRepo.GetByID(ctx, alert.TargetID)
	if err != nil {
		return fmt.Errorf("获取目标信息失败: %w", err)
	}

	// 解析目标配置的通知渠道ID列表
	var notifyChannelIDs []uint64
	if target.NotifyChannelIDs != nil && len(target.NotifyChannelIDs) > 0 {
		if err := json.Unmarshal(target.NotifyChannelIDs, &notifyChannelIDs); err != nil {
			logger.Error("解析通知渠道ID失败", zap.Error(err))
		}
	}

	// 如果目标没有配置通知渠道
	if len(notifyChannelIDs) == 0 {
		logger.Debug("目标未配置通知渠道",
			zap.Uint64("target_id", alert.TargetID),
		)
		// 如果之前配置文件告警器发送失败，返回该错误
		if hasAnyChannel {
			return alerterErr
		}
		// 如果没有任何通知方式，返回 nil（避免误报）
		return nil
	}

	// 获取所有启用的通知渠道
	allChannels, err := s.notifierRepo.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("获取通知渠道失败: %w", err)
	}

	// 筛选出目标配置的渠道
	var targetChannels []*model.NotifyChannel
	for _, channel := range allChannels {
		for _, id := range notifyChannelIDs {
			if channel.ID == id {
				targetChannels = append(targetChannels, channel)
				break
			}
		}
	}

	if len(targetChannels) == 0 {
		logger.Debug("未找到可用的通知渠道",
			zap.Uint64("target_id", alert.TargetID),
			zap.Any("configured_channels", notifyChannelIDs),
		)
		// 如果配置了通知渠道但都不可用，且 alerter 也失败了
		if alerterErr != nil {
			return alerterErr
		}
		// 如果配置了通知渠道但都不可用，返回错误
		return fmt.Errorf("目标配置了 %d 个通知渠道，但都不可用", len(notifyChannelIDs))
	}

	hasAnyChannel = true
	var lastErr error
	for _, channel := range targetChannels {
		if err := s.sendToChannel(ctx, channel, alert); err != nil {
			logger.Error("发送通知失败",
				zap.String("channel", channel.Name),
				zap.Error(err),
			)
			lastErr = err
		} else {
			successCount++
			logger.Debug("通知发送成功",
				zap.String("channel", channel.Name),
				zap.Uint64("target_id", alert.TargetID),
			)
		}
	}

	// 如果至少有一个渠道发送成功，返回成功
	if successCount > 0 {
		logger.Info("至少有一个通知渠道发送成功",
			zap.Int("success_count", successCount),
			zap.Uint64("target_id", alert.TargetID),
		)
		return nil
	}

	// 所有渠道都失败了，返回最后一个错误
	if lastErr != nil {
		return lastErr
	}
	if alerterErr != nil {
		return alerterErr
	}
	return fmt.Errorf("所有通知渠道发送失败")
}

// sendToChannel 发送告警到指定通知渠道
func (s *AlertService) sendToChannel(ctx context.Context, channel *model.NotifyChannel, alert *model.Alert) error {
	switch channel.Type {
	case model.NotifyChannelTypeWeCom:
		return s.sendToWeCom(ctx, channel, alert)
	default:
		return fmt.Errorf("不支持的通知渠道类型: %s", channel.Type)
	}
}

// sendToWeCom 发送到企业微信
func (s *AlertService) sendToWeCom(ctx context.Context, channel *model.NotifyChannel, alert *model.Alert) error {
	content := s.formatAlertMessage(channel, alert)

	msg := struct {
		MsgType string `json:"msgtype"`
		Text    struct {
			Content       string   `json:"content"`
			MentionedList []string `json:"mentioned_list,omitempty"`
		} `json:"text"`
	}{
		MsgType: "text",
	}
	msg.Text.Content = content
	// 根据配置决定是否@所有人
	if channel.MentionAll {
		msg.Text.MentionedList = []string{"@all"}
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, channel.WebhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API 返回错误: HTTP %d", resp.StatusCode)
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("企业微信 API 错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return nil
}

// formatAlertMessage 格式化告警消息
func (s *AlertService) formatAlertMessage(channel *model.NotifyChannel, alert *model.Alert) string {
	// 如果有自定义模板，使用自定义模板
	if channel.MessageTpl != "" {
		data := map[string]string{
			"TargetName": alert.TargetName,
			"TargetType": alert.TargetType,
			"Message":    alert.Message,
			"FiredAt":    alert.FiredAt.Format("2006-01-02 15:04:05"),
		}
		if alert.ResolvedAt != nil {
			data["ResolvedAt"] = alert.ResolvedAt.Format("2006-01-02 15:04:05")
			data["Duration"] = formatDuration(alert.ResolvedAt.Sub(alert.FiredAt))
		}

		t, err := template.New("message").Parse(channel.MessageTpl)
		if err == nil {
			var buf bytes.Buffer
			if err := t.Execute(&buf, data); err == nil {
				return buf.String()
			}
		}
	}

	// 使用默认模板
	if alert.Status == model.AlertStatusFiring {
		return fmt.Sprintf(`🚨 告警通知 [%s]

目标：%s
类型：%s
原因：%s
时间：%s`,
			channel.Name,
			alert.TargetName,
			alert.TargetType,
			alert.Message,
			alert.FiredAt.Format("2006-01-02 15:04:05"),
		)
	}

	content := fmt.Sprintf(`✅ 恢复通知 [%s]

目标：%s
类型：%s
时间：%s`,
		channel.Name,
		alert.TargetName,
		alert.TargetType,
		alert.FiredAt.Format("2006-01-02 15:04:05"),
	)

	if alert.ResolvedAt != nil {
		duration := alert.ResolvedAt.Sub(alert.FiredAt)
		content += fmt.Sprintf("\n恢复时间：%s", alert.ResolvedAt.Format("2006-01-02 15:04:05"))
		content += fmt.Sprintf("\n故障时长：%s", formatDuration(duration))
	}

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
