package cache

import (
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
)

// CacheService provides high-level caching operations
type CacheService struct {
	redis *RedisCache
}

// NewCacheService creates a new cache service
func NewCacheService(redis *RedisCache) *CacheService {
	return &CacheService{
		redis: redis,
	}
}

// Dashboard caching methods

// SetDashboardData caches dashboard data for a specific role and date
func (cs *CacheService) SetDashboardData(userRole string, date string, data map[string]interface{}) error {
	key := GenerateDashboardKey(userRole, date)
	return cs.redis.SetWithTags(key, data, []string{DashboardTag}, MediumCacheDuration)
}

// GetDashboardData retrieves cached dashboard data
func (cs *CacheService) GetDashboardData(userRole string, date string) (map[string]interface{}, error) {
	key := GenerateDashboardKey(userRole, date)
	var data map[string]interface{}
	err := cs.redis.Get(key, &data)
	return data, err
}

// InvalidateDashboardCache invalidates all dashboard cache
func (cs *CacheService) InvalidateDashboardCache() error {
	return cs.redis.InvalidateByTag(DashboardTag)
}

// Inventory caching methods

// SetInventoryItems caches inventory items
func (cs *CacheService) SetInventoryItems(items []models.InventoryItem) error {
	key := GenerateInventoryKey("items")
	return cs.redis.SetWithTags(key, items, []string{InventoryTag}, MediumCacheDuration)
}

// GetInventoryItems retrieves cached inventory items
func (cs *CacheService) GetInventoryItems() ([]models.InventoryItem, error) {
	key := GenerateInventoryKey("items")
	var items []models.InventoryItem
	err := cs.redis.Get(key, &items)
	return items, err
}

// SetLowStockItems caches low stock items
func (cs *CacheService) SetLowStockItems(items []models.InventoryItem) error {
	key := GenerateInventoryKey("low_stock")
	return cs.redis.SetWithTags(key, items, []string{InventoryTag}, ShortCacheDuration)
}

// GetLowStockItems retrieves cached low stock items
func (cs *CacheService) GetLowStockItems() ([]models.InventoryItem, error) {
	key := GenerateInventoryKey("low_stock")
	var items []models.InventoryItem
	err := cs.redis.Get(key, &items)
	return items, err
}

// InvalidateInventoryCache invalidates all inventory cache
func (cs *CacheService) InvalidateInventoryCache() error {
	return cs.redis.InvalidateByTag(InventoryTag)
}

// Menu caching methods

// SetMenuPlan caches menu plan for a specific date
func (cs *CacheService) SetMenuPlan(date string, menuPlan *models.MenuPlan) error {
	key := GenerateMenuKey(date)
	return cs.redis.SetWithTags(key, menuPlan, []string{MenuTag}, LongCacheDuration)
}

// GetMenuPlan retrieves cached menu plan
func (cs *CacheService) GetMenuPlan(date string) (*models.MenuPlan, error) {
	key := GenerateMenuKey(date)
	var menuPlan models.MenuPlan
	err := cs.redis.Get(key, &menuPlan)
	return &menuPlan, err
}

// SetRecipes caches active recipes
func (cs *CacheService) SetRecipes(recipes []models.Recipe) error {
	key := GenerateMenuKey("recipes")
	return cs.redis.SetWithTags(key, recipes, []string{MenuTag}, LongCacheDuration)
}

// GetRecipes retrieves cached recipes
func (cs *CacheService) GetRecipes() ([]models.Recipe, error) {
	key := GenerateMenuKey("recipes")
	var recipes []models.Recipe
	err := cs.redis.Get(key, &recipes)
	return recipes, err
}

// InvalidateMenuCache invalidates all menu cache
func (cs *CacheService) InvalidateMenuCache() error {
	return cs.redis.InvalidateByTag(MenuTag)
}

// Supplier caching methods

// SetSupplierPerformance caches supplier performance data
func (cs *CacheService) SetSupplierPerformance(supplierID uint, data map[string]interface{}) error {
	key := fmt.Sprintf("%sperformance:%d", SupplierCachePrefix, supplierID)
	return cs.redis.SetWithTags(key, data, []string{SupplierTag}, LongCacheDuration)
}

// GetSupplierPerformance retrieves cached supplier performance data
func (cs *CacheService) GetSupplierPerformance(supplierID uint) (map[string]interface{}, error) {
	key := fmt.Sprintf("%sperformance:%d", SupplierCachePrefix, supplierID)
	var data map[string]interface{}
	err := cs.redis.Get(key, &data)
	return data, err
}

// SetActiveSuppliers caches active suppliers
func (cs *CacheService) SetActiveSuppliers(suppliers []models.Supplier) error {
	key := fmt.Sprintf("%sactive", SupplierCachePrefix)
	return cs.redis.SetWithTags(key, suppliers, []string{SupplierTag}, LongCacheDuration)
}

// GetActiveSuppliers retrieves cached active suppliers
func (cs *CacheService) GetActiveSuppliers() ([]models.Supplier, error) {
	key := fmt.Sprintf("%sactive", SupplierCachePrefix)
	var suppliers []models.Supplier
	err := cs.redis.Get(key, &suppliers)
	return suppliers, err
}

// InvalidateSupplierCache invalidates all supplier cache
func (cs *CacheService) InvalidateSupplierCache() error {
	return cs.redis.InvalidateByTag(SupplierTag)
}

// Financial caching methods

// SetFinancialReport caches financial report data
func (cs *CacheService) SetFinancialReport(reportType, period string, data map[string]interface{}) error {
	key := GenerateFinancialKey(reportType, period)
	return cs.redis.SetWithTags(key, data, []string{FinancialTag}, MediumCacheDuration)
}

// GetFinancialReport retrieves cached financial report data
func (cs *CacheService) GetFinancialReport(reportType, period string) (map[string]interface{}, error) {
	key := GenerateFinancialKey(reportType, period)
	var data map[string]interface{}
	err := cs.redis.Get(key, &data)
	return data, err
}

// SetCashFlowSummary caches cash flow summary
func (cs *CacheService) SetCashFlowSummary(period string, summary map[string]interface{}) error {
	key := fmt.Sprintf("%scash_flow_summary:%s", FinancialCachePrefix, period)
	return cs.redis.SetWithTags(key, summary, []string{FinancialTag}, MediumCacheDuration)
}

// GetCashFlowSummary retrieves cached cash flow summary
func (cs *CacheService) GetCashFlowSummary(period string) (map[string]interface{}, error) {
	key := fmt.Sprintf("%scash_flow_summary:%s", FinancialCachePrefix, period)
	var summary map[string]interface{}
	err := cs.redis.Get(key, &summary)
	return summary, err
}

// InvalidateFinancialCache invalidates all financial cache
func (cs *CacheService) InvalidateFinancialCache() error {
	return cs.redis.InvalidateByTag(FinancialTag)
}

// Notification caching methods

// SetUserNotifications caches user notifications
func (cs *CacheService) SetUserNotifications(userID uint, notifications []models.Notification) error {
	key := GenerateNotificationKey(userID)
	return cs.redis.SetWithTags(key, notifications, []string{NotificationTag}, ShortCacheDuration)
}

// GetUserNotifications retrieves cached user notifications
func (cs *CacheService) GetUserNotifications(userID uint) ([]models.Notification, error) {
	key := GenerateNotificationKey(userID)
	var notifications []models.Notification
	err := cs.redis.Get(key, &notifications)
	return notifications, err
}

// SetUnreadNotificationCount caches unread notification count for a user
func (cs *CacheService) SetUnreadNotificationCount(userID uint, count int64) error {
	key := fmt.Sprintf("%sunread_count:%d", NotificationCachePrefix, userID)
	return cs.redis.SetWithTags(key, count, []string{NotificationTag}, ShortCacheDuration)
}

// GetUnreadNotificationCount retrieves cached unread notification count
func (cs *CacheService) GetUnreadNotificationCount(userID uint) (int64, error) {
	key := fmt.Sprintf("%sunread_count:%d", NotificationCachePrefix, userID)
	var count int64
	err := cs.redis.Get(key, &count)
	return count, err
}

// InvalidateUserNotifications invalidates notifications for a specific user
func (cs *CacheService) InvalidateUserNotifications(userID uint) error {
	pattern := fmt.Sprintf("%s%d*", NotificationCachePrefix, userID)
	return cs.redis.DeletePattern(pattern)
}

// InvalidateNotificationCache invalidates all notification cache
func (cs *CacheService) InvalidateNotificationCache() error {
	return cs.redis.InvalidateByTag(NotificationTag)
}

// User caching methods

// SetUserProfile caches user profile data
func (cs *CacheService) SetUserProfile(userID uint, user *models.User) error {
	key := GenerateUserKey(userID)
	return cs.redis.SetWithTags(key, user, []string{UserTag}, LongCacheDuration)
}

// GetUserProfile retrieves cached user profile data
func (cs *CacheService) GetUserProfile(userID uint) (*models.User, error) {
	key := GenerateUserKey(userID)
	var user models.User
	err := cs.redis.Get(key, &user)
	return &user, err
}

// InvalidateUserProfile invalidates cache for a specific user
func (cs *CacheService) InvalidateUserProfile(userID uint) error {
	key := GenerateUserKey(userID)
	return cs.redis.Delete(key)
}

// InvalidateUserCache invalidates all user cache
func (cs *CacheService) InvalidateUserCache() error {
	return cs.redis.InvalidateByTag(UserTag)
}

// General utility methods

// GetCachedResponse retrieves a cached HTTP response
func (cs *CacheService) GetCachedResponse(key string, response interface{}) error {
	return cs.redis.Get(key, response)
}

// SetCachedResponse stores an HTTP response in cache
func (cs *CacheService) SetCachedResponse(key string, response interface{}, duration time.Duration) error {
	return cs.redis.SetWithTags(key, response, []string{"http_cache"}, duration)
}

// WarmupCache preloads frequently accessed data into cache
func (cs *CacheService) WarmupCache() error {
	// This method would be called during application startup
	// to preload frequently accessed data
	
	// Example: Preload active suppliers, recipes, etc.
	// Implementation would depend on specific business requirements
	
	return nil
}

// ClearAllCache clears all cached data
func (cs *CacheService) ClearAllCache() error {
	tags := []string{DashboardTag, InventoryTag, MenuTag, SupplierTag, FinancialTag, NotificationTag, UserTag}
	
	for _, tag := range tags {
		if err := cs.redis.InvalidateByTag(tag); err != nil {
			return err
		}
	}
	
	return nil
}

// GetCacheStats returns cache statistics
func (cs *CacheService) GetCacheStats() (map[string]string, error) {
	return cs.redis.GetStats()
}

// Cache invalidation helpers for common operations

// InvalidateOnInventoryChange invalidates caches that depend on inventory data
func (cs *CacheService) InvalidateOnInventoryChange() error {
	// Invalidate inventory cache
	if err := cs.InvalidateInventoryCache(); err != nil {
		return err
	}
	
	// Invalidate dashboard cache (which includes inventory metrics)
	if err := cs.InvalidateDashboardCache(); err != nil {
		return err
	}
	
	return nil
}

// InvalidateOnMenuChange invalidates caches that depend on menu data
func (cs *CacheService) InvalidateOnMenuChange() error {
	// Invalidate menu cache
	if err := cs.InvalidateMenuCache(); err != nil {
		return err
	}
	
	// Invalidate dashboard cache
	if err := cs.InvalidateDashboardCache(); err != nil {
		return err
	}
	
	return nil
}

// InvalidateOnFinancialChange invalidates caches that depend on financial data
func (cs *CacheService) InvalidateOnFinancialChange() error {
	// Invalidate financial cache
	if err := cs.InvalidateFinancialCache(); err != nil {
		return err
	}
	
	// Invalidate dashboard cache
	if err := cs.InvalidateDashboardCache(); err != nil {
		return err
	}
	
	return nil
}

// InvalidateOnSupplierChange invalidates caches that depend on supplier data
func (cs *CacheService) InvalidateOnSupplierChange() error {
	// Invalidate supplier cache
	if err := cs.InvalidateSupplierCache(); err != nil {
		return err
	}
	
	// Invalidate dashboard cache
	if err := cs.InvalidateDashboardCache(); err != nil {
		return err
	}
	
	return nil
}