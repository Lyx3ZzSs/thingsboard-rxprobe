package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/thingsboard-rxprobe/internal/model"
	"github.com/thingsboard-rxprobe/internal/prober"
	"github.com/thingsboard-rxprobe/pkg/logger"
	"go.uber.org/zap"
)

// AlertChecker 告警检查接口，用于检查目标是否有未恢复的告警
type AlertChecker interface {
	HasUnresolvedAlert(ctx context.Context, targetID uint64) bool
}

// Scheduler 调度器
type Scheduler struct {
	cron           *cron.Cron
	tasks          sync.Map // map[uint64]*ProbeTask
	proberFactory  *prober.Factory
	alertChecker   AlertChecker
	resultChan     chan *ProbeResultEvent
	alertChan      chan *AlertEvent
	stopChan       chan struct{}
	mu             sync.RWMutex
	running        bool
	alertThreshold int // 告警阈值（连续失败次数）
}

// ProbeTask 探测任务
type ProbeTask struct {
	Target     *model.ProbeTarget
	EntryID    cron.EntryID
	FailCount  int
	LastResult *prober.ProbeResult
	mu         sync.Mutex
}

// ProbeResultEvent 探测结果事件
type ProbeResultEvent struct {
	TargetID  uint64
	Target    *model.ProbeTarget
	Result    *prober.ProbeResult
	Timestamp time.Time
}

// AlertEvent 告警事件
type AlertEvent struct {
	Target    *model.ProbeTarget
	Result    *prober.ProbeResult
	Status    model.AlertStatus
	FailCount int
}

// NewScheduler 创建调度器
func NewScheduler(factory *prober.Factory, alertThreshold int, alertChecker AlertChecker) *Scheduler {
	if alertThreshold <= 0 {
		alertThreshold = 3 // 默认值
	}
	return &Scheduler{
		cron:           cron.New(cron.WithSeconds()),
		proberFactory:  factory,
		alertChecker:   alertChecker,
		resultChan:     make(chan *ProbeResultEvent, 1000),
		alertChan:      make(chan *AlertEvent, 100),
		stopChan:       make(chan struct{}),
		alertThreshold: alertThreshold,
	}
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.cron.Start()
	// 注意：不再在这里调用 processResults
	// 告警逻辑改为在 executeProbe 中同步处理
	// resultChan 只由 alert_service 消费用于保存结果
	logger.Info("调度器已启动")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.cron.Stop()
	logger.Info("调度器已停止")
}

// AddTask 添加任务
func (s *Scheduler) AddTask(target *model.ProbeTarget) error {
	p, ok := s.proberFactory.Get(target.Type)
	if !ok {
		return fmt.Errorf("不支持的探针类型: %s", target.Type)
	}

	// 解析配置
	var config map[string]any
	if err := json.Unmarshal(target.Config, &config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	probeTarget := prober.Target{
		ID:       fmt.Sprintf("%d", target.ID),
		Name:     target.Name,
		Type:     target.Type,
		Config:   config,
		Timeout:  time.Duration(target.TimeoutSeconds) * time.Second,
		Interval: time.Duration(target.IntervalSeconds) * time.Second,
	}

	// 验证配置
	if err := p.Validate(probeTarget); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	task := &ProbeTask{Target: target}

	// 检查是否需要初始化 FailCount（用于正确触发恢复通知）
	// 条件：目标状态为 unhealthy，或者数据库中有未恢复的告警记录
	shouldInitFailCount := target.Status == model.TargetStatusUnhealthy
	if !shouldInitFailCount && s.alertChecker != nil {
		shouldInitFailCount = s.alertChecker.HasUnresolvedAlert(context.Background(), target.ID)
	}

	if shouldInitFailCount {
		task.FailCount = s.alertThreshold
		logger.Info("初始化失败计数（存在未恢复告警或状态异常）",
			zap.Uint64("target_id", target.ID),
			zap.String("status", target.Status),
			zap.Int("fail_count", task.FailCount),
		)
	}

	// 计算 cron 表达式
	spec := fmt.Sprintf("@every %ds", target.IntervalSeconds)

	entryID, err := s.cron.AddFunc(spec, func() {
		s.executeProbe(task, probeTarget)
	})
	if err != nil {
		return fmt.Errorf("添加定时任务失败: %w", err)
	}

	task.EntryID = entryID
	s.tasks.Store(target.ID, task)

	logger.Info("添加探测任务",
		zap.Uint64("target_id", target.ID),
		zap.String("name", target.Name),
		zap.String("type", target.Type),
		zap.Int("interval", target.IntervalSeconds),
	)

	return nil
}

// RemoveTask 移除任务
func (s *Scheduler) RemoveTask(targetID uint64) {
	if v, ok := s.tasks.Load(targetID); ok {
		task := v.(*ProbeTask)
		s.cron.Remove(task.EntryID)
		s.tasks.Delete(targetID)
		logger.Info("移除探测任务", zap.Uint64("target_id", targetID))
	}
}

// UpdateTask 更新任务
func (s *Scheduler) UpdateTask(target *model.ProbeTarget) error {
	s.RemoveTask(target.ID)
	if target.Enabled {
		return s.AddTask(target)
	}
	return nil
}

// GetResultChan 获取结果通道
func (s *Scheduler) GetResultChan() <-chan *ProbeResultEvent {
	return s.resultChan
}

// GetAlertChan 获取告警通道
func (s *Scheduler) GetAlertChan() <-chan *AlertEvent {
	return s.alertChan
}

// executeProbe 执行探测
func (s *Scheduler) executeProbe(task *ProbeTask, target prober.Target) {
	ctx, cancel := context.WithTimeout(context.Background(), target.Timeout)
	defer cancel()

	p, ok := s.proberFactory.Get(target.Type)
	if !ok {
		logger.Error("探针类型不存在", zap.String("type", target.Type))
		return
	}

	result, err := p.Probe(ctx, target)
	if err != nil {
		result = &prober.ProbeResult{
			Success:   false,
			Message:   err.Error(),
			CheckedAt: time.Now(),
		}
	}

	event := &ProbeResultEvent{
		TargetID:  task.Target.ID,
		Target:    task.Target,
		Result:    result,
		Timestamp: time.Now(),
	}

	// 先同步处理告警逻辑（更新失败计数、触发告警）
	s.handleResult(event)

	// 再发送到 resultChan 供 alert_service 保存结果
	select {
	case s.resultChan <- event:
	default:
		logger.Warn("结果通道已满，丢弃结果", zap.Uint64("target_id", task.Target.ID))
	}
}

// handleResult 处理单个结果（更新失败计数、触发告警）
func (s *Scheduler) handleResult(event *ProbeResultEvent) {
	v, ok := s.tasks.Load(event.TargetID)
	if !ok {
		return
	}

	task := v.(*ProbeTask)
	task.mu.Lock()
	defer task.mu.Unlock()

	task.LastResult = event.Result

	if !event.Result.Success {
		task.FailCount++
		// 连续失败达到阈值时触发告警
		if task.FailCount == s.alertThreshold {
			select {
			case s.alertChan <- &AlertEvent{
				Target:    task.Target,
				Result:    event.Result,
				Status:    model.AlertStatusFiring,
				FailCount: task.FailCount,
			}:
			default:
				logger.Warn("告警通道已满", zap.Uint64("target_id", event.TargetID))
			}
		}
	} else {
		if task.FailCount >= s.alertThreshold {
			// 恢复通知
			select {
			case s.alertChan <- &AlertEvent{
				Target:    task.Target,
				Result:    event.Result,
				Status:    model.AlertStatusResolved,
				FailCount: task.FailCount,
			}:
			default:
				logger.Warn("告警通道已满", zap.Uint64("target_id", event.TargetID))
			}
		}
		task.FailCount = 0
	}
}

// GetTaskStatus 获取任务状态
func (s *Scheduler) GetTaskStatus(targetID uint64) (*ProbeTask, bool) {
	if v, ok := s.tasks.Load(targetID); ok {
		return v.(*ProbeTask), true
	}
	return nil, false
}

// GetAllTasks 获取所有任务
func (s *Scheduler) GetAllTasks() []*ProbeTask {
	var tasks []*ProbeTask
	s.tasks.Range(func(key, value any) bool {
		tasks = append(tasks, value.(*ProbeTask))
		return true
	})
	return tasks
}
