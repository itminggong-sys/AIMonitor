//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aimonitor-agents/common"
)

// WindowsRabbitMQMonitor Windows版本的RabbitMQ监控器
type WindowsRabbitMQMonitor struct {
	*RabbitMQMonitor
	client   *http.Client
	baseURL  string
	username string
	password string
	host     string
	port     int
	managementPort int
}

// NewWindowsRabbitMQMonitor 创建Windows版本的RabbitMQ监控器
func NewWindowsRabbitMQMonitor(agent *common.Agent) *WindowsRabbitMQMonitor {
	baseMonitor := NewRabbitMQMonitor(agent)
	return &WindowsRabbitMQMonitor{
		RabbitMQMonitor: baseMonitor,
		client:          &http.Client{Timeout: 10 * time.Second},
		host:            "localhost",
		port:            5672,
		managementPort:  15672,
		username:        "guest",
		password:        "guest",
	}
}

// initRabbitMQConnection 初始化RabbitMQ连接
func (m *WindowsRabbitMQMonitor) initRabbitMQConnection() error {
	m.baseURL = fmt.Sprintf("http://%s:%d/api", m.host, m.managementPort)
	
	// 测试连接
	_, err := m.makeRequest("/overview")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ management API: %v", err)
	}
	
	return nil
}

// makeRequest 发送HTTP请求到RabbitMQ管理API
func (m *WindowsRabbitMQMonitor) makeRequest(endpoint string) ([]byte, error) {
	url := m.baseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.SetBasicAuth(m.username, m.password)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	return ioutil.ReadAll(resp.Body)
}

// getNodeInfo 获取真实的节点信息
func (m *WindowsRabbitMQMonitor) getNodeInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取节点列表
	data, err := m.makeRequest("/nodes")
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %v", err)
	}
	
	var nodes []map[string]interface{}
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse nodes response: %v", err)
	}
	
	result["total_nodes"] = len(nodes)
	
	nodeDetails := make([]map[string]interface{}, 0)
	runningNodes := 0
	totalMemory := int64(0)
	totalDisk := int64(0)
	
	for _, node := range nodes {
		nodeInfo := map[string]interface{}{
			"name": node["name"],
			"type": node["type"],
		}
		
		if running, ok := node["running"].(bool); ok {
			nodeInfo["running"] = running
			if running {
				runningNodes++
			}
		}
		
		if memUsed, ok := node["mem_used"].(float64); ok {
			nodeInfo["memory_used"] = int64(memUsed)
			totalMemory += int64(memUsed)
		}
		
		if memLimit, ok := node["mem_limit"].(float64); ok {
			nodeInfo["memory_limit"] = int64(memLimit)
			if memUsed, ok := node["mem_used"].(float64); ok {
				nodeInfo["memory_usage_percent"] = (memUsed / memLimit) * 100
			}
		}
		
		if diskFree, ok := node["disk_free"].(float64); ok {
			nodeInfo["disk_free"] = int64(diskFree)
			totalDisk += int64(diskFree)
		}
		
		if diskFreeLimit, ok := node["disk_free_limit"].(float64); ok {
			nodeInfo["disk_free_limit"] = int64(diskFreeLimit)
		}
		
		if fdUsed, ok := node["fd_used"].(float64); ok {
			nodeInfo["file_descriptors_used"] = int64(fdUsed)
		}
		
		if fdTotal, ok := node["fd_total"].(float64); ok {
			nodeInfo["file_descriptors_total"] = int64(fdTotal)
			if fdUsed, ok := node["fd_used"].(float64); ok {
				nodeInfo["file_descriptors_usage_percent"] = (fdUsed / fdTotal) * 100
			}
		}
		
		if socketsUsed, ok := node["sockets_used"].(float64); ok {
			nodeInfo["sockets_used"] = int64(socketsUsed)
		}
		
		if socketsTotal, ok := node["sockets_total"].(float64); ok {
			nodeInfo["sockets_total"] = int64(socketsTotal)
			if socketsUsed, ok := node["sockets_used"].(float64); ok {
				nodeInfo["sockets_usage_percent"] = (socketsUsed / socketsTotal) * 100
			}
		}
		
		if procUsed, ok := node["proc_used"].(float64); ok {
			nodeInfo["processes_used"] = int64(procUsed)
		}
		
		if procTotal, ok := node["proc_total"].(float64); ok {
			nodeInfo["processes_total"] = int64(procTotal)
			if procUsed, ok := node["proc_used"].(float64); ok {
				nodeInfo["processes_usage_percent"] = (procUsed / procTotal) * 100
			}
		}
		
		if uptime, ok := node["uptime"].(float64); ok {
			nodeInfo["uptime_milliseconds"] = int64(uptime)
			nodeInfo["uptime_seconds"] = int64(uptime) / 1000
		}
		
		nodeDetails = append(nodeDetails, nodeInfo)
	}
	
	result["node_details"] = nodeDetails
	result["running_nodes"] = runningNodes
	result["total_memory_used"] = totalMemory
	result["total_disk_free"] = totalDisk
	
	return result, nil
}

// getConnectionInfo 获取真实的连接信息
func (m *WindowsRabbitMQMonitor) getConnectionInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取连接列表
	data, err := m.makeRequest("/connections")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %v", err)
	}
	
	var connections []map[string]interface{}
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil, fmt.Errorf("failed to parse connections response: %v", err)
	}
	
	result["total_connections"] = len(connections)
	
	connectionDetails := make([]map[string]interface{}, 0)
	runningConnections := 0
	blockedConnections := 0
	clientConnections := make(map[string]int)
	userConnections := make(map[string]int)
	
	for _, conn := range connections {
		connInfo := map[string]interface{}{
			"name": conn["name"],
		}
		
		if state, ok := conn["state"].(string); ok {
			connInfo["state"] = state
			if state == "running" {
				runningConnections++
			} else if state == "blocked" {
				blockedConnections++
			}
		}
		
		if user, ok := conn["user"].(string); ok {
			connInfo["user"] = user
			userConnections[user]++
		}
		
		if vhost, ok := conn["vhost"].(string); ok {
			connInfo["vhost"] = vhost
		}
		
		if peerHost, ok := conn["peer_host"].(string); ok {
			connInfo["peer_host"] = peerHost
			clientConnections[peerHost]++
		}
		
		if peerPort, ok := conn["peer_port"].(float64); ok {
			connInfo["peer_port"] = int(peerPort)
		}
		
		if protocol, ok := conn["protocol"].(string); ok {
			connInfo["protocol"] = protocol
		}
		
		if channels, ok := conn["channels"].(float64); ok {
			connInfo["channels"] = int(channels)
		}
		
		if sendPending, ok := conn["send_pend"].(float64); ok {
			connInfo["send_pending"] = int64(sendPending)
		}
		
		if recvOct, ok := conn["recv_oct"].(float64); ok {
			connInfo["bytes_received"] = int64(recvOct)
		}
		
		if sendOct, ok := conn["send_oct"].(float64); ok {
			connInfo["bytes_sent"] = int64(sendOct)
		}
		
		if connectedAt, ok := conn["connected_at"].(float64); ok {
			connInfo["connected_at"] = time.Unix(int64(connectedAt)/1000, 0).Format(time.RFC3339)
		}
		
		connectionDetails = append(connectionDetails, connInfo)
	}
	
	result["connection_details"] = connectionDetails
	result["running_connections"] = runningConnections
	result["blocked_connections"] = blockedConnections
	result["client_connections"] = clientConnections
	result["user_connections"] = userConnections
	
	return result, nil
}

// getQueueInfo 获取真实的队列信息
func (m *WindowsRabbitMQMonitor) getQueueInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取队列列表
	data, err := m.makeRequest("/queues")
	if err != nil {
		return nil, fmt.Errorf("failed to get queues: %v", err)
	}
	
	var queues []map[string]interface{}
	if err := json.Unmarshal(data, &queues); err != nil {
		return nil, fmt.Errorf("failed to parse queues response: %v", err)
	}
	
	result["total_queues"] = len(queues)
	
	queueDetails := make([]map[string]interface{}, 0)
	totalMessages := int64(0)
	totalConsumers := int64(0)
	totalMemory := int64(0)
	queueStates := make(map[string]int)
	vhostQueues := make(map[string]int)
	
	for _, queue := range queues {
		queueInfo := map[string]interface{}{
			"name": queue["name"],
		}
		
		if vhost, ok := queue["vhost"].(string); ok {
			queueInfo["vhost"] = vhost
			vhostQueues[vhost]++
		}
		
		if state, ok := queue["state"].(string); ok {
			queueInfo["state"] = state
			queueStates[state]++
		}
		
		if durable, ok := queue["durable"].(bool); ok {
			queueInfo["durable"] = durable
		}
		
		if autoDelete, ok := queue["auto_delete"].(bool); ok {
			queueInfo["auto_delete"] = autoDelete
		}
		
		if messages, ok := queue["messages"].(float64); ok {
			queueInfo["messages"] = int64(messages)
			totalMessages += int64(messages)
		}
		
		if messagesReady, ok := queue["messages_ready"].(float64); ok {
			queueInfo["messages_ready"] = int64(messagesReady)
		}
		
		if messagesUnack, ok := queue["messages_unacknowledged"].(float64); ok {
			queueInfo["messages_unacknowledged"] = int64(messagesUnack)
		}
		
		if consumers, ok := queue["consumers"].(float64); ok {
			queueInfo["consumers"] = int64(consumers)
			totalConsumers += int64(consumers)
		}
		
		if memory, ok := queue["memory"].(float64); ok {
			queueInfo["memory"] = int64(memory)
			totalMemory += int64(memory)
		}
		
		// 消息统计
		if msgStats, ok := queue["message_stats"].(map[string]interface{}); ok {
			if publishTotal, ok := msgStats["publish"].(float64); ok {
				queueInfo["messages_published_total"] = int64(publishTotal)
			}
			if deliverTotal, ok := msgStats["deliver_get"].(float64); ok {
				queueInfo["messages_delivered_total"] = int64(deliverTotal)
			}
			if ackTotal, ok := msgStats["ack"].(float64); ok {
				queueInfo["messages_acknowledged_total"] = int64(ackTotal)
			}
			if redeliverTotal, ok := msgStats["redeliver"].(float64); ok {
				queueInfo["messages_redelivered_total"] = int64(redeliverTotal)
			}
		}
		
		// 消息速率
		if msgStatsDetails, ok := queue["message_stats"].(map[string]interface{}); ok {
			if publishDetails, ok := msgStatsDetails["publish_details"].(map[string]interface{}); ok {
				if rate, ok := publishDetails["rate"].(float64); ok {
					queueInfo["publish_rate"] = rate
				}
			}
			if deliverDetails, ok := msgStatsDetails["deliver_get_details"].(map[string]interface{}); ok {
				if rate, ok := deliverDetails["rate"].(float64); ok {
					queueInfo["deliver_rate"] = rate
				}
			}
		}
		
		queueDetails = append(queueDetails, queueInfo)
	}
	
	result["queue_details"] = queueDetails
	result["total_messages"] = totalMessages
	result["total_consumers"] = totalConsumers
	result["total_queue_memory"] = totalMemory
	result["queue_states"] = queueStates
	result["vhost_queues"] = vhostQueues
	
	return result, nil
}

// getExchangeInfo 获取真实的交换器信息
func (m *WindowsRabbitMQMonitor) getExchangeInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取交换器列表
	data, err := m.makeRequest("/exchanges")
	if err != nil {
		return nil, fmt.Errorf("failed to get exchanges: %v", err)
	}
	
	var exchanges []map[string]interface{}
	if err := json.Unmarshal(data, &exchanges); err != nil {
		return nil, fmt.Errorf("failed to parse exchanges response: %v", err)
	}
	
	result["total_exchanges"] = len(exchanges)
	
	exchangeDetails := make([]map[string]interface{}, 0)
	exchangeTypes := make(map[string]int)
	vhostExchanges := make(map[string]int)
	
	for _, exchange := range exchanges {
		exchangeInfo := map[string]interface{}{
			"name": exchange["name"],
		}
		
		if vhost, ok := exchange["vhost"].(string); ok {
			exchangeInfo["vhost"] = vhost
			vhostExchanges[vhost]++
		}
		
		if exchangeType, ok := exchange["type"].(string); ok {
			exchangeInfo["type"] = exchangeType
			exchangeTypes[exchangeType]++
		}
		
		if durable, ok := exchange["durable"].(bool); ok {
			exchangeInfo["durable"] = durable
		}
		
		if autoDelete, ok := exchange["auto_delete"].(bool); ok {
			exchangeInfo["auto_delete"] = autoDelete
		}
		
		if internal, ok := exchange["internal"].(bool); ok {
			exchangeInfo["internal"] = internal
		}
		
		// 消息统计
		if msgStats, ok := exchange["message_stats"].(map[string]interface{}); ok {
			if publishIn, ok := msgStats["publish_in"].(float64); ok {
				exchangeInfo["messages_published_in"] = int64(publishIn)
			}
			if publishOut, ok := msgStats["publish_out"].(float64); ok {
				exchangeInfo["messages_published_out"] = int64(publishOut)
			}
		}
		
		// 消息速率
		if msgStatsDetails, ok := exchange["message_stats"].(map[string]interface{}); ok {
			if publishInDetails, ok := msgStatsDetails["publish_in_details"].(map[string]interface{}); ok {
				if rate, ok := publishInDetails["rate"].(float64); ok {
					exchangeInfo["publish_in_rate"] = rate
				}
			}
			if publishOutDetails, ok := msgStatsDetails["publish_out_details"].(map[string]interface{}); ok {
				if rate, ok := publishOutDetails["rate"].(float64); ok {
					exchangeInfo["publish_out_rate"] = rate
				}
			}
		}
		
		exchangeDetails = append(exchangeDetails, exchangeInfo)
	}
	
	result["exchange_details"] = exchangeDetails
	result["exchange_types"] = exchangeTypes
	result["vhost_exchanges"] = vhostExchanges
	
	return result, nil
}

// getVirtualHostInfo 获取真实的虚拟主机信息
func (m *WindowsRabbitMQMonitor) getVirtualHostInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取虚拟主机列表
	data, err := m.makeRequest("/vhosts")
	if err != nil {
		return nil, fmt.Errorf("failed to get vhosts: %v", err)
	}
	
	var vhosts []map[string]interface{}
	if err := json.Unmarshal(data, &vhosts); err != nil {
		return nil, fmt.Errorf("failed to parse vhosts response: %v", err)
	}
	
	result["total_vhosts"] = len(vhosts)
	
	vhostDetails := make([]map[string]interface{}, 0)
	totalMessages := int64(0)
	
	for _, vhost := range vhosts {
		vhostInfo := map[string]interface{}{
			"name": vhost["name"],
		}
		
		if tracing, ok := vhost["tracing"].(bool); ok {
			vhostInfo["tracing"] = tracing
		}
		
		// 消息统计
		if msgStats, ok := vhost["message_stats"].(map[string]interface{}); ok {
			if publish, ok := msgStats["publish"].(float64); ok {
				vhostInfo["messages_published"] = int64(publish)
				totalMessages += int64(publish)
			}
			if deliver, ok := msgStats["deliver_get"].(float64); ok {
				vhostInfo["messages_delivered"] = int64(deliver)
			}
			if ack, ok := msgStats["ack"].(float64); ok {
				vhostInfo["messages_acknowledged"] = int64(ack)
			}
		}
		
		// 消息速率
		if msgStatsDetails, ok := vhost["message_stats"].(map[string]interface{}); ok {
			if publishDetails, ok := msgStatsDetails["publish_details"].(map[string]interface{}); ok {
				if rate, ok := publishDetails["rate"].(float64); ok {
					vhostInfo["publish_rate"] = rate
				}
			}
			if deliverDetails, ok := msgStatsDetails["deliver_get_details"].(map[string]interface{}); ok {
				if rate, ok := deliverDetails["rate"].(float64); ok {
					vhostInfo["deliver_rate"] = rate
				}
			}
		}
		
		vhostDetails = append(vhostDetails, vhostInfo)
	}
	
	result["vhost_details"] = vhostDetails
	result["total_messages_published"] = totalMessages
	
	return result, nil
}

// getClusterInfo 获取真实的集群信息
func (m *WindowsRabbitMQMonitor) getClusterInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取集群名称
	data, err := m.makeRequest("/cluster-name")
	if err == nil {
		var clusterName map[string]interface{}
		if err := json.Unmarshal(data, &clusterName); err == nil {
			if name, ok := clusterName["name"].(string); ok {
				result["cluster_name"] = name
			}
		}
	}
	
	// 获取概览信息
	overviewData, err := m.makeRequest("/overview")
	if err == nil {
		var overview map[string]interface{}
		if err := json.Unmarshal(overviewData, &overview); err == nil {
			if managementVersion, ok := overview["management_version"].(string); ok {
				result["management_version"] = managementVersion
			}
			if rabbitmqVersion, ok := overview["rabbitmq_version"].(string); ok {
				result["rabbitmq_version"] = rabbitmqVersion
			}
			if erlangVersion, ok := overview["erlang_version"].(string); ok {
				result["erlang_version"] = erlangVersion
			}
			
			// 对象计数
			if objectTotals, ok := overview["object_totals"].(map[string]interface{}); ok {
				if connections, ok := objectTotals["connections"].(float64); ok {
					result["total_connections"] = int64(connections)
				}
				if channels, ok := objectTotals["channels"].(float64); ok {
					result["total_channels"] = int64(channels)
				}
				if exchanges, ok := objectTotals["exchanges"].(float64); ok {
					result["total_exchanges"] = int64(exchanges)
				}
				if queues, ok := objectTotals["queues"].(float64); ok {
					result["total_queues"] = int64(queues)
				}
				if consumers, ok := objectTotals["consumers"].(float64); ok {
					result["total_consumers"] = int64(consumers)
				}
			}
			
			// 队列总计
			if queueTotals, ok := overview["queue_totals"].(map[string]interface{}); ok {
				if messages, ok := queueTotals["messages"].(float64); ok {
					result["total_messages"] = int64(messages)
				}
				if messagesReady, ok := queueTotals["messages_ready"].(float64); ok {
					result["total_messages_ready"] = int64(messagesReady)
				}
				if messagesUnack, ok := queueTotals["messages_unacknowledged"].(float64); ok {
					result["total_messages_unacknowledged"] = int64(messagesUnack)
				}
			}
			
			// 消息统计
			if msgStats, ok := overview["message_stats"].(map[string]interface{}); ok {
				if publish, ok := msgStats["publish"].(float64); ok {
					result["messages_published_total"] = int64(publish)
				}
				if deliver, ok := msgStats["deliver_get"].(float64); ok {
					result["messages_delivered_total"] = int64(deliver)
				}
				if ack, ok := msgStats["ack"].(float64); ok {
					result["messages_acknowledged_total"] = int64(ack)
				}
				if redeliver, ok := msgStats["redeliver"].(float64); ok {
					result["messages_redelivered_total"] = int64(redeliver)
				}
			}
			
			// 消息速率
			if msgStatsDetails, ok := overview["message_stats"].(map[string]interface{}); ok {
				if publishDetails, ok := msgStatsDetails["publish_details"].(map[string]interface{}); ok {
					if rate, ok := publishDetails["rate"].(float64); ok {
						result["publish_rate"] = rate
					}
				}
				if deliverDetails, ok := msgStatsDetails["deliver_get_details"].(map[string]interface{}); ok {
					if rate, ok := deliverDetails["rate"].(float64); ok {
						result["deliver_rate"] = rate
					}
				}
			}
		}
	}
	
	return result, nil
}

// getMessageStats 获取真实的消息统计
func (m *WindowsRabbitMQMonitor) getMessageStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 获取通道统计
	data, err := m.makeRequest("/channels")
	if err == nil {
		var channels []map[string]interface{}
		if err := json.Unmarshal(data, &channels); err == nil {
			result["total_channels"] = len(channels)
			
			channelDetails := make([]map[string]interface{}, 0)
			totalUncommitted := int64(0)
			totalUnconfirmed := int64(0)
			totalUnacknowledged := int64(0)
			
			for _, channel := range channels {
				channelInfo := map[string]interface{}{
					"name": channel["name"],
				}
				
				if user, ok := channel["user"].(string); ok {
					channelInfo["user"] = user
				}
				
				if vhost, ok := channel["vhost"].(string); ok {
					channelInfo["vhost"] = vhost
				}
				
				if state, ok := channel["state"].(string); ok {
					channelInfo["state"] = state
				}
				
				if number, ok := channel["number"].(float64); ok {
					channelInfo["number"] = int(number)
				}
				
				if consumerCount, ok := channel["consumer_count"].(float64); ok {
					channelInfo["consumer_count"] = int64(consumerCount)
				}
				
				if messagesUncommitted, ok := channel["messages_uncommitted"].(float64); ok {
					channelInfo["messages_uncommitted"] = int64(messagesUncommitted)
					totalUncommitted += int64(messagesUncommitted)
				}
				
				if messagesUnconfirmed, ok := channel["messages_unconfirmed"].(float64); ok {
					channelInfo["messages_unconfirmed"] = int64(messagesUnconfirmed)
					totalUnconfirmed += int64(messagesUnconfirmed)
				}
				
				if messagesUnacknowledged, ok := channel["messages_unacknowledged"].(float64); ok {
					channelInfo["messages_unacknowledged"] = int64(messagesUnacknowledged)
					totalUnacknowledged += int64(messagesUnacknowledged)
				}
				
				// 消息统计
				if msgStats, ok := channel["message_stats"].(map[string]interface{}); ok {
					if publish, ok := msgStats["publish"].(float64); ok {
						channelInfo["messages_published"] = int64(publish)
					}
					if deliver, ok := msgStats["deliver_get"].(float64); ok {
						channelInfo["messages_delivered"] = int64(deliver)
					}
					if ack, ok := msgStats["ack"].(float64); ok {
						channelInfo["messages_acknowledged"] = int64(ack)
					}
				}
				
				channelDetails = append(channelDetails, channelInfo)
			}
			
			result["channel_details"] = channelDetails
			result["total_uncommitted_messages"] = totalUncommitted
			result["total_unconfirmed_messages"] = totalUnconfirmed
			result["total_unacknowledged_messages"] = totalUnacknowledged
		}
	}
	
	return result, nil
}

// 重写collectMetrics方法以使用Windows特定的实现
func (m *WindowsRabbitMQMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// 初始化RabbitMQ连接
	if err := m.initRabbitMQConnection(); err != nil {
		m.agent.Logger.Error("Failed to initialize RabbitMQ connection: %v", err)
		return m.RabbitMQMonitor.collectMetrics() // 回退到模拟数据
	}
	
	// 添加连接信息
	metrics["host"] = m.host
	metrics["port"] = m.port
	metrics["management_port"] = m.managementPort
	metrics["username"] = m.username
	metrics["collection_time"] = time.Now().Format(time.RFC3339)
	
	// 使用Windows特定的方法收集指标
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
	
	if vhostInfo, err := m.getVirtualHostInfo(); err == nil {
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
	
	// 如果所有API调用都失败，回退到模拟数据
	if len(metrics) <= 5 { // 只有基本连接信息
		m.agent.Logger.Warn("All RabbitMQ API calls failed, falling back to simulated data")
		return m.RabbitMQMonitor.collectMetrics()
	}
	
	return metrics
}