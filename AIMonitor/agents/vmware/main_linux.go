//go:build linux
// +build linux

package main

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"aimonitor-agents/common"
)

// LinuxVMwareMonitor Linux版本的VMware监控器
type LinuxVMwareMonitor struct {
	*VMwareMonitor
	client *govmomi.Client
	finder *find.Finder
	perfManager *performance.Manager
	vcenterURL string
	username string
	password string
	insecure bool
}

// NewLinuxVMwareMonitor 创建Linux版本的VMware监控器
func NewLinuxVMwareMonitor(agent *common.Agent) *LinuxVMwareMonitor {
	baseMonitor := NewVMwareMonitor(agent)
	return &LinuxVMwareMonitor{
		VMwareMonitor: baseMonitor,
		vcenterURL: "https://vcenter.example.com/sdk",
		username: "administrator@vsphere.local",
		password: "password",
		insecure: true,
	}
}

// initClient 初始化VMware客户端
func (m *LinuxVMwareMonitor) initClient() error {
	if m.client != nil {
		return nil
	}

	ctx := context.Background()

	// 解析vCenter URL
	u, err := soap.ParseURL(m.vcenterURL)
	if err != nil {
		return fmt.Errorf("failed to parse vCenter URL: %v", err)
	}

	// 设置认证信息
	u.User = url.UserPassword(m.username, m.password)

	// 创建客户端
	client, err := govmomi.NewClient(ctx, u, m.insecure)
	if err != nil {
		return fmt.Errorf("failed to create vCenter client: %v", err)
	}

	m.client = client
	m.finder = find.NewFinder(client.Client, true)
	m.perfManager = performance.NewManager(client.Client)

	return nil
}

// getVCenterInfo 获取真实的vCenter服务器信息
func (m *LinuxVMwareMonitor) getVCenterInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	result := make(map[string]interface{})

	// 获取服务实例
	si := object.NewServiceInstance(m.client.Client)
	content, err := si.RetrieveContent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve service content: %v", err)
	}

	// 获取关于信息
	result["name"] = content.About.Name
	result["version"] = content.About.Version
	result["build"] = content.About.Build
	result["full_name"] = content.About.FullName
	result["vendor"] = content.About.Vendor
	result["api_type"] = content.About.ApiType
	result["api_version"] = content.About.ApiVersion
	result["instance_uuid"] = content.About.InstanceUuid
	result["license_product_name"] = content.About.LicenseProductName
	result["license_product_version"] = content.About.LicenseProductVersion

	// 获取会话信息
	sessionManager := object.NewSessionManager(m.client.Client)
	currentSession, err := sessionManager.UserSession(ctx)
	if err == nil && currentSession != nil {
		result["session_user"] = currentSession.UserName
		result["session_login_time"] = currentSession.LoginTime.Format(time.RFC3339)
		result["session_last_active"] = currentSession.LastActiveTime.Format(time.RFC3339)
	}

	return result, nil
}

// getDatacenterInfo 获取真实的数据中心信息
func (m *LinuxVMwareMonitor) getDatacenterInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	result := make(map[string]interface{})

	// 查找所有数据中心
	datacenters, err := m.finder.DatacenterList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list datacenters: %v", err)
	}

	result["total_datacenters"] = len(datacenters)

	if len(datacenters) > 0 {
		// 获取第一个数据中心的详细信息
		dc := datacenters[0]
		result["primary_datacenter_name"] = dc.Name()
		result["primary_datacenter_path"] = dc.InventoryPath

		// 设置数据中心为默认搜索范围
		m.finder.SetDatacenter(dc)

		// 获取数据中心下的文件夹信息
		var dcMo mo.Datacenter
		err = dc.Properties(ctx, dc.Reference(), []string{"vmFolder", "hostFolder", "datastoreFolder", "networkFolder"}, &dcMo)
		if err == nil {
			result["vm_folder"] = dcMo.VmFolder.Value
			result["host_folder"] = dcMo.HostFolder.Value
			result["datastore_folder"] = dcMo.DatastoreFolder.Value
			result["network_folder"] = dcMo.NetworkFolder.Value
		}
	}

	return result, nil
}

// getClusterInfo 获取真实的集群信息
func (m *LinuxVMwareMonitor) getClusterInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	result := make(map[string]interface{})

	// 查找所有集群
	clusters, err := m.finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %v", err)
	}

	result["total_clusters"] = len(clusters)

	if len(clusters) > 0 {
		// 获取集群详细信息
		totalCPUCores := int32(0)
		totalMemoryMB := int64(0)
		totalHosts := 0
		totalVMs := 0

		for i, cluster := range clusters {
			var clusterMo mo.ClusterComputeResource
			err = cluster.Properties(ctx, cluster.Reference(), []string{
				"name", "summary", "host", "configuration",
			}, &clusterMo)
			if err != nil {
				continue
			}

			if i == 0 {
				// 第一个集群的详细信息
				result["primary_cluster_name"] = clusterMo.Name
				result["primary_cluster_total_cpu_mhz"] = clusterMo.Summary.GetComputeResourceSummary().TotalCpu
				result["primary_cluster_total_memory_mb"] = clusterMo.Summary.GetComputeResourceSummary().TotalMemory / 1024 / 1024
				result["primary_cluster_num_cpu_cores"] = clusterMo.Summary.GetComputeResourceSummary().NumCpuCores
				result["primary_cluster_num_hosts"] = clusterMo.Summary.GetComputeResourceSummary().NumHosts
				result["primary_cluster_effective_cpu"] = clusterMo.Summary.GetComputeResourceSummary().EffectiveCpu
				result["primary_cluster_effective_memory"] = clusterMo.Summary.GetComputeResourceSummary().EffectiveMemory / 1024 / 1024

				// DRS配置
				if clusterMo.Configuration != nil {
					result["primary_cluster_drs_enabled"] = clusterMo.Configuration.DrsConfig.Enabled
					result["primary_cluster_drs_behavior"] = string(clusterMo.Configuration.DrsConfig.DefaultVmBehavior)
					result["primary_cluster_ha_enabled"] = clusterMo.Configuration.DasConfig.Enabled
				}
			}

			// 累计统计
			totalCPUCores += clusterMo.Summary.GetComputeResourceSummary().NumCpuCores
			totalMemoryMB += clusterMo.Summary.GetComputeResourceSummary().TotalMemory / 1024 / 1024
			totalHosts += int(clusterMo.Summary.GetComputeResourceSummary().NumHosts)
			totalVMs += int(clusterMo.Summary.GetComputeResourceSummary().NumVmsTotal)
		}

		result["total_cpu_cores"] = totalCPUCores
		result["total_memory_mb"] = totalMemoryMB
		result["total_memory_gb"] = totalMemoryMB / 1024
		result["total_hosts"] = totalHosts
		result["total_vms"] = totalVMs
	}

	return result, nil
}

// getESXiHostInfo 获取真实的ESXi主机信息
func (m *LinuxVMwareMonitor) getESXiHostInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	result := make(map[string]interface{})

	// 查找所有主机
	hosts, err := m.finder.HostSystemList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts: %v", err)
	}

	result["total_hosts"] = len(hosts)

	if len(hosts) > 0 {
		// 统计主机状态
		connectedHosts := 0
		disconnectedHosts := 0
		maintenanceHosts := 0
		totalCPUMhz := int32(0)
		totalMemoryMB := int64(0)
		totalVMs := 0

		for i, host := range hosts {
			var hostMo mo.HostSystem
			err = host.Properties(ctx, host.Reference(), []string{
				"name", "summary", "runtime", "hardware", "vm",
			}, &hostMo)
			if err != nil {
				continue
			}

			if i == 0 {
				// 第一个主机的详细信息
				result["primary_host_name"] = hostMo.Name
				result["primary_host_version"] = hostMo.Summary.Config.Product.Version
				result["primary_host_build"] = hostMo.Summary.Config.Product.Build
				result["primary_host_vendor"] = hostMo.Summary.Hardware.Vendor
				result["primary_host_model"] = hostMo.Summary.Hardware.Model
				result["primary_host_cpu_model"] = hostMo.Summary.Hardware.CpuModel
				result["primary_host_num_cpu_cores"] = hostMo.Summary.Hardware.NumCpuCores
				result["primary_host_num_cpu_threads"] = hostMo.Summary.Hardware.NumCpuThreads
				result["primary_host_cpu_mhz"] = hostMo.Summary.Hardware.CpuMhz
				result["primary_host_memory_mb"] = hostMo.Summary.Hardware.MemorySize / 1024 / 1024
				result["primary_host_memory_gb"] = hostMo.Summary.Hardware.MemorySize / 1024 / 1024 / 1024
				result["primary_host_num_nics"] = hostMo.Summary.Hardware.NumNics
				result["primary_host_num_hbas"] = hostMo.Summary.Hardware.NumHBAs
				result["primary_host_connection_state"] = string(hostMo.Runtime.ConnectionState)
				result["primary_host_power_state"] = string(hostMo.Runtime.PowerState)
				result["primary_host_maintenance_mode"] = hostMo.Runtime.InMaintenanceMode
				result["primary_host_uptime_seconds"] = hostMo.Summary.QuickStats.Uptime
				result["primary_host_cpu_usage_mhz"] = hostMo.Summary.QuickStats.OverallCpuUsage
				result["primary_host_memory_usage_mb"] = hostMo.Summary.QuickStats.OverallMemoryUsage
			}

			// 统计主机状态
			switch hostMo.Runtime.ConnectionState {
			case types.HostSystemConnectionStateConnected:
				connectedHosts++
			case types.HostSystemConnectionStateDisconnected:
				disconnectedHosts++
			}

			if hostMo.Runtime.InMaintenanceMode {
				maintenanceHosts++
			}

			// 累计资源
			totalCPUMhz += hostMo.Summary.Hardware.CpuMhz * hostMo.Summary.Hardware.NumCpuCores
			totalMemoryMB += hostMo.Summary.Hardware.MemorySize / 1024 / 1024
			totalVMs += len(hostMo.Vm)
		}

		result["connected_hosts"] = connectedHosts
		result["disconnected_hosts"] = disconnectedHosts
		result["maintenance_hosts"] = maintenanceHosts
		result["total_cpu_mhz"] = totalCPUMhz
		result["total_memory_mb"] = totalMemoryMB
		result["total_memory_gb"] = totalMemoryMB / 1024
		result["total_vms_on_hosts"] = totalVMs
	}

	return result, nil
}

// getVirtualMachineInfo 获取真实的虚拟机信息
func (m *LinuxVMwareMonitor) getVirtualMachineInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	result := make(map[string]interface{})

	// 查找所有虚拟机
	vms, err := m.finder.VirtualMachineList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list virtual machines: %v", err)
	}

	result["total_vms"] = len(vms)

	if len(vms) > 0 {
		// 统计虚拟机状态
		poweredOnVMs := 0
		poweredOffVMs := 0
		suspendedVMs := 0
		totalCPUs := int32(0)
		totalMemoryMB := int64(0)
		totalDiskGB := int64(0)

		for i, vm := range vms {
			var vmMo mo.VirtualMachine
			err = vm.Properties(ctx, vm.Reference(), []string{
				"name", "summary", "runtime", "config", "guest",
			}, &vmMo)
			if err != nil {
				continue
			}

			if i == 0 {
				// 第一个虚拟机的详细信息
				result["primary_vm_name"] = vmMo.Name
				result["primary_vm_power_state"] = string(vmMo.Runtime.PowerState)
				result["primary_vm_connection_state"] = string(vmMo.Runtime.ConnectionState)
				result["primary_vm_guest_os"] = vmMo.Summary.Config.GuestFullName
				result["primary_vm_num_cpu"] = vmMo.Summary.Config.NumCpu
				result["primary_vm_memory_mb"] = vmMo.Summary.Config.MemorySizeMB
				result["primary_vm_num_disks"] = vmMo.Summary.Config.NumVirtualDisks
				result["primary_vm_num_nics"] = vmMo.Summary.Config.NumEthernetCards
				result["primary_vm_committed_storage_mb"] = vmMo.Summary.Storage.Committed / 1024 / 1024
				result["primary_vm_uncommitted_storage_mb"] = vmMo.Summary.Storage.Uncommitted / 1024 / 1024
				result["primary_vm_cpu_usage_mhz"] = vmMo.Summary.QuickStats.OverallCpuUsage
				result["primary_vm_memory_usage_mb"] = vmMo.Summary.QuickStats.GuestMemoryUsage
				result["primary_vm_host_memory_usage_mb"] = vmMo.Summary.QuickStats.HostMemoryUsage
				result["primary_vm_uptime_seconds"] = vmMo.Summary.QuickStats.UptimeSeconds

				// VMware Tools信息
				if vmMo.Guest != nil {
					result["primary_vm_tools_status"] = string(vmMo.Guest.ToolsStatus)
					result["primary_vm_tools_version"] = vmMo.Guest.ToolsVersion
					result["primary_vm_guest_hostname"] = vmMo.Guest.HostName
					result["primary_vm_guest_ip"] = vmMo.Guest.IpAddress
				}
			}

			// 统计虚拟机状态
			switch vmMo.Runtime.PowerState {
			case types.VirtualMachinePowerStatePoweredOn:
				poweredOnVMs++
			case types.VirtualMachinePowerStatePoweredOff:
				poweredOffVMs++
			case types.VirtualMachinePowerStateSuspended:
				suspendedVMs++
			}

			// 累计资源
			totalCPUs += vmMo.Summary.Config.NumCpu
			totalMemoryMB += int64(vmMo.Summary.Config.MemorySizeMB)
			totalDiskGB += (vmMo.Summary.Storage.Committed + vmMo.Summary.Storage.Uncommitted) / 1024 / 1024 / 1024
		}

		result["powered_on_vms"] = poweredOnVMs
		result["powered_off_vms"] = poweredOffVMs
		result["suspended_vms"] = suspendedVMs
		result["total_vm_cpus"] = totalCPUs
		result["total_vm_memory_mb"] = totalMemoryMB
		result["total_vm_memory_gb"] = totalMemoryMB / 1024
		result["total_vm_disk_gb"] = totalDiskGB
	}

	return result, nil
}

// getDatastoreInfo 获取真实的存储信息
func (m *LinuxVMwareMonitor) getDatastoreInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	result := make(map[string]interface{})

	// 查找所有数据存储
	datastores, err := m.finder.DatastoreList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list datastores: %v", err)
	}

	result["total_datastores"] = len(datastores)

	if len(datastores) > 0 {
		totalCapacityGB := int64(0)
		totalFreeSpaceGB := int64(0)
		totalUsedSpaceGB := int64(0)

		for i, ds := range datastores {
			var dsMo mo.Datastore
			err = ds.Properties(ctx, ds.Reference(), []string{
				"name", "summary", "info", "host", "vm",
			}, &dsMo)
			if err != nil {
				continue
			}

			if i == 0 {
				// 第一个数据存储的详细信息
				result["primary_datastore_name"] = dsMo.Name
				result["primary_datastore_type"] = dsMo.Summary.Type
				result["primary_datastore_capacity_gb"] = dsMo.Summary.Capacity / 1024 / 1024 / 1024
				result["primary_datastore_free_space_gb"] = dsMo.Summary.FreeSpace / 1024 / 1024 / 1024
				result["primary_datastore_used_space_gb"] = (dsMo.Summary.Capacity - dsMo.Summary.FreeSpace) / 1024 / 1024 / 1024
				result["primary_datastore_accessible"] = dsMo.Summary.Accessible
				result["primary_datastore_maintenance_mode"] = string(dsMo.Summary.MaintenanceMode)
				result["primary_datastore_multiple_host_access"] = dsMo.Summary.MultipleHostAccess
				result["primary_datastore_num_hosts"] = len(dsMo.Host)
				result["primary_datastore_num_vms"] = len(dsMo.Vm)

				// 计算使用率
				if dsMo.Summary.Capacity > 0 {
					usagePercent := float64(dsMo.Summary.Capacity-dsMo.Summary.FreeSpace) / float64(dsMo.Summary.Capacity) * 100
					result["primary_datastore_usage_percent"] = usagePercent
				}
			}

			// 累计存储统计
			totalCapacityGB += dsMo.Summary.Capacity / 1024 / 1024 / 1024
			totalFreeSpaceGB += dsMo.Summary.FreeSpace / 1024 / 1024 / 1024
			totalUsedSpaceGB += (dsMo.Summary.Capacity - dsMo.Summary.FreeSpace) / 1024 / 1024 / 1024
		}

		result["total_capacity_gb"] = totalCapacityGB
		result["total_free_space_gb"] = totalFreeSpaceGB
		result["total_used_space_gb"] = totalUsedSpaceGB

		// 计算总体使用率
		if totalCapacityGB > 0 {
			result["total_usage_percent"] = float64(totalUsedSpaceGB) / float64(totalCapacityGB) * 100
		}
	}

	return result, nil
}

// getNetworkInfo 获取真实的网络信息
func (m *LinuxVMwareMonitor) getNetworkInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	result := make(map[string]interface{})

	// 查找所有网络
	networks, err := m.finder.NetworkList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %v", err)
	}

	result["total_networks"] = len(networks)

	// 查找分布式虚拟交换机
	dvSwitches, err := m.finder.DistributedVirtualSwitchList(ctx, "*")
	if err == nil {
		result["total_distributed_switches"] = len(dvSwitches)

		if len(dvSwitches) > 0 {
			// 获取第一个分布式交换机的信息
			dvs := dvSwitches[0]
			var dvsMo mo.DistributedVirtualSwitch
			err = dvs.Properties(ctx, dvs.Reference(), []string{
				"name", "summary", "config",
			}, &dvsMo)
			if err == nil {
				result["primary_dvs_name"] = dvsMo.Name
				result["primary_dvs_num_ports"] = dvsMo.Summary.NumPorts
				result["primary_dvs_num_hosts"] = len(dvsMo.Summary.HostMember)
				result["primary_dvs_version"] = dvsMo.Summary.ProductInfo.Version
			}
		}
	} else {
		result["total_distributed_switches"] = 0
	}

	// 统计标准交换机（通过主机获取）
	hosts, err := m.finder.HostSystemList(ctx, "*")
	if err == nil && len(hosts) > 0 {
		totalStandardSwitches := 0
		totalPortGroups := 0

		for _, host := range hosts {
			var hostMo mo.HostSystem
			err = host.Properties(ctx, host.Reference(), []string{
				"config.network",
			}, &hostMo)
			if err != nil {
				continue
			}

			if hostMo.Config != nil && hostMo.Config.Network != nil {
				totalStandardSwitches += len(hostMo.Config.Network.Vswitch)
				totalPortGroups += len(hostMo.Config.Network.Portgroup)
			}
		}

		result["total_standard_switches"] = totalStandardSwitches
		result["total_port_groups"] = totalPortGroups
	}

	return result, nil
}

// getPerformanceStats 获取真实的性能统计
func (m *LinuxVMwareMonitor) getPerformanceStats() (map[string]interface{}, error) {
	ctx := context.Background()
	result := make(map[string]interface{})

	// 获取性能计数器信息
	counters, err := m.perfManager.CounterInfoByName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance counters: %v", err)
	}

	result["available_counters"] = len(counters)

	// 获取一些关键的性能指标
	keyCounters := []string{
		"cpu.usage.average",
		"mem.usage.average",
		"disk.usage.average",
		"net.usage.average",
	}

	availableKeyCounters := 0
	for _, counterName := range keyCounters {
		if _, exists := counters[counterName]; exists {
			availableKeyCounters++
		}
	}

	result["available_key_counters"] = availableKeyCounters
	result["key_counters_list"] = keyCounters

	// 获取性能提供者信息
	providers, err := m.perfManager.ProviderSummary(ctx)
	if err == nil {
		result["performance_providers"] = len(providers)
	}

	return result, nil
}

// 重写collectMetrics方法以使用Linux特定的实现
func (m *LinuxVMwareMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 初始化客户端
	if err := m.initClient(); err != nil {
		m.agent.Logger.Error("Failed to initialize VMware client: %v", err)
		return m.VMwareMonitor.collectMetrics() // 回退到模拟数据
	}

	// 使用Linux特定的方法收集指标
	if vcenterInfo, err := m.getVCenterInfo(); err == nil {
		for k, v := range vcenterInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get vCenter info: %v", err)
	}

	if datacenterInfo, err := m.getDatacenterInfo(); err == nil {
		for k, v := range datacenterInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get datacenter info: %v", err)
	}

	if clusterInfo, err := m.getClusterInfo(); err == nil {
		for k, v := range clusterInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get cluster info: %v", err)
	}

	if hostInfo, err := m.getESXiHostInfo(); err == nil {
		for k, v := range hostInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get ESXi host info: %v", err)
	}

	if vmInfo, err := m.getVirtualMachineInfo(); err == nil {
		for k, v := range vmInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get virtual machine info: %v", err)
	}

	if datastoreInfo, err := m.getDatastoreInfo(); err == nil {
		for k, v := range datastoreInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get datastore info: %v", err)
	}

	if networkInfo, err := m.getNetworkInfo(); err == nil {
		for k, v := range networkInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get network info: %v", err)
	}

	if perfStats, err := m.getPerformanceStats(); err == nil {
		for k, v := range perfStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get performance stats: %v", err)
	}

	return metrics
}