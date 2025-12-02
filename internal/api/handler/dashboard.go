package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thingsboard-rxprobe/internal/repository"
	"github.com/thingsboard-rxprobe/internal/service"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	probeService *service.ProbeService
	alertService *service.AlertService
}

// NewDashboardHandler 创建仪表盘处理器
func NewDashboardHandler(probeService *service.ProbeService, alertService *service.AlertService) *DashboardHandler {
	return &DashboardHandler{
		probeService: probeService,
		alertService: alertService,
	}
}

// GetSummary 获取仪表盘概览
func (h *DashboardHandler) GetSummary(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取所有目标
	targets, total, err := h.probeService.ListTargets(ctx, repository.ListQuery{
		Page: 1,
		Size: 1000,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 统计各状态数量
	var healthyCount, unhealthyCount, unknownCount int64
	for _, t := range targets {
		switch t.Status {
		case "healthy":
			healthyCount++
		case "unhealthy":
			unhealthyCount++
		default:
			unknownCount++
		}
	}

	// 获取最近告警
	alerts, _, err := h.alertService.ListRecords(ctx, repository.AlertRecordQuery{
		Page: 1,
		Size: 10,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{
		"total_targets":   total,
		"healthy_count":   healthyCount,
		"unhealthy_count": unhealthyCount,
		"unknown_count":   unknownCount,
		"recent_alerts":   alerts,
	})
}

// GetMetrics 获取监控指标
func (h *DashboardHandler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取所有目标及其最新状态
	targets, _, err := h.probeService.ListTargets(ctx, repository.ListQuery{
		Page: 1,
		Size: 100,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	metrics := make([]gin.H, 0, len(targets))
	for _, t := range targets {
		stats, _ := h.probeService.GetTargetStats(ctx, t.ID)
		metrics = append(metrics, gin.H{
			"id":              t.ID,
			"name":            t.Name,
			"type":            t.Type,
			"status":          t.Status,
			"last_latency_ms": t.LastLatencyMs,
			"last_check_at":   t.LastCheckAt,
			"stats":           stats,
		})
	}

	Success(c, metrics)
}
