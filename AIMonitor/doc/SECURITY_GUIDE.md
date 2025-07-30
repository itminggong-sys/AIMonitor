# AI智能监控系统安全指南

## 概述

本文档详细说明AI智能监控系统的企业级安全架构、安全策略、威胁防护和安全最佳实践。系统基于React+Go技术栈，集成OpenAI和Claude AI服务，采用零信任安全模型，确保系统在各种安全威胁下的稳定运行。

### 安全特性概览
- **零信任架构**: 默认不信任，持续验证
- **多层防护**: 7层安全防护体系
- **AI安全**: 专门的AI服务安全防护
- **实时监控**: 24/7安全态势感知
- **自动响应**: 智能威胁检测和响应
- **合规支持**: 满足SOC2、ISO27001等标准

## 安全架构

### 多层安全防护

```
┌─────────────────────────────────────────────────────────────┐
│                        用户层                                │
├─────────────────────────────────────────────────────────────┤
│  Web应用防火墙(WAF) │ DDoS防护 │ CDN安全 │ SSL/TLS加密      │
├─────────────────────────────────────────────────────────────┤
│                      应用层安全                              │
│  身份认证 │ 授权控制 │ 输入验证 │ 输出编码 │ 会话管理        │
├─────────────────────────────────────────────────────────────┤
│                      服务层安全                              │
│  API网关 │ 服务间认证 │ 限流控制 │ 审计日志 │ 加密通信       │
├─────────────────────────────────────────────────────────────┤
│                      数据层安全                              │
│  数据加密 │ 访问控制 │ 备份加密 │ 数据脱敏 │ 完整性校验     │
├─────────────────────────────────────────────────────────────┤
│                    基础设施安全                              │
│  网络隔离 │ 主机加固 │ 容器安全 │ 密钥管理 │ 安全监控       │
└─────────────────────────────────────────────────────────────┘
```

### 安全组件架构

```go
// internal/security/security.go
package security

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "time"
)

// SecurityManager 安全管理器
type SecurityManager struct {
    authenticator *Authenticator
    authorizer    *Authorizer
    encryptor     *Encryptor
    auditor       *Auditor
    validator     *InputValidator
    rateLimiter   *RateLimiter
}

func NewSecurityManager() *SecurityManager {
    return &SecurityManager{
        authenticator: NewAuthenticator(),
        authorizer:    NewAuthorizer(),
        encryptor:     NewEncryptor(),
        auditor:       NewAuditor(),
        validator:     NewInputValidator(),
        rateLimiter:   NewRateLimiter(),
    }
}

// SecurityContext 安全上下文
type SecurityContext struct {
    UserID      string            `json:"user_id"`
    SessionID   string            `json:"session_id"`
    IP          string            `json:"ip"`
    UserAgent   string            `json:"user_agent"`
    Permissions []string          `json:"permissions"`
    Metadata    map[string]string `json:"metadata"`
    Timestamp   time.Time         `json:"timestamp"`
}

// ValidateSecurityContext 验证安全上下文
func (sm *SecurityManager) ValidateSecurityContext(ctx context.Context, secCtx *SecurityContext) error {
    // 验证会话
    if err := sm.authenticator.ValidateSession(ctx, secCtx.SessionID); err != nil {
        return err
    }
    
    // 检查权限
    if err := sm.authorizer.CheckPermissions(ctx, secCtx.UserID, secCtx.Permissions); err != nil {
        return err
    }
    
    // 限流检查
    if err := sm.rateLimiter.CheckLimit(ctx, secCtx.IP, secCtx.UserID); err != nil {
        return err
    }
    
    // 记录审计日志
    sm.auditor.LogAccess(ctx, secCtx)
    
    return nil
}
```

## 身份认证与授权

### JWT认证实现

```go
// internal/security/auth.go
package security

import (
    "crypto/rsa"
    "errors"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

type Authenticator struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
    issuer     string
    expiry     time.Duration
}

type Claims struct {
    UserID      string   `json:"user_id"`
    Username    string   `json:"username"`
    Email       string   `json:"email"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    SessionID   string   `json:"session_id"`
    jwt.RegisteredClaims
}

func NewAuthenticator(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *Authenticator {
    return &Authenticator{
        privateKey: privateKey,
        publicKey:  publicKey,
        issuer:     "aimonitor",
        expiry:     24 * time.Hour,
    }
}

// GenerateToken 生成JWT令牌
func (a *Authenticator) GenerateToken(user *User) (string, error) {
    now := time.Now()
    sessionID := generateSessionID()
    
    claims := &Claims{
        UserID:      user.ID,
        Username:    user.Username,
        Email:       user.Email,
        Roles:       user.Roles,
        Permissions: user.Permissions,
        SessionID:   sessionID,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    a.issuer,
            Subject:   user.ID,
            Audience:  []string{"aimonitor-api"},
            ExpiresAt: jwt.NewNumericDate(now.Add(a.expiry)),
            NotBefore: jwt.NewNumericDate(now),
            IssuedAt:  jwt.NewNumericDate(now),
            ID:        sessionID,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(a.privateKey)
}

// ValidateToken 验证JWT令牌
func (a *Authenticator) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, errors.New("无效的签名方法")
        }
        return a.publicKey, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("无效的令牌")
}

// RefreshToken 刷新令牌
func (a *Authenticator) RefreshToken(tokenString string) (string, error) {
    claims, err := a.ValidateToken(tokenString)
    if err != nil {
        return "", err
    }
    
    // 检查令牌是否即将过期（剩余时间少于1小时）
    if time.Until(claims.ExpiresAt.Time) > time.Hour {
        return "", errors.New("令牌尚未到刷新时间")
    }
    
    // 生成新令牌
    user := &User{
        ID:          claims.UserID,
        Username:    claims.Username,
        Email:       claims.Email,
        Roles:       claims.Roles,
        Permissions: claims.Permissions,
    }
    
    return a.GenerateToken(user)
}

func generateSessionID() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}
```

### RBAC权限控制

```go
// internal/security/rbac.go
package security

import (
    "context"
    "fmt"
    "strings"
)

type Authorizer struct {
    roleService       *RoleService
    permissionService *PermissionService
}

type Role struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Permissions []string `json:"permissions"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Permission struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Resource    string `json:"resource"`
    Action      string `json:"action"`
    Description string `json:"description"`
}

func NewAuthorizer(roleService *RoleService, permissionService *PermissionService) *Authorizer {
    return &Authorizer{
        roleService:       roleService,
        permissionService: permissionService,
    }
}

// CheckPermission 检查单个权限
func (a *Authorizer) CheckPermission(ctx context.Context, userID, resource, action string) error {
    userRoles, err := a.roleService.GetUserRoles(ctx, userID)
    if err != nil {
        return err
    }
    
    for _, role := range userRoles {
        rolePermissions, err := a.permissionService.GetRolePermissions(ctx, role.ID)
        if err != nil {
            continue
        }
        
        for _, perm := range rolePermissions {
            if a.matchPermission(perm, resource, action) {
                return nil
            }
        }
    }
    
    return fmt.Errorf("用户 %s 没有权限执行 %s:%s", userID, resource, action)
}

// CheckPermissions 检查多个权限
func (a *Authorizer) CheckPermissions(ctx context.Context, userID string, requiredPerms []string) error {
    userPermissions, err := a.getUserAllPermissions(ctx, userID)
    if err != nil {
        return err
    }
    
    for _, requiredPerm := range requiredPerms {
        if !a.hasPermission(userPermissions, requiredPerm) {
            return fmt.Errorf("用户 %s 缺少权限: %s", userID, requiredPerm)
        }
    }
    
    return nil
}

// matchPermission 匹配权限
func (a *Authorizer) matchPermission(perm *Permission, resource, action string) bool {
    // 支持通配符匹配
    resourceMatch := perm.Resource == "*" || perm.Resource == resource
    actionMatch := perm.Action == "*" || perm.Action == action
    
    return resourceMatch && actionMatch
}

// getUserAllPermissions 获取用户所有权限
func (a *Authorizer) getUserAllPermissions(ctx context.Context, userID string) ([]string, error) {
    userRoles, err := a.roleService.GetUserRoles(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    permissionSet := make(map[string]bool)
    
    for _, role := range userRoles {
        rolePermissions, err := a.permissionService.GetRolePermissions(ctx, role.ID)
        if err != nil {
            continue
        }
        
        for _, perm := range rolePermissions {
            permKey := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
            permissionSet[permKey] = true
        }
    }
    
    permissions := make([]string, 0, len(permissionSet))
    for perm := range permissionSet {
        permissions = append(permissions, perm)
    }
    
    return permissions, nil
}

// hasPermission 检查是否拥有权限
func (a *Authorizer) hasPermission(userPermissions []string, requiredPerm string) bool {
    for _, perm := range userPermissions {
        if perm == requiredPerm || perm == "*:*" {
            return true
        }
        
        // 检查通配符匹配
        parts := strings.Split(requiredPerm, ":")
        if len(parts) == 2 {
            resource, action := parts[0], parts[1]
            if perm == fmt.Sprintf("%s:*", resource) || perm == fmt.Sprintf("*:%s", action) {
                return true
            }
        }
    }
    
    return false
}
```

## 数据加密与保护

### 数据加密实现

```go
// internal/security/encryption.go
package security

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "io"
    
    "golang.org/x/crypto/pbkdf2"
)

type Encryptor struct {
    key []byte
}

func NewEncryptor(password string, salt []byte) *Encryptor {
    key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
    return &Encryptor{key: key}
}

// Encrypt 加密数据
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    // 生成随机IV
    ciphertext := make([]byte, aes.BlockSize+len(plaintext))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }
    
    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))
    
    return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
    data, err := base64.URLEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }
    
    if len(data) < aes.BlockSize {
        return "", errors.New("密文太短")
    }
    
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    iv := data[:aes.BlockSize]
    data = data[aes.BlockSize:]
    
    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(data, data)
    
    return string(data), nil
}

// EncryptSensitiveData 加密敏感数据
func (e *Encryptor) EncryptSensitiveData(data map[string]interface{}) (map[string]interface{}, error) {
    sensitiveFields := []string{"password", "token", "secret", "key", "credential"}
    
    result := make(map[string]interface{})
    
    for key, value := range data {
        if e.isSensitiveField(key, sensitiveFields) {
            if strValue, ok := value.(string); ok {
                encrypted, err := e.Encrypt(strValue)
                if err != nil {
                    return nil, err
                }
                result[key] = encrypted
            } else {
                result[key] = value
            }
        } else {
            result[key] = value
        }
    }
    
    return result, nil
}

// isSensitiveField 检查是否为敏感字段
func (e *Encryptor) isSensitiveField(fieldName string, sensitiveFields []string) bool {
    fieldLower := strings.ToLower(fieldName)
    for _, sensitive := range sensitiveFields {
        if strings.Contains(fieldLower, sensitive) {
            return true
        }
    }
    return false
}
```

### 数据脱敏

```go
// internal/security/masking.go
package security

import (
    "regexp"
    "strings"
)

type DataMasker struct {
    patterns map[string]*regexp.Regexp
}

func NewDataMasker() *DataMasker {
    dm := &DataMasker{
        patterns: make(map[string]*regexp.Regexp),
    }
    
    // 预定义脱敏模式
    dm.patterns["email"] = regexp.MustCompile(`([a-zA-Z0-9._%+-]+)@([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
    dm.patterns["phone"] = regexp.MustCompile(`(\d{3})(\d{4})(\d{4})`)
    dm.patterns["idcard"] = regexp.MustCompile(`(\d{6})(\d{8})(\d{4})`)
    dm.patterns["bankcard"] = regexp.MustCompile(`(\d{4})(\d{8,12})(\d{4})`)
    dm.patterns["ip"] = regexp.MustCompile(`(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})`)
    
    return dm
}

// MaskEmail 脱敏邮箱
func (dm *DataMasker) MaskEmail(email string) string {
    return dm.patterns["email"].ReplaceAllStringFunc(email, func(match string) string {
        parts := strings.Split(match, "@")
        if len(parts) != 2 {
            return match
        }
        
        username := parts[0]
        domain := parts[1]
        
        if len(username) <= 2 {
            return "***@" + domain
        }
        
        masked := username[:1] + strings.Repeat("*", len(username)-2) + username[len(username)-1:]
        return masked + "@" + domain
    })
}

// MaskPhone 脱敏手机号
func (dm *DataMasker) MaskPhone(phone string) string {
    return dm.patterns["phone"].ReplaceAllString(phone, "$1****$3")
}

// MaskIDCard 脱敏身份证
func (dm *DataMasker) MaskIDCard(idcard string) string {
    return dm.patterns["idcard"].ReplaceAllString(idcard, "$1********$3")
}

// MaskBankCard 脱敏银行卡
func (dm *DataMasker) MaskBankCard(bankcard string) string {
    return dm.patterns["bankcard"].ReplaceAllString(bankcard, "$1****$3")
}

// MaskIP 脱敏IP地址
func (dm *DataMasker) MaskIP(ip string) string {
    return dm.patterns["ip"].ReplaceAllString(ip, "$1.$2.*.*")
}

// MaskData 通用数据脱敏
func (dm *DataMasker) MaskData(data map[string]interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    
    for key, value := range data {
        if strValue, ok := value.(string); ok {
            switch {
            case strings.Contains(strings.ToLower(key), "email"):
                result[key] = dm.MaskEmail(strValue)
            case strings.Contains(strings.ToLower(key), "phone"):
                result[key] = dm.MaskPhone(strValue)
            case strings.Contains(strings.ToLower(key), "idcard") || strings.Contains(strings.ToLower(key), "id_card"):
                result[key] = dm.MaskIDCard(strValue)
            case strings.Contains(strings.ToLower(key), "bankcard") || strings.Contains(strings.ToLower(key), "bank_card"):
                result[key] = dm.MaskBankCard(strValue)
            case strings.Contains(strings.ToLower(key), "ip"):
                result[key] = dm.MaskIP(strValue)
            default:
                result[key] = value
            }
        } else {
            result[key] = value
        }
    }
    
    return result
}
```

## 输入验证与防护

### 输入验证器

```go
// internal/security/validator.go
package security

import (
    "errors"
    "fmt"
    "net/url"
    "regexp"
    "strings"
    "unicode"
)

type InputValidator struct {
    sqlInjectionPatterns []string
    xssPatterns         []string
    pathTraversalPattern *regexp.Regexp
}

func NewInputValidator() *InputValidator {
    return &InputValidator{
        sqlInjectionPatterns: []string{
            `(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`,
            `(?i)(script|javascript|vbscript|onload|onerror|onclick)`,
            `(?i)(or|and)\s+\d+\s*=\s*\d+`,
            `(?i)'\s*(or|and)\s*'`,
            `(?i)--`,
            `(?i)/\*.*\*/`,
        },
        xssPatterns: []string{
            `<script[^>]*>.*?</script>`,
            `javascript:`,
            `vbscript:`,
            `onload\s*=`,
            `onerror\s*=`,
            `onclick\s*=`,
            `<iframe[^>]*>`,
            `<object[^>]*>`,
            `<embed[^>]*>`,
        },
        pathTraversalPattern: regexp.MustCompile(`\.\./|\.\\`),
    }
}

// ValidateInput 验证输入
func (iv *InputValidator) ValidateInput(input string, validationType string) error {
    switch validationType {
    case "sql":
        return iv.ValidateSQL(input)
    case "xss":
        return iv.ValidateXSS(input)
    case "path":
        return iv.ValidatePath(input)
    case "email":
        return iv.ValidateEmail(input)
    case "url":
        return iv.ValidateURL(input)
    default:
        return iv.ValidateGeneral(input)
    }
}

// ValidateSQL 检查SQL注入
func (iv *InputValidator) ValidateSQL(input string) error {
    for _, pattern := range iv.sqlInjectionPatterns {
        matched, _ := regexp.MatchString(pattern, input)
        if matched {
            return errors.New("检测到潜在的SQL注入攻击")
        }
    }
    return nil
}

// ValidateXSS 检查XSS攻击
func (iv *InputValidator) ValidateXSS(input string) error {
    for _, pattern := range iv.xssPatterns {
        matched, _ := regexp.MatchString(`(?i)`+pattern, input)
        if matched {
            return errors.New("检测到潜在的XSS攻击")
        }
    }
    return nil
}

// ValidatePath 检查路径遍历
func (iv *InputValidator) ValidatePath(input string) error {
    if iv.pathTraversalPattern.MatchString(input) {
        return errors.New("检测到潜在的路径遍历攻击")
    }
    return nil
}

// ValidateEmail 验证邮箱格式
func (iv *InputValidator) ValidateEmail(email string) error {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return errors.New("无效的邮箱格式")
    }
    return nil
}

// ValidateURL 验证URL格式
func (iv *InputValidator) ValidateURL(urlStr string) error {
    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("无效的URL格式: %v", err)
    }
    
    if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
        return errors.New("URL必须使用HTTP或HTTPS协议")
    }
    
    return nil
}

// ValidateGeneral 通用输入验证
func (iv *InputValidator) ValidateGeneral(input string) error {
    // 检查长度
    if len(input) > 10000 {
        return errors.New("输入长度超过限制")
    }
    
    // 检查是否包含控制字符
    for _, r := range input {
        if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
            return errors.New("输入包含非法控制字符")
        }
    }
    
    // 检查SQL注入
    if err := iv.ValidateSQL(input); err != nil {
        return err
    }
    
    // 检查XSS
    if err := iv.ValidateXSS(input); err != nil {
        return err
    }
    
    // 检查路径遍历
    if err := iv.ValidatePath(input); err != nil {
        return err
    }
    
    return nil
}

// SanitizeInput 清理输入
func (iv *InputValidator) SanitizeInput(input string) string {
    // 移除HTML标签
    htmlRegex := regexp.MustCompile(`<[^>]*>`)
    sanitized := htmlRegex.ReplaceAllString(input, "")
    
    // 转义特殊字符
    sanitized = strings.ReplaceAll(sanitized, "<", "&lt;")
    sanitized = strings.ReplaceAll(sanitized, ">", "&gt;")
    sanitized = strings.ReplaceAll(sanitized, "&", "&amp;")
    sanitized = strings.ReplaceAll(sanitized, "\"", "&quot;")
    sanitized = strings.ReplaceAll(sanitized, "'", "&#x27;")
    
    return sanitized
}
```

## 限流与防护

### 限流器实现

```go
// internal/security/ratelimit.go
package security

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    
    // 配置
    defaultRate  rate.Limit
    defaultBurst int
    cleanupInterval time.Duration
}

type RateLimitConfig struct {
    RequestsPerSecond int           `yaml:"requests_per_second"`
    BurstSize         int           `yaml:"burst_size"`
    WindowSize        time.Duration `yaml:"window_size"`
    CleanupInterval   time.Duration `yaml:"cleanup_interval"`
}

func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
    rl := &RateLimiter{
        limiters:        make(map[string]*rate.Limiter),
        defaultRate:     rate.Limit(config.RequestsPerSecond),
        defaultBurst:    config.BurstSize,
        cleanupInterval: config.CleanupInterval,
    }
    
    // 启动清理协程
    go rl.cleanup()
    
    return rl
}

// CheckLimit 检查限流
func (rl *RateLimiter) CheckLimit(ctx context.Context, ip, userID string) error {
    // IP限流
    if err := rl.checkIPLimit(ip); err != nil {
        return err
    }
    
    // 用户限流
    if userID != "" {
        if err := rl.checkUserLimit(userID); err != nil {
            return err
        }
    }
    
    return nil
}

// checkIPLimit 检查IP限流
func (rl *RateLimiter) checkIPLimit(ip string) error {
    key := fmt.Sprintf("ip:%s", ip)
    limiter := rl.getLimiter(key)
    
    if !limiter.Allow() {
        return fmt.Errorf("IP %s 请求频率过高", ip)
    }
    
    return nil
}

// checkUserLimit 检查用户限流
func (rl *RateLimiter) checkUserLimit(userID string) error {
    key := fmt.Sprintf("user:%s", userID)
    limiter := rl.getLimiter(key)
    
    if !limiter.Allow() {
        return fmt.Errorf("用户 %s 请求频率过高", userID)
    }
    
    return nil
}

// getLimiter 获取限流器
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
    rl.mu.RLock()
    limiter, exists := rl.limiters[key]
    rl.mu.RUnlock()
    
    if exists {
        return limiter
    }
    
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    // 双重检查
    if limiter, exists := rl.limiters[key]; exists {
        return limiter
    }
    
    // 创建新的限流器
    limiter = rate.NewLimiter(rl.defaultRate, rl.defaultBurst)
    rl.limiters[key] = limiter
    
    return limiter
}

// cleanup 清理过期的限流器
func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(rl.cleanupInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        rl.mu.Lock()
        
        // 清理长时间未使用的限流器
        for key, limiter := range rl.limiters {
            // 如果限流器的令牌桶已满，说明长时间未使用
            if limiter.Tokens() == float64(rl.defaultBurst) {
                delete(rl.limiters, key)
            }
        }
        
        rl.mu.Unlock()
    }
}
```

### DDoS防护

```go
// internal/security/ddos.go
package security

import (
    "context"
    "sync"
    "time"
)

type DDoSProtector struct {
    ipStats     map[string]*IPStats
    mu          sync.RWMutex
    threshold   int
    timeWindow  time.Duration
    blockTime   time.Duration
    blockedIPs  map[string]time.Time
}

type IPStats struct {
    RequestCount int
    FirstRequest time.Time
    LastRequest  time.Time
    Blocked      bool
}

func NewDDoSProtector(threshold int, timeWindow, blockTime time.Duration) *DDoSProtector {
    ddos := &DDoSProtector{
        ipStats:    make(map[string]*IPStats),
        threshold:  threshold,
        timeWindow: timeWindow,
        blockTime:  blockTime,
        blockedIPs: make(map[string]time.Time),
    }
    
    // 启动清理协程
    go ddos.cleanup()
    
    return ddos
}

// CheckRequest 检查请求是否被阻止
func (ddos *DDoSProtector) CheckRequest(ctx context.Context, ip string) error {
    ddos.mu.Lock()
    defer ddos.mu.Unlock()
    
    now := time.Now()
    
    // 检查IP是否被阻止
    if blockTime, blocked := ddos.blockedIPs[ip]; blocked {
        if now.Sub(blockTime) < ddos.blockTime {
            return fmt.Errorf("IP %s 被临时阻止", ip)
        }
        // 解除阻止
        delete(ddos.blockedIPs, ip)
        delete(ddos.ipStats, ip)
    }
    
    // 获取或创建IP统计
    stats, exists := ddos.ipStats[ip]
    if !exists {
        stats = &IPStats{
            RequestCount: 0,
            FirstRequest: now,
            LastRequest:  now,
        }
        ddos.ipStats[ip] = stats
    }
    
    // 检查时间窗口
    if now.Sub(stats.FirstRequest) > ddos.timeWindow {
        // 重置统计
        stats.RequestCount = 0
        stats.FirstRequest = now
    }
    
    stats.RequestCount++
    stats.LastRequest = now
    
    // 检查是否超过阈值
    if stats.RequestCount > ddos.threshold {
        stats.Blocked = true
        ddos.blockedIPs[ip] = now
        return fmt.Errorf("IP %s 请求频率过高，已被阻止", ip)
    }
    
    return nil
}

// cleanup 清理过期数据
func (ddos *DDoSProtector) cleanup() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        ddos.mu.Lock()
        
        now := time.Now()
        
        // 清理过期的IP统计
        for ip, stats := range ddos.ipStats {
            if now.Sub(stats.LastRequest) > ddos.timeWindow*2 {
                delete(ddos.ipStats, ip)
            }
        }
        
        // 清理过期的阻止记录
        for ip, blockTime := range ddos.blockedIPs {
            if now.Sub(blockTime) > ddos.blockTime {
                delete(ddos.blockedIPs, ip)
            }
        }
        
        ddos.mu.Unlock()
    }
}

// GetBlockedIPs 获取被阻止的IP列表
func (ddos *DDoSProtector) GetBlockedIPs() []string {
    ddos.mu.RLock()
    defer ddos.mu.RUnlock()
    
    var blockedIPs []string
    for ip := range ddos.blockedIPs {
        blockedIPs = append(blockedIPs, ip)
    }
    
    return blockedIPs
}

// UnblockIP 解除IP阻止
func (ddos *DDoSProtector) UnblockIP(ip string) {
    ddos.mu.Lock()
    defer ddos.mu.Unlock()
    
    delete(ddos.blockedIPs, ip)
    delete(ddos.ipStats, ip)
}
```

## 安全审计

### 审计日志系统

```go
// internal/security/audit.go
package security

import (
    "context"
    "encoding/json"
    "time"
)

type Auditor struct {
    logger AuditLogger
    db     AuditDB
}

type AuditEvent struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    Category    string                 `json:"category"`
    UserID      string                 `json:"user_id"`
    Username    string                 `json:"username"`
    IP          string                 `json:"ip"`
    UserAgent   string                 `json:"user_agent"`
    Resource    string                 `json:"resource"`
    Action      string                 `json:"action"`
    Result      string                 `json:"result"`
    Details     map[string]interface{} `json:"details"`
    Timestamp   time.Time              `json:"timestamp"`
    RequestID   string                 `json:"request_id"`
    SessionID   string                 `json:"session_id"`
}

func NewAuditor(logger AuditLogger, db AuditDB) *Auditor {
    return &Auditor{
        logger: logger,
        db:     db,
    }
}

// LogAccess 记录访问日志
func (a *Auditor) LogAccess(ctx context.Context, secCtx *SecurityContext) {
    event := AuditEvent{
        ID:        generateEventID(),
        Type:      "access",
        Category:  "authentication",
        UserID:    secCtx.UserID,
        IP:        secCtx.IP,
        UserAgent: secCtx.UserAgent,
        Action:    "access",
        Result:    "success",
        Timestamp: time.Now(),
        SessionID: secCtx.SessionID,
    }
    
    a.logEvent(ctx, event)
}

// LogLogin 记录登录日志
func (a *Auditor) LogLogin(ctx context.Context, userID, username, ip, userAgent string, success bool) {
    result := "success"
    if !success {
        result = "failure"
    }
    
    event := AuditEvent{
        ID:        generateEventID(),
        Type:      "login",
        Category:  "authentication",
        UserID:    userID,
        Username:  username,
        IP:        ip,
        UserAgent: userAgent,
        Action:    "login",
        Result:    result,
        Timestamp: time.Now(),
    }
    
    a.logEvent(ctx, event)
}

// LogOperation 记录操作日志
func (a *Auditor) LogOperation(ctx context.Context, userID, resource, action string, details map[string]interface{}, success bool) {
    result := "success"
    if !success {
        result = "failure"
    }
    
    event := AuditEvent{
        ID:        generateEventID(),
        Type:      "operation",
        Category:  "business",
        UserID:    userID,
        Resource:  resource,
        Action:    action,
        Result:    result,
        Details:   details,
        Timestamp: time.Now(),
    }
    
    a.logEvent(ctx, event)
}

// LogSecurityEvent 记录安全事件
func (a *Auditor) LogSecurityEvent(ctx context.Context, eventType, description string, details map[string]interface{}) {
    event := AuditEvent{
        ID:        generateEventID(),
        Type:      eventType,
        Category:  "security",
        Action:    "security_event",
        Result:    "detected",
        Details:   details,
        Timestamp: time.Now(),
    }
    
    if description != "" {
        if event.Details == nil {
            event.Details = make(map[string]interface{})
        }
        event.Details["description"] = description
    }
    
    a.logEvent(ctx, event)
}

// logEvent 记录事件
func (a *Auditor) logEvent(ctx context.Context, event AuditEvent) {
    // 记录到日志文件
    eventJSON, _ := json.Marshal(event)
    a.logger.Log(string(eventJSON))
    
    // 记录到数据库
    go func() {
        if err := a.db.InsertAuditEvent(context.Background(), event); err != nil {
            a.logger.Log(fmt.Sprintf("Failed to insert audit event: %v", err))
        }
    }()
}

// QueryAuditEvents 查询审计事件
func (a *Auditor) QueryAuditEvents(ctx context.Context, filter AuditFilter) ([]AuditEvent, error) {
    return a.db.QueryAuditEvents(ctx, filter)
}

type AuditFilter struct {
    UserID    string    `json:"user_id"`
    Type      string    `json:"type"`
    Category  string    `json:"category"`
    Action    string    `json:"action"`
    Result    string    `json:"result"`
    StartTime time.Time `json:"start_time"`
    EndTime   time.Time `json:"end_time"`
    Limit     int       `json:"limit"`
    Offset    int       `json:"offset"`
}
```

## 安全配置

### 安全配置文件

```yaml
# config/security.yaml
security:
  # JWT配置
  jwt:
    private_key_path: "/etc/aimonitor/keys/private.pem"
    public_key_path: "/etc/aimonitor/keys/public.pem"
    issuer: "aimonitor"
    expiry: "24h"
    refresh_threshold: "1h"
  
  # 密码策略
  password:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_symbols: true
    max_age_days: 90
    history_count: 5
  
  # 会话管理
  session:
    timeout: "30m"
    max_concurrent: 5
    secure_cookie: true
    same_site: "strict"
  
  # 限流配置
  rate_limit:
    requests_per_second: 100
    burst_size: 200
    window_size: "1m"
    cleanup_interval: "5m"
  
  # DDoS防护
  ddos:
    threshold: 1000
    time_window: "1m"
    block_time: "10m"
  
  # 加密配置
  encryption:
    algorithm: "AES-256-GCM"
    key_rotation_days: 30
    salt_length: 32
  
  # 审计配置
  audit:
    enabled: true
    log_file: "/var/log/aimonitor/audit.log"
    retention_days: 365
    sensitive_fields:
      - "password"
      - "token"
      - "secret"
      - "key"
  
  # CORS配置
  cors:
    allowed_origins:
      - "https://monitor.example.com"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
    allowed_headers:
      - "Authorization"
      - "Content-Type"
    max_age: 3600
  
  # HTTPS配置
  tls:
    cert_file: "/etc/aimonitor/certs/server.crt"
    key_file: "/etc/aimonitor/certs/server.key"
    min_version: "1.2"
    cipher_suites:
      - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
      - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
```

### 安全中间件

```go
// internal/middleware/security.go
package middleware

import (
    "net/http"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
)

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 防止点击劫持
        c.Header("X-Frame-Options", "DENY")
        
        // 防止MIME类型嗅探
        c.Header("X-Content-Type-Options", "nosniff")
        
        // XSS保护
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // 强制HTTPS
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        
        // 内容安全策略
        c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
        
        // 引用策略
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // 权限策略
        c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        c.Next()
    }
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(rateLimiter *security.RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        userID := getUserIDFromContext(c)
        
        if err := rateLimiter.CheckLimit(c.Request.Context(), ip, userID); err != nil {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "请求频率过高",
                "message": err.Error(),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// DDoSProtectionMiddleware DDoS防护中间件
func DDoSProtectionMiddleware(ddosProtector *security.DDoSProtector) gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        
        if err := ddosProtector.CheckRequest(c.Request.Context(), ip); err != nil {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "访问被拒绝",
                "message": err.Error(),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// InputValidationMiddleware 输入验证中间件
func InputValidationMiddleware(validator *security.InputValidator) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 验证查询参数
        for key, values := range c.Request.URL.Query() {
            for _, value := range values {
                if err := validator.ValidateInput(value, "general"); err != nil {
                    c.JSON(http.StatusBadRequest, gin.H{
                        "error": "无效的输入",
                        "field": key,
                        "message": err.Error(),
                    })
                    c.Abort()
                    return
                }
            }
        }
        
        // 验证路径参数
        for _, param := range c.Params {
            if err := validator.ValidateInput(param.Value, "path"); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "无效的路径参数",
                    "field": param.Key,
                    "message": err.Error(),
                })
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}

// AuditMiddleware 审计中间件
func AuditMiddleware(auditor *security.Auditor) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 记录请求信息
        userID := getUserIDFromContext(c)
        ip := c.ClientIP()
        userAgent := c.GetHeader("User-Agent")
        method := c.Request.Method
        path := c.Request.URL.Path
        
        c.Next()
        
        // 记录响应信息
        duration := time.Since(start)
        statusCode := c.Writer.Status()
        
        // 记录审计日志
        details := map[string]interface{}{
            "method":      method,
            "path":        path,
            "status_code": statusCode,
            "duration_ms": duration.Milliseconds(),
        }
        
        success := statusCode < 400
        auditor.LogOperation(c.Request.Context(), userID, path, method, details, success)
    }
}

func getUserIDFromContext(c *gin.Context) string {
    if userID, exists := c.Get("user_id"); exists {
        if uid, ok := userID.(string); ok {
            return uid
        }
    }
    return ""
}
```

## 安全监控

### 安全事件监控

```go
// internal/security/monitor.go
package security

import (
    "context"
    "sync"
    "time"
)

type SecurityMonitor struct {
    eventChan    chan SecurityEvent
    rules        []MonitorRule
    alertManager *AlertManager
    mu           sync.RWMutex
}

type MonitorRule struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    EventType   string        `json:"event_type"`
    Condition   string        `json:"condition"`
    Threshold   int           `json:"threshold"`
    TimeWindow  time.Duration `json:"time_window"`
    Severity    string        `json:"severity"`
    Enabled     bool          `json:"enabled"`
    Actions     []string      `json:"actions"`
}

func NewSecurityMonitor(alertManager *AlertManager) *SecurityMonitor {
    sm := &SecurityMonitor{
        eventChan:    make(chan SecurityEvent, 1000),
        alertManager: alertManager,
    }
    
    // 加载默认规则
    sm.loadDefaultRules()
    
    // 启动监控协程
    go sm.monitor()
    
    return sm
}

// ProcessEvent 处理安全事件
func (sm *SecurityMonitor) ProcessEvent(event SecurityEvent) {
    select {
    case sm.eventChan <- event:
    default:
        // 事件队列已满，丢弃事件
    }
}

// monitor 监控协程
func (sm *SecurityMonitor) monitor() {
    eventBuffer := make(map[string][]SecurityEvent)
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case event := <-sm.eventChan:
            sm.processSecurityEvent(event, eventBuffer)
            
        case <-ticker.C:
            sm.evaluateRules(eventBuffer)
            sm.cleanupBuffer(eventBuffer)
        }
    }
}

// processSecurityEvent 处理安全事件
func (sm *SecurityMonitor) processSecurityEvent(event SecurityEvent, buffer map[string][]SecurityEvent) {
    sm.mu.RLock()
    rules := sm.rules
    sm.mu.RUnlock()
    
    for _, rule := range rules {
        if !rule.Enabled {
            continue
        }
        
        if sm.matchRule(event, rule) {
            key := fmt.Sprintf("%s:%s", rule.ID, event.Type)
            buffer[key] = append(buffer[key], event)
        }
    }
}

// evaluateRules 评估规则
func (sm *SecurityMonitor) evaluateRules(buffer map[string][]SecurityEvent) {
    sm.mu.RLock()
    rules := sm.rules
    sm.mu.RUnlock()
    
    for _, rule := range rules {
        if !rule.Enabled {
            continue
        }
        
        key := fmt.Sprintf("%s:%s", rule.ID, rule.EventType)
        events, exists := buffer[key]
        if !exists {
            continue
        }
        
        // 检查时间窗口内的事件数量
        now := time.Now()
        validEvents := make([]SecurityEvent, 0)
        
        for _, event := range events {
            if now.Sub(event.Timestamp) <= rule.TimeWindow {
                validEvents = append(validEvents, event)
            }
        }
        
        if len(validEvents) >= rule.Threshold {
            sm.triggerAlert(rule, validEvents)
            // 清空已触发的事件
            buffer[key] = nil
        } else {
            buffer[key] = validEvents
        }
    }
}

// triggerAlert 触发告警
func (sm *SecurityMonitor) triggerAlert(rule MonitorRule, events []SecurityEvent) {
    alert := SecurityAlert{
        ID:          generateAlertID(),
        RuleID:      rule.ID,
        RuleName:    rule.Name,
        Severity:    rule.Severity,
        EventCount:  len(events),
        FirstEvent:  events[0].Timestamp,
        LastEvent:   events[len(events)-1].Timestamp,
        Events:      events,
        Timestamp:   time.Now(),
        Status:      "active",
    }
    
    // 发送告警
    sm.alertManager.SendSecurityAlert(alert)
    
    // 执行自动化响应
    for _, action := range rule.Actions {
        sm.executeAction(action, alert)
    }
}

// executeAction 执行自动化响应
func (sm *SecurityMonitor) executeAction(action string, alert SecurityAlert) {
    switch action {
    case "block_ip":
        sm.blockSuspiciousIPs(alert)
    case "disable_user":
        sm.disableSuspiciousUsers(alert)
    case "notify_admin":
        sm.notifyAdministrators(alert)
    case "increase_monitoring":
        sm.increaseMonitoring(alert)
    }
}

// loadDefaultRules 加载默认规则
func (sm *SecurityMonitor) loadDefaultRules() {
    defaultRules := []MonitorRule{
        {
            ID:         "login_failure",
            Name:       "登录失败过多",
            EventType:  "login_failed",
            Threshold:  5,
            TimeWindow: 5 * time.Minute,
            Severity:   "medium",
            Enabled:    true,
            Actions:    []string{"block_ip", "notify_admin"},
        },
        {
            ID:         "brute_force",
            Name:       "暴力破解攻击",
            EventType:  "brute_force",
            Threshold:  10,
            TimeWindow: 1 * time.Minute,
            Severity:   "high",
            Enabled:    true,
            Actions:    []string{"block_ip", "disable_user", "notify_admin"},
        },
        {
            ID:         "privilege_escalation",
            Name:       "权限提升尝试",
            EventType:  "privilege_escalation",
            Threshold:  3,
            TimeWindow: 10 * time.Minute,
            Severity:   "critical",
            Enabled:    true,
            Actions:    []string{"disable_user", "notify_admin", "increase_monitoring"},
        },
        {
            ID:         "data_access_anomaly",
            Name:       "异常数据访问",
            EventType:  "data_access",
            Threshold:  100,
            TimeWindow: 5 * time.Minute,
            Severity:   "medium",
            Enabled:    true,
            Actions:    []string{"notify_admin", "increase_monitoring"},
        },
    }
    
    sm.mu.Lock()
    sm.rules = defaultRules
    sm.mu.Unlock()
}

type SecurityAlert struct {
    ID         string          `json:"id"`
    RuleID     string          `json:"rule_id"`
    RuleName   string          `json:"rule_name"`
    Severity   string          `json:"severity"`
    EventCount int             `json:"event_count"`
    FirstEvent time.Time       `json:"first_event"`
    LastEvent  time.Time       `json:"last_event"`
    Events     []SecurityEvent `json:"events"`
    Timestamp  time.Time       `json:"timestamp"`
    Status     string          `json:"status"`
}
```

## 威胁检测

### 异常行为检测

```go
// internal/security/anomaly.go
package security

import (
    "context"
    "math"
    "time"
)

type AnomalyDetector struct {
    userProfiles map[string]*UserProfile
    threshold    float64
}

type UserProfile struct {
    UserID           string            `json:"user_id"`
    NormalBehavior   BehaviorPattern   `json:"normal_behavior"`
    RecentBehavior   []BehaviorRecord  `json:"recent_behavior"`
    LastUpdated      time.Time         `json:"last_updated"`
    AnomalyScore     float64           `json:"anomaly_score"`
}

type BehaviorPattern struct {
    LoginTimes       []time.Time `json:"login_times"`
    AccessPatterns   []string    `json:"access_patterns"`
    IPAddresses      []string    `json:"ip_addresses"`
    UserAgents       []string    `json:"user_agents"`
    RequestFrequency float64     `json:"request_frequency"`
    SessionDuration  time.Duration `json:"session_duration"`
}

type BehaviorRecord struct {
    Timestamp       time.Time `json:"timestamp"`
    Action          string    `json:"action"`
    Resource        string    `json:"resource"`
    IP              string    `json:"ip"`
    UserAgent       string    `json:"user_agent"`
    SessionDuration time.Duration `json:"session_duration"`
}

func NewAnomalyDetector(threshold float64) *AnomalyDetector {
    return &AnomalyDetector{
        userProfiles: make(map[string]*UserProfile),
        threshold:    threshold,
    }
}

// DetectAnomaly 检测异常行为
func (ad *AnomalyDetector) DetectAnomaly(ctx context.Context, userID string, behavior BehaviorRecord) (bool, float64, error) {
    profile, exists := ad.userProfiles[userID]
    if !exists {
        // 创建新的用户画像
        profile = &UserProfile{
            UserID:         userID,
            RecentBehavior: make([]BehaviorRecord, 0),
            LastUpdated:    time.Now(),
        }
        ad.userProfiles[userID] = profile
    }
    
    // 计算异常分数
    anomalyScore := ad.calculateAnomalyScore(profile, behavior)
    
    // 更新用户画像
    ad.updateUserProfile(profile, behavior)
    
    // 判断是否异常
    isAnomaly := anomalyScore > ad.threshold
    
    return isAnomaly, anomalyScore, nil
}

// calculateAnomalyScore 计算异常分数
func (ad *AnomalyDetector) calculateAnomalyScore(profile *UserProfile, behavior BehaviorRecord) float64 {
    var score float64
    
    // 时间异常检测
    timeScore := ad.calculateTimeAnomalyScore(profile, behavior)
    score += timeScore * 0.3
    
    // IP地址异常检测
    ipScore := ad.calculateIPAnomalyScore(profile, behavior)
    score += ipScore * 0.2
    
    // 用户代理异常检测
    uaScore := ad.calculateUserAgentAnomalyScore(profile, behavior)
    score += uaScore * 0.1
    
    // 访问模式异常检测
    patternScore := ad.calculatePatternAnomalyScore(profile, behavior)
    score += patternScore * 0.4
    
    return score
}

// calculateTimeAnomalyScore 计算时间异常分数
func (ad *AnomalyDetector) calculateTimeAnomalyScore(profile *UserProfile, behavior BehaviorRecord) float64 {
    if len(profile.NormalBehavior.LoginTimes) == 0 {
        return 0
    }
    
    currentHour := behavior.Timestamp.Hour()
    
    // 统计正常登录时间分布
    hourCounts := make(map[int]int)
    for _, loginTime := range profile.NormalBehavior.LoginTimes {
        hour := loginTime.Hour()
        hourCounts[hour]++
    }
    
    // 计算当前时间的正常性
    totalLogins := len(profile.NormalBehavior.LoginTimes)
    currentHourCount := hourCounts[currentHour]
    
    if currentHourCount == 0 {
        return 1.0 // 完全异常
    }
    
    normalProbability := float64(currentHourCount) / float64(totalLogins)
    return 1.0 - normalProbability
}

// calculateIPAnomalyScore 计算IP异常分数
func (ad *AnomalyDetector) calculateIPAnomalyScore(profile *UserProfile, behavior BehaviorRecord) float64 {
    for _, ip := range profile.NormalBehavior.IPAddresses {
        if ip == behavior.IP {
            return 0 // 正常IP
        }
    }
    return 1.0 // 新IP地址
}

// calculateUserAgentAnomalyScore 计算用户代理异常分数
func (ad *AnomalyDetector) calculateUserAgentAnomalyScore(profile *UserProfile, behavior BehaviorRecord) float64 {
    for _, ua := range profile.NormalBehavior.UserAgents {
        if ua == behavior.UserAgent {
            return 0 // 正常用户代理
        }
    }
    return 0.5 // 新用户代理
}

// calculatePatternAnomalyScore 计算访问模式异常分数
func (ad *AnomalyDetector) calculatePatternAnomalyScore(profile *UserProfile, behavior BehaviorRecord) float64 {
    pattern := behavior.Action + ":" + behavior.Resource
    
    for _, normalPattern := range profile.NormalBehavior.AccessPatterns {
        if normalPattern == pattern {
            return 0 // 正常访问模式
        }
    }
    
    // 检查是否为敏感操作
    sensitiveActions := []string{"delete", "admin", "config", "user_management"}
    for _, sensitive := range sensitiveActions {
        if strings.Contains(strings.ToLower(behavior.Action), sensitive) {
            return 1.0 // 敏感操作异常
        }
    }
    
    return 0.3 // 一般新模式
}

// updateUserProfile 更新用户画像
func (ad *AnomalyDetector) updateUserProfile(profile *UserProfile, behavior BehaviorRecord) {
    // 添加到最近行为记录
    profile.RecentBehavior = append(profile.RecentBehavior, behavior)
    
    // 保持最近100条记录
    if len(profile.RecentBehavior) > 100 {
        profile.RecentBehavior = profile.RecentBehavior[1:]
    }
    
    // 更新正常行为模式（基于最近的行为）
    ad.updateNormalBehavior(profile)
    
    profile.LastUpdated = time.Now()
}

// updateNormalBehavior 更新正常行为模式
func (ad *AnomalyDetector) updateNormalBehavior(profile *UserProfile) {
    if len(profile.RecentBehavior) < 10 {
        return // 数据不足
    }
    
    // 更新登录时间
    var loginTimes []time.Time
    ipMap := make(map[string]bool)
    uaMap := make(map[string]bool)
    patternMap := make(map[string]bool)
    
    for _, record := range profile.RecentBehavior {
        if record.Action == "login" {
            loginTimes = append(loginTimes, record.Timestamp)
        }
        
        ipMap[record.IP] = true
        uaMap[record.UserAgent] = true
        
        pattern := record.Action + ":" + record.Resource
        patternMap[pattern] = true
    }
    
    // 转换为切片
    var ips, uas, patterns []string
    for ip := range ipMap {
        ips = append(ips, ip)
    }
    for ua := range uaMap {
        uas = append(uas, ua)
    }
    for pattern := range patternMap {
        patterns = append(patterns, pattern)
    }
    
    profile.NormalBehavior = BehaviorPattern{
        LoginTimes:     loginTimes,
        IPAddresses:    ips,
        UserAgents:     uas,
        AccessPatterns: patterns,
    }
}
```

## 安全最佳实践

### 开发安全规范

1. **代码安全**
   - 使用参数化查询防止SQL注入
   - 对所有用户输入进行验证和清理
   - 使用安全的加密算法和密钥管理
   - 定期更新依赖库和框架

2. **认证安全**
   - 实施强密码策略
   - 使用多因素认证
   - 实施会话管理和超时机制
   - 记录所有认证事件

3. **授权安全**
   - 实施最小权限原则
   - 使用基于角色的访问控制
   - 定期审查用户权限
   - 实施权限分离

4. **数据安全**
   - 加密敏感数据
   - 实施数据脱敏
   - 定期备份数据
   - 实施数据完整性检查

5. **网络安全**
   - 使用HTTPS加密通信
   - 实施网络分段
   - 配置防火墙规则
   - 监控网络流量

### 安全检查清单

#### 部署前检查

- [ ] 所有密码和密钥已更改为生产环境值
- [ ] 调试模式已关闭
- [ ] 错误信息不包含敏感信息
- [ ] 所有依赖库已更新到最新安全版本
- [ ] 安全配置已正确设置
- [ ] SSL/TLS证书已正确配置
- [ ] 防火墙规则已配置
- [ ] 监控和告警已启用

#### 运行时检查

- [ ] 定期检查安全日志
- [ ] 监控异常登录活动
- [ ] 检查权限变更
- [ ] 验证数据完整性
- [ ] 检查系统资源使用
- [ ] 更新安全补丁
- [ ] 备份验证
- [ ] 灾难恢复测试

### 应急响应流程

#### 安全事件响应

1. **事件识别**
   - 监控系统检测到异常
   - 用户报告安全问题
   - 第三方安全通知

2. **事件分类**
   - 确定事件类型和严重程度
   - 评估影响范围
   - 分配响应团队

3. **事件遏制**
   - 隔离受影响系统
   - 阻止攻击源
   - 保护关键数据

4. **事件调查**
   - 收集证据
   - 分析攻击路径
   - 确定根本原因

5. **事件恢复**
   - 修复安全漏洞
   - 恢复系统服务
   - 验证系统安全性

6. **事后总结**
   - 编写事件报告
   - 更新安全策略
   - 改进防护措施

## 总结

本安全指南提供了AI监控系统的全面安全保护方案，包括：

- **多层安全防护架构**：从基础设施到应用层的全方位保护
- **身份认证与授权**：JWT认证和RBAC权限控制
- **数据加密与保护**：敏感数据加密和脱敏处理
- **输入验证与防护**：防止各种注入攻击
- **限流与DDoS防护**：保护系统免受恶意攻击
- **安全审计与监控**：全面的安全事件记录和监控
- **威胁检测**：基于机器学习的异常行为检测
- **安全最佳实践**：开发和运维安全规范

通过实施这些安全措施，可以确保AI监控系统在面对各种安全威胁时的稳定性和可靠性。安全是一个持续的过程，需要定期评估、更新和改进安全策略。