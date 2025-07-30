package middleware

import (
	"net/http"
	"time"

	"ai-monitor/internal/auth"
	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 全局变量用于存储依赖
var (
	globalJWTManager   *auth.JWTManager
	globalCacheManager *cache.CacheManager
	globalConfig       *config.Config
)

// Prometheus指标
var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
)

// InitializeMiddleware 初始化中间件依赖
func InitializeMiddleware(jwtManager *auth.JWTManager, cacheManager *cache.CacheManager, cfg *config.Config) {
	globalJWTManager = jwtManager
	globalCacheManager = cacheManager
	globalConfig = cfg
}

// Auth 认证中间件（简化版）
func Auth() gin.HandlerFunc {
	if globalJWTManager == nil {
		// 如果没有初始化，返回一个空的中间件
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return AuthMiddleware(globalJWTManager)
}

// RequireRole 角色验证中间件（简化版）
func RequireRole(roles ...string) gin.HandlerFunc {
	return RoleMiddleware(roles...)
}

// Security 安全头中间件（简化版）
func Security() gin.HandlerFunc {
	return SecurityHeadersMiddleware()
}

// RateLimit 限流中间件（简化版）
func RateLimit() gin.HandlerFunc {
	if globalCacheManager == nil {
		// 如果没有初始化，返回一个空的中间件
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return RateLimitMiddleware(globalCacheManager, 100) // 默认每分钟100次请求
}

// RequestID 请求ID中间件（简化版）
func RequestID() gin.HandlerFunc {
	return RequestIDMiddleware()
}

// Metrics Prometheus指标中间件（简化版）
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start).Seconds()

		// 记录指标
		httpRequestsTotal.WithLabelValues(method, path, http.StatusText(status)).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// Recovery 恢复中间件（简化版）
func Recovery() gin.HandlerFunc {
	return RecoveryMiddleware()
}

// Logger 日志中间件（简化版）
func Logger() gin.HandlerFunc {
	return RequestLoggerMiddleware()
}

// AuditLog 审计日志中间件（简化版）
func AuditLog() gin.HandlerFunc {
	return AuditLogMiddleware()
}

// CORS CORS中间件（简化版）
func CORS() gin.HandlerFunc {
	if globalConfig != nil && globalConfig.Security.CORS.AllowedOrigins != nil {
		return CORSMiddleware(&globalConfig.Security.CORS)
	}
	// 默认CORS配置
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "43200")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Timeout 超时中间件（简化版）
func Timeout(duration time.Duration) gin.HandlerFunc {
	return TimeoutMiddleware(duration)
}

// OptionalAuth 可选认证中间件（简化版）
func OptionalAuth() gin.HandlerFunc {
	if globalJWTManager == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return OptionalAuthMiddleware(globalJWTManager)
}

// ValidateJSON JSON验证中间件（简化版）
func ValidateJSON() gin.HandlerFunc {
	return ValidateJSONMiddleware()
}

// IPWhitelist IP白名单中间件（简化版）
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return IPWhitelistMiddleware(allowedIPs)
}

// MaintenanceMode 维护模式中间件（简化版）
func MaintenanceMode(enabled bool) gin.HandlerFunc {
	return MaintenanceModeMiddleware(enabled)
}