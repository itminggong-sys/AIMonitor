//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"aimonitor-agents/common"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// WindowsHyperVMonitor Windows版本的Hyper-V监控器
type WindowsHyperVMonitor struct {
	*HyperVMonitor
	hyperVPath string
	oleConn    *ole.IUnknown
}

// NewWindowsHyperVMonitor 创建Windows版本的Hyper-V监控器
func NewWindowsHyperVMonitor(agent *common.Agent) *WindowsHyperVMonitor {
	baseMonitor := NewHyperVMonitor(agent)
	
	// 默认Hyper-V路径
	hyperVPath := "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"
	
	// 从配置中获取路径
	if path, exists := agent.Config["hyperv_path"]; exists {
		if pathStr, ok := path.(string); ok {
			hyperVPath = pathStr
		}
	}
	
	return &WindowsHyperVMonitor{
		HyperVMonitor: baseMonitor,
		hyperVPath:    hyperVPath,
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

// initializeOLE 初始化OLE连接
func (m *WindowsHyperVMonitor) initializeOLE() error {
	if m.oleConn != nil {
		return nil // 已经初始化
	}
	
	err := ole.CoInitialize(0)
	if err != nil {
		return fmt.Errorf("failed to initialize OLE: %v", err)
	}
	
	return nil
}

// cleanupOLE 清理OLE连接
func (m *WindowsHyperVMonitor) cleanupOLE() {
	if m.oleConn != nil {
		m.oleConn.Release()
		m.oleConn = nil
	}
	ole.CoUninitialize()
}

// executePowerShellCommand 执行PowerShell命令
func (m *WindowsHyperVMonitor) executePowerShellCommand(command string) (string, error) {
	cmd := exec.Command("powershell", "-Command", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("PowerShell command failed: %v", err)
	}
	return string(output), nil
}

// getHyperVHostInfo 获取Hyper-V主机信息
func (m *WindowsHyperVMonitor) getHyperVHostInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 检查Hyper-V角色是否安装
	command := "Get-WindowsFeature -Name Hyper-V | Select-Object Name, InstallState"
	output, err := m.executePowerShellCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to check Hyper-V feature: %v", err)
	}
	
	result["hyperv_feature_installed"] = strings.Contains(output, "Installed")
	
	// 获取Hyper-V主机信息
	command = "Get-VMHost | Select-Object ComputerName, LogicalProcessorCount, TotalMemory, MemoryCapacity, VirtualHardDiskPath, VirtualMachinePath"
	output, err = m.executePowerShellCommand(command)
	if err == nil {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			
			if strings.Contains(line, "ComputerName") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					result["computer_name"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "LogicalProcessorCount") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if count, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
						result["logical_processor_count"] = count
					}
				}
			} else if strings.Contains(line, "TotalMemory") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if memory, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
						result["total_memory_bytes"] = memory
						result["total_memory_gb"] = memory / 1024 / 1024 / 1024
					}
				}
			} else if strings.Contains(line, "VirtualHardDiskPath") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					result["virtual_hard_disk_path"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "VirtualMachinePath") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					result["virtual_machine_path"] = strings.TrimSpace(parts[1])
				}
			}
		}
	} else {
		m.agent.Logger.Error("Failed to get VM host info: %v", err)
	}
	
	// 获取Hyper-V服务状态
	command = "Get-Service -Name vmms, vmcompute | Select-Object Name, Status"
	output, err = m.executePowerShellCommand(command)
	if err == nil {
		services := make(map[string]string)
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "vmms") {
				if strings.Contains(line, "Running") {
					services["vmms"] = "Running"
				} else {
					services["vmms"] = "Stopped"
				}
			} else if strings.Contains(line, "vmcompute") {
				if strings.Contains(line, "Running") {
					services["vmcompute"] = "Running"
				} else {
					services["vmcompute"] = "Stopped"
				}
			}
		}
		result["hyperv_services"] = services
		result["hyperv_running"] = services["vmms"] == "Running"
	} else {
		result["hyperv_running"] = false
	}
	
	return result, nil
}

// getVirtualMachineInfo 获取虚拟机信息
func (m *WindowsHyperVMonitor) getVirtualMachineInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取所有虚拟机
	command := "Get-VM | Select-Object Name, State, CPUUsage, MemoryAssigned, MemoryDemand, MemoryStatus, Uptime, Status, Version"
	output, err := m.executePowerShellCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM info: %v", err)
	}
	
	vms := make([]map[string]interface{}, 0)
	lines := strings.Split(output, "\n")
	
	var currentVM map[string]interface{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentVM != nil {
				vms = append(vms, currentVM)
				currentVM = nil
			}
			continue
		}
		
		if strings.Contains(line, "Name") && strings.Contains(line, ":") {
			currentVM = make(map[string]interface{})
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				currentVM["name"] = strings.TrimSpace(parts[1])
			}
		} else if currentVM != nil {
			if strings.Contains(line, "State") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentVM["state"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "CPUUsage") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if cpu, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
						currentVM["cpu_usage"] = cpu
					}
				}
			} else if strings.Contains(line, "MemoryAssigned") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if memory, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
						currentVM["memory_assigned"] = memory
						currentVM["memory_assigned_mb"] = memory / 1024 / 1024
					}
				}
			} else if strings.Contains(line, "MemoryDemand") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if memory, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
						currentVM["memory_demand"] = memory
						currentVM["memory_demand_mb"] = memory / 1024 / 1024
					}
				}
			} else if strings.Contains(line, "Uptime") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentVM["uptime"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "Version") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentVM["version"] = strings.TrimSpace(parts[1])
				}
			}
		}
	}
	
	// 添加最后一个VM
	if currentVM != nil {
		vms = append(vms, currentVM)
	}
	
	result["virtual_machines"] = vms
	result["vm_count"] = len(vms)
	
	// 统计VM状态
	stateCount := make(map[string]int)
	for _, vm := range vms {
		if state, ok := vm["state"].(string); ok {
			stateCount[state]++
		}
	}
	result["vm_state_count"] = stateCount
	
	return result, nil
}

// getVirtualSwitchInfo 获取虚拟交换机信息
func (m *WindowsHyperVMonitor) getVirtualSwitchInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	command := "Get-VMSwitch | Select-Object Name, SwitchType, NetAdapterInterfaceDescription, AllowManagementOS"
	output, err := m.executePowerShellCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual switch info: %v", err)
	}
	
	switches := make([]map[string]interface{}, 0)
	lines := strings.Split(output, "\n")
	
	var currentSwitch map[string]interface{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentSwitch != nil {
				switches = append(switches, currentSwitch)
				currentSwitch = nil
			}
			continue
		}
		
		if strings.Contains(line, "Name") && strings.Contains(line, ":") {
			currentSwitch = make(map[string]interface{})
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				currentSwitch["name"] = strings.TrimSpace(parts[1])
			}
		} else if currentSwitch != nil {
			if strings.Contains(line, "SwitchType") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentSwitch["switch_type"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "NetAdapterInterfaceDescription") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentSwitch["net_adapter"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "AllowManagementOS") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentSwitch["allow_management_os"] = strings.TrimSpace(parts[1]) == "True"
				}
			}
		}
	}
	
	// 添加最后一个交换机
	if currentSwitch != nil {
		switches = append(switches, currentSwitch)
	}
	
	result["virtual_switches"] = switches
	result["switch_count"] = len(switches)
	
	// 统计交换机类型
	typeCount := make(map[string]int)
	for _, sw := range switches {
		if swType, ok := sw["switch_type"].(string); ok {
			typeCount[swType]++
		}
	}
	result["switch_type_count"] = typeCount
	
	return result, nil
}

// getStorageInfo 获取存储信息
func (m *WindowsHyperVMonitor) getStorageInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取虚拟硬盘信息
	command := "Get-VHD -Path (Get-VM | Get-VMHardDiskDrive | Select-Object -ExpandProperty Path) | Select-Object Path, Size, FileSize, VhdType"
	output, err := m.executePowerShellCommand(command)
	if err == nil {
		vhds := make([]map[string]interface{}, 0)
		lines := strings.Split(output, "\n")
		
		var currentVHD map[string]interface{}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				if currentVHD != nil {
					vhds = append(vhds, currentVHD)
					currentVHD = nil
				}
				continue
			}
			
			if strings.Contains(line, "Path") && strings.Contains(line, ":") {
				currentVHD = make(map[string]interface{})
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentVHD["path"] = strings.TrimSpace(parts[1])
				}
			} else if currentVHD != nil {
				if strings.Contains(line, "Size") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						if size, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
							currentVHD["size_bytes"] = size
							currentVHD["size_gb"] = size / 1024 / 1024 / 1024
						}
					}
				} else if strings.Contains(line, "FileSize") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						if fileSize, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
							currentVHD["file_size_bytes"] = fileSize
							currentVHD["file_size_gb"] = fileSize / 1024 / 1024 / 1024
						}
					}
				} else if strings.Contains(line, "VhdType") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						currentVHD["vhd_type"] = strings.TrimSpace(parts[1])
					}
				}
			}
		}
		
		// 添加最后一个VHD
		if currentVHD != nil {
			vhds = append(vhds, currentVHD)
		}
		
		result["virtual_hard_disks"] = vhds
		result["vhd_count"] = len(vhds)
		
		// 计算总存储
		totalSize := int64(0)
		totalFileSize := int64(0)
		for _, vhd := range vhds {
			if size, ok := vhd["size_bytes"].(int64); ok {
				totalSize += size
			}
			if fileSize, ok := vhd["file_size_bytes"].(int64); ok {
				totalFileSize += fileSize
			}
		}
		result["total_allocated_storage_bytes"] = totalSize
		result["total_allocated_storage_gb"] = totalSize / 1024 / 1024 / 1024
		result["total_used_storage_bytes"] = totalFileSize
		result["total_used_storage_gb"] = totalFileSize / 1024 / 1024 / 1024
	} else {
		m.agent.Logger.Error("Failed to get VHD info: %v", err)
	}
	
	return result, nil
}

// getReplicationInfo 获取复制信息
func (m *WindowsHyperVMonitor) getReplicationInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	command := "Get-VMReplication | Select-Object VMName, State, Mode, FrequencySec, PrimaryServer, ReplicaServer, Health"
	output, err := m.executePowerShellCommand(command)
	if err != nil {
		// 复制功能可能未配置，这是正常的
		result["replication_enabled"] = false
		result["replication_count"] = 0
		return result, nil
	}
	
	replications := make([]map[string]interface{}, 0)
	lines := strings.Split(output, "\n")
	
	var currentRepl map[string]interface{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentRepl != nil {
				replications = append(replications, currentRepl)
				currentRepl = nil
			}
			continue
		}
		
		if strings.Contains(line, "VMName") && strings.Contains(line, ":") {
			currentRepl = make(map[string]interface{})
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				currentRepl["vm_name"] = strings.TrimSpace(parts[1])
			}
		} else if currentRepl != nil {
			if strings.Contains(line, "State") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentRepl["state"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "Mode") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentRepl["mode"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "FrequencySec") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if freq, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
						currentRepl["frequency_seconds"] = freq
					}
				}
			} else if strings.Contains(line, "PrimaryServer") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentRepl["primary_server"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "ReplicaServer") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentRepl["replica_server"] = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "Health") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentRepl["health"] = strings.TrimSpace(parts[1])
				}
			}
		}
	}
	
	// 添加最后一个复制
	if currentRepl != nil {
		replications = append(replications, currentRepl)
	}
	
	result["replication_enabled"] = len(replications) > 0
	result["replication_count"] = len(replications)
	result["replications"] = replications
	
	return result, nil
}

// getClusterInfo 获取集群信息
func (m *WindowsHyperVMonitor) getClusterInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 检查是否安装了故障转移集群功能
	command := "Get-WindowsFeature -Name Failover-Clustering | Select-Object InstallState"
	output, err := m.executePowerShellCommand(command)
	if err != nil || !strings.Contains(output, "Installed") {
		result["cluster_enabled"] = false
		return result, nil
	}
	
	// 获取集群信息
	command = "Get-Cluster | Select-Object Name, Domain"
	output, err = m.executePowerShellCommand(command)
	if err != nil {
		result["cluster_enabled"] = false
		return result, nil
	}
	
	result["cluster_enabled"] = true
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Name") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				result["cluster_name"] = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Domain") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				result["cluster_domain"] = strings.TrimSpace(parts[1])
			}
		}
	}
	
	// 获取集群节点信息
	command = "Get-ClusterNode | Select-Object Name, State"
	output, err = m.executePowerShellCommand(command)
	if err == nil {
		nodes := make([]map[string]interface{}, 0)
		lines = strings.Split(output, "\n")
		
		var currentNode map[string]interface{}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				if currentNode != nil {
					nodes = append(nodes, currentNode)
					currentNode = nil
				}
				continue
			}
			
			if strings.Contains(line, "Name") && strings.Contains(line, ":") {
				currentNode = make(map[string]interface{})
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentNode["name"] = strings.TrimSpace(parts[1])
				}
			} else if currentNode != nil && strings.Contains(line, "State") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					currentNode["state"] = strings.TrimSpace(parts[1])
				}
			}
		}
		
		// 添加最后一个节点
		if currentNode != nil {
			nodes = append(nodes, currentNode)
		}
		
		result["cluster_nodes"] = nodes
		result["node_count"] = len(nodes)
	}
	
	return result, nil
}

// getPerformanceCounters 获取性能计数器
func (m *WindowsHyperVMonitor) getPerformanceCounters() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取Hyper-V相关的性能计数器
	counters := []string{
		"\\Hyper-V Hypervisor\\Virtual Processors",
		"\\Hyper-V Hypervisor\\Logical Processors",
		"\\Hyper-V Dynamic Memory VM(*)\\Current Pressure",
		"\\Hyper-V Virtual Storage Device(*)\\Read Bytes/sec",
		"\\Hyper-V Virtual Storage Device(*)\\Write Bytes/sec",
		"\\Hyper-V Virtual Network Adapter(*)\\Bytes Received/sec",
		"\\Hyper-V Virtual Network Adapter(*)\\Bytes Sent/sec",
	}
	
	for _, counter := range counters {
		cmd := exec.Command("typeperf", "-sc", "1", counter)
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, ",") && !strings.Contains(line, "PDH") {
					fields := strings.Split(line, ",")
					if len(fields) >= 2 {
						counterName := strings.Trim(fields[0], `"`)
						counterValue := strings.Trim(fields[1], `"`)
						
						// 简化计数器名称
						simpleName := ""
						if strings.Contains(counterName, "Virtual Processors") {
							simpleName = "virtual_processors"
						} else if strings.Contains(counterName, "Logical Processors") {
							simpleName = "logical_processors"
						} else if strings.Contains(counterName, "Current Pressure") {
							simpleName = "memory_pressure"
						} else if strings.Contains(counterName, "Read Bytes/sec") {
							simpleName = "storage_read_bytes_per_sec"
						} else if strings.Contains(counterName, "Write Bytes/sec") {
							simpleName = "storage_write_bytes_per_sec"
						} else if strings.Contains(counterName, "Bytes Received/sec") {
							simpleName = "network_received_bytes_per_sec"
						} else if strings.Contains(counterName, "Bytes Sent/sec") {
							simpleName = "network_sent_bytes_per_sec"
						}
						
						if simpleName != "" {
							if value, err := strconv.ParseFloat(counterValue, 64); err == nil {
								result[simpleName] = value
							}
						}
					}
				}
			}
		}
	}
	
	return result, nil
}

// 重写collectMetrics方法以使用Windows特定的实现
func (m *WindowsHyperVMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// 添加基本信息
	metrics["hyperv_path"] = m.hyperVPath
	metrics["collection_time"] = time.Now().Format(time.RFC3339)
	
	// 初始化OLE（如果需要）
	if err := m.initializeOLE(); err != nil {
		m.agent.Logger.Error("Failed to initialize OLE: %v", err)
	}
	defer m.cleanupOLE()
	
	// 使用Windows特定的方法收集指标
	if hostInfo, err := m.getHyperVHostInfo(); err == nil {
		for k, v := range hostInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get host info: %v", err)
		// 如果获取主机信息失败，设置基本状态
		metrics["hyperv_running"] = false
		metrics["host_error"] = err.Error()
	}
	
	if vmInfo, err := m.getVirtualMachineInfo(); err == nil {
		for k, v := range vmInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get VM info: %v", err)
	}
	
	if switchInfo, err := m.getVirtualSwitchInfo(); err == nil {
		for k, v := range switchInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get virtual switch info: %v", err)
	}
	
	if storageInfo, err := m.getStorageInfo(); err == nil {
		for k, v := range storageInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get storage info: %v", err)
	}
	
	if replInfo, err := m.getReplicationInfo(); err == nil {
		for k, v := range replInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get replication info: %v", err)
	}
	
	if clusterInfo, err := m.getClusterInfo(); err == nil {
		for k, v := range clusterInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get cluster info: %v", err)
	}
	
	if perfCounters, err := m.getPerformanceCounters(); err == nil {
		for k, v := range perfCounters {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get performance counters: %v", err)
	}
	
	// 如果大部分指标收集失败，回退到模拟数据
	if len(metrics) <= 2 { // 只有基本信息
		m.agent.Logger.Warn("Most metric collection failed, falling back to simulated data")
		return m.HyperVMonitor.collectMetrics()
	}
	
	return metrics
}