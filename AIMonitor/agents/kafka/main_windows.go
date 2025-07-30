//go:build windows
// +build windows

package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"aimonitor-agents/common"
)

// WindowsKafkaMonitor Windows版本的Kafka监控器
type WindowsKafkaMonitor struct {
	*KafkaMonitor
	client       sarama.Client
	clusterAdmin sarama.ClusterAdmin
	brokers      []string
	topics       []string
}

// NewWindowsKafkaMonitor 创建Windows版本的Kafka监控器
func NewWindowsKafkaMonitor(agent *common.Agent) *WindowsKafkaMonitor {
	baseMonitor := NewKafkaMonitor(agent)
	return &WindowsKafkaMonitor{
		KafkaMonitor: baseMonitor,
		brokers:      []string{"localhost:9092"},
	}
}

// initKafkaClient 初始化Kafka客户端
func (m *WindowsKafkaMonitor) initKafkaClient() error {
	if m.client != nil {
		return nil
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Return.Errors = true
	config.Metadata.RefreshFrequency = 10 * time.Second
	config.Metadata.Full = true

	// 创建客户端
	client, err := sarama.NewClient(m.brokers, config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka client: %v", err)
	}
	m.client = client

	// 创建集群管理员
	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to create cluster admin: %v", err)
	}
	m.clusterAdmin = admin

	return nil
}

// getBrokerInfo 获取真实的Broker信息
func (m *WindowsKafkaMonitor) getBrokerInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	brokers := m.client.Brokers()
	result["total_brokers"] = len(brokers)

	brokerDetails := make([]map[string]interface{}, 0)
	for _, broker := range brokers {
		brokerInfo := map[string]interface{}{
			"id":      broker.ID(),
			"address": broker.Addr(),
		}

		// 检查Broker连接状态
		connected, err := broker.Connected()
		if err == nil {
			brokerInfo["connected"] = connected
		} else {
			brokerInfo["connected"] = false
			brokerInfo["connection_error"] = err.Error()
		}

		// 尝试打开连接以获取更多信息
		if !connected {
			err = broker.Open(m.client.Config())
			if err == nil {
				brokerInfo["open_success"] = true
				defer broker.Close()
			} else {
				brokerInfo["open_error"] = err.Error()
			}
		}

		brokerDetails = append(brokerDetails, brokerInfo)
	}

	result["broker_details"] = brokerDetails

	// 获取控制器信息
	controllerID, err := m.client.Controller()
	if err == nil {
		result["controller_id"] = controllerID.ID()
		result["controller_address"] = controllerID.Addr()
	} else {
		result["controller_error"] = err.Error()
	}

	return result, nil
}

// getTopicInfo 获取真实的Topic信息
func (m *WindowsKafkaMonitor) getTopicInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取所有Topic
	topics, err := m.client.Topics()
	if err != nil {
		return nil, fmt.Errorf("failed to get topics: %v", err)
	}

	result["total_topics"] = len(topics)
	m.topics = topics

	// 获取Topic详细信息
	topicDetails := make(map[string]interface{})
	totalPartitions := 0
	totalReplicas := 0

	for _, topic := range topics {
		// 跳过内部Topic
		if strings.HasPrefix(topic, "__") {
			continue
		}

		// 获取Topic元数据
		partitions, err := m.client.Partitions(topic)
		if err != nil {
			continue
		}

		topicInfo := map[string]interface{}{
			"partitions": len(partitions),
		}

		totalPartitions += len(partitions)

		// 获取每个分区的副本信息
		replicaCount := 0
		leaderCount := 0
		partitionDetails := make([]map[string]interface{}, 0)

		for _, partition := range partitions {
			replicas, err := m.client.Replicas(topic, partition)
			if err == nil {
				replicaCount += len(replicas)
				totalReplicas += len(replicas)

				// 获取Leader信息
				leader, err := m.client.Leader(topic, partition)
				partitionInfo := map[string]interface{}{
					"partition_id": partition,
					"replicas":     len(replicas),
					"replica_ids":  replicas,
				}

				if err == nil {
					partitionInfo["leader_id"] = leader.ID()
					partitionInfo["leader_address"] = leader.Addr()
					leaderCount++
				} else {
					partitionInfo["leader_error"] = err.Error()
				}

				partitionDetails = append(partitionDetails, partitionInfo)
			}
		}

		topicInfo["total_replicas"] = replicaCount
		topicInfo["leaders_available"] = leaderCount
		topicInfo["partition_details"] = partitionDetails

		topicDetails[topic] = topicInfo
	}

	result["topic_details"] = topicDetails
	result["total_partitions"] = totalPartitions
	result["total_replicas"] = totalReplicas

	return result, nil
}

// getMessageStats 获取真实的消息统计
func (m *WindowsKafkaMonitor) getMessageStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if len(m.topics) == 0 {
		return result, nil
	}

	totalMessages := int64(0)
	totalSize := int64(0)
	topicStats := make(map[string]interface{})

	for _, topic := range m.topics {
		// 跳过内部Topic
		if strings.HasPrefix(topic, "__") {
			continue
		}

		partitions, err := m.client.Partitions(topic)
		if err != nil {
			continue
		}

		topicMessages := int64(0)
		topicSize := int64(0)
		partitionStats := make([]map[string]interface{}, 0)

		for _, partition := range partitions {
			// 获取最新偏移量
			newestOffset, err := m.client.GetOffset(topic, partition, sarama.OffsetNewest)
			if err != nil {
				continue
			}

			// 获取最旧偏移量
			oldestOffset, err := m.client.GetOffset(topic, partition, sarama.OffsetOldest)
			if err != nil {
				continue
			}

			messageCount := newestOffset - oldestOffset
			topicMessages += messageCount
			totalMessages += messageCount

			// 估算大小（假设每条消息平均1KB）
			estimatedSize := messageCount * 1024
			topicSize += estimatedSize
			totalSize += estimatedSize

			partitionStat := map[string]interface{}{
				"partition_id":    partition,
				"oldest_offset":   oldestOffset,
				"newest_offset":   newestOffset,
				"message_count":   messageCount,
				"estimated_size":  estimatedSize,
			}

			partitionStats = append(partitionStats, partitionStat)
		}

		topicStat := map[string]interface{}{
			"message_count":     topicMessages,
			"estimated_size":    topicSize,
			"partition_count":   len(partitions),
			"partition_stats":   partitionStats,
		}

		topicStats[topic] = topicStat
	}

	result["total_messages"] = totalMessages
	result["total_estimated_size"] = totalSize
	result["topic_stats"] = topicStats

	return result, nil
}

// getConsumerGroupInfo 获取真实的消费者组信息
func (m *WindowsKafkaMonitor) getConsumerGroupInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取所有消费者组
	groups, err := m.clusterAdmin.ListConsumerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer groups: %v", err)
	}

	result["total_consumer_groups"] = len(groups)

	groupDetails := make(map[string]interface{})
	activeGroups := 0
	emptyGroups := 0

	for groupID := range groups {
		// 获取消费者组详细信息
		groupDesc, err := m.clusterAdmin.DescribeConsumerGroups([]string{groupID})
		if err != nil {
			continue
		}

		if desc, ok := groupDesc[groupID]; ok {
			groupInfo := map[string]interface{}{
				"group_id":       groupID,
				"state":          desc.State,
				"protocol_type":  desc.ProtocolType,
				"protocol":       desc.Protocol,
				"member_count":   len(desc.Members),
			}

			if desc.State == "Stable" {
				activeGroups++
			}

			if len(desc.Members) == 0 {
				emptyGroups++
			}

			// 获取成员详细信息
			memberDetails := make([]map[string]interface{}, 0)
			for memberID, member := range desc.Members {
				memberInfo := map[string]interface{}{
					"member_id":   memberID,
					"client_id":   member.ClientId,
					"client_host": member.ClientHost,
				}
				memberDetails = append(memberDetails, memberInfo)
			}
			groupInfo["members"] = memberDetails

			// 获取消费者组偏移量信息
			offsets, err := m.clusterAdmin.ListConsumerGroupOffsets(groupID, nil)
			if err == nil {
				offsetDetails := make(map[string]interface{})
				totalLag := int64(0)

				for topic, partitions := range offsets.Blocks {
					topicOffsets := make(map[string]interface{})
					for partition, offsetInfo := range partitions {
						// 获取当前分区的最新偏移量
						latestOffset, err := m.client.GetOffset(topic, partition, sarama.OffsetNewest)
						if err == nil {
							lag := latestOffset - offsetInfo.Offset
							totalLag += lag

							topicOffsets[strconv.Itoa(int(partition))] = map[string]interface{}{
								"offset":        offsetInfo.Offset,
								"latest_offset": latestOffset,
								"lag":           lag,
								"metadata":      offsetInfo.Metadata,
							}
						}
					}
					offsetDetails[topic] = topicOffsets
				}

				groupInfo["offsets"] = offsetDetails
				groupInfo["total_lag"] = totalLag
			}

			groupDetails[groupID] = groupInfo
		}
	}

	result["group_details"] = groupDetails
	result["active_groups"] = activeGroups
	result["empty_groups"] = emptyGroups

	return result, nil
}

// getPerformanceMetrics 获取真实的性能指标
func (m *WindowsKafkaMonitor) getPerformanceMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 测量连接延迟
	start := time.Now()
	_, err := m.client.Coordinator("test-group")
	latency := time.Since(start)

	result["connection_latency_ms"] = latency.Milliseconds()
	if err != nil {
		result["connection_error"] = err.Error()
	}

	// 测量元数据刷新时间
	start = time.Now()
	err = m.client.RefreshMetadata()
	metadataRefreshTime := time.Since(start)

	result["metadata_refresh_time_ms"] = metadataRefreshTime.Milliseconds()
	if err != nil {
		result["metadata_refresh_error"] = err.Error()
	}

	// 获取集群配置信息
	if m.clusterAdmin != nil {
		// 获取Broker配置
		brokers := m.client.Brokers()
		if len(brokers) > 0 {
			brokerID := brokers[0].ID()
			configs, err := m.clusterAdmin.DescribeConfig(sarama.ConfigResource{
				Type: sarama.BrokerResource,
				Name: strconv.Itoa(int(brokerID)),
			})
			if err == nil && len(configs) > 0 {
				for _, config := range configs {
					configInfo := make(map[string]interface{})
					for name, entry := range config.Config {
						// 只记录一些关键配置
						if strings.Contains(name, "log.retention") ||
							strings.Contains(name, "num.network.threads") ||
							strings.Contains(name, "num.io.threads") ||
							strings.Contains(name, "socket.send.buffer.bytes") ||
							strings.Contains(name, "socket.receive.buffer.bytes") {
							configInfo[name] = entry.Value
						}
					}
					result["broker_config"] = configInfo
					break
				}
			}
		}
	}

	// 计算集群健康度
	healthScore := 100.0
	if latency.Milliseconds() > 1000 {
		healthScore -= 20
	}
	if metadataRefreshTime.Milliseconds() > 5000 {
		healthScore -= 20
	}
	if err != nil {
		healthScore -= 30
	}

	result["cluster_health_score"] = healthScore
	result["collection_timestamp"] = time.Now().Unix()

	return result, nil
}

// 重写collectMetrics方法以使用Windows特定的实现
func (m *WindowsKafkaMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 初始化Kafka客户端
	if err := m.initKafkaClient(); err != nil {
		m.agent.Logger.Error("Failed to initialize Kafka client: %v", err)
		return m.KafkaMonitor.collectMetrics() // 回退到模拟数据
	}

	// 添加连接信息
	metrics["brokers"] = m.brokers
	metrics["collection_time"] = time.Now().Format(time.RFC3339)

	// 使用Windows特定的方法收集指标
	if brokerInfo, err := m.getBrokerInfo(); err == nil {
		for k, v := range brokerInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get broker info: %v", err)
	}

	if topicInfo, err := m.getTopicInfo(); err == nil {
		for k, v := range topicInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get topic info: %v", err)
	}

	if messageStats, err := m.getMessageStats(); err == nil {
		for k, v := range messageStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get message stats: %v", err)
	}

	if consumerGroupInfo, err := m.getConsumerGroupInfo(); err == nil {
		for k, v := range consumerGroupInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get consumer group info: %v", err)
	}

	if performanceMetrics, err := m.getPerformanceMetrics(); err == nil {
		for k, v := range performanceMetrics {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get performance metrics: %v", err)
	}

	// 如果所有API调用都失败，回退到模拟数据
	if len(metrics) <= 2 { // 只有基本连接信息
		m.agent.Logger.Warn("All Kafka API calls failed, falling back to simulated data")
		return m.KafkaMonitor.collectMetrics()
	}

	return metrics
}

// Close 关闭连接
func (m *WindowsKafkaMonitor) Close() error {
	var errs []error

	if m.clusterAdmin != nil {
		if err := m.clusterAdmin.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if m.client != nil {
		if err := m.client.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Kafka connections: %v", errs)
	}

	return nil
}