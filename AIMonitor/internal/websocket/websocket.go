package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"ai-monitor/internal/auth"
	"ai-monitor/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketManager WebSocket管理器
type WebSocketManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
	jwtManager *auth.JWTManager
}

// Client WebSocket客户端
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	userID   uuid.UUID
	manager  *WebSocketManager
	subscriptions map[string]bool // 订阅的主题
	mutex    sync.RWMutex
}

// Message WebSocket消息
type Message struct {
	Type      string      `json:"type"`
	Topic     string      `json:"topic,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	ID        string      `json:"id,omitempty"`
}

// AlertMessage 告警消息
type AlertMessage struct {
	Alert     *models.Alert     `json:"alert"`
	Rule      *models.AlertRule `json:"rule"`
	Severity  string            `json:"severity"`
	Message   string            `json:"message"`
}

// MetricMessage 指标消息
type MetricMessage struct {
	TargetID   string                 `json:"target_id"`
	TargetName string                 `json:"target_name"`
	Metrics    map[string]interface{} `json:"metrics"`
	Timestamp  time.Time              `json:"timestamp"`
}

// SystemMessage 系统消息
type SystemMessage struct {
	Level   string `json:"level"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// SubscribeMessage 订阅消息
type SubscribeMessage struct {
	Action string   `json:"action"` // subscribe, unsubscribe
	Topics []string `json:"topics"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 在生产环境中应该检查Origin
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// NewWebSocketManager 创建WebSocket管理器
func NewWebSocketManager(jwtManager *auth.JWTManager) *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		jwtManager: jwtManager,
	}
}

// Run 运行WebSocket管理器
func (manager *WebSocketManager) Run() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()
			log.Printf("WebSocket client connected: %s", client.userID)

			// 发送欢迎消息
			welcomeMsg := Message{
				Type:      "system",
				Data:      SystemMessage{Level: "info", Title: "Connected", Content: "WebSocket connection established"},
				Timestamp: time.Now(),
				ID:        uuid.New().String(),
			}
			client.SendMessage(welcomeMsg)

		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				close(client.send)
				log.Printf("WebSocket client disconnected: %s", client.userID)
			}
			manager.mutex.Unlock()

		case message := <-manager.broadcast:
			manager.mutex.RLock()
			for client := range manager.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(manager.clients, client)
				}
			}
			manager.mutex.RUnlock()
		}
	}
}

// HandleWebSocket 处理WebSocket连接
func (manager *WebSocketManager) HandleWebSocket(c *gin.Context) {
	// 验证JWT令牌（可选）
	var userID uuid.UUID
	if token := c.Query("token"); token != "" {
		if claims, err := manager.jwtManager.VerifyToken(token); err == nil {
			userID = claims.UserID
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	} else {
		// 匿名连接
		userID = uuid.New()
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// 创建客户端
	client := &Client{
		conn:          conn,
		send:          make(chan []byte, 256),
		userID:        userID,
		manager:       manager,
		subscriptions: make(map[string]bool),
	}

	// 注册客户端
	manager.register <- client

	// 启动客户端的读写协程
	go client.writePump()
	go client.readPump()
}

// SendMessage 发送消息给客户端
func (client *Client) SendMessage(message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		close(client.send)
	}
}

// Subscribe 订阅主题
func (client *Client) Subscribe(topics []string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	for _, topic := range topics {
		client.subscriptions[topic] = true
	}

	log.Printf("Client %s subscribed to topics: %v", client.userID, topics)
}

// Unsubscribe 取消订阅主题
func (client *Client) Unsubscribe(topics []string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	for _, topic := range topics {
		delete(client.subscriptions, topic)
	}

	log.Printf("Client %s unsubscribed from topics: %v", client.userID, topics)
}

// IsSubscribed 检查是否订阅了主题
func (client *Client) IsSubscribed(topic string) bool {
	client.mutex.RLock()
	defer client.mutex.RUnlock()
	return client.subscriptions[topic]
}

// readPump 读取消息
func (client *Client) readPump() {
	defer func() {
		client.manager.unregister <- client
		client.conn.Close()
	}()

	// 设置读取超时
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageData, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// 处理客户端消息
		client.handleMessage(messageData)
	}
}

// writePump 写入消息
func (client *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中的消息
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (client *Client) handleMessage(messageData []byte) {
	var message Message
	if err := json.Unmarshal(messageData, &message); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	switch message.Type {
	case "subscribe":
		var subMsg SubscribeMessage
		if data, err := json.Marshal(message.Data); err == nil {
			if err := json.Unmarshal(data, &subMsg); err == nil {
				if subMsg.Action == "subscribe" {
					client.Subscribe(subMsg.Topics)
				} else if subMsg.Action == "unsubscribe" {
					client.Unsubscribe(subMsg.Topics)
				}
			}
		}

	case "ping":
		// 响应ping消息
		pongMsg := Message{
			Type:      "pong",
			Timestamp: time.Now(),
			ID:        message.ID,
		}
		client.SendMessage(pongMsg)

	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

// BroadcastAlert 广播告警消息
func (manager *WebSocketManager) BroadcastAlert(alert *models.Alert, rule *models.AlertRule) {
	alertMsg := AlertMessage{
		Alert:    alert,
		Rule:     rule,
		Severity: alert.Severity,
		Message:  alert.Summary, // 使用Summary字段而不是Message
	}

	message := Message{
		Type:      "alert",
		Topic:     "alerts",
		Data:      alertMsg,
		Timestamp: time.Now(),
		ID:        uuid.New().String(),
	}

	manager.BroadcastToTopic("alerts", message)
}

// BroadcastMetrics 广播指标消息
func (manager *WebSocketManager) BroadcastMetrics(targetID, targetName string, metrics map[string]interface{}) {
	metricMsg := MetricMessage{
		TargetID:   targetID,
		TargetName: targetName,
		Metrics:    metrics,
		Timestamp:  time.Now(),
	}

	message := Message{
		Type:      "metrics",
		Topic:     "metrics",
		Data:      metricMsg,
		Timestamp: time.Now(),
		ID:        uuid.New().String(),
	}

	manager.BroadcastToTopic("metrics", message)
}

// BroadcastSystem 广播系统消息
func (manager *WebSocketManager) BroadcastSystem(level, title, content string) {
	sysMsg := SystemMessage{
		Level:   level,
		Title:   title,
		Content: content,
	}

	message := Message{
		Type:      "system",
		Topic:     "system",
		Data:      sysMsg,
		Timestamp: time.Now(),
		ID:        uuid.New().String(),
	}

	manager.BroadcastToTopic("system", message)
}

// BroadcastToTopic 向订阅了特定主题的客户端广播消息
func (manager *WebSocketManager) BroadcastToTopic(topic string, message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	for client := range manager.clients {
		if client.IsSubscribed(topic) {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(manager.clients, client)
			}
		}
	}
}

// BroadcastToUser 向特定用户发送消息
func (manager *WebSocketManager) BroadcastToUser(userID uuid.UUID, message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	for client := range manager.clients {
		if client.userID == userID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(manager.clients, client)
			}
		}
	}
}

// GetConnectedClients 获取连接的客户端数量
func (manager *WebSocketManager) GetConnectedClients() int {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return len(manager.clients)
}

// GetClientsByTopic 获取订阅了特定主题的客户端数量
func (manager *WebSocketManager) GetClientsByTopic(topic string) int {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	count := 0
	for client := range manager.clients {
		if client.IsSubscribed(topic) {
			count++
		}
	}
	return count
}

// Close 关闭WebSocket管理器
func (manager *WebSocketManager) Close() {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for client := range manager.clients {
		close(client.send)
		client.conn.Close()
	}

	close(manager.broadcast)
	close(manager.register)
	close(manager.unregister)
}