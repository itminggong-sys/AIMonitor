# AI智能监控系统开发指南

## 文档概述

本文档为AI智能监控系统的开发人员提供完整的开发指南，包括环境搭建、代码规范、API开发、前端开发等内容。

## 版本信息

- **系统版本**: v3.8.5
- **Go版本**: 1.21.0
- **Node.js版本**: 18.17.0
- **React版本**: 18.2.0
- **TypeScript版本**: 5.0.4

## 📋 目录

1. [开发环境搭建](#开发环境搭建)
2. [项目结构](#项目结构)
3. [代码规范](#代码规范)
4. [API开发](#api开发)
5. [前端开发](#前端开发)
6. [数据库开发](#数据库开发)
7. [测试指南](#测试指南)
8. [调试技巧](#调试技巧)
9. [性能优化](#性能优化)
10. [部署流程](#部署流程)
11. [贡献指南](#贡献指南)

## 🛠️ 开发环境搭建

### 系统要求

| 组件 | 最低版本 | 推荐版本 | 当前版本 | 说明 |
|------|----------|----------|----------|------|
| **Go** | 1.19+ | 1.21+ | 1.21.0 | 后端开发语言 |
| **Node.js** | 16+ | 18+ | 18.17.0 | 前端构建工具 |
| **PostgreSQL** | 12+ | 15+ | 15.3 | 主数据库 |
| **Redis** | 6+ | 7+ | 7.0.11 | 缓存数据库 |
| **Git** | 2.30+ | 最新版 | 2.40+ | 版本控制 |
| **Docker** | 20+ | 最新版 | 24.0.2 | 容器化部署 |

### 环境配置

#### 1. Go 开发环境

```bash
# 安装 Go (Linux/macOS)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GO111MODULE=on' >> ~/.bashrc
echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### 2. Node.js 开发环境

```bash
# 使用 nvm 安装 Node.js
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc
nvm install 18
nvm use 18

# 配置 npm 镜像
npm config set registry https://registry.npmmirror.com

# 验证安装
node --version
npm --version
```

#### 3. 数据库环境

```bash
# PostgreSQL 安装 (Ubuntu/Debian)
sudo apt update
sudo apt install postgresql postgresql-contrib

# 启动服务
sudo systemctl start postgresql
sudo systemctl enable postgresql

# 创建开发数据库
sudo -u postgres psql
CREATE DATABASE ai_monitor_dev;
CREATE USER ai_monitor WITH PASSWORD 'dev_password';
GRANT ALL PRIVILEGES ON DATABASE ai_monitor_dev TO ai_monitor;
\q

# Redis 安装
sudo apt install redis-server
sudo systemctl start redis
sudo systemctl enable redis
```

#### 4. 开发工具推荐

**IDE/编辑器**：
- **GoLand** (JetBrains) - Go 开发首选
- **VS Code** - 轻量级，插件丰富
- **Vim/Neovim** - 命令行编辑器

**必装插件** (VS Code)：
```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-typescript-next",
    "bradlc.vscode-tailwindcss",
    "esbenp.prettier-vscode",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml",
    "ms-python.python",
    "ms-vscode.vscode-docker"
  ]
}
```

### 项目克隆与初始化

```bash
# 克隆项目
git clone https://github.com/your-org/ai-monitor.git
cd ai-monitor

# 初始化后端依赖
go mod download
go mod tidy

# 初始化前端依赖
cd frontend
npm install
cd ..

# 复制配置文件
cp config/config.example.yaml config/config.yaml

# 编辑配置文件
vim config/config.yaml
```

### 开发配置文件

```yaml
# config/config.dev.yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"  # debug/release
  
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  name: "ai_monitor_dev"
  username: "ai_monitor"
  password: "dev_password"
  ssl_mode: "disable"
  max_open_conns: 10
  max_idle_conns: 5
  
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  
logging:
  level: "debug"
  format: "text"  # text/json
  output: "stdout"  # stdout/file
  
ai:
  openai:
    api_key: "your-dev-api-key"
    base_url: "https://api.openai.com/v1"
    model: "gpt-3.5-turbo"
  
jwt:
  secret: "dev-jwt-secret-key"
  expire_hours: 24
```

## 📁 项目结构

```
ai-monitor/
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go
├── internal/               # 内部包（不对外暴露）
│   ├── handlers/          # HTTP 处理器
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── monitoring_handler.go
│   │   ├── alert_handler.go
│   │   ├── ai_handler.go
│   │   ├── middleware_handler.go
│   │   ├── apm_handler.go
│   │   ├── container_handler.go
│   │   └── config_handler.go
│   ├── middleware/        # 中间件
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logger.go
│   │   └── rate_limit.go
│   ├── router/            # 路由定义
│   │   └── router.go
│   ├── config/            # 配置管理
│   │   └── config.go
│   ├── models/            # 数据模型
│   │   ├── user.go
│   │   ├── alert.go
│   │   ├── monitoring.go
│   │   └── middleware.go
│   ├── services/          # 业务逻辑服务
│   │   ├── user_service.go
│   │   ├── monitoring_service.go
│   │   ├── alert_service.go
│   │   ├── ai_service.go
│   │   ├── middleware_service.go
│   │   ├── apm_service.go
│   │   ├── container_service.go
│   │   └── discovery_service.go
│   ├── database/          # 数据库相关
│   │   ├── migrations/    # 数据库迁移
│   │   └── connection.go
│   ├── utils/             # 工具函数
│   └── websocket/         # WebSocket 处理
├── pkg/                   # 可复用的包
│   ├── ai/               # AI 服务集成
│   ├── cache/            # 缓存抽象
│   ├── logger/           # 日志工具
│   └── validator/        # 数据验证
├── web/                   # 前端代码
│   ├── src/
│   │   ├── components/   # React 组件
│   │   │   ├── Dashboard/
│   │   │   ├── Monitoring/
│   │   │   ├── Alerts/
│   │   │   ├── AIAnalysis/
│   │   │   ├── Middleware/
│   │   │   ├── APM/
│   │   │   ├── Containers/
│   │   │   ├── Virtualization/
│   │   │   ├── KnowledgeBase/
│   │   │   ├── Settings/
│   │   │   ├── Profile/
│   │   │   ├── InstallGuide/
│   │   │   ├── APIKeys/
│   │   │   └── Layout/
│   │   ├── hooks/        # 自定义 Hooks
│   │   ├── services/     # API 服务
│   │   ├── store/        # 状态管理
│   │   ├── types/        # TypeScript 类型
│   │   └── utils/        # 工具函数
│   ├── public/           # 静态资源
│   └── package.json
├── agents/                # 监控代理
│   ├── windows/          # Windows 代理
│   ├── linux/            # Linux 代理
│   ├── apache/           # Apache 监控
│   ├── elasticsearch/    # ES 监控
│   ├── hyperv/           # Hyper-V 监控
│   ├── postgresql/       # PostgreSQL 监控
│   ├── vmware/           # VMware 监控
│   └── apm/              # APM 代理
├── configs/               # 配置文件
│   ├── config.yaml
│   └── config.dev.yaml
├── scripts/               # 构建和部署脚本
├── deploy/                # 部署配置
├── doc/                   # 项目文档
├── tests/                 # 测试文件
├── go.mod                 # Go 模块定义
├── go.sum                 # Go 依赖锁定
├── Dockerfile             # Docker 构建文件
├── docker-compose.yml     # Docker Compose 配置
├── quick-install.bat      # Windows 一键安装脚本
├── quick-install.sh       # Linux/macOS 一键安装脚本
└── README.md              # 项目说明
```

### 目录说明

| 目录 | 用途 | 规范 |
|------|------|------|
| `cmd/` | 应用程序入口点 | 每个可执行程序一个子目录 |
| `internal/` | 项目内部代码 | 不能被其他项目导入 |
| `pkg/` | 可复用的库代码 | 可以被其他项目导入 |
| `api/` | API 定义和文档 | OpenAPI/Swagger 规范 |
| `web/` | Web 静态资源 | 前端构建产物 |
| `configs/` | 配置文件模板 | 不包含敏感信息 |
| `deployments/` | 部署配置 | Docker, K8s 等 |
| `test/` | 测试文件 | 单元测试、集成测试 |

## 📝 代码规范

### Go 代码规范

#### 1. 命名规范

```go
// ✅ 正确的命名
type UserService struct {
    db     *sql.DB
    cache  cache.Cache
    logger logger.Logger
}

func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*User, error) {
    // 实现
}

// ❌ 错误的命名
type userservice struct {  // 应该使用 PascalCase
    DB     *sql.DB        // 私有字段应该使用 camelCase
    Cache  cache.Cache
}

func (s *userservice) getUserById(ctx context.Context, userId int64) (*User, error) {
    // 公开方法应该使用 PascalCase
}
```

#### 2. 错误处理

```go
// ✅ 正确的错误处理
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    if err := s.validateCreateUserRequest(req); err != nil {
        return nil, fmt.Errorf("validate request: %w", err)
    }
    
    user := &User{
        Name:  req.Name,
        Email: req.Email,
    }
    
    if err := s.db.CreateUser(ctx, user); err != nil {
        return nil, fmt.Errorf("create user in database: %w", err)
    }
    
    return user, nil
}

// ❌ 错误的错误处理
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    s.validateCreateUserRequest(req)  // 忽略错误
    
    user := &User{
        Name:  req.Name,
        Email: req.Email,
    }
    
    s.db.CreateUser(ctx, user)  // 忽略错误
    return user, nil
}
```

#### 3. 接口设计

```go
// ✅ 正确的接口设计
type UserRepository interface {
    GetByID(ctx context.Context, id int64) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, filter *UserFilter) ([]*User, error)
}

// 接口实现
type postgresUserRepository struct {
    db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
    return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    // 实现
}
```

#### 4. 结构体标签

```go
// ✅ 正确的结构体标签
type User struct {
    ID        int64     `json:"id" db:"id" validate:"required"`
    Name      string    `json:"name" db:"name" validate:"required,min=2,max=50"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// API 请求/响应结构
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 前端代码规范

#### 1. 组件命名

```typescript
// ✅ 正确的组件命名
// components/UserProfile/UserProfile.tsx
import React from 'react';
import { User } from '../../types/user';

interface UserProfileProps {
  user: User;
  onEdit: (user: User) => void;
}

export const UserProfile: React.FC<UserProfileProps> = ({ user, onEdit }) => {
  return (
    <div className="user-profile">
      <h2>{user.name}</h2>
      <p>{user.email}</p>
      <button onClick={() => onEdit(user)}>编辑</button>
    </div>
  );
};

// components/UserProfile/index.ts
export { UserProfile } from './UserProfile';
```

#### 2. Hooks 使用

```typescript
// ✅ 正确的 Hooks 使用
// hooks/useUser.ts
import { useState, useEffect } from 'react';
import { userService } from '../services/userService';
import { User } from '../types/user';

export const useUser = (userId: number) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        setLoading(true);
        const userData = await userService.getById(userId);
        setUser(userData);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : '获取用户失败');
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, [userId]);

  return { user, loading, error };
};
```

#### 3. 状态管理

```typescript
// ✅ 正确的状态管理 (Zustand)
// store/userStore.ts
import { create } from 'zustand';
import { User } from '../types/user';
import { userService } from '../services/userService';

interface UserState {
  users: User[];
  currentUser: User | null;
  loading: boolean;
  error: string | null;
  
  // Actions
  fetchUsers: () => Promise<void>;
  fetchUser: (id: number) => Promise<void>;
  createUser: (userData: CreateUserRequest) => Promise<void>;
  updateUser: (id: number, userData: UpdateUserRequest) => Promise<void>;
  deleteUser: (id: number) => Promise<void>;
}

export const useUserStore = create<UserState>((set, get) => ({
  users: [],
  currentUser: null,
  loading: false,
  error: null,

  fetchUsers: async () => {
    set({ loading: true, error: null });
    try {
      const users = await userService.getAll();
      set({ users, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },

  fetchUser: async (id: number) => {
    set({ loading: true, error: null });
    try {
      const user = await userService.getById(id);
      set({ currentUser: user, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },

  // 其他 actions...
}));
```

## 🔌 API开发

### RESTful API 设计

#### 1. 路由设计

```go
// internal/api/routes/routes.go
package routes

import (
    "github.com/gin-gonic/gin"
    "ai-monitor/internal/api/handlers"
    "ai-monitor/internal/api/middleware"
)

func SetupRoutes(r *gin.Engine, h *handlers.Handlers) {
    // 健康检查
    r.GET("/health", h.Health.Check)
    
    // API v1
    v1 := r.Group("/api/v1")
    {
        // 认证相关
        auth := v1.Group("/auth")
        {
            auth.POST("/login", h.Auth.Login)
            auth.POST("/logout", h.Auth.Logout)
            auth.POST("/refresh", h.Auth.RefreshToken)
        }
        
        // 需要认证的路由
        protected := v1.Group("/")
        protected.Use(middleware.AuthRequired())
        {
            // 用户管理
            users := protected.Group("/users")
            {
                users.GET("", h.User.List)           // GET /api/v1/users
                users.POST("", h.User.Create)        // POST /api/v1/users
                users.GET("/:id", h.User.GetByID)    // GET /api/v1/users/:id
                users.PUT("/:id", h.User.Update)     // PUT /api/v1/users/:id
                users.DELETE("/:id", h.User.Delete)  // DELETE /api/v1/users/:id
            }
            
            // 监控指标
            metrics := protected.Group("/metrics")
            {
                metrics.GET("", h.Metric.List)
                metrics.POST("", h.Metric.Create)
                metrics.GET("/:id", h.Metric.GetByID)
            }
            
            // 告警管理
            alerts := protected.Group("/alerts")
            {
                alerts.GET("", h.Alert.List)
                alerts.POST("", h.Alert.Create)
                alerts.PUT("/:id/status", h.Alert.UpdateStatus)
            }
        }
    }
    
    // WebSocket
    r.GET("/ws", h.WebSocket.HandleConnection)
}
```

#### 2. 处理器实现

```go
// internal/api/handlers/user.go
package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "ai-monitor/internal/services"
    "ai-monitor/pkg/logger"
)

type UserHandler struct {
    userService *services.UserService
    logger      logger.Logger
}

func NewUserHandler(userService *services.UserService, logger logger.Logger) *UserHandler {
    return &UserHandler{
        userService: userService,
        logger:      logger,
    }
}

// List 获取用户列表
func (h *UserHandler) List(c *gin.Context) {
    // 解析查询参数
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    
    filter := &services.UserFilter{
        Page:     page,
        PageSize: pageSize,
        Name:     c.Query("name"),
        Email:    c.Query("email"),
    }
    
    // 调用服务层
    users, total, err := h.userService.List(c.Request.Context(), filter)
    if err != nil {
        h.logger.Error("Failed to list users", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "获取用户列表失败",
        })
        return
    }
    
    // 返回响应
    c.JSON(http.StatusOK, gin.H{
        "data": gin.H{
            "users": users,
            "pagination": gin.H{
                "page":       page,
                "page_size":  pageSize,
                "total":      total,
                "total_pages": (total + pageSize - 1) / pageSize,
            },
        },
    })
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
    var req services.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "请求参数无效",
            "details": err.Error(),
        })
        return
    }
    
    // 验证请求数据
    if err := h.validateCreateUserRequest(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "数据验证失败",
            "details": err.Error(),
        })
        return
    }
    
    // 调用服务层
    user, err := h.userService.Create(c.Request.Context(), &req)
    if err != nil {
        h.logger.Error("Failed to create user", "error", err, "request", req)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "创建用户失败",
        })
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "data": user,
    })
}

// GetByID 根据ID获取用户
func (h *UserHandler) GetByID(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "用户ID无效",
        })
        return
    }
    
    user, err := h.userService.GetByID(c.Request.Context(), id)
    if err != nil {
        if err == services.ErrUserNotFound {
            c.JSON(http.StatusNotFound, gin.H{
                "error": "用户不存在",
            })
            return
        }
        
        h.logger.Error("Failed to get user", "error", err, "id", id)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "获取用户失败",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data": user,
    })
}

func (h *UserHandler) validateCreateUserRequest(req *services.CreateUserRequest) error {
    // 实现验证逻辑
    return nil
}
```

#### 3. 中间件开发

```go
// internal/api/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "ai-monitor/pkg/jwt"
    "ai-monitor/pkg/logger"
)

func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "缺少认证令牌",
            })
            c.Abort()
            return
        }
        
        // 检查 Bearer 前缀
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "认证令牌格式无效",
            })
            c.Abort()
            return
        }
        
        // 验证 JWT
        claims, err := jwt.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "认证令牌无效",
            })
            c.Abort()
            return
        }
        
        // 将用户信息存储到上下文
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        
        c.Next()
    }
}

// CORS 中间件
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        
        c.Next()
    }
}

// 请求日志中间件
func RequestLogger(logger logger.Logger) gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        logger.Info("HTTP Request",
            "method", param.Method,
            "path", param.Path,
            "status", param.StatusCode,
            "latency", param.Latency,
            "ip", param.ClientIP,
            "user_agent", param.Request.UserAgent(),
        )
        return ""
    })
}
```

### API 文档

#### Swagger 注释

```go
// @title AI Monitor API
// @version 1.0
// @description AI Monitor 系统 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

// @Summary 获取用户列表
// @Description 分页获取用户列表，支持按名称和邮箱过滤
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param name query string false "用户名过滤"
// @Param email query string false "邮箱过滤"
// @Success 200 {object} UserListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
    // 实现
}

// @Summary 创建用户
// @Description 创建新用户
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "用户信息"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
    // 实现
}
```

## 🎨 前端开发

### 组件开发

#### 1. 基础组件

```typescript
// components/Button/Button.tsx
import React from 'react';
import classNames from 'classnames';
import './Button.scss';

export interface ButtonProps {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'small' | 'medium' | 'large';
  disabled?: boolean;
  loading?: boolean;
  onClick?: () => void;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
}

export const Button: React.FC<ButtonProps> = ({
  children,
  variant = 'primary',
  size = 'medium',
  disabled = false,
  loading = false,
  onClick,
  type = 'button',
  className,
}) => {
  const buttonClass = classNames(
    'btn',
    `btn--${variant}`,
    `btn--${size}`,
    {
      'btn--disabled': disabled,
      'btn--loading': loading,
    },
    className
  );

  return (
    <button
      type={type}
      className={buttonClass}
      disabled={disabled || loading}
      onClick={onClick}
    >
      {loading && <span className="btn__spinner" />}
      <span className="btn__content">{children}</span>
    </button>
  );
};
```

```scss
// components/Button/Button.scss
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  
  &:focus {
    outline: none;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.5);
  }
  
  // 尺寸
  &--small {
    padding: 8px 16px;
    font-size: 14px;
    height: 32px;
  }
  
  &--medium {
    padding: 12px 24px;
    font-size: 16px;
    height: 40px;
  }
  
  &--large {
    padding: 16px 32px;
    font-size: 18px;
    height: 48px;
  }
  
  // 变体
  &--primary {
    background-color: #3b82f6;
    color: white;
    
    &:hover:not(.btn--disabled) {
      background-color: #2563eb;
    }
  }
  
  &--secondary {
    background-color: #6b7280;
    color: white;
    
    &:hover:not(.btn--disabled) {
      background-color: #4b5563;
    }
  }
  
  &--danger {
    background-color: #ef4444;
    color: white;
    
    &:hover:not(.btn--disabled) {
      background-color: #dc2626;
    }
  }
  
  // 状态
  &--disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  &--loading {
    cursor: wait;
    
    .btn__content {
      opacity: 0.7;
    }
  }
  
  &__spinner {
    width: 16px;
    height: 16px;
    border: 2px solid transparent;
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-right: 8px;
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
```

#### 2. 业务组件

```typescript
// components/UserTable/UserTable.tsx
import React from 'react';
import { User } from '../../types/user';
import { Button } from '../Button';
import { Table, TableColumn } from '../Table';
import './UserTable.scss';

interface UserTableProps {
  users: User[];
  loading?: boolean;
  onEdit: (user: User) => void;
  onDelete: (user: User) => void;
  onView: (user: User) => void;
}

export const UserTable: React.FC<UserTableProps> = ({
  users,
  loading = false,
  onEdit,
  onDelete,
  onView,
}) => {
  const columns: TableColumn<User>[] = [
    {
      key: 'id',
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      key: 'name',
      title: '姓名',
      dataIndex: 'name',
      sorter: true,
    },
    {
      key: 'email',
      title: '邮箱',
      dataIndex: 'email',
      sorter: true,
    },
    {
      key: 'status',
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => (
        <span className={`status status--${status.toLowerCase()}`}>
          {status === 'active' ? '活跃' : '禁用'}
        </span>
      ),
    },
    {
      key: 'created_at',
      title: '创建时间',
      dataIndex: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString(),
      sorter: true,
    },
    {
      key: 'actions',
      title: '操作',
      width: 200,
      render: (_, user) => (
        <div className="user-table__actions">
          <Button
            size="small"
            variant="secondary"
            onClick={() => onView(user)}
          >
            查看
          </Button>
          <Button
            size="small"
            onClick={() => onEdit(user)}
          >
            编辑
          </Button>
          <Button
            size="small"
            variant="danger"
            onClick={() => onDelete(user)}
          >
            删除
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="user-table">
      <Table
        columns={columns}
        dataSource={users}
        loading={loading}
        rowKey="id"
      />
    </div>
  );
};
```

### 状态管理

#### Zustand Store

```typescript
// store/authStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { authService } from '../services/authService';
import { User } from '../types/user';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  
  // Actions
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(n  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      loading: false,
      error: null,

      login: async (email: string, password: string) => {
        set({ loading: true, error: null });
        try {
          const response = await authService.login({ email, password });
          set({
            user: response.user,
            token: response.token,
            isAuthenticated: true,
            loading: false,
          });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : '登录失败',
            loading: false,
          });
        }
      },

      logout: () => {
        authService.logout();
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          error: null,
        });
      },

      refreshToken: async () => {
        const { token } = get();
        if (!token) return;

        try {
          const response = await authService.refreshToken(token);
          set({
            token: response.token,
            user: response.user,
          });
        } catch (error) {
          // Token 刷新失败，退出登录
          get().logout();
        }
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
```

### API 服务

```typescript
// services/apiClient.ts
import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { useAuthStore } from '../store/authStore';

class ApiClient {
  private client: AxiosInstance;

  constructor(baseURL: string) {
    this.client = axios.create({
      baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // 请求拦截器
    this.client.interceptors.request.use(
      (config) => {
        const token = useAuthStore.getState().token;
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // 响应拦截器
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;

          try {
            await useAuthStore.getState().refreshToken();
            const token = useAuthStore.getState().token;
            if (token) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
              return this.client(originalRequest);
            }
          } catch (refreshError) {
            useAuthStore.getState().logout();
            window.location.href = '/login';
          }
        }

        return Promise.reject(error);
      }
    );
  }

  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response: AxiosResponse<T> = await this.client.get(url, config);
    return response.data;
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response: AxiosResponse<T> = await this.client.post(url, data, config);
    return response.data;
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response: AxiosResponse<T> = await this.client.put(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response: AxiosResponse<T> = await this.client.delete(url, config);
    return response.data;
  }
}

export const apiClient = new ApiClient(process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080/api/v1');
```

```typescript
// services/userService.ts
import { apiClient } from './apiClient';
import { User, CreateUserRequest, UpdateUserRequest, UserListResponse } from '../types/user';

export class UserService {
  async getAll(params?: {
    page?: number;
    pageSize?: number;
    name?: string;
    email?: string;
  }): Promise<UserListResponse> {
    return apiClient.get('/users', { params });
  }

  async getById(id: number): Promise<User> {
    return apiClient.get(`/users/${id}`);
  }

  async create(data: CreateUserRequest): Promise<User> {
    return apiClient.post('/users', data);
  }

  async update(id: number, data: UpdateUserRequest): Promise<User> {
    return apiClient.put(`/users/${id}`, data);
  }

  async delete(id: number): Promise<void> {
    return apiClient.delete(`/users/${id}`);
  }
}

export const userService = new UserService();
```

## 🗄️ 数据库开发

### 数据库迁移

```go
// internal/database/migrations/001_create_users_table.go
package migrations

import (
    "database/sql"
    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigration(upCreateUsersTable, downCreateUsersTable)
}

func upCreateUsersTable(tx *sql.Tx) error {
    query := `
    CREATE TABLE users (
        id BIGSERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        status VARCHAR(20) DEFAULT 'active',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );
    
    CREATE INDEX idx_users_email ON users(email);
    CREATE INDEX idx_users_status ON users(status);
    CREATE INDEX idx_users_created_at ON users(created_at);
    `
    
    _, err := tx.Exec(query)
    return err
}

func downCreateUsersTable(tx *sql.Tx) error {
    _, err := tx.Exec("DROP TABLE IF EXISTS users;")
    return err
}
```

### 数据模型

```go
// internal/database/models/user.go
package models

import (
    "time"
    "database/sql/driver"
    "fmt"
)

type UserStatus string

const (
    UserStatusActive   UserStatus = "active"
    UserStatusInactive UserStatus = "inactive"
    UserStatusBanned   UserStatus = "banned"
)

func (us UserStatus) Value() (driver.Value, error) {
    return string(us), nil
}

func (us *UserStatus) Scan(value interface{}) error {
    if value == nil {
        *us = UserStatusActive
        return nil
    }
    
    switch s := value.(type) {
    case string:
        *us = UserStatus(s)
    case []byte:
        *us = UserStatus(s)
    default:
        return fmt.Errorf("cannot scan %T into UserStatus", value)
    }
    
    return nil
}

type User struct {
    ID           int64      `json:"id" db:"id"`
    Name         string     `json:"name" db:"name"`
    Email        string     `json:"email" db:"email"`
    PasswordHash string     `json:"-" db:"password_hash"`
    Status       UserStatus `json:"status" db:"status"`
    CreatedAt    time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateUserParams struct {
    Name     string `db:"name"`
    Email    string `db:"email"`
    Password string `db:"password_hash"`
}

type UpdateUserParams struct {
    ID     int64      `db:"id"`
    Name   *string    `db:"name"`
    Email  *string    `db:"email"`
    Status *UserStatus `db:"status"`
}

type UserFilter struct {
    Name     string
    Email    string
    Status   UserStatus
    Page     int
    PageSize int
}
```

### Repository 层

```go
// internal/database/repositories/user_repository.go
package repositories

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    
    "github.com/jmoiron/sqlx"
    "ai-monitor/internal/database/models"
)

type UserRepository interface {
    GetByID(ctx context.Context, id int64) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Create(ctx context.Context, params *models.CreateUserParams) (*models.User, error)
    Update(ctx context.Context, params *models.UpdateUserParams) (*models.User, error)
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, filter *models.UserFilter) ([]*models.User, int, error)
}

type userRepository struct {
    db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
    query := `
        SELECT id, name, email, password_hash, status, created_at, updated_at
        FROM users
        WHERE id = $1
    `
    
    var user models.User
    err := r.db.GetContext(ctx, &user, query, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    
    return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    query := `
        SELECT id, name, email, password_hash, status, created_at, updated_at
        FROM users
        WHERE email = $1
    `
    
    var user models.User
    err := r.db.GetContext(ctx, &user, query, email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("get user by email: %w", err)
    }
    
    return &user, nil
}

func (r *userRepository) Create(ctx context.Context, params *models.CreateUserParams) (*models.User, error) {
    query := `
        INSERT INTO users (name, email, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, name, email, password_hash, status, created_at, updated_at
    `
    
    var user models.User
    err := r.db.GetContext(ctx, &user, query, params.Name, params.Email, params.Password)
    if err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }
    
    return &user, nil
}

func (r *userRepository) Update(ctx context.Context, params *models.UpdateUserParams) (*models.User, error) {
    setParts := []string{}
    args := []interface{}{}
    argIndex := 1
    
    if params.Name != nil {
        setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
        args = append(args, *params.Name)
        argIndex++
    }
    
    if params.Email != nil {
        setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
        args = append(args, *params.Email)
        argIndex++
    }
    
    if params.Status != nil {
        setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
        args = append(args, *params.Status)
        argIndex++
    }
    
    if len(setParts) == 0 {
        return r.GetByID(ctx, params.ID)
    }
    
    setParts = append(setParts, fmt.Sprintf("updated_at = NOW()"))
    args = append(args, params.ID)
    
    query := fmt.Sprintf(`
        UPDATE users
        SET %s
        WHERE id = $%d
        RETURNING id, name, email, password_hash, status, created_at, updated_at
    `, strings.Join(setParts, ", "), argIndex)
    
    var user models.User
    err := r.db.GetContext(ctx, &user, query, args...)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("update user: %w", err)
    }
    
    return &user, nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
    query := `DELETE FROM users WHERE id = $1`
    
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("delete user: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return ErrUserNotFound
    }
    
    return nil
}

func (r *userRepository) List(ctx context.Context, filter *models.UserFilter) ([]*models.User, int, error) {
    whereParts := []string{}
    args := []interface{}
    argIndex := 1
    
    if filter.Name != "" {
        whereParts = append(whereParts, fmt.Sprintf("name ILIKE $%d", argIndex))
        args = append(args, "%"+filter.Name+"%")
        argIndex++
    }
    
    if filter.Email != "" {
        whereParts = append(whereParts, fmt.Sprintf("email ILIKE $%d", argIndex))
        args = append(args, "%"+filter.Email+"%")
        argIndex++
    }
    
    if filter.Status != "" {
        whereParts = append(whereParts, fmt.Sprintf("status = $%d", argIndex))
        args = append(args, filter.Status)
        argIndex++
    }
    
    whereClause := ""
    if len(whereParts) > 0 {
        whereClause = "WHERE " + strings.Join(whereParts, " AND ")
    }
    
    // 获取总数
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
    var total int
    err := r.db.GetContext(ctx, &total, countQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("count users: %w", err)
    }
    
    // 获取数据
    offset := (filter.Page - 1) * filter.PageSize
    dataQuery := fmt.Sprintf(`
        SELECT id, name, email, password_hash, status, created_at, updated_at
        FROM users
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, whereClause, argIndex, argIndex+1)
    
    args = append(args, filter.PageSize, offset)
    
    var users []*models.User
    err = r.db.SelectContext(ctx, &users, dataQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("select users: %w", err)
    }
    
    return users, total, nil
}

var (
    ErrUserNotFound = fmt.Errorf("user not found")
)
```

## 🧪 测试指南

### 单元测试

#### Go 单元测试

```go
// internal/services/user_service_test.go
package services

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "ai-monitor/internal/database/models"
    "ai-monitor/internal/database/repositories/mocks"
    "ai-monitor/pkg/logger"
)

func TestUserService_GetByID(t *testing.T) {
    // 准备测试数据
    userID := int64(1)
    expectedUser := &models.User{
        ID:        userID,
        Name:      "Test User",
        Email:     "test@example.com",
        Status:    models.UserStatusActive,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    // 创建 mock
    mockRepo := new(mocks.UserRepository)
    mockLogger := logger.NewNoop()
    
    // 设置 mock 期望
    mockRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)
    
    // 创建服务
    service := NewUserService(mockRepo, mockLogger)
    
    // 执行测试
    ctx := context.Background()
    user, err := service.GetByID(ctx, userID)
    
    // 验证结果
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    mockRepo.AssertExpectations(t)
}

func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name        string
        request     *CreateUserRequest
        setupMock   func(*mocks.UserRepository)
        expectedErr string
    }{
        {
            name: "成功创建用户",
            request: &CreateUserRequest{
                Name:     "New User",
                Email:    "new@example.com",
                Password: "password123",
            },
            setupMock: func(repo *mocks.UserRepository) {
                repo.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, repositories.ErrUserNotFound)
                repo.On("Create", mock.Anything, mock.AnythingOfType("*models.CreateUserParams")).Return(&models.User{
                    ID:    1,
                    Name:  "New User",
                    Email: "new@example.com",
                }, nil)
            },
        },
        {
            name: "邮箱已存在",
            request: &CreateUserRequest{
                Name:     "Duplicate User",
                Email:    "existing@example.com",
                Password: "password123",
            },
            setupMock: func(repo *mocks.UserRepository) {
                repo.On("GetByEmail", mock.Anything, "existing@example.com").Return(&models.User{}, nil)
            },
            expectedErr: "邮箱已存在",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(mocks.UserRepository)
            mockLogger := logger.NewNoop()
            
            tt.setupMock(mockRepo)
            
            service := NewUserService(mockRepo, mockLogger)
            
            ctx := context.Background()
            user, err := service.Create(ctx, tt.request)
            
            if tt.expectedErr != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErr)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

#### 前端单元测试

```typescript
// components/Button/Button.test.tsx
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { Button } from './Button';

describe('Button', () => {
  it('renders with correct text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const handleClick = jest.fn();
    render(<Button onClick={handleClick}>Click me</Button>);
    
    fireEvent.click(screen.getByText('Click me'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('is disabled when disabled prop is true', () => {
    render(<Button disabled>Click me</Button>);
    expect(screen.getByText('Click me')).toBeDisabled();
  });

  it('shows loading state', () => {
    render(<Button loading>Click me</Button>);
    expect(screen.getByText('Click me')).toBeDisabled();
    expect(document.querySelector('.btn__spinner')).toBeInTheDocument();
  });

  it('applies correct variant classes', () => {
    const { rerender } = render(<Button variant="primary">Primary</Button>);
    expect(screen.getByText('Primary')).toHaveClass('btn--primary');

    rerender(<Button variant="danger">Danger</Button>);
    expect(screen.getByText('Danger')).toHaveClass('btn--danger');
  });
});
```

```typescript
// hooks/useUser.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { useUser } from './useUser';
import { userService } from '../services/userService';

// Mock userService
jest.mock('../services/userService');
const mockUserService = userService as jest.Mocked<typeof userService>;

describe('useUser', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('fetches user successfully', async () => {
    const mockUser = {
      id: 1,
      name: 'Test User',
      email: 'test@example.com',
    };

    mockUserService.getById.mockResolvedValue(mockUser);

    const { result } = renderHook(() => useUser(1));

    expect(result.current.loading).toBe(true);
    expect(result.current.user).toBe(null);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.user).toEqual(mockUser);
    expect(result.current.error).toBe(null);
  });

  it('handles fetch error', async () => {
    const errorMessage = 'Failed to fetch user';
    mockUserService.getById.mockRejectedValue(new Error(errorMessage));

    const { result } = renderHook(() => useUser(1));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.user).toBe(null);
    expect(result.current.error).toBe(errorMessage);
  });
});
```

### 集成测试

```go
// tests/integration/user_api_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "ai-monitor/internal/api/routes"
    "ai-monitor/internal/config"
    "ai-monitor/internal/database"
)

type UserAPITestSuite struct {
    suite.Suite
    router *gin.Engine
    db     *database.DB
}

func (suite *UserAPITestSuite) SetupSuite() {
    // 设置测试配置
    cfg := &config.Config{
        Database: config.DatabaseConfig{
            Type: "postgres",
            Host: "localhost",
            Port: 5432,
            Name: "ai_monitor_test",
            Username: "test",
            Password: "test",
        },
    }
    
    // 初始化测试数据库
    db, err := database.New(cfg.Database)
    suite.Require().NoError(err)
    suite.db = db
    
    // 运行迁移
    err = db.Migrate()
    suite.Require().NoError(err)
    
    // 设置路由
    gin.SetMode(gin.TestMode)
    suite.router = gin.New()
    routes.SetupRoutes(suite.router, handlers)
}

func (suite *UserAPITestSuite) TearDownSuite() {
    // 清理测试数据
    suite.db.Close()
}

func (suite *UserAPITestSuite) SetupTest() {
    // 每个测试前清理数据
    suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}

func (suite *UserAPITestSuite) TestCreateUser() {
    // 准备请求数据
    userData := map[string]interface{}{
        "name":     "Test User",
        "email":    "test@example.com",
        "password": "password123",
    }
    
    jsonData, _ := json.Marshal(userData)
    
    // 发送请求
    req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+suite.getAuthToken())
    
    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)
    
    // 验证响应
    assert.Equal(suite.T(), http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    
    data := response["data"].(map[string]interface{})
    assert.Equal(suite.T(), "Test User", data["name"])
    assert.Equal(suite.T(), "test@example.com", data["email"])
    assert.NotEmpty(suite.T(), data["id"])
}

func (suite *UserAPITestSuite) TestGetUser() {
    // 先创建一个用户
    userID := suite.createTestUser()
    
    // 获取用户
    req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", userID), nil)
    req.Header.Set("Authorization", "Bearer "+suite.getAuthToken())
    
    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)
    
    // 验证响应
    assert.Equal(suite.T(), http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    
    data := response["data"].(map[string]interface{})
    assert.Equal(suite.T(), float64(userID), data["id"])
}

func (suite *UserAPITestSuite) createTestUser() int64 {
    // 辅助方法：创建测试用户
    // 实现省略...
    return 1
}

func (suite *UserAPITestSuite) getAuthToken() string {
    // 辅助方法：获取认证令牌
    // 实现省略...
    return "test-token"
}

func TestUserAPITestSuite(t *testing.T) {
    suite.Run(t, new(UserAPITestSuite))
}
```

### 端到端测试

```typescript
// e2e/user-management.spec.ts
import { test, expect } from '@playwright/test';

test.describe('用户管理', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@example.com');
    await page.fill('[data-testid="password"]', 'admin123');
    await page.click('[data-testid="login-button"]');
    
    // 等待跳转到首页
    await expect(page).toHaveURL('/dashboard');
  });

  test('创建新用户', async ({ page }) => {
    // 导航到用户管理页面
    await page.click('[data-testid="users-menu"]');
    await expect(page).toHaveURL('/users');

    // 点击创建用户按钮
    await page.click('[data-testid="create-user-button"]');

    // 填写用户信息
    await page.fill('[data-testid="user-name"]', 'Test User');
    await page.fill('[data-testid="user-email"]', 'testuser@example.com');
    await page.fill('[data-testid="user-password"]', 'password123');

    // 提交表单
    await page.click('[data-testid="submit-button"]');

    // 验证成功消息
    await expect(page.locator('[data-testid="success-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="success-message"]')).toContainText('用户创建成功');

    // 验证用户出现在列表中
    await expect(page.locator('[data-testid="user-table"]')).toContainText('Test User');
    await expect(page.locator('[data-testid="user-table"]')).toContainText('testuser@example.com');
  });

  test('编辑用户信息', async ({ page }) => {
    // 导航到用户管理页面
    await page.click('[data-testid="users-menu"]');
    
    // 点击第一个用户的编辑按钮
    await page.click('[data-testid="edit-user-1"]');

    // 修改用户名
    await page.fill('[data-testid="user-name"]', 'Updated User');

    // 提交表单
    await page.click('[data-testid="submit-button"]');

    // 验证更新成功
    await expect(page.locator('[data-testid="success-message"]')).toContainText('用户更新成功');
    await expect(page.locator('[data-testid="user-table"]')).toContainText('Updated User');
  });

  test('删除用户', async ({ page }) => {
    // 导航到用户管理页面
    await page.click('[data-testid="users-menu"]');
    
    // 获取用户数量
    const userRows = await page.locator('[data-testid="user-row"]').count();

    // 点击删除按钮
    await page.click('[data-testid="delete-user-1"]');

    // 确认删除
    await page.click('[data-testid="confirm-delete"]');

    // 验证用户被删除
    await expect(page.locator('[data-testid="success-message"]')).toContainText('用户删除成功');
    await expect(page.locator('[data-testid="user-row"]')).toHaveCount(userRows - 1);
  });
});
```

### 测试配置

```json
// package.json (前端测试配置)
{
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui"
  },
  "jest": {
    "testEnvironment": "jsdom",
    "setupFilesAfterEnv": ["<rootDir>/src/setupTests.ts"],
    "moduleNameMapping": {
      "^@/(.*)$": "<rootDir>/src/$1"
    },
    "collectCoverageFrom": [
      "src/**/*.{ts,tsx}",
      "!src/**/*.d.ts",
      "!src/index.tsx",
      "!src/reportWebVitals.ts"
    ],
    "coverageThreshold": {
      "global": {
        "branches": 80,
        "functions": 80,
        "lines": 80,
        "statements": 80
      }
    }
  }
}
```

```yaml
# .github/workflows/test.yml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: ai_monitor_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/ai_monitor_test?sslmode=disable
        REDIS_URL: redis://localhost:6379
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  frontend-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: 18
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
    
    - name: Install dependencies
      run: |
        cd frontend
        npm ci
    
    - name: Run tests
      run: |
        cd frontend
        npm run test:coverage
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./frontend/coverage/lcov.info

  e2e-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: 18
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
    
    - name: Install dependencies
      run: |
        cd frontend
        npm ci
        npx playwright install
    
    - name: Start application
      run: |
        # 启动后端服务
        docker-compose -f docker-compose.test.yml up -d
        # 等待服务启动
        sleep 30
    
    - name: Run E2E tests
      run: |
        cd frontend
        npm run test:e2e
    
    - name: Upload test results
      uses: actions/upload-artifact@v3
      if: failure()
      with:
        name: playwright-report
        path: frontend/playwright-report/
```

## 🐛 调试技巧

### Go 调试

#### 1. 使用 Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试应用程序
dlv debug cmd/server/main.go

# 在特定行设置断点
(dlv) break main.go:25

# 继续执行
(dlv) continue

# 查看变量
(dlv) print variableName

# 查看调用栈
(dlv) stack

# 单步执行
(dlv) next
(dlv) step
```

#### 2. 日志调试

```go
// 结构化日志
logger.Info("Processing user request",
    "user_id", userID,
    "action", "create",
    "request_id", requestID,
)

// 错误日志
logger.Error("Failed to create user",
    "error", err,
    "user_data", userData,
    "stack", string(debug.Stack()),
)

// 性能日志
start := time.Now()
defer func() {
    logger.Debug("Operation completed",
        "operation", "create_user",
        "duration", time.Since(start),
    )
}()
```

#### 3. 性能分析

```go
// 启用 pprof
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // 应用程序代码
}
```

```bash
# CPU 性能分析
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine 分析
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 前端调试

#### 1. React DevTools

```typescript
// 组件调试
const UserProfile = ({ user }) => {
  // 使用 React DevTools 查看 props 和 state
  console.log('UserProfile rendered', { user });
  
  return (
    <div>
      {/* 组件内容 */}
    </div>
  );
};
```

#### 2. 网络请求调试

```typescript
// API 调试拦截器
axios.interceptors.request.use(
  (config) => {
    console.log('🚀 Request:', config.method?.toUpperCase(), config.url, config.data);
    return config;
  },
  (error) => {
    console.error('❌ Request Error:', error);
    return Promise.reject(error);
  }
);

axios.interceptors.response.use(
  (response) => {
    console.log('✅ Response:', response.status, response.config.url, response.data);
    return response;
  },
  (error) => {
    console.error('❌ Response Error:', error.response?.status, error.config?.url, error.response?.data);
    return Promise.reject(error);
  }
);
```

#### 3. 状态调试

```typescript
// Zustand 调试
import { subscribeWithSelector } from 'zustand/middleware';
import { devtools } from 'zustand/middleware';

export const useUserStore = create<UserState>()(n  devtools(
    subscribeWithSelector((set, get) => ({
      // store 实现
    })),
    {
      name: 'user-store',
    }
  )
);

// 手动调试
const UserComponent = () => {
  const { users, loading, error } = useUserStore();
  
  // 调试状态变化
  useEffect(() => {
    console.log('User store state changed:', { users, loading, error });
  }, [users, loading, error]);
  
  return (
    // 组件内容
  );
};
```

### 数据库调试

```go
// SQL 查询日志
db, err := sqlx.Connect("postgres", dsn)
if err != nil {
    return nil, err
}

// 启用查询日志
if config.Debug {
    db = db.Unsafe() // 允许不安全的查询（仅开发环境）
    db.MapperFunc(strings.ToLower)
}

// 查询调试
func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
    query := `SELECT * FROM users WHERE id = $1`
    
    // 记录查询
    start := time.Now()
    defer func() {
        r.logger.Debug("SQL Query",
            "query", query,
            "args", []interface{}{id},
            "duration", time.Since(start),
        )
    }()
    
    var user models.User
    err := r.db.GetContext(ctx, &user, query, id)
    return &user, err
}
```

## ⚡ 性能优化

### 后端性能优化

#### 1. 数据库优化

```sql
-- 索引优化
CREATE INDEX CONCURRENTLY idx_users_email_status ON users(email, status);
CREATE INDEX CONCURRENTLY idx_metrics_timestamp ON metrics(timestamp DESC);
CREATE INDEX CONCURRENTLY idx_alerts_created_at ON alerts(created_at DESC) WHERE status = 'active';

-- 查询优化
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com' AND status = 'active';

-- 分区表（大数据量）
CREATE TABLE metrics_2024_01 PARTITION OF metrics
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

```go
// 连接池优化
func setupDatabase(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
    db, err := sqlx.Connect(cfg.Type, cfg.DSN)
    if err != nil {
        return nil, err
    }
    
    // 连接池配置
    db.SetMaxOpenConns(cfg.MaxOpenConns)     // 最大连接数
    db.SetMaxIdleConns(cfg.MaxIdleConns)     // 最大空闲连接数
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime) // 连接最大生命周期
    db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime) // 连接最大空闲时间
    
    return db, nil
}

// 批量操作
func (r *userRepository) CreateBatch(ctx context.Context, users []*models.User) error {
    if len(users) == 0 {
        return nil
    }
    
    query := `INSERT INTO users (name, email, password_hash) VALUES `
    values := []interface{}{}
    
    for i, user := range users {
        if i > 0 {
            query += ", "
        }
        query += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
        values = append(values, user.Name, user.Email, user.PasswordHash)
    }
    
    _, err := r.db.ExecContext(ctx, query, values...)
    return err
}
```

#### 2. 缓存策略

```go
// Redis 缓存
type CacheService struct {
    redis  *redis.Client
    logger logger.Logger
}

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
    data, err := c.redis.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return ErrCacheNotFound
        }
        return err
    }
    
    return json.Unmarshal([]byte(data), dest)
}

func (c *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return c.redis.Set(ctx, key, data, expiration).Err()
}

// 缓存装饰器
func (s *UserService) GetByIDWithCache(ctx context.Context, id int64) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // 尝试从缓存获取
    var user User
    err := s.cache.Get(ctx, cacheKey, &user)
    if err == nil {
        return &user, nil
    }
    
    // 缓存未命中，从数据库获取
    user, err = s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    s.cache.Set(ctx, cacheKey, user, 5*time.Minute)
    
    return user, nil
}
```

#### 3. 并发优化

```go
// Worker Pool 模式
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    resultChan chan Result
    wg         sync.WaitGroup
}

func NewWorkerPool(workers int, queueSize int) *WorkerPool {
    return &WorkerPool{
        workers:    workers,
        jobQueue:   make(chan Job, queueSize),
        resultChan: make(chan Result, queueSize),
    }
}

func (wp *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker(ctx)
    }
}

func (wp *WorkerPool) worker(ctx context.Context) {
    defer wp.wg.Done()
    
    for {
        select {
        case job := <-wp.jobQueue:
            result := job.Process()
            wp.resultChan <- result
        case <-ctx.Done():
            return
        }
    }
}

// 限流
type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(r, b),
    }
}

func (rl *RateLimiter) Allow() bool {
    return rl.limiter.Allow()
}

// 中间件
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "请求过于频繁",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 前端性能优化

#### 1. 组件优化

```typescript
// React.memo 优化
const UserCard = React.memo<UserCardProps>(({ user, onEdit }) => {
  return (
    <div className="user-card">
      <h3>{user.name}</h3>
      <p>{user.email}</p>
      <button onClick={() => onEdit(user)}>编辑</button>
    </div>
  );
}, (prevProps, nextProps) => {
  // 自定义比较函数
  return prevProps.user.id === nextProps.user.id &&
         prevProps.user.name === nextProps.user.name &&
         prevProps.user.email === nextProps.user.email;
});

// useMemo 优化计算
const UserList = ({ users, filter }) => {
  const filteredUsers = useMemo(() => {
    return users.filter(user => 
      user.name.toLowerCase().includes(filter.toLowerCase())
    );
  }, [users, filter]);

  return (
    <div>
      {filteredUsers.map(user => (
        <UserCard key={user.id} user={user} />
      ))}
    </div>
  );
};

// useCallback 优化函数
const UserManagement = () => {
  const [users, setUsers] = useState([]);
  
  const handleUserEdit = useCallback((user: User) => {
    // 编辑逻辑
  }, []);
  
  const handleUserDelete = useCallback((userId: number) => {
    setUsers(prev => prev.filter(u => u.id !== userId));
  }, []);
  
  return (
    <UserList 
      users={users}
      onEdit={handleUserEdit}
      onDelete={handleUserDelete}
    />
  );
};
```

#### 2. 懒加载和代码分割

```typescript
// 路由懒加载
import { lazy, Suspense } from 'react';

const UserManagement = lazy(() => import('./pages/UserManagement'));
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Settings = lazy(() => import('./pages/Settings'));

const App = () => {
  return (
    <Router>
      <Suspense fallback={<div>Loading...</div>}>
        <Routes>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/users" element={<UserManagement />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </Suspense>
    </Router>
  );
};

// 组件懒加载
const LazyChart = lazy(() => import('./components/Chart'));

const Dashboard = () => {
  const [showChart, setShowChart] = useState(false);
  
  return (
    <div>
      <h1>Dashboard</h1>
      <button onClick={() => setShowChart(true)}>显示图表</button>
      
      {showChart && (
        <Suspense fallback={<div>Loading chart...</div>}>
          <LazyChart />
        </Suspense>
      )}
    </div>
  );
};
```

#### 3. 虚拟滚动

```typescript
// 虚拟列表组件
import { FixedSizeList as List } from 'react-window';

interface VirtualUserListProps {
  users: User[];
  height: number;
  itemHeight: number;
}

const VirtualUserList: React.FC<VirtualUserListProps> = ({ users, height, itemHeight }) => {
  const Row = ({ index, style }) => (
    <div style={style}>
      <UserCard user={users[index]} />
    </div>
  );

  return (
    <List
      height={height}
      itemCount={users.length}
      itemSize={itemHeight}
      width="100%"
    >
      {Row}
    </List>
  );
};

// 无限滚动
const InfiniteUserList = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [page, setPage] = useState(1);

  const loadMore = useCallback(async () => {
    if (loading || !hasMore) return;
    
    setLoading(true);
    try {
      const newUsers = await userService.getAll({ page, pageSize: 20 });
      setUsers(prev => [...prev, ...newUsers.data]);
      setPage(prev => prev + 1);
      setHasMore(newUsers.data.length === 20);
    } catch (error) {
      console.error('Failed to load users:', error);
    } finally {
      setLoading(false);
    }
  }, [page, loading, hasMore]);

  useEffect(() => {
    loadMore();
  }, []);

  return (
    <InfiniteScroll
      dataLength={users.length}
      next={loadMore}
      hasMore={hasMore}
      loader={<div>Loading...</div>}
    >
      {users.map(user => (
        <UserCard key={user.id} user={user} />
      ))}
    </InfiniteScroll>
  );
};
```

## 🚀 部署流程

### 开发环境部署

```bash
#!/bin/bash
# scripts/dev-deploy.sh

set -e

echo "🚀 开始开发环境部署..."

# 检查依赖
command -v go >/dev/null 2>&1 || { echo "Go 未安装"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Node.js 未安装"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Docker 未安装"; exit 1; }

# 启动数据库服务
echo "📦 启动数据库服务..."
docker-compose -f docker-compose.dev.yml up -d postgres redis

# 等待数据库启动
echo "⏳ 等待数据库启动..."
sleep 10

# 运行数据库迁移
echo "🗄️ 运行数据库迁移..."
go run cmd/migrate/main.go up

# 安装前端依赖
echo "📦 安装前端依赖..."
cd frontend
npm install
cd ..

# 构建前端
echo "🏗️ 构建前端..."
cd frontend
npm run build
cd ..

# 启动后端服务
echo "🚀 启动后端服务..."
go run cmd/server/main.go &
BACKEND_PID=$!

# 启动前端开发服务器
echo "🎨 启动前端开发服务器..."
cd frontend
npm start &
FRONTEND_PID=$!
cd ..

echo "✅ 开发环境部署完成！"
echo "🌐 前端地址: http://localhost:3000"
echo "🔧 后端地址: http://localhost:8080"
echo "📊 API 文档: http://localhost:8080/swagger/index.html"

# 清理函数
cleanup() {
    echo "🧹 清理进程..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
    docker-compose -f docker-compose.dev.yml down
}

# 捕获退出信号
trap cleanup EXIT INT TERM

# 等待用户中断
echo "按 Ctrl+C 停止服务"
wait
```

### 生产环境部署

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.prod
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
      - DATABASE_URL=postgres://ai_monitor:${DB_PASSWORD}@postgres:5432/ai_monitor?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=ai_monitor
      - POSTGRES_USER=ai_monitor
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ai_monitor"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
      - ./frontend/build:/usr/share/nginx/html
    depends_on:
      - app
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

```dockerfile
# Dockerfile.prod
# 多阶段构建
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# Go 构建阶段
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

# 复制二进制文件
COPY --from=backend-builder /app/main .
COPY --from=frontend-builder /app/frontend/build ./web

# 复制配置文件
COPY config/ ./config/

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./main"]
```

### CI/CD 流水线

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [ main ]
  release:
    types: [ published ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Run tests
      run: |
        make test

  build-and-push:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    
    steps:
    - name: Checkout
      uses: actions/checkout@v3
    
    - name: Log in to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v4
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
    
    - name: Build and push
      uses: docker/build-push-action@v4
      with:
        context: .
        file: ./Dockerfile.prod
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - name: Deploy to production
      uses: appleboy/ssh-action@v0.1.5
      with:
        host: ${{ secrets.PROD_HOST }}
        username: ${{ secrets.PROD_USER }}
        key: ${{ secrets.PROD_SSH_KEY }}
        script: |
          cd /opt/ai-monitor
          docker-compose pull
          docker-compose up -d
          docker system prune -f
```

## 🤝 贡献指南

### 开发流程

1. **Fork 项目**
   ```bash
   git clone https://github.com/your-username/ai-monitor.git
   cd ai-monitor
   git remote add upstream https://github.com/original-repo/ai-monitor.git
   ```

2. **创建功能分支**
   ```bash
   git checkout -b feature/user-management
   ```

3. **开发和测试**
   ```bash
   # 运行测试
   make test
   
   # 代码格式化
   make fmt
   
   # 代码检查
   make lint
   ```

4. **提交代码**
   ```bash
   git add .
   git commit -m "feat: add user management functionality"
   ```

5. **推送和创建 PR**
   ```bash
   git push origin feature/user-management
   ```

### 代码提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**类型说明**：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

**示例**：
```
feat(user): add user profile management

- Add user profile editing functionality
- Implement avatar upload
- Add validation for user data

Closes #123
```

### Pull Request 模板

```markdown
## 变更描述

简要描述此 PR 的变更内容。

## 变更类型

- [ ] Bug 修复
- [ ] 新功能
- [ ] 重构
- [ ] 文档更新
- [ ] 性能优化
- [ ] 其他

## 测试

- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 手动测试完成
- [ ] 代码覆盖率满足要求

## 检查清单

- [ ] 代码遵循项目规范
- [ ] 已添加必要的测试
- [ ] 文档已更新
- [ ] 无破坏性变更
- [ ] 已测试向后兼容性

## 相关 Issue

Closes #(issue number)

## 截图（如适用）

## 额外说明

任何需要特别说明的内容。
```

### 代码审查指南

**审查者检查清单**：

1. **功能性**
   - [ ] 代码实现了预期功能
   - [ ] 边界条件处理正确
   - [ ] 错误处理完善

2. **代码质量**
   - [ ] 代码清晰易读
   - [ ] 命名规范
   - [ ] 注释充分
   - [ ] 无重复代码

3. **性能**
   - [ ] 无明显性能问题
   - [ ] 数据库查询优化
   - [ ] 内存使用合理

4. **安全性**
   - [ ] 输入验证
   - [ ] 权限检查
   - [ ] 无安全漏洞

5. **测试**
   - [ ] 测试覆盖充分
   - [ ] 测试用例合理
   - [ ] 测试通过

---

## 📚 相关资源

- [Go 官方文档](https://golang.org/doc/)
- [React 官方文档](https://reactjs.org/docs/)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)
- [Redis 文档](https://redis.io/documentation)
- [Docker 文档](https://docs.docker.com/)
- [Kubernetes 文档](https://kubernetes.io/docs/)

## 🆘 获取帮助

- 📧 邮件: dev@ai-monitor.com
- 💬 Slack: #ai-monitor-dev
- 🐛 Bug 报告: [GitHub Issues](https://github.com/your-org/ai-monitor/issues)
- 📖 Wiki: [项目 Wiki](https://github.com/your-org/ai-monitor/wiki)

---

*本开发指南会持续更新，请定期查看最新版本。*