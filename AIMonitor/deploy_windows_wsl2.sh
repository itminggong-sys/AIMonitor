#!/bin/bash

# AI Monitor Windows WSL2 + Docker Desktop 部署脚本
# 适用于: Windows 10/11 + WSL2 + Docker Desktop
# 作者: AI Monitor Team
# 版本: v1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示欢迎信息
show_welcome() {
    clear
    echo -e "${GREEN}"
    echo "================================================="
    echo "    AI Monitor Windows WSL2 部署脚本 v1.0.0"
    echo "================================================="
    echo -e "${NC}"
    echo "适用环境:"
    echo "  ✓ Windows 10/11 + WSL2"
    echo "  ✓ Docker Desktop for Windows"
    echo "  ✓ 内存: 4GB+ (推荐8GB+)"
    echo "  ✓ 存储: 10GB+ 可用空间"
    echo ""
    echo "部署内容:"
    echo "  ✓ 后端服务 (Go + Gin) - 端口8080"
    echo "  ✓ 前端服务 (React + Vite) - 端口3000"
    echo "  ✓ PostgreSQL 数据库 - 端口5432"
    echo "  ✓ Redis 缓存 - 端口6379"
    echo "  ✓ Prometheus 监控 - 端口9090"
    echo "  ✓ Elasticsearch 搜索 - 端口9200"
    echo ""
    read -p "按回车键开始部署..." -r
}

# 检查WSL2环境
check_wsl2_environment() {
    log_info "检查WSL2环境..."
    
    # 检查是否在WSL环境中
    if ! grep -q "microsoft" /proc/version 2>/dev/null; then
        log_error "此脚本需要在WSL2环境中运行"
        log_info "请在Windows中打开WSL2终端后运行此脚本"
        exit 1
    fi
    
    log_success "WSL2环境检查通过"
    
    # 显示WSL信息
    if command -v lsb_release &> /dev/null; then
        WSL_DISTRO=$(lsb_release -d | cut -f2)
        log_info "WSL发行版: $WSL_DISTRO"
    fi
}

# 检查Docker Desktop
check_docker_desktop() {
    log_info "检查Docker Desktop状态..."
    
    # 检查docker命令是否可用
    if ! command -v docker &> /dev/null; then
        log_error "未找到Docker命令"
        log_info "请确保Docker Desktop已安装并启用WSL2集成"
        log_info "安装步骤:"
        echo "  1. 下载Docker Desktop: https://www.docker.com/products/docker-desktop"
        echo "  2. 安装并启动Docker Desktop"
        echo "  3. 在设置中启用WSL2集成"
        echo "  4. 重启WSL2终端"
        exit 1
    fi
    
    # 检查docker服务是否运行
    if ! docker ps &> /dev/null; then
        log_error "Docker服务未运行"
        log_info "请启动Docker Desktop应用程序"
        log_info "等待Docker Desktop启动后重新运行此脚本"
        exit 1
    fi
    
    # 检查docker-compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "未找到docker-compose命令"
        log_info "请确保Docker Desktop版本支持docker-compose"
        exit 1
    fi
    
    log_success "Docker Desktop检查通过"
    
    # 显示Docker信息
    DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | cut -d',' -f1)
    log_info "Docker版本: $DOCKER_VERSION"
}

# 检查系统资源
check_system_resources() {
    log_info "检查系统资源..."
    
    # 检查内存
    TOTAL_MEM=$(free -m | awk 'NR==2{printf "%.0f", $2/1024}')
    if [ "$TOTAL_MEM" -lt 4 ]; then
        log_warning "系统内存不足4GB，可能影响性能"
    else
        log_success "内存检查通过: ${TOTAL_MEM}GB"
    fi
    
    # 检查磁盘空间
    AVAILABLE_SPACE=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
    if [ "$AVAILABLE_SPACE" -lt 10 ]; then
        log_warning "可用磁盘空间不足10GB，可能影响部署"
    else
        log_success "磁盘空间检查通过: ${AVAILABLE_SPACE}GB可用"
    fi
}

# 检查端口占用
check_ports() {
    log_info "检查端口占用情况..."
    
    PORTS=("3000" "5432" "6379" "8080" "9090" "9200" "9300")
    OCCUPIED_PORTS=()
    
    for port in "${PORTS[@]}"; do
        if netstat -tuln 2>/dev/null | grep -q ":$port " || ss -tuln 2>/dev/null | grep -q ":$port "; then
            OCCUPIED_PORTS+=("$port")
        fi
    done
    
    if [ ${#OCCUPIED_PORTS[@]} -gt 0 ]; then
        log_warning "以下端口已被占用: ${OCCUPIED_PORTS[*]}"
        log_info "如果继续部署，这些服务可能无法启动"
        read -p "是否继续部署? (y/N): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "部署已取消"
            exit 0
        fi
    else
        log_success "所有必需端口都可用"
    fi
}

# 准备项目环境
prepare_project() {
    log_info "准备项目环境..."
    
    # 检查项目文件
    if [ ! -f "docker-compose.yml" ]; then
        log_error "未找到docker-compose.yml文件"
        log_info "请确保在项目根目录中运行此脚本"
        exit 1
    fi
    
    if [ ! -f "go.mod" ]; then
        log_error "未找到go.mod文件"
        log_info "请确保在项目根目录中运行此脚本"
        exit 1
    fi
    
    if [ ! -d "web" ]; then
        log_error "未找到web目录"
        log_info "请确保在项目根目录中运行此脚本"
        exit 1
    fi
    
    log_success "项目文件检查通过"
    
    # 创建必要的目录
    mkdir -p logs
    
    # 设置环境变量文件
    if [ ! -f ".env" ]; then
        log_info "创建环境变量文件..."
        cat > .env << 'EOF'
# AI Monitor 环境配置
DB_HOST=postgres
DB_PORT=5432
DB_NAME=ai_monitor
DB_USER=ai_monitor
DB_PASSWORD=password

REDIS_HOST=redis
REDIS_PORT=6379

GIN_MODE=release
LOG_LEVEL=info

# 前端配置
VITE_API_BASE_URL=http://localhost:8080
NODE_ENV=development
EOF
        log_success "环境变量文件已创建"
    fi
}

# 清理旧容器和镜像
cleanup_old_containers() {
    log_info "清理旧的容器和镜像..."
    
    # 停止并删除旧容器
    if docker-compose ps -q 2>/dev/null | grep -q .; then
        log_info "停止现有服务..."
        docker-compose down --remove-orphans 2>/dev/null || true
    fi
    
    # 清理未使用的镜像和容器
    log_info "清理Docker缓存..."
    docker system prune -f --volumes 2>/dev/null || true
    
    log_success "清理完成"
}

# 构建和启动服务
deploy_services() {
    log_info "开始构建和部署服务..."
    
    # 拉取基础镜像
    log_info "拉取Docker基础镜像..."
    docker pull postgres:15-alpine
    docker pull redis:7-alpine
    docker pull golang:1.21-alpine
    docker pull node:18-alpine
    docker pull prom/prometheus:latest
    docker pull docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    
    # 构建并启动服务
    log_info "构建并启动所有服务..."
    
    # 使用docker-compose构建和启动
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d --build
    else
        docker compose up -d --build
    fi
    
    log_success "服务部署完成"
}

# 等待服务启动
wait_for_services() {
    log_info "等待服务启动..."
    
    # 等待数据库启动
    log_info "等待PostgreSQL启动..."
    for i in {1..30}; do
        if docker exec ai-monitor-postgres pg_isready -U ai_monitor -d ai_monitor &>/dev/null; then
            log_success "PostgreSQL已启动"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "PostgreSQL启动超时"
            return 1
        fi
        sleep 2
    done
    
    # 等待Redis启动
    log_info "等待Redis启动..."
    for i in {1..30}; do
        if docker exec ai-monitor-redis redis-cli ping &>/dev/null; then
            log_success "Redis已启动"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "Redis启动超时"
            return 1
        fi
        sleep 2
    done
    
    # 等待后端服务启动
    log_info "等待后端服务启动..."
    for i in {1..60}; do
        if curl -s http://localhost:8080/health &>/dev/null; then
            log_success "后端服务已启动"
            break
        fi
        if [ $i -eq 60 ]; then
            log_warning "后端服务启动超时，请检查日志"
            break
        fi
        sleep 3
    done
    
    # 等待前端服务启动
    log_info "等待前端服务启动..."
    for i in {1..60}; do
        if curl -s http://localhost:3000 &>/dev/null; then
            log_success "前端服务已启动"
            break
        fi
        if [ $i -eq 60 ]; then
            log_warning "前端服务启动超时，请检查日志"
            break
        fi
        sleep 3
    done
}

# 显示服务状态
show_service_status() {
    log_info "检查服务状态..."
    
    echo ""
    echo "=== 服务状态 ==="
    if command -v docker-compose &> /dev/null; then
        docker-compose ps
    else
        docker compose ps
    fi
    
    echo ""
    echo "=== 端口检查 ==="
    SERVICES=(
        "前端服务:3000"
        "后端API:8080"
        "PostgreSQL:5432"
        "Redis:6379"
        "Prometheus:9090"
        "Elasticsearch:9200"
    )
    
    for service in "${SERVICES[@]}"; do
        name=$(echo $service | cut -d':' -f1)
        port=$(echo $service | cut -d':' -f2)
        if netstat -tuln 2>/dev/null | grep -q ":$port " || ss -tuln 2>/dev/null | grep -q ":$port "; then
            echo -e "  ✓ $name (端口$port): ${GREEN}运行中${NC}"
        else
            echo -e "  ✗ $name (端口$port): ${RED}未运行${NC}"
        fi
    done
}

# 显示访问信息
show_access_info() {
    echo ""
    echo -e "${GREEN}=================================================${NC}"
    echo -e "${GREEN}           AI Monitor 部署成功！${NC}"
    echo -e "${GREEN}=================================================${NC}"
    echo ""
    echo "🌐 访问地址:"
    echo "  • 前端界面: http://localhost:3000"
    echo "  • 后端API:  http://localhost:8080"
    echo "  • API文档:  http://localhost:8080/swagger/index.html"
    echo "  • 健康检查: http://localhost:8080/health"
    echo ""
    echo "📊 监控服务:"
    echo "  • Prometheus: http://localhost:9090"
    echo "  • Elasticsearch: http://localhost:9200"
    echo ""
    echo "🗄️ 数据库连接:"
    echo "  • PostgreSQL: localhost:5432"
    echo "    - 数据库: ai_monitor"
    echo "    - 用户名: ai_monitor"
    echo "    - 密码: password"
    echo "  • Redis: localhost:6379"
    echo ""
    echo "🔧 管理命令:"
    echo "  • 查看日志: docker-compose logs -f [服务名]"
    echo "  • 重启服务: docker-compose restart [服务名]"
    echo "  • 停止服务: docker-compose down"
    echo "  • 启动服务: docker-compose up -d"
    echo ""
    echo "📝 默认账户 (如果有登录页面):"
    echo "  • 用户名: admin"
    echo "  • 密码: admin123"
    echo ""
    echo -e "${YELLOW}⚠️  重要提醒:${NC}"
    echo "  1. 首次访问可能需要等待1-2分钟"
    echo "  2. 如果服务无法访问，请检查防火墙设置"
    echo "  3. 生产环境请修改默认密码"
    echo ""
}

# 显示故障排除信息
show_troubleshooting() {
    echo "🔍 故障排除:"
    echo ""
    echo "如果遇到问题，请尝试以下步骤:"
    echo ""
    echo "1. 检查服务日志:"
    echo "   docker-compose logs [服务名]"
    echo ""
    echo "2. 重启所有服务:"
    echo "   docker-compose down && docker-compose up -d"
    echo ""
    echo "3. 清理并重新部署:"
    echo "   docker-compose down -v"
    echo "   docker system prune -f"
    echo "   ./deploy_windows_wsl2.sh"
    echo ""
    echo "4. 检查Docker Desktop状态:"
    echo "   确保Docker Desktop正在运行"
    echo ""
    echo "5. 检查WSL2资源分配:"
    echo "   在.wslconfig中增加内存和CPU分配"
    echo ""
}

# 主函数
main() {
    # 显示欢迎信息
    show_welcome
    
    # 环境检查
    check_wsl2_environment
    check_docker_desktop
    check_system_resources
    check_ports
    
    # 项目准备
    prepare_project
    
    # 清理旧环境
    cleanup_old_containers
    
    # 部署服务
    deploy_services
    
    # 等待服务启动
    wait_for_services
    
    # 显示结果
    show_service_status
    show_access_info
    show_troubleshooting
    
    log_success "部署完成！请访问 http://localhost:3000 查看前端界面"
}

# 错误处理
trap 'log_error "部署过程中发生错误，请检查上面的错误信息"' ERR

# 运行主函数
main "$@"