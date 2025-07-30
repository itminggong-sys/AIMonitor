package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ai-monitor/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContainerHandler 容器监控处理器
type ContainerHandler struct {
	containerService *services.ContainerService
}

// NewContainerHandler 创建容器监控处理器
func NewContainerHandler(containerService *services.ContainerService) *ContainerHandler {
	return &ContainerHandler{
		containerService: containerService,
	}
}

// GetDockerContainers 获取Docker容器列表
// @Summary 获取Docker容器列表
// @Description 获取Docker容器监控列表
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param status query string false "容器状态" Enums(running,stopped,paused,exited,error)
// @Param image query string false "镜像名称过滤"
// @Param name query string false "容器名称过滤"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/docker [get]
func (h *ContainerHandler) GetDockerContainers(c *gin.Context) {
	// 解析查询参数
	_ = c.Query("status")
	_ = c.Query("image")
	_ = c.Query("name")

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	containers, err := h.containerService.GetDockerContainers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	// 简化响应，实际的过滤和分页逻辑需要在服务层实现
	response := PaginatedResponse{
		Data: containers,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(containers),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetDockerContainerDetail 获取Docker容器详情
// @Summary 获取Docker容器详情
// @Description 获取指定Docker容器的详细信息
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param container_id path string true "容器ID"
// @Success 200 {object} services.DockerContainerInfo
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/docker/{container_id} [get]
func (h *ContainerHandler) GetDockerContainerDetail(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid container ID",
			Message: "Container ID is required",
		})
		return
	}

	container, err := h.containerService.GetDockerContainer(containerID)
	if err != nil {
		if err.Error() == "container not found" {
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

	c.JSON(http.StatusOK, container)
}

// GetKubernetesPods 获取Kubernetes Pod列表
// @Summary 获取Kubernetes Pod列表
// @Description 获取Kubernetes Pod监控列表
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param namespace query string false "命名空间过滤"
// @Param node_name query string false "节点名称过滤"
// @Param status query string false "Pod状态" Enums(running,pending,succeeded,failed,unknown)
// @Param label_selector query string false "标签选择器"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/kubernetes/pods [get]
func (h *ContainerHandler) GetKubernetesPods(c *gin.Context) {
	// 解析查询参数
	namespace := c.Query("namespace")
	_ = c.Query("node_name")
	_ = c.Query("status")
	_ = c.Query("label_selector")

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	pods, err := h.containerService.GetKubernetesPods(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	// 简化响应，实际的过滤和分页逻辑需要在服务层实现
	response := PaginatedResponse{
		Data: pods,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(pods),
			Pages:    1,
		},
	}

	c.JSON(http.StatusOK, response)
}

// ListContainerMonitors 获取容器监控列表
// @Summary 获取容器监控列表
// @Description 获取所有容器监控实例列表
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type query string false "容器类型" Enums(docker,kubernetes)
// @Param status query string false "状态" Enums(active,inactive,error)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/monitors [get]
func (h *ContainerHandler) ListContainerMonitors(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	_ = c.Query("type")
	_ = c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 这里需要实现获取容器监控列表的逻辑
	// 暂时返回空列表
	response := PaginatedResponse{
		Data: []interface{}{},
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    0,
			Pages:    0,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetContainerMonitor 获取容器监控详情
// @Summary 获取容器监控详情
// @Description 获取指定容器监控实例的详细信息
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param id path string true "监控实例ID"
// @Success 200 {object} services.ContainerMonitorInfo
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/monitors/{id} [get]
func (h *ContainerHandler) GetContainerMonitor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	// 这里需要实现获取容器监控详情的逻辑
	// 暂时返回示例响应
	now := time.Now()
	response := map[string]interface{}{
		"id":         id,
		"name":       "示例容器监控",
		"type":       "docker",
		"status":     "active",
		"created_at": now.Add(-24 * time.Hour),
		"updated_at": now,
	}

	c.JSON(http.StatusOK, response)
}

// GetKubernetesNodes 获取Kubernetes节点列表
// @Summary 获取Kubernetes节点列表
// @Description 获取Kubernetes节点监控列表
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param status query string false "节点状态" Enums(ready,notready,unknown)
// @Param role query string false "节点角色" Enums(master,worker)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/kubernetes/nodes [get]
func (h *ContainerHandler) GetKubernetesNodes(c *gin.Context) {
	// 解析查询参数
	_ = c.Query("status")
	_ = c.Query("role")

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	nodes, err := h.containerService.GetKubernetesNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	// 简化响应，实际的过滤和分页逻辑需要在服务层实现
	totalPages := (len(nodes) + pageSize - 1) / pageSize
	response := PaginatedResponse{
		Data: nodes,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    len(nodes),
			Pages:    totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetKubernetesNamespaces 获取Kubernetes命名空间列表
// @Summary 获取Kubernetes命名空间列表
// @Description 获取Kubernetes命名空间监控列表
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param status query string false "命名空间状态" Enums(active,terminating)
// @Success 200 {object} []services.KubernetesNamespace
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/kubernetes/namespaces [get]
func (h *ContainerHandler) GetKubernetesNamespaces(c *gin.Context) {
	_ = c.Query("status")

	namespaces, err := h.containerService.GetKubernetesNamespaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, namespaces)
}

// GetClusterMetrics 获取集群指标
// @Summary 获取集群指标
// @Description 获取Kubernetes集群的整体指标信息
// @Tags 容器监控
// @Accept json
// @Produce json
// @Success 200 {object} services.ClusterMetrics
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/kubernetes/cluster/metrics [get]
func (h *ContainerHandler) GetClusterMetrics(c *gin.Context) {
	metrics, err := h.containerService.GetKubernetesClusterMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetResourceUsage 获取资源使用情况
// @Summary 获取资源使用情况
// @Description 获取容器或Pod的资源使用情况统计
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param platform query string true "平台类型" Enums(docker,kubernetes)
// @Param resource_type query string false "资源类型" Enums(cpu,memory,disk,network)
// @Param namespace query string false "命名空间(仅Kubernetes)"
// @Param start_time query string false "开始时间" format(date-time)
// @Param end_time query string false "结束时间" format(date-time)
// @Success 200 {object} services.ResourceUsage
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/resource-usage [get]
func (h *ContainerHandler) GetResourceUsage(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid platform",
			Message: "Platform is required (docker or kubernetes)",
		})
		return
	}

	resourceType := c.Query("resource_type")
	namespace := c.Query("namespace")

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

	usage, err := h.containerService.GetResourceUsage(platform, resourceType, namespace, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// CreateContainerMonitor 创建容器监控
// @Summary 创建容器监控
// @Description 创建新的容器监控配置
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param request body CreateContainerMonitorRequest true "创建请求"
// @Success 201 {object} ContainerMonitorResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/monitors [post]
func (h *ContainerHandler) CreateContainerMonitor(c *gin.Context) {
	var req CreateContainerMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 这里需要实现创建容器监控的逻辑
	// 暂时返回成功响应
	response := ContainerMonitorResponse{
		ID:          uuid.New(),
		ContainerID: req.ContainerID,
		Name:        req.Name,
		Image:       req.Image,
		Platform:    req.Platform,
		Status:      "running",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateContainerMonitor 更新容器监控
// @Summary 更新容器监控
// @Description 更新指定的容器监控配置
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param id path string true "监控ID"
// @Param request body UpdateContainerMonitorRequest true "更新请求"
// @Success 200 {object} ContainerMonitorResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/monitors/{id} [put]
func (h *ContainerHandler) UpdateContainerMonitor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	var req UpdateContainerMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 这里需要实现更新容器监控的逻辑
	// 暂时返回成功响应
	response := ContainerMonitorResponse{
		ID:        id,
		Name:      req.Name,
		Status:    "running",
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// DeleteContainerMonitor 删除容器监控
// @Summary 删除容器监控
// @Description 删除指定的容器监控配置
// @Tags 容器监控
// @Accept json
// @Produce json
// @Param id path string true "监控ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/containers/monitors/{id} [delete]
func (h *ContainerHandler) DeleteContainerMonitor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	// 这里需要实现删除容器监控的逻辑
	_ = id // 避免未使用变量警告

	c.Status(http.StatusNoContent)
}

// 请求和响应结构体

// CreateContainerMonitorRequest 创建容器监控请求
type CreateContainerMonitorRequest struct {
	ContainerID string            `json:"container_id" binding:"required" example:"abc123def456"`
	Name        string            `json:"name" binding:"required" example:"nginx-web"`
	Image       string            `json:"image" binding:"required" example:"nginx:1.21"`
	Platform    string            `json:"platform" binding:"required,oneof=docker kubernetes" example:"docker"`
	Namespace   string            `json:"namespace" example:"default"`
	Labels      map[string]string `json:"labels" example:"{\"app\":\"nginx\",\"env\":\"prod\"}"`
	Annotations map[string]string `json:"annotations" example:"{\"description\":\"Web server\"}"`
}

// UpdateContainerMonitorRequest 更新容器监控请求
type UpdateContainerMonitorRequest struct {
	Name        string            `json:"name" example:"nginx-web"`
	Labels      map[string]string `json:"labels" example:"{\"app\":\"nginx\",\"env\":\"prod\"}"`
	Annotations map[string]string `json:"annotations" example:"{\"description\":\"Web server\"}"`
}

// ContainerMonitorResponse 容器监控响应
type ContainerMonitorResponse struct {
	ID          uuid.UUID         `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ContainerID string            `json:"container_id" example:"abc123def456"`
	Name        string            `json:"name" example:"nginx-web"`
	Image       string            `json:"image" example:"nginx:1.21"`
	Platform    string            `json:"platform" example:"docker"`
	Namespace   string            `json:"namespace,omitempty" example:"default"`
	Status      string            `json:"status" example:"running"`
	StartedAt   *string           `json:"started_at,omitempty" example:"2024-01-01T12:00:00Z"`
	Labels      map[string]string `json:"labels,omitempty" example:"{\"app\":\"nginx\",\"env\":\"prod\"}"`
	Annotations map[string]string `json:"annotations,omitempty" example:"{\"description\":\"Web server\"}"`
	CreatedAt   string            `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   string            `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}