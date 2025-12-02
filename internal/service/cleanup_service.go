package service

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/thingsboard-rxprobe/internal/repository"
	"github.com/thingsboard-rxprobe/pkg/logger"
	"go.uber.org/zap"
)

// CleanupService 数据清理服务
type CleanupService struct {
	resultRepo          *repository.ResultRepository
	alertRepo           *repository.AlertRepository
	cron                *cron.Cron
	resultRetentionDays int
	alertRetentionDays  int
}

// NewCleanupService 创建清理服务
func NewCleanupService(
	resultRepo *repository.ResultRepository,
	alertRepo *repository.AlertRepository,
	resultRetentionDays int,
) *CleanupService {
	// 告警记录默认保留 90 天
	alertRetentionDays := 90
	if resultRetentionDays > alertRetentionDays {
		alertRetentionDays = resultRetentionDays
	}

	return &CleanupService{
		resultRepo:          resultRepo,
		alertRepo:           alertRepo,
		cron:                cron.New(),
		resultRetentionDays: resultRetentionDays,
		alertRetentionDays:  alertRetentionDays,
	}
}

// Start 启动清理服务
func (s *CleanupService) Start() {
	// 每天凌晨 3:00 执行清理
	_, err := s.cron.AddFunc("0 3 * * *", func() {
		s.cleanup()
	})
	if err != nil {
		logger.Error("添加清理任务失败", zap.Error(err))
		return
	}

	s.cron.Start()
	logger.Info("清理服务已启动",
		zap.Int("result_retention_days", s.resultRetentionDays),
		zap.Int("alert_retention_days", s.alertRetentionDays),
		zap.String("schedule", "每天 03:00"),
	)

	// 启动时执行一次清理
	go func() {
		time.Sleep(5 * time.Second) // 等待服务完全启动
		s.cleanup()
	}()
}

// Stop 停止清理服务
func (s *CleanupService) Stop() {
	s.cron.Stop()
	logger.Info("清理服务已停止")
}

// cleanup 执行清理
func (s *CleanupService) cleanup() {
	ctx := context.Background()

	logger.Info("开始执行数据清理...")

	// 1. 清理过期的探测结果
	resultDeleted, err := s.resultRepo.DeleteOld(ctx, s.resultRetentionDays)
	if err != nil {
		logger.Error("清理探测结果失败", zap.Error(err))
	} else {
		logger.Info("清理探测结果完成",
			zap.Int64("deleted", resultDeleted),
			zap.Int("retention_days", s.resultRetentionDays),
		)
	}

	// 2. 清理过期的告警记录
	alertDeleted, err := s.alertRepo.DeleteOld(ctx, s.alertRetentionDays)
	if err != nil {
		logger.Error("清理告警记录失败", zap.Error(err))
	} else {
		logger.Info("清理告警记录完成",
			zap.Int64("deleted", alertDeleted),
			zap.Int("retention_days", s.alertRetentionDays),
		)
	}

	logger.Info("数据清理完成",
		zap.Int64("total_deleted", resultDeleted+alertDeleted),
	)
}

// CleanupNow 立即执行清理（手动触发）
func (s *CleanupService) CleanupNow() {
	s.cleanup()
}
