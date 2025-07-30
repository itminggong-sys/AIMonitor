package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"
	"ai-monitor/internal/models"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// AIService AI服务
type AIService struct {
	db           *gorm.DB
	cacheManager *cache.CacheManager
	config       *config.Config
	openaiClient *openai.Client
}

// NewAIService 创建AI服务
func NewAIService(db *gorm.DB, cacheManager *cache.CacheManager, config *config.Config) *AIService {
	var openaiClient *openai.Client
	if config.AIModels.OpenAI.APIKey != "" {
		// 创建OpenAI客户端配置
		clientConfig := openai.DefaultConfig(config.AIModels.OpenAI.APIKey)
		
		// 如果配置了自定义BaseURL，则使用自定义URL（支持Ollama等本地服务）
		if config.AIModels.OpenAI.BaseURL != "" {
			clientConfig.BaseURL = config.AIModels.OpenAI.BaseURL
		}
		
		openaiClient = openai.NewClientWithConfig(clientConfig)
	}

	return &AIService{
		db:           db,
		cacheManager: cacheManager,
		config:       config,
		openaiClient: openaiClient,
	}
}

// AIAnalysisRequest AI分析请求
type AIAnalysisRequest struct {
	Type         string                 `json:"type" binding:"required,oneof=alert_analysis performance_analysis trend_analysis capacity_planning"`
	AlertID      uuid.UUID              `json:"alert_id,omitempty"`
	RuleID       uuid.UUID              `json:"rule_id,omitempty"`
	TargetType   string                 `json:"target_type"`
	TargetID     string                 `json:"target_id"`
	MetricName   string                 `json:"metric_name"`
	CurrentValue float64                `json:"current_value"`
	Threshold    float64                `json:"threshold"`
	Condition    string                 `json:"condition"`
	Severity     string                 `json:"severity"`
	Tags         map[string]interface{} `json:"tags"`
	Timestamp    time.Time              `json:"timestamp"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// PerformanceAnalysisRequest 性能分析请求
type PerformanceAnalysisRequest struct {
	TargetType   string                 `json:"target_type" binding:"required"`
	TargetID     string                 `json:"target_id" binding:"required"`
	MetricNames  []string               `json:"metric_names"`
	TimeRange    string                 `json:"time_range" binding:"required"`
	StartTime    *time.Time             `json:"start_time"`
	EndTime      *time.Time             `json:"end_time"`
	Tags         map[string]interface{} `json:"tags"`
	Context      map[string]interface{} `json:"context"`
}

// AIAnalysisResponse AI分析响应
type AIAnalysisResponse struct {
	ID               uuid.UUID              `json:"id"`
	Type             string                 `json:"type"`
	TargetType       string                 `json:"target_type"`
	TargetID         string                 `json:"target_id"`
	AnalysisResult   string                 `json:"analysis_result"`
	RootCause        string                 `json:"root_cause"`
	Recommendations  []string               `json:"recommendations"`
	SeverityLevel    string                 `json:"severity_level"`
	ConfidenceScore  float64                `json:"confidence_score"`
	Tags             map[string]interface{} `json:"tags"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
}

// KnowledgeBaseRequest 知识库请求
type KnowledgeBaseRequest struct {
	Title       string                 `json:"title" binding:"required"`
	Content     string                 `json:"content" binding:"required"`
	Category    string                 `json:"category" binding:"required"`
	Tags        []string               `json:"tags"`
	MetricTypes []string               `json:"metric_types"`
	Severity    string                 `json:"severity"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// KnowledgeBaseResponse 知识库响应
type KnowledgeBaseResponse struct {
	ID          uuid.UUID              `json:"id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	MetricTypes []string               `json:"metric_types"`
	Severity    string                 `json:"severity"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedBy   uuid.UUID              `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AnalyzeAlert 分析告警
func (s *AIService) AnalyzeAlert(req *AIAnalysisRequest) (*AIAnalysisResponse, error) {
	if s.openaiClient == nil {
		return nil, errors.New("AI service not configured")
	}

	// 检查缓存
	cacheKey := s.generateCacheKey("alert_analysis", req)
	if s.cacheManager != nil {
		var cached string
		if err := s.cacheManager.Get(context.Background(), cacheKey, &cached); err == nil {
			var response AIAnalysisResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return &response, nil
			}
		}
	}

	// 执行异常检测算法
	anomalyScore, err := s.detectAnomaly(req)
	if err != nil {
		return nil, fmt.Errorf("anomaly detection failed: %w", err)
	}

	// 执行趋势预测分析
	prediction, err := s.predictTrend(req)
	if err != nil {
		return nil, fmt.Errorf("trend prediction failed: %w", err)
	}

	// 构建增强的分析上下文
	context_str := s.buildEnhancedAnalysisContext(req, anomalyScore, prediction)

	// 查询相关知识库
	knowledgeContext, err := s.getRelevantKnowledge(req.MetricName, req.Severity)
	if err == nil && knowledgeContext != "" {
		context_str += "\n\n相关知识库信息:\n" + knowledgeContext
	}

	// 调用AI模型
	analysisResult, err := s.callOpenAI(context_str, "alert_analysis")
	if err != nil {
		return nil, fmt.Errorf("failed to call AI model: %w", err)
	}

	// 解析AI响应
	parsedResult, err := s.parseAIResponse(analysisResult)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 创建分析结果记录
	analysis := models.AIAnalysisResult{
		AnalysisType:    req.Type,
		Model:           "openai-gpt-4",
		Response:        analysisResult,
		Confidence:      parsedResult.ConfidenceScore,
		Status:          "completed",
	}

	// 序列化标签和元数据到metadata中
	// 注意：models.AIAnalysisResult中没有Tags字段，将标签信息存储在metadata中

	metadata := map[string]interface{}{
		"alert_id":      req.AlertID,
		"rule_id":       req.RuleID,
		"metric_name":   req.MetricName,
		"current_value": req.CurrentValue,
		"threshold":     req.Threshold,
		"condition":     req.Condition,
		"timestamp":     req.Timestamp,
	}
	metadataJSON, _ := json.Marshal(metadata)
	analysis.Metadata = string(metadataJSON)

	// 保存到数据库
	if err := s.db.Create(&analysis).Error; err != nil {
		return nil, fmt.Errorf("failed to save analysis result: %w", err)
	}

	// 构建响应
	response := &AIAnalysisResponse{
		ID:               analysis.ID,
		Type:             analysis.AnalysisType,
		TargetType:       req.TargetType,
		TargetID:         req.TargetID,
		AnalysisResult:   analysis.Response,
		RootCause:        parsedResult.RootCause,
		Recommendations:  parsedResult.Recommendations,
		SeverityLevel:    parsedResult.SeverityLevel,
		ConfidenceScore:  analysis.Confidence,
		Tags:             req.Tags,
		Metadata:         metadata,
		CreatedAt:        analysis.CreatedAt,
	}

	// 缓存结果
	if s.cacheManager != nil {
		if data, err := json.Marshal(response); err == nil {
			s.cacheManager.Set(context.Background(), cacheKey, string(data), 30*time.Minute)
		}
	}

	return response, nil
}

// AnalyzePerformance 性能分析
func (s *AIService) AnalyzePerformance(req *AIAnalysisRequest) (*AIAnalysisResponse, error) {
	if s.openaiClient == nil {
		return nil, errors.New("AI service not configured")
	}

	// 构建性能分析上下文
	context_str := s.buildPerformanceAnalysisContext(req)

	// 调用AI模型
	analysisResult, err := s.callOpenAI(context_str, "performance_analysis")
	if err != nil {
		return nil, fmt.Errorf("failed to call AI model: %w", err)
	}

	// 解析AI响应
	parsedResult, err := s.parseAIResponse(analysisResult)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 创建分析结果记录
	analysis := models.AIAnalysisResult{
		AnalysisType:    req.Type,
		Model:           "openai-gpt-4",
		Response:        analysisResult,
		Confidence:      parsedResult.ConfidenceScore,
		Status:          "completed",
	}

	// 序列化元数据
	metadata := map[string]interface{}{
		"metric_name":   req.MetricName,
		"current_value": req.CurrentValue,
		"timestamp":     req.Timestamp,
		"context":       req.Context,
	}
	metadataJSON, _ := json.Marshal(metadata)
	analysis.Metadata = string(metadataJSON)

	// 保存到数据库
	if err := s.db.Create(&analysis).Error; err != nil {
		return nil, fmt.Errorf("failed to save analysis result: %w", err)
	}

	// 构建响应
	return &AIAnalysisResponse{
		ID:               analysis.ID,
		Type:             analysis.AnalysisType,
		TargetType:       req.TargetType,
		TargetID:         req.TargetID,
		AnalysisResult:   analysis.Response,
		RootCause:        parsedResult.RootCause,
		Recommendations:  parsedResult.Recommendations,
		SeverityLevel:    parsedResult.SeverityLevel,
		ConfidenceScore:  analysis.Confidence,
		Tags:             req.Tags,
		Metadata:         metadata,
		CreatedAt:        analysis.CreatedAt,
	}, nil
}

// ListAnalysis 获取AI分析历史
func (s *AIService) ListAnalysis(page, pageSize int, analysisType, targetType string) ([]*AIAnalysisResponse, int64, error) {
	query := s.db.Model(&models.AIAnalysisResult{})

	// 分析类型过滤
	if analysisType != "" {
		query = query.Where("analysis_type = ?", analysisType)
	}

	// 目标类型过滤（通过metadata字段）
	if targetType != "" {
		query = query.Where("metadata LIKE ?", "%\"target_type\":\""+targetType+"\"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count analysis results: %w", err)
	}

	// 分页查询
	var results []models.AIAnalysisResult
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list analysis results: %w", err)
	}

	// 转换为响应格式
	responses := make([]*AIAnalysisResponse, len(results))
	for i, result := range results {
		responses[i] = s.toAIAnalysisResponse(&result)
	}

	return responses, total, nil
}

// toAIAnalysisResponse 转换为AI分析响应格式
func (s *AIService) toAIAnalysisResponse(result *models.AIAnalysisResult) *AIAnalysisResponse {
	var metadata map[string]interface{}
	if result.Metadata != "" {
		json.Unmarshal([]byte(result.Metadata), &metadata)
	}

	// 解析AI响应
	parsedResult, _ := s.parseAIResponse(result.Response)

	return &AIAnalysisResponse{
		ID:               result.ID,
		Type:             result.AnalysisType,
		TargetType:       "", // 从metadata中获取
		TargetID:         "", // 从metadata中获取
		AnalysisResult:   result.Response,
		RootCause:        parsedResult.RootCause,
		Recommendations:  parsedResult.Recommendations,
		SeverityLevel:    parsedResult.SeverityLevel,
		ConfidenceScore:  result.Confidence,
		Tags:             nil, // 从metadata中获取
		Metadata:         metadata,
		CreatedAt:        result.CreatedAt,
	}
}

// GetAnalysisHistory 获取分析历史
func (s *AIService) GetAnalysisHistory(page, pageSize int, analysisType, targetType string) ([]*AIAnalysisResponse, int64, error) {
	query := s.db.Model(&models.AIAnalysisResult{})

	// 类型过滤
	if analysisType != "" {
		query = query.Where("analysis_type = ?", analysisType)
	}

	// 目标类型过滤 - 注意：models.AIAnalysisResult中没有target_type字段
	// 如果需要按目标类型过滤，需要通过metadata字段或其他方式实现
	if targetType != "" {
		// 暂时注释掉，因为数据库模型中没有这个字段
		// query = query.Where("target_type = ?", targetType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count analysis results: %w", err)
	}

	// 分页查询
	var results []models.AIAnalysisResult
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list analysis results: %w", err)
	}

	// 转换为响应格式
	responses := make([]*AIAnalysisResponse, len(results))
	for i, result := range results {
		responses[i] = s.toAnalysisResponse(&result)
	}

	return responses, total, nil
}

// CreateKnowledgeBase 创建知识库条目
func (s *AIService) CreateKnowledgeBase(req *KnowledgeBaseRequest, createdBy uuid.UUID) (*KnowledgeBaseResponse, error) {
	// 序列化标签和指标类型
	tagsJSON, _ := json.Marshal(req.Tags)
	metricTypesJSON, _ := json.Marshal(req.MetricTypes)
	// metadataJSON, _ := json.Marshal(req.Metadata)

	// 创建知识库条目
	kb := models.KnowledgeBase{
		Title:       req.Title,
		Content:     req.Content,
		Category:    req.Category,
		Tags:        string(tagsJSON),
		Metrics:     string(metricTypesJSON), // 使用Metrics字段而不是MetricTypes
		Severity:    req.Severity,
		// 注意：models.KnowledgeBase中没有Metadata字段，可以将元数据存储在其他字段中
		CreatedBy:   createdBy,
	}

	if err := s.db.Create(&kb).Error; err != nil {
		return nil, fmt.Errorf("failed to create knowledge base entry: %w", err)
	}

	return s.toKnowledgeBaseResponse(&kb), nil
}

// GetKnowledgeBase 获取知识库条目
func (s *AIService) GetKnowledgeBase(kbID uuid.UUID) (*KnowledgeBaseResponse, error) {
	var kb models.KnowledgeBase
	if err := s.db.First(&kb, kbID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("knowledge base entry not found")
		}
		return nil, fmt.Errorf("failed to get knowledge base entry: %w", err)
	}

	return s.toKnowledgeBaseResponse(&kb), nil
}

// ListKnowledgeBase 获取知识库列表
func (s *AIService) ListKnowledgeBase(page, pageSize int, category, search string) ([]*KnowledgeBaseResponse, int64, error) {
	query := s.db.Model(&models.KnowledgeBase{})

	// 分类过滤
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 搜索条件
	if search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count knowledge base entries: %w", err)
	}

	// 分页查询
	var entries []models.KnowledgeBase
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list knowledge base entries: %w", err)
	}

	// 转换为响应格式
	responses := make([]*KnowledgeBaseResponse, len(entries))
	for i, entry := range entries {
		responses[i] = s.toKnowledgeBaseResponse(&entry)
	}

	return responses, total, nil
}

// buildAlertAnalysisContext 构建告警分析上下文
func (s *AIService) buildAlertAnalysisContext(req *AIAnalysisRequest) string {
	context_str := fmt.Sprintf(`
你是一个专业的系统监控和运维专家，请分析以下告警信息并提供详细的分析报告：

告警信息：
- 目标类型：%s
- 目标ID：%s
- 指标名称：%s
- 当前值：%.2f
- 阈值：%.2f
- 条件：%s
- 严重级别：%s
- 时间：%s
`,
		req.TargetType,
		req.TargetID,
		req.MetricName,
		req.CurrentValue,
		req.Threshold,
		req.Condition,
		req.Severity,
		req.Timestamp.Format("2006-01-02 15:04:05"),
	)

	if req.Tags != nil && len(req.Tags) > 0 {
		context_str += "\n标签信息：\n"
		for key, value := range req.Tags {
			context_str += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}

	context_str += `

请提供以下分析：
1. 根本原因分析
2. 影响评估
3. 解决建议（至少3个具体的操作步骤）
4. 预防措施
5. 严重级别评估（critical/high/medium/low）
6. 置信度评分（0-1之间的小数）

请以JSON格式返回分析结果，格式如下：
{
  "root_cause": "根本原因分析",
  "impact_assessment": "影响评估",
  "recommendations": ["建议1", "建议2", "建议3"],
  "prevention_measures": "预防措施",
  "severity_level": "严重级别",
  "confidence_score": 0.85
}
`

	return context_str
}

// buildPerformanceAnalysisContext 构建性能分析上下文
func (s *AIService) buildPerformanceAnalysisContext(req *AIAnalysisRequest) string {
	context_str := fmt.Sprintf(`
你是一个专业的系统性能分析专家，请分析以下性能数据并提供优化建议：

性能数据：
- 目标类型：%s
- 目标ID：%s
- 指标名称：%s
- 当前值：%.2f
- 时间：%s
`,
		req.TargetType,
		req.TargetID,
		req.MetricName,
		req.CurrentValue,
		req.Timestamp.Format("2006-01-02 15:04:05"),
	)

	if req.Context != nil && len(req.Context) > 0 {
		context_str += "\n上下文信息：\n"
		for key, value := range req.Context {
			context_str += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}

	context_str += `

请提供以下分析：
1. 性能状态评估
2. 瓶颈识别
3. 优化建议（至少3个具体的优化方案）
4. 容量规划建议
5. 风险评估
6. 置信度评分（0-1之间的小数）

请以JSON格式返回分析结果。
`

	return context_str
}

// callOpenAI 调用OpenAI API
func (s *AIService) callOpenAI(prompt, analysisType string) (string, error) {
	ctx := context.Background()

	// 根据分析类型选择模型
	model := s.config.AIModels.OpenAI.Model
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}

	// 构建消息
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "你是一个专业的系统监控和运维专家，具有丰富的故障诊断和性能优化经验。请提供准确、实用的分析和建议。",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	// 创建请求
	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   s.config.AIModels.OpenAI.MaxTokens,
		Temperature: float32(s.config.AIModels.OpenAI.Temperature),
	}

	// 调用API
	resp, err := s.openaiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// parseAIResponse 解析AI响应
func (s *AIService) parseAIResponse(response string) (*ParsedAIResponse, error) {
	// 尝试提取JSON部分
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonStart >= jsonEnd {
		// 如果没有找到JSON，返回默认解析结果
		return &ParsedAIResponse{
			RootCause:       "AI分析结果解析失败",
			Recommendations: []string{"请检查AI模型响应格式"},
			SeverityLevel:   "medium",
			ConfidenceScore: 0.5,
		}, nil
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var parsed ParsedAIResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		// JSON解析失败，返回原始响应
		return &ParsedAIResponse{
			RootCause:       response,
			Recommendations: []string{"请参考完整的AI分析报告"},
			SeverityLevel:   "medium",
			ConfidenceScore: 0.7,
		}, nil
	}

	// 验证和修正数据
	if parsed.SeverityLevel == "" {
		parsed.SeverityLevel = "medium"
	}
	if parsed.ConfidenceScore <= 0 || parsed.ConfidenceScore > 1 {
		parsed.ConfidenceScore = 0.7
	}
	if len(parsed.Recommendations) == 0 {
		parsed.Recommendations = []string{"请参考AI分析报告中的详细建议"}
	}

	return &parsed, nil
}

// getRelevantKnowledge 获取相关知识库信息
func (s *AIService) getRelevantKnowledge(metricName, severity string) (string, error) {
	var entries []models.KnowledgeBase
	query := s.db.Where("metric_types LIKE ? OR severity = ?", "%"+metricName+"%", severity)
	if err := query.Limit(3).Find(&entries).Error; err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "", nil
	}

	var knowledge strings.Builder
	for i, entry := range entries {
		if i > 0 {
			knowledge.WriteString("\n\n")
		}
		knowledge.WriteString(fmt.Sprintf("标题：%s\n内容：%s", entry.Title, entry.Content))
	}

	return knowledge.String(), nil
}

// generateCacheKey 生成缓存键
func (s *AIService) generateCacheKey(analysisType string, req *AIAnalysisRequest) string {
	return fmt.Sprintf("ai_analysis:%s:%s:%s:%s:%.2f",
		analysisType,
		req.TargetType,
		req.TargetID,
		req.MetricName,
		req.CurrentValue,
	)
}

// toAnalysisResponse 转换为分析响应格式
func (s *AIService) toAnalysisResponse(result *models.AIAnalysisResult) *AIAnalysisResponse {
	var tags map[string]interface{}
	var metadata map[string]interface{}

	// 注意：models.AIAnalysisResult中没有Tags字段
	// if result.Tags != "" {
	//	json.Unmarshal([]byte(result.Tags), &tags)
	// }
	if result.Metadata != "" {
		json.Unmarshal([]byte(result.Metadata), &metadata)
	}

	// 注意：models.AIAnalysisResult中没有Recommendations字段
	// 推荐建议需要从Response字段中解析
	recommendations := []string{}

	return &AIAnalysisResponse{
		ID:               result.ID,
		Type:             result.AnalysisType,
		TargetType:       "", // 从metadata中提取或设置默认值
		TargetID:         "", // 从metadata中提取或设置默认值
		AnalysisResult:   result.Response,
		RootCause:        "", // 需要从Response中解析
		Recommendations:  recommendations, // 需要从Response中解析
		SeverityLevel:    "", // 需要从Response中解析
		ConfidenceScore:  result.Confidence,
		Tags:             tags,
		Metadata:         metadata,
		CreatedAt:        result.CreatedAt,
	}
}

// toKnowledgeBaseResponse 转换为知识库响应格式
func (s *AIService) toKnowledgeBaseResponse(kb *models.KnowledgeBase) *KnowledgeBaseResponse {
	var tags []string
	var metricTypes []string
	var metadata map[string]interface{}

	if kb.Tags != "" {
		json.Unmarshal([]byte(kb.Tags), &tags)
	}
	if kb.Metrics != "" {
		json.Unmarshal([]byte(kb.Metrics), &metricTypes)
	}
	// if kb.Metadata != "" {
	//	json.Unmarshal([]byte(kb.Metadata), &metadata)
	// }

	return &KnowledgeBaseResponse{
		ID:          kb.ID,
		Title:       kb.Title,
		Content:     kb.Content,
		Category:    kb.Category,
		Tags:        tags,
		MetricTypes: metricTypes,
		Severity:    kb.Severity,
		Metadata:    metadata,
		CreatedBy:   kb.CreatedBy,
		CreatedAt:   kb.CreatedAt,
		UpdatedAt:   kb.UpdatedAt,
	}
}

// ParsedAIResponse 解析后的AI响应
type ParsedAIResponse struct {
	RootCause         string   `json:"root_cause"`
	ImpactAssessment  string   `json:"impact_assessment"`
	Recommendations   []string `json:"recommendations"`
	PreventionMeasures string   `json:"prevention_measures"`
	SeverityLevel     string   `json:"severity_level"`
	ConfidenceScore   float64  `json:"confidence_score"`
}

// AnomalyDetectionResult 异常检测结果
type AnomalyDetectionResult struct {
	IsAnomaly       bool    `json:"is_anomaly"`
	AnomalyScore    float64 `json:"anomaly_score"`
	Threshold       float64 `json:"threshold"`
	DeviationLevel  string  `json:"deviation_level"`
	HistoricalMean  float64 `json:"historical_mean"`
	HistoricalStdDev float64 `json:"historical_std_dev"`
	Confidence      float64 `json:"confidence"`
}

// TrendPredictionResult 趋势预测结果
type TrendPredictionResult struct {
	TrendDirection   string    `json:"trend_direction"`
	PredictedValues  []float64 `json:"predicted_values"`
	ConfidenceInterval []float64 `json:"confidence_interval"`
	Seasonality      bool      `json:"seasonality"`
	RiskLevel        string    `json:"risk_level"`
	TimeHorizon      string    `json:"time_horizon"`
	Accuracy         float64   `json:"accuracy"`
}

// UpdateKnowledgeBase 更新知识库条目
func (s *AIService) UpdateKnowledgeBase(kbID uuid.UUID, req *KnowledgeBaseRequest) (*KnowledgeBaseResponse, error) {
	// 首先检查知识库条目是否存在
	var kb models.KnowledgeBase
	if err := s.db.First(&kb, kbID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("knowledge base entry not found")
		}
		return nil, fmt.Errorf("failed to get knowledge base entry: %w", err)
	}

	// 序列化标签和指标类型
	tagsJSON, _ := json.Marshal(req.Tags)
	metricTypesJSON, _ := json.Marshal(req.MetricTypes)

	// 更新字段
	updateData := map[string]interface{}{
		"title":    req.Title,
		"content":  req.Content,
		"category": req.Category,
		"tags":     string(tagsJSON),
		"metrics":  string(metricTypesJSON),
		"severity": req.Severity,
	}

	if err := s.db.Model(&kb).Updates(updateData).Error; err != nil {
		return nil, fmt.Errorf("failed to update knowledge base entry: %w", err)
	}

	// 重新获取更新后的数据
	if err := s.db.First(&kb, kbID).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated knowledge base entry: %w", err)
	}

	return s.toKnowledgeBaseResponse(&kb), nil
}

// DeleteKnowledgeBase 删除知识库条目
func (s *AIService) DeleteKnowledgeBase(kbID uuid.UUID) error {
	// 首先检查知识库条目是否存在
	var kb models.KnowledgeBase
	if err := s.db.First(&kb, kbID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("knowledge base entry not found")
		}
		return fmt.Errorf("failed to get knowledge base entry: %w", err)
	}

	// 删除知识库条目
	if err := s.db.Delete(&kb).Error; err != nil {
		return fmt.Errorf("failed to delete knowledge base entry: %w", err)
	}

	return nil
}

// detectAnomaly 异常检测算法
func (s *AIService) detectAnomaly(req *AIAnalysisRequest) (*AnomalyDetectionResult, error) {
	// 获取历史数据进行异常检测
	historicalData, err := s.getHistoricalMetrics(req.TargetType, req.TargetID, req.MetricName, 30) // 获取30天历史数据
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	if len(historicalData) < 10 {
		// 数据不足，返回基础检测结果
		return &AnomalyDetectionResult{
			IsAnomaly:      false,
			AnomalyScore:   0.0,
			Threshold:      2.0,
			DeviationLevel: "normal",
			Confidence:     0.5,
		}, nil
	}

	// 计算统计指标
	mean := s.calculateMean(historicalData)
	stdDev := s.calculateStdDev(historicalData, mean)

	// Z-Score异常检测
	zScore := (req.CurrentValue - mean) / stdDev
	anomalyScore := math.Abs(zScore)

	// 动态阈值调整
	threshold := 2.0
	if req.Severity == "critical" {
		threshold = 1.5 // 更敏感的检测
	} else if req.Severity == "low" {
		threshold = 2.5 // 较不敏感的检测
	}

	// 确定偏差级别
	deviationLevel := "normal"
	if anomalyScore > threshold {
		if anomalyScore > threshold*2 {
			deviationLevel = "severe"
		} else if anomalyScore > threshold*1.5 {
			deviationLevel = "high"
		} else {
			deviationLevel = "moderate"
		}
	}

	// 计算置信度
	confidence := math.Min(0.95, float64(len(historicalData))/100.0+0.5)

	return &AnomalyDetectionResult{
		IsAnomaly:        anomalyScore > threshold,
		AnomalyScore:     anomalyScore,
		Threshold:        threshold,
		DeviationLevel:   deviationLevel,
		HistoricalMean:   mean,
		HistoricalStdDev: stdDev,
		Confidence:       confidence,
	}, nil
}

// predictTrend 趋势预测分析
func (s *AIService) predictTrend(req *AIAnalysisRequest) (*TrendPredictionResult, error) {
	// 获取历史数据进行趋势分析
	historicalData, err := s.getHistoricalMetrics(req.TargetType, req.TargetID, req.MetricName, 7) // 获取7天历史数据
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	if len(historicalData) < 5 {
		// 数据不足，返回基础预测结果
		return &TrendPredictionResult{
			TrendDirection: "stable",
			PredictedValues: []float64{req.CurrentValue},
			RiskLevel:      "low",
			TimeHorizon:    "1h",
			Accuracy:       0.5,
		}, nil
	}

	// 简单线性回归预测
	n := float64(len(historicalData))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, value := range historicalData {
		x := float64(i)
		sumX += x
		sumY += value
		sumXY += x * value
		sumX2 += x * x
	}

	// 计算回归系数
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// 预测未来值
	predictedValues := make([]float64, 6) // 预测未来6个时间点
	for i := 0; i < 6; i++ {
		x := n + float64(i)
		predictedValues[i] = slope*x + intercept
	}

	// 确定趋势方向
	trendDirection := "stable"
	if slope > 0.1 {
		trendDirection = "increasing"
	} else if slope < -0.1 {
		trendDirection = "decreasing"
	}

	// 评估风险级别
	riskLevel := "low"
	if math.Abs(slope) > 1.0 {
		riskLevel = "high"
	} else if math.Abs(slope) > 0.5 {
		riskLevel = "medium"
	}

	// 计算预测准确度
	accuracy := math.Max(0.6, 1.0-math.Abs(slope)/10.0)

	return &TrendPredictionResult{
		TrendDirection:     trendDirection,
		PredictedValues:    predictedValues,
		ConfidenceInterval: []float64{0.8, 0.9}, // 简化的置信区间
		Seasonality:        s.detectSeasonality(historicalData),
		RiskLevel:          riskLevel,
		TimeHorizon:        "6h",
		Accuracy:           accuracy,
	}, nil
}

// buildEnhancedAnalysisContext 构建增强的分析上下文
func (s *AIService) buildEnhancedAnalysisContext(req *AIAnalysisRequest, anomaly *AnomalyDetectionResult, prediction *TrendPredictionResult) string {
	context_str := fmt.Sprintf(`
你是一个专业的AI驱动系统监控和运维专家，请基于以下综合信息进行深度分析：

=== 基础告警信息 ===
- 目标类型：%s
- 目标ID：%s
- 指标名称：%s
- 当前值：%.2f
- 阈值：%.2f
- 条件：%s
- 严重级别：%s
- 时间：%s

=== AI异常检测结果 ===
- 异常状态：%t
- 异常评分：%.2f
- 检测阈值：%.2f
- 偏差级别：%s
- 历史均值：%.2f
- 历史标准差：%.2f
- 检测置信度：%.2f

=== AI趋势预测分析 ===
- 趋势方向：%s
- 风险级别：%s
- 预测准确度：%.2f
- 时间范围：%s
- 季节性模式：%t
`,
		req.TargetType, req.TargetID, req.MetricName,
		req.CurrentValue, req.Threshold, req.Condition, req.Severity,
		req.Timestamp.Format("2006-01-02 15:04:05"),
		anomaly.IsAnomaly, anomaly.AnomalyScore, anomaly.Threshold,
		anomaly.DeviationLevel, anomaly.HistoricalMean, anomaly.HistoricalStdDev, anomaly.Confidence,
		prediction.TrendDirection, prediction.RiskLevel, prediction.Accuracy,
		prediction.TimeHorizon, prediction.Seasonality,
	)

	if req.Tags != nil && len(req.Tags) > 0 {
		context_str += "\n=== 标签信息 ===\n"
		for key, value := range req.Tags {
			context_str += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}

	context_str += `

=== 请提供以下深度分析 ===
1. 智能根因分析（结合异常检测和趋势预测结果）
2. 多维度影响评估（业务影响、技术影响、用户体验影响）
3. 分层解决方案（紧急处理、中期优化、长期预防）
4. 预测性维护建议（基于趋势分析）
5. 智能告警优化建议
6. 综合风险评估（critical/high/medium/low）
7. AI分析置信度评分（0-1之间的小数）

请以JSON格式返回分析结果，格式如下：
{
  "root_cause": "智能根因分析结果",
  "impact_assessment": "多维度影响评估",
  "recommendations": ["紧急处理方案", "中期优化方案", "长期预防方案"],
  "prevention_measures": "预测性维护建议",
  "severity_level": "综合风险级别",
  "confidence_score": 0.85
}
`

	return context_str
}

// 辅助函数
func (s *AIService) getHistoricalMetrics(targetType, targetID, metricName string, days int) ([]float64, error) {
	// 模拟获取历史数据
	// 在实际实现中，这里应该从时序数据库（如InfluxDB）中获取数据
	data := make([]float64, days*24) // 每小时一个数据点
	base := 50.0
	for i := range data {
		// 模拟带有趋势和噪声的数据
		trend := float64(i) * 0.1
		noise := (rand.Float64() - 0.5) * 10
		data[i] = base + trend + noise
	}
	return data, nil
}

func (s *AIService) calculateMean(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func (s *AIService) calculateStdDev(data []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += (v - mean) * (v - mean)
	}
	return math.Sqrt(sum / float64(len(data)))
}

func (s *AIService) detectSeasonality(data []float64) bool {
	// 简化的季节性检测
	if len(data) < 24 {
		return false
	}
	// 检查是否存在周期性模式
	return true // 简化实现
}

// GetKnowledgeBaseStats 获取知识库统计信息
func (s *AIService) GetKnowledgeBaseStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总条目数
	var totalEntries int64
	if err := s.db.Model(&models.KnowledgeBase{}).Count(&totalEntries).Error; err != nil {
		return nil, fmt.Errorf("failed to count total knowledge base entries: %w", err)
	}
	stats["total_entries"] = totalEntries

	// 按分类统计
	type CategoryStat struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	var categoryStats []CategoryStat
	if err := s.db.Model(&models.KnowledgeBase{}).Select("category, COUNT(*) as count").Group("category").Scan(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}

	categories := make(map[string]int64)
	for _, stat := range categoryStats {
		categories[stat.Category] = stat.Count
	}
	stats["categories"] = categories

	// 按严重级别统计
	type SeverityStat struct {
		Severity string `json:"severity"`
		Count    int64  `json:"count"`
	}
	var severityStats []SeverityStat
	if err := s.db.Model(&models.KnowledgeBase{}).Select("severity, COUNT(*) as count").Group("severity").Scan(&severityStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get severity stats: %w", err)
	}

	severities := make(map[string]int64)
	for _, stat := range severityStats {
		severities[stat.Severity] = stat.Count
	}
	stats["severities"] = severities

	// 最近7天创建的条目数
	var recentEntries int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := s.db.Model(&models.KnowledgeBase{}).Where("created_at >= ?", sevenDaysAgo).Count(&recentEntries).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent entries: %w", err)
	}
	stats["recent_entries"] = recentEntries

	// 今日创建的条目数
	var todayEntries int64
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&models.KnowledgeBase{}).Where("created_at >= ?", today).Count(&todayEntries).Error; err != nil {
		return nil, fmt.Errorf("failed to count today entries: %w", err)
	}
	stats["today_entries"] = todayEntries

	return stats, nil
}

// ExportKnowledgeBase 导出知识库为Markdown格式
func (s *AIService) ExportKnowledgeBase(category string) (string, error) {
	query := s.db.Model(&models.KnowledgeBase{})

	// 分类过滤
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 获取所有知识库条目
	var entries []models.KnowledgeBase
	if err := query.Order("category, created_at DESC").Find(&entries).Error; err != nil {
		return "", fmt.Errorf("failed to get knowledge base entries: %w", err)
	}

	if len(entries) == 0 {
		return "", errors.New("no knowledge base entries found")
	}

	// 构建Markdown内容
	var markdown strings.Builder
	markdown.WriteString("# 知识库导出\n\n")
	markdown.WriteString(fmt.Sprintf("导出时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	markdown.WriteString(fmt.Sprintf("总条目数: %d\n\n", len(entries)))

	// 按分类分组
	currentCategory := ""
	for _, entry := range entries {
		if entry.Category != currentCategory {
			currentCategory = entry.Category
			markdown.WriteString(fmt.Sprintf("## %s\n\n", currentCategory))
		}

		// 解析标签和指标类型
		var tags []string
		var metricTypes []string
		if entry.Tags != "" {
			json.Unmarshal([]byte(entry.Tags), &tags)
		}
		if entry.Metrics != "" {
			json.Unmarshal([]byte(entry.Metrics), &metricTypes)
		}

		// 写入条目信息
		markdown.WriteString(fmt.Sprintf("### %s\n\n", entry.Title))
		markdown.WriteString(fmt.Sprintf("**严重级别**: %s\n\n", entry.Severity))
		if len(tags) > 0 {
			markdown.WriteString(fmt.Sprintf("**标签**: %s\n\n", strings.Join(tags, ", ")))
		}
		if len(metricTypes) > 0 {
			markdown.WriteString(fmt.Sprintf("**指标类型**: %s\n\n", strings.Join(metricTypes, ", ")))
		}
		markdown.WriteString(fmt.Sprintf("**创建时间**: %s\n\n", entry.CreatedAt.Format("2006-01-02 15:04:05")))
		markdown.WriteString(fmt.Sprintf("**内容**:\n\n%s\n\n", entry.Content))
		markdown.WriteString("---\n\n")
	}

	return markdown.String(), nil
}

// GetAIStats 获取AI分析统计
func (s *AIService) GetAIStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总分析次数
	var totalAnalysis int64
	if err := s.db.Model(&models.AIAnalysisResult{}).Count(&totalAnalysis).Error; err != nil {
		return nil, fmt.Errorf("failed to count total analysis: %w", err)
	}
	stats["total_analysis"] = totalAnalysis

	// 按类型统计
	typeStats := make(map[string]int64)
	types := []string{"alert_analysis", "performance_analysis", "trend_analysis", "capacity_planning"}
	for _, analysisType := range types {
		var count int64
		if err := s.db.Model(&models.AIAnalysisResult{}).Where("analysis_type = ?", analysisType).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count %s analysis: %w", analysisType, err)
		}
		typeStats[analysisType] = count
	}
	stats["type_stats"] = typeStats

	// 今日分析次数
	var todayAnalysis int64
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&models.AIAnalysisResult{}).Where("created_at >= ?", today).Count(&todayAnalysis).Error; err != nil {
		return nil, fmt.Errorf("failed to count today analysis: %w", err)
	}
	stats["today_analysis"] = todayAnalysis

	// 知识库条目数
	var knowledgeCount int64
	if err := s.db.Model(&models.KnowledgeBase{}).Count(&knowledgeCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count knowledge base: %w", err)
	}
	stats["knowledge_count"] = knowledgeCount

	return stats, nil
}