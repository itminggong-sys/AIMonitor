package handlers

import (
	"net/http"

	"ai-monitor/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	configService *services.ConfigService
	auditService  *services.AuditService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(configService *services.ConfigService, auditService *services.AuditService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
		auditService:  auditService,
	}
}

// AlertConfigRequest 告警配置请求
type AlertConfigRequest struct {
	// 邮件告警配置
	EmailEnabled    bool   `json:"email_enabled"`
	SMTPHost        string `json:"smtp_host"`
	SMTPPort        int    `json:"smtp_port"`
	SMTPUsername    string `json:"smtp_username"`
	SMTPPassword    string `json:"smtp_password"`
	SMTPFromEmail   string `json:"smtp_from_email"`
	SMTPToEmails    string `json:"smtp_to_emails"` // 逗号分隔的邮箱列表
	SMTPUseTLS      bool   `json:"smtp_use_tls"`
	
	// 短信告警配置
	SMSEnabled      bool   `json:"sms_enabled"`
	SMSProvider     string `json:"sms_provider"` // aliyun, tencent, huawei, baidu
	SMSAccessKeyID  string `json:"sms_access_key_id"`
	SMSAccessKeySecret string `json:"sms_access_key_secret"`
	SMSSignName     string `json:"sms_sign_name"`
	SMSTemplateCode string `json:"sms_template_code"`
	SMSPhoneNumbers string `json:"sms_phone_numbers"` // 逗号分隔的手机号列表
}

// AlertConfigResponse 告警配置响应
type AlertConfigResponse struct {
	// 邮件告警配置
	EmailEnabled    bool   `json:"email_enabled"`
	SMTPHost        string `json:"smtp_host"`
	SMTPPort        int    `json:"smtp_port"`
	SMTPUsername    string `json:"smtp_username"`
	SMTPPassword    string `json:"smtp_password,omitempty"` // 敏感信息可选返回
	SMTPFromEmail   string `json:"smtp_from_email"`
	SMTPToEmails    string `json:"smtp_to_emails"`
	SMTPUseTLS      bool   `json:"smtp_use_tls"`
	
	// 短信告警配置
	SMSEnabled      bool   `json:"sms_enabled"`
	SMSProvider     string `json:"sms_provider"`
	SMSAccessKeyID  string `json:"sms_access_key_id"`
	SMSAccessKeySecret string `json:"sms_access_key_secret,omitempty"` // 敏感信息可选返回
	SMSSignName     string `json:"sms_sign_name"`
	SMSTemplateCode string `json:"sms_template_code"`
	SMSPhoneNumbers string `json:"sms_phone_numbers"`
}

// SystemSettingsRequest 系统设置请求
type SystemSettingsRequest struct {
	SystemName      string `json:"system_name"`
	RefreshInterval int    `json:"refresh_interval"`
	Timezone        string `json:"timezone"`
	Language        string `json:"language"`
	Theme           string `json:"theme"`
	DebugMode       bool   `json:"debug_mode"`
	LogLevel        string `json:"log_level"`
	MaxConcurrentRequests int `json:"max_concurrent_requests"`
	CustomConfig    string `json:"custom_config"` // JSON格式的自定义配置
}

// SystemSettingsResponse 系统设置响应
type SystemSettingsResponse struct {
	SystemName      string `json:"system_name"`
	RefreshInterval int    `json:"refresh_interval"`
	Timezone        string `json:"timezone"`
	Language        string `json:"language"`
	Theme           string `json:"theme"`
	DebugMode       bool   `json:"debug_mode"`
	LogLevel        string `json:"log_level"`
	MaxConcurrentRequests int `json:"max_concurrent_requests"`
	CustomConfig    string `json:"custom_config"`
}

// AIServiceConfigRequest AI服务配置请求
type AIServiceConfigRequest struct {
	Provider     string  `json:"provider"`     // openai, azure, anthropic, ollama
	APIKey       string  `json:"api_key"`
	BaseURL      string  `json:"base_url"`
	Model        string  `json:"model"`
	MaxTokens    int     `json:"max_tokens"`
	Temperature  float64 `json:"temperature"`
	StreamOutput bool    `json:"stream_output"`
	// Ollama特有配置
	OllamaServerURL string `json:"ollama_server_url"`
	OllamaTimeout   int    `json:"ollama_timeout"`
}

// AIServiceConfigResponse AI服务配置响应
type AIServiceConfigResponse struct {
	Provider     string  `json:"provider"`
	APIKey       string  `json:"api_key,omitempty"` // 敏感信息可选返回
	BaseURL      string  `json:"base_url"`
	Model        string  `json:"model"`
	MaxTokens    int     `json:"max_tokens"`
	Temperature  float64 `json:"temperature"`
	StreamOutput bool    `json:"stream_output"`
	OllamaServerURL string `json:"ollama_server_url"`
	OllamaTimeout   int    `json:"ollama_timeout"`
}

// GetAlertConfig 获取告警配置
// @Summary 获取告警配置
// @Description 获取邮件和短信告警配置
// @Tags Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} AlertConfigResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/config/alert [get]
func (h *ConfigHandler) GetAlertConfig(c *gin.Context) {
	// 获取邮件配置
	emailEnabled, _ := h.getConfigBool("alert.email.enabled", false)
	smtpHost, _ := h.getConfigString("alert.email.smtp_host", "")
	smtpPort, _ := h.getConfigInt("alert.email.smtp_port", 587)
	smtpUsername, _ := h.getConfigString("alert.email.smtp_username", "")
	_, _ = h.getConfigString("alert.email.smtp_password", "")
	smtpFromEmail, _ := h.getConfigString("alert.email.smtp_from_email", "")
	smtpToEmails, _ := h.getConfigString("alert.email.smtp_to_emails", "")
	smtpUseTLS, _ := h.getConfigBool("alert.email.smtp_use_tls", true)

	// 获取短信配置
	smsEnabled, _ := h.getConfigBool("alert.sms.enabled", false)
	smsProvider, _ := h.getConfigString("alert.sms.provider", "aliyun")
	smsAccessKeyID, _ := h.getConfigString("alert.sms.access_key_id", "")
	_, _ = h.getConfigString("alert.sms.access_key_secret", "")
	smsSignName, _ := h.getConfigString("alert.sms.sign_name", "")
	smsTemplateCode, _ := h.getConfigString("alert.sms.template_code", "")
	smsPhoneNumbers, _ := h.getConfigString("alert.sms.phone_numbers", "")

	response := AlertConfigResponse{
		EmailEnabled:    emailEnabled,
		SMTPHost:        smtpHost,
		SMTPPort:        smtpPort,
		SMTPUsername:    smtpUsername,
		SMTPFromEmail:   smtpFromEmail,
		SMTPToEmails:    smtpToEmails,
		SMTPUseTLS:      smtpUseTLS,
		SMSEnabled:      smsEnabled,
		SMSProvider:     smsProvider,
		SMSAccessKeyID:  smsAccessKeyID,
		SMSSignName:     smsSignName,
		SMSTemplateCode: smsTemplateCode,
		SMSPhoneNumbers: smsPhoneNumbers,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取告警配置成功",
		"data": response,
	})
}

// UpdateAlertConfig 更新告警配置
// @Summary 更新告警配置
// @Description 更新邮件和短信告警配置
// @Tags Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body AlertConfigRequest true "告警配置"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/config/alert [put]
func (h *ConfigHandler) UpdateAlertConfig(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req AlertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 更新邮件配置
	configs := map[string]interface{}{
		"alert.email.enabled":         req.EmailEnabled,
		"alert.email.smtp_host":       req.SMTPHost,
		"alert.email.smtp_port":       req.SMTPPort,
		"alert.email.smtp_username":   req.SMTPUsername,
		"alert.email.smtp_password":   req.SMTPPassword,
		"alert.email.smtp_from_email": req.SMTPFromEmail,
		"alert.email.smtp_to_emails":  req.SMTPToEmails,
		"alert.email.smtp_use_tls":    req.SMTPUseTLS,
		"alert.sms.enabled":           req.SMSEnabled,
		"alert.sms.provider":          req.SMSProvider,
		"alert.sms.access_key_id":     req.SMSAccessKeyID,
		"alert.sms.access_key_secret": req.SMSAccessKeySecret,
		"alert.sms.sign_name":         req.SMSSignName,
		"alert.sms.template_code":     req.SMSTemplateCode,
		"alert.sms.phone_numbers":     req.SMSPhoneNumbers,
	}

	// 批量更新配置
	for key, value := range configs {
		if err := h.updateConfigValue(key, value, "alert", userID); err != nil {
			h.auditService.LogAuditFromContext(c, "update_alert_config", "config", key, "failure", err.Error(), map[string]interface{}{
				"config_key": key,
			})
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update config",
				Message: err.Error(),
			})
			return
		}
	}

	h.auditService.LogAuditFromContext(c, "update_alert_config", "config", "alert", "success", "", map[string]interface{}{
		"config_count": len(configs),
	})

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "告警配置更新成功",
	})
}

// TestEmailConfig 测试邮件配置
// @Summary 测试邮件配置
// @Description 发送测试邮件验证配置
// @Tags Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/config/alert/test-email [post]
func (h *ConfigHandler) TestEmailConfig(c *gin.Context) {
	// TODO: 实现邮件测试功能
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "测试邮件发送成功",
	})
}

// TestSMSConfig 测试短信配置
// @Summary 测试短信配置
// @Description 发送测试短信验证配置
// @Tags Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/config/alert/test-sms [post]
func (h *ConfigHandler) TestSMSConfig(c *gin.Context) {
	// TODO: 实现短信测试功能
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "测试短信发送成功",
	})
}

// GetSystemSettings 获取系统设置
// @Summary 获取系统设置
// @Description 获取系统基础设置
// @Tags Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} SystemSettingsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/config/system-settings [get]
func (h *ConfigHandler) GetSystemSettings(c *gin.Context) {
	systemName, _ := h.getConfigString("system.name", "AI Monitor")
	refreshInterval, _ := h.getConfigInt("system.refresh_interval", 30)
	timezone, _ := h.getConfigString("system.timezone", "Asia/Shanghai")
	language, _ := h.getConfigString("system.language", "zh-CN")
	theme, _ := h.getConfigString("system.theme", "light")
	debugMode, _ := h.getConfigBool("system.debug_mode", false)
	logLevel, _ := h.getConfigString("system.log_level", "info")
	maxConcurrentRequests, _ := h.getConfigInt("system.max_concurrent_requests", 100)
	customConfig, _ := h.getConfigString("system.custom_config", "{}")

	response := SystemSettingsResponse{
		SystemName:      systemName,
		RefreshInterval: refreshInterval,
		Timezone:        timezone,
		Language:        language,
		Theme:           theme,
		DebugMode:       debugMode,
		LogLevel:        logLevel,
		MaxConcurrentRequests: maxConcurrentRequests,
		CustomConfig:    customConfig,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取系统设置成功",
		"data": response,
	})
}

// UpdateSystemSettings 更新系统设置
// @Summary 更新系统设置
// @Description 更新系统基础设置
// @Tags Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body SystemSettingsRequest true "系统设置"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/config/system-settings [put]
func (h *ConfigHandler) UpdateSystemSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req SystemSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 更新系统配置
	configs := map[string]interface{}{
		"system.name":                    req.SystemName,
		"system.refresh_interval":        req.RefreshInterval,
		"system.timezone":                req.Timezone,
		"system.language":                req.Language,
		"system.theme":                   req.Theme,
		"system.debug_mode":              req.DebugMode,
		"system.log_level":               req.LogLevel,
		"system.max_concurrent_requests": req.MaxConcurrentRequests,
		"system.custom_config":           req.CustomConfig,
	}

	// 批量更新配置
	for key, value := range configs {
		if err := h.updateConfigValue(key, value, "system", userID); err != nil {
			h.auditService.LogAuditFromContext(c, "update_system_settings", "config", key, "failure", err.Error(), map[string]interface{}{
				"config_key": key,
			})
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update config",
				Message: err.Error(),
			})
			return
		}
	}

	h.auditService.LogAuditFromContext(c, "update_system_settings", "config", "system", "success", "", map[string]interface{}{
		"config_count": len(configs),
	})

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "系统设置更新成功",
	})
}

// GetAIServiceConfig 获取AI服务配置
// @Summary 获取AI服务配置
// @Description 获取AI服务配置信息
// @Tags Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} AIServiceConfigResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/config/ai-service [get]
func (h *ConfigHandler) GetAIServiceConfig(c *gin.Context) {
	provider, _ := h.getConfigString("ai.provider", "openai")
	_, _ = h.getConfigString("ai.api_key", "")
	baseURL, _ := h.getConfigString("ai.base_url", "")
	model, _ := h.getConfigString("ai.model", "gpt-3.5-turbo")
	maxTokens, _ := h.getConfigInt("ai.max_tokens", 2048)
	temperature, _ := h.getConfigFloat("ai.temperature", 0.7)
	streamOutput, _ := h.getConfigBool("ai.stream_output", true)
	ollamaServerURL, _ := h.getConfigString("ai.ollama_server_url", "http://localhost:11434")
	ollamaTimeout, _ := h.getConfigInt("ai.ollama_timeout", 30)

	response := AIServiceConfigResponse{
		Provider:        provider,
		BaseURL:         baseURL,
		Model:           model,
		MaxTokens:       maxTokens,
		Temperature:     temperature,
		StreamOutput:    streamOutput,
		OllamaServerURL: ollamaServerURL,
		OllamaTimeout:   ollamaTimeout,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取AI服务配置成功",
		"data": response,
	})
}

// UpdateAIServiceConfig 更新AI服务配置
// @Summary 更新AI服务配置
// @Description 更新AI服务配置信息
// @Tags Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body AIServiceConfigRequest true "AI服务配置"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/config/ai-service [put]
func (h *ConfigHandler) UpdateAIServiceConfig(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req AIServiceConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 更新AI服务配置
	configs := map[string]interface{}{
		"ai.provider":           req.Provider,
		"ai.api_key":            req.APIKey,
		"ai.base_url":           req.BaseURL,
		"ai.model":              req.Model,
		"ai.max_tokens":         req.MaxTokens,
		"ai.temperature":        req.Temperature,
		"ai.stream_output":      req.StreamOutput,
		"ai.ollama_server_url":  req.OllamaServerURL,
		"ai.ollama_timeout":     req.OllamaTimeout,
	}

	// 批量更新配置
	for key, value := range configs {
		if err := h.updateConfigValue(key, value, "ai", userID); err != nil {
			h.auditService.LogAuditFromContext(c, "update_ai_service_config", "config", key, "failure", err.Error(), map[string]interface{}{
				"config_key": key,
			})
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to update config",
				Message: err.Error(),
			})
			return
		}
	}

	h.auditService.LogAuditFromContext(c, "update_ai_service_config", "config", "ai", "success", "", map[string]interface{}{
		"config_count": len(configs),
	})

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "AI服务配置更新成功",
	})
}

// 辅助方法

// getConfigString 获取字符串配置
func (h *ConfigHandler) getConfigString(key, defaultValue string) (string, error) {
	config, err := h.configService.GetConfig(key)
	if err != nil {
		return defaultValue, err
	}
	if str, ok := config.Value.(string); ok {
		return str, nil
	}
	return defaultValue, nil
}

// getConfigInt 获取整数配置
func (h *ConfigHandler) getConfigInt(key string, defaultValue int) (int, error) {
	config, err := h.configService.GetConfig(key)
	if err != nil {
		return defaultValue, err
	}
	if val, ok := config.Value.(float64); ok {
		return int(val), nil
	}
	return defaultValue, nil
}

// getConfigFloat 获取浮点数配置
func (h *ConfigHandler) getConfigFloat(key string, defaultValue float64) (float64, error) {
	config, err := h.configService.GetConfig(key)
	if err != nil {
		return defaultValue, err
	}
	if val, ok := config.Value.(float64); ok {
		return val, nil
	}
	return defaultValue, nil
}

// getConfigBool 获取布尔配置
func (h *ConfigHandler) getConfigBool(key string, defaultValue bool) (bool, error) {
	config, err := h.configService.GetConfig(key)
	if err != nil {
		return defaultValue, err
	}
	if val, ok := config.Value.(bool); ok {
		return val, nil
	}
	return defaultValue, nil
}

// updateConfigValue 更新配置值
func (h *ConfigHandler) updateConfigValue(key string, value interface{}, category string, userID uuid.UUID) error {
	// 检查配置是否存在
	_, err := h.configService.GetConfig(key)
	if err != nil {
		// 配置不存在，创建新配置
		req := &services.ConfigRequest{
			Key:         key,
			Value:       value,
			Category:    category,
			Description: "",
			IsSecret:    false,
		}
		_, err = h.configService.CreateConfig(req)
		return err
	}

	// 配置存在，更新配置
	req := &services.ConfigRequest{
		Key:         key,
		Value:       value,
		Category:    category,
		Description: "",
		IsSecret:    false,
	}
	_, err = h.configService.UpdateConfig(key, req)
	return err
}