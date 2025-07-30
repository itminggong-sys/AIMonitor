package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-monitor/internal/config"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端实例
var RedisClient *redis.Client
var memoryCache *MemoryCache
var useMemoryFallback bool

// Initialize 初始化Redis连接
func Initialize(cfg *config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:               cfg.GetRedisAddr(),
		Password:           cfg.Password,
		DB:                 cfg.DB,
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		DialTimeout:        cfg.DialTimeout,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		PoolTimeout:        cfg.PoolTimeout,
		ConnMaxIdleTime:    cfg.IdleTimeout,
		ConnMaxLifetime:    cfg.IdleCheckFrequency,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Redis connection failed: %v, using memory cache fallback\n", err)
		memoryCache = NewMemoryCache()
		useMemoryFallback = true
		return nil // 不返回错误，使用内存缓存
	}

	fmt.Println("Redis connected successfully")
	useMemoryFallback = false
	return nil
}

// Close 关闭Redis连接
func Close() error {
	if useMemoryFallback && memoryCache != nil {
		memoryCache.Close()
	}
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// HealthCheck Redis健康检查
func HealthCheck() error {
	if RedisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	return nil
}

// CacheManager 缓存管理器
type CacheManager struct {
	client *redis.Client
}

// NewCacheManager 创建缓存管理器
func NewCacheManager() *CacheManager {
	return &CacheManager{
		client: RedisClient,
	}
}

// Set 设置缓存
func (c *CacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if useMemoryFallback {
		return memoryCache.Set(ctx, key, value, expiration)
	}
	
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (c *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	if useMemoryFallback {
		return memoryCache.Get(ctx, key, dest)
	}
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	return json.Unmarshal([]byte(data), dest)
}

// Delete 删除缓存
func (c *CacheManager) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Expire 设置过期时间
func (c *CacheManager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (c *CacheManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Increment 递增
func (c *CacheManager) Increment(ctx context.Context, key string) (int64, error) {
	if useMemoryFallback {
		return memoryCache.Increment(ctx, key)
	}
	return c.client.Incr(ctx, key).Result()
}

// IncrementBy 按指定值递增
func (c *CacheManager) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decrement 递减
func (c *CacheManager) Decrement(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// DecrementBy 按指定值递减
func (c *CacheManager) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// HSet 设置哈希字段
func (c *CacheManager) HSet(ctx context.Context, key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.client.HSet(ctx, key, field, data).Err()
}

// HGet 获取哈希字段
func (c *CacheManager) HGet(ctx context.Context, key string, field string, dest interface{}) error {
	data, err := c.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("field not found")
		}
		return fmt.Errorf("failed to get hash field: %w", err)
	}

	return json.Unmarshal([]byte(data), dest)
}

// HDel 删除哈希字段
func (c *CacheManager) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// HExists 检查哈希字段是否存在
func (c *CacheManager) HExists(ctx context.Context, key string, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (c *CacheManager) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// LPush 从左侧推入列表
func (c *CacheManager) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

// RPush 从右侧推入列表
func (c *CacheManager) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.RPush(ctx, key, values...).Err()
}

// LPop 从左侧弹出列表元素
func (c *CacheManager) LPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, key).Result()
}

// RPop 从右侧弹出列表元素
func (c *CacheManager) RPop(ctx context.Context, key string) (string, error) {
	return c.client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (c *CacheManager) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// LRange 获取列表范围内的元素
func (c *CacheManager) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// SAdd 添加集合成员
func (c *CacheManager) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// SRem 移除集合成员
func (c *CacheManager) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// SIsMember 检查是否为集合成员
func (c *CacheManager) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// SMembers 获取所有集合成员
func (c *CacheManager) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SCard 获取集合成员数量
func (c *CacheManager) SCard(ctx context.Context, key string) (int64, error) {
	return c.client.SCard(ctx, key).Result()
}

// ZAdd 添加有序集合成员
func (c *CacheManager) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	return c.client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// ZRem 移除有序集合成员
func (c *CacheManager) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.ZRem(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围内的成员
func (c *CacheManager) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 获取有序集合范围内的成员及分数
func (c *CacheManager) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return c.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZCard 获取有序集合成员数量
func (c *CacheManager) ZCard(ctx context.Context, key string) (int64, error) {
	return c.client.ZCard(ctx, key).Result()
}

// ZScore 获取有序集合成员分数
func (c *CacheManager) ZScore(ctx context.Context, key string, member string) (float64, error) {
	return c.client.ZScore(ctx, key, member).Result()
}

// Publish 发布消息
func (c *CacheManager) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return c.client.Publish(ctx, channel, data).Err()
}

// Subscribe 订阅频道
func (c *CacheManager) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.client.Subscribe(ctx, channels...)
}

// PSubscribe 模式订阅
func (c *CacheManager) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return c.client.PSubscribe(ctx, patterns...)
}

// Keys 获取匹配模式的键
func (c *CacheManager) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// FlushDB 清空当前数据库
func (c *CacheManager) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// FlushAll 清空所有数据库
func (c *CacheManager) FlushAll(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

// Info 获取Redis信息
func (c *CacheManager) Info(ctx context.Context, section ...string) (string, error) {
	return c.client.Info(ctx, section...).Result()
}

// DBSize 获取数据库键数量
func (c *CacheManager) DBSize(ctx context.Context) (int64, error) {
	return c.client.DBSize(ctx).Result()
}

// Pipeline 创建管道
func (c *CacheManager) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// TxPipeline 创建事务管道
func (c *CacheManager) TxPipeline() redis.Pipeliner {
	return c.client.TxPipeline()
}

// Watch 监视键
func (c *CacheManager) Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	return c.client.Watch(ctx, fn, keys...)
}

// Eval 执行Lua脚本
func (c *CacheManager) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}

// EvalSha 执行Lua脚本SHA
func (c *CacheManager) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.EvalSha(ctx, sha1, keys, args...).Result()
}

// ScriptLoad 加载Lua脚本
func (c *CacheManager) ScriptLoad(ctx context.Context, script string) (string, error) {
	return c.client.ScriptLoad(ctx, script).Result()
}

// GetClient 获取Redis客户端
func (c *CacheManager) GetClient() *redis.Client {
	return c.client
}

// 缓存键前缀常量
const (
	UserCachePrefix         = "user:"
	SessionCachePrefix      = "session:"
	AlertCachePrefix        = "alert:"
	MetricCachePrefix       = "metric:"
	DashboardCachePrefix    = "dashboard:"
	ConfigCachePrefix       = "config:"
	AIAnalysisCachePrefix   = "ai_analysis:"
	KnowledgeCachePrefix    = "knowledge:"
	NotificationCachePrefix = "notification:"
	RateLimitCachePrefix    = "rate_limit:"
	LockCachePrefix         = "lock:"
)

// 生成缓存键的辅助函数
func UserCacheKey(userID string) string {
	return UserCachePrefix + userID
}

func SessionCacheKey(sessionID string) string {
	return SessionCachePrefix + sessionID
}

func AlertCacheKey(alertID string) string {
	return AlertCachePrefix + alertID
}

func MetricCacheKey(targetID, metric string) string {
	return MetricCachePrefix + targetID + ":" + metric
}

func DashboardCacheKey(dashboardID string) string {
	return DashboardCachePrefix + dashboardID
}

func ConfigCacheKey(key string) string {
	return ConfigCachePrefix + key
}

func AIAnalysisCacheKey(alertID string) string {
	return AIAnalysisCachePrefix + alertID
}

func KnowledgeCacheKey(category string) string {
	return KnowledgeCachePrefix + category
}

func NotificationCacheKey(alertID string) string {
	return NotificationCachePrefix + alertID
}

func RateLimitCacheKey(userID, action string) string {
	return RateLimitCachePrefix + userID + ":" + action
}

func LockCacheKey(resource string) string {
	return LockCachePrefix + resource
}

// AuditStatsCacheKey 生成审计统计缓存键
func AuditStatsCacheKey(days int) string {
	return fmt.Sprintf("%saudit_stats_%d", ConfigCachePrefix, days)
}

// ConfigCategoryCacheKey 生成配置分类缓存键
func ConfigCategoryCacheKey(category string) string {
	return fmt.Sprintf("config:category:%s", category)
}

// SystemMetricsCacheKey 生成系统指标缓存键
func SystemMetricsCacheKey(targetID string) string {
	return fmt.Sprintf("system:metrics:%s", targetID)
}

func AlertRulesCacheKey(targetType, targetID, metricName string) string {
	return "alert_rules:" + targetType + ":" + targetID + ":" + metricName
}

// JWTBlacklistKey 生成JWT黑名单缓存键
func JWTBlacklistKey(tokenID string) string {
	return "jwt_blacklist:" + tokenID
}