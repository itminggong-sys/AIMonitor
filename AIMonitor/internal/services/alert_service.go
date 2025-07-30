package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ai-monitor/internal/cache"
	"ai-monitor/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AlertService 告警服务
type AlertService struct {
	db           *gorm.DB
	cacheManager *cache.CacheManager
	notifyService *NotificationService
	aiService    *AIService
}

// NewAlertService 创建告警服务
func NewAlertService(db *gorm.DB, cacheManager *cache.CacheManager, notifyService *NotificationService, aiService *AIService) *AlertService {
	return &AlertService{
		db:           db,
		cacheManager: cacheManager,
		notifyService: notifyService,
		aiService:    aiService,
	}
}

// CreateAlertRuleRequest 创建告警规则请求
type CreateAlertRuleRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	TargetType  string                 `json:"target_type" binding:"required,oneof=host service application"`
	TargetID    string                 `json:"target_id"`
	MetricName  string                 `json:"metric_name" binding:"required"`
	Condition   string                 `json:"condition" binding:"required,oneof=> >= < <= == !="`
	Threshold   float64                `json:"threshold" binding:"required"`
	Duration    int                    `json:"duration" binding:"required,min=1"`
	Severity    string                 `json:"severity" binding:"required,oneof=critical high medium low"`
	Enabled     bool                   `json:"enabled"`
	Tags        map[string]interface{} `json:"tags"`
	Channels    []string               `json:"channels"`
}

// UpdateAlertRuleRequest 更新告警规则请求
type UpdateAlertRuleRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Condition   string                 `json:"condition" binding:"omitempty,oneof=> >= < <= == !="`
	Threshold   *float64               `json:"threshold"`
	Duration    *int                   `json:"duration" binding:"omitempty,min=1"`
	Severity    string                 `json:"severity" binding:"omitempty,oneof=critical high medium low"`
	Enabled     *bool                  `json:"enabled"`
	Tags        map[string]interface{} `json:"tags"`
	Channels    []string               `json:"channels"`
}

// AlertRuleResponse 告警规则响应
type AlertRuleResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TargetType  string                 `json:"target_type"`
	TargetID    string                 `json:"target_id"`
	MetricName  string                 `json:"metric_name"`
	Condition   string                 `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Duration    int                    `json:"duration"`
	Severity    string                 `json:"severity"`
	Enabled     bool                   `json:"enabled"`
	Tags        map[string]interface{} `json:"tags"`
	Channels    []string               `json:"channels"`
	CreatedBy   uuid.UUID              `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AlertResponse 告警响应
type AlertResponse struct {
	ID          uuid.UUID              `json:"id"`
	RuleID      uuid.UUID              `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	TargetType  string                 `json:"target_type"`
	TargetID    string                 `json:"target_id"`
	MetricName  string                 `json:"metric_name"`
	CurrentValue float64               `json:"current_value"`
	Threshold   float64                `json:"threshold"`
	Condition   string                 `json:"condition"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Tags        map[string]interface{} `json:"tags"`
	StartedAt   time.Time              `json:"started_at"`
	ResolvedAt  *time.Time             `json:"resolved_at"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// MetricData 指标数据
type MetricData struct {
	TargetType string                 `json:"target_type"`
	TargetID   string                 `json:"target_id"`
	MetricName string                 `json:"metric_name"`
	Value      float64                `json:"value"`
	Tags       map[string]interface{} `json:"tags"`
	Timestamp  time.Time              `json:"timestamp"`
}

// CreateAlertRule 创建告警规则
func (s *AlertService) CreateAlertRule(req *CreateAlertRuleRequest, createdBy uuid.UUID) (*AlertRuleResponse, error) {
	// 检查规则名称是否已存在
	var existingRule models.AlertRule
	if err := s.db.Where("name = ?", req.Name).First(&existingRule).Error; err == nil {
		return nil, errors.New("alert rule name already exists")
	}

	// 序列化标签
	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	// 序列化通知渠道
	channelsJSON, err := json.Marshal(req.Channels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal channels: %w", err)
	}

	// 创建告警规则
	rule := models.AlertRule{
		Name:        req.Name,
		Description: req.Description,
		Metric:      req.MetricName, // 使用Metric字段而不是MetricName
		Condition:   req.Condition,
		Threshold:   req.Threshold,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Enabled:     req.Enabled,
		// 注意：AlertRule模型中没有Tags和Channels字段，可以使用Labels和Annotations
		Labels:      string(tagsJSON),
		Annotations: string(channelsJSON),
		CreatedBy:   createdBy,
	}

	if err := s.db.Create(&rule).Error; err != nil {
		return nil, fmt.Errorf("failed to create alert rule: %w", err)
	}

	// 清除缓存
	s.clearAlertRuleCache()

	return s.toAlertRuleResponse(&rule), nil
}

// GetAlertRule 获取告警规则
func (s *AlertService) GetAlertRule(ruleID uuid.UUID) (*AlertRuleResponse, error) {
	var rule models.AlertRule
	if err := s.db.First(&rule, ruleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("alert rule not found")
		}
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	return s.toAlertRuleResponse(&rule), nil
}

// UpdateAlertRule 更新告警规则
func (s *AlertService) UpdateAlertRule(ruleID uuid.UUID, req *UpdateAlertRuleRequest) (*AlertRuleResponse, error) {
	var rule models.AlertRule
	if err := s.db.First(&rule, ruleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("alert rule not found")
		}
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	// 更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		// 检查名称是否已存在
		var existingRule models.AlertRule
		if err := s.db.Where("name = ? AND id != ?", req.Name, ruleID).First(&existingRule).Error; err == nil {
			return nil, errors.New("alert rule name already exists")
		}
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Condition != "" {
		updates["condition"] = req.Condition
	}
	if req.Threshold != nil {
		updates["threshold"] = *req.Threshold
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.Severity != "" {
		updates["severity"] = req.Severity
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Tags != nil {
		tagsJSON, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		updates["tags"] = string(tagsJSON)
	}
	if req.Channels != nil {
		channelsJSON, err := json.Marshal(req.Channels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal channels: %w", err)
		}
		updates["channels"] = string(channelsJSON)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&rule).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update alert rule: %w", err)
		}
	}

	// 重新加载数据
	if err := s.db.First(&rule, ruleID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload alert rule: %w", err)
	}

	// 清除缓存
	s.clearAlertRuleCache()

	return s.toAlertRuleResponse(&rule), nil
}

// DeleteAlertRule 删除告警规则
func (s *AlertService) DeleteAlertRule(ruleID uuid.UUID) error {
	var rule models.AlertRule
	if err := s.db.First(&rule, ruleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("alert rule not found")
		}
		return fmt.Errorf("failed to get alert rule: %w", err)
	}

	// 软删除规则
	if err := s.db.Delete(&rule).Error; err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	// 清除缓存
	s.clearAlertRuleCache()

	return nil
}

// ListAlertRules 获取告警规则列表
func (s *AlertService) ListAlertRules(page, pageSize int, search string, enabled *bool) ([]*AlertRuleResponse, int64, error) {
	query := s.db.Model(&models.AlertRule{})

	// 搜索条件
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 启用状态过滤
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count alert rules: %w", err)
	}

	// 分页查询
	var rules []models.AlertRule
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&rules).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list alert rules: %w", err)
	}

	// 转换为响应格式
	responses := make([]*AlertRuleResponse, len(rules))
	for i, rule := range rules {
		responses[i] = s.toAlertRuleResponse(&rule)
	}

	return responses, total, nil
}

// ProcessMetricData 处理指标数据并检查告警
func (s *AlertService) ProcessMetricData(data *MetricData) error {
	// 获取相关的告警规则
	rules, err := s.getActiveAlertRules(data.TargetType, data.TargetID, data.MetricName)
	if err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	// 检查每个规则
	for _, rule := range rules {
		if err := s.checkAlertRule(rule, data); err != nil {
			// 记录错误但不中断处理
			// Error checking alert rule
		}
	}

	return nil
}

// checkAlertRule 检查告警规则
func (s *AlertService) checkAlertRule(rule *models.AlertRule, data *MetricData) error {
	// 评估条件
	triggered := s.evaluateCondition(rule.Condition, data.Value, rule.Threshold)

	if triggered {
		// 检查是否已存在活跃告警
		var existingAlert models.Alert
		err := s.db.Where("rule_id = ? AND target_id = ? AND status = ?", rule.ID, data.TargetID, "firing").First(&existingAlert).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check existing alert: %w", err)
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新告警
			return s.createAlert(rule, data)
		} else {
			// 更新现有告警
			return s.updateAlert(&existingAlert, data)
		}
	} else {
		// 检查是否需要解决告警
		return s.resolveAlert(rule.ID, data.TargetID)
	}
}

// evaluateCondition 评估告警条件
func (s *AlertService) evaluateCondition(condition string, currentValue, threshold float64) bool {
	switch condition {
	case ">":
		return currentValue > threshold
	case ">=":
		return currentValue >= threshold
	case "<":
		return currentValue < threshold
	case "<=":
		return currentValue <= threshold
	case "==":
		return currentValue == threshold
	case "!=":
		return currentValue != threshold
	default:
		return false
	}
}

// createAlert 创建告警
func (s *AlertService) createAlert(rule *models.AlertRule, data *MetricData) error {
	// 序列化标签
	tagsJSON, err := json.Marshal(data.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	// 生成告警消息
	message := fmt.Sprintf("%s %s %s %.2f (current: %.2f)",
		data.MetricName, rule.Condition, strconv.FormatFloat(rule.Threshold, 'f', 2, 64), rule.Threshold, data.Value)

	// 创建告警
	alert := models.Alert{
		RuleID:      rule.ID,
		Fingerprint: fmt.Sprintf("%s-%s-%s", rule.ID, data.TargetID, data.MetricName), // 生成指纹
		Severity:    rule.Severity,
		Status:      "firing",
		Summary:     message,
		Description: fmt.Sprintf("Alert for metric %s on target %s", data.MetricName, data.TargetID),
		Value:       data.Value,
		Labels:      string(tagsJSON),
		StartsAt:    data.Timestamp,
	}

	if err := s.db.Create(&alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// 发送通知
	go s.sendAlertNotification(&alert, rule)

	// 触发AI分析
	go s.triggerAIAnalysis(&alert, rule, data)

	return nil
}

// updateAlert 更新告警
func (s *AlertService) updateAlert(alert *models.Alert, data *MetricData) error {
	// 更新当前值和时间戳
	updates := map[string]interface{}{
		"current_value": data.Value,
		"updated_at":    time.Now(),
	}

	// 更新标签
	if data.Tags != nil {
		tagsJSON, err := json.Marshal(data.Tags)
		if err != nil {
			return fmt.Errorf("failed to marshal tags: %w", err)
		}
		updates["tags"] = string(tagsJSON)
	}

	return s.db.Model(alert).Updates(updates).Error
}

// resolveAlert 解决告警
func (s *AlertService) resolveAlert(ruleID uuid.UUID, targetID string) error {
	var alert models.Alert
	err := s.db.Where("rule_id = ? AND target_id = ? AND status = ?", ruleID, targetID, "firing").First(&alert).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 没有活跃告警
		}
		return fmt.Errorf("failed to find alert: %w", err)
	}

	// 更新告警状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":      "resolved",
		"resolved_at": &now,
		"updated_at":  now,
	}

	if err := s.db.Model(&alert).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	// 发送解决通知
	go s.sendResolvedNotification(&alert)

	return nil
}

// sendAlertNotification 发送告警通知
func (s *AlertService) sendAlertNotification(alert *models.Alert, rule *models.AlertRule) {
	if s.notifyService == nil {
		return
	}

	// 解析通知渠道
	var channels []string
	if err := json.Unmarshal([]byte(rule.Annotations), &channels); err != nil {
		// Failed to unmarshal channels
		return
	}

	// 构建通知内容
	notification := &NotificationRequest{
		Title:    fmt.Sprintf("[%s] %s", strings.ToUpper(alert.Severity), rule.Name),
		Content:  alert.Summary,
		Severity: alert.Severity,
		Tags: map[string]interface{}{
			"alert_id":    alert.ID,
			"rule_id":     rule.ID,
			"fingerprint": alert.Fingerprint,
			"status":      alert.Status,
		},
		Channels: channels,
	}

	// 发送通知
	if err := s.notifyService.SendNotification(notification); err != nil {
		// Failed to send alert notification
	}
}

// sendResolvedNotification 发送解决通知
func (s *AlertService) sendResolvedNotification(alert *models.Alert) {
	if s.notifyService == nil {
		return
	}

	// 获取规则信息
	var rule models.AlertRule
	if err := s.db.First(&rule, alert.RuleID).Error; err != nil {
		// Failed to get alert rule
		return
	}

	// 解析通知渠道
	var channels []string
	if err := json.Unmarshal([]byte(rule.Annotations), &channels); err != nil {
		// Failed to unmarshal channels
		return
	}

	// 构建通知内容
	notification := &NotificationRequest{
		Title:    fmt.Sprintf("[RESOLVED] %s", rule.Name),
		Content:  fmt.Sprintf("Alert has been resolved: %s", alert.Summary),
		Severity: "info",
		Tags: map[string]interface{}{
			"alert_id":    alert.ID,
			"rule_id":     rule.ID,
			"fingerprint": alert.Fingerprint,
			"status":      alert.Status,
			"resolved":    true,
		},
		Channels: channels,
	}

	// 发送通知
	if err := s.notifyService.SendNotification(notification); err != nil {
		// Failed to send resolved notification
	}
}

// triggerAIAnalysis 触发AI分析
func (s *AlertService) triggerAIAnalysis(alert *models.Alert, rule *models.AlertRule, data *MetricData) {
	if s.aiService == nil {
		return
	}

	// 构建分析请求
	request := &AIAnalysisRequest{
		Type:        "alert_analysis",
		AlertID:     alert.ID,
		RuleID:      rule.ID,
		TargetType:  data.TargetType,
		TargetID:    data.TargetID,
		MetricName:  data.MetricName,
		CurrentValue: data.Value,
		Threshold:   rule.Threshold,
		Condition:   rule.Condition,
		Severity:    rule.Severity,
		Tags:        data.Tags,
		Timestamp:   data.Timestamp,
	}

	// 执行AI分析
	if _, err := s.aiService.AnalyzeAlert(request); err != nil {
		// Failed to trigger AI analysis
	}
}

// getActiveAlertRules 获取活跃的告警规则
func (s *AlertService) getActiveAlertRules(targetType, targetID, metricName string) ([]*models.AlertRule, error) {
	// 尝试从缓存获取
	cacheKey := cache.AlertRulesCacheKey(targetType, targetID, metricName)
	if s.cacheManager != nil {
		var rules []*models.AlertRule
		if err := s.cacheManager.Get(context.Background(), cacheKey, &rules); err == nil {
			return rules, nil
		}
	}

	// 从数据库查询
	var rules []models.AlertRule
	query := s.db.Where("enabled = ? AND metric = ?", true, metricName)

	if err := query.Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}

	// 转换为指针切片
	rulePointers := make([]*models.AlertRule, len(rules))
	for i := range rules {
		rulePointers[i] = &rules[i]
	}

	// 缓存结果
	if s.cacheManager != nil {
		if data, err := json.Marshal(rulePointers); err == nil {
			s.cacheManager.Set(context.Background(), cacheKey, string(data), 5*time.Minute)
		}
	}

	return rulePointers, nil
}

// GetAlert 获取告警
func (s *AlertService) GetAlert(alertID uuid.UUID) (*AlertResponse, error) {
	var alert models.Alert
	if err := s.db.First(&alert, alertID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("alert not found")
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return s.toAlertResponse(&alert), nil
}

// ListAlerts 获取告警列表
func (s *AlertService) ListAlerts(page, pageSize int, status, severity, targetType string) ([]*AlertResponse, int64, error) {
	query := s.db.Model(&models.Alert{})

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 严重级别过滤
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	// 注意：Alert模型中没有target_type字段，此过滤条件被注释
	// if targetType != "" {
	//	query = query.Where("target_type = ?", targetType)
	// }

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	// 分页查询
	var alerts []models.Alert
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list alerts: %w", err)
	}

	// 转换为响应格式
	responses := make([]*AlertResponse, len(alerts))
	for i, alert := range alerts {
		responses[i] = s.toAlertResponse(&alert)
	}

	return responses, total, nil
}

// AcknowledgeAlert 确认告警
func (s *AlertService) AcknowledgeAlert(alertID uuid.UUID, userID uuid.UUID) (*AlertResponse, error) {
	var alert models.Alert
	if err := s.db.First(&alert, alertID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("alert not found")
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	if alert.Status != "firing" {
		return nil, errors.New("only firing alerts can be acknowledged")
	}

	// 更新告警状态
	now := time.Now()
	updates := map[string]interface{}{
		"acknowledged":    true,
		"acknowledged_by": userID,
		"acknowledged_at": &now,
		"updated_at":      now,
	}

	if err := s.db.Model(&alert).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	// 重新获取更新后的告警
	if err := s.db.First(&alert, alertID).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated alert: %w", err)
	}

	return s.toAlertResponse(&alert), nil
}

// ResolveAlert 解决告警
func (s *AlertService) ResolveAlert(alertID uuid.UUID, userID uuid.UUID) (*AlertResponse, error) {
	var alert models.Alert
	if err := s.db.First(&alert, alertID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("alert not found")
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	if alert.Status == "resolved" {
		return nil, errors.New("alert is already resolved")
	}

	// 更新告警状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":     "resolved",
		"ends_at":    &now,
		"updated_at": now,
	}

	if err := s.db.Model(&alert).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to resolve alert: %w", err)
	}

	// 发送解决通知
	go s.sendResolvedNotification(&alert)

	// 重新获取更新后的告警
	if err := s.db.First(&alert, alertID).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated alert: %w", err)
	}

	return s.toAlertResponse(&alert), nil
}

// GetAlertStats 获取告警统计
func (s *AlertService) GetAlertStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总告警数
	var totalAlerts int64
	if err := s.db.Model(&models.Alert{}).Count(&totalAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total alerts: %w", err)
	}
	stats["total_alerts"] = totalAlerts

	// 活跃告警数
	var activeAlerts int64
	if err := s.db.Model(&models.Alert{}).Where("status IN ?", []string{"firing", "acknowledged"}).Count(&activeAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count active alerts: %w", err)
	}
	stats["active_alerts"] = activeAlerts

	// 按严重级别统计
	severityStats := make(map[string]int64)
	severities := []string{"critical", "high", "medium", "low"}
	for _, severity := range severities {
		var count int64
		if err := s.db.Model(&models.Alert{}).Where("severity = ? AND status IN ?", severity, []string{"firing", "acknowledged"}).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count %s alerts: %w", severity, err)
		}
		severityStats[severity] = count
	}
	stats["severity_stats"] = severityStats

	// 今日新增告警数
	var todayAlerts int64
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&models.Alert{}).Where("created_at >= ?", today).Count(&todayAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count today alerts: %w", err)
	}
	stats["today_alerts"] = todayAlerts

	return stats, nil
}

// toAlertRuleResponse 转换为告警规则响应格式
func (s *AlertService) toAlertRuleResponse(rule *models.AlertRule) *AlertRuleResponse {
	var tags map[string]interface{}
	var channels []string

	if rule.Labels != "" {
		json.Unmarshal([]byte(rule.Labels), &tags)
	}
	if rule.Annotations != "" {
		json.Unmarshal([]byte(rule.Annotations), &channels)
	}

	return &AlertRuleResponse{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		TargetType:  "", // AlertRule模型中没有此字段
		TargetID:    "", // AlertRule模型中没有此字段
		MetricName:  rule.Metric, // 使用Metric字段
		Condition:   rule.Condition,
		Threshold:   rule.Threshold,
		Duration:    rule.Duration,
		Severity:    rule.Severity,
		Enabled:     rule.Enabled,
		Tags:        tags,
		Channels:    channels,
		CreatedBy:   rule.CreatedBy,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}

// toAlertResponse 转换为告警响应格式
func (s *AlertService) toAlertResponse(alert *models.Alert) *AlertResponse {
	var tags map[string]interface{}
	if alert.Labels != "" {
		json.Unmarshal([]byte(alert.Labels), &tags)
	}

	// 获取规则名称
	var rule models.AlertRule
	ruleName := "Unknown"
	if err := s.db.Select("name").First(&rule, alert.RuleID).Error; err == nil {
		ruleName = rule.Name
	}

	return &AlertResponse{
		ID:           alert.ID,
		RuleID:       alert.RuleID,
		RuleName:     ruleName,
		TargetType:   "", // Alert模型中没有此字段
		TargetID:     "", // Alert模型中没有此字段
		MetricName:   "", // Alert模型中没有此字段
		CurrentValue: alert.Value, // 使用Value字段
		Threshold:    0, // Alert模型中没有此字段
		Condition:    "", // Alert模型中没有此字段
		Severity:     alert.Severity,
		Status:       alert.Status,
		Message:      alert.Summary, // 使用Summary字段
		Tags:         tags,
		StartedAt:    alert.StartsAt, // 使用StartsAt字段
		ResolvedAt:   alert.EndsAt, // 使用EndsAt字段
		CreatedAt:    alert.CreatedAt,
		UpdatedAt:    alert.UpdatedAt,
	}
}

// clearAlertRuleCache 清除告警规则缓存
func (s *AlertService) clearAlertRuleCache() {
	if s.cacheManager != nil {
		// 清除所有告警规则相关的缓存
		pattern := "alert_rules:*"
		keys, err := s.cacheManager.Keys(context.Background(), pattern)
		if err == nil && len(keys) > 0 {
			s.cacheManager.Delete(context.Background(), keys...)
		}
	}
}