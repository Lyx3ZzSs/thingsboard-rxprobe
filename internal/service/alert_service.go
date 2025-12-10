package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/thingsboard-rxprobe/internal/alerter"
	"github.com/thingsboard-rxprobe/internal/model"
	"github.com/thingsboard-rxprobe/internal/prober"
	"github.com/thingsboard-rxprobe/internal/repository"
	"github.com/thingsboard-rxprobe/internal/scheduler"
	"github.com/thingsboard-rxprobe/pkg/logger"
	"go.uber.org/zap"
)

// AlertService 告警服务
type AlertService struct {
	alertRepo  *repository.AlertRepository
	targetRepo *repository.TargetRepository
	resultRepo *repository.ResultRepository
	alerter    alerter.Alerter
	scheduler  *scheduler.Scheduler
	silenceMap sync.Map // map[uint64]time.Time 静默记录
	stopChan   chan struct{}
}

// NewAlertService 创建告警服务
func NewAlertService(
	alertRepo *repository.AlertRepository,
	targetRepo *repository.TargetRepository,
	resultRepo *repository.ResultRepository,
	alerter alerter.Alerter,
	sch *scheduler.Scheduler,
) *AlertService {
	return &AlertService{
		alertRepo:  alertRepo,
		targetRepo: targetRepo,
		resultRepo: resultRepo,
		alerter:    alerter,
		scheduler:  sch,
		stopChan:   make(chan struct{}),
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
	// 检查是否在静默期
	if s.isSilenced(event.Target.ID) {
		logger.Debug("告警处于静默期",
			zap.Uint64("target_id", event.Target.ID),
		)
		return
	}

	var configMap map[string]any
	if err := json.Unmarshal(event.Target.Config, &configMap); err != nil {
		logger.Error("解析配置失败", zap.Error(err))
	}

	if event.Status == model.AlertStatusFiring {
		// 创建告警记录
		record := &model.AlertRecord{
			TargetID:   event.Target.ID,
			TargetName: event.Target.Name,
			TargetType: event.Target.Type,
			Status:     model.AlertStatusFiring,
			Message:    event.Result.Message,
			LatencyMs:  event.Result.Latency.Milliseconds(),
			FiredAt:    time.Now(),
		}

		if err := s.alertRepo.CreateRecord(ctx, record); err != nil {
			logger.Error("创建告警记录失败", zap.Error(err))
		}

		// 发送告警通知
		if s.alerter != nil {
			alert := &model.Alert{
				ID:         record.ID,
				TargetID:   event.Target.ID,
				TargetName: event.Target.Name,
				TargetType: event.Target.Type,
				Status:     model.AlertStatusFiring,
				Message:    event.Result.Message,
				Latency:    event.Result.Latency,
				FiredAt:    time.Now(),
			}

			if err := s.alerter.Send(ctx, alert); err != nil {
				logger.Error("发送告警失败", zap.Error(err))
			} else {
				record.Notified = true
				s.alertRepo.UpdateRecord(ctx, record)
				logger.Info("告警发送成功",
					zap.Uint64("target_id", event.Target.ID),
					zap.String("target_name", event.Target.Name),
				)
			}
		}

		// 更新目标状态
		s.targetRepo.UpdateStatus(ctx, event.Target.ID, model.TargetStatusUnhealthy, event.Result.Latency.Milliseconds(), event.Result.Message)

	} else if event.Status == model.AlertStatusResolved {
		// 查找并恢复告警记录
		record, err := s.alertRepo.GetLastFiringRecord(ctx, event.Target.ID)
		if err == nil && record != nil {
			s.alertRepo.ResolveRecord(ctx, record.ID)

			// 发送恢复通知
			if s.alerter != nil {
				now := time.Now()
				alert := &model.Alert{
					ID:         record.ID,
					TargetID:   event.Target.ID,
					TargetName: event.Target.Name,
					TargetType: event.Target.Type,
					Status:     model.AlertStatusResolved,
					Message:    "服务已恢复正常",
					Latency:    event.Result.Latency,
					FiredAt:    record.FiredAt,
					ResolvedAt: &now,
				}

				if err := s.alerter.Send(ctx, alert); err != nil {
					logger.Error("发送恢复通知失败", zap.Error(err))
				} else {
					logger.Info("恢复通知发送成功",
						zap.Uint64("target_id", event.Target.ID),
						zap.String("target_name", event.Target.Name),
					)
				}
			}
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
