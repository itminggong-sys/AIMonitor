# AI智能监控系统快速开始指南

## 文档概述

本文档为新用户提供最简单的系统部署和使用指南，帮助您在最短时间内启动AI智能监控系统。

## 版本信息

- **系统版本**: v3.8.5
- **部署方式**: Docker Compose
- **默认端口**: 前端3000，后端8080
- **数据库**: PostgreSQL 15.3

## 📋 目录

1. [环境准备](#环境准备)
2. [一键部署](#一键部署)
3. [首次配置](#首次配置)
4. [基础使用](#基础使用)
5. [常见问题](#常见问题)

## 🚀 环境准备

### 最小系统要求

- **CPU**: 2核心
- **内存**: 4GB RAM
- **存储**: 50GB 可用空间
- **网络**: 100Mbps
- **操作系统**: 
  - Linux: Ubuntu 20.04+, CentOS 8+
  - Windows: Windows 10+, Windows Server 2019+
  - macOS: 11.0+

### 必需软件

#### Docker部署（推荐）
- Docker 24.0+
- Docker Compose 2.20+

#### 源码部署
- Go 1.21+
- Node.js 18+
- PostgreSQL 13+ 或 MySQL 8.0+
- Redis 6.0+

## 🎯 一键部署（超简单）

### 🚀 小白用户专用（零技术门槛）

我们为非技术用户提供了超级简单的一键部署方式，只需要运行一个脚本，系统会自动安装所有依赖并完成部署。

#### 🎯 Windows用户（推荐）

```batch
# 1. 下载项目到本地（如：C:\AIMonitor）
# 2. 以管理员身份打开PowerShell或命令提示符
# 3. 进入项目目录
cd C:\AIMonitor

# 4. 运行Docker Compose部署
docker-compose up -d

# 5. 检查服务状态
docker-compose ps
```

**脚本会自动完成：**
- ✅ 检查系统环境
- ✅ 安装Docker Desktop
- ✅ 安装PostgreSQL数据库
- ✅ 安装Redis缓存
- ✅ 构建前后端应用
- ✅ 启动所有服务
- ✅ 创建默认管理员账户

#### 🐧 Linux/macOS用户（推荐）

```bash
# 1. 克隆项目到本地
git clone https://github.com/your-org/AIMonitor.git
cd AIMonitor

# 2. 启动服务
docker-compose up -d

# 3. 检查服务状态
docker-compose ps
```

**脚本会自动完成：**
- ✅ 检查系统环境和依赖
- ✅ 安装Docker和Docker Compose
- ✅ 安装数据库和缓存服务
- ✅ 构建应用Docker镜像
- ✅ 启动所有服务容器
- ✅ 配置系统服务自启动

#### 🐳 技术用户快速部署

如果您已经安装了Docker，可以直接使用Docker Compose：

```bash
# 生产环境部署
docker-compose -f docker-compose.prod.yml up -d

# 开发环境部署
docker-compose up -d
```

## ⚙️ 首次配置

### 1. 验证部署状态

```bash
# 检查服务状态
docker-compose ps

# 检查服务健康状态
curl http://localhost:8080/api/v1/health
curl http://localhost:3000
```

### 2. 访问系统

打开浏览器访问以下地址：

| 服务 | 地址 | 默认账号 |
|------|------|----------|
| **AI Monitor主界面** | http://localhost:3000 | admin / admin123 |
| **后端API服务** | http://localhost:8080 | - |
| **WebSocket服务** | ws://localhost:8080/ws | - |
| **API文档** | http://localhost:8080/swagger/index.html | 无需登录 |

### 3. **首次登录配置**

1. **登录系统**
   - 访问 http://localhost:3000
   - 使用默认账号：admin / admin123
   - 首次登录后请立即修改密码

2. **修改默认密码**
   - 点击右上角用户头像
   - 选择「设置」页面
   - 在用户管理中修改密码并保存

3. **配置AI服务**（可选）
   - 进入「设置」页面
   - 在AI配置部分添加API密钥
   - 支持OpenAI GPT-4和Claude 3
   - 测试连接并保存

## 📊 基础使用

### 1. 添加监控目标

#### 自动发现（推荐）

1. 进入「发现管理」页面
2. 点击「新建发现任务」
3. 配置扫描参数：
   ```
   目标类型: 服务器
   IP范围: 192.168.1.1-192.168.1.100
   扫描端口: 22,80,443,3306,6379
   认证方式: SSH密钥/用户名密码
   ```
4. 点击「开始发现」
5. 等待发现完成，选择要监控的设备
6. 点击「添加到监控」

#### 手动添加

1. 进入「设备管理」页面
2. 点击「添加设备」
3. 填写设备信息：
   ```
   设备名称: Web服务器01
   IP地址: 192.168.1.10
   设备类型: 服务器
   操作系统: Linux
   认证信息: SSH用户名/密码
   ```
4. 点击「测试连接」
5. 连接成功后点击「保存」

### 2. 安装监控Agent

#### 自动安装（推荐）

系统会在添加设备时自动部署Agent，无需手动操作。

#### 手动安装

1. 进入「Agent管理」页面
2. 点击「下载Agent」
3. 选择对应的操作系统版本
4. 按照安装指南部署Agent

### 3. 配置告警规则

1. 进入「告警管理」→「告警规则」
2. 点击「新建规则」
3. 配置告警条件：
   ```
   规则名称: CPU使用率过高
   监控指标: cpu_usage_percent
   告警条件: > 80
   持续时间: 5分钟
   告警级别: 警告
   ```
4. 配置通知方式（邮件、钉钉、企业微信等）
5. 点击「保存并启用」

### 4. 查看监控数据

#### 实时监控

1. 进入「监控概览」查看整体状态
2. 点击具体设备查看详细指标
3. 使用时间选择器查看历史数据

#### Grafana仪表板

1. 访问 http://localhost:3000
2. 使用 admin/admin123 登录
3. 查看预置的监控仪表板
4. 根据需要创建自定义仪表板

## 🔧 常见问题

### 部署相关

**Q: Docker部署失败，提示端口被占用**

A: 检查端口占用情况并释放端口：
```bash
# 检查端口占用
netstat -tlnp | grep :3001
netstat -tlnp | grep :8080

# 停止占用端口的服务
sudo systemctl stop nginx  # 如果80端口被占用
sudo systemctl stop apache2

# 或修改docker-compose.yml中的端口映射
```

**Q: 服务启动后无法访问**

A: 检查防火墙和服务状态：
```bash
# 检查服务状态
docker compose ps

# 检查防火墙
sudo ufw status
sudo firewall-cmd --list-ports

# 开放必要端口
sudo ufw allow 3001
sudo ufw allow 8080
```

### 监控相关

**Q: Agent连接失败**

A: 检查网络连接和认证信息：
```bash
# 测试SSH连接
ssh username@target_ip

# 检查防火墙
telnet target_ip 22

# 验证认证信息
ssh -i /path/to/key username@target_ip
```

**Q: 监控数据不更新**

A: 检查Agent状态和配置：
```bash
# 检查Agent进程
ps aux | grep ai-monitor-agent

# 查看Agent日志
tail -f /var/log/ai-monitor-agent.log

# 重启Agent
sudo systemctl restart ai-monitor-agent
```

### 性能相关

**Q: 系统响应慢**

A: 优化配置和资源：
```bash
# 检查系统资源
top
df -h
free -m

# 优化数据库
# 增加PostgreSQL内存配置
# 清理历史数据

# 调整监控频率
# 减少采集频率从30秒改为60秒
```

**Q: 磁盘空间不足**

A: 清理历史数据：
```bash
# 清理Docker镜像
docker system prune -a

# 清理日志文件
sudo find /var/log -name "*.log" -mtime +30 -delete

# 配置日志轮转
sudo logrotate -f /etc/logrotate.conf
```

## 📞 获取帮助

如果遇到问题无法解决，可以通过以下方式获取帮助：

1. **查看详细文档**
   - [部署指南](DEPLOYMENT_GUIDE.md)
   - [配置指南](CONFIGURATION_GUIDE.md)
   - [故障排除](TROUBLESHOOTING_GUIDE.md)

2. **社区支持**
   - [GitHub Issues](https://github.com/your-org/ai-monitor/issues)
   - [讨论区](https://github.com/your-org/ai-monitor/discussions)

3. **商业支持**
   - 邮件：support@ai-monitor.com
   - 技术支持热线：400-xxx-xxxx

## 🎉 下一步

恭喜！您已经成功部署并配置了AI Monitor系统。接下来您可以：

1. **探索高级功能**
   - 配置AI智能分析
   - 设置自动化运维脚本
   - 创建自定义仪表板

2. **扩展监控范围**
   - 添加更多监控目标
   - 配置应用性能监控
   - 集成第三方系统

3. **优化系统性能**
   - 调整监控策略
   - 配置数据保留策略
   - 设置高可用部署

---

**🌟 如果这个快速开始指南对您有帮助，请给我们一个Star！**