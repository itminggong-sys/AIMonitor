# AI智能监控系统配置指南

## 文档概述

本文档详细说明AI智能监控系统的各项配置选项，包括系统配置、数据库配置、AI模型配置、监控配置等。适用于系统管理员和运维人员。

## 版本信息

- **系统版本**: v3.8.5
- **配置文件版本**: v1.0
- **最后更新**: 2024-01-01
- **适用环境**: 生产环境、测试环境

## 配置文件结构

```
configs/
├── config.yaml              # 主配置文件
├── config.yaml.example       # 配置模板文件
├── database.yaml            # 数据库配置
├── redis.yaml               # Redis配置
└── monitoring.yaml          # 监控配置
```

# AI Monitor 配置指南

## 📋 目录

1. [一键部署后配置](#一键部署后配置) - **小白用户必读**
2. [配置文件概览](#配置文件概览)
3. [核心配置](#核心配置)
4. [数据库配置](#数据库配置)
5. [监控配置](#监控配置)
6. [AI服务配置](#ai服务配置)
7. [告警配置](#告警配置)
8. [安全配置](#安全配置)
9. [性能优化](#性能优化)
10. [环境变量](#环境变量)

## 🚀 一键部署后配置（小白用户必读）

### 📋 部署完成检查清单

使用一键部署脚本（`quick-install.bat` 或 `quick-install.sh`）完成部署后，请按以下步骤进行基本配置：

#### ✅ 1. 验证服务状态

```bash
# 检查所有服务是否正常运行
docker-compose ps

# 应该看到以下服务都处于 "Up" 状态：
# - ai-monitor-backend
# - ai-monitor-frontend  
# - postgres
# - redis
# - nginx
```

#### ✅ 2. 首次登录系统

1. 打开浏览器访问：`http://localhost:8080`
2. 使用默认账户登录：
   - **用户名**: `admin`
   - **密码**: `admin123`

#### ✅ 3. 修改默认密码（重要）

1. 登录后点击右上角用户头像
2. 选择「个人设置」→「修改密码」
3. 设置强密码并保存

#### ✅ 4. 配置AI服务（可选但推荐）

1. 进入「系统设置」→「AI配置」
2. 配置以下任一AI服务：

**OpenAI配置：**
```yaml
ai:
  openai:
    api_key: "your-openai-api-key"
    model: "gpt-4"
    base_url: "https://api.openai.com/v1"
```

**Claude配置：**
```yaml
ai:
  claude:
    api_key: "your-claude-api-key"
    model: "claude-3-sonnet-20240229"
```

#### ✅ 5. 添加第一个监控目标

1. 进入「设备管理」→「添加设备」
2. 填写基本信息：
   ```
   设备名称: 我的服务器
   IP地址: 192.168.1.100
   设备类型: Linux服务器
   SSH端口: 22
   用户名: root
   密码: ******
   ```
3. 点击「测试连接」确认可达性
4. 保存设备，系统会自动部署监控Agent

#### ✅ 6. 配置告警通知（推荐）

1. 进入「告警管理」→「通知配置」
2. 配置邮件通知：
   ```yaml
   email:
     smtp_host: "smtp.gmail.com"
     smtp_port: 587
     username: "your-email@gmail.com"
     password: "your-app-password"
     from: "your-email@gmail.com"
   ```
3. 测试邮件发送功能

### 🔧 常用配置文件位置

一键部署后，主要配置文件位置：

```
项目根目录/
├── config.yaml              # 主配置文件（自动生成）
├── docker-compose.yml       # Docker编排配置
├── .env                     # 环境变量配置
└── data/
    ├── postgres/            # 数据库数据目录
    ├── redis/               # Redis数据目录
    └── logs/                # 日志目录
```

### 🚨 重要安全提醒

- ✅ **立即修改默认密码**
- ✅ **配置防火墙规则**（仅开放必要端口）
- ✅ **定期备份数据库**
- ✅ **监控系统资源使用**
- ✅ **及时更新系统补丁**

## 📁 配置文件概览

### 主要配置文件

```
ai-monitor/
├── config.yaml              # 主配置文件
├── deploy/
│   ├── docker-compose.yml   # Docker编排配置
│   ├── docker-deploy.yml    # 一体化Docker部署
│   ├── nginx.conf           # Nginx配置
│   └── redis.conf           # Redis配置
├── web/
│   └── .env                 # 前端环境配置
└── agents/
    └── config/
        ├── windows.yaml     # Windows Agent配置
        ├── linux.yaml       # Linux Agent配置
        └── macos.yaml       # macOS Agent配置
```

### 配置优先级

1. **环境变量** (最高优先级)
2. **命令行参数**
3. **配置文件**
4. **默认值** (最低优先级)

## ⚙️ 核心配置

### config.yaml 主配置

```yaml
# 服务器配置
server:
  host: "0.0.0.0"           # 监听地址
  port: 8080                # 监听端口
  mode: "release"           # 运行模式: debug/release/test
  read_timeout: 30s         # 读取超时
  write_timeout: 30s        # 写入超时
  max_header_bytes: 1048576 # 最大请求头大小
  
# 日志配置
logging:
  level: "info"             # 日志级别: debug/info/warn/error
  format: "json"            # 日志格式: json/text
  output: "stdout"          # 输出方式: stdout/file
  file_path: "/var/log/ai-monitor.log"  # 日志文件路径
  max_size: 100             # 单个日志文件最大大小(MB)
  max_backups: 10           # 保留的日志文件数量
  max_age: 30               # 日志文件保留天数
  compress: true            # 是否压缩旧日志

# 跨域配置
cors:
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:3001"
    - "https://your-domain.com"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "*"
  allow_credentials: true
```

### 运行模式说明

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| **debug** | 开发模式，详细日志，热重载 | 本地开发 |
| **release** | 生产模式，优化性能，简化日志 | 生产环境 |
| **test** | 测试模式，内存数据库 | 单元测试 |

## 🗄️ 数据库配置

### PostgreSQL 配置（推荐）

```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  username: "ai_monitor"
  password: "your_password"
  database: "ai_monitor"
  sslmode: "disable"        # SSL模式: disable/require/verify-full
  timezone: "Asia/Shanghai"
  
  # 连接池配置
  max_open_conns: 100       # 最大打开连接数
  max_idle_conns: 10        # 最大空闲连接数
  conn_max_lifetime: 3600s  # 连接最大生存时间
  conn_max_idle_time: 300s  # 连接最大空闲时间
  
  # 性能优化
  slow_query_threshold: 1s  # 慢查询阈值
  log_level: "warn"         # 数据库日志级别
```

### MySQL 配置

```yaml
database:
  type: "mysql"
  host: "localhost"
  port: 3306
  username: "ai_monitor"
  password: "your_password"
  database: "ai_monitor"
  charset: "utf8mb4"
  parse_time: true
  loc: "Asia/Shanghai"
  
  # MySQL特定配置
  max_allowed_packet: 67108864  # 最大数据包大小
  sql_mode: "STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO"
```

### SQLite 配置（开发/测试）

```yaml
database:
  type: "sqlite"
  database: "./data/ai_monitor.db"
  
  # SQLite特定配置
  cache_size: 2000          # 缓存页数
  busy_timeout: 5000        # 忙等超时(ms)
  journal_mode: "WAL"       # 日志模式: DELETE/TRUNCATE/PERSIST/MEMORY/WAL/OFF
  synchronous: "NORMAL"     # 同步模式: OFF/NORMAL/FULL/EXTRA
```

### Redis 配置

```yaml
redis:
  host: "localhost"
  port: 6379
  password: ""              # Redis密码
  database: 0               # 数据库编号
  
  # 连接池配置
  pool_size: 100            # 连接池大小
  min_idle_conns: 10        # 最小空闲连接
  max_conn_age: 3600s       # 连接最大年龄
  pool_timeout: 30s         # 连接池超时
  idle_timeout: 300s        # 空闲超时
  idle_check_frequency: 60s # 空闲检查频率
  
  # 集群配置（可选）
  cluster:
    enabled: false
    nodes:
      - "localhost:7000"
      - "localhost:7001"
      - "localhost:7002"
    max_redirects: 3
    read_only: false
```

## 📊 监控配置

### Prometheus 配置

```yaml
prometheus:
  enabled: true
  endpoint: "http://localhost:9090"
  push_gateway: "http://localhost:9091"
  
  # 指标配置
  metrics:
    namespace: "ai_monitor"   # 指标命名空间
    subsystem: "server"      # 子系统名称
    
  # 采集配置
  scrape_configs:
    - job_name: "ai-monitor"
      static_configs:
        - targets: ["localhost:8080"]
      scrape_interval: 30s
      metrics_path: "/metrics"
```

### Elasticsearch 配置

```yaml
elasticsearch:
  enabled: true
  endpoints:
    - "http://localhost:9200"
  username: "elastic"
  password: "your_password"
  
  # 索引配置
  indices:
    logs: "ai-monitor-logs"      # 日志索引
    metrics: "ai-monitor-metrics" # 指标索引
    events: "ai-monitor-events"   # 事件索引
    
  # 性能配置
  bulk_size: 1000           # 批量写入大小
  flush_interval: 10s       # 刷新间隔
  max_retries: 3            # 最大重试次数
  
  # 索引模板
  index_template:
    number_of_shards: 1
    number_of_replicas: 0
    refresh_interval: "30s"
```

### Agent 配置

```yaml
agent:
  # 通用配置
  server_url: "http://localhost:8080"
  api_key: "your_api_key"
  agent_id: "auto"          # auto为自动生成
  
  # 采集配置
  collection:
    interval: 30s           # 采集间隔
    timeout: 10s            # 采集超时
    batch_size: 100         # 批量大小
    
  # 采集项配置
  collectors:
    system:
      enabled: true
      cpu: true
      memory: true
      disk: true
      network: true
      process: true
      
    application:
      enabled: true
      jvm: true             # Java应用
      nodejs: true          # Node.js应用
      python: true          # Python应用
      
    custom:
      enabled: true
      scripts_path: "./scripts"
      
  # 缓存配置
  cache:
    enabled: true
    size: 1000              # 缓存条目数
    ttl: 300s               # 缓存TTL
```

## 🤖 AI服务配置

### OpenAI 配置

```yaml
ai:
  openai:
    enabled: true
    api_key: "sk-your-openai-api-key"
    base_url: "https://api.openai.com/v1"  # 可自定义API地址
    organization: ""        # 组织ID（可选）
    
    # 模型配置
    models:
      chat: "gpt-4"          # 对话模型
      embedding: "text-embedding-ada-002"  # 嵌入模型
      
    # 请求配置
    timeout: 30s
    max_retries: 3
    retry_delay: 1s
    
    # 速率限制
    rate_limit:
      requests_per_minute: 60
      tokens_per_minute: 40000
```

### Claude 配置

```yaml
ai:
  claude:
    enabled: false
    api_key: "your-claude-api-key"
    base_url: "https://api.anthropic.com"
    
    # 模型配置
    models:
      chat: "claude-3-sonnet-20240229"
      
    # 请求配置
    timeout: 30s
    max_retries: 3
    max_tokens: 4096
```

### AI功能配置

```yaml
ai:
  features:
    # 智能告警分析
    alert_analysis:
      enabled: true
      model: "gpt-4"
      confidence_threshold: 0.8
      
    # 异常检测
    anomaly_detection:
      enabled: true
      sensitivity: "medium"  # low/medium/high
      window_size: "1h"
      
    # 自动化建议
    automation_suggestions:
      enabled: true
      categories:
        - "performance"
        - "security"
        - "cost"
        
    # 报告生成
    report_generation:
      enabled: true
      schedule: "0 9 * * 1"  # 每周一9点
      recipients:
        - "admin@company.com"
```

## 🚨 告警配置

### 告警规则配置

```yaml
alerting:
  enabled: true
  
  # 默认规则
  default_rules:
    cpu_high:
      metric: "cpu_usage_percent"
      operator: ">"
      threshold: 80
      duration: "5m"
      severity: "warning"
      
    memory_high:
      metric: "memory_usage_percent"
      operator: ">"
      threshold: 85
      duration: "5m"
      severity: "warning"
      
    disk_full:
      metric: "disk_usage_percent"
      operator: ">"
      threshold: 90
      duration: "1m"
      severity: "critical"
      
  # 告警抑制
  inhibit_rules:
    - source_match:
        severity: "critical"
      target_match:
        severity: "warning"
      equal: ["instance"]
```

### 通知配置

```yaml
notifications:
  # 邮件通知
  email:
    enabled: true
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    from: "AI Monitor <noreply@ai-monitor.com>"
    
    # 邮件模板
    templates:
      alert: "./templates/alert_email.html"
      report: "./templates/report_email.html"
      
  # 钉钉通知
  dingtalk:
    enabled: false
    webhook_url: "https://oapi.dingtalk.com/robot/send?access_token=your_token"
    secret: "your_secret"
    
  # 企业微信通知
  wechat:
    enabled: false
    webhook_url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=your_key"
    
  # Slack通知
  slack:
    enabled: false
    webhook_url: "https://hooks.slack.com/services/your/webhook/url"
    channel: "#alerts"
    username: "AI Monitor"
```

## 🔒 安全配置

### JWT 配置

```yaml
jwt:
  secret: "your-super-secret-jwt-key"  # 建议使用环境变量
  issuer: "ai-monitor"
  audience: "ai-monitor-users"
  expires_in: 24h           # Token过期时间
  refresh_expires_in: 168h  # 刷新Token过期时间
  
  # 算法配置
  algorithm: "HS256"        # 签名算法
  
  # 安全选项
  require_exp: true         # 要求过期时间
  require_iat: true         # 要求签发时间
  require_nbf: false        # 要求生效时间
```

### 认证配置

```yaml
auth:
  # 密码策略
  password_policy:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_symbols: false
    max_age_days: 90
    
  # 登录限制
  login_limits:
    max_attempts: 5         # 最大尝试次数
    lockout_duration: 30m   # 锁定时间
    
  # 会话配置
  session:
    timeout: 24h            # 会话超时
    max_concurrent: 3       # 最大并发会话
    
  # LDAP集成（可选）
  ldap:
    enabled: false
    server: "ldap://localhost:389"
    bind_dn: "cn=admin,dc=company,dc=com"
    bind_password: "admin_password"
    search_base: "ou=users,dc=company,dc=com"
    search_filter: "(uid=%s)"
```

### API安全配置

```yaml
api:
  # 速率限制
  rate_limiting:
    enabled: true
    requests_per_minute: 100
    burst: 200
    
  # API密钥管理
  api_keys:
    enabled: true
    header_name: "X-API-Key"
    query_param: "api_key"
    
  # CORS配置
  cors:
    enabled: true
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["*"]
    max_age: 86400
    
  # 请求大小限制
  limits:
    max_request_size: 10MB
    max_multipart_memory: 32MB
```

## ⚡ 性能优化

### 缓存配置

```yaml
cache:
  # 内存缓存
  memory:
    enabled: true
    max_size: 1000          # 最大条目数
    ttl: 300s               # 默认TTL
    cleanup_interval: 60s   # 清理间隔
    
  # Redis缓存
  redis:
    enabled: true
    key_prefix: "ai_monitor:"
    default_ttl: 3600s
    
    # 缓存策略
    strategies:
      user_sessions: 86400s   # 用户会话
      api_responses: 300s     # API响应
      metrics_data: 60s       # 指标数据
```

### 数据库优化

```yaml
database:
  # 连接池优化
  pool:
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600s
    
  # 查询优化
  query:
    slow_query_threshold: 1s
    explain_slow_queries: true
    
  # 数据保留策略
  retention:
    metrics_data: 90d       # 指标数据保留90天
    log_data: 30d           # 日志数据保留30天
    event_data: 365d        # 事件数据保留1年
    
  # 分区策略
  partitioning:
    enabled: true
    strategy: "time"        # 按时间分区
    interval: "1d"          # 每日分区
```

### 监控数据优化

```yaml
monitoring:
  # 数据采集优化
  collection:
    batch_size: 1000        # 批量大小
    flush_interval: 10s     # 刷新间隔
    compression: true       # 数据压缩
    
  # 数据聚合
  aggregation:
    enabled: true
    intervals:
      - "1m"                # 1分钟聚合
      - "5m"                # 5分钟聚合
      - "1h"                # 1小时聚合
      - "1d"                # 1天聚合
      
  # 数据下采样
  downsampling:
    enabled: true
    rules:
      - source_resolution: "1m"
        target_resolution: "5m"
        retention: "7d"
      - source_resolution: "5m"
        target_resolution: "1h"
        retention: "30d"
```

## 🌍 环境变量

### 核心环境变量

```bash
# 服务配置
AI_MONITOR_HOST=0.0.0.0
AI_MONITOR_PORT=8080
AI_MONITOR_MODE=release

# 数据库配置
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=ai_monitor
DB_PASSWORD=your_password
DB_DATABASE=ai_monitor

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DATABASE=0

# JWT配置
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRES_IN=24h

# AI服务配置
OPENAI_API_KEY=sk-your-openai-api-key
OPENAI_BASE_URL=https://api.openai.com/v1
CLAUDE_API_KEY=your-claude-api-key

# 监控配置
PROMETHEUS_ENDPOINT=http://localhost:9090
ELASTICSEARCH_ENDPOINT=http://localhost:9200

# 邮件配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# 安全配置
API_RATE_LIMIT=100
MAX_REQUEST_SIZE=10MB

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
```

### Docker环境变量

```bash
# Docker Compose环境变量文件 (.env)
COMPOSE_PROJECT_NAME=ai-monitor
COMPOSE_FILE=docker-deploy.yml

# 服务版本
AI_MONITOR_VERSION=latest
POSTGRES_VERSION=15-alpine
REDIS_VERSION=7-alpine
PROMETHEUS_VERSION=latest
GRAFANA_VERSION=latest

# 数据目录
DATA_DIR=./data
LOGS_DIR=./logs
CONFIG_DIR=./config

# 网络配置
NETWORK_NAME=ai-monitor-network
SUBNET=172.20.0.0/16
```

### 前端环境变量

```bash
# web/.env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080
VITE_APP_TITLE=AI Monitor
VITE_APP_VERSION=1.0.0

# 功能开关
VITE_ENABLE_AI_FEATURES=true
VITE_ENABLE_DARK_MODE=true
VITE_ENABLE_I18N=true

# 第三方服务
VITE_SENTRY_DSN=your-sentry-dsn
VITE_GOOGLE_ANALYTICS_ID=your-ga-id
```

## 📝 配置验证

### 配置检查命令

```bash
# 检查配置文件语法
./ai-monitor config validate

# 显示当前配置
./ai-monitor config show

# 测试数据库连接
./ai-monitor config test-db

# 测试Redis连接
./ai-monitor config test-redis

# 测试AI服务连接
./ai-monitor config test-ai
```

### 配置模板生成

```bash
# 生成默认配置文件
./ai-monitor config init

# 生成生产环境配置
./ai-monitor config init --env production

# 生成开发环境配置
./ai-monitor config init --env development
```

---

## 📞 配置支持

如果在配置过程中遇到问题，请参考：

1. **[故障排除指南](TROUBLESHOOTING_GUIDE.md)**
2. **[API文档](API_DOCUMENTATION.md)**
3. **[部署指南](DEPLOYMENT_GUIDE.md)**

或联系技术支持：support@ai-monitor.com