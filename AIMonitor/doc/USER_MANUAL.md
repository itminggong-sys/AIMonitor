# AI智能监控系统用户手册

## 目录

1. [系统概述](#系统概述)
2. [快速开始](#快速开始)
3. [用户管理](#用户管理)
4. [监控配置](#监控配置)
5. [告警管理](#告警管理)
6. [仪表板使用](#仪表板使用)
7. [AI分析功能](#AI分析功能)
8. [系统配置](#系统配置)
9. [常见问题](#常见问题)
10. [故障排除](#故障排除)

## 系统概述

### 系统简介

AI智能监控系统是一个基于React+Go技术栈构建的现代化监控平台，集成OpenAI和Claude AI能力，为企业提供全栈监控、智能分析和自动化运维解决方案。

### 核心功能

- **🚀 现代化架构**：React 18 + Go 1.21 + TypeScript 5构建的高性能系统
- **🤖 AI智能分析**：集成OpenAI GPT-4和Claude 3，提供智能故障诊断和预测
- **📊 全栈监控**：支持系统、应用、中间件、容器等全方位监控
- **🔔 智能告警**：多级告警策略、自动收敛和智能降噪
- **🛡️ 企业级安全**：JWT认证、RBAC权限控制、审计日志
- **📱 响应式设计**：支持桌面端和移动端，适配各种屏幕尺寸
- **⚡ 实时数据**：WebSocket实时推送，毫秒级数据更新
- **🎨 现代UI**：基于Ant Design 5和ECharts 5.4的美观界面

### 技术架构

**前端技术栈**：
- React 18 SPA架构
- TypeScript 5类型安全
- Ant Design 5组件库
- ECharts 5.4数据可视化
- Vite构建工具

**后端技术栈**：
- Go 1.21高性能服务
- Gin Web框架
- 微服务架构设计
- gRPC服务通信

**AI集成**：
- OpenAI GPT-4智能分析
- Claude 3辅助诊断
- 自然语言查询
- 智能报告生成

**数据存储**：
- MySQL关系型数据库
- Redis缓存和会话
- InfluxDB时序数据
- Elasticsearch日志搜索

**监控组件**：
- Prometheus指标采集
- Grafana可视化
- 多平台Agent支持
- 自定义采集器

### 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web前端界面   │    │   移动端应用    │    │   API接口       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────────────────────┼─────────────────────────────────┐
│                          AI监控系统后端                          │
├─────────────────┬───────────────┼───────────────┬─────────────────┤
│   用户管理模块   │   监控数据模块 │   告警管理模块 │   AI分析模块    │
└─────────────────┴───────────────┼───────────────┴─────────────────┘
                                 │
┌─────────────────────────────────┼─────────────────────────────────┐
│                            数据存储层                            │
├─────────────────┬───────────────┼───────────────┬─────────────────┤
│   PostgreSQL    │     Redis     │   时序数据库   │   文件存储      │
└─────────────────┴───────────────┴───────────────┴─────────────────┘
```

### 用户角色

- **超级管理员**：拥有所有权限，可以管理系统配置和用户
- **管理员**：可以管理监控配置、告警规则和用户权限
- **操作员**：可以查看监控数据、处理告警和配置仪表板
- **观察者**：只能查看监控数据和仪表板，无法进行配置操作

## 快速开始

### 🚀 部署完成后的首次使用

如果您已经使用一键部署脚本（`quick-install.bat` 或 `quick-install.sh`）完成了系统部署，请按以下步骤开始使用：

#### ✅ 1. 验证部署状态

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

#### ✅ 2. 访问系统

1. 打开现代浏览器（推荐 Chrome、Edge 或 Firefox）
2. 访问：`http://localhost:3000`（前端界面）
3. 后端API地址：`http://localhost:8080`
4. 您将看到 AI Monitor 的登录界面

### 系统要求

**浏览器要求**：
- Chrome 90+, Firefox 88+, Safari 14+, Edge 90+
- 屏幕分辨率：1280x720 或更高
- JavaScript：必须启用
- 网络：稳定的互联网连接

**已通过一键部署自动安装**：
- ✅ Docker 和 Docker Compose
- ✅ PostgreSQL 数据库
- ✅ Redis 缓存服务
- ✅ Nginx 反向代理
- ✅ 前后端应用服务

**可选配置**：
- OpenAI API密钥（用于AI智能分析）
- Claude API密钥（用于AI辅助诊断）
- 邮件服务配置（用于告警通知）

### 系统登录

1. **访问系统**
   - 打开现代浏览器，访问：`http://localhost:3000`
   - 系统支持响应式设计，自动适配设备屏幕

2. **默认管理员账号**
   - 用户名：`admin`
   - 密码：`admin123`
   - 首次登录后请立即修改密码

3. **用户认证**
   - 输入用户名和密码
   - 支持记住登录状态（24小时有效期）
   - 点击"登录"按钮进入系统

3. **安全特性**
   - JWT Token认证，自动刷新
   - 登录失败5次自动锁定账户
   - 支持双因素认证（2FA）
   - 异地登录安全提醒

**默认管理员账户**（一键部署自动创建）：
- 用户名：`admin`
- 密码：`admin123`
- 角色：超级管理员

**⚠️ 安全提醒**：首次登录后请立即修改默认密码！

![登录界面](images/login.png)

### 首次登录设置

首次登录后，建议完成以下设置：

1. **修改默认密码**
   - 点击右上角用户头像
   - 选择"个人设置"
   - 在"安全设置"中修改密码

2. **配置个人信息**
   - 填写姓名、邮箱等基本信息
   - 设置通知偏好

3. **了解界面布局**
   - 顶部导航栏：主要功能模块
   - 左侧菜单：详细功能列表
   - 主内容区：具体功能界面
   - 右侧面板：通知和快捷操作

### 主界面概览

系统采用现代化的React SPA设计，提供直观的用户体验：

**主要布局区域**：
- **顶部导航栏**：系统logo、全局搜索、通知中心、用户菜单
- **左侧菜单**：可折叠的功能模块导航，支持多级菜单
- **主内容区**：动态加载的页面内容，支持标签页切换
- **右侧面板**：实时通知、快捷操作、系统状态

**交互特性**：
- **响应式设计**：自动适配桌面端和移动端
- **主题切换**：支持亮色/暗色主题
- **国际化**：支持中文/英文界面切换
- **快捷键**：支持键盘快捷操作
- **实时更新**：WebSocket实时数据推送

```
┌─────────────────────────────────────────────────────────────────┐
│  🏠 AI监控  🔍搜索  📊概览  🚨告警  📈仪表板  ⚙️配置  👤admin ▼ 🔔│
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────┐ │                                               │📱│
│ │📊监控概览│ │                主内容区域                      │📋│
│ │🎯监控目标│ │         (支持标签页切换)                      │🔔│
│ │🚨告警规则│ │                                               │⚡│
│ │📈仪表板  │ │                                               │📊│
│ │👥用户管理│ │                                               │  │
│ │⚙️系统设置│ │                                               │  │
│ │📝审计日志│ │                                               │  │
│ │🤖AI分析 │ │                                               │  │
│ └─────────┘ │                                               │  │
└─────────────────────────────────────────────────────────────────┘
```

## 用户管理

### 用户注册

管理员可以为新用户创建账户：

1. 进入"用户管理" > "用户列表"
2. 点击"添加用户"按钮
3. 填写用户信息：
   - 用户名（必填，唯一）
   - 邮箱（必填，用于通知）
   - 姓名（必填）
   - 角色（必选）
   - 部门（可选）
   - 电话（可选）
4. 点击"保存"完成创建

### 用户权限管理

#### 角色权限矩阵

| 功能模块 | 超级管理员 | 管理员 | 操作员 | 观察者 |
|---------|-----------|--------|--------|--------|
| 用户管理 | ✓ | ✓ | ✗ | ✗ |
| 系统配置 | ✓ | ✓ | ✗ | ✗ |
| 监控配置 | ✓ | ✓ | ✓ | ✗ |
| 告警管理 | ✓ | ✓ | ✓ | ✗ |
| 数据查看 | ✓ | ✓ | ✓ | ✓ |
| 仪表板编辑 | ✓ | ✓ | ✓ | ✗ |
| 审计日志 | ✓ | ✓ | ✗ | ✗ |

#### 权限分配

1. 进入"用户管理" > "角色管理"
2. 选择要编辑的角色
3. 在权限列表中勾选相应权限
4. 点击"保存"应用更改

### 个人设置

用户可以自行管理个人信息：

1. **基本信息**
   - 姓名、邮箱、电话等
   - 头像上传
   - 时区设置

2. **安全设置**
   - 修改密码
   - 启用两步验证
   - 查看登录历史

3. **通知设置**
   - 邮件通知偏好
   - 短信通知设置
   - 推送通知配置

## 监控配置

### 监控目标管理

#### 添加监控目标

1. 进入"监控配置" > "监控目标"
2. 点击"添加目标"按钮
3. 选择监控类型：
   - **服务器监控**：CPU、内存、磁盘、网络
   - **应用监控**：HTTP服务、数据库、消息队列
   - **网络监控**：网络连通性、延迟、带宽
   - **自定义监控**：通过API上报的自定义指标

4. 配置监控参数：

```yaml
# 服务器监控配置示例
name: "Web服务器-01"
type: "server"
host: "192.168.1.100"
port: 22
auth:
  type: "ssh_key"
  username: "monitor"
  private_key_path: "/etc/ssh/monitor_key"
metrics:
  - cpu_usage
  - memory_usage
  - disk_usage
  - network_io
interval: 60  # 采集间隔（秒）
timeout: 30   # 超时时间（秒）
```

#### 监控指标配置

每种监控类型支持不同的指标：

**服务器指标**：
- CPU使用率、负载平均值
- 内存使用率、交换空间
- 磁盘使用率、I/O性能
- 网络流量、连接数
- 进程数、文件描述符

**应用指标**：
- HTTP响应时间、状态码分布
- 数据库连接数、查询性能
- 消息队列长度、处理速度
- 自定义业务指标

**网络指标**：
- Ping延迟、丢包率
- 端口连通性
- 带宽使用率
- DNS解析时间

### 数据采集配置

#### 采集频率设置

根据监控需求设置合适的采集频率：

- **高频监控**（10-30秒）：关键业务系统
- **标准监控**（1-5分钟）：一般服务器和应用
- **低频监控**（10-30分钟）：批处理任务、备份系统

#### 数据保留策略

```yaml
# 数据保留配置
retention:
  raw_data: "7d"      # 原始数据保留7天
  hourly_data: "30d"  # 小时聚合数据保留30天
  daily_data: "1y"    # 日聚合数据保留1年
  monthly_data: "5y"  # 月聚合数据保留5年
```

## 告警管理

### 智能告警规则配置

#### 创建智能告警规则

1. 进入"🚨告警管理" > "告警规则"
2. 点击"➕新建规则"按钮
3. 配置规则基本信息：
   - 规则名称和描述
   - 优先级（低、中、高、紧急）
   - AI增强选项（动态阈值、异常检测）
   - 启用状态和生效时间

4. 设置智能触发条件：

```yaml
# 智能告警规则示例
name: "CPU使用率智能告警"
description: "基于AI的CPU使用率异常检测"
priority: "high"
ai_enhanced: true
conditions:
  - metric: "cpu_usage"
    operator: ">"
    threshold: 80
    dynamic_threshold: true  # AI动态阈值
    duration: "5m"
    sensitivity: "medium"    # 敏感度：low/medium/high
  - metric: "load_average"
    operator: ">"
    threshold: 4
    duration: "3m"
    anomaly_detection: true  # 异常检测
logic: "AND"  # 条件逻辑：AND/OR
ai_analysis:
  root_cause_analysis: true
  impact_assessment: true
  auto_suggestion: true
```

#### 智能告警级别定义

- 🔴 **紧急 (Critical)**：系统完全不可用，AI自动触发应急响应流程
- 🟠 **高 (High)**：严重影响系统功能，AI提供快速解决方案
- 🟡 **中 (Medium)**：影响部分功能，AI进行趋势分析和预警
- 🔵 **低 (Low)**：轻微问题，AI自动归类和批量处理

#### AI增强特性

- **动态阈值学习**：基于历史数据自动调整告警阈值
- **异常模式识别**：识别非线性和复杂的异常模式
- **告警收敛算法**：智能合并相关告警，减少噪音
- **预测性告警**：提前预警可能发生的问题

### 通知配置

#### 通知渠道

系统支持多种通知方式：

1. **邮件通知**
```yaml
email:
  smtp_server: "smtp.company.com"
  smtp_port: 587
  username: "alert@company.com"
  password: "password"
  from: "AI监控系统 <alert@company.com>"
```

2. **短信通知**
```yaml
sms:
  provider: "aliyun"  # 支持阿里云、腾讯云等
  access_key: "your_access_key"
  secret_key: "your_secret_key"
  template_id: "SMS_123456789"
```

3. **Webhook通知**
```yaml
webhook:
  url: "https://hooks.slack.com/services/xxx"
  method: "POST"
  headers:
    Content-Type: "application/json"
  template: |
    {
      "text": "告警：{{.AlertName}}",
      "attachments": [
        {
          "color": "{{.Color}}",
          "fields": [
            {
              "title": "级别",
              "value": "{{.Priority}}",
              "short": true
            },
            {
              "title": "时间",
              "value": "{{.Timestamp}}",
              "short": true
            }
          ]
        }
      ]
    }
```

#### 通知策略

配置不同级别告警的通知策略：

```yaml
notification_policies:
  - name: "紧急告警"
    priority: "critical"
    channels: ["email", "sms", "webhook"]
    escalation:
      - delay: "0m"
        recipients: ["oncall_engineer"]
      - delay: "5m"
        recipients: ["team_lead"]
      - delay: "15m"
        recipients: ["manager"]
  
  - name: "高级告警"
    priority: "high"
    channels: ["email", "webhook"]
    escalation:
      - delay: "0m"
        recipients: ["team_members"]
      - delay: "30m"
        recipients: ["team_lead"]
```

### 告警处理

#### 告警状态管理

告警具有以下状态：

- **触发**：告警条件满足，产生新告警
- **确认**：运维人员确认收到告警
- **处理中**：正在处理告警问题
- **已解决**：问题已解决，告警关闭
- **已忽略**：告警被标记为误报或可忽略

#### 告警处理流程

1. **接收告警**
   - 系统自动发送通知
   - 运维人员收到告警信息

2. **确认告警**
   - 登录系统查看详细信息
   - 点击"确认"按钮
   - 添加处理备注

3. **问题诊断**
   - 查看相关监控数据
   - 分析告警原因
   - 制定解决方案

4. **问题解决**
   - 执行解决方案
   - 验证问题是否解决
   - 更新告警状态

5. **总结归档**
   - 记录处理过程
   - 更新知识库
   - 优化告警规则

## 仪表板使用

### 预置仪表板

系统提供多个预置仪表板：

1. **系统概览**
   - 整体健康状态
   - 关键指标趋势
   - 告警统计

2. **服务器监控**
   - CPU、内存、磁盘使用率
   - 网络流量
   - 系统负载

3. **应用性能**
   - 响应时间
   - 吞吐量
   - 错误率

4. **网络监控**
   - 网络延迟
   - 带宽使用
   - 连通性状态

### 自定义仪表板

#### 创建仪表板

1. 进入"仪表板"模块
2. 点击"新建仪表板"
3. 设置仪表板属性：
   - 名称和描述
   - 访问权限
   - 刷新间隔
   - 时间范围

#### 添加图表组件

支持多种图表类型：

1. **时间序列图**
   - 适用于趋势分析
   - 支持多指标对比
   - 可设置阈值线

2. **饼图/环形图**
   - 适用于比例展示
   - 资源使用分布
   - 状态统计

3. **柱状图/条形图**
   - 适用于分类对比
   - 排行榜展示
   - 历史对比

4. **数值面板**
   - 显示单一指标
   - 支持阈值颜色
   - 趋势指示器

5. **表格**
   - 详细数据列表
   - 支持排序筛选
   - 可配置列显示

6. **热力图**
   - 时间维度分析
   - 密度分布展示
   - 异常点识别

#### 图表配置示例

```json
{
  "type": "timeseries",
  "title": "CPU使用率趋势",
  "datasource": "prometheus",
  "targets": [
    {
      "expr": "cpu_usage{instance=~\"web-.*\"}",
      "legend": "{{instance}}",
      "refId": "A"
    }
  ],
  "options": {
    "legend": {
      "displayMode": "table",
      "placement": "right"
    },
    "tooltip": {
      "mode": "multi"
    },
    "thresholds": [
      {
        "value": 80,
        "color": "orange"
      },
      {
        "value": 90,
        "color": "red"
      }
    ]
  }
}
```

### 仪表板分享

#### 权限控制

- **私有**：仅创建者可见
- **团队共享**：团队成员可见
- **公开**：所有用户可见
- **只读链接**：生成匿名访问链接

#### 导出功能

- **PDF导出**：生成报告文档
- **图片导出**：保存为PNG/JPG
- **数据导出**：导出为CSV/Excel
- **配置导出**：导出仪表板配置JSON

## AI分析功能

### AI分析功能

### 智能告警分析

#### 多维度告警关联分析

AI系统采用深度学习算法自动分析告警之间的复杂关联关系：

1. **时间关联分析**：
   - 识别时间窗口内的告警集群
   - 分析告警发生的时序模式
   - 检测周期性告警规律

2. **空间关联分析**：
   - 识别同一系统、服务或地域的告警
   - 分析网络拓扑相关的告警传播
   - 检测跨系统的级联故障

3. **因果关联分析**：
   - 基于历史数据学习因果关系
   - 构建动态的因果关系图谱
   - 实时更新因果关系权重

4. **模式识别与学习**：
   - 识别重复出现的告警模式
   - 学习新的异常模式
   - 自动更新告警规则

#### AI驱动的根因分析

系统集成OpenAI GPT-4和Claude 3，提供企业级智能根因分析：

1. **多源数据融合**：
   - 收集告警、日志、指标、配置变更等多维数据
   - 整合外部数据源（CMDB、变更记录等）
   - 实时数据流处理和分析

2. **智能依赖分析**：
   - 自动发现系统组件依赖关系
   - 构建动态服务依赖图
   - 分析依赖链路的健康状态

3. **时序事件分析**：
   - 分析事件发生的精确时间顺序
   - 识别关键时间节点和触发事件
   - 构建事件时间线和影响链路

4. **概率推理引擎**：
   - 使用贝叶斯网络计算故障概率
   - 多假设并行验证
   - 置信度评分和不确定性量化

5. **智能解决方案生成**：
   - 基于知识库匹配解决方案
   - AI生成个性化修复建议
   - 自动化修复脚本推荐
   - 预防措施和优化建议

#### 根因分析报告示例

```json
{
  "analysis_id": "RCA-2024-001",
  "timestamp": "2024-01-15T10:30:00Z",
  "incident_summary": "Web服务响应时间异常增长",
  "root_causes": [
    {
      "cause": "数据库连接池耗尽",
      "confidence": 0.92,
      "evidence": [
        "数据库连接数达到上限",
        "应用日志显示连接超时错误",
        "数据库慢查询增加"
      ],
      "impact": "高",
      "solutions": [
        "增加数据库连接池大小",
        "优化慢查询SQL",
        "实施连接池监控"
      ]
    }
  ],
  "timeline": [
    {
      "time": "10:25:00",
      "event": "数据库慢查询开始增加"
    },
    {
      "time": "10:28:00",
      "event": "连接池使用率达到90%"
    },
    {
      "time": "10:30:00",
      "event": "Web服务响应时间告警触发"
    }
  ],
  "ai_insights": {
    "pattern_match": "与历史事件#2023-156相似",
    "prediction": "如不处理，预计30分钟内服务完全不可用",
    "prevention": "建议实施自动扩容和熔断机制"
  }
}
```

### 异常检测

#### 基线学习

AI系统会自动学习正常行为模式：

- **周期性模式**：识别日、周、月的周期性变化
- **趋势模式**：识别长期增长或下降趋势
- **季节性模式**：识别季节性变化规律
- **突发模式**：识别突发事件的影响模式

#### 异常类型

系统可以检测多种异常类型：

1. **点异常**：单个数据点的异常
2. **上下文异常**：在特定上下文中的异常
3. **集体异常**：一组数据的异常
4. **趋势异常**：趋势变化的异常

### 预测分析

#### 容量预测

基于历史数据预测资源使用趋势：

- **CPU使用率预测**：预测未来CPU使用情况
- **内存使用预测**：预测内存需求变化
- **磁盘空间预测**：预测磁盘空间耗尽时间
- **网络流量预测**：预测网络带宽需求

#### 故障预测

提前识别可能的系统故障：

- **硬件故障预测**：基于硬件指标预测故障
- **性能下降预测**：预测性能瓶颈
- **服务中断预测**：预测服务可用性风险

### AI分析报告

#### 自动报告生成

系统可以自动生成分析报告：

1. **日报**：每日系统健康状况总结
2. **周报**：一周内的趋势分析和异常总结
3. **月报**：月度性能分析和容量规划建议
4. **事件报告**：重大事件的详细分析报告

#### 报告内容

- **执行摘要**：关键发现和建议
- **详细分析**：数据分析和图表展示
- **趋势预测**：未来趋势预测
- **行动建议**：具体的改进建议
- **风险评估**：潜在风险识别

## 系统配置

### 基础配置

#### 系统参数

1. **时区设置**
   - 系统默认时区
   - 用户个人时区
   - 数据显示时区

2. **语言设置**
   - 界面语言
   - 日期格式
   - 数字格式

3. **主题设置**
   - 浅色主题
   - 深色主题
   - 自定义主题

#### 性能配置

```yaml
# 性能配置示例
performance:
  max_concurrent_queries: 100
  query_timeout: 30s
  cache_ttl: 300s
  batch_size: 1000
  worker_pool_size: 10
```

### 数据源配置

#### Prometheus集成

```yaml
prometheus:
  url: "http://prometheus:9090"
  timeout: 30s
  basic_auth:
    username: "admin"
    password: "password"
  custom_headers:
    X-Custom-Header: "value"
```

#### InfluxDB集成

```yaml
influxdb:
  url: "http://influxdb:8086"
  database: "monitoring"
  username: "monitor"
  password: "password"
  retention_policy: "autogen"
```

#### Elasticsearch集成

```yaml
elasticsearch:
  hosts:
    - "http://es1:9200"
    - "http://es2:9200"
  index_pattern: "logs-*"
  username: "elastic"
  password: "password"
```

### 安全配置

#### 认证配置

1. **本地认证**
```yaml
auth:
  type: "local"
  password_policy:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_symbols: true
    max_age_days: 90
```

2. **LDAP认证**
```yaml
auth:
  type: "ldap"
  ldap:
    server: "ldap://ldap.company.com:389"
    bind_dn: "cn=admin,dc=company,dc=com"
    bind_password: "password"
    user_search_base: "ou=users,dc=company,dc=com"
    user_search_filter: "(uid=%s)"
    group_search_base: "ou=groups,dc=company,dc=com"
```

3. **OAuth2认证**
```yaml
auth:
  type: "oauth2"
  oauth2:
    provider: "google"
    client_id: "your_client_id"
    client_secret: "your_client_secret"
    redirect_url: "http://your-domain/auth/callback"
    scopes: ["openid", "profile", "email"]
```

#### 访问控制

```yaml
access_control:
  session_timeout: 8h
  max_login_attempts: 5
  lockout_duration: 30m
  ip_whitelist:
    - "192.168.1.0/24"
    - "10.0.0.0/8"
  rate_limiting:
    requests_per_minute: 60
    burst_size: 10
```

### 备份配置

#### 自动备份

```yaml
backup:
  enabled: true
  schedule: "0 2 * * *"  # 每天凌晨2点
  retention_days: 30
  storage:
    type: "s3"
    bucket: "monitoring-backups"
    region: "us-west-2"
    access_key: "your_access_key"
    secret_key: "your_secret_key"
  include:
    - "database"
    - "configurations"
    - "dashboards"
  exclude:
    - "raw_metrics"  # 排除原始指标数据
```

## 常见问题

### 登录问题

**Q: 忘记密码怎么办？**

A: 有以下几种解决方案：
1. 使用"忘记密码"功能，系统会发送重置链接到注册邮箱
2. 联系管理员重置密码
3. 如果是管理员账户，可以通过命令行工具重置：
```bash
./aimonitor admin reset-password --username admin
```

**Q: 登录后提示权限不足？**

A: 检查以下几点：
1. 确认用户角色是否正确
2. 检查用户是否被禁用
3. 联系管理员检查权限配置

### 监控问题

**Q: 监控数据不更新？**

A: 排查步骤：
1. 检查监控目标是否在线
2. 验证网络连接是否正常
3. 检查认证信息是否正确
4. 查看采集器日志：
```bash
docker logs aimonitor-collector
```

**Q: 监控指标显示异常？**

A: 可能的原因：
1. 数据源配置错误
2. 指标计算公式有误
3. 时间范围设置不当
4. 数据采集延迟

### 告警问题

**Q: 告警不触发？**

A: 检查以下配置：
1. 告警规则是否启用
2. 阈值设置是否合理
3. 持续时间配置是否正确
4. 数据是否正常采集

**Q: 告警通知收不到？**

A: 排查步骤：
1. 检查通知渠道配置
2. 验证邮箱/手机号是否正确
3. 检查垃圾邮件文件夹
4. 查看通知发送日志

### 性能问题

**Q: 系统响应慢？**

A: 优化建议：
1. 检查数据库性能
2. 优化查询时间范围
3. 增加缓存配置
4. 检查系统资源使用

**Q: 仪表板加载慢？**

A: 解决方案：
1. 减少图表数量
2. 优化查询语句
3. 使用数据聚合
4. 调整刷新频率

## 故障排除

### 日志查看

#### 应用日志

```bash
# 查看应用日志
docker logs -f aimonitor-app

# 查看特定时间段的日志
docker logs --since="2024-01-01T00:00:00" --until="2024-01-01T23:59:59" aimonitor-app

# 查看错误日志
docker logs aimonitor-app 2>&1 | grep ERROR
```

## 主要功能模块详解

### 监控概览模块

监控概览是系统的核心仪表板，提供全方位的监控数据展示和AI智能分析：

**核心指标面板**：
- **系统性能**：CPU、内存、磁盘、网络实时状态
- **应用健康**：服务可用性、响应时间、错误率
- **中间件状态**：数据库、缓存、消息队列连接状态
- **容器监控**：Docker容器、Kubernetes集群状态
- **AI分析摘要**：智能故障预测和性能优化建议

**可视化图表**：
- **实时图表**：基于ECharts 5.4的高性能图表
- **时间序列**：支持1分钟到30天的历史数据查看
- **多维度对比**：支持多个监控目标同时对比
- **自定义视图**：可保存个人偏好的图表配置
- **响应式设计**：自动适配不同屏幕尺寸

**智能功能**：
- **异常检测**：AI自动识别性能异常和趋势变化
- **预测分析**：基于历史数据预测未来趋势
- **智能告警**：自动生成告警规则建议
- **性能优化**：AI提供系统优化建议

**操作指南**：
1. 点击左侧菜单「📊监控概览」
2. 使用顶部时间选择器设置查看范围
3. 通过筛选器选择监控目标和指标类型
4. 点击图表可查看详细数据和AI分析
5. 使用「🤖AI分析」按钮获取智能建议

### 监控目标管理

**目标类型支持**：
- **物理服务器**：Linux/Windows服务器监控
- **虚拟机**：VMware、Hyper-V等虚拟化平台
- **容器**：Docker容器和Kubernetes Pod
- **云服务**：AWS、Azure、阿里云等云资源
- **网络设备**：交换机、路由器、防火墙
- **应用服务**：Web服务、数据库、中间件

**批量管理功能**：
- **批量导入**：支持CSV、Excel文件批量导入
- **自动发现**：网络扫描自动发现监控目标
- **模板配置**：预定义监控模板快速部署
- **分组管理**：按业务、环境、地域等维度分组

### 实时监控功能

**实时数据流**：
- **WebSocket连接**：毫秒级数据推送
- **数据压缩**：优化网络传输效率
- **断线重连**：自动恢复连接机制
- **缓存机制**：本地缓存减少服务器压力

**性能指标**：
```yaml
# 系统指标配置
system_metrics:
  cpu:
    - usage_percent
    - load_average_1m
    - load_average_5m
    - load_average_15m
  memory:
    - usage_percent
    - available_bytes
    - swap_usage_percent
  disk:
    - usage_percent
    - read_iops
    - write_iops
    - read_throughput
    - write_throughput
  network:
    - bytes_sent
    - bytes_recv
    - packets_sent
    - packets_recv
    - errors
```

#### 数据库日志

```bash
# PostgreSQL日志
docker logs -f aimonitor-postgres

# Redis日志
docker logs -f aimonitor-redis
```

#### 系统日志

```bash
# 系统日志位置
/var/log/aimonitor/
├── app.log          # 应用日志
├── access.log       # 访问日志
├── error.log        # 错误日志
├── audit.log        # 审计日志
└── performance.log  # 性能日志
```

### 健康检查

#### 系统健康状态

```bash
# 检查系统健康状态
curl http://localhost:8080/api/health

# 检查各组件状态
curl http://localhost:8080/api/health/detailed
```

#### 数据库连接测试

```bash
# 测试数据库连接
./aimonitor db test-connection

# 检查数据库状态
./aimonitor db status
```

### 性能诊断

#### 系统资源监控

```bash
# CPU使用率
top -p $(pgrep aimonitor)

# 内存使用
ps aux | grep aimonitor

# 磁盘I/O
iotop -p $(pgrep aimonitor)

# 网络连接
netstat -tulpn | grep aimonitor
```

#### 应用性能分析

```bash
# 启用性能分析
./aimonitor --profile=cpu

# 生成性能报告
go tool pprof http://localhost:8080/debug/pprof/profile
```

### 数据恢复

#### 数据库恢复

```bash
# 从备份恢复数据库
pg_restore -h localhost -U postgres -d aimonitor backup.sql

# 恢复特定表
pg_restore -h localhost -U postgres -d aimonitor -t users backup.sql
```

#### 配置恢复

```bash
# 恢复配置文件
cp /backup/config/* /etc/aimonitor/

# 重启服务
systemctl restart aimonitor
```

### 联系支持

如果以上方法无法解决问题，请联系技术支持：

- **邮箱**：support@aimonitor.com
- **电话**：400-123-4567
- **在线支持**：https://support.aimonitor.com
- **文档中心**：https://docs.aimonitor.com

提供以下信息有助于快速解决问题：

1. 系统版本信息
2. 错误日志和截图
3. 问题复现步骤
4. 系统环境信息
5. 相关配置文件

---

**版本信息**：v1.0.0  
**更新日期**：2024年1月  
**文档维护**：AI监控系统开发团队