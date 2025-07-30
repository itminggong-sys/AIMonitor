//go:build windows
// +build windows

package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"aimonitor-agents/common"
)

// WindowsPostgreSQLMonitor Windows版本的PostgreSQL监控器
type WindowsPostgreSQLMonitor struct {
	*PostgreSQLMonitor
	db           *sql.DB
	connectionString string
	host         string
	port         int
	database     string
	username     string
	password     string
}

// NewWindowsPostgreSQLMonitor 创建Windows版本的PostgreSQL监控器
func NewWindowsPostgreSQLMonitor(agent *common.Agent) *WindowsPostgreSQLMonitor {
	baseMonitor := NewPostgreSQLMonitor(agent)
	return &WindowsPostgreSQLMonitor{
		PostgreSQLMonitor: baseMonitor,
		host:              "localhost",
		port:              5432,
		database:          "postgres",
		username:          "postgres",
		password:          "password",
	}
}

// initDatabase 初始化数据库连接
func (m *WindowsPostgreSQLMonitor) initDatabase() error {
	if m.db != nil {
		return nil
	}

	// 构建连接字符串
	m.connectionString = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		m.host, m.port, m.username, m.password, m.database)

	// 创建数据库连接
	db, err := sql.Open("postgres", m.connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	m.db = db
	return nil
}

// getConnectionInfo 获取真实的连接信息
func (m *WindowsPostgreSQLMonitor) getConnectionInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取数据库版本
	var version string
	err := m.db.QueryRow("SELECT version()").Scan(&version)
	if err == nil {
		result["postgresql_version"] = version
		
		// 解析版本号
		if strings.Contains(version, "PostgreSQL") {
			parts := strings.Fields(version)
			if len(parts) >= 2 {
				result["postgresql_version_number"] = parts[1]
			}
		}
	} else {
		result["version_error"] = err.Error()
	}

	// 获取当前连接数
	var currentConnections int
	err = m.db.QueryRow("SELECT count(*) FROM pg_stat_activity").Scan(&currentConnections)
	if err == nil {
		result["current_connections"] = currentConnections
	}

	// 获取最大连接数
	var maxConnections int
	err = m.db.QueryRow("SHOW max_connections").Scan(&maxConnections)
	if err == nil {
		result["max_connections"] = maxConnections
		if currentConnections > 0 {
			result["connection_usage_percent"] = float64(currentConnections) / float64(maxConnections) * 100
		}
	}

	// 获取活跃连接数
	var activeConnections int
	err = m.db.QueryRow("SELECT count(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&activeConnections)
	if err == nil {
		result["active_connections"] = activeConnections
	}

	// 获取空闲连接数
	var idleConnections int
	err = m.db.QueryRow("SELECT count(*) FROM pg_stat_activity WHERE state = 'idle'").Scan(&idleConnections)
	if err == nil {
		result["idle_connections"] = idleConnections
	}

	// 获取等待连接数
	var waitingConnections int
	err = m.db.QueryRow("SELECT count(*) FROM pg_stat_activity WHERE wait_event_type IS NOT NULL").Scan(&waitingConnections)
	if err == nil {
		result["waiting_connections"] = waitingConnections
	}

	// 获取数据库启动时间
	var startTime time.Time
	err = m.db.QueryRow("SELECT pg_postmaster_start_time()").Scan(&startTime)
	if err == nil {
		result["start_time"] = startTime.Format(time.RFC3339)
		result["uptime_seconds"] = int64(time.Since(startTime).Seconds())
	}

	// 获取当前数据库名称
	var currentDatabase string
	err = m.db.QueryRow("SELECT current_database()").Scan(&currentDatabase)
	if err == nil {
		result["current_database"] = currentDatabase
	}

	// 获取当前用户
	var currentUser string
	err = m.db.QueryRow("SELECT current_user").Scan(&currentUser)
	if err == nil {
		result["current_user"] = currentUser
	}

	return result, nil
}

// getDatabaseStats 获取真实的数据库统计信息
func (m *WindowsPostgreSQLMonitor) getDatabaseStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取数据库列表和统计信息
	query := `
		SELECT 
			datname,
			numbackends,
			xact_commit,
			xact_rollback,
			blks_read,
			blks_hit,
			tup_returned,
			tup_fetched,
			tup_inserted,
			tup_updated,
			tup_deleted,
			conflicts,
			temp_files,
			temp_bytes,
			deadlocks,
			blk_read_time,
			blk_write_time
		FROM pg_stat_database 
		WHERE datname NOT IN ('template0', 'template1')
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database stats: %v", err)
	}
	defer rows.Close()

	databaseStats := make(map[string]interface{})
	totalDatabases := 0
	totalConnections := 0
	totalCommits := int64(0)
	totalRollbacks := int64(0)
	totalBlocksRead := int64(0)
	totalBlocksHit := int64(0)

	for rows.Next() {
		var (
			datname                                                    string
			numbackends                                                int
			xactCommit, xactRollback, blksRead, blksHit               int64
			tupReturned, tupFetched, tupInserted, tupUpdated, tupDeleted int64
			conflicts, tempFiles, tempBytes, deadlocks               int64
			blkReadTime, blkWriteTime                                float64
		)

		err := rows.Scan(
			&datname, &numbackends, &xactCommit, &xactRollback,
			&blksRead, &blksHit, &tupReturned, &tupFetched,
			&tupInserted, &tupUpdated, &tupDeleted, &conflicts,
			&tempFiles, &tempBytes, &deadlocks, &blkReadTime, &blkWriteTime,
		)
		if err != nil {
			continue
		}

		totalDatabases++
		totalConnections += numbackends
		totalCommits += xactCommit
		totalRollbacks += xactRollback
		totalBlocksRead += blksRead
		totalBlocksHit += blksHit

		// 计算缓存命中率
		cacheHitRatio := float64(0)
		if blksRead+blksHit > 0 {
			cacheHitRatio = float64(blksHit) / float64(blksRead+blksHit) * 100
		}

		databaseStats[datname] = map[string]interface{}{
			"connections":        numbackends,
			"commits":            xactCommit,
			"rollbacks":          xactRollback,
			"blocks_read":        blksRead,
			"blocks_hit":         blksHit,
			"cache_hit_ratio":    cacheHitRatio,
			"tuples_returned":    tupReturned,
			"tuples_fetched":     tupFetched,
			"tuples_inserted":    tupInserted,
			"tuples_updated":     tupUpdated,
			"tuples_deleted":     tupDeleted,
			"conflicts":          conflicts,
			"temp_files":         tempFiles,
			"temp_bytes":         tempBytes,
			"deadlocks":          deadlocks,
			"block_read_time":    blkReadTime,
			"block_write_time":   blkWriteTime,
		}
	}

	result["database_stats"] = databaseStats
	result["total_databases"] = totalDatabases
	result["total_connections"] = totalConnections
	result["total_commits"] = totalCommits
	result["total_rollbacks"] = totalRollbacks
	result["total_blocks_read"] = totalBlocksRead
	result["total_blocks_hit"] = totalBlocksHit

	// 计算总体缓存命中率
	if totalBlocksRead+totalBlocksHit > 0 {
		result["overall_cache_hit_ratio"] = float64(totalBlocksHit) / float64(totalBlocksRead+totalBlocksHit) * 100
	}

	return result, nil
}

// getTableStats 获取真实的表统计信息
func (m *WindowsPostgreSQLMonitor) getTableStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取表统计信息
	query := `
		SELECT 
			schemaname,
			tablename,
			seq_scan,
			seq_tup_read,
			idx_scan,
			idx_tup_fetch,
			n_tup_ins,
			n_tup_upd,
			n_tup_del,
			n_tup_hot_upd,
			n_live_tup,
			n_dead_tup,
			n_mod_since_analyze,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze,
			vacuum_count,
			autovacuum_count,
			analyze_count,
			autoanalyze_count
		FROM pg_stat_user_tables 
		ORDER BY seq_scan + COALESCE(idx_scan, 0) DESC 
		LIMIT 20
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table stats: %v", err)
	}
	defer rows.Close()

	tableStats := make([]map[string]interface{}, 0)
	totalTables := 0
	totalSeqScans := int64(0)
	totalIdxScans := int64(0)
	totalLiveTuples := int64(0)
	totalDeadTuples := int64(0)

	for rows.Next() {
		var (
			schemaname, tablename                                    string
			seqScan, seqTupRead, idxScan, idxTupFetch               sql.NullInt64
			nTupIns, nTupUpd, nTupDel, nTupHotUpd                   int64
			nLiveTup, nDeadTup, nModSinceAnalyze                    int64
			lastVacuum, lastAutovacuum, lastAnalyze, lastAutoanalyze sql.NullTime
			vacuumCount, autovacuumCount, analyzeCount, autoanalyzeCount int64
		)

		err := rows.Scan(
			&schemaname, &tablename, &seqScan, &seqTupRead,
			&idxScan, &idxTupFetch, &nTupIns, &nTupUpd,
			&nTupDel, &nTupHotUpd, &nLiveTup, &nDeadTup,
			&nModSinceAnalyze, &lastVacuum, &lastAutovacuum,
			&lastAnalyze, &lastAutoanalyze, &vacuumCount,
			&autovacuumCount, &analyzeCount, &autoanalyzeCount,
		)
		if err != nil {
			continue
		}

		totalTables++
		if seqScan.Valid {
			totalSeqScans += seqScan.Int64
		}
		if idxScan.Valid {
			totalIdxScans += idxScan.Int64
		}
		totalLiveTuples += nLiveTup
		totalDeadTuples += nDeadTup

		tableStat := map[string]interface{}{
			"schema":                schemaname,
			"table":                 tablename,
			"sequential_scans":      seqScan.Int64,
			"sequential_tuples_read": seqTupRead.Int64,
			"index_scans":           idxScan.Int64,
			"index_tuples_fetched":  idxTupFetch.Int64,
			"tuples_inserted":       nTupIns,
			"tuples_updated":        nTupUpd,
			"tuples_deleted":        nTupDel,
			"tuples_hot_updated":    nTupHotUpd,
			"live_tuples":           nLiveTup,
			"dead_tuples":           nDeadTup,
			"modified_since_analyze": nModSinceAnalyze,
			"vacuum_count":          vacuumCount,
			"autovacuum_count":      autovacuumCount,
			"analyze_count":         analyzeCount,
			"autoanalyze_count":     autoanalyzeCount,
		}

		// 添加时间戳
		if lastVacuum.Valid {
			tableStat["last_vacuum"] = lastVacuum.Time.Format(time.RFC3339)
		}
		if lastAutovacuum.Valid {
			tableStat["last_autovacuum"] = lastAutovacuum.Time.Format(time.RFC3339)
		}
		if lastAnalyze.Valid {
			tableStat["last_analyze"] = lastAnalyze.Time.Format(time.RFC3339)
		}
		if lastAutoanalyze.Valid {
			tableStat["last_autoanalyze"] = lastAutoanalyze.Time.Format(time.RFC3339)
		}

		// 计算死元组比例
		if nLiveTup+nDeadTup > 0 {
			tableStat["dead_tuple_ratio"] = float64(nDeadTup) / float64(nLiveTup+nDeadTup) * 100
		}

		tableStats = append(tableStats, tableStat)
	}

	result["table_stats"] = tableStats
	result["total_user_tables"] = totalTables
	result["total_sequential_scans"] = totalSeqScans
	result["total_index_scans"] = totalIdxScans
	result["total_live_tuples"] = totalLiveTuples
	result["total_dead_tuples"] = totalDeadTuples

	// 计算死元组比例
	if totalLiveTuples+totalDeadTuples > 0 {
		result["overall_dead_tuple_ratio"] = float64(totalDeadTuples) / float64(totalLiveTuples+totalDeadTuples) * 100
	}

	return result, nil
}

// getIndexStats 获取真实的索引统计信息
func (m *WindowsPostgreSQLMonitor) getIndexStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取索引统计信息
	query := `
		SELECT 
			schemaname,
			tablename,
			indexrelname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		FROM pg_stat_user_indexes 
		ORDER BY idx_scan DESC 
		LIMIT 20
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query index stats: %v", err)
	}
	defer rows.Close()

	indexStats := make([]map[string]interface{}, 0)
	totalIndexes := 0
	totalIndexScans := int64(0)
	totalIndexTuplesRead := int64(0)
	totalIndexTuplesFetched := int64(0)

	for rows.Next() {
		var (
			schemaname, tablename, indexname string
			idxScan, idxTupRead, idxTupFetch int64
		)

		err := rows.Scan(&schemaname, &tablename, &indexname, &idxScan, &idxTupRead, &idxTupFetch)
		if err != nil {
			continue
		}

		totalIndexes++
		totalIndexScans += idxScan
		totalIndexTuplesRead += idxTupRead
		totalIndexTuplesFetched += idxTupFetch

		indexStat := map[string]interface{}{
			"schema":              schemaname,
			"table":               tablename,
			"index":               indexname,
			"scans":               idxScan,
			"tuples_read":         idxTupRead,
			"tuples_fetched":      idxTupFetch,
		}

		// 计算索引效率
		if idxTupRead > 0 {
			indexStat["efficiency_ratio"] = float64(idxTupFetch) / float64(idxTupRead) * 100
		}

		indexStats = append(indexStats, indexStat)
	}

	result["index_stats"] = indexStats
	result["total_indexes"] = totalIndexes
	result["total_index_scans"] = totalIndexScans
	result["total_index_tuples_read"] = totalIndexTuplesRead
	result["total_index_tuples_fetched"] = totalIndexTuplesFetched

	return result, nil
}

// getLockInfo 获取真实的锁信息
func (m *WindowsPostgreSQLMonitor) getLockInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取锁信息
	query := `
		SELECT 
			mode,
			locktype,
			granted,
			COUNT(*) as count
		FROM pg_locks 
		GROUP BY mode, locktype, granted
		ORDER BY count DESC
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query lock info: %v", err)
	}
	defer rows.Close()

	lockStats := make([]map[string]interface{}, 0)
	totalLocks := 0
	grantedLocks := 0
	waitingLocks := 0

	for rows.Next() {
		var (
			mode, locktype string
			granted        bool
			count          int
		)

		err := rows.Scan(&mode, &locktype, &granted, &count)
		if err != nil {
			continue
		}

		totalLocks += count
		if granted {
			grantedLocks += count
		} else {
			waitingLocks += count
		}

		lockStat := map[string]interface{}{
			"mode":     mode,
			"type":     locktype,
			"granted":  granted,
			"count":    count,
		}

		lockStats = append(lockStats, lockStat)
	}

	result["lock_stats"] = lockStats
	result["total_locks"] = totalLocks
	result["granted_locks"] = grantedLocks
	result["waiting_locks"] = waitingLocks

	// 获取阻塞查询信息
	blockingQuery := `
		SELECT 
			blocked_locks.pid AS blocked_pid,
			blocked_activity.usename AS blocked_user,
			blocking_locks.pid AS blocking_pid,
			blocking_activity.usename AS blocking_user,
			blocked_activity.query AS blocked_statement,
			blocking_activity.query AS current_statement_in_blocking_process
		FROM pg_catalog.pg_locks blocked_locks
		JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
		JOIN pg_catalog.pg_locks blocking_locks 
			ON blocking_locks.locktype = blocked_locks.locktype
			AND blocking_locks.DATABASE IS NOT DISTINCT FROM blocked_locks.DATABASE
			AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
			AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
			AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
			AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
			AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
			AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
			AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
			AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
			AND blocking_locks.pid != blocked_locks.pid
		JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
		WHERE NOT blocked_locks.GRANTED
	`

	blockingRows, err := m.db.Query(blockingQuery)
	if err == nil {
		defer blockingRows.Close()
		blockingQueries := make([]map[string]interface{}, 0)

		for blockingRows.Next() {
			var (
				blockedPid, blockingPid                                   int
				blockedUser, blockingUser, blockedStmt, blockingStmt string
			)

			err := blockingRows.Scan(&blockedPid, &blockedUser, &blockingPid, &blockingUser, &blockedStmt, &blockingStmt)
			if err == nil {
				blockingQueries = append(blockingQueries, map[string]interface{}{
					"blocked_pid":       blockedPid,
					"blocked_user":      blockedUser,
					"blocking_pid":      blockingPid,
					"blocking_user":     blockingUser,
					"blocked_query":     blockedStmt,
					"blocking_query":    blockingStmt,
				})
			}
		}

		result["blocking_queries"] = blockingQueries
		result["blocking_queries_count"] = len(blockingQueries)
	}

	return result, nil
}

// getReplicationInfo 获取真实的复制信息
func (m *WindowsPostgreSQLMonitor) getReplicationInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 检查是否为主服务器
	var isInRecovery bool
	err := m.db.QueryRow("SELECT pg_is_in_recovery()").Scan(&isInRecovery)
	if err == nil {
		result["is_in_recovery"] = isInRecovery
		result["server_role"] = map[bool]string{true: "standby", false: "primary"}[isInRecovery]
	}

	if !isInRecovery {
		// 主服务器 - 获取复制槽信息
		replicationQuery := `
			SELECT 
				slot_name,
				plugin,
				slot_type,
				datoid,
				active,
				xmin,
				catalog_xmin,
				restart_lsn,
				confirmed_flush_lsn
			FROM pg_replication_slots
		`

		rows, err := m.db.Query(replicationQuery)
		if err == nil {
			defer rows.Close()
			replicationSlots := make([]map[string]interface{}, 0)

			for rows.Next() {
				var (
					slotName, plugin, slotType                    string
					datoid                                        sql.NullInt64
					active                                        bool
					xmin, catalogXmin                             sql.NullString
					restartLsn, confirmedFlushLsn                 sql.NullString
				)

				err := rows.Scan(&slotName, &plugin, &slotType, &datoid, &active, &xmin, &catalogXmin, &restartLsn, &confirmedFlushLsn)
				if err == nil {
					slot := map[string]interface{}{
						"slot_name": slotName,
						"plugin":    plugin,
						"slot_type": slotType,
						"active":    active,
					}

					if datoid.Valid {
						slot["database_oid"] = datoid.Int64
					}
					if xmin.Valid {
						slot["xmin"] = xmin.String
					}
					if catalogXmin.Valid {
						slot["catalog_xmin"] = catalogXmin.String
					}
					if restartLsn.Valid {
						slot["restart_lsn"] = restartLsn.String
					}
					if confirmedFlushLsn.Valid {
						slot["confirmed_flush_lsn"] = confirmedFlushLsn.String
					}

					replicationSlots = append(replicationSlots, slot)
				}
			}

			result["replication_slots"] = replicationSlots
			result["replication_slots_count"] = len(replicationSlots)
		}

		// 获取WAL发送者信息
		walSenderQuery := `
			SELECT 
				pid,
				usename,
				application_name,
				client_addr,
				client_hostname,
				client_port,
				backend_start,
				backend_xmin,
				state,
				sent_lsn,
				write_lsn,
				flush_lsn,
				replay_lsn,
				write_lag,
				flush_lag,
				replay_lag,
				sync_priority,
				sync_state
			FROM pg_stat_replication
		`

		walRows, err := m.db.Query(walSenderQuery)
		if err == nil {
			defer walRows.Close()
			walSenders := make([]map[string]interface{}, 0)

			for walRows.Next() {
				var (
					pid, clientPort, syncPriority                         int
					usename, appName, clientAddr, clientHostname, state, syncState string
					backendStart                                          time.Time
					backendXmin                                           sql.NullString
					sentLsn, writeLsn, flushLsn, replayLsn               sql.NullString
					writeLag, flushLag, replayLag                        sql.NullString
				)

				err := walRows.Scan(
					&pid, &usename, &appName, &clientAddr, &clientHostname, &clientPort,
					&backendStart, &backendXmin, &state, &sentLsn, &writeLsn, &flushLsn,
					&replayLsn, &writeLag, &flushLag, &replayLag, &syncPriority, &syncState,
				)
				if err == nil {
					walSender := map[string]interface{}{
						"pid":              pid,
						"username":         usename,
						"application_name": appName,
						"client_addr":      clientAddr,
						"client_hostname":  clientHostname,
						"client_port":      clientPort,
						"backend_start":    backendStart.Format(time.RFC3339),
						"state":            state,
						"sync_priority":    syncPriority,
						"sync_state":       syncState,
					}

					// 添加可选字段
					if backendXmin.Valid {
						walSender["backend_xmin"] = backendXmin.String
					}
					if sentLsn.Valid {
						walSender["sent_lsn"] = sentLsn.String
					}
					if writeLsn.Valid {
						walSender["write_lsn"] = writeLsn.String
					}
					if flushLsn.Valid {
						walSender["flush_lsn"] = flushLsn.String
					}
					if replayLsn.Valid {
						walSender["replay_lsn"] = replayLsn.String
					}
					if writeLag.Valid {
						walSender["write_lag"] = writeLag.String
					}
					if flushLag.Valid {
						walSender["flush_lag"] = flushLag.String
					}
					if replayLag.Valid {
						walSender["replay_lag"] = replayLag.String
					}

					walSenders = append(walSenders, walSender)
				}
			}

			result["wal_senders"] = walSenders
			result["wal_senders_count"] = len(walSenders)
		}
	} else {
		// 备用服务器 - 获取恢复信息
		var receivedLsn, lastMsgSendTime, lastMsgReceiptTime, latestEndLsn sql.NullString
		var lastMsgReceiptTimeVal sql.NullTime

		recoveryQuery := `
			SELECT 
				pg_last_wal_receive_lsn(),
				pg_last_wal_replay_lsn(),
				pg_last_xact_replay_timestamp()
		`

		err = m.db.QueryRow(recoveryQuery).Scan(&receivedLsn, &latestEndLsn, &lastMsgReceiptTimeVal)
		if err == nil {
			if receivedLsn.Valid {
				result["last_wal_receive_lsn"] = receivedLsn.String
			}
			if latestEndLsn.Valid {
				result["last_wal_replay_lsn"] = latestEndLsn.String
			}
			if lastMsgReceiptTimeVal.Valid {
				result["last_xact_replay_timestamp"] = lastMsgReceiptTimeVal.Time.Format(time.RFC3339)
			}
		}
	}

	return result, nil
}

// getPerformanceMetrics 获取真实的性能指标
func (m *WindowsPostgreSQLMonitor) getPerformanceMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取查询性能统计
	queryStatsQuery := `
		SELECT 
			queryid,
			query,
			calls,
			total_time,
			mean_time,
			min_time,
			max_time,
			stddev_time,
			rows,
			100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
		FROM pg_stat_statements 
		ORDER BY total_time DESC 
		LIMIT 10
	`

	rows, err := m.db.Query(queryStatsQuery)
	if err == nil {
		defer rows.Close()
		queryStats := make([]map[string]interface{}, 0)

		for rows.Next() {
			var (
				queryid                                                    int64
				query                                                      string
				calls                                                      int64
				totalTime, meanTime, minTime, maxTime, stddevTime, hitPercent sql.NullFloat64
				rowsReturned                                               int64
			)

			err := rows.Scan(&queryid, &query, &calls, &totalTime, &meanTime, &minTime, &maxTime, &stddevTime, &rowsReturned, &hitPercent)
			if err == nil {
				queryStat := map[string]interface{}{
					"query_id":     queryid,
					"query":        query,
					"calls":        calls,
					"rows":         rowsReturned,
				}

				if totalTime.Valid {
					queryStat["total_time"] = totalTime.Float64
				}
				if meanTime.Valid {
					queryStat["mean_time"] = meanTime.Float64
				}
				if minTime.Valid {
					queryStat["min_time"] = minTime.Float64
				}
				if maxTime.Valid {
					queryStat["max_time"] = maxTime.Float64
				}
				if stddevTime.Valid {
					queryStat["stddev_time"] = stddevTime.Float64
				}
				if hitPercent.Valid {
					queryStat["cache_hit_percent"] = hitPercent.Float64
				}

				queryStats = append(queryStats, queryStat)
			}
		}

		result["top_queries"] = queryStats
	} else {
		// pg_stat_statements扩展可能未安装
		result["query_stats_error"] = "pg_stat_statements extension not available"
	}

	// 获取等待事件统计
	waitEventsQuery := `
		SELECT 
			wait_event_type,
			wait_event,
			COUNT(*) as count
		FROM pg_stat_activity 
		WHERE wait_event_type IS NOT NULL 
		GROUP BY wait_event_type, wait_event 
		ORDER BY count DESC
	`

	waitRows, err := m.db.Query(waitEventsQuery)
	if err == nil {
		defer waitRows.Close()
		waitEvents := make([]map[string]interface{}, 0)

		for waitRows.Next() {
			var waitEventType, waitEvent string
			var count int

			err := waitRows.Scan(&waitEventType, &waitEvent, &count)
			if err == nil {
				waitEvents = append(waitEvents, map[string]interface{}{
					"wait_event_type": waitEventType,
					"wait_event":      waitEvent,
					"count":           count,
				})
			}
		}

		result["wait_events"] = waitEvents
	}

	// 获取检查点统计
	var checkpoints, timedCheckpoints, reqCheckpoints, bufWritten, bufCheckpoint int64
	var checkpointWriteTime, checkpointSyncTime float64

	checkpointQuery := `
		SELECT 
			checkpoints_timed,
			checkpoints_req,
			checkpoint_write_time,
			checkpoint_sync_time,
			buffers_checkpoint,
			buffers_clean,
			buffers_backend
		FROM pg_stat_bgwriter
	`

	err = m.db.QueryRow(checkpointQuery).Scan(
		&timedCheckpoints, &reqCheckpoints, &checkpointWriteTime, &checkpointSyncTime,
		&bufCheckpoint, &bufWritten, &bufWritten,
	)
	if err == nil {
		checkpoints = timedCheckpoints + reqCheckpoints
		result["checkpoints_total"] = checkpoints
		result["checkpoints_timed"] = timedCheckpoints
		result["checkpoints_requested"] = reqCheckpoints
		result["checkpoint_write_time"] = checkpointWriteTime
		result["checkpoint_sync_time"] = checkpointSyncTime
		result["buffers_checkpoint"] = bufCheckpoint
		result["buffers_written"] = bufWritten

		// 计算检查点频率
		if checkpoints > 0 {
			result["timed_checkpoint_ratio"] = float64(timedCheckpoints) / float64(checkpoints) * 100
		}
	}

	return result, nil
}

// 重写collectMetrics方法以使用Windows特定的实现
func (m *WindowsPostgreSQLMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 初始化数据库连接
	if err := m.initDatabase(); err != nil {
		m.agent.Logger.Error("Failed to initialize PostgreSQL connection: %v", err)
		return m.PostgreSQLMonitor.collectMetrics() // 回退到模拟数据
	}

	// 添加连接信息
	metrics["host"] = m.host
	metrics["port"] = m.port
	metrics["database"] = m.database
	metrics["username"] = m.username
	metrics["collection_time"] = time.Now().Format(time.RFC3339)

	// 使用Windows特定的方法收集指标
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

	// 如果所有数据库查询都失败，回退到模拟数据
	if len(metrics) <= 5 { // 只有基本连接信息
		m.agent.Logger.Warn("All PostgreSQL queries failed, falling back to simulated data")
		return m.PostgreSQLMonitor.collectMetrics()
	}

	return metrics
}

// Close 关闭数据库连接
func (m *WindowsPostgreSQLMonitor) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}