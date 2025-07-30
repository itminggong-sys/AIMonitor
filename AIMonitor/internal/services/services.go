package services

import (
	"context"
	"fmt"

	"ai-monitor/internal/auth"
	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"

	"gorm.io/gorm"
)

// Services 服务集合
type Services struct {
	UserService         *UserService
	AlertService        *AlertService
	NotificationService *NotificationService
	AIService           *AIService
	MonitoringService   *MonitoringService
	ConfigService       *ConfigService
	AuditService        *AuditService
	MiddlewareService   *MiddlewareService
	APMService          *APMService
	ContainerService    *ContainerService
	AgentService        *AgentService
	APIKeyService       *APIKeyService
	DiscoveryService    *DiscoveryService

	// 数据库连接
	DB *gorm.DB
	// 缓存管理器
	cacheManager *cache.CacheManager
	// JWT管理器
	JWTManager *auth.JWTManager
}

// NewServices 创建服务集合
func NewServices(cfg *config.Config, db *gorm.DB) (*Services, error) {
	// 初始化缓存管理器
	cacheManager := cache.NewCacheManager()

	// 初始化JWT管理器
	jwtManager := auth.NewJWTManager(&cfg.JWT)

	// 创建基础服务
	userService := NewUserService(db, cacheManager, jwtManager)
	notificationService := NewNotificationService(db, cacheManager, cfg)
	aiService := NewAIService(db, cacheManager, cfg)
	alertService := NewAlertService(db, cacheManager, notificationService, aiService)
	configService := NewConfigService(db, cacheManager, cfg)
	auditService := NewAuditService(db, cacheManager, cfg)
	agentService := NewAgentService(db, cacheManager, cfg)
	apikeyService := NewAPIKeyService(db)

	// 创建需要依赖其他服务的服务
	monitoringService, err := NewMonitoringService(db, cacheManager, cfg, alertService)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitoring service: %w", err)
	}

	middlewareService, err := NewMiddlewareService(db, cacheManager, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create middleware service: %w", err)
	}

	apmService, err := NewAPMService(db, cacheManager, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create apm service: %w", err)
	}

	containerService, err := NewContainerService(db, cacheManager, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create container service: %w", err)
	}

	// 创建发现服务
	discoveryService := NewDiscoveryService(db, cacheManager, cfg, agentService, apikeyService)

	return &Services{
		UserService:         userService,
		AlertService:        alertService,
		NotificationService: notificationService,
		AIService:           aiService,
		MonitoringService:   monitoringService,
		ConfigService:       configService,
		AuditService:        auditService,
		MiddlewareService:   middlewareService,
		APMService:          apmService,
		ContainerService:    containerService,
		AgentService:        agentService,
		APIKeyService:       apikeyService,
		DiscoveryService:    discoveryService,
		DB:                  db,
		cacheManager:        cacheManager,
		JWTManager:          jwtManager,
	}, nil
}

// Start 启动所有服务
func (s *Services) Start(ctx context.Context) error {
	// 缓存管理器不需要显式启动
	// 这里可以添加其他需要启动的服务
	// 例如：定时任务、后台处理器等

	return nil
}

// Stop 停止所有服务
func (s *Services) Stop() {
	// 缓存管理器不需要显式停止
	// 这里可以添加其他需要停止的服务
}

// GetCacheManager 获取缓存管理器
func (s *Services) GetCacheManager() *cache.CacheManager {
	return s.cacheManager
}

// GetJWTManager 获取JWT管理器
func (s *Services) GetJWTManager() *auth.JWTManager {
	return s.JWTManager
}