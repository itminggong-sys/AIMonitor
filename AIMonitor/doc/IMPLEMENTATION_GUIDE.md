# AI监控系统技术实现指南

## 文档概述

本文档基于AI监控系统的架构设计和功能需求，提供详细的技术实现指南，包括代码结构、关键算法、部署方案和最佳实践。

## 项目结构说明

### 目录结构
```
AIMonitor/
├── cmd/
│   └── server/                 # 应用程序入口
│       └── main.go
├── configs/
│   └── config.yaml            # 配置文件
├── internal/
│   ├── auth/                  # 认证模块
│   ├── cache/                 # 缓存模块
│   ├── config/                # 配置管理
│   ├── database/              # 数据库连接
│   ├── handlers/              # HTTP处理器
│   ├── metrics/               # Prometheus指标
│   ├── middleware/            # 中间件
│   ├── models/                # 数据模型
│   ├── routes/                # 路由配置
│   ├── scheduler/             # 定时任务
│   ├── services/              # 业务逻辑
│   │   ├── monitoring/        # 监控服务
│   │   ├── alert/             # 告警服务
│   │   ├── ai/                # AI分析服务
│   │   ├── middleware_monitor/ # 中间件监控
│   │   ├── apm/               # APM服务
│   │   ├── container/         # 容器监控
│   │   └── agent/             # Agent管理
│   ├── utils/                 # 工具函数
│   └── websocket/             # WebSocket服务
├── agents/                    # Agent相关
│   ├── windows/               # Windows Agent
│   ├── linux/                 # Linux Agent
│   ├── container/             # 容器Agent
│   └── apm/                   # APM Agent
├── doc/                       # 文档目录
└── go.mod                     # Go模块文件
```

## 核心模块实现

### 1. 数据采集模块 (Monitoring Service)

#### 1.1 Prometheus集成

**文件位置**: `internal/services/monitoring_service.go`

**核心功能**:
- 监控目标管理
- 指标数据查询和存储
- 系统指标采集
- 仪表板管理

**关键实现**:
```go
type MonitoringService struct {
    db             *gorm.DB
    redis          *redis.Client
    prometheusAPI  v1.API
    logger         *logrus.Logger
}

// 查询指标数据
func (s *MonitoringService) QueryMetrics(ctx context.Context, query string, timeRange TimeRange) (*MetricsData, error) {
    // 构建Prometheus查询
    promQuery := model.Value(query)
    
    // 执行查询
    result, warnings, err := s.prometheusAPI.QueryRange(ctx, promQuery, v1.Range{
        Start: timeRange.Start,
        End:   timeRange.End,
        Step:  timeRange.Step,
    })
    
    if err != nil {
        return nil, fmt.Errorf("prometheus query failed: %w", err)
    }
    
    // 处理查询结果
    return s.processPrometheusResult(result, warnings)
}
```

#### 1.2 跨平台数据采集

**支持平台**:
- Windows: WMI、Performance Counters
- Linux: /proc、/sys、systemd
- macOS: System Profiler、IOKit
- ESXi: vSphere API

**实现策略**:
```go
type PlatformCollector interface {
    CollectSystemMetrics() (*SystemMetrics, error)
    CollectApplicationMetrics() (*ApplicationMetrics, error)
    CollectNetworkMetrics() (*NetworkMetrics, error)
}

type WindowsCollector struct {
    wmiClient *wmi.Client
}

type LinuxCollector struct {
    procFS string
    sysFS  string
}

type MacOSCollector struct {
    systemProfiler *exec.Cmd
}
```

### 2. 告警管理模块 (Alert Service)

#### 2.1 告警规则引擎

**文件位置**: `internal/services/alert_service.go`

**核心功能**:
- 告警规则管理
- 告警触发检测
- 告警生命周期管理
- 通知发送

**关键实现**:
```go
type AlertService struct {
    db                *gorm.DB
    redis             *redis.Client
    notificationSvc   *NotificationService
    aiSvc            *AIService
    logger           *logrus.Logger
}

// 检查告警规则
func (s *AlertService) CheckAlertRules(ctx context.Context) error {
    rules, err := s.GetActiveAlertRules(ctx)
    if err != nil {
        return fmt.Errorf("failed to get alert rules: %w", err)
    }
    
    for _, rule := range rules {
        if err := s.evaluateRule(ctx, rule); err != nil {
            s.logger.WithError(err).WithField("rule_id", rule.ID).Error("Failed to evaluate rule")
        }
    }
    
    return nil
}

// 评估告警规则
func (s *AlertService) evaluateRule(ctx context.Context, rule *models.AlertRule) error {
    // 查询指标数据
    metrics, err := s.queryMetricsForRule(ctx, rule)
    if err != nil {
        return fmt.Errorf("failed to query metrics: %w", err)
    }
    
    // 检查阈值
    if s.checkThreshold(metrics, rule) {
        return s.triggerAlert(ctx, rule, metrics)
    }
    
    return nil
}
```

#### 2.2 智能阈值管理

**动态阈值算法**:
```go
type ThresholdCalculator struct {
    historicalData []float64
    sensitivity    float64
}

// 计算动态阈值
func (tc *ThresholdCalculator) CalculateDynamicThreshold() (float64, error) {
    if len(tc.historicalData) < 10 {
        return 0, errors.New("insufficient historical data")
    }
    
    // 计算统计指标
    mean := tc.calculateMean()
    stdDev := tc.calculateStdDev(mean)
    
    // 基于标准差的动态阈值
    threshold := mean + (tc.sensitivity * stdDev)
    
    return threshold, nil
}
```

### 3. AI分析模块 (AI Service)

#### 3.1 AI模型集成

**文件位置**: `internal/services/ai_service.go`

**支持的AI服务**:
- OpenAI GPT系列
- Claude
- 国产大模型（通义千问、文心一言等）

**核心实现**:
```go
type AIService struct {
    db           *gorm.DB
    redis        *redis.Client
    openaiClient *openai.Client
    logger       *logrus.Logger
    config       *AIConfig
}

// AI告警分析
func (s *AIService) AnalyzeAlert(ctx context.Context, alert *models.Alert) (*AIAnalysis, error) {
    // 构建分析提示
    prompt := s.buildAlertAnalysisPrompt(alert)
    
    // 调用AI模型
    response, err := s.callAIModel(ctx, prompt)
    if err != nil {
        return nil, fmt.Errorf("AI analysis failed: %w", err)
    }
    
    // 解析AI响应
    analysis, err := s.parseAIResponse(response)
    if err != nil {
        return nil, fmt.Errorf("failed to parse AI response: %w", err)
    }
    
    // 保存分析结果
    if err := s.saveAnalysis(ctx, alert.ID, analysis); err != nil {
        s.logger.WithError(err).Error("Failed to save AI analysis")
    }
    
    return analysis, nil
}
```

#### 3.2 智能分析算法

**异常检测算法**:
```go
type AnomalyDetector struct {
    model      *isolation.Forest
    threshold  float64
    windowSize int
}

// 检测异常
func (ad *AnomalyDetector) DetectAnomaly(data []float64) (bool, float64, error) {
    if len(data) < ad.windowSize {
        return false, 0, errors.New("insufficient data")
    }
    
    // 特征提取
    features := ad.extractFeatures(data)
    
    // 异常评分
    score := ad.model.AnomalyScore(features)
    
    // 判断是否异常
    isAnomaly := score > ad.threshold
    
    return isAnomaly, score, nil
}

### 4. 中间件监控模块 (Middleware Monitor Service)

#### 4.1 数据库监控

**文件位置**: `internal/services/middleware_monitor/database_monitor.go`

**支持的数据库**:
- MySQL
- PostgreSQL
- MongoDB
- ClickHouse
- Redis

**核心实现**:
```go
type DatabaseMonitor struct {
    db     *gorm.DB
    redis  *redis.Client
    logger *logrus.Logger
}

// MySQL监控
func (dm *DatabaseMonitor) MonitorMySQL(ctx context.Context, config *MySQLConfig) (*MySQLMetrics, error) {
    db, err := sql.Open("mysql", config.DSN)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
    }
    defer db.Close()
    
    metrics := &MySQLMetrics{}
    
    // 查询连接数
    if err := db.QueryRow("SHOW STATUS LIKE 'Threads_connected'").Scan(&metrics.Connections); err != nil {
        return nil, fmt.Errorf("failed to get connection count: %w", err)
    }
    
    // 查询QPS
    if err := db.QueryRow("SHOW STATUS LIKE 'Queries'").Scan(&metrics.QPS); err != nil {
        return nil, fmt.Errorf("failed to get QPS: %w", err)
    }
    
    // 查询慢查询数
    if err := db.QueryRow("SHOW STATUS LIKE 'Slow_queries'").Scan(&metrics.SlowQueries); err != nil {
        return nil, fmt.Errorf("failed to get slow queries: %w", err)
    }
    
    return metrics, nil
}

// Redis监控
func (dm *DatabaseMonitor) MonitorRedis(ctx context.Context, config *RedisConfig) (*RedisMetrics, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     config.Addr,
        Password: config.Password,
        DB:       config.DB,
    })
    defer client.Close()
    
    info, err := client.Info(ctx).Result()
    if err != nil {
        return nil, fmt.Errorf("failed to get Redis info: %w", err)
    }
    
    metrics := &RedisMetrics{}
    // 解析Redis INFO信息
    metrics.UsedMemory = dm.parseRedisInfo(info, "used_memory")
    metrics.ConnectedClients = dm.parseRedisInfo(info, "connected_clients")
    metrics.TotalCommandsProcessed = dm.parseRedisInfo(info, "total_commands_processed")
    
    return metrics, nil
}
```

#### 4.2 消息队列监控

**支持的消息队列**:
- Apache Kafka
- RabbitMQ
- NATS

```go
// Kafka监控
func (dm *DatabaseMonitor) MonitorKafka(ctx context.Context, config *KafkaConfig) (*KafkaMetrics, error) {
    client, err := sarama.NewClient(config.Brokers, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create Kafka client: %w", err)
    }
    defer client.Close()
    
    metrics := &KafkaMetrics{}
    
    // 获取主题列表
    topics, err := client.Topics()
    if err != nil {
        return nil, fmt.Errorf("failed to get topics: %w", err)
    }
    metrics.TopicCount = len(topics)
    
    // 获取分区信息
    for _, topic := range topics {
        partitions, err := client.Partitions(topic)
        if err != nil {
            continue
        }
        metrics.PartitionCount += len(partitions)
    }
    
    return metrics, nil
}
```

### 5. APM应用性能监控模块 (APM Service)

#### 5.1 探针管理

**文件位置**: `internal/services/apm/probe_manager.go`

**支持的语言**:
- Java (基于JavaAgent)
- Go (基于OpenTelemetry)
- Python (基于OpenTelemetry)
- Node.js (基于OpenTelemetry)
- .NET (基于OpenTelemetry)

**核心实现**:
```go
type APMService struct {
    db           *gorm.DB
    redis        *redis.Client
    traceStorage *TraceStorage
    logger       *logrus.Logger
}

// 处理链路追踪数据
func (apm *APMService) ProcessTrace(ctx context.Context, trace *Trace) error {
    // 验证链路数据
    if err := apm.validateTrace(trace); err != nil {
        return fmt.Errorf("invalid trace data: %w", err)
    }
    
    // 存储链路数据
    if err := apm.traceStorage.Store(ctx, trace); err != nil {
        return fmt.Errorf("failed to store trace: %w", err)
    }
    
    // 分析性能指标
    metrics := apm.analyzeTraceMetrics(trace)
    
    // 检查性能异常
    if apm.detectPerformanceAnomaly(metrics) {
        return apm.triggerPerformanceAlert(ctx, trace, metrics)
    }
    
    return nil
}

// 分析链路性能指标
func (apm *APMService) analyzeTraceMetrics(trace *Trace) *TraceMetrics {
    metrics := &TraceMetrics{
        TraceID:      trace.TraceID,
        Duration:     trace.Duration,
        SpanCount:    len(trace.Spans),
        ErrorCount:   0,
        ServiceCount: make(map[string]int),
    }
    
    for _, span := range trace.Spans {
        // 统计服务调用次数
        metrics.ServiceCount[span.ServiceName]++
        
        // 统计错误数量
        if span.Status == "error" {
            metrics.ErrorCount++
        }
        
        // 分析慢调用
        if span.Duration > apm.slowThreshold {
            metrics.SlowSpans = append(metrics.SlowSpans, span)
        }
    }
    
    return metrics
}
```

#### 5.2 用户体验监控

```go
// 处理用户体验数据
func (apm *APMService) ProcessUserExperience(ctx context.Context, ux *UserExperience) error {
    // 计算核心Web指标
    webVitals := &WebVitals{
        LCP: ux.LargestContentfulPaint,  // 最大内容绘制
        FID: ux.FirstInputDelay,         // 首次输入延迟
        CLS: ux.CumulativeLayoutShift,   // 累积布局偏移
    }
    
    // 评估用户体验等级
    score := apm.calculateUXScore(webVitals)
    
    // 存储用户体验数据
    return apm.storeUserExperience(ctx, ux, score)
}
```

### 6. 容器监控模块 (Container Monitor Service)

#### 6.1 Docker监控

**文件位置**: `internal/services/container/docker_monitor.go`

**核心实现**:
```go
type ContainerMonitor struct {
    dockerClient *client.Client
    k8sClient    kubernetes.Interface
    db           *gorm.DB
    logger       *logrus.Logger
}

// Docker容器监控
func (cm *ContainerMonitor) MonitorDockerContainers(ctx context.Context) error {
    containers, err := cm.dockerClient.ContainerList(ctx, types.ContainerListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list containers: %w", err)
    }
    
    for _, container := range containers {
        metrics, err := cm.collectContainerMetrics(ctx, container.ID)
        if err != nil {
            cm.logger.WithError(err).WithField("container_id", container.ID).Error("Failed to collect container metrics")
            continue
        }
        
        // 存储容器指标
        if err := cm.storeContainerMetrics(ctx, metrics); err != nil {
            cm.logger.WithError(err).Error("Failed to store container metrics")
        }
    }
    
    return nil
}

// 收集容器指标
func (cm *ContainerMonitor) collectContainerMetrics(ctx context.Context, containerID string) (*ContainerMetrics, error) {
    stats, err := cm.dockerClient.ContainerStats(ctx, containerID, false)
    if err != nil {
        return nil, fmt.Errorf("failed to get container stats: %w", err)
    }
    defer stats.Body.Close()
    
    var statsData types.StatsJSON
    if err := json.NewDecoder(stats.Body).Decode(&statsData); err != nil {
        return nil, fmt.Errorf("failed to decode stats: %w", err)
    }
    
    metrics := &ContainerMetrics{
        ContainerID: containerID,
        CPUUsage:    cm.calculateCPUUsage(&statsData),
        MemoryUsage: statsData.MemoryStats.Usage,
        NetworkIO:   cm.calculateNetworkIO(&statsData),
        DiskIO:      cm.calculateDiskIO(&statsData),
        Timestamp:   time.Now(),
    }
    
    return metrics, nil
}
```

#### 6.2 Kubernetes监控

```go
// Kubernetes集群监控
func (cm *ContainerMonitor) MonitorKubernetesCluster(ctx context.Context) error {
    // 监控节点
    if err := cm.monitorNodes(ctx); err != nil {
        return fmt.Errorf("failed to monitor nodes: %w", err)
    }
    
    // 监控Pod
    if err := cm.monitorPods(ctx); err != nil {
        return fmt.Errorf("failed to monitor pods: %w", err)
    }
    
    // 监控服务
    if err := cm.monitorServices(ctx); err != nil {
        return fmt.Errorf("failed to monitor services: %w", err)
    }
    
    return nil
}

// 监控Pod
func (cm *ContainerMonitor) monitorPods(ctx context.Context) error {
    pods, err := cm.k8sClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list pods: %w", err)
    }
    
    for _, pod := range pods.Items {
        metrics := &PodMetrics{
            Name:      pod.Name,
            Namespace: pod.Namespace,
            Phase:     string(pod.Status.Phase),
            NodeName:  pod.Spec.NodeName,
            Timestamp: time.Now(),
        }
        
        // 收集Pod资源使用情况
        if err := cm.collectPodResourceMetrics(ctx, &pod, metrics); err != nil {
            cm.logger.WithError(err).WithField("pod", pod.Name).Error("Failed to collect pod metrics")
        }
        
        // 存储Pod指标
        if err := cm.storePodMetrics(ctx, metrics); err != nil {
            cm.logger.WithError(err).Error("Failed to store pod metrics")
        }
    }
    
    return nil
}
```

### 7. Agent管理模块 (Agent Management Service)

#### 7.1 Agent下载管理

**文件位置**: `internal/services/agent/agent_manager.go`

**核心实现**:
```go
type AgentManager struct {
    db           *gorm.DB
    redis        *redis.Client
    fileStorage  *FileStorage
    logger       *logrus.Logger
}

// 获取Agent下载信息
func (am *AgentManager) GetAgentDownload(ctx context.Context, req *AgentDownloadRequest) (*AgentDownloadResponse, error) {
    // 验证请求参数
    if err := am.validateDownloadRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // 查找匹配的Agent版本
    agent, err := am.findMatchingAgent(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to find matching agent: %w", err)
    }
    
    // 生成下载URL
    downloadURL, err := am.generateDownloadURL(ctx, agent)
    if err != nil {
        return nil, fmt.Errorf("failed to generate download URL: %w", err)
    }
    
    // 记录下载日志
    am.recordDownloadLog(ctx, agent, req)
    
    return &AgentDownloadResponse{
        AgentID:     agent.ID,
        Version:     agent.Version,
        Platform:    agent.Platform,
        Architecture: agent.Architecture,
        DownloadURL: downloadURL,
        Checksum:    agent.Checksum,
        Size:        agent.Size,
    }, nil
}

// 上传Agent包
func (am *AgentManager) UploadAgent(ctx context.Context, req *UploadAgentRequest) (*Agent, error) {
    // 验证Agent包
    if err := am.validateAgentPackage(req.File); err != nil {
        return nil, fmt.Errorf("invalid agent package: %w", err)
    }
    
    // 计算文件校验和
    checksum, err := am.calculateChecksum(req.File)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate checksum: %w", err)
    }
    
    // 存储文件
    filePath, err := am.fileStorage.Store(ctx, req.File)
    if err != nil {
        return nil, fmt.Errorf("failed to store file: %w", err)
    }
    
    // 创建Agent记录
    agent := &models.Agent{
        Name:         req.Name,
        Version:      req.Version,
        Platform:     req.Platform,
        Architecture: req.Architecture,
        FilePath:     filePath,
        Checksum:     checksum,
        Size:         req.File.Size,
        Status:       "active",
    }
    
    if err := am.db.Create(agent).Error; err != nil {
        return nil, fmt.Errorf("failed to create agent record: %w", err)
    }
    
    return agent, nil
}
```

#### 7.2 Agent部署管理

```go
// Agent自动部署
func (am *AgentManager) DeployAgent(ctx context.Context, req *DeployAgentRequest) error {
    // 获取目标主机信息
    host, err := am.getHost(ctx, req.HostID)
    if err != nil {
        return fmt.Errorf("failed to get host: %w", err)
    }
    
    // 选择合适的Agent版本
    agent, err := am.selectAgentForHost(ctx, host)
    if err != nil {
        return fmt.Errorf("failed to select agent: %w", err)
    }
    
    // 建立SSH连接
    sshClient, err := am.createSSHConnection(host)
    if err != nil {
        return fmt.Errorf("failed to create SSH connection: %w", err)
    }
    defer sshClient.Close()
    
    // 下载Agent到目标主机
    if err := am.downloadAgentToHost(ctx, sshClient, agent); err != nil {
        return fmt.Errorf("failed to download agent: %w", err)
    }
    
    // 安装Agent
    if err := am.installAgentOnHost(ctx, sshClient, agent); err != nil {
        return fmt.Errorf("failed to install agent: %w", err)
    }
    
    // 启动Agent服务
    if err := am.startAgentService(ctx, sshClient); err != nil {
        return fmt.Errorf("failed to start agent service: %w", err)
    }
    
    // 记录部署状态
    return am.recordDeploymentStatus(ctx, req.HostID, agent.ID, "success")
}
```
```

### 4. 通知服务模块 (Notification Service)

#### 4.1 多渠道通知

**文件位置**: `internal/services/notification_service.go`

**支持的通知渠道**:
- 邮件 (SMTP)
- Webhook
- Slack
- 钉钉
- 企业微信

**核心实现**:
```go
type NotificationService struct {
    db     *gorm.DB
    redis  *redis.Client
    logger *logrus.Logger
}

// 发送通知
func (s *NotificationService) SendNotification(ctx context.Context, req *SendNotificationRequest) error {
    channel, err := s.GetNotificationChannel(ctx, req.ChannelID)
    if err != nil {
        return fmt.Errorf("failed to get notification channel: %w", err)
    }
    
    // 根据渠道类型发送通知
    switch channel.Type {
    case "email":
        return s.sendEmailNotification(ctx, channel, req)
    case "webhook":
        return s.sendWebhookNotification(ctx, channel, req)
    case "slack":
        return s.sendSlackNotification(ctx, channel, req)
    case "dingtalk":
        return s.sendDingTalkNotification(ctx, channel, req)
    default:
        return fmt.Errorf("unsupported notification type: %s", channel.Type)
    }
}
```

#### 4.2 邮件通知实现

```go
// 发送邮件通知
func (s *NotificationService) sendEmailNotification(ctx context.Context, channel *models.NotificationChannel, req *SendNotificationRequest) error {
    config := &EmailConfig{}
    if err := json.Unmarshal(channel.Config, config); err != nil {
        return fmt.Errorf("failed to parse email config: %w", err)
    }
    
    // 构建邮件内容
    htmlContent, err := s.generateEmailHTML(req)
    if err != nil {
        return fmt.Errorf("failed to generate email HTML: %w", err)
    }
    
    // 发送邮件
    m := gomail.NewMessage()
    m.SetHeader("From", config.From)
    m.SetHeader("To", req.Recipients...)
    m.SetHeader("Subject", req.Subject)
    m.SetBody("text/html", htmlContent)
    
    d := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)
    d.TLSConfig = &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}
    
    return d.DialAndSend(m)
}
```

### 5. 用户管理模块 (User Service)

#### 5.1 认证和授权

**文件位置**: `internal/services/user_service.go`

**核心功能**:
- 用户注册和登录
- JWT令牌管理
- 权限控制
- 用户资料管理

**关键实现**:
```go
type UserService struct {
    db     *gorm.DB
    redis  *redis.Client
    logger *logrus.Logger
    config *AuthConfig
}

// 用户登录
func (s *UserService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // 验证用户凭据
    user, err := s.validateCredentials(ctx, req.Username, req.Password)
    if err != nil {
        return nil, fmt.Errorf("authentication failed: %w", err)
    }
    
    // 生成JWT令牌
    accessToken, err := s.generateAccessToken(user)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }
    
    refreshToken, err := s.generateRefreshToken(user)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }
    
    // 记录登录日志
    s.recordLoginLog(ctx, user, req.ClientIP)
    
    return &LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    s.config.AccessTokenExpiry,
        User:         s.sanitizeUser(user),
    }, nil
}
```

#### 5.2 JWT令牌管理

```go
// 生成访问令牌
func (s *UserService) generateAccessToken(user *models.User) (string, error) {
    claims := &jwt.MapClaims{
        "user_id":  user.ID,
        "username": user.Username,
        "role":     user.Role,
        "exp":      time.Now().Add(s.config.AccessTokenExpiry).Unix(),
        "iat":      time.Now().Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.config.JWTSecret))
}

// 验证令牌
func (s *UserService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(s.config.JWTSecret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return &claims, nil
    }
    
    return nil, errors.New("invalid token")
}
```

### 6. 中间件模块 (Middleware)

#### 6.1 认证中间件

**文件位置**: `internal/middleware/middleware.go`

```go
// JWT认证中间件
func JWTAuthMiddleware(userService *services.UserService) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := userService.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // 设置用户信息到上下文
        c.Set("user_id", (*claims)["user_id"])
        c.Set("username", (*claims)["username"])
        c.Set("role", (*claims)["role"])
        
        c.Next()
    }
}
```

#### 6.2 权限验证中间件

```go
// 权限验证中间件
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "Role not found"})
            c.Abort()
            return
        }
        
        if !hasPermission(role.(string), permission) {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 检查权限
func hasPermission(role, permission string) bool {
    permissions := map[string][]string{
        "admin": {"read", "write", "delete", "manage_users", "manage_system"},
        "user":  {"read", "write"},
        "guest": {"read"},
    }
    
    rolePermissions, exists := permissions[role]
    if !exists {
        return false
    }
    
    for _, p := range rolePermissions {
        if p == permission {
            return true
        }
    }
    
    return false
}
```

### 7. WebSocket模块 (WebSocket)

#### 7.1 实时数据推送

**文件位置**: `internal/websocket/websocket.go`

```go
type WebSocketManager struct {
    clients    map[string]*Client
    register   chan *Client
    unregister chan *Client
    broadcast  chan *Message
    mutex      sync.RWMutex
}

// 启动WebSocket管理器
func (manager *WebSocketManager) Run() {
    for {
        select {
        case client := <-manager.register:
            manager.registerClient(client)
            
        case client := <-manager.unregister:
            manager.unregisterClient(client)
            
        case message := <-manager.broadcast:
            manager.broadcastMessage(message)
        }
    }
}

// 广播告警消息
func (manager *WebSocketManager) BroadcastAlert(alert *models.Alert) {
    message := &Message{
        Type: "alert",
        Data: AlertMessage{
            ID:          alert.ID,
            Title:       alert.Title,
            Description: alert.Description,
            Level:       alert.Level,
            Status:      alert.Status,
            CreatedAt:   alert.CreatedAt,
        },
        Timestamp: time.Now(),
    }
    
    manager.broadcast <- message
}
```

### 8. 定时任务模块 (Scheduler)

#### 8.1 任务调度器

**文件位置**: `internal/scheduler/scheduler.go`

```go
type Scheduler struct {
    cron            *cron.Cron
    alertService    *services.AlertService
    monitoringService *services.MonitoringService
    auditService    *services.AuditService
    logger          *logrus.Logger
}

// 启动调度器
func (s *Scheduler) Start() error {
    // 告警规则检查 - 每30秒执行一次
    _, err := s.cron.AddFunc("*/30 * * * * *", s.checkAlertRules)
    if err != nil {
        return fmt.Errorf("failed to add alert check job: %w", err)
    }
    
    // 系统指标收集 - 每分钟执行一次
    _, err = s.cron.AddFunc("0 * * * * *", s.collectSystemMetrics)
    if err != nil {
        return fmt.Errorf("failed to add metrics collection job: %w", err)
    }
    
    // 缓存清理 - 每小时执行一次
    _, err = s.cron.AddFunc("0 0 * * * *", s.cleanupCache)
    if err != nil {
        return fmt.Errorf("failed to add cache cleanup job: %w", err)
    }
    
    s.cron.Start()
    s.logger.Info("Scheduler started successfully")
    
    return nil
}
```

## 数据库设计

### 核心数据表

#### 用户表 (users)
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 告警规则表 (alert_rules)
```sql
CREATE TABLE alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    query TEXT NOT NULL,
    threshold_value DECIMAL(10,2),
    threshold_operator VARCHAR(10),
    severity VARCHAR(20) DEFAULT 'warning',
    enabled BOOLEAN DEFAULT true,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 告警表 (alerts)
```sql
CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    rule_id INTEGER REFERENCES alert_rules(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    level VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    value DECIMAL(10,2),
    labels JSONB,
    annotations JSONB,
    starts_at TIMESTAMP,
    ends_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 配置管理

### 配置文件结构

**文件位置**: `configs/config.yaml`

```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"  # debug, release, test
  read_timeout: 30s
  write_timeout: 30s

# 数据库配置
database:
  host: "localhost"
  port: 5432
  username: "aimonitor"
  password: "password"
  dbname: "aimonitor"
  sslmode: "disable"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s

# Redis配置
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5

# Prometheus配置
prometheus:
  url: "http://localhost:9090"
  timeout: 30s
  max_samples: 50000000

# AI配置
ai:
  provider: "openai"  # openai, claude, local
  openai:
    api_key: "your-openai-api-key"
    model: "gpt-3.5-turbo"
    max_tokens: 1000
    temperature: 0.7
  claude:
    api_key: "your-claude-api-key"
    model: "claude-3-sonnet-20240229"

# 日志配置
logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
  output: "stdout"  # stdout, file
  file_path: "logs/app.log"
  max_size: 100  # MB
  max_backups: 5
  max_age: 30  # days

# JWT配置
jwt:
  secret: "your-jwt-secret-key"
  access_token_expiry: 24h
  refresh_token_expiry: 168h  # 7 days

# 邮件配置
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"
  from: "AI Monitor <your-email@gmail.com>"
```

## 部署指南

### Docker部署

#### Dockerfile
```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o aimonitor cmd/server/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/aimonitor .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./aimonitor"]
```

#### docker-compose.yml
```yaml
version: '3.8'

services:
  aimonitor:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - PROMETHEUS_URL=http://prometheus:9090
    depends_on:
      - postgres
      - redis
      - prometheus
    volumes:
      - ./configs:/root/configs
      - ./logs:/root/logs

  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: aimonitor
      POSTGRES_USER: aimonitor
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
```

### Kubernetes部署

#### deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aimonitor
  labels:
    app: aimonitor
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aimonitor
  template:
    metadata:
      labels:
        app: aimonitor
    spec:
      containers:
      - name: aimonitor
        image: aimonitor:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
        - name: PROMETHEUS_URL
          value: "http://prometheus-service:9090"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: aimonitor-service
spec:
  selector:
    app: aimonitor
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

## 监控和运维

### 健康检查

```go
// 健康检查处理器
func (h *Handlers) HealthCheck(c *gin.Context) {
    health := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Services:  make(map[string]string),
    }
    
    // 检查数据库连接
    if err := h.db.Ping(); err != nil {
        health.Status = "unhealthy"
        health.Services["database"] = "down"
    } else {
        health.Services["database"] = "up"
    }
    
    // 检查Redis连接
    if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
        health.Status = "unhealthy"
        health.Services["redis"] = "down"
    } else {
        health.Services["redis"] = "up"
    }
    
    // 检查Prometheus连接
    if _, err := h.prometheusAPI.LabelNames(c.Request.Context(), nil, time.Now().Add(-time.Hour), time.Now()); err != nil {
        health.Status = "unhealthy"
        health.Services["prometheus"] = "down"
    } else {
        health.Services["prometheus"] = "up"
    }
    
    statusCode := http.StatusOK
    if health.Status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }
    
    c.JSON(statusCode, health)
}
```

### 性能监控

```go
// Prometheus指标定义
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
    
    databaseConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "database_connections",
            Help: "Number of database connections",
        },
        []string{"state"},
    )
)

// 注册指标
func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(databaseConnections)
}
```

## 最佳实践

### 1. 错误处理

```go
// 统一错误处理
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    return e.Message
}

// 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        if err, ok := recovered.(string); ok {
            c.JSON(http.StatusInternalServerError, &APIError{
                Code:    http.StatusInternalServerError,
                Message: "Internal Server Error",
                Details: err,
            })
        }
        c.Abort()
    })
}
```

### 2. 日志记录

```go
// 结构化日志
func (s *Service) logWithContext(ctx context.Context) *logrus.Entry {
    entry := s.logger.WithContext(ctx)
    
    if userID, ok := ctx.Value("user_id").(uint); ok {
        entry = entry.WithField("user_id", userID)
    }
    
    if requestID, ok := ctx.Value("request_id").(string); ok {
        entry = entry.WithField("request_id", requestID)
    }
    
    return entry
}

// 使用示例
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    logger := s.logWithContext(ctx)
    logger.WithField("username", req.Username).Info("Creating new user")
    
    // 业务逻辑...
    
    logger.WithField("user_id", user.ID).Info("User created successfully")
    return user, nil
}
```

### 3. 缓存策略

```go
// 缓存装饰器
func WithCache[T any](cache *redis.Client, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
    var result T
    
    // 尝试从缓存获取
    cached, err := cache.Get(context.Background(), key).Result()
    if err == nil {
        if err := json.Unmarshal([]byte(cached), &result); err == nil {
            return result, nil
        }
    }
    
    // 缓存未命中，执行函数
    result, err = fn()
    if err != nil {
        return result, err
    }
    
    // 写入缓存
    if data, err := json.Marshal(result); err == nil {
        cache.Set(context.Background(), key, data, ttl)
    }
    
    return result, nil
}
```

### 4. 数据库事务

```go
// 事务装饰器
func WithTransaction(db *gorm.DB, fn func(*gorm.DB) error) error {
    tx := db.Begin()
    if tx.Error != nil {
        return tx.Error
    }
    
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit().Error
}
```

## 安全考虑

### 1. 输入验证

```go
// 请求验证
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

// 自定义验证器
func validatePassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    
    // 至少包含一个大写字母、一个小写字母、一个数字
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    
    return hasUpper && hasLower && hasNumber
}
```

### 2. 敏感数据处理

```go
// 密码加密
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// 密码验证
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// 敏感信息脱敏
func (u *User) Sanitize() *User {
    u.PasswordHash = ""
    return u
}
```

## 测试策略

### 单元测试

```go
func TestUserService_CreateUser(t *testing.T) {
    // 设置测试数据库
    db := setupTestDB(t)
    defer teardownTestDB(t, db)
    
    // 创建服务实例
    userService := &UserService{
        db:     db,
        logger: logrus.New(),
    }
    
    // 测试用例
    tests := []struct {
        name    string
        request *CreateUserRequest
        wantErr bool
    }{
        {
            name: "valid user",
            request: &CreateUserRequest{
                Username: "testuser",
                Email:    "test@example.com",
                Password: "Password123",
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            request: &CreateUserRequest{
                Username: "testuser",
                Email:    "invalid-email",
                Password: "Password123",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, err := userService.CreateUser(context.Background(), tt.request)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
                assert.Equal(t, tt.request.Username, user.Username)
            }
        })
    }
}
```

### 集成测试

```go
func TestAPI_CreateUser(t *testing.T) {
    // 设置测试服务器
    router := setupTestRouter(t)
    server := httptest.NewServer(router)
    defer server.Close()
    
    // 测试请求
    reqBody := `{"username":"testuser","email":"test@example.com","password":"Password123"}`
    resp, err := http.Post(server.URL+"/api/v1/users", "application/json", strings.NewReader(reqBody))
    
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    
    // 验证响应
    var user User
    err = json.NewDecoder(resp.Body).Decode(&user)
    assert.NoError(t, err)
    assert.Equal(t, "testuser", user.Username)
}
```

## 总结

本实现指南提供了AI监控系统的完整技术实现方案，包括：

1. **模块化架构**: 清晰的代码组织和模块划分
2. **核心功能**: 数据采集、告警管理、AI分析、通知服务等
3. **技术栈**: Go + Gin + PostgreSQL + Redis + Prometheus
4. **部署方案**: Docker和Kubernetes部署配置
5. **最佳实践**: 错误处理、日志记录、缓存策略、安全考虑
6. **测试策略**: 单元测试和集成测试

通过遵循本指南，开发团队可以构建一个高性能、可扩展、安全可靠的AI监控系统。