package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thingsboard-rxprobe/internal/alerter"
	"github.com/thingsboard-rxprobe/internal/api"
	"github.com/thingsboard-rxprobe/internal/config"
	"github.com/thingsboard-rxprobe/internal/prober"
	"github.com/thingsboard-rxprobe/internal/repository"
	"github.com/thingsboard-rxprobe/internal/scheduler"
	"github.com/thingsboard-rxprobe/internal/service"
	"github.com/thingsboard-rxprobe/pkg/database"
	"github.com/thingsboard-rxprobe/pkg/logger"
	"go.uber.org/zap"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "configs/config.yaml", "配置文件路径")
}

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.Format)
	defer logger.Sync()

	logger.Info("正在启动 Thingsboard RxProbe 探针系统...")

	// 初始化数据库
	if err := database.Init(database.Config{
		Driver:       cfg.Database.Driver,
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		DBName:       cfg.Database.DBName,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	}); err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer database.Close()

	db := database.GetDB()

	// 创建仓库
	userRepo := repository.NewUserRepository(db)
	targetRepo := repository.NewTargetRepository(db)
	resultRepo := repository.NewResultRepository(db)
	alertRepo := repository.NewAlertRepository(db)

	// 创建探针工厂
	proberFactory := prober.NewFactory()

	// 创建调度器
	sch := scheduler.NewScheduler(proberFactory)

	// 创建告警器
	var alerterInstance alerter.Alerter
	if cfg.Alerter.WeCom.Enabled && cfg.Alerter.WeCom.WebhookURL != "" {
		alerterInstance = alerter.NewWeComAlerter(cfg.Alerter.WeCom.WebhookURL)
		logger.Info("企业微信告警已启用")
	}

	// 创建服务
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHours)
	probeService := service.NewProbeService(targetRepo, resultRepo, proberFactory, sch)
	alertService := service.NewAlertService(alertRepo, targetRepo, resultRepo, alerterInstance, sch)
	cleanupService := service.NewCleanupService(resultRepo, alertRepo, cfg.Scheduler.ResultRetentionDays)

	// 初始化默认管理员
	if err := authService.InitDefaultAdmin(context.Background(), "admin123"); err != nil {
		logger.Error("初始化默认管理员失败", zap.Error(err))
	}

	// 创建路由
	router := api.NewRouter(authService, probeService, alertService)
	engine := router.Setup(cfg.Server.Mode)

	// 创建 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	// 启动调度器和告警服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sch.Start(ctx)
	alertService.Start(ctx)
	cleanupService.Start()

	// 加载已启用的探测目标
	if err := probeService.LoadEnabledTargets(ctx); err != nil {
		logger.Error("加载探测目标失败", zap.Error(err))
	}

	// 启动 HTTP 服务器
	go func() {
		logger.Info("HTTP 服务器启动",
			zap.String("addr", addr),
			zap.String("mode", cfg.Server.Mode),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP 服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	sch.Stop()
	alertService.Stop()
	cleanupService.Stop()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("服务器关闭失败", zap.Error(err))
	}

	logger.Info("服务器已关闭")
}
