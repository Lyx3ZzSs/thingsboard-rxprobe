package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/thingsboard-rxprobe/internal/model"
	"github.com/thingsboard-rxprobe/internal/prober"
	"github.com/thingsboard-rxprobe/internal/repository"
	"github.com/thingsboard-rxprobe/internal/scheduler"
	"github.com/thingsboard-rxprobe/pkg/logger"
	"go.uber.org/zap"
)

// ProbeService 探测服务
type ProbeService struct {
	targetRepo *repository.TargetRepository
	resultRepo *repository.ResultRepository
	alertRepo  *repository.AlertRepository
	factory    *prober.Factory
	scheduler  *scheduler.Scheduler
}

// NewProbeService 创建探测服务
func NewProbeService(
	targetRepo *repository.TargetRepository,
	resultRepo *repository.ResultRepository,
	alertRepo *repository.AlertRepository,
	factory *prober.Factory,
	sch *scheduler.Scheduler,
) *ProbeService {
	return &ProbeService{
		targetRepo: targetRepo,
		resultRepo: resultRepo,
		alertRepo:  alertRepo,
		factory:    factory,
		scheduler:  sch,
	}
}

// CreateTarget 创建探测目标
func (s *ProbeService) CreateTarget(ctx context.Context, req *model.CreateTargetRequest) (*model.ProbeTarget, error) {
	// 验证探针类型
	p, ok := s.factory.Get(req.Type)
	if !ok {
		return nil, fmt.Errorf("不支持的探针类型: %s", req.Type)
	}

	// 验证配置
	target := prober.Target{
		Type:   req.Type,
		Config: req.Config,
	}
	if err := p.Validate(target); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	// 序列化配置
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %w", err)
	}

	// 设置默认值
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 5
	}
	if req.IntervalSeconds <= 0 {
		req.IntervalSeconds = 30
	}

	// 根据 enabled 状态设置初始 status
	initialStatus := model.TargetStatusUnknown
	initialMessage := "等待首次探测"
	if !req.Enabled {
		initialStatus = model.TargetStatusDisabled
		initialMessage = "监控已禁用"
	}

	probeTarget := &model.ProbeTarget{
		Name:            req.Name,
		Type:            req.Type,
		Config:          configJSON,
		TimeoutSeconds:  req.TimeoutSeconds,
		IntervalSeconds: req.IntervalSeconds,
		Enabled:         req.Enabled,
		Status:          initialStatus,
		LastMessage:     initialMessage,
	}

	if err := s.targetRepo.Create(ctx, probeTarget); err != nil {
		return nil, fmt.Errorf("创建目标失败: %w", err)
	}

	// 如果启用，添加到调度器
	if probeTarget.Enabled {
		if err := s.scheduler.AddTask(probeTarget); err != nil {
			logger.Error("添加调度任务失败", zap.Error(err))
		}
	}

	return probeTarget, nil
}

// UpdateTarget 更新探测目标
func (s *ProbeService) UpdateTarget(ctx context.Context, id uint64, req *model.UpdateTargetRequest) (*model.ProbeTarget, error) {
	target, err := s.targetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("目标不存在: %w", err)
	}

	if req.Name != "" {
		target.Name = req.Name
	}
	if req.Config != nil {
		// 验证配置
		p, ok := s.factory.Get(target.Type)
		if !ok {
			return nil, fmt.Errorf("不支持的探针类型: %s", target.Type)
		}
		probeTarget := prober.Target{
			Type:   target.Type,
			Config: req.Config,
		}
		if err := p.Validate(probeTarget); err != nil {
			return nil, fmt.Errorf("配置验证失败: %w", err)
		}

		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, fmt.Errorf("序列化配置失败: %w", err)
		}
		target.Config = configJSON
	}
	if req.TimeoutSeconds > 0 {
		target.TimeoutSeconds = req.TimeoutSeconds
	}
	if req.IntervalSeconds > 0 {
		target.IntervalSeconds = req.IntervalSeconds
	}
	if req.Enabled != nil {
		oldEnabled := target.Enabled
		target.Enabled = *req.Enabled

		// 当禁用时，将状态设置为 disabled
		if !*req.Enabled {
			target.Status = model.TargetStatusDisabled
			target.LastMessage = "监控已禁用"
		} else if oldEnabled != *req.Enabled {
			// 从禁用变为启用：重置所有状态信息，等待重新探测
			target.Status = model.TargetStatusUnknown
			target.LastMessage = "等待首次探测"
			target.LastCheckAt = nil
			target.LastLatencyMs = 0
		}
		// 如果原本就是启用状态，保持现有的健康状态不变
	}

	if err := s.targetRepo.Update(ctx, target); err != nil {
		return nil, fmt.Errorf("更新目标失败: %w", err)
	}

	// 更新调度器任务
	if err := s.scheduler.UpdateTask(target); err != nil {
		logger.Error("更新调度任务失败", zap.Error(err))
	}

	return target, nil
}

// DeleteTarget 删除探测目标
func (s *ProbeService) DeleteTarget(ctx context.Context, id uint64) error {
	// 从调度器移除
	s.scheduler.RemoveTask(id)

	// 删除关联的探测结果
	if deletedResults, err := s.resultRepo.DeleteByTargetID(ctx, id); err != nil {
		logger.Error("删除探测结果失败",
			zap.Uint64("target_id", id),
			zap.Error(err),
		)
		return fmt.Errorf("删除探测结果失败: %w", err)
	} else {
		logger.Info("已删除探测结果",
			zap.Uint64("target_id", id),
			zap.Int64("count", deletedResults),
		)
	}

	// 删除关联的告警记录
	if deletedAlerts, err := s.alertRepo.DeleteByTargetID(ctx, id); err != nil {
		logger.Error("删除告警记录失败",
			zap.Uint64("target_id", id),
			zap.Error(err),
		)
		return fmt.Errorf("删除告警记录失败: %w", err)
	} else {
		logger.Info("已删除告警记录",
			zap.Uint64("target_id", id),
			zap.Int64("count", deletedAlerts),
		)
	}

	// 删除探测目标
	if err := s.targetRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除目标失败: %w", err)
	}

	logger.Info("探测目标已删除", zap.Uint64("target_id", id))
	return nil
}

// GetTarget 获取探测目标
func (s *ProbeService) GetTarget(ctx context.Context, id uint64) (*model.ProbeTarget, error) {
	return s.targetRepo.GetByID(ctx, id)
}

// ListTargets 获取探测目标列表
func (s *ProbeService) ListTargets(ctx context.Context, query repository.ListQuery) ([]*model.ProbeTarget, int64, error) {
	return s.targetRepo.List(ctx, query)
}

// TestTarget 测试探测目标
func (s *ProbeService) TestTarget(ctx context.Context, req *model.TestTargetRequest) (*prober.ProbeResult, error) {
	p, ok := s.factory.Get(req.Type)
	if !ok {
		return nil, fmt.Errorf("不支持的探针类型: %s", req.Type)
	}

	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	target := prober.Target{
		Type:    req.Type,
		Config:  req.Config,
		Timeout: timeout,
	}

	if err := p.Validate(target); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return p.Probe(ctx, target)
}

// GetProbeSchema 获取探针配置 Schema
func (s *ProbeService) GetProbeSchema(probeType string) (map[string]prober.FieldSchema, error) {
	p, ok := s.factory.Get(probeType)
	if !ok {
		return nil, fmt.Errorf("不支持的探针类型: %s", probeType)
	}

	if provider, ok := p.(prober.SchemaProvider); ok {
		return provider.ConfigSchema(), nil
	}

	return make(map[string]prober.FieldSchema), nil
}

// GetProbeTypes 获取所有支持的探针类型
func (s *ProbeService) GetProbeTypes() []map[string]string {
	types := []map[string]string{
		{"value": "postgresql", "label": "PostgreSQL", "icon": "database"},
		{"value": "cassandra", "label": "Cassandra", "icon": "database"},
		{"value": "redis", "label": "Redis", "icon": "server"},
		{"value": "kafka", "label": "Kafka", "icon": "message"},
		{"value": "http", "label": "HTTP", "icon": "globe"},
		{"value": "tcp", "label": "TCP", "icon": "network"},
	}
	return types
}

// GetTargetResults 获取目标探测结果
func (s *ProbeService) GetTargetResults(ctx context.Context, targetID uint64, query repository.ResultQuery) ([]*model.ProbeResult, int64, error) {
	query.TargetID = targetID
	return s.resultRepo.List(ctx, query)
}

// GetTargetStats 获取目标统计信息
func (s *ProbeService) GetTargetStats(ctx context.Context, targetID uint64) (map[string]any, error) {
	successRate, err := s.resultRepo.GetSuccessRate(ctx, targetID, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	avgLatency, err := s.resultRepo.GetAverageLatency(ctx, targetID, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"success_rate_24h":   successRate,
		"avg_latency_ms_24h": avgLatency,
	}, nil
}

// LoadEnabledTargets 加载所有启用的目标到调度器
func (s *ProbeService) LoadEnabledTargets(ctx context.Context) error {
	targets, err := s.targetRepo.ListEnabled(ctx)
	if err != nil {
		return err
	}

	for _, target := range targets {
		if err := s.scheduler.AddTask(target); err != nil {
			logger.Error("加载调度任务失败",
				zap.Uint64("target_id", target.ID),
				zap.Error(err),
			)
		}
	}

	logger.Info("已加载探测任务", zap.Int("count", len(targets)))
	return nil
}
