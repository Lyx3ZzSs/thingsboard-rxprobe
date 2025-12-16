package repository

import (
	"context"
	"time"

	"github.com/thingsboard-rxprobe/internal/model"
	"gorm.io/gorm"
)

// TargetRepository 探测目标仓库
type TargetRepository struct {
	db *gorm.DB
}

// NewTargetRepository 创建探测目标仓库
func NewTargetRepository(db *gorm.DB) *TargetRepository {
	return &TargetRepository{db: db}
}

// Create 创建探测目标
func (r *TargetRepository) Create(ctx context.Context, target *model.ProbeTarget) error {
	return r.db.WithContext(ctx).Create(target).Error
}

// Update 更新探测目标
func (r *TargetRepository) Update(ctx context.Context, target *model.ProbeTarget) error {
	return r.db.WithContext(ctx).Save(target).Error
}

// Delete 删除探测目标
func (r *TargetRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.ProbeTarget{}, id).Error
}

// GetByID 根据 ID 获取探测目标
func (r *TargetRepository) GetByID(ctx context.Context, id uint64) (*model.ProbeTarget, error) {
	var target model.ProbeTarget
	err := r.db.WithContext(ctx).First(&target, id).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

// ListQuery 列表查询参数
type ListQuery struct {
	Keyword string
	Type    string
	Status  string
	Group   string
	Enabled *bool
	Page    int
	Size    int
}

// List 获取探测目标列表
func (r *TargetRepository) List(ctx context.Context, query ListQuery) ([]*model.ProbeTarget, int64, error) {
	var targets []*model.ProbeTarget
	var total int64

	db := r.db.WithContext(ctx).Model(&model.ProbeTarget{})

	if query.Keyword != "" {
		db = db.Where("name LIKE ?", "%"+query.Keyword+"%")
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Group != "" {
		db = db.Where("\"group\" = ?", query.Group)
	}
	if query.Enabled != nil {
		db = db.Where("enabled = ?", *query.Enabled)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Page > 0 && query.Size > 0 {
		offset := (query.Page - 1) * query.Size
		db = db.Offset(offset).Limit(query.Size)
	}

	if err := db.Order("id DESC").Find(&targets).Error; err != nil {
		return nil, 0, err
	}

	return targets, total, nil
}

// ListEnabled 获取所有启用的探测目标
func (r *TargetRepository) ListEnabled(ctx context.Context) ([]*model.ProbeTarget, error) {
	var targets []*model.ProbeTarget
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&targets).Error
	return targets, err
}

// UpdateStatus 更新探测状态
func (r *TargetRepository) UpdateStatus(ctx context.Context, id uint64, status string, latencyMs int64, message string) error {
	return r.db.WithContext(ctx).Model(&model.ProbeTarget{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":          status,
			"last_latency_ms": latencyMs,
			"last_message":    message,
			"last_check_at":   time.Now(),
		}).Error
}
