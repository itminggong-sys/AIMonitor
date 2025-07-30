package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ai-monitor/internal/auth"
	"ai-monitor/internal/cache"
	"ai-monitor/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db           *gorm.DB
	cacheManager *cache.CacheManager
	jwtManager   *auth.JWTManager
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB, cacheManager *cache.CacheManager, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		db:           db,
		cacheManager: cacheManager,
		jwtManager:   jwtManager,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6"`
	FullName string   `json:"full_name"`
	Phone    string   `json:"phone"`
	Roles    []string `json:"roles"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	FullName string   `json:"full_name"`
	Phone    string   `json:"phone"`
	Avatar   string   `json:"avatar"`
	Status   string   `json:"status" binding:"omitempty,oneof=active inactive locked"`
	Roles    []string `json:"roles"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         *UserResponse `json:"user"`
	TokenInfo    *auth.TokenInfo `json:"token_info"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	Avatar      string    `json:"avatar"`
	Phone       string    `json:"phone"`
	Status      string    `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LoginCount  int       `json:"login_count"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *CreateUserRequest) (*UserResponse, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("username or email already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Phone:    req.Phone,
		Status:   "active",
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建用户
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 分配角色
	if len(req.Roles) > 0 {
		var roles []models.Role
		if err := tx.Where("name IN ?", req.Roles).Find(&roles).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to find roles: %w", err)
		}

		if err := tx.Model(&user).Association("Roles").Append(roles); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to assign roles: %w", err)
		}
	}

	tx.Commit()

	// 重新加载用户数据
	if err := s.db.Preload("Roles.Permissions").First(&user, user.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	return s.toUserResponse(&user), nil
}

// GetUser 获取用户
func (s *UserService) GetUser(userID uuid.UUID) (*UserResponse, error) {
	var user models.User
	if err := s.db.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserResponse(&user), nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles.Permissions").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles.Permissions").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(userID uuid.UUID, req *UpdateUserRequest) (*UserResponse, error) {
	var user models.User
	if err := s.db.Preload("Roles").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户信息
	updates := map[string]interface{}{}
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// 更新角色
	if req.Roles != nil {
		var roles []models.Role
		if len(req.Roles) > 0 {
			if err := tx.Where("name IN ?", req.Roles).Find(&roles).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to find roles: %w", err)
			}
		}

		if err := tx.Model(&user).Association("Roles").Replace(roles); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update roles: %w", err)
		}
	}

	tx.Commit()

	// 重新加载用户数据
	if err := s.db.Preload("Roles.Permissions").First(&user, user.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	// 清除缓存
	s.clearUserCache(userID)

	return s.toUserResponse(&user), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(userID uuid.UUID) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 软删除用户
	if err := s.db.Delete(&user).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// 清除缓存
	s.clearUserCache(userID)

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int, search string) ([]*UserResponse, int64, error) {
	query := s.db.Model(&models.User{}).Preload("Roles.Permissions")

	// 搜索条件
	if search != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ? OR full_name ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	var users []models.User
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// 转换为响应格式
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.toUserResponse(&user)
	}

	return responses, total, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uuid.UUID, req *ChangePasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest) (*LoginResponse, error) {
	// 获取用户
	user, err := s.GetUserByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 生成令牌
	tokenInfo, err := s.jwtManager.GenerateTokenInfo(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 更新登录信息
	now := time.Now()
	s.db.Model(user).Updates(map[string]interface{}{
		"last_login_at": now,
		"login_count":   gorm.Expr("login_count + 1"),
	})

	return &LoginResponse{
		User:      s.toUserResponse(user),
		TokenInfo: tokenInfo,
	}, nil
}

// RefreshToken 刷新令牌
func (s *UserService) RefreshToken(req *RefreshTokenRequest) (*auth.TokenInfo, error) {
	// 验证刷新令牌
	claims, err := s.jwtManager.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("not a refresh token")
	}

	// 获取用户
	user, err := s.GetUserByUsername(claims.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	// 生成新的令牌对
	return s.jwtManager.GenerateTokenInfo(user)
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(userID uuid.UUID) (*UserResponse, error) {
	return s.GetUser(userID)
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*UserResponse, error) {
	// 转换为 UpdateUserRequest
	updateReq := &UpdateUserRequest{
		FullName: req.FullName,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		// 不允许通过此接口修改状态和角色
		Status: "",
		Roles:  nil,
	}
	return s.UpdateUser(userID, updateReq)
}

// ResetPassword 重置密码（管理员功能）
func (s *UserService) ResetPassword(userID uuid.UUID, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// LockUser 锁定用户
func (s *UserService) LockUser(userID uuid.UUID) error {
	return s.updateUserStatus(userID, "locked")
}

// UnlockUser 解锁用户
func (s *UserService) UnlockUser(userID uuid.UUID) error {
	return s.updateUserStatus(userID, "active")
}

// DeactivateUser 停用用户
func (s *UserService) DeactivateUser(userID uuid.UUID) error {
	return s.updateUserStatus(userID, "inactive")
}

// ActivateUser 激活用户
func (s *UserService) ActivateUser(userID uuid.UUID) error {
	return s.updateUserStatus(userID, "active")
}

// updateUserStatus 更新用户状态
func (s *UserService) updateUserStatus(userID uuid.UUID, status string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.db.Model(&user).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// 清除缓存
	s.clearUserCache(userID)

	return nil
}

// toUserResponse 转换为用户响应格式
func (s *UserService) toUserResponse(user *models.User) *UserResponse {
	roles := make([]string, len(user.Roles))
	permissions := make([]string, 0)

	for i, role := range user.Roles {
		roles[i] = role.Name
		for _, permission := range role.Permissions {
			permissions = append(permissions, permission.Name)
		}
	}

	// 去重权限
	permissions = removeDuplicateStrings(permissions)

	return &UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		Avatar:      user.Avatar,
		Phone:       user.Phone,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		LoginCount:  user.LoginCount,
		Roles:       roles,
		Permissions: permissions,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// clearUserCache 清除用户缓存
func (s *UserService) clearUserCache(userID uuid.UUID) {
	if s.cacheManager != nil {
		key := cache.UserCacheKey(userID.String())
		s.cacheManager.Delete(context.Background(), key)
	}
}

// removeDuplicateStrings 去除字符串切片中的重复项
func removeDuplicateStrings(slice []string) []string {
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

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总用户数
	var totalUsers int64
	if err := s.db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}
	stats["total_users"] = totalUsers

	// 活跃用户数
	var activeUsers int64
	if err := s.db.Model(&models.User{}).Where("status = ?", "active").Count(&activeUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}
	stats["active_users"] = activeUsers

	// 锁定用户数
	var lockedUsers int64
	if err := s.db.Model(&models.User{}).Where("status = ?", "locked").Count(&lockedUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count locked users: %w", err)
	}
	stats["locked_users"] = lockedUsers

	// 今日新增用户数
	var todayUsers int64
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&todayUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count today users: %w", err)
	}
	stats["today_users"] = todayUsers

	// 本月新增用户数
	var monthUsers int64
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	if err := s.db.Model(&models.User{}).Where("created_at >= ?", monthStart).Count(&monthUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count month users: %w", err)
	}
	stats["month_users"] = monthUsers

	return stats, nil
}