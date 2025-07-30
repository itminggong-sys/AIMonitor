package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"ai-monitor/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKeyService API密钥管理服务
type APIKeyService struct {
	db *gorm.DB
}

// NewAPIKeyService 创建API密钥管理服务
func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{
		db: db,
	}
}

// CreateAPIKeyRequest 创建API密钥请求
type CreateAPIKeyRequest struct {
	Name        string     `json:"name" binding:"required"`
	Key         string     `json:"key" binding:"required"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// UpdateAPIKeyRequest 更新API密钥请求
type UpdateAPIKeyRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// APIKeyResponse API密钥响应
type APIKeyResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Key         string     `json:"key"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	UsageCount  int64      `json:"usage_count"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	UpdatedBy   uuid.UUID  `json:"updated_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateAPIKey 创建API密钥
func (s *APIKeyService) CreateAPIKey(req *CreateAPIKeyRequest, createdBy uuid.UUID) (*APIKeyResponse, error) {
	// 检查密钥是否已存在
	var existingKey models.APIKey
	if err := s.db.Where("key = ?", req.Key).First(&existingKey).Error; err == nil {
		return nil, errors.New("API key already exists")
	}

	// 检查名称是否已存在
	if err := s.db.Where("name = ?", req.Name).First(&existingKey).Error; err == nil {
		return nil, errors.New("API key name already exists")
	}

	// 创建API密钥
	apiKey := models.APIKey{
		Name:        req.Name,
		Key:         req.Key,
		Description: req.Description,
		Status:      "active",
		ExpiresAt:   req.ExpiresAt,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	if err := s.db.Create(&apiKey).Error; err != nil {
		return nil, err
	}

	return s.toAPIKeyResponse(&apiKey), nil
}

// GetAPIKey 获取API密钥
func (s *APIKeyService) GetAPIKey(id uuid.UUID) (*APIKeyResponse, error) {
	var apiKey models.APIKey
	if err := s.db.First(&apiKey, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API key not found")
		}
		return nil, err
	}

	return s.toAPIKeyResponse(&apiKey), nil
}

// GetAPIKeyByKey 根据密钥获取API密钥信息
func (s *APIKeyService) GetAPIKeyByKey(key string) (*APIKeyResponse, error) {
	var apiKey models.APIKey
	if err := s.db.Where("key = ? AND status = 'active'", key).First(&apiKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API key not found or inactive")
		}
		return nil, err
	}

	// 检查是否过期
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("API key expired")
	}

	return s.toAPIKeyResponse(&apiKey), nil
}

// ListAPIKeys 获取API密钥列表
func (s *APIKeyService) ListAPIKeys(page, pageSize int) ([]*APIKeyResponse, int64, error) {
	var apiKeys []models.APIKey
	var total int64

	// 获取总数
	if err := s.db.Model(&models.APIKey{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&apiKeys).Error; err != nil {
		return nil, 0, err
	}

	responses := make([]*APIKeyResponse, len(apiKeys))
	for i, apiKey := range apiKeys {
		responses[i] = s.toAPIKeyResponse(&apiKey)
	}

	return responses, total, nil
}

// UpdateAPIKey 更新API密钥
func (s *APIKeyService) UpdateAPIKey(id uuid.UUID, req *UpdateAPIKeyRequest, updatedBy uuid.UUID) (*APIKeyResponse, error) {
	var apiKey models.APIKey
	if err := s.db.First(&apiKey, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API key not found")
		}
		return nil, err
	}

	// 检查名称是否已被其他密钥使用
	if req.Name != "" && req.Name != apiKey.Name {
		var existingKey models.APIKey
		if err := s.db.Where("name = ? AND id != ?", req.Name, id).First(&existingKey).Error; err == nil {
			return nil, errors.New("API key name already exists")
		}
		apiKey.Name = req.Name
	}

	if req.Description != "" {
		apiKey.Description = req.Description
	}

	if req.Status != "" {
		apiKey.Status = req.Status
	}

	if req.ExpiresAt != nil {
		apiKey.ExpiresAt = req.ExpiresAt
	}

	apiKey.UpdatedBy = updatedBy

	if err := s.db.Save(&apiKey).Error; err != nil {
		return nil, err
	}

	return s.toAPIKeyResponse(&apiKey), nil
}

// DeleteAPIKey 删除API密钥
func (s *APIKeyService) DeleteAPIKey(id uuid.UUID) error {
	var apiKey models.APIKey
	if err := s.db.First(&apiKey, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("API key not found")
		}
		return err
	}

	return s.db.Delete(&apiKey).Error
}

// ValidateAPIKey 验证API密钥
func (s *APIKeyService) ValidateAPIKey(key string) (*APIKeyResponse, error) {
	apiKeyResp, err := s.GetAPIKeyByKey(key)
	if err != nil {
		return nil, err
	}

	// 更新使用统计
	now := time.Now()
	if err := s.db.Model(&models.APIKey{}).Where("id = ?", apiKeyResp.ID).Updates(map[string]interface{}{
		"last_used_at": now,
		"usage_count":  gorm.Expr("usage_count + 1"),
	}).Error; err != nil {
		return nil, err
	}

	apiKeyResp.LastUsedAt = &now
	apiKeyResp.UsageCount++

	return apiKeyResp, nil
}

// GenerateRandomKey 生成随机密钥
func (s *APIKeyService) GenerateRandomKey(length int) (string, error) {
	if length <= 0 {
		length = 32 // 默认32字符
	}

	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// UpdateUsage 更新API密钥使用统计
func (s *APIKeyService) UpdateUsage(id uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&models.APIKey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_used_at": now,
		"usage_count":  gorm.Expr("usage_count + 1"),
	}).Error
}

// toAPIKeyResponse 转换为响应格式
func (s *APIKeyService) toAPIKeyResponse(apiKey *models.APIKey) *APIKeyResponse {
	return &APIKeyResponse{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		Key:         apiKey.Key,
		Description: apiKey.Description,
		Status:      apiKey.Status,
		ExpiresAt:   apiKey.ExpiresAt,
		LastUsedAt:  apiKey.LastUsedAt,
		UsageCount:  apiKey.UsageCount,
		CreatedBy:   apiKey.CreatedBy,
		UpdatedBy:   apiKey.UpdatedBy,
		CreatedAt:   apiKey.CreatedAt,
		UpdatedAt:   apiKey.UpdatedAt,
	}
}