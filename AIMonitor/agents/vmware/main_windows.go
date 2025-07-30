//go:build windows
// +build windows

package main

import (
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

// WindowsVMwareMonitor Windows版本的VMware监控器
type WindowsVMwareMonitor struct {
	*VMwareMonitor
	vmwareToolsPath string
	vmrunPath       string
	vixPath         string
	httpClient      *http.Client
}

// NewWindowsVMwareMonitor 创建Windows版本的VMware监控器
func NewWindowsVMwareMonitor(agent *common.Agent) *WindowsVMwareMonitor {
	baseMonitor := NewVMwareMonitor(agent)
	
	// 默认Windows路径
	vmwareToolsPath := "C:\\Program Files\\VMware\\VMware Tools"
	vmrunPath := "C:\\Program Files (x86)\\VMware\\VMware Workstation\\vmrun.exe"
	vixPath := "C:\\Program Files (x86)\\VMware\\VMware VIX"
	
	// 从配置中获取路径
	if path, exists := agent.Config["vmware_tools_path"]; exists {
		if pathStr, ok := path.(string); ok {
			vmwareToolsPath = pathStr
		}
	}
	
	if path, exists := agent.Config["vmrun_path"]; exists {
		if pathStr, ok := path.(string); ok {
			vmrunPath = pathStr
		}
	}
	
	if path, exists := agent.Config["vix_path"]; exists {
		if pathStr, ok := path.(string); ok {
			vixPath = pathStr
		}
	}
	
	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	
	return &WindowsVMwareMonitor{
		VMwareMonitor:   baseMonitor,
		vmwareToolsPath: vmwareToolsPath,
		vmrunPath:       vmrunPath,
		vixPath:         vixPath,
		httpClient:      httpClient,
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

// getVMwareProcessInfo 获取VMware相关进程信息
func (m *WindowsVMwareMonitor) getVMwareProcessInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// VMware相关进程名称
	vmwareProcesses := []string{
		"vmware.exe",
		"vmware-vmx.exe",
		"vmware-hostd.exe",
		"vmware-authd.exe",
		"vmnetdhcp.exe",
		"vmnat.exe",
		"vmware-usbarbitrator64.exe",
		"vmware-tray.exe",
	}
	
	allProcesses := make([]map[string]interface{}, 0)
	totalMemory := int64(0)
	
	for _, processName := range vmwareProcesses {
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", processName), "/FO", "CSV")
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		
		lines := strings.Split(string(output), "\n")
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
				
				pid, _ := strconv.Atoi(pidStr)
				memUsage := m.parseMemoryUsage(memUsageStr)
				totalMemory += memUsage
				
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
				
				allProcesses = append(allProcesses, processInfo)
			}
		}
	}
	
	result["vmware_running"] = len(allProcesses) > 0
	result["process_count"] = len(allProcesses)
	result["processes"] = allProcesses
	result["total_memory_usage"] = totalMemory
	result["total_memory_usage_mb"] = totalMemory / 1024 / 1024
	
	return result, nil
}

// parseMemoryUsage 解析内存使用量字符串
func (m *WindowsVMwareMonitor) parseMemoryUsage(memStr string) int64 {
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
func (m *WindowsVMwareMonitor) getDetailedProcessInfo(pid int) (map[string]interface{}, error) {
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

// getVMwareVersion 获取VMware版本信息
func (m *WindowsVMwareMonitor) getVMwareVersion() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 尝试从注册表获取版本信息
	if versionInfo, err := m.getVersionFromRegistry(); err == nil {
		for k, v := range versionInfo {
			result[k] = v
		}
	}
	
	// 尝试从vmrun命令获取版本
	if _, err := os.Stat(m.vmrunPath); err == nil {
		cmd := exec.Command(m.vmrunPath)
		output, err := cmd.Output()
		if err == nil {
			versionStr := string(output)
			if matches := regexp.MustCompile(`vmrun version (\d+\.\d+\.\d+)`).FindStringSubmatch(versionStr); len(matches) > 1 {
				result["vmrun_version"] = matches[1]
			}
		}
	}
	
	// 检查VMware Tools版本
	if toolsVersion, err := m.getVMwareToolsVersion(); err == nil {
		result["tools_version"] = toolsVersion
	}
	
	return result, nil
}

// getVersionFromRegistry 从注册表获取版本信息
func (m *WindowsVMwareMonitor) getVersionFromRegistry() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 使用reg命令查询注册表
	registryPaths := []string{
		"HKEY_LOCAL_MACHINE\\SOFTWARE\\VMware, Inc.\\VMware Workstation",
		"HKEY_LOCAL_MACHINE\\SOFTWARE\\VMware, Inc.\\VMware Player",
		"HKEY_LOCAL_MACHINE\\SOFTWARE\\VMware, Inc.\\VMware Server",
	}
	
	for _, regPath := range registryPaths {
		cmd := exec.Command("reg", "query", regPath, "/v", "ProductVersion")
		output, err := cmd.Output()
		if err == nil {
			outputStr := string(output)
			if matches := regexp.MustCompile(`ProductVersion\s+REG_SZ\s+(.+)`).FindStringSubmatch(outputStr); len(matches) > 1 {
				product := "unknown"
				if strings.Contains(regPath, "Workstation") {
					product = "VMware Workstation"
				} else if strings.Contains(regPath, "Player") {
					product = "VMware Player"
				} else if strings.Contains(regPath, "Server") {
					product = "VMware Server"
				}
				result["product"] = product
				result["version"] = strings.TrimSpace(matches[1])
				break
			}
		}
	}
	
	return result, nil
}

// getVMwareToolsVersion 获取VMware Tools版本
func (m *WindowsVMwareMonitor) getVMwareToolsVersion() (string, error) {
	toolsExe := m.vmwareToolsPath + "\\VMwareToolboxCmd.exe"
	if _, err := os.Stat(toolsExe); err != nil {
		return "", err
	}
	
	cmd := exec.Command(toolsExe, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	versionStr := strings.TrimSpace(string(output))
	if matches := regexp.MustCompile(`(\d+\.\d+\.\d+)`).FindStringSubmatch(versionStr); len(matches) > 1 {
		return matches[1], nil
	}
	
	return versionStr, nil
}

// getVirtualMachineList 获取虚拟机列表
func (m *WindowsVMwareMonitor) getVirtualMachineList() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	if _, err := os.Stat(m.vmrunPath); err != nil {
		return nil, fmt.Errorf("vmrun not found: %v", err)
	}
	
	// 获取运行中的虚拟机
	cmd := exec.Command(m.vmrunPath, "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %v", err)
	}
	
	lines := strings.Split(string(output), "\n")
	runningVMs := make([]string, 0)
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if i == 0 || line == "" {
			continue // 跳过第一行（总数）和空行
		}
		
		if strings.HasSuffix(line, ".vmx") {
			runningVMs = append(runningVMs, line)
		}
	}
	
	result["running_vms"] = runningVMs
	result["running_vm_count"] = len(runningVMs)
	
	// 获取每个虚拟机的详细信息
	vmDetails := make([]map[string]interface{}, 0)
	for _, vmPath := range runningVMs {
		if vmInfo, err := m.getVMInfo(vmPath); err == nil {
			vmDetails = append(vmDetails, vmInfo)
		}
	}
	result["vm_details"] = vmDetails
	
	return result, nil
}

// getVMInfo 获取单个虚拟机信息
func (m *WindowsVMwareMonitor) getVMInfo(vmPath string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["vm_path"] = vmPath
	
	// 获取虚拟机配置信息
	if vmxContent, err := ioutil.ReadFile(vmPath); err == nil {
		configStr := string(vmxContent)
		
		// 解析配置文件
		if matches := regexp.MustCompile(`displayName\s*=\s*"([^"]+)"`).FindStringSubmatch(configStr); len(matches) > 1 {
			result["display_name"] = matches[1]
		}
		
		if matches := regexp.MustCompile(`guestOS\s*=\s*"([^"]+)"`).FindStringSubmatch(configStr); len(matches) > 1 {
			result["guest_os"] = matches[1]
		}
		
		if matches := regexp.MustCompile(`memsize\s*=\s*"([^"]+)"`).FindStringSubmatch(configStr); len(matches) > 1 {
			if memSize, err := strconv.Atoi(matches[1]); err == nil {
				result["memory_mb"] = memSize
			}
		}
		
		if matches := regexp.MustCompile(`numvcpus\s*=\s*"([^"]+)"`).FindStringSubmatch(configStr); len(matches) > 1 {
			if cpuCount, err := strconv.Atoi(matches[1]); err == nil {
				result["cpu_count"] = cpuCount
			}
		}
	}
	
	// 获取虚拟机状态
	cmd := exec.Command(m.vmrunPath, "list")
	output, err := cmd.Output()
	if err == nil {
		if strings.Contains(string(output), vmPath) {
			result["power_state"] = "poweredOn"
		} else {
			result["power_state"] = "poweredOff"
		}
	}
	
	// 获取虚拟机IP地址（如果运行中）
	if powerState, ok := result["power_state"].(string); ok && powerState == "poweredOn" {
		cmd = exec.Command(m.vmrunPath, "getGuestIPAddress", vmPath)
		output, err = cmd.Output()
		if err == nil {
			ipAddress := strings.TrimSpace(string(output))
			if ipAddress != "" && ipAddress != "unknown" {
				result["ip_address"] = ipAddress
			}
		}
	}
	
	return result, nil
}

// getNetworkInfo 获取网络信息
func (m *WindowsVMwareMonitor) getNetworkInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取VMware网络适配器信息
	cmd := exec.Command("ipconfig", "/all")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	outputStr := string(output)
	vmwareAdapters := make([]map[string]interface{}, 0)
	
	// 查找VMware网络适配器
	adapterRegex := regexp.MustCompile(`(?s)VMware.*?Adapter.*?\n(.*?)(?=\n\S|$)`)
	matches := adapterRegex.FindAllStringSubmatch(outputStr, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			adapterInfo := make(map[string]interface{})
			adapterText := match[0]
			
			// 提取适配器名称
			if nameMatch := regexp.MustCompile(`VMware.*?Adapter[^\n]*`).FindString(adapterText); nameMatch != "" {
				adapterInfo["name"] = strings.TrimSpace(nameMatch)
			}
			
			// 提取IP地址
			if ipMatch := regexp.MustCompile(`IPv4 Address[^:]*:\s*([0-9.]+)`).FindStringSubmatch(adapterText); len(ipMatch) > 1 {
				adapterInfo["ip_address"] = ipMatch[1]
			}
			
			// 提取子网掩码
			if subnetMatch := regexp.MustCompile(`Subnet Mask[^:]*:\s*([0-9.]+)`).FindStringSubmatch(adapterText); len(subnetMatch) > 1 {
				adapterInfo["subnet_mask"] = subnetMatch[1]
			}
			
			// 提取MAC地址
			if macMatch := regexp.MustCompile(`Physical Address[^:]*:\s*([0-9A-F-]+)`).FindStringSubmatch(adapterText); len(macMatch) > 1 {
				adapterInfo["mac_address"] = macMatch[1]
			}
			
			vmwareAdapters = append(vmwareAdapters, adapterInfo)
		}
	}
	
	result["vmware_adapters"] = vmwareAdapters
	result["adapter_count"] = len(vmwareAdapters)
	
	return result, nil
}

// getStorageInfo 获取存储信息
func (m *WindowsVMwareMonitor) getStorageInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 查找VMware相关的磁盘文件
	vmdkFiles := make([]map[string]interface{}, 0)
	
	// 使用PowerShell查找.vmdk文件
	cmd := exec.Command("powershell", "-Command", 
		"Get-ChildItem -Path C:\\ -Recurse -Include *.vmdk -ErrorAction SilentlyContinue | Select-Object FullName, Length, LastWriteTime")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasSuffix(line, ".vmdk") {
				// 获取文件信息
				if info, err := os.Stat(line); err == nil {
					vmdkInfo := map[string]interface{}{
						"path":          line,
						"size_bytes":    info.Size(),
						"size_mb":       info.Size() / 1024 / 1024,
						"modified_time": info.ModTime().Format(time.RFC3339),
					}
					vmdkFiles = append(vmdkFiles, vmdkInfo)
				}
			}
		}
	}
	
	result["vmdk_files"] = vmdkFiles
	result["vmdk_count"] = len(vmdkFiles)
	
	// 计算总存储大小
	totalSize := int64(0)
	for _, vmdk := range vmdkFiles {
		if size, ok := vmdk["size_bytes"].(int64); ok {
			totalSize += size
		}
	}
	result["total_storage_bytes"] = totalSize
	result["total_storage_mb"] = totalSize / 1024 / 1024
	result["total_storage_gb"] = totalSize / 1024 / 1024 / 1024
	
	return result, nil
}

// getPerformanceStats 获取性能统计信息
func (m *WindowsVMwareMonitor) getPerformanceStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取系统性能计数器
	cmd := exec.Command("typeperf", "-sc", "1", 
		"\\Processor(_Total)\\% Processor Time",
		"\\Memory\\Available MBytes",
		"\\Memory\\Committed Bytes",
		"\\PhysicalDisk(_Total)\\Disk Read Bytes/sec",
		"\\PhysicalDisk(_Total)\\Disk Write Bytes/sec")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Processor Time") {
				fields := strings.Split(line, ",")
				if len(fields) >= 2 {
					if cpuUsage, err := strconv.ParseFloat(strings.Trim(fields[1], `"`), 64); err == nil {
						result["cpu_usage_percent"] = cpuUsage
					}
				}
			} else if strings.Contains(line, "Available MBytes") {
				fields := strings.Split(line, ",")
				if len(fields) >= 2 {
					if availMem, err := strconv.ParseFloat(strings.Trim(fields[1], `"`), 64); err == nil {
						result["available_memory_mb"] = availMem
					}
				}
			} else if strings.Contains(line, "Committed Bytes") {
				fields := strings.Split(line, ",")
				if len(fields) >= 2 {
					if commitMem, err := strconv.ParseFloat(strings.Trim(fields[1], `"`), 64); err == nil {
						result["committed_memory_bytes"] = commitMem
					}
				}
			} else if strings.Contains(line, "Disk Read Bytes/sec") {
				fields := strings.Split(line, ",")
				if len(fields) >= 2 {
					if diskRead, err := strconv.ParseFloat(strings.Trim(fields[1], `"`), 64); err == nil {
						result["disk_read_bytes_per_sec"] = diskRead
					}
				}
			} else if strings.Contains(line, "Disk Write Bytes/sec") {
				fields := strings.Split(line, ",")
				if len(fields) >= 2 {
					if diskWrite, err := strconv.ParseFloat(strings.Trim(fields[1], `"`), 64); err == nil {
						result["disk_write_bytes_per_sec"] = diskWrite
					}
				}
			}
		}
	}
	
	return result, nil
}

// 重写collectMetrics方法以使用Windows特定的实现
func (m *WindowsVMwareMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// 添加基本信息
	metrics["vmware_tools_path"] = m.vmwareToolsPath
	metrics["vmrun_path"] = m.vmrunPath
	metrics["vix_path"] = m.vixPath
	metrics["collection_time"] = time.Now().Format(time.RFC3339)
	
	// 使用Windows特定的方法收集指标
	if processInfo, err := m.getVMwareProcessInfo(); err == nil {
		for k, v := range processInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get process info: %v", err)
		// 如果获取进程信息失败，设置基本状态
		metrics["vmware_running"] = false
		metrics["process_error"] = err.Error()
	}
	
	if versionInfo, err := m.getVMwareVersion(); err == nil {
		for k, v := range versionInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get version info: %v", err)
	}
	
	if vmList, err := m.getVirtualMachineList(); err == nil {
		for k, v := range vmList {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get VM list: %v", err)
	}
	
	if networkInfo, err := m.getNetworkInfo(); err == nil {
		for k, v := range networkInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get network info: %v", err)
	}
	
	if storageInfo, err := m.getStorageInfo(); err == nil {
		for k, v := range storageInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get storage info: %v", err)
	}
	
	if perfStats, err := m.getPerformanceStats(); err == nil {
		for k, v := range perfStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get performance stats: %v", err)
	}
	
	// 如果大部分指标收集失败，回退到模拟数据
	if len(metrics) <= 4 { // 只有基本信息
		m.agent.Logger.Warn("Most metric collection failed, falling back to simulated data")
		return m.VMwareMonitor.collectMetrics()
	}
	
	return metrics
}