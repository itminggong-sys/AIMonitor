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
	agent.Info.Type = "apache"
	agent.Info.Name = "Apache Monitor"

	// 创建Apache监控器
	monitor := NewApacheMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Apache Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("Apache Agent stopped.")
}

// ApacheMonitor Apache监控器
type ApacheMonitor struct {
	agent *common.Agent
}

// NewApacheMonitor 创建Apache监控器
func NewApacheMonitor(agent *common.Agent) *ApacheMonitor {
	return &ApacheMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *ApacheMonitor) StartMonitoring() {
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

// collectMetrics 收集Apache指标
func (m *ApacheMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 服务器状态信息
	statusInfo, err := m.getStatusInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Apache status info: %v", err)
	} else {
		metrics["total_accesses"] = statusInfo["total_accesses"]
		metrics["total_kbytes"] = statusInfo["total_kbytes"]
		metrics["cpu_load"] = statusInfo["cpu_load"]
		metrics["uptime"] = statusInfo["uptime"]
		metrics["requests_per_sec"] = statusInfo["requests_per_sec"]
		metrics["bytes_per_sec"] = statusInfo["bytes_per_sec"]
		metrics["bytes_per_request"] = statusInfo["bytes_per_request"]
	}

	// 工作进程信息
	workerInfo, err := m.getWorkerInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Apache worker info: %v", err)
	} else {
		metrics["busy_workers"] = workerInfo["busy_workers"]
		metrics["idle_workers"] = workerInfo["idle_workers"]
		metrics["total_workers"] = workerInfo["total_workers"]
		metrics["worker_utilization"] = workerInfo["worker_utilization"]
	}

	// 连接信息
	connectionInfo, err := m.getConnectionInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Apache connection info: %v", err)
	} else {
		metrics["connections_total"] = connectionInfo["connections_total"]
		metrics["connections_async_writing"] = connectionInfo["connections_async_writing"]
		metrics["connections_async_keep_alive"] = connectionInfo["connections_async_keep_alive"]
		metrics["connections_async_closing"] = connectionInfo["connections_async_closing"]
	}

	// 虚拟主机统计
	vhostStats, err := m.getVhostStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Apache vhost stats: %v", err)
	} else {
		metrics["vhosts_total"] = vhostStats["vhosts_total"]
		metrics["vhost_requests"] = vhostStats["vhost_requests"]
		metrics["vhost_bytes_served"] = vhostStats["vhost_bytes_served"]
	}

	// 模块信息
	moduleInfo, err := m.getModuleInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Apache module info: %v", err)
	} else {
		metrics["loaded_modules"] = moduleInfo["loaded_modules"]
		metrics["ssl_enabled"] = moduleInfo["ssl_enabled"]
		metrics["rewrite_enabled"] = moduleInfo["rewrite_enabled"]
		metrics["compression_enabled"] = moduleInfo["compression_enabled"]
	}

	// 请求统计
	requestStats, err := m.getRequestStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Apache request stats: %v", err)
	} else {
		metrics["2xx_responses"] = requestStats["2xx_responses"]
		metrics["3xx_responses"] = requestStats["3xx_responses"]
		metrics["4xx_responses"] = requestStats["4xx_responses"]
		metrics["5xx_responses"] = requestStats["5xx_responses"]
		metrics["avg_response_time"] = requestStats["avg_response_time"]
	}

	// SSL信息
	sslInfo, err := m.getSSLInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Apache SSL info: %v", err)
	} else {
		metrics["ssl_sessions"] = sslInfo["ssl_sessions"]
		metrics["ssl_handshakes"] = sslInfo["ssl_handshakes"]
		metrics["ssl_handshake_failures"] = sslInfo["ssl_handshake_failures"]
	}

	// 缓存统计
	cacheStats, err := m.getCacheStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Apache cache stats: %v", err)
	} else {
		metrics["cache_hits"] = cacheStats["cache_hits"]
		metrics["cache_misses"] = cacheStats["cache_misses"]
		metrics["cache_hit_ratio"] = cacheStats["cache_hit_ratio"]
	}

	return metrics
}

// getStatusInfo 获取状态信息
func (m *ApacheMonitor) getStatusInfo() (map[string]interface{}, error) {
	// 这里应该从Apache server-status获取状态信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total_accesses":     100000,
		"total_kbytes":       1048576, // 1GB
		"cpu_load":           15.5,
		"uptime":             86400, // 1 day
		"requests_per_sec":   25.5,
		"bytes_per_sec":      12288, // 12KB/s
		"bytes_per_request": 512,
	}, nil
}

// getWorkerInfo 获取工作进程信息
func (m *ApacheMonitor) getWorkerInfo() (map[string]interface{}, error) {
	// 这里应该从Apache server-status获取工作进程信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"busy_workers":       25,
		"idle_workers":       75,
		"total_workers":      100,
		"worker_utilization": 25.0,
	}, nil
}

// getConnectionInfo 获取连接信息
func (m *ApacheMonitor) getConnectionInfo() (map[string]interface{}, error) {
	// 这里应该从Apache server-status获取连接信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"connections_total":              150,
		"connections_async_writing":      10,
		"connections_async_keep_alive":   50,
		"connections_async_closing":      5,
	}, nil
}

// getVhostStats 获取虚拟主机统计
func (m *ApacheMonitor) getVhostStats() (map[string]interface{}, error) {
	// 这里应该解析Apache配置文件和日志
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"vhosts_total":       5,
		"vhost_requests":     50000,
		"vhost_bytes_served": 524288000, // 500MB
	}, nil
}

// getModuleInfo 获取模块信息
func (m *ApacheMonitor) getModuleInfo() (map[string]interface{}, error) {
	// 这里应该从Apache server-info获取模块信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"loaded_modules":       45,
		"ssl_enabled":          true,
		"rewrite_enabled":      true,
		"compression_enabled": true,
	}, nil
}

// getRequestStats 获取请求统计
func (m *ApacheMonitor) getRequestStats() (map[string]interface{}, error) {
	// 这里应该解析Apache访问日志
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"2xx_responses":      85000,
		"3xx_responses":      8000,
		"4xx_responses":      5000,
		"5xx_responses":      2000,
		"avg_response_time": 0.35,
	}, nil
}

// getSSLInfo 获取SSL信息
func (m *ApacheMonitor) getSSLInfo() (map[string]interface{}, error) {
	// 这里应该从Apache SSL模块获取SSL统计
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"ssl_sessions":           5000,
		"ssl_handshakes":         5200,
		"ssl_handshake_failures": 50,
	}, nil
}

// getCacheStats 获取缓存统计
func (m *ApacheMonitor) getCacheStats() (map[string]interface{}, error) {
	// 这里应该从Apache缓存模块获取缓存统计
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"cache_hits":      15000,
		"cache_misses":    3000,
		"cache_hit_ratio": 83.3,
	}, nil
}