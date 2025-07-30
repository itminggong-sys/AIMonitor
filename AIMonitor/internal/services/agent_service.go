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

// AgentService Agent管理服务
type AgentService struct {
	db           *gorm.DB
	cacheManager *cache.CacheManager
	config       *config.Config
}

// NewAgentService 创建Agent管理服务
func NewAgentService(db *gorm.DB, cacheManager *cache.CacheManager, config *config.Config) *AgentService {
	return &AgentService{
		db:           db,
		cacheManager: cacheManager,
		config:       config,
	}
}

// Agent Agent信息
type Agent struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Version      string                 `json:"version"`
	Platform     string                 `json:"platform"`
	Architecture string                 `json:"architecture"`
	Hostname     string                 `json:"hostname"`
	IPAddress    string                 `json:"ip_address"`
	Port         int                    `json:"port"`
	Status       string                 `json:"status"`
	LastSeen     time.Time              `json:"last_seen"`
	Config       map[string]interface{} `json:"config"`
	Metrics      AgentMetrics           `json:"metrics"`
	Tags         map[string]string      `json:"tags"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// AgentMetrics Agent指标
type AgentMetrics struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkIO    float64 `json:"network_io"`
	Uptime       float64 `json:"uptime"`
	CollectRate  float64 `json:"collect_rate"`
	ErrorRate    float64 `json:"error_rate"`
	LastCollect  time.Time `json:"last_collect"`
}

// AgentConfig Agent配置
type AgentConfig struct {
	ID              uuid.UUID              `json:"id"`
	AgentID         uuid.UUID              `json:"agent_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Config          map[string]interface{} `json:"config"`
	Version         string                 `json:"version"`
	Status          string                 `json:"status"`
	AppliedAt       *time.Time             `json:"applied_at"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// AgentDeployment Agent部署信息
type AgentDeployment struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	AgentType    string                 `json:"agent_type"`
	Version      string                 `json:"version"`
	Targets      []DeploymentTarget     `json:"targets"`
	Config       map[string]interface{} `json:"config"`
	Status       string                 `json:"status"`
	Progress     DeploymentProgress     `json:"progress"`
	CreatedBy    uuid.UUID              `json:"created_by"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// DeploymentTarget 部署目标
type DeploymentTarget struct {
	ID          uuid.UUID              `json:"id"`
	Hostname    string                 `json:"hostname"`
	IPAddress   string                 `json:"ip_address"`
	Platform    string                 `json:"platform"`
	Architecture string                `json:"architecture"`
	Credentials map[string]interface{} `json:"credentials"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	DeployedAt  *time.Time             `json:"deployed_at"`
}

// DeploymentProgress 部署进度
type DeploymentProgress struct {
	Total     int `json:"total"`
	Success   int `json:"success"`
	Failed    int `json:"failed"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Percent   float64 `json:"percent"`
}

// AgentPackage Agent安装包
type AgentPackage struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Version      string    `json:"version"`
	Platform     string    `json:"platform"`
	Architecture string    `json:"architecture"`
	Size         int64     `json:"size"`
	Checksum     string    `json:"checksum"`
	DownloadURL  string    `json:"download_url"`
	Description  string    `json:"description"`
	ReleaseNotes string    `json:"release_notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateAgentRequest 创建Agent请求
type CreateAgentRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Type         string                 `json:"type" binding:"required,oneof=node_exporter mysql_exporter redis_exporter kafka_exporter custom"`
	Version      string                 `json:"version" binding:"required"`
	Platform     string                 `json:"platform" binding:"required,oneof=linux windows macos"`
	Architecture string                 `json:"architecture" binding:"required,oneof=amd64 arm64 386"`
	Hostname     string                 `json:"hostname" binding:"required"`
	IPAddress    string                 `json:"ip_address" binding:"required"`
	Port         int                    `json:"port"`
	Config       map[string]interface{} `json:"config"`
	Tags         map[string]string      `json:"tags"`
}

// UpdateAgentRequest 更新Agent请求
type UpdateAgentRequest struct {
	Name      string                 `json:"name"`
	Version   string                 `json:"version"`
	Hostname  string                 `json:"hostname"`
	IPAddress string                 `json:"ip_address"`
	Port      *int                   `json:"port"`
	Config    map[string]interface{} `json:"config"`
	Tags      map[string]string      `json:"tags"`
}

// CreateDeploymentRequest 创建部署请求
type CreateDeploymentRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	AgentType   string                 `json:"agent_type" binding:"required"`
	Version     string                 `json:"version" binding:"required"`
	Targets     []DeploymentTarget     `json:"targets" binding:"required,min=1"`
	Config      map[string]interface{} `json:"config"`
}

// AgentHeartbeat Agent心跳信息
type AgentHeartbeat struct {
	AgentID   uuid.UUID    `json:"agent_id"`
	Timestamp time.Time    `json:"timestamp"`
	Metrics   AgentMetrics `json:"metrics"`
	Status    string       `json:"status"`
	Message   string       `json:"message"`
}

// CreateAgent 创建Agent
func (s *AgentService) CreateAgent(req *CreateAgentRequest) (*Agent, error) {
	// 检查Agent名称是否已存在
	var existingAgent models.MonitoringTarget
	if err := s.db.Where("name = ? AND type = 'agent'", req.Name).First(&existingAgent).Error; err == nil {
		return nil, errors.New("agent name already exists")
	}

	// 序列化配置和标签
	configJSON, _ := json.Marshal(req.Config)
	tagsJSON, _ := json.Marshal(req.Tags)

	// 创建监控目标记录
	target := models.MonitoringTarget{
		Name:        req.Name,
		Type:        "agent",
		Platform:    req.Platform,
		Address:     req.IPAddress,
		Port:        req.Port,
		Credentials: string(configJSON),
		Labels:      string(tagsJSON),
		Status:      "pending",
	}

	if err := s.db.Create(&target).Error; err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// 转换为Agent响应
	agent := &Agent{
		ID:           target.ID,
		Name:         target.Name,
		Type:         req.Type,
		Version:      req.Version,
		Platform:     req.Platform,
		Architecture: req.Architecture,
		Hostname:     req.Hostname,
		IPAddress:    req.IPAddress,
		Port:         req.Port,
		Status:       target.Status,
		Config:       req.Config,
		Tags:         req.Tags,
		CreatedAt:    target.CreatedAt,
		UpdatedAt:    target.UpdatedAt,
	}

	return agent, nil
}

// GetAgent 获取Agent信息
func (s *AgentService) GetAgent(agentID uuid.UUID) (*Agent, error) {
	var target models.MonitoringTarget
	if err := s.db.Where("id = ? AND type = 'agent'", agentID).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("agent not found")
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// 解析配置和标签
	var config map[string]interface{}
	var tags map[string]string
	json.Unmarshal([]byte(target.Credentials), &config)
	json.Unmarshal([]byte(target.Labels), &tags)

	// 获取Agent指标
	metrics, _ := s.getAgentMetrics(agentID)

	agent := &Agent{
		ID:        target.ID,
		Name:      target.Name,
		Type:      "agent",
		Platform:  target.Platform,
		IPAddress: target.Address,
		Port:      target.Port,
		Status:    target.Status,
		LastSeen:  *target.LastSeen,
		Config:    config,
		Tags:      tags,
		Metrics:   *metrics,
		CreatedAt: target.CreatedAt,
		UpdatedAt: target.UpdatedAt,
	}

	return agent, nil
}

// UpdateAgent 更新Agent信息
func (s *AgentService) UpdateAgent(agentID uuid.UUID, req *UpdateAgentRequest) (*Agent, error) {
	var target models.MonitoringTarget
	if err := s.db.Where("id = ? AND type = 'agent'", agentID).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("agent not found")
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// 更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		// 检查名称是否已存在
		var existingAgent models.MonitoringTarget
		if err := s.db.Where("name = ? AND type = 'agent' AND id != ?", req.Name, agentID).First(&existingAgent).Error; err == nil {
			return nil, errors.New("agent name already exists")
		}
		updates["name"] = req.Name
	}
	if req.IPAddress != "" {
		updates["address"] = req.IPAddress
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		updates["credentials"] = string(configJSON)
	}
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		updates["labels"] = string(tagsJSON)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&target).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update agent: %w", err)
		}
	}

	// 重新获取Agent信息
	return s.GetAgent(agentID)
}

// DeleteAgent 删除Agent
func (s *AgentService) DeleteAgent(agentID uuid.UUID) error {
	var target models.MonitoringTarget
	if err := s.db.Where("id = ? AND type = 'agent'", agentID).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("agent not found")
		}
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// 软删除Agent
	if err := s.db.Delete(&target).Error; err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return nil
}

// ListAgents 获取Agent列表
func (s *AgentService) ListAgents(page, pageSize int, agentType, status string) ([]*Agent, int64, error) {
	query := s.db.Model(&models.MonitoringTarget{}).Where("type = 'agent'")

	// 类型过滤
	if agentType != "" {
		query = query.Where("platform = ?", agentType)
	}

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count agents: %w", err)
	}

	// 分页查询
	var targets []models.MonitoringTarget
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&targets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list agents: %w", err)
	}

	// 转换为Agent响应
	agents := make([]*Agent, len(targets))
	for i, target := range targets {
		var config map[string]interface{}
		var tags map[string]string
		json.Unmarshal([]byte(target.Credentials), &config)
		json.Unmarshal([]byte(target.Labels), &tags)

		metrics, _ := s.getAgentMetrics(target.ID)

		agents[i] = &Agent{
			ID:        target.ID,
			Name:      target.Name,
			Type:      "agent",
			Platform:  target.Platform,
			IPAddress: target.Address,
			Port:      target.Port,
			Status:    target.Status,
			Config:    config,
			Tags:      tags,
			Metrics:   *metrics,
			CreatedAt: target.CreatedAt,
			UpdatedAt: target.UpdatedAt,
		}
		if target.LastSeen != nil {
			agents[i].LastSeen = *target.LastSeen
		}
	}

	return agents, total, nil
}

// CreateDeployment 创建部署
func (s *AgentService) CreateDeployment(req *CreateDeploymentRequest, userID uuid.UUID) (*AgentDeployment, error) {
	// 创建部署记录
	deployment := &AgentDeployment{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		AgentType:   req.AgentType,
		Version:     req.Version,
		Targets:     req.Targets,
		Config:      req.Config,
		Status:      "pending",
		Progress: DeploymentProgress{
			Total:   len(req.Targets),
			Pending: len(req.Targets),
		},
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 初始化目标状态
	for i := range deployment.Targets {
		deployment.Targets[i].ID = uuid.New()
		deployment.Targets[i].Status = "pending"
	}

	// 保存到数据库（这里需要创建相应的数据模型）
	// 暂时返回创建的部署信息

	// 异步执行部署
	go s.executeDeployment(deployment)

	return deployment, nil
}

// GetDeployment 获取部署信息
func (s *AgentService) GetDeployment(deploymentID uuid.UUID) (*AgentDeployment, error) {
	// 从数据库获取部署信息
	// 这里需要实现具体的数据库查询逻辑
	return &AgentDeployment{}, nil
}

// ListDeployments 获取部署列表
func (s *AgentService) ListDeployments(page, pageSize int, status string) ([]*AgentDeployment, int64, error) {
	// 从数据库获取部署列表
	// 这里需要实现具体的数据库查询逻辑
	return []*AgentDeployment{}, 0, nil
}

// GetAgentPackages 获取Agent安装包列表
func (s *AgentService) GetAgentPackages(agentType, platform, architecture string) ([]*AgentPackage, error) {
	ctx := context.Background()

	// 检查缓存
	cacheKey := fmt.Sprintf("agent_packages:%s:%s:%s", agentType, platform, architecture)
	if s.cacheManager != nil {
		var cached string
		if err := s.cacheManager.Get(ctx, cacheKey, &cached); err == nil {
			var packages []*AgentPackage
			if err := json.Unmarshal([]byte(cached), &packages); err == nil {
				return packages, nil
			}
		}
	}

	// 从数据库获取安装包列表
	// 这里需要实现具体的数据库查询逻辑
	packages := []*AgentPackage{
		{
			ID:           uuid.New(),
			Name:         "node_exporter",
			Type:         "node_exporter",
			Version:      "1.6.1",
			Platform:     "linux",
			Architecture: "amd64",
			Size:         10485760,
			Checksum:     "sha256:abc123...",
			DownloadURL:  "/api/v1/agents/download/node_exporter/1.6.1/linux/amd64",
			Description:  "Prometheus Node Exporter for Linux AMD64",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// 缓存结果
	if s.cacheManager != nil {
		if data, err := json.Marshal(packages); err == nil {
			s.cacheManager.Set(ctx, cacheKey, string(data), 30*time.Minute)
		}
	}

	return packages, nil
}

// ProcessHeartbeat 处理Agent心跳
func (s *AgentService) ProcessHeartbeat(heartbeat *AgentHeartbeat) error {
	// 更新Agent最后心跳时间和状态
	updates := map[string]interface{}{
		"last_seen": heartbeat.Timestamp,
		"status":    heartbeat.Status,
	}

	if err := s.db.Model(&models.MonitoringTarget{}).Where("id = ? AND type = 'agent'", heartbeat.AgentID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update agent heartbeat: %w", err)
	}

	// 缓存Agent指标
	if s.cacheManager != nil {
		cacheKey := fmt.Sprintf("agent_metrics:%s", heartbeat.AgentID)
		if data, err := json.Marshal(heartbeat.Metrics); err == nil {
			s.cacheManager.Set(context.Background(), cacheKey, string(data), 5*time.Minute)
		}
	}

	return nil
}

// GetAgentConfig 获取Agent配置
func (s *AgentService) GetAgentConfig(agentID uuid.UUID) (*AgentConfig, error) {
	// 从数据库获取Agent配置
	// 这里需要实现具体的数据库查询逻辑
	return &AgentConfig{}, nil
}

// UpdateAgentConfig 更新Agent配置
func (s *AgentService) UpdateAgentConfig(agentID uuid.UUID, config map[string]interface{}) (*AgentConfig, error) {
	// 更新Agent配置
	// 这里需要实现具体的配置更新逻辑
	return &AgentConfig{}, nil
}

// 私有方法实现

// getAgentMetrics 获取Agent指标
func (s *AgentService) getAgentMetrics(agentID uuid.UUID) (*AgentMetrics, error) {
	ctx := context.Background()

	// 检查缓存
	cacheKey := fmt.Sprintf("agent_metrics:%s", agentID)
	if s.cacheManager != nil {
		var cached string
		if err := s.cacheManager.Get(ctx, cacheKey, &cached); err == nil {
			var metrics AgentMetrics
			if err := json.Unmarshal([]byte(cached), &metrics); err == nil {
				return &metrics, nil
			}
		}
	}

	// 返回默认指标
	return &AgentMetrics{
		CPUUsage:    0.0,
		MemoryUsage: 0.0,
		DiskUsage:   0.0,
		NetworkIO:   0.0,
		Uptime:      0.0,
		CollectRate: 0.0,
		ErrorRate:   0.0,
		LastCollect: time.Now(),
	}, nil
}

// executeDeployment 执行部署
func (s *AgentService) executeDeployment(deployment *AgentDeployment) {
	// 更新部署状态为运行中
	deployment.Status = "running"
	deployment.Progress.Running = deployment.Progress.Total
	deployment.Progress.Pending = 0

	// 模拟部署过程
	for i := range deployment.Targets {
		target := &deployment.Targets[i]
		
		// 模拟部署延迟
		time.Sleep(2 * time.Second)
		
		// 模拟部署结果（90%成功率）
		if i < len(deployment.Targets)*9/10 {
			target.Status = "success"
			target.Message = "Agent deployed successfully"
			now := time.Now()
			target.DeployedAt = &now
			deployment.Progress.Success++
		} else {
			target.Status = "failed"
			target.Message = "Failed to connect to target host"
			deployment.Progress.Failed++
		}
		
		deployment.Progress.Running--
		deployment.Progress.Percent = float64(deployment.Progress.Success+deployment.Progress.Failed) / float64(deployment.Progress.Total) * 100
	}

	// 更新最终状态
	if deployment.Progress.Failed == 0 {
		deployment.Status = "completed"
	} else if deployment.Progress.Success == 0 {
		deployment.Status = "failed"
	} else {
		deployment.Status = "partial"
	}

	deployment.UpdatedAt = time.Now()

	// 这里应该更新数据库中的部署状态
	// 暂时省略数据库更新逻辑
}