# AI Monitor 智能监控系统 - 统一部署指南

## 📋 目录

- [系统概述](#系统概述)
- [系统要求](#系统要求)
- [快速部署](#快速部署)
- [手动部署](#手动部署)
- [Docker部署](#docker部署)
- [Kubernetes部署](#kubernetes部署)
- [配置管理](#配置管理)
- [监控与运维](#监控与运维)
- [故障排除](#故障排除)
- [安全配置](#安全配置)

## 🎯 系统概述

### 技术栈

**前端技术栈：**
- React 18.2+ (UI框架)
- TypeScript 4.9+ (类型安全)
- Ant Design 5.0+ (UI组件库)
- Vite 4.0+ (构建工具)
- React Router 6.0+ (路由管理)
- Zustand 4.0+ (状态管理)
- React Query 4.0+ (数据获取)
- ECharts 5.4+ (图表库)

**后端技术栈：**
- Go 1.21+ (主要语言)
- Gin 1.9+ (Web框架)
- GORM 1.25+ (ORM框架)
- PostgreSQL 15+ (主数据库)
- Redis 7.0+ (缓存/消息队列)
- JWT (身份认证)
- Swagger (API文档)

**基础设施：**
- Docker 24.0+ & Docker Compose 2.20+
- Nginx 1.24+ (反向代理)
- Kubernetes 1.28+ (容器编排)
- Prometheus 2.40+ (监控)
- Grafana 9.0+ (可视化)
- Jaeger 1.45+ (链路追踪)

**AI集成：**
- OpenAI GPT-4 (智能分析)
- Claude 3 (辅助分析)
- 本地LLM支持 (Ollama)

## 💻 系统要求

### 最低配置
- **CPU**: 2核心
- **内存**: 4GB RAM
- **存储**: 20GB 可用空间
- **网络**: 稳定的互联网连接

### 推荐配置
- **CPU**: 4核心或更多
- **内存**: 8GB RAM 或更多
- **存储**: 50GB SSD
- **网络**: 100Mbps 带宽

### 支持的操作系统
- **Linux**: Ubuntu 20.04+, CentOS 8+, Debian 11+
- **Windows**: Windows 10/11, Windows Server 2019+
- **macOS**: macOS 12+ (Monterey)

## 🚀 快速部署

我们为不同用户群体提供了多种部署方式，从零技术背景的小白用户到专业运维人员都能找到适合的方案。

### 🎯 方式一：Windows用户（推荐）

**适用于：** Windows 10/11 或 Windows Server 2019+

```batch
# 1. 以管理员身份打开命令提示符或PowerShell
# 2. 进入项目根目录
cd C:\path\to\AIMonitor

# 3. 运行一键安装脚本（推荐）
.\quick-install.bat

# 或使用PowerShell脚本
.\deploy_windows.ps1
```

**脚本会自动完成：**
- ✅ 检查系统环境和权限
- ✅ 自动安装Docker Desktop
- ✅ 自动安装PostgreSQL和Redis
- ✅ 配置数据库和缓存
- ✅ 构建和启动前后端服务
- ✅ 配置Nginx反向代理
- ✅ 设置防火墙规则
- ✅ 创建默认管理员账户

### 🐧 方式二：Linux/macOS用户（推荐）

**适用于：** Ubuntu 20.04+, CentOS 8+, macOS 12+

```bash
# 1. 进入项目根目录
cd /path/to/AIMonitor

# 2. 给脚本执行权限
chmod +x quick-install.sh

# 3. 运行一键部署脚本（推荐）
sudo ./quick-install.sh

# 或使用智能安装脚本
sudo ./auto-deploy.sh
```

**脚本会自动完成：**
- ✅ 检查系统环境和依赖
- ✅ 自动安装Docker和Docker Compose
- ✅ 自动安装PostgreSQL和Redis
- ✅ 配置数据库连接和初始化
- ✅ 构建前后端Docker镜像
- ✅ 启动所有服务容器
- ✅ 配置Nginx负载均衡
- ✅ 设置系统服务自启动

### 🐳 方式三：Docker一键部署（技术用户）

**适用于：** 已安装Docker的用户

```bash
# 使用生产级Docker Compose配置
docker-compose -f docker-compose.prod.yml up -d

# 或使用开发环境配置
docker-compose up -d
```

### 📋 部署完成后的访问信息

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

### 🔧 部署脚本高级选项

如果您需要自定义配置，可以使用以下高级选项：

#### Windows高级部署
```batch
# 自定义端口和数据库密码
install.bat --port 9090 --db-password mypassword

# 启用SSL和域名配置
install.bat --ssl --domain monitor.company.com
```

#### Linux/macOS高级部署
```bash
# 自定义配置部署
./deploy.sh --port 9090 --db-password mypassword --jwt-secret mysecret

# 生产环境部署（包含SSL）
./deploy.sh --mode production --domain monitor.company.com --ssl
```

**支持的参数：**
- `--port`: 自定义Web服务端口 (默认: 8080)
- `--db-password`: 数据库密码
- `--jwt-secret`: JWT密钥
- `--domain`: 域名配置
- `--ssl`: 启用SSL证书
- `--mode`: 部署模式 (dev/test/prod)

## 🔧 手动部署

### 1. 环境准备

#### 安装Go

```bash
# Linux
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# macOS
brew install go

# Windows
# 下载并安装 https://go.dev/dl/go1.21.0.windows-amd64.msi
```

#### 安装Node.js

```bash
# Linux (使用NodeSource)
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# macOS
brew install node

# Windows
# 下载并安装 https://nodejs.org/dist/v18.17.0/node-v18.17.0-x64.msi
```

#### 安装PostgreSQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql postgresql-contrib

# CentOS/RHEL
sudo yum install postgresql-server postgresql-contrib
sudo postgresql-setup initdb
sudo systemctl enable postgresql
sudo systemctl start postgresql

# macOS
brew install postgresql
brew services start postgresql
```

#### 安装Redis

```bash
# Ubuntu/Debian
sudo apt install redis-server

# CentOS/RHEL
sudo yum install redis
sudo systemctl enable redis
sudo systemctl start redis

# macOS
brew install redis
brew services start redis
```

### 2. 数据库配置

```bash
# 创建数据库和用户
sudo -u postgres psql << EOF
CREATE DATABASE aimonitor;
CREATE USER aimonitor WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE aimonitor TO aimonitor;
\q
EOF
```

### 3. 项目部署

```bash
# 克隆项目
git clone https://github.com/your-org/aimonitor.git
cd aimonitor

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，配置数据库连接等信息

# 安装后端依赖
go mod download

# 运行数据库迁移
go run cmd/migrate/main.go up

# 构建后端
go build -o bin/server cmd/server/main.go

# 安装前端依赖
cd frontend
npm install

# 构建前端
npm run build
cd ..

# 启动服务
./bin/server
```

## 🐳 Docker部署

### 1. Docker Compose配置

#### 开发环境配置

`docker-compose.yml`:

```yaml
version: '3.8'

services:
  # PostgreSQL数据库
  postgres:
    image: postgres:15-alpine
    container_name: aimonitor-postgres
    environment:
      POSTGRES_DB: aimonitor
      POSTGRES_USER: aimonitor
      POSTGRES_PASSWORD: ${DB_PASSWORD:-password}
      TZ: Asia/Shanghai
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    networks:
      - aimonitor-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U aimonitor"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis缓存
  redis:
    image: redis:7-alpine
    container_name: aimonitor-redis
    command: redis-server --requirepass ${REDIS_PASSWORD:-password}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    networks:
      - aimonitor-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # 后端应用
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: aimonitor-app
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=aimonitor
      - DB_PASSWORD=${DB_PASSWORD:-password}
      - DB_NAME=aimonitor
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD:-password}
      - JWT_SECRET=${JWT_SECRET:-your-secret-key}
      - GIN_MODE=release
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - aimonitor-network
    restart: unless-stopped
    volumes:
      - ./logs:/app/logs

  # 前端应用
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: aimonitor-frontend
    ports:
      - "3000:80"
    depends_on:
      - app
    networks:
      - aimonitor-network
    restart: unless-stopped

  # Nginx反向代理
  nginx:
    image: nginx:alpine
    container_name: aimonitor-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - app
      - frontend
    networks:
      - aimonitor-network
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:

networks:
  aimonitor-network:
    driver: bridge
```

### 2. 生产环境配置

`docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: aimonitor
      POSTGRES_USER: aimonitor
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    networks:
      - aimonitor-network
    restart: always
    deploy:
      resources:
        limits:
          memory: 1G
        reservations:
          memory: 512M

  redis:
    image: redis:7-alpine
    command: >
      redis-server
      --requirepass ${REDIS_PASSWORD}
      --maxmemory 256mb
      --maxmemory-policy allkeys-lru
      --save 900 1
      --save 300 10
      --save 60 10000
    volumes:
      - redis_data:/data
    networks:
      - aimonitor-network
    restart: always
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M

  app:
    image: aimonitor:latest
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=aimonitor
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=aimonitor
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - GIN_MODE=release
      - LOG_LEVEL=info
    depends_on:
      - postgres
      - redis
    networks:
      - aimonitor-network
    restart: always
    deploy:
      replicas: 2
      resources:
        limits:
          memory: 1G
        reservations:
          memory: 512M
    volumes:
      - ./logs:/app/logs
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.prod.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
      - ./logs/nginx:/var/log/nginx
    depends_on:
      - app
    networks:
      - aimonitor-network
    restart: always
    deploy:
      resources:
        limits:
          memory: 256M
        reservations:
          memory: 128M

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local

networks:
  aimonitor-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### 3. 环境变量文件

`.env`:

```bash
# 数据库配置
DB_PASSWORD=your_secure_db_password_here

# Redis配置
REDIS_PASSWORD=your_secure_redis_password_here

# JWT配置
JWT_SECRET=your_jwt_secret_key_here_32_characters_long

# Grafana配置
GRAFANA_PASSWORD=your_secure_grafana_password_here

# AI配置
OPENAI_API_KEY=your_openai_api_key_here

# 邮件配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=AI Monitor <your_email@gmail.com>

# Webhook配置
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
DINGTALK_WEBHOOK_URL=https://oapi.dingtalk.com/robot/send?access_token=...
WECHAT_WEBHOOK_URL=https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=...

# 其他配置
TZ=Asia/Shanghai
LOG_LEVEL=info
```

### 4. Docker部署命令

```bash
# 构建并启动所有服务
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app

# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v

# 重启特定服务
docker-compose restart app

# 更新服务
docker-compose pull
docker-compose up -d
```

## ☸️ Kubernetes部署

### 1. 命名空间

`k8s/namespace.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: aimonitor
  labels:
    name: aimonitor
```

### 2. ConfigMap配置

`k8s/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: aimonitor-config
  namespace: aimonitor
data:
  config.yaml: |
    server:
      port: 8080
      mode: release
    database:
      host: postgres-service
      port: "5432"
      name: aimonitor
      user: aimonitor
      sslmode: disable
      maxIdleConns: 10
      maxOpenConns: 100
    redis:
      host: redis-service
      port: "6379"
      db: 0
    log:
      level: info
      format: json
    monitoring:
      enabled: true
      prometheus:
        enabled: true
        port: 9090
```

### 3. Secret配置

`k8s/secret.yaml`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aimonitor-secret
  namespace: aimonitor
type: Opaque
data:
  db-password: <base64-encoded-password>
  redis-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-jwt-secret>
  openai-api-key: <base64-encoded-api-key>
```

### 4. 应用部署

`k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aimonitor-app
  namespace: aimonitor
  labels:
    app: aimonitor
    component: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aimonitor
      component: backend
  template:
    metadata:
      labels:
        app: aimonitor
        component: backend
    spec:
      containers:
      - name: aimonitor
        image: aimonitor:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: postgres-service
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: aimonitor-secret
              key: db-password
        - name: REDIS_HOST
          value: redis-service
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: aimonitor-secret
              key: jwt-secret
        volumeMounts:
        - name: config
          mountPath: /app/config
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: aimonitor-config
```

### 5. Service配置

`k8s/service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: aimonitor-service
  namespace: aimonitor
spec:
  selector:
    app: aimonitor
    component: backend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

### 6. Ingress配置

`k8s/ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: aimonitor-ingress
  namespace: aimonitor
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - monitor.yourdomain.com
    secretName: aimonitor-tls
  rules:
  - host: monitor.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: aimonitor-service
            port:
              number: 80
```

### 7. Kubernetes部署命令

```bash
# 创建命名空间
kubectl apply -f k8s/namespace.yaml

# 创建配置和密钥
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# 部署数据库
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml

# 部署监控组件
kubectl apply -f k8s/monitoring.yaml

# 部署应用
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# 配置Ingress
kubectl apply -f k8s/ingress.yaml

# 查看部署状态
kubectl get pods -n aimonitor
kubectl get services -n aimonitor
kubectl get ingress -n aimonitor

# 查看日志
kubectl logs -f deployment/aimonitor-app -n aimonitor

# 扩容
kubectl scale deployment aimonitor-app --replicas=5 -n aimonitor

# 滚动更新
kubectl set image deployment/aimonitor-app aimonitor=aimonitor:v1.1.0 -n aimonitor

# 回滚
kubectl rollout undo deployment/aimonitor-app -n aimonitor
```

## ⚙️ 配置管理

### 1. 配置文件结构

```
config/
├── config.yaml          # 主配置文件
├── config.dev.yaml      # 开发环境配置
├── config.test.yaml     # 测试环境配置
├── config.prod.yaml     # 生产环境配置
└── secrets/
    ├── jwt.key          # JWT私钥
    ├── tls.crt          # TLS证书
    └── tls.key          # TLS私钥
```

### 2. 环境变量优先级

1. 命令行参数
2. 环境变量
3. 配置文件
4. 默认值

### 3. 主配置文件示例

`config/config.yaml`:

```yaml
server:
  port: 8080
  mode: debug
  readTimeout: 60s
  writeTimeout: 60s
  maxHeaderBytes: 1048576

database:
  host: localhost
  port: 5432
  user: aimonitor
  password: password
  name: aimonitor
  sslmode: disable
  maxIdleConns: 10
  maxOpenConns: 100
  connMaxLifetime: 3600s

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  poolSize: 10
  minIdleConns: 5

jwt:
  secret: your-secret-key
  expireTime: 24h
  refreshTime: 168h

log:
  level: info
  format: json
  output: stdout
  file:
    enabled: false
    path: logs/app.log
    maxSize: 100
    maxBackups: 5
    maxAge: 30

monitoring:
  enabled: true
  prometheus:
    enabled: true
    port: 9090
    path: /metrics
  jaeger:
    enabled: true
    endpoint: http://localhost:14268/api/traces

ai:
  openai:
    apiKey: your-api-key
    model: gpt-4
    maxTokens: 4000
  claude:
    apiKey: your-api-key
    model: claude-3-sonnet

notification:
  email:
    enabled: true
    smtp:
      host: smtp.gmail.com
      port: 587
      username: your-email@gmail.com
      password: your-app-password
      from: AI Monitor <your-email@gmail.com>
  webhook:
    slack:
      enabled: false
      url: https://hooks.slack.com/services/...
    dingtalk:
      enabled: false
      url: https://oapi.dingtalk.com/robot/send?access_token=...
```

## 📊 监控与运维

### 1. Prometheus监控配置

`monitoring/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'aimonitor'
    static_configs:
      - targets: ['app:8080']
    metrics_path: /metrics
    scrape_interval: 10s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  - job_name: 'node'
    static_configs:
      - targets: ['node-exporter:9100']
```

### 2. Grafana仪表板

系统提供预配置的Grafana仪表板：

- **系统概览**: 整体系统健康状态
- **应用性能**: API响应时间、吞吐量
- **数据库监控**: PostgreSQL性能指标
- **缓存监控**: Redis使用情况
- **基础设施**: 服务器资源使用

### 3. 日志管理

使用ELK Stack进行日志收集和分析：

```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.8.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    ports:
      - "9200:9200"

  logstash:
    image: docker.elastic.co/logstash/logstash:8.8.0
    volumes:
      - ./logstash/pipeline:/usr/share/logstash/pipeline
    ports:
      - "5044:5044"
    depends_on:
      - elasticsearch

  kibana:
    image: docker.elastic.co/kibana/kibana:8.8.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch

volumes:
  elasticsearch_data:
```

## 🔒 安全配置

### 1. SSL/TLS配置

#### 生成自签名证书（开发环境）

```bash
# 创建SSL目录
mkdir -p nginx/ssl

# 生成私钥
openssl genrsa -out nginx/ssl/server.key 2048

# 生成证书签名请求
openssl req -new -key nginx/ssl/server.key -out nginx/ssl/server.csr

# 生成自签名证书
openssl x509 -req -days 365 -in nginx/ssl/server.csr -signkey nginx/ssl/server.key -out nginx/ssl/server.crt
```

#### Let's Encrypt证书（生产环境）

```bash
# 安装certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d monitor.yourdomain.com

# 自动续期
sudo crontab -e
# 添加以下行
0 12 * * * /usr/bin/certbot renew --quiet
```

### 2. 防火墙配置

```bash
# Ubuntu/Debian
sudo ufw enable
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8080/tcp

# CentOS/RHEL
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

### 3. 数据库安全

```sql
-- 创建只读用户
CREATE USER aimonitor_readonly WITH PASSWORD 'readonly_password';
GRANT CONNECT ON DATABASE aimonitor TO aimonitor_readonly;
GRANT USAGE ON SCHEMA public TO aimonitor_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO aimonitor_readonly;

-- 设置连接限制
ALTER USER aimonitor CONNECTION LIMIT 50;

-- 启用行级安全
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
```

## 🚨 故障排除

### 常见问题

#### 1. 服务无法启动

**问题**: Docker容器启动失败

**解决方案**:
```bash
# 查看容器日志
docker-compose logs app

# 检查端口占用
sudo netstat -tlnp | grep :8080

# 检查磁盘空间
df -h

# 清理Docker资源
docker system prune -a
```

#### 2. 数据库连接失败

**问题**: 应用无法连接到PostgreSQL

**解决方案**:
```bash
# 检查数据库服务状态
docker-compose ps postgres

# 测试数据库连接
psql -h localhost -U aimonitor -d aimonitor

# 检查网络连接
docker network ls
docker network inspect aimonitor_default
```

#### 3. 前端页面无法访问

**问题**: 浏览器显示502错误

**解决方案**:
```bash
# 检查Nginx配置
nginx -t

# 查看Nginx日志
docker-compose logs nginx

# 检查后端API状态
curl http://localhost:8080/health
```

#### 4. 性能问题

**问题**: 系统响应缓慢

**解决方案**:
```bash
# 检查系统资源
top
htop
iostat -x 1

# 检查数据库性能
# 在PostgreSQL中执行
SELECT * FROM pg_stat_activity WHERE state = 'active';

# 检查Redis性能
redis-cli info stats
```

### 日志分析

#### 应用日志位置
- 应用日志: `./logs/app.log`
- Nginx日志: `./logs/nginx/`
- PostgreSQL日志: Docker容器内
- Redis日志: Docker容器内

#### 常用日志命令
```bash
# 实时查看应用日志
tail -f logs/app.log

# 查看错误日志
grep "ERROR" logs/app.log

# 查看最近的访问日志
tail -100 logs/nginx/access.log

# 分析错误模式
awk '/ERROR/ {print $1, $2, $NF}' logs/app.log | sort | uniq -c
```

### 性能优化

#### 数据库优化
```sql
-- 分析慢查询
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- 创建索引
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
CREATE INDEX CONCURRENTLY idx_logs_created_at ON logs(created_at);

-- 分析表统计信息
ANALYZE;
```

#### 应用优化
```bash
# 调整Go运行时参数
export GOGC=100
export GOMAXPROCS=4

# 调整数据库连接池
export DB_MAX_OPEN_CONNS=50
export DB_MAX_IDLE_CONNS=25

# 启用Redis持久化
export REDIS_SAVE="900 1 300 10 60 10000"
```

## 📚 相关文档

- [用户手册](USER_MANUAL.md) - 系统使用指南
- [开发指南](DEVELOPMENT_GUIDE.md) - 开发环境搭建
- [API文档](API_REFERENCE.md) - 接口说明
- [配置参考](CONFIGURATION_REFERENCE.md) - 详细配置说明
- [监控指南](MONITORING_GUIDE.md) - 监控配置
- [安全指南](SECURITY_GUIDE.md) - 安全最佳实践

## 🆘 技术支持

如果您在部署过程中遇到问题，可以通过以下方式获取帮助：

1. **查看文档**: 首先查阅相关文档和FAQ
2. **检查日志**: 查看应用和系统日志获取错误信息
3. **社区支持**: 在GitHub Issues中提交问题
4. **技术支持**: 联系技术支持团队

---

**恭喜！您已经成功部署了 AI Monitor 智能监控系统！** 🎉

现在您可以开始使用系统的各项功能，包括实时监控、智能告警、性能分析等。建议您先阅读用户手册了解系统的基本功能，然后根据实际需求进行个性化配置。