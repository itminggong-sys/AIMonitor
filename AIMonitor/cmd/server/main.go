package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"
	"ai-monitor/internal/database"
	"ai-monitor/internal/logger"
	"ai-monitor/internal/middleware"
	"ai-monitor/internal/router"
	"ai-monitor/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title AI监控系统 API
// @version 1.0
// @description AI智能监控系统的RESTful API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 初始化配置
	// 检查是否指定了配置文件
	configFile := "config"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	
	cfg, err := config.LoadWithFile(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		Output:     cfg.Logging.Output,
		FilePath:   cfg.Logging.FilePath,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	})
	log := logrus.WithField("component", "main")

	log.Info("Starting AI Monitor System...")

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := database.Initialize(&cfg.Database); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		if database.DB != nil {
			if sqlDB, err := database.DB.DB(); err == nil {
				sqlDB.Close()
			}
		}
	}()

	// 初始化Redis缓存
	if err := cache.Initialize(&cfg.Redis); err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
	defer cache.Close()

	// 运行数据库迁移
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// 初始化服务
	services, err := services.NewServices(cfg, database.DB)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// 启动后台服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := services.Start(ctx); err != nil {
		log.Fatalf("Failed to start services: %v", err)
	}

	// 初始化中间件依赖
	middleware.InitializeMiddleware(services.GetJWTManager(), services.GetCacheManager(), cfg)

	// 初始化路由
	r := router.Setup(cfg, services)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// 启动服务器
	go func() {
		log.Infof("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 停止后台服务
	cancel()
	services.Stop()

	// 关闭HTTP服务器
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	} else {
		log.Info("Server exited gracefully")
	}
}

// 健康检查处理器
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"service":   "ai-monitor",
	})
}

// 版本信息处理器
func versionInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":     "1.0.0",
		"build_time":  "2024-01-01T00:00:00Z",
		"git_commit":  "unknown",
		"go_version":  "go1.21",
		"service":     "ai-monitor",
		"description": "AI智能监控系统",
	})
}