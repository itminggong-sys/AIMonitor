package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai-monitor/internal/auth"
	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"
	"ai-monitor/internal/database"
	"ai-monitor/internal/models"
	"ai-monitor/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// 提取令牌
		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  "INVALID_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := jwtManager.VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// 检查是否为访问令牌
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token type",
				"code":  "INVALID_TOKEN_TYPE",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Set("user_claims", claims)

		c.Next()
	}
}

// APIKeyOrJWTAuthMiddleware 支持API Key或JWT认证的中间件
func APIKeyOrJWTAuthMiddleware(jwtManager *auth.JWTManager, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// 检查是否为Bearer token格式
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			
			// 首先尝试作为JWT验证
			claims, err := jwtManager.VerifyToken(token)
			if err == nil && claims.TokenType == "access" {
				// JWT认证成功
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("email", claims.Email)
				c.Set("roles", claims.Roles)
				c.Set("permissions", claims.Permissions)
				c.Set("user_claims", claims)
				c.Set("auth_type", "jwt")
				c.Next()
				return
			}
			
			// JWT验证失败，尝试作为API Key验证
			apiKeyService := services.NewAPIKeyService(db)
			apiKey, err := apiKeyService.ValidateAPIKey(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token or API key",
					"code":  "INVALID_AUTH",
				})
				c.Abort()
				return
			}
			
			// API Key认证成功
			c.Set("api_key_id", apiKey.ID)
			c.Set("api_key_name", apiKey.Name)
			c.Set("created_by", apiKey.CreatedBy)
			c.Set("auth_type", "api_key")
			
			// 更新API Key使用统计
			go func() {
				apiKeyService.UpdateUsage(apiKey.ID)
			}()
			
			c.Next()
			return
		}
		
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid authorization header format",
			"code":  "INVALID_AUTH_HEADER",
		})
		c.Abort()
	}
}

// PermissionMiddleware 权限验证中间件
func PermissionMiddleware(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户声明
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.UserClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user claims",
				"code":  "INVALID_CLAIMS",
			})
			c.Abort()
			return
		}

		// 管理员拥有所有权限
		if userClaims.IsAdmin() {
			c.Next()
			return
		}

		// 检查是否有任意一个所需权限
		if !userClaims.HasAnyPermission(requiredPermissions...) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "INSUFFICIENT_PERMISSIONS",
				"required_permissions": requiredPermissions,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleMiddleware 角色验证中间件
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户声明
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.UserClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user claims",
				"code":  "INVALID_CLAIMS",
			})
			c.Abort()
			return
		}

		// 检查是否有任意一个所需角色
		if !userClaims.HasAnyRole(requiredRoles...) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient role permissions",
				"code":  "INSUFFICIENT_ROLE_PERMISSIONS",
				"required_roles": requiredRoles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnlyMiddleware 仅管理员中间件
func AdminOnlyMiddleware() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// CORSMiddleware CORS中间件
func CORSMiddleware(cfg *config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 检查是否允许该来源
		allowed := false
		for _, allowedOrigin := range cfg.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
		c.Header("Access-Control-Max-Age", strconv.Itoa(int(cfg.MaxAge.Seconds())))

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLoggerMiddleware 请求日志中间件
func RequestLoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// AuditLogMiddleware 审计日志中间件
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 获取用户信息
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		c.Next()

		// 记录审计日志
		go func() {
			auditLog := models.AuditLog{
				Action:     fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
				Resource:   extractResourceFromPath(c.Request.URL.Path),
				ResourceID: c.Param("id"),
				IPAddress:  c.ClientIP(),
				UserAgent:  c.Request.UserAgent(),
				Status:     getStatusFromCode(c.Writer.Status()),
			}

			if userID != nil {
				if uid, ok := userID.(uuid.UUID); ok {
					auditLog.UserID = uid
				}
			}

			// 添加详细信息
			details := map[string]interface{}{
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"query":       c.Request.URL.RawQuery,
				"status_code": c.Writer.Status(),
				"latency":     time.Since(start).String(),
				"user_agent":  c.Request.UserAgent(),
				"username":    username,
			}

			if detailsJSON, err := json.Marshal(details); err == nil {
				auditLog.Details = string(detailsJSON)
			}

			// 保存到数据库
			if database.DB != nil {
				database.DB.Create(&auditLog)
			}
		}()
	}
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(cacheManager *cache.CacheManager, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID或IP地址作为限制键
		var key string
		if userID, exists := c.Get("user_id"); exists {
			key = fmt.Sprintf("rate_limit:user:%v", userID)
		} else {
			key = fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
		}

		ctx := context.Background()
		
		// 获取当前请求数
		var currentRequests int
		err := cacheManager.Get(ctx, key, &currentRequests)
		if err != nil {
			// 键不存在，设置为1
			cacheManager.Set(ctx, key, 1, time.Minute)
			c.Next()
			return
		}

		requests := currentRequests
		if requests >= requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
				"limit": requestsPerMinute,
			})
			c.Abort()
			return
		}

		// 增加请求计数
		cacheManager.Increment(ctx, key)
		c.Next()
	}
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// 支持Google Fonts、内联样式和CSS导入的CSP策略
		c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://fonts.gstatic.com; font-src 'self' https://fonts.gstatic.com; script-src 'self' 'unsafe-inline'; connect-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logrus.WithFields(logrus.Fields{
			"error":      recovered,
			"request_id": c.GetString("request_id"),
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"ip":         c.ClientIP(),
		}).Error("Panic recovered")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	})
}

// HealthCheckMiddleware 健康检查中间件
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().UTC(),
				"version":   "1.0.0",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 辅助函数

// extractResourceFromPath 从路径中提取资源名称
func extractResourceFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 1 {
		return parts[1] // 返回第二部分作为资源名称
	}
	return "unknown"
}

// getStatusFromCode 根据状态码获取状态
func getStatusFromCode(code int) string {
	if code >= 200 && code < 300 {
		return "success"
	}
	return "failed"
}



// OptionalAuthMiddleware 可选认证中间件（不强制要求认证）
func OptionalAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.Next()
			return
		}

		claims, err := jwtManager.VerifyToken(token)
		if err != nil {
			c.Next()
			return
		}

		if claims.TokenType == "access" {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("roles", claims.Roles)
			c.Set("permissions", claims.Permissions)
			c.Set("user_claims", claims)
		}

		c.Next()
	}
}

// ValidateJSONMiddleware JSON验证中间件
func ValidateJSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				var jsonData interface{}
				if err := c.ShouldBindJSON(&jsonData); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid JSON format",
						"code":  "INVALID_JSON",
						"details": err.Error(),
					})
					c.Abort()
					return
				}
				// 重新设置请求体，因为ShouldBindJSON会消费它
				c.Set("json_data", jsonData)
			}
		}
		c.Next()
	}
}

// IPWhitelistMiddleware IP白名单中间件
func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// 检查IP是否在白名单中
		allowed := false
		for _, ip := range allowedIPs {
			if ip == clientIP || ip == "*" {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "IP address not allowed",
				"code":  "IP_NOT_ALLOWED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// MaintenanceModeMiddleware 维护模式中间件
func MaintenanceModeMiddleware(isMaintenanceMode bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isMaintenanceMode {
			// 允许健康检查和管理员访问
			if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/admin" {
				c.Next()
				return
			}

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "System is under maintenance",
				"code":  "MAINTENANCE_MODE",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}