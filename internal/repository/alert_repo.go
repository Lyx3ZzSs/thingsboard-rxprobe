package repository

import (
	"context"
	"time"

	"github.com/thingsboard-rxprobe/internal/model"
	"gorm.io/gorm"
)

// AlertRepository 告警仓库
type AlertRepository struct {
	db *gorm.DB
}

// NewAlertRepository 创建告警仓库
func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// CreateRecord 创建告警记录
func (r *AlertRepository) CreateRecord(ctx context.Context, record *model.AlertRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// UpdateRecord 更新告警记录
func (r *AlertRepository) UpdateRecord(ctx context.Context, record *model.AlertRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// AlertRecordQuery 告警记录查询参数
type AlertRecordQuery struct {
	TargetID  uint64
	Status    string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	Size      int
}

// ListRecords 获取告警记录列表
func (r *AlertRepository) ListRecords(ctx context.Context, query AlertRecordQuery) ([]*model.AlertRecord, int64, error) {
	var records []*model.AlertRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&model.AlertRecord{})

	if query.TargetID > 0 {
		db = db.Where("target_id = ?", query.TargetID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("fired_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("fired_at <= ?", query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Page > 0 && query.Size > 0 {
		offset := (query.Page - 1) * query.Size
		db = db.Offset(offset).Limit(query.Size)
	}

	if err := db.Order("fired_at DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetRecordByID 根据 ID 获取告警记录
func (r *AlertRepository) GetRecordByID(ctx context.Context, id uint64) (*model.AlertRecord, error) {
	var record model.AlertRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetLastFiringRecord 获取目标最后一条未恢复的告警记录
func (r *AlertRepository) GetLastFiringRecord(ctx context.Context, targetID uint64) (*model.AlertRecord, error) {
	var record model.AlertRecord
	err := r.db.WithContext(ctx).
		Where("target_id = ? AND status = ?", targetID, model.AlertStatusFiring).
		Order("fired_at DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ResolveRecord 恢复告警记录
func (r *AlertRepository) ResolveRecord(ctx context.Context, id uint64) error {
	return r.ResolveRecordAt(ctx, id, time.Now())
}

// ResolveRecordAt 恢复告警记录，并设置恢复时间
func (r *AlertRepository) ResolveRecordAt(ctx context.Context, id uint64, resolvedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.AlertRecord{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      model.AlertStatusResolved,
			"resolved_at": resolvedAt,
		}).Error
}

// CountByStatus 按状态统计告警数量
func (r *AlertRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	type Result struct {
		Status string
		Count  int64
	}
	var results []Result

	err := r.db.WithContext(ctx).Model(&model.AlertRecord{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}

// DeleteOld 删除旧的告警记录
func (r *AlertRepository) DeleteOld(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := r.db.WithContext(ctx).
		Where("fired_at < ?", cutoff).
		Delete(&model.AlertRecord{})
	return result.RowsAffected, result.Error
}

// DeleteByTargetID 删除指定目标的所有告警记录
func (r *AlertRepository) DeleteByTargetID(ctx context.Context, targetID uint64) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("target_id = ?", targetID).
		Delete(&model.AlertRecord{})
	return result.RowsAffected, result.Error
}

// HasUnresolvedAlert 检查目标是否有未恢复的告警记录
// 实现 scheduler.AlertChecker 接口
func (r *AlertRepository) HasUnresolvedAlert(ctx context.Context, targetID uint64) bool {
	var count int64
	r.db.WithContext(ctx).Model(&model.AlertRecord{}).
		Where("target_id = ? AND status = ?", targetID, model.AlertStatusFiring).
		Count(&count)
	return count > 0
}
