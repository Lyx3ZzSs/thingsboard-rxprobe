package api

import (
	"github.com/gin-gonic/gin"
	"github.com/thingsboard-rxprobe/internal/api/handler"
	"github.com/thingsboard-rxprobe/internal/api/middleware"
	"github.com/thingsboard-rxprobe/internal/repository"
	"github.com/thingsboard-rxprobe/internal/service"
)

// Router 路由管理器
type Router struct {
	engine       *gin.Engine
	probeService *service.ProbeService
	alertService *service.AlertService
	authService  *service.AuthService
	notifierRepo *repository.NotifierRepository
}

// NewRouter 创建路由管理器
func NewRouter(
	probeService *service.ProbeService,
	alertService *service.AlertService,
	authService *service.AuthService,
	notifierRepo *repository.NotifierRepository,
) *Router {
	return &Router{
		probeService: probeService,
		alertService: alertService,
		authService:  authService,
		notifierRepo: notifierRepo,
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
	notifierHandler := handler.NewNotifierHandler(r.notifierRepo)
	authHandler := handler.NewAuthHandler(r.authService)

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// 公开路由（不需要认证）
		public := v1.Group("")
		{
			// 认证相关
			public.POST("/auth/login", authHandler.Login)
			public.GET("/auth/check-init", authHandler.CheckInit)
			public.POST("/auth/init", authHandler.InitSystem)
		}

		// 需要认证的路由
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(r.authService))
		{
			// 认证相关（需要登录）
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.POST("/auth/change-password", authHandler.ChangePassword)

			// 探针类型和 Schema
			protected.GET("/probe/types", probeHandler.GetProbeTypes)
			protected.GET("/probe/schema/:type", probeHandler.GetProbeSchema)
			protected.POST("/probe/test", probeHandler.TestTarget)

			// 探测目标管理
			targets := protected.Group("/targets")
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
			alerts := protected.Group("/alerts")
			{
				alerts.GET("", alertHandler.ListRecords)
				alerts.GET("/:id", alertHandler.GetRecord)
				alerts.PUT("/:id/silence", alertHandler.SilenceAlert)
			}

			// 仪表盘
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/summary", dashboardHandler.GetSummary)
				dashboard.GET("/metrics", dashboardHandler.GetMetrics)
			}

			// 通知渠道
			notifiers := protected.Group("/notifiers")
			{
				notifiers.GET("", notifierHandler.List)
				notifiers.POST("", notifierHandler.Create)
				notifiers.GET("/types", notifierHandler.GetTypes)
				notifiers.POST("/test", notifierHandler.Test)
				notifiers.GET("/:id", notifierHandler.Get)
				notifiers.PUT("/:id", notifierHandler.Update)
				notifiers.DELETE("/:id", notifierHandler.Delete)
			}
		}
	}

	r.engine = engine
	return engine
}

// GetEngine 获取 Gin 引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
