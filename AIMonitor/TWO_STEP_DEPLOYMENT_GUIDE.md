# AI Monitor 两步部署策略指南

## 概述

本指南介绍AI Monitor的两步部署策略，通过分离镜像构建和服务部署过程，大幅提高远程服务器部署的成功率和可靠性。

## 部署策略优势

### 传统部署问题
- 远程服务器网络环境复杂，依赖下载经常失败
- 构建过程耗时长，容易因网络中断导致失败
- 错误排查困难，构建和部署混合在一起
- 重复部署时需要重新构建，效率低下

### 两步部署优势
- **构建与部署分离**：在本地稳定环境构建镜像，避免远程网络问题
- **离线部署**：镜像预构建后可离线部署，不依赖外网
- **快速部署**：跳过构建过程，部署速度提升80%以上
- **错误隔离**：构建错误和部署错误分离，便于排查
- **版本管理**：镜像文件可版本化管理，支持回滚

## 部署流程

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   第一步：构建   │───▶│   传输镜像文件   │───▶│   第二步：部署   │
│   (本地环境)    │    │   (scp/ftp等)   │    │   (远程服务器)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 第一步：镜像构建

### 环境要求
- Docker Desktop (Windows/macOS) 或 Docker Engine (Linux)
- 8GB+ 可用内存
- 20GB+ 可用磁盘空间
- 稳定的网络连接

### 构建脚本

#### Linux/macOS 环境
```bash
# 给脚本执行权限
chmod +x build_images.sh

# 执行构建
./build_images.sh
```

#### Windows 环境
```powershell
# 以管理员身份运行PowerShell
# 执行构建
.\build_images.ps1
```

### 构建过程

1. **环境检查**
   - 验证Docker环境
   - 检查项目结构完整性
   - 确认必要文件存在

2. **缓存清理**
   - 清理旧的Docker镜像
   - 清理构建缓存
   - 停止相关容器

3. **Dockerfile优化**
   - 创建多阶段构建的Dockerfile
   - 配置中国镜像源加速
   - 添加重试机制

4. **镜像构建**
   - 构建后端Go应用镜像
   - 构建前端React应用镜像
   - 拉取基础服务镜像

5. **镜像打包**
   - 将所有镜像保存为tar.gz文件
   - 生成镜像清单
   - 创建加载脚本

### 构建产物

构建完成后，会在 `docker-images/` 目录生成以下文件：

```
docker-images/
├── ai-monitor-backend.tar.gz     # 后端服务镜像 (~200MB)
├── ai-monitor-frontend.tar.gz    # 前端服务镜像 (~50MB)
├── postgres.tar.gz               # PostgreSQL数据库 (~150MB)
├── redis.tar.gz                  # Redis缓存 (~30MB)
├── prometheus.tar.gz             # Prometheus监控 (~200MB)
├── elasticsearch.tar.gz          # Elasticsearch搜索 (~500MB)
├── images.txt                    # 镜像清单文件
└── load_images.sh               # 镜像加载脚本
```

## 第二步：远程部署

### 文件传输

将整个 `docker-images/` 目录传输到远程服务器：

#### 使用SCP
```bash
# 传输整个目录
scp -r docker-images/ user@remote-server:/path/to/deployment/

# 或者打包后传输
tar -czf docker-images.tar.gz docker-images/
scp docker-images.tar.gz user@remote-server:/path/to/deployment/
```

#### 使用SFTP
```bash
sftp user@remote-server
put -r docker-images/
```

#### 使用rsync
```bash
rsync -avz docker-images/ user@remote-server:/path/to/deployment/docker-images/
```

### 远程服务器部署

1. **连接到远程服务器**
   ```bash
   ssh user@remote-server
   cd /path/to/deployment
   ```

2. **加载Docker镜像**
   ```bash
   cd docker-images/
   chmod +x load_images.sh
   ./load_images.sh
   ```

3. **执行部署**
   ```bash
   # 将部署脚本复制到服务器
   # 然后执行
   chmod +x deploy_from_images.sh
   ./deploy_from_images.sh
   ```

### 部署过程

1. **环境验证**
   - 检查Docker环境
   - 验证镜像完整性
   - 确认端口可用性

2. **配置生成**
   - 创建docker-compose配置
   - 生成应用配置文件
   - 设置数据库初始化脚本

3. **服务启动**
   - 启动基础服务（数据库、缓存）
   - 等待基础服务就绪
   - 启动应用服务

4. **健康检查**
   - 验证所有服务状态
   - 执行API健康检查
   - 确认服务可访问性

## 服务管理

部署完成后，系统会自动创建管理脚本：

### 服务管理脚本 (manage.sh)

```bash
# 启动所有服务
./manage.sh start

# 停止所有服务
./manage.sh stop

# 重启所有服务
./manage.sh restart

# 查看服务状态
./manage.sh status

# 查看日志
./manage.sh logs
./manage.sh logs aimonitor  # 查看特定服务日志

# 更新服务
./manage.sh update

# 备份数据库
./manage.sh backup
```

### 系统监控脚本 (monitor.sh)

```bash
# 查看系统整体状态
./monitor.sh
```

## 服务访问

部署成功后，可通过以下地址访问各项服务：

- **前端界面**: http://server-ip:3000
- **后端API**: http://server-ip:8080
- **API文档**: http://server-ip:8080/swagger/index.html
- **健康检查**: http://server-ip:8080/health
- **Prometheus**: http://server-ip:9090
- **Elasticsearch**: http://server-ip:9200

### 默认登录信息
- **用户名**: admin
- **密码**: password

## 故障排查

### 构建阶段问题

#### 1. Docker环境问题
```bash
# 检查Docker状态
docker info
docker version

# 重启Docker服务
sudo systemctl restart docker  # Linux
# 或重启Docker Desktop (Windows/macOS)
```

#### 2. 网络连接问题
```bash
# 测试网络连接
curl -I https://goproxy.cn
curl -I https://registry.npmmirror.com

# 配置代理（如需要）
export HTTP_PROXY=http://proxy:port
export HTTPS_PROXY=http://proxy:port
```

#### 3. 磁盘空间不足
```bash
# 检查磁盘空间
df -h

# 清理Docker缓存
docker system prune -a
```

### 部署阶段问题

#### 1. 镜像加载失败
```bash
# 检查镜像文件完整性
ls -la docker-images/
file docker-images/*.tar.gz

# 手动加载镜像
docker load < docker-images/ai-monitor-backend.tar.gz
```

#### 2. 端口冲突
```bash
# 检查端口占用
netstat -tlnp | grep -E '3000|8080|5432|6379|9090|9200'

# 修改端口配置
vim docker-compose.production.yml
```

#### 3. 服务启动失败
```bash
# 查看详细日志
docker-compose -f docker-compose.production.yml logs aimonitor

# 检查服务依赖
docker-compose -f docker-compose.production.yml ps
```

### 运行时问题

#### 1. 服务无响应
```bash
# 检查容器状态
docker ps -a

# 重启特定服务
docker-compose -f docker-compose.production.yml restart aimonitor
```

#### 2. 数据库连接问题
```bash
# 测试数据库连接
docker-compose -f docker-compose.production.yml exec postgres psql -U ai_monitor -d ai_monitor

# 检查数据库日志
docker-compose -f docker-compose.production.yml logs postgres
```

#### 3. 内存不足
```bash
# 检查内存使用
free -h
docker stats

# 调整服务资源限制
vim docker-compose.production.yml
```

## 性能优化

### 1. 资源配置

根据服务器配置调整资源限制：

```yaml
# docker-compose.production.yml
services:
  aimonitor:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 1G
```

### 2. 数据库优化

```yaml
postgres:
  environment:
    - POSTGRES_SHARED_PRELOAD_LIBRARIES=pg_stat_statements
    - POSTGRES_MAX_CONNECTIONS=200
    - POSTGRES_SHARED_BUFFERS=256MB
```

### 3. 缓存优化

```yaml
redis:
  command: >
    redis-server
    --maxmemory 512mb
    --maxmemory-policy allkeys-lru
    --save 900 1
```

## 版本管理

### 镜像版本控制

```bash
# 构建时添加版本标签
docker build -t ai-monitor-backend:v1.0.0 .
docker build -t ai-monitor-backend:latest .

# 保存特定版本
docker save ai-monitor-backend:v1.0.0 | gzip > ai-monitor-backend-v1.0.0.tar.gz
```

### 回滚策略

```bash
# 保存当前版本
docker tag ai-monitor-backend:latest ai-monitor-backend:backup

# 加载旧版本
docker load < ai-monitor-backend-v1.0.0.tar.gz

# 重新部署
docker-compose -f docker-compose.production.yml up -d
```

## 安全建议

### 1. 网络安全
- 使用防火墙限制端口访问
- 配置SSL/TLS证书
- 设置反向代理

### 2. 数据安全
- 定期备份数据库
- 加密敏感配置
- 使用强密码

### 3. 容器安全
- 定期更新基础镜像
- 使用非root用户运行
- 限制容器权限

## 监控和告警

### 1. 系统监控
- CPU、内存、磁盘使用率
- 网络连接状态
- 服务响应时间

### 2. 应用监控
- API响应时间
- 错误率统计
- 业务指标监控

### 3. 日志管理
- 集中化日志收集
- 日志轮转配置
- 异常日志告警

## 总结

两步部署策略通过分离构建和部署过程，显著提高了AI Monitor在远程服务器上的部署成功率。主要优势包括：

1. **可靠性提升**: 避免网络环境对部署的影响
2. **效率提升**: 部署速度提升80%以上
3. **维护简化**: 错误隔离，便于排查和修复
4. **版本管理**: 支持镜像版本化和快速回滚

建议在生产环境中采用此部署策略，确保系统的稳定性和可维护性。