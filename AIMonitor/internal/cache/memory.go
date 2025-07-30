package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MemoryCache 内存缓存实现
type MemoryCache struct {
	data   map[string]*cacheItem
	mutex  sync.RWMutex
	cleanup *time.Ticker
	stop    chan bool
}

type cacheItem struct {
	value      []byte
	expiration time.Time
	hasExpiry  bool
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		data:    make(map[string]*cacheItem),
		cleanup: time.NewTicker(5 * time.Minute),
		stop:    make(chan bool),
	}
	
	// 启动清理协程
	go mc.cleanupExpired()
	
	return mc
}

// cleanupExpired 清理过期项
func (mc *MemoryCache) cleanupExpired() {
	for {
		select {
		case <-mc.cleanup.C:
			mc.mutex.Lock()
			now := time.Now()
			for key, item := range mc.data {
				if item.hasExpiry && now.After(item.expiration) {
					delete(mc.data, key)
				}
			}
			mc.mutex.Unlock()
		case <-mc.stop:
			return
		}
	}
}

// Close 关闭内存缓存
func (mc *MemoryCache) Close() {
	mc.cleanup.Stop()
	close(mc.stop)
}

// Set 设置缓存
func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	item := &cacheItem{
		value: data,
	}
	
	if expiration > 0 {
		item.expiration = time.Now().Add(expiration)
		item.hasExpiry = true
	}
	
	mc.data[key] = item
	return nil
}

// Get 获取缓存
func (mc *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	mc.mutex.RLock()
	item, exists := mc.data[key]
	mc.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("key not found")
	}
	
	// 检查是否过期
	if item.hasExpiry && time.Now().After(item.expiration) {
		mc.mutex.Lock()
		delete(mc.data, key)
		mc.mutex.Unlock()
		return fmt.Errorf("key not found")
	}
	
	return json.Unmarshal(item.value, dest)
}

// Delete 删除缓存
func (mc *MemoryCache) Delete(ctx context.Context, keys ...string) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	for _, key := range keys {
		delete(mc.data, key)
	}
	return nil
}

// Exists 检查键是否存在
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mutex.RLock()
	item, exists := mc.data[key]
	mc.mutex.RUnlock()
	
	if !exists {
		return false, nil
	}
	
	// 检查是否过期
	if item.hasExpiry && time.Now().After(item.expiration) {
		mc.mutex.Lock()
		delete(mc.data, key)
		mc.mutex.Unlock()
		return false, nil
	}
	
	return true, nil
}

// Expire 设置过期时间
func (mc *MemoryCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	item, exists := mc.data[key]
	if !exists {
		return fmt.Errorf("key not found")
	}
	
	item.expiration = time.Now().Add(expiration)
	item.hasExpiry = true
	return nil
}

// TTL 获取剩余过期时间
func (mc *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	mc.mutex.RLock()
	item, exists := mc.data[key]
	mc.mutex.RUnlock()
	
	if !exists {
		return -2 * time.Second, nil // Redis convention: -2 for non-existent key
	}
	
	if !item.hasExpiry {
		return -1 * time.Second, nil // Redis convention: -1 for no expiry
	}
	
	remaining := time.Until(item.expiration)
	if remaining <= 0 {
		mc.mutex.Lock()
		delete(mc.data, key)
		mc.mutex.Unlock()
		return -2 * time.Second, nil
	}
	
	return remaining, nil
}

// Increment 递增
func (mc *MemoryCache) Increment(ctx context.Context, key string) (int64, error) {
	return mc.IncrementBy(ctx, key, 1)
}

// IncrementBy 按指定值递增
func (mc *MemoryCache) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	item, exists := mc.data[key]
	var currentValue int64 = 0
	
	if exists {
		// 检查是否过期
		if item.hasExpiry && time.Now().After(item.expiration) {
			delete(mc.data, key)
		} else {
			// 尝试解析当前值
			if err := json.Unmarshal(item.value, &currentValue); err != nil {
				currentValue = 0
			}
		}
	}
	
	newValue := currentValue + value
	data, _ := json.Marshal(newValue)
	
	newItem := &cacheItem{
		value: data,
	}
	
	if exists && item.hasExpiry {
		newItem.expiration = item.expiration
		newItem.hasExpiry = true
	}
	
	mc.data[key] = newItem
	return newValue, nil
}

// Decrement 递减
func (mc *MemoryCache) Decrement(ctx context.Context, key string) (int64, error) {
	return mc.DecrementBy(ctx, key, 1)
}

// DecrementBy 按指定值递减
func (mc *MemoryCache) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return mc.IncrementBy(ctx, key, -value)
}