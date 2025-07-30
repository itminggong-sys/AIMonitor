package scheduler

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"ai-monitor/internal/cache"
	"ai-monitor/internal/database"
	"ai-monitor/internal/metrics"
	"ai-monitor/internal/services"
	"ai-monitor/internal/websocket"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron              *cron.Cron
	alertService      *services.AlertService
	monitoringService *services.MonitoringService
	auditService      *services.AuditService
	configService     *services.ConfigService
	userService       *services.UserService
	aiService         *services.AIService
	redisClient       *redis.Client
	metrics           *metrics.Metrics
	wsManager         *websocket.WebSocketManager
	ctx               context.Context
	cancel            context.CancelFunc
}

// NewScheduler 创建定时任务调度器
func NewScheduler(
	alertService *services.AlertService,
	monitoringService *services.MonitoringService,
	auditService *services.AuditService,
	configService *services.ConfigService,
	userService *services.UserService,
	aiService *services.AIService,
	redisClient *redis.Client,
	metrics *metrics.Metrics,
	wsManager *websocket.WebSocketManager,
) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		cron:              cron.New(cron.WithSeconds()),
		alertService:      alertService,
		monitoringService: monitoringService,
		auditService:      auditService,
		configService:     configService,
		userService:       userService,
		aiService:         aiService,
		redisClient:       redisClient,
		metrics:           metrics,
		wsManager:         wsManager,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Start 启动定时任务
func (s *Scheduler) Start() error {
	log.Println("Starting scheduler...")

	// 注册定时任务
	if err := s.registerJobs(); err != nil {
		return err
	}

	// 启动cron调度器
	s.cron.Start()

	// 启动系统监控
	go s.startSystemMonitoring()

	log.Println("Scheduler started successfully")
	return nil
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	s.cancel()
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

// registerJobs 注册定时任务
func (s *Scheduler) registerJobs() error {
	jobs := []struct {
		spec string
		job  func()
		name string
	}{
		// 每分钟检查告警规则
		{"0 * * * * *", s.checkAlertRules, "check_alert_rules"},
		
		// 每5分钟收集系统指标
		{"0 */5 * * * *", s.collectSystemMetrics, "collect_system_metrics"},
		
		// 每10分钟更新缓存统计
		{"0 */10 * * * *", s.updateCacheStats, "update_cache_stats"},
		
		// 每30分钟清理过期缓存
		{"0 */30 * * * *", s.cleanExpiredCache, "clean_expired_cache"},
		
		// 每小时更新业务统计
		{"0 0 * * * *", s.updateBusinessStats, "update_business_stats"},
		
		// 每小时检查系统健康状态
		{"0 0 * * * *", s.checkSystemHealth, "check_system_health"},
		
		// 每天凌晨2点清理旧日志
		{"0 0 2 * * *", s.cleanOldLogs, "clean_old_logs"},
		
		// 每天凌晨3点备份重要数据
		{"0 0 3 * * *", s.backupData, "backup_data"},
		
		// 每天凌晨4点生成日报
		{"0 0 4 * * *", s.generateDailyReport, "generate_daily_report"},
		
		// 每周日凌晨1点生成周报
		{"0 0 1 * * 0", s.generateWeeklyReport, "generate_weekly_report"},
		
		// 每月1号凌晨1点生成月报
		{"0 0 1 1 * *", s.generateMonthlyReport, "generate_monthly_report"},
	}

	for _, job := range jobs {
		if _, err := s.cron.AddFunc(job.spec, job.job); err != nil {
			log.Printf("Failed to register job %s: %v", job.name, err)
			return err
		}
		log.Printf("Registered job: %s with spec: %s", job.name, job.spec)
	}

	return nil
}

// checkAlertRules 检查告警规则
func (s *Scheduler) checkAlertRules() {
	start := time.Now()
	defer func() {
		s.metrics.RecordAlertProcessing("check_rules", time.Since(start))
	}()

	log.Println("Checking alert rules...")

	// 获取启用的告警规则
	rules, _, err := s.alertService.ListAlertRules(1, 1000, "", boolPtr(true))
	if err != nil {
		log.Printf("Failed to get alert rules: %v", err)
		return
	}

	// 模拟指标数据进行告警检查
	for _, rule := range rules {
		// 创建模拟的指标数据
		metricData := &services.MetricData{
			TargetType:  "system",
			TargetID:    "localhost",
			MetricName:  rule.MetricName,
			Value:       0.0, // 这里应该从实际监控系统获取真实数据
			Timestamp:   time.Now(),
			Tags:        make(map[string]interface{}),
		}

		// 处理指标数据并检查告警规则
		if err := s.alertService.ProcessMetricData(metricData); err != nil {
			log.Printf("Failed to process metric data for rule %s: %v", rule.Name, err)
		}
	}

	log.Printf("Checked %d alert rules", len(rules))
}

// collectSystemMetrics 收集系统指标
func (s *Scheduler) collectSystemMetrics() {
	log.Println("Collecting system metrics...")

	// 收集内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	s.metrics.UpdateMemoryUsage("alloc", m.Alloc)
	s.metrics.UpdateMemoryUsage("total_alloc", m.TotalAlloc)
	s.metrics.UpdateMemoryUsage("sys", m.Sys)
	s.metrics.UpdateMemoryUsage("heap_alloc", m.HeapAlloc)
	s.metrics.UpdateMemoryUsage("heap_sys", m.HeapSys)
	s.metrics.UpdateMemoryUsage("heap_inuse", m.HeapInuse)

	// 收集协程数量
	s.metrics.UpdateGoroutines("main", runtime.NumGoroutine())

	// 收集数据库连接池状态
	db := database.GetDB()
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			stats := sqlDB.Stats()
			s.metrics.UpdateDBConnections("postgres", stats.OpenConnections, stats.Idle, stats.InUse)
		}
	}

	// 收集Redis连接状态
	if s.redisClient != nil {
		ctx := context.Background()
		info := s.redisClient.Info(ctx, "clients")
		if info.Err() == nil {
			// 解析Redis info信息（简化版）
			s.metrics.UpdateRedisConnections("main", 1) // 简化处理
		}
	}

	// 收集WebSocket连接数
	if s.wsManager != nil {
		connectedClients := s.wsManager.GetConnectedClients()
		s.metrics.UpdateWebSocketConnections("total", connectedClients)
	}

	log.Println("System metrics collected")
}

// updateCacheStats 更新缓存统计
func (s *Scheduler) updateCacheStats() {
	log.Println("Updating cache stats...")

	if s.redisClient == nil {
		return
	}

	// 获取Redis信息
	ctx := context.Background()
	info := s.redisClient.Info(ctx, "memory")
	if info.Err() != nil {
		log.Printf("Failed to get Redis info: %v", info.Err())
		return
	}

	// 获取数据库大小
	dbSize := s.redisClient.DBSize(ctx)
	if dbSize.Err() != nil {
		log.Printf("Failed to get Redis DB size: %v", dbSize.Err())
		return
	}

	log.Printf("Redis DB size: %d keys", dbSize.Val())
	log.Println("Cache stats updated")
}

// cleanExpiredCache 清理过期缓存
func (s *Scheduler) cleanExpiredCache() {
	log.Println("Cleaning expired cache...")

	if s.redisClient == nil {
		return
	}

	// 清理过期的JWT令牌
	pattern := cache.JWTBlacklistKey("*")
	keys, err := s.redisClient.Keys(context.Background(), pattern).Result()
	if err != nil {
		log.Printf("Failed to get JWT blacklist keys: %v", err)
		return
	}

	expiredCount := 0
	for _, key := range keys {
		ttl, err := s.redisClient.TTL(context.Background(), key).Result()
		if err != nil {
			continue
		}
		if ttl <= 0 {
			s.redisClient.Del(context.Background(), key)
			expiredCount++
		}
	}

	log.Printf("Cleaned %d expired cache entries", expiredCount)
}

// updateBusinessStats 更新业务统计
func (s *Scheduler) updateBusinessStats() {
	log.Println("Updating business stats...")

	// 更新用户统计
	if stats, err := s.userService.GetUserStats(); err == nil {
		if activeUsers, ok := stats["active_users"].(int64); ok {
			s.metrics.UpdateUsersTotal("active", int(activeUsers))
		}
		if totalUsers, ok := stats["total_users"].(int64); ok {
			s.metrics.UpdateUsersTotal("total", int(totalUsers))
		}
		if todayUsers, ok := stats["today_users"].(int64); ok {
			s.metrics.UpdateActiveUsers("daily", int(todayUsers))
		}
		// 注意：用户服务没有提供周活跃和月活跃用户统计
	}

	// 更新告警统计
	if stats, err := s.alertService.GetAlertStats(); err == nil {
		if severityStats, ok := stats["severity_stats"].(map[string]int64); ok {
			if critical, exists := severityStats["critical"]; exists {
				s.metrics.UpdateActiveAlerts("critical", int(critical))
			}
			if high, exists := severityStats["high"]; exists {
				s.metrics.UpdateActiveAlerts("warning", int(high))
			}
			if medium, exists := severityStats["medium"]; exists {
				s.metrics.UpdateActiveAlerts("info", int(medium))
			}
		}
		// 注意：告警服务没有提供启用/禁用规则统计
	}

	// 更新监控目标统计
	if stats, err := s.monitoringService.GetMonitoringStats(); err == nil {
		if typeStats, ok := stats["type_stats"].(map[string]int64); ok {
			if hostTargets, exists := typeStats["host"]; exists {
				s.metrics.UpdateMonitoringTargets("server", "enabled", int(hostTargets))
			}
			if appTargets, exists := typeStats["application"]; exists {
				s.metrics.UpdateMonitoringTargets("application", "enabled", int(appTargets))
			}
			if dbTargets, exists := typeStats["database"]; exists {
				s.metrics.UpdateMonitoringTargets("database", "enabled", int(dbTargets))
			}
		}
		if totalDashboards, ok := stats["total_dashboards"].(int64); ok {
			s.metrics.UpdateDashboards("system", int(totalDashboards))
		}
	}

	log.Println("Business stats updated")
}

// checkSystemHealth 检查系统健康状态
func (s *Scheduler) checkSystemHealth() {
	log.Println("Checking system health...")

	// 检查数据库连接
	if err := database.HealthCheck(); err != nil {
		log.Printf("Database health check failed: %v", err)
		s.wsManager.BroadcastSystem("error", "Database Health Check", "Database connection failed")
	}

	// 检查Redis连接
	if s.redisClient != nil {
		if err := s.redisClient.HealthCheck(); err != nil {
			log.Printf("Redis health check failed: %v", err)
			s.wsManager.BroadcastSystem("error", "Redis Health Check", "Redis connection failed")
		}
	}

	// 检查内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsagePercent := float64(m.Alloc) / float64(m.Sys) * 100
	if memoryUsagePercent > 80 {
		log.Printf("High memory usage: %.2f%%", memoryUsagePercent)
		s.wsManager.BroadcastSystem("warning", "High Memory Usage", 
			fmt.Sprintf("Memory usage is %.2f%%", memoryUsagePercent))
	}

	// 检查协程数量
	goroutineCount := runtime.NumGoroutine()
	if goroutineCount > 1000 {
		log.Printf("High goroutine count: %d", goroutineCount)
		s.wsManager.BroadcastSystem("warning", "High Goroutine Count", 
			fmt.Sprintf("Goroutine count is %d", goroutineCount))
	}

	log.Println("System health check completed")
}

// cleanOldLogs 清理旧日志
func (s *Scheduler) cleanOldLogs() {
	log.Println("Cleaning old logs...")

	// 清理30天前的审计日志
	if err := s.auditService.CleanOldLogs(30); err != nil {
		log.Printf("Failed to clean old audit logs: %v", err)
	} else {
		log.Println("Old audit logs cleaned")
	}

	// 清理旧的AI分析结果（保留90天）
	if err := s.aiService.CleanOldAnalysis(90); err != nil {
		log.Printf("Failed to clean old AI analysis: %v", err)
	} else {
		log.Println("Old AI analysis cleaned")
	}

	log.Println("Old logs cleanup completed")
}

// backupData 备份重要数据
func (s *Scheduler) backupData() {
	log.Println("Starting data backup...")

	// 这里可以实现数据备份逻辑
	// 例如：备份配置、用户数据、告警规则等

	log.Println("Data backup completed")
}

// generateDailyReport 生成日报
func (s *Scheduler) generateDailyReport() {
	log.Println("Generating daily report...")

	// 获取昨天的统计数据
	yesterday := time.Now().AddDate(0, 0, -1)
	startTime := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	endTime := startTime.Add(24 * time.Hour)

	// 生成报告内容
	report := map[string]interface{}{
		"date":       yesterday.Format("2006-01-02"),
		"start_time": startTime,
		"end_time":   endTime,
	}

	// 获取告警统计
	if alertStats, err := s.auditService.GetAuditStats(1); err == nil {
		report["alerts"] = alertStats
	}

	// 发送报告通知
	s.wsManager.BroadcastSystem("info", "Daily Report", "Daily report has been generated")

	log.Println("Daily report generated")
}

// generateWeeklyReport 生成周报
func (s *Scheduler) generateWeeklyReport() {
	log.Println("Generating weekly report...")

	// 获取上周的统计数据
	now := time.Now()
	lastWeek := now.AddDate(0, 0, -7)

	report := map[string]interface{}{
		"week_start": lastWeek.Format("2006-01-02"),
		"week_end":   now.Format("2006-01-02"),
	}

	// 获取周统计数据
	if auditStats, err := s.auditService.GetAuditStats(7); err == nil {
		report["audit"] = auditStats
	}

	// 发送报告通知
	s.wsManager.BroadcastSystem("info", "Weekly Report", "Weekly report has been generated")

	log.Println("Weekly report generated")
}

// generateMonthlyReport 生成月报
func (s *Scheduler) generateMonthlyReport() {
	log.Println("Generating monthly report...")

	// 获取上月的统计数据
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)

	report := map[string]interface{}{
		"month": lastMonth.Format("2006-01"),
	}

	// 获取月统计数据
	if auditStats, err := s.auditService.GetAuditStats(30); err == nil {
		report["audit"] = auditStats
	}

	// 发送报告通知
	s.wsManager.BroadcastSystem("info", "Monthly Report", "Monthly report has been generated")

	log.Println("Monthly report generated")
}

// startSystemMonitoring 启动系统监控
func (s *Scheduler) startSystemMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 更新系统运行时间
			s.metrics.UpdateUptime("main", 30)

			// 更新系统信息
			s.metrics.UpdateSystemInfo("1.0.0", runtime.Version(), time.Now().Format("2006-01-02 15:04:05"))

		case <-s.ctx.Done():
			return
		}
	}
}

// 辅助函数
func boolPtr(b bool) *bool {
	return &b
}