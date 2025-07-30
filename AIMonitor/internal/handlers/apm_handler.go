package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ai-monitor/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APMHandler APM应用性能监控处理器
type APMHandler struct {
	apmService *services.APMService
}

// NewAPMHandler 创建APM处理器
func NewAPMHandler(apmService *services.APMService) *APMHandler {
	return &APMHandler{
		apmService: apmService,
	}
}

// GetTraces 获取链路追踪列表
// @Summary 获取链路追踪列表
// @Description 获取应用的链路追踪数据列表
// @Tags APM监控
// @Accept json
// @Produce json
// @Param service_name query string false "服务名称"
// @Param operation_name query string false "操作名称"
// @Param status query string false "状态" Enums(ok,error,timeout)
// @Param start_time query string false "开始时间" format(date-time)
// @Param end_time query string false "结束时间" format(date-time)
// @Param min_duration query int false "最小耗时(微秒)"
// @Param max_duration query int false "最大耗时(微秒)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/traces [get]
func (h *APMHandler) GetTraces(c *gin.Context) {
	// 解析查询参数
	query := services.TraceQuery{
		Service:   c.Query("service_name"),
		Operation: c.Query("operation_name"),
		Tags:      make(map[string]string),
	}

	// 解析时间范围
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			query.StartTime = startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			query.EndTime = endTime
		}
	}

	// 解析耗时范围
	if minDurationStr := c.Query("min_duration"); minDurationStr != "" {
		if minDuration, err := strconv.ParseInt(minDurationStr, 10, 64); err == nil {
			query.MinDuration = time.Duration(minDuration) * time.Microsecond
		}
	}
	if maxDurationStr := c.Query("max_duration"); maxDurationStr != "" {
		if maxDuration, err := strconv.ParseInt(maxDurationStr, 10, 64); err == nil {
			query.MaxDuration = time.Duration(maxDuration) * time.Microsecond
		}
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 设置分页限制
	query.Limit = pageSize

	traces, err := h.apmService.GetTraces(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	// 简化响应，因为实际的分页逻辑需要在服务层实现
	response := PaginatedResponse{
		Data: traces,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(traces),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetTraceDetail 获取链路追踪详情
// @Summary 获取链路追踪详情
// @Description 获取指定链路追踪的详细信息
// @Tags APM监控
// @Accept json
// @Produce json
// @Param trace_id path string true "链路追踪ID"
// @Success 200 {object} services.TraceDetail
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/traces/{trace_id} [get]
func (h *APMHandler) GetTraceDetail(c *gin.Context) {
	traceID := c.Param("trace_id")
	if traceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid trace ID",
			Message: "Trace ID is required",
		})
		return
	}

	traceDetail, err := h.apmService.GetTraceDetail(traceID)
	if err != nil {
		if err.Error() == "trace not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, traceDetail)
}

// GetServiceTopology 获取服务拓扑图
// @Summary 获取服务拓扑图
// @Description 获取应用服务间的调用拓扑关系
// @Tags APM监控
// @Accept json
// @Produce json
// @Param start_time query string false "开始时间" format(date-time)
// @Param end_time query string false "结束时间" format(date-time)
// @Param service_name query string false "服务名称过滤"
// @Success 200 {object} services.ServiceTopology
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/topology [get]
func (h *APMHandler) GetServiceTopology(c *gin.Context) {
	// 解析时间范围
	var startTime, endTime *time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	_ = c.Query("service_name")
	// 暂时不使用时间范围参数
	_ = startTime
	_ = endTime

	// 获取服务拓扑图
	topology, err := h.apmService.GetServiceMap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, topology)
}

// GetServices 获取服务列表
// @Summary 获取服务列表
// @Description 获取所有监控的服务列表
// @Tags APM监控
// @Accept json
// @Produce json
// @Success 200 {object} []services.ServiceInfo
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/services [get]
func (h *APMHandler) GetServices(c *gin.Context) {
	services, err := h.apmService.GetServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, services)
}

// GetServiceDetail 获取服务详情
// @Summary 获取服务详情
// @Description 获取指定服务的详细信息
// @Tags APM监控
// @Accept json
// @Produce json
// @Param service_name path string true "服务名称"
// @Success 200 {object} services.ServiceDetail
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/services/{service_name} [get]
func (h *APMHandler) GetServiceDetail(c *gin.Context) {
	serviceName := c.Param("service_name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid service name",
			Message: "Service name is required",
		})
		return
	}

	serviceDetail, err := h.apmService.GetServiceDetail(serviceName)
	if err != nil {
		if err.Error() == "service not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, serviceDetail)
}

// GetServicePerformance 获取服务性能指标
// @Summary 获取服务性能指标
// @Description 获取指定服务的性能指标数据
// @Tags APM监控
// @Accept json
// @Produce json
// @Param service_name path string true "服务名称"
// @Param start_time query string false "开始时间" format(date-time)
// @Param end_time query string false "结束时间" format(date-time)
// @Success 200 {object} services.ServicePerformance
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/services/{service_name}/performance [get]
func (h *APMHandler) GetServicePerformance(c *gin.Context) {
	serviceName := c.Param("service_name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid service name",
			Message: "Service name is required",
		})
		return
	}

	// 解析时间范围
	var startTime, endTime *time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	var startTimeVal, endTimeVal time.Time
	if startTime != nil {
		startTimeVal = *startTime
	}
	if endTime != nil {
		endTimeVal = *endTime
	}

	performance, err := h.apmService.GetServicePerformance(serviceName, startTimeVal, endTimeVal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// GetServiceOverview 获取服务性能概览
// @Summary 获取服务性能概览
// @Description 获取指定服务的性能概览信息
// @Tags APM监控
// @Accept json
// @Produce json
// @Param service_name path string true "服务名称"
// @Param start_time query string false "开始时间" format(date-time)
// @Param end_time query string false "结束时间" format(date-time)
// @Success 200 {object} services.ServiceOverview
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/services/{service_name}/overview [get]
func (h *APMHandler) GetServiceOverview(c *gin.Context) {
	serviceName := c.Param("service_name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid service name",
			Message: "Service name is required",
		})
		return
	}

	// 解析时间范围
	var startTime, endTime *time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	// 使用默认时间范围如果没有提供
	if startTime == nil {
		t := time.Now().Add(-24 * time.Hour)
		startTime = &t
	}
	if endTime == nil {
		t := time.Now()
		endTime = &t
	}

	overview, err := h.apmService.GetServicePerformance(serviceName, *startTime, *endTime)
	if err != nil {
		if err.Error() == "service not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetServiceList 获取服务列表
// @Summary 获取服务列表
// @Description 获取所有监控的服务列表
// @Tags APM监控
// @Accept json
// @Produce json
// @Param environment query string false "环境过滤"
// @Param language query string false "语言过滤"
// @Param status query string false "状态过滤" Enums(active,inactive,error)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/services [get]
func (h *APMHandler) GetServiceList(c *gin.Context) {
	// 解析查询参数
	_ = c.Query("environment")
	_ = c.Query("language")
	_ = c.Query("status")

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	services, err := h.apmService.GetServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	// 简化响应，实际的过滤和分页逻辑需要在服务层实现
	response := PaginatedResponse{
		Data: services,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(services),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, response)
}

// CreateService 创建服务监控
// @Summary 创建服务监控
// @Description 创建新的服务监控配置
// @Tags APM监控
// @Accept json
// @Produce json
// @Param request body CreateServiceRequest true "创建请求"
// @Success 201 {object} ServiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/services [post]
func (h *APMHandler) CreateService(c *gin.Context) {
	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 这里需要实现创建服务监控的逻辑
	// 暂时返回成功响应
	response := ServiceResponse{
		ID:          uuid.New(),
		Name:        req.Name,
		Version:     req.Version,
		Environment: req.Environment,
		Language:    req.Language,
		Framework:   req.Framework,
		Status:      "active",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateService 更新服务监控
// @Summary 更新服务监控
// @Description 更新指定的服务监控配置
// @Tags APM监控
// @Accept json
// @Produce json
// @Param id path string true "服务ID"
// @Param request body UpdateServiceRequest true "更新请求"
// @Success 200 {object} ServiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/services/{id} [put]
func (h *APMHandler) UpdateService(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 这里需要实现更新服务监控的逻辑
	// 暂时返回成功响应
	response := ServiceResponse{
		ID:          id,
		Name:        req.Name,
		Version:     req.Version,
		Environment: req.Environment,
		Language:    req.Language,
		Framework:   req.Framework,
		Status:      "active",
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// DeleteService 删除服务监控
// @Summary 删除服务监控
// @Description 删除指定的服务监控配置
// @Tags APM监控
// @Accept json
// @Produce json
// @Param id path string true "服务ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apm/services/{id} [delete]
func (h *APMHandler) DeleteService(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	// 这里需要实现删除服务监控的逻辑
	_ = id // 避免未使用变量警告

	c.Status(http.StatusNoContent)
}

// 请求和响应结构体

// CreateServiceRequest 创建服务监控请求
type CreateServiceRequest struct {
	Name        string            `json:"name" binding:"required" example:"user-service"`
	Version     string            `json:"version" example:"1.0.0"`
	Environment string            `json:"environment" example:"production"`
	Language    string            `json:"language" example:"go"`
	Framework   string            `json:"framework" example:"gin"`
	Tags        map[string]string `json:"tags" example:"{\"team\":\"backend\",\"owner\":\"john\"}"`
}

// UpdateServiceRequest 更新服务监控请求
type UpdateServiceRequest struct {
	Name        string            `json:"name" example:"user-service"`
	Version     string            `json:"version" example:"1.0.1"`
	Environment string            `json:"environment" example:"production"`
	Language    string            `json:"language" example:"go"`
	Framework   string            `json:"framework" example:"gin"`
	Tags        map[string]string `json:"tags" example:"{\"team\":\"backend\",\"owner\":\"john\"}"`
}

// ServiceResponse 服务监控响应
type ServiceResponse struct {
	ID          uuid.UUID         `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string            `json:"name" example:"user-service"`
	Version     string            `json:"version" example:"1.0.0"`
	Environment string            `json:"environment" example:"production"`
	Language    string            `json:"language" example:"go"`
	Framework   string            `json:"framework" example:"gin"`
	Status      string            `json:"status" example:"active"`
	LastSeen    *string           `json:"last_seen,omitempty" example:"2024-01-01T12:00:00Z"`
	Tags        map[string]string `json:"tags,omitempty" example:"{\"team\":\"backend\",\"owner\":\"john\"}"`
	CreatedAt   string            `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   string            `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}