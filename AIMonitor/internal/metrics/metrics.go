package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics 指标收集器
type Metrics struct {
	// HTTP请求指标
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestsInFlight  *prometheus.GaugeVec

	// 数据库指标
	DBConnectionsTotal    *prometheus.GaugeVec
	DBConnectionsIdle     *prometheus.GaugeVec
	DBConnectionsInUse    *prometheus.GaugeVec
	DBQueryDuration       *prometheus.HistogramVec
	DBQueriesTotal        *prometheus.CounterVec

	// Redis指标
	RedisConnectionsTotal *prometheus.GaugeVec
	RedisCommandDuration  *prometheus.HistogramVec
	RedisCommandsTotal    *prometheus.CounterVec
	RedisCacheHits        *prometheus.CounterVec
	RedisCacheMisses      *prometheus.CounterVec

	// 告警指标
	AlertsTotal           *prometheus.CounterVec
	ActiveAlerts          *prometheus.GaugeVec
	AlertRulesTotal       *prometheus.GaugeVec
	AlertProcessingTime   *prometheus.HistogramVec

	// AI分析指标
	AIRequestsTotal       *prometheus.CounterVec
	AIRequestDuration     *prometheus.HistogramVec
	AITokensUsed          *prometheus.CounterVec
	AIAnalysisTotal       *prometheus.CounterVec

	// WebSocket指标
	WebSocketConnections  *prometheus.GaugeVec
	WebSocketMessages     *prometheus.CounterVec

	// 系统指标
	SystemInfo            *prometheus.GaugeVec
	Uptime                *prometheus.CounterVec
	MemoryUsage           *prometheus.GaugeVec
	Goroutines            *prometheus.GaugeVec

	// 业务指标
	UsersTotal            *prometheus.GaugeVec
	ActiveUsers           *prometheus.GaugeVec
	MonitoringTargets     *prometheus.GaugeVec
	DashboardsTotal       *prometheus.GaugeVec
}

// NewMetrics 创建指标收集器
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP请求指标
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "endpoint"},
		),

		// 数据库指标
		DBConnectionsTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "db_connections_total",
				Help: "Total number of database connections",
			},
			[]string{"database", "state"},
		),
		DBConnectionsIdle: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "db_connections_idle",
				Help: "Number of idle database connections",
			},
			[]string{"database"},
		),
		DBConnectionsInUse: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "db_connections_in_use",
				Help: "Number of database connections in use",
			},
			[]string{"database"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
			},
			[]string{"database", "operation"},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"database", "operation", "status"},
		),

		// Redis指标
		RedisConnectionsTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "redis_connections_total",
				Help: "Total number of Redis connections",
			},
			[]string{"instance"},
		),
		RedisCommandDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "redis_command_duration_seconds",
				Help:    "Redis command duration in seconds",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
			},
			[]string{"instance", "command"},
		),
		RedisCommandsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_commands_total",
				Help: "Total number of Redis commands",
			},
			[]string{"instance", "command", "status"},
		),
		RedisCacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_cache_hits_total",
				Help: "Total number of Redis cache hits",
			},
			[]string{"instance", "key_pattern"},
		),
		RedisCacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_cache_misses_total",
				Help: "Total number of Redis cache misses",
			},
			[]string{"instance", "key_pattern"},
		),

		// 告警指标
		AlertsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "alerts_total",
				Help: "Total number of alerts",
			},
			[]string{"severity", "status", "rule_name"},
		),
		ActiveAlerts: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "active_alerts",
				Help: "Number of active alerts",
			},
			[]string{"severity"},
		),
		AlertRulesTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "alert_rules_total",
				Help: "Total number of alert rules",
			},
			[]string{"enabled"},
		),
		AlertProcessingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "alert_processing_duration_seconds",
				Help:    "Alert processing duration in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"rule_name"},
		),

		// AI分析指标
		AIRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_requests_total",
				Help: "Total number of AI requests",
			},
			[]string{"provider", "model", "type", "status"},
		),
		AIRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ai_request_duration_seconds",
				Help:    "AI request duration in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
			},
			[]string{"provider", "model", "type"},
		),
		AITokensUsed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_tokens_used_total",
				Help: "Total number of AI tokens used",
			},
			[]string{"provider", "model", "type"},
		),
		AIAnalysisTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_analysis_total",
				Help: "Total number of AI analysis",
			},
			[]string{"analysis_type", "status"},
		),

		// WebSocket指标
		WebSocketConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "websocket_connections",
				Help: "Number of active WebSocket connections",
			},
			[]string{"authenticated"},
		),
		WebSocketMessages: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_messages_total",
				Help: "Total number of WebSocket messages",
			},
			[]string{"direction", "type"},
		),

		// 系统指标
		SystemInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "system_info",
				Help: "System information",
			},
			[]string{"version", "go_version", "build_time"},
		),
		Uptime: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "uptime_seconds_total",
				Help: "Total uptime in seconds",
			},
			[]string{"instance"},
		),
		MemoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"type"},
		),
		Goroutines: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "goroutines",
				Help: "Number of goroutines",
			},
			[]string{"instance"},
		),

		// 业务指标
		UsersTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "users_total",
				Help: "Total number of users",
			},
			[]string{"status"},
		),
		ActiveUsers: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "active_users",
				Help: "Number of active users",
			},
			[]string{"period"},
		),
		MonitoringTargets: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "monitoring_targets_total",
				Help: "Total number of monitoring targets",
			},
			[]string{"type", "enabled"},
		),
		DashboardsTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "dashboards_total",
				Help: "Total number of dashboards",
			},
			[]string{"type"},
		),
	}
}

// RecordHTTPRequest 记录HTTP请求指标
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// IncHTTPRequestsInFlight 增加正在处理的HTTP请求数
func (m *Metrics) IncHTTPRequestsInFlight(method, endpoint string) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Inc()
}

// DecHTTPRequestsInFlight 减少正在处理的HTTP请求数
func (m *Metrics) DecHTTPRequestsInFlight(method, endpoint string) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Dec()
}

// UpdateDBConnections 更新数据库连接指标
func (m *Metrics) UpdateDBConnections(database string, total, idle, inUse int) {
	m.DBConnectionsTotal.WithLabelValues(database, "total").Set(float64(total))
	m.DBConnectionsIdle.WithLabelValues(database).Set(float64(idle))
	m.DBConnectionsInUse.WithLabelValues(database).Set(float64(inUse))
}

// RecordDBQuery 记录数据库查询指标
func (m *Metrics) RecordDBQuery(database, operation, status string, duration time.Duration) {
	m.DBQueriesTotal.WithLabelValues(database, operation, status).Inc()
	m.DBQueryDuration.WithLabelValues(database, operation).Observe(duration.Seconds())
}

// UpdateRedisConnections 更新Redis连接指标
func (m *Metrics) UpdateRedisConnections(instance string, connections int) {
	m.RedisConnectionsTotal.WithLabelValues(instance).Set(float64(connections))
}

// RecordRedisCommand 记录Redis命令指标
func (m *Metrics) RecordRedisCommand(instance, command, status string, duration time.Duration) {
	m.RedisCommandsTotal.WithLabelValues(instance, command, status).Inc()
	m.RedisCommandDuration.WithLabelValues(instance, command).Observe(duration.Seconds())
}

// RecordCacheHit 记录缓存命中
func (m *Metrics) RecordCacheHit(instance, keyPattern string) {
	m.RedisCacheHits.WithLabelValues(instance, keyPattern).Inc()
}

// RecordCacheMiss 记录缓存未命中
func (m *Metrics) RecordCacheMiss(instance, keyPattern string) {
	m.RedisCacheMisses.WithLabelValues(instance, keyPattern).Inc()
}

// RecordAlert 记录告警指标
func (m *Metrics) RecordAlert(severity, status, ruleName string) {
	m.AlertsTotal.WithLabelValues(severity, status, ruleName).Inc()
}

// UpdateActiveAlerts 更新活跃告警数
func (m *Metrics) UpdateActiveAlerts(severity string, count int) {
	m.ActiveAlerts.WithLabelValues(severity).Set(float64(count))
}

// UpdateAlertRules 更新告警规则数
func (m *Metrics) UpdateAlertRules(enabled string, count int) {
	m.AlertRulesTotal.WithLabelValues(enabled).Set(float64(count))
}

// RecordAlertProcessing 记录告警处理时间
func (m *Metrics) RecordAlertProcessing(ruleName string, duration time.Duration) {
	m.AlertProcessingTime.WithLabelValues(ruleName).Observe(duration.Seconds())
}

// RecordAIRequest 记录AI请求指标
func (m *Metrics) RecordAIRequest(provider, model, requestType, status string, duration time.Duration, tokens int) {
	m.AIRequestsTotal.WithLabelValues(provider, model, requestType, status).Inc()
	m.AIRequestDuration.WithLabelValues(provider, model, requestType).Observe(duration.Seconds())
	m.AITokensUsed.WithLabelValues(provider, model, requestType).Add(float64(tokens))
}

// RecordAIAnalysis 记录AI分析指标
func (m *Metrics) RecordAIAnalysis(analysisType, status string) {
	m.AIAnalysisTotal.WithLabelValues(analysisType, status).Inc()
}

// UpdateWebSocketConnections 更新WebSocket连接数
func (m *Metrics) UpdateWebSocketConnections(authenticated string, count int) {
	m.WebSocketConnections.WithLabelValues(authenticated).Set(float64(count))
}

// RecordWebSocketMessage 记录WebSocket消息
func (m *Metrics) RecordWebSocketMessage(direction, messageType string) {
	m.WebSocketMessages.WithLabelValues(direction, messageType).Inc()
}

// UpdateSystemInfo 更新系统信息
func (m *Metrics) UpdateSystemInfo(version, goVersion, buildTime string) {
	m.SystemInfo.WithLabelValues(version, goVersion, buildTime).Set(1)
}

// UpdateUptime 更新运行时间
func (m *Metrics) UpdateUptime(instance string, uptime float64) {
	m.Uptime.WithLabelValues(instance).Add(uptime)
}

// UpdateMemoryUsage 更新内存使用情况
func (m *Metrics) UpdateMemoryUsage(memType string, bytes uint64) {
	m.MemoryUsage.WithLabelValues(memType).Set(float64(bytes))
}

// UpdateGoroutines 更新协程数
func (m *Metrics) UpdateGoroutines(instance string, count int) {
	m.Goroutines.WithLabelValues(instance).Set(float64(count))
}

// UpdateUsersTotal 更新用户总数
func (m *Metrics) UpdateUsersTotal(status string, count int) {
	m.UsersTotal.WithLabelValues(status).Set(float64(count))
}

// UpdateActiveUsers 更新活跃用户数
func (m *Metrics) UpdateActiveUsers(period string, count int) {
	m.ActiveUsers.WithLabelValues(period).Set(float64(count))
}

// UpdateMonitoringTargets 更新监控目标数
func (m *Metrics) UpdateMonitoringTargets(targetType, enabled string, count int) {
	m.MonitoringTargets.WithLabelValues(targetType, enabled).Set(float64(count))
}

// UpdateDashboards 更新仪表板数
func (m *Metrics) UpdateDashboards(dashboardType string, count int) {
	m.DashboardsTotal.WithLabelValues(dashboardType).Set(float64(count))
}

// GetKeyPattern 获取缓存键模式
func GetKeyPattern(key string) string {
	// 简化的键模式提取，实际应用中可能需要更复杂的逻辑
	if len(key) > 20 {
		return key[:20] + "..."
	}
	return key
}