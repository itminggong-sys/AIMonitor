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
	agent.Info.Type = "rabbitmq"
	agent.Info.Name = "RabbitMQ Monitor"

	// 创建RabbitMQ监控器
	monitor := NewRabbitMQMonitor(agent)

	// 启动Agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 启动监控
	go monitor.StartMonitoring()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("RabbitMQ Agent started. Press Ctrl+C to stop.")
	<-sigChan

	// 停止Agent
	agent.Stop()
	fmt.Println("RabbitMQ Agent stopped.")
}

// RabbitMQMonitor RabbitMQ监控器
type RabbitMQMonitor struct {
	agent *common.Agent
}

// NewRabbitMQMonitor 创建RabbitMQ监控器
func NewRabbitMQMonitor(agent *common.Agent) *RabbitMQMonitor {
	return &RabbitMQMonitor{
		agent: agent,
	}
}

// StartMonitoring 开始监控
func (m *RabbitMQMonitor) StartMonitoring() {
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

// collectMetrics 收集RabbitMQ指标
func (m *RabbitMQMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 节点信息
	nodeInfo, err := m.getNodeInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get RabbitMQ node info: %v", err)
	} else {
		metrics["node_running"] = nodeInfo["node_running"]
		metrics["memory_used"] = nodeInfo["memory_used"]
		metrics["memory_limit"] = nodeInfo["memory_limit"]
		metrics["disk_free"] = nodeInfo["disk_free"]
		metrics["disk_free_limit"] = nodeInfo["disk_free_limit"]
		metrics["fd_used"] = nodeInfo["fd_used"]
		metrics["fd_total"] = nodeInfo["fd_total"]
		metrics["sockets_used"] = nodeInfo["sockets_used"]
		metrics["sockets_total"] = nodeInfo["sockets_total"]
	}

	// 连接信息
	connectionInfo, err := m.getConnectionInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get RabbitMQ connection info: %v", err)
	} else {
		metrics["connections_total"] = connectionInfo["connections_total"]
		metrics["channels_total"] = connectionInfo["channels_total"]
		metrics["consumers_total"] = connectionInfo["consumers_total"]
	}

	// 队列信息
	queueInfo, err := m.getQueueInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get RabbitMQ queue info: %v", err)
	} else {
		metrics["queues_total"] = queueInfo["queues_total"]
		metrics["messages_total"] = queueInfo["messages_total"]
		metrics["messages_ready"] = queueInfo["messages_ready"]
		metrics["messages_unacknowledged"] = queueInfo["messages_unacknowledged"]
		metrics["message_publish_rate"] = queueInfo["message_publish_rate"]
		metrics["message_deliver_rate"] = queueInfo["message_deliver_rate"]
		metrics["message_ack_rate"] = queueInfo["message_ack_rate"]
	}

	// 交换器信息
	exchangeInfo, err := m.getExchangeInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get RabbitMQ exchange info: %v", err)
	} else {
		metrics["exchanges_total"] = exchangeInfo["exchanges_total"]
		metrics["exchange_publish_in_rate"] = exchangeInfo["exchange_publish_in_rate"]
		metrics["exchange_publish_out_rate"] = exchangeInfo["exchange_publish_out_rate"]
	}

	// 虚拟主机信息
	vhostInfo, err := m.getVhostInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get RabbitMQ vhost info: %v", err)
	} else {
		metrics["vhosts_total"] = vhostInfo["vhosts_total"]
		metrics["vhost_messages_total"] = vhostInfo["vhost_messages_total"]
	}

	// 集群信息
	clusterInfo, err := m.getClusterInfo()
	if err != nil {
		m.agent.Logger.Error("Failed to get RabbitMQ cluster info: %v", err)
	} else {
		metrics["cluster_nodes_total"] = clusterInfo["cluster_nodes_total"]
		metrics["cluster_nodes_running"] = clusterInfo["cluster_nodes_running"]
		metrics["cluster_partitions"] = clusterInfo["cluster_partitions"]
	}

	return metrics
}

// getNodeInfo 获取节点信息
func (m *RabbitMQMonitor) getNodeInfo() (map[string]interface{}, error) {
	// 这里应该调用RabbitMQ Management API获取节点信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"node_running":     true,
		"memory_used":      104857600,  // 100MB
		"memory_limit":     1073741824, // 1GB
		"disk_free":        10737418240, // 10GB
		"disk_free_limit": 1073741824,  // 1GB
		"fd_used":          100,
		"fd_total":         1024,
		"sockets_used":     50,
		"sockets_total":    829,
	}, nil
}

// getConnectionInfo 获取连接信息
func (m *RabbitMQMonitor) getConnectionInfo() (map[string]interface{}, error) {
	// 这里应该调用RabbitMQ Management API获取连接信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"connections_total": 25,
		"channels_total":    50,
		"consumers_total":   30,
	}, nil
}

// getQueueInfo 获取队列信息
func (m *RabbitMQMonitor) getQueueInfo() (map[string]interface{}, error) {
	// 这里应该调用RabbitMQ Management API获取队列信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"queues_total":              20,
		"messages_total":            1000,
		"messages_ready":            800,
		"messages_unacknowledged": 200,
		"message_publish_rate":      10.5,
		"message_deliver_rate":      9.8,
		"message_ack_rate":          9.5,
	}, nil
}

// getExchangeInfo 获取交换器信息
func (m *RabbitMQMonitor) getExchangeInfo() (map[string]interface{}, error) {
	// 这里应该调用RabbitMQ Management API获取交换器信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"exchanges_total":         15,
		"exchange_publish_in_rate": 12.3,
		"exchange_publish_out_rate": 11.8,
	}, nil
}

// getVhostInfo 获取虚拟主机信息
func (m *RabbitMQMonitor) getVhostInfo() (map[string]interface{}, error) {
	// 这里应该调用RabbitMQ Management API获取虚拟主机信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"vhosts_total":         3,
		"vhost_messages_total": 1500,
	}, nil
}

// getClusterInfo 获取集群信息
func (m *RabbitMQMonitor) getClusterInfo() (map[string]interface{}, error) {
	// 这里应该调用RabbitMQ Management API获取集群信息
	// 为了简化，这里返回模拟数据
	return map[string]interface{}{
		"cluster_nodes_total":   3,
		"cluster_nodes_running": 3,
		"cluster_partitions":    0,
	}, nil
}