#!/bin/bash

# =============================================================================
# AI Monitor 镜像构建脚本
# Version: 2.0.0
# Description: 分离式部署 - 第一步：构建前后端Docker镜像
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
    echo -e "${GREEN}    AI Monitor 镜像构建工具 v2.0${NC}"
    echo -e "${GREEN}==========================================${NC}"
    echo ""
    echo "本工具将构建前后端Docker镜像并保存到本地目录"
    echo "构建完成后可以传输到远程服务器进行部署"
    echo ""
}

# 检查项目结构
check_project_structure() {
    log_info "检查项目结构"
    
    if [ ! -f "go.mod" ]; then
        log_error "go.mod 文件不存在，请在项目根目录运行此脚本"
        exit 1
    fi
    
    if [ ! -f "cmd/server/main.go" ]; then
        log_error "cmd/server/main.go 文件不存在，请确保项目结构完整"
        exit 1
    fi
    
    if [ ! -d "web" ]; then
        log_error "web 目录不存在，请确保前端项目存在"
        exit 1
    fi
    
    if [ ! -f "web/package.json" ]; then
        log_error "web/package.json 文件不存在，请确保前端项目完整"
        exit 1
    fi
    
    log_success "项目结构检查通过"
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
    
    log_success "Docker环境检查通过"
}

# 创建镜像输出目录
create_output_directory() {
    log_info "创建镜像输出目录"
    
    OUTPUT_DIR="./docker-images"
    
    if [ -d "$OUTPUT_DIR" ]; then
        log_warning "输出目录已存在，清理旧文件"
        rm -rf "$OUTPUT_DIR"/*
    else
        mkdir -p "$OUTPUT_DIR"
    fi
    
    log_success "输出目录创建完成: $OUTPUT_DIR"
}

# 清理Docker缓存
clean_docker_cache() {
    log_info "清理Docker构建缓存"
    
    # 停止相关容器
    docker-compose down --remove-orphans 2>/dev/null || true
    
    # 删除相关镜像
    docker images | grep -E 'aimonitor|ai-monitor' | awk '{print $3}' | xargs -r docker rmi -f 2>/dev/null || true
    
    # 清理构建缓存
    docker builder prune -f 2>/dev/null || true
    
    log_success "Docker缓存清理完成"
}

# 创建优化的后端Dockerfile
create_backend_dockerfile() {
    log_info "创建优化的后端Dockerfile"
    
    cat > Dockerfile.backend << 'EOF'
# 多阶段构建 - 后端服务
FROM golang:1.23-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git ca-certificates tzdata curl

# 设置工作目录
WORKDIR /app

# 配置Go环境 - 使用中国镜像源
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOSUMDB=sum.golang.org
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖 - 添加重试机制
RUN echo "开始Go模块依赖下载..." && \
    for i in 1 2 3 4 5; do \
        echo "第 $i 次尝试下载依赖" && \
        go mod download -x && break || \
        (echo "下载失败，等待10秒后重试..." && sleep 10); \
    done && \
    echo "依赖下载完成" && \
    go mod verify

# 复制源代码
COPY . .

# 构建应用
RUN echo "开始应用构建..." && \
    go build -ldflags "-w -s" -o main cmd/server/main.go && \
    echo "应用构建完成"

# 运行时阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata curl && \
    addgroup -g 1000 appgroup && \
    adduser -D -s /bin/sh -u 1000 -G appgroup appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 设置文件权限
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./main"]
EOF
    
    log_success "后端Dockerfile创建完成"
}

# 创建优化的前端Dockerfile
create_frontend_dockerfile() {
    log_info "创建优化的前端Dockerfile"
    
    cat > web/Dockerfile.frontend << 'EOF'
# 多阶段构建 - 前端服务
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置npm镜像源
RUN npm config set registry https://registry.npmmirror.com/

# 复制package文件
COPY package*.json ./

# 安装依赖 - 添加重试机制
RUN echo "开始安装前端依赖..." && \
    for i in 1 2 3; do \
        echo "第 $i 次尝试安装依赖" && \
        npm install && break || \
        (echo "安装失败，清理缓存后重试..." && npm cache clean --force && sleep 5); \
    done && \
    echo "依赖安装完成"

# 复制源代码
COPY . .

# 构建应用
RUN echo "开始前端构建..." && \
    npm run build && \
    echo "前端构建完成"

# 运行时阶段
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制nginx配置
COPY nginx.conf /etc/nginx/nginx.conf

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:80 || exit 1

# 暴露端口
EXPOSE 80

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]
EOF
    
    # 创建nginx配置文件
    cat > web/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    sendfile        on;
    keepalive_timeout  65;
    
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;
    
    server {
        listen       80;
        server_name  localhost;
        
        location / {
            root   /usr/share/nginx/html;
            index  index.html index.htm;
            try_files $uri $uri/ /index.html;
        }
        
        location /api {
            proxy_pass http://aimonitor:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
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
        
        error_page   500 502 503 504  /50x.html;
        location = /50x.html {
            root   /usr/share/nginx/html;
        }
    }
}
EOF
    
    log_success "前端Dockerfile和nginx配置创建完成"
}

# 构建后端镜像
build_backend_image() {
    log_info "构建后端Docker镜像"
    
    IMAGE_NAME="ai-monitor-backend:latest"
    
    docker build -f Dockerfile.backend -t "$IMAGE_NAME" . --no-cache
    
    if [ $? -eq 0 ]; then
        log_success "后端镜像构建成功: $IMAGE_NAME"
    else
        log_error "后端镜像构建失败"
        exit 1
    fi
}

# 构建前端镜像
build_frontend_image() {
    log_info "构建前端Docker镜像"
    
    IMAGE_NAME="ai-monitor-frontend:latest"
    
    cd web
    docker build -f Dockerfile.frontend -t "$IMAGE_NAME" . --no-cache
    cd ..
    
    if [ $? -eq 0 ]; then
        log_success "前端镜像构建成功: $IMAGE_NAME"
    else
        log_error "前端镜像构建失败"
        exit 1
    fi
}

# 保存镜像到文件
save_images() {
    log_info "保存Docker镜像到文件"
    
    # 保存后端镜像
    log_info "保存后端镜像..."
    docker save ai-monitor-backend:latest | gzip > "$OUTPUT_DIR/ai-monitor-backend.tar.gz"
    
    # 保存前端镜像
    log_info "保存前端镜像..."
    docker save ai-monitor-frontend:latest | gzip > "$OUTPUT_DIR/ai-monitor-frontend.tar.gz"
    
    # 保存基础镜像
    log_info "保存基础镜像..."
    docker save postgres:15-alpine | gzip > "$OUTPUT_DIR/postgres.tar.gz"
    docker save redis:7-alpine | gzip > "$OUTPUT_DIR/redis.tar.gz"
    docker save prom/prometheus:latest | gzip > "$OUTPUT_DIR/prometheus.tar.gz"
    docker save docker.elastic.co/elasticsearch/elasticsearch:8.11.0 | gzip > "$OUTPUT_DIR/elasticsearch.tar.gz"
    
    log_success "所有镜像保存完成"
}

# 创建镜像清单
create_image_manifest() {
    log_info "创建镜像清单文件"
    
    cat > "$OUTPUT_DIR/images.txt" << EOF
# AI Monitor Docker镜像清单
# 构建时间: $(date)
# 构建版本: 2.0.0

## 应用镜像
ai-monitor-backend.tar.gz    # 后端服务镜像
ai-monitor-frontend.tar.gz   # 前端服务镜像

## 基础服务镜像
postgres.tar.gz             # PostgreSQL数据库
redis.tar.gz                # Redis缓存
prometheus.tar.gz           # Prometheus监控
elasticsearch.tar.gz        # Elasticsearch搜索引擎

## 使用说明
# 1. 将所有.tar.gz文件传输到目标服务器
# 2. 在目标服务器上运行 load_images.sh 加载镜像
# 3. 运行 deploy.sh 启动服务
EOF
    
    log_success "镜像清单创建完成"
}

# 创建镜像加载脚本
create_load_script() {
    log_info "创建镜像加载脚本"
    
    cat > "$OUTPUT_DIR/load_images.sh" << 'EOF'
#!/bin/bash

# AI Monitor 镜像加载脚本

set -e

echo "开始加载Docker镜像..."

# 检查Docker环境
if ! command -v docker &> /dev/null; then
    echo "错误: Docker 未安装"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo "错误: Docker 服务未运行"
    exit 1
fi

# 加载镜像
echo "加载后端镜像..."
docker load < ai-monitor-backend.tar.gz

echo "加载前端镜像..."
docker load < ai-monitor-frontend.tar.gz

echo "加载PostgreSQL镜像..."
docker load < postgres.tar.gz

echo "加载Redis镜像..."
docker load < redis.tar.gz

echo "加载Prometheus镜像..."
docker load < prometheus.tar.gz

echo "加载Elasticsearch镜像..."
docker load < elasticsearch.tar.gz

echo "所有镜像加载完成！"
echo "可以运行 deploy.sh 开始部署服务"
EOF
    
    chmod +x "$OUTPUT_DIR/load_images.sh"
    log_success "镜像加载脚本创建完成"
}

# 显示构建结果
show_build_result() {
    echo ""
    echo -e "${GREEN}🎉 镜像构建完成！${NC}"
    echo "=========================================="
    echo -e "${BLUE}📦 构建产物:${NC}"
    echo "  输出目录: $OUTPUT_DIR"
    echo "  镜像文件:"
    ls -lh "$OUTPUT_DIR"/*.tar.gz | awk '{print "    " $9 " (" $5 ")"}'
    echo ""
    echo -e "${YELLOW}📋 下一步操作:${NC}"
    echo "  1. 将 $OUTPUT_DIR 目录传输到远程服务器"
    echo "  2. 在远程服务器上运行: ./load_images.sh"
    echo "  3. 运行部署脚本启动服务"
    echo ""
    echo -e "${GREEN}✅ 构建成功！镜像已准备就绪${NC}"
    echo ""
}

# 主函数
main() {
    show_welcome
    check_project_structure
    check_docker
    create_output_directory
    clean_docker_cache
    create_backend_dockerfile
    create_frontend_dockerfile
    build_backend_image
    build_frontend_image
    save_images
    create_image_manifest
    create_load_script
    show_build_result
}

# 错误处理
trap 'log_error "构建过程中发生错误，请检查上述日志信息"' ERR

# 执行主函数
main