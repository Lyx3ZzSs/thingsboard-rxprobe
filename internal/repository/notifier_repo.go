package repository

import (
	"context"

	"github.com/thingsboard-rxprobe/internal/model"
	"gorm.io/gorm"
)

// NotifierRepository 通知渠道仓库
type NotifierRepository struct {
	db *gorm.DB
}

// NewNotifierRepository 创建通知渠道仓库
func NewNotifierRepository(db *gorm.DB) *NotifierRepository {
	return &NotifierRepository{db: db}
}

// Create 创建通知渠道
func (r *NotifierRepository) Create(ctx context.Context, channel *model.NotifyChannel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

// Update 更新通知渠道
func (r *NotifierRepository) Update(ctx context.Context, channel *model.NotifyChannel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

// Delete 删除通知渠道
func (r *NotifierRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.NotifyChannel{}, id).Error
}

// GetByID 根据 ID 获取通知渠道
func (r *NotifierRepository) GetByID(ctx context.Context, id uint64) (*model.NotifyChannel, error) {
	var channel model.NotifyChannel
	err := r.db.WithContext(ctx).First(&channel, id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// List 获取通知渠道列表
func (r *NotifierRepository) List(ctx context.Context) ([]*model.NotifyChannel, error) {
	var channels []*model.NotifyChannel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// ListEnabled 获取已启用的通知渠道列表
func (r *NotifierRepository) ListEnabled(ctx context.Context) ([]*model.NotifyChannel, error) {
	var channels []*model.NotifyChannel
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// ListByType 根据类型获取通知渠道列表
func (r *NotifierRepository) ListByType(ctx context.Context, channelType model.NotifyChannelType) ([]*model.NotifyChannel, error) {
	var channels []*model.NotifyChannel
	err := r.db.WithContext(ctx).Where("type = ?", channelType).Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}
