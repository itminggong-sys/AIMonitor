package services

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DiscoveryService 服务发现服务
type DiscoveryService struct {
	db           *gorm.DB
	cacheManager *cache.CacheManager
	config       *config.Config
	agentService *AgentService
	apiKeyService *APIKeyService
	mu           sync.RWMutex
	discoveryTasks map[string]*DiscoveryTask
}

// DiscoveryTask 发现任务
type DiscoveryTask struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // network_scan, ssh_scan, agent_discovery
	Status      string                 `json:"status"` // pending, running, completed, failed
	Targets     []DiscoveryTarget      `json:"targets"`
	Results     []DiscoveryResult      `json:"results"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Progress    int                    `json:"progress"` // 0-100
	Message     string                 `json:"message"`
}

// DiscoveryTarget 发现目标
type DiscoveryTarget struct {
	Host     string            `json:"host"`
	PortRange string           `json:"port_range,omitempty"`
	Type     string            `json:"type"` // server, container, middleware, apm
	Credentials map[string]string `json:"credentials,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
}

// DiscoveryResult 发现结果
type DiscoveryResult struct {
	Host        string                 `json:"host"`
	Port        int                    `json:"port,omitempty"`
	Type        string                 `json:"type"`
	Service     string                 `json:"service,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Status      string                 `json:"status"` // discovered, registered, failed
	AgentID     *uuid.UUID             `json:"agent_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
	DiscoveredAt time.Time             `json:"discovered_at"`
}

// CreateDiscoveryTaskRequest 创建发现任务请求
type CreateDiscoveryTaskRequest struct {
	Name    string                 `json:"name" binding:"required"`
	Type    string                 `json:"type" binding:"required,oneof=network_scan ssh_scan agent_discovery"`
	Targets []DiscoveryTarget      `json:"targets" binding:"required,min=1"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

// NewDiscoveryService 创建服务发现服务
func NewDiscoveryService(db *gorm.DB, cacheManager *cache.CacheManager, config *config.Config, agentService *AgentService, apiKeyService *APIKeyService) *DiscoveryService {
	return &DiscoveryService{
		db:             db,
		cacheManager:   cacheManager,
		config:         config,
		agentService:   agentService,
		apiKeyService:  apiKeyService,
		discoveryTasks: make(map[string]*DiscoveryTask),
	}
}

// CreateDiscoveryTask 创建发现任务
func (s *DiscoveryService) CreateDiscoveryTask(req *CreateDiscoveryTaskRequest) (*DiscoveryTask, error) {
	task := &DiscoveryTask{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Type:      req.Type,
		Status:    "pending",
		Targets:   req.Targets,
		Results:   make([]DiscoveryResult, 0),
		Config:    req.Config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Progress:  0,
	}

	s.mu.Lock()
	s.discoveryTasks[task.ID] = task
	s.mu.Unlock()

	// 异步执行发现任务
	go s.executeDiscoveryTask(task)

	return task, nil
}

// GetDiscoveryTask 获取发现任务
func (s *DiscoveryService) GetDiscoveryTask(taskID string) (*DiscoveryTask, error) {
	s.mu.RLock()
	task, exists := s.discoveryTasks[taskID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("discovery task not found")
	}

	return task, nil
}

// ListDiscoveryTasks 获取发现任务列表
func (s *DiscoveryService) ListDiscoveryTasks() ([]*DiscoveryTask, error) {
	s.mu.RLock()
	tasks := make([]*DiscoveryTask, 0, len(s.discoveryTasks))
	for _, task := range s.discoveryTasks {
		tasks = append(tasks, task)
	}
	s.mu.RUnlock()

	return tasks, nil
}

// executeDiscoveryTask 执行发现任务
func (s *DiscoveryService) executeDiscoveryTask(task *DiscoveryTask) {
	s.updateTaskStatus(task.ID, "running", "开始执行发现任务")

	switch task.Type {
	case "network_scan":
		s.executeNetworkScan(task)
	case "ssh_scan":
		s.executeSSHScan(task)
	case "agent_discovery":
		s.executeAgentDiscovery(task)
	default:
		s.updateTaskStatus(task.ID, "failed", fmt.Sprintf("不支持的发现类型: %s", task.Type))
		return
	}

	s.updateTaskStatus(task.ID, "completed", "发现任务完成")
	now := time.Now()
	task.CompletedAt = &now
}

// executeNetworkScan 执行网络扫描
func (s *DiscoveryService) executeNetworkScan(task *DiscoveryTask) {
	totalTargets := len(task.Targets)
	processed := 0

	for _, target := range task.Targets {
		s.scanNetworkTarget(task, target)
		processed++
		progress := (processed * 100) / totalTargets
		s.updateTaskProgress(task.ID, progress, fmt.Sprintf("已扫描 %d/%d 个目标", processed, totalTargets))
	}
}

// scanNetworkTarget 扫描网络目标
func (s *DiscoveryService) scanNetworkTarget(task *DiscoveryTask, target DiscoveryTarget) {
	// 解析端口范围
	ports := s.parsePortRange(target.PortRange)
	if len(ports) == 0 {
		// 默认扫描常用端口
		ports = []int{22, 80, 443, 3306, 5432, 6379, 9092, 9200, 27017}
	}

	for _, port := range ports {
		if s.isPortOpen(target.Host, port) {
			service := s.identifyService(target.Host, port)
			result := DiscoveryResult{
				Host:         target.Host,
				Port:         port,
				Type:         target.Type,
				Service:      service,
				Status:       "discovered",
				DiscoveredAt: time.Now(),
				Metadata: map[string]interface{}{
					"scan_type": "network",
					"tags":      target.Tags,
				},
			}

			// 尝试自动注册Agent
			if s.shouldAutoRegister(service) {
				s.autoRegisterAgent(task, &result, target)
			}

			task.Results = append(task.Results, result)
		}
	}
}

// executeSSHScan 执行SSH扫描
func (s *DiscoveryService) executeSSHScan(task *DiscoveryTask) {
	totalTargets := len(task.Targets)
	processed := 0

	for _, target := range task.Targets {
		s.scanSSHTarget(task, target)
		processed++
		progress := (processed * 100) / totalTargets
		s.updateTaskProgress(task.ID, progress, fmt.Sprintf("已扫描 %d/%d 个SSH目标", processed, totalTargets))
	}
}

// scanSSHTarget 扫描SSH目标
func (s *DiscoveryService) scanSSHTarget(task *DiscoveryTask, target DiscoveryTarget) {
	// 这里实现SSH连接和系统信息收集
	// 为了简化，这里返回模拟结果
	result := DiscoveryResult{
		Host:         target.Host,
		Port:         22,
		Type:         "server",
		Service:      "ssh",
		Status:       "discovered",
		DiscoveredAt: time.Now(),
		Metadata: map[string]interface{}{
			"scan_type": "ssh",
			"os":        "linux", // 实际应该通过SSH获取
			"tags":      target.Tags,
		},
	}

	// 尝试部署Agent
	s.deployAgentViaSSH(task, &result, target)
	task.Results = append(task.Results, result)
}

// executeAgentDiscovery 执行Agent发现
func (s *DiscoveryService) executeAgentDiscovery(task *DiscoveryTask) {
	// 扫描网络中已安装但未注册的Agent
	totalTargets := len(task.Targets)
	processed := 0

	for _, target := range task.Targets {
		s.discoverExistingAgents(task, target)
		processed++
		progress := (processed * 100) / totalTargets
		s.updateTaskProgress(task.ID, progress, fmt.Sprintf("已发现 %d/%d 个目标", processed, totalTargets))
	}
}

// discoverExistingAgents 发现现有Agent
func (s *DiscoveryService) discoverExistingAgents(task *DiscoveryTask, target DiscoveryTarget) {
	// 扫描Agent默认端口（假设Agent在8080端口提供健康检查）
	agentPort := 8080
	if s.isPortOpen(target.Host, agentPort) {
		// 尝试连接Agent健康检查接口
		if s.isAgentHealthy(target.Host, agentPort) {
			result := DiscoveryResult{
				Host:         target.Host,
				Port:         agentPort,
				Type:         "agent",
				Service:      "aimonitor-agent",
				Status:       "discovered",
				DiscoveredAt: time.Now(),
				Metadata: map[string]interface{}{
					"scan_type": "agent_discovery",
					"tags":      target.Tags,
				},
			}

			// 尝试注册Agent
			s.registerDiscoveredAgent(task, &result, target)
			task.Results = append(task.Results, result)
		}
	}
}

// 辅助方法

// parsePortRange 解析端口范围
func (s *DiscoveryService) parsePortRange(portRange string) []int {
	if portRange == "" {
		return nil
	}

	var ports []int
	parts := strings.Split(portRange, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// 范围格式: 80-90
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				if err1 == nil && err2 == nil && start <= end {
					for i := start; i <= end; i++ {
						ports = append(ports, i)
					}
				}
			}
		} else {
			// 单个端口
			if port, err := strconv.Atoi(part); err == nil {
				ports = append(ports, port)
			}
		}
	}

	return ports
}

// isPortOpen 检查端口是否开放
func (s *DiscoveryService) isPortOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// identifyService 识别服务类型
func (s *DiscoveryService) identifyService(host string, port int) string {
	serviceMap := map[int]string{
		22:    "ssh",
		80:    "http",
		443:   "https",
		3306:  "mysql",
		5432:  "postgresql",
		6379:  "redis",
		9092:  "kafka",
		9200:  "elasticsearch",
		27017: "mongodb",
		8080:  "aimonitor-agent",
	}

	if service, exists := serviceMap[port]; exists {
		return service
	}
	return "unknown"
}

// shouldAutoRegister 判断是否应该自动注册
func (s *DiscoveryService) shouldAutoRegister(service string) bool {
	// 只对特定服务自动注册
	autoRegisterServices := []string{"mysql", "redis", "postgresql", "mongodb", "elasticsearch"}
	for _, autoService := range autoRegisterServices {
		if service == autoService {
			return true
		}
	}
	return false
}

// autoRegisterAgent 自动注册Agent
func (s *DiscoveryService) autoRegisterAgent(task *DiscoveryTask, result *DiscoveryResult, target DiscoveryTarget) {
	// 创建API密钥
	apiKey, err := s.apiKeyService.CreateAPIKey(&CreateAPIKeyRequest{
		Name:        fmt.Sprintf("auto-discovery-%s-%s", result.Service, result.Host),
		Key:         fmt.Sprintf("ak_%s_%s_%d", result.Service, result.Host, time.Now().Unix()),
		Description: fmt.Sprintf("自动发现的%s服务", result.Service),
		ExpiresAt:   nil, // 永不过期
	}, uuid.New())
	if err != nil {
		result.Error = fmt.Sprintf("创建API密钥失败: %v", err)
		result.Status = "failed"
		return
	}

	// 创建Agent注册请求
	agentReq := &CreateAgentRequest{
		Name:         fmt.Sprintf("%s-agent-%s", result.Service, result.Host),
		Type:         "custom",
		Version:      "1.0.0",
		Platform:     "linux",
		Architecture: "amd64",
		Hostname:     result.Host,
		IPAddress:    result.Host,
		Port:         result.Port,
		Tags:         target.Tags,
	}

	agent, err := s.agentService.CreateAgent(agentReq)
	if err != nil {
		result.Error = fmt.Sprintf("注册Agent失败: %v", err)
		result.Status = "failed"
		return
	}

	result.AgentID = &agent.ID
	result.Status = "registered"
	result.Metadata["api_key"] = apiKey.Key
	result.Metadata["agent_name"] = agent.Name
}

// deployAgentViaSSH 通过SSH部署Agent
func (s *DiscoveryService) deployAgentViaSSH(task *DiscoveryTask, result *DiscoveryResult, target DiscoveryTarget) {
	// 这里实现SSH连接和Agent部署逻辑
	// 为了简化，这里只是标记为需要手动部署
	result.Status = "discovered"
	result.Metadata["deployment_method"] = "ssh"
	result.Metadata["requires_manual_setup"] = true
}

// isAgentHealthy 检查Agent健康状态
func (s *DiscoveryService) isAgentHealthy(host string, port int) bool {
	// 尝试访问Agent健康检查接口
	// 这里简化实现，实际应该发送HTTP请求到 /health 接口
	return s.isPortOpen(host, port)
}

// registerDiscoveredAgent 注册发现的Agent
func (s *DiscoveryService) registerDiscoveredAgent(task *DiscoveryTask, result *DiscoveryResult, target DiscoveryTarget) {
	// 创建API密钥
	apiKey, err := s.apiKeyService.CreateAPIKey(&CreateAPIKeyRequest{
		Name:        fmt.Sprintf("discovered-agent-%s", result.Host),
		Key:         fmt.Sprintf("ak_discovered_%s_%d", result.Host, time.Now().Unix()),
		Description: fmt.Sprintf("发现的Agent: %s", result.Host),
		ExpiresAt:   nil,
	}, uuid.New())
	if err != nil {
		result.Error = fmt.Sprintf("创建API密钥失败: %v", err)
		result.Status = "failed"
		return
	}

	// 注册Agent
	agentReq := &CreateAgentRequest{
		Name:         fmt.Sprintf("agent-%s", result.Host),
		Type:         "custom",
		Version:      "1.0.0",
		Platform:     "linux",
		Architecture: "amd64",
		Hostname:     result.Host,
		IPAddress:    result.Host,
		Port:         result.Port,
		Tags:         target.Tags,
	}

	agent, err := s.agentService.CreateAgent(agentReq)
	if err != nil {
		result.Error = fmt.Sprintf("注册Agent失败: %v", err)
		result.Status = "failed"
		return
	}

	result.AgentID = &agent.ID
	result.Status = "registered"
	result.Metadata["api_key"] = apiKey.Key
	result.Metadata["agent_name"] = agent.Name
}

// updateTaskStatus 更新任务状态
func (s *DiscoveryService) updateTaskStatus(taskID, status, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.discoveryTasks[taskID]; exists {
		task.Status = status
		task.Message = message
		task.UpdatedAt = time.Now()
	}
}

// updateTaskProgress 更新任务进度
func (s *DiscoveryService) updateTaskProgress(taskID string, progress int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.discoveryTasks[taskID]; exists {
		task.Progress = progress
		task.Message = message
		task.UpdatedAt = time.Now()
	}
}