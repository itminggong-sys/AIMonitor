#!/bin/bash

# =============================================================================
# AI Monitor 镜像部署脚本
# Version: 2.0.0
# Description: 分离式部署 - 第二步：从本地镜像部署服务
# Author: AI Assistant
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取当前时间戳
get_timestamp() {
    date '+%Y-%m-%d %H:%M:%S'
}

# 日志函数
log_info() {
    echo -e "[$(get_timestamp)] ${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "[$(get_timestamp)] ${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "[$(get_timestamp)] ${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "[$(get_timestamp)] ${RED}❌ $1${NC}"
}

# 显示欢迎信息
show_welcome() {
    echo -e "${GREEN}==========================================${NC}"
    echo -e "${GREEN}    AI Monitor 镜像部署工具 v2.0${NC}"
    echo -e "${GREEN}==========================================${NC}"
    echo ""
    echo "本工具将从预构建的Docker镜像部署AI Monitor系统"
    echo "确保已经运行了 load_images.sh 加载所有镜像"
    echo ""
}

# 检查Docker环境
check_docker() {
    log_info "检查Docker环境"
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装Docker"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker 服务未运行，请启动Docker服务"
        exit 1
    fi
    
    # 检查docker-compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "docker-compose 未安装，请先安装docker-compose"
        exit 1
    fi
    
    log_success "Docker环境检查通过"
}

# 检查必需的镜像
check_required_images() {
    log_info "检查必需的Docker镜像"
    
    REQUIRED_IMAGES=(
        "ai-monitor-backend:latest"
        "ai-monitor-frontend:latest"
        "postgres:15-alpine"
        "redis:7-alpine"
        "prom/prometheus:latest"
        "docker.elastic.co/elasticsearch/elasticsearch:8.11.0"
    )
    
    MISSING_IMAGES=()
    
    for image in "${REQUIRED_IMAGES[@]}"; do
        if ! docker images --format "table {{.Repository}}:{{.Tag}}" | grep -q "^$image$"; then
            MISSING_IMAGES+=("$image")
        fi
    done
    
    if [ ${#MISSING_IMAGES[@]} -gt 0 ]; then
        log_error "缺少以下Docker镜像:"
        for image in "${MISSING_IMAGES[@]}"; do
            echo "  - $image"
        done
        echo ""
        log_error "请先运行 load_images.sh 加载所有镜像"
        exit 1
    fi
    
    log_success "所有必需镜像检查通过"
}

# 创建必要的目录和文件
create_deployment_files() {
    log_info "创建部署配置文件"
    
    # 创建configs目录
    mkdir -p configs
    
    # 创建config.yaml
    if [ ! -f "configs/config.yaml" ]; then
        cat > configs/config.yaml << 'EOF'
server:
  port: 8080
  mode: release

database:
  host: postgres
  port: 5432
  user: ai_monitor
  password: password
  dbname: ai_monitor
  sslmode: disable

redis:
  host: redis
  port: 6379
  password: ""
  db: 0

logging:
  level: info
  file: logs/app.log

metrics:
  enabled: true
  path: /metrics
EOF
    fi
    
    # 创建prometheus配置
    if [ ! -f "configs/prometheus.yml" ]; then
        cat > configs/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'ai-monitor'
    static_configs:
      - targets: ['aimonitor:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF
    fi
    
    # 创建数据库初始化脚本
    if [ ! -f "init.sql" ]; then
        cat > init.sql << 'EOF'
-- AI Monitor 数据库初始化脚本

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建监控数据表
CREATE TABLE IF NOT EXISTS monitor_data (
    id SERIAL PRIMARY KEY,
    agent_id VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(10,2),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tags JSONB
);

-- 创建告警规则表
CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    condition TEXT NOT NULL,
    threshold DECIMAL(10,2),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建告警历史表
CREATE TABLE IF NOT EXISTS alert_history (
    id SERIAL PRIMARY KEY,
    rule_id INTEGER REFERENCES alert_rules(id),
    message TEXT,
    level VARCHAR(20),
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

-- 插入默认管理员用户
INSERT INTO users (username, password, email, role) 
VALUES ('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin@example.com', 'admin')
ON CONFLICT (username) DO NOTHING;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_monitor_data_timestamp ON monitor_data(timestamp);
CREATE INDEX IF NOT EXISTS idx_monitor_data_agent_id ON monitor_data(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_history_triggered_at ON alert_history(triggered_at);
EOF
    fi
    
    # 创建Redis配置
    mkdir -p deploy
    if [ ! -f "deploy/redis.conf" ]; then
        cat > deploy/redis.conf << 'EOF'
# Redis配置文件
bind 0.0.0.0
port 6379
timeout 0
tcp-keepalive 300

# 内存管理
maxmemory 256mb
maxmemory-policy allkeys-lru

# 持久化
save 900 1
save 300 10
save 60 10000

# 日志
loglevel notice
logfile ""

# 安全
# requirepass your_password_here
EOF
    fi
    
    log_success "部署配置文件创建完成"
}

# 创建docker-compose配置
create_docker_compose() {
    log_info "创建docker-compose配置文件"
    
    cat > docker-compose.production.yml << 'EOF'
services:
  postgres:
    image: postgres:15-alpine
    container_name: ai-monitor-postgres
    environment:
      POSTGRES_DB: ai_monitor
      POSTGRES_USER: ai_monitor
      POSTGRES_PASSWORD: password
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ai_monitor -d ai_monitor"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 30s
    networks:
      - ai-monitor-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  redis:
    image: redis:7-alpine
    container_name: ai-monitor-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      - ./deploy/redis.conf:/usr/local/etc/redis/redis.conf
    command: redis-server /usr/local/etc/redis/redis.conf
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 5
    networks:
      - ai-monitor-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  aimonitor:
    image: ai-monitor-backend:latest
    container_name: ai-monitor-backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - GIN_MODE=release
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 60s
    networks:
      - ai-monitor-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  frontend:
    image: ai-monitor-frontend:latest
    container_name: ai-monitor-frontend
    ports:
      - "3000:80"
    depends_on:
      aimonitor:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:80"]
      interval: 30s
      timeout: 10s
      retries: 5
    networks:
      - ai-monitor-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  prometheus:
    image: prom/prometheus:latest
    container_name: ai-monitor-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    restart: unless-stopped
    networks:
      - ai-monitor-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    container_name: ai-monitor-elasticsearch
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
      - xpack.security.enabled=false
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:9200/_cluster/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 5
    networks:
      - ai-monitor-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local
  prometheus_data:
    driver: local
  elasticsearch_data:
    driver: local

networks:
  ai-monitor-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
EOF
    
    log_success "docker-compose配置文件创建完成"
}

# 停止旧服务
stop_old_services() {
    log_info "停止旧服务"
    
    # 尝试停止可能存在的服务
    docker-compose -f docker-compose.production.yml down --remove-orphans 2>/dev/null || true
    docker-compose down --remove-orphans 2>/dev/null || true
    
    # 清理悬挂的容器
    docker container prune -f 2>/dev/null || true
    
    log_success "旧服务停止完成"
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录"
    
    mkdir -p logs
    mkdir -p data
    
    # 设置目录权限
    chmod 755 logs
    chmod 755 data
    
    log_success "目录创建完成"
}

# 启动服务
start_services() {
    log_info "启动AI Monitor服务"
    
    # 启动所有服务
    docker-compose -f docker-compose.production.yml up -d
    
    if [ $? -eq 0 ]; then
        log_success "服务启动命令执行成功"
    else
        log_error "服务启动失败"
        exit 1
    fi
}

# 等待服务启动
wait_for_services() {
    log_info "等待服务完全启动"
    
    # 等待基础服务启动
    sleep 30
    
    # 检查服务状态
    log_info "检查服务状态"
    docker-compose -f docker-compose.production.yml ps
    
    log_success "服务状态检查完成"
}

# 健康检查
health_check() {
    log_info "执行服务健康检查"
    
    MAX_ATTEMPTS=15
    ATTEMPT=1
    
    # 检查后端服务
    while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
        log_info "健康检查尝试 $ATTEMPT/$MAX_ATTEMPTS"
        
        if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
            log_success "后端服务健康检查通过"
            break
        else
            if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
                log_error "后端服务健康检查失败"
                log_info "查看后端服务日志:"
                docker-compose -f docker-compose.production.yml logs --tail=50 aimonitor
                return 1
            fi
            log_warning "后端服务尚未就绪，等待15秒后重试..."
            sleep 15
            ATTEMPT=$((ATTEMPT + 1))
        fi
    done
    
    # 检查前端服务
    if curl -f -s http://localhost:3000 > /dev/null 2>&1; then
        log_success "前端服务健康检查通过"
    else
        log_warning "前端服务可能尚未完全就绪"
    fi
    
    # 检查数据库连接
    if docker-compose -f docker-compose.production.yml exec -T postgres pg_isready -U ai_monitor -d ai_monitor > /dev/null 2>&1; then
        log_success "数据库连接检查通过"
    else
        log_warning "数据库连接检查失败"
    fi
    
    log_success "健康检查完成"
}

# 显示部署结果
show_deployment_result() {
    echo ""
    echo -e "${GREEN}🎉 AI Monitor 部署完成！${NC}"
    echo "=========================================="
    echo -e "${BLUE}📋 服务访问地址:${NC}"
    
    # 获取服务器IP
    SERVER_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "localhost")
    
    echo "  前端界面: http://$SERVER_IP:3000"
    echo "  后端API: http://$SERVER_IP:8080"
    echo "  API文档: http://$SERVER_IP:8080/swagger/index.html"
    echo "  健康检查: http://$SERVER_IP:8080/health"
    echo "  Prometheus: http://$SERVER_IP:9090"
    echo "  Elasticsearch: http://$SERVER_IP:9200"
    echo ""
    echo -e "${YELLOW}🔧 管理命令:${NC}"
    echo "  查看服务状态: docker-compose -f docker-compose.production.yml ps"
    echo "  查看所有日志: docker-compose -f docker-compose.production.yml logs"
    echo "  查看后端日志: docker-compose -f docker-compose.production.yml logs aimonitor"
    echo "  查看前端日志: docker-compose -f docker-compose.production.yml logs frontend"
    echo "  重启服务: docker-compose -f docker-compose.production.yml restart"
    echo "  停止服务: docker-compose -f docker-compose.production.yml down"
    echo "  更新服务: docker-compose -f docker-compose.production.yml up -d"
    echo ""
    echo -e "${GREEN}📊 默认登录信息:${NC}"
    echo "  用户名: admin"
    echo "  密码: password"
    echo ""
    echo -e "${GREEN}✅ 部署成功！系统已准备就绪${NC}"
    echo ""
}

# 创建管理脚本
create_management_scripts() {
    log_info "创建管理脚本"
    
    # 创建服务管理脚本
    cat > manage.sh << 'EOF'
#!/bin/bash

# AI Monitor 服务管理脚本

COMPOSE_FILE="docker-compose.production.yml"

case "$1" in
    start)
        echo "启动AI Monitor服务..."
        docker-compose -f $COMPOSE_FILE up -d
        ;;
    stop)
        echo "停止AI Monitor服务..."
        docker-compose -f $COMPOSE_FILE down
        ;;
    restart)
        echo "重启AI Monitor服务..."
        docker-compose -f $COMPOSE_FILE restart
        ;;
    status)
        echo "AI Monitor服务状态:"
        docker-compose -f $COMPOSE_FILE ps
        ;;
    logs)
        if [ -n "$2" ]; then
            docker-compose -f $COMPOSE_FILE logs -f "$2"
        else
            docker-compose -f $COMPOSE_FILE logs -f
        fi
        ;;
    update)
        echo "更新AI Monitor服务..."
        docker-compose -f $COMPOSE_FILE pull
        docker-compose -f $COMPOSE_FILE up -d
        ;;
    backup)
        echo "备份数据库..."
        docker-compose -f $COMPOSE_FILE exec postgres pg_dump -U ai_monitor ai_monitor > "backup_$(date +%Y%m%d_%H%M%S).sql"
        echo "备份完成"
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|logs [service]|update|backup}"
        echo "示例:"
        echo "  $0 start          # 启动所有服务"
        echo "  $0 stop           # 停止所有服务"
        echo "  $0 restart        # 重启所有服务"
        echo "  $0 status         # 查看服务状态"
        echo "  $0 logs           # 查看所有日志"
        echo "  $0 logs aimonitor # 查看后端日志"
        echo "  $0 update         # 更新服务"
        echo "  $0 backup         # 备份数据库"
        exit 1
        ;;
esac
EOF
    
    chmod +x manage.sh
    
    # 创建监控脚本
    cat > monitor.sh << 'EOF'
#!/bin/bash

# AI Monitor 系统监控脚本

echo "=== AI Monitor 系统状态 ==="
echo "时间: $(date)"
echo ""

echo "=== Docker 服务状态 ==="
docker-compose -f docker-compose.production.yml ps
echo ""

echo "=== 系统资源使用情况 ==="
echo "CPU使用率:"
top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1
echo ""
echo "内存使用情况:"
free -h
echo ""
echo "磁盘使用情况:"
df -h
echo ""

echo "=== Docker 容器资源使用 ==="
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
echo ""

echo "=== 服务健康检查 ==="
echo -n "后端服务: "
if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ 正常"
else
    echo "❌ 异常"
fi

echo -n "前端服务: "
if curl -f -s http://localhost:3000 > /dev/null 2>&1; then
    echo "✅ 正常"
else
    echo "❌ 异常"
fi

echo -n "Prometheus: "
if curl -f -s http://localhost:9090 > /dev/null 2>&1; then
    echo "✅ 正常"
else
    echo "❌ 异常"
fi

echo -n "Elasticsearch: "
if curl -f -s http://localhost:9200 > /dev/null 2>&1; then
    echo "✅ 正常"
else
    echo "❌ 异常"
fi
EOF
    
    chmod +x monitor.sh
    
    log_success "管理脚本创建完成"
}

# 主函数
main() {
    show_welcome
    check_docker
    check_required_images
    create_deployment_files
    create_docker_compose
    stop_old_services
    create_directories
    start_services
    wait_for_services
    health_check
    create_management_scripts
    show_deployment_result
}

# 错误处理
trap 'log_error "部署过程中发生错误，请检查上述日志信息"' ERR

# 执行主函数
main