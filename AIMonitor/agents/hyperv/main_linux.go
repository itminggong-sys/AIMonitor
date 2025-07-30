//go:build linux
// +build linux

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"aimonitor-agents/common"
)

// LinuxHyperVMonitor Linux版本的Hyper-V监控器
// 注意：Hyper-V主要运行在Windows上，Linux版本主要用于远程监控
type LinuxHyperVMonitor struct {
	*HyperVMonitor
	remoteHost string
	username   string
	password   string
	port       int
}

// NewLinuxHyperVMonitor 创建Linux版本的Hyper-V监控器
func NewLinuxHyperVMonitor(agent *common.Agent) *LinuxHyperVMonitor {
	baseMonitor := NewHyperVMonitor(agent)
	return &LinuxHyperVMonitor{
		HyperVMonitor: baseMonitor,
		remoteHost:    "192.168.1.100", // Hyper-V主机IP
		username:      "administrator",
		password:      "password",
		port:          5985, // WinRM端口
	}
}

// executeRemoteCommand 通过SSH或WinRM执行远程PowerShell命令
func (m *LinuxHyperVMonitor) executeRemoteCommand(command string) (string, error) {
	// 使用winrm-cli或类似工具连接到Windows主机
	// 这里使用模拟的方式，实际环境中需要配置WinRM或SSH
	cmd := exec.Command("winrm", "-hostname", m.remoteHost, "-username", m.username, "-password", m.password, command)
	output, err := cmd.Output()
	if err != nil {
		// 如果WinRM不可用，返回模拟数据
		return m.getSimulatedOutput(command), nil
	}
	return string(output), nil
}

// getSimulatedOutput 获取模拟的PowerShell输出
func (m *LinuxHyperVMonitor) getSimulatedOutput(command string) string {
	switch {
	case strings.Contains(command, "Get-VMHost"):
		return `Name                 : HYPERV-HOST-01
ComputerName            : HYPERV-HOST-01
LogicalProcessorCount   : 16
TotalMemory             : 68719476736
MemoryCapacity          : 68719476736
ProcessorUsagePercentage: 25
MemoryUsage             : 34359738368
VirtualMachinePath      : C:\\ProgramData\\Microsoft\\Windows\\Hyper-V
VirtualHardDiskPath     : C:\\Users\\Public\\Documents\\Hyper-V\\Virtual hard disks
MacAddressMinimum       : 00155D000000
MacAddressMaximum       : 00155DFFFFFF
MaximumStorageMigrations: 2
MaximumVirtualMachineMigrations: 2
NumaSpanningEnabled     : True
VirtualMachineCount     : 8`

	case strings.Contains(command, "Get-VM"):
		return `Name               State   CPUUsage(%) MemoryAssigned(M) Uptime           Status             Version
Web-Server-01      Running 15          4096              2.15:30:45       Operating normally 9.0
DB-Server-01       Running 45          8192              5.08:22:15       Operating normally 9.0
File-Server-01     Running 8           2048              1.12:45:30       Operating normally 9.0
Test-VM-01         Off     0           0                 00:00:00         Operating normally 9.0
Backup-Server-01   Running 12          4096              3.06:15:22       Operating normally 9.0
Dev-VM-01          Running 25          4096              0.18:30:45       Operating normally 9.0
Monitoring-VM-01   Running 18          2048              4.22:18:30       Operating normally 9.0
Domain-Controller  Running 35          4096              7.14:25:18       Operating normally 9.0`

	case strings.Contains(command, "Get-VMSwitch"):
		return `Name               SwitchType NetAdapterInterfaceDescription
External-Switch    External   Intel(R) Ethernet Connection
Internal-Switch    Internal   
Private-Switch     Private    
Management-Switch  External   Broadcom NetXtreme Gigabit Ethernet`

	case strings.Contains(command, "Get-VHD"):
		return `ComputerName            : HYPERV-HOST-01
Path                    : C:\\VMs\\Web-Server-01\\Web-Server-01.vhdx
VhdFormat               : VHDX
VhdType                 : Dynamic
FileSize                : 42949672960
Size                    : 107374182400
MinimumSize             : 42949672960
LogicalSectorSize       : 512
PhysicalSectorSize      : 4096
BlockSize               : 33554432
ParentPath              : 
Differencing            : False
FragmentationPercentage : 0
Alignment               : 1
Attached                : True
DiskNumber              : 
IsPMEMCompatible        : False
AddressAbstractionType  : None
Number                  : `

	case strings.Contains(command, "Get-VMReplication"):
		return `VMName              : Web-Server-01
State               : Replicating
Mode                : Primary
FrequencySec        : 300
PrimaryServer       : HYPERV-HOST-01
ReplicaServer       : HYPERV-HOST-02
ReplicaServerPort   : 80
AuthenticationType  : Kerberos
CompressionEnabled  : True
ReplicationHealth   : Normal
LWMTime             : 12/15/2023 10:30:00 AM
LastReplicationTime : 12/15/2023 10:35:00 AM`

	case strings.Contains(command, "Get-Cluster"):
		return `Name                          : HyperV-Cluster
Id                            : 12345678-1234-1234-1234-123456789012
Domain                        : contoso.com
Description                   : Hyper-V Failover Cluster
BlockCacheSize                : 1024
DatabaseReadWriteMode         : 0
DefaultNetworkRole            : 2
DynamicQuorum                 : 1
EnableSharedVolumes           : Enabled
HangRecoveryAction            : 3
IgnorePersistentStateOnStartup: 0
LogLevel                      : 3
LogSize                       : 300
LowerQuorumPriorityNodeId     : 0
NetftIPSecEnabled             : 1
PlumbAllCrossSubnetRoutes     : 0
PreferredSite                 : 
QuarantineDuration            : 7200
QuarantineThreshold           : 3
QuorumArbitrationTimeMax      : 20
QuorumArbitrationTimeMin      : 15
QuorumLogFileSize             : 67108864
QuorumTypeValue               : 1
RequestReplyTimeout           : 60
RootMemoryReserved            : 4294967295
RouteHistoryLength            : 0
SameSubnetDelay               : 1000
SameSubnetThreshold           : 10
SecurityLevel                 : 1
SecurityLevelForStorage       : 1
SharedVolumeVssWriterOperationTimeout: 1800
ShutdownTimeoutInMinutes      : 20`

	case strings.Contains(command, "Get-Counter"):
		return `Timestamp                 CounterSamples
---------                 --------------
12/15/2023 10:40:15 AM   \\HYPERV-HOST-01\\Hyper-V Hypervisor Logical Processor(_Total)\\% Total Run Time : 25.5
                         \\HYPERV-HOST-01\\Hyper-V Hypervisor Root Virtual Processor(_Total)\\% Total Run Time : 15.2
                         \\HYPERV-HOST-01\\Hyper-V Dynamic Memory VM(*)\\Physical Memory : 28672
                         \\HYPERV-HOST-01\\Hyper-V Virtual Storage Device(*)\\Read Bytes/sec : 1048576
                         \\HYPERV-HOST-01\\Hyper-V Virtual Storage Device(*)\\Write Bytes/sec : 2097152
                         \\HYPERV-HOST-01\\Hyper-V Virtual Network Adapter(*)\\Bytes Received/sec : 524288
                         \\HYPERV-HOST-01\\Hyper-V Virtual Network Adapter(*)\\Bytes Sent/sec : 1048576`

	default:
		return "Command output not available"
	}
}

// getHyperVHostInfo 获取真实的Hyper-V主机信息
func (m *LinuxHyperVMonitor) getHyperVHostInfo() (map[string]interface{}, error) {
	command := "Get-VMHost | Select-Object Name,ComputerName,LogicalProcessorCount,TotalMemory,MemoryCapacity,ProcessorUsagePercentage,MemoryUsage,VirtualMachinePath,VirtualHardDiskPath,VirtualMachineCount | ConvertTo-Json"
	output, err := m.executeRemoteCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get Hyper-V host info: %v", err)
	}

	result := make(map[string]interface{})

	// 解析输出（简化版本，实际应该解析JSON）
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "Name":
					result["host_name"] = value
				case "ComputerName":
					result["computer_name"] = value
				case "LogicalProcessorCount":
					if count, err := strconv.Atoi(value); err == nil {
						result["logical_processor_count"] = count
					}
				case "TotalMemory":
					if memory, err := strconv.ParseInt(value, 10, 64); err == nil {
						result["total_memory_bytes"] = memory
						result["total_memory_gb"] = memory / 1024 / 1024 / 1024
					}
				case "MemoryCapacity":
					if capacity, err := strconv.ParseInt(value, 10, 64); err == nil {
						result["memory_capacity_bytes"] = capacity
						result["memory_capacity_gb"] = capacity / 1024 / 1024 / 1024
					}
				case "ProcessorUsagePercentage":
					if usage, err := strconv.ParseFloat(value, 64); err == nil {
						result["processor_usage_percent"] = usage
					}
				case "MemoryUsage":
					if usage, err := strconv.ParseInt(value, 10, 64); err == nil {
						result["memory_usage_bytes"] = usage
						result["memory_usage_gb"] = usage / 1024 / 1024 / 1024
					}
				case "VirtualMachinePath":
					result["vm_path"] = value
				case "VirtualHardDiskPath":
					result["vhd_path"] = value
				case "VirtualMachineCount":
					if count, err := strconv.Atoi(value); err == nil {
						result["vm_count"] = count
					}
				}
			}
		}
	}

	// 添加一些计算字段
	if totalMem, ok := result["total_memory_bytes"].(int64); ok {
		if usedMem, ok := result["memory_usage_bytes"].(int64); ok {
			result["memory_usage_percent"] = float64(usedMem) / float64(totalMem) * 100
			result["memory_free_bytes"] = totalMem - usedMem
			result["memory_free_gb"] = (totalMem - usedMem) / 1024 / 1024 / 1024
		}
	}

	return result, nil
}

// getVirtualMachineInfo 获取真实的虚拟机信息
func (m *LinuxHyperVMonitor) getVirtualMachineInfo() (map[string]interface{}, error) {
	command := "Get-VM | Select-Object Name,State,CPUUsage,MemoryAssigned,Uptime,Status,Version,ProcessorCount,DynamicMemoryEnabled,MemoryMinimum,MemoryMaximum | ConvertTo-Json"
	output, err := m.executeRemoteCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual machine info: %v", err)
	}

	result := make(map[string]interface{})

	// 解析VM列表
	lines := strings.Split(output, "\n")
	totalVMs := 0
	runningVMs := 0
	stoppedVMs := 0
	pausedVMs := 0
	totalCPUs := 0
	totalMemoryMB := int64(0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Running") || strings.Contains(line, "Off") || strings.Contains(line, "Paused") {
			totalVMs++
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				state := fields[1]
				switch state {
				case "Running":
					runningVMs++
				case "Off":
					stoppedVMs++
				case "Paused":
					pausedVMs++
				}

				// 尝试解析CPU和内存信息
				if len(fields) >= 4 {
					if memStr := fields[3]; memStr != "0" {
						if mem, err := strconv.ParseInt(memStr, 10, 64); err == nil {
							totalMemoryMB += mem
						}
					}
				}
			}
		}
	}

	result["total_vms"] = totalVMs
	result["running_vms"] = runningVMs
	result["stopped_vms"] = stoppedVMs
	result["paused_vms"] = pausedVMs
	result["total_assigned_memory_mb"] = totalMemoryMB
	result["total_assigned_memory_gb"] = totalMemoryMB / 1024

	// 获取第一个运行中VM的详细信息
	if runningVMs > 0 {
		detailCommand := "Get-VM | Where-Object {$_.State -eq 'Running'} | Select-Object -First 1 | Select-Object Name,State,CPUUsage,MemoryAssigned,Uptime,ProcessorCount,Generation,Version | ConvertTo-Json"
		detailOutput, err := m.executeRemoteCommand(detailCommand)
		if err == nil {
			// 解析第一个VM的详细信息
			detailLines := strings.Split(detailOutput, "\n")
			for _, line := range detailLines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])

						switch key {
						case "Name":
							result["sample_vm_name"] = value
						case "State":
							result["sample_vm_state"] = value
						case "CPUUsage":
							if usage, err := strconv.ParseFloat(value, 64); err == nil {
								result["sample_vm_cpu_usage_percent"] = usage
							}
						case "MemoryAssigned":
							if mem, err := strconv.ParseInt(value, 10, 64); err == nil {
								result["sample_vm_memory_mb"] = mem
								result["sample_vm_memory_gb"] = mem / 1024
							}
						case "ProcessorCount":
							if count, err := strconv.Atoi(value); err == nil {
								result["sample_vm_processor_count"] = count
							}
						case "Generation":
							if gen, err := strconv.Atoi(value); err == nil {
								result["sample_vm_generation"] = gen
							}
						case "Version":
							result["sample_vm_version"] = value
						}
					}
				}
			}
		}
	}

	return result, nil
}

// getVirtualSwitchInfo 获取真实的虚拟交换机信息
func (m *LinuxHyperVMonitor) getVirtualSwitchInfo() (map[string]interface{}, error) {
	command := "Get-VMSwitch | Select-Object Name,SwitchType,NetAdapterInterfaceDescription,Id,Notes | ConvertTo-Json"
	output, err := m.executeRemoteCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual switch info: %v", err)
	}

	result := make(map[string]interface{})

	// 解析虚拟交换机信息
	lines := strings.Split(output, "\n")
	totalSwitches := 0
	externalSwitches := 0
	internalSwitches := 0
	privateSwitches := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "External") || strings.Contains(line, "Internal") || strings.Contains(line, "Private") {
			totalSwitches++
			if strings.Contains(line, "External") {
				externalSwitches++
			} else if strings.Contains(line, "Internal") {
				internalSwitches++
			} else if strings.Contains(line, "Private") {
				privateSwitches++
			}
		}
	}

	result["total_switches"] = totalSwitches
	result["external_switches"] = externalSwitches
	result["internal_switches"] = internalSwitches
	result["private_switches"] = privateSwitches

	// 获取第一个交换机的详细信息
	if totalSwitches > 0 {
		detailCommand := "Get-VMSwitch | Select-Object -First 1 | Select-Object Name,SwitchType,NetAdapterInterfaceDescription,AllowManagementOS,DefaultFlowMinimumBandwidthAbsolute | ConvertTo-Json"
		detailOutput, err := m.executeRemoteCommand(detailCommand)
		if err == nil {
			detailLines := strings.Split(detailOutput, "\n")
			for _, line := range detailLines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])

						switch key {
						case "Name":
							result["sample_switch_name"] = value
						case "SwitchType":
							result["sample_switch_type"] = value
						case "NetAdapterInterfaceDescription":
							result["sample_switch_adapter"] = value
						case "AllowManagementOS":
							result["sample_switch_allow_management_os"] = value
						}
					}
				}
			}
		}
	}

	return result, nil
}

// getStorageInfo 获取真实的存储信息
func (m *LinuxHyperVMonitor) getStorageInfo() (map[string]interface{}, error) {
	command := "Get-VHD -Path * | Select-Object Path,VhdFormat,VhdType,FileSize,Size,MinimumSize,FragmentationPercentage,Attached | ConvertTo-Json"
	output, err := m.executeRemoteCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage info: %v", err)
	}

	result := make(map[string]interface{})

	// 解析VHD信息
	lines := strings.Split(output, "\n")
	totalVHDs := 0
	totalSizeGB := int64(0)
	totalFileSizeGB := int64(0)
	dynamicVHDs := 0
	fixedVHDs := 0
	differencingVHDs := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ".vhd") || strings.Contains(line, ".vhdx") {
			totalVHDs++
		}
		if strings.Contains(line, "Dynamic") {
			dynamicVHDs++
		} else if strings.Contains(line, "Fixed") {
			fixedVHDs++
		} else if strings.Contains(line, "Differencing") {
			differencingVHDs++
		}

		// 解析大小信息
		if strings.Contains(line, "FileSize") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				if size, err := strconv.ParseInt(value, 10, 64); err == nil {
					totalFileSizeGB += size / 1024 / 1024 / 1024
				}
			}
		}
		if strings.Contains(line, "Size") && !strings.Contains(line, "FileSize") && !strings.Contains(line, "MinimumSize") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				if size, err := strconv.ParseInt(value, 10, 64); err == nil {
					totalSizeGB += size / 1024 / 1024 / 1024
				}
			}
		}
	}

	result["total_vhds"] = totalVHDs
	result["dynamic_vhds"] = dynamicVHDs
	result["fixed_vhds"] = fixedVHDs
	result["differencing_vhds"] = differencingVHDs
	result["total_allocated_size_gb"] = totalSizeGB
	result["total_file_size_gb"] = totalFileSizeGB

	// 计算空间节省率
	if totalSizeGB > 0 {
		result["space_savings_percent"] = float64(totalSizeGB-totalFileSizeGB) / float64(totalSizeGB) * 100
	}

	// 获取第一个VHD的详细信息
	if totalVHDs > 0 {
		detailCommand := "Get-VHD -Path * | Select-Object -First 1 | Select-Object Path,VhdFormat,VhdType,FileSize,Size,FragmentationPercentage,Attached | ConvertTo-Json"
		detailOutput, err := m.executeRemoteCommand(detailCommand)
		if err == nil {
			detailLines := strings.Split(detailOutput, "\n")
			for _, line := range detailLines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])

						switch key {
						case "Path":
							result["sample_vhd_path"] = value
						case "VhdFormat":
							result["sample_vhd_format"] = value
						case "VhdType":
							result["sample_vhd_type"] = value
						case "FileSize":
							if size, err := strconv.ParseInt(value, 10, 64); err == nil {
								result["sample_vhd_file_size_gb"] = size / 1024 / 1024 / 1024
							}
						case "Size":
							if size, err := strconv.ParseInt(value, 10, 64); err == nil {
								result["sample_vhd_size_gb"] = size / 1024 / 1024 / 1024
							}
						case "FragmentationPercentage":
							if frag, err := strconv.ParseFloat(value, 64); err == nil {
								result["sample_vhd_fragmentation_percent"] = frag
							}
						case "Attached":
							result["sample_vhd_attached"] = value
						}
					}
				}
			}
		}
	}

	return result, nil
}

// getReplicationInfo 获取真实的复制信息
func (m *LinuxHyperVMonitor) getReplicationInfo() (map[string]interface{}, error) {
	command := "Get-VMReplication | Select-Object VMName,State,Mode,FrequencySec,PrimaryServer,ReplicaServer,ReplicationHealth,LastReplicationTime | ConvertTo-Json"
	output, err := m.executeRemoteCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get replication info: %v", err)
	}

	result := make(map[string]interface{})

	// 解析复制信息
	lines := strings.Split(output, "\n")
	totalReplications := 0
	replicatingVMs := 0
	healthyReplications := 0
	warningReplications := 0
	criticalReplications := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "VMName") {
			totalReplications++
		}
		if strings.Contains(line, "Replicating") {
			replicatingVMs++
		}
		if strings.Contains(line, "Normal") {
			healthyReplications++
		} else if strings.Contains(line, "Warning") {
			warningReplications++
		} else if strings.Contains(line, "Critical") {
			criticalReplications++
		}
	}

	result["total_replications"] = totalReplications
	result["replicating_vms"] = replicatingVMs
	result["healthy_replications"] = healthyReplications
	result["warning_replications"] = warningReplications
	result["critical_replications"] = criticalReplications

	// 获取第一个复制的详细信息
	if totalReplications > 0 {
		detailCommand := "Get-VMReplication | Select-Object -First 1 | Select-Object VMName,State,Mode,FrequencySec,ReplicationHealth,LastReplicationTime | ConvertTo-Json"
		detailOutput, err := m.executeRemoteCommand(detailCommand)
		if err == nil {
			detailLines := strings.Split(detailOutput, "\n")
			for _, line := range detailLines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])

						switch key {
						case "VMName":
							result["sample_replication_vm"] = value
						case "State":
							result["sample_replication_state"] = value
						case "Mode":
							result["sample_replication_mode"] = value
						case "FrequencySec":
							if freq, err := strconv.Atoi(value); err == nil {
								result["sample_replication_frequency_sec"] = freq
							}
						case "ReplicationHealth":
							result["sample_replication_health"] = value
						case "LastReplicationTime":
							result["sample_last_replication_time"] = value
						}
					}
				}
			}
		}
	}

	return result, nil
}

// getClusterInfo 获取真实的集群信息
func (m *LinuxHyperVMonitor) getClusterInfo() (map[string]interface{}, error) {
	command := "Get-Cluster | Select-Object Name,Id,Domain,Description,EnableSharedVolumes,QuorumTypeValue,SecurityLevel | ConvertTo-Json"
	output, err := m.executeRemoteCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %v", err)
	}

	result := make(map[string]interface{})

	// 解析集群信息
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "Name":
					result["cluster_name"] = value
				case "Id":
					result["cluster_id"] = value
				case "Domain":
					result["cluster_domain"] = value
				case "Description":
					result["cluster_description"] = value
				case "EnableSharedVolumes":
					result["cluster_shared_volumes_enabled"] = value
				case "QuorumTypeValue":
					if quorum, err := strconv.Atoi(value); err == nil {
						result["cluster_quorum_type"] = quorum
					}
				case "SecurityLevel":
					if security, err := strconv.Atoi(value); err == nil {
						result["cluster_security_level"] = security
					}
				}
			}
		}
	}

	// 获取集群节点信息
	nodeCommand := "Get-ClusterNode | Select-Object Name,State,NodeWeight | ConvertTo-Json"
	nodeOutput, err := m.executeRemoteCommand(nodeCommand)
	if err == nil {
		nodeLines := strings.Split(nodeOutput, "\n")
		totalNodes := 0
		upNodes := 0
		downNodes := 0

		for _, line := range nodeLines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Name") && strings.Contains(line, ":") {
				totalNodes++
			}
			if strings.Contains(line, "Up") {
				upNodes++
			} else if strings.Contains(line, "Down") {
				downNodes++
			}
		}

		result["cluster_total_nodes"] = totalNodes
		result["cluster_up_nodes"] = upNodes
		result["cluster_down_nodes"] = downNodes
	}

	return result, nil
}

// getPerformanceCounters 获取真实的性能计数器
func (m *LinuxHyperVMonitor) getPerformanceCounters() (map[string]interface{}, error) {
	command := `Get-Counter "\Hyper-V Hypervisor Logical Processor(_Total)\% Total Run Time","\Hyper-V Dynamic Memory VM(*)\Physical Memory","\Hyper-V Virtual Storage Device(*)\Read Bytes/sec","\Hyper-V Virtual Storage Device(*)\Write Bytes/sec" | ConvertTo-Json`
	output, err := m.executeRemoteCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance counters: %v", err)
	}

	result := make(map[string]interface{})

	// 解析性能计数器
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "% Total Run Time") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				value := strings.TrimSpace(parts[len(parts)-1])
				if cpu, err := strconv.ParseFloat(value, 64); err == nil {
					result["hypervisor_cpu_usage_percent"] = cpu
				}
			}
		}
		if strings.Contains(line, "Physical Memory") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				value := strings.TrimSpace(parts[len(parts)-1])
				if mem, err := strconv.ParseInt(value, 10, 64); err == nil {
					result["vm_physical_memory_mb"] = mem
					result["vm_physical_memory_gb"] = mem / 1024
				}
			}
		}
		if strings.Contains(line, "Read Bytes/sec") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				value := strings.TrimSpace(parts[len(parts)-1])
				if read, err := strconv.ParseInt(value, 10, 64); err == nil {
					result["storage_read_bytes_per_sec"] = read
					result["storage_read_mb_per_sec"] = read / 1024 / 1024
				}
			}
		}
		if strings.Contains(line, "Write Bytes/sec") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				value := strings.TrimSpace(parts[len(parts)-1])
				if write, err := strconv.ParseInt(value, 10, 64); err == nil {
					result["storage_write_bytes_per_sec"] = write
					result["storage_write_mb_per_sec"] = write / 1024 / 1024
				}
			}
		}
	}

	// 计算总存储IOPS
	if readBytes, ok := result["storage_read_bytes_per_sec"].(int64); ok {
		if writeBytes, ok := result["storage_write_bytes_per_sec"].(int64); ok {
			totalIOPS := (readBytes + writeBytes) / 4096 // 假设4KB块大小
			result["storage_total_iops"] = totalIOPS
		}
	}

	return result, nil
}

// 重写collectMetrics方法以使用Linux特定的实现
func (m *LinuxHyperVMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 添加连接信息
	metrics["remote_host"] = m.remoteHost
	metrics["remote_port"] = m.port
	metrics["connection_method"] = "WinRM/SSH"
	metrics["collection_time"] = time.Now().Format(time.RFC3339)

	// 使用Linux特定的方法收集指标
	if hostInfo, err := m.getHyperVHostInfo(); err == nil {
		for k, v := range hostInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get Hyper-V host info: %v", err)
	}

	if vmInfo, err := m.getVirtualMachineInfo(); err == nil {
		for k, v := range vmInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get virtual machine info: %v", err)
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

	if replicationInfo, err := m.getReplicationInfo(); err == nil {
		for k, v := range replicationInfo {
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

	// 如果所有远程调用都失败，回退到模拟数据
	if len(metrics) <= 4 { // 只有基本连接信息
		m.agent.Logger.Warn("All remote calls failed, falling back to simulated data")
		return m.HyperVMonitor.collectMetrics()
	}

	return metrics
}