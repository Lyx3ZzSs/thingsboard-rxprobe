package repository

import (
	"context"
	"time"

	"github.com/thingsboard-rxprobe/internal/model"
	"gorm.io/gorm"
)

// ResultRepository 探测结果仓库
type ResultRepository struct {
	db *gorm.DB
}

// NewResultRepository 创建探测结果仓库
func NewResultRepository(db *gorm.DB) *ResultRepository {
	return &ResultRepository{db: db}
}

// Create 创建探测结果
func (r *ResultRepository) Create(ctx context.Context, result *model.ProbeResult) error {
	return r.db.WithContext(ctx).Create(result).Error
}

// ResultQuery 结果查询参数
type ResultQuery struct {
	TargetID  uint64
	Success   *bool
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	Size      int
}

// List 获取探测结果列表
func (r *ResultRepository) List(ctx context.Context, query ResultQuery) ([]*model.ProbeResult, int64, error) {
	var results []*model.ProbeResult
	var total int64

	db := r.db.WithContext(ctx).Model(&model.ProbeResult{})

	if query.TargetID > 0 {
		db = db.Where("target_id = ?", query.TargetID)
	}
	if query.Success != nil {
		db = db.Where("success = ?", *query.Success)
	}
	if query.StartTime != nil {
		db = db.Where("checked_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("checked_at <= ?", query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Page > 0 && query.Size > 0 {
		offset := (query.Page - 1) * query.Size
		db = db.Offset(offset).Limit(query.Size)
	}

	if err := db.Order("checked_at DESC").Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetLatest 获取目标最新的探测结果
func (r *ResultRepository) GetLatest(ctx context.Context, targetID uint64, limit int) ([]*model.ProbeResult, error) {
	var results []*model.ProbeResult
	err := r.db.WithContext(ctx).
		Where("target_id = ?", targetID).
		Order("checked_at DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

// DeleteOld 删除旧的探测结果
func (r *ResultRepository) DeleteOld(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := r.db.WithContext(ctx).
		Where("checked_at < ?", cutoff).
		Delete(&model.ProbeResult{})
	return result.RowsAffected, result.Error
}

// DeleteByTargetID 删除指定目标的所有探测结果
func (r *ResultRepository) DeleteByTargetID(ctx context.Context, targetID uint64) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("target_id = ?", targetID).
		Delete(&model.ProbeResult{})
	return result.RowsAffected, result.Error
}

// GetSuccessRate 获取成功率统计
func (r *ResultRepository) GetSuccessRate(ctx context.Context, targetID uint64, duration time.Duration) (float64, error) {
	startTime := time.Now().Add(-duration)

	var total, success int64

	err := r.db.WithContext(ctx).Model(&model.ProbeResult{}).
		Where("target_id = ? AND checked_at >= ?", targetID, startTime).
		Count(&total).Error
	if err != nil {
		return 0, err
	}

	if total == 0 {
		return 100, nil
	}

	err = r.db.WithContext(ctx).Model(&model.ProbeResult{}).
		Where("target_id = ? AND checked_at >= ? AND success = ?", targetID, startTime, true).
		Count(&success).Error
	if err != nil {
		return 0, err
	}

	return float64(success) / float64(total) * 100, nil
}

// GetAverageLatency 获取平均延迟
func (r *ResultRepository) GetAverageLatency(ctx context.Context, targetID uint64, duration time.Duration) (float64, error) {
	startTime := time.Now().Add(-duration)

	var avgLatency float64
	err := r.db.WithContext(ctx).Model(&model.ProbeResult{}).
		Where("target_id = ? AND checked_at >= ? AND success = ?", targetID, startTime, true).
		Select("COALESCE(AVG(latency_ms), 0)").
		Scan(&avgLatency).Error

	return avgLatency, err
}
