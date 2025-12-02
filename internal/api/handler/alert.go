package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thingsboard-rxprobe/internal/model"
	"github.com/thingsboard-rxprobe/internal/repository"
	"github.com/thingsboard-rxprobe/internal/service"
)

// AlertHandler 告警处理器
type AlertHandler struct {
	alertService *service.AlertService
}

// NewAlertHandler 创建告警处理器
func NewAlertHandler(alertService *service.AlertService) *AlertHandler {
	return &AlertHandler{alertService: alertService}
}

// ListRules 获取告警规则列表
func (h *AlertHandler) ListRules(c *gin.Context) {
	rules, err := h.alertService.ListRules(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	Success(c, rules)
}

// CreateRule 创建告警规则
func (h *AlertHandler) CreateRule(c *gin.Context) {
	var req model.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	rule, err := h.alertService.CreateRule(c.Request.Context(), &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, rule)
}

// UpdateRule 更新告警规则
func (h *AlertHandler) UpdateRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	var req model.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	rule, err := h.alertService.UpdateRule(c.Request.Context(), id, &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, rule)
}

// DeleteRule 删除告警规则
func (h *AlertHandler) DeleteRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	if err := h.alertService.DeleteRule(c.Request.Context(), id); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, nil)
}

// ListRecords 获取告警记录列表
func (h *AlertHandler) ListRecords(c *gin.Context) {
	var query repository.AlertRecordQuery
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.Size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))
	query.Status = c.Query("status")

	if targetIDStr := c.Query("target_id"); targetIDStr != "" {
		targetID, _ := strconv.ParseUint(targetIDStr, 10, 64)
		query.TargetID = targetID
	}

	records, total, err := h.alertService.ListRecords(c.Request.Context(), query)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessList(c, records, total, query.Page, query.Size)
}

// GetRecord 获取告警记录详情
func (h *AlertHandler) GetRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	record, err := h.alertService.GetRecordByID(c.Request.Context(), id)
	if err != nil {
		Error(c, http.StatusNotFound, "记录不存在")
		return
	}

	Success(c, record)
}

// SilenceAlert 静默告警
func (h *AlertHandler) SilenceAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	var req struct {
		DurationMinutes int `json:"duration_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 30
	}

	duration := time.Duration(req.DurationMinutes) * time.Minute
	h.alertService.SilenceAlert(id, duration)

	Success(c, nil)
}
