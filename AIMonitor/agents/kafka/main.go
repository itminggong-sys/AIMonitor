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
	agent.Info.Type = "kafka"
	agent.Info.Name = "Kafka Monitor"

	// 创建Kafka监控器
	monitor := NewKafkaMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Kafka Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("Kafka Agent stopped.")
}

// KafkaMonitor Kafka监控器
type KafkaMonitor struct {
	agent *common.Agent
}

// NewKafkaMonitor 创建Kafka监控器
func NewKafkaMonitor(agent *common.Agent) *KafkaMonitor {
	return &KafkaMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *KafkaMonitor) StartMonitoring() {
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

// collectMetrics 收集Kafka指标
func (m *KafkaMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Broker信息
	brokerInfo, err := m.getBrokerInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Kafka broker info: %v", err)
	} else {
		metrics["broker_count"] = brokerInfo["broker_count"]
		metrics["controller_id"] = brokerInfo["controller_id"]
		metrics["active_brokers"] = brokerInfo["active_brokers"]
	}

	// Topic信息
	topicInfo, err := m.getTopicInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Kafka topic info: %v", err)
	} else {
		metrics["topic_count"] = topicInfo["topic_count"]
		metrics["partition_count"] = topicInfo["partition_count"]
		metrics["under_replicated_partitions"] = topicInfo["under_replicated_partitions"]
		metrics["offline_partitions"] = topicInfo["offline_partitions"]
	}

	// 消息统计
	messageStats, err := m.getMessageStats()
	if err != nil {
		m.agent.Logger.Error("Failed to get Kafka message stats: %v", err)
	} else {
		metrics["messages_in_per_sec"] = messageStats["messages_in_per_sec"]
		metrics["bytes_in_per_sec"] = messageStats["bytes_in_per_sec"]
		metrics["bytes_out_per_sec"] = messageStats["bytes_out_per_sec"]
		metrics["total_produce_requests"] = messageStats["total_produce_requests"]
		metrics["total_fetch_requests"] = messageStats["total_fetch_requests"]
	}

	// Consumer Group信息
	consumerInfo, err := m.getConsumerInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Kafka consumer info: %v", err)
	} else {
		metrics["consumer_groups"] = consumerInfo["consumer_groups"]
		metrics["active_consumers"] = consumerInfo["active_consumers"]
		metrics["consumer_lag"] = consumerInfo["consumer_lag"]
	}

	// 性能指标
	performanceInfo, err := m.getPerformanceInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get Kafka performance info: %v", err)
	} else {
		metrics["request_handler_avg_idle"] = performanceInfo["request_handler_avg_idle"]
		metrics["network_processor_avg_idle"] = performanceInfo["network_processor_avg_idle"]
		metrics["produce_request_time_avg"] = performanceInfo["produce_request_time_avg"]
		metrics["fetch_request_time_avg"] = performanceInfo["fetch_request_time_avg"]
	}

	return metrics
}

// getBrokerInfo 获取Broker信息
func (m *KafkaMonitor) getBrokerInfo() (map[string]interface{}, error) {
	// 这里应该连接Kafka并获取Broker信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"broker_count":   3,
		"controller_id":  1,
		"active_brokers": 3,
	}, nil
}

// getTopicInfo 获取Topic信息
func (m *KafkaMonitor) getTopicInfo() (map[string]interface{}, error) {
	// 这里应该连接Kafka并获取Topic信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"topic_count":                  50,
		"partition_count":              150,
		"under_replicated_partitions": 0,
		"offline_partitions":          0,
	}, nil
}

// getMessageStats 获取消息统计
func (m *KafkaMonitor) getMessageStats() (map[string]interface{}, error) {
	// 这里应该连接Kafka并获取消息统计
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"messages_in_per_sec":    1000.0,
		"bytes_in_per_sec":       1048576.0, // 1MB/s
		"bytes_out_per_sec":      2097152.0, // 2MB/s
		"total_produce_requests": 50000,
		"total_fetch_requests":   75000,
	}, nil
}

// getConsumerInfo 获取Consumer信息
func (m *KafkaMonitor) getConsumerInfo() (map[string]interface{}, error) {
	// 这里应该连接Kafka并获取Consumer信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"consumer_groups":  10,
		"active_consumers": 25,
		"consumer_lag":     100,
	}, nil
}

// getPerformanceInfo 获取性能信息
func (m *KafkaMonitor) getPerformanceInfo() (map[string]interface{}, error) {
	// 这里应该连接Kafka并获取性能信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"request_handler_avg_idle":     0.8,
		"network_processor_avg_idle":   0.9,
		"produce_request_time_avg":     5.2,
		"fetch_request_time_avg":       3.8,
	}, nil
}