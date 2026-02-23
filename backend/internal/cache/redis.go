package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache provides Redis caching functionality
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// CacheConfig holds Redis configuration
type CacheConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config CacheConfig) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()

	// Test connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis cache connection established")

	return &RedisCache{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// Set stores a value in cache with expiration
func (rc *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return rc.client.Set(rc.ctx, key, jsonValue, expiration).Err()
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(key string, dest interface{}) error {
	val, err := rc.client.Get(rc.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get value from cache: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key from cache
func (rc *RedisCache) Delete(key string) error {
	return rc.client.Del(rc.ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (rc *RedisCache) DeletePattern(pattern string) error {
	keys, err := rc.client.Keys(rc.ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return rc.client.Del(rc.ctx, keys...).Err()
	}

	return nil
}

// Exists checks if a key exists in cache
func (rc *RedisCache) Exists(key string) (bool, error) {
	count, err := rc.client.Exists(rc.ctx, key).Result()
	return count > 0, err
}

// SetHash stores a hash in cache
func (rc *RedisCache) SetHash(key string, fields map[string]interface{}, expiration time.Duration) error {
	pipe := rc.client.Pipeline()
	pipe.HMSet(rc.ctx, key, fields)
	pipe.Expire(rc.ctx, key, expiration)
	_, err := pipe.Exec(rc.ctx)
	return err
}

// GetHash retrieves a hash from cache
func (rc *RedisCache) GetHash(key string) (map[string]string, error) {
	return rc.client.HGetAll(rc.ctx, key).Result()
}

// SetList stores a list in cache
func (rc *RedisCache) SetList(key string, values []interface{}, expiration time.Duration) error {
	pipe := rc.client.Pipeline()
	pipe.Del(rc.ctx, key) // Clear existing list
	if len(values) > 0 {
		pipe.LPush(rc.ctx, key, values...)
	}
	pipe.Expire(rc.ctx, key, expiration)
	_, err := pipe.Exec(rc.ctx)
	return err
}

// GetList retrieves a list from cache
func (rc *RedisCache) GetList(key string) ([]string, error) {
	return rc.client.LRange(rc.ctx, key, 0, -1).Result()
}

// Increment increments a counter
func (rc *RedisCache) Increment(key string, expiration time.Duration) (int64, error) {
	pipe := rc.client.Pipeline()
	incr := pipe.Incr(rc.ctx, key)
	pipe.Expire(rc.ctx, key, expiration)
	_, err := pipe.Exec(rc.ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// SetWithTags stores a value with tags for group invalidation
func (rc *RedisCache) SetWithTags(key string, value interface{}, tags []string, expiration time.Duration) error {
	// Store the main value
	if err := rc.Set(key, value, expiration); err != nil {
		return err
	}

	// Store tag associations
	for _, tag := range tags {
		tagKey := fmt.Sprintf("tag:%s", tag)
		rc.client.SAdd(rc.ctx, tagKey, key)
		rc.client.Expire(rc.ctx, tagKey, expiration+time.Hour) // Tags expire later
	}

	return nil
}

// InvalidateByTag removes all keys associated with a tag
func (rc *RedisCache) InvalidateByTag(tag string) error {
	tagKey := fmt.Sprintf("tag:%s", tag)
	keys, err := rc.client.SMembers(rc.ctx, tagKey).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		// Delete all keys associated with the tag
		rc.client.Del(rc.ctx, keys...)
		// Delete the tag set itself
		rc.client.Del(rc.ctx, tagKey)
	}

	return nil
}

// GetStats returns cache statistics
func (rc *RedisCache) GetStats() (map[string]string, error) {
	info, err := rc.client.Info(rc.ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]string)
	// Parse Redis INFO output (simplified)
	stats["info"] = info

	// Get memory usage
	memory, err := rc.client.Info(rc.ctx, "memory").Result()
	if err == nil {
		stats["memory"] = memory
	}

	return stats, nil
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// Cache key generators for different data types
const (
	DashboardCachePrefix    = "dashboard:"
	InventoryCachePrefix    = "inventory:"
	MenuCachePrefix         = "menu:"
	SupplierCachePrefix     = "supplier:"
	FinancialCachePrefix    = "financial:"
	NotificationCachePrefix = "notification:"
	UserCachePrefix         = "user:"
)

// GenerateDashboardKey generates cache key for dashboard data
func GenerateDashboardKey(userRole string, date string) string {
	return fmt.Sprintf("%s%s:%s", DashboardCachePrefix, userRole, date)
}

// GenerateInventoryKey generates cache key for inventory data
func GenerateInventoryKey(itemType string) string {
	return fmt.Sprintf("%s%s", InventoryCachePrefix, itemType)
}

// GenerateMenuKey generates cache key for menu data
func GenerateMenuKey(date string) string {
	return fmt.Sprintf("%s%s", MenuCachePrefix, date)
}

// GenerateSupplierKey generates cache key for supplier data
func GenerateSupplierKey(supplierID uint) string {
	return fmt.Sprintf("%s%d", SupplierCachePrefix, supplierID)
}

// GenerateFinancialKey generates cache key for financial data
func GenerateFinancialKey(reportType, period string) string {
	return fmt.Sprintf("%s%s:%s", FinancialCachePrefix, reportType, period)
}

// GenerateNotificationKey generates cache key for user notifications
func GenerateNotificationKey(userID uint) string {
	return fmt.Sprintf("%s%d", NotificationCachePrefix, userID)
}

// GenerateUserKey generates cache key for user data
func GenerateUserKey(userID uint) string {
	return fmt.Sprintf("%s%d", UserCachePrefix, userID)
}

// Cache tags for group invalidation
const (
	DashboardTag    = "dashboard"
	InventoryTag    = "inventory"
	MenuTag         = "menu"
	SupplierTag     = "supplier"
	FinancialTag    = "financial"
	NotificationTag = "notification"
	UserTag         = "user"
)

// Common cache durations
const (
	ShortCacheDuration  = 5 * time.Minute
	MediumCacheDuration = 30 * time.Minute
	LongCacheDuration   = 2 * time.Hour
	DayCacheDuration    = 24 * time.Hour
)

// ErrCacheMiss is returned when a key is not found in cache
var ErrCacheMiss = fmt.Errorf("cache miss")