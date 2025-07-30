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
	agent.Info.Type = "nginx"
	agent.Info.Name = "Nginx Monitor"

	// 创建Nginx监控器
	monitor := NewNginxMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Nginx Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("Nginx Agent stopped.")
}

// NginxMonitor Nginx监控器
type NginxMonitor struct {
	agent *common.Agent
}

// NewNginxMonitor 创建Nginx监控器
func NewNginxMonitor(agent *common.Agent) *NginxMonitor {
	return &NginxMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *NginxMonitor) StartMonitoring() {
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

// collectMetrics 收集Nginx指标
func (m *NginxMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 基本状态信息
	statusInfo, err := m.getStatusInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Nginx status info: %v", err)
	} else {
		metrics["active_connections"] = statusInfo["active_connections"]
		metrics["accepts"] = statusInfo["accepts"]
		metrics["handled"] = statusInfo["handled"]
		metrics["requests"] = statusInfo["requests"]
		metrics["reading"] = statusInfo["reading"]
		metrics["writing"] = statusInfo["writing"]
		metrics["waiting"] = statusInfo["waiting"]
	}

	// 请求统计
	requestStats, err := m.getRequestStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Nginx request stats: %v", err)
	} else {
		metrics["requests_per_second"] = requestStats["requests_per_second"]
		metrics["avg_request_time"] = requestStats["avg_request_time"]
		metrics["2xx_responses"] = requestStats["2xx_responses"]
		metrics["3xx_responses"] = requestStats["3xx_responses"]
		metrics["4xx_responses"] = requestStats["4xx_responses"]
		metrics["5xx_responses"] = requestStats["5xx_responses"]
	}

	// 上游服务器状态
	upstreamStats, err := m.getUpstreamStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Nginx upstream stats: %v", err)
	} else {
		metrics["upstream_servers_total"] = upstreamStats["upstream_servers_total"]
		metrics["upstream_servers_up"] = upstreamStats["upstream_servers_up"]
		metrics["upstream_servers_down"] = upstreamStats["upstream_servers_down"]
		metrics["upstream_requests"] = upstreamStats["upstream_requests"]
		metrics["upstream_responses"] = upstreamStats["upstream_responses"]
		metrics["upstream_avg_response_time"] = upstreamStats["upstream_avg_response_time"]
	}

	// SSL/TLS信息
	sslInfo, err := m.getSSLInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Nginx SSL info: %v", err)
	} else {
		metrics["ssl_handshakes"] = sslInfo["ssl_handshakes"]
		metrics["ssl_handshakes_failed"] = sslInfo["ssl_handshakes_failed"]
		metrics["ssl_session_reuses"] = sslInfo["ssl_session_reuses"]
	}

	// 缓存统计
	cacheStats, err := m.getCacheStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Nginx cache stats: %v", err)
	} else {
		metrics["cache_hits"] = cacheStats["cache_hits"]
		metrics["cache_misses"] = cacheStats["cache_misses"]
		metrics["cache_bypasses"] = cacheStats["cache_bypasses"]
		metrics["cache_expires"] = cacheStats["cache_expires"]
		metrics["cache_hit_ratio"] = cacheStats["cache_hit_ratio"]
	}

	// 进程信息
	processInfo, err := m.getProcessInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Nginx process info: %v", err)
	} else {
		metrics["worker_processes"] = processInfo["worker_processes"]
		metrics["worker_connections"] = processInfo["worker_connections"]
		metrics["master_process_running"] = processInfo["master_process_running"]
	}

	// 配置信息
	configInfo, err := m.getConfigInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Nginx config info: %v", err)
	} else {
		metrics["server_blocks"] = configInfo["server_blocks"]
		metrics["location_blocks"] = configInfo["location_blocks"]
		metrics["config_test_successful"] = configInfo["config_test_successful"]
	}

	return metrics
}

// getStatusInfo 获取状态信息
func (m *NginxMonitor) getStatusInfo() (map[string]interface{}, error) {
	// 这里应该从nginx status模块获取状态信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"active_connections": 100,
		"accepts":            50000,
		"handled":            49950,
		"requests":           150000,
		"reading":            5,
		"writing":            10,
		"waiting":            85,
	}, nil
}

// getRequestStats 获取请求统计
func (m *NginxMonitor) getRequestStats() (map[string]interface{}, error) {
	// 这里应该解析nginx访问日志或使用nginx-module-vts等模块
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"requests_per_second": 50.5,
		"avg_request_time":   0.25,
		"2xx_responses":      45000,
		"3xx_responses":      3000,
		"4xx_responses":      1500,
		"5xx_responses":      500,
	}, nil
}

// getUpstreamStats 获取上游服务器统计
func (m *NginxMonitor) getUpstreamStats() (map[string]interface{}, error) {
	// 这里应该从nginx upstream模块获取统计信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"upstream_servers_total":       6,
		"upstream_servers_up":          5,
		"upstream_servers_down":        1,
		"upstream_requests":            25000,
		"upstream_responses":           24800,
		"upstream_avg_response_time": 0.15,
	}, nil
}

// getSSLInfo 获取SSL信息
func (m *NginxMonitor) getSSLInfo() (map[string]interface{}, error) {
	// 这里应该从nginx SSL模块获取SSL统计信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"ssl_handshakes":        5000,
		"ssl_handshakes_failed": 50,
		"ssl_session_reuses":    4500,
	}, nil
}

// getCacheStats 获取缓存统计
func (m *NginxMonitor) getCacheStats() (map[string]interface{}, error) {
	// 这里应该从nginx缓存模块获取缓存统计信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"cache_hits":     8000,
		"cache_misses":   2000,
		"cache_bypasses": 500,
		"cache_expires":  300,
		"cache_hit_ratio": 80.0,
	}, nil
}

// getProcessInfo 获取进程信息
func (m *NginxMonitor) getProcessInfo() (map[string]interface{}, error) {
	// 这里应该检查nginx进程状态
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"worker_processes":        4,
		"worker_connections":      1024,
		"master_process_running": true,
	}, nil
}

// getConfigInfo 获取配置信息
func (m *NginxMonitor) getConfigInfo() (map[string]interface{}, error) {
	// 这里应该解析nginx配置文件
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"server_blocks":           10,
		"location_blocks":         25,
		"config_test_successful": true,
	}, nil
}