package handlers

import (
	"net/http"
	"strconv"

	"ai-monitor/internal/services"

	"github.com/gin-gonic/gin"
)

// Response 通用响应结构体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// DiscoveryHandler 服务发现处理器
type DiscoveryHandler struct {
	discoveryService *services.DiscoveryService
}

// NewDiscoveryHandler 创建服务发现处理器
func NewDiscoveryHandler(discoveryService *services.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{
		discoveryService: discoveryService,
	}
}

// CreateDiscoveryTask 创建发现任务
// @Summary 创建服务发现任务
// @Description 创建一个新的服务发现任务，支持网络扫描、SSH扫描和Agent发现
// @Tags Discovery
// @Accept json
// @Produce json
// @Param request body services.CreateDiscoveryTaskRequest true "发现任务请求"
// @Success 200 {object} Response{data=services.DiscoveryTask} "创建成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/discovery/tasks [post]
func (h *DiscoveryHandler) CreateDiscoveryTask(c *gin.Context) {
	var req services.CreateDiscoveryTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	task, err := h.discoveryService.CreateDiscoveryTask(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "创建发现任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "发现任务创建成功",
		Data:    task,
	})
}

// GetDiscoveryTask 获取发现任务详情
// @Summary 获取发现任务详情
// @Description 根据任务ID获取发现任务的详细信息和执行结果
// @Tags Discovery
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} Response{data=services.DiscoveryTask} "获取成功"
// @Failure 404 {object} Response "任务不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/discovery/tasks/{id} [get]
func (h *DiscoveryHandler) GetDiscoveryTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "任务ID不能为空",
		})
		return
	}

	task, err := h.discoveryService.GetDiscoveryTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "发现任务不存在: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取发现任务成功",
		Data:    task,
	})
}

// ListDiscoveryTasks 获取发现任务列表
// @Summary 获取发现任务列表
// @Description 获取所有发现任务的列表，支持分页和状态过滤
// @Tags Discovery
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param status query string false "任务状态过滤"
// @Success 200 {object} Response{data=[]services.DiscoveryTask} "获取成功"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/discovery/tasks [get]
func (h *DiscoveryHandler) ListDiscoveryTasks(c *gin.Context) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	status := c.Query("status")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	tasks, err := h.discoveryService.ListDiscoveryTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取发现任务列表失败: " + err.Error(),
		})
		return
	}

	// 状态过滤
	if status != "" {
		filteredTasks := make([]*services.DiscoveryTask, 0)
		for _, task := range tasks {
			if task.Status == status {
				filteredTasks = append(filteredTasks, task)
			}
		}
		tasks = filteredTasks
	}

	// 分页处理
	total := len(tasks)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		tasks = []*services.DiscoveryTask{}
	} else {
		if end > total {
			end = total
		}
		tasks = tasks[start:end]
	}

	response := map[string]interface{}{
		"tasks": tasks,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取发现任务列表成功",
		Data:    response,
	})
}

// GetDiscoveryTaskResults 获取发现任务结果
// @Summary 获取发现任务结果
// @Description 获取指定发现任务的执行结果，包括发现的服务和注册状态
// @Tags Discovery
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param status query string false "结果状态过滤"
// @Success 200 {object} Response{data=[]services.DiscoveryResult} "获取成功"
// @Failure 404 {object} Response "任务不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/discovery/tasks/{id}/results [get]
func (h *DiscoveryHandler) GetDiscoveryTaskResults(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "任务ID不能为空",
		})
		return
	}

	task, err := h.discoveryService.GetDiscoveryTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "发现任务不存在: " + err.Error(),
		})
		return
	}

	results := task.Results
	status := c.Query("status")

	// 状态过滤
	if status != "" {
		filteredResults := make([]services.DiscoveryResult, 0)
		for _, result := range results {
			if result.Status == status {
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取发现任务结果成功",
		Data:    results,
	})
}

// GetDiscoveryTaskProgress 获取发现任务进度
// @Summary 获取发现任务进度
// @Description 获取指定发现任务的执行进度和状态信息
// @Tags Discovery
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} Response{data=map[string]interface{}} "获取成功"
// @Failure 404 {object} Response "任务不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/discovery/tasks/{id}/progress [get]
func (h *DiscoveryHandler) GetDiscoveryTaskProgress(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "任务ID不能为空",
		})
		return
	}

	task, err := h.discoveryService.GetDiscoveryTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "发现任务不存在: " + err.Error(),
		})
		return
	}

	progress := map[string]interface{}{
		"task_id":     task.ID,
		"status":      task.Status,
		"progress":    task.Progress,
		"message":     task.Message,
		"created_at":  task.CreatedAt,
		"updated_at":  task.UpdatedAt,
		"completed_at": task.CompletedAt,
		"total_targets": len(task.Targets),
		"total_results": len(task.Results),
		"discovered_count": h.countResultsByStatus(task.Results, "discovered"),
		"registered_count": h.countResultsByStatus(task.Results, "registered"),
		"failed_count":     h.countResultsByStatus(task.Results, "failed"),
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取发现任务进度成功",
		Data:    progress,
	})
}

// countResultsByStatus 统计指定状态的结果数量
func (h *DiscoveryHandler) countResultsByStatus(results []services.DiscoveryResult, status string) int {
	count := 0
	for _, result := range results {
		if result.Status == status {
			count++
		}
	}
	return count
}

// GetDiscoveryStats 获取发现统计信息
// @Summary 获取发现统计信息
// @Description 获取服务发现的统计信息，包括任务数量、发现的服务类型等
// @Tags Discovery
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=map[string]interface{}} "获取成功"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /api/v1/discovery/stats [get]
func (h *DiscoveryHandler) GetDiscoveryStats(c *gin.Context) {
	tasks, err := h.discoveryService.ListDiscoveryTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取发现统计信息失败: " + err.Error(),
		})
		return
	}

	// 统计信息
	stats := map[string]interface{}{
		"total_tasks":      len(tasks),
		"pending_tasks":    h.countTasksByStatus(tasks, "pending"),
		"running_tasks":    h.countTasksByStatus(tasks, "running"),
		"completed_tasks":  h.countTasksByStatus(tasks, "completed"),
		"failed_tasks":     h.countTasksByStatus(tasks, "failed"),
		"task_types":       h.getTaskTypeStats(tasks),
		"service_types":    h.getServiceTypeStats(tasks),
		"discovery_summary": h.getDiscoverySummary(tasks),
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取发现统计信息成功",
		Data:    stats,
	})
}

// countTasksByStatus 统计指定状态的任务数量
func (h *DiscoveryHandler) countTasksByStatus(tasks []*services.DiscoveryTask, status string) int {
	count := 0
	for _, task := range tasks {
		if task.Status == status {
			count++
		}
	}
	return count
}

// getTaskTypeStats 获取任务类型统计
func (h *DiscoveryHandler) getTaskTypeStats(tasks []*services.DiscoveryTask) map[string]int {
	stats := make(map[string]int)
	for _, task := range tasks {
		stats[task.Type]++
	}
	return stats
}

// getServiceTypeStats 获取服务类型统计
func (h *DiscoveryHandler) getServiceTypeStats(tasks []*services.DiscoveryTask) map[string]int {
	stats := make(map[string]int)
	for _, task := range tasks {
		for _, result := range task.Results {
			if result.Service != "" {
				stats[result.Service]++
			}
		}
	}
	return stats
}

// getDiscoverySummary 获取发现摘要
func (h *DiscoveryHandler) getDiscoverySummary(tasks []*services.DiscoveryTask) map[string]int {
	totalResults := 0
	discovered := 0
	registered := 0
	failed := 0

	for _, task := range tasks {
		totalResults += len(task.Results)
		for _, result := range task.Results {
			switch result.Status {
			case "discovered":
				discovered++
			case "registered":
				registered++
			case "failed":
				failed++
			}
		}
	}

	return map[string]int{
		"total_results": totalResults,
		"discovered":    discovered,
		"registered":    registered,
		"failed":        failed,
	}
}