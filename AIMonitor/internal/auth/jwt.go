package auth

import (
	"errors"
	"fmt"
	"time"

	"ai-monitor/internal/config"
	"ai-monitor/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager JWT管理器
type JWTManager struct {
	secretKey             string
	accessTokenDuration   time.Duration
	refreshTokenDuration  time.Duration
	issuer                string
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey:             cfg.SecretKey,
		accessTokenDuration:   cfg.AccessTokenExpiry,
		refreshTokenDuration:  cfg.RefreshTokenExpiry,
		issuer:                cfg.Issuer,
	}
}

// UserClaims 用户声明
type UserClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	TokenType   string    `json:"token_type"` // access 或 refresh
	jwt.RegisteredClaims
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (j *JWTManager) GenerateTokenPair(user *models.User) (accessToken, refreshToken string, err error) {
	// 获取用户角色和权限
	roles := make([]string, len(user.Roles))
	permissions := make([]string, 0)
	
	for i, role := range user.Roles {
		roles[i] = role.Name
		for _, permission := range role.Permissions {
			permissions = append(permissions, permission.Name)
		}
	}

	// 去重权限
	permissions = removeDuplicates(permissions)

	// 生成访问令牌
	accessToken, err = j.generateToken(user, roles, permissions, "access", j.accessTokenDuration)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err = j.generateToken(user, roles, permissions, "refresh", j.refreshTokenDuration)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateToken 生成令牌
func (j *JWTManager) generateToken(user *models.User, roles, permissions []string, tokenType string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := UserClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Roles:       roles,
		Permissions: permissions,
		TokenType:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   user.ID.String(),
			Audience:  []string{"ai-monitor"},
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// VerifyToken 验证令牌
func (j *JWTManager) VerifyToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 检查令牌是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	// 检查令牌是否还未生效
	if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
		return nil, errors.New("token not valid yet")
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func (j *JWTManager) RefreshToken(refreshTokenString string, user *models.User) (accessToken, newRefreshToken string, err error) {
	// 验证刷新令牌
	claims, err := j.VerifyToken(refreshTokenString)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// 检查是否为刷新令牌
	if claims.TokenType != "refresh" {
		return "", "", errors.New("not a refresh token")
	}

	// 检查用户ID是否匹配
	if claims.UserID != user.ID {
		return "", "", errors.New("token user mismatch")
	}

	// 生成新的令牌对
	return j.GenerateTokenPair(user)
}

// ExtractTokenFromHeader 从请求头中提取令牌
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must start with Bearer")
	}

	return authHeader[len(bearerPrefix):], nil
}

// HasPermission 检查用户是否有指定权限
func (c *UserClaims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasRole 检查用户是否有指定角色
func (c *UserClaims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole 检查用户是否有任意指定角色
func (c *UserClaims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles 检查用户是否有所有指定角色
func (c *UserClaims) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !c.HasRole(role) {
			return false
		}
	}
	return true
}

// HasAnyPermission 检查用户是否有任意指定权限
func (c *UserClaims) HasAnyPermission(permissions ...string) bool {
	for _, permission := range permissions {
		if c.HasPermission(permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查用户是否有所有指定权限
func (c *UserClaims) HasAllPermissions(permissions ...string) bool {
	for _, permission := range permissions {
		if !c.HasPermission(permission) {
			return false
		}
	}
	return true
}

// IsAdmin 检查用户是否为管理员
func (c *UserClaims) IsAdmin() bool {
	return c.HasRole("admin")
}

// IsOperator 检查用户是否为运维人员
func (c *UserClaims) IsOperator() bool {
	return c.HasRole("operator")
}

// IsViewer 检查用户是否为只读用户
func (c *UserClaims) IsViewer() bool {
	return c.HasRole("viewer")
}

// CanManageUsers 检查用户是否可以管理用户
func (c *UserClaims) CanManageUsers() bool {
	return c.HasAnyPermission("user.create", "user.update", "user.delete")
}

// CanManageAlerts 检查用户是否可以管理告警
func (c *UserClaims) CanManageAlerts() bool {
	return c.HasAnyPermission("alert.create", "alert.update", "alert.delete")
}

// CanViewMonitoring 检查用户是否可以查看监控数据
func (c *UserClaims) CanViewMonitoring() bool {
	return c.HasPermission("monitoring.read")
}

// CanManageSystem 检查用户是否可以管理系统
func (c *UserClaims) CanManageSystem() bool {
	return c.HasPermission("system.config")
}

// CanUseAI 检查用户是否可以使用AI功能
func (c *UserClaims) CanUseAI() bool {
	return c.HasPermission("ai.analysis")
}

// GetUserInfo 获取用户基本信息
func (c *UserClaims) GetUserInfo() map[string]interface{} {
	return map[string]interface{}{
		"user_id":     c.UserID,
		"username":    c.Username,
		"email":       c.Email,
		"roles":       c.Roles,
		"permissions": c.Permissions,
		"issued_at":   c.IssuedAt,
		"expires_at":  c.ExpiresAt,
	}
}

// TokenInfo 令牌信息
type TokenInfo struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	TokenType             string    `json:"token_type"`
	ExpiresIn             int64     `json:"expires_in"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

// GenerateTokenInfo 生成令牌信息
func (j *JWTManager) GenerateTokenInfo(user *models.User) (*TokenInfo, error) {
	accessToken, refreshToken, err := j.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &TokenInfo{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		ExpiresIn:             int64(j.accessTokenDuration.Seconds()),
		AccessTokenExpiresAt:  now.Add(j.accessTokenDuration),
		RefreshTokenExpiresAt: now.Add(j.refreshTokenDuration),
	}, nil
}

// ValidateTokenType 验证令牌类型
func (j *JWTManager) ValidateTokenType(tokenString, expectedType string) error {
	claims, err := j.VerifyToken(tokenString)
	if err != nil {
		return err
	}

	if claims.TokenType != expectedType {
		return fmt.Errorf("expected %s token, got %s", expectedType, claims.TokenType)
	}

	return nil
}

// GetTokenClaims 获取令牌声明（不验证过期时间）
func (j *JWTManager) GetTokenClaims(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// IsTokenExpired 检查令牌是否过期
func (j *JWTManager) IsTokenExpired(tokenString string) bool {
	claims, err := j.GetTokenClaims(tokenString)
	if err != nil {
		return true
	}

	if claims.ExpiresAt == nil {
		return false
	}

	return claims.ExpiresAt.Time.Before(time.Now())
}

// GetTokenRemainingTime 获取令牌剩余时间
func (j *JWTManager) GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	claims, err := j.GetTokenClaims(tokenString)
	if err != nil {
		return 0, err
	}

	if claims.ExpiresAt == nil {
		return 0, errors.New("token has no expiration time")
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

// 辅助函数

// removeDuplicates 去除字符串切片中的重复项
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// BlacklistManager 令牌黑名单管理器
type BlacklistManager struct {
	// 这里可以使用Redis或内存存储黑名单
	// 为简化实现，这里使用内存map
	blacklist map[string]time.Time
}

// NewBlacklistManager 创建黑名单管理器
func NewBlacklistManager() *BlacklistManager {
	return &BlacklistManager{
		blacklist: make(map[string]time.Time),
	}
}

// AddToBlacklist 添加令牌到黑名单
func (b *BlacklistManager) AddToBlacklist(tokenID string, expiresAt time.Time) {
	b.blacklist[tokenID] = expiresAt
}

// IsBlacklisted 检查令牌是否在黑名单中
func (b *BlacklistManager) IsBlacklisted(tokenID string) bool {
	expiresAt, exists := b.blacklist[tokenID]
	if !exists {
		return false
	}

	// 如果令牌已过期，从黑名单中移除
	if time.Now().After(expiresAt) {
		delete(b.blacklist, tokenID)
		return false
	}

	return true
}

// CleanupExpired 清理过期的黑名单条目
func (b *BlacklistManager) CleanupExpired() {
	now := time.Now()
	for tokenID, expiresAt := range b.blacklist {
		if now.After(expiresAt) {
			delete(b.blacklist, tokenID)
		}
	}
}