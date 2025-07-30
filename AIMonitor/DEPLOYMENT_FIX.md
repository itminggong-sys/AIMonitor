# AI Monitor 部署问题修复指南

## 问题描述

在远程CentOS部署完成后，发现以下两个问题：

1. **API路径重复问题**: 浏览器请求显示 `http://192.168.36.33:3000/api/api/v1/auth/login`，出现了重复的 `/api` 路径
2. **构建产物包含node_modules**: 使用构建脚本后，web目录下存在node_modules文件夹，导致发布时报错

## 问题原因分析

### 1. API路径重复问题

**根本原因**: 前端环境变量配置与nginx代理配置冲突

- 前端 `.env` 文件中设置了 `VITE_API_BASE_URL=/api`
- 前端代码中使用 `${API_BASE_URL}/api/v1/...` 构建请求URL
- nginx配置中已经将 `/api` 路径代理到后端服务
- 最终导致URL变成 `/api/api/v1/...`

### 2. node_modules包含问题

**根本原因**: Docker构建时缺少 `.dockerignore` 文件

- 前端Dockerfile中使用 `COPY . .` 复制所有文件
- 没有 `.dockerignore` 文件排除 `node_modules` 目录
- 导致本地的 `node_modules` 被复制到Docker镜像中

## 修复方案

### 1. 修复API路径重复问题

**修改前端环境变量配置**:

```bash
# web/.env
# 修改前
VITE_API_BASE_URL=/api

# 修改后  
VITE_API_BASE_URL=
```

**说明**: 
- 将 `VITE_API_BASE_URL` 设置为空字符串
- 前端代码中的 `${API_BASE_URL}/api/v1/...` 会变成 `/api/v1/...`
- nginx代理配置中的 `/api` 路径会正确转发到后端

### 2. 修复node_modules包含问题

**创建 `.dockerignore` 文件**:

```bash
# web/.dockerignore
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
dist/
build/
.env.local
.env.development.local
.env.test.local
.env.production.local
.vscode/
.idea/
*.swp
*.swo
*~
.DS_Store
Thumbs.db
.git/
.gitignore
Dockerfile*
docker-compose*
.dockerignore
logs/
*.log
coverage/
.eslintcache
.npm
*.tgz
.yarn-integrity
.cache
.parcel-cache
```

**说明**:
- 排除 `node_modules/` 目录，避免本地依赖被复制
- 排除各种缓存、日志、IDE配置文件
- 确保Docker镜像只包含必要的源代码文件

### 3. 完善nginx配置

**添加WebSocket代理支持**:

```nginx
location /ws {
    proxy_pass http://aimonitor:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

### 4. 修复docker-compose配置

**更新开发环境配置**:

```yaml
# docker-compose.yml
frontend:
  environment:
    - NODE_ENV=development
    - VITE_API_BASE_URL=  # 改为空字符串
```

**说明**:
- 在容器环境中，前端通过nginx代理访问后端
- 不需要指定具体的后端地址
- 依赖nginx的/api路径代理配置

## 部署步骤

### 1. 重新构建镜像

```bash
# 在项目根目录执行
./build_images.sh
```

### 2. 传输镜像到远程服务器

```bash
# 将 docker-images 目录传输到远程服务器
scp -r docker-images/ user@192.168.36.33:/path/to/deployment/
```

### 3. 在远程服务器加载镜像

```bash
# 在远程服务器上执行
cd /path/to/deployment/docker-images/
./load_images.sh
```

### 4. 重新部署服务

```bash
# 停止现有服务
docker-compose down

# 启动新服务
docker-compose up -d
```

## 验证修复效果

### 1. 检查API请求路径

打开浏览器开发者工具，查看网络请求：
- ✅ 正确: `http://192.168.36.33:3000/api/v1/auth/login`
- ❌ 错误: `http://192.168.36.33:3000/api/api/v1/auth/login`

### 2. 检查镜像大小

```bash
# 查看前端镜像大小，应该明显减小
docker images | grep ai-monitor-frontend
```

### 3. 功能测试

- 登录功能正常
- WebSocket连接正常
- 所有API接口调用正常

## 预防措施

### 1. 开发环境配置

建议在开发环境中使用完整的API地址：

```bash
# web/.env.development
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080
```

### 2. 生产环境配置

生产环境使用相对路径，依赖nginx代理：

```bash
# web/.env.production
VITE_API_BASE_URL=
VITE_WS_BASE_URL=/ws
```

### 3. 构建检查

每次构建前检查：
- `.dockerignore` 文件存在且配置正确
- 环境变量配置符合部署环境要求
- nginx配置包含所有必要的代理规则

## 总结

通过以上修复：

1. **解决了API路径重复问题**: 修改前端环境变量配置，避免路径冲突
2. **解决了构建产物问题**: 添加 `.dockerignore` 文件，优化Docker镜像
3. **完善了代理配置**: 确保API和WebSocket代理正常工作
4. **提供了部署指导**: 明确的重新部署步骤和验证方法

这些修复确保了AI Monitor系统在生产环境中的稳定运行。