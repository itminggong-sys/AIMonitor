package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ai-monitor/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ErrorResponse 错误响应结构体
type ErrorResponse struct {
	Error   string `json:"error" example:"Bad Request"`
	Message string `json:"message" example:"Invalid request parameters"`
}

// PaginatedResponse 分页响应结构体
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page     int `json:"page" example:"1"`
	PageSize int `json:"page_size" example:"20"`
	Total    int `json:"total" example:"100"`
	Pages    int `json:"pages" example:"5"`
}

// Handlers 处理器集合
type Handlers struct {
	userService       *services.UserService
	alertService      *services.AlertService
	notificationService *services.NotificationService
	aiService         *services.AIService
	monitoringService *services.MonitoringService
	configService     *services.ConfigService
	auditService      *services.AuditService
	// 新增服务
	middlewareService *services.MiddlewareService
	apmService        *services.APMService
	containerService  *services.ContainerService
	agentService      *services.AgentService
	apiKeyService     *services.APIKeyService
	// 新增处理器
	middlewareHandler *MiddlewareHandler
	apmHandler        *APMHandler
	containerHandler  *ContainerHandler
	agentHandler      *AgentHandler
	configHandler     *ConfigHandler
	apiKeyHandler     *APIKeyHandler
	discoveryHandler  *DiscoveryHandler
	// 添加Services字段以便访问所有服务
	Services          *services.Services
}

// NewHandlers 创建处理器
func NewHandlers(services *services.Services) *Handlers {
	// 创建新增的处理器
	middlewareHandler := NewMiddlewareHandler(services.MiddlewareService)
	apmHandler := NewAPMHandler(services.APMService)
	containerHandler := NewContainerHandler(services.ContainerService)
	agentHandler := NewAgentHandler(services.AgentService)
	configHandler := NewConfigHandler(services.ConfigService, services.AuditService)
	apiKeyHandler := NewAPIKeyHandler(services.APIKeyService)
	discoveryHandler := NewDiscoveryHandler(services.DiscoveryService)

	return &Handlers{
		userService:         services.UserService,
		alertService:        services.AlertService,
		notificationService: services.NotificationService,
		aiService:           services.AIService,
		monitoringService:   services.MonitoringService,
		configService:       services.ConfigService,
		auditService:        services.AuditService,
		// 新增服务
		middlewareService: services.MiddlewareService,
		apmService:        services.APMService,
		containerService:  services.ContainerService,
		agentService:      services.AgentService,
		apiKeyService:     services.APIKeyService,
		// 新增处理器
		middlewareHandler: middlewareHandler,
		apmHandler:        apmHandler,
		containerHandler:  containerHandler,
		agentHandler:      agentHandler,
		configHandler:     configHandler,
		apiKeyHandler:     apiKeyHandler,
		discoveryHandler:  discoveryHandler,
		// 添加Services字段
		Services:          services,
	}
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务健康状态
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

// GetVersion 获取版本信息
// @Summary 获取版本信息
// @Description 获取系统版本信息
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /version [get]
func (h *Handlers) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":     "1.0.0",
		"build_time":  "2024-01-01T00:00:00Z",
		"git_commit": "unknown",
		"go_version":  "1.21",
	})
}

// ===== 用户管理相关处理器 =====

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户
// @Tags User
// @Accept json
// @Produce json
// @Param request body services.CreateUserRequest true "注册信息"
// @Success 201 {object} services.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *Handlers) Register(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "register", "user", "", "failure", err.Error(), map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "register", "user", user.ID.String(), "success", "", map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	})

	c.JSON(http.StatusCreated, user)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取令牌
// @Tags User
// @Accept json
// @Produce json
// @Param request body services.LoginRequest true "登录信息"
// @Success 200 {object} services.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "请求参数错误",
			"error": err.Error(),
		})
		return
	}

	response, err := h.userService.Login(&req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "login", "user", "", "failure", err.Error(), map[string]interface{}{
			"username": req.Username,
		})
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"message": err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "login", "user", response.User.ID.String(), "success", "", map[string]interface{}{
		"username": response.User.Username,
	})

	// 构造符合前端期望的响应格式
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "登录成功",
		"data": gin.H{
			"token": response.TokenInfo.AccessToken,
			"refreshToken": response.TokenInfo.RefreshToken,
			"user": response.User,
		},
	})
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags User
// @Accept json
// @Produce json
// @Param request body services.RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} services.RefreshTokenResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *Handlers) RefreshToken(c *gin.Context) {
	var req services.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.userService.RefreshToken(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，使令牌失效
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *Handlers) Logout(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	// 这里可以实现令牌黑名单逻辑
	h.auditService.LogAuditFromContext(c, "logout", "user", userID.String(), "success", "", nil)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetUserProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前用户资料
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.UserResponse
// @Failure 401 {object} map[string]interface{}
// @Router /users/profile [get]
func (h *Handlers) GetUserProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	user, err := h.userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUserProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前用户资料
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.UpdateProfileRequest true "更新信息"
// @Success 200 {object} services.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Router /users/profile [put]
func (h *Handlers) UpdateUserProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "update_profile", "user", userID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "update_profile", "user", userID.String(), "success", "", nil)
	c.JSON(http.StatusOK, user)
}

// GetProfile 获取用户资料（保持向后兼容）
// @Summary 获取用户资料
// @Description 获取当前用户资料
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.UserResponse
// @Failure 401 {object} map[string]interface{}
// @Router /users/profile [get]
func (h *Handlers) GetProfile(c *gin.Context) {
	h.GetUserProfile(c)
}

// UpdateProfile 更新用户资料（保持向后兼容）
// @Summary 更新用户资料
// @Description 更新当前用户资料
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.UpdateProfileRequest true "更新信息"
// @Success 200 {object} services.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Router /users/profile [put]
func (h *Handlers) UpdateProfile(c *gin.Context) {
	h.UpdateUserProfile(c)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户密码
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.ChangePasswordRequest true "密码信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /users/password [put]
func (h *Handlers) ChangePassword(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userService.ChangePassword(userID, &req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "change_password", "user", userID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "change_password", "user", userID.String(), "success", "", nil)
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表（需要管理员权限）
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "用户状态"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /users [get]
func (h *Handlers) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	users, total, err := h.userService.ListUsers(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ===== 告警管理相关处理器 =====

// CreateAlertRequest 创建告警请求
type CreateAlertRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Severity    string `json:"severity" binding:"required"`
	TargetType  string `json:"target_type" binding:"required"`
	TargetID    string `json:"target_id" binding:"required"`
	Conditions  string `json:"conditions" binding:"required"`
}

// UpdateAlertRequest 更新告警请求
type UpdateAlertRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Severity    *string `json:"severity"`
	Conditions  *string `json:"conditions"`
}

// CreateAlertRule 创建告警规则
// @Summary 创建告警规则
// @Description 创建新的告警规则
// @Tags Alert
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateAlertRuleRequest true "告警规则信息"
// @Success 201 {object} services.AlertRuleResponse
// @Failure 400 {object} map[string]interface{}
// @Router /alerts/rules [post]
func (h *Handlers) CreateAlertRule(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req services.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.alertService.CreateAlertRule(&req, userID)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "create_alert_rule", "alert_rule", "", "failure", err.Error(), map[string]interface{}{
			"rule_name": req.Name,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "create_alert_rule", "alert_rule", rule.ID.String(), "success", "", map[string]interface{}{
		"rule_name": rule.Name,
	})

	c.JSON(http.StatusCreated, rule)
}

// GetAlertRule 获取告警规则
// @Summary 获取告警规则
// @Description 根据ID获取告警规则详情
// @Tags Alert
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "规则ID"
// @Success 200 {object} services.AlertRuleResponse
// @Failure 404 {object} map[string]interface{}
// @Router /alerts/rules/{id} [get]
func (h *Handlers) GetAlertRule(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	rule, err := h.alertService.GetAlertRule(ruleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// ListAlertRules 获取告警规则列表
// @Summary 获取告警规则列表
// @Description 获取告警规则列表
// @Tags Alert
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param enabled query bool false "是否启用"
// @Success 200 {object} map[string]interface{}
// @Router /alerts/rules [get]
func (h *Handlers) ListAlertRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	
	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if e, err := strconv.ParseBool(enabledStr); err == nil {
			enabled = &e
		}
	}

	rules, total, err := h.alertService.ListAlertRules(page, pageSize, search, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ListAlerts 获取告警列表
// @Summary 获取告警列表
// @Description 获取告警实例列表
// @Tags Alert
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "告警状态"
// @Param severity query string false "严重程度"
// @Success 200 {object} map[string]interface{}
// @Router /alerts [get]
func (h *Handlers) ListAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	severity := c.Query("severity")
	targetType := c.Query("target_type")

	alerts, total, err := h.alertService.ListAlerts(page, pageSize, status, severity, targetType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  total,
		"page":   page,
		"page_size": pageSize,
	})
}

// AcknowledgeAlert 确认告警
// @Summary 确认告警
// @Description 确认告警实例
// @Tags Alert
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "告警ID"
// @Success 200 {object} services.AlertResponse
// @Failure 404 {object} map[string]interface{}
// @Router /alerts/{id}/acknowledge [post]
func (h *Handlers) AcknowledgeAlert(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	alert, err := h.alertService.AcknowledgeAlert(alertID, userID)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "acknowledge_alert", "alert", alertID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "acknowledge_alert", "alert", alertID.String(), "success", "", nil)
	c.JSON(http.StatusOK, alert)
}

// ResolveAlert 解决告警
// @Summary 解决告警
// @Description 解决告警实例
// @Tags Alert
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "告警ID"
// @Success 200 {object} services.AlertResponse
// @Failure 404 {object} map[string]interface{}
// @Router /alerts/{id}/resolve [post]
func (h *Handlers) ResolveAlert(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	alert, err := h.alertService.ResolveAlert(alertID, userID)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "resolve_alert", "alert", alertID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "resolve_alert", "alert", alertID.String(), "success", "", nil)
	c.JSON(http.StatusOK, alert)
}

// ===== 监控数据相关处理器 =====

// CreateMonitoringTarget 创建监控目标
// @Summary 创建监控目标
// @Description 创建新的监控目标
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateTargetRequest true "监控目标信息"
// @Success 201 {object} services.TargetResponse
// @Failure 400 {object} map[string]interface{}
// @Router /monitoring/targets [post]
func (h *Handlers) CreateMonitoringTarget(c *gin.Context) {
	var req services.CreateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	target, err := h.monitoringService.CreateTarget(&req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "create_monitoring_target", "monitoring_target", "", "failure", err.Error(), map[string]interface{}{
			"target_name": req.Name,
			"target_type": req.Type,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "create_monitoring_target", "monitoring_target", target.ID.String(), "success", "", map[string]interface{}{
		"target_name": target.Name,
		"target_type": target.Type,
	})

	c.JSON(http.StatusCreated, target)
}

// ListMonitoringTargets 获取监控目标列表
// @Summary 获取监控目标列表
// @Description 获取监控目标列表
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type query string false "目标类型"
// @Param enabled query bool false "是否启用"
// @Success 200 {object} map[string]interface{}
// @Router /monitoring/targets [get]
func (h *Handlers) ListMonitoringTargets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	targetType := c.Query("type")
	
	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if e, err := strconv.ParseBool(enabledStr); err == nil {
			enabled = &e
		}
	}

	targets, total, err := h.monitoringService.ListTargets(page, pageSize, targetType, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"targets": targets,
		"total":   total,
		"page":    page,
		"page_size": pageSize,
	})
}

// QueryMetrics 查询指标数据
// @Summary 查询指标数据
// @Description 查询Prometheus指标数据
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.MetricQueryRequest true "查询请求"
// @Success 200 {object} []services.MetricQueryResponse
// @Failure 400 {object} map[string]interface{}
// @Router /monitoring/metrics/query [post]
func (h *Handlers) QueryMetrics(c *gin.Context) {
	var req services.MetricQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metrics, err := h.monitoringService.QueryMetrics(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetSystemMetrics 获取系统指标
// @Summary 获取系统指标
// @Description 获取指定目标的系统指标
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param target_id path string true "目标ID"
// @Success 200 {object} services.SystemMetrics
// @Failure 400 {object} map[string]interface{}
// @Router /monitoring/targets/{target_id}/metrics [get]
func (h *Handlers) GetSystemMetrics(c *gin.Context) {
	targetID := c.Param("target_id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target ID is required"})
		return
	}

	metrics, err := h.monitoringService.GetSystemMetrics(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// CreateDashboard 创建仪表板
// @Summary 创建仪表板
// @Description 创建新的仪表板
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.DashboardRequest true "仪表板信息"
// @Success 201 {object} services.DashboardResponse
// @Failure 400 {object} map[string]interface{}
// @Router /monitoring/dashboards [post]
func (h *Handlers) CreateDashboard(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req services.DashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dashboard, err := h.monitoringService.CreateDashboard(&req, userID)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "create_dashboard", "dashboard", "", "failure", err.Error(), map[string]interface{}{
			"dashboard_name": req.Name,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "create_dashboard", "dashboard", dashboard.ID.String(), "success", "", map[string]interface{}{
		"dashboard_name": dashboard.Name,
	})

	c.JSON(http.StatusCreated, dashboard)
}

// ===== AI分析相关处理器 =====

// AnalyzeAlert AI分析告警
// @Summary AI分析告警
// @Description 使用AI分析告警原因和解决方案
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "告警ID"
// @Success 200 {object} services.AIAnalysisResponse
// @Failure 400 {object} map[string]interface{}
// @Router /ai/analyze/alert/{id} [post]
func (h *Handlers) AnalyzeAlert(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	// 获取告警信息
	alert, err := h.alertService.GetAlert(alertID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	// 构建AI分析请求
	req := &services.AIAnalysisRequest{
		Type:         "alert_analysis",
		AlertID:      alertID,
		RuleID:       alert.RuleID,
		TargetType:   alert.TargetType,
		TargetID:     alert.TargetID,
		MetricName:   alert.MetricName,
		CurrentValue: alert.CurrentValue,
		Threshold:    alert.Threshold,
		Condition:    alert.Condition,
		Severity:     alert.Severity,
		Tags:         alert.Tags,
		Timestamp:    time.Now(),
	}

	analysis, err := h.aiService.AnalyzeAlert(req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "ai_analyze_alert", "alert", alertID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "ai_analyze_alert", "alert", alertID.String(), "success", "", map[string]interface{}{
		"analysis_id": analysis.ID,
	})

	c.JSON(http.StatusOK, analysis)
}

// AnalyzePerformance AI性能分析
// @Summary AI性能分析
// @Description 使用AI分析系统性能
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.PerformanceAnalysisRequest true "性能分析请求"
// @Success 200 {object} services.AIAnalysisResponse
// @Failure 400 {object} map[string]interface{}
// @Router /ai/analyze/performance [post]
func (h *Handlers) AnalyzePerformance(c *gin.Context) {
	var req services.PerformanceAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为AI分析请求
	aiReq := &services.AIAnalysisRequest{
		Type:       "performance_analysis",
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Tags:       req.Tags,
		Timestamp:  time.Now(),
		Context:    req.Context,
	}

	analysis, err := h.aiService.AnalyzePerformance(aiReq)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "ai_analyze_performance", "performance", "", "failure", err.Error(), map[string]interface{}{
			"target_id": req.TargetID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "ai_analyze_performance", "performance", "", "success", "", map[string]interface{}{
		"analysis_id": analysis.ID,
		"target_id":   req.TargetID,
	})

	c.JSON(http.StatusOK, analysis)
}

// ListAIAnalysis 获取AI分析历史
// @Summary 获取AI分析历史
// @Description 获取AI分析历史记录
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param type query string false "分析类型"
// @Param target_type query string false "目标类型"
// @Success 200 {object} map[string]interface{}
// @Router /ai/analysis [get]
func (h *Handlers) ListAIAnalysis(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	analysisType := c.Query("type")
	targetType := c.Query("target_type")

	analysis, total, err := h.aiService.ListAnalysis(page, pageSize, analysisType, targetType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       analysis,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ===== 系统配置相关处理器 =====

// GetConfig 获取配置
// @Summary 获取配置
// @Description 根据键获取配置值
// @Tags Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Success 200 {object} services.ConfigResponse
// @Failure 404 {object} map[string]interface{}
// @Router /config/{key} [get]
func (h *Handlers) GetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config key is required"})
		return
	}

	config, err := h.configService.GetConfig(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfig 更新配置
// @Summary 更新配置
// @Description 更新配置值
// @Tags Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Param request body services.ConfigRequest true "配置信息"
// @Success 200 {object} services.ConfigResponse
// @Failure 400 {object} map[string]interface{}
// @Router /config/{key} [put]
func (h *Handlers) UpdateConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config key is required"})
		return
	}

	var req services.ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.configService.UpdateConfig(key, &req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "update_config", "config", key, "failure", err.Error(), map[string]interface{}{
			"config_key": key,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "update_config", "config", key, "success", "", map[string]interface{}{
		"config_key": key,
	})

	c.JSON(http.StatusOK, config)
}

// ListConfigs 获取配置列表
// @Summary 获取配置列表
// @Description 获取配置列表
// @Tags Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param category query string false "配置分类"
// @Success 200 {object} map[string]interface{}
// @Router /config [get]
func (h *Handlers) ListConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	
	var isPublic *bool
	if publicStr := c.Query("is_public"); publicStr != "" {
		if p, err := strconv.ParseBool(publicStr); err == nil {
			isPublic = &p
		}
	}

	configs, total, err := h.configService.ListConfigs(page, pageSize, category, isPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
		"total":   total,
		"page":    page,
		"page_size": pageSize,
	})
}

// UpdateDatabaseConfig 更新数据库配置
// @Summary 更新数据库配置
// @Description 更新数据库连接配置
// @Tags Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.DatabaseConfigRequest true "数据库配置"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /config/database [put]
func (h *Handlers) UpdateDatabaseConfig(c *gin.Context) {
	var req services.DatabaseConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.configService.UpdateDatabaseConfig(&req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "update_database_config", "config", "database", "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "update_database_config", "config", "database", "success", "", nil)
	c.JSON(http.StatusOK, gin.H{"message": "Database config updated successfully"})
}

// ===== 审计日志相关处理器 =====

// ListAuditLogs 获取审计日志列表
// @Summary 获取审计日志列表
// @Description 获取审计日志列表
// @Tags Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param action query string false "操作类型"
// @Param resource query string false "资源类型"
// @Param result query string false "操作结果"
// @Success 200 {object} map[string]interface{}
// @Router /audit/logs [get]
func (h *Handlers) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filter := &services.AuditLogFilter{
		Action:   c.Query("action"),
		Resource: c.Query("resource"),
		Result:   c.Query("result"),
	}

	// 解析时间范围
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	logs, total, err := h.auditService.ListAuditLogs(page, pageSize, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// GetAuditStats 获取审计统计
// @Summary 获取审计统计
// @Description 获取审计日志统计信息
// @Tags Audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "统计天数" default(7)
// @Success 200 {object} services.AuditStats
// @Router /audit/stats [get]
func (h *Handlers) GetAuditStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	stats, err := h.auditService.GetAuditStats(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ExportAuditLogs 导出审计日志
// @Summary 导出审计日志
// @Description 导出审计日志为文件
// @Tags Audit
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param format query string false "导出格式" default(json) Enums(json,csv)
// @Success 200 {file} file
// @Router /audit/logs/export [get]
func (h *Handlers) ExportAuditLogs(c *gin.Context) {
	format := c.DefaultQuery("format", "json")

	filter := &services.AuditLogFilter{
		Action:   c.Query("action"),
		Resource: c.Query("resource"),
		Result:   c.Query("result"),
	}

	// 解析时间范围
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	data, err := h.auditService.ExportAuditLogs(filter, format)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "export_audit_logs", "audit_log", "", "failure", err.Error(), map[string]interface{}{
			"format": format,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditService.LogAuditFromContext(c, "export_audit_logs", "audit_log", "", "success", "", map[string]interface{}{
		"format": format,
	})

	// 设置响应头
	filename := "audit_logs_" + time.Now().Format("20060102_150405")
	if format == "csv" {
		filename += ".csv"
		c.Header("Content-Type", "text/csv")
	} else {
		filename += ".json"
		c.Header("Content-Type", "application/json")
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)

	c.Data(http.StatusOK, c.GetHeader("Content-Type"), data)
}

// ===== 中间件监控相关处理器 =====

// GetMySQLMetrics 获取MySQL指标
func (h *Handlers) GetMySQLMetrics(c *gin.Context) {
	h.middlewareHandler.GetMySQLMetrics(c)
}

// GetRedisMetrics 获取Redis指标
func (h *Handlers) GetRedisMetrics(c *gin.Context) {
	h.middlewareHandler.GetRedisMetrics(c)
}

// GetKafkaMetrics 获取Kafka指标
func (h *Handlers) GetKafkaMetrics(c *gin.Context) {
	h.middlewareHandler.GetKafkaMetrics(c)
}

// ListMiddleware 获取中间件列表
func (h *Handlers) ListMiddleware(c *gin.Context) {
	h.middlewareHandler.ListMiddleware(c)
}

// CreateMiddleware 创建中间件监控
func (h *Handlers) CreateMiddleware(c *gin.Context) {
	h.middlewareHandler.CreateMiddleware(c)
}

// GetMiddleware 获取中间件详情
func (h *Handlers) GetMiddleware(c *gin.Context) {
	h.middlewareHandler.GetMiddleware(c)
}

// UpdateMiddleware 更新中间件监控
func (h *Handlers) UpdateMiddleware(c *gin.Context) {
	h.middlewareHandler.UpdateMiddleware(c)
}

// DeleteMiddleware 删除中间件监控
func (h *Handlers) DeleteMiddleware(c *gin.Context) {
	h.middlewareHandler.DeleteMiddleware(c)
}

// ===== APM应用性能监控相关处理器 =====

// GetTraces 获取链路追踪列表
func (h *Handlers) GetTraces(c *gin.Context) {
	h.apmHandler.GetTraces(c)
}

// GetTraceDetail 获取链路追踪详情
func (h *Handlers) GetTraceDetail(c *gin.Context) {
	h.apmHandler.GetTraceDetail(c)
}

// GetServices 获取服务列表
func (h *Handlers) GetServices(c *gin.Context) {
	h.apmHandler.GetServices(c)
}

// GetServiceDetail 获取服务详情
func (h *Handlers) GetServiceDetail(c *gin.Context) {
	h.apmHandler.GetServiceDetail(c)
}

// GetServicePerformance 获取服务性能概览
func (h *Handlers) GetServicePerformance(c *gin.Context) {
	h.apmHandler.GetServicePerformance(c)
}

// CreateService 创建服务监控
func (h *Handlers) CreateService(c *gin.Context) {
	h.apmHandler.CreateService(c)
}

// UpdateService 更新服务监控
func (h *Handlers) UpdateService(c *gin.Context) {
	h.apmHandler.UpdateService(c)
}

// DeleteService 删除服务监控
func (h *Handlers) DeleteService(c *gin.Context) {
	h.apmHandler.DeleteService(c)
}

// GetServiceTopology 获取服务拓扑图
func (h *Handlers) GetServiceTopology(c *gin.Context) {
	h.apmHandler.GetServiceTopology(c)
}

// ===== 容器监控相关处理器 =====

// GetDockerContainers 获取Docker容器列表
func (h *Handlers) GetDockerContainers(c *gin.Context) {
	h.containerHandler.GetDockerContainers(c)
}

// GetDockerContainerDetail 获取Docker容器详情
func (h *Handlers) GetDockerContainerDetail(c *gin.Context) {
	h.containerHandler.GetDockerContainerDetail(c)
}

// GetKubernetesPods 获取Kubernetes Pod列表
func (h *Handlers) GetKubernetesPods(c *gin.Context) {
	h.containerHandler.GetKubernetesPods(c)
}

// GetKubernetesNodes 获取Kubernetes节点列表
func (h *Handlers) GetKubernetesNodes(c *gin.Context) {
	h.containerHandler.GetKubernetesNodes(c)
}

// GetKubernetesNamespaces 获取Kubernetes命名空间列表
func (h *Handlers) GetKubernetesNamespaces(c *gin.Context) {
	h.containerHandler.GetKubernetesNamespaces(c)
}

// GetClusterMetrics 获取集群指标
func (h *Handlers) GetClusterMetrics(c *gin.Context) {
	h.containerHandler.GetClusterMetrics(c)
}

// ListContainerMonitors 获取容器监控列表
func (h *Handlers) ListContainerMonitors(c *gin.Context) {
	h.containerHandler.ListContainerMonitors(c)
}

// CreateContainerMonitor 创建容器监控
func (h *Handlers) CreateContainerMonitor(c *gin.Context) {
	h.containerHandler.CreateContainerMonitor(c)
}

// GetContainerMonitor 获取容器监控详情
func (h *Handlers) GetContainerMonitor(c *gin.Context) {
	h.containerHandler.GetContainerMonitor(c)
}

// UpdateContainerMonitor 更新容器监控
func (h *Handlers) UpdateContainerMonitor(c *gin.Context) {
	h.containerHandler.UpdateContainerMonitor(c)
}

// DeleteContainerMonitor 删除容器监控
func (h *Handlers) DeleteContainerMonitor(c *gin.Context) {
	h.containerHandler.DeleteContainerMonitor(c)
}

// GetResourceUsage 获取资源使用情况
func (h *Handlers) GetResourceUsage(c *gin.Context) {
	h.containerHandler.GetResourceUsage(c)
}

// ===== Agent管理相关处理器 =====

// ListAgents 获取Agent列表
func (h *Handlers) ListAgents(c *gin.Context) {
	h.agentHandler.ListAgents(c)
}

// CreateAgent 创建Agent
func (h *Handlers) CreateAgent(c *gin.Context) {
	h.agentHandler.CreateAgent(c)
}

// GetAgent 获取Agent详情
func (h *Handlers) GetAgent(c *gin.Context) {
	h.agentHandler.GetAgent(c)
}

// UpdateAgent 更新Agent
func (h *Handlers) UpdateAgent(c *gin.Context) {
	h.agentHandler.UpdateAgent(c)
}

// DeleteAgent 删除Agent
func (h *Handlers) DeleteAgent(c *gin.Context) {
	h.agentHandler.DeleteAgent(c)
}

// GetAgentConfig 获取Agent配置
func (h *Handlers) GetAgentConfig(c *gin.Context) {
	h.agentHandler.GetAgentConfig(c)
}

// UpdateAgentConfig 更新Agent配置
func (h *Handlers) UpdateAgentConfig(c *gin.Context) {
	h.agentHandler.UpdateAgentConfig(c)
}

// ListDeployments 获取部署列表
func (h *Handlers) ListDeployments(c *gin.Context) {
	h.agentHandler.ListDeployments(c)
}

// CreateDeployment 创建部署
func (h *Handlers) CreateDeployment(c *gin.Context) {
	h.agentHandler.CreateDeployment(c)
}

// GetDeployment 获取部署详情
func (h *Handlers) GetDeployment(c *gin.Context) {
	h.agentHandler.GetDeployment(c)
}

// GetAgentPackages 获取Agent安装包列表
func (h *Handlers) GetAgentPackages(c *gin.Context) {
	h.agentHandler.GetAgentPackages(c)
}

// DownloadAgent 下载Agent
func (h *Handlers) DownloadAgent(c *gin.Context) {
	h.agentHandler.DownloadAgent(c)
}

// DownloadAgentPackage 下载Agent安装包
func (h *Handlers) DownloadAgentPackage(c *gin.Context) {
	h.agentHandler.DownloadAgentPackage(c)
}

// GetAgentInstallGuide 获取Agent安装指南
func (h *Handlers) GetAgentInstallGuide(c *gin.Context) {
	h.agentHandler.GetAgentInstallGuide(c)
}

// ProcessHeartbeat 处理Agent心跳
func (h *Handlers) ProcessHeartbeat(c *gin.Context) {
	h.agentHandler.ProcessHeartbeat(c)
}

// ===== 监控管理相关处理器 =====

// GetServers 获取服务器监控列表
func (h *Handlers) GetServers(c *gin.Context) {
	// TODO: 实现服务器监控列表获取
	c.JSON(http.StatusOK, gin.H{"message": "GetServers not implemented yet"})
}

// CreateServer 创建服务器监控
func (h *Handlers) CreateServer(c *gin.Context) {
	// TODO: 实现服务器监控创建
	c.JSON(http.StatusOK, gin.H{"message": "CreateServer not implemented yet"})
}

// UpdateServer 更新服务器监控
func (h *Handlers) UpdateServer(c *gin.Context) {
	// TODO: 实现服务器监控更新
	c.JSON(http.StatusOK, gin.H{"message": "UpdateServer not implemented yet"})
}

// DeleteServer 删除服务器监控
func (h *Handlers) DeleteServer(c *gin.Context) {
	// TODO: 实现服务器监控删除
	c.JSON(http.StatusOK, gin.H{"message": "DeleteServer not implemented yet"})
}

// GetProcesses 获取进程监控列表
func (h *Handlers) GetProcesses(c *gin.Context) {
	// TODO: 实现进程监控列表获取
	c.JSON(http.StatusOK, gin.H{"message": "GetProcesses not implemented yet"})
}

// CreateProcess 创建进程监控
func (h *Handlers) CreateProcess(c *gin.Context) {
	// TODO: 实现进程监控创建
	c.JSON(http.StatusOK, gin.H{"message": "CreateProcess not implemented yet"})
}

// UpdateProcess 更新进程监控
func (h *Handlers) UpdateProcess(c *gin.Context) {
	// TODO: 实现进程监控更新
	c.JSON(http.StatusOK, gin.H{"message": "UpdateProcess not implemented yet"})
}

// DeleteProcess 删除进程监控
func (h *Handlers) DeleteProcess(c *gin.Context) {
	// TODO: 实现进程监控删除
	c.JSON(http.StatusOK, gin.H{"message": "DeleteProcess not implemented yet"})
}

// ===== 系统配置相关处理器 =====

// GetConfigs 获取配置列表
func (h *Handlers) GetConfigs(c *gin.Context) {
	// TODO: 实现配置列表获取
	c.JSON(http.StatusOK, gin.H{"message": "GetConfigs not implemented yet"})
}

// CreateConfig 创建配置
func (h *Handlers) CreateConfig(c *gin.Context) {
	// TODO: 实现配置创建
	c.JSON(http.StatusOK, gin.H{"message": "CreateConfig not implemented yet"})
}

// DeleteConfig 删除配置
func (h *Handlers) DeleteConfig(c *gin.Context) {
	// TODO: 实现配置删除
	c.JSON(http.StatusOK, gin.H{"message": "DeleteConfig not implemented yet"})
}

// GetDatabaseConfig 获取数据库配置
func (h *Handlers) GetDatabaseConfig(c *gin.Context) {
	// TODO: 实现数据库配置获取
	c.JSON(http.StatusOK, gin.H{"message": "GetDatabaseConfig not implemented yet"})
}

// GetRedisConfig 获取Redis配置
func (h *Handlers) GetRedisConfig(c *gin.Context) {
	// TODO: 实现Redis配置获取
	c.JSON(http.StatusOK, gin.H{"message": "GetRedisConfig not implemented yet"})
}

// UpdateRedisConfig 更新Redis配置
func (h *Handlers) UpdateRedisConfig(c *gin.Context) {
	// TODO: 实现Redis配置更新
	c.JSON(http.StatusOK, gin.H{"message": "UpdateRedisConfig not implemented yet"})
}

// GetAIModelConfig 获取AI模型配置
func (h *Handlers) GetAIModelConfig(c *gin.Context) {
	h.configHandler.GetAIServiceConfig(c)
}

// UpdateAIModelConfig 更新AI模型配置
func (h *Handlers) UpdateAIModelConfig(c *gin.Context) {
	h.configHandler.UpdateAIServiceConfig(c)
}

// GetEmailConfig 获取邮件配置
func (h *Handlers) GetEmailConfig(c *gin.Context) {
	h.configHandler.GetAlertConfig(c)
}

// UpdateEmailConfig 更新邮件配置
func (h *Handlers) UpdateEmailConfig(c *gin.Context) {
	h.configHandler.UpdateAlertConfig(c)
}

// GetPrometheusConfig 获取Prometheus配置
func (h *Handlers) GetPrometheusConfig(c *gin.Context) {
	// TODO: 实现Prometheus配置获取
	c.JSON(http.StatusOK, gin.H{"message": "GetPrometheusConfig not implemented yet"})
}

// UpdatePrometheusConfig 更新Prometheus配置
func (h *Handlers) UpdatePrometheusConfig(c *gin.Context) {
	// TODO: 实现Prometheus配置更新
	c.JSON(http.StatusOK, gin.H{"message": "UpdatePrometheusConfig not implemented yet"})
}

// GetSystemConfig 获取系统配置
func (h *Handlers) GetSystemConfig(c *gin.Context) {
	h.configHandler.GetSystemSettings(c)
}

// UpdateSystemConfig 更新系统配置
func (h *Handlers) UpdateSystemConfig(c *gin.Context) {
	h.configHandler.UpdateSystemSettings(c)
}

// GetAlertConfig 获取告警配置
func (h *Handlers) GetAlertConfig(c *gin.Context) {
	h.configHandler.GetAlertConfig(c)
}

// UpdateAlertConfig 更新告警配置
func (h *Handlers) UpdateAlertConfig(c *gin.Context) {
	h.configHandler.UpdateAlertConfig(c)
}

// TestEmailConfig 测试邮件配置
func (h *Handlers) TestEmailConfig(c *gin.Context) {
	h.configHandler.TestEmailConfig(c)
}

// TestSMSConfig 测试短信配置
func (h *Handlers) TestSMSConfig(c *gin.Context) {
	h.configHandler.TestSMSConfig(c)
}

// GetAIServiceConfig 获取AI服务配置
func (h *Handlers) GetAIServiceConfig(c *gin.Context) {
	h.configHandler.GetAIServiceConfig(c)
}

// UpdateAIServiceConfig 更新AI服务配置
func (h *Handlers) UpdateAIServiceConfig(c *gin.Context) {
	h.configHandler.UpdateAIServiceConfig(c)
}

// ===== 用户管理相关处理器 =====

// GetUser 获取单个用户
func (h *Handlers) GetUser(c *gin.Context) {
	// TODO: 实现单个用户获取
	c.JSON(http.StatusOK, gin.H{"message": "GetUser not implemented yet"})
}

// UpdateUserStatus 更新用户状态
func (h *Handlers) UpdateUserStatus(c *gin.Context) {
	// TODO: 实现用户状态更新
	c.JSON(http.StatusOK, gin.H{"message": "UpdateUserStatus not implemented yet"})
}

// UpdateUserRoles 更新用户角色
func (h *Handlers) UpdateUserRoles(c *gin.Context) {
	// TODO: 实现用户角色更新
	c.JSON(http.StatusOK, gin.H{"message": "UpdateUserRoles not implemented yet"})
}

// GetRoles 获取角色列表
func (h *Handlers) GetRoles(c *gin.Context) {
	// TODO: 实现角色列表获取
	c.JSON(http.StatusOK, gin.H{"message": "GetRoles not implemented yet"})
}

// CreateRole 创建角色
func (h *Handlers) CreateRole(c *gin.Context) {
	// TODO: 实现角色创建
	c.JSON(http.StatusOK, gin.H{"message": "CreateRole not implemented yet"})
}

// UpdateRole 更新角色
func (h *Handlers) UpdateRole(c *gin.Context) {
	// TODO: 实现角色更新
	c.JSON(http.StatusOK, gin.H{"message": "UpdateRole not implemented yet"})
}

// DeleteRole 删除角色
func (h *Handlers) DeleteRole(c *gin.Context) {
	// TODO: 实现角色删除
	c.JSON(http.StatusOK, gin.H{"message": "DeleteRole not implemented yet"})
}

// GetPermissions 获取权限列表
func (h *Handlers) GetPermissions(c *gin.Context) {
	// TODO: 实现权限列表获取
	c.JSON(http.StatusOK, gin.H{"message": "GetPermissions not implemented yet"})
}

// CreateSystemBackup 创建系统备份
func (h *Handlers) CreateSystemBackup(c *gin.Context) {
	// TODO: 实现系统备份创建
	c.JSON(http.StatusOK, gin.H{"message": "CreateSystemBackup not implemented yet"})
}

// GetSystemBackups 获取系统备份列表
func (h *Handlers) GetSystemBackups(c *gin.Context) {
	// TODO: 实现系统备份列表获取
	c.JSON(http.StatusOK, gin.H{"message": "GetSystemBackups not implemented yet"})
}

// ===== 告警管理相关处理器 =====

// GetAlerts 获取告警列表
func (h *Handlers) GetAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	severity := c.Query("severity")
	targetType := c.Query("target_type")

	alerts, total, err := h.alertService.ListAlerts(page, pageSize, status, severity, targetType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	response := PaginatedResponse{
		Data: alerts,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    int(total),
			Pages:    (int(total) + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, response)
}

// CreateAlert 创建告警
func (h *Handlers) CreateAlert(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 创建告警规则请求
	ruleReq := &services.CreateAlertRuleRequest{
		Name:        req.Name,
		Description: req.Description,
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		MetricName:  "custom_metric", // 默认指标名
		Condition:   ">",             // 默认条件
		Threshold:   0,               // 默认阈值
		Duration:    300,             // 默认持续时间5分钟
		Severity:    req.Severity,
		Enabled:     true,
		Tags:        map[string]interface{}{"conditions": req.Conditions},
		Channels:    []string{"default"},
	}

	alert, err := h.alertService.CreateAlertRule(ruleReq, userID)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "create_alert", "alert", "", "failure", err.Error(), map[string]interface{}{
			"alert_name": req.Name,
		})
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "create_alert", "alert", alert.ID.String(), "success", "", map[string]interface{}{
		"alert_name": alert.Name,
	})

	c.JSON(http.StatusCreated, alert)
}

// GetAlert 获取告警详情
func (h *Handlers) GetAlert(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	alert, err := h.alertService.GetAlert(alertID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Not Found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// UpdateAlert 更新告警
func (h *Handlers) UpdateAlert(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	var req UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 创建告警规则更新请求
	ruleReq := &services.UpdateAlertRuleRequest{}
	if req.Name != nil {
		ruleReq.Name = *req.Name
	}
	if req.Description != nil {
		ruleReq.Description = *req.Description
	}
	if req.Severity != nil {
		ruleReq.Severity = *req.Severity
	}
	if req.Conditions != nil {
		ruleReq.Tags = map[string]interface{}{"conditions": *req.Conditions}
	}

	alert, err := h.alertService.UpdateAlertRule(alertID, ruleReq)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "update_alert", "alert", alertID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "update_alert", "alert", alertID.String(), "success", "", nil)
	c.JSON(http.StatusOK, alert)
}

// DeleteAlert 删除告警
func (h *Handlers) DeleteAlert(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	err = h.alertService.DeleteAlertRule(alertID)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "delete_alert", "alert", alertID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "delete_alert", "alert", alertID.String(), "success", "", nil)
	c.Status(http.StatusNoContent)
}

// GetAlertRules 获取告警规则列表
func (h *Handlers) GetAlertRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	enabledStr := c.Query("enabled")
	var enabled *bool
	if enabledStr != "" {
		if enabledStr == "true" {
			e := true
			enabled = &e
		} else if enabledStr == "false" {
			e := false
			enabled = &e
		}
	}

	rules, total, err := h.alertService.ListAlertRules(page, pageSize, search, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	response := PaginatedResponse{
		Data: rules,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    int(total),
			Pages:    (int(total) + pageSize - 1) / pageSize,
		},
	}
 
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取告警规则列表成功",
		"data": response,
	})
}

// UpdateAlertRule 更新告警规则
func (h *Handlers) UpdateAlertRule(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	var req services.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	rule, err := h.alertService.UpdateAlertRule(ruleID, &req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "update_alert_rule", "alert_rule", ruleID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "update_alert_rule", "alert_rule", ruleID.String(), "success", "", nil)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "更新告警规则成功",
		"data": rule,
	})
}

// DeleteAlertRule 删除告警规则
func (h *Handlers) DeleteAlertRule(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	err = h.alertService.DeleteAlertRule(ruleID)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "delete_alert_rule", "alert_rule", ruleID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "delete_alert_rule", "alert_rule", ruleID.String(), "success", "", nil)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "删除告警规则成功",
	})
}

// EnableAlertRule 启用告警规则
func (h *Handlers) EnableAlertRule(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	enabled := true
	req := &services.UpdateAlertRuleRequest{
		Enabled: &enabled,
	}

	_, err = h.alertService.UpdateAlertRule(ruleID, req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "enable_alert_rule", "alert_rule", ruleID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "enable_alert_rule", "alert_rule", ruleID.String(), "success", "", nil)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "启用告警规则成功",
	})
}

// DisableAlertRule 禁用告警规则
func (h *Handlers) DisableAlertRule(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	enabled := false
	req := &services.UpdateAlertRuleRequest{
		Enabled: &enabled,
	}

	_, err = h.alertService.UpdateAlertRule(ruleID, req)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "disable_alert_rule", "alert_rule", ruleID.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "disable_alert_rule", "alert_rule", ruleID.String(), "success", "", nil)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "禁用告警规则成功",
	})
}

// ===== 监控目标管理相关处理器 =====

// GetTargets 获取监控目标列表
func (h *Handlers) GetTargets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	targetType := c.Query("type")
	status := c.Query("status")

	_ = targetType // 避免未使用变量错误
	_ = status     // 避免未使用变量错误

	// 模拟数据
	targets := []map[string]interface{}{
		{
			"id":          "target-1",
			"name":        "Web Server 1",
			"type":        "host",
			"address":     "192.168.1.100",
			"status":      "online",
			"last_seen":   time.Now().Add(-5 * time.Minute),
			"created_at":  time.Now().Add(-24 * time.Hour),
		},
		{
			"id":          "target-2",
			"name":        "Database Server",
			"type":        "service",
			"address":     "192.168.1.101",
			"status":      "online",
			"last_seen":   time.Now().Add(-2 * time.Minute),
			"created_at":  time.Now().Add(-48 * time.Hour),
		},
	}

	response := PaginatedResponse{
		Data: targets,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(targets),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取监控目标列表成功",
		"data": response,
	})
}

// CreateTarget 创建监控目标
func (h *Handlers) CreateTarget(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 模拟创建目标
	target := map[string]interface{}{
		"id":         uuid.New().String(),
		"name":       req["name"],
		"type":       req["type"],
		"address":    req["address"],
		"status":     "pending",
		"created_at": time.Now(),
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "创建监控目标成功",
		"data": target,
	})
}

// GetTarget 获取监控目标详情
func (h *Handlers) GetTarget(c *gin.Context) {
	targetID := c.Param("id")

	// 模拟获取目标详情
	target := map[string]interface{}{
		"id":          targetID,
		"name":        "Web Server 1",
		"type":        "host",
		"address":     "192.168.1.100",
		"status":      "online",
		"last_seen":   time.Now().Add(-5 * time.Minute),
		"created_at":  time.Now().Add(-24 * time.Hour),
		"metrics": map[string]interface{}{
			"cpu_usage":    75.5,
			"memory_usage": 68.2,
			"disk_usage":   45.8,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取监控目标详情成功",
		"data": target,
	})
}

// UpdateTarget 更新监控目标
func (h *Handlers) UpdateTarget(c *gin.Context) {
	targetID := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 模拟更新目标
	target := map[string]interface{}{
		"id":         targetID,
		"name":       req["name"],
		"type":       req["type"],
		"address":    req["address"],
		"status":     "online",
		"updated_at": time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "更新监控目标成功",
		"data": target,
	})
}

// DeleteTarget 删除监控目标
func (h *Handlers) DeleteTarget(c *gin.Context) {
	targetID := c.Param("id")
	_ = targetID // 避免未使用变量错误

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "删除监控目标成功",
	})
}

// GetTargetMetrics 获取目标指标
func (h *Handlers) GetTargetMetrics(c *gin.Context) {
	targetID := c.Param("id")
	metricName := c.Query("metric")
	startTime := c.Query("start")
	endTime := c.Query("end")

	_ = targetID   // 避免未使用变量错误
	_ = metricName // 避免未使用变量错误
	_ = startTime  // 避免未使用变量错误
	_ = endTime    // 避免未使用变量错误

	// 模拟指标数据
	metrics := map[string]interface{}{
		"metric_name": metricName,
		"target_id":   targetID,
		"data_points": []map[string]interface{}{
			{
				"timestamp": time.Now().Add(-10 * time.Minute).Unix(),
				"value":     75.5,
			},
			{
				"timestamp": time.Now().Add(-5 * time.Minute).Unix(),
				"value":     78.2,
			},
			{
				"timestamp": time.Now().Unix(),
				"value":     72.1,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取目标指标成功",
		"data": metrics,
	})
}

// GetTargetStatus 获取目标状态
func (h *Handlers) GetTargetStatus(c *gin.Context) {
	targetID := c.Param("id")
	_ = targetID // 避免未使用变量错误

	// 模拟状态数据
	status := map[string]interface{}{
		"target_id":  targetID,
		"status":     "online",
		"last_seen":  time.Now().Add(-2 * time.Minute),
		"uptime":     "99.9%",
		"response_time": 45.2,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取目标状态成功",
		"data": status,
	})
}

// ===== 指标查询相关处理器 =====

// QueryRangeMetrics 查询范围指标
func (h *Handlers) QueryRangeMetrics(c *gin.Context) {
	query := c.Query("query")
	startTime := c.Query("start")
	endTime := c.Query("end")
	step := c.Query("step")

	_ = query     // 避免未使用变量错误
	_ = startTime // 避免未使用变量错误
	_ = endTime   // 避免未使用变量错误
	_ = step      // 避免未使用变量错误

	// 模拟查询结果
	result := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "matrix",
			"result": []map[string]interface{}{
				{
					"metric": map[string]string{
						"__name__": "cpu_usage",
						"instance": "localhost:9090",
					},
					"values": [][]interface{}{
						{time.Now().Add(-10 * time.Minute).Unix(), "75.5"},
						{time.Now().Add(-5 * time.Minute).Unix(), "78.2"},
						{time.Now().Unix(), "72.1"},
					},
				},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "查询范围指标成功",
		"data": result,
	})
}

// GetMetricLabels 获取指标标签
func (h *Handlers) GetMetricLabels(c *gin.Context) {
	metricName := c.Query("metric")
	_ = metricName // 避免未使用变量错误

	// 模拟标签数据
	labels := []string{
		"instance",
		"job",
		"environment",
		"region",
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取指标标签成功",
		"data": labels,
	})
}

// GetMetricValues 获取指标值
func (h *Handlers) GetMetricValues(c *gin.Context) {
	labelName := c.Query("label")
	metricName := c.Query("metric")

	_ = labelName  // 避免未使用变量错误
	_ = metricName // 避免未使用变量错误

	// 模拟标签值数据
	values := []string{
		"localhost:9090",
		"localhost:9091",
		"localhost:9092",
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取指标值成功",
		"data": values,
	})
}

// ===== AI分析相关处理器 =====

// AnalyzeData 分析数据
func (h *Handlers) AnalyzeData(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 模拟分析结果
	analysis := map[string]interface{}{
		"id":         uuid.New().String(),
		"type":       req["type"],
		"status":     "completed",
		"result": map[string]interface{}{
			"anomalies": []map[string]interface{}{
				{
					"timestamp": time.Now().Unix(),
					"metric":    "cpu_usage",
					"value":     95.5,
					"severity":  "high",
				},
			},
			"insights": []string{
				"CPU使用率异常高",
				"建议检查系统负载",
			},
		},
		"created_at": time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "数据分析成功",
		"data": analysis,
	})
}

// GetAnalyses 获取分析列表
func (h *Handlers) GetAnalyses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 模拟分析数据
	analyses := []map[string]interface{}{
		{
			"id":         "analysis-1",
			"type":       "anomaly_detection",
			"status":     "completed",
			"created_at": time.Now().Add(-1 * time.Hour),
		},
		{
			"id":         "analysis-2",
			"type":       "trend_prediction",
			"status":     "running",
			"created_at": time.Now().Add(-30 * time.Minute),
		},
	}

	response := PaginatedResponse{
		Data: analyses,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(analyses),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取分析列表成功",
		"data": response,
	})
}

// GetAnalysis 获取分析详情
func (h *Handlers) GetAnalysis(c *gin.Context) {
	analysisID := c.Param("id")

	// 模拟分析详情
	analysis := map[string]interface{}{
		"id":     analysisID,
		"type":   "anomaly_detection",
		"status": "completed",
		"result": map[string]interface{}{
			"anomalies": []map[string]interface{}{
				{
					"timestamp": time.Now().Unix(),
					"metric":    "cpu_usage",
					"value":     95.5,
					"severity":  "high",
				},
			},
		},
		"created_at": time.Now().Add(-1 * time.Hour),
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取分析详情成功",
		"data": analysis,
	})
}

// DeleteAnalysis 删除分析
func (h *Handlers) DeleteAnalysis(c *gin.Context) {
	analysisID := c.Param("id")
	_ = analysisID // 避免未使用变量错误

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "删除分析成功",
	})
}

// PredictTrend 预测趋势
func (h *Handlers) PredictTrend(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 模拟趋势预测
	prediction := map[string]interface{}{
		"metric":     req["metric"],
		"timeframe": req["timeframe"],
		"prediction": []map[string]interface{}{
			{
				"timestamp": time.Now().Add(1 * time.Hour).Unix(),
				"value":     78.5,
				"confidence": 0.85,
			},
			{
				"timestamp": time.Now().Add(2 * time.Hour).Unix(),
				"value":     82.1,
				"confidence": 0.78,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "趋势预测成功",
		"data": prediction,
	})
}

// GetInsights 获取洞察
func (h *Handlers) GetInsights(c *gin.Context) {
	// 模拟洞察数据
	insights := []map[string]interface{}{
		{
			"type":        "performance",
			"title":       "CPU使用率异常",
			"description": "过去1小时内CPU使用率持续超过90%",
			"severity":    "high",
			"timestamp":   time.Now().Add(-30 * time.Minute),
		},
		{
			"type":        "capacity",
			"title":       "内存使用趋势",
			"description": "内存使用率呈上升趋势，预计2小时后达到阈值",
			"severity":    "medium",
			"timestamp":   time.Now().Add(-15 * time.Minute),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取洞察成功",
		"data": insights,
	})
}

// ===== 中间件管理相关处理器 =====

// GetMiddlewareList 获取中间件列表
func (h *Handlers) GetMiddlewareList(c *gin.Context) {
	// 模拟中间件数据
	middlewares := []map[string]interface{}{
		{
			"name":        "auth",
			"description": "身份验证中间件",
			"enabled":     true,
			"order":       1,
		},
		{
			"name":        "cors",
			"description": "跨域请求中间件",
			"enabled":     true,
			"order":       2,
		},
		{
			"name":        "rate_limit",
			"description": "限流中间件",
			"enabled":     false,
			"order":       3,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取中间件列表成功",
		"data": middlewares,
	})
}

// ===== APM相关处理器 =====

// GetOperations 获取操作列表
func (h *Handlers) GetOperations(c *gin.Context) {
	service := c.Query("service")
	startTime := c.Query("start")
	endTime := c.Query("end")

	_ = service   // 避免未使用变量错误
	_ = startTime // 避免未使用变量错误
	_ = endTime   // 避免未使用变量错误

	// 模拟操作数据
	operations := []map[string]interface{}{
		{
			"name":         "GET /api/users",
			"service":      "user-service",
			"avg_duration": 45.2,
			"request_count": 1250,
			"error_rate":   0.02,
		},
		{
			"name":         "POST /api/orders",
			"service":      "order-service",
			"avg_duration": 128.7,
			"request_count": 856,
			"error_rate":   0.01,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取操作列表成功",
		"data": operations,
	})
}

// GetServiceMap 获取服务地图
func (h *Handlers) GetServiceMap(c *gin.Context) {
	// 模拟服务地图数据
	serviceMap := map[string]interface{}{
		"nodes": []map[string]interface{}{
			{
				"id":       "user-service",
				"name":     "用户服务",
				"type":     "service",
				"status":   "healthy",
				"requests": 1250,
			},
			{
				"id":       "order-service",
				"name":     "订单服务",
				"type":     "service",
				"status":   "healthy",
				"requests": 856,
			},
			{
				"id":       "database",
				"name":     "数据库",
				"type":     "database",
				"status":   "healthy",
				"requests": 2106,
			},
		},
		"edges": []map[string]interface{}{
			{
				"source":   "user-service",
				"target":   "database",
				"requests": 1250,
				"latency":  25.3,
			},
			{
				"source":   "order-service",
				"target":   "database",
				"requests": 856,
				"latency":  32.1,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取服务地图成功",
		"data": serviceMap,
	})
}

// ===== 代理管理相关处理器 =====

// GetAgents 获取代理列表
func (h *Handlers) GetAgents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	_ = status // 避免未使用变量错误

	// 模拟代理数据
	agents := []map[string]interface{}{
		{
			"id":         "agent-1",
			"name":       "Web Server Agent",
			"host":       "192.168.1.100",
			"status":     "online",
			"version":    "1.0.0",
			"last_seen":  time.Now().Add(-2 * time.Minute),
			"created_at": time.Now().Add(-24 * time.Hour),
		},
		{
			"id":         "agent-2",
			"name":       "Database Agent",
			"host":       "192.168.1.101",
			"status":     "offline",
			"version":    "1.0.0",
			"last_seen":  time.Now().Add(-15 * time.Minute),
			"created_at": time.Now().Add(-48 * time.Hour),
		},
	}

	response := PaginatedResponse{
		Data: agents,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(agents),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取代理列表成功",
		"data": response,
	})
}

// HandleAgentHeartbeat 处理代理心跳
func (h *Handlers) HandleAgentHeartbeat(c *gin.Context) {
	agentID := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 模拟处理心跳
	heartbeat := map[string]interface{}{
		"agent_id":   agentID,
		"status":     "received",
		"timestamp":  time.Now(),
		"next_check": time.Now().Add(30 * time.Second),
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "代理心跳处理成功",
		"data": heartbeat,
	})
}

// GetDeployments 获取部署列表
func (h *Handlers) GetDeployments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	_ = status // 避免未使用变量错误

	// 模拟部署数据
	deployments := []map[string]interface{}{
		{
			"id":          "deploy-1",
			"name":        "Web Application v1.2.0",
			"version":     "1.2.0",
			"status":      "running",
			"environment": "production",
			"deployed_at": time.Now().Add(-2 * time.Hour),
		},
		{
			"id":          "deploy-2",
			"name":        "API Service v2.1.0",
			"version":     "2.1.0",
			"status":      "deploying",
			"environment": "staging",
			"deployed_at": time.Now().Add(-30 * time.Minute),
		},
	}

	response := PaginatedResponse{
		Data: deployments,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(deployments),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取部署列表成功",
		"data": response,
	})
}

// ===== 管理员相关处理器 =====

// GetUsers 获取用户列表
func (h *Handlers) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	role := c.Query("role")
	status := c.Query("status")

	_ = role   // 避免未使用变量错误
	_ = status // 避免未使用变量错误

	// 模拟用户数据
	users := []map[string]interface{}{
		{
			"id":         "user-1",
			"username":   "admin",
			"email":      "admin@example.com",
			"role":       "admin",
			"status":     "active",
			"created_at": time.Now().Add(-30 * 24 * time.Hour),
			"last_login": time.Now().Add(-2 * time.Hour),
		},
		{
			"id":         "user-2",
			"username":   "operator",
			"email":      "operator@example.com",
			"role":       "operator",
			"status":     "active",
			"created_at": time.Now().Add(-15 * 24 * time.Hour),
			"last_login": time.Now().Add(-1 * time.Hour),
		},
	}

	response := PaginatedResponse{
		Data: users,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(users),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取用户列表成功",
		"data": response,
	})
}

// CreateUser 创建用户
func (h *Handlers) CreateUser(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 模拟创建用户
	user := map[string]interface{}{
		"id":         uuid.New().String(),
		"username":   req["username"],
		"email":      req["email"],
		"role":       req["role"],
		"status":     "active",
		"created_at": time.Now(),
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "用户创建成功",
		"data": user,
	})
}

// UpdateUser 更新用户
func (h *Handlers) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 模拟更新用户
	user := map[string]interface{}{
		"id":         userID,
		"username":   req["username"],
		"email":      req["email"],
		"role":       req["role"],
		"status":     req["status"],
		"updated_at": time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "用户更新成功",
		"data": user,
	})
}

// DeleteUser 删除用户
func (h *Handlers) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	_ = userID // 避免未使用变量错误

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "用户删除成功",
	})
}

// GetSystemInfo 获取系统信息
func (h *Handlers) GetSystemInfo(c *gin.Context) {
	// 模拟系统信息
	systemInfo := map[string]interface{}{
		"version":     "1.0.0",
		"build_time":  "2024-01-01T00:00:00Z",
		"go_version":  "go1.21.0",
		"uptime":      "72h30m15s",
		"memory_usage": map[string]interface{}{
			"allocated": "256MB",
			"total":     "512MB",
			"usage":     "50%",
		},
		"cpu_usage": "15.5%",
		"goroutines": 45,
		"database": map[string]interface{}{
			"status":      "connected",
			"connections": 10,
			"max_connections": 100,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取系统信息成功",
		"data": systemInfo,
	})
}

// GetAuditLogs 获取审计日志
func (h *Handlers) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	action := c.Query("action")
	user := c.Query("user")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	_ = action    // 避免未使用变量错误
	_ = user      // 避免未使用变量错误
	_ = startTime // 避免未使用变量错误
	_ = endTime   // 避免未使用变量错误

	// 模拟审计日志数据
	logs := []map[string]interface{}{
		{
			"id":          "log-1",
			"user_id":     "user-1",
			"username":    "admin",
			"action":      "login",
			"resource":    "auth",
			"resource_id": "",
			"status":      "success",
			"ip_address":  "192.168.1.100",
			"user_agent":  "Mozilla/5.0...",
			"timestamp":   time.Now().Add(-1 * time.Hour),
		},
		{
			"id":          "log-2",
			"user_id":     "user-1",
			"username":    "admin",
			"action":      "create_alert_rule",
			"resource":    "alert_rule",
			"resource_id": "rule-123",
			"status":      "success",
			"ip_address":  "192.168.1.100",
			"user_agent":  "Mozilla/5.0...",
			"timestamp":   time.Now().Add(-30 * time.Minute),
		},
	}

	response := PaginatedResponse{
		Data: logs,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(logs),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取审计日志成功",
		"data": response,
	})
}

// ===== WebSocket相关处理器 =====

// HandleAlertWebSocket 处理告警WebSocket连接
func (h *Handlers) HandleAlertWebSocket(c *gin.Context) {
	// 这里应该实现WebSocket升级和处理逻辑
	// 由于这是一个模拟实现，我们只返回一个简单的响应
	response := map[string]interface{}{
		"message": "WebSocket endpoint for alerts",
		"type":    "alert_websocket",
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "告警WebSocket连接成功",
		"data": response,
	})
}

// HandleMetricsWebSocket 处理指标WebSocket连接
func (h *Handlers) HandleMetricsWebSocket(c *gin.Context) {
	// 这里应该实现WebSocket升级和处理逻辑
	// 由于这是一个模拟实现，我们只返回一个简单的响应
	response := map[string]interface{}{
		"message": "WebSocket endpoint for metrics",
		"type":    "metrics_websocket",
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "指标WebSocket连接成功",
		"data": response,
	})
}

// HandleLogsWebSocket 处理日志WebSocket连接
func (h *Handlers) HandleLogsWebSocket(c *gin.Context) {
	// 这里应该实现WebSocket升级和处理逻辑
	// 由于这是一个模拟实现，我们只返回一个简单的响应
	response := map[string]interface{}{
		"message": "WebSocket endpoint for logs",
		"type":    "logs_websocket",
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "日志WebSocket连接成功",
		"data": response,
	})
}

// ===== 知识库管理相关处理器 =====

// GetKnowledgeBases 获取知识库列表
func (h *Handlers) GetKnowledgeBases(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	knowledgeBases, total, err := h.aiService.ListKnowledgeBase(page, pageSize, category, search)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "get_knowledge_bases", "knowledge_base", "", "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	response := PaginatedResponse{
		Data: knowledgeBases,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    int(total),
			Pages:    int((total + int64(pageSize) - 1) / int64(pageSize)),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取知识库列表成功",
		"data": response,
	})
}

// CreateKnowledgeBase 创建知识库条目
func (h *Handlers) CreateKnowledgeBase(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req services.KnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	knowledgeBase, err := h.aiService.CreateKnowledgeBase(&req, userID)
	if err != nil {
		h.auditService.LogAuditFromContext(c, "create_knowledge_base", "knowledge_base", "", "failure", err.Error(), map[string]interface{}{
			"title": req.Title,
			"category": req.Category,
		})
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "create_knowledge_base", "knowledge_base", knowledgeBase.ID.String(), "success", "", map[string]interface{}{
		"title": req.Title,
		"category": req.Category,
	})

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "知识库条目创建成功",
		"data": knowledgeBase,
	})
}

// GetKnowledgeBase 获取知识库条目详情
func (h *Handlers) GetKnowledgeBase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	knowledgeBase, err := h.aiService.GetKnowledgeBase(id)
	if err != nil {
		if err.Error() == "knowledge base entry not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		h.auditService.LogAuditFromContext(c, "get_knowledge_base", "knowledge_base", id.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取知识库条目成功",
		"data": knowledgeBase,
	})
}

// UpdateKnowledgeBase 更新知识库条目
func (h *Handlers) UpdateKnowledgeBase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	var req services.KnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 更新知识库条目
	knowledgeBase, err := h.aiService.UpdateKnowledgeBase(id, &req)
	if err != nil {
		if err.Error() == "knowledge base entry not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		h.auditService.LogAuditFromContext(c, "update_knowledge_base", "knowledge_base", id.String(), "failure", err.Error(), map[string]interface{}{
			"title": req.Title,
			"category": req.Category,
			"updated_by": userID,
		})
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "update_knowledge_base", "knowledge_base", id.String(), "success", "", map[string]interface{}{
		"title": req.Title,
		"category": req.Category,
		"updated_by": userID,
	})

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "知识库条目更新成功",
		"data": knowledgeBase,
	})
}

// DeleteKnowledgeBase 删除知识库条目
func (h *Handlers) DeleteKnowledgeBase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	// 删除知识库条目
	err = h.aiService.DeleteKnowledgeBase(id)
	if err != nil {
		if err.Error() == "knowledge base entry not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		h.auditService.LogAuditFromContext(c, "delete_knowledge_base", "knowledge_base", id.String(), "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	h.auditService.LogAuditFromContext(c, "delete_knowledge_base", "knowledge_base", id.String(), "success", "", nil)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "知识库条目删除成功",
	})
}

// GetKnowledgeBaseStats 获取知识库统计信息
func (h *Handlers) GetKnowledgeBaseStats(c *gin.Context) {
	stats, err := h.aiService.GetKnowledgeBaseStats()
	if err != nil {
		h.auditService.LogAuditFromContext(c, "get_knowledge_base_stats", "knowledge_base", "", "failure", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取知识库统计成功",
		"data": stats,
	})
}

// ExportKnowledgeBase 导出知识库
func (h *Handlers) ExportKnowledgeBase(c *gin.Context) {
	category := c.Query("category")
	format := c.DefaultQuery("format", "markdown")

	if format != "markdown" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid format",
			Message: "Only markdown format is supported",
		})
		return
	}

	content, err := h.aiService.ExportKnowledgeBase(category)
	if err != nil {
		if err.Error() == "no knowledge base entries found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		h.auditService.LogAuditFromContext(c, "export_knowledge_base", "knowledge_base", "", "failure", err.Error(), map[string]interface{}{
			"category": category,
			"format": format,
		})
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	// 设置响应头
	filename := "knowledge_base_export.md"
	if category != "" {
		filename = fmt.Sprintf("knowledge_base_%s_export.md", category)
	}

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(content)))

	h.auditService.LogAuditFromContext(c, "export_knowledge_base", "knowledge_base", "", "success", "", map[string]interface{}{
		"category": category,
		"format": format,
		"content_length": len(content),
	})

	c.String(http.StatusOK, content)
}

// ===== API Key管理相关处理器 =====

// CreateAPIKey 创建API Key
func (h *Handlers) CreateAPIKey(c *gin.Context) {
	h.apiKeyHandler.CreateAPIKey(c)
}

// ListAPIKeys 获取API Key列表
func (h *Handlers) ListAPIKeys(c *gin.Context) {
	h.apiKeyHandler.ListAPIKeys(c)
}

// GetAPIKey 获取API Key详情
func (h *Handlers) GetAPIKey(c *gin.Context) {
	h.apiKeyHandler.GetAPIKey(c)
}

// UpdateAPIKey 更新API Key
func (h *Handlers) UpdateAPIKey(c *gin.Context) {
	h.apiKeyHandler.UpdateAPIKey(c)
}

// DeleteAPIKey 删除API Key
func (h *Handlers) DeleteAPIKey(c *gin.Context) {
	h.apiKeyHandler.DeleteAPIKey(c)
}

// GenerateAPIKey 生成API Key
func (h *Handlers) GenerateAPIKey(c *gin.Context) {
	h.apiKeyHandler.GenerateAPIKey(c)
}

// ValidateAPIKey 验证API Key
func (h *Handlers) ValidateAPIKey(c *gin.Context) {
	h.apiKeyHandler.ValidateAPIKey(c)
}

// ===== 服务发现相关处理器 =====

// CreateDiscoveryTask 创建发现任务
func (h *Handlers) CreateDiscoveryTask(c *gin.Context) {
	h.discoveryHandler.CreateDiscoveryTask(c)
}

// GetDiscoveryTask 获取发现任务详情
func (h *Handlers) GetDiscoveryTask(c *gin.Context) {
	h.discoveryHandler.GetDiscoveryTask(c)
}

// ListDiscoveryTasks 获取发现任务列表
func (h *Handlers) ListDiscoveryTasks(c *gin.Context) {
	h.discoveryHandler.ListDiscoveryTasks(c)
}

// GetDiscoveryTaskResults 获取发现任务结果
func (h *Handlers) GetDiscoveryTaskResults(c *gin.Context) {
	h.discoveryHandler.GetDiscoveryTaskResults(c)
}

// GetDiscoveryTaskProgress 获取发现任务进度
func (h *Handlers) GetDiscoveryTaskProgress(c *gin.Context) {
	h.discoveryHandler.GetDiscoveryTaskProgress(c)
}

// GetDiscoveryStats 获取发现统计信息
func (h *Handlers) GetDiscoveryStats(c *gin.Context) {
	h.discoveryHandler.GetDiscoveryStats(c)
}