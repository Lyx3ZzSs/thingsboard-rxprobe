package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thingsboard-rxprobe/internal/model"
	"github.com/thingsboard-rxprobe/internal/repository"
)

// NotifierHandler 通知渠道处理器
type NotifierHandler struct {
	repo *repository.NotifierRepository
}

// NewNotifierHandler 创建通知渠道处理器
func NewNotifierHandler(repo *repository.NotifierRepository) *NotifierHandler {
	return &NotifierHandler{repo: repo}
}

// CreateNotifierRequest 创建通知渠道请求
type CreateNotifierRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	WebhookURL  string `json:"webhook_url" binding:"required"`
	MessageTpl  string `json:"message_tpl"`
	MentionAll  bool   `json:"mention_all"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// UpdateNotifierRequest 更新通知渠道请求
type UpdateNotifierRequest struct {
	Name        string `json:"name"`
	WebhookURL  string `json:"webhook_url"`
	MessageTpl  string `json:"message_tpl"`
	MentionAll  *bool  `json:"mention_all"`
	Enabled     *bool  `json:"enabled"`
	Description string `json:"description"`
}

// List 获取通知渠道列表
func (h *NotifierHandler) List(c *gin.Context) {
	channels, err := h.repo.List(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取通知渠道列表失败")
		return
	}
	Success(c, channels)
}

// Create 创建通知渠道
func (h *NotifierHandler) Create(c *gin.Context) {
	var req CreateNotifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 验证通知类型
	if req.Type != string(model.NotifyChannelTypeWeCom) {
		Error(c, http.StatusBadRequest, "不支持的通知类型")
		return
	}

	// 验证 webhook URL
	if !strings.HasPrefix(req.WebhookURL, "http://") && !strings.HasPrefix(req.WebhookURL, "https://") {
		Error(c, http.StatusBadRequest, "Webhook URL 格式无效")
		return
	}

	channel := &model.NotifyChannel{
		Name:        req.Name,
		Type:        model.NotifyChannelType(req.Type),
		WebhookURL:  req.WebhookURL,
		MessageTpl:  req.MessageTpl,
		MentionAll:  req.MentionAll,
		Enabled:     req.Enabled,
		Description: req.Description,
	}

	if err := h.repo.Create(c.Request.Context(), channel); err != nil {
		Error(c, http.StatusInternalServerError, "创建通知渠道失败")
		return
	}

	Success(c, channel)
}

// Get 获取通知渠道详情
func (h *NotifierHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	channel, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, http.StatusNotFound, "通知渠道不存在")
		return
	}

	Success(c, channel)
}

// Update 更新通知渠道
func (h *NotifierHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	channel, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, http.StatusNotFound, "通知渠道不存在")
		return
	}

	var req UpdateNotifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 更新字段
	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.WebhookURL != "" {
		if !strings.HasPrefix(req.WebhookURL, "http://") && !strings.HasPrefix(req.WebhookURL, "https://") {
			Error(c, http.StatusBadRequest, "Webhook URL 格式无效")
			return
		}
		channel.WebhookURL = req.WebhookURL
	}
	if req.MessageTpl != "" {
		channel.MessageTpl = req.MessageTpl
	}
	if req.MentionAll != nil {
		channel.MentionAll = *req.MentionAll
	}
	if req.Enabled != nil {
		channel.Enabled = *req.Enabled
	}
	if req.Description != "" {
		channel.Description = req.Description
	}

	if err := h.repo.Update(c.Request.Context(), channel); err != nil {
		Error(c, http.StatusInternalServerError, "更新通知渠道失败")
		return
	}

	Success(c, channel)
}

// Delete 删除通知渠道
func (h *NotifierHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		Error(c, http.StatusInternalServerError, "删除通知渠道失败")
		return
	}

	Success(c, nil)
}

// TestNotifierRequest 测试通知渠道请求
type TestNotifierRequest struct {
	WebhookURL string `json:"webhook_url" binding:"required"`
	Type       string `json:"type" binding:"required"`
	MentionAll bool   `json:"mention_all"`
}

// Test 测试通知渠道
func (h *NotifierHandler) Test(c *gin.Context) {
	var req TestNotifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 验证通知类型
	if req.Type != string(model.NotifyChannelTypeWeCom) {
		Error(c, http.StatusBadRequest, "不支持的通知类型")
		return
	}

	// 构建测试消息（使用默认格式）
	message := fmt.Sprintf(`🚨 测试通知

目标：测试目标
类型：HTTP
原因：这是一条测试消息
时间：%s`, time.Now().Format("2006-01-02 15:04:05"))

	// 发送测试消息
	if err := sendWeComMessage(c.Request.Context(), req.WebhookURL, message, req.MentionAll); err != nil {
		Error(c, http.StatusInternalServerError, "发送测试消息失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "测试消息发送成功"})
}

// GetTypes 获取支持的通知类型
func (h *NotifierHandler) GetTypes(c *gin.Context) {
	types := []gin.H{
		{
			"type":        "wecom",
			"name":        "企业微信",
			"description": "通过企业微信群机器人发送通知",
		},
	}
	Success(c, types)
}

// sendWeComMessage 发送企业微信消息
func sendWeComMessage(ctx context.Context, webhookURL, content string, mentionAll bool) error {
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
	if mentionAll {
		msg.Text.MentionedList = []string{"@all"}
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(data))
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
