# AI智能监控系统API文档

## 文档概述

本文档详细描述基于Go Gin框架构建的AI智能监控系统的所有REST API接口，包括请求格式、响应格式、错误码和使用示例。系统提供完整的监控、告警、AI分析和管理功能。

## 基础信息

- **Base URL**: `http://localhost:8080`
- **API版本**: `v3.8.5`
- **Content-Type**: `application/json`
- **认证方式**: JWT Bearer Token
- **框架**: Gin 1.9.1
- **Go版本**: 1.21.0
- **部署方式**: Docker Compose
- **Swagger文档**: `/swagger/index.html`

## API实现状态

- **认证接口**: ✅ 100% 完成 (登录、登出、刷新、个人资料)
- **用户管理**: ✅ 100% 完成 (用户CRUD、密码修改、权限管理)
- **系统监控**: ✅ 100% 完成 (CPU、内存、磁盘、网络、进程监控)
- **中间件监控**: ✅ 100% 完成 (MySQL、Redis、Kafka、Elasticsearch等)
- **APM监控**: ✅ 100% 完成 (应用性能、链路追踪、服务地图)
- **容器监控**: ✅ 100% 完成 (Docker、Kubernetes容器管理)
- **虚拟化监控**: ✅ 100% 完成 (VMware、Hyper-V虚拟机监控)
- **告警管理**: ✅ 100% 完成 (告警规则、历史记录、通知管理)
- **AI分析**: ✅ 100% 完成 (智能分析、趋势预测、知识库管理)
- **服务发现**: ✅ 100% 完成 (自动发现、Agent管理)
- **配置管理**: ✅ 100% 完成 (系统配置、API密钥管理)
- **安装指南**: ✅ 100% 完成 (Agent下载、安装指导)

## 认证说明

系统采用JWT (JSON Web Token) 认证机制，所有需要认证的API请求都必须在请求头中包含有效的JWT token：

### JWT Token格式
```http
Authorization: Bearer <access_token>
```

### JWT Token 特性
- **过期时间**: 24小时
- **刷新机制**: 支持refresh token自动续期
- **加密算法**: HS256
- **权限控制**: 基于RBAC角色权限模型

### Token获取示例

**请求**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**响应**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "created_at": "2024-01-01T00:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-02T00:00:00Z"
  }
}
```

## 通用响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "error message",
  "details": "detailed error information"
}
```

## 错误码说明

### HTTP状态码

| 状态码 | 说明 | 常见原因 | 解决方案 |
|--------|------|----------|----------|
| 200 | 请求成功 | - | - |
| 201 | 创建成功 | 资源创建成功 | - |
| 400 | 请求参数错误 | 参数格式不正确、必填参数缺失 | 检查请求参数格式和完整性 |
| 401 | 未授权访问 | Token无效、过期或缺失 | 重新登录获取有效Token |
| 403 | 权限不足 | 用户角色权限不够 | 联系管理员分配相应权限 |
| 404 | 资源不存在 | 请求的资源不存在 | 检查请求路径和资源ID |
| 409 | 资源冲突 | 资源已存在或状态冲突 | 检查资源状态或使用更新接口 |
| 422 | 数据验证失败 | 数据格式正确但业务逻辑验证失败 | 检查业务规则和数据有效性 |
| 429 | 请求频率限制 | 超过API调用频率限制 | 降低请求频率或联系管理员 |
| 500 | 服务器内部错误 | 服务器异常 | 联系技术支持 |
| 502 | 网关错误 | 上游服务不可用 | 检查服务状态或稍后重试 |
| 503 | 服务不可用 | 服务维护或过载 | 稍后重试或联系管理员 |

### 业务错误码

| 错误码 | 说明 | 示例场景 |
|--------|------|----------|
| 10001 | 用户名或密码错误 | 登录时凭据不正确 |
| 10002 | 用户不存在 | 查询不存在的用户 |
| 10003 | 用户已存在 | 注册时用户名重复 |
| 10004 | 密码强度不够 | 密码不符合安全策略 |
| 20001 | 告警规则不存在 | 查询不存在的告警规则 |
| 20002 | 告警规则已存在 | 创建重复的告警规则 |
| 20003 | 告警状态无效 | 告警状态转换不合法 |
| 30001 | 监控目标不存在 | 查询不存在的监控目标 |
| 30002 | 监控数据获取失败 | 数据源连接异常 |
| 40001 | AI分析服务不可用 | OpenAI/Claude API异常 |
| 40002 | AI分析超时 | 分析请求处理超时 |
| 50001 | 配置项不存在 | 查询不存在的配置 |
| 50002 | 配置格式错误 | 配置值格式不正确 |

## API接口分类

### 接口分类概览

| 分类 | 描述 | 接口数量 | 权限要求 |
|------|------|----------|----------|
| 系统接口 | 健康检查、版本信息等基础接口 | 2 | 无 |
| 认证接口 | 用户登录、登出、令牌管理 | 4 | 无/用户 |
| 用户管理 | 用户CRUD、密码修改、个人资料 | 8 | 用户/管理员 |
| 系统监控 | CPU、内存、磁盘、网络、进程监控 | 10 | 用户 |
| 中间件监控 | MySQL、Redis、Kafka、Elasticsearch等 | 15 | 用户 |
| APM监控 | 应用性能、链路追踪、服务地图 | 8 | 用户 |
| 容器监控 | Docker、Kubernetes容器管理 | 12 | 用户 |
| 虚拟化监控 | VMware、Hyper-V虚拟机监控 | 6 | 用户 |
| 告警管理 | 告警规则、历史记录、通知管理 | 10 | 用户 |
| AI分析 | 智能分析、趋势预测、知识库管理 | 8 | 用户 |
| 服务发现 | 自动发现、Agent管理 | 6 | 用户 |
| 配置管理 | 系统配置、API密钥管理 | 5 | 管理员 |
| 安装指南 | Agent下载、安装指导 | 4 | 无 |
| WebSocket | 实时数据推送 | 1 | 用户 |

### 权限说明

- **无**: 无需认证即可访问
- **用户**: 需要有效的JWT Token
- **管理员**: 需要管理员角色权限

## API接口列表

### 1. 系统接口

#### 1.1 健康检查

**接口地址**: `GET /health`

**接口描述**: 检查系统健康状态

**请求参数**: 无

**响应示例**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "database": "up",
    "redis": "up",
    "prometheus": "up"
  }
}
```

#### 1.2 版本信息

**接口地址**: `GET /version`

**接口描述**: 获取系统版本信息

**请求参数**: 无

**响应示例**:
```json
{
  "version": "1.0.0",
  "build_time": "2024-01-01T12:00:00Z",
  "git_commit": "abc123",
  "go_version": "go1.21.0"
}
```

### 2. 中间件监控接口

#### 2.1 数据库监控

##### 2.1.1 获取MySQL监控数据

**接口地址**: `GET /api/v1/middleware/mysql/{instance_id}/metrics`

**接口描述**: 获取MySQL实例监控指标

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `instance_id`: MySQL实例ID

**查询参数**:
- `start_time`: 开始时间 (RFC3339格式)
- `end_time`: 结束时间 (RFC3339格式)
- `step`: 时间间隔 (秒)

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "instance_id": "mysql-001",
    "metrics": {
      "connections": {
        "current": 45,
        "max": 151,
        "usage_rate": 0.298
      },
      "qps": {
        "current": 1250,
        "avg": 1180,
        "peak": 2500
      },
      "slow_queries": {
        "count": 12,
        "rate": 0.0096
      },
      "innodb": {
        "buffer_pool_usage": 0.85,
        "lock_waits": 3
      }
    },
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

##### 2.1.2 获取Redis监控数据

**接口地址**: `GET /api/v1/middleware/redis/{instance_id}/metrics`

**接口描述**: 获取Redis实例监控指标

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `instance_id`: Redis实例ID

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "instance_id": "redis-001",
    "metrics": {
      "memory": {
        "used": 1073741824,
        "max": 2147483648,
        "usage_rate": 0.5
      },
      "clients": {
        "connected": 25,
        "blocked": 0
      },
      "commands": {
        "processed": 1000000,
        "per_second": 500
      },
      "keyspace": {
        "keys": 50000,
        "expires": 5000
      }
    },
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

#### 2.2 消息队列监控

##### 2.2.1 获取Kafka监控数据

**接口地址**: `GET /api/v1/middleware/kafka/{cluster_id}/metrics`

**接口描述**: 获取Kafka集群监控指标

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `cluster_id`: Kafka集群ID

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "cluster_id": "kafka-cluster-001",
    "metrics": {
      "brokers": {
        "total": 3,
        "online": 3
      },
      "topics": {
        "count": 50,
        "partitions": 150
      },
      "messages": {
        "in_per_sec": 10000,
        "out_per_sec": 9500
      },
      "consumer_groups": {
        "active": 25,
        "lag": 1000
      }
    },
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### 3. APM应用性能监控接口

#### 3.1 链路追踪

##### 3.1.1 获取链路列表

**接口地址**: `GET /api/v1/apm/traces`

**接口描述**: 获取分布式链路追踪列表

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `service_name`: 服务名称 (可选)
- `operation_name`: 操作名称 (可选)
- `start_time`: 开始时间 (RFC3339格式)
- `end_time`: 结束时间 (RFC3339格式)
- `min_duration`: 最小耗时 (毫秒)
- `max_duration`: 最大耗时 (毫秒)
- `limit`: 返回数量限制 (默认100)
- `offset`: 偏移量 (默认0)

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "traces": [
      {
        "trace_id": "abc123def456",
        "root_span": {
          "span_id": "span001",
          "operation_name": "GET /api/users",
          "service_name": "user-service",
          "start_time": "2024-01-01T12:00:00Z",
          "duration": 250,
          "status": "ok"
        },
        "span_count": 8,
        "service_count": 3,
        "duration": 250,
        "error_count": 0
      }
    ],
    "total": 1500,
    "limit": 100,
    "offset": 0
  }
}
```

##### 3.1.2 获取链路详情

**接口地址**: `GET /api/v1/apm/traces/{trace_id}`

**接口描述**: 获取指定链路的详细信息

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `trace_id`: 链路ID

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "trace_id": "abc123def456",
    "spans": [
      {
        "span_id": "span001",
        "parent_span_id": null,
        "operation_name": "GET /api/users",
        "service_name": "user-service",
        "start_time": "2024-01-01T12:00:00.000Z",
        "duration": 250,
        "status": "ok",
        "tags": {
          "http.method": "GET",
          "http.url": "/api/users",
          "http.status_code": 200
        },
        "logs": [
          {
            "timestamp": "2024-01-01T12:00:00.100Z",
            "level": "info",
            "message": "Processing user request"
          }
        ]
      }
    ],
    "services": [
      {
        "name": "user-service",
        "span_count": 3,
        "error_count": 0
      }
    ]
  }
}
```

#### 3.2 应用性能指标

##### 3.2.1 获取服务性能概览

**接口地址**: `GET /api/v1/apm/services/{service_name}/overview`

**接口描述**: 获取服务性能概览

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `service_name`: 服务名称

**查询参数**:
- `start_time`: 开始时间 (RFC3339格式)
- `end_time`: 结束时间 (RFC3339格式)

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "service_name": "user-service",
    "metrics": {
      "throughput": {
        "requests_per_minute": 1200,
        "trend": "up"
      },
      "response_time": {
        "avg": 150,
        "p95": 300,
        "p99": 500
      },
      "error_rate": {
        "rate": 0.02,
        "count": 24
      },
      "apdex": {
        "score": 0.95,
        "threshold": 100
      }
    },
    "endpoints": [
      {
        "name": "GET /api/users",
        "requests_per_minute": 800,
        "avg_response_time": 120,
        "error_rate": 0.01
      }
    ]
  }
}
```

### 4. 容器监控接口

#### 4.1 Docker容器监控

##### 4.1.1 获取容器列表

**接口地址**: `GET /api/v1/containers`

**接口描述**: 获取Docker容器列表

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `status`: 容器状态 (running, stopped, all)
- `image`: 镜像名称过滤
- `limit`: 返回数量限制
- `offset`: 偏移量

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "containers": [
      {
        "id": "container123",
        "name": "web-server",
        "image": "nginx:1.21",
        "status": "running",
        "created_at": "2024-01-01T10:00:00Z",
        "ports": ["80:8080"],
        "labels": {
          "app": "web",
          "env": "production"
        }
      }
    ],
    "total": 50
  }
}
```

##### 4.1.2 获取容器监控指标

**接口地址**: `GET /api/v1/containers/{container_id}/metrics`

**接口描述**: 获取指定容器的监控指标

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `container_id`: 容器ID

**查询参数**:
- `start_time`: 开始时间
- `end_time`: 结束时间
- `step`: 时间间隔

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "container_id": "container123",
    "metrics": {
      "cpu": {
        "usage_percent": 25.5,
        "limit_cores": 2.0,
        "throttled_periods": 0
      },
      "memory": {
        "usage_bytes": 536870912,
        "limit_bytes": 1073741824,
        "usage_percent": 50.0
      },
      "network": {
        "rx_bytes": 1048576,
        "tx_bytes": 2097152,
        "rx_packets": 1000,
        "tx_packets": 1500
      },
      "disk": {
        "read_bytes": 10485760,
        "write_bytes": 5242880
      }
    },
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

#### 4.2 Kubernetes监控

##### 4.2.1 获取集群概览

**接口地址**: `GET /api/v1/kubernetes/clusters/{cluster_id}/overview`

**接口描述**: 获取Kubernetes集群概览

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `cluster_id`: 集群ID

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "cluster_id": "k8s-cluster-001",
    "cluster_info": {
      "name": "production-cluster",
      "version": "v1.28.0",
      "status": "healthy"
    },
    "nodes": {
      "total": 5,
      "ready": 5,
      "not_ready": 0
    },
    "pods": {
      "total": 150,
      "running": 145,
      "pending": 3,
      "failed": 2
    },
    "namespaces": {
      "total": 10,
      "active": 10
    },
    "resources": {
      "cpu": {
        "total_cores": 20,
        "used_cores": 12.5,
        "usage_percent": 62.5
      },
      "memory": {
        "total_bytes": 85899345920,
        "used_bytes": 51539607552,
        "usage_percent": 60.0
      }
    }
  }
}
```

### 5. Agent管理接口

#### 5.1 Agent下载

##### 5.1.1 获取可用Agent列表

**接口地址**: `GET /api/v1/agents`

**接口描述**: 获取可下载的Agent列表

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `platform`: 平台类型 (windows, linux, container, apm)
- `architecture`: 架构 (amd64, arm64)
- `version`: 版本号

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "agents": [
      {
        "id": "agent-001",
        "name": "Windows System Monitor Agent",
        "version": "1.0.0",
        "platform": "windows",
        "architecture": "amd64",
        "type": "system",
        "size": 45678901,
        "checksum": "sha256:abc123...",
        "release_date": "2024-01-01T00:00:00Z",
        "description": "Windows系统监控Agent，支持CPU、内存、磁盘、网络监控"
      }
    ],
    "total": 12
  }
}
```

##### 5.1.2 下载Agent

**接口地址**: `GET /api/v1/agents/{agent_id}/download`

**接口描述**: 下载指定Agent

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `agent_id`: Agent ID

**响应**: 二进制文件流

**响应头**:
```
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="agent-windows-amd64-v1.0.0.exe"
Content-Length: 45678901
X-Checksum: sha256:abc123...
```

#### 5.2 Agent管理

##### 5.2.1 获取已部署Agent列表

**接口地址**: `GET /api/v1/agents/deployed`

**接口描述**: 获取已部署的Agent列表

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `status`: Agent状态 (online, offline, error)
- `platform`: 平台类型
- `limit`: 返回数量限制
- `offset`: 偏移量

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "agents": [
      {
        "id": "deployed-agent-001",
        "host_id": "host-001",
        "host_name": "web-server-01",
        "agent_version": "1.0.0",
        "platform": "linux",
        "status": "online",
        "last_heartbeat": "2024-01-01T12:00:00Z",
        "deployed_at": "2024-01-01T10:00:00Z",
        "metrics": {
          "cpu_usage": 25.5,
          "memory_usage": 60.0,
          "disk_usage": 45.0
        }
      }
    ],
    "total": 100,
    "summary": {
      "online": 95,
      "offline": 3,
      "error": 2
    }
  }
}
```

### 6. 用户管理接口

#### 2.1 用户注册

**接口地址**: `POST /api/v1/auth/register`

**接口描述**: 用户注册

**请求参数**:
```json
{
  "username": "string",     // 用户名，3-50字符，字母数字
  "email": "string",        // 邮箱地址
  "password": "string",     // 密码，至少8位，包含大小写字母和数字
  "confirm_password": "string" // 确认密码
}
```

**响应示例**:
```json
{
  "code": 201,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "role": "user",
      "status": "active",
      "created_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 2.2 用户登录

**接口地址**: `POST /api/v1/auth/login`

**接口描述**: 用户登录

**请求参数**:
```json
{
  "username": "string",  // 用户名或邮箱
  "password": "string"   // 密码
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "role": "user",
      "status": "active"
    }
  }
}
```

#### 2.3 刷新令牌

**接口地址**: `POST /api/v1/auth/refresh`

**接口描述**: 刷新访问令牌

**请求参数**:
```json
{
  "refresh_token": "string"  // 刷新令牌
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

#### 2.4 获取用户资料

**接口地址**: `GET /api/v1/users/profile`

**接口描述**: 获取当前用户资料

**请求头**: `Authorization: Bearer <token>`

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "role": "user",
      "status": "active",
      "last_login_at": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 2.5 更新用户资料

**接口地址**: `PUT /api/v1/users/profile`

**接口描述**: 更新当前用户资料

**请求头**: `Authorization: Bearer <token>`

**请求参数**:
```json
{
  "email": "string",     // 邮箱地址（可选）
  "avatar": "string"     // 头像URL（可选）
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Profile updated successfully",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "newemail@example.com",
      "role": "user",
      "status": "active",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 2.6 修改密码

**接口地址**: `PUT /api/v1/users/password`

**接口描述**: 修改用户密码

**请求头**: `Authorization: Bearer <token>`

**请求参数**:
```json
{
  "old_password": "string",     // 旧密码
  "new_password": "string",     // 新密码
  "confirm_password": "string"  // 确认新密码
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Password changed successfully",
  "data": null
}
```

#### 2.7 获取用户列表（管理员）

**接口地址**: `GET /api/v1/users`

**接口描述**: 获取用户列表（需要管理员权限）

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10）
- `search`: 搜索关键词
- `role`: 角色筛选
- `status`: 状态筛选

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "users": [
      {
        "id": 1,
        "username": "testuser",
        "email": "test@example.com",
        "role": "user",
        "status": "active",
        "last_login_at": "2024-01-01T12:00:00Z",
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "pages": 10
    }
  }
}
```

### 3. 告警管理接口

#### 3.1 创建告警规则

**接口地址**: `POST /api/v1/alerts/rules`

**接口描述**: 创建告警规则

**请求头**: `Authorization: Bearer <token>`

**请求参数**:
```json
{
  "name": "string",              // 规则名称
  "description": "string",       // 规则描述
  "query": "string",             // Prometheus查询语句
  "threshold_value": 80.0,       // 阈值
  "threshold_operator": ">",     // 比较操作符
  "severity": "warning",         // 严重程度：info, warning, critical
  "enabled": true,               // 是否启用
  "labels": {                    // 标签
    "team": "ops",
    "service": "web"
  },
  "annotations": {               // 注释
    "summary": "High CPU usage",
    "description": "CPU usage is above 80%"
  }
}
```

**响应示例**:
```json
{
  "code": 201,
  "message": "Alert rule created successfully",
  "data": {
    "rule": {
      "id": 1,
      "name": "High CPU Usage",
      "description": "Alert when CPU usage exceeds 80%",
      "query": "cpu_usage_percent > 80",
      "threshold_value": 80.0,
      "threshold_operator": ">",
      "severity": "warning",
      "enabled": true,
      "created_by": 1,
      "created_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 3.2 获取告警规则

**接口地址**: `GET /api/v1/alerts/rules/{id}`

**接口描述**: 获取指定告警规则

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `id`: 告警规则ID

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "rule": {
      "id": 1,
      "name": "High CPU Usage",
      "description": "Alert when CPU usage exceeds 80%",
      "query": "cpu_usage_percent > 80",
      "threshold_value": 80.0,
      "threshold_operator": ">",
      "severity": "warning",
      "enabled": true,
      "labels": {
        "team": "ops",
        "service": "web"
      },
      "annotations": {
        "summary": "High CPU usage",
        "description": "CPU usage is above 80%"
      },
      "created_by": 1,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 3.3 获取告警规则列表

**接口地址**: `GET /api/v1/alerts/rules`

**接口描述**: 获取告警规则列表

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10）
- `search`: 搜索关键词
- `severity`: 严重程度筛选
- `enabled`: 启用状态筛选

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "rules": [
      {
        "id": 1,
        "name": "High CPU Usage",
        "description": "Alert when CPU usage exceeds 80%",
        "severity": "warning",
        "enabled": true,
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 50,
      "pages": 5
    }
  }
}
```

#### 3.4 更新告警规则

**接口地址**: `PUT /api/v1/alerts/rules/{id}`

**接口描述**: 更新告警规则

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `id`: 告警规则ID

**请求参数**: 同创建告警规则

**响应示例**:
```json
{
  "code": 200,
  "message": "Alert rule updated successfully",
  "data": {
    "rule": {
      "id": 1,
      "name": "High CPU Usage",
      "description": "Updated description",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 3.5 删除告警规则

**接口地址**: `DELETE /api/v1/alerts/rules/{id}`

**接口描述**: 删除告警规则

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `id`: 告警规则ID

**响应示例**:
```json
{
  "code": 200,
  "message": "Alert rule deleted successfully",
  "data": null
}
```

#### 3.6 获取告警列表

**接口地址**: `GET /api/v1/alerts`

**接口描述**: 获取告警列表

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10）
- `status`: 状态筛选（active, resolved, silenced）
- `severity`: 严重程度筛选
- `start_time`: 开始时间
- `end_time`: 结束时间

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "alerts": [
      {
        "id": 1,
        "rule_id": 1,
        "title": "High CPU Usage Alert",
        "description": "CPU usage is 85%",
        "level": "warning",
        "status": "active",
        "value": 85.0,
        "labels": {
          "instance": "server-01",
          "job": "node-exporter"
        },
        "starts_at": "2024-01-01T12:00:00Z",
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 25,
      "pages": 3
    }
  }
}
```

#### 3.7 确认告警

**接口地址**: `POST /api/v1/alerts/{id}/acknowledge`

**接口描述**: 确认告警

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `id`: 告警ID

**请求参数**:
```json
{
  "comment": "string"  // 确认备注（可选）
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Alert acknowledged successfully",
  "data": {
    "alert": {
      "id": 1,
      "status": "acknowledged",
      "acknowledged_by": 1,
      "acknowledged_at": "2024-01-01T12:00:00Z",
      "comment": "Investigating the issue"
    }
  }
}
```

#### 3.8 解决告警

**接口地址**: `POST /api/v1/alerts/{id}/resolve`

**接口描述**: 解决告警

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `id`: 告警ID

**请求参数**:
```json
{
  "comment": "string"  // 解决备注（可选）
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Alert resolved successfully",
  "data": {
    "alert": {
      "id": 1,
      "status": "resolved",
      "resolved_by": 1,
      "resolved_at": "2024-01-01T12:00:00Z",
      "comment": "Issue fixed by restarting service"
    }
  }
}
```

### 4. 监控数据接口

#### 4.1 创建监控目标

**接口地址**: `POST /api/v1/monitoring/targets`

**接口描述**: 创建监控目标

**请求头**: `Authorization: Bearer <token>`

**请求参数**:
```json
{
  "name": "string",        // 目标名称
  "type": "string",        // 目标类型：server, application, database
  "endpoint": "string",    // 监控端点
  "labels": {              // 标签
    "environment": "production",
    "team": "ops"
  },
  "scrape_interval": "30s", // 采集间隔
  "enabled": true          // 是否启用
}
```

**响应示例**:
```json
{
  "code": 201,
  "message": "Monitoring target created successfully",
  "data": {
    "target": {
      "id": 1,
      "name": "Web Server 01",
      "type": "server",
      "endpoint": "http://192.168.1.100:9100/metrics",
      "labels": {
        "environment": "production",
        "team": "ops"
      },
      "scrape_interval": "30s",
      "enabled": true,
      "status": "up",
      "created_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 4.2 获取监控目标列表

**接口地址**: `GET /api/v1/monitoring/targets`

**接口描述**: 获取监控目标列表

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10）
- `type`: 类型筛选
- `status`: 状态筛选
- `search`: 搜索关键词

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "targets": [
      {
        "id": 1,
        "name": "Web Server 01",
        "type": "server",
        "endpoint": "http://192.168.1.100:9100/metrics",
        "status": "up",
        "last_scrape": "2024-01-01T12:00:00Z",
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 15,
      "pages": 2
    }
  }
}
```

#### 4.3 查询指标数据

**接口地址**: `GET /api/v1/monitoring/metrics`

**接口描述**: 查询指标数据

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `query`: Prometheus查询语句（必需）
- `start`: 开始时间（RFC3339格式）
- `end`: 结束时间（RFC3339格式）
- `step`: 步长（如：30s, 1m, 5m）

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "result_type": "matrix",
    "result": [
      {
        "metric": {
          "__name__": "cpu_usage_percent",
          "instance": "192.168.1.100:9100",
          "job": "node-exporter"
        },
        "values": [
          [1704110400, "75.5"],
          [1704110430, "78.2"],
          [1704110460, "82.1"]
        ]
      }
    ]
  }
}
```

#### 4.4 获取系统指标

**接口地址**: `GET /api/v1/monitoring/system-metrics`

**接口描述**: 获取系统指标概览

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `instance`: 实例标识（可选）
- `duration`: 时间范围（如：1h, 24h, 7d）

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "cpu": {
      "usage_percent": 75.5,
      "load_1m": 1.2,
      "load_5m": 1.1,
      "load_15m": 1.0
    },
    "memory": {
      "total_bytes": 8589934592,
      "used_bytes": 6442450944,
      "available_bytes": 2147483648,
      "usage_percent": 75.0
    },
    "disk": {
      "total_bytes": 107374182400,
      "used_bytes": 53687091200,
      "available_bytes": 53687091200,
      "usage_percent": 50.0
    },
    "network": {
      "bytes_sent": 1073741824,
      "bytes_recv": 2147483648,
      "packets_sent": 1000000,
      "packets_recv": 2000000
    }
  }
}
```

#### 4.5 创建仪表板

**接口地址**: `POST /api/v1/monitoring/dashboards`

**接口描述**: 创建监控仪表板

**请求头**: `Authorization: Bearer <token>`

**请求参数**:
```json
{
  "name": "string",        // 仪表板名称
  "description": "string", // 描述
  "config": {              // 仪表板配置
    "panels": [
      {
        "title": "CPU Usage",
        "type": "graph",
        "query": "cpu_usage_percent",
        "position": {
          "x": 0,
          "y": 0,
          "width": 6,
          "height": 4
        }
      }
    ]
  },
  "tags": ["system", "performance"] // 标签
}
```

**响应示例**:
```json
{
  "code": 201,
  "message": "Dashboard created successfully",
  "data": {
    "dashboard": {
      "id": 1,
      "name": "System Overview",
      "description": "System performance overview",
      "config": {
        "panels": [...]
      },
      "tags": ["system", "performance"],
      "created_by": 1,
      "created_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

### 5. AI分析接口

#### 5.1 AI告警分析

**接口地址**: `POST /api/v1/ai/analyze-alert`

**接口描述**: 使用AI分析告警

**请求头**: `Authorization: Bearer <token>`

**请求参数**:
```json
{
  "alert_id": 1,           // 告警ID
  "context": "string"      // 额外上下文信息（可选）
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "AI analysis completed",
  "data": {
    "analysis": {
      "id": 1,
      "alert_id": 1,
      "analysis_type": "alert",
      "summary": "High CPU usage detected on web server",
      "root_cause": "Increased traffic causing CPU spike",
      "impact_assessment": "Medium impact - may affect response times",
      "recommendations": [
        "Scale up the server resources",
        "Implement load balancing",
        "Optimize application code"
      ],
      "confidence_score": 0.85,
      "created_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 5.2 AI性能分析

**接口地址**: `POST /api/v1/ai/analyze-performance`

**接口描述**: 使用AI分析系统性能

**请求头**: `Authorization: Bearer <token>`

**请求参数**:
```json
{
  "target_id": 1,          // 监控目标ID
  "time_range": {          // 时间范围
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-01T23:59:59Z"
  },
  "metrics": [             // 要分析的指标
    "cpu_usage_percent",
    "memory_usage_percent",
    "disk_usage_percent"
  ]
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Performance analysis completed",
  "data": {
    "analysis": {
      "id": 2,
      "target_id": 1,
      "analysis_type": "performance",
      "summary": "System performance is within normal range",
      "bottlenecks": [
        {
          "metric": "cpu_usage_percent",
          "severity": "medium",
          "description": "CPU usage spikes during peak hours"
        }
      ],
      "trends": [
        {
          "metric": "memory_usage_percent",
          "trend": "increasing",
          "description": "Memory usage has been gradually increasing"
        }
      ],
      "recommendations": [
        "Consider upgrading CPU for better performance",
        "Monitor memory usage closely",
        "Implement caching to reduce CPU load"
      ],
      "confidence_score": 0.92,
      "created_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 5.3 获取AI分析历史

**接口地址**: `GET /api/v1/ai/analyses`

**接口描述**: 获取AI分析历史记录

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10）
- `type`: 分析类型筛选（alert, performance）
- `start_time`: 开始时间
- `end_time`: 结束时间

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "analyses": [
      {
        "id": 1,
        "analysis_type": "alert",
        "summary": "High CPU usage detected",
        "confidence_score": 0.85,
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 50,
      "pages": 5
    }
  }
}
```

### 6. 系统配置接口

#### 6.1 获取配置

**接口地址**: `GET /api/v1/config/{key}`

**接口描述**: 获取系统配置

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `key`: 配置键名

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "config": {
      "key": "smtp_settings",
      "value": {
        "host": "smtp.gmail.com",
        "port": 587,
        "username": "admin@example.com",
        "password": "***"
      },
      "category": "email",
      "description": "SMTP email settings",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 6.2 更新配置

**接口地址**: `PUT /api/v1/config/{key}`

**接口描述**: 更新系统配置

**请求头**: `Authorization: Bearer <token>`

**路径参数**:
- `key`: 配置键名

**请求参数**:
```json
{
  "value": {},             // 配置值
  "description": "string"  // 配置描述（可选）
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Configuration updated successfully",
  "data": {
    "config": {
      "key": "smtp_settings",
      "value": {
        "host": "smtp.gmail.com",
        "port": 587,
        "username": "admin@example.com"
      },
      "updated_at": "2024-01-01T12:00:00Z"
    }
  }
}
```

#### 6.3 获取配置列表

**接口地址**: `GET /api/v1/config`

**接口描述**: 获取配置列表

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `category`: 配置分类筛选
- `search`: 搜索关键词

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "configs": [
      {
        "key": "smtp_settings",
        "category": "email",
        "description": "SMTP email settings",
        "updated_at": "2024-01-01T12:00:00Z"
      }
    ]
  }
}
```

### 7. 审计日志接口

#### 7.1 获取审计日志

**接口地址**: `GET /api/v1/audit/logs`

**接口描述**: 获取审计日志列表

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10）
- `user_id`: 用户ID筛选
- `action`: 操作类型筛选
- `resource`: 资源类型筛选
- `start_time`: 开始时间
- `end_time`: 结束时间

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "logs": [
      {
        "id": 1,
        "user_id": 1,
        "username": "admin",
        "action": "CREATE",
        "resource": "alert_rule",
        "resource_id": "1",
        "details": {
          "rule_name": "High CPU Usage",
          "threshold": 80
        },
        "ip_address": "192.168.1.100",
        "user_agent": "Mozilla/5.0...",
        "status": "success",
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1000,
      "pages": 100
    }
  }
}
```

#### 7.2 获取审计统计

**接口地址**: `GET /api/v1/audit/stats`

**接口描述**: 获取审计统计信息

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `period`: 统计周期（day, week, month）
- `start_time`: 开始时间
- `end_time`: 结束时间

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "stats": {
      "total_logs": 10000,
      "success_count": 9500,
      "error_count": 500,
      "success_rate": 95.0,
      "today_logs": 150,
      "top_actions": [
        {
          "action": "LOGIN",
          "count": 2000
        },
        {
          "action": "CREATE",
          "count": 1500
        }
      ],
      "top_users": [
        {
          "user_id": 1,
          "username": "admin",
          "count": 500
        }
      ],
      "hourly_stats": [
        {
          "hour": "2024-01-01T00:00:00Z",
          "count": 50
        }
      ]
    }
  }
}
```

#### 7.3 导出审计日志

**接口地址**: `GET /api/v1/audit/export`

**接口描述**: 导出审计日志

**请求头**: `Authorization: Bearer <token>`

**查询参数**:
- `format`: 导出格式（json, csv）
- `start_time`: 开始时间
- `end_time`: 结束时间
- `filters`: 筛选条件

**响应**: 文件下载

### 8. WebSocket接口

#### 8.1 WebSocket连接

**接口地址**: `WS /ws`

**接口描述**: 建立WebSocket连接

**连接参数**:
- `token`: JWT访问令牌（查询参数）

**消息格式**:

**订阅消息**:
```json
{
  "type": "subscribe",
  "topics": ["alerts", "metrics", "system"]
}
```

**取消订阅消息**:
```json
{
  "type": "unsubscribe",
  "topics": ["alerts"]
}
```

**告警推送消息**:
```json
{
  "type": "alert",
  "data": {
    "id": 1,
    "title": "High CPU Usage",
    "description": "CPU usage is 85%",
    "level": "warning",
    "status": "active",
    "created_at": "2024-01-01T12:00:00Z"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**指标推送消息**:
```json
{
  "type": "metric",
  "data": {
    "name": "cpu_usage_percent",
    "value": 75.5,
    "labels": {
      "instance": "server-01",
      "job": "node-exporter"
    },
    "timestamp": "2024-01-01T12:00:00Z"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 使用示例

### 1. 用户登录和获取数据

```bash
# 1. 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

# 2. 使用返回的token获取用户资料
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 3. 查询指标数据
curl -X GET "http://localhost:8080/api/v1/monitoring/metrics?query=cpu_usage_percent&start=2024-01-01T00:00:00Z&end=2024-01-01T23:59:59Z&step=5m" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 2. 创建告警规则

```bash
curl -X POST http://localhost:8080/api/v1/alerts/rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "name": "High CPU Usage",
    "description": "Alert when CPU usage exceeds 80%",
    "query": "cpu_usage_percent > 80",
    "threshold_value": 80.0,
    "threshold_operator": ">",
    "severity": "warning",
    "enabled": true,
    "labels": {
      "team": "ops",
      "service": "web"
    },
    "annotations": {
      "summary": "High CPU usage detected",
      "description": "CPU usage is above 80%"
    }
  }'
```

### 3. AI分析告警

```bash
curl -X POST http://localhost:8080/api/v1/ai/analyze-alert \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "alert_id": 1,
    "context": "Server has been experiencing high load recently"
  }'
```

### 4. WebSocket连接示例

```javascript
// JavaScript WebSocket客户端示例
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = function() {
  console.log('WebSocket connected');
  
  // 订阅告警和指标
  ws.send(JSON.stringify({
    type: 'subscribe',
    topics: ['alerts', 'metrics']
  }));
};

ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('Received message:', message);
  
  if (message.type === 'alert') {
    // 处理告警消息
    console.log('New alert:', message.data);
  } else if (message.type === 'metric') {
    // 处理指标消息
    console.log('New metric:', message.data);
  }
};

ws.onerror = function(error) {
  console.error('WebSocket error:', error);
};

ws.onclose = function() {
  console.log('WebSocket disconnected');
};
```

## 错误处理

### 常见错误码

| 错误码 | 错误信息 | 说明 |
|--------|----------|------|
| 400001 | Invalid request parameters | 请求参数无效 |
| 401001 | Authentication required | 需要认证 |
| 401002 | Invalid token | 无效的令牌 |
| 401003 | Token expired | 令牌已过期 |
| 403001 | Insufficient permissions | 权限不足 |
| 404001 | Resource not found | 资源不存在 |
| 409001 | Resource already exists | 资源已存在 |
| 500001 | Internal server error | 服务器内部错误 |
| 500002 | Database error | 数据库错误 |
| 500003 | External service error | 外部服务错误 |

### 错误响应示例

```json
{
  "code": 401002,
  "message": "Invalid token",
  "details": "The provided JWT token is malformed or invalid",
  "timestamp": "2024-01-01T12:00:00Z",
  "path": "/api/v1/users/profile"
}
```

## 限流说明

### 限流策略

| 接口类型 | 限制 | 时间窗口 |
|----------|------|----------|
| 登录接口 | 5次/IP | 15分钟 |
| 普通API | 1000次/用户 | 1小时 |
| AI分析接口 | 10次/用户 | 1小时 |
| WebSocket | 100个连接/用户 | - |

### 限流响应

```json
{
  "code": 429,
  "message": "Rate limit exceeded",
  "details": "Too many requests, please try again later",
  "retry_after": 3600
}
```

## 版本更新

### v1.0.0 (2024-01-01)
- 初始版本发布
- 基础用户管理功能
- 告警管理功能
- 监控数据查询功能
- AI分析功能
- WebSocket实时推送

### 后续版本计划
- v1.1.0: 增加更多AI分析功能
- v1.2.0: 支持更多监控数据源
- v1.3.0: 增强用户权限管理
- v2.0.0: 重构API架构，支持GraphQL

---

**注意**: 本文档会随着系统功能的更新而持续更新，请关注版本变更说明。