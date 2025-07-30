package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"
	"ai-monitor/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConfigService 配置服务
type ConfigService struct {
	db           *gorm.DB
	cacheManager *cache.CacheManager
	config       *config.Config
}

// NewConfigService 创建配置服务
func NewConfigService(db *gorm.DB, cacheManager *cache.CacheManager, config *config.Config) *ConfigService {
	return &ConfigService{
		db:           db,
		cacheManager: cacheManager,
		config:       config,
	}
}

// ConfigRequest 配置请求
type ConfigRequest struct {
	Key         string      `json:"key" binding:"required"`
	Value       interface{} `json:"value" binding:"required"`
	Description string      `json:"description"`
	Category    string      `json:"category" binding:"required"`
	IsSecret    bool        `json:"is_secret"`
}

// ConfigResponse 配置响应
type ConfigResponse struct {
	ID          uuid.UUID   `json:"id"`
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	IsSecret    bool        `json:"is_secret"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// DatabaseConfigRequest 数据库配置请求
type DatabaseConfigRequest struct {
	Host         string `json:"host" binding:"required"`
	Port         int    `json:"port" binding:"required"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Database     string `json:"database" binding:"required"`
	SSLMode      string `json:"ssl_mode"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxLifetime  int    `json:"max_lifetime"`
}

// RedisConfigRequest Redis配置请求
type RedisConfigRequest struct {
	Host         string `json:"host" binding:"required"`
	Port         int    `json:"port" binding:"required"`
	Password     string `json:"password"`
	Database     int    `json:"database"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
	MaxRetries   int    `json:"max_retries"`
	DialTimeout  int    `json:"dial_timeout"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// AIModelConfigRequest AI模型配置请求
type AIModelConfigRequest struct {
	Provider    string `json:"provider" binding:"required,oneof=openai claude"`
	APIKey      string `json:"api_key" binding:"required"`
	BaseURL     string `json:"base_url"`
	Model       string `json:"model" binding:"required"`
	MaxTokens   int    `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Timeout     int    `json:"timeout"`
}

// EmailConfigRequest 邮件配置请求
type EmailConfigRequest struct {
	SMTPHost     string `json:"smtp_host" binding:"required"`
	SMTPPort     int    `json:"smtp_port" binding:"required"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	FromEmail    string `json:"from_email" binding:"required,email"`
	FromName     string `json:"from_name"`
	UseTLS       bool   `json:"use_tls"`
	UseSSL       bool   `json:"use_ssl"`
	AuthType     string `json:"auth_type"`
	PoolSize     int    `json:"pool_size"`
	KeepAlive    int    `json:"keep_alive"`
}

// PrometheusConfigRequest Prometheus配置请求
type PrometheusConfigRequest struct {
	URL             string `json:"url" binding:"required"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Timeout         int    `json:"timeout"`
	RetentionPeriod string `json:"retention_period"`
	ScrapeInterval  string `json:"scrape_interval"`
	QueryTimeout    string `json:"query_timeout"`
}

// SystemConfigRequest 系统配置请求
type SystemConfigRequest struct {
	SystemName        string `json:"system_name"`
	SystemDescription string `json:"system_description"`
	SystemVersion     string `json:"system_version"`
	MaintenanceMode   bool   `json:"maintenance_mode"`
	DebugMode         bool   `json:"debug_mode"`
	LogLevel          string `json:"log_level"`
	MaxFileSize       int64  `json:"max_file_size"`
	SessionTimeout    int    `json:"session_timeout"`
	PasswordPolicy    string `json:"password_policy"`
	LoginAttempts     int    `json:"login_attempts"`
	LockoutDuration   int    `json:"lockout_duration"`
}

// CreateConfig 创建配置
func (s *ConfigService) CreateConfig(req *ConfigRequest) (*ConfigResponse, error) {
	// 检查配置键是否已存在
	var existingConfig models.SystemConfig
	if err := s.db.Where("key = ?", req.Key).First(&existingConfig).Error; err == nil {
		return nil, errors.New("config key already exists")
	}

	// 序列化值
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config value: %w", err)
	}

	// 创建配置
	config := models.SystemConfig{
		Key:         req.Key,
		Value:       string(valueJSON),
		Description: req.Description,
		Category:    req.Category,
		IsSecret:    req.IsSecret,
	}

	if err := s.db.Create(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	// 清除缓存
	s.clearConfigCache(req.Key)

	return s.toConfigResponse(&config), nil
}

// GetConfig 获取配置
func (s *ConfigService) GetConfig(key string) (*ConfigResponse, error) {
	// 检查缓存
	cacheKey := cache.ConfigCacheKey(key)
	if s.cacheManager != nil {
		ctx := context.Background()
		var response ConfigResponse
		if err := s.cacheManager.Get(ctx, cacheKey, &response); err == nil {
			return &response, nil
		}
	}

	var config models.SystemConfig
	if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("config not found")
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	response := s.toConfigResponse(&config)

	// 缓存结果
	if s.cacheManager != nil {
		ctx := context.Background()
		if data, err := json.Marshal(response); err == nil {
			s.cacheManager.Set(ctx, cacheKey, string(data), 10*time.Minute)
		}
	}

	return response, nil
}

// UpdateConfig 更新配置
func (s *ConfigService) UpdateConfig(key string, req *ConfigRequest) (*ConfigResponse, error) {
	var config models.SystemConfig
	if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("config not found")
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// 序列化新值
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config value: %w", err)
	}

	// 更新配置
	updates := map[string]interface{}{
		"value":       string(valueJSON),
		"description": req.Description,
		"category":    req.Category,
		"is_secret":   req.IsSecret,
	}

	if err := s.db.Model(&config).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update config: %w", err)
	}

	// 重新加载数据
	if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to reload config: %w", err)
	}

	// 清除缓存
	s.clearConfigCache(key)

	return s.toConfigResponse(&config), nil
}

// DeleteConfig 删除配置
func (s *ConfigService) DeleteConfig(key string) error {
	var config models.SystemConfig
	if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("config not found")
		}
		return fmt.Errorf("failed to get config: %w", err)
	}

	// 软删除配置
	if err := s.db.Delete(&config).Error; err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	// 清除缓存
	s.clearConfigCache(key)

	return nil
}

// ListConfigs 获取配置列表
func (s *ConfigService) ListConfigs(page, pageSize int, category string, isPublic *bool) ([]*ConfigResponse, int64, error) {
	query := s.db.Model(&models.SystemConfig{})

	// 分类过滤
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 公开状态过滤
	if isPublic != nil {
		query = query.Where("is_secret = ?", !*isPublic)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count configs: %w", err)
	}

	// 分页查询
	var configs []models.SystemConfig
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("category, key").Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list configs: %w", err)
	}

	// 转换为响应格式
	responses := make([]*ConfigResponse, len(configs))
	for i, config := range configs {
		responses[i] = s.toConfigResponse(&config)
	}

	return responses, total, nil
}

// UpdateDatabaseConfig 更新数据库配置
func (s *ConfigService) UpdateDatabaseConfig(req *DatabaseConfigRequest) error {
	// 验证数据库连接
	if err := s.validateDatabaseConnection(req); err != nil {
		return fmt.Errorf("database connection validation failed: %w", err)
	}

	// 更新配置
	configs := map[string]interface{}{
		"database.host":           req.Host,
		"database.port":           req.Port,
		"database.username":       req.Username,
		"database.password":       req.Password,
		"database.database":       req.Database,
		"database.ssl_mode":       req.SSLMode,
		"database.max_open_conns": req.MaxOpenConns,
		"database.max_idle_conns": req.MaxIdleConns,
		"database.max_lifetime":   req.MaxLifetime,
	}

	return s.updateMultipleConfigs(configs, "database")
}

// UpdateRedisConfig 更新Redis配置
func (s *ConfigService) UpdateRedisConfig(req *RedisConfigRequest) error {
	// 验证Redis连接
	if err := s.validateRedisConnection(req); err != nil {
		return fmt.Errorf("redis connection validation failed: %w", err)
	}

	// 更新配置
	configs := map[string]interface{}{
		"redis.host":           req.Host,
		"redis.port":           req.Port,
		"redis.password":       req.Password,
		"redis.database":       req.Database,
		"redis.pool_size":      req.PoolSize,
		"redis.min_idle_conns": req.MinIdleConns,
		"redis.max_retries":    req.MaxRetries,
		"redis.dial_timeout":   req.DialTimeout,
		"redis.read_timeout":   req.ReadTimeout,
		"redis.write_timeout":  req.WriteTimeout,
	}

	return s.updateMultipleConfigs(configs, "redis")
}

// UpdateAIModelConfig 更新AI模型配置
func (s *ConfigService) UpdateAIModelConfig(req *AIModelConfigRequest) error {
	// 验证AI模型连接
	if err := s.validateAIModelConnection(req); err != nil {
		return fmt.Errorf("ai model connection validation failed: %w", err)
	}

	// 更新配置
	configs := map[string]interface{}{
		"ai." + req.Provider + ".api_key":     req.APIKey,
		"ai." + req.Provider + ".base_url":    req.BaseURL,
		"ai." + req.Provider + ".model":       req.Model,
		"ai." + req.Provider + ".max_tokens":  req.MaxTokens,
		"ai." + req.Provider + ".temperature": req.Temperature,
		"ai." + req.Provider + ".timeout":     req.Timeout,
	}

	return s.updateMultipleConfigs(configs, "ai")
}

// UpdateEmailConfig 更新邮件配置
func (s *ConfigService) UpdateEmailConfig(req *EmailConfigRequest) error {
	// 验证邮件配置
	if err := s.validateEmailConnection(req); err != nil {
		return fmt.Errorf("email connection validation failed: %w", err)
	}

	// 更新配置
	configs := map[string]interface{}{
		"email.smtp_host":  req.SMTPHost,
		"email.smtp_port":  req.SMTPPort,
		"email.username":   req.Username,
		"email.password":   req.Password,
		"email.from_email": req.FromEmail,
		"email.from_name":  req.FromName,
		"email.use_tls":    req.UseTLS,
		"email.use_ssl":    req.UseSSL,
		"email.auth_type":  req.AuthType,
		"email.pool_size":  req.PoolSize,
		"email.keep_alive": req.KeepAlive,
	}

	return s.updateMultipleConfigs(configs, "email")
}

// UpdatePrometheusConfig 更新Prometheus配置
func (s *ConfigService) UpdatePrometheusConfig(req *PrometheusConfigRequest) error {
	// 验证Prometheus连接
	if err := s.validatePrometheusConnection(req); err != nil {
		return fmt.Errorf("prometheus connection validation failed: %w", err)
	}

	// 更新配置
	configs := map[string]interface{}{
		"prometheus.url":              req.URL,
		"prometheus.username":         req.Username,
		"prometheus.password":         req.Password,
		"prometheus.timeout":          req.Timeout,
		"prometheus.retention_period": req.RetentionPeriod,
		"prometheus.scrape_interval":  req.ScrapeInterval,
		"prometheus.query_timeout":    req.QueryTimeout,
	}

	return s.updateMultipleConfigs(configs, "prometheus")
}

// UpdateSystemConfig 更新系统配置
func (s *ConfigService) UpdateSystemConfig(req *SystemConfigRequest) error {
	// 更新配置
	configs := map[string]interface{}{
		"system.name":             req.SystemName,
		"system.description":      req.SystemDescription,
		"system.version":          req.SystemVersion,
		"system.maintenance_mode": req.MaintenanceMode,
		"system.debug_mode":       req.DebugMode,
		"system.log_level":        req.LogLevel,
		"system.max_file_size":    req.MaxFileSize,
		"system.session_timeout":  req.SessionTimeout,
		"system.password_policy":  req.PasswordPolicy,
		"system.login_attempts":   req.LoginAttempts,
		"system.lockout_duration": req.LockoutDuration,
	}

	return s.updateMultipleConfigs(configs, "system")
}

// GetConfigsByCategory 按分类获取配置
func (s *ConfigService) GetConfigsByCategory(category string) (map[string]interface{}, error) {
	// 检查缓存
	cacheKey := cache.ConfigCategoryCacheKey(category)
	if s.cacheManager != nil {
		ctx := context.Background()
		var result map[string]interface{}
		if err := s.cacheManager.Get(ctx, cacheKey, &result); err == nil {
			return result, nil
		}
	}

	var configs []models.SystemConfig
	if err := s.db.Where("category = ?", category).Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to get configs by category: %w", err)
	}

	result := make(map[string]interface{})
	for _, config := range configs {
		var value interface{}
		valueStr := config.Value
		
		if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
			value = valueStr // 如果解析失败，使用原始字符串
		}
		result[config.Key] = value
	}

	// 缓存结果
	if s.cacheManager != nil {
		ctx := context.Background()
		if data, err := json.Marshal(result); err == nil {
			s.cacheManager.Set(ctx, cacheKey, string(data), 10*time.Minute)
		}
	}

	return result, nil
}

// TestDatabaseConnection 测试数据库连接
func (s *ConfigService) TestDatabaseConnection(req *DatabaseConfigRequest) error {
	return s.validateDatabaseConnection(req)
}

// TestRedisConnection 测试Redis连接
func (s *ConfigService) TestRedisConnection(req *RedisConfigRequest) error {
	return s.validateRedisConnection(req)
}

// TestAIModelConnection 测试AI模型连接
func (s *ConfigService) TestAIModelConnection(req *AIModelConfigRequest) error {
	return s.validateAIModelConnection(req)
}

// TestEmailConnection 测试邮件连接
func (s *ConfigService) TestEmailConnection(req *EmailConfigRequest) error {
	return s.validateEmailConnection(req)
}

// TestPrometheusConnection 测试Prometheus连接
func (s *ConfigService) TestPrometheusConnection(req *PrometheusConfigRequest) error {
	return s.validatePrometheusConnection(req)
}

// validateDatabaseConnection 验证数据库连接
func (s *ConfigService) validateDatabaseConnection(req *DatabaseConfigRequest) error {
	// TODO: 实现数据库连接验证逻辑
	// 这里应该尝试连接到指定的数据库
	return nil
}

// validateRedisConnection 验证Redis连接
func (s *ConfigService) validateRedisConnection(req *RedisConfigRequest) error {
	// TODO: 实现Redis连接验证逻辑
	// 这里应该尝试连接到指定的Redis
	return nil
}

// validateAIModelConnection 验证AI模型连接
func (s *ConfigService) validateAIModelConnection(req *AIModelConfigRequest) error {
	// TODO: 实现AI模型连接验证逻辑
	// 这里应该尝试调用指定的AI模型API
	return nil
}

// validateEmailConnection 验证邮件连接
func (s *ConfigService) validateEmailConnection(req *EmailConfigRequest) error {
	// TODO: 实现邮件连接验证逻辑
	// 这里应该尝试连接到指定的SMTP服务器
	return nil
}

// validatePrometheusConnection 验证Prometheus连接
func (s *ConfigService) validatePrometheusConnection(req *PrometheusConfigRequest) error {
	// TODO: 实现Prometheus连接验证逻辑
	// 这里应该尝试连接到指定的Prometheus服务器
	return nil
}

// updateMultipleConfigs 批量更新配置
func (s *ConfigService) updateMultipleConfigs(configs map[string]interface{}, category string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for key, value := range configs {
		// 序列化值
		valueJSON, err := json.Marshal(value)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to marshal config value for key %s: %w", key, err)
		}

		// 更新或创建配置
		var config models.SystemConfig
		if err := tx.Where("key = ?", key).First(&config).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建新配置
				config = models.SystemConfig{
					Key:         key,
					Value:       string(valueJSON),
					Category:    category,
					IsSecret:    false,
				}
				if err := tx.Create(&config).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create config for key %s: %w", key, err)
				}
			} else {
				tx.Rollback()
				return fmt.Errorf("failed to get config for key %s: %w", key, err)
			}
		} else {
			// 更新现有配置
			if err := tx.Model(&config).Update("value", string(valueJSON)).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update config for key %s: %w", key, err)
			}
		}

		// 清除缓存
		s.clearConfigCache(key)
	}

	// 清除分类缓存
	s.clearConfigCategoryCache(category)

	return tx.Commit().Error
}

// clearConfigCache 清除配置缓存
func (s *ConfigService) clearConfigCache(key string) {
	if s.cacheManager != nil {
		ctx := context.Background()
		cacheKey := cache.ConfigCacheKey(key)
		s.cacheManager.Delete(ctx, cacheKey)
	}
}

// clearConfigCategoryCache 清除分类配置缓存
func (s *ConfigService) clearConfigCategoryCache(category string) {
	if s.cacheManager != nil {
		ctx := context.Background()
		cacheKey := cache.ConfigCategoryCacheKey(category)
		s.cacheManager.Delete(ctx, cacheKey)
	}
}

// toConfigResponse 转换为配置响应格式
func (s *ConfigService) toConfigResponse(config *models.SystemConfig) *ConfigResponse {
	var value interface{}
	valueStr := config.Value
	
	if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
		value = valueStr // 如果解析失败，使用原始字符串
	}

	return &ConfigResponse{
		ID:          config.ID,
		Key:         config.Key,
		Value:       value,
		Description: config.Description,
		Category:    config.Category,
		IsSecret:    config.IsSecret,
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
	}
}

// GetConfigStats 获取配置统计信息
func (s *ConfigService) GetConfigStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 配置总数
	var totalConfigs int64
	if err := s.db.Model(&models.SystemConfig{}).Count(&totalConfigs).Error; err != nil {
		return nil, fmt.Errorf("failed to count total configs: %w", err)
	}
	stats["total_configs"] = totalConfigs

	// 按分类统计
	categoryStats := make(map[string]int64)
	categories := []string{"system", "database", "redis", "ai", "email", "prometheus", "alert", "monitoring"}
	for _, category := range categories {
		var count int64
		if err := s.db.Model(&models.SystemConfig{}).Where("category = ?", category).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count %s configs: %w", category, err)
		}
		categoryStats[category] = count
	}
	stats["category_stats"] = categoryStats

	// 公开配置数
	var publicConfigs int64
	if err := s.db.Model(&models.SystemConfig{}).Where("is_secret = ?", false).Count(&publicConfigs).Error; err != nil {
		return nil, fmt.Errorf("failed to count public configs: %w", err)
	}
	stats["public_configs"] = publicConfigs

	// 私密配置数
	var secretConfigs int64
	if err := s.db.Model(&models.SystemConfig{}).Where("is_secret = ?", true).Count(&secretConfigs).Error; err != nil {
		return nil, fmt.Errorf("failed to count secret configs: %w", err)
	}
	stats["secret_configs"] = secretConfigs

	return stats, nil
}