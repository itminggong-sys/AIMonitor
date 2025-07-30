package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"
	"ai-monitor/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditService 审计服务
type AuditService struct {
	db           *gorm.DB
	cacheManager *cache.CacheManager
	config       *config.Config
}

// NewAuditService 创建审计服务
func NewAuditService(db *gorm.DB, cacheManager *cache.CacheManager, config *config.Config) *AuditService {
	return &AuditService{
		db:           db,
		cacheManager: cacheManager,
		config:       config,
	}
}

// AuditLogRequest 审计日志请求
type AuditLogRequest struct {
	Action      string                 `json:"action" binding:"required"`
	Resource    string                 `json:"resource" binding:"required"`
	ResourceID  string                 `json:"resource_id"`
	Details     map[string]interface{} `json:"details"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Result      string                 `json:"result" binding:"required,oneof=success failure"`
	ErrorMsg    string                 `json:"error_msg"`
}

// AuditLogResponse 审计日志响应
type AuditLogResponse struct {
	ID         uuid.UUID              `json:"id"`
	UserID     *uuid.UUID             `json:"user_id"`
	Username   string                 `json:"username"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id"`
	Details    map[string]interface{} `json:"details"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Result     string                 `json:"result"`
	ErrorMsg   string                 `json:"error_msg"`
	CreatedAt  time.Time              `json:"created_at"`
}

// AuditLogFilter 审计日志过滤器
type AuditLogFilter struct {
	UserID     *uuid.UUID `json:"user_id"`
	Action     string     `json:"action"`
	Resource   string     `json:"resource"`
	ResourceID string     `json:"resource_id"`
	Result     string     `json:"result"`
	IPAddress  string     `json:"ip_address"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
}

// AuditStats 审计统计
type AuditStats struct {
	TotalLogs       int64                    `json:"total_logs"`
	SuccessLogs     int64                    `json:"success_logs"`
	FailureLogs     int64                    `json:"failure_logs"`
	TodayLogs       int64                    `json:"today_logs"`
	TopActions      []ActionStat             `json:"top_actions"`
	TopResources    []ResourceStat           `json:"top_resources"`
	TopUsers        []UserStat               `json:"top_users"`
	HourlyStats     []HourlyStat             `json:"hourly_stats"`
	DailyStats      []DailyStat              `json:"daily_stats"`
	ErrorAnalysis   []ErrorAnalysis          `json:"error_analysis"`
	SecurityEvents  []SecurityEvent          `json:"security_events"`
}

// ActionStat 操作统计
type ActionStat struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

// ResourceStat 资源统计
type ResourceStat struct {
	Resource string `json:"resource"`
	Count    int64  `json:"count"`
}

// UserStat 用户统计
type UserStat struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Count    int64     `json:"count"`
}

// HourlyStat 小时统计
type HourlyStat struct {
	Hour    int   `json:"hour"`
	Count   int64 `json:"count"`
	Success int64 `json:"success"`
	Failure int64 `json:"failure"`
}

// DailyStat 日统计
type DailyStat struct {
	Date    string `json:"date"`
	Count   int64  `json:"count"`
	Success int64  `json:"success"`
	Failure int64  `json:"failure"`
}

// ErrorAnalysis 错误分析
type ErrorAnalysis struct {
	ErrorMsg string `json:"error_msg"`
	Count    int64  `json:"count"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

// SecurityEvent 安全事件
type SecurityEvent struct {
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Count       int64     `json:"count"`
	LastSeen    time.Time `json:"last_seen"`
}

// LogAudit 记录审计日志
func (s *AuditService) LogAudit(userID *uuid.UUID, req *AuditLogRequest) error {
	// 序列化详细信息
	detailsJSON, _ := json.Marshal(req.Details)

	// 创建审计日志
	auditLog := models.AuditLog{
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Details:    string(detailsJSON),
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		Status:     req.Result,
		Error:      req.ErrorMsg,
	}

	// 设置用户ID
	if userID != nil {
		auditLog.UserID = *userID
	}

	if err := s.db.Create(&auditLog).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	// 异步处理安全事件检测
	go s.detectSecurityEvents(&auditLog)

	return nil
}

// LogAuditFromContext 从Gin上下文记录审计日志
func (s *AuditService) LogAuditFromContext(c *gin.Context, action, resource, resourceID, result, errorMsg string, details map[string]interface{}) error {
	// 获取用户ID
	var userID *uuid.UUID
	if userIDValue, exists := c.Get("user_id"); exists {
		if uid, ok := userIDValue.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// 构建请求
	req := &AuditLogRequest{
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Result:     result,
		ErrorMsg:   errorMsg,
	}

	return s.LogAudit(userID, req)
}

// GetAuditLog 获取审计日志
func (s *AuditService) GetAuditLog(logID uuid.UUID) (*AuditLogResponse, error) {
	var auditLog models.AuditLog
	if err := s.db.Preload("User").First(&auditLog, logID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("audit log not found")
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return s.toAuditLogResponse(&auditLog), nil
}

// ListAuditLogs 获取审计日志列表
func (s *AuditService) ListAuditLogs(page, pageSize int, filter *AuditLogFilter) ([]*AuditLogResponse, int64, error) {
	query := s.db.Model(&models.AuditLog{}).Preload("User")

	// 应用过滤器
	if filter != nil {
		if filter.UserID != nil {
			query = query.Where("user_id = ?", *filter.UserID)
		}
		if filter.Action != "" {
			query = query.Where("action ILIKE ?", "%"+filter.Action+"%")
		}
		if filter.Resource != "" {
			query = query.Where("resource ILIKE ?", "%"+filter.Resource+"%")
		}
		if filter.ResourceID != "" {
			query = query.Where("resource_id = ?", filter.ResourceID)
		}
		if filter.Result != "" {
			query = query.Where("status = ?", filter.Result)
		}
		if filter.IPAddress != "" {
			query = query.Where("ip_address = ?", filter.IPAddress)
		}
		if filter.StartTime != nil {
			query = query.Where("created_at >= ?", *filter.StartTime)
		}
		if filter.EndTime != nil {
			query = query.Where("created_at <= ?", *filter.EndTime)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// 分页查询
	var auditLogs []models.AuditLog
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&auditLogs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	// 转换为响应格式
	responses := make([]*AuditLogResponse, len(auditLogs))
	for i, auditLog := range auditLogs {
		responses[i] = s.toAuditLogResponse(&auditLog)
	}

	return responses, total, nil
}

// DeleteAuditLogs 删除审计日志
func (s *AuditService) DeleteAuditLogs(logIDs []uuid.UUID) error {
	if err := s.db.Where("id IN ?", logIDs).Delete(&models.AuditLog{}).Error; err != nil {
		return fmt.Errorf("failed to delete audit logs: %w", err)
	}
	return nil
}

// CleanupOldLogs 清理旧日志
func (s *AuditService) CleanupOldLogs(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	result := s.db.Where("created_at < ?", cutoffTime).Delete(&models.AuditLog{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old audit logs: %w", result.Error)
	}

	// 记录清理操作
	cleanupDetails := map[string]interface{}{
		"retention_days": retentionDays,
		"cutoff_time":    cutoffTime,
		"deleted_count":  result.RowsAffected,
	}

	cleanupReq := &AuditLogRequest{
		Action:   "cleanup_logs",
		Resource: "audit_log",
		Details:  cleanupDetails,
		Result:   "success",
	}

	return s.LogAudit(nil, cleanupReq)
}

// GetAuditStats 获取审计统计信息
func (s *AuditService) GetAuditStats(days int) (*AuditStats, error) {
	// 检查缓存
	cacheKey := cache.AuditStatsCacheKey(days)
	if s.cacheManager != nil {
		ctx := context.Background()
		var stats AuditStats
		if err := s.cacheManager.Get(ctx, cacheKey, &stats); err == nil {
			return &stats, nil
		}
	}

	stats := &AuditStats{}
	startTime := time.Now().AddDate(0, 0, -days)

	// 总日志数
	if err := s.db.Model(&models.AuditLog{}).Where("created_at >= ?", startTime).Count(&stats.TotalLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to count total logs: %w", err)
	}

	// 成功日志数
	if err := s.db.Model(&models.AuditLog{}).Where("created_at >= ? AND status = ?", startTime, "success").Count(&stats.SuccessLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to count success logs: %w", err)
	}

	// 失败日志数
	if err := s.db.Model(&models.AuditLog{}).Where("created_at >= ? AND status = ?", startTime, "failed").Count(&stats.FailureLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to count failure logs: %w", err)
	}

	// 今日日志数
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&models.AuditLog{}).Where("created_at >= ?", today).Count(&stats.TodayLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to count today logs: %w", err)
	}

	// 获取各种统计数据
	var err error
	stats.TopActions, err = s.getTopActions(startTime, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top actions: %w", err)
	}

	stats.TopResources, err = s.getTopResources(startTime, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top resources: %w", err)
	}

	stats.TopUsers, err = s.getTopUsers(startTime, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top users: %w", err)
	}

	stats.HourlyStats, err = s.getHourlyStats(startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly stats: %w", err)
	}

	stats.DailyStats, err = s.getDailyStats(startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}

	stats.ErrorAnalysis, err = s.getErrorAnalysis(startTime, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get error analysis: %w", err)
	}

	stats.SecurityEvents, err = s.getSecurityEvents(startTime, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}

	// 缓存结果
	if s.cacheManager != nil {
		ctx := context.Background()
		if data, err := json.Marshal(stats); err == nil {
			s.cacheManager.Set(ctx, cacheKey, string(data), 5*time.Minute)
		}
	}

	return stats, nil
}

// ExportAuditLogs 导出审计日志
func (s *AuditService) ExportAuditLogs(filter *AuditLogFilter, format string) ([]byte, error) {
	// 获取所有匹配的日志
	logs, _, err := s.ListAuditLogs(1, 10000, filter) // 限制最大导出数量
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs for export: %w", err)
	}

	switch strings.ToLower(format) {
	case "json":
		return json.MarshalIndent(logs, "", "  ")
	case "csv":
		return s.exportToCSV(logs)
	default:
		return nil, errors.New("unsupported export format")
	}
}

// getTopActions 获取热门操作
func (s *AuditService) getTopActions(startTime time.Time, limit int) ([]ActionStat, error) {
	var results []ActionStat
	err := s.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("created_at >= ?", startTime).
		Group("action").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// getTopResources 获取热门资源
func (s *AuditService) getTopResources(startTime time.Time, limit int) ([]ResourceStat, error) {
	var results []ResourceStat
	err := s.db.Model(&models.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("created_at >= ?", startTime).
		Group("resource").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// getTopUsers 获取活跃用户
func (s *AuditService) getTopUsers(startTime time.Time, limit int) ([]UserStat, error) {
	var results []UserStat
	err := s.db.Table("audit_logs").
		Select("audit_logs.user_id, users.username, COUNT(*) as count").
		Joins("LEFT JOIN users ON audit_logs.user_id = users.id").
		Where("audit_logs.created_at >= ? AND audit_logs.user_id IS NOT NULL", startTime).
		Group("audit_logs.user_id, users.username").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// getHourlyStats 获取小时统计
func (s *AuditService) getHourlyStats(startTime time.Time) ([]HourlyStat, error) {
	var results []HourlyStat
	err := s.db.Model(&models.AuditLog{}).
		Select("EXTRACT(hour FROM created_at) as hour, COUNT(*) as count, SUM(CASE WHEN result = 'success' THEN 1 ELSE 0 END) as success, SUM(CASE WHEN result = 'failure' THEN 1 ELSE 0 END) as failure").
		Where("created_at >= ?", startTime).
		Group("EXTRACT(hour FROM created_at)").
		Order("hour").
		Scan(&results).Error
	return results, err
}

// getDailyStats 获取日统计
func (s *AuditService) getDailyStats(startTime time.Time) ([]DailyStat, error) {
	var results []DailyStat
	err := s.db.Model(&models.AuditLog{}).
		Select("DATE(created_at) as date, COUNT(*) as count, SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success, SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failure").
		Where("created_at >= ?", startTime).
		Group("DATE(created_at)").
		Order("date").
		Scan(&results).Error
	return results, err
}

// getErrorAnalysis 获取错误分析
func (s *AuditService) getErrorAnalysis(startTime time.Time, limit int) ([]ErrorAnalysis, error) {
	var results []ErrorAnalysis
	err := s.db.Model(&models.AuditLog{}).
		Select("error, action, resource, COUNT(*) as count").
		Where("created_at >= ? AND status = 'failed' AND error != ''", startTime).
		Group("error, action, resource").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// getSecurityEvents 获取安全事件
func (s *AuditService) getSecurityEvents(startTime time.Time, limit int) ([]SecurityEvent, error) {
	// 这里可以定义各种安全事件的检测规则
	securityActions := []string{"login_failed", "password_change", "permission_denied", "suspicious_activity"}
	
	var results []SecurityEvent
	err := s.db.Model(&models.AuditLog{}).
		Select("action as event_type, CONCAT('Security event: ', action) as description, ip_address, user_agent, COUNT(*) as count, MAX(created_at) as last_seen").
		Where("created_at >= ? AND action IN ?", startTime, securityActions).
		Group("action, ip_address, user_agent").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// detectSecurityEvents 检测安全事件
func (s *AuditService) detectSecurityEvents(auditLog *models.AuditLog) {
	// 检测多次登录失败
	if auditLog.Action == "login" && auditLog.Status == "failed" {
		s.checkMultipleLoginFailures(auditLog)
	}

	// 检测异常IP访问
	if auditLog.IPAddress != "" {
		s.checkSuspiciousIP(auditLog)
	}

	// 检测权限提升
	if strings.Contains(auditLog.Action, "permission") || strings.Contains(auditLog.Action, "role") {
		s.checkPrivilegeEscalation(auditLog)
	}
}

// checkMultipleLoginFailures 检查多次登录失败
func (s *AuditService) checkMultipleLoginFailures(auditLog *models.AuditLog) {
	// 检查过去15分钟内的登录失败次数
	var count int64
	fifteenMinutesAgo := time.Now().Add(-15 * time.Minute)
	s.db.Model(&models.AuditLog{}).
		Where("action = ? AND status = ? AND ip_address = ? AND created_at >= ?", "login", "failed", auditLog.IPAddress, fifteenMinutesAgo).
		Count(&count)

	if count >= 5 { // 15分钟内失败5次
		// 记录安全事件
		securityReq := &AuditLogRequest{
			Action:   "security_event_multiple_login_failures",
			Resource: "security",
			Details: map[string]interface{}{
				"failure_count": count,
				"ip_address":    auditLog.IPAddress,
				"time_window":   "15 minutes",
			},
			IPAddress: auditLog.IPAddress,
			UserAgent: auditLog.UserAgent,
			Result:    "success",
		}
		s.LogAudit(nil, securityReq)
	}
}

// checkSuspiciousIP 检查可疑IP
func (s *AuditService) checkSuspiciousIP(auditLog *models.AuditLog) {
	// 检查IP是否在黑名单中或来自可疑地区
	// 这里可以集成IP地理位置服务和威胁情报
	// 简化实现：检查是否是新IP
	var count int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	s.db.Model(&models.AuditLog{}).
		Where("ip_address = ? AND created_at >= ?", auditLog.IPAddress, sevenDaysAgo).
		Count(&count)

	if count == 1 { // 新IP首次访问
		securityReq := &AuditLogRequest{
			Action:   "security_event_new_ip_access",
			Resource: "security",
			Details: map[string]interface{}{
				"ip_address": auditLog.IPAddress,
				"user_agent": auditLog.UserAgent,
				"action":     auditLog.Action,
			},
			IPAddress: auditLog.IPAddress,
			UserAgent: auditLog.UserAgent,
			Result:    "success",
		}
		s.LogAudit(&auditLog.UserID, securityReq)
	}
}

// checkPrivilegeEscalation 检查权限提升
func (s *AuditService) checkPrivilegeEscalation(auditLog *models.AuditLog) {
	// UserID 不是指针类型，直接使用

	// 检查用户是否在短时间内进行了权限相关操作
	var count int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	s.db.Model(&models.AuditLog{}).
		Where("user_id = ? AND (action ILIKE '%permission%' OR action ILIKE '%role%') AND created_at >= ?", auditLog.UserID, oneHourAgo).
		Count(&count)

	if count >= 3 { // 1小时内3次权限操作
		securityReq := &AuditLogRequest{
			Action:   "security_event_privilege_escalation",
			Resource: "security",
			Details: map[string]interface{}{
				"user_id":       auditLog.UserID,
				"operation_count": count,
				"time_window":   "1 hour",
				"latest_action": auditLog.Action,
			},
			IPAddress: auditLog.IPAddress,
			UserAgent: auditLog.UserAgent,
			Result:    "success",
		}
		s.LogAudit(&auditLog.UserID, securityReq)
	}
}

// exportToCSV 导出为CSV格式
func (s *AuditService) exportToCSV(logs []*AuditLogResponse) ([]byte, error) {
	// 简化的CSV导出实现
	var csv strings.Builder
	csv.WriteString("ID,UserID,Username,Action,Resource,ResourceID,IPAddress,Result,ErrorMsg,CreatedAt\n")
	
	for _, log := range logs {
		userID := ""
		if log.UserID != nil {
			userID = log.UserID.String()
		}
		csv.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			log.ID.String(),
			userID,
			log.Username,
			log.Action,
			log.Resource,
			log.ResourceID,
			log.IPAddress,
			log.Result,
			log.ErrorMsg,
			log.CreatedAt.Format(time.RFC3339),
		))
	}
	
	return []byte(csv.String()), nil
}

// toAuditLogResponse 转换为审计日志响应格式
func (s *AuditService) toAuditLogResponse(auditLog *models.AuditLog) *AuditLogResponse {
	var details map[string]interface{}
	if auditLog.Details != "" {
		json.Unmarshal([]byte(auditLog.Details), &details)
	}

	username := ""
	if auditLog.User.ID != uuid.Nil {
		username = auditLog.User.Username
	}

	return &AuditLogResponse{
		ID:         auditLog.ID,
		UserID:     &auditLog.UserID,
		Username:   username,
		Action:     auditLog.Action,
		Resource:   auditLog.Resource,
		ResourceID: auditLog.ResourceID,
		Details:    details,
		IPAddress:  auditLog.IPAddress,
		UserAgent:  auditLog.UserAgent,
		Result:     auditLog.Status,
		ErrorMsg:   auditLog.Error,
		CreatedAt:  auditLog.CreatedAt,
	}
}