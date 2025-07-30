package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// DownloadAgentPackage 下载Agent安装包（包含可执行文件、配置文件和安装脚本）
// @Summary 下载Agent完整安装包
// @Description 下载包含可执行文件、配置文件和安装脚本的完整Agent安装包
// @Tags Agent管理
// @Accept json
// @Produce application/zip
// @Param type path string true "Agent类型" Enums(windows,linux,redis,mysql,docker,kafka,apache,nginx,postgresql,elasticsearch,rabbitmq,hyperv,vmware,apm,service-discovery,api-key)
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/download/{type} [get]
func (h *AgentHandler) DownloadAgentPackage(c *gin.Context) {
	agentType := c.Param("type")

	if agentType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid parameters",
			Message: "Agent type is required",
		})
		return
	}

	// 根据Agent类型确定文件路径
	var buildDir string
	switch agentType {
	case "windows":
		buildDir = "./agents/build/windows"
	case "linux":
		buildDir = "./agents/build/linux"
	case "redis":
		buildDir = "./agents/build/redis"
	case "mysql":
		buildDir = "./agents/build/mysql"
	case "docker":
		buildDir = "./agents/build/docker"
	case "kafka":
		buildDir = "./agents/build/kafka"
	case "apache":
		buildDir = "./agents/build/apache"
	case "nginx":
		buildDir = "./agents/build/nginx"
	case "postgresql":
		buildDir = "./agents/build/postgresql"
	case "elasticsearch":
		buildDir = "./agents/build/elasticsearch"
	case "rabbitmq":
		buildDir = "./agents/build/rabbitmq"
	case "hyperv":
		buildDir = "./agents/build/hyperv"
	case "vmware":
		buildDir = "./agents/build/vmware"
	case "apm":
		buildDir = "./agents/build/apm"
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid agent type",
			Message: "Supported types: windows, linux, redis, mysql, docker, kafka, apache, nginx, postgresql, elasticsearch, rabbitmq, hyperv, vmware, apm, service-discovery, api-key",
		})
		return
	}

	// 检查目录是否存在
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Not Found",
			Message: "Agent package not found",
		})
		return
	}

	// 创建临时ZIP文件
	zipFileName := fmt.Sprintf("aimonitor-%s-agent.zip", agentType)
	tempZipPath := filepath.Join(os.TempDir(), zipFileName)

	// 创建ZIP文件
	zipFile, err := os.Create(tempZipPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create zip file",
		})
		return
	}
	defer zipFile.Close()
	defer os.Remove(tempZipPath) // 清理临时文件

	// 创建ZIP写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加文件到ZIP
	err = filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(buildDir, path)
		if err != nil {
			return err
		}

		// 创建ZIP文件条目
		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// 打开源文件
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// 复制文件内容到ZIP
		_, err = io.Copy(zipEntry, srcFile)
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create zip archive",
		})
		return
	}

	// 关闭ZIP写入器以确保所有数据都被写入
	zipWriter.Close()
	zipFile.Close()

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+zipFileName)
	c.Header("Content-Type", "application/zip")

	// 发送ZIP文件
	c.File(tempZipPath)
}

// GetAgentInstallGuide 获取Agent安装指南
// @Summary 获取Agent安装指南
// @Description 获取指定类型Agent的安装指南
// @Tags Agent管理
// @Accept json
// @Produce json
// @Param type path string true "Agent类型" Enums(windows,linux,redis,mysql,docker,kafka,apache,nginx,postgresql,elasticsearch,rabbitmq,hyperv,vmware,apm,service-discovery,api-key)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/agents/install-guide/{type} [get]
func (h *AgentHandler) GetAgentInstallGuide(c *gin.Context) {
	agentType := c.Param("type")

	if agentType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid parameters",
			Message: "Agent type is required",
		})
		return
	}

	// 根据Agent类型生成安装指南
	var guide map[string]interface{}
	switch strings.ToLower(agentType) {
	case "windows":
		guide = map[string]interface{}{
			"title":       "Windows系统监控Agent安装指南",
			"description": "用于监控Windows系统资源的Agent程序",
			"requirements": []string{
				"Windows 10/Server 2016 或更高版本",
				"管理员权限",
				"网络连接到监控服务器",
			},
			"installation_steps": []string{
				"1. 下载Agent安装包",
				"2. 解压到目标目录（如 C:\\AIMonitor\\Agent）",
				"3. 编辑 config.yaml 配置文件",
				"4. 以管理员身份运行 install.bat",
				"5. 验证服务是否正常启动",
			},
			"configuration": map[string]interface{}{
				"server_host": "监控服务器地址",
				"server_port": "监控服务器端口（默认8080）",
				"api_key":     "API密钥",
				"agent_name":  "Agent名称（唯一标识）",
			},
			"commands": map[string]string{
				"install":   "install.bat",
				"uninstall": "uninstall.bat",
				"start":     "sc start AIMonitorAgent",
				"stop":      "sc stop AIMonitorAgent",
				"status":    "sc query AIMonitorAgent",
			},
		}
	case "linux":
		guide = map[string]interface{}{
			"title":       "Linux系统监控Agent安装指南",
			"description": "用于监控Linux系统资源的Agent程序",
			"requirements": []string{
				"Linux发行版（Ubuntu 18.04+, CentOS 7+, RHEL 7+等）",
				"root权限",
				"网络连接到监控服务器",
			},
			"installation_steps": []string{
				"1. 下载Agent安装包",
				"2. 解压到目标目录（如 /opt/aimonitor/agent）",
				"3. 编辑 config.yaml 配置文件",
				"4. 以root身份运行 sudo ./install.sh",
				"5. 验证服务是否正常启动",
			},
			"configuration": map[string]interface{}{
				"server_host": "监控服务器地址",
				"server_port": "监控服务器端口（默认8080）",
				"api_key":     "API密钥",
				"agent_name":  "Agent名称（唯一标识）",
			},
			"commands": map[string]string{
				"install":   "sudo ./install.sh",
				"uninstall": "sudo ./uninstall.sh",
				"start":     "sudo systemctl start aimonitor-agent",
				"stop":      "sudo systemctl stop aimonitor-agent",
				"status":    "sudo systemctl status aimonitor-agent",
				"logs":      "sudo journalctl -u aimonitor-agent -f",
			},
		}
	case "redis":
		guide = map[string]interface{}{
			"title":       "Redis监控Agent安装指南",
			"description": "用于监控Redis实例的专用Agent程序",
			"requirements": []string{
				"Redis服务器访问权限",
				"网络连接到Redis服务器和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载Redis Agent安装包",
				"2. 解压到目标目录",
				"3. 编辑 config.yaml 配置Redis连接信息",
				"4. 运行Agent程序",
				"5. 验证Redis监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"redis_host":     "Redis服务器地址",
				"redis_port":     "Redis端口（默认6379）",
				"redis_password": "Redis密码（如果有）",
				"redis_db":       "Redis数据库编号（默认0）",
				"server_host":    "监控服务器地址",
				"server_port":    "监控服务器端口",
				"api_key":        "API密钥",
			},
		}
	case "mysql":
		guide = map[string]interface{}{
			"title":       "MySQL监控Agent安装指南",
			"description": "用于监控MySQL数据库的专用Agent程序",
			"requirements": []string{
				"MySQL服务器访问权限",
				"具有监控权限的MySQL用户账号",
				"网络连接到MySQL服务器和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载MySQL Agent安装包",
				"2. 解压到目标目录",
				"3. 在MySQL中创建监控用户",
				"4. 编辑 config.yaml 配置MySQL连接信息",
				"5. 运行Agent程序",
				"6. 验证MySQL监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"mysql_host":     "MySQL服务器地址",
				"mysql_port":     "MySQL端口（默认3306）",
				"mysql_username": "MySQL用户名",
				"mysql_password": "MySQL密码",
				"mysql_database": "监控的数据库名",
				"server_host":    "监控服务器地址",
				"server_port":    "监控服务器端口",
				"api_key":        "API密钥",
			},
			"mysql_user_setup": []string{
				"CREATE USER 'monitor'@'%' IDENTIFIED BY 'password';",
				"GRANT PROCESS, REPLICATION CLIENT, SELECT ON *.* TO 'monitor'@'%';",
				"FLUSH PRIVILEGES;",
			},
		}
	case "docker":
		guide = map[string]interface{}{
			"title":       "Docker监控Agent安装指南",
			"description": "用于监控Docker容器和镜像的专用Agent程序",
			"requirements": []string{
				"Docker服务运行中",
				"Docker API访问权限",
				"网络连接到监控服务器",
			},
			"installation_steps": []string{
				"1. 下载Docker Agent安装包",
				"2. 解压到目标目录",
				"3. 编辑 config.yaml 配置Docker连接信息",
				"4. 运行Agent程序",
				"5. 验证Docker监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"docker_endpoint": "Docker API端点（如 unix:///var/run/docker.sock）",
				"docker_version":  "Docker API版本",
				"server_host":     "监控服务器地址",
				"server_port":     "监控服务器端口",
				"api_key":         "API密钥",
			},
		}
	case "kafka":
		guide = map[string]interface{}{
			"title":       "Kafka监控Agent安装指南",
			"description": "用于监控Apache Kafka集群的专用Agent程序",
			"requirements": []string{
				"Kafka集群访问权限",
				"JMX端口开放（默认9999）",
				"网络连接到Kafka集群和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载Kafka Agent安装包",
				"2. 解压到目标目录",
				"3. 编辑 config.yaml 配置Kafka连接信息",
				"4. 确保Kafka JMX端口开放",
				"5. 运行Agent程序",
				"6. 验证Kafka监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"kafka_brokers":    "Kafka Broker地址列表（逗号分隔）",
				"kafka_jmx_port":   "JMX端口（默认9999）",
				"kafka_version":    "Kafka版本",
				"server_host":      "监控服务器地址",
				"server_port":      "监控服务器端口",
				"api_key":          "API密钥",
			},
		}
	case "apache":
		guide = map[string]interface{}{
			"title":       "Apache监控Agent安装指南",
			"description": "用于监控Apache HTTP服务器的专用Agent程序",
			"requirements": []string{
				"Apache HTTP服务器运行中",
				"mod_status模块已启用",
				"网络连接到Apache服务器和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载Apache Agent安装包",
				"2. 解压到目标目录",
				"3. 启用Apache mod_status模块",
				"4. 编辑 config.yaml 配置Apache连接信息",
				"5. 运行Agent程序",
				"6. 验证Apache监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"apache_host":     "Apache服务器地址",
				"apache_port":     "Apache端口（默认80）",
				"status_url":      "状态页面URL（如/server-status）",
				"server_host":     "监控服务器地址",
				"server_port":     "监控服务器端口",
				"api_key":         "API密钥",
			},
			"apache_setup": []string{
				"LoadModule status_module modules/mod_status.so",
				"<Location \"/server-status\">",
				"    SetHandler server-status",
				"    Require local",
				"</Location>",
			},
		}
	case "nginx":
		guide = map[string]interface{}{
			"title":       "Nginx监控Agent安装指南",
			"description": "用于监控Nginx Web服务器的专用Agent程序",
			"requirements": []string{
				"Nginx服务器运行中",
				"nginx-module-vts或stub_status模块已启用",
				"网络连接到Nginx服务器和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载Nginx Agent安装包",
				"2. 解压到目标目录",
				"3. 配置Nginx状态页面",
				"4. 编辑 config.yaml 配置Nginx连接信息",
				"5. 运行Agent程序",
				"6. 验证Nginx监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"nginx_host":      "Nginx服务器地址",
				"nginx_port":      "Nginx端口（默认80）",
				"status_url":      "状态页面URL（如/nginx_status）",
				"server_host":     "监控服务器地址",
				"server_port":     "监控服务器端口",
				"api_key":         "API密钥",
			},
			"nginx_setup": []string{
				"location /nginx_status {",
				"    stub_status on;",
				"    access_log off;",
				"    allow 127.0.0.1;",
				"    deny all;",
				"}",
			},
		}
	case "postgresql":
		guide = map[string]interface{}{
			"title":       "PostgreSQL监控Agent安装指南",
			"description": "用于监控PostgreSQL数据库的专用Agent程序",
			"requirements": []string{
				"PostgreSQL服务器访问权限",
				"具有监控权限的PostgreSQL用户账号",
				"网络连接到PostgreSQL服务器和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载PostgreSQL Agent安装包",
				"2. 解压到目标目录",
				"3. 在PostgreSQL中创建监控用户",
				"4. 编辑 config.yaml 配置PostgreSQL连接信息",
				"5. 运行Agent程序",
				"6. 验证PostgreSQL监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"postgres_host":     "PostgreSQL服务器地址",
				"postgres_port":     "PostgreSQL端口（默认5432）",
				"postgres_username": "PostgreSQL用户名",
				"postgres_password": "PostgreSQL密码",
				"postgres_database": "监控的数据库名",
				"server_host":       "监控服务器地址",
				"server_port":       "监控服务器端口",
				"api_key":           "API密钥",
			},
			"postgres_user_setup": []string{
				"CREATE USER monitor WITH PASSWORD 'password';",
				"GRANT CONNECT ON DATABASE postgres TO monitor;",
				"GRANT USAGE ON SCHEMA public TO monitor;",
				"GRANT SELECT ON ALL TABLES IN SCHEMA public TO monitor;",
			},
		}
	case "elasticsearch":
		guide = map[string]interface{}{
			"title":       "Elasticsearch监控Agent安装指南",
			"description": "用于监控Elasticsearch集群的专用Agent程序",
			"requirements": []string{
				"Elasticsearch集群访问权限",
				"REST API端口开放（默认9200）",
				"网络连接到Elasticsearch集群和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载Elasticsearch Agent安装包",
				"2. 解压到目标目录",
				"3. 编辑 config.yaml 配置Elasticsearch连接信息",
				"4. 配置认证信息（如果启用了安全功能）",
				"5. 运行Agent程序",
				"6. 验证Elasticsearch监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"elasticsearch_hosts": "Elasticsearch节点地址列表",
				"elasticsearch_port":  "Elasticsearch端口（默认9200）",
				"username":            "用户名（如果启用了安全功能）",
				"password":            "密码（如果启用了安全功能）",
				"use_ssl":             "是否使用SSL连接",
				"server_host":         "监控服务器地址",
				"server_port":         "监控服务器端口",
				"api_key":             "API密钥",
			},
		}
	case "rabbitmq":
		guide = map[string]interface{}{
			"title":       "RabbitMQ监控Agent安装指南",
			"description": "用于监控RabbitMQ消息队列的专用Agent程序",
			"requirements": []string{
				"RabbitMQ服务器运行中",
				"Management插件已启用",
				"网络连接到RabbitMQ服务器和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载RabbitMQ Agent安装包",
				"2. 解压到目标目录",
				"3. 启用RabbitMQ Management插件",
				"4. 编辑 config.yaml 配置RabbitMQ连接信息",
				"5. 运行Agent程序",
				"6. 验证RabbitMQ监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"rabbitmq_host":     "RabbitMQ服务器地址",
				"rabbitmq_port":     "RabbitMQ管理端口（默认15672）",
				"rabbitmq_username": "RabbitMQ用户名",
				"rabbitmq_password": "RabbitMQ密码",
				"server_host":       "监控服务器地址",
				"server_port":       "监控服务器端口",
				"api_key":           "API密钥",
			},
			"rabbitmq_setup": []string{
				"rabbitmq-plugins enable rabbitmq_management",
				"systemctl restart rabbitmq-server",
			},
		}
	case "hyperv":
		guide = map[string]interface{}{
			"title":       "Hyper-V监控Agent安装指南",
			"description": "用于监控Microsoft Hyper-V虚拟化平台的专用Agent程序",
			"requirements": []string{
				"Windows Server with Hyper-V角色",
				"管理员权限",
				"PowerShell执行权限",
				"网络连接到监控服务器",
			},
			"installation_steps": []string{
				"1. 下载Hyper-V Agent安装包",
				"2. 解压到目标目录",
				"3. 编辑 config.yaml 配置连接信息",
				"4. 以管理员身份运行Agent程序",
				"5. 验证Hyper-V监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"hyperv_host":   "Hyper-V主机地址",
				"server_host":   "监控服务器地址",
				"server_port":   "监控服务器端口",
				"api_key":       "API密钥",
				"agent_name":    "Agent名称",
			},
		}
	case "vmware":
		guide = map[string]interface{}{
			"title":       "VMware监控Agent安装指南",
			"description": "用于监控VMware vSphere环境的专用Agent程序",
			"requirements": []string{
				"VMware vCenter Server或ESXi主机访问权限",
				"具有监控权限的VMware用户账号",
				"网络连接到VMware环境和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载VMware Agent安装包",
				"2. 解压到目标目录",
				"3. 编辑 config.yaml 配置VMware连接信息",
				"4. 运行Agent程序",
				"5. 验证VMware监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"vcenter_host":     "vCenter Server地址",
				"vcenter_username": "VMware用户名",
				"vcenter_password": "VMware密码",
				"insecure_ssl":     "是否忽略SSL证书验证",
				"server_host":      "监控服务器地址",
				"server_port":      "监控服务器端口",
				"api_key":          "API密钥",
			},
		}
	case "apm":
		guide = map[string]interface{}{
			"title":       "APM监控Agent安装指南",
			"description": "用于应用性能监控的专用Agent程序",
			"requirements": []string{
				"目标应用程序运行中",
				"应用程序支持APM集成",
				"网络连接到应用服务器和监控服务器",
			},
			"installation_steps": []string{
				"1. 下载APM Agent安装包",
				"2. 解压到目标目录",
				"3. 编辑 config.yaml 配置APM连接信息",
				"4. 集成到目标应用程序",
				"5. 重启应用程序",
				"6. 验证APM监控数据是否正常上报",
			},
			"configuration": map[string]interface{}{
				"app_name":     "应用程序名称",
				"app_version":  "应用程序版本",
				"environment":  "运行环境（dev/test/prod）",
				"server_host":  "监控服务器地址",
				"server_port":  "监控服务器端口",
				"api_key":      "API密钥",
			},
		}
	case "service-discovery":
		guide = map[string]interface{}{
			"title":       "服务发现功能配置指南",
			"description": "配置AI Monitor系统的网络扫描和自动发现功能，主动发现网络中的设备和服务并建立监控连接",
			"requirements": []string{
				"AI Monitor系统已部署并运行",
				"管理员权限",
				"网络连接正常",
				"目标网络段可访问",
				"必要的网络端口开放（SSH、SNMP、WMI等）",
			},
			"discovery_types": map[string]interface{}{
				"network_scan": "网络扫描 - 通过IP段扫描发现在线设备",
				"port_scan": "端口扫描 - 检测设备开放的服务端口",
				"snmp_discovery": "SNMP发现 - 通过SNMP协议获取设备信息",
				"ssh_discovery": "SSH发现 - 通过SSH连接发现Linux/Unix系统",
				"wmi_discovery": "WMI发现 - 通过WMI协议发现Windows系统",
				"agent_discovery": "Agent发现 - 发现已安装但未注册的Agent",
			},
			"discovery_workflow": "AI Monitor主动扫描网络 → 发现设备和服务 → 自动建立监控连接 → 开始数据采集",
			"installation_steps": []string{
				"1. 确认AI Monitor系统运行状态",
				"   - 访问 http://your-server:8080/health 检查系统健康状态",
				"   - 确认数据库连接正常",
				"2. 登录AI Monitor管理界面",
				"   - 使用管理员账号登录 http://your-server:8080",
				"   - 进入系统设置 → 服务发现配置页面",
				"3. 配置网络扫描范围",
				"   - 设置目标IP段（如：192.168.1.0/24）",
				"   - 配置扫描端口范围（如：22,80,443,3389,161）",
				"   - 设置扫描间隔（建议5-30分钟）",
				"4. 配置发现协议",
				"   - SNMP：设置Community字符串",
				"   - SSH：配置认证凭据",
				"   - WMI：设置Windows认证信息",
				"5. 启用自动发现功能",
				"   - 点击'启用网络扫描'按钮",
				"   - 确认配置信息并保存",
				"6. 配置发现规则和过滤条件",
				"   - 设置设备类型过滤（服务器、网络设备、数据库等）",
				"   - 配置自动监控策略",
				"   - 设置告警阈值",
				"7. 测试网络扫描功能",
				"   - 执行手动扫描测试",
				"   - 检查扫描日志",
				"8. 查看发现的设备列表并确认监控状态",
				"   - 查看自动发现的设备和服务",
				"   - 确认监控数据正常采集",
				"   - 验证告警功能",
			},
			"configuration": map[string]interface{}{
				"scan_networks": "扫描网络段，支持CIDR格式，如：192.168.1.0/24,10.0.0.0/16",
				"scan_ports": "扫描端口列表，如：22,80,443,3389,161,8080-8090",
				"scan_interval": "扫描间隔，单位：分钟，建议值：5-30",
				"discovery_timeout": "发现超时时间，单位：秒，建议值：5-10",
				"snmp_community": "SNMP Community字符串，默认：public",
				"ssh_credentials": "SSH认证信息，格式：username:password或使用密钥",
				"wmi_credentials": "WMI认证信息，格式：domain\\username:password",
				"auto_monitor": "是否自动添加发现的设备到监控，true/false",
				"device_filters": "设备过滤规则，支持IP范围、设备类型等条件",
			},
			"web_ui_steps": []string{
				"1. 访问 http://your-server:8080/settings/network-discovery",
				"2. 点击'配置网络扫描'按钮",
				"3. 设置扫描网络段和端口范围",
				"4. 配置发现协议和认证信息",
				"5. 启用自动网络扫描功能",
				"6. 执行手动扫描测试",
				"7. 在设备列表中查看自动发现的设备",
			},
			"api_examples": map[string]string{
				"start_scan":        "POST /api/v1/discovery/scan\n{\"networks\": [\"192.168.1.0/24\"], \"ports\": [22, 80, 443, 3389, 161]}",
				"list_devices":      "GET /api/v1/discovery/devices",
				"scan_status":       "GET /api/v1/discovery/scan/status",
				"add_to_monitor":    "POST /api/v1/discovery/monitor\n{\"device_id\": \"device_123\", \"monitor_type\": \"snmp\"}",
				"scan_config":       "PUT /api/v1/discovery/config\n{\"scan_interval\": 15, \"auto_monitor\": true}",
			},
			"troubleshooting": []string{
				"问题：网络扫描无法发现设备",
				"解决：检查网络连通性、防火墙设置和IP段配置",
				"问题：SNMP发现失败",
				"解决：验证SNMP Community字符串和目标设备SNMP配置",
				"问题：SSH连接认证失败",
				"解决：检查SSH凭据、端口开放状态和目标系统SSH服务",
				"问题：WMI发现Windows设备失败",
				"解决：确认WMI服务运行、防火墙规则和认证权限",
				"问题：扫描速度过慢",
				"解决：调整扫描间隔、减少端口范围或优化网络配置",
				"问题：发现的设备信息不完整",
				"解决：检查设备协议支持情况和认证配置",
			},
		}
	case "api-key":
		guide = map[string]interface{}{
			"title":       "API密钥管理指南",
			"description": "管理AI Monitor系统的API密钥，包括生成、更新、删除和权限控制",
			"requirements": []string{
				"AI Monitor系统已部署并运行",
				"管理员账号权限",
				"网络连接正常",
				"HTTPS连接（生产环境推荐）",
			},
			"key_types": map[string]interface{}{
				"agent_key": "Agent密钥 - 用于Agent与服务器通信认证",
				"api_key": "API密钥 - 用于第三方应用调用API接口",
				"webhook_key": "Webhook密钥 - 用于接收外部系统回调",
				"readonly_key": "只读密钥 - 仅允许查询操作，不能修改数据",
			},
			"installation_steps": []string{
				"1. 创建API密钥",
				"   - 登录AI Monitor Web管理界面",
				"   - 进入'系统设置' -> 'API密钥管理'",
				"   - 点击'创建新密钥'按钮",
				"   - 填写密钥名称和描述信息",
				"   - 选择密钥类型和过期时间",
				"   - 系统自动生成安全的密钥字符串",
				"2. 配置密钥权限",
				"   - 为密钥分配适当的权限范围",
				"   - Agent密钥：数据上报、心跳检测权限",
				"   - API密钥：接口调用、数据查询权限",
				"   - 管理密钥：完整的系统管理权限",
				"   - 遵循最小权限原则",
				"3. 分发密钥",
				"   - 将密钥安全地传输给目标系统",
				"   - 避免通过明文邮件或聊天工具传输",
				"   - 建议使用加密文件或安全的密钥管理工具",
				"   - 记录密钥分发日志",
				"4. 配置Agent",
				"   - 在Agent配置文件中设置API密钥",
				"   - 配置服务器地址和端口",
				"   - 设置数据上报间隔和重试策略",
				"   - 验证配置文件格式正确性",
				"5. 验证连接",
				"   - 启动Agent服务",
				"   - 检查Agent日志确认连接成功",
				"   - 在Web界面查看Agent在线状态",
				"   - 验证数据上报功能正常",
				"6. 监控密钥使用",
				"   - 定期检查密钥使用情况",
				"   - 监控异常的API调用模式",
				"   - 设置密钥使用告警规则",
				"   - 记录和审计密钥访问日志",
				"7. 定期轮换",
				"   - 制定密钥轮换策略（建议90天）",
				"   - 提前通知相关人员密钥更新计划",
				"   - 创建新密钥并更新Agent配置",
				"   - 验证新密钥工作正常后删除旧密钥",
				"8. 备份恢复",
				"   - 定期备份密钥配置信息",
				"   - 建立密钥恢复流程",
				"   - 测试备份数据的完整性",
				"   - 制定密钥泄露应急响应预案",
			},
			"configuration": map[string]interface{}{
				"key_name":        "密钥名称（便于识别）",
				"permissions":     "权限范围（read/write/admin）",
				"expiry_date":     "过期时间（可选）",
				"ip_whitelist":    "IP白名单（可选）",
				"rate_limit":      "请求频率限制",
			},
			"web_ui_steps": []string{
				"1. 访问 http://your-server:8080/settings/api-keys",
				"2. 点击'新增API密钥'按钮",
				"3. 填写密钥基本信息和权限设置",
				"4. 点击'生成密钥'按钮",
				"5. 复制生成的密钥并妥善保存",
				"6. 在密钥列表中管理现有密钥",
			},
			"api_examples": map[string]string{
				"create_key":   "POST /api/v1/auth/api-keys",
				"list_keys":    "GET /api/v1/auth/api-keys",
				"update_key":   "PUT /api/v1/auth/api-keys/{id}",
				"delete_key":   "DELETE /api/v1/auth/api-keys/{id}",
				"validate_key": "POST /api/v1/auth/validate",
			},
			"security_notes": []string{
				"密钥生成后请立即保存，系统不会再次显示完整密钥",
				"定期轮换API密钥以提高安全性",
				"为不同用途创建不同权限的密钥",
				"监控密钥使用情况，及时发现异常访问",
				"删除不再使用的密钥",
			},
			"troubleshooting": []string{
				"检查密钥是否已过期",
				"验证密钥权限是否足够",
				"确认IP地址在白名单中",
				"检查请求频率是否超出限制",
				"查看API访问日志获取详细信息",
			},
		}
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid agent type",
			Message: "Supported types: windows, linux, redis, mysql, docker, kafka, apache, nginx, postgresql, elasticsearch, rabbitmq, hyperv, vmware, apm, service-discovery, api-key",
		})
		return
	}

	c.JSON(http.StatusOK, guide)
}