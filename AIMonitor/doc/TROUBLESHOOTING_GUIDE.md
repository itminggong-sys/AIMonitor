# AI监控系统故障排除指南

## 概述

本文档提供了AI监控系统常见问题的诊断和解决方案，帮助运维人员快速定位和修复系统故障。

## 目录

1. [系统启动问题](#系统启动问题)
2. [数据库连接问题](#数据库连接问题)
3. [Redis连接问题](#redis连接问题)
4. [API接口问题](#api接口问题)
5. [WebSocket连接问题](#websocket连接问题)
6. [AI服务问题](#ai服务问题)
7. [性能问题](#性能问题)
8. [监控数据问题](#监控数据问题)
9. [告警问题](#告警问题)
10. [日志分析](#日志分析)
11. [网络问题](#网络问题)
12. [安全问题](#安全问题)

## 系统启动问题

### 问题：应用启动失败

#### 症状
- 应用无法启动
- 启动过程中出现错误
- 进程立即退出

#### 诊断步骤

1. **检查配置文件**
```bash
# 验证配置文件语法
go run cmd/config-validator/main.go

# 检查配置文件权限
ls -la configs/
```

2. **检查端口占用**
```bash
# Windows
netstat -ano | findstr :8080

# Linux/macOS
lsof -i :8080
netstat -tulpn | grep :8080
```

3. **检查环境变量**
```bash
# 验证必需的环境变量
echo $DATABASE_URL
echo $REDIS_URL
echo $JWT_SECRET
```

4. **查看启动日志**
```bash
# 查看应用日志
tail -f logs/app.log

# 查看系统日志
# Windows
Get-EventLog -LogName Application -Source "AIMonitor"

# Linux
journalctl -u aimonitor -f
```

#### 常见解决方案

1. **配置文件错误**
```yaml
# 确保配置文件格式正确
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
```

2. **端口冲突**
```bash
# 修改配置文件中的端口
# 或者停止占用端口的进程
kill -9 <PID>
```

3. **权限问题**
```bash
# 设置正确的文件权限
chmod 644 configs/*.yaml
chmod 755 bin/aimonitor
```

### 问题：依赖服务不可用

#### 症状
- 数据库连接失败
- Redis连接失败
- 外部API不可达

#### 诊断脚本
```bash
#!/bin/bash
# scripts/health_check.sh

set -e

echo "检查系统健康状态..."

# 检查数据库连接
echo "检查数据库连接..."
psql $DATABASE_URL -c "SELECT 1;" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ 数据库连接正常"
else
    echo "✗ 数据库连接失败"
    exit 1
fi

# 检查Redis连接
echo "检查Redis连接..."
redis-cli -u $REDIS_URL ping > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ Redis连接正常"
else
    echo "✗ Redis连接失败"
    exit 1
fi

# 检查应用健康接口
echo "检查应用健康接口..."
curl -f http://localhost:8080/api/v1/health > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ 应用健康检查通过"
else
    echo "✗ 应用健康检查失败"
    exit 1
fi

echo "所有检查通过！"
```

## 数据库连接问题

### 问题：数据库连接超时

#### 症状
- 连接数据库超时
- 查询执行缓慢
- 连接池耗尽

#### 诊断工具
```go
// internal/database/diagnostics.go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

type DatabaseDiagnostics struct {
    db *sql.DB
}

func NewDatabaseDiagnostics(db *sql.DB) *DatabaseDiagnostics {
    return &DatabaseDiagnostics{db: db}
}

func (dd *DatabaseDiagnostics) CheckConnection(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    return dd.db.PingContext(ctx)
}

func (dd *DatabaseDiagnostics) GetConnectionStats() sql.DBStats {
    return dd.db.Stats()
}

func (dd *DatabaseDiagnostics) CheckSlowQueries(ctx context.Context) ([]SlowQuery, error) {
    query := `
        SELECT query, calls, total_time, mean_time, rows
        FROM pg_stat_statements
        WHERE mean_time > 1000  -- 超过1秒的查询
        ORDER BY mean_time DESC
        LIMIT 10
    `
    
    rows, err := dd.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var slowQueries []SlowQuery
    for rows.Next() {
        var sq SlowQuery
        if err := rows.Scan(&sq.Query, &sq.Calls, &sq.TotalTime, &sq.MeanTime, &sq.Rows); err != nil {
            continue
        }
        slowQueries = append(slowQueries, sq)
    }
    
    return slowQueries, nil
}

type SlowQuery struct {
    Query     string  `json:"query"`
    Calls     int64   `json:"calls"`
    TotalTime float64 `json:"total_time"`
    MeanTime  float64 `json:"mean_time"`
    Rows      int64   `json:"rows"`
}

func (dd *DatabaseDiagnostics) GenerateReport(ctx context.Context) (*DatabaseReport, error) {
    report := &DatabaseReport{
        Timestamp: time.Now(),
    }
    
    // 检查连接
    if err := dd.CheckConnection(ctx); err != nil {
        report.ConnectionStatus = "FAILED"
        report.ConnectionError = err.Error()
    } else {
        report.ConnectionStatus = "OK"
    }
    
    // 获取连接统计
    stats := dd.GetConnectionStats()
    report.ConnectionStats = ConnectionStats{
        OpenConnections: stats.OpenConnections,
        InUse:          stats.InUse,
        Idle:           stats.Idle,
        WaitCount:      stats.WaitCount,
        WaitDuration:   stats.WaitDuration,
    }
    
    // 检查慢查询
    slowQueries, err := dd.CheckSlowQueries(ctx)
    if err != nil {
        report.SlowQueryError = err.Error()
    } else {
        report.SlowQueries = slowQueries
    }
    
    return report, nil
}

type DatabaseReport struct {
    Timestamp        time.Time         `json:"timestamp"`
    ConnectionStatus string            `json:"connection_status"`
    ConnectionError  string            `json:"connection_error,omitempty"`
    ConnectionStats  ConnectionStats   `json:"connection_stats"`
    SlowQueries      []SlowQuery       `json:"slow_queries"`
    SlowQueryError   string            `json:"slow_query_error,omitempty"`
}

type ConnectionStats struct {
    OpenConnections int           `json:"open_connections"`
    InUse          int           `json:"in_use"`
    Idle           int           `json:"idle"`
    WaitCount      int64         `json:"wait_count"`
    WaitDuration   time.Duration `json:"wait_duration"`
}
```

#### 解决方案

1. **优化连接池配置**
```yaml
database:
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 1h
  conn_max_idle_time: 30m
```

2. **查询优化**
```sql
-- 添加索引
CREATE INDEX CONCURRENTLY idx_alerts_created_at ON alerts(created_at);
CREATE INDEX CONCURRENTLY idx_metrics_timestamp ON metrics(timestamp);

-- 分析表统计信息
ANALYZE alerts;
ANALYZE metrics;
```

3. **连接监控**
```go
// 定期监控连接状态
func (s *Server) monitorDatabaseConnections() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            stats := s.db.Stats()
            
            // 记录连接统计
            log.Printf("DB连接统计 - 打开: %d, 使用中: %d, 空闲: %d, 等待: %d",
                stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
            
            // 检查连接池是否接近耗尽
            if float64(stats.InUse)/float64(stats.OpenConnections) > 0.8 {
                log.Warn("数据库连接池使用率过高")
            }
        }
    }
}
```

### 问题：数据库锁等待

#### 症状
- 查询长时间等待
- 事务超时
- 死锁错误

#### 诊断查询
```sql
-- 查看当前锁等待
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
WHERE NOT blocked_locks.GRANTED;

-- 查看长时间运行的查询
SELECT 
    pid,
    now() - pg_stat_activity.query_start AS duration,
    query,
    state
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes';
```

## Redis连接问题

### 问题：Redis连接失败

#### 症状
- 无法连接到Redis
- 连接超时
- 认证失败

#### 诊断工具
```go
// internal/cache/diagnostics.go
package cache

import (
    "context"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type RedisDiagnostics struct {
    client *redis.Client
}

func NewRedisDiagnostics(client *redis.Client) *RedisDiagnostics {
    return &RedisDiagnostics{client: client}
}

func (rd *RedisDiagnostics) CheckConnection(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    return rd.client.Ping(ctx).Err()
}

func (rd *RedisDiagnostics) GetInfo(ctx context.Context) (map[string]string, error) {
    result, err := rd.client.Info(ctx).Result()
    if err != nil {
        return nil, err
    }
    
    info := make(map[string]string)
    lines := strings.Split(result, "\r\n")
    
    for _, line := range lines {
        if strings.Contains(line, ":") {
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                info[parts[0]] = parts[1]
            }
        }
    }
    
    return info, nil
}

func (rd *RedisDiagnostics) CheckMemoryUsage(ctx context.Context) (*RedisMemoryInfo, error) {
    info, err := rd.GetInfo(ctx)
    if err != nil {
        return nil, err
    }
    
    memInfo := &RedisMemoryInfo{}
    
    if usedMemory, exists := info["used_memory"]; exists {
        fmt.Sscanf(usedMemory, "%d", &memInfo.UsedMemory)
    }
    
    if maxMemory, exists := info["maxmemory"]; exists {
        fmt.Sscanf(maxMemory, "%d", &memInfo.MaxMemory)
    }
    
    if memInfo.MaxMemory > 0 {
        memInfo.UsagePercent = float64(memInfo.UsedMemory) / float64(memInfo.MaxMemory) * 100
    }
    
    return memInfo, nil
}

type RedisMemoryInfo struct {
    UsedMemory   int64   `json:"used_memory"`
    MaxMemory    int64   `json:"max_memory"`
    UsagePercent float64 `json:"usage_percent"`
}

func (rd *RedisDiagnostics) GenerateReport(ctx context.Context) (*RedisReport, error) {
    report := &RedisReport{
        Timestamp: time.Now(),
    }
    
    // 检查连接
    if err := rd.CheckConnection(ctx); err != nil {
        report.ConnectionStatus = "FAILED"
        report.ConnectionError = err.Error()
        return report, nil
    }
    
    report.ConnectionStatus = "OK"
    
    // 获取内存信息
    memInfo, err := rd.CheckMemoryUsage(ctx)
    if err != nil {
        report.MemoryError = err.Error()
    } else {
        report.MemoryInfo = memInfo
    }
    
    // 获取基本信息
    info, err := rd.GetInfo(ctx)
    if err != nil {
        report.InfoError = err.Error()
    } else {
        report.ServerInfo = map[string]string{
            "redis_version": info["redis_version"],
            "uptime_in_seconds": info["uptime_in_seconds"],
            "connected_clients": info["connected_clients"],
            "total_commands_processed": info["total_commands_processed"],
        }
    }
    
    return report, nil
}

type RedisReport struct {
    Timestamp        time.Time          `json:"timestamp"`
    ConnectionStatus string             `json:"connection_status"`
    ConnectionError  string             `json:"connection_error,omitempty"`
    MemoryInfo       *RedisMemoryInfo   `json:"memory_info,omitempty"`
    MemoryError      string             `json:"memory_error,omitempty"`
    ServerInfo       map[string]string  `json:"server_info,omitempty"`
    InfoError        string             `json:"info_error,omitempty"`
}
```

#### 解决方案

1. **检查Redis服务状态**
```bash
# 检查Redis进程
ps aux | grep redis

# 检查Redis服务状态
systemctl status redis

# 测试连接
redis-cli ping
```

2. **配置优化**
```redis
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
timeout 300
tcp-keepalive 60
```

3. **连接池配置**
```yaml
redis:
  pool_size: 10
  min_idle_conns: 5
  pool_timeout: 30s
  idle_timeout: 5m
  idle_check_frequency: 1m
```

## API接口问题

### 问题：API响应缓慢

#### 症状
- 接口响应时间过长
- 超时错误
- 高延迟

#### 性能分析工具
```go
// internal/middleware/performance.go
package middleware

import (
    "context"
    "net/http"
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
)

var (
    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP请求持续时间",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path", "status"},
    )
    
    httpRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "HTTP请求总数",
        },
        []string{"method", "path", "status"},
    )
)

func init() {
    prometheus.MustRegister(httpDuration, httpRequests)
}

func PerformanceMonitoring() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 处理请求
        c.Next()
        
        // 记录指标
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        httpDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Observe(duration)
        
        httpRequests.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Inc()
        
        // 记录慢请求
        if duration > 1.0 {
            log.Printf("慢请求: %s %s - %v", 
                c.Request.Method, c.Request.URL.Path, time.Since(start))
        }
    }
}

// 请求追踪中间件
func RequestTracing() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := generateRequestID()
        c.Header("X-Request-ID", requestID)
        
        // 添加到上下文
        ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
        c.Request = c.Request.WithContext(ctx)
        
        start := time.Now()
        
        log.Printf("[%s] 开始处理请求: %s %s", 
            requestID, c.Request.Method, c.Request.URL.Path)
        
        c.Next()
        
        log.Printf("[%s] 请求处理完成: %s %s - %v - %d", 
            requestID, c.Request.Method, c.Request.URL.Path, 
            time.Since(start), c.Writer.Status())
    }
}

func generateRequestID() string {
    return fmt.Sprintf("%d-%s", time.Now().UnixNano(), 
        randomString(8))
}
```

#### 诊断步骤

1. **检查API响应时间**
```bash
# 使用curl测试API响应时间
curl -w "@curl-format.txt" -o /dev/null -s "http://localhost:8080/api/v1/health"

# curl-format.txt内容:
# time_namelookup:  %{time_namelookup}\n
# time_connect:     %{time_connect}\n
# time_appconnect:  %{time_appconnect}\n
# time_pretransfer: %{time_pretransfer}\n
# time_redirect:    %{time_redirect}\n
# time_starttransfer: %{time_starttransfer}\n
# ----------\n
# time_total:       %{time_total}\n
```

2. **分析慢查询**
```go
// 慢查询分析
func (s *Server) analyzeSlowAPIs() {
    // 从Prometheus获取慢API数据
    query := `
        histogram_quantile(0.95, 
            rate(http_request_duration_seconds_bucket[5m])
        ) > 1
    `
    
    // 执行查询并分析结果
    // ...
}
```

#### 解决方案

1. **添加缓存**
```go
func (h *Handler) GetDashboardWithCache(c *gin.Context) {
    dashboardID := c.Param("id")
    cacheKey := fmt.Sprintf("dashboard:%s", dashboardID)
    
    // 尝试从缓存获取
    if cached, found := h.cache.Get(c.Request.Context(), cacheKey); found {
        c.JSON(http.StatusOK, cached)
        return
    }
    
    // 从数据库获取
    dashboard, err := h.service.GetDashboard(c.Request.Context(), dashboardID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // 缓存结果
    h.cache.Set(c.Request.Context(), cacheKey, dashboard, 5*time.Minute)
    
    c.JSON(http.StatusOK, dashboard)
}
```

2. **数据库查询优化**
```go
// 使用分页
func (s *Service) GetAlerts(ctx context.Context, page, size int) (*AlertList, error) {
    offset := (page - 1) * size
    
    query := `
        SELECT id, title, description, severity, status, created_at
        FROM alerts
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `
    
    rows, err := s.db.QueryContext(ctx, query, size, offset)
    // ...
}

// 使用索引
func (s *Service) GetAlertsByStatus(ctx context.Context, status string) ([]*Alert, error) {
    query := `
        SELECT id, title, description, severity, status, created_at
        FROM alerts
        WHERE status = $1
        ORDER BY created_at DESC
    `
    
    rows, err := s.db.QueryContext(ctx, query, status)
    // ...
}
```

3. **异步处理**
```go
func (h *Handler) CreateAlertAsync(c *gin.Context) {
    var req CreateAlertRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 立即返回响应
    alertID := generateAlertID()
    c.JSON(http.StatusAccepted, gin.H{
        "alert_id": alertID,
        "status": "processing",
    })
    
    // 异步处理
    go func() {
        if err := h.service.CreateAlert(context.Background(), &req); err != nil {
            log.Printf("创建告警失败: %v", err)
        }
    }()
}
```

### 问题：API错误率高

#### 症状
- 大量4xx/5xx错误
- 请求失败
- 异常响应

#### 错误监控
```go
// internal/middleware/error_tracking.go
package middleware

import (
    "bytes"
    "io/ioutil"
    "net/http"
    
    "github.com/gin-gonic/gin"
)

type ErrorTracker struct {
    errors chan ErrorEvent
}

type ErrorEvent struct {
    RequestID   string            `json:"request_id"`
    Method      string            `json:"method"`
    Path        string            `json:"path"`
    StatusCode  int               `json:"status_code"`
    Error       string            `json:"error"`
    RequestBody string            `json:"request_body,omitempty"`
    Headers     map[string]string `json:"headers"`
    Timestamp   time.Time         `json:"timestamp"`
}

func NewErrorTracker() *ErrorTracker {
    et := &ErrorTracker{
        errors: make(chan ErrorEvent, 1000),
    }
    
    go et.processErrors()
    
    return et
}

func (et *ErrorTracker) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 读取请求体
        var requestBody string
        if c.Request.Body != nil {
            bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
            requestBody = string(bodyBytes)
            c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
        }
        
        c.Next()
        
        // 检查是否有错误
        if c.Writer.Status() >= 400 {
            requestID, _ := c.Get("request_id")
            
            headers := make(map[string]string)
            for key, values := range c.Request.Header {
                if len(values) > 0 {
                    headers[key] = values[0]
                }
            }
            
            errorMsg := ""
            if len(c.Errors) > 0 {
                errorMsg = c.Errors.Last().Error()
            }
            
            event := ErrorEvent{
                RequestID:   requestID.(string),
                Method:      c.Request.Method,
                Path:        c.Request.URL.Path,
                StatusCode:  c.Writer.Status(),
                Error:       errorMsg,
                RequestBody: requestBody,
                Headers:     headers,
                Timestamp:   time.Now(),
            }
            
            select {
            case et.errors <- event:
            default:
                log.Println("错误事件队列已满")
            }
        }
    }
}

func (et *ErrorTracker) processErrors() {
    for event := range et.errors {
        // 记录错误日志
        log.Printf("API错误: [%s] %s %s - %d - %s", 
            event.RequestID, event.Method, event.Path, 
            event.StatusCode, event.Error)
        
        // 发送到监控系统
        et.sendToMonitoring(event)
        
        // 检查是否需要告警
        et.checkAlertThreshold(event)
    }
}

func (et *ErrorTracker) sendToMonitoring(event ErrorEvent) {
    // 发送到Prometheus、Grafana等监控系统
    // ...
}

func (et *ErrorTracker) checkAlertThreshold(event ErrorEvent) {
    // 检查错误率是否超过阈值
    // 如果超过，发送告警
    // ...
}
```

## WebSocket连接问题

### 问题：WebSocket连接断开

#### 症状
- 连接频繁断开
- 消息丢失
- 连接超时

#### 诊断工具
```go
// internal/websocket/diagnostics.go
package websocket

import (
    "sync"
    "time"
)

type ConnectionDiagnostics struct {
    connections map[string]*ConnectionInfo
    mu          sync.RWMutex
}

type ConnectionInfo struct {
    ID            string    `json:"id"`
    UserID        string    `json:"user_id"`
    ConnectedAt   time.Time `json:"connected_at"`
    LastPingAt    time.Time `json:"last_ping_at"`
    MessagesSent  int64     `json:"messages_sent"`
    MessagesRecv  int64     `json:"messages_received"`
    Errors        int64     `json:"errors"`
    Status        string    `json:"status"`
}

func NewConnectionDiagnostics() *ConnectionDiagnostics {
    return &ConnectionDiagnostics{
        connections: make(map[string]*ConnectionInfo),
    }
}

func (cd *ConnectionDiagnostics) AddConnection(connID, userID string) {
    cd.mu.Lock()
    defer cd.mu.Unlock()
    
    cd.connections[connID] = &ConnectionInfo{
        ID:          connID,
        UserID:      userID,
        ConnectedAt: time.Now(),
        LastPingAt:  time.Now(),
        Status:      "connected",
    }
}

func (cd *ConnectionDiagnostics) RemoveConnection(connID string) {
    cd.mu.Lock()
    defer cd.mu.Unlock()
    
    if conn, exists := cd.connections[connID]; exists {
        conn.Status = "disconnected"
        delete(cd.connections, connID)
    }
}

func (cd *ConnectionDiagnostics) UpdatePing(connID string) {
    cd.mu.Lock()
    defer cd.mu.Unlock()
    
    if conn, exists := cd.connections[connID]; exists {
        conn.LastPingAt = time.Now()
    }
}

func (cd *ConnectionDiagnostics) IncrementMessagesSent(connID string) {
    cd.mu.Lock()
    defer cd.mu.Unlock()
    
    if conn, exists := cd.connections[connID]; exists {
        conn.MessagesSent++
    }
}

func (cd *ConnectionDiagnostics) IncrementMessagesReceived(connID string) {
    cd.mu.Lock()
    defer cd.mu.Unlock()
    
    if conn, exists := cd.connections[connID]; exists {
        conn.MessagesRecv++
    }
}

func (cd *ConnectionDiagnostics) IncrementErrors(connID string) {
    cd.mu.Lock()
    defer cd.mu.Unlock()
    
    if conn, exists := cd.connections[connID]; exists {
        conn.Errors++
    }
}

func (cd *ConnectionDiagnostics) GetConnectionStats() map[string]*ConnectionInfo {
    cd.mu.RLock()
    defer cd.mu.RUnlock()
    
    stats := make(map[string]*ConnectionInfo)
    for id, conn := range cd.connections {
        stats[id] = &ConnectionInfo{
            ID:            conn.ID,
            UserID:        conn.UserID,
            ConnectedAt:   conn.ConnectedAt,
            LastPingAt:    conn.LastPingAt,
            MessagesSent:  conn.MessagesSent,
            MessagesRecv:  conn.MessagesRecv,
            Errors:        conn.Errors,
            Status:        conn.Status,
        }
    }
    
    return stats
}

func (cd *ConnectionDiagnostics) GetSummary() ConnectionSummary {
    cd.mu.RLock()
    defer cd.mu.RUnlock()
    
    summary := ConnectionSummary{
        TotalConnections: len(cd.connections),
        Timestamp:       time.Now(),
    }
    
    for _, conn := range cd.connections {
        summary.TotalMessagesSent += conn.MessagesSent
        summary.TotalMessagesReceived += conn.MessagesRecv
        summary.TotalErrors += conn.Errors
        
        // 检查僵尸连接
        if time.Since(conn.LastPingAt) > 5*time.Minute {
            summary.StaleConnections++
        }
    }
    
    return summary
}

type ConnectionSummary struct {
    TotalConnections       int       `json:"total_connections"`
    StaleConnections       int       `json:"stale_connections"`
    TotalMessagesSent      int64     `json:"total_messages_sent"`
    TotalMessagesReceived  int64     `json:"total_messages_received"`
    TotalErrors           int64     `json:"total_errors"`
    Timestamp             time.Time `json:"timestamp"`
}
```

#### 解决方案

1. **心跳检测**
```go
func (c *Connection) startHeartbeat() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := c.writeMessage(websocket.PingMessage, []byte{}); err != nil {
                log.Printf("发送心跳失败: %v", err)
                return
            }
        case <-c.done:
            return
        }
    }
}

func (c *Connection) handlePong(appData string) {
    c.lastPongAt = time.Now()
    c.diagnostics.UpdatePing(c.id)
}
```

2. **连接重试机制**
```javascript
// 客户端重连逻辑
class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectInterval = 1000;
        this.connect();
    }
    
    connect() {
        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = () => {
            console.log('WebSocket连接已建立');
            this.reconnectAttempts = 0;
        };
        
        this.ws.onclose = (event) => {
            console.log('WebSocket连接已关闭:', event.code, event.reason);
            this.handleReconnect();
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket错误:', error);
        };
        
        this.ws.onmessage = (event) => {
            this.handleMessage(JSON.parse(event.data));
        };
    }
    
    handleReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = this.reconnectInterval * Math.pow(2, this.reconnectAttempts - 1);
            
            console.log(`${delay}ms后尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            
            setTimeout(() => {
                this.connect();
            }, delay);
        } else {
            console.error('达到最大重连次数，停止重连');
        }
    }
}
```

3. **消息队列**
```go
type MessageQueue struct {
    messages []Message
    mu       sync.Mutex
    maxSize  int
}

func (mq *MessageQueue) Enqueue(msg Message) {
    mq.mu.Lock()
    defer mq.mu.Unlock()
    
    if len(mq.messages) >= mq.maxSize {
        // 移除最旧的消息
        mq.messages = mq.messages[1:]
    }
    
    mq.messages = append(mq.messages, msg)
}

func (mq *MessageQueue) DequeueAll() []Message {
    mq.mu.Lock()
    defer mq.mu.Unlock()
    
    messages := make([]Message, len(mq.messages))
    copy(messages, mq.messages)
    mq.messages = mq.messages[:0]
    
    return messages
}
```

## AI服务问题

### 问题：AI分析超时

#### 症状
- AI分析请求超时
- 响应时间过长
- 分析结果不准确

#### 诊断工具
```go
// internal/ai/diagnostics.go
package ai

import (
    "context"
    "sync"
    "time"
)

type AIServiceDiagnostics struct {
    requests map[string]*RequestInfo
    mu       sync.RWMutex
    metrics  AIMetrics
}

type RequestInfo struct {
    ID          string        `json:"id"`
    Type        string        `json:"type"`
    StartTime   time.Time     `json:"start_time"`
    EndTime     *time.Time    `json:"end_time,omitempty"`
    Duration    time.Duration `json:"duration"`
    Status      string        `json:"status"`
    Error       string        `json:"error,omitempty"`
    InputSize   int           `json:"input_size"`
    OutputSize  int           `json:"output_size"`
}

type AIMetrics struct {
    TotalRequests     int64         `json:"total_requests"`
    SuccessfulRequests int64        `json:"successful_requests"`
    FailedRequests    int64         `json:"failed_requests"`
    TimeoutRequests   int64         `json:"timeout_requests"`
    AverageLatency    time.Duration `json:"average_latency"`
    MaxLatency        time.Duration `json:"max_latency"`
    MinLatency        time.Duration `json:"min_latency"`
}

func NewAIServiceDiagnostics() *AIServiceDiagnostics {
    return &AIServiceDiagnostics{
        requests: make(map[string]*RequestInfo),
        metrics:  AIMetrics{MinLatency: time.Hour}, // 初始化为一个大值
    }
}

func (asd *AIServiceDiagnostics) StartRequest(id, requestType string, inputSize int) {
    asd.mu.Lock()
    defer asd.mu.Unlock()
    
    asd.requests[id] = &RequestInfo{
        ID:        id,
        Type:      requestType,
        StartTime: time.Now(),
        Status:    "processing",
        InputSize: inputSize,
    }
    
    asd.metrics.TotalRequests++
}

func (asd *AIServiceDiagnostics) CompleteRequest(id string, outputSize int, err error) {
    asd.mu.Lock()
    defer asd.mu.Unlock()
    
    req, exists := asd.requests[id]
    if !exists {
        return
    }
    
    now := time.Now()
    req.EndTime = &now
    req.Duration = now.Sub(req.StartTime)
    req.OutputSize = outputSize
    
    if err != nil {
        req.Status = "failed"
        req.Error = err.Error()
        asd.metrics.FailedRequests++
        
        if isTimeoutError(err) {
            asd.metrics.TimeoutRequests++
        }
    } else {
        req.Status = "completed"
        asd.metrics.SuccessfulRequests++
    }
    
    // 更新延迟统计
    asd.updateLatencyMetrics(req.Duration)
    
    // 清理旧请求
    delete(asd.requests, id)
}

func (asd *AIServiceDiagnostics) updateLatencyMetrics(duration time.Duration) {
    if duration > asd.metrics.MaxLatency {
        asd.metrics.MaxLatency = duration
    }
    
    if duration < asd.metrics.MinLatency {
        asd.metrics.MinLatency = duration
    }
    
    // 计算平均延迟（简化版本）
    if asd.metrics.SuccessfulRequests > 0 {
        total := time.Duration(asd.metrics.SuccessfulRequests) * asd.metrics.AverageLatency
        asd.metrics.AverageLatency = (total + duration) / time.Duration(asd.metrics.SuccessfulRequests)
    } else {
        asd.metrics.AverageLatency = duration
    }
}

func (asd *AIServiceDiagnostics) GetMetrics() AIMetrics {
    asd.mu.RLock()
    defer asd.mu.RUnlock()
    
    return asd.metrics
}

func (asd *AIServiceDiagnostics) GetActiveRequests() []*RequestInfo {
    asd.mu.RLock()
    defer asd.mu.RUnlock()
    
    var active []*RequestInfo
    for _, req := range asd.requests {
        if req.Status == "processing" {
            active = append(active, req)
        }
    }
    
    return active
}

func isTimeoutError(err error) bool {
    return strings.Contains(err.Error(), "timeout") || 
           strings.Contains(err.Error(), "deadline exceeded")
}
```

#### 解决方案

1. **超时控制**
```go
func (s *AIService) AnalyzeWithTimeout(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
    // 设置超时
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // 使用通道进行超时控制
    resultChan := make(chan *AnalysisResponse, 1)
    errorChan := make(chan error, 1)
    
    go func() {
        result, err := s.performAnalysis(ctx, req)
        if err != nil {
            errorChan <- err
        } else {
            resultChan <- result
        }
    }()
    
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errorChan:
        return nil, err
    case <-ctx.Done():
        return nil, fmt.Errorf("AI分析超时: %v", ctx.Err())
    }
}
```

2. **请求队列管理**
```go
type AIRequestQueue struct {
    queue    chan AnalysisRequest
    workers  int
    timeout  time.Duration
    metrics  *AIServiceDiagnostics
}

func NewAIRequestQueue(workers int, queueSize int, timeout time.Duration) *AIRequestQueue {
    arq := &AIRequestQueue{
        queue:   make(chan AnalysisRequest, queueSize),
        workers: workers,
        timeout: timeout,
        metrics: NewAIServiceDiagnostics(),
    }
    
    // 启动工作协程
    for i := 0; i < workers; i++ {
        go arq.worker(i)
    }
    
    return arq
}

func (arq *AIRequestQueue) worker(id int) {
    for req := range arq.queue {
        requestID := generateRequestID()
        arq.metrics.StartRequest(requestID, req.Type, len(req.Data))
        
        ctx, cancel := context.WithTimeout(context.Background(), arq.timeout)
        
        result, err := arq.processRequest(ctx, req)
        
        outputSize := 0
        if result != nil {
            outputSize = len(result.Result)
        }
        
        arq.metrics.CompleteRequest(requestID, outputSize, err)
        
        cancel()
        
        // 发送结果
        select {
        case req.ResultChan <- AIResponse{Result: result, Error: err}:
        case <-time.After(5 * time.Second):
            log.Printf("发送AI分析结果超时")
        }
    }
}

func (arq *AIRequestQueue) Submit(req AnalysisRequest) (*AnalysisResponse, error) {
    resultChan := make(chan AIResponse, 1)
    req.ResultChan = resultChan
    
    select {
    case arq.queue <- req:
        // 等待结果
        select {
        case result := <-resultChan:
            return result.Result, result.Error
        case <-time.After(arq.timeout + 10*time.Second):
            return nil, fmt.Errorf("AI分析请求超时")
        }
    case <-time.After(5 * time.Second):
        return nil, fmt.Errorf("AI分析队列已满")
    }
}
```

3. **模型优化**
```go
// 模型缓存
type ModelCache struct {
    models map[string]*AIModel
    mu     sync.RWMutex
    ttl    time.Duration
}

func (mc *ModelCache) GetModel(modelType string) (*AIModel, error) {
    mc.mu.RLock()
    model, exists := mc.models[modelType]
    mc.mu.RUnlock()
    
    if exists && time.Since(model.LoadedAt) < mc.ttl {
        return model, nil
    }
    
    // 加载模型
    mc.mu.Lock()
    defer mc.mu.Unlock()
    
    // 双重检查
    if model, exists := mc.models[modelType]; exists && time.Since(model.LoadedAt) < mc.ttl {
        return model, nil
    }
    
    newModel, err := mc.loadModel(modelType)
    if err != nil {
        return nil, err
    }
    
    mc.models[modelType] = newModel
    return newModel, nil
}
```

## 性能问题

### 问题：系统响应缓慢

#### 症状
- 整体响应时间增加
- CPU使用率高
- 内存使用率高
- 磁盘I/O高

#### 系统监控脚本
```bash
#!/bin/bash
# scripts/system_monitor.sh

set -e

LOG_FILE="/var/log/aimonitor/system_monitor.log"
THRESHOLD_CPU=80
THRESHOLD_MEMORY=85
THRESHOLD_DISK=90

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> $LOG_FILE
}

check_cpu() {
    CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
    CPU_USAGE=${CPU_USAGE%.*}
    
    if [ $CPU_USAGE -gt $THRESHOLD_CPU ]; then
        log_message "WARNING: CPU使用率过高: ${CPU_USAGE}%"
        
        # 获取CPU使用率最高的进程
        TOP_PROCESSES=$(ps aux --sort=-%cpu | head -10)
        log_message "CPU使用率最高的进程:\n$TOP_PROCESSES"
        
        return 1
    fi
    
    return 0
}

check_memory() {
    MEMORY_USAGE=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
    
    if [ $MEMORY_USAGE -gt $THRESHOLD_MEMORY ]; then
        log_message "WARNING: 内存使用率过高: ${MEMORY_USAGE}%"
        
        # 获取内存使用率最高的进程
        TOP_PROCESSES=$(ps aux --sort=-%mem | head -10)
        log_message "内存使用率最高的进程:\n$TOP_PROCESSES"
        
        return 1
    fi
    
    return 0
}

check_disk() {
    DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ $DISK_USAGE -gt $THRESHOLD_DISK ]; then
        log_message "WARNING: 磁盘使用率过高: ${DISK_USAGE}%"
        
        # 获取最大的文件和目录
        LARGE_FILES=$(find /var/log -type f -size +100M 2>/dev/null | head -10)
        log_message "大文件:\n$LARGE_FILES"
        
        return 1
    fi
    
    return 0
}

check_network() {
    # 检查网络连接数
    CONNECTIONS=$(netstat -an | wc -l)
    log_message "当前网络连接数: $CONNECTIONS"
    
    # 检查TIME_WAIT连接
    TIME_WAIT=$(netstat -an | grep TIME_WAIT | wc -l)
    if [ $TIME_WAIT -gt 1000 ]; then
        log_message "WARNING: TIME_WAIT连接过多: $TIME_WAIT"
    fi
}

check_application() {
    # 检查应用进程
    if ! pgrep -f "aimonitor" > /dev/null; then
        log_message "ERROR: AI监控应用进程未运行"
        return 1
    fi
    
    # 检查应用健康接口
    if ! curl -f http://localhost:8080/api/v1/health > /dev/null 2>&1; then
        log_message "ERROR: 应用健康检查失败"
        return 1
    fi
    
    return 0
}

main() {
    log_message "开始系统监控检查"
    
    ISSUES=0
    
    if ! check_cpu; then
        ISSUES=$((ISSUES + 1))
    fi
    
    if ! check_memory; then
        ISSUES=$((ISSUES + 1))
    fi
    
    if ! check_disk; then
        ISSUES=$((ISSUES + 1))
    fi
    
    check_network
    
    if ! check_application; then
        ISSUES=$((ISSUES + 1))
    fi
    
    if [ $ISSUES -gt 0 ]; then
        log_message "发现 $ISSUES 个问题"
        exit 1
    else
        log_message "系统监控检查通过"
        exit 0
    fi
}

main
```

#### 性能分析工具
```go
// internal/profiling/profiler.go
package profiling

import (
    "context"
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "time"
)

type Profiler struct {
    server *http.Server
}

func NewProfiler(port int) *Profiler {
    mux := http.NewServeMux()
    
    // 添加pprof路由
    mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
    mux.HandleFunc("/debug/pprof/cmdline", http.DefaultServeMux.ServeHTTP)
    mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
    mux.HandleFunc("/debug/pprof/symbol", http.DefaultServeMux.ServeHTTP)
    mux.HandleFunc("/debug/pprof/trace", http.DefaultServeMux.ServeHTTP)
    
    // 添加自定义性能指标
    mux.HandleFunc("/debug/stats", handleStats)
    mux.HandleFunc("/debug/gc", handleGC)
    
    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    
    return &Profiler{server: server}
}

func (p *Profiler) Start() error {
    go func() {
        if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("性能分析服务器启动失败: %v", err)
        }
    }()
    
    return nil
}

func (p *Profiler) Stop(ctx context.Context) error {
    return p.server.Shutdown(ctx)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    stats := map[string]interface{}{
        "goroutines": runtime.NumGoroutine(),
        "memory": map[string]interface{}{
            "alloc":       m.Alloc,
            "total_alloc": m.TotalAlloc,
            "sys":         m.Sys,
            "heap_alloc":  m.HeapAlloc,
            "heap_sys":    m.HeapSys,
            "heap_idle":   m.HeapIdle,
            "heap_inuse":  m.HeapInuse,
        },
        "gc": map[string]interface{}{
            "num_gc":        m.NumGC,
            "pause_total":   m.PauseTotalNs,
            "last_gc":       time.Unix(0, int64(m.LastGC)),
            "gc_cpu_fraction": m.GCCPUFraction,
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

func handleGC(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        // 手动触发GC
        runtime.GC()
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("GC triggered"))
    } else {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        gcStats := map[string]interface{}{
            "num_gc":         m.NumGC,
            "pause_total_ns": m.PauseTotalNs,
            "last_gc":        time.Unix(0, int64(m.LastGC)),
            "gc_cpu_fraction": m.GCCPUFraction,
            "next_gc":        m.NextGC,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(gcStats)
    }
}
```

#### 解决方案

1. **CPU优化**
```go
// 限制Goroutine数量
type WorkerPool struct {
    workers chan struct{}
}

func NewWorkerPool(size int) *WorkerPool {
    return &WorkerPool{
        workers: make(chan struct{}, size),
    }
}

func (wp *WorkerPool) Submit(task func()) {
    wp.workers <- struct{}{}
    go func() {
        defer func() { <-wp.workers }()
        task()
    }()
}
```

2. **内存优化**
```go
// 对象池
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func processData(data []byte) {
    buffer := bufferPool.Get().([]byte)
    defer bufferPool.Put(buffer)
    
    // 使用buffer处理数据
    // ...
}
```

3. **数据库连接优化**
```yaml
database:
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 1h
  conn_max_idle_time: 30m
```

## 监控数据问题

### 问题：监控数据丢失

#### 症状
- 监控指标缺失
- 数据不连续
- 历史数据丢失

#### 数据完整性检查
```go
// internal/monitoring/data_integrity.go
package monitoring

import (
    "context"
    "fmt"
    "time"
)

type DataIntegrityChecker struct {
    db MetricsDB
}

func (dic *DataIntegrityChecker) CheckDataGaps(ctx context.Context, metricName string, start, end time.Time) ([]DataGap, error) {
    query := `
        SELECT 
            timestamp,
            LAG(timestamp) OVER (ORDER BY timestamp) as prev_timestamp
        FROM metrics 
        WHERE metric_name = $1 
        AND timestamp BETWEEN $2 AND $3
        ORDER BY timestamp
    `
    
    rows, err := dic.db.QueryContext(ctx, query, metricName, start, end)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var gaps []DataGap
    expectedInterval := 1 * time.Minute // 假设数据间隔为1分钟
    
    for rows.Next() {
        var timestamp, prevTimestamp *time.Time
        if err := rows.Scan(&timestamp, &prevTimestamp); err != nil {
            continue
        }
        
        if prevTimestamp != nil {
            gap := timestamp.Sub(*prevTimestamp)
            if gap > expectedInterval*2 { // 超过2倍间隔认为是数据缺失
                gaps = append(gaps, DataGap{
                    MetricName: metricName,
                    StartTime:  *prevTimestamp,
                    EndTime:    *timestamp,
                    Duration:   gap,
                })
            }
        }
    }
    
    return gaps, nil
}

type DataGap struct {
    MetricName string        `json:"metric_name"`
    StartTime  time.Time     `json:"start_time"`
    EndTime    time.Time     `json:"end_time"`
    Duration   time.Duration `json:"duration"`
}
```

#### 解决方案

1. **数据备份策略**
```bash
#!/bin/bash
# scripts/backup_metrics.sh

BACKUP_DIR="/backup/metrics"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/metrics_$DATE.sql"

mkdir -p $BACKUP_DIR

# 备份数据库
pg_dump $DATABASE_URL > $BACKUP_FILE

# 压缩备份文件
gzip $BACKUP_FILE

# 清理7天前的备份
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "数据备份完成: ${BACKUP_FILE}.gz"
```

2. **数据恢复**
```go
func (s *Service) RecoverMissingData(ctx context.Context, metricName string, start, end time.Time) error {
    // 从备份源重新获取数据
    data, err := s.fetchDataFromSource(ctx, metricName, start, end)
    if err != nil {
        return err
    }
    
    // 批量插入数据
    return s.batchInsertMetrics(ctx, data)
}
```

## 告警问题

### 问题：告警不触发

#### 症状
- 满足条件但不发送告警
- 告警延迟
- 告警重复发送

#### 告警系统诊断
```go
// internal/alert/diagnostics.go
package alert

import (
    "context"
    "time"
)

type AlertDiagnostics struct {
    ruleEngine *RuleEngine
    notifier   *Notifier
}

func (ad *AlertDiagnostics) DiagnoseRule(ctx context.Context, ruleID string) (*RuleDiagnostic, error) {
    rule, err := ad.ruleEngine.GetRule(ruleID)
    if err != nil {
        return nil, err
    }
    
    diagnostic := &RuleDiagnostic{
        RuleID:    ruleID,
        RuleName:  rule.Name,
        Status:    rule.Status,
        Timestamp: time.Now(),
    }
    
    // 检查规则配置
    if err := ad.validateRuleConfig(rule); err != nil {
        diagnostic.ConfigErrors = append(diagnostic.ConfigErrors, err.Error())
    }
    
    // 检查数据源
    if err := ad.checkDataSource(ctx, rule); err != nil {
        diagnostic.DataSourceErrors = append(diagnostic.DataSourceErrors, err.Error())
    }
    
    // 检查最近的评估结果
    evaluations, err := ad.getRecentEvaluations(ctx, ruleID, 10)
    if err != nil {
        diagnostic.EvaluationError = err.Error()
    } else {
        diagnostic.RecentEvaluations = evaluations
    }
    
    return diagnostic, nil
}

type RuleDiagnostic struct {
    RuleID             string            `json:"rule_id"`
    RuleName           string            `json:"rule_name"`
    Status             string            `json:"status"`
    Timestamp          time.Time         `json:"timestamp"`
    ConfigErrors       []string          `json:"config_errors"`
    DataSourceErrors   []string          `json:"data_source_errors"`
    EvaluationError    string            `json:"evaluation_error,omitempty"`
    RecentEvaluations  []RuleEvaluation  `json:"recent_evaluations"`
}

type RuleEvaluation struct {
    Timestamp time.Time `json:"timestamp"`
    Result    bool      `json:"result"`
    Value     float64   `json:"value"`
    Threshold float64   `json:"threshold"`
    Error     string    `json:"error,omitempty"`
}
```

#### 解决方案

1. **告警规则验证**
```go
func (ad *AlertDiagnostics) validateRuleConfig(rule *AlertRule) error {
    // 检查查询语句
    if rule.Query == "" {
        return fmt.Errorf("查询语句不能为空")
    }
    
    // 检查阈值
    if rule.Threshold == 0 {
        return fmt.Errorf("阈值不能为0")
    }
    
    // 检查评估间隔
    if rule.EvaluationInterval < time.Minute {
        return fmt.Errorf("评估间隔不能小于1分钟")
    }
    
    // 检查通知配置
    if len(rule.NotificationChannels) == 0 {
        return fmt.Errorf("至少需要配置一个通知渠道")
    }
    
    return nil
}
```

2. **告警测试工具**
```go
func (ad *AlertDiagnostics) TestAlert(ctx context.Context, ruleID string) (*AlertTestResult, error) {
    rule, err := ad.ruleEngine.GetRule(ruleID)
    if err != nil {
        return nil, err
    }
    
    result := &AlertTestResult{
        RuleID:    ruleID,
        Timestamp: time.Now(),
    }
    
    // 执行查询
    value, err := ad.ruleEngine.ExecuteQuery(ctx, rule.Query)
    if err != nil {
        result.QueryError = err.Error()
        return result, nil
    }
    
    result.QueryResult = value
    result.Threshold = rule.Threshold
    result.Triggered = ad.evaluateCondition(value, rule.Condition, rule.Threshold)
    
    // 测试通知
    if result.Triggered {
        for _, channel := range rule.NotificationChannels {
            if err := ad.testNotification(ctx, channel, rule); err != nil {
                result.NotificationErrors = append(result.NotificationErrors, 
                    fmt.Sprintf("%s: %v", channel, err))
            }
        }
    }
    
    return result, nil
}

type AlertTestResult struct {
    RuleID             string    `json:"rule_id"`
    Timestamp          time.Time `json:"timestamp"`
    QueryResult        float64   `json:"query_result"`
    Threshold          float64   `json:"threshold"`
    Triggered          bool      `json:"triggered"`
    QueryError         string    `json:"query_error,omitempty"`
    NotificationErrors []string  `json:"notification_errors,omitempty"`
}
```

## 日志分析

### 日志聚合和分析

```bash
#!/bin/bash
# scripts/log_analysis.sh

LOG_DIR="/var/log/aimonitor"
OUTPUT_DIR="/tmp/log_analysis"
DATE=$(date +%Y-%m-%d)

mkdir -p $OUTPUT_DIR

echo "开始日志分析..."

# 错误日志统计
echo "=== 错误日志统计 ===" > $OUTPUT_DIR/error_summary.txt
grep -i "error" $LOG_DIR/app.log | \
    awk '{print $1, $2, $NF}' | \
    sort | uniq -c | sort -nr >> $OUTPUT_DIR/error_summary.txt

# API响应时间分析
echo "=== API响应时间分析 ===" > $OUTPUT_DIR/api_performance.txt
grep "请求处理完成" $LOG_DIR/app.log | \
    awk '{print $(NF-1)}' | \
    sed 's/ms//' | \
    sort -n | \
    awk '
    {
        times[NR] = $1
        sum += $1
    }
    END {
        print "总请求数:", NR
        print "平均响应时间:", sum/NR "ms"
        print "最小响应时间:", times[1] "ms"
        print "最大响应时间:", times[NR] "ms"
        print "P50:", times[int(NR*0.5)] "ms"
        print "P95:", times[int(NR*0.95)] "ms"
        print "P99:", times[int(NR*0.99)] "ms"
    }' >> $OUTPUT_DIR/api_performance.txt

# 数据库连接分析
echo "=== 数据库连接分析 ===" > $OUTPUT_DIR/db_connections.txt
grep "DB连接统计" $LOG_DIR/app.log | \
    tail -100 | \
    awk '{print $3, $5, $7, $9}' >> $OUTPUT_DIR/db_connections.txt

# 内存使用分析
echo "=== 内存使用分析 ===" > $OUTPUT_DIR/memory_usage.txt
grep "内存使用" $LOG_DIR/app.log | \
    awk '{print $1, $2, $NF}' | \
    tail -50 >> $OUTPUT_DIR/memory_usage.txt

echo "日志分析完成，结果保存在 $OUTPUT_DIR"
```

### 实时日志监控

```go
// internal/logging/monitor.go
package logging

import (
    "bufio"
    "context"
    "os"
    "regexp"
    "strings"
    "time"
)

type LogMonitor struct {
    patterns map[string]*regexp.Regexp
    alerts   chan LogAlert
    ctx      context.Context
    cancel   context.CancelFunc
}

type LogAlert struct {
    Level     string    `json:"level"`
    Pattern   string    `json:"pattern"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
    Count     int       `json:"count"`
}

func NewLogMonitor() *LogMonitor {
    ctx, cancel := context.WithCancel(context.Background())
    
    lm := &LogMonitor{
        patterns: make(map[string]*regexp.Regexp),
        alerts:   make(chan LogAlert, 100),
        ctx:      ctx,
        cancel:   cancel,
    }
    
    // 添加默认监控模式
    lm.AddPattern("error", regexp.MustCompile(`(?i)error|exception|failed|panic`))
    lm.AddPattern("timeout", regexp.MustCompile(`(?i)timeout|deadline exceeded`))
    lm.AddPattern("connection", regexp.MustCompile(`(?i)connection.*failed|connection.*refused`))
    
    return lm
}

func (lm *LogMonitor) AddPattern(name string, pattern *regexp.Regexp) {
    lm.patterns[name] = pattern
}

func (lm *LogMonitor) MonitorFile(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)
    
    // 跳到文件末尾
    file.Seek(0, 2)
    
    go func() {
        for scanner.Scan() {
            select {
            case <-lm.ctx.Done():
                return
            default:
                lm.processLogLine(scanner.Text())
            }
        }
    }()
    
    return nil
}

func (lm *LogMonitor) processLogLine(line string) {
    for patternName, pattern := range lm.patterns {
        if pattern.MatchString(line) {
            alert := LogAlert{
                Level:     "warning",
                Pattern:   patternName,
                Message:   line,
                Timestamp: time.Now(),
                Count:     1,
            }
            
            // 检查严重程度
            if strings.Contains(strings.ToLower(line), "panic") || 
               strings.Contains(strings.ToLower(line), "fatal") {
                alert.Level = "critical"
            }
            
            select {
            case lm.alerts <- alert:
            default:
                // 告警队列已满，丢弃
            }
            
            break
        }
    }
}

func (lm *LogMonitor) GetAlerts() <-chan LogAlert {
    return lm.alerts
}

func (lm *LogMonitor) Stop() {
    lm.cancel()
}
```

## 网络问题

### 问题：网络连接超时

#### 网络诊断工具

```bash
#!/bin/bash
# scripts/network_diagnosis.sh

TARGET_HOST="$1"
TARGET_PORT="$2"

if [ -z "$TARGET_HOST" ] || [ -z "$TARGET_PORT" ]; then
    echo "用法: $0 <主机> <端口>"
    exit 1
fi

echo "网络诊断报告 - $(date)"
echo "目标: $TARGET_HOST:$TARGET_PORT"
echo "=============================="

# DNS解析测试
echo "1. DNS解析测试:"
nslookup $TARGET_HOST
echo

# Ping测试
echo "2. Ping测试:"
ping -c 4 $TARGET_HOST
echo

# 端口连通性测试
echo "3. 端口连通性测试:"
if command -v nc >/dev/null 2>&1; then
    nc -zv $TARGET_HOST $TARGET_PORT
else
    telnet $TARGET_HOST $TARGET_PORT
fi
echo

# 路由跟踪
echo "4. 路由跟踪:"
traceroute $TARGET_HOST
echo

# 网络统计
echo "5. 网络统计:"
netstat -i
echo

# 防火墙检查
echo "6. 防火墙状态:"
if command -v ufw >/dev/null 2>&1; then
    ufw status
elif command -v iptables >/dev/null 2>&1; then
    iptables -L
fi
```

## 安全问题

### 问题：认证失败

#### 安全审计工具

```go
// internal/security/audit.go
package security

import (
    "context"
    "time"
)

type SecurityAuditor struct {
    db AuditDB
}

type SecurityEvent struct {
    ID          string    `json:"id"`
    Type        string    `json:"type"`
    UserID      string    `json:"user_id"`
    IP          string    `json:"ip"`
    UserAgent   string    `json:"user_agent"`
    Resource    string    `json:"resource"`
    Action      string    `json:"action"`
    Result      string    `json:"result"`
    Timestamp   time.Time `json:"timestamp"`
    Details     string    `json:"details"`
}

func (sa *SecurityAuditor) LogSecurityEvent(ctx context.Context, event SecurityEvent) error {
    event.ID = generateEventID()
    event.Timestamp = time.Now()
    
    return sa.db.InsertSecurityEvent(ctx, event)
}

func (sa *SecurityAuditor) DetectSuspiciousActivity(ctx context.Context) ([]SuspiciousActivity, error) {
    var activities []SuspiciousActivity
    
    // 检测暴力破解攻击
    bruteForce, err := sa.detectBruteForceAttacks(ctx)
    if err != nil {
        return nil, err
    }
    activities = append(activities, bruteForce...)
    
    // 检测异常登录
    anomalousLogins, err := sa.detectAnomalousLogins(ctx)
    if err != nil {
        return nil, err
    }
    activities = append(activities, anomalousLogins...)
    
    // 检测权限提升
    privilegeEscalation, err := sa.detectPrivilegeEscalation(ctx)
    if err != nil {
        return nil, err
    }
    activities = append(activities, privilegeEscalation...)
    
    return activities, nil
}

type SuspiciousActivity struct {
    Type        string    `json:"type"`
    Description string    `json:"description"`
    Severity    string    `json:"severity"`
    UserID      string    `json:"user_id"`
    IP          string    `json:"ip"`
    Count       int       `json:"count"`
    FirstSeen   time.Time `json:"first_seen"`
    LastSeen    time.Time `json:"last_seen"`
}

func (sa *SecurityAuditor) detectBruteForceAttacks(ctx context.Context) ([]SuspiciousActivity, error) {
    query := `
        SELECT ip, user_id, COUNT(*) as attempts, MIN(timestamp) as first_attempt, MAX(timestamp) as last_attempt
        FROM security_events
        WHERE type = 'login_failed'
        AND timestamp > NOW() - INTERVAL '1 hour'
        GROUP BY ip, user_id
        HAVING COUNT(*) >= 5
    `
    
    rows, err := sa.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var activities []SuspiciousActivity
    
    for rows.Next() {
        var ip, userID string
        var attempts int
        var firstAttempt, lastAttempt time.Time
        
        if err := rows.Scan(&ip, &userID, &attempts, &firstAttempt, &lastAttempt); err != nil {
            continue
        }
        
        activities = append(activities, SuspiciousActivity{
            Type:        "brute_force",
            Description: fmt.Sprintf("检测到来自IP %s的暴力破解攻击，目标用户: %s", ip, userID),
            Severity:    "high",
            UserID:      userID,
            IP:          ip,
            Count:       attempts,
            FirstSeen:   firstAttempt,
            LastSeen:    lastAttempt,
        })
    }
    
    return activities, nil
}
```

## 总结

本故障排除指南涵盖了AI监控系统的主要问题类型和解决方案：

### 快速诊断流程

1. **确定问题范围**
   - 系统级问题还是应用级问题
   - 影响范围和严重程度
   - 问题发生时间和频率

2. **收集诊断信息**
   - 查看系统日志
   - 检查监控指标
   - 运行诊断脚本

3. **分析根本原因**
   - 性能瓶颈分析
   - 资源使用分析
   - 依赖服务状态

4. **实施解决方案**
   - 临时缓解措施
   - 根本性修复
   - 预防措施

5. **验证修复效果**
   - 功能验证
   - 性能验证
   - 稳定性验证

### 预防措施

- 定期系统健康检查
- 监控关键指标和告警
- 定期备份重要数据
- 保持系统和依赖更新
- 进行容量规划
- 建立应急响应流程

通过遵循本指南，可以快速诊断和解决AI监控系统的各种问题，确保系统的稳定运行。