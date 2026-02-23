package middleware

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/erp-sppg/backend/internal/cache"
	"github.com/gin-gonic/gin"
)

// CacheMiddleware provides HTTP response caching
func CacheMiddleware(cacheService *cache.CacheService, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip caching for non-GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Generate cache key based on URL and query parameters
		cacheKey := generateCacheKey(c)

		// Try to get cached response
		var cachedResponse CachedResponse
		err := getCachedResponse(cacheService, cacheKey, &cachedResponse)
		if err == nil {
			// Cache hit - return cached response
			c.Header("X-Cache", "HIT")
			c.Header("X-Cache-Key", cacheKey)
			
			// Set cached headers
			for key, value := range cachedResponse.Headers {
				c.Header(key, value)
			}
			
			c.Data(cachedResponse.StatusCode, cachedResponse.ContentType, cachedResponse.Body)
			c.Abort()
			return
		}

		// Cache miss - continue with request processing
		c.Header("X-Cache", "MISS")
		c.Header("X-Cache-Key", cacheKey)

		// Capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:          make([]byte, 0),
			statusCode:    http.StatusOK,
		}
		c.Writer = writer

		c.Next()

		// Cache the response if it's successful
		if writer.statusCode >= 200 && writer.statusCode < 300 {
			cachedResponse := CachedResponse{
				StatusCode:  writer.statusCode,
				ContentType: writer.Header().Get("Content-Type"),
				Headers:     make(map[string]string),
				Body:        writer.body,
				CachedAt:    time.Now(),
			}

			// Copy important headers
			for _, header := range []string{"Content-Type", "Content-Encoding", "ETag"} {
				if value := writer.Header().Get(header); value != "" {
					cachedResponse.Headers[header] = value
				}
			}

			// Cache the response
			setCachedResponse(cacheService, cacheKey, cachedResponse, duration)
		}
	}
}

// DashboardCacheMiddleware provides specialized caching for dashboard endpoints
func DashboardCacheMiddleware(cacheService *cache.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip caching for non-GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Get user role from context
		userRole, exists := c.Get("user_role")
		if !exists {
			c.Next()
			return
		}

		// Generate cache key based on role and date
		today := time.Now().Format("2006-01-02")

		// Try to get cached dashboard data
		data, err := cacheService.GetDashboardData(userRole.(string), today)
		if err == nil {
			// Cache hit
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    data,
			})
			c.Abort()
			return
		}

		// Cache miss - continue with request processing
		c.Header("X-Cache", "MISS")
		c.Next()
	}
}

// InventoryCacheMiddleware provides specialized caching for inventory endpoints
func InventoryCacheMiddleware(cacheService *cache.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Check for low stock alerts endpoint
		if strings.Contains(c.Request.URL.Path, "alerts") {
			items, err := cacheService.GetLowStockItems()
			if err == nil {
				c.Header("X-Cache", "HIT")
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    items,
				})
				c.Abort()
				return
			}
		} else {
			// Regular inventory items
			items, err := cacheService.GetInventoryItems()
			if err == nil {
				c.Header("X-Cache", "HIT")
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    items,
				})
				c.Abort()
				return
			}
		}

		c.Header("X-Cache", "MISS")
		c.Next()
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode  int               `json:"status_code"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	CachedAt    time.Time         `json:"cached_at"`
}

// responseWriter captures response data for caching
type responseWriter struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// generateCacheKey creates a unique cache key for the request
func generateCacheKey(c *gin.Context) string {
	// Include path, query parameters, and user ID (if available)
	key := c.Request.URL.Path
	
	if c.Request.URL.RawQuery != "" {
		key += "?" + c.Request.URL.RawQuery
	}

	// Include user ID for user-specific caching
	if userID, exists := c.Get("user_id"); exists {
		key += fmt.Sprintf(":user:%d", userID)
	}

	// Create MD5 hash of the key to keep it short
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("http:%x", hash)
}

// CacheInvalidationMiddleware invalidates cache on data modifications
func CacheInvalidationMiddleware(cacheService *cache.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store original path for later use
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		// Only invalidate cache for successful modifications
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Determine what cache to invalidate based on the endpoint
			switch {
			case strings.Contains(path, "/inventory"):
				if method != "GET" {
					cacheService.InvalidateOnInventoryChange()
				}
			case strings.Contains(path, "/menu") || strings.Contains(path, "/recipe"):
				if method != "GET" {
					cacheService.InvalidateOnMenuChange()
				}
			case strings.Contains(path, "/cash-flow") || strings.Contains(path, "/financial"):
				if method != "GET" {
					cacheService.InvalidateOnFinancialChange()
				}
			case strings.Contains(path, "/supplier"):
				if method != "GET" {
					cacheService.InvalidateOnSupplierChange()
				}
			case strings.Contains(path, "/notification"):
				if method != "GET" {
					// Invalidate notifications for the specific user
					if userID, exists := c.Get("user_id"); exists {
						cacheService.InvalidateUserNotifications(userID.(uint))
					}
				}
			}
		}
	}
}

// ConditionalCacheMiddleware applies caching based on conditions
func ConditionalCacheMiddleware(cacheService *cache.CacheService, condition func(*gin.Context) bool, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !condition(c) {
			c.Next()
			return
		}

		CacheMiddleware(cacheService, duration)(c)
	}
}

// Cache condition functions

// CacheForAuthenticatedUsers returns true if user is authenticated
func CacheForAuthenticatedUsers(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// CacheForSpecificRoles returns a condition function that checks user role
func CacheForSpecificRoles(roles ...string) func(*gin.Context) bool {
	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role] = true
	}

	return func(c *gin.Context) bool {
		userRole, exists := c.Get("user_role")
		if !exists {
			return false
		}
		return roleMap[userRole.(string)]
	}
}

// CacheForReadOnlyOperations returns true for GET requests only
func CacheForReadOnlyOperations(c *gin.Context) bool {
	return c.Request.Method == "GET"
}

// Helper function to set cached response
func setCachedResponse(cacheService *cache.CacheService, key string, response CachedResponse, duration time.Duration) error {
	return cacheService.SetCachedResponse(key, response, duration)
}

// Helper function to get cached response
func getCachedResponse(cacheService *cache.CacheService, key string, response *CachedResponse) error {
	return cacheService.GetCachedResponse(key, response)
}