package api

import (
	"github.com/gin-gonic/gin"
	"github.com/thingsboard-rxprobe/internal/api/handler"
	"github.com/thingsboard-rxprobe/internal/api/middleware"
	"github.com/thingsboard-rxprobe/internal/service"
)

// Router 路由管理器
type Router struct {
	engine       *gin.Engine
	probeService *service.ProbeService
	alertService *service.AlertService
}

// NewRouter 创建路由管理器
func NewRouter(
	probeService *service.ProbeService,
	alertService *service.AlertService,
) *Router {
	return &Router{
		probeService: probeService,
		alertService: alertService,
	}
}

// Setup 设置路由
func (r *Router) Setup(mode string) *gin.Engine {
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORSMiddleware())

	// 创建处理器
	probeHandler := handler.NewProbeHandler(r.probeService)
	alertHandler := handler.NewAlertHandler(r.alertService)
	dashboardHandler := handler.NewDashboardHandler(r.probeService, r.alertService)

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// 探针类型和 Schema
		v1.GET("/probe/types", probeHandler.GetProbeTypes)
		v1.GET("/probe/schema/:type", probeHandler.GetProbeSchema)
		v1.POST("/probe/test", probeHandler.TestTarget)

		// 探测目标管理
		targets := v1.Group("/targets")
		{
			targets.GET("", probeHandler.ListTargets)
			targets.POST("", probeHandler.CreateTarget)
			targets.GET("/:id", probeHandler.GetTarget)
			targets.PUT("/:id", probeHandler.UpdateTarget)
			targets.DELETE("/:id", probeHandler.DeleteTarget)
			targets.GET("/:id/results", probeHandler.GetTargetResults)
			targets.GET("/:id/stats", probeHandler.GetTargetStats)
		}

		// 告警记录
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", alertHandler.ListRecords)
			alerts.GET("/:id", alertHandler.GetRecord)
			alerts.PUT("/:id/silence", alertHandler.SilenceAlert)
		}

		// 仪表盘
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/summary", dashboardHandler.GetSummary)
			dashboard.GET("/metrics", dashboardHandler.GetMetrics)
		}
	}

	r.engine = engine
	return engine
}

// GetEngine 获取 Gin 引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
