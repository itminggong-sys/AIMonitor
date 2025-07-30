# AIMonitor Agent Programs

本目录包含各种平台和中间件的监控代理程序，用于向AIMonitor平台注册并提供监控数据。

## 支持的Agent类型

### 系统监控Agent
- **Windows Agent**: 监控Windows服务器的CPU、内存、磁盘、网络等指标
- **Linux Agent**: 监控Linux服务器的系统资源和性能指标
- **macOS Agent**: 监控macOS系统的资源使用情况

### 中间件监控Agent
- **Redis Agent**: 监控Redis缓存服务的性能和状态
- **MySQL Agent**: 监控MySQL数据库的性能指标
- **PostgreSQL Agent**: 监控PostgreSQL数据库的运行状态
- **Nginx Agent**: 监控Nginx Web服务器的访问和性能
- **Apache Agent**: 监控Apache Web服务器
- **Kafka Agent**: 监控Kafka消息队列
- **RabbitMQ Agent**: 监控RabbitMQ消息队列
- **Elasticsearch Agent**: 监控Elasticsearch搜索引擎
- **MongoDB Agent**: 监控MongoDB数据库

### 容器监控Agent
- **Docker Agent**: 监控Docker容器的资源使用和状态
- **Kubernetes Agent**: 监控Kubernetes集群和Pod
- **Podman Agent**: 监控Podman容器

### 虚拟化平台Agent
- **VMware Agent**: 监控VMware虚拟机
- **Hyper-V Agent**: 监控Hyper-V虚拟机
- **KVM Agent**: 监控KVM虚拟机
- **Proxmox Agent**: 监控Proxmox虚拟化平台

### APM监控Agent
- **Java APM Agent**: 监控Java应用性能
- **Node.js APM Agent**: 监控Node.js应用
- **Python APM Agent**: 监控Python应用
- **Go APM Agent**: 监控Go应用
- **PHP APM Agent**: 监控PHP应用
- **.NET APM Agent**: 监控.NET应用

## 安装和使用

1. 从AIMonitor平台下载对应的Agent程序
2. 根据平台类型执行相应的安装脚本
3. 配置Agent连接参数
4. 启动Agent服务
5. 在AIMonitor平台查看监控数据

## Agent配置

所有Agent都支持以下基本配置：

```yaml
server:
  host: "your-aimonitor-host"
  port: 8080
  api_key: "your-api-key"
  ssl: true

agent:
  name: "agent-name"
  type: "agent-type"
  interval: 30s
  timeout: 10s

metrics:
  enabled: true
  interval: 30s
  
logging:
  level: "info"
  file: "/var/log/aimonitor-agent.log"
```

## 开发指南

如需开发自定义Agent，请参考现有Agent的实现，遵循以下规范：

1. 使用统一的配置格式
2. 实现标准的注册和心跳机制
3. 支持指标数据的标准化格式
4. 包含错误处理和重连机制
5. 提供详细的日志记录