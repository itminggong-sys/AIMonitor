#!/bin/bash

echo "Installing AIMonitor Linux Agent..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "This script requires root privileges."
    echo "Please run with sudo: sudo ./install.sh"
    exit 1
fi

# Create installation directory
INSTALL_DIR="/opt/aimonitor/agent"
mkdir -p "$INSTALL_DIR"

# Create logs directory
LOG_DIR="/var/log"
mkdir -p "$LOG_DIR"

# Copy files
echo "Copying agent files..."
cp aimonitor-linux-agent "$INSTALL_DIR/"
cp config.yaml "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/aimonitor-linux-agent"

# Create systemd service file
echo "Creating systemd service..."
cat > /etc/systemd/system/aimonitor-agent.service << EOF
[Unit]
Description=AIMonitor Agent Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/aimonitor-linux-agent -config $INSTALL_DIR/config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and enable service
echo "Enabling service..."
systemctl daemon-reload
systemctl enable aimonitor-agent.service

# Start service
echo "Starting service..."
systemctl start aimonitor-agent.service

if systemctl is-active --quiet aimonitor-agent.service; then
    echo "Service started successfully"
else
    echo "Failed to start service. Check logs with: journalctl -u aimonitor-agent.service"
fi

# Create uninstall script
echo "Creating uninstall script..."
cat > "$INSTALL_DIR/uninstall.sh" << EOF
#!/bin/bash
echo "Stopping AIMonitor Agent service..."
systemctl stop aimonitor-agent.service
echo "Disabling service..."
systemctl disable aimonitor-agent.service
echo "Removing service file..."
rm -f /etc/systemd/system/aimonitor-agent.service
systemctl daemon-reload
echo "Removing files..."
rm -rf "$INSTALL_DIR"
echo "Uninstall completed."
EOF

chmod +x "$INSTALL_DIR/uninstall.sh"

echo ""
echo "Installation completed!"
echo ""
echo "Agent installed to: $INSTALL_DIR"
echo "Logs will be written to: $LOG_DIR/aimonitor-agent.log"
echo ""
echo "To configure the agent, edit: $INSTALL_DIR/config.yaml"
echo "To uninstall, run: $INSTALL_DIR/uninstall.sh"
echo ""
echo "Service commands:"
echo "  Start:   systemctl start aimonitor-agent.service"
echo "  Stop:    systemctl stop aimonitor-agent.service"
echo "  Status:  systemctl status aimonitor-agent.service"
echo "  Logs:    journalctl -u aimonitor-agent.service -f"
echo ""
echo "Please update the config.yaml file with your AIMonitor server details."
echo "Then restart the service: systemctl restart aimonitor-agent.service"
echo ""