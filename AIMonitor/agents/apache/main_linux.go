//go:build linux
// +build linux

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"aimonitor-agents/common"
)

// LinuxApacheMonitor Linux版本的Apache监控器
type LinuxApacheMonitor struct {
	*ApacheMonitor
	statusURL     string
	infoURL       string
	configPath    string
	accessLogPath string
	errorLogPath  string
	pidFile       string
	ctl           string // apache控制命令
}

// NewLinuxApacheMonitor 创建Linux版本的Apache监控器
func NewLinuxApacheMonitor(agent *common.Agent) *LinuxApacheMonitor {
	baseMonitor := NewApacheMonitor(agent)
	return &LinuxApacheMonitor{
		ApacheMonitor: baseMonitor,
		statusURL:     "http://localhost/server-status?auto",
		infoURL:       "http://localhost/server-info",
		configPath:    "/etc/apache2/apache2.conf",
		accessLogPath: "/var/log/apache2/access.log",
		errorLogPath:  "/var/log/apache2/error.log",
		pidFile:       "/var/run/apache2/apache2.pid",
		ctl:           "apache2ctl",
	}
}

// getServerStatus 获取真实的服务器状态
func (m *LinuxApacheMonitor) getServerStatus() (map[string]interface{}, error) {
	// 检查Apache进程是否运行
	isRunning := m.isApacheRunning()
	result := map[string]interface{}{
		"status": "stopped",
		"uptime": 0,
	}

	if !isRunning {
		return result, nil
	}

	result["status"] = "running"

	// 尝试从server-status获取详细信息
	if statusData := m.getServerStatusData(); statusData != nil {
		for k, v := range statusData {
			result[k] = v
		}
	}

	// 获取Apache版本
	if version := m.getApacheVersion(); version != "" {
		result["version"] = version
	}

	// 获取进程启动时间
	if uptime := m.getApacheUptime(); uptime > 0 {
		result["uptime"] = uptime
	}

	return result, nil
}

// isApacheRunning 检查Apache是否运行
func (m *LinuxApacheMonitor) isApacheRunning() bool {
	// 方法1: 检查PID文件
	if _, err := os.Stat(m.pidFile); err == nil {
		if pidData, err := ioutil.ReadFile(m.pidFile); err == nil {
			pid := strings.TrimSpace(string(pidData))
			if pid != "" {
				// 检查进程是否存在
				cmd := exec.Command("kill", "-0", pid)
				if err := cmd.Run(); err == nil {
					return true
				}
			}
		}
	}

	// 方法2: 使用pgrep查找apache进程
	cmd := exec.Command("pgrep", "apache2")
	if err := cmd.Run(); err == nil {
		return true
	}

	// 方法3: 尝试httpd
	cmd = exec.Command("pgrep", "httpd")
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

// getServerStatusData 从server-status获取数据
func (m *LinuxApacheMonitor) getServerStatusData() map[string]interface{} {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(m.statusURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	// 解析server-status输出
	lines := strings.Split(string(body), "\n")
	result := make(map[string]interface{})

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析键值对
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Total Accesses":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["total_accesses"] = val
			}
		case "Total kBytes":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				result["total_kbytes"] = val
				result["total_bytes"] = val * 1024
			}
		case "CPULoad":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				result["cpu_load"] = val
			}
		case "Uptime":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["uptime"] = val
			}
		case "ReqPerSec":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				result["requests_per_sec"] = val
			}
		case "BytesPerSec":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				result["bytes_per_sec"] = val
			}
		case "BytesPerReq":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				result["bytes_per_req"] = val
			}
		case "BusyWorkers":
			if val, err := strconv.Atoi(value); err == nil {
				result["busy_workers"] = val
			}
		case "IdleWorkers":
			if val, err := strconv.Atoi(value); err == nil {
				result["idle_workers"] = val
			}
		case "Scoreboard":
			// 解析记分板
			scoreboard := m.parseScoreboard(value)
			for k, v := range scoreboard {
				result[k] = v
			}
		}
	}

	return result
}

// parseScoreboard 解析Apache记分板
func (m *LinuxApacheMonitor) parseScoreboard(scoreboard string) map[string]interface{} {
	result := make(map[string]interface{})
	counts := make(map[rune]int)

	for _, char := range scoreboard {
		counts[char]++
	}

	// Apache记分板字符含义
	result["waiting_for_connection"] = counts['_']
	result["starting_up"] = counts['S']
	result["reading_request"] = counts['R']
	result["sending_reply"] = counts['W']
	result["keepalive"] = counts['K']
	result["dns_lookup"] = counts['D']
	result["closing_connection"] = counts['C']
	result["logging"] = counts['L']
	result["gracefully_finishing"] = counts['G']
	result["idle_cleanup"] = counts['I']
	result["open_slot"] = counts['.']

	return result
}

// getApacheVersion 获取Apache版本
func (m *LinuxApacheMonitor) getApacheVersion() string {
	cmd := exec.Command(m.ctl, "-v")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// 解析版本信息
	versionRegex := regexp.MustCompile(`Apache/([\d\.]+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getApacheUptime 获取Apache运行时间
func (m *LinuxApacheMonitor) getApacheUptime() int64 {
	if pidData, err := ioutil.ReadFile(m.pidFile); err == nil {
		pid := strings.TrimSpace(string(pidData))
		if pid != "" {
			// 读取进程启动时间
			statFile := fmt.Sprintf("/proc/%s/stat", pid)
			if statData, err := ioutil.ReadFile(statFile); err == nil {
				fields := strings.Fields(string(statData))
				if len(fields) > 21 {
					if starttime, err := strconv.ParseInt(fields[21], 10, 64); err == nil {
						// 获取系统启动时间
						if uptimeData, err := ioutil.ReadFile("/proc/uptime"); err == nil {
							uptimeStr := strings.Fields(string(uptimeData))[0]
							if systemUptime, err := strconv.ParseFloat(uptimeStr, 64); err == nil {
								// 计算进程运行时间
								clockTicks := int64(100) // 通常是100 Hz
								processUptime := systemUptime - float64(starttime)/float64(clockTicks)
								return int64(processUptime)
							}
						}
					}
				}
			}
		}
	}
	return 0
}

// getWorkerInfo 获取真实的工作进程信息
func (m *LinuxApacheMonitor) getWorkerInfo() (map[string]interface{}, error) {
	// 获取所有Apache进程
	cmd := exec.Command("pgrep", "-f", "apache2|httpd")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	pids := strings.Fields(string(output))
	totalProcesses := len(pids)
	workerProcesses := 0
	totalMemory := int64(0)

	for _, pid := range pids {
		// 检查是否为worker进程
		cmdlineFile := fmt.Sprintf("/proc/%s/cmdline", pid)
		if cmdlineData, err := ioutil.ReadFile(cmdlineFile); err == nil {
			cmdline := string(cmdlineData)
			if !strings.Contains(cmdline, "apache2") && !strings.Contains(cmdline, "httpd") {
				continue
			}
			workerProcesses++
		}

		// 获取内存使用
		statmFile := fmt.Sprintf("/proc/%s/statm", pid)
		if statmData, err := ioutil.ReadFile(statmFile); err == nil {
			fields := strings.Fields(string(statmData))
			if len(fields) > 1 {
				if rss, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					// RSS in pages, convert to bytes (assuming 4KB pages)
					totalMemory += rss * 4096
				}
			}
		}
	}

	return map[string]interface{}{
		"total_processes":    totalProcesses,
		"worker_processes":   workerProcesses,
		"memory_usage_bytes": totalMemory,
		"memory_usage_mb":    totalMemory / 1024 / 1024,
	}, nil
}

// getConnectionInfo 获取真实的连接信息
func (m *LinuxApacheMonitor) getConnectionInfo() (map[string]interface{}, error) {
	// 从server-status获取连接信息
	if statusData := m.getServerStatusData(); statusData != nil {
		result := make(map[string]interface{})
		
		// 从记分板数据计算连接信息
		if busyWorkers, ok := statusData["busy_workers"]; ok {
			result["active_connections"] = busyWorkers
		}
		if idleWorkers, ok := statusData["idle_workers"]; ok {
			result["idle_connections"] = idleWorkers
		}
		
		// 计算总连接数
		if active, ok := result["active_connections"].(int); ok {
			if idle, ok := result["idle_connections"].(int); ok {
				result["total_connections"] = active + idle
			}
		}
		
		return result, nil
	}

	// 回退方案：使用netstat统计连接
	return m.getConnectionsFromNetstat()
}

// getConnectionsFromNetstat 从netstat获取连接信息
func (m *LinuxApacheMonitor) getConnectionsFromNetstat() (map[string]interface{}, error) {
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	established := 0
	listening := 0
	timeWait := 0

	for _, line := range lines {
		if strings.Contains(line, ":80 ") || strings.Contains(line, ":443 ") {
			if strings.Contains(line, "ESTABLISHED") {
				established++
			} else if strings.Contains(line, "LISTEN") {
				listening++
			} else if strings.Contains(line, "TIME_WAIT") {
				timeWait++
			}
		}
	}

	return map[string]interface{}{
		"established_connections": established,
		"listening_connections":   listening,
		"time_wait_connections":   timeWait,
		"total_connections":       established + listening + timeWait,
	}, nil
}

// getVirtualHostStats 获取真实的虚拟主机统计
func (m *LinuxApacheMonitor) getVirtualHostStats() (map[string]interface{}, error) {
	// 读取配置文件
	content, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		return nil, err
	}

	// 查找VirtualHost配置
	vhostRegex := regexp.MustCompile(`<VirtualHost\s+([^>]+)>`)
	matches := vhostRegex.FindAllStringSubmatch(string(content), -1)

	vhosts := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			vhosts = append(vhosts, strings.TrimSpace(match[1]))
		}
	}

	// 查找ServerName配置
	serverNameRegex := regexp.MustCompile(`ServerName\s+([^\s]+)`)
	serverNames := serverNameRegex.FindAllStringSubmatch(string(content), -1)

	domains := make([]string, 0, len(serverNames))
	for _, match := range serverNames {
		if len(match) > 1 {
			domains = append(domains, strings.TrimSpace(match[1]))
		}
	}

	return map[string]interface{}{
		"total_vhosts":   len(vhosts),
		"vhost_configs":  vhosts,
		"server_names":   domains,
		"enabled_sites":  len(vhosts), // 简化处理
		"disabled_sites": 0,
	}, nil
}

// getModuleInfo 获取真实的模块信息
func (m *LinuxApacheMonitor) getModuleInfo() (map[string]interface{}, error) {
	cmd := exec.Command(m.ctl, "-M")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	loadedModules := make([]string, 0)
	staticModules := 0
	sharedModules := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Loaded Modules:") {
			continue
		}

		// 解析模块行
		parts := strings.Fields(line)
		if len(parts) > 0 {
			moduleName := parts[0]
			loadedModules = append(loadedModules, moduleName)

			if strings.Contains(line, "(static)") {
				staticModules++
			} else if strings.Contains(line, "(shared)") {
				sharedModules++
			}
		}
	}

	return map[string]interface{}{
		"total_modules":  len(loadedModules),
		"static_modules": staticModules,
		"shared_modules": sharedModules,
		"loaded_modules": loadedModules,
	}, nil
}

// getRequestStats 获取真实的请求统计
func (m *LinuxApacheMonitor) getRequestStats() (map[string]interface{}, error) {
	// 尝试从server-status获取
	if statusData := m.getServerStatusData(); statusData != nil {
		result := make(map[string]interface{})
		
		// 复制相关统计
		if totalAccesses, ok := statusData["total_accesses"]; ok {
			result["total_requests"] = totalAccesses
		}
		if totalBytes, ok := statusData["total_bytes"]; ok {
			result["total_bytes"] = totalBytes
		}
		if reqPerSec, ok := statusData["requests_per_sec"]; ok {
			result["requests_per_second"] = reqPerSec
		}
		if bytesPerSec, ok := statusData["bytes_per_sec"]; ok {
			result["bytes_per_second"] = bytesPerSec
		}
		
		return result, nil
	}

	// 回退到日志分析
	return m.parseAccessLog()
}

// parseAccessLog 解析访问日志
func (m *LinuxApacheMonitor) parseAccessLog() (map[string]interface{}, error) {
	file, err := os.Open(m.accessLogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open access log: %v", err)
	}
	defer file.Close()

	// 统计最近的请求
	var totalRequests int64
	statusCodeCounts := make(map[string]int)
	methodCounts := make(map[string]int)
	bytesTransferred := int64(0)

	// 读取日志文件的最后几行
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, 1000) // 最多读取1000行

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > 1000 {
			// 保持最新的1000行
			lines = lines[1:]
		}
	}

	// 解析日志行
	for _, line := range lines {
		if line == "" {
			continue
		}

		totalRequests++

		// 简单的日志解析 (假设是标准的combined格式)
		// 提取状态码
		statusRegex := regexp.MustCompile(`" (\d{3}) `)
		if matches := statusRegex.FindStringSubmatch(line); len(matches) > 1 {
			statusCodeCounts[matches[1]]++
		}

		// 提取HTTP方法
		methodRegex := regexp.MustCompile(`"(GET|POST|PUT|DELETE|HEAD|OPTIONS|PATCH) `)
		if matches := methodRegex.FindStringSubmatch(line); len(matches) > 1 {
			methodCounts[matches[1]]++
		}

		// 提取传输字节数
		bytesRegex := regexp.MustCompile(` (\d+) "[^"]*" "[^"]*"$`)
		if matches := bytesRegex.FindStringSubmatch(line); len(matches) > 1 {
			if bytes, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
				bytesTransferred += bytes
			}
		}
	}

	result := map[string]interface{}{
		"total_requests":    totalRequests,
		"bytes_transferred": bytesTransferred,
		"status_codes":      statusCodeCounts,
		"http_methods":      methodCounts,
	}

	// 计算成功率
	successRequests := statusCodeCounts["200"] + statusCodeCounts["201"] + statusCodeCounts["204"]
	if totalRequests > 0 {
		result["success_rate"] = float64(successRequests) / float64(totalRequests) * 100
	}

	return result, nil
}

// getSSLInfo 获取真实的SSL信息
func (m *LinuxApacheMonitor) getSSLInfo() (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		return nil, err
	}

	// 查找SSL配置
	sslEngineRegex := regexp.MustCompile(`SSLEngine\s+on`)
	sslCertRegex := regexp.MustCompile(`SSLCertificateFile\s+([^\s]+)`)
	sslKeyRegex := regexp.MustCompile(`SSLCertificateKeyFile\s+([^\s]+)`)
	listen443Regex := regexp.MustCompile(`Listen\s+443`)

	sslEnabled := len(sslEngineRegex.FindAllString(string(content), -1)) > 0
	sslCerts := sslCertRegex.FindAllStringSubmatch(string(content), -1)
	sslKeys := sslKeyRegex.FindAllStringSubmatch(string(content), -1)
	sslListeners := listen443Regex.FindAllString(string(content), -1)

	result := map[string]interface{}{
		"ssl_enabled":     sslEnabled,
		"ssl_certificates": len(sslCerts),
		"ssl_keys":         len(sslKeys),
		"ssl_listeners":    len(sslListeners),
	}

	// 检查证书有效期
	if len(sslCerts) > 0 {
		certPath := strings.TrimSpace(sslCerts[0][1])
		if expiry := m.getCertificateExpiry(certPath); expiry != "" {
			result["certificate_expiry"] = expiry
		}
	}

	return result, nil
}

// getCertificateExpiry 获取证书过期时间
func (m *LinuxApacheMonitor) getCertificateExpiry(certPath string) string {
	cmd := exec.Command("openssl", "x509", "-in", certPath, "-noout", "-enddate")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// 解析输出: notAfter=Dec 31 23:59:59 2024 GMT
	line := strings.TrimSpace(string(output))
	if strings.HasPrefix(line, "notAfter=") {
		return strings.TrimPrefix(line, "notAfter=")
	}
	return ""
}

// getCacheStats 获取真实的缓存统计
func (m *LinuxApacheMonitor) getCacheStats() (map[string]interface{}, error) {
	// 检查是否启用了mod_cache
	cmd := exec.Command(m.ctl, "-M")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	cacheEnabled := strings.Contains(string(output), "cache_module")

	result := map[string]interface{}{
		"cache_enabled": cacheEnabled,
	}

	if !cacheEnabled {
		return result, nil
	}

	// 查找缓存配置
	content, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		return result, nil
	}

	// 查找CacheRoot配置
	cacheRootRegex := regexp.MustCompile(`CacheRoot\s+([^\s]+)`)
	matches := cacheRootRegex.FindAllStringSubmatch(string(content), -1)

	if len(matches) > 0 {
		cacheRoot := matches[0][1]
		if stat, err := os.Stat(cacheRoot); err == nil && stat.IsDir() {
			if size := m.getDirSize(cacheRoot); size > 0 {
				result["cache_size_bytes"] = size
				result["cache_size_mb"] = size / 1024 / 1024
			}
			if count := m.getDirFileCount(cacheRoot); count > 0 {
				result["cache_files"] = count
			}
		}
	}

	return result, nil
}

// getDirSize 获取目录大小
func (m *LinuxApacheMonitor) getDirSize(path string) int64 {
	cmd := exec.Command("du", "-sb", path)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	fields := strings.Fields(string(output))
	if len(fields) > 0 {
		if size, err := strconv.ParseInt(fields[0], 10, 64); err == nil {
			return size
		}
	}
	return 0
}

// getDirFileCount 获取目录文件数量
func (m *LinuxApacheMonitor) getDirFileCount(path string) int {
	cmd := exec.Command("find", path, "-type", "f", "-exec", "echo", "1", ";")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

// 重写collectMetrics方法以使用Linux特定的实现
func (m *LinuxApacheMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 使用Linux特定的方法收集指标
	if serverStatus, err := m.getServerStatus(); err == nil {
		for k, v := range serverStatus {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get server status: %v", err)
	}

	if workerInfo, err := m.getWorkerInfo(); err == nil {
		for k, v := range workerInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get worker info: %v", err)
	}

	if connectionInfo, err := m.getConnectionInfo(); err == nil {
		for k, v := range connectionInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get connection info: %v", err)
	}

	if vhostStats, err := m.getVirtualHostStats(); err == nil {
		for k, v := range vhostStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get virtual host stats: %v", err)
	}

	if moduleInfo, err := m.getModuleInfo(); err == nil {
		for k, v := range moduleInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get module info: %v", err)
	}

	if requestStats, err := m.getRequestStats(); err == nil {
		for k, v := range requestStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get request stats: %v", err)
	}

	if sslInfo, err := m.getSSLInfo(); err == nil {
		for k, v := range sslInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get SSL info: %v", err)
	}

	if cacheStats, err := m.getCacheStats(); err == nil {
		for k, v := range cacheStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get cache stats: %v", err)
	}

	return metrics
}