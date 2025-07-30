package router

import (
	"net/http"
	"time"

	"ai-monitor/internal/config"
	"ai-monitor/internal/handlers"
	"ai-monitor/internal/middleware"
	"ai-monitor/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup 设置路由
func Setup(cfg *config.Config, services *services.Services) *gin.Engine {
	// 创建Gin引擎
	r := gin.New()

	// 创建handlers
	h := handlers.NewHandlers(services)

	// 先设置静态文件路由（不受中间件影响）
	setupStaticRoutes(r)

	// 添加中间件
	setupMiddleware(r, cfg)

	// 设置其他路由
	setupRoutes(r, h, cfg)

	return r
}

// setupMiddleware 设置中间件
func setupMiddleware(r *gin.Engine, cfg *config.Config) {
	// 恢复中间件
	r.Use(middleware.Recovery())

	// 日志中间件
	r.Use(middleware.Logger())

	// CORS中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 安全中间件
	r.Use(middleware.Security())

	// 限流中间件
	r.Use(middleware.RateLimit())

	// 请求ID中间件
	r.Use(middleware.RequestID())

	// 指标中间件
	r.Use(middleware.Metrics())
}

// setupStaticRoutes 设置静态文件路由（不受中间件影响）
func setupStaticRoutes(r *gin.Engine) {
	// 静态文件服务
	staticGroup := r.Group("/static")
	staticGroup.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000")
		// 添加CORS头部
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Next()
	})
	staticGroup.Static("/", "./web/static")
	
	// 为了兼容前端构建的assets路径，添加assets路由
	assetsGroup := r.Group("/assets")
	assetsGroup.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000")
		// 添加CORS头部
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Next()
	})
	assetsGroup.Static("/", "./web/static/assets")
	
	r.StaticFile("/favicon.ico", "./web/static/favicon.ico")
}

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine, h *handlers.Handlers, cfg *config.Config) {
	// 获取数据库连接用于API Key认证
	db := h.Services.DB
	// Swagger文档
	if cfg.Server.Mode != "release" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 健康检查
	r.GET("/health", h.HealthCheck)
	r.GET("/version", h.GetVersion)

	// Prometheus指标
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// pprof性能分析（仅在开发模式下）
	if cfg.Server.Mode == "debug" {
		pprof.Register(r)
	}

	// API兼容路由组（为了兼容前端请求）
	apiCompat := r.Group("/api")
	{
		// 认证路由（无需认证）
		authCompat := apiCompat.Group("/auth")
		{
			authCompat.POST("/register", h.Register)
			authCompat.POST("/login", h.Login)
			authCompat.POST("/refresh", h.RefreshToken)
			authCompat.POST("/logout", middleware.Auth(), h.Logout)
		}
	}

	// API v1路由组
	api := r.Group("/api/v1")
	{
		// 认证路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
			auth.POST("/refresh", h.RefreshToken)
			auth.POST("/logout", middleware.Auth(), h.Logout)
		}

		// 用户管理路由（需要认证）
		users := api.Group("/users")
		users.Use(middleware.Auth())
		{
			users.GET("/profile", h.GetUserProfile)
			users.PUT("/profile", h.UpdateUserProfile)
			users.PUT("/password", h.ChangePassword)
		}

		// 告警管理路由（需要认证）
		alerts := api.Group("/alerts")
		alerts.Use(middleware.Auth())
		{
			alerts.GET("", h.GetAlerts)
			alerts.POST("", h.CreateAlert)
			alerts.GET("/:id", h.GetAlert)
			alerts.PUT("/:id", h.UpdateAlert)
			alerts.DELETE("/:id", h.DeleteAlert)
			alerts.POST("/:id/acknowledge", h.AcknowledgeAlert)
			alerts.POST("/:id/resolve", h.ResolveAlert)

			// 告警规则
			alerts.GET("/rules", h.GetAlertRules)
			alerts.POST("/rules", h.CreateAlertRule)
			alerts.GET("/rules/:id", h.GetAlertRule)
			alerts.PUT("/rules/:id", h.UpdateAlertRule)
			alerts.DELETE("/rules/:id", h.DeleteAlertRule)
			alerts.POST("/rules/:id/enable", h.EnableAlertRule)
			alerts.POST("/rules/:id/disable", h.DisableAlertRule)
		}

		// 监控数据路由（需要认证）
		monitoring := api.Group("/monitoring")
		monitoring.Use(middleware.Auth())
		{
			// 监控目标管理
			monitoring.GET("/targets", h.GetTargets)
			monitoring.POST("/targets", h.CreateTarget)
			monitoring.GET("/targets/:id", h.GetTarget)
			monitoring.PUT("/targets/:id", h.UpdateTarget)
			monitoring.DELETE("/targets/:id", h.DeleteTarget)
			monitoring.GET("/targets/:id/metrics", h.GetTargetMetrics)
			monitoring.GET("/targets/:id/status", h.GetTargetStatus)

			// 服务器监控管理
			monitoring.GET("/servers", h.GetServers)
			monitoring.POST("/servers", h.CreateServer)
			monitoring.PUT("/servers/:id", h.UpdateServer)
			monitoring.DELETE("/servers/:id", h.DeleteServer)

			// 进程监控管理
			monitoring.GET("/processes", h.GetProcesses)
			monitoring.POST("/processes", h.CreateProcess)
			monitoring.PUT("/processes/:id", h.UpdateProcess)
			monitoring.DELETE("/processes/:id", h.DeleteProcess)

			// 指标查询
			monitoring.GET("/metrics", h.QueryMetrics)
			monitoring.GET("/metrics/range", h.QueryRangeMetrics)
			monitoring.GET("/metrics/labels", h.GetMetricLabels)
			monitoring.GET("/metrics/values", h.GetMetricValues)
		}

		// AI分析路由（需要认证）
		ai := api.Group("/ai")
		ai.Use(middleware.Auth())
		{
			ai.POST("/analyze", h.AnalyzeData)
			ai.GET("/analyses", h.GetAnalyses)
			ai.GET("/analyses/:id", h.GetAnalysis)
			ai.DELETE("/analyses/:id", h.DeleteAnalysis)
			ai.POST("/predict", h.PredictTrend)
			ai.GET("/insights", h.GetInsights)

			// 知识库管理
			ai.GET("/knowledge-base", h.GetKnowledgeBases)
			ai.POST("/knowledge-base", h.CreateKnowledgeBase)
			ai.GET("/knowledge-base/:id", h.GetKnowledgeBase)
			ai.PUT("/knowledge-base/:id", h.UpdateKnowledgeBase)
			ai.DELETE("/knowledge-base/:id", h.DeleteKnowledgeBase)
			ai.GET("/knowledge-base/stats", h.GetKnowledgeBaseStats)
			ai.GET("/knowledge-base/export", h.ExportKnowledgeBase)
		}

		// 中间件监控路由（需要认证）
		middlewareGroup := api.Group("/middleware")
		middlewareGroup.Use(middleware.Auth())
		{
			middlewareGroup.GET("/mysql/metrics", h.GetMySQLMetrics)
			middlewareGroup.GET("/redis/metrics", h.GetRedisMetrics)
			middlewareGroup.GET("/kafka/metrics", h.GetKafkaMetrics)
			middlewareGroup.GET("/list", h.GetMiddlewareList)
			middlewareGroup.POST("", h.CreateMiddleware)
			middlewareGroup.PUT("/:id", h.UpdateMiddleware)
			middlewareGroup.DELETE("/:id", h.DeleteMiddleware)
		}

		// APM应用性能监控路由（需要认证）
		apm := api.Group("/apm")
		apm.Use(middleware.Auth())
		{
			apm.GET("/services", h.GetServices)
			apm.GET("/services/:name/performance", h.GetServicePerformance)
			apm.GET("/services/:name/operations", h.GetOperations)
			apm.GET("/service-map", h.GetServiceMap)
			apm.POST("/services", h.CreateService)
			apm.PUT("/services/:name", h.UpdateService)
			apm.DELETE("/services/:name", h.DeleteService)
		}

		// 容器监控路由（需要认证）
		containers := api.Group("/containers")
		containers.Use(middleware.Auth())
		{
			containers.GET("/docker", h.GetDockerContainers)
			containers.GET("/kubernetes/pods", h.GetKubernetesPods)
			containers.GET("/kubernetes/nodes", h.GetKubernetesNodes)
			containers.GET("/kubernetes/namespaces", h.GetKubernetesNamespaces)
			containers.GET("/cluster/metrics", h.GetClusterMetrics)
			containers.GET("/resource-usage", h.GetResourceUsage)
			containers.POST("", h.CreateContainerMonitor)
			containers.PUT("/:id", h.UpdateContainerMonitor)
			containers.DELETE("/:id", h.DeleteContainerMonitor)
		}

		// Agent下载和安装指南路由（无需认证）
		api.GET("/agents/download/:type", h.DownloadAgentPackage)
		api.GET("/agents/install-guide/:type", h.GetAgentInstallGuide)

		// Agent路由组
		agents := api.Group("/agents")
		{
			// Agent注册和心跳路由（支持API Key或JWT认证）
			agents.POST("", middleware.APIKeyOrJWTAuthMiddleware(h.Services.JWTManager, db), h.CreateAgent)
			agents.POST("/heartbeat", middleware.APIKeyOrJWTAuthMiddleware(h.Services.JWTManager, db), h.HandleAgentHeartbeat)
			
			// Agent管理路由（需要JWT认证）
			agents.GET("", middleware.Auth(), h.GetAgents)
			agents.GET("/:id", middleware.Auth(), h.GetAgent)
			agents.PUT("/:id", middleware.Auth(), h.UpdateAgent)
			agents.DELETE("/:id", middleware.Auth(), h.DeleteAgent)
			agents.GET("/:id/config", middleware.Auth(), h.GetAgentConfig)
			agents.PUT("/:id/config", middleware.Auth(), h.UpdateAgentConfig)
			agents.GET("/packages", middleware.Auth(), h.GetAgentPackages)

			// Agent部署（需要JWT认证）
			agents.POST("/deployments", middleware.Auth(), h.CreateDeployment)
			agents.GET("/deployments", middleware.Auth(), h.GetDeployments)
			agents.GET("/deployments/:id", middleware.Auth(), h.GetDeployment)
		}

		// API Key管理路由（需要JWT认证）
		apiKeys := api.Group("/api-keys")
		apiKeys.Use(middleware.Auth())
		{
			apiKeys.POST("", h.CreateAPIKey)
			apiKeys.GET("", h.ListAPIKeys)
			apiKeys.GET("/:id", h.GetAPIKey)
			apiKeys.PUT("/:id", h.UpdateAPIKey)
			apiKeys.DELETE("/:id", h.DeleteAPIKey)
			apiKeys.POST("/generate", h.GenerateAPIKey)
			apiKeys.POST("/validate", h.ValidateAPIKey)
		}

		// 服务发现路由（需要JWT认证）
		discovery := api.Group("/discovery")
		discovery.Use(middleware.Auth())
		{
			// 发现任务管理
			discovery.POST("/tasks", h.CreateDiscoveryTask)
			discovery.GET("/tasks", h.ListDiscoveryTasks)
			discovery.GET("/tasks/:id", h.GetDiscoveryTask)
			discovery.GET("/tasks/:id/results", h.GetDiscoveryTaskResults)
			discovery.GET("/tasks/:id/progress", h.GetDiscoveryTaskProgress)
			
			// 发现统计
			discovery.GET("/stats", h.GetDiscoveryStats)
		}

		// 系统配置路由（需要认证）
		config := api.Group("/config")
		config.Use(middleware.Auth())
		{
			// 基础配置管理
			config.GET("", h.GetConfigs)
			config.GET("/:key", h.GetConfig)
			config.PUT("/:key", h.UpdateConfig)
			config.POST("", h.CreateConfig)
			config.DELETE("/:key", h.DeleteConfig)

			// 专项配置管理
			config.GET("/database", h.GetDatabaseConfig)
			config.PUT("/database", h.UpdateDatabaseConfig)
			config.GET("/redis", h.GetRedisConfig)
			config.PUT("/redis", h.UpdateRedisConfig)
			config.GET("/ai-model", h.GetAIModelConfig)
			config.PUT("/ai-model", h.UpdateAIModelConfig)
			config.GET("/email", h.GetEmailConfig)
			config.PUT("/email", h.UpdateEmailConfig)
			config.GET("/prometheus", h.GetPrometheusConfig)
			config.PUT("/prometheus", h.UpdatePrometheusConfig)
			config.GET("/system", h.GetSystemConfig)
			config.PUT("/system", h.UpdateSystemConfig)

			// 新增配置路由
			config.GET("/alert", h.GetAlertConfig)
			config.PUT("/alert", h.UpdateAlertConfig)
			config.POST("/alert/test-email", h.TestEmailConfig)
			config.POST("/alert/test-sms", h.TestSMSConfig)
			config.GET("/ai-service", h.GetAIServiceConfig)
			config.PUT("/ai-service", h.UpdateAIServiceConfig)
			config.GET("/system-settings", h.GetSystemConfig)
			config.PUT("/system-settings", h.UpdateSystemConfig)
		}

		// 管理员专用路由（需要管理员权限）
		admin := api.Group("/admin")
		admin.Use(middleware.Auth(), middleware.RequireRole("admin"))
		{
			// 用户管理
			admin.GET("/users", h.GetUsers)
			admin.POST("/users", h.CreateUser)
			admin.GET("/users/:id", h.GetUser)
			admin.PUT("/users/:id", h.UpdateUser)
			admin.DELETE("/users/:id", h.DeleteUser)
			admin.PUT("/users/:id/status", h.UpdateUserStatus)
			admin.PUT("/users/:id/roles", h.UpdateUserRoles)

			// 角色权限管理
			admin.GET("/roles", h.GetRoles)
			admin.POST("/roles", h.CreateRole)
			admin.PUT("/roles/:id", h.UpdateRole)
			admin.DELETE("/roles/:id", h.DeleteRole)
			admin.GET("/permissions", h.GetPermissions)

			// 系统管理
			admin.GET("/system/info", h.GetSystemInfo)
			admin.GET("/audit/logs", h.GetAuditLogs)
			admin.POST("/system/backup", h.CreateSystemBackup)
			admin.GET("/system/backups", h.GetSystemBackups)
		}
	}

	// WebSocket路由
	ws := r.Group("/ws")
	ws.Use(middleware.Auth())
	{
		ws.GET("/alerts", h.HandleAlertWebSocket)
		ws.GET("/metrics", h.HandleMetricsWebSocket)
		ws.GET("/logs", h.HandleLogsWebSocket)
	}

	// 公开API（无需认证）
	public := r.Group("/public")
	{
		public.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
				"service":   "ai-monitor",
			})
		})
	}

	// 前端路由（SPA支持）
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/static/index.html")
	})
}