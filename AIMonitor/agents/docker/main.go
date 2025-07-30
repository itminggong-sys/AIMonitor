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
	agent.Info.Type = "docker"
	agent.Info.Name = "Docker Monitor"

	// 创建Docker监控器
	monitor := NewDockerMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Docker Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("Docker Agent stopped.")
}

// DockerMonitor Docker监控器
type DockerMonitor struct {
	agent *common.Agent
}

// NewDockerMonitor 创建Docker监控器
func NewDockerMonitor(agent *common.Agent) *DockerMonitor {
	return &DockerMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *DockerMonitor) StartMonitoring() {
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

// collectMetrics 收集Docker指标
func (m *DockerMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Docker系统信息
	systemInfo, err := m.getSystemInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Docker system info: %v", err)
	} else {
		metrics["containers_total"] = systemInfo["containers_total"]
		metrics["containers_running"] = systemInfo["containers_running"]
		metrics["containers_paused"] = systemInfo["containers_paused"]
		metrics["containers_stopped"] = systemInfo["containers_stopped"]
		metrics["images_total"] = systemInfo["images_total"]
	}

	// 容器资源使用情况
	containerStats, err := m.getContainerStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get container stats: %v", err)
	} else {
		metrics["total_cpu_usage_percent"] = containerStats["total_cpu_usage_percent"]
		metrics["total_memory_usage_bytes"] = containerStats["total_memory_usage_bytes"]
		metrics["total_memory_limit_bytes"] = containerStats["total_memory_limit_bytes"]
		metrics["total_network_rx_bytes"] = containerStats["total_network_rx_bytes"]
		metrics["total_network_tx_bytes"] = containerStats["total_network_tx_bytes"]
		metrics["total_block_read_bytes"] = containerStats["total_block_read_bytes"]
		metrics["total_block_write_bytes"] = containerStats["total_block_write_bytes"]
	}

	// 镜像信息
	imageInfo, err := m.getImageInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get image info: %v", err)
	} else {
		metrics["total_image_size_bytes"] = imageInfo["total_image_size_bytes"]
		metrics["dangling_images"] = imageInfo["dangling_images"]
	}

	// 卷信息
	volumeInfo, err := m.getVolumeInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get volume info: %v", err)
	} else {
		metrics["volumes_total"] = volumeInfo["volumes_total"]
		metrics["volumes_size_bytes"] = volumeInfo["volumes_size_bytes"]
	}

	// 网络信息
	networkInfo, err := m.getNetworkInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get network info: %v", err)
	} else {
		metrics["networks_total"] = networkInfo["networks_total"]
		metrics["bridge_networks"] = networkInfo["bridge_networks"]
		metrics["overlay_networks"] = networkInfo["overlay_networks"]
	}

	// Docker守护进程信息
	daemonInfo, err := m.getDaemonInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get daemon info: %v", err)
	} else {
		metrics["docker_version"] = daemonInfo["docker_version"]
		metrics["api_version"] = daemonInfo["api_version"]
		metrics["kernel_version"] = daemonInfo["kernel_version"]
		metrics["operating_system"] = daemonInfo["operating_system"]
		metrics["architecture"] = daemonInfo["architecture"]
	}

	return metrics
}

// getSystemInfo 获取Docker系统信息
func (m *DockerMonitor) getSystemInfo() (map[string]interface{}, error) {
	// 这里应该使用Docker API获取系统信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"containers_total":   15,
		"containers_running": 10,
		"containers_paused":  1,
		"containers_stopped": 4,
		"images_total":       25,
	}, nil
}

// getContainerStats 获取容器统计信息
func (m *DockerMonitor) getContainerStats() (map[string]interface{}, error) {
	// 这里应该使用Docker API获取所有运行容器的统计信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total_cpu_usage_percent":    45.5,
		"total_memory_usage_bytes":   2147483648, // 2GB
		"total_memory_limit_bytes":   8589934592, // 8GB
		"total_network_rx_bytes":     1073741824, // 1GB
		"total_network_tx_bytes":     536870912,  // 512MB
		"total_block_read_bytes":     2147483648, // 2GB
		"total_block_write_bytes":    1073741824, // 1GB
	}, nil
}

// getImageInfo 获取镜像信息
func (m *DockerMonitor) getImageInfo() (map[string]interface{}, error) {
	// 这里应该使用Docker API获取镜像信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total_image_size_bytes": 10737418240, // 10GB
		"dangling_images":        3,
	}, nil
}

// getVolumeInfo 获取卷信息
func (m *DockerMonitor) getVolumeInfo() (map[string]interface{}, error) {
	// 这里应该使用Docker API获取卷信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"volumes_total":      8,
		"volumes_size_bytes": 5368709120, // 5GB
	}, nil
}

// getNetworkInfo 获取网络信息
func (m *DockerMonitor) getNetworkInfo() (map[string]interface{}, error) {
	// 这里应该使用Docker API获取网络信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"networks_total":    5,
		"bridge_networks":   3,
		"overlay_networks": 2,
	}, nil
}

// getDaemonInfo 获取Docker守护进程信息
func (m *DockerMonitor) getDaemonInfo() (map[string]interface{}, error) {
	// 这里应该使用Docker API获取守护进程信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"docker_version":   "24.0.7",
		"api_version":      "1.43",
		"kernel_version":   "5.15.0-91-generic",
		"operating_system": "Ubuntu 22.04.3 LTS",
		"architecture":     "x86_64",
	}, nil
}