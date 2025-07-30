@echo off
echo Building AIMonitor Agents...

:: Create build directory
if not exist "build" mkdir build
if not exist "build\windows" mkdir build\windows
if not exist "build\linux" mkdir build\linux
if not exist "build\redis" mkdir build\redis
if not exist "build\mysql" mkdir build\mysql
if not exist "build\docker" mkdir build\docker

:: Initialize go modules
echo Initializing Go modules...
go mod tidy

:: Build Windows Agent
echo Building Windows Agent...
go build -o build\windows\aimonitor-windows-agent.exe .\windows\main.go
if %errorlevel% equ 0 (
    echo Windows Agent built successfully
    copy windows\config.yaml build\windows\
    copy windows\install.bat build\windows\
) else (
    echo Failed to build Windows Agent
)

:: Build Linux Agent
echo Building Linux Agent...
set GOOS=linux
set GOARCH=amd64
go build -o build\linux\aimonitor-linux-agent .\linux\main.go
if %errorlevel% equ 0 (
    echo Linux Agent built successfully
    copy linux\config.yaml build\linux\
    copy linux\install.sh build\linux\
) else (
    echo Failed to build Linux Agent
)
set GOOS=
set GOARCH=

:: Build Redis Agent
echo Building Redis Agent...
go build -o build\redis\aimonitor-redis-agent.exe .\redis\main.go
if %errorlevel% equ 0 (
    echo Redis Agent built successfully
    copy redis\config.yaml build\redis\
) else (
    echo Failed to build Redis Agent
)

:: Build MySQL Agent
echo Building MySQL Agent...
go build -o build\mysql\aimonitor-mysql-agent.exe .\mysql\main.go
if %errorlevel% equ 0 (
    echo MySQL Agent built successfully
    copy mysql\config.yaml build\mysql\
) else (
    echo Failed to build MySQL Agent
)

:: Build Docker Agent
echo Building Docker Agent...
go build -o build\docker\aimonitor-docker-agent.exe .\docker\main.go
if %errorlevel% equ 0 (
    echo Docker Agent built successfully
    copy docker\config.yaml build\docker\
) else (
    echo Failed to build Docker Agent
)

echo.
echo Build completed! Check the build directory for executables.
echo.
pause