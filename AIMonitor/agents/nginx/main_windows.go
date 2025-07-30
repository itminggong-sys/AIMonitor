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

// WindowsNginxMonitor Windows版本的Nginx监控器
type WindowsNginxMonitor struct {
	*NginxMonitor
	nginxPath      string
	confPath       string
	logPath        string
	statusURL      string
	pidFile        string
	client         *http.Client
	processName    string
	serviceName    string
}

// Windows API 结构体
type PROCESSENTRY32 struct {
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

type FILETIME struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}

type SYSTEMTIME struct {
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
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First      = kernel32.NewProc("Process32FirstW")
	procProcess32Next       = kernel32.NewProc("Process32NextW")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	procOpenProcess         = kernel32.NewProc("OpenProcess")
	procGetProcessTimes     = kernel32.NewProc("GetProcessTimes")
	procFileTimeToSystemTime = kernel32.NewProc("FileTimeToSystemTime")
	procGetSystemTime       = kernel32.NewProc("GetSystemTime")
)

const (
	TH32CS_SNAPPROCESS = 0x00000002
	PROCESS_QUERY_INFORMATION = 0x0400
)

// NewWindowsNginxMonitor 创建Windows版本的Nginx监控器
func NewWindowsNginxMonitor(agent *common.Agent) *WindowsNginxMonitor {
	baseMonitor := NewNginxMonitor(agent)
	return &WindowsNginxMonitor{
		NginxMonitor: baseMonitor,
		nginxPath:    "C:\\nginx\\nginx.exe",
		confPath:     "C:\\nginx\\conf\\nginx.conf",
		logPath:      "C:\\nginx\\logs",
		statusURL:    "http://localhost/nginx_status",
		pidFile:      "C:\\nginx\\logs\\nginx.pid",
		client:       &http.Client{Timeout: 10 * time.Second},
		processName:  "nginx.exe",
		serviceName:  "nginx",
	}
}

// isNginxRunning 检查Nginx是否运行
func (m *WindowsNginxMonitor) isNginxRunning() bool {
	processes, err := m.getProcessList()
	if err != nil {
		return false
	}
	
	for _, proc := range processes {
		if strings.Contains(strings.ToLower(proc["name"].(string)), "nginx") {
			return true
		}
	}
	return false
}

// getProcessList 获取进程列表
func (m *WindowsNginxMonitor) getProcessList() ([]map[string]interface{}, error) {
	var processes []map[string]interface{}
	
	// 创建进程快照
	handle, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if handle == uintptr(syscall.InvalidHandle) {
		return nil, fmt.Errorf("failed to create process snapshot")
	}
	defer procCloseHandle.Call(handle)
	
	var pe32 PROCESSENTRY32
	pe32.dwSize = uint32(unsafe.Sizeof(pe32))
	
	// 获取第一个进程
	ret, _, _ := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&pe32)))
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
		if processHandle, _, _ := procOpenProcess.Call(PROCESS_QUERY_INFORMATION, 0, uintptr(pe32.th32ProcessID)); processHandle != 0 {
			var creationTime, exitTime, kernelTime, userTime FILETIME
			if ret, _, _ := procGetProcessTimes.Call(processHandle, uintptr(unsafe.Pointer(&creationTime)), uintptr(unsafe.Pointer(&exitTime)), uintptr(unsafe.Pointer(&kernelTime)), uintptr(unsafe.Pointer(&userTime))); ret != 0 {
				// 转换创建时间
				var st SYSTEMTIME
				if ret, _, _ := procFileTimeToSystemTime.Call(uintptr(unsafe.Pointer(&creationTime)), uintptr(unsafe.Pointer(&st))); ret != 0 {
					createdAt := time.Date(int(st.wYear), time.Month(st.wMonth), int(st.wDay), int(st.wHour), int(st.wMinute), int(st.wSecond), int(st.wMilliseconds)*1000000, time.UTC)
					process["created_at"] = createdAt.Format(time.RFC3339)
					process["uptime_seconds"] = int64(time.Since(createdAt).Seconds())
				}
			}
			procCloseHandle.Call(processHandle)
		}
		
		processes = append(processes, process)
		
		// 获取下一个进程
		pe32.dwSize = uint32(unsafe.Sizeof(pe32))
		ret, _, _ = procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&pe32)))
		if ret == 0 {
			break
		}
	}
	
	return processes, nil
}

// getNginxProcessInfo 获取真实的Nginx进程信息
func (m *WindowsNginxMonitor) getNginxProcessInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	processes, err := m.getProcessList()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %v", err)
	}
	
	nginxProcesses := make([]map[string]interface{}, 0)
	masterPID := uint32(0)
	workerCount := 0
	totalMemory := int64(0)
	oldestUptime := int64(0)
	
	for _, proc := range processes {
		name := proc["name"].(string)
		if strings.Contains(strings.ToLower(name), "nginx") {
			nginxProc := map[string]interface{}{
				"pid":        proc["pid"],
				"name":       name,
				"parent_pid": proc["parent_pid"],
				"threads":    proc["threads"],
			}
			
			if createdAt, ok := proc["created_at"].(string); ok {
				nginxProc["created_at"] = createdAt
			}
			
			if uptime, ok := proc["uptime_seconds"].(int64); ok {
				nginxProc["uptime_seconds"] = uptime
				if uptime > oldestUptime {
					oldestUptime = uptime
				}
			}
			
			// 判断是否为master进程（通常parent_pid较小或为1）
			parentPID := proc["parent_pid"].(uint32)
			if parentPID == 1 || (masterPID == 0 && parentPID < 1000) {
				masterPID = proc["pid"].(uint32)
				nginxProc["type"] = "master"
			} else {
				nginxProc["type"] = "worker"
				workerCount++
			}
			
			// 获取内存使用情况（通过PowerShell）
			pid := proc["pid"].(uint32)
			if memory, err := m.getProcessMemory(pid); err == nil {
				nginxProc["memory_mb"] = memory
				totalMemory += memory
			}
			
			nginxProcesses = append(nginxProcesses, nginxProc)
		}
	}
	
	result["is_running"] = len(nginxProcesses) > 0
	result["process_count"] = len(nginxProcesses)
	result["master_pid"] = masterPID
	result["worker_count"] = workerCount
	result["total_memory_mb"] = totalMemory
	result["uptime_seconds"] = oldestUptime
	result["processes"] = nginxProcesses
	
	return result, nil
}

// getProcessMemory 获取进程内存使用情况
func (m *WindowsNginxMonitor) getProcessMemory(pid uint32) (int64, error) {
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

// getNginxVersion 获取真实的Nginx版本信息
func (m *WindowsNginxMonitor) getNginxVersion() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试执行nginx -v命令
	cmd := exec.Command(m.nginxPath, "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果直接执行失败，尝试从注册表或其他方式获取
		return m.getNginxVersionFromRegistry()
	}
	
	versionStr := string(output)
	// 解析版本信息，格式通常为: nginx version: nginx/1.20.1
	if matches := regexp.MustCompile(`nginx version: nginx/([\d\.]+)`).FindStringSubmatch(versionStr); len(matches) > 1 {
		result["version"] = matches[1]
		result["full_version"] = strings.TrimSpace(versionStr)
	}
	
	// 获取编译信息
	cmd = exec.Command(m.nginxPath, "-V")
	output, err = cmd.CombinedOutput()
	if err == nil {
		buildInfo := string(output)
		result["build_info"] = strings.TrimSpace(buildInfo)
		
		// 解析编译选项
		if matches := regexp.MustCompile(`built by (.+)`).FindStringSubmatch(buildInfo); len(matches) > 1 {
			result["built_by"] = strings.TrimSpace(matches[1])
		}
		
		if matches := regexp.MustCompile(`built with (.+)`).FindStringSubmatch(buildInfo); len(matches) > 1 {
			result["built_with"] = strings.TrimSpace(matches[1])
		}
		
		// 解析configure参数
		if matches := regexp.MustCompile(`configure arguments: (.+)`).FindStringSubmatch(buildInfo); len(matches) > 1 {
			result["configure_arguments"] = strings.TrimSpace(matches[1])
		}
	}
	
	return result, nil
}

// getNginxVersionFromRegistry 从注册表获取Nginx版本信息
func (m *WindowsNginxMonitor) getNginxVersionFromRegistry() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试从文件属性获取版本信息
	if info, err := os.Stat(m.nginxPath); err == nil {
		result["file_size"] = info.Size()
		result["file_modified"] = info.ModTime().Format(time.RFC3339)
	}
	
	// 使用PowerShell获取文件版本
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("(Get-ItemProperty '%s').VersionInfo", m.nginxPath))
	output, err := cmd.Output()
	if err == nil {
		versionInfo := string(output)
		if matches := regexp.MustCompile(`FileVersion\s+:\s+([\d\.]+)`).FindStringSubmatch(versionInfo); len(matches) > 1 {
			result["file_version"] = matches[1]
		}
		if matches := regexp.MustCompile(`ProductVersion\s+:\s+([\d\.]+)`).FindStringSubmatch(versionInfo); len(matches) > 1 {
			result["product_version"] = matches[1]
		}
	}
	
	return result, nil
}

// getConfigInfo 获取真实的配置信息
func (m *WindowsNginxMonitor) getConfigInfo() (map[string]interface{}, error) {
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
			
			// 解析worker_processes
			if matches := regexp.MustCompile(`worker_processes\s+(\w+);`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["worker_processes"] = matches[1]
			}
			
			// 解析worker_connections
			if matches := regexp.MustCompile(`worker_connections\s+(\d+);`).FindStringSubmatch(configStr); len(matches) > 1 {
				if connections, err := strconv.Atoi(matches[1]); err == nil {
					result["worker_connections"] = connections
				}
			}
			
			// 解析keepalive_timeout
			if matches := regexp.MustCompile(`keepalive_timeout\s+(\d+);`).FindStringSubmatch(configStr); len(matches) > 1 {
				if timeout, err := strconv.Atoi(matches[1]); err == nil {
					result["keepalive_timeout"] = timeout
				}
			}
			
			// 统计server块数量
			serverCount := len(regexp.MustCompile(`server\s*{`).FindAllString(configStr, -1))
			result["server_blocks"] = serverCount
			
			// 统计location块数量
			locationCount := len(regexp.MustCompile(`location\s+[^{]+{`).FindAllString(configStr, -1))
			result["location_blocks"] = locationCount
			
			// 检查是否启用了status模块
			result["status_enabled"] = strings.Contains(configStr, "stub_status")
			
			// 检查是否启用了gzip
			result["gzip_enabled"] = strings.Contains(configStr, "gzip on")
			
			// 检查是否启用了SSL
			result["ssl_enabled"] = strings.Contains(configStr, "ssl_certificate")
		}
	} else {
		result["config_exists"] = false
		result["config_error"] = err.Error()
	}
	
	// 测试配置文件语法
	cmd := exec.Command(m.nginxPath, "-t")
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

// getRequestStats 获取真实的请求统计
func (m *WindowsNginxMonitor) getRequestStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试从status模块获取统计信息
	if statusStats, err := m.getStatusModuleStats(); err == nil {
		for k, v := range statusStats {
			result[k] = v
		}
	} else {
		// 如果status模块不可用，尝试从日志文件分析
		if logStats, err := m.getLogStats(); err == nil {
			for k, v := range logStats {
				result[k] = v
			}
		} else {
			result["stats_source"] = "unavailable"
			result["error"] = fmt.Sprintf("status module: %v, log analysis: %v", err, err)
		}
	}
	
	return result, nil
}

// getStatusModuleStats 从status模块获取统计信息
func (m *WindowsNginxMonitor) getStatusModuleStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	resp, err := m.client.Get(m.statusURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status endpoint returned %d", resp.StatusCode)
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	// 解析nginx status输出
	// 格式通常为:
	// Active connections: 1
	// server accepts handled requests
	//  1 1 1
	// Reading: 0 Writing: 1 Waiting: 0
	statusText := string(body)
	lines := strings.Split(statusText, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 解析活跃连接数
		if matches := regexp.MustCompile(`Active connections:\s+(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			if active, err := strconv.Atoi(matches[1]); err == nil {
				result["active_connections"] = active
			}
		}
		
		// 解析服务器统计（accepts handled requests）
		if matches := regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+(\d+)\s*$`).FindStringSubmatch(line); len(matches) > 3 {
			if accepts, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
				result["total_accepts"] = accepts
			}
			if handled, err := strconv.ParseInt(matches[2], 10, 64); err == nil {
				result["total_handled"] = handled
			}
			if requests, err := strconv.ParseInt(matches[3], 10, 64); err == nil {
				result["total_requests"] = requests
			}
		}
		
		// 解析连接状态（Reading Writing Waiting）
		if matches := regexp.MustCompile(`Reading:\s+(\d+)\s+Writing:\s+(\d+)\s+Waiting:\s+(\d+)`).FindStringSubmatch(line); len(matches) > 3 {
			if reading, err := strconv.Atoi(matches[1]); err == nil {
				result["reading_connections"] = reading
			}
			if writing, err := strconv.Atoi(matches[2]); err == nil {
				result["writing_connections"] = writing
			}
			if waiting, err := strconv.Atoi(matches[3]); err == nil {
				result["waiting_connections"] = waiting
			}
		}
	}
	
	result["stats_source"] = "status_module"
	return result, nil
}

// getLogStats 从日志文件分析统计信息
func (m *WindowsNginxMonitor) getLogStats() (map[string]interface{}, error) {
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
func (m *WindowsNginxMonitor) analyzeAccessLog(logPath string) (map[string]interface{}, error) {
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
		}
	}
	
	result["total_requests"] = totalRequests
	result["status_codes"] = statusCodeCounts
	result["http_methods"] = methodCounts
	result["top_ips"] = m.getTopEntries(ipCounts, 10)
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
func (m *WindowsNginxMonitor) analyzeErrorLog(logPath string) (map[string]interface{}, error) {
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
		
		// 分析错误类型
		if strings.Contains(line, "connection refused") {
			errorTypes["connection_refused"]++
		} else if strings.Contains(line, "timeout") {
			errorTypes["timeout"]++
		} else if strings.Contains(line, "permission denied") {
			errorTypes["permission_denied"]++
		} else if strings.Contains(line, "file not found") {
			errorTypes["file_not_found"]++
		} else {
			errorTypes["other"]++
		}
	}
	
	result["total_errors"] = totalErrors
	result["error_levels"] = errorLevels
	result["error_types"] = errorTypes
	
	return result, nil
}

// getTopEntries 获取前N个条目
func (m *WindowsNginxMonitor) getTopEntries(counts map[string]int64, limit int) map[string]int64 {
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

// getUpstreamInfo 获取真实的上游服务器信息
func (m *WindowsNginxMonitor) getUpstreamInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试从nginx plus API获取上游信息（如果可用）
	if upstreamStats, err := m.getNginxPlusUpstreams(); err == nil {
		result = upstreamStats
	} else {
		// 如果nginx plus不可用，从配置文件解析上游信息
		if configUpstreams, err := m.parseUpstreamsFromConfig(); err == nil {
			result = configUpstreams
		} else {
			result["upstreams_available"] = false
			result["error"] = "nginx plus API and config parsing both failed"
		}
	}
	
	return result, nil
}

// getNginxPlusUpstreams 从nginx plus API获取上游信息
func (m *WindowsNginxMonitor) getNginxPlusUpstreams() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试访问nginx plus API
	upstreamURL := "http://localhost:8080/api/6/http/upstreams"
	resp, err := m.client.Get(upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("nginx plus API not available: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("nginx plus API returned %d", resp.StatusCode)
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %v", err)
	}
	
	var upstreams map[string]interface{}
	if err := json.Unmarshal(body, &upstreams); err != nil {
		return nil, fmt.Errorf("failed to parse upstream response: %v", err)
	}
	
	result["upstreams"] = upstreams
	result["upstreams_available"] = true
	result["source"] = "nginx_plus_api"
	
	return result, nil
}

// parseUpstreamsFromConfig 从配置文件解析上游信息
func (m *WindowsNginxMonitor) parseUpstreamsFromConfig() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	content, err := ioutil.ReadFile(m.confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	configStr := string(content)
	upstreams := make(map[string]interface{})
	
	// 使用正则表达式查找upstream块
	upstreamRegex := regexp.MustCompile(`upstream\s+(\w+)\s*{([^}]+)}`)
	matches := upstreamRegex.FindAllStringSubmatch(configStr, -1)
	
	for _, match := range matches {
		if len(match) >= 3 {
			upstreamName := match[1]
			upstreamContent := match[2]
			
			upstreamInfo := map[string]interface{}{
				"name": upstreamName,
			}
			
			// 解析服务器列表
			serverRegex := regexp.MustCompile(`server\s+([^;\s]+)([^;]*);`)
			serverMatches := serverRegex.FindAllStringSubmatch(upstreamContent, -1)
			
			servers := make([]map[string]interface{}, 0)
			for _, serverMatch := range serverMatches {
				if len(serverMatch) >= 2 {
					server := map[string]interface{}{
						"address": serverMatch[1],
						"status":  "unknown", // 无法从配置文件确定状态
					}
					
					// 解析服务器参数
					if len(serverMatch) >= 3 {
						params := strings.TrimSpace(serverMatch[2])
						if strings.Contains(params, "weight=") {
							if weightMatch := regexp.MustCompile(`weight=(\d+)`).FindStringSubmatch(params); len(weightMatch) > 1 {
								if weight, err := strconv.Atoi(weightMatch[1]); err == nil {
									server["weight"] = weight
								}
							}
						}
						if strings.Contains(params, "backup") {
							server["backup"] = true
						}
						if strings.Contains(params, "down") {
							server["status"] = "down"
						}
					}
					
					servers = append(servers, server)
				}
			}
			
			upstreamInfo["servers"] = servers
			upstreamInfo["server_count"] = len(servers)
			
			upstreams[upstreamName] = upstreamInfo
		}
	}
	
	result["upstreams"] = upstreams
	result["upstream_count"] = len(upstreams)
	result["upstreams_available"] = len(upstreams) > 0
	result["source"] = "config_file"
	
	return result, nil
}

// getSSLInfo 获取真实的SSL信息
func (m *WindowsNginxMonitor) getSSLInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 从配置文件解析SSL配置
	content, err := ioutil.ReadFile(m.confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	configStr := string(content)
	sslCertificates := make([]map[string]interface{}, 0)
	
	// 查找SSL证书配置
	certRegex := regexp.MustCompile(`ssl_certificate\s+([^;]+);`)
	certMatches := certRegex.FindAllStringSubmatch(configStr, -1)
	
	keyRegex := regexp.MustCompile(`ssl_certificate_key\s+([^;]+);`)
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
	if matches := regexp.MustCompile(`ssl_protocols\s+([^;]+);`).FindStringSubmatch(configStr); len(matches) > 1 {
		protocols := strings.Fields(strings.TrimSpace(matches[1]))
		result["ssl_protocols"] = protocols
	}
	
	// 检查SSL密码套件
	if matches := regexp.MustCompile(`ssl_ciphers\s+([^;]+);`).FindStringSubmatch(configStr); len(matches) > 1 {
		result["ssl_ciphers"] = strings.TrimSpace(matches[1])
	}
	
	return result, nil
}

// getCertificateDetails 获取证书详细信息
func (m *WindowsNginxMonitor) getCertificateDetails(certPath string) (map[string]interface{}, error) {
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
func (m *WindowsNginxMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// 添加基本信息
	metrics["nginx_path"] = m.nginxPath
	metrics["config_path"] = m.confPath
	metrics["log_path"] = m.logPath
	metrics["collection_time"] = time.Now().Format(time.RFC3339)
	
	// 使用Windows特定的方法收集指标
	if processInfo, err := m.getNginxProcessInfo(); err == nil {
		for k, v := range processInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get process info: %v", err)
		// 如果获取进程信息失败，设置基本状态
		metrics["is_running"] = false
		metrics["process_error"] = err.Error()
	}
	
	if versionInfo, err := m.getNginxVersion(); err == nil {
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
	
	if requestStats, err := m.getRequestStats(); err == nil {
		for k, v := range requestStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get request stats: %v", err)
	}
	
	if upstreamInfo, err := m.getUpstreamInfo(); err == nil {
		for k, v := range upstreamInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get upstream info: %v", err)
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
		return m.NginxMonitor.collectMetrics()
	}
	
	return metrics
}