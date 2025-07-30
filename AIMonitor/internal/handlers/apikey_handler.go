package handlers

import (
	"net/http"
	"strconv"

	"ai-monitor/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIKeyHandler API密钥管理处理器
type APIKeyHandler struct {
	apikeyService *services.APIKeyService
}

// NewAPIKeyHandler 创建API密钥管理处理器
func NewAPIKeyHandler(apikeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apikeyService: apikeyService,
	}
}

// CreateAPIKey 创建API密钥
// @Summary 创建API密钥
// @Description 创建新的API密钥用于Agent认证
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Param request body services.CreateAPIKeyRequest true "创建请求"
// @Success 201 {object} services.APIKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apikeys [post]
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req services.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	createdBy, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Invalid user ID format",
		})
		return
	}

	apiKey, err := h.apikeyService.CreateAPIKey(&req, createdBy)
	if err != nil {
		if err.Error() == "API key already exists" || err.Error() == "API key name already exists" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, apiKey)
}

// GetAPIKey 获取API密钥信息
// @Summary 获取API密钥信息
// @Description 获取指定API密钥的详细信息
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Param id path string true "API密钥ID"
// @Success 200 {object} services.APIKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apikeys/{id} [get]
func (h *APIKeyHandler) GetAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	apiKey, err := h.apikeyService.GetAPIKey(id)
	if err != nil {
		if err.Error() == "API key not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// ListAPIKeys 获取API密钥列表
// @Summary 获取API密钥列表
// @Description 获取API密钥列表，支持分页
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} PaginatedResponse{data=[]services.APIKeyResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apikeys [get]
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	apiKeys, total, err := h.apikeyService.ListAPIKeys(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data: apiKeys,
		Pagination: PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    int(total),
			Pages:    int((total + int64(pageSize) - 1) / int64(pageSize)),
		},
	})
}

// UpdateAPIKey 更新API密钥信息
// @Summary 更新API密钥信息
// @Description 更新指定API密钥的信息
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Param id path string true "API密钥ID"
// @Param request body services.UpdateAPIKeyRequest true "更新请求"
// @Success 200 {object} services.APIKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apikeys/{id} [put]
func (h *APIKeyHandler) UpdateAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	var req services.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	updatedBy, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Invalid user ID format",
		})
		return
	}

	apiKey, err := h.apikeyService.UpdateAPIKey(id, &req, updatedBy)
	if err != nil {
		if err.Error() == "API key not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		if err.Error() == "API key name already exists" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// DeleteAPIKey 删除API密钥
// @Summary 删除API密钥
// @Description 删除指定的API密钥
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Param id path string true "API密钥ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apikeys/{id} [delete]
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Message: "ID must be a valid UUID",
		})
		return
	}

	err = h.apikeyService.DeleteAPIKey(id)
	if err != nil {
		if err.Error() == "API key not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ValidateAPIKeyRequest 验证API密钥请求
type ValidateAPIKeyRequest struct {
	Key string `json:"key" binding:"required"`
}

// ValidateAPIKey 验证API密钥
// @Summary 验证API密钥
// @Description 验证API密钥是否有效
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Param request body ValidateAPIKeyRequest true "验证请求"
// @Success 200 {object} services.APIKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apikeys/validate [post]
func (h *APIKeyHandler) ValidateAPIKey(c *gin.Context) {
	var req ValidateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	key := req.Key
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "API key is required",
		})
		return
	}

	apiKey, err := h.apikeyService.ValidateAPIKey(key)
	if err != nil {
		if err.Error() == "API key not found or inactive" || err.Error() == "API key expired" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// GenerateAPIKey 生成随机密钥
// @Summary 生成随机密钥
// @Description 生成指定长度的随机密钥
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Param length query int false "密钥长度" default(32)
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/apikeys/generate [get]
func (h *APIKeyHandler) GenerateAPIKey(c *gin.Context) {
	lengthStr := c.DefaultQuery("length", "32")
	length, err := strconv.Atoi(lengthStr)
	if err != nil || length < 8 || length > 128 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "Length must be between 8 and 128",
		})
		return
	}

	key, err := h.apikeyService.GenerateRandomKey(length)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key": key,
	})
}

// APIKeyAuthMiddleware API密钥认证中间件
func (h *APIKeyHandler) APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先尝试JWT认证
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token := authHeader[7:]
				// 尝试验证JWT token - 这里需要实现JWT验证逻辑
				// 暂时跳过JWT验证，直接尝试API Key验证
				// TODO: 实现JWT token验证
				
				// JWT验证失败，尝试API Key验证
				if _, err := h.apikeyService.ValidateAPIKey(token); err == nil {
					c.Set("auth_type", "apikey")
					c.Set("api_key", token)
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "Valid JWT token or API key required",
		})
		c.Abort()
	}
}