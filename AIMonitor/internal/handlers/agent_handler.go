package handlers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"ai-monitor/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AgentHandler Agent管理处理器
type AgentHandler struct {
	agentService *services.AgentService
}

// NewAgentHandler 创建Agent管理处理器
func NewAgentHandler(agentService *services.AgentService) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

// CreateAgent 创建Agent
// @Summary 创建Agent
// @Description 创建新的监控Agent
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param request body services.CreateAgentRequest true "创建请求"
// @Success 201 {object} services.Agent
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents [post]
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req services.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	agent, err := h.agentService.CreateAgent(&req)
	if err != nil {
		if err.Error() == "agent name already exists" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
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

	c.JSON(http.StatusCreated, agent)
}

// GetAgent 获取Agent信息
// @Summary 获取Agent信息
// @Description 获取指定Agent的详细信息
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} services.Agent
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/{id} [get]
func (h *AgentHandler) GetAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	agent, err := h.agentService.GetAgent(id)
	if err != nil {
		if err.Error() == "agent not found" {
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

	c.JSON(http.StatusOK, agent)
}

// UpdateAgent 更新Agent信息
// @Summary 更新Agent信息
// @Description 更新指定Agent的信息
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param request body services.UpdateAgentRequest true "更新请求"
// @Success 200 {object} services.Agent
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/{id} [put]
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	var req services.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	agent, err := h.agentService.UpdateAgent(id, &req)
	if err != nil {
		if err.Error() == "agent not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		if err.Error() == "agent name already exists" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
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

	c.JSON(http.StatusOK, agent)
}

// DeleteAgent 删除Agent
// @Summary 删除Agent
// @Description 删除指定的Agent
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/{id} [delete]
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	err = h.agentService.DeleteAgent(id)
	if err != nil {
		if err.Error() == "agent not found" {
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

	c.Status(http.StatusNoContent)
}

// ListAgents 获取Agent列表
// @Summary 获取Agent列表
// @Description 获取Agent列表，支持分页和过滤
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type query string false "Agent类型" Enums(node_exporter,mysql_exporter,redis_exporter,kafka_exporter,custom)
// @Param status query string false "状态" Enums(pending,active,inactive,error)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents [get]
func (h *AgentHandler) ListAgents(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	agentType := c.Query("type")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	agents, total, err := h.agentService.ListAgents(page, pageSize, agentType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	response := PaginatedResponse{
		Data: agents,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    int(total),
			Pages:    totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// CreateDeployment 创建部署
// @Summary 创建Agent部署
// @Description 创建新的Agent部署任务
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param request body services.CreateDeploymentRequest true "创建请求"
// @Success 201 {object} services.AgentDeployment
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/deployments [post]
func (h *AgentHandler) CreateDeployment(c *gin.Context) {
	var req services.CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 获取用户ID（从JWT token或session中获取）
	userID := uuid.New() // 这里应该从认证中间件获取实际用户ID

	deployment, err := h.agentService.CreateDeployment(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, deployment)
}

// GetDeployment 获取部署信息
// @Summary 获取部署信息
// @Description 获取指定部署的详细信息
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param id path string true "部署ID"
// @Success 200 {object} services.AgentDeployment
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/deployments/{id} [get]
func (h *AgentHandler) GetDeployment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	deployment, err := h.agentService.GetDeployment(id)
	if err != nil {
		if err.Error() == "deployment not found" {
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

	c.JSON(http.StatusOK, deployment)
}

// ListDeployments 获取部署列表
// @Summary 获取部署列表
// @Description 获取Agent部署列表，支持分页和过滤
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "状态" Enums(pending,running,completed,failed,partial)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/deployments [get]
func (h *AgentHandler) ListDeployments(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	deployments, total, err := h.agentService.ListDeployments(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	response := PaginatedResponse{
		Data: deployments,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    int(total),
			Pages:    totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetAgentPackages 获取Agent安装包列表
// @Summary 获取Agent安装包列表
// @Description 获取可用的Agent安装包列表
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param type query string false "Agent类型" Enums(node_exporter,mysql_exporter,redis_exporter,kafka_exporter,custom)
// @Param platform query string false "平台" Enums(linux,windows,macos)
// @Param architecture query string false "架构" Enums(amd64,arm64,386)
// @Success 200 {object} []services.AgentPackage
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/packages [get]
func (h *AgentHandler) GetAgentPackages(c *gin.Context) {
	agentType := c.Query("type")
	platform := c.Query("platform")
	architecture := c.Query("architecture")

	packages, err := h.agentService.GetAgentPackages(agentType, platform, architecture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, packages)
}

// ProcessHeartbeat 处理Agent心跳
// @Summary 处理Agent心跳
// @Description 接收并处理Agent的心跳数据
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param request body services.AgentHeartbeat true "心跳数据"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/heartbeat [post]
func (h *AgentHandler) ProcessHeartbeat(c *gin.Context) {
	var heartbeat services.AgentHeartbeat
	if err := c.ShouldBindJSON(&heartbeat); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 设置心跳时间戳
	heartbeat.Timestamp = time.Now()

	err := h.agentService.ProcessHeartbeat(&heartbeat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"status": "success",
		"message": "Heartbeat processed successfully",
	})
}

// GetAgentConfig 获取Agent配置
// @Summary 获取Agent配置
// @Description 获取指定Agent的配置信息
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} services.AgentConfig
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/{id}/config [get]
func (h *AgentHandler) GetAgentConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	config, err := h.agentService.GetAgentConfig(id)
	if err != nil {
		if err.Error() == "agent not found" {
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

	c.JSON(http.StatusOK, config)
}

// UpdateAgentConfig 更新Agent配置
// @Summary 更新Agent配置
// @Description 更新指定Agent的配置信息
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param request body map[string]interface{} true "配置数据"
// @Success 200 {object} services.AgentConfig
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/{id}/config [put]
func (h *AgentHandler) UpdateAgentConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	var configData map[string]interface{}
	if err := c.ShouldBindJSON(&configData); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	config, err := h.agentService.UpdateAgentConfig(id, configData)
	if err != nil {
		if err.Error() == "agent not found" {
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

	c.JSON(http.StatusOK, config)
}

// DownloadAgent 下载Agent安装包
// @Summary 下载Agent安装包
// @Description 下载指定的Agent安装包
// @Tags Agent管理
// @Accept json
// @Produce application/octet-stream
// @Param type path string true "Agent类型" Enums(windows,linux,redis,mysql,docker,kafka,apache,nginx,postgresql,elasticsearch,rabbitmq,hyperv,vmware,apm)
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/download/{type} [get]
func (h *AgentHandler) DownloadAgent(c *gin.Context) {
	agentType := c.Param("type")

	if agentType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid parameters",
			Message: "Agent type is required",
		})
		return
	}

	// 根据Agent类型确定文件路径和名称
	var filePath, fileName string
	switch agentType {
	case "windows":
		filePath = "./agents/build/windows/aimonitor-windows-agent.exe"
		fileName = "aimonitor-windows-agent.exe"
	case "linux":
		filePath = "./agents/build/linux/aimonitor-linux-agent"
		fileName = "aimonitor-linux-agent"
	case "redis":
		filePath = "./agents/build/redis/aimonitor-redis-agent.exe"
		fileName = "aimonitor-redis-agent.exe"
	case "mysql":
		filePath = "./agents/build/mysql/aimonitor-mysql-agent.exe"
		fileName = "aimonitor-mysql-agent.exe"
	case "docker":
		filePath = "./agents/build/docker/aimonitor-docker-agent.exe"
		fileName = "aimonitor-docker-agent.exe"
	case "kafka":
		filePath = "./agents/build/kafka/aimonitor-kafka-agent.exe"
		fileName = "aimonitor-kafka-agent.exe"
	case "apache":
		filePath = "./agents/build/apache/aimonitor-apache-agent.exe"
		fileName = "aimonitor-apache-agent.exe"
	case "nginx":
		filePath = "./agents/build/nginx/aimonitor-nginx-agent.exe"
		fileName = "aimonitor-nginx-agent.exe"
	case "postgresql":
		filePath = "./agents/build/postgresql/aimonitor-postgresql-agent.exe"
		fileName = "aimonitor-postgresql-agent.exe"
	case "elasticsearch":
		filePath = "./agents/build/elasticsearch/aimonitor-elasticsearch-agent.exe"
		fileName = "aimonitor-elasticsearch-agent.exe"
	case "rabbitmq":
		filePath = "./agents/build/rabbitmq/aimonitor-rabbitmq-agent.exe"
		fileName = "aimonitor-rabbitmq-agent.exe"
	case "hyperv":
		filePath = "./agents/build/hyperv/aimonitor-hyperv-agent.exe"
		fileName = "aimonitor-hyperv-agent.exe"
	case "vmware":
		filePath = "./agents/build/vmware/aimonitor-vmware-agent.exe"
		fileName = "aimonitor-vmware-agent.exe"
	case "apm":
		filePath = "./agents/build/apm/aimonitor-apm-agent.exe"
		fileName = "aimonitor-apm-agent.exe"
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid agent type",
			Message: "Supported types: windows, linux, redis, mysql, docker, kafka, apache, nginx, postgresql, elasticsearch, rabbitmq, hyperv, vmware, apm",
		})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Not Found",
			Message: "Agent package not found",
		})
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件
	c.File(filePath)
}