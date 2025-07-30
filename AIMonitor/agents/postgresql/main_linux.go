//go:build linux
// +build linux

package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"aimonitor-agents/common"
)

// LinuxPostgreSQLMonitor Linux版本的PostgreSQL监控器
type LinuxPostgreSQLMonitor struct {
	*PostgreSQLMonitor
	db *sql.DB
}

// NewLinuxPostgreSQLMonitor 创建Linux版本的PostgreSQL监控器
func NewLinuxPostgreSQLMonitor(agent *common.Agent) *LinuxPostgreSQLMonitor {
	baseMonitor := NewPostgreSQLMonitor(agent)
	return &LinuxPostgreSQLMonitor{
		PostgreSQLMonitor: baseMonitor,
	}
}

// initDBConnection 初始化数据库连接
func (m *LinuxPostgreSQLMonitor) initDBConnection() error {
	// 从配置文件读取连接信息
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "password"
	dbname := "postgres"
	sslmode := "disable"

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %v", err)
	}

	m.db = db
	return nil
}

// getConnectionInfo 获取真实的连接信息
func (m *LinuxPostgreSQLMonitor) getConnectionInfo() (map[string]interface{}, error) {
	if m.db == nil {
		if err := m.initDBConnection(); err != nil {
			return nil, err
		}
	}

	// 获取数据库版本
	var version string
	err := m.db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %v", err)
	}

	// 获取连接统计
	var maxConnections, activeConnections, idleConnections int
	err = m.db.QueryRow(`
		SELECT 
			setting::int as max_connections
		FROM pg_settings 
		WHERE name = 'max_connections'
	`).Scan(&maxConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to get max connections: %v", err)
	}

	err = m.db.QueryRow(`
		SELECT 
			COUNT(*) as active_connections
		FROM pg_stat_activity 
		WHERE state = 'active'
	`).Scan(&activeConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %v", err)
	}

	err = m.db.QueryRow(`
		SELECT 
			COUNT(*) as idle_connections
		FROM pg_stat_activity 
		WHERE state = 'idle'
	`).Scan(&idleConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to get idle connections: %v", err)
	}

	// 获取数据库启动时间
	var startTime string
	err = m.db.QueryRow("SELECT pg_postmaster_start_time()").Scan(&startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get start time: %v", err)
	}

	return map[string]interface{}{
		"server_version":     version,
		"server_status":      "online",
		"uptime":             startTime,
		"max_connections":    maxConnections,
		"active_connections": activeConnections,
		"idle_connections":   idleConnections,
		"used_connections":   activeConnections + idleConnections,
		"connection_usage_percent": float64(activeConnections+idleConnections) / float64(maxConnections) * 100,
	}, nil
}

// getDatabaseStats 获取真实的数据库统计
func (m *LinuxPostgreSQLMonitor) getDatabaseStats() (map[string]interface{}, error) {
	if m.db == nil {
		if err := m.initDBConnection(); err != nil {
			return nil, err
		}
	}

	// 获取数据库统计信息
	rows, err := m.db.Query(`
		SELECT 
			COUNT(*) as database_count,
			SUM(pg_database_size(datname))::bigint as total_size_bytes
		FROM pg_database 
		WHERE datistemplate = false
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %v", err)
	}
	defer rows.Close()

	var databaseCount int
	var totalSizeBytes int64
	if rows.Next() {
		err = rows.Scan(&databaseCount, &totalSizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan database stats: %v", err)
		}
	}

	// 获取事务统计
	var commitCount, rollbackCount int64
	err = m.db.QueryRow(`
		SELECT 
			SUM(xact_commit) as total_commits,
			SUM(xact_rollback) as total_rollbacks
		FROM pg_stat_database
	`).Scan(&commitCount, &rollbackCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction stats: %v", err)
	}

	// 获取缓存命中率
	var cacheHitRatio float64
	err = m.db.QueryRow(`
		SELECT 
			ROUND(
				SUM(blks_hit) * 100.0 / NULLIF(SUM(blks_hit) + SUM(blks_read), 0), 2
			) as cache_hit_ratio
		FROM pg_stat_database
	`).Scan(&cacheHitRatio)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache hit ratio: %v", err)
	}

	return map[string]interface{}{
		"database_count":      databaseCount,
		"total_size_bytes":    totalSizeBytes,
		"total_size_mb":       totalSizeBytes / 1024 / 1024,
		"total_commits":       commitCount,
		"total_rollbacks":     rollbackCount,
		"commit_ratio":        float64(commitCount) / float64(commitCount+rollbackCount) * 100,
		"cache_hit_ratio":     cacheHitRatio,
		"active_databases":    databaseCount, // 简化处理
	}, nil
}

// getTableStats 获取真实的表统计
func (m *LinuxPostgreSQLMonitor) getTableStats() (map[string]interface{}, error) {
	if m.db == nil {
		if err := m.initDBConnection(); err != nil {
			return nil, err
		}
	}

	// 获取表统计信息
	var tableCount, totalRows int64
	var totalTableSize int64
	err := m.db.QueryRow(`
		SELECT 
			COUNT(*) as table_count,
			COALESCE(SUM(n_tup_ins + n_tup_upd + n_tup_del), 0) as total_rows,
			COALESCE(SUM(pg_total_relation_size(schemaname||'.'||tablename)), 0) as total_size
		FROM pg_stat_user_tables
	`).Scan(&tableCount, &totalRows, &totalTableSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get table stats: %v", err)
	}

	// 获取最大表信息
	var largestTable string
	var largestTableSize int64
	err = m.db.QueryRow(`
		SELECT 
			schemaname||'.'||tablename as table_name,
			pg_total_relation_size(schemaname||'.'||tablename) as table_size
		FROM pg_stat_user_tables 
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC 
		LIMIT 1
	`).Scan(&largestTable, &largestTableSize)
	if err != nil {
		// 如果没有用户表，设置默认值
		largestTable = "N/A"
		largestTableSize = 0
	}

	// 获取表操作统计
	var totalInserts, totalUpdates, totalDeletes int64
	err = m.db.QueryRow(`
		SELECT 
			COALESCE(SUM(n_tup_ins), 0) as total_inserts,
			COALESCE(SUM(n_tup_upd), 0) as total_updates,
			COALESCE(SUM(n_tup_del), 0) as total_deletes
		FROM pg_stat_user_tables
	`).Scan(&totalInserts, &totalUpdates, &totalDeletes)
	if err != nil {
		return nil, fmt.Errorf("failed to get table operation stats: %v", err)
	}

	return map[string]interface{}{
		"table_count":         tableCount,
		"total_rows":          totalRows,
		"total_size_bytes":    totalTableSize,
		"total_size_mb":       totalTableSize / 1024 / 1024,
		"largest_table":       largestTable,
		"largest_table_size_mb": largestTableSize / 1024 / 1024,
		"total_inserts":       totalInserts,
		"total_updates":       totalUpdates,
		"total_deletes":       totalDeletes,
		"avg_table_size_mb":   float64(totalTableSize) / float64(tableCount) / 1024 / 1024,
	}, nil
}

// getIndexStats 获取真实的索引统计
func (m *LinuxPostgreSQLMonitor) getIndexStats() (map[string]interface{}, error) {
	if m.db == nil {
		if err := m.initDBConnection(); err != nil {
			return nil, err
		}
	}

	// 获取索引统计信息
	var indexCount int64
	var totalIndexSize int64
	var indexHitRatio float64
	err := m.db.QueryRow(`
		SELECT 
			COUNT(*) as index_count,
			COALESCE(SUM(pg_relation_size(indexrelid)), 0) as total_index_size,
			COALESCE(
				ROUND(
					SUM(idx_blks_hit) * 100.0 / NULLIF(SUM(idx_blks_hit) + SUM(idx_blks_read), 0), 2
				), 0
			) as index_hit_ratio
		FROM pg_stat_user_indexes
	`).Scan(&indexCount, &totalIndexSize, &indexHitRatio)
	if err != nil {
		return nil, fmt.Errorf("failed to get index stats: %v", err)
	}

	// 获取未使用的索引数量
	var unusedIndexes int64
	err = m.db.QueryRow(`
		SELECT COUNT(*) 
		FROM pg_stat_user_indexes 
		WHERE idx_scan = 0
	`).Scan(&unusedIndexes)
	if err != nil {
		return nil, fmt.Errorf("failed to get unused indexes count: %v", err)
	}

	// 获取索引扫描统计
	var totalIndexScans, totalIndexTupleReads int64
	err = m.db.QueryRow(`
		SELECT 
			COALESCE(SUM(idx_scan), 0) as total_index_scans,
			COALESCE(SUM(idx_tup_read), 0) as total_index_tuple_reads
		FROM pg_stat_user_indexes
	`).Scan(&totalIndexScans, &totalIndexTupleReads)
	if err != nil {
		return nil, fmt.Errorf("failed to get index scan stats: %v", err)
	}

	return map[string]interface{}{
		"index_count":            indexCount,
		"total_index_size_bytes": totalIndexSize,
		"total_index_size_mb":    totalIndexSize / 1024 / 1024,
		"index_hit_ratio":        indexHitRatio,
		"unused_indexes":         unusedIndexes,
		"used_indexes":           indexCount - unusedIndexes,
		"total_index_scans":      totalIndexScans,
		"total_index_tuple_reads": totalIndexTupleReads,
		"avg_index_size_mb":      float64(totalIndexSize) / float64(indexCount) / 1024 / 1024,
		"index_usage_ratio":      float64(indexCount-unusedIndexes) / float64(indexCount) * 100,
	}, nil
}

// getLockInfo 获取真实的锁信息
func (m *LinuxPostgreSQLMonitor) getLockInfo() (map[string]interface{}, error) {
	if m.db == nil {
		if err := m.initDBConnection(); err != nil {
			return nil, err
		}
	}

	// 获取锁统计信息
	rows, err := m.db.Query(`
		SELECT 
			mode,
			COUNT(*) as lock_count
		FROM pg_locks 
		GROUP BY mode
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get lock info: %v", err)
	}
	defer rows.Close()

	lockCounts := make(map[string]int)
	totalLocks := 0
	for rows.Next() {
		var mode string
		var count int
		err = rows.Scan(&mode, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lock info: %v", err)
		}
		lockCounts[mode] = count
		totalLocks += count
	}

	// 获取等待锁的查询数量
	var waitingQueries int
	err = m.db.QueryRow(`
		SELECT COUNT(*) 
		FROM pg_stat_activity 
		WHERE wait_event_type = 'Lock'
	`).Scan(&waitingQueries)
	if err != nil {
		return nil, fmt.Errorf("failed to get waiting queries count: %v", err)
	}

	return map[string]interface{}{
		"total_locks":       totalLocks,
		"access_share_locks": lockCounts["AccessShareLock"],
		"row_share_locks":   lockCounts["RowShareLock"],
		"row_exclusive_locks": lockCounts["RowExclusiveLock"],
		"share_locks":       lockCounts["ShareLock"],
		"exclusive_locks":   lockCounts["ExclusiveLock"],
		"waiting_queries":   waitingQueries,
		"blocked_queries":   waitingQueries, // 简化处理
		"deadlocks":         0,              // 需要从pg_stat_database获取
	}, nil
}

// getReplicationInfo 获取真实的复制信息
func (m *LinuxPostgreSQLMonitor) getReplicationInfo() (map[string]interface{}, error) {
	if m.db == nil {
		if err := m.initDBConnection(); err != nil {
			return nil, err
		}
	}

	// 检查是否为主服务器
	var isInRecovery bool
	err := m.db.QueryRow("SELECT pg_is_in_recovery()").Scan(&isInRecovery)
	if err != nil {
		return nil, fmt.Errorf("failed to check recovery status: %v", err)
	}

	result := map[string]interface{}{
		"is_master":    !isInRecovery,
		"is_slave":     isInRecovery,
		"replication_enabled": false,
	}

	if !isInRecovery {
		// 主服务器 - 获取复制槽信息
		var replicationSlots int
		err = m.db.QueryRow(`
			SELECT COUNT(*) 
			FROM pg_replication_slots
		`).Scan(&replicationSlots)
		if err == nil {
			result["replication_slots"] = replicationSlots
			result["replication_enabled"] = replicationSlots > 0
		}

		// 获取WAL发送进程信息
		var walSenders int
		err = m.db.QueryRow(`
			SELECT COUNT(*) 
			FROM pg_stat_replication
		`).Scan(&walSenders)
		if err == nil {
			result["wal_senders"] = walSenders
			result["connected_slaves"] = walSenders
		}
	} else {
		// 从服务器 - 获取复制延迟信息
		var replicationLag string
		err = m.db.QueryRow(`
			SELECT COALESCE(
				EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))::text,
				'0'
			)
		`).Scan(&replicationLag)
		if err == nil {
			if lag, parseErr := strconv.ParseFloat(replicationLag, 64); parseErr == nil {
				result["replication_lag_seconds"] = lag
			}
		}
	}

	return result, nil
}

// getPerformanceMetrics 获取真实的性能指标
func (m *LinuxPostgreSQLMonitor) getPerformanceMetrics() (map[string]interface{}, error) {
	if m.db == nil {
		if err := m.initDBConnection(); err != nil {
			return nil, err
		}
	}

	// 获取查询性能统计
	var avgQueryTime, slowQueries float64
	err := m.db.QueryRow(`
		SELECT 
			COALESCE(AVG(mean_time), 0) as avg_query_time,
			COALESCE(SUM(calls), 0) as total_queries
		FROM pg_stat_statements 
		WHERE mean_time > 1000
	`).Scan(&avgQueryTime, &slowQueries)
	if err != nil {
		// pg_stat_statements可能未启用，使用默认值
		avgQueryTime = 0
		slowQueries = 0
	}

	// 获取缓冲区统计
	var sharedBuffers, effectiveCacheSize string
	err = m.db.QueryRow(`
		SELECT 
			setting as shared_buffers
		FROM pg_settings 
		WHERE name = 'shared_buffers'
	`).Scan(&sharedBuffers)
	if err != nil {
		sharedBuffers = "unknown"
	}

	err = m.db.QueryRow(`
		SELECT 
			setting as effective_cache_size
		FROM pg_settings 
		WHERE name = 'effective_cache_size'
	`).Scan(&effectiveCacheSize)
	if err != nil {
		effectiveCacheSize = "unknown"
	}

	// 获取检查点统计
	var checkpoints, checkpointWriteTime float64
	err = m.db.QueryRow(`
		SELECT 
			checkpoints_timed + checkpoints_req as total_checkpoints,
			checkpoint_write_time
		FROM pg_stat_bgwriter
	`).Scan(&checkpoints, &checkpointWriteTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint stats: %v", err)
	}

	return map[string]interface{}{
		"avg_query_time_ms":     avgQueryTime,
		"slow_queries":          slowQueries,
		"shared_buffers":        sharedBuffers,
		"effective_cache_size":  effectiveCacheSize,
		"total_checkpoints":     checkpoints,
		"checkpoint_write_time": checkpointWriteTime,
		"queries_per_second":    0, // 需要计算
		"transactions_per_second": 0, // 需要计算
		"buffer_hit_ratio":     0, // 已在getDatabaseStats中计算
	}, nil
}

// Close 关闭数据库连接
func (m *LinuxPostgreSQLMonitor) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// 重写collectMetrics方法以使用Linux特定的实现
func (m *LinuxPostgreSQLMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 使用Linux特定的方法收集指标
	if connectionInfo, err := m.getConnectionInfo(); err == nil {
		for k, v := range connectionInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get connection info: %v", err)
	}

	if databaseStats, err := m.getDatabaseStats(); err == nil {
		for k, v := range databaseStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get database stats: %v", err)
	}

	if tableStats, err := m.getTableStats(); err == nil {
		for k, v := range tableStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get table stats: %v", err)
	}

	if indexStats, err := m.getIndexStats(); err == nil {
		for k, v := range indexStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get index stats: %v", err)
	}

	if lockInfo, err := m.getLockInfo(); err == nil {
		for k, v := range lockInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get lock info: %v", err)
	}

	if replicationInfo, err := m.getReplicationInfo(); err == nil {
		for k, v := range replicationInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get replication info: %v", err)
	}

	if performanceMetrics, err := m.getPerformanceMetrics(); err == nil {
		for k, v := range performanceMetrics {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get performance metrics: %v", err)
	}

	return metrics
}