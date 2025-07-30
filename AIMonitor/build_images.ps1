# =============================================================================
# AI Monitor Image Build Script (PowerShell)
# Version: 2.0.0
# Description: Two-step deployment - Step 1: Build frontend and backend Docker images
# Author: AI Assistant
# =============================================================================

# Error handling
$ErrorActionPreference = "Stop"

# Get current timestamp
function Get-Timestamp {
    return Get-Date -Format "yyyy-MM-dd HH:mm:ss"
}

# Log functions
function Write-Info {
    param([string]$Message)
    Write-Host "[$(Get-Timestamp)] INFO: $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[$(Get-Timestamp)] SUCCESS: $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[$(Get-Timestamp)] WARNING: $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[$(Get-Timestamp)] ERROR: $Message" -ForegroundColor Red
}

# Show welcome message
function Show-Welcome {
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host "    AI Monitor Image Build Tool v2.0" -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "This tool will build frontend and backend Docker images and save them locally"
    Write-Host "After completion, you can transfer them to remote servers for deployment"
    Write-Host ""
}

# Check project structure
function Test-ProjectStructure {
    Write-Info "Checking project structure"
    
    if (-not (Test-Path "go.mod")) {
        Write-Error "go.mod file not found, please run this script in the project root directory"
        exit 1
    }
    
    if (-not (Test-Path "cmd/server/main.go")) {
        Write-Error "cmd/server/main.go file not found, please ensure project structure is complete"
        exit 1
    }
    
    if (-not (Test-Path "web")) {
        Write-Error "web directory not found, please ensure frontend project exists"
        exit 1
    }
    
    if (-not (Test-Path "web/package.json")) {
        Write-Error "web/package.json file not found, please ensure frontend project is complete"
        exit 1
    }
    
    Write-Success "Project structure check passed"
}

# Check Docker environment
function Test-Docker {
    Write-Info "Checking Docker environment"
    
    try {
        $null = Get-Command docker -ErrorAction Stop
    } catch {
        Write-Error "Docker not installed, please install Docker Desktop first"
        exit 1
    }
    
    try {
        docker info | Out-Null
    } catch {
        Write-Error "Docker service not running, please start Docker Desktop"
        exit 1
    }
    
    Write-Success "Docker environment check passed"
}

# Create image output directory
function New-OutputDirectory {
    Write-Info "Creating image output directory"
    
    $script:OutputDir = "./docker-images"
    
    if (Test-Path $OutputDir) {
        Write-Warning "Output directory already exists, cleaning old files"
        Remove-Item "$OutputDir/*" -Recurse -Force -ErrorAction SilentlyContinue
    } else {
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    }
    
    Write-Success "Output directory created: $OutputDir"
}

# Clear Docker cache
function Clear-DockerCache {
    Write-Info "Clearing Docker build cache"
    
    try {
        # Stop related containers
        docker-compose down --remove-orphans 2>$null
        
        # Remove related images
        $images = docker images --format "{{.Repository}}:{{.Tag}} {{.ID}}" | Where-Object { $_ -match 'aimonitor|ai-monitor' }
        if ($images) {
            $imageIds = $images | ForEach-Object { ($_ -split ' ')[1] }
            docker rmi $imageIds -f 2>$null
        }
        
        # Clear build cache
        docker builder prune -f 2>$null
        
        Write-Success "Docker cache cleanup completed"
    } catch {
        Write-Warning "Some Docker cleanup operations failed, continuing..."
    }
}

# Create optimized backend Dockerfile
function New-BackendDockerfile {
    Write-Info "Creating optimized backend Dockerfile"
    
    $dockerfileLines = @(
        "# Multi-stage build - Backend service",
        "FROM golang:1.23-alpine AS builder",
        "",
        "# Install necessary tools",
        "RUN apk add --no-cache git ca-certificates tzdata curl",
        "",
        "# Set working directory",
        "WORKDIR /app",
        "",
        "# Configure Go environment - Use Chinese mirror sources",
        "ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct",
        "ENV GOSUMDB=sum.golang.org",
        "ENV GO111MODULE=on",
        "ENV CGO_ENABLED=0",
        "ENV GOOS=linux",
        "ENV GOARCH=amd64",
        "",
        "# Copy go.mod and go.sum",
        "COPY go.mod go.sum ./",
        "",
        "# Download dependencies with retry mechanism",
         "RUN echo 'Starting Go module dependency download...' && \",
         "    for i in 1 2 3 4 5; do \",
         "        echo 'Attempt `$i to download dependencies' && \",
         "        go mod download -x && break || \",
         "        (echo 'Download failed, waiting 10 seconds before retry...' && sleep 10); \",
         "    done && \",
         "    echo 'Dependency download completed' && \",
         "    go mod verify",
        "",
        "# Copy source code",
        "COPY . .",
        "",
        "# Build application",
         "RUN echo 'Starting application build...' && \",
         "    go build -ldflags '-w -s' -o main cmd/server/main.go && \",
         "    echo 'Application build completed'",
        "",
        "# Runtime stage",
        "FROM alpine:latest",
        "",
        "# Install runtime dependencies",
         "RUN apk --no-cache add ca-certificates tzdata curl && \",
         "    addgroup -g 1000 appgroup && \",
         "    adduser -D -s /bin/sh -u 1000 -G appgroup appuser",
        "",
        "# Set working directory",
        "WORKDIR /app",
        "",
        "# Copy files from build stage",
        "COPY --from=builder /app/main .",
        "COPY --from=builder /app/configs ./configs",
        "",
        "# Set file permissions",
        "RUN chown -R appuser:appgroup /app",
        "",
        "# Switch to non-root user",
        "USER appuser",
        "",
        "# Health check",
         "HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \",
         "    CMD curl -f http://localhost:8080/health || exit 1",
        "",
        "# Expose port",
        "EXPOSE 8080",
        "",
        "# Start application",
         "CMD [`"./main`"]"
    )
    
    $dockerfileLines | Out-File -FilePath "Dockerfile.backend" -Encoding UTF8
    Write-Success "Backend Dockerfile created"
}

# Create optimized frontend Dockerfile
function New-FrontendDockerfile {
    Write-Info "Creating optimized frontend Dockerfile"
    
    $frontendDockerfileLines = @(
        "# Multi-stage build - Frontend service",
        "FROM node:18-alpine AS builder",
        "",
        "# Set working directory",
        "WORKDIR /app",
        "",
        "# Set npm mirror source",
        "RUN npm config set registry https://registry.npmmirror.com/",
        "",
        "# Copy package files",
        "COPY package*.json ./",
        "",
        "# Install build dependencies and set environment",
           "RUN apk add --no-cache python3 make g++ && \",
           "    ln -sf /usr/bin/python3 /usr/bin/python",
          "",
          "# Set environment variables for npm",
          "ENV PYTHON=/usr/bin/python3",
          "ENV npm_config_python=/usr/bin/python3",
          "",
          "# Install dependencies with retry mechanism",
          "RUN echo 'Starting frontend dependency installation...' && \",
          "    for i in 1 2 3; do \",
          "        echo 'Attempt `$i to install dependencies' && \",
          "        npm install && break || \",
          "        (echo 'Installation failed, cleaning cache and retrying...' && npm cache clean --force && sleep 5); \",
          "    done && \",
          "    echo 'Dependency installation completed'",
        "",
        "# Copy source code",
        "COPY . .",
        "",
        "# Build application",
        "RUN echo 'Starting frontend build...' && \",
        "    npm run build && \",
        "    echo 'Frontend build completed'",
        "",
        "# Runtime stage",
        "FROM nginx:alpine",
        "",
        "# Copy build artifacts",
        "COPY --from=builder /app/dist /usr/share/nginx/html",
        "",
        "# Copy nginx configuration",
        "COPY nginx.conf /etc/nginx/nginx.conf",
        "",
        "# Health check",
        "HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \",
        "    CMD curl -f http://localhost:80 || exit 1",
        "",
        "# Expose port",
        "EXPOSE 80",
        "",
        "# Start nginx",
         "CMD [`"nginx`", `"-g`", `"daemon off;`"]"
    )
    
    $frontendDockerfileLines | Out-File -FilePath "web/Dockerfile.frontend" -Encoding UTF8
    
    # Create nginx configuration file
    $nginxConfigLines = @(
        "events {",
        "    worker_connections 1024;",
        "}",
        "",
        "http {",
        "    include       /etc/nginx/mime.types;",
        "    default_type  application/octet-stream;",
        "    ",
        "    sendfile        on;",
        "    keepalive_timeout  65;",
        "    ",
        "    gzip on;",
        "    gzip_vary on;",
        "    gzip_min_length 1024;",
        "    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;",
        "    ",
        "    server {",
        "        listen       80;",
        "        server_name  localhost;",
        "        ",
        "        location / {",
        "            root   /usr/share/nginx/html;",
        "            index  index.html index.htm;",
        "            try_files `$uri `$uri/ /index.html;",
        "        }",
        "        ",
        "        location /api {",
        "            proxy_pass http://aimonitor:8080;",
        "            proxy_set_header Host `$host;",
        "            proxy_set_header X-Real-IP `$remote_addr;",
        "            proxy_set_header X-Forwarded-For `$proxy_add_x_forwarded_for;",
        "            proxy_set_header X-Forwarded-Proto `$scheme;",
        "        }",
        "        ",
        "        error_page   500 502 503 504  /50x.html;",
        "        location = /50x.html {",
        "            root   /usr/share/nginx/html;",
        "        }",
        "    }",
        "}"
    )
    
    $nginxConfigLines | Out-File -FilePath "web/nginx.conf" -Encoding UTF8
    
    Write-Success "Frontend Dockerfile and nginx configuration created"
}

# Build backend image
function Build-BackendImage {
    Write-Info "Building backend Docker image"
    
    $imageName = "ai-monitor-backend:latest"
    
    docker build -f Dockerfile.backend -t $imageName . --no-cache
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Backend image build successful: $imageName"
    } else {
        Write-Error "Backend image build failed"
        exit 1
    }
}

# Build frontend image
function Build-FrontendImage {
    Write-Info "Building frontend Docker image"
    
    $imageName = "ai-monitor-frontend:latest"
    
    Push-Location web
    try {
        docker build -f Dockerfile.frontend -t $imageName . --no-cache
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Frontend image build successful: $imageName"
        } else {
            Write-Error "Frontend image build failed"
            exit 1
        }
    } finally {
        Pop-Location
    }
}

# Save images to files
function Save-Images {
    Write-Info "Saving Docker images to files"
    
    # Save backend image
    Write-Info "Saving backend image..."
    docker save ai-monitor-backend:latest -o "$OutputDir/ai-monitor-backend.tar"
    
    # Save frontend image
    Write-Info "Saving frontend image..."
    docker save ai-monitor-frontend:latest -o "$OutputDir/ai-monitor-frontend.tar"
    
    # Save base images
    Write-Info "Saving base images..."
    docker save postgres:15-alpine -o "$OutputDir/postgres.tar"
    docker save redis:7-alpine -o "$OutputDir/redis.tar"
    docker save prom/prometheus:latest -o "$OutputDir/prometheus.tar"
    docker save docker.elastic.co/elasticsearch/elasticsearch:8.11.0 -o "$OutputDir/elasticsearch.tar"
    
    Write-Success "All images saved"
}

# Create image manifest
function New-ImageManifest {
    Write-Info "Creating image manifest file"
    
    $manifestLines = @(
        "# AI Monitor Docker Image Manifest",
        "# Build time: $(Get-Date)",
        "# Build version: 2.0.0",
        "",
        "## Application Images",
        "ai-monitor-backend.tar    # Backend service image",
        "ai-monitor-frontend.tar   # Frontend service image",
        "",
        "## Base Service Images",
        "postgres.tar             # PostgreSQL database",
        "redis.tar                # Redis cache",
        "prometheus.tar           # Prometheus monitoring",
        "elasticsearch.tar        # Elasticsearch search engine",
        "",
        "## Usage Instructions",
        "# 1. Transfer all .tar files to target server",
        "# 2. Run load_images.sh on target server to load images",
        "# 3. Run deploy_from_images.sh to start services"
    )
    
    $manifestLines | Out-File -FilePath "$OutputDir/images.txt" -Encoding UTF8
    Write-Success "Image manifest created"
}

# Create image loading script
function New-LoadScript {
    Write-Info "Creating image loading script"
    
    $loadScriptLines = @(
        "#!/bin/bash",
        "",
        "# AI Monitor Image Loading Script",
        "",
        "set -e",
        "",
        "echo 'Starting Docker image loading...'",
        "",
        "# Check Docker environment",
        "if ! command -v docker &> /dev/null; then",
        "    echo 'Error: Docker not installed'",
        "    exit 1",
        "fi",
        "",
        "if ! docker info &> /dev/null; then",
        "    echo 'Error: Docker service not running'",
        "    exit 1",
        "fi",
        "",
        "# Load images",
        "echo 'Loading backend image...'",
        "docker load < ai-monitor-backend.tar",
        "",
        "echo 'Loading frontend image...'",
        "docker load < ai-monitor-frontend.tar",
        "",
        "echo 'Loading PostgreSQL image...'",
        "docker load < postgres.tar",
        "",
        "echo 'Loading Redis image...'",
        "docker load < redis.tar",
        "",
        "echo 'Loading Prometheus image...'",
        "docker load < prometheus.tar",
        "",
        "echo 'Loading Elasticsearch image...'",
        "docker load < elasticsearch.tar",
        "",
        "echo 'All images loaded successfully!'",
        "echo 'You can now run deploy_from_images.sh to start services'"
    )
    
    $loadScriptLines | Out-File -FilePath "$OutputDir/load_images.sh" -Encoding UTF8
    Write-Success "Image loading script created"
}

# Show build results
function Show-BuildResult {
    Write-Host ""
    Write-Host "Image build completed!" -ForegroundColor Green
    Write-Host "=========================================="
    Write-Host "Build artifacts:" -ForegroundColor Blue
    Write-Host "  Output directory: $OutputDir"
    Write-Host "  Image files:"
    
    Get-ChildItem "$OutputDir/*.tar" | ForEach-Object {
        $size = [math]::Round($_.Length / 1MB, 2)
        Write-Host "    $($_.Name) ($size MB)"
    }
    
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "  1. Transfer $OutputDir directory to remote server"
    Write-Host "  2. Run on remote server: ./load_images.sh"
    Write-Host "  3. Run deploy_from_images.sh to start services"
    Write-Host ""
    Write-Host "Build successful! Images are ready for deployment" -ForegroundColor Green
    Write-Host ""
}

# Main function
function Main {
    try {
        Show-Welcome
        Test-ProjectStructure
        Test-Docker
        New-OutputDirectory
        Clear-DockerCache
        New-BackendDockerfile
        New-FrontendDockerfile
        Build-BackendImage
        Build-FrontendImage
        Save-Images
        New-ImageManifest
        New-LoadScript
        Show-BuildResult
    } catch {
        Write-Error "Error occurred during build process: $($_.Exception.Message)"
        Write-Host "Please check the log information above" -ForegroundColor Red
        exit 1
    }
}

# Execute main function
Main