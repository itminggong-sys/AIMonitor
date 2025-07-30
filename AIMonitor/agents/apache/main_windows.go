//go:build windows
// +build windows

package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"aimonitor-agents/common"
)

// WindowsApacheMonitor Windows版本的Apache监控器
type WindowsApacheMonitor struct {
	*ApacheMonitor
	apachePath     string
	confPath       string
	logPath        string
	statusURL      string
	infoURL        string
	client         *http.Client
	processName    string
	serviceName    string
	httpdPath      string
}

// Windows API 结构体（重用Nginx中的定义）
type PROCESSENTRY32_APACHE struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [260]uint16
}

type FILETIME_APACHE struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}

type SYSTEMTIME_APACHE struct {
	wYear         uint16
	wMonth        uint16
	wDayOfWeek    uint16
	wDay          uint16
	wHour         uint16
	wMinute       uint16
	wSecond       uint16
	wMilliseconds uint16
}

// Windows API 函数
var (
	kernel32_apache                = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot_apache = kernel32_apache.NewProc("CreateToolhelp32Snapshot")
	procProcess32First_apache      = kernel32_apache.NewProc("Process32FirstW")
	procProcess32Next_apache       = kernel32_apache.NewProc("Process32NextW")
	procCloseHandle_apache         = kernel32_apache.NewProc("CloseHandle")
	procOpenProcess_apache         = kernel32_apache.NewProc("OpenProcess")
	procGetProcessTimes_apache     = kernel32_apache.NewProc("GetProcessTimes")
	procFileTimeToSystemTime_apache = kernel32_apache.NewProc("FileTimeToSystemTime")
	procGetSystemTime_apache       = kernel32_apache.NewProc("GetSystemTime")
)

const (
	TH32CS_SNAPPROCESS_APACHE = 0x00000002
	PROCESS_QUERY_INFORMATION_APACHE = 0x0400
)

// NewWindowsApacheMonitor 创建Windows版本的Apache监控器
func NewWindowsApacheMonitor(agent *common.Agent) *WindowsApacheMonitor {
	baseMonitor := NewApacheMonitor(agent)
	return &WindowsApacheMonitor{
		ApacheMonitor: baseMonitor,
		apachePath:    "C:\\Apache24\\bin\\httpd.exe",
		httpdPath:     "C:\\Apache24\\bin\\httpd.exe",
		confPath:      "C:\\Apache24\\conf\\httpd.conf",
		logPath:       "C:\\Apache24\\logs",
		statusURL:     "http://localhost/server-status",
		infoURL:       "http://localhost/server-info",
		client:        &http.Client{Timeout: 10 * time.Second},
		processName:   "httpd.exe",
		serviceName:   "Apache2.4",
	}
}

// isApacheRunning 检查Apache是否运行
func (m *WindowsApacheMonitor) isApacheRunning() bool {
	processes, err := m.getProcessList()
	if err != nil {
		return false
	}
	
	for _, proc := range processes {
		name := strings.ToLower(proc["name"].(string))
		if strings.Contains(name, "httpd") || strings.Contains(name, "apache") {
			return true
		}
	}
	return false
}

// getProcessList 获取进程列表
func (m *WindowsApacheMonitor) getProcessList() ([]map[string]interface{}, error) {
	var processes []map[string]interface{}
	
	// 创建进程快照
	handle, _, _ := procCreateToolhelp32Snapshot_apache.Call(TH32CS_SNAPPROCESS_APACHE, 0)
	if handle == uintptr(syscall.InvalidHandle) {
		return nil, fmt.Errorf("failed to create process snapshot")
	}
	defer procCloseHandle_apache.Call(handle)
	
	var pe32 PROCESSENTRY32_APACHE
	pe32.dwSize = uint32(unsafe.Sizeof(pe32))
	
	// 获取第一个进程
	ret, _, _ := procProcess32First_apache.Call(handle, uintptr(unsafe.Pointer(&pe32)))
	if ret == 0 {
		return nil, fmt.Errorf("failed to get first process")
	}
	
	for {
		// 转换进程名
		name := syscall.UTF16ToString(pe32.szExeFile[:])
		
		process := map[string]interface{}{
			"pid":       pe32.th32ProcessID,
			"name":      name,
			"parent_pid": pe32.th32ParentProcessID,
			"threads":   pe32.cntThreads,
		}
		
		// 获取进程时间信息
		if processHandle, _, _ := procOpenProcess_apache.Call(PROCESS_QUERY_INFORMATION_APACHE, 0, uintptr(pe32.th32ProcessID)); processHandle != 0 {
			var creationTime, exitTime, kernelTime, userTime FILETIME_APACHE
			if ret, _, _ := procGetProcessTimes_apache.Call(processHandle, uintptr(unsafe.Pointer(&creationTime)), uintptr(unsafe.Pointer(&exitTime)), uintptr(unsafe.Pointer(&kernelTime)), uintptr(unsafe.Pointer(&userTime))); ret != 0 {
				// 转换创建时间
				var st SYSTEMTIME_APACHE
				if ret, _, _ := procFileTimeToSystemTime_apache.Call(uintptr(unsafe.Pointer(&creationTime)), uintptr(unsafe.Pointer(&st))); ret != 0 {
					createdAt := time.Date(int(st.wYear), time.Month(st.wMonth), int(st.wDay), int(st.wHour), int(st.wMinute), int(st.wSecond), int(st.wMilliseconds)*1000000, time.UTC)
					process["created_at"] = createdAt.Format(time.RFC3339)
					process["uptime_seconds"] = int64(time.Since(createdAt).Seconds())
				}
			}
			procCloseHandle_apache.Call(processHandle)
		}
		
		processes = append(processes, process)
		
		// 获取下一个进程
		pe32.dwSize = uint32(unsafe.Sizeof(pe32))
		ret, _, _ = procProcess32Next_apache.Call(handle, uintptr(unsafe.Pointer(&pe32)))
		if ret == 0 {
			break
		}
	}
	
	return processes, nil
}

// getApacheProcessInfo 获取真实的Apache进程信息
func (m *WindowsApacheMonitor) getApacheProcessInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	processes, err := m.getProcessList()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %v", err)
	}
	
	apacheProcesses := make([]map[string]interface{}, 0)
	masterPID := uint32(0)
	workerCount := 0
	totalMemory := int64(0)
	oldestUptime := int64(0)
	
	for _, proc := range processes {
		name := strings.ToLower(proc["name"].(string))
		if strings.Contains(name, "httpd") || strings.Contains(name, "apache") {
			apacheProc := map[string]interface{}{
				"pid":        proc["pid"],
				"name":       proc["name"],
				"parent_pid": proc["parent_pid"],
				"threads":    proc["threads"],
			}
			
			if createdAt, ok := proc["created_at"].(string); ok {
				apacheProc["created_at"] = createdAt
			}
			
			if uptime, ok := proc["uptime_seconds"].(int64); ok {
				apacheProc["uptime_seconds"] = uptime
				if uptime > oldestUptime {
					oldestUptime = uptime
				}
			}
			
			// 判断是否为master进程（通常parent_pid较小或为1）
			parentPID := proc["parent_pid"].(uint32)
			if parentPID == 1 || (masterPID == 0 && parentPID < 1000) {
				masterPID = proc["pid"].(uint32)
				apacheProc["type"] = "master"
			} else {
				apacheProc["type"] = "worker"
				workerCount++
			}
			
			// 获取内存使用情况
			pid := proc["pid"].(uint32)
			if memory, err := m.getProcessMemory(pid); err == nil {
				apacheProc["memory_mb"] = memory
				totalMemory += memory
			}
			
			apacheProcesses = append(apacheProcesses, apacheProc)
		}
	}
	
	result["is_running"] = len(apacheProcesses) > 0
	result["process_count"] = len(apacheProcesses)
	result["master_pid"] = masterPID
	result["worker_count"] = workerCount
	result["total_memory_mb"] = totalMemory
	result["uptime_seconds"] = oldestUptime
	result["processes"] = apacheProcesses
	
	return result, nil
}

// getProcessMemory 获取进程内存使用情况
func (m *WindowsApacheMonitor) getProcessMemory(pid uint32) (int64, error) {
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Get-Process -Id %d | Select-Object WorkingSet", pid))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	// 解析输出
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "WorkingSet") {
			continue
		}
		if line != "" && line != "----------" {
			if memory, err := strconv.ParseInt(strings.TrimSpace(line), 10, 64); err == nil {
				return memory / 1024 / 1024, nil // 转换为MB
			}
		}
	}
	
	return 0, fmt.Errorf("failed to parse memory usage")
}

// getApacheVersion 获取真实的Apache版本信息
func (m *WindowsApacheMonitor) getApacheVersion() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试执行httpd -v命令
	cmd := exec.Command(m.httpdPath, "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果直接执行失败，尝试从注册表或其他方式获取
		return m.getApacheVersionFromRegistry()
	}
	
	versionStr := string(output)
	// 解析版本信息，格式通常为: Server version: Apache/2.4.41 (Win64)
	if matches := regexp.MustCompile(`Server version: Apache/([\d\.]+)`).FindStringSubmatch(versionStr); len(matches) > 1 {
		result["version"] = matches[1]
	}
	
	if matches := regexp.MustCompile(`Server version: (.+)`).FindStringSubmatch(versionStr); len(matches) > 1 {
		result["full_version"] = strings.TrimSpace(matches[1])
	}
	
	if matches := regexp.MustCompile(`Server built: (.+)`).FindStringSubmatch(versionStr); len(matches) > 1 {
		result["built_date"] = strings.TrimSpace(matches[1])
	}
	
	// 获取编译信息
	cmd = exec.Command(m.httpdPath, "-V")
	output, err = cmd.CombinedOutput()
	if err == nil {
		buildInfo := string(output)
		result["build_info"] = strings.TrimSpace(buildInfo)
		
		// 解析编译选项
		if matches := regexp.MustCompile(`Server MPM:\s+(.+)`).FindStringSubmatch(buildInfo); len(matches) > 1 {
			result["mpm_module"] = strings.TrimSpace(matches[1])
		}
		
		if matches := regexp.MustCompile(`Server compiled with\.\.\.\s+(.+)`).FindStringSubmatch(buildInfo); len(matches) > 1 {
			result["compiled_with"] = strings.TrimSpace(matches[1])
		}
		
		// 解析configure参数
		lines := strings.Split(buildInfo, "\n")
		configureArgs := make([]string, 0)
		inConfigureSection := false
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Server compiled with") {
				inConfigureSection = true
				continue
			}
			if inConfigureSection && line != "" {
				if strings.HasPrefix(line, "-D") || strings.HasPrefix(line, "--") {
					configureArgs = append(configureArgs, line)
				}
			}
		}
		
		if len(configureArgs) > 0 {
			result["configure_arguments"] = configureArgs
		}
	}
	
	return result, nil
}

// getApacheVersionFromRegistry 从注册表获取Apache版本信息
func (m *WindowsApacheMonitor) getApacheVersionFromRegistry() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试从文件属性获取版本信息
	if info, err := os.Stat(m.httpdPath); err == nil {
		result["file_size"] = info.Size()
		result["file_modified"] = info.ModTime().Format(time.RFC3339)
	}
	
	// 使用PowerShell获取文件版本
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("(Get-ItemProperty '%s').VersionInfo", m.httpdPath))
	output, err := cmd.Output()
	if err == nil {
		versionInfo := string(output)
		if matches := regexp.MustCompile(`FileVersion\s+:\s+([\d\.]+)`).FindStringSubmatch(versionInfo); len(matches) > 1 {
			result["file_version"] = matches[1]
		}
		if matches := regexp.MustCompile(`ProductVersion\s+:\s+([\d\.]+)`).FindStringSubmatch(versionInfo); len(matches) > 1 {
			result["product_version"] = matches[1]
		}
		if matches := regexp.MustCompile(`ProductName\s+:\s+(.+)`).FindStringSubmatch(versionInfo); len(matches) > 1 {
			result["product_name"] = strings.TrimSpace(matches[1])
		}
	}
	
	return result, nil
}

// getConfigInfo 获取真实的配置信息
func (m *WindowsApacheMonitor) getConfigInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 检查配置文件是否存在
	if info, err := os.Stat(m.confPath); err == nil {
		result["config_file"] = m.confPath
		result["config_size"] = info.Size()
		result["config_modified"] = info.ModTime().Format(time.RFC3339)
		result["config_exists"] = true
		
		// 读取配置文件内容进行分析
		if content, err := ioutil.ReadFile(m.confPath); err == nil {
			configStr := string(content)
			result["config_lines"] = len(strings.Split(configStr, "\n"))
			
			// 解析ServerRoot
			if matches := regexp.MustCompile(`ServerRoot\s+"?([^"\s]+)"?`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["server_root"] = matches[1]
			}
			
			// 解析Listen指令
			listenMatches := regexp.MustCompile(`Listen\s+([^\s]+)`).FindAllStringSubmatch(configStr, -1)
			listenPorts := make([]string, 0)
			for _, match := range listenMatches {
				if len(match) > 1 {
					listenPorts = append(listenPorts, match[1])
				}
			}
			result["listen_ports"] = listenPorts
			
			// 解析ServerName
			if matches := regexp.MustCompile(`ServerName\s+([^\s]+)`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["server_name"] = matches[1]
			}
			
			// 解析DocumentRoot
			if matches := regexp.MustCompile(`DocumentRoot\s+"?([^"\s]+)"?`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["document_root"] = matches[1]
			}
			
			// 解析MaxRequestWorkers
			if matches := regexp.MustCompile(`MaxRequestWorkers\s+(\d+)`).FindStringSubmatch(configStr); len(matches) > 1 {
				if workers, err := strconv.Atoi(matches[1]); err == nil {
					result["max_request_workers"] = workers
				}
			}
			
			// 解析ThreadsPerChild
			if matches := regexp.MustCompile(`ThreadsPerChild\s+(\d+)`).FindStringSubmatch(configStr); len(matches) > 1 {
				if threads, err := strconv.Atoi(matches[1]); err == nil {
					result["threads_per_child"] = threads
				}
			}
			
			// 统计VirtualHost数量
			vhostCount := len(regexp.MustCompile(`<VirtualHost\s+[^>]+>`).FindAllString(configStr, -1))
			result["virtual_hosts"] = vhostCount
			
			// 统计Directory块数量
			directoryCount := len(regexp.MustCompile(`<Directory\s+[^>]+>`).FindAllString(configStr, -1))
			result["directory_blocks"] = directoryCount
			
			// 检查是否启用了status模块
			result["status_enabled"] = strings.Contains(configStr, "mod_status") || strings.Contains(configStr, "server-status")
			
			// 检查是否启用了info模块
			result["info_enabled"] = strings.Contains(configStr, "mod_info") || strings.Contains(configStr, "server-info")
			
			// 检查是否启用了SSL
			result["ssl_enabled"] = strings.Contains(configStr, "mod_ssl") || strings.Contains(configStr, "SSLEngine")
			
			// 检查是否启用了rewrite模块
			result["rewrite_enabled"] = strings.Contains(configStr, "mod_rewrite")
			
			// 统计加载的模块
			loadModuleMatches := regexp.MustCompile(`LoadModule\s+(\w+)`).FindAllStringSubmatch(configStr, -1)
			loadedModules := make([]string, 0)
			for _, match := range loadModuleMatches {
				if len(match) > 1 {
					loadedModules = append(loadedModules, match[1])
				}
			}
			result["loaded_modules"] = loadedModules
			result["loaded_modules_count"] = len(loadedModules)
		}
	} else {
		result["config_exists"] = false
		result["config_error"] = err.Error()
	}
	
	// 测试配置文件语法
	cmd := exec.Command(m.httpdPath, "-t")
	output, err := cmd.CombinedOutput()
	if err == nil {
		result["config_test"] = "passed"
		result["config_test_output"] = strings.TrimSpace(string(output))
	} else {
		result["config_test"] = "failed"
		result["config_test_error"] = strings.TrimSpace(string(output))
	}
	
	return result, nil
}

// getConnectionInfo 获取真实的连接信息
func (m *WindowsApacheMonitor) getConnectionInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试从server-status获取连接信息
	if statusInfo, err := m.getServerStatusInfo(); err == nil {
		for k, v := range statusInfo {
			result[k] = v
		}
	} else {
		// 如果server-status不可用，使用netstat获取连接信息
		if netstatInfo, err := m.getNetstatInfo(); err == nil {
			for k, v := range netstatInfo {
				result[k] = v
			}
		} else {
			result["connection_info_available"] = false
			result["error"] = fmt.Sprintf("server-status: %v, netstat: %v", err, err)
		}
	}
	
	return result, nil
}

// getServerStatusInfo 从server-status获取信息
func (m *WindowsApacheMonitor) getServerStatusInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试获取机器可读的状态信息
	statusURL := m.statusURL + "?auto"
	resp, err := m.client.Get(statusURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get server status: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server-status returned %d", resp.StatusCode)
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read status response: %v", err)
	}
	
	// 解析server-status输出
	statusText := string(body)
	lines := strings.Split(statusText, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "Total Accesses":
			if accesses, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["total_accesses"] = accesses
			}
		case "Total kBytes":
			if kbytes, err := strconv.ParseFloat(value, 64); err == nil {
				result["total_kbytes"] = kbytes
				result["total_bytes"] = int64(kbytes * 1024)
			}
		case "CPULoad":
			if cpuLoad, err := strconv.ParseFloat(value, 64); err == nil {
				result["cpu_load"] = cpuLoad
			}
		case "Uptime":
			if uptime, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["uptime_seconds"] = uptime
			}
		case "ReqPerSec":
			if reqPerSec, err := strconv.ParseFloat(value, 64); err == nil {
				result["requests_per_second"] = reqPerSec
			}
		case "BytesPerSec":
			if bytesPerSec, err := strconv.ParseFloat(value, 64); err == nil {
				result["bytes_per_second"] = bytesPerSec
			}
		case "BytesPerReq":
			if bytesPerReq, err := strconv.ParseFloat(value, 64); err == nil {
				result["bytes_per_request"] = bytesPerReq
			}
		case "BusyWorkers":
			if busyWorkers, err := strconv.Atoi(value); err == nil {
				result["busy_workers"] = busyWorkers
			}
		case "IdleWorkers":
			if idleWorkers, err := strconv.Atoi(value); err == nil {
				result["idle_workers"] = idleWorkers
			}
		case "Scoreboard":
			result["scoreboard"] = value
			// 分析scoreboard
			scoreboardStats := m.analyzeScoreboard(value)
			for k, v := range scoreboardStats {
				result[k] = v
			}
		}
	}
	
	result["status_source"] = "server_status"
	return result, nil
}

// analyzeScoreboard 分析Apache scoreboard
func (m *WindowsApacheMonitor) analyzeScoreboard(scoreboard string) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Scoreboard字符含义:
	// "_" = Waiting for Connection
	// "S" = Starting up
	// "R" = Reading Request
	// "W" = Sending Reply
	// "K" = Keepalive (read)
	// "D" = DNS Lookup
	// "C" = Closing connection
	// "L" = Logging
	// "G" = Gracefully finishing
	// "I" = Idle cleanup of worker
	// "." = Open slot with no current process
	
	counts := make(map[string]int)
	for _, char := range scoreboard {
		switch char {
		case '_':
			counts["waiting_for_connection"]++
		case 'S':
			counts["starting_up"]++
		case 'R':
			counts["reading_request"]++
		case 'W':
			counts["sending_reply"]++
		case 'K':
			counts["keepalive"]++
		case 'D':
			counts["dns_lookup"]++
		case 'C':
			counts["closing_connection"]++
		case 'L':
			counts["logging"]++
		case 'G':
			counts["gracefully_finishing"]++
		case 'I':
			counts["idle_cleanup"]++
		case '.':
			counts["open_slot"]++
		}
	}
	
	result["scoreboard_stats"] = counts
	result["total_slots"] = len(scoreboard)
	
	return result
}

// getNetstatInfo 使用netstat获取连接信息
func (m *WindowsApacheMonitor) getNetstatInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 使用netstat获取Apache相关的连接
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run netstat: %v", err)
	}
	
	lines := strings.Split(string(output), "\n")
	connections := make([]map[string]interface{}, 0)
	listenPorts := make([]string, 0)
	establishedCount := 0
	listeningCount := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "TCP") {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		
		protocol := fields[0]
		localAddr := fields[1]
		remoteAddr := fields[2]
		state := fields[3]
		
		// 检查是否是HTTP相关端口（80, 443, 8080等）
		if strings.Contains(localAddr, ":80") || strings.Contains(localAddr, ":443") || strings.Contains(localAddr, ":8080") {
			connection := map[string]interface{}{
				"protocol":     protocol,
				"local_address": localAddr,
				"remote_address": remoteAddr,
				"state":        state,
			}
			
			connections = append(connections, connection)
			
			if state == "LISTENING" {
				listeningCount++
				listenPorts = append(listenPorts, localAddr)
			} else if state == "ESTABLISHED" {
				establishedCount++
			}
		}
	}
	
	result["connections"] = connections
	result["connection_count"] = len(connections)
	result["established_connections"] = establishedCount
	result["listening_connections"] = listeningCount
	result["listen_ports"] = listenPorts
	result["status_source"] = "netstat"
	
	return result, nil
}

// getVirtualHostInfo 获取真实的虚拟主机信息
func (m *WindowsApacheMonitor) getVirtualHostInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试从server-info获取虚拟主机信息
	if infoData, err := m.getServerInfoData(); err == nil {
		for k, v := range infoData {
			result[k] = v
		}
	} else {
		// 如果server-info不可用，从配置文件解析
		if configVhosts, err := m.parseVirtualHostsFromConfig(); err == nil {
			for k, v := range configVhosts {
				result[k] = v
			}
		} else {
			result["virtual_hosts_available"] = false
			result["error"] = fmt.Sprintf("server-info: %v, config parsing: %v", err, err)
		}
	}
	
	return result, nil
}

// getServerInfoData 从server-info获取信息
func (m *WindowsApacheMonitor) getServerInfoData() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	resp, err := m.client.Get(m.infoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server-info returned %d", resp.StatusCode)
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read info response: %v", err)
	}
	
	// 解析HTML内容（简单的文本解析）
	infoText := string(body)
	
	// 提取服务器信息
	if matches := regexp.MustCompile(`Server Version: ([^<]+)`).FindStringSubmatch(infoText); len(matches) > 1 {
		result["server_version"] = strings.TrimSpace(matches[1])
	}
	
	if matches := regexp.MustCompile(`Server Built: ([^<]+)`).FindStringSubmatch(infoText); len(matches) > 1 {
		result["server_built"] = strings.TrimSpace(matches[1])
	}
	
	// 提取模块信息
	moduleMatches := regexp.MustCompile(`<dt><strong>([^<]+)</strong></dt>`).FindAllStringSubmatch(infoText, -1)
	modules := make([]string, 0)
	for _, match := range moduleMatches {
		if len(match) > 1 {
			moduleName := strings.TrimSpace(match[1])
			if strings.HasSuffix(moduleName, "_module") {
				modules = append(modules, moduleName)
			}
		}
	}
	result["loaded_modules"] = modules
	result["loaded_modules_count"] = len(modules)
	
	result["info_source"] = "server_info"
	return result, nil
}

// parseVirtualHostsFromConfig 从配置文件解析虚拟主机
func (m *WindowsApacheMonitor) parseVirtualHostsFromConfig() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	content, err := ioutil.ReadFile(m.confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	configStr := string(content)
	virtualHosts := make([]map[string]interface{}, 0)
	
	// 使用正则表达式查找VirtualHost块
	vhostRegex := regexp.MustCompile(`<VirtualHost\s+([^>]+)>([\s\S]*?)</VirtualHost>`)
	matches := vhostRegex.FindAllStringSubmatch(configStr, -1)
	
	for _, match := range matches {
		if len(match) >= 3 {
			vhostAddress := strings.TrimSpace(match[1])
			vhostContent := match[2]
			
			vhostInfo := map[string]interface{}{
				"address": vhostAddress,
			}
			
			// 解析ServerName
			if serverNameMatch := regexp.MustCompile(`ServerName\s+([^\s]+)`).FindStringSubmatch(vhostContent); len(serverNameMatch) > 1 {
				vhostInfo["server_name"] = serverNameMatch[1]
			}
			
			// 解析ServerAlias
			serverAliasMatches := regexp.MustCompile(`ServerAlias\s+([^\n]+)`).FindAllStringSubmatch(vhostContent, -1)
			aliases := make([]string, 0)
			for _, aliasMatch := range serverAliasMatches {
				if len(aliasMatch) > 1 {
					aliasFields := strings.Fields(aliasMatch[1])
					aliases = append(aliases, aliasFields...)
				}
			}
			if len(aliases) > 0 {
				vhostInfo["server_aliases"] = aliases
			}
			
			// 解析DocumentRoot
			if docRootMatch := regexp.MustCompile(`DocumentRoot\s+"?([^"\s]+)"?`).FindStringSubmatch(vhostContent); len(docRootMatch) > 1 {
				vhostInfo["document_root"] = docRootMatch[1]
			}
			
			// 检查是否启用SSL
			vhostInfo["ssl_enabled"] = strings.Contains(vhostContent, "SSLEngine on")
			
			// 解析错误日志
			if errorLogMatch := regexp.MustCompile(`ErrorLog\s+"?([^"\s]+)"?`).FindStringSubmatch(vhostContent); len(errorLogMatch) > 1 {
				vhostInfo["error_log"] = errorLogMatch[1]
			}
			
			// 解析访问日志
			if accessLogMatch := regexp.MustCompile(`CustomLog\s+"?([^"\s]+)"?`).FindStringSubmatch(vhostContent); len(accessLogMatch) > 1 {
				vhostInfo["access_log"] = accessLogMatch[1]
			}
			
			virtualHosts = append(virtualHosts, vhostInfo)
		}
	}
	
	result["virtual_hosts"] = virtualHosts
	result["virtual_host_count"] = len(virtualHosts)
	result["virtual_hosts_available"] = len(virtualHosts) > 0
	result["info_source"] = "config_file"
	
	return result, nil
}

// getRequestStats 获取真实的请求统计
func (m *WindowsApacheMonitor) getRequestStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试从server-status获取统计信息
	if statusStats, err := m.getServerStatusInfo(); err == nil {
		for k, v := range statusStats {
			result[k] = v
		}
	} else {
		// 如果server-status不可用，尝试从日志文件分析
		if logStats, err := m.getLogStats(); err == nil {
			for k, v := range logStats {
				result[k] = v
			}
		} else {
			result["stats_source"] = "unavailable"
			result["error"] = fmt.Sprintf("server-status: %v, log analysis: %v", err, err)
		}
	}
	
	return result, nil
}

// getLogStats 从日志文件分析统计信息
func (m *WindowsApacheMonitor) getLogStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	accessLogPath := filepath.Join(m.logPath, "access.log")
	errorLogPath := filepath.Join(m.logPath, "error.log")
	
	// 分析访问日志
	if accessStats, err := m.analyzeAccessLog(accessLogPath); err == nil {
		for k, v := range accessStats {
			result[k] = v
		}
	}
	
	// 分析错误日志
	if errorStats, err := m.analyzeErrorLog(errorLogPath); err == nil {
		for k, v := range errorStats {
			result[k] = v
		}
	}
	
	result["stats_source"] = "log_analysis"
	return result, nil
}

// analyzeAccessLog 分析访问日志
func (m *WindowsApacheMonitor) analyzeAccessLog(logPath string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	file, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open access log: %v", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	totalRequests := int64(0)
	statusCodeCounts := make(map[string]int64)
	methodCounts := make(map[string]int64)
	ipCounts := make(map[string]int64)
	bytesTransferred := int64(0)
	userAgents := make(map[string]int64)
	
	// 只分析最近的1000行以提高性能
	lines := make([]string, 0, 1000)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > 1000 {
			lines = lines[1:] // 保持最新的1000行
		}
	}
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		totalRequests++
		
		// 解析日志行（假设是标准的combined格式）
		// IP - - [timestamp] "METHOD /path HTTP/1.1" status size "referer" "user-agent"
		parts := strings.Fields(line)
		if len(parts) >= 9 {
			// IP地址
			ip := parts[0]
			ipCounts[ip]++
			
			// HTTP方法
			if len(parts[5]) > 1 {
				method := strings.Trim(parts[5], `"`)
				methodCounts[method]++
			}
			
			// 状态码
			statusCode := parts[8]
			statusCodeCounts[statusCode]++
			
			// 传输字节数
			if len(parts) >= 10 && parts[9] != "-" {
				if bytes, err := strconv.ParseInt(parts[9], 10, 64); err == nil {
					bytesTransferred += bytes
				}
			}
			
			// User-Agent（如果存在）
			if len(parts) >= 12 {
				userAgent := strings.Trim(parts[11], `"`)
				if userAgent != "-" {
					userAgents[userAgent]++
				}
			}
		}
	}
	
	result["total_requests"] = totalRequests
	result["status_codes"] = statusCodeCounts
	result["http_methods"] = methodCounts
	result["top_ips"] = m.getTopEntries(ipCounts, 10)
	result["top_user_agents"] = m.getTopEntries(userAgents, 5)
	result["bytes_transferred"] = bytesTransferred
	
	// 计算错误率
	errorRequests := int64(0)
	for status, count := range statusCodeCounts {
		if strings.HasPrefix(status, "4") || strings.HasPrefix(status, "5") {
			errorRequests += count
		}
	}
	if totalRequests > 0 {
		result["error_rate"] = float64(errorRequests) / float64(totalRequests) * 100
	}
	
	return result, nil
}

// analyzeErrorLog 分析错误日志
func (m *WindowsApacheMonitor) analyzeErrorLog(logPath string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	file, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open error log: %v", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	totalErrors := int64(0)
	errorLevels := make(map[string]int64)
	errorTypes := make(map[string]int64)
	errorModules := make(map[string]int64)
	
	// 只分析最近的500行错误日志
	lines := make([]string, 0, 500)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > 500 {
			lines = lines[1:]
		}
	}
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		totalErrors++
		
		// 解析错误级别
		if matches := regexp.MustCompile(`\[(\w+)\]`).FindStringSubmatch(line); len(matches) > 1 {
			level := matches[1]
			errorLevels[level]++
		}
		
		// 解析模块名
		if matches := regexp.MustCompile(`\[([\w:]+)\]`).FindStringSubmatch(line); len(matches) > 1 {
			module := matches[1]
			if strings.Contains(module, ":") {
				parts := strings.Split(module, ":")
				if len(parts) > 0 {
					errorModules[parts[0]]++
				}
			}
		}
		
		// 分析错误类型
		if strings.Contains(line, "File does not exist") {
			errorTypes["file_not_found"]++
		} else if strings.Contains(line, "Permission denied") {
			errorTypes["permission_denied"]++
		} else if strings.Contains(line, "Connection refused") {
			errorTypes["connection_refused"]++
		} else if strings.Contains(line, "Timeout") {
			errorTypes["timeout"]++
		} else if strings.Contains(line, "Internal error") {
			errorTypes["internal_error"]++
		} else {
			errorTypes["other"]++
		}
	}
	
	result["total_errors"] = totalErrors
	result["error_levels"] = errorLevels
	result["error_types"] = errorTypes
	result["error_modules"] = errorModules
	
	return result, nil
}

// getTopEntries 获取前N个条目
func (m *WindowsApacheMonitor) getTopEntries(counts map[string]int64, limit int) map[string]int64 {
	type entry struct {
		key   string
		value int64
	}
	
	entries := make([]entry, 0, len(counts))
	for k, v := range counts {
		entries = append(entries, entry{k, v})
	}
	
	// 简单排序（冒泡排序，适用于小数据集）
	for i := 0; i < len(entries)-1; i++ {
		for j := 0; j < len(entries)-i-1; j++ {
			if entries[j].value < entries[j+1].value {
				entries[j], entries[j+1] = entries[j+1], entries[j]
			}
		}
	}
	
	result := make(map[string]int64)
	for i, entry := range entries {
		if i >= limit {
			break
		}
		result[entry.key] = entry.value
	}
	
	return result
}

// getSSLInfo 获取真实的SSL信息
func (m *WindowsApacheMonitor) getSSLInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 从配置文件解析SSL配置
	content, err := ioutil.ReadFile(m.confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	configStr := string(content)
	sslCertificates := make([]map[string]interface{}, 0)
	
	// 查找SSL证书配置
	certRegex := regexp.MustCompile(`SSLCertificateFile\s+"?([^"\s]+)"?`)
	certMatches := certRegex.FindAllStringSubmatch(configStr, -1)
	
	keyRegex := regexp.MustCompile(`SSLCertificateKeyFile\s+"?([^"\s]+)"?`)
	keyMatches := keyRegex.FindAllStringSubmatch(configStr, -1)
	
	// 处理找到的证书
	for i, certMatch := range certMatches {
		if len(certMatch) >= 2 {
			certPath := strings.TrimSpace(certMatch[1])
			certInfo := map[string]interface{}{
				"certificate_path": certPath,
			}
			
			// 对应的私钥
			if i < len(keyMatches) && len(keyMatches[i]) >= 2 {
				certInfo["private_key_path"] = strings.TrimSpace(keyMatches[i][1])
			}
			
			// 检查证书文件是否存在并获取信息
			if certDetails, err := m.getCertificateDetails(certPath); err == nil {
				for k, v := range certDetails {
					certInfo[k] = v
				}
			} else {
				certInfo["error"] = err.Error()
			}
			
			sslCertificates = append(sslCertificates, certInfo)
		}
	}
	
	result["ssl_enabled"] = len(sslCertificates) > 0
	result["certificate_count"] = len(sslCertificates)
	result["certificates"] = sslCertificates
	
	// 检查SSL协议配置
	if matches := regexp.MustCompile(`SSLProtocol\s+([^\n]+)`).FindStringSubmatch(configStr); len(matches) > 1 {
		protocols := strings.Fields(strings.TrimSpace(matches[1]))
		result["ssl_protocols"] = protocols
	}
	
	// 检查SSL密码套件
	if matches := regexp.MustCompile(`SSLCipherSuite\s+([^\n]+)`).FindStringSubmatch(configStr); len(matches) > 1 {
		result["ssl_cipher_suite"] = strings.TrimSpace(matches[1])
	}
	
	return result, nil
}

// getCertificateDetails 获取证书详细信息
func (m *WindowsApacheMonitor) getCertificateDetails(certPath string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 检查文件是否存在
	if info, err := os.Stat(certPath); err == nil {
		result["file_size"] = info.Size()
		result["file_modified"] = info.ModTime().Format(time.RFC3339)
		result["file_exists"] = true
	} else {
		result["file_exists"] = false
		return result, fmt.Errorf("certificate file not found: %v", err)
	}
	
	// 使用OpenSSL命令获取证书信息（如果可用）
	cmd := exec.Command("openssl", "x509", "-in", certPath, "-text", "-noout")
	output, err := cmd.Output()
	if err == nil {
		certText := string(output)
		
		// 解析证书信息
		if matches := regexp.MustCompile(`Subject: (.+)`).FindStringSubmatch(certText); len(matches) > 1 {
			result["subject"] = strings.TrimSpace(matches[1])
		}
		
		if matches := regexp.MustCompile(`Issuer: (.+)`).FindStringSubmatch(certText); len(matches) > 1 {
			result["issuer"] = strings.TrimSpace(matches[1])
		}
		
		if matches := regexp.MustCompile(`Not Before: (.+)`).FindStringSubmatch(certText); len(matches) > 1 {
			result["not_before"] = strings.TrimSpace(matches[1])
		}
		
		if matches := regexp.MustCompile(`Not After : (.+)`).FindStringSubmatch(certText); len(matches) > 1 {
			notAfter := strings.TrimSpace(matches[1])
			result["not_after"] = notAfter
			
			// 计算证书到期时间
			if expireTime, err := time.Parse("Jan 2 15:04:05 2006 MST", notAfter); err == nil {
				daysUntilExpiry := int(time.Until(expireTime).Hours() / 24)
				result["days_until_expiry"] = daysUntilExpiry
				result["is_expired"] = daysUntilExpiry < 0
				result["expires_soon"] = daysUntilExpiry < 30 && daysUntilExpiry >= 0
			}
		}
		
		// 解析SAN（Subject Alternative Names）
		if matches := regexp.MustCompile(`DNS:([^,\s]+)`).FindAllStringSubmatch(certText, -1); len(matches) > 0 {
			sans := make([]string, 0)
			for _, match := range matches {
				if len(match) > 1 {
					sans = append(sans, match[1])
				}
			}
			result["subject_alt_names"] = sans
		}
	} else {
		// 如果OpenSSL不可用，尝试使用Go的crypto/tls包
		if certData, err := ioutil.ReadFile(certPath); err == nil {
			if cert, err := tls.X509KeyPair(certData, nil); err == nil {
				if len(cert.Certificate) > 0 {
					if x509Cert, err := x509.ParseCertificate(cert.Certificate[0]); err == nil {
						result["subject"] = x509Cert.Subject.String()
						result["issuer"] = x509Cert.Issuer.String()
						result["not_before"] = x509Cert.NotBefore.Format(time.RFC3339)
						result["not_after"] = x509Cert.NotAfter.Format(time.RFC3339)
						
						daysUntilExpiry := int(time.Until(x509Cert.NotAfter).Hours() / 24)
						result["days_until_expiry"] = daysUntilExpiry
						result["is_expired"] = daysUntilExpiry < 0
						result["expires_soon"] = daysUntilExpiry < 30 && daysUntilExpiry >= 0
						
						if len(x509Cert.DNSNames) > 0 {
							result["subject_alt_names"] = x509Cert.DNSNames
						}
					}
				}
			}
		}
	}
	
	return result, nil
}

// 重写collectMetrics方法以使用Windows特定的实现
func (m *WindowsApacheMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// 添加基本信息
	metrics["apache_path"] = m.apachePath
	metrics["httpd_path"] = m.httpdPath
	metrics["config_path"] = m.confPath
	metrics["log_path"] = m.logPath
	metrics["collection_time"] = time.Now().Format(time.RFC3339)
	
	// 使用Windows特定的方法收集指标
	if processInfo, err := m.getApacheProcessInfo(); err == nil {
		for k, v := range processInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get process info: %v", err)
		// 如果获取进程信息失败，设置基本状态
		metrics["is_running"] = false
		metrics["process_error"] = err.Error()
	}
	
	if versionInfo, err := m.getApacheVersion(); err == nil {
		for k, v := range versionInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get version info: %v", err)
	}
	
	if configInfo, err := m.getConfigInfo(); err == nil {
		for k, v := range configInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get config info: %v", err)
	}
	
	if connectionInfo, err := m.getConnectionInfo(); err == nil {
		for k, v := range connectionInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get connection info: %v", err)
	}
	
	if vhostInfo, err := m.getVirtualHostInfo(); err == nil {
		for k, v := range vhostInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get virtual host info: %v", err)
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
	
	// 如果大部分指标收集失败，回退到模拟数据
	if len(metrics) <= 5 { // 只有基本信息
		m.agent.Logger.Warn("Most metric collection failed, falling back to simulated data")
		return m.ApacheMonitor.collectMetrics()
	}
	
	return metrics
}