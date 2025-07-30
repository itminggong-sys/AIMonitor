@echo off
echo Installing AIMonitor Windows Agent...

:: Check if running as administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo This script requires administrator privileges.
    echo Please run as administrator.
    pause
    exit /b 1
)

:: Create installation directory
set INSTALL_DIR=C:\Program Files\AIMonitor\Agent
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

:: Create logs directory
set LOG_DIR=C:\logs
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

:: Copy files
echo Copying agent files...
copy aimonitor-windows-agent.exe "%INSTALL_DIR%\"
copy config.yaml "%INSTALL_DIR%\"

:: Create service wrapper script
echo Creating service wrapper...
(
echo @echo off
echo cd /d "%INSTALL_DIR%"
echo aimonitor-windows-agent.exe -config config.yaml
) > "%INSTALL_DIR%\start-agent.bat"

:: Install as Windows service using sc command
echo Installing Windows service...
sc create "AIMonitorAgent" binPath= "\"%INSTALL_DIR%\start-agent.bat\"" start= auto DisplayName= "AIMonitor Agent Service"
if %errorlevel% equ 0 (
    echo Service installed successfully
    echo Starting service...
    sc start "AIMonitorAgent"
    if %errorlevel% equ 0 (
        echo Service started successfully
    ) else (
        echo Failed to start service. You can start it manually from Services.msc
    )
) else (
    echo Failed to install service
)

:: Create uninstall script
echo Creating uninstall script...
(
echo @echo off
echo echo Stopping AIMonitor Agent service...
echo sc stop "AIMonitorAgent"
echo echo Removing service...
echo sc delete "AIMonitorAgent"
echo echo Removing files...
echo rmdir /s /q "%INSTALL_DIR%"
echo echo Uninstall completed.
echo pause
) > "%INSTALL_DIR%\uninstall.bat"

echo.
echo Installation completed!
echo.
echo Agent installed to: %INSTALL_DIR%
echo Logs will be written to: %LOG_DIR%\aimonitor-agent.log
echo.
echo To configure the agent, edit: %INSTALL_DIR%\config.yaml
echo To uninstall, run: %INSTALL_DIR%\uninstall.bat
echo.
echo Please update the config.yaml file with your AIMonitor server details before starting the service.
echo.
pause