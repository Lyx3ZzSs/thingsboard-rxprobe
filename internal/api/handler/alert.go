package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
	recordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	// 路由是 /alerts/:id/silence，这里的 id 是告警记录 ID
	// 静默需要作用在目标维度上，因此需要先反查 target_id
	record, err := h.alertService.GetRecordByID(c.Request.Context(), recordID)
	if err != nil || record == nil {
		Error(c, http.StatusNotFound, "记录不存在")
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
	h.alertService.SilenceAlert(record.TargetID, duration)

	Success(c, nil)
}
