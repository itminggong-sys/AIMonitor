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
	agent.Info.Type = "elasticsearch"
	agent.Info.Name = "Elasticsearch Monitor"

	// 创建Elasticsearch监控器
	monitor := NewElasticsearchMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Elasticsearch Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("Elasticsearch Agent stopped.")
}

// ElasticsearchMonitor Elasticsearch监控器
type ElasticsearchMonitor struct {
	agent *common.Agent
}

// NewElasticsearchMonitor 创建Elasticsearch监控器
func NewElasticsearchMonitor(agent *common.Agent) *ElasticsearchMonitor {
	return &ElasticsearchMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *ElasticsearchMonitor) StartMonitoring() {
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

// collectMetrics 收集Elasticsearch指标
func (m *ElasticsearchMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 集群健康状态
	clusterHealth, err := m.getClusterHealth()
	if err != nil {
		m.agent.Logger.Error("Failed to get Elasticsearch cluster health: %v", err)
	} else {
		metrics["cluster_status"] = clusterHealth["cluster_status"]
		metrics["number_of_nodes"] = clusterHealth["number_of_nodes"]
		metrics["number_of_data_nodes"] = clusterHealth["number_of_data_nodes"]
		metrics["active_primary_shards"] = clusterHealth["active_primary_shards"]
		metrics["active_shards"] = clusterHealth["active_shards"]
		metrics["relocating_shards"] = clusterHealth["relocating_shards"]
		metrics["initializing_shards"] = clusterHealth["initializing_shards"]
		metrics["unassigned_shards"] = clusterHealth["unassigned_shards"]
		metrics["delayed_unassigned_shards"] = clusterHealth["delayed_unassigned_shards"]
		metrics["number_of_pending_tasks"] = clusterHealth["number_of_pending_tasks"]
		metrics["number_of_in_flight_fetch"] = clusterHealth["number_of_in_flight_fetch"]
		metrics["task_max_waiting_in_queue_millis"] = clusterHealth["task_max_waiting_in_queue_millis"]
		metrics["active_shards_percent_as_number"] = clusterHealth["active_shards_percent_as_number"]
	}

	// 节点统计
	nodeStats, err := m.getNodeStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Elasticsearch node stats: %v", err)
	} else {
		metrics["jvm_heap_used_percent"] = nodeStats["jvm_heap_used_percent"]
		metrics["jvm_heap_used_bytes"] = nodeStats["jvm_heap_used_bytes"]
		metrics["jvm_heap_max_bytes"] = nodeStats["jvm_heap_max_bytes"]
		metrics["jvm_gc_collectors_young_collection_count"] = nodeStats["jvm_gc_collectors_young_collection_count"]
		metrics["jvm_gc_collectors_young_collection_time_ms"] = nodeStats["jvm_gc_collectors_young_collection_time_ms"]
		metrics["jvm_gc_collectors_old_collection_count"] = nodeStats["jvm_gc_collectors_old_collection_count"]
		metrics["jvm_gc_collectors_old_collection_time_ms"] = nodeStats["jvm_gc_collectors_old_collection_time_ms"]
		metrics["process_cpu_percent"] = nodeStats["process_cpu_percent"]
		metrics["process_open_file_descriptors"] = nodeStats["process_open_file_descriptors"]
		metrics["process_max_file_descriptors"] = nodeStats["process_max_file_descriptors"]
	}

	// 索引统计
	indexStats, err := m.getIndexStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Elasticsearch index stats: %v", err)
	} else {
		metrics["indices_count"] = indexStats["indices_count"]
		metrics["indices_docs_count"] = indexStats["indices_docs_count"]
		metrics["indices_docs_deleted"] = indexStats["indices_docs_deleted"]
		metrics["indices_store_size_bytes"] = indexStats["indices_store_size_bytes"]
		metrics["indices_indexing_index_total"] = indexStats["indices_indexing_index_total"]
		metrics["indices_indexing_index_time_ms"] = indexStats["indices_indexing_index_time_ms"]
		metrics["indices_indexing_delete_total"] = indexStats["indices_indexing_delete_total"]
		metrics["indices_indexing_delete_time_ms"] = indexStats["indices_indexing_delete_time_ms"]
		metrics["indices_search_query_total"] = indexStats["indices_search_query_total"]
		metrics["indices_search_query_time_ms"] = indexStats["indices_search_query_time_ms"]
		metrics["indices_search_fetch_total"] = indexStats["indices_search_fetch_total"]
		metrics["indices_search_fetch_time_ms"] = indexStats["indices_search_fetch_time_ms"]
	}

	// 文件系统统计
	fsStats, err := m.getFilesystemStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Elasticsearch filesystem stats: %v", err)
	} else {
		metrics["fs_total_bytes"] = fsStats["fs_total_bytes"]
		metrics["fs_free_bytes"] = fsStats["fs_free_bytes"]
		metrics["fs_available_bytes"] = fsStats["fs_available_bytes"]
		metrics["fs_used_percent"] = fsStats["fs_used_percent"]
	}

	// 线程池统计
	threadPoolStats, err := m.getThreadPoolStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Elasticsearch thread pool stats: %v", err)
	} else {
		metrics["thread_pool_search_threads"] = threadPoolStats["thread_pool_search_threads"]
		metrics["thread_pool_search_queue"] = threadPoolStats["thread_pool_search_queue"]
		metrics["thread_pool_search_active"] = threadPoolStats["thread_pool_search_active"]
		metrics["thread_pool_search_rejected"] = threadPoolStats["thread_pool_search_rejected"]
		metrics["thread_pool_index_threads"] = threadPoolStats["thread_pool_index_threads"]
		metrics["thread_pool_index_queue"] = threadPoolStats["thread_pool_index_queue"]
		metrics["thread_pool_index_active"] = threadPoolStats["thread_pool_index_active"]
		metrics["thread_pool_index_rejected"] = threadPoolStats["thread_pool_index_rejected"]
	}

	// 缓存统计
	cacheStats, err := m.getCacheStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Elasticsearch cache stats: %v", err)
	} else {
		metrics["indices_query_cache_memory_size_bytes"] = cacheStats["indices_query_cache_memory_size_bytes"]
		metrics["indices_query_cache_total_count"] = cacheStats["indices_query_cache_total_count"]
		metrics["indices_query_cache_hit_count"] = cacheStats["indices_query_cache_hit_count"]
		metrics["indices_query_cache_miss_count"] = cacheStats["indices_query_cache_miss_count"]
		metrics["indices_query_cache_cache_size"] = cacheStats["indices_query_cache_cache_size"]
		metrics["indices_query_cache_cache_count"] = cacheStats["indices_query_cache_cache_count"]
		metrics["indices_query_cache_evictions"] = cacheStats["indices_query_cache_evictions"]
	}

	return metrics
}

// getClusterHealth 获取集群健康状态
func (m *ElasticsearchMonitor) getClusterHealth() (map[string]interface{}, error) {
	// 这里应该调用Elasticsearch _cluster/health API
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"cluster_status":                      "green",
		"number_of_nodes":                     3,
		"number_of_data_nodes":                3,
		"active_primary_shards":               15,
		"active_shards":                       30,
		"relocating_shards":                   0,
		"initializing_shards":                 0,
		"unassigned_shards":                   0,
		"delayed_unassigned_shards":           0,
		"number_of_pending_tasks":             0,
		"number_of_in_flight_fetch":           0,
		"task_max_waiting_in_queue_millis":    0,
		"active_shards_percent_as_number":     100.0,
	}, nil
}

// getNodeStats 获取节点统计
func (m *ElasticsearchMonitor) getNodeStats() (map[string]interface{}, error) {
	// 这里应该调用Elasticsearch _nodes/stats API
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"jvm_heap_used_percent":                      45.5,
		"jvm_heap_used_bytes":                        2147483648, // 2GB
		"jvm_heap_max_bytes":                         4294967296, // 4GB
		"jvm_gc_collectors_young_collection_count":   1000,
		"jvm_gc_collectors_young_collection_time_ms": 5000,
		"jvm_gc_collectors_old_collection_count":     50,
		"jvm_gc_collectors_old_collection_time_ms":   2000,
		"process_cpu_percent":                        25,
		"process_open_file_descriptors":              500,
		"process_max_file_descriptors":               65536,
	}, nil
}

// getIndexStats 获取索引统计
func (m *ElasticsearchMonitor) getIndexStats() (map[string]interface{}, error) {
	// 这里应该调用Elasticsearch _stats API
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"indices_count":                     10,
		"indices_docs_count":                1000000,
		"indices_docs_deleted":              5000,
		"indices_store_size_bytes":          10737418240, // 10GB
		"indices_indexing_index_total":      500000,
		"indices_indexing_index_time_ms":    120000,
		"indices_indexing_delete_total":     2500,
		"indices_indexing_delete_time_ms":   1000,
		"indices_search_query_total":        250000,
		"indices_search_query_time_ms":      75000,
		"indices_search_fetch_total":        200000,
		"indices_search_fetch_time_ms":      30000,
	}, nil
}

// getFilesystemStats 获取文件系统统计
func (m *ElasticsearchMonitor) getFilesystemStats() (map[string]interface{}, error) {
	// 这里应该从节点统计中获取文件系统信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"fs_total_bytes":     107374182400, // 100GB
		"fs_free_bytes":      53687091200,  // 50GB
		"fs_available_bytes": 53687091200,  // 50GB
		"fs_used_percent":    50.0,
	}, nil
}

// getThreadPoolStats 获取线程池统计
func (m *ElasticsearchMonitor) getThreadPoolStats() (map[string]interface{}, error) {
	// 这里应该从节点统计中获取线程池信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"thread_pool_search_threads":  13,
		"thread_pool_search_queue":    0,
		"thread_pool_search_active":   2,
		"thread_pool_search_rejected": 0,
		"thread_pool_index_threads":   4,
		"thread_pool_index_queue":     0,
		"thread_pool_index_active":    1,
		"thread_pool_index_rejected":  0,
	}, nil
}

// getCacheStats 获取缓存统计
func (m *ElasticsearchMonitor) getCacheStats() (map[string]interface{}, error) {
	// 这里应该从节点统计中获取缓存信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"indices_query_cache_memory_size_bytes": 104857600, // 100MB
		"indices_query_cache_total_count":       50000,
		"indices_query_cache_hit_count":         45000,
		"indices_query_cache_miss_count":        5000,
		"indices_query_cache_cache_size":        1000,
		"indices_query_cache_cache_count":       1000,
		"indices_query_cache_evictions":         100,
	}, nil
}