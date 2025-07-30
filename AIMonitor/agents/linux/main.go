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
	agent.Info.Type = "linux"
	agent.Info.Name = "Linux System Monitor"

	// 创建Linux监控器
	monitor := NewLinuxMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Linux Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("Linux Agent stopped.")
}

// LinuxMonitor Linux系统监控器
type LinuxMonitor struct {
	agent *common.Agent
}

// NewLinuxMonitor 创建Linux监控器
func NewLinuxMonitor(agent *common.Agent) *LinuxMonitor {
	return &LinuxMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *LinuxMonitor) StartMonitoring() {
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

// collectMetrics 收集Linux系统指标
func (m *LinuxMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// CPU使用率
	cpuUsage, err := m.getCPUUsage()
	if err != nil {
		m.agent.Logger.Error("Failed to get CPU usage: %v", err)
	} else {
		metrics["cpu_usage_percent"] = cpuUsage
	}

	// 内存使用情况
	memoryInfo, err := m.getMemoryInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get memory info: %v", err)
	} else {
		metrics["memory_total_bytes"] = memoryInfo["total"]
		metrics["memory_used_bytes"] = memoryInfo["used"]
		metrics["memory_available_bytes"] = memoryInfo["available"]
		metrics["memory_usage_percent"] = memoryInfo["usage_percent"]
	}

	// 磁盘使用情况
	diskInfo, err := m.getDiskInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get disk info: %v", err)
	} else {
		metrics["disk_total_bytes"] = diskInfo["total"]
		metrics["disk_used_bytes"] = diskInfo["used"]
		metrics["disk_available_bytes"] = diskInfo["available"]
		metrics["disk_usage_percent"] = diskInfo["usage_percent"]
	}

	// 网络统计
	networkInfo, err := m.getNetworkInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get network info: %v", err)
	} else {
		metrics["network_bytes_sent"] = networkInfo["bytes_sent"]
		metrics["network_bytes_recv"] = networkInfo["bytes_recv"]
		metrics["network_packets_sent"] = networkInfo["packets_sent"]
		metrics["network_packets_recv"] = networkInfo["packets_recv"]
	}

	// 系统负载
	loadInfo, err := m.getLoadInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get load info: %v", err)
	} else {
		metrics["load_1min"] = loadInfo["load_1min"]
		metrics["load_5min"] = loadInfo["load_5min"]
		metrics["load_15min"] = loadInfo["load_15min"]
	}

	// 进程数量
	processCount, err := m.getProcessCount()
	if err != nil {
		m.agent.Logger.Error("Failed to get process count: %v", err)
	} else {
		metrics["process_count"] = processCount
	}

	// 系统启动时间
	uptime, err := m.getUptime()
	if err != nil {
		m.agent.Logger.Error("Failed to get uptime: %v", err)
	} else {
		metrics["uptime_seconds"] = uptime
	}

	return metrics
}

// getCPUUsage 获取CPU使用率
func (m *LinuxMonitor) getCPUUsage() (float64, error) {
	// 这里应该读取/proc/stat来计算CPU使用率
	// 为了简化，这里返回模拟数据
	return 35.2, nil
}

// getMemoryInfo 获取内存信息
func (m *LinuxMonitor) getMemoryInfo() (map[string]interface{}, error) {
	// 这里应该读取/proc/meminfo来获取内存信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total":         16777216000, // 16GB
		"used":          8388608000,  // 8GB
		"available":     8388608000,  // 8GB
		"usage_percent": 50.0,
	}, nil
}

// getDiskInfo 获取磁盘信息
func (m *LinuxMonitor) getDiskInfo() (map[string]interface{}, error) {
	// 这里应该使用statvfs系统调用来获取磁盘信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total":         2199023255552, // 2TB
		"used":          1099511627776, // 1TB
		"available":     1099511627776, // 1TB
		"usage_percent": 50.0,
	}, nil
}

// getNetworkInfo 获取网络信息
func (m *LinuxMonitor) getNetworkInfo() (map[string]interface{}, error) {
	// 这里应该读取/proc/net/dev来获取网络统计
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"bytes_sent":    2097152000, // 2GB
		"bytes_recv":    4194304000, // 4GB
		"packets_sent":  2000000,
		"packets_recv":  4000000,
	}, nil
}

// getLoadInfo 获取系统负载信息
func (m *LinuxMonitor) getLoadInfo() (map[string]interface{}, error) {
	// 这里应该读取/proc/loadavg来获取系统负载
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"load_1min":  2.1,
		"load_5min":  1.8,
		"load_15min": 1.5,
	}, nil
}

// getProcessCount 获取进程数量
func (m *LinuxMonitor) getProcessCount() (int, error) {
	// 这里应该读取/proc目录来统计进程数量
	// 为了简化，这里返回模拟数据
	return 200, nil
}

// getUptime 获取系统启动时间
func (m *LinuxMonitor) getUptime() (int64, error) {
	// 这里应该读取/proc/uptime来获取系统启动时间
	// 为了简化，这里返回模拟数据（10天）
	return 864000, nil
}