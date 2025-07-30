package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ai-monitor/internal/cache"
	"ai-monitor/internal/config"
	"ai-monitor/internal/models"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"gorm.io/gorm"
)

// MonitoringService 监控服务
type MonitoringService struct {
	db             *gorm.DB
	cacheManager   *cache.CacheManager
	config         *config.Config
	prometheusAPI  v1.API
	alertService   *AlertService
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(db *gorm.DB, cacheManager *cache.CacheManager, config *config.Config, alertService *AlertService) (*MonitoringService, error) {
	// 创建Prometheus客户端
	client, err := api.NewClient(api.Config{
		Address: config.Prometheus.URL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %w", err)
	}

	prometheusAPI := v1.NewAPI(client)

	return &MonitoringService{
		db:             db,
		cacheManager:   cacheManager,
		config:         config,
		prometheusAPI:  prometheusAPI,
		alertService:   alertService,
	}, nil
}

// CreateTargetRequest 创建监控目标请求
type CreateTargetRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Type        string                 `json:"type" binding:"required,oneof=host service application database"`
	Address     string                 `json:"address" binding:"required"`
	Port        int                    `json:"port"`
	Description string                 `json:"description"`
	Tags        map[string]interface{} `json:"tags"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
}

// UpdateTargetRequest 更新监控目标请求
type UpdateTargetRequest struct {
	Name        string                 `json:"name"`
	Address     string                 `json:"address"`
	Port        *int                   `json:"port"`
	Description string                 `json:"description"`
	Tags        map[string]interface{} `json:"tags"`
	Config      map[string]interface{} `json:"config"`
	Enabled     *bool                  `json:"enabled"`
}

// TargetResponse 监控目标响应
type TargetResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Address     string                 `json:"address"`
	Port        int                    `json:"port"`
	Description string                 `json:"description"`
	Tags        map[string]interface{} `json:"tags"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
	Status      string                 `json:"status"`
	LastSeen    *time.Time             `json:"last_seen"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// MetricQueryRequest 指标查询请求
type MetricQueryRequest struct {
	Query     string    `json:"query" binding:"required"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Step      string    `json:"step"`
}

// MetricQueryResponse 指标查询响应
type MetricQueryResponse struct {
	MetricName string                   `json:"metric_name"`
	Data       []MetricDataPoint        `json:"data"`
	Labels     map[string]string        `json:"labels"`
	Metadata   map[string]interface{}   `json:"metadata"`
}

// MetricDataPoint 指标数据点
type MetricDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// DashboardRequest 仪表板请求
type DashboardRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config" binding:"required"`
	Tags        []string               `json:"tags"`
	IsPublic    bool                   `json:"is_public"`
}

// DashboardResponse 仪表板响应
type DashboardResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Tags        []string               `json:"tags"`
	IsPublic    bool                   `json:"is_public"`
	CreatedBy   uuid.UUID              `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPU     CPUMetrics     `json:"cpu"`
	Memory  MemoryMetrics  `json:"memory"`
	Disk    DiskMetrics    `json:"disk"`
	Network NetworkMetrics `json:"network"`
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	LoadAvg1     float64 `json:"load_avg_1"`
	LoadAvg5     float64 `json:"load_avg_5"`
	LoadAvg15    float64 `json:"load_avg_15"`
	Cores        int     `json:"cores"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	UsagePercent  float64 `json:"usage_percent"`
	TotalBytes    int64   `json:"total_bytes"`
	UsedBytes     int64   `json:"used_bytes"`
	AvailableBytes int64   `json:"available_bytes"`
	BufferBytes   int64   `json:"buffer_bytes"`
	CachedBytes   int64   `json:"cached_bytes"`
}

// DiskMetrics 磁盘指标
type DiskMetrics struct {
	UsagePercent   float64 `json:"usage_percent"`
	TotalBytes     int64   `json:"total_bytes"`
	UsedBytes      int64   `json:"used_bytes"`
	AvailableBytes int64   `json:"available_bytes"`
	ReadIOPS       float64 `json:"read_iops"`
	WriteIOPS      float64 `json:"write_iops"`
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	BytesReceived    int64   `json:"bytes_received"`
	BytesSent        int64   `json:"bytes_sent"`
	PacketsReceived  int64   `json:"packets_received"`
	PacketsSent      int64   `json:"packets_sent"`
	ReceiveThroughput float64 `json:"receive_throughput"`
	TransmitThroughput float64 `json:"transmit_throughput"`
}

// CreateTarget 创建监控目标
func (s *MonitoringService) CreateTarget(req *CreateTargetRequest) (*TargetResponse, error) {
	// 检查目标名称是否已存在
	var existingTarget models.MonitoringTarget
	if err := s.db.Where("name = ?", req.Name).First(&existingTarget).Error; err == nil {
		return nil, errors.New("monitoring target name already exists")
	}

	// 序列化标签和配置
	tagsJSON, _ := json.Marshal(req.Tags)
	configJSON, _ := json.Marshal(req.Config)

	// 创建监控目标
	target := models.MonitoringTarget{
		Name:        req.Name,
		Type:        req.Type,
		Address:     req.Address,
		Port:        req.Port,
		Labels:      string(tagsJSON),
		Metrics:     string(configJSON),
		Status:      "active",
	}

	if err := s.db.Create(&target).Error; err != nil {
		return nil, fmt.Errorf("failed to create monitoring target: %w", err)
	}

	return s.toTargetResponse(&target), nil
}

// GetTarget 获取监控目标
func (s *MonitoringService) GetTarget(targetID uuid.UUID) (*TargetResponse, error) {
	var target models.MonitoringTarget
	if err := s.db.First(&target, targetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("monitoring target not found")
		}
		return nil, fmt.Errorf("failed to get monitoring target: %w", err)
	}

	return s.toTargetResponse(&target), nil
}

// UpdateTarget 更新监控目标
func (s *MonitoringService) UpdateTarget(targetID uuid.UUID, req *UpdateTargetRequest) (*TargetResponse, error) {
	var target models.MonitoringTarget
	if err := s.db.First(&target, targetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("monitoring target not found")
		}
		return nil, fmt.Errorf("failed to get monitoring target: %w", err)
	}

	// 更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		// 检查名称是否已存在
		var existingTarget models.MonitoringTarget
		if err := s.db.Where("name = ? AND id != ?", req.Name, targetID).First(&existingTarget).Error; err == nil {
			return nil, errors.New("monitoring target name already exists")
		}
		updates["name"] = req.Name
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		updates["labels"] = string(tagsJSON)
	}
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		updates["metrics"] = string(configJSON)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&target).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update monitoring target: %w", err)
		}
	}

	// 重新加载数据
	if err := s.db.First(&target, targetID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload monitoring target: %w", err)
	}

	return s.toTargetResponse(&target), nil
}

// DeleteTarget 删除监控目标
func (s *MonitoringService) DeleteTarget(targetID uuid.UUID) error {
	var target models.MonitoringTarget
	if err := s.db.First(&target, targetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("monitoring target not found")
		}
		return fmt.Errorf("failed to get monitoring target: %w", err)
	}

	// 软删除目标
	if err := s.db.Delete(&target).Error; err != nil {
		return fmt.Errorf("failed to delete monitoring target: %w", err)
	}

	return nil
}

// ListTargets 获取监控目标列表
func (s *MonitoringService) ListTargets(page, pageSize int, targetType string, enabled *bool) ([]*TargetResponse, int64, error) {
	query := s.db.Model(&models.MonitoringTarget{})

	// 类型过滤
	if targetType != "" {
		query = query.Where("type = ?", targetType)
	}

	// 状态过滤（使用 status 字段代替 enabled）
	if enabled != nil {
		if *enabled {
			query = query.Where("status = ?", "active")
		} else {
			query = query.Where("status != ?", "active")
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count monitoring targets: %w", err)
	}

	// 分页查询
	var targets []models.MonitoringTarget
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&targets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list monitoring targets: %w", err)
	}

	// 转换为响应格式
	responses := make([]*TargetResponse, len(targets))
	for i, target := range targets {
		responses[i] = s.toTargetResponse(&target)
	}

	return responses, total, nil
}

// QueryMetrics 查询指标数据
func (s *MonitoringService) QueryMetrics(req *MetricQueryRequest) ([]*MetricQueryResponse, error) {
	ctx := context.Background()

	// 检查缓存
	cacheKey := s.generateMetricCacheKey(req)
	if s.cacheManager != nil {
		var responses []*MetricQueryResponse
		if err := s.cacheManager.Get(ctx, cacheKey, &responses); err == nil {
			return responses, nil
		}
	}

	var result model.Value
	var err error

	// 根据时间范围选择查询方式
	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		// 即时查询
		result, _, err = s.prometheusAPI.Query(ctx, req.Query, time.Now())
	} else {
		// 范围查询
		step, _ := time.ParseDuration(req.Step)
		if step == 0 {
			step = time.Minute
		}
		result, _, err = s.prometheusAPI.QueryRange(ctx, req.Query, v1.Range{
			Start: req.StartTime,
			End:   req.EndTime,
			Step:  step,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query prometheus: %w", err)
	}

	// 解析结果
	responses := s.parsePrometheusResult(result)

	// 缓存结果
	if s.cacheManager != nil {
		if data, err := json.Marshal(responses); err == nil {
			s.cacheManager.Set(ctx, cacheKey, string(data), 5*time.Minute)
		}
	}

	return responses, nil
}

// GetSystemMetrics 获取系统指标
func (s *MonitoringService) GetSystemMetrics(targetID string) (*SystemMetrics, error) {
	ctx := context.Background()

	// 检查缓存
	cacheKey := cache.SystemMetricsCacheKey(targetID)
	if s.cacheManager != nil {
		var metrics SystemMetrics
		if err := s.cacheManager.Get(ctx, cacheKey, &metrics); err == nil {
			return &metrics, nil
		}
	}

	metrics := &SystemMetrics{}

	// 查询CPU指标
	cpuMetrics, err := s.getCPUMetrics(targetID)
	if err == nil {
		metrics.CPU = *cpuMetrics
	}

	// 查询内存指标
	memoryMetrics, err := s.getMemoryMetrics(targetID)
	if err == nil {
		metrics.Memory = *memoryMetrics
	}

	// 查询磁盘指标
	diskMetrics, err := s.getDiskMetrics(targetID)
	if err == nil {
		metrics.Disk = *diskMetrics
	}

	// 查询网络指标
	networkMetrics, err := s.getNetworkMetrics(targetID)
	if err == nil {
		metrics.Network = *networkMetrics
	}

	// 缓存结果
	if s.cacheManager != nil {
		if data, err := json.Marshal(metrics); err == nil {
			s.cacheManager.Set(ctx, cacheKey, string(data), 2*time.Minute)
		}
	}

	return metrics, nil
}

// CreateDashboard 创建仪表板
func (s *MonitoringService) CreateDashboard(req *DashboardRequest, createdBy uuid.UUID) (*DashboardResponse, error) {
	// 检查仪表板名称是否已存在
	var existingDashboard models.Dashboard
	if err := s.db.Where("name = ?", req.Name).First(&existingDashboard).Error; err == nil {
		return nil, errors.New("dashboard name already exists")
	}

	// 序列化配置和标签
	configJSON, _ := json.Marshal(req.Config)
	tagsJSON, _ := json.Marshal(req.Tags)

	// 创建仪表板
	dashboard := models.Dashboard{
		Name:        req.Name,
		Description: req.Description,
		Config:      string(configJSON),
		Tags:        string(tagsJSON),
		IsPublic:    req.IsPublic,
		CreatedBy:   createdBy,
	}

	if err := s.db.Create(&dashboard).Error; err != nil {
		return nil, fmt.Errorf("failed to create dashboard: %w", err)
	}

	return s.toDashboardResponse(&dashboard), nil
}

// GetDashboard 获取仪表板
func (s *MonitoringService) GetDashboard(dashboardID uuid.UUID) (*DashboardResponse, error) {
	var dashboard models.Dashboard
	if err := s.db.First(&dashboard, dashboardID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dashboard not found")
		}
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	return s.toDashboardResponse(&dashboard), nil
}

// ListDashboards 获取仪表板列表
func (s *MonitoringService) ListDashboards(page, pageSize int, isPublic *bool, createdBy *uuid.UUID) ([]*DashboardResponse, int64, error) {
	query := s.db.Model(&models.Dashboard{})

	// 公开状态过滤
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	// 创建者过滤
	if createdBy != nil {
		query = query.Where("created_by = ?", *createdBy)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count dashboards: %w", err)
	}

	// 分页查询
	var dashboards []models.Dashboard
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dashboards).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list dashboards: %w", err)
	}

	// 转换为响应格式
	responses := make([]*DashboardResponse, len(dashboards))
	for i, dashboard := range dashboards {
		responses[i] = s.toDashboardResponse(&dashboard)
	}

	return responses, total, nil
}

// StoreMetricData 存储指标数据
func (s *MonitoringService) StoreMetricData(targetID, metricName string, value float64, tags map[string]interface{}) error {
	// 序列化标签
	tagsJSON, _ := json.Marshal(tags)

	// 解析 targetID 为 UUID
	targetUUID, err := uuid.Parse(targetID)
	if err != nil {
		return fmt.Errorf("invalid target ID: %w", err)
	}

	// 创建指标数据记录
	metricData := models.MetricData{
		TargetID:  targetUUID,
		Metric:    metricName,
		Value:     value,
		Labels:    string(tagsJSON),
		Timestamp: time.Now(),
	}

	if err := s.db.Create(&metricData).Error; err != nil {
		return fmt.Errorf("failed to store metric data: %w", err)
	}

	// 触发告警检查
	if s.alertService != nil {
		alertData := &MetricData{
			TargetType: "host", // 默认类型，可以根据实际情况调整
			TargetID:   targetID,
			MetricName: metricName,
			Value:      value,
			Tags:       tags,
			Timestamp:  time.Now(),
		}
		go s.alertService.ProcessMetricData(alertData)
	}

	return nil
}

// getCPUMetrics 获取CPU指标
func (s *MonitoringService) getCPUMetrics(targetID string) (*CPUMetrics, error) {
	ctx := context.Background()

	// CPU使用率
	usageQuery := fmt.Sprintf(`100 - (avg(irate(node_cpu_seconds_total{instance="%s",mode="idle"}[5m])) * 100)`, targetID)
	usageResult, _, err := s.prometheusAPI.Query(ctx, usageQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query CPU usage: %w", err)
	}

	// 负载平均值
	loadAvg1Query := fmt.Sprintf(`node_load1{instance="%s"}`, targetID)
	loadAvg1Result, _, err := s.prometheusAPI.Query(ctx, loadAvg1Query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query load avg 1: %w", err)
	}

	loadAvg5Query := fmt.Sprintf(`node_load5{instance="%s"}`, targetID)
	loadAvg5Result, _, err := s.prometheusAPI.Query(ctx, loadAvg5Query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query load avg 5: %w", err)
	}

	loadAvg15Query := fmt.Sprintf(`node_load15{instance="%s"}`, targetID)
	loadAvg15Result, _, err := s.prometheusAPI.Query(ctx, loadAvg15Query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query load avg 15: %w", err)
	}

	// CPU核心数
	coresQuery := fmt.Sprintf(`count(count(node_cpu_seconds_total{instance="%s"}) by (cpu))`, targetID)
	coresResult, _, err := s.prometheusAPI.Query(ctx, coresQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query CPU cores: %w", err)
	}

	return &CPUMetrics{
		UsagePercent: s.extractFloatValue(usageResult),
		LoadAvg1:     s.extractFloatValue(loadAvg1Result),
		LoadAvg5:     s.extractFloatValue(loadAvg5Result),
		LoadAvg15:    s.extractFloatValue(loadAvg15Result),
		Cores:        int(s.extractFloatValue(coresResult)),
	}, nil
}

// getMemoryMetrics 获取内存指标
func (s *MonitoringService) getMemoryMetrics(targetID string) (*MemoryMetrics, error) {
	ctx := context.Background()

	// 内存总量
	totalQuery := fmt.Sprintf(`node_memory_MemTotal_bytes{instance="%s"}`, targetID)
	totalResult, _, err := s.prometheusAPI.Query(ctx, totalQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query memory total: %w", err)
	}

	// 可用内存
	availableQuery := fmt.Sprintf(`node_memory_MemAvailable_bytes{instance="%s"}`, targetID)
	availableResult, _, err := s.prometheusAPI.Query(ctx, availableQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query memory available: %w", err)
	}

	// 缓冲区
	bufferQuery := fmt.Sprintf(`node_memory_Buffers_bytes{instance="%s"}`, targetID)
	bufferResult, _, err := s.prometheusAPI.Query(ctx, bufferQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query memory buffers: %w", err)
	}

	// 缓存
	cachedQuery := fmt.Sprintf(`node_memory_Cached_bytes{instance="%s"}`, targetID)
	cachedResult, _, err := s.prometheusAPI.Query(ctx, cachedQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query memory cached: %w", err)
	}

	total := s.extractFloatValue(totalResult)
	available := s.extractFloatValue(availableResult)
	used := total - available
	usagePercent := (used / total) * 100

	return &MemoryMetrics{
		UsagePercent:   usagePercent,
		TotalBytes:     int64(total),
		UsedBytes:      int64(used),
		AvailableBytes: int64(available),
		BufferBytes:    int64(s.extractFloatValue(bufferResult)),
		CachedBytes:    int64(s.extractFloatValue(cachedResult)),
	}, nil
}

// getDiskMetrics 获取磁盘指标
func (s *MonitoringService) getDiskMetrics(targetID string) (*DiskMetrics, error) {
	ctx := context.Background()

	// 磁盘总量
	totalQuery := fmt.Sprintf(`node_filesystem_size_bytes{instance="%s",fstype!="tmpfs"}`, targetID)
	totalResult, _, err := s.prometheusAPI.Query(ctx, totalQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query disk total: %w", err)
	}

	// 磁盘可用空间
	availableQuery := fmt.Sprintf(`node_filesystem_avail_bytes{instance="%s",fstype!="tmpfs"}`, targetID)
	availableResult, _, err := s.prometheusAPI.Query(ctx, availableQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query disk available: %w", err)
	}

	// 读IOPS
	readIOPSQuery := fmt.Sprintf(`rate(node_disk_reads_completed_total{instance="%s"}[5m])`, targetID)
	readIOPSResult, _, err := s.prometheusAPI.Query(ctx, readIOPSQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query disk read IOPS: %w", err)
	}

	// 写IOPS
	writeIOPSQuery := fmt.Sprintf(`rate(node_disk_writes_completed_total{instance="%s"}[5m])`, targetID)
	writeIOPSResult, _, err := s.prometheusAPI.Query(ctx, writeIOPSQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query disk write IOPS: %w", err)
	}

	total := s.extractFloatValue(totalResult)
	available := s.extractFloatValue(availableResult)
	used := total - available
	usagePercent := (used / total) * 100

	return &DiskMetrics{
		UsagePercent:   usagePercent,
		TotalBytes:     int64(total),
		UsedBytes:      int64(used),
		AvailableBytes: int64(available),
		ReadIOPS:       s.extractFloatValue(readIOPSResult),
		WriteIOPS:      s.extractFloatValue(writeIOPSResult),
	}, nil
}

// getNetworkMetrics 获取网络指标
func (s *MonitoringService) getNetworkMetrics(targetID string) (*NetworkMetrics, error) {
	ctx := context.Background()

	// 接收字节数
	bytesReceivedQuery := fmt.Sprintf(`node_network_receive_bytes_total{instance="%s",device!="lo"}`, targetID)
	bytesReceivedResult, _, err := s.prometheusAPI.Query(ctx, bytesReceivedQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query network bytes received: %w", err)
	}

	// 发送字节数
	bytesSentQuery := fmt.Sprintf(`node_network_transmit_bytes_total{instance="%s",device!="lo"}`, targetID)
	bytesSentResult, _, err := s.prometheusAPI.Query(ctx, bytesSentQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query network bytes sent: %w", err)
	}

	// 接收包数
	packetsReceivedQuery := fmt.Sprintf(`node_network_receive_packets_total{instance="%s",device!="lo"}`, targetID)
	packetsReceivedResult, _, err := s.prometheusAPI.Query(ctx, packetsReceivedQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query network packets received: %w", err)
	}

	// 发送包数
	packetsSentQuery := fmt.Sprintf(`node_network_transmit_packets_total{instance="%s",device!="lo"}`, targetID)
	packetsSentResult, _, err := s.prometheusAPI.Query(ctx, packetsSentQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query network packets sent: %w", err)
	}

	// 接收吞吐量
	receiveThroughputQuery := fmt.Sprintf(`rate(node_network_receive_bytes_total{instance="%s",device!="lo"}[5m])`, targetID)
	receiveThroughputResult, _, err := s.prometheusAPI.Query(ctx, receiveThroughputQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query network receive throughput: %w", err)
	}

	// 发送吞吐量
	transmitThroughputQuery := fmt.Sprintf(`rate(node_network_transmit_bytes_total{instance="%s",device!="lo"}[5m])`, targetID)
	transmitThroughputResult, _, err := s.prometheusAPI.Query(ctx, transmitThroughputQuery, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query network transmit throughput: %w", err)
	}

	return &NetworkMetrics{
		BytesReceived:      int64(s.extractFloatValue(bytesReceivedResult)),
		BytesSent:          int64(s.extractFloatValue(bytesSentResult)),
		PacketsReceived:    int64(s.extractFloatValue(packetsReceivedResult)),
		PacketsSent:        int64(s.extractFloatValue(packetsSentResult)),
		ReceiveThroughput:  s.extractFloatValue(receiveThroughputResult),
		TransmitThroughput: s.extractFloatValue(transmitThroughputResult),
	}, nil
}

// extractFloatValue 从Prometheus结果中提取浮点值
func (s *MonitoringService) extractFloatValue(result model.Value) float64 {
	switch v := result.(type) {
	case model.Vector:
		if len(v) > 0 {
			return float64(v[0].Value)
		}
	case *model.Scalar:
		return float64(v.Value)
	}
	return 0
}

// parsePrometheusResult 解析Prometheus查询结果
func (s *MonitoringService) parsePrometheusResult(result model.Value) []*MetricQueryResponse {
	var responses []*MetricQueryResponse

	switch v := result.(type) {
	case model.Vector:
		for _, sample := range v {
			labels := make(map[string]string)
			for k, v := range sample.Metric {
				labels[string(k)] = string(v)
			}

			response := &MetricQueryResponse{
				MetricName: string(sample.Metric["__name__"]),
				Data: []MetricDataPoint{
					{
						Timestamp: sample.Timestamp.Time(),
						Value:     float64(sample.Value),
					},
				},
				Labels: labels,
			}
			responses = append(responses, response)
		}
	case model.Matrix:
		for _, sampleStream := range v {
			labels := make(map[string]string)
			for k, v := range sampleStream.Metric {
				labels[string(k)] = string(v)
			}

			data := make([]MetricDataPoint, len(sampleStream.Values))
			for i, pair := range sampleStream.Values {
				data[i] = MetricDataPoint{
					Timestamp: pair.Timestamp.Time(),
					Value:     float64(pair.Value),
				}
			}

			response := &MetricQueryResponse{
				MetricName: string(sampleStream.Metric["__name__"]),
				Data:       data,
				Labels:     labels,
			}
			responses = append(responses, response)
		}
	}

	return responses
}

// generateMetricCacheKey 生成指标缓存键
func (s *MonitoringService) generateMetricCacheKey(req *MetricQueryRequest) string {
	return fmt.Sprintf("metrics:%s:%d:%d:%s",
		req.Query,
		req.StartTime.Unix(),
		req.EndTime.Unix(),
		req.Step,
	)
}

// toTargetResponse 转换为监控目标响应格式
func (s *MonitoringService) toTargetResponse(target *models.MonitoringTarget) *TargetResponse {
	var tags map[string]interface{}
	var config map[string]interface{}

	if target.Labels != "" {
		json.Unmarshal([]byte(target.Labels), &tags)
	}
	if target.Metrics != "" {
		json.Unmarshal([]byte(target.Metrics), &config)
	}

	// 根据 status 字段确定 enabled 状态
	enabled := target.Status == "active"

	return &TargetResponse{
		ID:          target.ID,
		Name:        target.Name,
		Type:        target.Type,
		Address:     target.Address,
		Port:        target.Port,
		Description: "", // MonitoringTarget 模型中没有 Description 字段
		Tags:        tags,
		Config:      config,
		Enabled:     enabled,
		Status:      target.Status,
		LastSeen:    target.LastSeen,
		CreatedAt:   target.CreatedAt,
		UpdatedAt:   target.UpdatedAt,
	}
}

// toDashboardResponse 转换为仪表板响应格式
func (s *MonitoringService) toDashboardResponse(dashboard *models.Dashboard) *DashboardResponse {
	var config map[string]interface{}
	var tags []string

	if dashboard.Config != "" {
		json.Unmarshal([]byte(dashboard.Config), &config)
	}
	if dashboard.Tags != "" {
		json.Unmarshal([]byte(dashboard.Tags), &tags)
	}

	return &DashboardResponse{
		ID:          dashboard.ID,
		Name:        dashboard.Name,
		Description: dashboard.Description,
		Config:      config,
		Tags:        tags,
		IsPublic:    dashboard.IsPublic,
		CreatedBy:   dashboard.CreatedBy,
		CreatedAt:   dashboard.CreatedAt,
		UpdatedAt:   dashboard.UpdatedAt,
	}
}

// GetMonitoringStats 获取监控统计信息
func (s *MonitoringService) GetMonitoringStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 监控目标总数
	var totalTargets int64
	if err := s.db.Model(&models.MonitoringTarget{}).Count(&totalTargets).Error; err != nil {
		return nil, fmt.Errorf("failed to count total targets: %w", err)
	}
	stats["total_targets"] = totalTargets

	// 活跃目标数
	var activeTargets int64
	if err := s.db.Model(&models.MonitoringTarget{}).Where("status = ?", "active").Count(&activeTargets).Error; err != nil {
		return nil, fmt.Errorf("failed to count active targets: %w", err)
	}
	stats["active_targets"] = activeTargets

	// 按类型统计
	typeStats := make(map[string]int64)
	types := []string{"host", "service", "application", "database"}
	for _, targetType := range types {
		var count int64
		if err := s.db.Model(&models.MonitoringTarget{}).Where("type = ?", targetType).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count %s targets: %w", targetType, err)
		}
		typeStats[targetType] = count
	}
	stats["type_stats"] = typeStats

	// 仪表板总数
	var totalDashboards int64
	if err := s.db.Model(&models.Dashboard{}).Count(&totalDashboards).Error; err != nil {
		return nil, fmt.Errorf("failed to count total dashboards: %w", err)
	}
	stats["total_dashboards"] = totalDashboards

	// 今日指标数据量
	var todayMetrics int64
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&models.MetricData{}).Where("timestamp >= ?", today).Count(&todayMetrics).Error; err != nil {
		return nil, fmt.Errorf("failed to count today metrics: %w", err)
	}
	stats["today_metrics"] = todayMetrics

	return stats, nil
}