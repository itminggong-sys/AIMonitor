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
	agent.Info.Type = "postgresql"
	agent.Info.Name = "PostgreSQL Monitor"

	// 创建PostgreSQL监控器
	monitor := NewPostgreSQLMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("PostgreSQL Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("PostgreSQL Agent stopped.")
}

// PostgreSQLMonitor PostgreSQL监控器
type PostgreSQLMonitor struct {
	agent *common.Agent
}

// NewPostgreSQLMonitor 创建PostgreSQL监控器
func NewPostgreSQLMonitor(agent *common.Agent) *PostgreSQLMonitor {
	return &PostgreSQLMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *PostgreSQLMonitor) StartMonitoring() {
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

// collectMetrics 收集PostgreSQL指标
func (m *PostgreSQLMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 连接信息
	connectionInfo, err := m.getConnectionInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get PostgreSQL connection info: %v", err)
	} else {
		metrics["active_connections"] = connectionInfo["active_connections"]
		metrics["idle_connections"] = connectionInfo["idle_connections"]
		metrics["max_connections"] = connectionInfo["max_connections"]
		metrics["connection_usage_percent"] = connectionInfo["connection_usage_percent"]
	}

	// 数据库统计
	databaseStats, err := m.getDatabaseStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get PostgreSQL database stats: %v", err)
	} else {
		metrics["database_count"] = databaseStats["database_count"]
		metrics["total_size_bytes"] = databaseStats["total_size_bytes"]
		metrics["transactions_committed"] = databaseStats["transactions_committed"]
		metrics["transactions_rolled_back"] = databaseStats["transactions_rolled_back"]
		metrics["blocks_read"] = databaseStats["blocks_read"]
		metrics["blocks_hit"] = databaseStats["blocks_hit"]
		metrics["cache_hit_ratio"] = databaseStats["cache_hit_ratio"]
	}

	// 表统计
	tableStats, err := m.getTableStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get PostgreSQL table stats: %v", err)
	} else {
		metrics["total_tables"] = tableStats["total_tables"]
		metrics["total_rows"] = tableStats["total_rows"]
		metrics["sequential_scans"] = tableStats["sequential_scans"]
		metrics["index_scans"] = tableStats["index_scans"]
		metrics["rows_inserted"] = tableStats["rows_inserted"]
		metrics["rows_updated"] = tableStats["rows_updated"]
		metrics["rows_deleted"] = tableStats["rows_deleted"]
	}

	// 索引统计
	indexStats, err := m.getIndexStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get PostgreSQL index stats: %v", err)
	} else {
		metrics["total_indexes"] = indexStats["total_indexes"]
		metrics["index_size_bytes"] = indexStats["index_size_bytes"]
		metrics["index_scans_total"] = indexStats["index_scans_total"]
		metrics["unused_indexes"] = indexStats["unused_indexes"]
	}

	// 锁信息
	lockInfo, err := m.getLockInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get PostgreSQL lock info: %v", err)
	} else {
		metrics["active_locks"] = lockInfo["active_locks"]
		metrics["waiting_locks"] = lockInfo["waiting_locks"]
		metrics["deadlocks"] = lockInfo["deadlocks"]
	}

	// 复制信息
	replicationInfo, err := m.getReplicationInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get PostgreSQL replication info: %v", err)
	} else {
		metrics["is_master"] = replicationInfo["is_master"]
		metrics["replica_count"] = replicationInfo["replica_count"]
		metrics["replication_lag_bytes"] = replicationInfo["replication_lag_bytes"]
		metrics["wal_files_count"] = replicationInfo["wal_files_count"]
	}

	// 性能指标
	performanceInfo, err := m.getPerformanceInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get PostgreSQL performance info: %v", err)
	} else {
		metrics["queries_per_second"] = performanceInfo["queries_per_second"]
		metrics["avg_query_time_ms"] = performanceInfo["avg_query_time_ms"]
		metrics["slow_queries"] = performanceInfo["slow_queries"]
		metrics["checkpoint_write_time"] = performanceInfo["checkpoint_write_time"]
		metrics["checkpoint_sync_time"] = performanceInfo["checkpoint_sync_time"]
	}

	return metrics
}

// getConnectionInfo 获取连接信息
func (m *PostgreSQLMonitor) getConnectionInfo() (map[string]interface{}, error) {
	// 这里应该连接PostgreSQL并查询pg_stat_activity
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"active_connections":        25,
		"idle_connections":          15,
		"max_connections":           100,
		"connection_usage_percent": 40.0,
	}, nil
}

// getDatabaseStats 获取数据库统计
func (m *PostgreSQLMonitor) getDatabaseStats() (map[string]interface{}, error) {
	// 这里应该连接PostgreSQL并查询pg_stat_database
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"database_count":           5,
		"total_size_bytes":         1073741824, // 1GB
		"transactions_committed":   50000,
		"transactions_rolled_back": 100,
		"blocks_read":              10000,
		"blocks_hit":               90000,
		"cache_hit_ratio":          90.0,
	}, nil
}

// getTableStats 获取表统计
func (m *PostgreSQLMonitor) getTableStats() (map[string]interface{}, error) {
	// 这里应该连接PostgreSQL并查询pg_stat_user_tables
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total_tables":     50,
		"total_rows":       1000000,
		"sequential_scans": 1000,
		"index_scans":      50000,
		"rows_inserted":    10000,
		"rows_updated":     5000,
		"rows_deleted":     1000,
	}, nil
}

// getIndexStats 获取索引统计
func (m *PostgreSQLMonitor) getIndexStats() (map[string]interface{}, error) {
	// 这里应该连接PostgreSQL并查询pg_stat_user_indexes
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"total_indexes":      100,
		"index_size_bytes":   104857600, // 100MB
		"index_scans_total": 75000,
		"unused_indexes":     5,
	}, nil
}

// getLockInfo 获取锁信息
func (m *PostgreSQLMonitor) getLockInfo() (map[string]interface{}, error) {
	// 这里应该连接PostgreSQL并查询pg_locks
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"active_locks":  50,
		"waiting_locks": 2,
		"deadlocks":     0,
	}, nil
}

// getReplicationInfo 获取复制信息
func (m *PostgreSQLMonitor) getReplicationInfo() (map[string]interface{}, error) {
	// 这里应该连接PostgreSQL并查询pg_stat_replication
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"is_master":              true,
		"replica_count":          2,
		"replication_lag_bytes": 1024,
		"wal_files_count":        10,
	}, nil
}

// getPerformanceInfo 获取性能信息
func (m *PostgreSQLMonitor) getPerformanceInfo() (map[string]interface{}, error) {
	// 这里应该连接PostgreSQL并查询相关性能视图
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"queries_per_second":     100.0,
		"avg_query_time_ms":      25.5,
		"slow_queries":           5,
		"checkpoint_write_time": 1500,
		"checkpoint_sync_time":   200,
	}, nil
}