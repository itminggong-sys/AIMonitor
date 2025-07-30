package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aimonitor-agents/common"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// 创建Agent
	agent, err := common.NewAgent(*configPath)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// 设置Agent类型
	agent.Info.Type = "hyperv"
	agent.Info.Name = "Hyper-V Monitor"

	// 创建Hyper-V监控器
	monitor := NewHyperVMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Hyper-V Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("Hyper-V Agent stopped.")
}

// HyperVMonitor Hyper-V监控器
type HyperVMonitor struct {
	agent *common.Agent
}

// NewHyperVMonitor 创建Hyper-V监控器
func NewHyperVMonitor(agent *common.Agent) *HyperVMonitor {
	return &HyperVMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *HyperVMonitor) StartMonitoring() {
	ticker := time.NewTicker(m.agent.Config.Metrics.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.agent.Ctx.Done():
			return
		case <-ticker.C:
			if m.agent.Config.Metrics.Enabled {
				metrics := m.collectMetrics()
				if err := m.agent.SendMetrics(metrics); err != nil {
					m.agent.Logger.Error("Failed to send metrics: %v", err)
				}
			}
		}
	}
}

// collectMetrics 收集Hyper-V指标
func (m *HyperVMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Hyper-V主机信息
	hostInfo, err := m.getHostInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Hyper-V host info: %v", err)
	} else {
		metrics["host_version"] = hostInfo["host_version"]
		metrics["host_edition"] = hostInfo["host_edition"]
		metrics["host_uptime"] = hostInfo["host_uptime"]
		metrics["host_status"] = hostInfo["host_status"]
		metrics["host_cpu_cores"] = hostInfo["host_cpu_cores"]
		metrics["host_logical_processors"] = hostInfo["host_logical_processors"]
		metrics["host_total_memory_mb"] = hostInfo["host_total_memory_mb"]
		metrics["host_available_memory_mb"] = hostInfo["host_available_memory_mb"]
		metrics["host_memory_usage_percent"] = hostInfo["host_memory_usage_percent"]
	}

	// 虚拟机信息
	vmInfo, err := m.getVMInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get VM info: %v", err)
	} else {
		metrics["vm_count"] = vmInfo["vm_count"]
		metrics["vm_running_count"] = vmInfo["vm_running_count"]
		metrics["vm_stopped_count"] = vmInfo["vm_stopped_count"]
		metrics["vm_paused_count"] = vmInfo["vm_paused_count"]
		metrics["vm_saved_count"] = vmInfo["vm_saved_count"]
		metrics["vm_starting_count"] = vmInfo["vm_starting_count"]
		metrics["vm_stopping_count"] = vmInfo["vm_stopping_count"]
		metrics["vm_total_cpu_cores"] = vmInfo["vm_total_cpu_cores"]
		metrics["vm_total_memory_mb"] = vmInfo["vm_total_memory_mb"]
		metrics["vm_cpu_usage_percent"] = vmInfo["vm_cpu_usage_percent"]
		metrics["vm_memory_usage_percent"] = vmInfo["vm_memory_usage_percent"]
		metrics["vm_integration_services_ok_count"] = vmInfo["vm_integration_services_ok_count"]
		metrics["vm_integration_services_error_count"] = vmInfo["vm_integration_services_error_count"]
	}

	// 虚拟交换机信息
	vSwitchInfo, err := m.getVSwitchInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get virtual switch info: %v", err)
	} else {
		metrics["vswitch_count"] = vSwitchInfo["vswitch_count"]
		metrics["vswitch_external_count"] = vSwitchInfo["vswitch_external_count"]
		metrics["vswitch_internal_count"] = vSwitchInfo["vswitch_internal_count"]
		metrics["vswitch_private_count"] = vSwitchInfo["vswitch_private_count"]
		metrics["vswitch_total_ports"] = vSwitchInfo["vswitch_total_ports"]
		metrics["vswitch_used_ports"] = vSwitchInfo["vswitch_used_ports"]
	}

	// 存储信息
	storageInfo, err := m.getStorageInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get storage info: %v", err)
	} else {
		metrics["vhd_count"] = storageInfo["vhd_count"]
		metrics["vhdx_count"] = storageInfo["vhdx_count"]
		metrics["vhd_total_size_gb"] = storageInfo["vhd_total_size_gb"]
		metrics["vhd_used_size_gb"] = storageInfo["vhd_used_size_gb"]
		metrics["checkpoint_count"] = storageInfo["checkpoint_count"]
		metrics["checkpoint_total_size_gb"] = storageInfo["checkpoint_total_size_gb"]
	}

	// 复制信息
	replicationInfo, err := m.getReplicationInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get replication info: %v", err)
	} else {
		metrics["replication_enabled_vm_count"] = replicationInfo["replication_enabled_vm_count"]
		metrics["replication_healthy_count"] = replicationInfo["replication_healthy_count"]
		metrics["replication_warning_count"] = replicationInfo["replication_warning_count"]
		metrics["replication_critical_count"] = replicationInfo["replication_critical_count"]
		metrics["replication_primary_count"] = replicationInfo["replication_primary_count"]
		metrics["replication_replica_count"] = replicationInfo["replication_replica_count"]
	}

	// 集群信息（如果启用了故障转移集群）
	clusterInfo, err := m.getClusterInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get cluster info: %v", err)
	} else {
		metrics["cluster_enabled"] = clusterInfo["cluster_enabled"]
		metrics["cluster_node_count"] = clusterInfo["cluster_node_count"]
		metrics["cluster_online_node_count"] = clusterInfo["cluster_online_node_count"]
		metrics["cluster_offline_node_count"] = clusterInfo["cluster_offline_node_count"]
		metrics["cluster_shared_volume_count"] = clusterInfo["cluster_shared_volume_count"]
		metrics["cluster_shared_volume_online_count"] = clusterInfo["cluster_shared_volume_online_count"]
	}

	// 性能计数器
	perfCounters, err := m.getPerformanceCounters()
	if err != nil {
		m.agent.Logger.Error("Failed to get performance counters: %v", err)
	} else {
		metrics["hyperv_logical_processor_total_runtime_percent"] = perfCounters["hyperv_logical_processor_total_runtime_percent"]
		metrics["hyperv_logical_processor_guest_runtime_percent"] = perfCounters["hyperv_logical_processor_guest_runtime_percent"]
		metrics["hyperv_logical_processor_hypervisor_runtime_percent"] = perfCounters["hyperv_logical_processor_hypervisor_runtime_percent"]
		metrics["hyperv_vm_vid_physical_pages_allocated"] = perfCounters["hyperv_vm_vid_physical_pages_allocated"]
		metrics["hyperv_vm_vid_remote_physical_pages"] = perfCounters["hyperv_vm_vid_remote_physical_pages"]
		metrics["hyperv_virtual_network_adapter_bytes_per_sec"] = perfCounters["hyperv_virtual_network_adapter_bytes_per_sec"]
		metrics["hyperv_virtual_storage_device_read_bytes_per_sec"] = perfCounters["hyperv_virtual_storage_device_read_bytes_per_sec"]
		metrics["hyperv_virtual_storage_device_write_bytes_per_sec"] = perfCounters["hyperv_virtual_storage_device_write_bytes_per_sec"]
	}

	return metrics
}

// getHostInfo 获取Hyper-V主机信息
func (m *HyperVMonitor) getHostInfo() (map[string]interface{}, error) {
	// 这里应该调用Windows PowerShell或WMI API
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"host_version":              "10.0.20348",
		"host_edition":              "Windows Server 2022 Datacenter",
		"host_uptime":               86400, // 秒
		"host_status":               "Running",
		"host_cpu_cores":            16,
		"host_logical_processors":   32,
		"host_total_memory_mb":      65536, // 64GB
		"host_available_memory_mb":  32768, // 32GB
		"host_memory_usage_percent": 50.0,
	}, nil
}

// getVMInfo 获取虚拟机信息
func (m *HyperVMonitor) getVMInfo() (map[string]interface{}, error) {
	// 这里应该调用Hyper-V PowerShell模块或WMI API
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"vm_count":                              25,
		"vm_running_count":                      20,
		"vm_stopped_count":                      3,
		"vm_paused_count":                       1,
		"vm_saved_count":                        1,
		"vm_starting_count":                     0,
		"vm_stopping_count":                     0,
		"vm_total_cpu_cores":                    80,
		"vm_total_memory_mb":                    204800, // 200GB
		"vm_cpu_usage_percent":                  45.0,
		"vm_memory_usage_percent":               60.0,
		"vm_integration_services_ok_count":      22,
		"vm_integration_services_error_count":   3,
	}, nil
}

// getVSwitchInfo 获取虚拟交换机信息
func (m *HyperVMonitor) getVSwitchInfo() (map[string]interface{}, error) {
	// 这里应该调用Hyper-V PowerShell模块
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"vswitch_count":          5,
		"vswitch_external_count": 2,
		"vswitch_internal_count": 2,
		"vswitch_private_count":  1,
		"vswitch_total_ports":    256,
		"vswitch_used_ports":     45,
	}, nil
}

// getStorageInfo 获取存储信息
func (m *HyperVMonitor) getStorageInfo() (map[string]interface{}, error) {
	// 这里应该调用Hyper-V PowerShell模块
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"vhd_count":                  15,
		"vhdx_count":                 35,
		"vhd_total_size_gb":          5120, // 5TB
		"vhd_used_size_gb":           3072, // 3TB
		"checkpoint_count":           8,
		"checkpoint_total_size_gb":   512, // 512GB
	}, nil
}

// getReplicationInfo 获取复制信息
func (m *HyperVMonitor) getReplicationInfo() (map[string]interface{}, error) {
	// 这里应该调用Hyper-V PowerShell模块
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"replication_enabled_vm_count": 10,
		"replication_healthy_count":    8,
		"replication_warning_count":    1,
		"replication_critical_count":   1,
		"replication_primary_count":    6,
		"replication_replica_count":    4,
	}, nil
}

// getClusterInfo 获取集群信息
func (m *HyperVMonitor) getClusterInfo() (map[string]interface{}, error) {
	// 这里应该调用故障转移集群PowerShell模块
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"cluster_enabled":                      true,
		"cluster_node_count":                   3,
		"cluster_online_node_count":            3,
		"cluster_offline_node_count":           0,
		"cluster_shared_volume_count":          4,
		"cluster_shared_volume_online_count":   4,
	}, nil
}

// getPerformanceCounters 获取性能计数器
func (m *HyperVMonitor) getPerformanceCounters() (map[string]interface{}, error) {
	// 这里应该调用Windows性能计数器API
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"hyperv_logical_processor_total_runtime_percent":       75.0,
		"hyperv_logical_processor_guest_runtime_percent":       60.0,
		"hyperv_logical_processor_hypervisor_runtime_percent":  15.0,
		"hyperv_vm_vid_physical_pages_allocated":               1048576, // 4GB in pages
		"hyperv_vm_vid_remote_physical_pages":                  262144,  // 1GB in pages
		"hyperv_virtual_network_adapter_bytes_per_sec":         104857600, // 100MB/s
		"hyperv_virtual_storage_device_read_bytes_per_sec":     52428800,  // 50MB/s
		"hyperv_virtual_storage_device_write_bytes_per_sec":    31457280,  // 30MB/s
	}, nil
}