package services

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"
	"ai-monitor/internal/models"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

// NotificationService 通知服务
type NotificationService struct {
	db           *gorm.DB
	cacheManager *cache.CacheManager
	config       *config.Config
}

// NewNotificationService 创建通知服务
func NewNotificationService(db *gorm.DB, cacheManager *cache.CacheManager, config *config.Config) *NotificationService {
	return &NotificationService{
		db:           db,
		cacheManager: cacheManager,
		config:       config,
	}
}

// NotificationRequest 通知请求
type NotificationRequest struct {
	Title    string                 `json:"title" binding:"required"`
	Content  string                 `json:"content" binding:"required"`
	Severity string                 `json:"severity" binding:"required,oneof=critical high medium low info"`
	Tags     map[string]interface{} `json:"tags"`
	Channels []string               `json:"channels" binding:"required"`
}

// CreateChannelRequest 创建通知渠道请求
type CreateChannelRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Type        string                 `json:"type" binding:"required,oneof=email sms webhook slack dingtalk"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config" binding:"required"`
	Enabled     bool                   `json:"enabled"`
}

// UpdateChannelRequest 更新通知渠道请求
type UpdateChannelRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Enabled     *bool                  `json:"enabled"`
}

// ChannelResponse 通知渠道响应
type ChannelResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID        uuid.UUID              `json:"id"`
	ChannelID uuid.UUID              `json:"channel_id"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Severity  string                 `json:"severity"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
	Tags      map[string]interface{} `json:"tags"`
	SentAt    *time.Time             `json:"sent_at"`
	CreatedAt time.Time              `json:"created_at"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string   `json:"smtp_host"`
	SMTPPort     int      `json:"smtp_port"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	FromAddress  string   `json:"from_address"`
	FromName     string   `json:"from_name"`
	ToAddresses  []string `json:"to_addresses"`
	CCAddresses  []string `json:"cc_addresses,omitempty"`
	BCCAddresses []string `json:"bcc_addresses,omitempty"`
	UseTLS       bool     `json:"use_tls"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Timeout int               `json:"timeout"`
}

// SlackConfig Slack配置
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username,omitempty"`
	IconEmoji  string `json:"icon_emoji,omitempty"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret,omitempty"`
	AtMobiles  []string `json:"at_mobiles,omitempty"`
	AtAll      bool   `json:"at_all,omitempty"`
}

// CreateChannel 创建通知渠道
func (s *NotificationService) CreateChannel(req *CreateChannelRequest) (*ChannelResponse, error) {
	// 检查渠道名称是否已存在
	var existingChannel models.NotificationChannel
	if err := s.db.Where("name = ?", req.Name).First(&existingChannel).Error; err == nil {
		return nil, errors.New("notification channel name already exists")
	}

	// 验证配置
	if err := s.validateChannelConfig(req.Type, req.Config); err != nil {
		return nil, fmt.Errorf("invalid channel config: %w", err)
	}

	// 序列化配置
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// 创建通知渠道
	channel := models.NotificationChannel{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Config:      string(configJSON),
		Enabled:     req.Enabled,
	}

	if err := s.db.Create(&channel).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification channel: %w", err)
	}

	return s.toChannelResponse(&channel), nil
}

// GetChannel 获取通知渠道
func (s *NotificationService) GetChannel(channelID uuid.UUID) (*ChannelResponse, error) {
	var channel models.NotificationChannel
	if err := s.db.First(&channel, channelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification channel not found")
		}
		return nil, fmt.Errorf("failed to get notification channel: %w", err)
	}

	return s.toChannelResponse(&channel), nil
}

// UpdateChannel 更新通知渠道
func (s *NotificationService) UpdateChannel(channelID uuid.UUID, req *UpdateChannelRequest) (*ChannelResponse, error) {
	var channel models.NotificationChannel
	if err := s.db.First(&channel, channelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification channel not found")
		}
		return nil, fmt.Errorf("failed to get notification channel: %w", err)
	}

	// 更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		// 检查名称是否已存在
		var existingChannel models.NotificationChannel
		if err := s.db.Where("name = ? AND id != ?", req.Name, channelID).First(&existingChannel).Error; err == nil {
			return nil, errors.New("notification channel name already exists")
		}
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Config != nil {
		// 验证配置
		if err := s.validateChannelConfig(channel.Type, req.Config); err != nil {
			return nil, fmt.Errorf("invalid channel config: %w", err)
		}
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		updates["config"] = string(configJSON)
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := s.db.Model(&channel).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update notification channel: %w", err)
		}
	}

	// 重新加载数据
	if err := s.db.First(&channel, channelID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload notification channel: %w", err)
	}

	return s.toChannelResponse(&channel), nil
}

// DeleteChannel 删除通知渠道
func (s *NotificationService) DeleteChannel(channelID uuid.UUID) error {
	var channel models.NotificationChannel
	if err := s.db.First(&channel, channelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("notification channel not found")
		}
		return fmt.Errorf("failed to get notification channel: %w", err)
	}

	// 软删除渠道
	if err := s.db.Delete(&channel).Error; err != nil {
		return fmt.Errorf("failed to delete notification channel: %w", err)
	}

	return nil
}

// ListChannels 获取通知渠道列表
func (s *NotificationService) ListChannels(page, pageSize int, channelType string, enabled *bool) ([]*ChannelResponse, int64, error) {
	query := s.db.Model(&models.NotificationChannel{})

	// 类型过滤
	if channelType != "" {
		query = query.Where("type = ?", channelType)
	}

	// 启用状态过滤
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notification channels: %w", err)
	}

	// 分页查询
	var channels []models.NotificationChannel
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&channels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list notification channels: %w", err)
	}

	// 转换为响应格式
	responses := make([]*ChannelResponse, len(channels))
	for i, channel := range channels {
		responses[i] = s.toChannelResponse(&channel)
	}

	return responses, total, nil
}

// SendNotification 发送通知
func (s *NotificationService) SendNotification(req *NotificationRequest) error {
	// 获取通知渠道
	channels, err := s.getChannelsByNames(req.Channels)
	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	// 并发发送通知
	errorChan := make(chan error, len(channels))
	for _, channel := range channels {
		go func(ch *models.NotificationChannel) {
			err := s.sendToChannel(ch, req)
			errorChan <- err
		}(channel)
	}

	// 收集错误
	var errors []string
	for i := 0; i < len(channels); i++ {
		if err := <-errorChan; err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// sendToChannel 发送到指定渠道
func (s *NotificationService) sendToChannel(channel *models.NotificationChannel, req *NotificationRequest) error {
	if !channel.Enabled {
		return fmt.Errorf("channel %s is disabled", channel.Name)
	}

	// 创建通知记录
	notification := models.AlertNotification{
		Channel:   channel.Type,
		Recipient: channel.Name,
		Status:    "pending",
	}

	// 注意：AlertNotification 模型中没有 Title, Content, Severity, Tags 字段
	// 这些信息应该存储在其他地方或者需要修改模型定义

	if err := s.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	// 根据渠道类型发送通知
	var err error
	switch channel.Type {
	case "email":
		err = s.sendEmail(channel, req)
	case "webhook":
		err = s.sendWebhook(channel, req)
	case "slack":
		err = s.sendSlack(channel, req)
	case "dingtalk":
		err = s.sendDingTalk(channel, req)
	default:
		err = fmt.Errorf("unsupported channel type: %s", channel.Type)
	}

	// 更新通知状态
	now := time.Now()
	updates := map[string]interface{}{
		"updated_at": now,
	}

	if err != nil {
		updates["status"] = "failed"
		updates["error"] = err.Error()
	} else {
		updates["status"] = "sent"
		updates["sent_at"] = &now
	}

	s.db.Model(&notification).Updates(updates)

	return err
}

// sendEmail 发送邮件
func (s *NotificationService) sendEmail(channel *models.NotificationChannel, req *NotificationRequest) error {
	var config EmailConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("failed to parse email config: %w", err)
	}

	// 创建邮件
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(config.FromAddress, config.FromName))
	m.SetHeader("To", config.ToAddresses...)
	if len(config.CCAddresses) > 0 {
		m.SetHeader("Cc", config.CCAddresses...)
	}
	if len(config.BCCAddresses) > 0 {
		m.SetHeader("Bcc", config.BCCAddresses...)
	}
	m.SetHeader("Subject", req.Title)

	// 生成邮件内容
	htmlContent, err := s.generateEmailHTML(req)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %w", err)
	}

	m.SetBody("text/html", htmlContent)
	m.AddAlternative("text/plain", req.Content)

	// 创建SMTP拨号器
	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.Username, config.Password)
	if config.UseTLS {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendWebhook 发送Webhook
func (s *NotificationService) sendWebhook(channel *models.NotificationChannel, req *NotificationRequest) error {
	var config WebhookConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("failed to parse webhook config: %w", err)
	}

	// 构建请求体
	payload := map[string]interface{}{
		"title":     req.Title,
		"content":   req.Content,
		"severity":  req.Severity,
		"tags":      req.Tags,
		"timestamp": time.Now().Unix(),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// 创建HTTP请求
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	req_http, err := http.NewRequest(config.Method, config.URL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// 设置请求头
	req_http.Header.Set("Content-Type", "application/json")
	for key, value := range config.Headers {
		req_http.Header.Set(key, value)
	}

	// 发送请求
	resp, err := client.Do(req_http)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status code: %d", resp.StatusCode)
	}

	return nil
}

// sendSlack 发送Slack通知
func (s *NotificationService) sendSlack(channel *models.NotificationChannel, req *NotificationRequest) error {
	var config SlackConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("failed to parse slack config: %w", err)
	}

	// 构建Slack消息
	payload := map[string]interface{}{
		"text":    req.Title,
		"channel": config.Channel,
	}

	if config.Username != "" {
		payload["username"] = config.Username
	}
	if config.IconEmoji != "" {
		payload["icon_emoji"] = config.IconEmoji
	}

	// 添加附件
	color := s.getSeverityColor(req.Severity)
	attachment := map[string]interface{}{
		"color": color,
		"text":  req.Content,
		"ts":    time.Now().Unix(),
	}

	if req.Tags != nil {
		fields := []map[string]interface{}{}
		for key, value := range req.Tags {
			fields = append(fields, map[string]interface{}{
				"title": key,
				"value": fmt.Sprintf("%v", value),
				"short": true,
			})
		}
		attachment["fields"] = fields
	}

	payload["attachments"] = []map[string]interface{}{attachment}

	// 发送请求
	return s.sendWebhookPayload(config.WebhookURL, payload)
}

// sendDingTalk 发送钉钉通知
func (s *NotificationService) sendDingTalk(channel *models.NotificationChannel, req *NotificationRequest) error {
	var config DingTalkConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("failed to parse dingtalk config: %w", err)
	}

	// 构建钉钉消息
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": fmt.Sprintf("%s\n\n%s", req.Title, req.Content),
		},
	}

	// 添加@功能
	if len(config.AtMobiles) > 0 || config.AtAll {
		at := map[string]interface{}{}
		if len(config.AtMobiles) > 0 {
			at["atMobiles"] = config.AtMobiles
		}
		if config.AtAll {
			at["isAtAll"] = true
		}
		payload["at"] = at
	}

	// 发送请求
	return s.sendWebhookPayload(config.WebhookURL, payload)
}

// sendWebhookPayload 发送Webhook载荷
func (s *NotificationService) sendWebhookPayload(url string, payload map[string]interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status code: %d", resp.StatusCode)
	}

	return nil
}

// TestChannel 测试通知渠道
func (s *NotificationService) TestChannel(channelID uuid.UUID) error {
	channel, err := s.GetChannel(channelID)
	if err != nil {
		return err
	}

	// 构建测试通知
	testReq := &NotificationRequest{
		Title:    "Test Notification",
		Content:  "This is a test notification from AI Monitor System.",
		Severity: "info",
		Tags: map[string]interface{}{
			"test": true,
			"time": time.Now().Format(time.RFC3339),
		},
		Channels: []string{channel.Name},
	}

	return s.SendNotification(testReq)
}

// GetNotificationHistory 获取通知历史
func (s *NotificationService) GetNotificationHistory(page, pageSize int, channelID *uuid.UUID, status string) ([]*NotificationResponse, int64, error) {
	query := s.db.Model(&models.AlertNotification{})

	// 渠道过滤
	if channelID != nil {
		query = query.Where("channel_id = ?", *channelID)
	}

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// 分页查询
	var notifications []models.AlertNotification
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}

	// 转换为响应格式
	responses := make([]*NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = s.toNotificationResponse(&notification)
	}

	return responses, total, nil
}

// validateChannelConfig 验证渠道配置
func (s *NotificationService) validateChannelConfig(channelType string, config map[string]interface{}) error {
	switch channelType {
	case "email":
		return s.validateEmailConfig(config)
	case "webhook":
		return s.validateWebhookConfig(config)
	case "slack":
		return s.validateSlackConfig(config)
	case "dingtalk":
		return s.validateDingTalkConfig(config)
	default:
		return fmt.Errorf("unsupported channel type: %s", channelType)
	}
}

// validateEmailConfig 验证邮件配置
func (s *NotificationService) validateEmailConfig(config map[string]interface{}) error {
	requiredFields := []string{"smtp_host", "smtp_port", "username", "password", "from_address", "to_addresses"}
	for _, field := range requiredFields {
		if _, exists := config[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

// validateWebhookConfig 验证Webhook配置
func (s *NotificationService) validateWebhookConfig(config map[string]interface{}) error {
	requiredFields := []string{"url", "method"}
	for _, field := range requiredFields {
		if _, exists := config[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

// validateSlackConfig 验证Slack配置
func (s *NotificationService) validateSlackConfig(config map[string]interface{}) error {
	requiredFields := []string{"webhook_url"}
	for _, field := range requiredFields {
		if _, exists := config[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

// validateDingTalkConfig 验证钉钉配置
func (s *NotificationService) validateDingTalkConfig(config map[string]interface{}) error {
	requiredFields := []string{"webhook_url"}
	for _, field := range requiredFields {
		if _, exists := config[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

// getChannelsByNames 根据名称获取通知渠道
func (s *NotificationService) getChannelsByNames(names []string) ([]*models.NotificationChannel, error) {
	var channels []*models.NotificationChannel
	if err := s.db.Where("name IN ? AND enabled = ?", names, true).Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	if len(channels) == 0 {
		return nil, errors.New("no enabled channels found")
	}

	return channels, nil
}

// generateEmailHTML 生成邮件HTML内容
func (s *NotificationService) generateEmailHTML(req *NotificationRequest) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { background-color: {{.HeaderColor}}; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .severity { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; color: white; background-color: {{.SeverityColor}}; }
        .tags { margin-top: 15px; }
        .tag { display: inline-block; background-color: #e9ecef; padding: 2px 6px; border-radius: 3px; font-size: 11px; margin: 2px; }
        .footer { background-color: #f8f9fa; padding: 15px; text-align: center; font-size: 12px; color: #6c757d; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
            <span class="severity">{{.Severity}}</span>
        </div>
        <div class="content">
            <p>{{.Content}}</p>
            {{if .Tags}}
            <div class="tags">
                <strong>Tags:</strong><br>
                {{range $key, $value := .Tags}}
                <span class="tag">{{$key}}: {{$value}}</span>
                {{end}}
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>AI Monitor System - {{.Timestamp}}</p>
        </div>
    </div>
</body>
</html>
`

	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"Title":         req.Title,
		"Content":       req.Content,
		"Severity":      strings.ToUpper(req.Severity),
		"Tags":          req.Tags,
		"Timestamp":     time.Now().Format("2006-01-02 15:04:05"),
		"HeaderColor":   s.getSeverityColor(req.Severity),
		"SeverityColor": s.getSeverityColor(req.Severity),
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// getSeverityColor 获取严重级别颜色
func (s *NotificationService) getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "#dc3545"
	case "high":
		return "#fd7e14"
	case "medium":
		return "#ffc107"
	case "low":
		return "#28a745"
	case "info":
		return "#17a2b8"
	default:
		return "#6c757d"
	}
}

// toChannelResponse 转换为通知渠道响应格式
func (s *NotificationService) toChannelResponse(channel *models.NotificationChannel) *ChannelResponse {
	var config map[string]interface{}
	if channel.Config != "" {
		json.Unmarshal([]byte(channel.Config), &config)
	}

	// 隐藏敏感信息
	if config != nil {
		if password, exists := config["password"]; exists && password != "" {
			config["password"] = "******"
		}
		if secret, exists := config["secret"]; exists && secret != "" {
			config["secret"] = "******"
		}
	}

	return &ChannelResponse{
		ID:          channel.ID,
		Name:        channel.Name,
		Type:        channel.Type,
		Description: channel.Description,
		Config:      config,
		Enabled:     channel.Enabled,
		CreatedAt:   channel.CreatedAt,
		UpdatedAt:   channel.UpdatedAt,
	}
}

// toNotificationResponse 转换为通知响应格式
func (s *NotificationService) toNotificationResponse(notification *models.AlertNotification) *NotificationResponse {
	// 注意：AlertNotification 模型中缺少一些字段，需要从其他地方获取或使用默认值
	return &NotificationResponse{
		ID:        notification.ID,
		ChannelID: uuid.Nil, // AlertNotification 模型中没有 ChannelID 字段
		Title:     "",       // AlertNotification 模型中没有 Title 字段
		Content:   "",       // AlertNotification 模型中没有 Content 字段
		Severity:  "",       // AlertNotification 模型中没有 Severity 字段
		Status:    notification.Status,
		Error:     notification.Error,
		Tags:      nil,      // AlertNotification 模型中没有 Tags 字段
		SentAt:    notification.SentAt,
		CreatedAt: notification.CreatedAt,
	}
}