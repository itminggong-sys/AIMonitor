# AI Monitor Windows PowerShell 部署脚本
# 适用于: Windows 10/11 + Docker Desktop
# 作者: AI Monitor Team
# 版本: v1.0.0

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色定义
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    switch ($Color) {
        "Red" { Write-Host $Message -ForegroundColor Red }
        "Green" { Write-Host $Message -ForegroundColor Green }
        "Yellow" { Write-Host $Message -ForegroundColor Yellow }
        "Blue" { Write-Host $Message -ForegroundColor Blue }
        "Cyan" { Write-Host $Message -ForegroundColor Cyan }
        default { Write-Host $Message }
    }
}

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Blue"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

# 显示欢迎信息
function Show-Welcome {
    Clear-Host
    Write-ColorOutput "=================================================" "Green"
    Write-ColorOutput "    AI Monitor Windows 部署脚本 v1.0.0" "Green"
    Write-ColorOutput "=================================================" "Green"
    Write-Host ""
    Write-Host "适用环境:"
    Write-Host "  ✓ Windows 10/11"
    Write-Host "  ✓ Docker Desktop for Windows"
    Write-Host "  ✓ 内存: 4GB+ (推荐8GB+)"
    Write-Host "  ✓ 存储: 10GB+ 可用空间"
    Write-Host ""
    Write-Host "部署内容:"
    Write-Host "  ✓ 后端服务 (Go + Gin) - 端口8080"
    Write-Host "  ✓ 前端服务 (React + Vite) - 端口3000"
    Write-Host "  ✓ PostgreSQL 数据库 - 端口5432"
    Write-Host "  ✓ Redis 缓存 - 端口6379"
    Write-Host "  ✓ Prometheus 监控 - 端口9090"
    Write-Host "  ✓ Elasticsearch 搜索 - 端口9200"
    Write-Host ""
    Read-Host "按回车键开始部署..."
}

# 检查管理员权限
function Test-AdminRights {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 检查Docker Desktop
function Test-DockerDesktop {
    Write-Info "检查Docker Desktop状态..."
    
    # 检查docker命令是否可用
    try {
        $dockerVersion = docker --version
        Write-Success "Docker已安装: $dockerVersion"
    }
    catch {
        Write-Error "未找到Docker命令"
        Write-Info "请安装Docker Desktop:"
        Write-Host "  1. 访问: https://www.docker.com/products/docker-desktop"
        Write-Host "  2. 下载并安装Docker Desktop"
        Write-Host "  3. 启动Docker Desktop"
        Write-Host "  4. 重新运行此脚本"
        exit 1
    }
    
    # 检查docker服务是否运行
    try {
        docker ps | Out-Null
        Write-Success "Docker服务正常运行"
    }
    catch {
        Write-Error "Docker服务未运行"
        Write-Info "请启动Docker Desktop应用程序"
        Write-Info "等待Docker Desktop启动后重新运行此脚本"
        exit 1
    }
    
    # 检查docker-compose
    try {
        $composeVersion = docker-compose --version
        Write-Success "Docker Compose已安装: $composeVersion"
    }
    catch {
        try {
            docker compose version | Out-Null
            Write-Success "Docker Compose (内置版本) 可用"
        }
        catch {
            Write-Error "未找到docker-compose命令"
            Write-Info "请确保Docker Desktop版本支持docker-compose"
            exit 1
        }
    }
}

# 检查系统资源
function Test-SystemResources {
    Write-Info "检查系统资源..."
    
    # 检查内存
    $totalMemoryGB = [math]::Round((Get-CimInstance Win32_PhysicalMemory | Measure-Object -Property capacity -Sum).sum / 1GB, 2)
    if ($totalMemoryGB -lt 4) {
        Write-Warning "系统内存不足4GB，可能影响性能"
    } else {
        Write-Success "内存检查通过: ${totalMemoryGB}GB"
    }
    
    # 检查磁盘空间
    $freeSpaceGB = [math]::Round((Get-PSDrive C).Free / 1GB, 2)
    if ($freeSpaceGB -lt 10) {
        Write-Warning "C盘可用空间不足10GB，可能影响部署"
    } else {
        Write-Success "磁盘空间检查通过: ${freeSpaceGB}GB可用"
    }
}

# 检查端口占用
function Test-Ports {
    Write-Info "检查端口占用情况..."
    
    $ports = @(3000, 5432, 6379, 8080, 9090, 9200, 9300)
    $occupiedPorts = @()
    
    foreach ($port in $ports) {
        $connection = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
        if ($connection) {
            $occupiedPorts += $port
        }
    }
    
    if ($occupiedPorts.Count -gt 0) {
        Write-Warning "以下端口已被占用: $($occupiedPorts -join ', ')"
        Write-Info "如果继续部署，这些服务可能无法启动"
        $continue = Read-Host "是否继续部署? (y/N)"
        if ($continue -notmatch '^[Yy]$') {
            Write-Info "部署已取消"
            exit 0
        }
    } else {
        Write-Success "所有必需端口都可用"
    }
}

# 准备项目环境
function Initialize-Project {
    Write-Info "准备项目环境..."
    
    # 检查项目文件
    if (-not (Test-Path "docker-compose.yml")) {
        Write-Error "未找到docker-compose.yml文件"
        Write-Info "请确保在项目根目录中运行此脚本"
        exit 1
    }
    
    if (-not (Test-Path "go.mod")) {
        Write-Error "未找到go.mod文件"
        Write-Info "请确保在项目根目录中运行此脚本"
        exit 1
    }
    
    if (-not (Test-Path "web")) {
        Write-Error "未找到web目录"
        Write-Info "请确保在项目根目录中运行此脚本"
        exit 1
    }
    
    Write-Success "项目文件检查通过"
    
    # 创建必要的目录
    if (-not (Test-Path "logs")) {
        New-Item -ItemType Directory -Path "logs" | Out-Null
    }
    
    # 设置环境变量文件
    if (-not (Test-Path ".env")) {
        Write-Info "创建环境变量文件..."
        $envContent = @"
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
"@
        Set-Content -Path ".env" -Value $envContent -Encoding UTF8
        Write-Success "环境变量文件已创建"
    }
}

# 清理旧容器和镜像
function Clear-OldContainers {
    Write-Info "清理旧的容器和镜像..."
    
    # 停止并删除旧容器
    try {
        $runningContainers = docker-compose ps -q 2>$null
        if ($runningContainers) {
            Write-Info "停止现有服务..."
            docker-compose down --remove-orphans 2>$null
        }
    }
    catch {
        # 忽略错误，可能没有运行的容器
    }
    
    # 清理未使用的镜像和容器
    Write-Info "清理Docker缓存..."
    try {
        docker system prune -f --volumes 2>$null
    }
    catch {
        Write-Warning "清理Docker缓存时出现警告，继续部署..."
    }
    
    Write-Success "清理完成"
}

# 构建和启动服务
function Deploy-Services {
    Write-Info "开始构建和部署服务..."
    
    # 拉取基础镜像
    Write-Info "拉取Docker基础镜像..."
    $baseImages = @(
        "postgres:15-alpine",
        "redis:7-alpine",
        "golang:1.21-alpine",
        "node:18-alpine",
        "prom/prometheus:latest",
        "docker.elastic.co/elasticsearch/elasticsearch:8.11.0"
    )
    
    foreach ($image in $baseImages) {
        Write-Info "拉取镜像: $image"
        docker pull $image
    }
    
    # 构建并启动服务
    Write-Info "构建并启动所有服务..."
    
    try {
        docker-compose up -d --build
    }
    catch {
        try {
            docker compose up -d --build
        }
        catch {
            Write-Error "服务启动失败"
            throw
        }
    }
    
    Write-Success "服务部署完成"
}

# 等待服务启动
function Wait-ForServices {
    Write-Info "等待服务启动..."
    
    # 等待数据库启动
    Write-Info "等待PostgreSQL启动..."
    for ($i = 1; $i -le 30; $i++) {
        try {
            docker exec ai-monitor-postgres pg_isready -U ai_monitor -d ai_monitor 2>$null | Out-Null
            Write-Success "PostgreSQL已启动"
            break
        }
        catch {
            if ($i -eq 30) {
                Write-Error "PostgreSQL启动超时"
                return $false
            }
            Start-Sleep -Seconds 2
        }
    }
    
    # 等待Redis启动
    Write-Info "等待Redis启动..."
    for ($i = 1; $i -le 30; $i++) {
        try {
            docker exec ai-monitor-redis redis-cli ping 2>$null | Out-Null
            Write-Success "Redis已启动"
            break
        }
        catch {
            if ($i -eq 30) {
                Write-Error "Redis启动超时"
                return $false
            }
            Start-Sleep -Seconds 2
        }
    }
    
    # 等待后端服务启动
    Write-Info "等待后端服务启动..."
    for ($i = 1; $i -le 60; $i++) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 5 -ErrorAction Stop
            Write-Success "后端服务已启动"
            break
        }
        catch {
            if ($i -eq 60) {
                Write-Warning "后端服务启动超时，请检查日志"
                break
            }
            Start-Sleep -Seconds 3
        }
    }
    
    # 等待前端服务启动
    Write-Info "等待前端服务启动..."
    for ($i = 1; $i -le 60; $i++) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:3000" -TimeoutSec 5 -ErrorAction Stop
            Write-Success "前端服务已启动"
            break
        }
        catch {
            if ($i -eq 60) {
                Write-Warning "前端服务启动超时，请检查日志"
                break
            }
            Start-Sleep -Seconds 3
        }
    }
    
    return $true
}

# 显示服务状态
function Show-ServiceStatus {
    Write-Info "检查服务状态..."
    
    Write-Host ""
    Write-Host "=== 服务状态 ==="
    try {
        docker-compose ps
    }
    catch {
        docker compose ps
    }
    
    Write-Host ""
    Write-Host "=== 端口检查 ==="
    $services = @(
        @{Name="前端服务"; Port=3000},
        @{Name="后端API"; Port=8080},
        @{Name="PostgreSQL"; Port=5432},
        @{Name="Redis"; Port=6379},
        @{Name="Prometheus"; Port=9090},
        @{Name="Elasticsearch"; Port=9200}
    )
    
    foreach ($service in $services) {
        $connection = Get-NetTCPConnection -LocalPort $service.Port -ErrorAction SilentlyContinue
        if ($connection) {
            Write-ColorOutput "  ✓ $($service.Name) (端口$($service.Port)): 运行中" "Green"
        } else {
            Write-ColorOutput "  ✗ $($service.Name) (端口$($service.Port)): 未运行" "Red"
        }
    }
}

# 显示访问信息
function Show-AccessInfo {
    Write-Host ""
    Write-ColorOutput "=================================================" "Green"
    Write-ColorOutput "           AI Monitor 部署成功！" "Green"
    Write-ColorOutput "=================================================" "Green"
    Write-Host ""
    Write-Host "🌐 访问地址:"
    Write-Host "  • 前端界面: http://localhost:3000"
    Write-Host "  • 后端API:  http://localhost:8080"
    Write-Host "  • API文档:  http://localhost:8080/swagger/index.html"
    Write-Host "  • 健康检查: http://localhost:8080/health"
    Write-Host ""
    Write-Host "📊 监控服务:"
    Write-Host "  • Prometheus: http://localhost:9090"
    Write-Host "  • Elasticsearch: http://localhost:9200"
    Write-Host ""
    Write-Host "🗄️ 数据库连接:"
    Write-Host "  • PostgreSQL: localhost:5432"
    Write-Host "    - 数据库: ai_monitor"
    Write-Host "    - 用户名: ai_monitor"
    Write-Host "    - 密码: password"
    Write-Host "  • Redis: localhost:6379"
    Write-Host ""
    Write-Host "🔧 管理命令:"
    Write-Host "  • 查看日志: docker-compose logs -f [服务名]"
    Write-Host "  • 重启服务: docker-compose restart [服务名]"
    Write-Host "  • 停止服务: docker-compose down"
    Write-Host "  • 启动服务: docker-compose up -d"
    Write-Host ""
    Write-Host "📝 默认账户 (如果有登录页面):"
    Write-Host "  • 用户名: admin"
    Write-Host "  • 密码: admin123"
    Write-Host ""
    Write-ColorOutput "⚠️  重要提醒:" "Yellow"
    Write-Host "  1. 首次访问可能需要等待1-2分钟"
    Write-Host "  2. 如果服务无法访问，请检查防火墙设置"
    Write-Host "  3. 生产环境请修改默认密码"
    Write-Host ""
}

# 显示故障排除信息
function Show-Troubleshooting {
    Write-Host "🔍 故障排除:"
    Write-Host ""
    Write-Host "如果遇到问题，请尝试以下步骤:"
    Write-Host ""
    Write-Host "1. 检查服务日志:"
    Write-Host "   docker-compose logs [服务名]"
    Write-Host ""
    Write-Host "2. 重启所有服务:"
    Write-Host "   docker-compose down; docker-compose up -d"
    Write-Host ""
    Write-Host "3. 清理并重新部署:"
    Write-Host "   docker-compose down -v"
    Write-Host "   docker system prune -f"
    Write-Host "   .\deploy_windows.ps1"
    Write-Host ""
    Write-Host "4. 检查Docker Desktop状态:"
    Write-Host "   确保Docker Desktop正在运行"
    Write-Host ""
    Write-Host "5. 检查Windows防火墙:"
    Write-Host "   允许Docker Desktop通过防火墙"
    Write-Host ""
}

# 主函数
function Main {
    try {
        # 显示欢迎信息
        Show-Welcome
        
        # 检查管理员权限
        if (-not (Test-AdminRights)) {
            Write-Warning "建议以管理员身份运行此脚本以获得最佳体验"
            $continue = Read-Host "是否继续? (y/N)"
            if ($continue -notmatch '^[Yy]$') {
                Write-Info "请右键点击PowerShell，选择'以管理员身份运行'"
                exit 0
            }
        }
        
        # 环境检查
        Test-DockerDesktop
        Test-SystemResources
        Test-Ports
        
        # 项目准备
        Initialize-Project
        
        # 清理旧环境
        Clear-OldContainers
        
        # 部署服务
        Deploy-Services
        
        # 等待服务启动
        $servicesReady = Wait-ForServices
        
        # 显示结果
        Show-ServiceStatus
        Show-AccessInfo
        Show-Troubleshooting
        
        Write-Success "部署完成！请访问 http://localhost:3000 查看前端界面"
        
        # 询问是否打开浏览器
        $openBrowser = Read-Host "是否现在打开浏览器访问前端界面? (y/N)"
        if ($openBrowser -match '^[Yy]$') {
            Start-Process "http://localhost:3000"
        }
    }
    catch {
        Write-Error "部署过程中发生错误: $($_.Exception.Message)"
        Write-Info "请检查上面的错误信息，或查看故障排除部分"
        exit 1
    }
}

# 运行主函数
Main