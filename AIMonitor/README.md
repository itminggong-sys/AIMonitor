# AI Monitor - 智能监控系统 v3.8.5

<div align="center">

![AI Monitor Logo](https://img.shields.io/badge/AI-Monitor-v3.8.5-blue?style=for-the-badge&logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDJMMTMuMDkgOC4yNkwyMCA5TDEzLjA5IDE1Ljc0TDEyIDIyTDEwLjkxIDE1Ljc0TDQgOUwxMC45MSA4LjI2TDEyIDJaIiBmaWxsPSJ3aGl0ZSIvPgo8L3N2Zz4K)

[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21.0-blue.svg)](https://golang.org)
[![React Version](https://img.shields.io/badge/react-18.2.0-blue.svg)](https://reactjs.org)
[![TypeScript](https://img.shields.io/badge/typescript-5.0.4-blue.svg)](https://typescriptlang.org)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](https://docker.com)
[![Production Ready](https://img.shields.io/badge/production-ready-green.svg)](#)

**下一代智能监控系统 - 集成AI能力的全栈监控解决方案**

[快速开始](#快速开始) • [功能特性](#功能特性) • [架构设计](#架构设计) • [部署指南](#部署指南) • [API文档](#api文档) • [贡献指南](#贡献指南)

</div>

## 🚀 项目简介

AI Monitor是一个现代化的智能监控系统，基于React 18 + Go 1.21技术栈构建，集成了OpenAI GPT-4和Claude 3等先进AI能力。系统提供全方位的基础设施监控、智能告警分析、自动化运维建议等功能，帮助企业构建高效、智能的IT运维体系。

### 🎯 版本特性 (v3.8.5)

- ✅ **生产就绪** - 完整的前后端实现，通过全面测试
- ✅ **Docker部署** - 一键部署，开箱即用
- ✅ **AI集成** - 支持OpenAI和Claude智能分析
- ✅ **实时监控** - WebSocket实时数据推送
- ✅ **现代化UI** - 基于Ant Design 5的响应式界面

### ✨ 核心亮点

- 🤖 **AI智能分析** - 集成GPT-4/Claude 3，提供智能告警分析和运维建议
- 🔄 **自动发现** - 支持网络扫描、服务识别、Agent自动注册
- 📊 **多维监控** - 服务器、数据库、中间件、应用全覆盖监控
- 🎯 **精准告警** - 智能告警规则，减少误报，提高运维效率
- 🌐 **现代化UI** - 基于React 18 + Ant Design 5的响应式界面
- 🐳 **容器化部署** - 支持Docker、Kubernetes等多种部署方式
- 🔒 **企业级安全** - JWT认证、RBAC权限、API限流等安全机制
- 📈 **高性能架构** - Go微服务架构，支持水平扩展

## 🎯 功能特性

### 监控能力

| 监控类型 | 支持对象 | 核心指标 |
|---------|---------|----------|
| **基础设施** | 服务器、虚拟机、容器 | CPU、内存、磁盘、网络、进程 |
| **数据库** | MySQL、PostgreSQL、Redis、MongoDB | 连接数、QPS、慢查询、复制延迟 |
| **中间件** | Nginx、Apache、Kafka、RabbitMQ | 请求量、响应时间、队列深度 |
| **应用服务** | HTTP服务、API接口、微服务 | 可用性、响应时间、错误率 |
| **网络设备** | 交换机、路由器、防火墙 | 端口状态、流量统计、SNMP指标 |

### AI智能功能

- **智能告警分析** - AI分析告警根因，提供解决建议
- **异常检测** - 基于机器学习的异常模式识别
- **容量预测** - 智能预测资源使用趋势
- **运维建议** - 基于历史数据的优化建议
- **自然语言查询** - 支持自然语言查询监控数据

### 自动化运维

- **自动发现** - 网络扫描、服务识别、Agent部署
- **自动注册** - 新设备自动接入监控
- **自动修复** - 预定义修复脚本自动执行
- **批量操作** - 支持批量配置、批量部署

## 🏗️ 架构设计

### 技术栈

#### 前端技术栈
- **框架**: React 18 + TypeScript 5
- **UI库**: Ant Design 5
- **状态管理**: Zustand
- **路由**: React Router 6
- **构建工具**: Vite 4
- **图表库**: ECharts、D3.js

#### 后端技术栈
- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL 13+ / MySQL 8.0+
- **缓存**: Redis 6.0+
- **消息队列**: Kafka / RabbitMQ
- **搜索引擎**: Elasticsearch 8.0+

#### 监控技术栈
- **指标收集**: Prometheus
- **可视化**: Grafana
- **告警**: Alertmanager
- **链路追踪**: Jaeger
- **日志聚合**: ELK Stack

#### AI集成
- **OpenAI**: GPT-4, GPT-3.5-turbo
- **Anthropic**: Claude 3
- **本地模型**: Ollama, LocalAI支持

### 系统架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Frontend  │    │   Mobile App    │    │   API Clients   │
│   (React 18)    │    │   (React Native)│    │   (Third-party)  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │      API Gateway          │
                    │   (Nginx/Kong/Traefik)   │
                    └─────────────┬─────────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
    ┌─────────┴─────────┐ ┌───────┴───────┐ ┌─────────┴─────────┐
    │   Auth Service    │ │  Core Service │ │  Monitor Service  │
    │   (JWT/OAuth)     │ │   (Business)  │ │   (Metrics/AI)    │
    └─────────┬─────────┘ └───────┬───────┘ └─────────┬─────────┘
              │                   │                   │
              └───────────────────┼───────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    │      Data Layer           │
                    └─────────────┬─────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
 ┌──────┴──────┐         ┌────────┴────────┐       ┌────────┴────────┐
 │ PostgreSQL  │         │     Redis       │       │  Elasticsearch  │
 │ (Metadata)  │         │   (Cache)       │       │    (Logs)       │
 └─────────────┘         └─────────────────┘       └─────────────────┘
```

## 🚀 快速开始（小白用户专用）

### 🎯 零技术门槛部署

我们为非技术用户提供了超级简单的一键部署方式，无需任何技术背景，只需运行一个脚本即可完成整个系统的部署。

### 环境要求

- **操作系统**: Windows 10+, Linux, macOS
- **内存**: 4GB+ (推荐8GB+)
- **存储**: 50GB+ 可用空间
- **网络**: 稳定的互联网连接

### 🚀 一键部署（推荐）

#### 🎯 Windows用户（最简单）

```batch
# 1. 下载项目到本地（如：C:\AIMonitor）
# 2. 以管理员身份打开命令提示符
# 3. 进入项目目录
cd C:\AIMonitor

# 4. 运行一键安装脚本（推荐小白用户）
quick-install.bat

# 或者运行完整安装脚本
install.bat
```

#### 🐧 Linux/macOS用户

```bash
# 1. 下载项目
git clone https://github.com/your-org/ai-monitor.git
cd ai-monitor

# 2. 运行一键安装脚本（推荐小白用户）
chmod +x quick-install.sh
./quick-install.sh

# 或者运行完整安装脚本
chmod +x deploy.sh
./deploy.sh
```

#### 🐳 技术用户快速部署

如果您已经安装了Docker，可以直接使用：

```bash
# 生产环境部署
docker-compose -f docker-compose.prod.yml up -d

# 开发环境部署
docker-compose up -d
```

### ✅ 脚本自动完成的工作

- ✅ 检查系统环境和权限
- ✅ 自动安装Docker和相关依赖
- ✅ 自动安装PostgreSQL数据库
- ✅ 自动安装Redis缓存服务
- ✅ 构建前后端应用
- ✅ 启动所有服务
- ✅ 配置Nginx反向代理
- ✅ 创建默认管理员账户
- ✅ 设置系统服务自启动

### 🎉 部署完成后访问系统

部署成功后，您将看到以下访问信息：

```
🎉 AI监控系统部署成功！

📱 访问地址：
- 主界面：http://localhost:8080
- 管理后台：http://localhost:8080/admin
- API文档：http://localhost:8080/api/docs
- 监控面板：http://localhost:3000 (Grafana)

🔐 默认账户：
- 管理员：admin / admin123
- 普通用户：user / user123

📊 数据库连接：
- PostgreSQL：localhost:5432
- Redis：localhost:6379

⚠️ 重要提醒：
1. 首次登录后请立即修改默认密码
2. 生产环境请配置SSL证书
3. 定期备份数据库数据
4. 监控系统资源使用情况
```

### 🔧 常用管理命令

```bash
# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 更新服务
docker-compose pull && docker-compose up -d
```

## 📖 详细文档

### 用户文档
- [📋 用户手册](doc/USER_MANUAL.md) - 系统使用指南
- [🚀 部署指南](doc/DEPLOYMENT_GUIDE.md) - 详细部署说明
- [⚙️ 配置指南](doc/CONFIGURATION_GUIDE.md) - 系统配置说明
- [🔧 故障排除](doc/TROUBLESHOOTING_GUIDE.md) - 常见问题解决

### 开发文档
- [🏗️ 架构设计](doc/ARCHITECTURE.md) - 系统架构详解
- [📝 API文档](doc/API_DOCUMENTATION.md) - RESTful API说明
- [🧪 测试指南](doc/TESTING_GUIDE.md) - 测试策略和方法
- [📊 性能指南](doc/PERFORMANCE_GUIDE.md) - 性能优化建议

### 运维文档
- [🔒 安全指南](doc/SECURITY_GUIDE.md) - 安全配置和最佳实践
- [📈 监控运维](doc/MONITORING_GUIDE.md) - 系统监控和运维
- [💾 备份恢复](doc/BACKUP_RECOVERY.md) - 数据备份和恢复

## 🛠️ 开发指南

### 本地开发环境

```bash
# 1. 启动后端服务
cd backend
go mod download
go run cmd/server/main.go

# 2. 启动前端服务
cd web
npm install
npm run dev

# 3. 启动中间件（可选）
docker compose up -d postgres redis prometheus
```

### 项目结构

```
ai-monitor/
├── web/                    # 前端项目
│   ├── src/
│   │   ├── components/     # 通用组件
│   │   ├── pages/         # 页面组件
│   │   ├── services/      # API服务
│   │   └── utils/         # 工具函数
│   └── package.json
├── cmd/                    # 后端入口
│   ├── server/            # 主服务
│   └── migrate/           # 数据库迁移
├── internal/               # 内部包
│   ├── api/               # API路由
│   ├── service/           # 业务逻辑
│   ├── model/             # 数据模型
│   └── config/            # 配置管理
├── agents/                 # 监控Agent
│   ├── windows/           # Windows Agent
│   ├── linux/             # Linux Agent
│   └── macos/             # macOS Agent
├── deploy/                 # 部署脚本
│   ├── docker-deploy.yml  # Docker Compose
│   ├── deploy.sh          # 主部署脚本
│   └── install.bat        # Windows安装
├── configs/                # 配置文件
├── doc/                    # 项目文档
└── README.md
```

### 贡献代码

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 🤝 社区支持

### 获取帮助

- 📖 [文档中心](doc/) - 查看详细文档
- 🐛 [问题反馈](https://github.com/your-org/ai-monitor/issues) - 报告Bug或提出建议
- 💬 [讨论区](https://github.com/your-org/ai-monitor/discussions) - 技术讨论和交流
- 📧 [邮件支持](mailto:support@ai-monitor.com) - 商业支持

### 贡献者

感谢所有为项目做出贡献的开发者！

<a href="https://github.com/your-org/ai-monitor/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=your-org/ai-monitor" />
</a>

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=your-org/ai-monitor&type=Date)](https://star-history.com/#your-org/ai-monitor&Date)

---

<div align="center">

**如果这个项目对您有帮助，请给我们一个 ⭐ Star！**

[⬆ 回到顶部](#ai-monitor---智能监控系统)

</div>