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
	agent.Info.Type = "windows"
	agent.Info.Name = "Windows System Monitor"

	// 创建Windows监控器
	monitor := NewWindowsMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Windows Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("Windows Agent stopped.")
}

// WindowsMonitor Windows系统监控器
type WindowsMonitor struct {
	agent *common.Agent
}

// NewWindowsMonitor 创建Windows监控器
func NewWindowsMonitor(agent *common.Agent) *WindowsMonitor {
	return &WindowsMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *WindowsMonitor) StartMonitoring() {
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

// collectMetrics 收集Windows系统指标
func (m *WindowsMonitor) collectMetrics() map[string]interface{} {
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
func (m *WindowsMonitor) getCPUUsage() (float64, error) {
	// 这里应该使用Windows API或WMI来获取CPU使用率
	// 为了简化，这里返回模拟数据
	return 25.5, nil
}

// getMemoryInfo 获取内存信息
func (m *WindowsMonitor) getMemoryInfo() (map[string]interface{}, error) {
	// 这里应该使用Windows API来获取内存信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total":         8589934592, // 8GB
		"used":          4294967296, // 4GB
		"available":     4294967296, // 4GB
		"usage_percent": 50.0,
	}, nil
}

// getDiskInfo 获取磁盘信息
func (m *WindowsMonitor) getDiskInfo() (map[string]interface{}, error) {
	// 这里应该使用Windows API来获取磁盘信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total":         1099511627776, // 1TB
		"used":          549755813888,  // 512GB
		"available":     549755813888,  // 512GB
		"usage_percent": 50.0,
	}, nil
}

// getNetworkInfo 获取网络信息
func (m *WindowsMonitor) getNetworkInfo() (map[string]interface{}, error) {
	// 这里应该使用Windows API来获取网络统计
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"bytes_sent":    1048576000, // 1GB
		"bytes_recv":    2097152000, // 2GB
		"packets_sent":  1000000,
		"packets_recv":  2000000,
	}, nil
}

// getLoadInfo 获取系统负载信息
func (m *WindowsMonitor) getLoadInfo() (map[string]interface{}, error) {
	// Windows没有传统的load average概念，这里可以用CPU队列长度等指标代替
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"load_1min":  1.5,
		"load_5min":  1.2,
		"load_15min": 1.0,
	}, nil
}

// getProcessCount 获取进程数量
func (m *WindowsMonitor) getProcessCount() (int, error) {
	// 这里应该使用Windows API来获取进程数量
	// 为了简化，这里返回模拟数据
	return 150, nil
}

// getUptime 获取系统启动时间
func (m *WindowsMonitor) getUptime() (int64, error) {
	// 这里应该使用Windows API来获取系统启动时间
	// 为了简化，这里返回模拟数据（7天）
	return 604800, nil
}