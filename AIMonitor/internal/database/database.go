package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"ai-monitor/internal/config"
	"ai-monitor/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	sqlite "github.com/glebarez/sqlite"
)

// DB 全局数据库实例
var DB *gorm.DB

// ensureDBDir 确保数据库文件目录存在
func ensureDBDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." {
		return nil // 当前目录，无需创建
	}
	return os.MkdirAll(dir, 0755)
}

// Initialize 初始化数据库连接
func Initialize(cfg *config.DatabaseConfig) error {
	// 构建数据库连接字符串
	dsn := cfg.GetDSN()

	// 配置GORM日志
	logLevel := logger.Info
	if cfg.Host == "localhost" || cfg.Host == "127.0.0.1" {
		logLevel = logger.Info
	} else {
		logLevel = logger.Warn
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 根据驱动类型选择数据库连接器
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "mysql":
		// 使用MySQL驱动
		dialector = mysql.Open(dsn)
	case "sqlite":
		// 确保SQLite数据库文件目录存在
		if err := ensureDBDir(dsn); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
		// 使用标准的sqlite.Open方法
		dialector = sqlite.Open(dsn + "?_pragma=foreign_keys(1)")
	case "postgres", "":
		// 默认使用PostgreSQL
		dialector = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// 连接数据库
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层sql.DB对象并配置连接池（仅对非SQLite数据库）
	if cfg.Driver != "sqlite" {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}

		// 设置连接池参数
		sqlDB.SetMaxOpenConns(cfg.MaxConnections)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

		// 测试连接
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
	}

	DB = db
	log.Println("Database connected successfully")
	return nil
}

// Migrate 执行数据库迁移
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 启用UUID扩展（仅PostgreSQL）
	if DB.Dialector.Name() == "postgres" {
		if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
			log.Printf("Warning: failed to create uuid-ossp extension: %v", err)
		}
	}

	// 自动迁移所有模型
	err := DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.AlertRule{},
		&models.Alert{},
		&models.AlertNotification{},
		&models.SystemConfig{},
		&models.AuditLog{},
		&models.AIAnalysisResult{},
		&models.KnowledgeBase{},
		&models.MonitoringTarget{},
		&models.MetricData{},
		&models.Dashboard{},
		&models.NotificationChannel{},
		&models.APIKey{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 创建索引
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// 初始化基础数据
	if err := seedData(); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// createIndexes 创建数据库索引
func createIndexes() error {
	// 用户表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON users(last_login_at)")

	// 告警规则表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alert_rules_severity ON alert_rules(severity)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alert_rules_metric ON alert_rules(metric)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alert_rules_created_by ON alert_rules(created_by)")

	// 告警实例表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_rule_id ON alerts(rule_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_starts_at ON alerts(starts_at)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_ends_at ON alerts(ends_at)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_silenced ON alerts(silenced)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_acknowledged ON alerts(acknowledged)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_fingerprint ON alerts(fingerprint)")

	// 告警通知表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alert_notifications_alert_id ON alert_notifications(alert_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alert_notifications_status ON alert_notifications(status)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_alert_notifications_channel ON alert_notifications(channel)")

	// 系统配置表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_system_configs_category ON system_configs(category)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_system_configs_key ON system_configs(key)")

	// 审计日志表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_address ON audit_logs(ip_address)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)")

	// AI分析结果表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_ai_analysis_results_alert_id ON ai_analysis_results(alert_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_ai_analysis_results_analysis_type ON ai_analysis_results(analysis_type)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_ai_analysis_results_status ON ai_analysis_results(status)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_ai_analysis_results_model ON ai_analysis_results(model)")

	// 知识库表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_knowledge_base_category ON knowledge_base(category)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_knowledge_base_severity ON knowledge_base(severity)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_knowledge_base_platform ON knowledge_base(platform)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_knowledge_base_status ON knowledge_base(status)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_knowledge_base_created_by ON knowledge_base(created_by)")

	// 监控目标表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_monitoring_targets_type ON monitoring_targets(type)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_monitoring_targets_platform ON monitoring_targets(platform)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_monitoring_targets_status ON monitoring_targets(status)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_monitoring_targets_created_by ON monitoring_targets(created_by)")

	// 指标数据表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_metric_data_target_id ON metric_data(target_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_metric_data_metric ON metric_data(metric)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_metric_data_timestamp ON metric_data(timestamp)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_metric_data_target_metric_timestamp ON metric_data(target_id, metric, timestamp)")

	// 仪表板表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_dashboards_created_by ON dashboards(created_by)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_dashboards_is_public ON dashboards(is_public)")

	// 通知渠道表索引
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_notification_channels_type ON notification_channels(type)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_notification_channels_enabled ON notification_channels(enabled)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_notification_channels_created_by ON notification_channels(created_by)")

	return nil
}

// seedData 初始化基础数据
func seedData() error {
	// 检查是否已经初始化过
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return nil // 已经有数据，跳过初始化
	}

	// 创建默认权限
	permissions := []models.Permission{
		{Name: "user.create", Resource: "user", Action: "create", Description: "创建用户"},
		{Name: "user.read", Resource: "user", Action: "read", Description: "查看用户"},
		{Name: "user.update", Resource: "user", Action: "update", Description: "更新用户"},
		{Name: "user.delete", Resource: "user", Action: "delete", Description: "删除用户"},
		{Name: "role.create", Resource: "role", Action: "create", Description: "创建角色"},
		{Name: "role.read", Resource: "role", Action: "read", Description: "查看角色"},
		{Name: "role.update", Resource: "role", Action: "update", Description: "更新角色"},
		{Name: "role.delete", Resource: "role", Action: "delete", Description: "删除角色"},
		{Name: "alert.create", Resource: "alert", Action: "create", Description: "创建告警规则"},
		{Name: "alert.read", Resource: "alert", Action: "read", Description: "查看告警"},
		{Name: "alert.update", Resource: "alert", Action: "update", Description: "更新告警规则"},
		{Name: "alert.delete", Resource: "alert", Action: "delete", Description: "删除告警规则"},
		{Name: "alert.acknowledge", Resource: "alert", Action: "acknowledge", Description: "确认告警"},
		{Name: "alert.silence", Resource: "alert", Action: "silence", Description: "静默告警"},
		{Name: "monitoring.read", Resource: "monitoring", Action: "read", Description: "查看监控数据"},
		{Name: "monitoring.manage", Resource: "monitoring", Action: "manage", Description: "管理监控目标"},
		{Name: "dashboard.create", Resource: "dashboard", Action: "create", Description: "创建仪表板"},
		{Name: "dashboard.read", Resource: "dashboard", Action: "read", Description: "查看仪表板"},
		{Name: "dashboard.update", Resource: "dashboard", Action: "update", Description: "更新仪表板"},
		{Name: "dashboard.delete", Resource: "dashboard", Action: "delete", Description: "删除仪表板"},
		{Name: "system.config", Resource: "system", Action: "config", Description: "系统配置管理"},
		{Name: "system.audit", Resource: "system", Action: "audit", Description: "查看审计日志"},
		{Name: "ai.analysis", Resource: "ai", Action: "analysis", Description: "AI分析功能"},
		{Name: "knowledge.create", Resource: "knowledge", Action: "create", Description: "创建知识库"},
		{Name: "knowledge.read", Resource: "knowledge", Action: "read", Description: "查看知识库"},
		{Name: "knowledge.update", Resource: "knowledge", Action: "update", Description: "更新知识库"},
		{Name: "knowledge.delete", Resource: "knowledge", Action: "delete", Description: "删除知识库"},
	}

	for _, permission := range permissions {
		DB.FirstOrCreate(&permission, models.Permission{Name: permission.Name})
	}

	// 创建默认角色
	adminRole := models.Role{
		Name:        "admin",
		Description: "系统管理员",
		Status:      "active",
	}
	DB.FirstOrCreate(&adminRole, models.Role{Name: "admin"})

	operatorRole := models.Role{
		Name:        "operator",
		Description: "运维人员",
		Status:      "active",
	}
	DB.FirstOrCreate(&operatorRole, models.Role{Name: "operator"})

	viewerRole := models.Role{
		Name:        "viewer",
		Description: "只读用户",
		Status:      "active",
	}
	DB.FirstOrCreate(&viewerRole, models.Role{Name: "viewer"})

	// 为管理员角色分配所有权限
	var allPermissions []models.Permission
	DB.Find(&allPermissions)
	DB.Model(&adminRole).Association("Permissions").Replace(allPermissions)

	// 为运维人员分配部分权限
	var operatorPermissions []models.Permission
	operatorPermissionNames := []string{
		"alert.create", "alert.read", "alert.update", "alert.acknowledge", "alert.silence",
		"monitoring.read", "monitoring.manage",
		"dashboard.create", "dashboard.read", "dashboard.update",
		"ai.analysis",
		"knowledge.create", "knowledge.read", "knowledge.update",
	}
	DB.Where("name IN ?", operatorPermissionNames).Find(&operatorPermissions)
	DB.Model(&operatorRole).Association("Permissions").Replace(operatorPermissions)

	// 为只读用户分配查看权限
	var viewerPermissions []models.Permission
	viewerPermissionNames := []string{
		"alert.read", "monitoring.read", "dashboard.read", "knowledge.read",
	}
	DB.Where("name IN ?", viewerPermissionNames).Find(&viewerPermissions)
	DB.Model(&viewerRole).Association("Permissions").Replace(viewerPermissions)

	// 创建默认管理员用户
	adminUser := models.User{
		Username: "admin",
		Email:    "admin@aimonitor.com",
		Password: "$2a$10$mMqqvYIZnqDx7elkfuj/FeJMlGU.meZRV3WL/9UzePcHabCY.vYVS", // admin123
		FullName: "系统管理员",
		Status:   "active",
	}
	result := DB.FirstOrCreate(&adminUser, models.User{Username: "admin"})
	if result.Error == nil {
		// 为管理员用户分配管理员角色
		DB.Model(&adminUser).Association("Roles").Append(&adminRole)
	}

	// 创建默认系统配置
	defaultConfigs := []models.SystemConfig{
		{Key: "system.name", Value: "AI监控系统", Description: "系统名称", Category: "system", DataType: "string"},
		{Key: "system.version", Value: "1.0.0", Description: "系统版本", Category: "system", DataType: "string"},
		{Key: "alert.default_severity", Value: "medium", Description: "默认告警级别", Category: "alert", DataType: "string"},
		{Key: "alert.max_notifications_per_hour", Value: "10", Description: "每小时最大通知数量", Category: "alert", DataType: "int"},
		{Key: "ai.default_model", Value: "gpt-3.5-turbo", Description: "默认AI模型", Category: "ai", DataType: "string"},
		{Key: "ai.max_tokens", Value: "2000", Description: "AI最大Token数", Category: "ai", DataType: "int"},
		{Key: "monitoring.default_scrape_interval", Value: "30s", Description: "默认采集间隔", Category: "monitoring", DataType: "string"},
		{Key: "dashboard.default_refresh_interval", Value: "30s", Description: "仪表板默认刷新间隔", Category: "dashboard", DataType: "string"},
	}

	for _, config := range defaultConfigs {
		DB.FirstOrCreate(&config, models.SystemConfig{Key: config.Key})
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// Transaction 执行事务
func Transaction(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}