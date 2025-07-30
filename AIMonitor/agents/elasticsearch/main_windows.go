//go:build windows
// +build windows

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"aimonitor-agents/common"
)

// WindowsElasticsearchMonitor Windows版本的Elasticsearch监控器
type WindowsElasticsearchMonitor struct {
	*ElasticsearchMonitor
	elasticsearchPath string
	configPath        string
	dataPath          string
	logPath           string
	httpClient        *http.Client
}

// NewWindowsElasticsearchMonitor 创建Windows版本的Elasticsearch监控器
func NewWindowsElasticsearchMonitor(agent *common.Agent) *WindowsElasticsearchMonitor {
	baseMonitor := NewElasticsearchMonitor(agent)
	
	// 默认Windows路径
	elasticsearchPath := "C:\\elasticsearch"
	configPath := "C:\\elasticsearch\\config\\elasticsearch.yml"
	dataPath := "C:\\elasticsearch\\data"
	logPath := "C:\\elasticsearch\\logs"
	
	// 从配置中获取路径
	if path, exists := agent.Config["elasticsearch_path"]; exists {
		if pathStr, ok := path.(string); ok {
			elasticsearchPath = pathStr
		}
	}
	
	if path, exists := agent.Config["config_path"]; exists {
		if pathStr, ok := path.(string); ok {
			configPath = pathStr
		}
	}
	
	if path, exists := agent.Config["data_path"]; exists {
		if pathStr, ok := path.(string); ok {
			dataPath = pathStr
		}
	}
	
	if path, exists := agent.Config["log_path"]; exists {
		if pathStr, ok := path.(string); ok {
			logPath = pathStr
		}
	}
	
	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	
	return &WindowsElasticsearchMonitor{
		ElasticsearchMonitor: baseMonitor,
		elasticsearchPath:    elasticsearchPath,
		configPath:           configPath,
		dataPath:             dataPath,
		logPath:              logPath,
		httpClient:           httpClient,
	}
}

// Windows API 结构体和常量
type PROCESS_MEMORY_COUNTERS struct {
	Cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	psapi                   = syscall.NewLazyDLL("psapi.dll")
	getCurrentProcess       = kernel32.NewProc("GetCurrentProcess")
	getProcessMemoryInfo    = psapi.NewProc("GetProcessMemoryInfo")
	getProcessTimes         = kernel32.NewProc("GetProcessTimes")
	openProcess             = kernel32.NewProc("OpenProcess")
	closeHandle             = kernel32.NewProc("CloseHandle")
	getProcessImageFileName = psapi.NewProc("GetProcessImageFileNameW")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

// getElasticsearchProcessInfo 获取Elasticsearch进程信息
func (m *WindowsElasticsearchMonitor) getElasticsearchProcessInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 使用tasklist命令查找Elasticsearch进程
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq elasticsearch.exe", "/FO", "CSV")
	output, err := cmd.Output()
	if err != nil {
		// 尝试查找java进程中的Elasticsearch
		cmd = exec.Command("tasklist", "/FI", "IMAGENAME eq java.exe", "/FO", "CSV")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get process list: %v", err)
		}
	}
	
	lines := strings.Split(string(output), "\n")
	processes := make([]map[string]interface{}, 0)
	
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // 跳过标题行和空行
		}
		
		// 解析CSV格式的输出
		fields := strings.Split(line, ",")
		if len(fields) >= 5 {
			imageName := strings.Trim(fields[0], `"`)
			pidStr := strings.Trim(fields[1], `"`)
			memUsageStr := strings.Trim(fields[4], `"`)
			
			// 检查是否是Elasticsearch相关进程
			if strings.Contains(strings.ToLower(imageName), "elasticsearch") ||
				(strings.Contains(strings.ToLower(imageName), "java") && m.isElasticsearchJavaProcess(pidStr)) {
				
				pid, _ := strconv.Atoi(pidStr)
				memUsage := m.parseMemoryUsage(memUsageStr)
				
				processInfo := map[string]interface{}{
					"image_name":   imageName,
					"pid":          pid,
					"memory_usage": memUsage,
				}
				
				// 获取更详细的进程信息
				if detailedInfo, err := m.getDetailedProcessInfo(pid); err == nil {
					for k, v := range detailedInfo {
						processInfo[k] = v
					}
				}
				
				processes = append(processes, processInfo)
			}
		}
	}
	
	result["is_running"] = len(processes) > 0
	result["process_count"] = len(processes)
	result["processes"] = processes
	
	if len(processes) > 0 {
		// 计算总内存使用量
		totalMemory := int64(0)
		for _, proc := range processes {
			if mem, ok := proc["memory_usage"].(int64); ok {
				totalMemory += mem
			}
		}
		result["total_memory_usage"] = totalMemory
		result["total_memory_usage_mb"] = totalMemory / 1024 / 1024
	}
	
	return result, nil
}

// isElasticsearchJavaProcess 检查Java进程是否是Elasticsearch
func (m *WindowsElasticsearchMonitor) isElasticsearchJavaProcess(pidStr string) bool {
	// 使用wmic获取进程命令行
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%s", pidStr), "get", "CommandLine", "/format:value")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	cmdLine := string(output)
	return strings.Contains(strings.ToLower(cmdLine), "elasticsearch") ||
		strings.Contains(strings.ToLower(cmdLine), "org.elasticsearch")
}

// parseMemoryUsage 解析内存使用量字符串
func (m *WindowsElasticsearchMonitor) parseMemoryUsage(memStr string) int64 {
	// 移除逗号和"K"后缀
	memStr = strings.ReplaceAll(memStr, ",", "")
	memStr = strings.ReplaceAll(memStr, " K", "")
	memStr = strings.TrimSpace(memStr)
	
	if mem, err := strconv.ParseInt(memStr, 10, 64); err == nil {
		return mem * 1024 // 转换为字节
	}
	return 0
}

// getDetailedProcessInfo 获取详细的进程信息
func (m *WindowsElasticsearchMonitor) getDetailedProcessInfo(pid int) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 打开进程句柄
	handle, _, _ := openProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		uintptr(0),
		uintptr(pid),
	)
	
	if handle == 0 {
		return result, fmt.Errorf("failed to open process %d", pid)
	}
	defer closeHandle.Call(handle)
	
	// 获取内存信息
	var memCounters PROCESS_MEMORY_COUNTERS
	memCounters.Cb = uint32(unsafe.Sizeof(memCounters))
	
	ret, _, _ := getProcessMemoryInfo.Call(
		handle,
		uintptr(unsafe.Pointer(&memCounters)),
		uintptr(memCounters.Cb),
	)
	
	if ret != 0 {
		result["working_set_size"] = int64(memCounters.WorkingSetSize)
		result["peak_working_set_size"] = int64(memCounters.PeakWorkingSetSize)
		result["pagefile_usage"] = int64(memCounters.PagefileUsage)
		result["peak_pagefile_usage"] = int64(memCounters.PeakPagefileUsage)
		result["page_fault_count"] = int64(memCounters.PageFaultCount)
	}
	
	// 获取进程时间信息
	var creationTime, exitTime, kernelTime, userTime FILETIME
	ret, _, _ = getProcessTimes.Call(
		handle,
		uintptr(unsafe.Pointer(&creationTime)),
		uintptr(unsafe.Pointer(&exitTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	
	if ret != 0 {
		// 转换FILETIME到时间
		creationTimeNs := (int64(creationTime.DwHighDateTime)<<32 + int64(creationTime.DwLowDateTime)) * 100
		creationTimeUnix := (creationTimeNs - 116444736000000000) / 10000000
		result["creation_time"] = time.Unix(creationTimeUnix, 0).Format(time.RFC3339)
		
		// 计算运行时间
		runningTime := time.Since(time.Unix(creationTimeUnix, 0))
		result["running_time_seconds"] = int64(runningTime.Seconds())
		result["running_time_formatted"] = runningTime.String()
		
		// CPU时间
		kernelTimeNs := (int64(kernelTime.DwHighDateTime)<<32 + int64(kernelTime.DwLowDateTime)) * 100
		userTimeNs := (int64(userTime.DwHighDateTime)<<32 + int64(userTime.DwLowDateTime)) * 100
		result["kernel_time_ns"] = kernelTimeNs
		result["user_time_ns"] = userTimeNs
		result["total_cpu_time_ns"] = kernelTimeNs + userTimeNs
	}
	
	return result, nil
}

// getElasticsearchVersion 获取Elasticsearch版本信息
func (m *WindowsElasticsearchMonitor) getElasticsearchVersion() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试通过HTTP API获取版本信息
	if versionInfo, err := m.getVersionFromAPI(); err == nil {
		return versionInfo, nil
	}
	
	// 尝试通过命令行获取版本
	elasticsearchExe := m.elasticsearchPath + "\\bin\\elasticsearch.exe"
	cmd := exec.Command(elasticsearchExe, "--version")
	output, err := cmd.Output()
	if err != nil {
		// 尝试java命令
		cmd = exec.Command("java", "-cp", m.elasticsearchPath+"\\lib\\*", "org.elasticsearch.Version")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get version: %v", err)
		}
	}
	
	versionStr := string(output)
	lines := strings.Split(versionStr, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 解析版本信息
		if strings.Contains(line, "Version:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				result["version"] = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Build:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				result["build_hash"] = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "JVM:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				result["jvm_version"] = strings.TrimSpace(parts[1])
			}
		}
	}
	
	// 如果没有解析到版本信息，尝试从整个输出中提取
	if _, exists := result["version"]; !exists {
		versionRegex := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
		if matches := versionRegex.FindStringSubmatch(versionStr); len(matches) > 1 {
			result["version"] = matches[1]
		}
	}
	
	return result, nil
}

// getVersionFromAPI 通过API获取版本信息
func (m *WindowsElasticsearchMonitor) getVersionFromAPI() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s://%s:%d", m.agent.Config["protocol"], m.agent.Config["host"], m.agent.Config["port"])
	
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}
	
	result := make(map[string]interface{})
	if version, exists := apiResponse["version"]; exists {
		if versionMap, ok := version.(map[string]interface{}); ok {
			for k, v := range versionMap {
				result[k] = v
			}
		}
	}
	
	return result, nil
}

// getConfigInfo 获取配置信息
func (m *WindowsElasticsearchMonitor) getConfigInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 检查配置文件是否存在
	if info, err := os.Stat(m.configPath); err == nil {
		result["config_file_exists"] = true
		result["config_file_size"] = info.Size()
		result["config_file_modified"] = info.ModTime().Format(time.RFC3339)
		
		// 读取配置文件内容
		if content, err := ioutil.ReadFile(m.configPath); err == nil {
			configStr := string(content)
			result["config_lines"] = len(strings.Split(configStr, "\n"))
			
			// 解析一些关键配置
			if matches := regexp.MustCompile(`cluster\.name:\s*(.+)`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["cluster_name"] = strings.TrimSpace(matches[1])
			}
			
			if matches := regexp.MustCompile(`node\.name:\s*(.+)`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["node_name"] = strings.TrimSpace(matches[1])
			}
			
			if matches := regexp.MustCompile(`path\.data:\s*(.+)`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["data_path_config"] = strings.TrimSpace(matches[1])
			}
			
			if matches := regexp.MustCompile(`path\.logs:\s*(.+)`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["logs_path_config"] = strings.TrimSpace(matches[1])
			}
			
			if matches := regexp.MustCompile(`network\.host:\s*(.+)`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["network_host"] = strings.TrimSpace(matches[1])
			}
			
			if matches := regexp.MustCompile(`http\.port:\s*(.+)`).FindStringSubmatch(configStr); len(matches) > 1 {
				result["http_port"] = strings.TrimSpace(matches[1])
			}
		} else {
			result["config_read_error"] = err.Error()
		}
	} else {
		result["config_file_exists"] = false
		result["config_file_error"] = err.Error()
	}
	
	// 检查数据目录
	if info, err := os.Stat(m.dataPath); err == nil {
		result["data_dir_exists"] = true
		result["data_dir_modified"] = info.ModTime().Format(time.RFC3339)
		
		// 计算数据目录大小
		if size, err := m.getDirSize(m.dataPath); err == nil {
			result["data_dir_size"] = size
			result["data_dir_size_mb"] = size / 1024 / 1024
		}
	} else {
		result["data_dir_exists"] = false
	}
	
	// 检查日志目录
	if info, err := os.Stat(m.logPath); err == nil {
		result["log_dir_exists"] = true
		result["log_dir_modified"] = info.ModTime().Format(time.RFC3339)
		
		// 计算日志目录大小
		if size, err := m.getDirSize(m.logPath); err == nil {
			result["log_dir_size"] = size
			result["log_dir_size_mb"] = size / 1024 / 1024
		}
	} else {
		result["log_dir_exists"] = false
	}
	
	return result, nil
}

// getDirSize 计算目录大小
func (m *WindowsElasticsearchMonitor) getDirSize(path string) (int64, error) {
	cmd := exec.Command("powershell", "-Command", 
		fmt.Sprintf("(Get-ChildItem -Path '%s' -Recurse -File | Measure-Object -Property Length -Sum).Sum", path))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	sizeStr := strings.TrimSpace(string(output))
	return strconv.ParseInt(sizeStr, 10, 64)
}

// sendHTTPRequest 发送HTTP请求
func (m *WindowsElasticsearchMonitor) sendHTTPRequest(endpoint string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s://%s:%d%s", 
		m.agent.Config["protocol"], 
		m.agent.Config["host"], 
		m.agent.Config["port"], 
		endpoint)
	
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	return result, nil
}

// getClusterHealth 获取集群健康状态
func (m *WindowsElasticsearchMonitor) getClusterHealth() (map[string]interface{}, error) {
	return m.sendHTTPRequest("/_cluster/health")
}

// getNodeStats 获取节点统计信息
func (m *WindowsElasticsearchMonitor) getNodeStats() (map[string]interface{}, error) {
	return m.sendHTTPRequest("/_nodes/stats")
}

// getIndexStats 获取索引统计信息
func (m *WindowsElasticsearchMonitor) getIndexStats() (map[string]interface{}, error) {
	return m.sendHTTPRequest("/_stats")
}

// getFilesystemStats 获取文件系统统计信息
func (m *WindowsElasticsearchMonitor) getFilesystemStats() (map[string]interface{}, error) {
	return m.sendHTTPRequest("/_nodes/stats/fs")
}

// getThreadPoolStats 获取线程池统计信息
func (m *WindowsElasticsearchMonitor) getThreadPoolStats() (map[string]interface{}, error) {
	return m.sendHTTPRequest("/_nodes/stats/thread_pool")
}

// getCacheStats 获取缓存统计信息
func (m *WindowsElasticsearchMonitor) getCacheStats() (map[string]interface{}, error) {
	return m.sendHTTPRequest("/_nodes/stats/indices/query_cache,request_cache,fielddata")
}

// 重写collectMetrics方法以使用Windows特定的实现
func (m *WindowsElasticsearchMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// 添加基本信息
	metrics["elasticsearch_path"] = m.elasticsearchPath
	metrics["config_path"] = m.configPath
	metrics["data_path"] = m.dataPath
	metrics["log_path"] = m.logPath
	metrics["collection_time"] = time.Now().Format(time.RFC3339)
	
	// 使用Windows特定的方法收集指标
	if processInfo, err := m.getElasticsearchProcessInfo(); err == nil {
		for k, v := range processInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get process info: %v", err)
		// 如果获取进程信息失败，设置基本状态
		metrics["is_running"] = false
		metrics["process_error"] = err.Error()
	}
	
	if versionInfo, err := m.getElasticsearchVersion(); err == nil {
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
	
	// 如果Elasticsearch正在运行，尝试获取API数据
	if isRunning, ok := metrics["is_running"].(bool); ok && isRunning {
		if clusterHealth, err := m.getClusterHealth(); err == nil {
			metrics["cluster_health"] = clusterHealth
		} else {
			m.agent.Logger.Error("Failed to get cluster health: %v", err)
		}
		
		if nodeStats, err := m.getNodeStats(); err == nil {
			metrics["node_stats"] = nodeStats
		} else {
			m.agent.Logger.Error("Failed to get node stats: %v", err)
		}
		
		if indexStats, err := m.getIndexStats(); err == nil {
			metrics["index_stats"] = indexStats
		} else {
			m.agent.Logger.Error("Failed to get index stats: %v", err)
		}
		
		if fsStats, err := m.getFilesystemStats(); err == nil {
			metrics["filesystem_stats"] = fsStats
		} else {
			m.agent.Logger.Error("Failed to get filesystem stats: %v", err)
		}
		
		if threadPoolStats, err := m.getThreadPoolStats(); err == nil {
			metrics["thread_pool_stats"] = threadPoolStats
		} else {
			m.agent.Logger.Error("Failed to get thread pool stats: %v", err)
		}
		
		if cacheStats, err := m.getCacheStats(); err == nil {
			metrics["cache_stats"] = cacheStats
		} else {
			m.agent.Logger.Error("Failed to get cache stats: %v", err)
		}
	}
	
	// 如果大部分指标收集失败，回退到模拟数据
	if len(metrics) <= 5 { // 只有基本信息
		m.agent.Logger.Warn("Most metric collection failed, falling back to simulated data")
		return m.ElasticsearchMonitor.collectMetrics()
	}
	
	return metrics
}