package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thingsboard-rxprobe/internal/model"
	"github.com/thingsboard-rxprobe/internal/repository"
	"github.com/thingsboard-rxprobe/internal/service"
)

// ProbeHandler 探针处理器
type ProbeHandler struct {
	probeService *service.ProbeService
}

// NewProbeHandler 创建探针处理器
func NewProbeHandler(probeService *service.ProbeService) *ProbeHandler {
	return &ProbeHandler{probeService: probeService}
}

// GetProbeTypes 获取探针类型列表
func (h *ProbeHandler) GetProbeTypes(c *gin.Context) {
	types := h.probeService.GetProbeTypes()
	Success(c, types)
}

// GetProbeSchema 获取探针配置 Schema
func (h *ProbeHandler) GetProbeSchema(c *gin.Context) {
	probeType := c.Param("type")

	schema, err := h.probeService.GetProbeSchema(probeType)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, schema)
}

// CreateTarget 创建探测目标
func (h *ProbeHandler) CreateTarget(c *gin.Context) {
	var req model.CreateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	target, err := h.probeService.CreateTarget(c.Request.Context(), &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, target)
}

// UpdateTarget 更新探测目标
func (h *ProbeHandler) UpdateTarget(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	var req model.UpdateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	target, err := h.probeService.UpdateTarget(c.Request.Context(), id, &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, target)
}

// DeleteTarget 删除探测目标
func (h *ProbeHandler) DeleteTarget(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	if err := h.probeService.DeleteTarget(c.Request.Context(), id); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, nil)
}

// GetTarget 获取探测目标
func (h *ProbeHandler) GetTarget(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	target, err := h.probeService.GetTarget(c.Request.Context(), id)
	if err != nil {
		Error(c, http.StatusNotFound, "目标不存在")
		return
	}

	Success(c, target)
}

// ListTargets 获取探测目标列表
func (h *ProbeHandler) ListTargets(c *gin.Context) {
	var query repository.ListQuery
	query.Keyword = c.Query("keyword")
	query.Type = c.Query("type")
	query.Status = c.Query("status")
	query.Group = c.Query("group")
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.Size, _ = strconv.Atoi(c.DefaultQuery("size", "10"))

	if enabledStr := c.Query("enabled"); enabledStr != "" {
		enabled := enabledStr == "true"
		query.Enabled = &enabled
	}

	targets, total, err := h.probeService.ListTargets(c.Request.Context(), query)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessList(c, targets, total, query.Page, query.Size)
}

// TestTarget 测试探测目标
func (h *ProbeHandler) TestTarget(c *gin.Context) {
	var req model.TestTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.probeService.TestTarget(c.Request.Context(), &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{
		"success":    result.Success,
		"latency_ms": result.Latency.Milliseconds(),
		"message":    result.Message,
		"metrics":    result.Metrics,
		"warnings":   result.Warnings,
	})
}

// GetTargetResults 获取目标探测结果
func (h *ProbeHandler) GetTargetResults(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	var query repository.ResultQuery
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.Size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))

	results, total, err := h.probeService.GetTargetResults(c.Request.Context(), id, query)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessList(c, results, total, query.Page, query.Size)
}

// GetTargetStats 获取目标统计信息
func (h *ProbeHandler) GetTargetStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的 ID")
		return
	}

	stats, err := h.probeService.GetTargetStats(c.Request.Context(), id)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, stats)
}
