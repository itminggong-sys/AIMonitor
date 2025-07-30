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
	agent.Info.Type = "redis"
	agent.Info.Name = "Redis Monitor"

	// 创建Redis监控器
	monitor := NewRedisMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Redis Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("Redis Agent stopped.")
}

// RedisMonitor Redis监控器
type RedisMonitor struct {
	agent *common.Agent
}

// NewRedisMonitor 创建Redis监控器
func NewRedisMonitor(agent *common.Agent) *RedisMonitor {
	return &RedisMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *RedisMonitor) StartMonitoring() {
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

// collectMetrics 收集Redis指标
func (m *RedisMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Redis连接信息
	connectionInfo, err := m.getConnectionInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Redis connection info: %v", err)
	} else {
		metrics["connected_clients"] = connectionInfo["connected_clients"]
		metrics["blocked_clients"] = connectionInfo["blocked_clients"]
		metrics["total_connections_received"] = connectionInfo["total_connections_received"]
		metrics["rejected_connections"] = connectionInfo["rejected_connections"]
	}

	// 内存使用情况
	memoryInfo, err := m.getMemoryInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Redis memory info: %v", err)
	} else {
		metrics["used_memory"] = memoryInfo["used_memory"]
		metrics["used_memory_rss"] = memoryInfo["used_memory_rss"]
		metrics["used_memory_peak"] = memoryInfo["used_memory_peak"]
		metrics["memory_fragmentation_ratio"] = memoryInfo["memory_fragmentation_ratio"]
	}

	// 键空间统计
	keyspaceInfo, err := m.getKeyspaceInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Redis keyspace info: %v", err)
	} else {
		metrics["total_keys"] = keyspaceInfo["total_keys"]
		metrics["expired_keys"] = keyspaceInfo["expired_keys"]
		metrics["evicted_keys"] = keyspaceInfo["evicted_keys"]
		metrics["keyspace_hits"] = keyspaceInfo["keyspace_hits"]
		metrics["keyspace_misses"] = keyspaceInfo["keyspace_misses"]
	}

	// 性能统计
	performanceInfo, err := m.getPerformanceInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Redis performance info: %v", err)
	} else {
		metrics["total_commands_processed"] = performanceInfo["total_commands_processed"]
		metrics["instantaneous_ops_per_sec"] = performanceInfo["instantaneous_ops_per_sec"]
		metrics["instantaneous_input_kbps"] = performanceInfo["instantaneous_input_kbps"]
		metrics["instantaneous_output_kbps"] = performanceInfo["instantaneous_output_kbps"]
	}

	// 持久化信息
	persistenceInfo, err := m.getPersistenceInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Redis persistence info: %v", err)
	} else {
		metrics["rdb_last_save_time"] = persistenceInfo["rdb_last_save_time"]
		metrics["rdb_changes_since_last_save"] = persistenceInfo["rdb_changes_since_last_save"]
		metrics["aof_enabled"] = persistenceInfo["aof_enabled"]
		metrics["aof_rewrite_in_progress"] = persistenceInfo["aof_rewrite_in_progress"]
	}

	// 复制信息
	replicationInfo, err := m.getReplicationInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Redis replication info: %v", err)
	} else {
		metrics["role"] = replicationInfo["role"]
		metrics["connected_slaves"] = replicationInfo["connected_slaves"]
		metrics["master_repl_offset"] = replicationInfo["master_repl_offset"]
	}

	return metrics
}

// getConnectionInfo 获取连接信息
func (m *RedisMonitor) getConnectionInfo() (map[string]interface{}, error) {
	// 这里应该连接Redis并执行INFO clients命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"connected_clients":          50,
		"blocked_clients":            2,
		"total_connections_received": 10000,
		"rejected_connections":       0,
	}, nil
}

// getMemoryInfo 获取内存信息
func (m *RedisMonitor) getMemoryInfo() (map[string]interface{}, error) {
	// 这里应该连接Redis并执行INFO memory命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"used_memory":                 104857600, // 100MB
		"used_memory_rss":             125829120, // 120MB
		"used_memory_peak":            134217728, // 128MB
		"memory_fragmentation_ratio": 1.2,
	}, nil
}

// getKeyspaceInfo 获取键空间信息
func (m *RedisMonitor) getKeyspaceInfo() (map[string]interface{}, error) {
	// 这里应该连接Redis并执行INFO keyspace命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total_keys":      10000,
		"expired_keys":    500,
		"evicted_keys":    10,
		"keyspace_hits":   50000,
		"keyspace_misses": 1000,
	}, nil
}

// getPerformanceInfo 获取性能信息
func (m *RedisMonitor) getPerformanceInfo() (map[string]interface{}, error) {
	// 这里应该连接Redis并执行INFO stats命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total_commands_processed":   1000000,
		"instantaneous_ops_per_sec":  100,
		"instantaneous_input_kbps":   50.5,
		"instantaneous_output_kbps": 75.2,
	}, nil
}

// getPersistenceInfo 获取持久化信息
func (m *RedisMonitor) getPersistenceInfo() (map[string]interface{}, error) {
	// 这里应该连接Redis并执行INFO persistence命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"rdb_last_save_time":         time.Now().Unix() - 3600, // 1小时前
		"rdb_changes_since_last_save": 100,
		"aof_enabled":                 true,
		"aof_rewrite_in_progress":     false,
	}, nil
}

// getReplicationInfo 获取复制信息
func (m *RedisMonitor) getReplicationInfo() (map[string]interface{}, error) {
	// 这里应该连接Redis并执行INFO replication命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"role":                "master",
		"connected_slaves":    2,
		"master_repl_offset": 1000000,
	}, nil
}