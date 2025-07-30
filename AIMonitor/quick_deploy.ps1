# =============================================================================
# AI Monitor Quick Deploy Script (PowerShell)
# Version: 2.0.0
# Description: One-click two-step deployment process
# Author: AI Assistant
# =============================================================================

# Set error handling
$ErrorActionPreference = "Stop"

# Color output function
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    
    switch ($Color) {
        "Red" { Write-Host "[$timestamp] ERROR: $Message" -ForegroundColor Red }
        "Green" { Write-Host "[$timestamp] SUCCESS: $Message" -ForegroundColor Green }
        "Yellow" { Write-Host "[$timestamp] WARNING: $Message" -ForegroundColor Yellow }
        "Blue" { Write-Host "[$timestamp] INFO: $Message" -ForegroundColor Blue }
        default { Write-Host "[$timestamp] $Message" -ForegroundColor White }
    }
}

# Log functions
function Log-Info { param([string]$Message) Write-ColorOutput $Message "Blue" }
function Log-Success { param([string]$Message) Write-ColorOutput $Message "Green" }
function Log-Warning { param([string]$Message) Write-ColorOutput $Message "Yellow" }
function Log-Error { param([string]$Message) Write-ColorOutput $Message "Red" }

# Show welcome message
function Show-Welcome {
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host "    AI Monitor Quick Deploy Tool v2.0" -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "This tool will guide you through the two-step deployment process:"
    Write-Host "1. Build Docker images (local environment)"
    Write-Host "2. Deploy to remote server (optional)"
    Write-Host ""
}

# Show menu
function Show-Menu {
    Write-Host "Please select deployment mode:" -ForegroundColor Blue
    Write-Host "1. Build images only (recommended for remote deployment)"
    Write-Host "2. Full local deployment (build + deploy)"
    Write-Host "3. Deploy only (use existing images)"
    Write-Host "4. View deployment guide"
    Write-Host "5. Exit"
    Write-Host ""
    $choice = Read-Host "Please enter option (1-5)"
    return $choice
}

# Check environment
function Test-Environment {
    Log-Info "Checking deployment environment"
    
    # Check Docker
    try {
        $null = Get-Command docker -ErrorAction Stop
        $null = docker info 2>$null
        Log-Success "Docker is available"
    }
    catch {
        Log-Error "Docker is not installed or not running"
        Write-Host "Installation guide: https://docs.docker.com/desktop/windows/" -ForegroundColor Cyan
        exit 1
    }
    
    # Check project structure
    $requiredFiles = @("go.mod", "cmd\server\main.go")
    $requiredDirs = @("web")
    
    foreach ($file in $requiredFiles) {
        if (-not (Test-Path $file)) {
            Log-Error "Missing required file: $file"
            Log-Error "Please run this script in the AI Monitor project root directory"
            exit 1
        }
    }
    
    foreach ($dir in $requiredDirs) {
        if (-not (Test-Path $dir -PathType Container)) {
            Log-Error "Missing required directory: $dir"
            Log-Error "Please run this script in the AI Monitor project root directory"
            exit 1
        }
    }
    
    Log-Success "Environment check passed"
}

# Build images
function Build-Images {
    Log-Info "Starting Docker image build"
    
    if (Test-Path "build_images.ps1") {
        try {
            & .\build_images.ps1
            if ($LASTEXITCODE -ne 0) {
                throw "Build script returned error code: $LASTEXITCODE"
            }
            Log-Success "Image build completed"
        }
        catch {
            Log-Error "Build script execution failed: $_"
            throw $_
        }
    }
    else {
        Log-Error "Build script build_images.ps1 not found"
        throw "Build script not found"
    }
}

# Deploy locally
function Deploy-Local {
    Log-Info "Starting local deployment"
    
    # Check if images exist
    try {
        $backendImage = docker images --format "table {{.Repository}}:{{.Tag}}" | Select-String "ai-monitor-backend:latest"
        if (-not $backendImage) {
            Log-Warning "Backend image not found, starting build..."
            Build-Images
        }
    }
    catch {
        Log-Error "Error checking images: $_"
        throw $_
    }
    
    if (Test-Path "deploy_from_images.sh") {
        try {
            if (Get-Command wsl -ErrorAction SilentlyContinue) {
                wsl ./deploy_from_images.sh
            }
            elseif (Get-Command bash -ErrorAction SilentlyContinue) {
                bash ./deploy_from_images.sh
            }
            else {
                Log-Error "WSL or Git Bash required to run deployment script"
                Log-Info "Please manually run: ./deploy_from_images.sh"
                throw "Missing WSL or Git Bash"
            }
            Log-Success "Local deployment completed"
        }
        catch {
            Log-Error "Deployment script execution failed: $_"
            throw $_
        }
    }
    else {
        Log-Error "Deployment script deploy_from_images.sh not found"
        throw "Deployment script not found"
    }
}

# Deploy only
function Deploy-Only {
    Log-Info "Starting service deployment"
    
    # Check required images
    $requiredImages = @(
        "ai-monitor-backend:latest",
        "ai-monitor-frontend:latest",
        "postgres:15-alpine",
        "redis:7-alpine"
    )
    
    try {
        $missingImages = @()
        $existingImages = docker images --format "{{.Repository}}:{{.Tag}}"
        
        foreach ($image in $requiredImages) {
            if ($existingImages -notcontains $image) {
                $missingImages += $image
            }
        }
        
        if ($missingImages.Count -gt 0) {
            Log-Error "Missing the following Docker images:"
            foreach ($image in $missingImages) {
                Write-Host "  - $image" -ForegroundColor Red
            }
            Write-Host ""
            Log-Error "Please build images or load image files first"
            throw "Missing required images"
        }
    }
    catch {
        Log-Error "Error checking images: $_"
        throw $_
    }
    
    if (Test-Path "deploy_from_images.sh") {
        try {
            if (Get-Command wsl -ErrorAction SilentlyContinue) {
                wsl ./deploy_from_images.sh
            }
            elseif (Get-Command bash -ErrorAction SilentlyContinue) {
                bash ./deploy_from_images.sh
            }
            else {
                Log-Error "WSL or Git Bash required to run deployment script"
                Log-Info "Please manually run: ./deploy_from_images.sh"
                throw "Missing WSL or Git Bash"
            }
            Log-Success "Service deployment completed"
        }
        catch {
            Log-Error "Deployment script execution failed: $_"
            throw $_
        }
    }
    else {
        Log-Error "Deployment script deploy_from_images.sh not found"
        throw "Deployment script not found"
    }
}

# Show deployment guide
function Show-Guide {
    Write-Host "=== AI Monitor Two-Step Deployment Guide ===" -ForegroundColor Blue
    Write-Host ""
    Write-Host "Step 1: Build Images (Local Environment)" -ForegroundColor Yellow
    Write-Host "1. Run build script in stable network environment"
    Write-Host "2. After completion, docker-images/ directory will be generated"
    Write-Host "3. This directory contains all required Docker image files"
    Write-Host ""
    Write-Host "Step 2: Remote Deployment" -ForegroundColor Yellow
    Write-Host "1. Transfer docker-images/ directory to remote server"
    Write-Host "   scp -r docker-images/ user@server:/path/to/deployment/"
    Write-Host "2. Load images on remote server"
    Write-Host "   cd docker-images"
    Write-Host "   chmod +x load_images.sh"
    Write-Host "   ./load_images.sh"
    Write-Host "3. Execute deployment script"
    Write-Host "   ./deploy_from_images.sh"
    Write-Host ""
    Write-Host "Advantages:" -ForegroundColor Green
    Write-Host "- Avoid remote network issues causing build failures"
    Write-Host "- Deployment speed improved by 80%+"
    Write-Host "- Support offline deployment"
    Write-Host "- Easy version management and rollback"
    Write-Host ""
    Write-Host "Detailed documentation: TWO_STEP_DEPLOYMENT_GUIDE.md" -ForegroundColor Blue
    Write-Host ""
    Read-Host "Press Enter to return to main menu..."
}

# Show remote deployment instructions
function Show-RemoteInstructions {
    Write-Host ""
    Write-Host "Image build completed!" -ForegroundColor Green
    Write-Host "=========================================="
    Write-Host "Build artifacts location:" -ForegroundColor Blue
    Write-Host "  .\docker-images\"
    Write-Host ""
    Write-Host "Remote deployment steps:" -ForegroundColor Yellow
    Write-Host "1. Transfer image files to remote server:"
    Write-Host "   scp -r docker-images/ user@your-server:/path/to/deployment/"
    Write-Host ""
    Write-Host "2. Execute on remote server:"
    Write-Host "   cd docker-images/"
    Write-Host "   chmod +x load_images.sh"
    Write-Host "   ./load_images.sh"
    Write-Host ""
    Write-Host "3. Copy deployment script and execute:"
    Write-Host "   # Copy deploy_from_images.sh to server"
    Write-Host "   chmod +x deploy_from_images.sh"
    Write-Host "   ./deploy_from_images.sh"
    Write-Host ""
    Write-Host "Ready! You can start remote deployment" -ForegroundColor Green
    Write-Host ""
}

# Show local deployment result
function Show-LocalResult {
    Write-Host ""
    Write-Host "Local deployment completed!" -ForegroundColor Green
    Write-Host "=========================================="
    Write-Host "Service access URLs:" -ForegroundColor Blue
    Write-Host "  Frontend: http://localhost:3000"
    Write-Host "  Backend API: http://localhost:8080"
    Write-Host "  API Documentation: http://localhost:8080/swagger/index.html"
    Write-Host "  Health Check: http://localhost:8080/health"
    Write-Host "  Prometheus: http://localhost:9090"
    Write-Host "  Elasticsearch: http://localhost:9200"
    Write-Host ""
    Write-Host "Management commands:" -ForegroundColor Yellow
    Write-Host "  Check status: ./manage.sh status"
    Write-Host "  View logs: ./manage.sh logs"
    Write-Host "  Restart services: ./manage.sh restart"
    Write-Host "  Stop services: ./manage.sh stop"
    Write-Host ""
    Write-Host "Default login credentials:" -ForegroundColor Green
    Write-Host "  Username: admin"
    Write-Host "  Password: password"
    Write-Host ""
}

# Main function
function Main {
    try {
        Show-Welcome
        Test-Environment
        
        while ($true) {
            $choice = Show-Menu
            
            switch ($choice) {
                "1" {
                    Write-Host ""
                    Log-Info "Starting Docker image build..."
                    try {
                        Build-Images
                        Show-RemoteInstructions
                    }
                    catch {
                        Log-Error "Image build failed: $_"
                        continue
                    }
                }
                "2" {
                    Write-Host ""
                    Log-Info "Starting full local deployment..."
                    try {
                        Build-Images
                        Deploy-Local
                        Show-LocalResult
                    }
                    catch {
                        Log-Error "Local deployment failed: $_"
                        continue
                    }
                }
                "3" {
                    Write-Host ""
                    Log-Info "Starting service deployment..."
                    try {
                        Deploy-Only
                        Show-LocalResult
                    }
                    catch {
                        Log-Error "Service deployment failed: $_"
                        continue
                    }
                }
                "4" {
                    Show-Guide
                }
                "5" {
                    Write-Host ""
                    Log-Info "Thank you for using AI Monitor deployment tool!"
                    exit 0
                }
                default {
                    Write-Host ""
                    Log-Warning "Invalid option, please select again"
                }
            }
            
            Write-Host ""
            Read-Host "Press Enter to continue..."
            Write-Host ""
        }
    }
    catch {
        Log-Error "Error occurred during deployment: $_"
        Log-Error "Please check the log information above"
        exit 1
    }
}

# Execute main function
Main