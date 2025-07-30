//go:build linux
// +build linux

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/streadway/amqp"
	"aimonitor-agents/common"
)

// LinuxRabbitMQMonitor Linux版本的RabbitMQ监控器
type LinuxRabbitMQMonitor struct {
	*RabbitMQMonitor
	conn       *amqp.Connection
	managementURL string
	username      string
	password      string
}

// NewLinuxRabbitMQMonitor 创建Linux版本的RabbitMQ监控器
func NewLinuxRabbitMQMonitor(agent *common.Agent) *LinuxRabbitMQMonitor {
	baseMonitor := NewRabbitMQMonitor(agent)
	return &LinuxRabbitMQMonitor{
		RabbitMQMonitor: baseMonitor,
		managementURL:   "http://localhost:15672",
		username:        "guest",
		password:        "guest",
	}
}

// initConnection 初始化RabbitMQ连接
func (m *LinuxRabbitMQMonitor) initConnection() error {
	// AMQP连接
	connStr := fmt.Sprintf("amqp://%s:%s@localhost:5672/", m.username, m.password)
	conn, err := amqp.Dial(connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	m.conn = conn
	return nil
}

// makeManagementAPIRequest 发送管理API请求
func (m *LinuxRabbitMQMonitor) makeManagementAPIRequest(endpoint string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", m.managementURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(m.username, m.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

// getNodeInfo 获取真实的节点信息
func (m *LinuxRabbitMQMonitor) getNodeInfo() (map[string]interface{}, error) {
	data, err := m.makeManagementAPIRequest("/api/nodes")
	if err != nil {
		return nil, fmt.Errorf("failed to get node info: %v", err)
	}

	var nodes []map[string]interface{}
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse node info: %v", err)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	node := nodes[0] // 获取第一个节点信息

	// 提取关键信息
	result := map[string]interface{}{
		"node_name":     node["name"],
		"node_type":     node["type"],
		"running":       node["running"],
		"uptime":        node["uptime"],
		"erlang_version": node["erlang_version"],
		"rabbitmq_version": node["rabbitmq_version"],
	}

	// 内存信息
	if memInfo, ok := node["mem_used"].(float64); ok {
		result["memory_used_bytes"] = int64(memInfo)
		result["memory_used_mb"] = int64(memInfo) / 1024 / 1024
	}

	if memLimit, ok := node["mem_limit"].(float64); ok {
		result["memory_limit_bytes"] = int64(memLimit)
		result["memory_limit_mb"] = int64(memLimit) / 1024 / 1024
		if memUsed, exists := result["memory_used_bytes"]; exists {
			usage := float64(memUsed.(int64)) / memLimit * 100
			result["memory_usage_percent"] = usage
		}
	}

	// 磁盘信息
	if diskFree, ok := node["disk_free"].(float64); ok {
		result["disk_free_bytes"] = int64(diskFree)
		result["disk_free_mb"] = int64(diskFree) / 1024 / 1024
	}

	if diskLimit, ok := node["disk_free_limit"].(float64); ok {
		result["disk_free_limit_bytes"] = int64(diskLimit)
		result["disk_free_limit_mb"] = int64(diskLimit) / 1024 / 1024
	}

	// 文件描述符
	if fdUsed, ok := node["fd_used"].(float64); ok {
		result["fd_used"] = int(fdUsed)
	}

	if fdTotal, ok := node["fd_total"].(float64); ok {
		result["fd_total"] = int(fdTotal)
		if fdUsed, exists := result["fd_used"]; exists {
			usage := float64(fdUsed.(int)) / fdTotal * 100
			result["fd_usage_percent"] = usage
		}
	}

	// Socket信息
	if socketsUsed, ok := node["sockets_used"].(float64); ok {
		result["sockets_used"] = int(socketsUsed)
	}

	if socketsTotal, ok := node["sockets_total"].(float64); ok {
		result["sockets_total"] = int(socketsTotal)
		if socketsUsed, exists := result["sockets_used"]; exists {
			usage := float64(socketsUsed.(int)) / socketsTotal * 100
			result["sockets_usage_percent"] = usage
		}
	}

	return result, nil
}

// getConnectionInfo 获取真实的连接信息
func (m *LinuxRabbitMQMonitor) getConnectionInfo() (map[string]interface{}, error) {
	data, err := m.makeManagementAPIRequest("/api/connections")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection info: %v", err)
	}

	var connections []map[string]interface{}
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil, fmt.Errorf("failed to parse connection info: %v", err)
	}

	totalConnections := len(connections)
	runningConnections := 0
	blockedConnections := 0
	totalChannels := 0

	for _, conn := range connections {
		if state, ok := conn["state"].(string); ok && state == "running" {
			runningConnections++
		}
		if blocked, ok := conn["blocked"].(bool); ok && blocked {
			blockedConnections++
		}
		if channels, ok := conn["channels"].(float64); ok {
			totalChannels += int(channels)
		}
	}

	return map[string]interface{}{
		"total_connections":   totalConnections,
		"running_connections": runningConnections,
		"blocked_connections": blockedConnections,
		"idle_connections":    totalConnections - runningConnections - blockedConnections,
		"total_channels":      totalChannels,
		"avg_channels_per_connection": float64(totalChannels) / float64(totalConnections),
	}, nil
}

// getQueueInfo 获取真实的队列信息
func (m *LinuxRabbitMQMonitor) getQueueInfo() (map[string]interface{}, error) {
	data, err := m.makeManagementAPIRequest("/api/queues")
	if err != nil {
		return nil, fmt.Errorf("failed to get queue info: %v", err)
	}

	var queues []map[string]interface{}
	if err := json.Unmarshal(data, &queues); err != nil {
		return nil, fmt.Errorf("failed to parse queue info: %v", err)
	}

	totalQueues := len(queues)
	totalMessages := 0
	readyMessages := 0
	unackedMessages := 0
	totalConsumers := 0
	largestQueueSize := 0
	largestQueueName := ""

	for _, queue := range queues {
		if messages, ok := queue["messages"].(float64); ok {
			msgCount := int(messages)
			totalMessages += msgCount
			if msgCount > largestQueueSize {
				largestQueueSize = msgCount
				if name, nameOk := queue["name"].(string); nameOk {
					largestQueueName = name
				}
			}
		}
		if ready, ok := queue["messages_ready"].(float64); ok {
			readyMessages += int(ready)
		}
		if unacked, ok := queue["messages_unacknowledged"].(float64); ok {
			unackedMessages += int(unacked)
		}
		if consumers, ok := queue["consumers"].(float64); ok {
			totalConsumers += int(consumers)
		}
	}

	return map[string]interface{}{
		"total_queues":        totalQueues,
		"total_messages":      totalMessages,
		"ready_messages":      readyMessages,
		"unacked_messages":    unackedMessages,
		"total_consumers":     totalConsumers,
		"largest_queue_size":  largestQueueSize,
		"largest_queue_name":  largestQueueName,
		"avg_messages_per_queue": float64(totalMessages) / float64(totalQueues),
		"avg_consumers_per_queue": float64(totalConsumers) / float64(totalQueues),
	}, nil
}

// getExchangeInfo 获取真实的交换器信息
func (m *LinuxRabbitMQMonitor) getExchangeInfo() (map[string]interface{}, error) {
	data, err := m.makeManagementAPIRequest("/api/exchanges")
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %v", err)
	}

	var exchanges []map[string]interface{}
	if err := json.Unmarshal(data, &exchanges); err != nil {
		return nil, fmt.Errorf("failed to parse exchange info: %v", err)
	}

	totalExchanges := len(exchanges)
	directExchanges := 0
	topicExchanges := 0
	fanoutExchanges := 0
	headersExchanges := 0
	customExchanges := 0

	for _, exchange := range exchanges {
		if exchangeType, ok := exchange["type"].(string); ok {
			switch exchangeType {
			case "direct":
				directExchanges++
			case "topic":
				topicExchanges++
			case "fanout":
				fanoutExchanges++
			case "headers":
				headersExchanges++
			default:
				customExchanges++
			}
		}
	}

	return map[string]interface{}{
		"total_exchanges":   totalExchanges,
		"direct_exchanges":  directExchanges,
		"topic_exchanges":   topicExchanges,
		"fanout_exchanges":  fanoutExchanges,
		"headers_exchanges": headersExchanges,
		"custom_exchanges":  customExchanges,
	}, nil
}

// getVhostInfo 获取真实的虚拟主机信息
func (m *LinuxRabbitMQMonitor) getVhostInfo() (map[string]interface{}, error) {
	data, err := m.makeManagementAPIRequest("/api/vhosts")
	if err != nil {
		return nil, fmt.Errorf("failed to get vhost info: %v", err)
	}

	var vhosts []map[string]interface{}
	if err := json.Unmarshal(data, &vhosts); err != nil {
		return nil, fmt.Errorf("failed to parse vhost info: %v", err)
	}

	totalVhosts := len(vhosts)
	vhostNames := make([]string, 0, totalVhosts)

	for _, vhost := range vhosts {
		if name, ok := vhost["name"].(string); ok {
			vhostNames = append(vhostNames, name)
		}
	}

	return map[string]interface{}{
		"total_vhosts": totalVhosts,
		"vhost_names":  vhostNames,
	}, nil
}

// getClusterInfo 获取真实的集群信息
func (m *LinuxRabbitMQMonitor) getClusterInfo() (map[string]interface{}, error) {
	// 获取集群名称
	data, err := m.makeManagementAPIRequest("/api/cluster-name")
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster name: %v", err)
	}

	var clusterName map[string]interface{}
	if err := json.Unmarshal(data, &clusterName); err != nil {
		return nil, fmt.Errorf("failed to parse cluster name: %v", err)
	}

	// 获取节点信息
	nodeData, err := m.makeManagementAPIRequest("/api/nodes")
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes info: %v", err)
	}

	var nodes []map[string]interface{}
	if err := json.Unmarshal(nodeData, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse nodes info: %v", err)
	}

	totalNodes := len(nodes)
	runningNodes := 0
	nodeNames := make([]string, 0, totalNodes)

	for _, node := range nodes {
		if running, ok := node["running"].(bool); ok && running {
			runningNodes++
		}
		if name, ok := node["name"].(string); ok {
			nodeNames = append(nodeNames, name)
		}
	}

	result := map[string]interface{}{
		"cluster_name":   clusterName["name"],
		"total_nodes":    totalNodes,
		"running_nodes":  runningNodes,
		"stopped_nodes":  totalNodes - runningNodes,
		"node_names":     nodeNames,
		"cluster_status": "running",
	}

	if runningNodes < totalNodes {
		result["cluster_status"] = "degraded"
	}
	if runningNodes == 0 {
		result["cluster_status"] = "stopped"
	}

	return result, nil
}

// getMessageStats 获取消息统计信息
func (m *LinuxRabbitMQMonitor) getMessageStats() (map[string]interface{}, error) {
	data, err := m.makeManagementAPIRequest("/api/overview")
	if err != nil {
		return nil, fmt.Errorf("failed to get overview: %v", err)
	}

	var overview map[string]interface{}
	if err := json.Unmarshal(data, &overview); err != nil {
		return nil, fmt.Errorf("failed to parse overview: %v", err)
	}

	result := make(map[string]interface{})

	// 消息统计
	if messageStats, ok := overview["message_stats"].(map[string]interface{}); ok {
		if publish, exists := messageStats["publish"]; exists {
			result["total_published"] = publish
		}
		if deliver, exists := messageStats["deliver_get"]; exists {
			result["total_delivered"] = deliver
		}
		if ack, exists := messageStats["ack"]; exists {
			result["total_acked"] = ack
		}
		if reject, exists := messageStats["reject"]; exists {
			result["total_rejected"] = reject
		}
	}

	// 队列总计
	if queueTotals, ok := overview["queue_totals"].(map[string]interface{}); ok {
		if messages, exists := queueTotals["messages"]; exists {
			result["total_messages"] = messages
		}
		if ready, exists := queueTotals["messages_ready"]; exists {
			result["messages_ready"] = ready
		}
		if unacked, exists := queueTotals["messages_unacknowledged"]; exists {
			result["messages_unacked"] = unacked
		}
	}

	// 对象总计
	if objectTotals, ok := overview["object_totals"].(map[string]interface{}); ok {
		if connections, exists := objectTotals["connections"]; exists {
			result["total_connections"] = connections
		}
		if channels, exists := objectTotals["channels"]; exists {
			result["total_channels"] = channels
		}
		if exchanges, exists := objectTotals["exchanges"]; exists {
			result["total_exchanges"] = exchanges
		}
		if queues, exists := objectTotals["queues"]; exists {
			result["total_queues"] = queues
		}
		if consumers, exists := objectTotals["consumers"]; exists {
			result["total_consumers"] = consumers
		}
	}

	return result, nil
}

// Close 关闭连接
func (m *LinuxRabbitMQMonitor) Close() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

// 重写collectMetrics方法以使用Linux特定的实现
func (m *LinuxRabbitMQMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 初始化连接（如果需要）
	if m.conn == nil {
		if err := m.initConnection(); err != nil {
			m.agent.Logger.Error("Failed to initialize RabbitMQ connection: %v", err)
			return m.RabbitMQMonitor.collectMetrics() // 回退到模拟数据
		}
	}

	// 使用Linux特定的方法收集指标
	if nodeInfo, err := m.getNodeInfo(); err == nil {
		for k, v := range nodeInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get node info: %v", err)
	}

	if connectionInfo, err := m.getConnectionInfo(); err == nil {
		for k, v := range connectionInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get connection info: %v", err)
	}

	if queueInfo, err := m.getQueueInfo(); err == nil {
		for k, v := range queueInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get queue info: %v", err)
	}

	if exchangeInfo, err := m.getExchangeInfo(); err == nil {
		for k, v := range exchangeInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get exchange info: %v", err)
	}

	if vhostInfo, err := m.getVhostInfo(); err == nil {
		for k, v := range vhostInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get vhost info: %v", err)
	}

	if clusterInfo, err := m.getClusterInfo(); err == nil {
		for k, v := range clusterInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get cluster info: %v", err)
	}

	if messageStats, err := m.getMessageStats(); err == nil {
		for k, v := range messageStats {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get message stats: %v", err)
	}

	return metrics
}