package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// AgentConfig Agent配置结构
type AgentConfig struct {
	Server struct {
		Host   string `yaml:"host"`
		Port   int    `yaml:"port"`
		APIKey string `yaml:"api_key"`
		SSL    bool   `yaml:"ssl"`
	} `yaml:"server"`
	Agent struct {
		Name     string        `yaml:"name"`
		Type     string        `yaml:"type"`
		Interval time.Duration `yaml:"interval"`
		Timeout  time.Duration `yaml:"timeout"`
	} `yaml:"agent"`
	Metrics struct {
		Enabled  bool          `yaml:"enabled"`
		Interval time.Duration `yaml:"interval"`
	} `yaml:"metrics"`
	Logging struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`
}

// AgentInfo Agent信息
type AgentInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Version      string            `json:"version"`
	Platform     string            `json:"platform"`
	Architecture string            `json:"architecture"`
	Hostname     string            `json:"hostname"`
	IPAddress    string            `json:"ip_address"`
	Port         int               `json:"port"`
	Config       map[string]interface{} `json:"config"`
	Tags         map[string]string `json:"tags"`
}

// MetricData 指标数据
type MetricData struct {
	AgentID   string                 `json:"agent_id"`
	Timestamp time.Time             `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
	Tags      map[string]string      `json:"tags"`
}

// HeartbeatData 心跳数据
type HeartbeatData struct {
	AgentID   string                 `json:"agent_id"`
	Timestamp time.Time             `json:"timestamp"`
	Status    string                 `json:"status"`
	Metrics   map[string]interface{} `json:"metrics"`
	Message   string                 `json:"message"`
}

// Agent 基础Agent结构
type Agent struct {
	Config     *AgentConfig
	Info       *AgentInfo
	HTTPClient *http.Client
	Ctx        context.Context
	Cancel     context.CancelFunc
	Logger     Logger
}

// Logger 日志接口
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// SimpleLogger 简单日志实现
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] "+msg+"\n", args...)
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] "+msg+"\n", args...)
}

func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}

// NewAgent 创建新的Agent
func NewAgent(configPath string) (*Agent, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	hostname, _ := os.Hostname()
	ctx, cancel := context.WithCancel(context.Background())

	agent := &Agent{
		Config: config,
		Info: &AgentInfo{
			ID:           uuid.New().String(),
			Name:         config.Agent.Name,
			Type:         config.Agent.Type,
			Version:      "1.0.0",
			Platform:     runtime.GOOS,
			Architecture: runtime.GOARCH,
			Hostname:     hostname,
			Tags:         make(map[string]string),
		},
		HTTPClient: &http.Client{
			Timeout: config.Agent.Timeout,
		},
		Ctx:    ctx,
		Cancel: cancel,
		Logger: &SimpleLogger{},
	}

	return agent, nil
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*AgentConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AgentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// 设置默认值
	if config.Agent.Interval == 0 {
		config.Agent.Interval = 30 * time.Second
	}
	if config.Agent.Timeout == 0 {
		config.Agent.Timeout = 10 * time.Second
	}
	if config.Metrics.Interval == 0 {
		config.Metrics.Interval = 30 * time.Second
	}

	return &config, nil
}

// Register 注册Agent到监控平台
func (a *Agent) Register() error {
	url := a.getAPIURL("/api/v1/agents")
	data, err := json.Marshal(a.Info)
	if err != nil {
		return fmt.Errorf("failed to marshal agent info: %w", err)
	}

	req, err := http.NewRequestWithContext(a.Ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Config.Server.APIKey)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	a.Logger.Info("Agent registered successfully: %s", a.Info.ID)
	return nil
}

// SendHeartbeat 发送心跳
func (a *Agent) SendHeartbeat(metrics map[string]interface{}) error {
	heartbeat := HeartbeatData{
		AgentID:   a.Info.ID,
		Timestamp: time.Now(),
		Status:    "online",
		Metrics:   metrics,
		Message:   "Agent is running normally",
	}

	url := a.getAPIURL("/api/v1/agents/heartbeat")
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat: %w", err)
	}

	req, err := http.NewRequestWithContext(a.Ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Config.Server.APIKey)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		a.Logger.Warn("Heartbeat failed with status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("heartbeat failed with status %d", resp.StatusCode)
	}

	return nil
}

// SendMetrics 发送指标数据
func (a *Agent) SendMetrics(metrics map[string]interface{}) error {
	metricData := MetricData{
		AgentID:   a.Info.ID,
		Timestamp: time.Now(),
		Metrics:   metrics,
		Tags:      a.Info.Tags,
	}

	url := a.getAPIURL("/api/v1/metrics")
	data, err := json.Marshal(metricData)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req, err := http.NewRequestWithContext(a.Ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Config.Server.APIKey)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		a.Logger.Warn("Metrics submission failed with status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("metrics submission failed with status %d", resp.StatusCode)
	}

	return nil
}

// Start 启动Agent
func (a *Agent) Start() error {
	// 注册Agent
	if err := a.Register(); err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	// 启动心跳和指标收集
	heartbeatTicker := time.NewTicker(a.Config.Agent.Interval)
	metricsTicker := time.NewTicker(a.Config.Metrics.Interval)

	go func() {
		for {
			select {
			case <-a.Ctx.Done():
				return
			case <-heartbeatTicker.C:
				if err := a.SendHeartbeat(nil); err != nil {
					a.Logger.Error("Failed to send heartbeat: %v", err)
				}
			case <-metricsTicker.C:
				if a.Config.Metrics.Enabled {
					// 这里应该由具体的Agent实现收集指标的逻辑
					a.Logger.Debug("Metrics collection interval reached")
				}
			}
		}
	}()

	a.Logger.Info("Agent started successfully")
	return nil
}

// Stop 停止Agent
func (a *Agent) Stop() {
	a.Cancel()
	a.Logger.Info("Agent stopped")
}

// getAPIURL 构建API URL
func (a *Agent) getAPIURL(path string) string {
	scheme := "http"
	if a.Config.Server.SSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d%s", scheme, a.Config.Server.Host, a.Config.Server.Port, path)
}