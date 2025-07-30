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
	agent.Info.Type = "mysql"
	agent.Info.Name = "MySQL Monitor"

	// 创建MySQL监控器
	monitor := NewMySQLMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("MySQL Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("MySQL Agent stopped.")
}

// MySQLMonitor MySQL监控器
type MySQLMonitor struct {
	agent *common.Agent
}

// NewMySQLMonitor 创建MySQL监控器
func NewMySQLMonitor(agent *common.Agent) *MySQLMonitor {
	return &MySQLMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *MySQLMonitor) StartMonitoring() {
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

// collectMetrics 收集MySQL指标
func (m *MySQLMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 连接信息
	connectionInfo, err := m.getConnectionInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get MySQL connection info: %v", err)
	} else {
		metrics["threads_connected"] = connectionInfo["threads_connected"]
		metrics["threads_running"] = connectionInfo["threads_running"]
		metrics["max_connections"] = connectionInfo["max_connections"]
		metrics["connection_errors_total"] = connectionInfo["connection_errors_total"]
	}

	// 查询统计
	queryInfo, err := m.getQueryInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get MySQL query info: %v", err)
	} else {
		metrics["queries"] = queryInfo["queries"]
		metrics["questions"] = queryInfo["questions"]
		metrics["slow_queries"] = queryInfo["slow_queries"]
		metrics["select_scan"] = queryInfo["select_scan"]
		metrics["select_full_join"] = queryInfo["select_full_join"]
	}

	// InnoDB统计
	innodbInfo, err := m.getInnoDBInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get MySQL InnoDB info: %v", err)
	} else {
		metrics["innodb_buffer_pool_size"] = innodbInfo["innodb_buffer_pool_size"]
		metrics["innodb_buffer_pool_pages_total"] = innodbInfo["innodb_buffer_pool_pages_total"]
		metrics["innodb_buffer_pool_pages_free"] = innodbInfo["innodb_buffer_pool_pages_free"]
		metrics["innodb_buffer_pool_pages_dirty"] = innodbInfo["innodb_buffer_pool_pages_dirty"]
		metrics["innodb_buffer_pool_read_requests"] = innodbInfo["innodb_buffer_pool_read_requests"]
		metrics["innodb_buffer_pool_reads"] = innodbInfo["innodb_buffer_pool_reads"]
	}

	// 表锁信息
	lockInfo, err := m.getLockInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get MySQL lock info: %v", err)
	} else {
		metrics["table_locks_immediate"] = lockInfo["table_locks_immediate"]
		metrics["table_locks_waited"] = lockInfo["table_locks_waited"]
		metrics["innodb_row_lock_waits"] = lockInfo["innodb_row_lock_waits"]
		metrics["innodb_row_lock_time"] = lockInfo["innodb_row_lock_time"]
	}

	// 复制信息
	replicationInfo, err := m.getReplicationInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get MySQL replication info: %v", err)
	} else {
		metrics["slave_running"] = replicationInfo["slave_running"]
		metrics["seconds_behind_master"] = replicationInfo["seconds_behind_master"]
		metrics["slave_io_running"] = replicationInfo["slave_io_running"]
		metrics["slave_sql_running"] = replicationInfo["slave_sql_running"]
	}

	// 二进制日志信息
	binlogInfo, err := m.getBinlogInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get MySQL binlog info: %v", err)
	} else {
		metrics["binlog_cache_use"] = binlogInfo["binlog_cache_use"]
		metrics["binlog_cache_disk_use"] = binlogInfo["binlog_cache_disk_use"]
		metrics["binlog_stmt_cache_use"] = binlogInfo["binlog_stmt_cache_use"]
	}

	// 性能指标
	performanceInfo, err := m.getPerformanceInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get MySQL performance info: %v", err)
	} else {
		metrics["uptime"] = performanceInfo["uptime"]
		metrics["bytes_sent"] = performanceInfo["bytes_sent"]
		metrics["bytes_received"] = performanceInfo["bytes_received"]
		metrics["com_select"] = performanceInfo["com_select"]
		metrics["com_insert"] = performanceInfo["com_insert"]
		metrics["com_update"] = performanceInfo["com_update"]
		metrics["com_delete"] = performanceInfo["com_delete"]
	}

	return metrics
}

// getConnectionInfo 获取连接信息
func (m *MySQLMonitor) getConnectionInfo() (map[string]interface{}, error) {
	// 这里应该连接MySQL并执行SHOW STATUS LIKE 'Threads_%'等命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"threads_connected":       25,
		"threads_running":         5,
		"max_connections":         151,
		"connection_errors_total": 0,
	}, nil
}

// getQueryInfo 获取查询信息
func (m *MySQLMonitor) getQueryInfo() (map[string]interface{}, error) {
	// 这里应该连接MySQL并执行SHOW STATUS LIKE 'Queries'等命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"queries":           1000000,
		"questions":         950000,
		"slow_queries":      100,
		"select_scan":       5000,
		"select_full_join": 50,
	}, nil
}

// getInnoDBInfo 获取InnoDB信息
func (m *MySQLMonitor) getInnoDBInfo() (map[string]interface{}, error) {
	// 这里应该连接MySQL并执行SHOW STATUS LIKE 'Innodb_%'等命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"innodb_buffer_pool_size":         134217728, // 128MB
		"innodb_buffer_pool_pages_total":  8192,
		"innodb_buffer_pool_pages_free":   1024,
		"innodb_buffer_pool_pages_dirty":  512,
		"innodb_buffer_pool_read_requests": 1000000,
		"innodb_buffer_pool_reads":        10000,
	}, nil
}

// getLockInfo 获取锁信息
func (m *MySQLMonitor) getLockInfo() (map[string]interface{}, error) {
	// 这里应该连接MySQL并执行SHOW STATUS LIKE '%lock%'等命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"table_locks_immediate": 500000,
		"table_locks_waited":    100,
		"innodb_row_lock_waits": 50,
		"innodb_row_lock_time":  1000,
	}, nil
}

// getReplicationInfo 获取复制信息
func (m *MySQLMonitor) getReplicationInfo() (map[string]interface{}, error) {
	// 这里应该连接MySQL并执行SHOW SLAVE STATUS命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"slave_running":         true,
		"seconds_behind_master": 0,
		"slave_io_running":      "Yes",
		"slave_sql_running":     "Yes",
	}, nil
}

// getBinlogInfo 获取二进制日志信息
func (m *MySQLMonitor) getBinlogInfo() (map[string]interface{}, error) {
	// 这里应该连接MySQL并执行SHOW STATUS LIKE 'Binlog_%'等命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"binlog_cache_use":      10000,
		"binlog_cache_disk_use": 100,
		"binlog_stmt_cache_use": 5000,
	}, nil
}

// getPerformanceInfo 获取性能信息
func (m *MySQLMonitor) getPerformanceInfo() (map[string]interface{}, error) {
	// 这里应该连接MySQL并执行SHOW STATUS等命令
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"uptime":       86400, // 1天
		"bytes_sent":   1048576000,
		"bytes_received": 524288000,
		"com_select":   100000,
		"com_insert":   10000,
		"com_update":   5000,
		"com_delete":   1000,
	}, nil
}