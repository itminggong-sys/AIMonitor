//go:build linux
// +build linux

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"aimonitor-agents/common"
)

// LinuxElasticsearchMonitor Linux版本的Elasticsearch监控器
type LinuxElasticsearchMonitor struct {
	*ElasticsearchMonitor
	client *elasticsearch.Client
	baseURL string
	username string
	password string
}

// NewLinuxElasticsearchMonitor 创建Linux版本的Elasticsearch监控器
func NewLinuxElasticsearchMonitor(agent *common.Agent) *LinuxElasticsearchMonitor {
	baseMonitor := NewElasticsearchMonitor(agent)
	return &LinuxElasticsearchMonitor{
		ElasticsearchMonitor: baseMonitor,
		baseURL: "http://localhost:9200",
		username: "",
		password: "",
	}
}

// initClient 初始化Elasticsearch客户端
func (m *LinuxElasticsearchMonitor) initClient() error {
	if m.client != nil {
		return nil
	}

	cfg := elasticsearch.Config{
		Addresses: []string{m.baseURL},
	}

	if m.username != "" && m.password != "" {
		cfg.Username = m.username
		cfg.Password = m.password
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch client: %v", err)
	}

	m.client = client
	return nil
}

// makeRequest 发送HTTP请求
func (m *LinuxElasticsearchMonitor) makeRequest(method, endpoint string, body []byte) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := m.baseURL + endpoint

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if m.username != "" && m.password != "" {
		req.SetBasicAuth(m.username, m.password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// getClusterHealth 获取真实的集群健康状态
func (m *LinuxElasticsearchMonitor) getClusterHealth() (map[string]interface{}, error) {
	data, err := m.makeRequest("GET", "/_cluster/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster health: %v", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(data, &health); err != nil {
		return nil, fmt.Errorf("failed to parse cluster health: %v", err)
	}

	// 提取关键健康指标
	result := make(map[string]interface{})
	for key, value := range health {
		switch key {
		case "cluster_name", "status", "timed_out", "number_of_nodes", 
			 "number_of_data_nodes", "active_primary_shards", "active_shards",
			 "relocating_shards", "initializing_shards", "unassigned_shards",
			 "delayed_unassigned_shards", "number_of_pending_tasks",
			 "number_of_in_flight_fetch", "task_max_waiting_in_queue_millis",
			 "active_shards_percent_as_number":
			result[key] = value
		}
	}

	return result, nil
}

// getNodeStats 获取真实的节点统计
func (m *LinuxElasticsearchMonitor) getNodeStats() (map[string]interface{}, error) {
	data, err := m.makeRequest("GET", "/_nodes/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get node stats: %v", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse node stats: %v", err)
	}

	result := make(map[string]interface{})

	// 提取集群级别的统计信息
	if clusterName, ok := stats["cluster_name"]; ok {
		result["cluster_name"] = clusterName
	}

	// 处理节点统计
	if nodes, ok := stats["nodes"].(map[string]interface{}); ok {
		totalNodes := len(nodes)
		result["total_nodes"] = totalNodes

		// 聚合所有节点的统计信息
		totalMemoryUsed := int64(0)
		totalMemoryMax := int64(0)
		totalDiskUsed := int64(0)
		totalDiskAvailable := int64(0)
		totalCPUPercent := 0.0
		totalJVMHeapUsed := int64(0)
		totalJVMHeapMax := int64(0)

		for _, nodeData := range nodes {
			if node, ok := nodeData.(map[string]interface{}); ok {
				// OS统计
				if os, ok := node["os"].(map[string]interface{}); ok {
					if mem, ok := os["mem"].(map[string]interface{}); ok {
						if used, ok := mem["used_in_bytes"].(float64); ok {
							totalMemoryUsed += int64(used)
						}
						if total, ok := mem["total_in_bytes"].(float64); ok {
							totalMemoryMax += int64(total)
						}
					}
					if cpu, ok := os["cpu"].(map[string]interface{}); ok {
						if percent, ok := cpu["percent"].(float64); ok {
							totalCPUPercent += percent
						}
					}
				}

				// 文件系统统计
				if fs, ok := node["fs"].(map[string]interface{}); ok {
					if total, ok := fs["total"].(map[string]interface{}); ok {
						if used, ok := total["used_in_bytes"].(float64); ok {
							totalDiskUsed += int64(used)
						}
						if available, ok := total["available_in_bytes"].(float64); ok {
							totalDiskAvailable += int64(available)
						}
					}
				}

				// JVM统计
				if jvm, ok := node["jvm"].(map[string]interface{}); ok {
					if mem, ok := jvm["mem"].(map[string]interface{}); ok {
						if heapUsed, ok := mem["heap_used_in_bytes"].(float64); ok {
							totalJVMHeapUsed += int64(heapUsed)
						}
						if heapMax, ok := mem["heap_max_in_bytes"].(float64); ok {
							totalJVMHeapMax += int64(heapMax)
						}
					}
				}
			}
		}

		// 计算平均值和总计
		result["total_memory_used_bytes"] = totalMemoryUsed
		result["total_memory_max_bytes"] = totalMemoryMax
		result["total_disk_used_bytes"] = totalDiskUsed
		result["total_disk_available_bytes"] = totalDiskAvailable
		result["avg_cpu_percent"] = totalCPUPercent / float64(totalNodes)
		result["total_jvm_heap_used_bytes"] = totalJVMHeapUsed
		result["total_jvm_heap_max_bytes"] = totalJVMHeapMax

		// 计算使用率
		if totalMemoryMax > 0 {
			result["memory_usage_percent"] = float64(totalMemoryUsed) / float64(totalMemoryMax) * 100
		}
		if totalDiskUsed+totalDiskAvailable > 0 {
			totalDisk := totalDiskUsed + totalDiskAvailable
			result["disk_usage_percent"] = float64(totalDiskUsed) / float64(totalDisk) * 100
		}
		if totalJVMHeapMax > 0 {
			result["jvm_heap_usage_percent"] = float64(totalJVMHeapUsed) / float64(totalJVMHeapMax) * 100
		}
	}

	return result, nil
}

// getIndexStats 获取真实的索引统计
func (m *LinuxElasticsearchMonitor) getIndexStats() (map[string]interface{}, error) {
	data, err := m.makeRequest("GET", "/_stats", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get index stats: %v", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse index stats: %v", err)
	}

	result := make(map[string]interface{})

	// 提取总体统计
	if all, ok := stats["_all"].(map[string]interface{}); ok {
		if primaries, ok := all["primaries"].(map[string]interface{}); ok {
			// 文档统计
			if docs, ok := primaries["docs"].(map[string]interface{}); ok {
				if count, ok := docs["count"].(float64); ok {
					result["total_docs"] = int64(count)
				}
				if deleted, ok := docs["deleted"].(float64); ok {
					result["deleted_docs"] = int64(deleted)
				}
			}

			// 存储统计
			if store, ok := primaries["store"].(map[string]interface{}); ok {
				if size, ok := store["size_in_bytes"].(float64); ok {
					result["total_size_bytes"] = int64(size)
					result["total_size_mb"] = int64(size) / 1024 / 1024
				}
			}

			// 索引统计
			if indexing, ok := primaries["indexing"].(map[string]interface{}); ok {
				if indexTotal, ok := indexing["index_total"].(float64); ok {
					result["total_indexing_operations"] = int64(indexTotal)
				}
				if indexTime, ok := indexing["index_time_in_millis"].(float64); ok {
					result["total_indexing_time_ms"] = int64(indexTime)
				}
				if deleteTotal, ok := indexing["delete_total"].(float64); ok {
					result["total_delete_operations"] = int64(deleteTotal)
				}
			}

			// 搜索统计
			if search, ok := primaries["search"].(map[string]interface{}); ok {
				if queryTotal, ok := search["query_total"].(float64); ok {
					result["total_search_queries"] = int64(queryTotal)
				}
				if queryTime, ok := search["query_time_in_millis"].(float64); ok {
					result["total_search_time_ms"] = int64(queryTime)
				}
				if fetchTotal, ok := search["fetch_total"].(float64); ok {
					result["total_fetch_operations"] = int64(fetchTotal)
				}
			}
		}
	}

	// 获取索引列表
	if indices, ok := stats["indices"].(map[string]interface{}); ok {
		result["total_indices"] = len(indices)
		
		// 找出最大的索引
		largestIndex := ""
		largestSize := int64(0)
		
		for indexName, indexData := range indices {
			if index, ok := indexData.(map[string]interface{}); ok {
				if primaries, ok := index["primaries"].(map[string]interface{}); ok {
					if store, ok := primaries["store"].(map[string]interface{}); ok {
						if size, ok := store["size_in_bytes"].(float64); ok {
							indexSize := int64(size)
							if indexSize > largestSize {
								largestSize = indexSize
								largestIndex = indexName
							}
						}
					}
				}
			}
		}
		
		result["largest_index_name"] = largestIndex
		result["largest_index_size_bytes"] = largestSize
		result["largest_index_size_mb"] = largestSize / 1024 / 1024
	}

	return result, nil
}

// getFilesystemStats 获取真实的文件系统统计
func (m *LinuxElasticsearchMonitor) getFilesystemStats() (map[string]interface{}, error) {
	data, err := m.makeRequest("GET", "/_nodes/stats/fs", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %v", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse filesystem stats: %v", err)
	}

	result := make(map[string]interface{})

	if nodes, ok := stats["nodes"].(map[string]interface{}); ok {
		totalBytes := int64(0)
		freeBytes := int64(0)
		availableBytes := int64(0)

		for _, nodeData := range nodes {
			if node, ok := nodeData.(map[string]interface{}); ok {
				if fs, ok := node["fs"].(map[string]interface{}); ok {
					if total, ok := fs["total"].(map[string]interface{}); ok {
						if totalInBytes, ok := total["total_in_bytes"].(float64); ok {
							totalBytes += int64(totalInBytes)
						}
						if freeInBytes, ok := total["free_in_bytes"].(float64); ok {
							freeBytes += int64(freeInBytes)
						}
						if availableInBytes, ok := total["available_in_bytes"].(float64); ok {
							availableBytes += int64(availableInBytes)
						}
					}
				}
			}
		}

		result["total_bytes"] = totalBytes
		result["free_bytes"] = freeBytes
		result["available_bytes"] = availableBytes
		result["used_bytes"] = totalBytes - freeBytes

		// 转换为MB和GB
		result["total_mb"] = totalBytes / 1024 / 1024
		result["free_mb"] = freeBytes / 1024 / 1024
		result["used_mb"] = (totalBytes - freeBytes) / 1024 / 1024
		result["total_gb"] = totalBytes / 1024 / 1024 / 1024

		// 计算使用率
		if totalBytes > 0 {
			result["usage_percent"] = float64(totalBytes-freeBytes) / float64(totalBytes) * 100
		}
	}

	return result, nil
}

// getThreadPoolStats 获取真实的线程池统计
func (m *LinuxElasticsearchMonitor) getThreadPoolStats() (map[string]interface{}, error) {
	data, err := m.makeRequest("GET", "/_nodes/stats/thread_pool", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread pool stats: %v", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse thread pool stats: %v", err)
	}

	result := make(map[string]interface{})

	if nodes, ok := stats["nodes"].(map[string]interface{}); ok {
		// 聚合所有节点的线程池统计
		poolStats := make(map[string]map[string]int64)

		for _, nodeData := range nodes {
			if node, ok := nodeData.(map[string]interface{}); ok {
				if threadPool, ok := node["thread_pool"].(map[string]interface{}); ok {
					for poolName, poolData := range threadPool {
						if pool, ok := poolData.(map[string]interface{}); ok {
							if poolStats[poolName] == nil {
								poolStats[poolName] = make(map[string]int64)
							}

							if threads, ok := pool["threads"].(float64); ok {
								poolStats[poolName]["threads"] += int64(threads)
							}
							if queue, ok := pool["queue"].(float64); ok {
								poolStats[poolName]["queue"] += int64(queue)
							}
							if active, ok := pool["active"].(float64); ok {
								poolStats[poolName]["active"] += int64(active)
							}
							if rejected, ok := pool["rejected"].(float64); ok {
								poolStats[poolName]["rejected"] += int64(rejected)
							}
							if completed, ok := pool["completed"].(float64); ok {
								poolStats[poolName]["completed"] += int64(completed)
							}
						}
					}
				}
			}
		}

		// 将聚合结果添加到返回值
		for poolName, stats := range poolStats {
			poolResult := make(map[string]interface{})
			for statName, value := range stats {
				poolResult[statName] = value
			}
			result[poolName] = poolResult
		}
	}

	return result, nil
}

// getCacheStats 获取真实的缓存统计
func (m *LinuxElasticsearchMonitor) getCacheStats() (map[string]interface{}, error) {
	data, err := m.makeRequest("GET", "/_nodes/stats/indices/query_cache,request_cache,fielddata", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %v", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse cache stats: %v", err)
	}

	result := make(map[string]interface{})

	if nodes, ok := stats["nodes"].(map[string]interface{}); ok {
		totalQueryCacheSize := int64(0)
		totalQueryCacheHits := int64(0)
		totalQueryCacheMisses := int64(0)
		totalRequestCacheSize := int64(0)
		totalRequestCacheHits := int64(0)
		totalRequestCacheMisses := int64(0)
		totalFielddataSize := int64(0)

		for _, nodeData := range nodes {
			if node, ok := nodeData.(map[string]interface{}); ok {
				if indices, ok := node["indices"].(map[string]interface{}); ok {
					// Query Cache
					if queryCache, ok := indices["query_cache"].(map[string]interface{}); ok {
						if size, ok := queryCache["memory_size_in_bytes"].(float64); ok {
							totalQueryCacheSize += int64(size)
						}
						if hits, ok := queryCache["hit_count"].(float64); ok {
							totalQueryCacheHits += int64(hits)
						}
						if misses, ok := queryCache["miss_count"].(float64); ok {
							totalQueryCacheMisses += int64(misses)
						}
					}

					// Request Cache
					if requestCache, ok := indices["request_cache"].(map[string]interface{}); ok {
						if size, ok := requestCache["memory_size_in_bytes"].(float64); ok {
							totalRequestCacheSize += int64(size)
						}
						if hits, ok := requestCache["hit_count"].(float64); ok {
							totalRequestCacheHits += int64(hits)
						}
						if misses, ok := requestCache["miss_count"].(float64); ok {
							totalRequestCacheMisses += int64(misses)
						}
					}

					// Fielddata Cache
					if fielddata, ok := indices["fielddata"].(map[string]interface{}); ok {
						if size, ok := fielddata["memory_size_in_bytes"].(float64); ok {
							totalFielddataSize += int64(size)
						}
					}
				}
			}
		}

		// Query Cache统计
		result["query_cache_size_bytes"] = totalQueryCacheSize
		result["query_cache_size_mb"] = totalQueryCacheSize / 1024 / 1024
		result["query_cache_hits"] = totalQueryCacheHits
		result["query_cache_misses"] = totalQueryCacheMisses
		if totalQueryCacheHits+totalQueryCacheMisses > 0 {
			result["query_cache_hit_ratio"] = float64(totalQueryCacheHits) / float64(totalQueryCacheHits+totalQueryCacheMisses) * 100
		}

		// Request Cache统计
		result["request_cache_size_bytes"] = totalRequestCacheSize
		result["request_cache_size_mb"] = totalRequestCacheSize / 1024 / 1024
		result["request_cache_hits"] = totalRequestCacheHits
		result["request_cache_misses"] = totalRequestCacheMisses
		if totalRequestCacheHits+totalRequestCacheMisses > 0 {
			result["request_cache_hit_ratio"] = float64(totalRequestCacheHits) / float64(totalRequestCacheHits+totalRequestCacheMisses) * 100
		}

		// Fielddata Cache统计
		result["fielddata_cache_size_bytes"] = totalFielddataSize
		result["fielddata_cache_size_mb"] = totalFielddataSize / 1024 / 1024

		// 总缓存大小
		totalCacheSize := totalQueryCacheSize + totalRequestCacheSize + totalFielddataSize
		result["total_cache_size_bytes"] = totalCacheSize
		result["total_cache_size_mb"] = totalCacheSize / 1024 / 1024
	}

	return result, nil
}

// getClusterInfo 获取集群基本信息
func (m *LinuxElasticsearchMonitor) getClusterInfo() (map[string]interface{}, error) {
	// 获取集群信息
	data, err := m.makeRequest("GET", "/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %v", err)
	}

	var info map[string]interface{}
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse cluster info: %v", err)
	}

	result := make(map[string]interface{})

	// 基本信息
	if name, ok := info["name"]; ok {
		result["node_name"] = name
	}
	if clusterName, ok := info["cluster_name"]; ok {
		result["cluster_name"] = clusterName
	}
	if clusterUUID, ok := info["cluster_uuid"]; ok {
		result["cluster_uuid"] = clusterUUID
	}

	// 版本信息
	if version, ok := info["version"].(map[string]interface{}); ok {
		if number, ok := version["number"]; ok {
			result["version"] = number
		}
		if buildHash, ok := version["build_hash"]; ok {
			result["build_hash"] = buildHash
		}
		if buildDate, ok := version["build_date"]; ok {
			result["build_date"] = buildDate
		}
		if luceneVersion, ok := version["lucene_version"]; ok {
			result["lucene_version"] = luceneVersion
		}
	}

	return result, nil
}

// 重写collectMetrics方法以使用Linux特定的实现
func (m *LinuxElasticsearchMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 初始化客户端
	if err := m.initClient(); err != nil {
		m.agent.Logger.Error("Failed to initialize Elasticsearch client: %v", err)
		return m.ElasticsearchMonitor.collectMetrics() // 回退到模拟数据
	}

	// 使用Linux特定的方法收集指标
	if clusterInfo, err := m.getClusterInfo(); err == nil {
		for k, v := range clusterInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get cluster info: %v", err)
	}

	if clusterHealth, err := m.getClusterHealth(); err == nil {
		for k, v := range clusterHealth {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get cluster health: %v", err)
	}

	if nodeStats, err := m.getNodeStats(); err == nil {
		for k, v := range nodeStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get node stats: %v", err)
	}

	if indexStats, err := m.getIndexStats(); err == nil {
		for k, v := range indexStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get index stats: %v", err)
	}

	if fsStats, err := m.getFilesystemStats(); err == nil {
		for k, v := range fsStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get filesystem stats: %v", err)
	}

	if threadPoolStats, err := m.getThreadPoolStats(); err == nil {
		for k, v := range threadPoolStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get thread pool stats: %v", err)
	}

	if cacheStats, err := m.getCacheStats(); err == nil {
		for k, v := range cacheStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get cache stats: %v", err)
	}

	return metrics
}