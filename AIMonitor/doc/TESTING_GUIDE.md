# AI监控系统测试指南

## 文档概述

本文档详细描述AI监控系统的测试策略、测试用例、测试流程和质量保证方法，确保系统的可靠性、性能和安全性。

## 目录

1. [测试策略](#测试策略)
2. [测试环境](#测试环境)
3. [单元测试](#单元测试)
4. [集成测试](#集成测试)
5. [API测试](#api测试)
6. [性能测试](#性能测试)
7. [安全测试](#安全测试)
8. [端到端测试](#端到端测试)
9. [测试自动化](#测试自动化)
10. [质量保证](#质量保证)

## 测试策略

### 测试金字塔

```
    /\     E2E Tests (10%)
   /  \    
  /____\   Integration Tests (20%)
 /______\  
/__________\ Unit Tests (70%)
```

### 测试类型

| 测试类型 | 覆盖范围 | 执行频率 | 工具 |
|----------|----------|----------|------|
| 单元测试 | 函数、方法 | 每次提交 | Go Test, Testify |
| 集成测试 | 模块间交互 | 每日构建 | Go Test, Docker |
| API测试 | REST API | 每次部署 | Postman, Newman |
| 性能测试 | 系统性能 | 每周 | JMeter, K6 |
| 安全测试 | 安全漏洞 | 每月 | OWASP ZAP, SonarQube |
| E2E测试 | 完整流程 | 发布前 | Selenium, Cypress |

### 测试原则

1. **测试驱动开发(TDD)**: 先写测试，再写代码
2. **持续集成**: 每次代码提交都运行测试
3. **快速反馈**: 测试结果快速反馈给开发者
4. **可重复性**: 测试结果可重复和可预测
5. **独立性**: 测试用例之间相互独立
6. **可维护性**: 测试代码易于维护和更新

## 测试环境

### 环境配置

#### 开发环境
- **用途**: 开发者本地测试
- **数据**: 模拟数据
- **配置**: 最小化配置

#### 测试环境
- **用途**: 自动化测试
- **数据**: 测试数据集
- **配置**: 接近生产环境

#### 预生产环境
- **用途**: 发布前验证
- **数据**: 生产数据副本
- **配置**: 生产环境配置

### 测试数据管理

#### 测试数据策略

```go
// 测试数据工厂
type TestDataFactory struct {
    db *gorm.DB
}

func (f *TestDataFactory) CreateUser(overrides ...func(*models.User)) *models.User {
    user := &models.User{
        Username: "testuser_" + uuid.New().String()[:8],
        Email:    "test@example.com",
        Password: "password123",
        Role:     "user",
        Status:   "active",
    }
    
    for _, override := range overrides {
        override(user)
    }
    
    f.db.Create(user)
    return user
}

func (f *TestDataFactory) CreateAlert(overrides ...func(*models.Alert)) *models.Alert {
    alert := &models.Alert{
        Title:       "Test Alert",
        Description: "Test alert description",
        Level:       "warning",
        Status:      "active",
        Value:       85.0,
    }
    
    for _, override := range overrides {
        override(alert)
    }
    
    f.db.Create(alert)
    return alert
}
```

#### 数据清理

```go
// 测试数据清理
func CleanupTestData(db *gorm.DB) {
    tables := []string{
        "audit_logs",
        "ai_analyses",
        "alerts",
        "alert_rules",
        "monitoring_targets",
        "users",
    }
    
    for _, table := range tables {
        db.Exec("TRUNCATE TABLE " + table + " CASCADE")
    }
}

// 测试基类
type BaseTestSuite struct {
    suite.Suite
    db      *gorm.DB
    factory *TestDataFactory
}

func (s *BaseTestSuite) SetupTest() {
    CleanupTestData(s.db)
    s.factory = &TestDataFactory{db: s.db}
}

func (s *BaseTestSuite) TearDownTest() {
    CleanupTestData(s.db)
}
```

## 单元测试

### 测试结构

```go
// internal/services/alert_service_test.go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

type AlertServiceTestSuite struct {
    suite.Suite
    service    *AlertService
    mockRepo   *MockAlertRepository
    mockNotify *MockNotificationService
}

func (s *AlertServiceTestSuite) SetupTest() {
    s.mockRepo = &MockAlertRepository{}
    s.mockNotify = &MockNotificationService{}
    s.service = NewAlertService(s.mockRepo, s.mockNotify)
}

func (s *AlertServiceTestSuite) TestCreateAlert_Success() {
    // Arrange
    alertData := &CreateAlertRequest{
        Title:       "High CPU Usage",
        Description: "CPU usage exceeds 80%",
        Level:       "warning",
    }
    
    expectedAlert := &models.Alert{
        ID:          1,
        Title:       alertData.Title,
        Description: alertData.Description,
        Level:       alertData.Level,
        Status:      "active",
    }
    
    s.mockRepo.On("Create", mock.AnythingOfType("*models.Alert")).Return(expectedAlert, nil)
    s.mockNotify.On("SendAlert", expectedAlert).Return(nil)
    
    // Act
    result, err := s.service.CreateAlert(alertData)
    
    // Assert
    assert.NoError(s.T(), err)
    assert.Equal(s.T(), expectedAlert.Title, result.Title)
    assert.Equal(s.T(), expectedAlert.Level, result.Level)
    s.mockRepo.AssertExpectations(s.T())
    s.mockNotify.AssertExpectations(s.T())
}

func (s *AlertServiceTestSuite) TestCreateAlert_ValidationError() {
    // Arrange
    alertData := &CreateAlertRequest{
        Title: "", // 空标题应该失败
        Level: "warning",
    }
    
    // Act
    result, err := s.service.CreateAlert(alertData)
    
    // Assert
    assert.Error(s.T(), err)
    assert.Nil(s.T(), result)
    assert.Contains(s.T(), err.Error(), "title is required")
}

func (s *AlertServiceTestSuite) TestCreateAlert_RepositoryError() {
    // Arrange
    alertData := &CreateAlertRequest{
        Title:       "High CPU Usage",
        Description: "CPU usage exceeds 80%",
        Level:       "warning",
    }
    
    s.mockRepo.On("Create", mock.AnythingOfType("*models.Alert")).Return(nil, errors.New("database error"))
    
    // Act
    result, err := s.service.CreateAlert(alertData)
    
    // Assert
    assert.Error(s.T(), err)
    assert.Nil(s.T(), result)
    assert.Contains(s.T(), err.Error(), "database error")
    s.mockRepo.AssertExpectations(s.T())
}

func TestAlertServiceTestSuite(t *testing.T) {
    suite.Run(t, new(AlertServiceTestSuite))
}
```

### Mock对象

```go
// mocks/alert_repository.go
type MockAlertRepository struct {
    mock.Mock
}

func (m *MockAlertRepository) Create(alert *models.Alert) (*models.Alert, error) {
    args := m.Called(alert)
    return args.Get(0).(*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetByID(id uint) (*models.Alert, error) {
    args := m.Called(id)
    return args.Get(0).(*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) List(filter *AlertFilter) ([]*models.Alert, error) {
    args := m.Called(filter)
    return args.Get(0).([]*models.Alert), args.Error(1)
}

func (m *MockAlertRepository) Update(alert *models.Alert) error {
    args := m.Called(alert)
    return args.Error(0)
}

func (m *MockAlertRepository) Delete(id uint) error {
    args := m.Called(id)
    return args.Error(0)
}
```

### 测试覆盖率

```bash
# 运行测试并生成覆盖率报告
go test -v -race -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 生成HTML覆盖率报告
go tool cover -html=coverage.out -o coverage.html

# 设置覆盖率阈值
go test -v -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' | awk '{if($1<80) exit 1}'
```

### 基准测试

```go
// 性能基准测试
func BenchmarkAlertService_CreateAlert(b *testing.B) {
    service := setupAlertService()
    alertData := &CreateAlertRequest{
        Title:       "Benchmark Alert",
        Description: "Benchmark test",
        Level:       "warning",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.CreateAlert(alertData)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkAlertService_GetAlerts(b *testing.B) {
    service := setupAlertService()
    filter := &AlertFilter{
        Status: "active",
        Limit:  10,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.GetAlerts(filter)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 集成测试

### 数据库集成测试

```go
// tests/integration/database_test.go
package integration

import (
    "testing"
    "github.com/stretchr/testify/suite"
    "gorm.io/gorm"
)

type DatabaseIntegrationTestSuite struct {
    BaseTestSuite
    alertRepo *repositories.AlertRepository
    userRepo  *repositories.UserRepository
}

func (s *DatabaseIntegrationTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    s.alertRepo = repositories.NewAlertRepository(s.db)
    s.userRepo = repositories.NewUserRepository(s.db)
}

func (s *DatabaseIntegrationTestSuite) TestCreateAndRetrieveAlert() {
    // 创建用户
    user := s.factory.CreateUser()
    
    // 创建告警
    alert := &models.Alert{
        Title:       "Integration Test Alert",
        Description: "Test alert for integration testing",
        Level:       "warning",
        Status:      "active",
        CreatedBy:   user.ID,
    }
    
    // 保存告警
    createdAlert, err := s.alertRepo.Create(alert)
    s.NoError(err)
    s.NotNil(createdAlert)
    s.NotZero(createdAlert.ID)
    
    // 检索告警
    retrievedAlert, err := s.alertRepo.GetByID(createdAlert.ID)
    s.NoError(err)
    s.Equal(alert.Title, retrievedAlert.Title)
    s.Equal(alert.Level, retrievedAlert.Level)
    s.Equal(user.ID, retrievedAlert.CreatedBy)
}

func (s *DatabaseIntegrationTestSuite) TestAlertWithRelations() {
    // 创建用户
    user := s.factory.CreateUser()
    
    // 创建告警规则
    rule := s.factory.CreateAlertRule(func(r *models.AlertRule) {
        r.CreatedBy = user.ID
    })
    
    // 创建告警
    alert := s.factory.CreateAlert(func(a *models.Alert) {
        a.RuleID = &rule.ID
        a.CreatedBy = user.ID
    })
    
    // 检索带关联的告警
    retrievedAlert, err := s.alertRepo.GetByIDWithRelations(alert.ID)
    s.NoError(err)
    s.NotNil(retrievedAlert.Rule)
    s.Equal(rule.Name, retrievedAlert.Rule.Name)
    s.NotNil(retrievedAlert.Creator)
    s.Equal(user.Username, retrievedAlert.Creator.Username)
}

func TestDatabaseIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(DatabaseIntegrationTestSuite))
}
```

### Redis集成测试

```go
// tests/integration/redis_test.go
package integration

import (
    "context"
    "testing"
    "time"
    "github.com/stretchr/testify/suite"
)

type RedisIntegrationTestSuite struct {
    suite.Suite
    redis *redis.Client
    cache *services.CacheService
}

func (s *RedisIntegrationTestSuite) SetupTest() {
    s.redis = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   1, // 使用测试数据库
    })
    
    s.cache = services.NewCacheService(s.redis)
    
    // 清理测试数据
    s.redis.FlushDB(context.Background())
}

func (s *RedisIntegrationTestSuite) TearDownTest() {
    s.redis.FlushDB(context.Background())
    s.redis.Close()
}

func (s *RedisIntegrationTestSuite) TestCacheSetAndGet() {
    key := "test:user:1"
    value := map[string]interface{}{
        "id":       1,
        "username": "testuser",
        "email":    "test@example.com",
    }
    
    // 设置缓存
    err := s.cache.Set(key, value, time.Minute)
    s.NoError(err)
    
    // 获取缓存
    var result map[string]interface{}
    err = s.cache.Get(key, &result)
    s.NoError(err)
    s.Equal(value["username"], result["username"])
    s.Equal(value["email"], result["email"])
}

func (s *RedisIntegrationTestSuite) TestCacheExpiration() {
    key := "test:expire"
    value := "test value"
    
    // 设置短期缓存
    err := s.cache.Set(key, value, 100*time.Millisecond)
    s.NoError(err)
    
    // 立即获取应该成功
    var result string
    err = s.cache.Get(key, &result)
    s.NoError(err)
    s.Equal(value, result)
    
    // 等待过期
    time.Sleep(150 * time.Millisecond)
    
    // 再次获取应该失败
    err = s.cache.Get(key, &result)
    s.Error(err)
}

func TestRedisIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(RedisIntegrationTestSuite))
}
```

### 外部服务集成测试

```go
// tests/integration/prometheus_test.go
package integration

import (
    "testing"
    "time"
    "github.com/stretchr/testify/suite"
)

type PrometheusIntegrationTestSuite struct {
    suite.Suite
    client *services.PrometheusClient
}

func (s *PrometheusIntegrationTestSuite) SetupTest() {
    s.client = services.NewPrometheusClient("http://localhost:9090")
}

func (s *PrometheusIntegrationTestSuite) TestQueryMetrics() {
    query := "up"
    
    result, err := s.client.Query(query)
    s.NoError(err)
    s.NotNil(result)
    s.NotEmpty(result.Data.Result)
}

func (s *PrometheusIntegrationTestSuite) TestQueryRange() {
    query := "rate(http_requests_total[5m])"
    start := time.Now().Add(-time.Hour)
    end := time.Now()
    step := time.Minute
    
    result, err := s.client.QueryRange(query, start, end, step)
    s.NoError(err)
    s.NotNil(result)
}

func TestPrometheusIntegrationTestSuite(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Prometheus integration test in short mode")
    }
    suite.Run(t, new(PrometheusIntegrationTestSuite))
}
```

## API测试

### HTTP API测试

```go
// tests/api/auth_test.go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/suite"
)

type AuthAPITestSuite struct {
    BaseAPITestSuite
}

func (s *AuthAPITestSuite) TestRegister_Success() {
    payload := map[string]interface{}{
        "username":         "testuser",
        "email":            "test@example.com",
        "password":         "password123",
        "confirm_password": "password123",
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    s.Equal(http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    s.NoError(err)
    s.Equal(float64(201), response["code"])
    s.Contains(response, "data")
    
    data := response["data"].(map[string]interface{})
    user := data["user"].(map[string]interface{})
    s.Equal("testuser", user["username"])
    s.Equal("test@example.com", user["email"])
}

func (s *AuthAPITestSuite) TestRegister_ValidationError() {
    payload := map[string]interface{}{
        "username": "testuser",
        "email":    "invalid-email", // 无效邮箱
        "password": "123",           // 密码太短
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    s.Equal(http.StatusBadRequest, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    s.NoError(err)
    s.Equal(float64(400), response["code"])
    s.Contains(response["message"], "validation")
}

func (s *AuthAPITestSuite) TestLogin_Success() {
    // 先创建用户
    user := s.factory.CreateUser(func(u *models.User) {
        u.Username = "testuser"
        u.Email = "test@example.com"
        u.Password = hashPassword("password123")
    })
    
    payload := map[string]interface{}{
        "username": user.Username,
        "password": "password123",
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    s.Equal(http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    s.NoError(err)
    s.Equal(float64(200), response["code"])
    
    data := response["data"].(map[string]interface{})
    s.Contains(data, "access_token")
    s.Contains(data, "refresh_token")
    s.Contains(data, "user")
}

func (s *AuthAPITestSuite) TestProtectedEndpoint_WithoutToken() {
    req := httptest.NewRequest("GET", "/api/v1/users/profile", nil)
    
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *AuthAPITestSuite) TestProtectedEndpoint_WithValidToken() {
    user := s.factory.CreateUser()
    token := s.generateJWTToken(user)
    
    req := httptest.NewRequest("GET", "/api/v1/users/profile", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    s.Equal(http.StatusOK, w.Code)
}

func TestAuthAPITestSuite(t *testing.T) {
    suite.Run(t, new(AuthAPITestSuite))
}
```

### API测试基类

```go
// tests/api/base_test.go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/suite"
    "gorm.io/gorm"
)

type BaseAPITestSuite struct {
    suite.Suite
    db      *gorm.DB
    router  *gin.Engine
    factory *TestDataFactory
}

func (s *BaseAPITestSuite) SetupSuite() {
    gin.SetMode(gin.TestMode)
    s.db = setupTestDatabase()
    s.router = setupTestRouter(s.db)
}

func (s *BaseAPITestSuite) SetupTest() {
    CleanupTestData(s.db)
    s.factory = &TestDataFactory{db: s.db}
}

func (s *BaseAPITestSuite) TearDownTest() {
    CleanupTestData(s.db)
}

func (s *BaseAPITestSuite) TearDownSuite() {
    sqlDB, _ := s.db.DB()
    sqlDB.Close()
}

func (s *BaseAPITestSuite) generateJWTToken(user *models.User) string {
    token, _ := auth.GenerateToken(user.ID, user.Username, user.Role)
    return token
}

func setupTestRouter(db *gorm.DB) *gin.Engine {
    router := gin.New()
    
    // 设置中间件
    router.Use(gin.Recovery())
    router.Use(middleware.CORS())
    
    // 设置路由
    api := router.Group("/api/v1")
    routes.SetupAuthRoutes(api, db)
    routes.SetupUserRoutes(api, db)
    routes.SetupAlertRoutes(api, db)
    
    return router
}
```

### Postman集合

```json
{
  "info": {
    "name": "AI Monitor API Tests",
    "description": "API测试集合",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080"
    },
    {
      "key": "accessToken",
      "value": ""
    }
  ],
  "item": [
    {
      "name": "Authentication",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"username\": \"testuser\",\n  \"email\": \"test@example.com\",\n  \"password\": \"password123\",\n  \"confirm_password\": \"password123\"\n}"
            },
            "url": {
              "raw": "{{baseUrl}}/api/v1/auth/register",
              "host": ["{{baseUrl}}"],
              "path": ["api", "v1", "auth", "register"]
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test('Status code is 201', function () {",
                  "    pm.response.to.have.status(201);",
                  "});",
                  "",
                  "pm.test('Response has user data', function () {",
                  "    var jsonData = pm.response.json();",
                  "    pm.expect(jsonData.data).to.have.property('user');",
                  "    pm.expect(jsonData.data.user).to.have.property('username');",
                  "});"
                ]
              }
            }
          ]
        },
        {
          "name": "Login",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"username\": \"admin\",\n  \"password\": \"admin123\"\n}"
            },
            "url": {
              "raw": "{{baseUrl}}/api/v1/auth/login",
              "host": ["{{baseUrl}}"],
              "path": ["api", "v1", "auth", "login"]
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test('Status code is 200', function () {",
                  "    pm.response.to.have.status(200);",
                  "});",
                  "",
                  "pm.test('Response has access token', function () {",
                  "    var jsonData = pm.response.json();",
                  "    pm.expect(jsonData.data).to.have.property('access_token');",
                  "    pm.collectionVariables.set('accessToken', jsonData.data.access_token);",
                  "});"
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}
```

## 性能测试

### JMeter测试计划

```xml
<?xml version="1.0" encoding="UTF-8"?>
<jmeterTestPlan version="1.2" properties="5.0" jmeter="5.4.1">
  <hashTree>
    <TestPlan guiclass="TestPlanGui" testclass="TestPlan" testname="AI Monitor Performance Test">
      <stringProp name="TestPlan.comments">AI监控系统性能测试</stringProp>
      <boolProp name="TestPlan.functional_mode">false</boolProp>
      <boolProp name="TestPlan.tearDown_on_shutdown">true</boolProp>
      <boolProp name="TestPlan.serialize_threadgroups">false</boolProp>
      <elementProp name="TestPlan.arguments" elementType="Arguments" guiclass="ArgumentsPanel" testclass="Arguments" testname="User Defined Variables">
        <collectionProp name="Arguments.arguments">
          <elementProp name="baseUrl" elementType="Argument">
            <stringProp name="Argument.name">baseUrl</stringProp>
            <stringProp name="Argument.value">http://localhost:8080</stringProp>
          </elementProp>
        </collectionProp>
      </elementProp>
    </TestPlan>
    <hashTree>
      <ThreadGroup guiclass="ThreadGroupGui" testclass="ThreadGroup" testname="API Load Test">
        <stringProp name="ThreadGroup.on_sample_error">continue</stringProp>
        <elementProp name="ThreadGroup.main_controller" elementType="LoopController" guiclass="LoopControlPanel" testclass="LoopController" testname="Loop Controller">
          <boolProp name="LoopController.continue_forever">false</boolProp>
          <stringProp name="LoopController.loops">10</stringProp>
        </elementProp>
        <stringProp name="ThreadGroup.num_threads">100</stringProp>
        <stringProp name="ThreadGroup.ramp_time">60</stringProp>
        <boolProp name="ThreadGroup.scheduler">false</boolProp>
        <stringProp name="ThreadGroup.duration"></stringProp>
        <stringProp name="ThreadGroup.delay"></stringProp>
      </ThreadGroup>
      <hashTree>
        <HTTPSamplerProxy guiclass="HttpTestSampleGui" testclass="HTTPSamplerProxy" testname="Login Request">
          <elementProp name="HTTPsampler.Arguments" elementType="Arguments" guiclass="HTTPArgumentsPanel" testclass="Arguments" testname="User Defined Variables">
            <collectionProp name="Arguments.arguments">
              <elementProp name="" elementType="HTTPArgument">
                <boolProp name="HTTPArgument.always_encode">false</boolProp>
                <stringProp name="Argument.value">{"username":"admin","password":"admin123"}</stringProp>
                <stringProp name="Argument.metadata">=</stringProp>
              </elementProp>
            </collectionProp>
          </elementProp>
          <stringProp name="HTTPSampler.domain">${baseUrl}</stringProp>
          <stringProp name="HTTPSampler.port"></stringProp>
          <stringProp name="HTTPSampler.protocol">http</stringProp>
          <stringProp name="HTTPSampler.contentEncoding"></stringProp>
          <stringProp name="HTTPSampler.path">/api/v1/auth/login</stringProp>
          <stringProp name="HTTPSampler.method">POST</stringProp>
          <boolProp name="HTTPSampler.follow_redirects">true</boolProp>
          <boolProp name="HTTPSampler.auto_redirects">false</boolProp>
          <boolProp name="HTTPSampler.use_keepalive">true</boolProp>
          <boolProp name="HTTPSampler.DO_MULTIPART_POST">false</boolProp>
          <stringProp name="HTTPSampler.embedded_url_re"></stringProp>
          <stringProp name="HTTPSampler.connect_timeout"></stringProp>
          <stringProp name="HTTPSampler.response_timeout"></stringProp>
        </HTTPSamplerProxy>
      </hashTree>
    </hashTree>
  </hashTree>
</jmeterTestPlan>
```

### K6性能测试

```javascript
// tests/performance/load_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// 自定义指标
const errorRate = new Rate('errors');

// 测试配置
export const options = {
  stages: [
    { duration: '2m', target: 100 }, // 2分钟内增加到100个用户
    { duration: '5m', target: 100 }, // 保持100个用户5分钟
    { duration: '2m', target: 200 }, // 2分钟内增加到200个用户
    { duration: '5m', target: 200 }, // 保持200个用户5分钟
    { duration: '2m', target: 0 },   // 2分钟内减少到0个用户
  ],
  thresholds: {
    http_req_duration: ['p(99)<1500'], // 99%的请求在1.5秒内完成
    http_req_failed: ['rate<0.1'],     // 错误率小于10%
    errors: ['rate<0.1'],              // 自定义错误率小于10%
  },
};

const BASE_URL = 'http://localhost:8080';

// 登录获取token
function login() {
  const payload = JSON.stringify({
    username: 'admin',
    password: 'admin123',
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const response = http.post(`${BASE_URL}/api/v1/auth/login`, payload, params);
  
  const success = check(response, {
    'login status is 200': (r) => r.status === 200,
    'login response has token': (r) => JSON.parse(r.body).data.access_token !== undefined,
  });
  
  errorRate.add(!success);
  
  if (success) {
    return JSON.parse(response.body).data.access_token;
  }
  return null;
}

// 获取告警列表
function getAlerts(token) {
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  };
  
  const response = http.get(`${BASE_URL}/api/v1/alerts`, params);
  
  const success = check(response, {
    'get alerts status is 200': (r) => r.status === 200,
    'get alerts response has data': (r) => JSON.parse(r.body).data !== undefined,
  });
  
  errorRate.add(!success);
}

// 创建告警
function createAlert(token) {
  const payload = JSON.stringify({
    title: `Load Test Alert ${Math.random()}`,
    description: 'Alert created during load testing',
    level: 'warning',
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  };
  
  const response = http.post(`${BASE_URL}/api/v1/alerts`, payload, params);
  
  const success = check(response, {
    'create alert status is 201': (r) => r.status === 201,
    'create alert response has id': (r) => JSON.parse(r.body).data.alert.id !== undefined,
  });
  
  errorRate.add(!success);
}

// 查询指标
function queryMetrics(token) {
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  };
  
  const response = http.get(`${BASE_URL}/api/v1/monitoring/metrics?query=up`, params);
  
  const success = check(response, {
    'query metrics status is 200': (r) => r.status === 200,
  });
  
  errorRate.add(!success);
}

// 主测试函数
export default function () {
  // 登录
  const token = login();
  if (!token) {
    return;
  }
  
  sleep(1);
  
  // 执行各种操作
  getAlerts(token);
  sleep(1);
  
  createAlert(token);
  sleep(1);
  
  queryMetrics(token);
  sleep(1);
}

// 测试设置
export function setup() {
  console.log('Starting load test...');
}

// 测试清理
export function teardown(data) {
  console.log('Load test completed.');
}
```

### 压力测试脚本

```bash
#!/bin/bash
# tests/performance/stress_test.sh

echo "开始压力测试..."

# 设置变量
BASE_URL="http://localhost:8080"
MAX_USERS=1000
DURATION=300  # 5分钟

# 创建结果目录
mkdir -p results

# 运行K6压力测试
echo "运行K6压力测试..."
k6 run --vus $MAX_USERS --duration ${DURATION}s tests/performance/stress_test.js > results/k6_stress_test.log

# 运行JMeter压力测试
echo "运行JMeter压力测试..."
jmeter -n -t tests/performance/stress_test.jmx -l results/jmeter_stress_test.jtl -e -o results/jmeter_report

# 生成报告
echo "生成性能报告..."
python3 tests/performance/generate_report.py

echo "压力测试完成，结果保存在results目录"
```

## 安全测试

### OWASP ZAP自动化测试

```python
# tests/security/zap_test.py
import time
import requests
from zapv2 import ZAPv2

class SecurityTest:
    def __init__(self):
        self.zap = ZAPv2(proxies={'http': 'http://127.0.0.1:8080', 'https': 'http://127.0.0.1:8080'})
        self.target_url = 'http://localhost:8080'
    
    def spider_scan(self):
        """爬虫扫描"""
        print('开始爬虫扫描...')
        scan_id = self.zap.spider.scan(self.target_url)
        
        # 等待扫描完成
        while int(self.zap.spider.status(scan_id)) < 100:
            print(f'爬虫扫描进度: {self.zap.spider.status(scan_id)}%')
            time.sleep(2)
        
        print('爬虫扫描完成')
    
    def active_scan(self):
        """主动扫描"""
        print('开始主动安全扫描...')
        scan_id = self.zap.ascan.scan(self.target_url)
        
        # 等待扫描完成
        while int(self.zap.ascan.status(scan_id)) < 100:
            print(f'主动扫描进度: {self.zap.ascan.status(scan_id)}%')
            time.sleep(5)
        
        print('主动扫描完成')
    
    def generate_report(self):
        """生成报告"""
        print('生成安全测试报告...')
        
        # 获取告警
        alerts = self.zap.core.alerts()
        
        # 生成HTML报告
        html_report = self.zap.core.htmlreport()
        with open('results/security_report.html', 'w') as f:
            f.write(html_report)
        
        # 生成JSON报告
        json_report = self.zap.core.jsonreport()
        with open('results/security_report.json', 'w') as f:
            f.write(json_report)
        
        print(f'发现 {len(alerts)} 个安全问题')
        
        # 打印高危漏洞
        high_risk_alerts = [alert for alert in alerts if alert['risk'] == 'High']
        if high_risk_alerts:
            print('高危漏洞:')
            for alert in high_risk_alerts:
                print(f'  - {alert["alert"]}: {alert["description"]}')
    
    def run_security_test(self):
        """运行完整安全测试"""
        try:
            self.spider_scan()
            self.active_scan()
            self.generate_report()
        except Exception as e:
            print(f'安全测试失败: {e}')

if __name__ == '__main__':
    security_test = SecurityTest()
    security_test.run_security_test()
```

### SQL注入测试

```go
// tests/security/sql_injection_test.go
package security

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/suite"
)

type SQLInjectionTestSuite struct {
    BaseSecurityTestSuite
}

func (s *SQLInjectionTestSuite) TestLoginSQLInjection() {
    sqlInjectionPayloads := []string{
        "admin' OR '1'='1",
        "admin'; DROP TABLE users; --",
        "admin' UNION SELECT * FROM users --",
        "admin' AND (SELECT COUNT(*) FROM users) > 0 --",
    }
    
    for _, payload := range sqlInjectionPayloads {
        s.Run("SQL Injection: "+payload, func() {
            loginData := map[string]interface{}{
                "username": payload,
                "password": "password",
            }
            
            body, _ := json.Marshal(loginData)
            req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
            req.Header.Set("Content-Type", "application/json")
            
            w := httptest.NewRecorder()
            s.router.ServeHTTP(w, req)
            
            // SQL注入应该被阻止，返回401或400
            s.True(w.Code == http.StatusUnauthorized || w.Code == http.StatusBadRequest,
                "SQL injection payload should be blocked: %s", payload)
        })
    }
}

func (s *SQLInjectionTestSuite) TestSearchSQLInjection() {
    user := s.factory.CreateUser()
    token := s.generateJWTToken(user)
    
    sqlInjectionPayloads := []string{
        "test' OR '1'='1",
        "test'; DELETE FROM alerts; --",
        "test' UNION SELECT password FROM users --",
    }
    
    for _, payload := range sqlInjectionPayloads {
        s.Run("Search SQL Injection: "+payload, func() {
            req := httptest.NewRequest("GET", "/api/v1/alerts?search="+payload, nil)
            req.Header.Set("Authorization", "Bearer "+token)
            
            w := httptest.NewRecorder()
            s.router.ServeHTTP(w, req)
            
            // 应该返回正常响应，但不应该执行SQL注入
            s.Equal(http.StatusOK, w.Code)
            
            var response map[string]interface{}
            err := json.Unmarshal(w.Body.Bytes(), &response)
            s.NoError(err)
            
            // 检查响应中不应该包含敏感信息
            responseStr := w.Body.String()
            s.NotContains(responseStr, "password")
            s.NotContains(responseStr, "hash")
        })
    }
}

func TestSQLInjectionTestSuite(t *testing.T) {
    suite.Run(t, new(SQLInjectionTestSuite))
}
```

### XSS测试

```go
// tests/security/xss_test.go
package security

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/suite"
)

type XSSTestSuite struct {
    BaseSecurityTestSuite
}

func (s *XSSTestSuite) TestCreateAlertXSS() {
    user := s.factory.CreateUser()
    token := s.generateJWTToken(user)
    
    xssPayloads := []string{
        "<script>alert('XSS')</script>",
        "<img src=x onerror=alert('XSS')>",
        "javascript:alert('XSS')",
        "<svg onload=alert('XSS')>",
        "<iframe src=javascript:alert('XSS')></iframe>",
    }
    
    for _, payload := range xssPayloads {
        s.Run("XSS Payload: "+payload, func() {
            alertData := map[string]interface{}{
                "title":       payload,
                "description": "Test alert with XSS payload",
                "level":       "warning",
            }
            
            body, _ := json.Marshal(alertData)
            req := httptest.NewRequest("POST", "/api/v1/alerts", bytes.NewBuffer(body))
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("Authorization", "Bearer "+token)
            
            w := httptest.NewRecorder()
            s.router.ServeHTTP(w, req)
            
            if w.Code == http.StatusCreated {
                // 如果创建成功，检查返回的数据是否被正确转义
                var response map[string]interface{}
                err := json.Unmarshal(w.Body.Bytes(), &response)
                s.NoError(err)
                
                data := response["data"].(map[string]interface{})
                alert := data["alert"].(map[string]interface{})
                title := alert["title"].(string)
                
                // 检查XSS payload是否被转义或过滤
                s.NotContains(title, "<script>")
                s.NotContains(title, "javascript:")
                s.NotContains(title, "onerror=")
                s.NotContains(title, "onload=")
            }
        })
    }
}

func (s *XSSTestSuite) TestGetAlertXSS() {
    user := s.factory.CreateUser()
    token := s.generateJWTToken(user)
    
    // 创建包含潜在XSS的告警
    alert := s.factory.CreateAlert(func(a *models.Alert) {
        a.Title = "<script>alert('XSS')</script>"
        a.Description = "<img src=x onerror=alert('XSS')>"
        a.CreatedBy = user.ID
    })
    
    req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/alerts/%d", alert.ID), nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    s.Equal(http.StatusOK, w.Code)
    
    // 检查响应中的XSS payload是否被正确处理
    responseBody := w.Body.String()
    s.NotContains(responseBody, "<script>")
    s.NotContains(responseBody, "onerror=")
}

func TestXSSTestSuite(t *testing.T) {
    suite.Run(t, new(XSSTestSuite))
}
```

## 端到端测试

### Selenium测试

```python
# tests/e2e/selenium_test.py
import time
import unittest
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.chrome.options import Options

class E2ETestSuite(unittest.TestCase):
    def setUp(self):
        chrome_options = Options()
        chrome_options.add_argument('--headless')  # 无头模式
        chrome_options.add_argument('--no-sandbox')
        chrome_options.add_argument('--disable-dev-shm-usage')
        
        self.driver = webdriver.Chrome(options=chrome_options)
        self.driver.implicitly_wait(10)
        self.base_url = 'http://localhost:3000'  # 前端应用地址
    
    def tearDown(self):
        self.driver.quit()
    
    def test_user_login_flow(self):
        """测试用户登录流程"""
        driver = self.driver
        
        # 访问登录页面
        driver.get(f'{self.base_url}/login')
        
        # 等待页面加载
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.ID, 'username'))
        )
        
        # 输入用户名和密码
        username_input = driver.find_element(By.ID, 'username')
        password_input = driver.find_element(By.ID, 'password')
        login_button = driver.find_element(By.ID, 'login-button')
        
        username_input.send_keys('admin')
        password_input.send_keys('admin123')
        login_button.click()
        
        # 等待跳转到仪表板
        WebDriverWait(driver, 10).until(
            EC.url_contains('/dashboard')
        )
        
        # 验证登录成功
        self.assertIn('/dashboard', driver.current_url)
        
        # 检查用户信息显示
        user_info = WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.CLASS_NAME, 'user-info'))
        )
        self.assertIn('admin', user_info.text)
    
    def test_alert_management_flow(self):
        """测试告警管理流程"""
        driver = self.driver
        
        # 先登录
        self._login()
        
        # 导航到告警页面
        driver.get(f'{self.base_url}/alerts')
        
        # 等待告警列表加载
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.CLASS_NAME, 'alert-list'))
        )
        
        # 点击创建告警按钮
        create_button = driver.find_element(By.ID, 'create-alert-button')
        create_button.click()
        
        # 等待创建表单出现
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.ID, 'alert-form'))
        )
        
        # 填写告警信息
        title_input = driver.find_element(By.ID, 'alert-title')
        description_input = driver.find_element(By.ID, 'alert-description')
        level_select = driver.find_element(By.ID, 'alert-level')
        submit_button = driver.find_element(By.ID, 'submit-alert')
        
        title_input.send_keys('E2E Test Alert')
        description_input.send_keys('This is a test alert created by E2E test')
        level_select.send_keys('warning')
        submit_button.click()
        
        # 等待告警创建成功
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.CLASS_NAME, 'success-message'))
        )
        
        # 验证告警出现在列表中
        alert_list = driver.find_element(By.CLASS_NAME, 'alert-list')
        self.assertIn('E2E Test Alert', alert_list.text)
    
    def test_monitoring_dashboard(self):
        """测试监控仪表板"""
        driver = self.driver
        
        # 先登录
        self._login()
        
        # 导航到监控仪表板
        driver.get(f'{self.base_url}/monitoring')
        
        # 等待图表加载
        WebDriverWait(driver, 15).until(
            EC.presence_of_element_located((By.CLASS_NAME, 'chart-container'))
        )
        
        # 检查各个监控图表是否存在
        charts = driver.find_elements(By.CLASS_NAME, 'chart-container')
        self.assertGreater(len(charts), 0)
        
        # 检查CPU使用率图表
        cpu_chart = driver.find_element(By.ID, 'cpu-usage-chart')
        self.assertTrue(cpu_chart.is_displayed())
        
        # 检查内存使用率图表
        memory_chart = driver.find_element(By.ID, 'memory-usage-chart')
        self.assertTrue(memory_chart.is_displayed())
        
        # 测试时间范围选择
        time_range_selector = driver.find_element(By.ID, 'time-range-selector')
        time_range_selector.click()
        
        # 选择1小时范围
        one_hour_option = driver.find_element(By.XPATH, "//option[@value='1h']")
        one_hour_option.click()
        
        # 等待图表更新
        time.sleep(2)
        
        # 验证图表已更新
        updated_chart = driver.find_element(By.ID, 'cpu-usage-chart')
        self.assertTrue(updated_chart.is_displayed())
    
    def test_ai_analysis_flow(self):
        """测试AI分析流程"""
        driver = self.driver
        
        # 先登录
        self._login()
        
        # 导航到AI分析页面
        driver.get(f'{self.base_url}/ai-analysis')
        
        # 等待页面加载
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.ID, 'ai-analysis-container'))
        )
        
        # 点击开始分析按钮
        analyze_button = driver.find_element(By.ID, 'start-analysis-button')
        analyze_button.click()
        
        # 等待分析完成
        WebDriverWait(driver, 30).until(
            EC.presence_of_element_located((By.CLASS_NAME, 'analysis-result'))
        )
        
        # 验证分析结果显示
        result_container = driver.find_element(By.CLASS_NAME, 'analysis-result')
        self.assertTrue(result_container.is_displayed())
        self.assertIn('分析完成', result_container.text)
    
    def _login(self):
        """辅助方法：登录"""
        driver = self.driver
        driver.get(f'{self.base_url}/login')
        
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.ID, 'username'))
        )
        
        username_input = driver.find_element(By.ID, 'username')
        password_input = driver.find_element(By.ID, 'password')
        login_button = driver.find_element(By.ID, 'login-button')
        
        username_input.send_keys('admin')
        password_input.send_keys('admin123')
        login_button.click()
        
        WebDriverWait(driver, 10).until(
            EC.url_contains('/dashboard')
        )

if __name__ == '__main__':
    unittest.main()
```

### Cypress测试

```javascript
// tests/e2e/cypress/integration/app_spec.js
describe('AI Monitor E2E Tests', () => {
  beforeEach(() => {
    // 访问应用
    cy.visit('/');
  });
  
  it('should login successfully', () => {
    // 访问登录页面
    cy.visit('/login');
    
    // 输入凭据
    cy.get('#username').type('admin');
    cy.get('#password').type('admin123');
    cy.get('#login-button').click();
    
    // 验证跳转到仪表板
    cy.url().should('include', '/dashboard');
    cy.get('.user-info').should('contain', 'admin');
  });
  
  it('should create and manage alerts', () => {
    // 登录
    cy.login('admin', 'admin123');
    
    // 导航到告警页面
    cy.visit('/alerts');
    
    // 创建新告警
    cy.get('#create-alert-button').click();
    cy.get('#alert-title').type('Cypress Test Alert');
    cy.get('#alert-description').type('Test alert created by Cypress');
    cy.get('#alert-level').select('warning');
    cy.get('#submit-alert').click();
    
    // 验证告警创建成功
    cy.get('.success-message').should('be.visible');
    cy.get('.alert-list').should('contain', 'Cypress Test Alert');
    
    // 编辑告警
    cy.get('[data-testid="edit-alert"]').first().click();
    cy.get('#alert-title').clear().type('Updated Cypress Test Alert');
    cy.get('#submit-alert').click();
    
    // 验证告警更新成功
    cy.get('.alert-list').should('contain', 'Updated Cypress Test Alert');
    
    // 删除告警
    cy.get('[data-testid="delete-alert"]').first().click();
    cy.get('#confirm-delete').click();
    
    // 验证告警删除成功
    cy.get('.alert-list').should('not.contain', 'Updated Cypress Test Alert');
  });
  
  it('should display monitoring charts', () => {
    // 登录
    cy.login('admin', 'admin123');
    
    // 导航到监控页面
    cy.visit('/monitoring');
    
    // 验证图表加载
    cy.get('.chart-container').should('be.visible');
    cy.get('#cpu-usage-chart').should('be.visible');
    cy.get('#memory-usage-chart').should('be.visible');
    
    // 测试时间范围选择
    cy.get('#time-range-selector').select('1h');
    cy.wait(2000); // 等待图表更新
    cy.get('#cpu-usage-chart').should('be.visible');
  });
  
  it('should perform AI analysis', () => {
    // 登录
    cy.login('admin', 'admin123');
    
    // 导航到AI分析页面
    cy.visit('/ai-analysis');
    
    // 开始分析
    cy.get('#start-analysis-button').click();
    
    // 等待分析完成
    cy.get('.analysis-result', { timeout: 30000 }).should('be.visible');
    cy.get('.analysis-result').should('contain', '分析完成');
  });
});

// 自定义命令
Cy.Commands.add('login', (username, password) => {
  cy.visit('/login');
  cy.get('#username').type(username);
  cy.get('#password').type(password);
  cy.get('#login-button').click();
  cy.url().should('include', '/dashboard');
});
```

## 测试自动化

### CI/CD集成

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: aimonitor_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:6
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.19
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_USER: postgres
        DB_PASSWORD: postgres
        DB_NAME: aimonitor_test
        REDIS_HOST: localhost
        REDIS_PORT: 6379
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
  
  integration-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: aimonitor_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:6
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.19
    
    - name: Run integration tests
      run: go test -v -tags=integration ./tests/integration/...
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_USER: postgres
        DB_PASSWORD: postgres
        DB_NAME: aimonitor_test
        REDIS_HOST: localhost
        REDIS_PORT: 6379
  
  api-tests:
    runs-on: ubuntu-latest
    needs: integration-tests
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '16'
    
    - name: Install Newman
      run: npm install -g newman
    
    - name: Start application
      run: |
        docker-compose -f docker-compose.test.yml up -d
        sleep 30
    
    - name: Run API tests
      run: |
        newman run tests/api/postman_collection.json \
          --environment tests/api/test_environment.json \
          --reporters cli,json \
          --reporter-json-export results/api_test_results.json
    
    - name: Upload API test results
      uses: actions/upload-artifact@v3
      with:
        name: api-test-results
        path: results/api_test_results.json
  
  security-tests:
    runs-on: ubuntu-latest
    needs: api-tests
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Run OWASP ZAP Baseline Scan
      uses: zaproxy/action-baseline@v0.7.0
      with:
        target: 'http://localhost:8080'
        rules_file_name: '.zap/rules.tsv'
        cmd_options: '-a'
  
  e2e-tests:
    runs-on: ubuntu-latest
    needs: api-tests
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '16'
    
    - name: Install dependencies
      run: |
        cd frontend
        npm ci
    
    - name: Start application
      run: |
        docker-compose -f docker-compose.test.yml up -d
        sleep 60
    
    - name: Run Cypress tests
      run: |
        cd frontend
        npm run cy:run
      env:
        CYPRESS_baseUrl: http://localhost:3000
    
    - name: Upload Cypress screenshots
      uses: actions/upload-artifact@v3
      if: failure()
      with:
        name: cypress-screenshots
        path: frontend/cypress/screenshots
    
    - name: Upload Cypress videos
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: cypress-videos
        path: frontend/cypress/videos
```

### 测试报告生成

```python
# scripts/generate_test_report.py
import json
import os
import datetime
from jinja2 import Template

class TestReportGenerator:
    def __init__(self):
        self.report_data = {
            'timestamp': datetime.datetime.now().isoformat(),
            'summary': {},
            'unit_tests': {},
            'integration_tests': {},
            'api_tests': {},
            'security_tests': {},
            'e2e_tests': {},
            'performance_tests': {}
        }
    
    def load_unit_test_results(self, coverage_file):
        """加载单元测试结果"""
        if os.path.exists(coverage_file):
            with open(coverage_file, 'r') as f:
                coverage_data = f.read()
            
            # 解析覆盖率
            lines = coverage_data.split('\n')
            total_line = [line for line in lines if 'total:' in line]
            if total_line:
                coverage = total_line[0].split()[-1]
                self.report_data['unit_tests']['coverage'] = coverage
    
    def load_api_test_results(self, results_file):
        """加载API测试结果"""
        if os.path.exists(results_file):
            with open(results_file, 'r') as f:
                api_results = json.load(f)
            
            self.report_data['api_tests'] = {
                'total_requests': api_results['run']['stats']['requests']['total'],
                'failed_requests': api_results['run']['stats']['requests']['failed'],
                'avg_response_time': api_results['run']['timings']['responseAverage'],
                'assertions': api_results['run']['stats']['assertions']
            }
    
    def load_security_test_results(self, results_file):
        """加载安全测试结果"""
        if os.path.exists(results_file):
            with open(results_file, 'r') as f:
                security_results = json.load(f)
            
            alerts = security_results.get('site', [{}])[0].get('alerts', [])
            
            self.report_data['security_tests'] = {
                'total_alerts': len(alerts),
                'high_risk': len([a for a in alerts if a.get('riskdesc', '').startswith('High')]),
                'medium_risk': len([a for a in alerts if a.get('riskdesc', '').startswith('Medium')]),
                'low_risk': len([a for a in alerts if a.get('riskdesc', '').startswith('Low')])
            }
    
    def generate_html_report(self, output_file):
        """生成HTML测试报告"""
        template_str = """
<!DOCTYPE html>
<html>
<head>
    <title>AI监控系统测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .success { color: green; }
        .warning { color: orange; }
        .error { color: red; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background-color: #f9f9f9; border-radius: 3px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>AI监控系统测试报告</h1>
        <p>生成时间: {{ timestamp }}</p>
    </div>
    
    <div class="section">
        <h2>测试概览</h2>
        <div class="metric">
            <strong>单元测试覆盖率:</strong> 
            <span class="{% if unit_tests.coverage and unit_tests.coverage|replace('%', '')|float >= 80 %}success{% else %}warning{% endif %}">
                {{ unit_tests.coverage or 'N/A' }}
            </span>
        </div>
        <div class="metric">
            <strong>API测试:</strong> 
            <span class="{% if api_tests.failed_requests == 0 %}success{% else %}error{% endif %}">
                {{ api_tests.total_requests - api_tests.failed_requests }}/{{ api_tests.total_requests }} 通过
            </span>
        </div>
        <div class="metric">
            <strong>安全测试:</strong> 
            <span class="{% if security_tests.high_risk == 0 %}success{% else %}error{% endif %}">
                {{ security_tests.high_risk }} 高危漏洞
            </span>
        </div>
    </div>
    
    <div class="section">
        <h2>单元测试</h2>
        <p>代码覆盖率: {{ unit_tests.coverage or 'N/A' }}</p>
    </div>
    
    <div class="section">
        <h2>API测试</h2>
        <table>
            <tr><th>指标</th><th>值</th></tr>
            <tr><td>总请求数</td><td>{{ api_tests.total_requests or 'N/A' }}</td></tr>
            <tr><td>失败请求数</td><td>{{ api_tests.failed_requests or 'N/A' }}</td></tr>
            <tr><td>平均响应时间</td><td>{{ api_tests.avg_response_time or 'N/A' }}ms</td></tr>
        </table>
    </div>
    
    <div class="section">
        <h2>安全测试</h2>
        <table>
            <tr><th>风险级别</th><th>数量</th></tr>
            <tr><td>高危</td><td class="{% if security_tests.high_risk == 0 %}success{% else %}error{% endif %}">{{ security_tests.high_risk or 0 }}</td></tr>
            <tr><td>中危</td><td class="{% if security_tests.medium_risk == 0 %}success{% else %}warning{% endif %}">{{ security_tests.medium_risk or 0 }}</td></tr>
            <tr><td>低危</td><td>{{ security_tests.low_risk or 0 }}</td></tr>
        </table>
    </div>
</body>
</html>
        """
        
        template = Template(template_str)
        html_content = template.render(**self.report_data)
        
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(html_content)
        
        print(f'测试报告已生成: {output_file}')

if __name__ == '__main__':
    generator = TestReportGenerator()
    
    # 加载各种测试结果
    generator.load_unit_test_results('coverage.out')
    generator.load_api_test_results('results/api_test_results.json')
    generator.load_security_test_results('results/security_report.json')
    
    # 生成报告
    generator.generate_html_report('results/test_report.html')
```

## 质量保证

### 代码质量检查

```bash
#!/bin/bash
# scripts/quality_check.sh

echo "开始代码质量检查..."

# Go代码格式检查
echo "检查Go代码格式..."
gofmt -l . | tee gofmt_issues.txt
if [ -s gofmt_issues.txt ]; then
    echo "发现格式问题，请运行 gofmt -w . 修复"
    exit 1
fi

# Go代码静态分析
echo "运行静态分析..."
go vet ./...
if [ $? -ne 0 ]; then
    echo "静态分析发现问题"
    exit 1
fi

# 运行golint
echo "运行golint..."
golint ./... | tee golint_issues.txt
if [ -s golint_issues.txt ]; then
    echo "发现代码风格问题"
    cat golint_issues.txt
fi

# 运行gosec安全检查
echo "运行安全检查..."
gosec ./...
if [ $? -ne 0 ]; then
    echo "发现安全问题"
    exit 1
fi

# 检查依赖漏洞
echo "检查依赖漏洞..."
go list -json -m all | nancy sleuth
if [ $? -ne 0 ]; then
    echo "发现依赖漏洞"
    exit 1
fi

echo "代码质量检查完成"
```

### 测试数据管理

```go
// tests/fixtures/fixtures.go
package fixtures

import (
    "encoding/json"
    "io/ioutil"
    "path/filepath"
)

type TestFixtures struct {
    Users      []map[string]interface{} `json:"users"`
    Alerts     []map[string]interface{} `json:"alerts"`
    AlertRules []map[string]interface{} `json:"alert_rules"`
}

func LoadFixtures(fixturesDir string) (*TestFixtures, error) {
    fixturesFile := filepath.Join(fixturesDir, "test_data.json")
    
    data, err := ioutil.ReadFile(fixturesFile)
    if err != nil {
        return nil, err
    }
    
    var fixtures TestFixtures
    err = json.Unmarshal(data, &fixtures)
    if err != nil {
        return nil, err
    }
    
    return &fixtures, nil
}

func (f *TestFixtures) GetUser(username string) map[string]interface{} {
    for _, user := range f.Users {
        if user["username"] == username {
            return user
        }
    }
    return nil
}

func (f *TestFixtures) GetAlert(title string) map[string]interface{} {
    for _, alert := range f.Alerts {
        if alert["title"] == title {
            return alert
        }
    }
    return nil
}
```

### 测试环境管理

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  app-test:
    build:
      context: .
      dockerfile: Dockerfile.test
    environment:
      - ENV=test
      - DB_HOST=postgres-test
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=aimonitor_test
      - REDIS_HOST=redis-test
      - REDIS_PORT=6379
    depends_on:
      - postgres-test
      - redis-test
    ports:
      - "8080:8080"
    volumes:
      - ./tests:/app/tests
  
  postgres-test:
    image: postgres:13
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=aimonitor_test
    ports:
      - "5432:5432"
    volumes:
      - postgres_test_data:/var/lib/postgresql/data
  
  redis-test:
    image: redis:6
    ports:
      - "6379:6379"
    volumes:
      - redis_test_data:/data
  
  prometheus-test:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./tests/config/prometheus.yml:/etc/prometheus/prometheus.yml

volumes:
  postgres_test_data:
  redis_test_data:
```

### 持续质量监控

```yaml
# .github/workflows/quality.yml
name: Code Quality

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  quality-check:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.19
    
    - name: Install quality tools
      run: |
        go install golang.org/x/lint/golint@latest
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1
    
    - name: Run quality checks
      run: |
        gofmt -l .
        go vet ./...
        golint ./...
        gosec ./...
        golangci-lint run
    
    - name: SonarCloud Scan
      uses: SonarSource/sonarcloud-github-action@master
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

## 总结

本测试指南提供了AI监控系统的完整测试策略和实施方案，包括：

1. **全面的测试覆盖**: 从单元测试到端到端测试的完整测试金字塔
2. **自动化测试**: CI/CD集成和自动化测试流水线
3. **质量保证**: 代码质量检查和持续监控
4. **安全测试**: 全面的安全漏洞检测
5. **性能测试**: 负载和压力测试
6. **测试数据管理**: 测试数据的创建、管理和清理

通过遵循本指南，可以确保AI监控系统的高质量交付和持续改进。