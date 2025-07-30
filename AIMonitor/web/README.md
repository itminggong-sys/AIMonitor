# AI Monitor System - Web Frontend

现代化的智能监控系统前端界面，基于 React 18 + TypeScript + Ant Design 5 构建。

## 🚀 特性

- **现代化技术栈**: React 18, TypeScript, Vite
- **优雅UI设计**: Ant Design 5 组件库
- **数据可视化**: ECharts + Recharts 图表库
- **状态管理**: Zustand + React Query
- **路由管理**: React Router 6
- **响应式设计**: 支持桌面端和移动端
- **国际化支持**: 中文界面优化
- **SEO优化**: React Helmet Async

## 📦 功能模块

### 核心监控
- 🏠 **仪表板**: 系统概览和关键指标
- 📊 **系统监控**: 服务器和进程监控
- 🚨 **告警管理**: 告警规则和历史记录
- 🤖 **AI分析**: 智能异常检测和趋势预测

### 专业监控
- 🔧 **中间件监控**: Redis, MySQL, Nginx等
- ⚡ **APM监控**: 应用性能和链路追踪
- 🐳 **容器监控**: Docker容器管理

### 系统管理
- ⚙️ **系统设置**: 配置和参数管理
- 👤 **个人资料**: 用户信息和偏好设置

## 🛠️ 技术架构

```
src/
├── components/          # 公共组件
│   ├── Layout/         # 主布局组件
│   └── NotificationPanel/ # 通知面板
├── pages/              # 页面组件
│   ├── Dashboard/      # 仪表板
│   ├── Monitoring/     # 系统监控
│   ├── Alerts/         # 告警管理
│   ├── AIAnalysis/     # AI分析
│   ├── Middleware/     # 中间件监控
│   ├── APM/           # APM监控
│   ├── Containers/     # 容器监控
│   ├── Settings/       # 系统设置
│   ├── Profile/        # 个人资料
│   └── Login/         # 登录页面
├── store/              # 状态管理
│   └── authStore.ts   # 认证状态
├── styles/             # 样式文件
│   └── index.css      # 全局样式
├── App.tsx            # 主应用组件
└── main.tsx           # 应用入口
```

## 🚀 快速开始

### 环境要求
- Node.js >= 16.0.0
- npm >= 8.0.0

### 安装依赖
```bash
npm install
```

### 开发模式
```bash
npm run dev
```
访问 http://localhost:3000

### 构建生产版本
```bash
npm run build
```

### 预览生产版本
```bash
npm run preview
```

### 代码检查
```bash
npm run lint
```

## 🔧 配置说明

### API 代理配置
开发环境下，API请求会自动代理到后端服务器:
- 前端: http://localhost:3000
- 后端: http://localhost:8080

### 环境变量
可以创建 `.env.local` 文件配置环境变量:
```
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_TITLE=AI Monitor System
```

## 📱 响应式设计

- **桌面端**: >= 1200px - 完整功能界面
- **平板端**: 768px - 1199px - 适配布局
- **移动端**: < 768px - 抽屉式导航

## 🎨 设计规范

### 颜色主题
- 主色调: #1890ff (蓝色)
- 成功色: #52c41a (绿色)
- 警告色: #faad14 (橙色)
- 错误色: #ff4d4f (红色)

### 组件规范
- 统一使用 Ant Design 组件
- 保持一致的间距和圆角
- 响应式栅格布局
- 现代化卡片阴影效果

## 🔐 认证系统

### 演示账号
- 用户名: `admin`
- 密码: `admin123`

### 权限控制
- 基于角色的访问控制 (RBAC)
- 路由级别权限验证
- API请求自动携带认证头

## 📊 数据可视化

### 图表类型
- **折线图**: 趋势分析 (ECharts)
- **柱状图**: 数据对比 (Recharts)
- **进度条**: 资源使用率
- **统计卡片**: 关键指标展示

## 🚀 部署建议

### Nginx 配置示例
```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /path/to/dist;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系我们

- 项目地址: [GitHub Repository]
- 问题反馈: [Issues]
- 邮箱: support@aimonitor.com

---

**AI Monitor System** - 让监控更智能，让运维更简单！ 🚀