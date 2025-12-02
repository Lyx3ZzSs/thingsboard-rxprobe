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

// Scheduler 调度器
type Scheduler struct {
	cron          *cron.Cron
	tasks         sync.Map // map[uint64]*ProbeTask
	proberFactory *prober.Factory
	resultChan    chan *ProbeResultEvent
	alertChan     chan *AlertEvent
	stopChan      chan struct{}
	mu            sync.RWMutex
	running       bool
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
func NewScheduler(factory *prober.Factory) *Scheduler {
	return &Scheduler{
		cron:          cron.New(cron.WithSeconds()),
		proberFactory: factory,
		resultChan:    make(chan *ProbeResultEvent, 1000),
		alertChan:     make(chan *AlertEvent, 100),
		stopChan:      make(chan struct{}),
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
	go s.processResults(ctx)
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

	// 如果目标状态是 unhealthy，初始化 FailCount 为 3
	// 这样当探测恢复正常时，能够正确触发恢复通知
	if target.Status == "unhealthy" {
		task.FailCount = 3
		logger.Info("目标状态为异常，初始化失败计数",
			zap.Uint64("target_id", target.ID),
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

	select {
	case s.resultChan <- &ProbeResultEvent{
		TargetID:  task.Target.ID,
		Target:    task.Target,
		Result:    result,
		Timestamp: time.Now(),
	}:
	default:
		logger.Warn("结果通道已满，丢弃结果", zap.Uint64("target_id", task.Target.ID))
	}
}

// processResults 处理结果
func (s *Scheduler) processResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case event := <-s.resultChan:
			s.handleResult(event)
		}
	}
}

// handleResult 处理单个结果
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
		// 连续失败3次触发告警
		if task.FailCount == 3 {
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
		if task.FailCount >= 3 {
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
