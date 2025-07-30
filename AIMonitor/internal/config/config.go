package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 系统配置结构
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Prometheus    PrometheusConfig    `mapstructure:"prometheus"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	AIModels      AIModelsConfig      `mapstructure:"ai_models"`
	Email         EmailConfig         `mapstructure:"email"`
	Alerting      AlertingConfig      `mapstructure:"alerting"`
	Collector     CollectorConfig     `mapstructure:"collector"`
	Logging       LoggingConfig       `mapstructure:"logging"`
	Security      SecurityConfig      `mapstructure:"security"`
	Monitoring    MonitoringConfig    `mapstructure:"monitoring"`
	Cache         CacheConfig         `mapstructure:"cache"`
	FeatureFlags  FeatureFlagsConfig  `mapstructure:"feature_flags"`
	Development   DevelopmentConfig   `mapstructure:"development"`
	Production    ProductionConfig    `mapstructure:"production"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	Mode           string        `mapstructure:"mode"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`          // 数据库驱动: mysql, postgres, sqlite
	DSN             string        `mapstructure:"dsn"`             // 数据库连接字符串（SQLite使用）
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxConnections  int           `mapstructure:"max_connections"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	Timeout         time.Duration `mapstructure:"timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host               string        `mapstructure:"host"`
	Port               int           `mapstructure:"port"`
	Password           string        `mapstructure:"password"`
	DB                 int           `mapstructure:"db"`
	PoolSize           int           `mapstructure:"pool_size"`
	MinIdleConns       int           `mapstructure:"min_idle_conns"`
	DialTimeout        time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout"`
	PoolTimeout        time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
	IdleCheckFrequency time.Duration `mapstructure:"idle_check_frequency"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey           string        `mapstructure:"secret_key"`
	AccessTokenExpiry   time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry  time.Duration `mapstructure:"refresh_token_expiry"`
	Issuer              string        `mapstructure:"issuer"`
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	URL          string        `mapstructure:"url"`
	Timeout      time.Duration `mapstructure:"timeout"`
	MaxSamples   int           `mapstructure:"max_samples"`
	QueryTimeout time.Duration `mapstructure:"query_timeout"`
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Addresses     []string      `mapstructure:"addresses"`
	Username      string        `mapstructure:"username"`
	Password      string        `mapstructure:"password"`
	CloudID       string        `mapstructure:"cloud_id"`
	APIKey        string        `mapstructure:"api_key"`
	Timeout       time.Duration `mapstructure:"timeout"`
	MaxRetries    int           `mapstructure:"max_retries"`
	RetryBackoff  time.Duration `mapstructure:"retry_backoff"`
}

// AIModelsConfig AI模型配置
type AIModelsConfig struct {
	OpenAI AIModelConfig `mapstructure:"openai"`
	Claude AIModelConfig `mapstructure:"claude"`
}

// AIModelConfig 单个AI模型配置
type AIModelConfig struct {
	APIKey      string          `mapstructure:"api_key"`
	BaseURL     string          `mapstructure:"base_url"`
	Model       string          `mapstructure:"model"`
	Temperature float64         `mapstructure:"temperature"`
	MaxTokens   int             `mapstructure:"max_tokens"`
	Timeout     time.Duration   `mapstructure:"timeout"`
	RateLimit   RateLimitConfig `mapstructure:"rate_limit"`
	CostLimit   CostLimitConfig `mapstructure:"cost_limit"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	TokensPerMinute   int `mapstructure:"tokens_per_minute"`
}

// CostLimitConfig 成本限制配置
type CostLimitConfig struct {
	DailyLimit     float64 `mapstructure:"daily_limit"`
	MonthlyLimit   float64 `mapstructure:"monthly_limit"`
	AlertThreshold float64 `mapstructure:"alert_threshold"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTP      SMTPConfig      `mapstructure:"smtp"`
	Templates TemplateConfig  `mapstructure:"templates"`
}

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	UseTLS   bool   `mapstructure:"use_tls"`
}

// TemplateConfig 模板配置
type TemplateConfig struct {
	AlertSubject string `mapstructure:"alert_subject"`
	AlertBody    string `mapstructure:"alert_body"`
}

// AlertingConfig 告警配置
type AlertingConfig struct {
	EvaluationInterval   time.Duration       `mapstructure:"evaluation_interval"`
	NotificationTimeout  time.Duration       `mapstructure:"notification_timeout"`
	GroupWait            time.Duration       `mapstructure:"group_wait"`
	GroupInterval        time.Duration       `mapstructure:"group_interval"`
	RepeatInterval       time.Duration       `mapstructure:"repeat_interval"`
	MaxAlertsPerGroup    int                 `mapstructure:"max_alerts_per_group"`
	Aggregation          AggregationConfig   `mapstructure:"aggregation"`
}

// AggregationConfig 聚合配置
type AggregationConfig struct {
	Enabled    bool          `mapstructure:"enabled"`
	TimeWindow time.Duration `mapstructure:"time_window"`
	MaxCount   int           `mapstructure:"max_count"`
	GroupBy    []string      `mapstructure:"group_by"`
}

// CollectorConfig 采集器配置
type CollectorConfig struct {
	ScrapeInterval     time.Duration              `mapstructure:"scrape_interval"`
	ScrapeTimeout      time.Duration              `mapstructure:"scrape_timeout"`
	EvaluationInterval time.Duration              `mapstructure:"evaluation_interval"`
	Platforms          map[string]PlatformConfig  `mapstructure:"platforms"`
}

// PlatformConfig 平台配置
type PlatformConfig struct {
	Enabled   bool     `mapstructure:"enabled"`
	Exporters []string `mapstructure:"exporters"`
	Metrics   []string `mapstructure:"metrics"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	CORS       CORSConfig       `mapstructure:"cors"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string      `mapstructure:"allowed_origins"`
	AllowedMethods   []string      `mapstructure:"allowed_methods"`
	AllowedHeaders   []string      `mapstructure:"allowed_headers"`
	ExposedHeaders   []string      `mapstructure:"exposed_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Key       string `mapstructure:"key"`
	Algorithm string `mapstructure:"algorithm"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
	Metrics     MetricsConfig     `mapstructure:"metrics"`
	Tracing     TracingConfig     `mapstructure:"tracing"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	Interval  time.Duration `mapstructure:"interval"`
	Timeout   time.Duration `mapstructure:"timeout"`
	Endpoints []string      `mapstructure:"endpoints"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Path      string `mapstructure:"path"`
	Namespace string `mapstructure:"namespace"`
}

// TracingConfig 链路追踪配置
type TracingConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	JaegerEndpoint  string `mapstructure:"jaeger_endpoint"`
	ServiceName     string `mapstructure:"service_name"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	DefaultTTL      time.Duration              `mapstructure:"default_ttl"`
	CleanupInterval time.Duration              `mapstructure:"cleanup_interval"`
	MaxSize         int                        `mapstructure:"max_size"`
	Configs         map[string]CacheItemConfig `mapstructure:"configs"`
}

// CacheItemConfig 缓存项配置
type CacheItemConfig struct {
	TTL     time.Duration `mapstructure:"ttl"`
	MaxSize int           `mapstructure:"max_size"`
}

// FeatureFlagsConfig 特性开关配置
type FeatureFlagsConfig struct {
	AIAnalysis        bool `mapstructure:"ai_analysis"`
	FailurePrediction bool `mapstructure:"failure_prediction"`
	AutoRemediation   bool `mapstructure:"auto_remediation"`
	AdvancedAnalytics bool `mapstructure:"advanced_analytics"`
	RealTimeStreaming bool `mapstructure:"real_time_streaming"`
	MultiTenant       bool `mapstructure:"multi_tenant"`
}

// DevelopmentConfig 开发环境配置
type DevelopmentConfig struct {
	Debug     bool `mapstructure:"debug"`
	HotReload bool `mapstructure:"hot_reload"`
	MockAI    bool `mapstructure:"mock_ai"`
	SeedData  bool `mapstructure:"seed_data"`
	Profiling bool `mapstructure:"profiling"`
}

// ProductionConfig 生产环境配置
type ProductionConfig struct {
	Debug     bool `mapstructure:"debug"`
	HotReload bool `mapstructure:"hot_reload"`
	MockAI    bool `mapstructure:"mock_ai"`
	SeedData  bool `mapstructure:"seed_data"`
	Profiling bool `mapstructure:"profiling"`
}

// Load 加载配置
func Load() (*Config, error) {
	return LoadWithFile("config")
}

// LoadWithFile 加载指定的配置文件
func LoadWithFile(configName string) (*Config, error) {
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath("/etc/ai-monitor")

	// 设置环境变量前缀
	viper.SetEnvPrefix("AI_MONITOR")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置默认值
func setDefaults() {
	// 服务器默认值
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// 数据库默认值
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_connections", 100)
	viper.SetDefault("database.max_idle_conns", 10)

	// Redis默认值
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	// JWT默认值
	viper.SetDefault("jwt.access_token_expiry", "24h")
	viper.SetDefault("jwt.refresh_token_expiry", "168h")
	viper.SetDefault("jwt.issuer", "ai-monitor")

	// 日志默认值
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// SQLite数据库不需要host验证
	if cfg.Database.Driver != "sqlite" && cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required")
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	// 如果直接指定了DSN，则使用DSN
	if c.DSN != "" {
		return c.DSN
	}
	
	// 根据驱动类型构建DSN
	switch c.Driver {
	case "mysql":
		// MySQL DSN格式
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.User, c.Password, c.Host, c.Port, c.Name)
	case "sqlite":
		return c.Name // SQLite使用文件路径作为DSN
	case "postgres", "":
		// 默认使用PostgreSQL格式
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
		)
	default:
		// 未知驱动，返回PostgreSQL格式
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
		)
	}
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}