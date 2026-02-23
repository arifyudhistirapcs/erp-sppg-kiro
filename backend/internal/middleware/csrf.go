package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRFToken represents a CSRF token with expiry
type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

// CSRFManager manages CSRF tokens
type CSRFManager struct {
	tokens map[string]*CSRFToken
	mu     sync.RWMutex
}

// NewCSRFManager creates a new CSRF manager
func NewCSRFManager() *CSRFManager {
	cm := &CSRFManager{
		tokens: make(map[string]*CSRFToken),
	}

	// Start cleanup goroutine
	go cm.cleanup()

	return cm
}

// cleanup removes expired tokens
func (cm *CSRFManager) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()
		for sessionID, token := range cm.tokens {
			if now.After(token.ExpiresAt) {
				delete(cm.tokens, sessionID)
			}
		}
		cm.mu.Unlock()
	}
}

// GenerateToken generates a new CSRF token for a session
func (cm *CSRFManager) GenerateToken(sessionID string) (string, error) {
	// Generate random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(bytes)

	// Store token with expiry
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.tokens[sessionID] = &CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Token expires in 1 hour
	}

	return token, nil
}

// ValidateToken validates a CSRF token for a session
func (cm *CSRFManager) ValidateToken(sessionID, token string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	storedToken, exists := cm.tokens[sessionID]
	if !exists {
		return false
	}

	// Check if token has expired
	if time.Now().After(storedToken.ExpiresAt) {
		return false
	}

	// Compare tokens
	return storedToken.Token == token
}

// RemoveToken removes a CSRF token for a session
func (cm *CSRFManager) RemoveToken(sessionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.tokens, sessionID)
}

// Global CSRF manager instance
var globalCSRFManager *CSRFManager
var csrfManagerOnce sync.Once

// GetCSRFManager returns the global CSRF manager instance
func GetCSRFManager() *CSRFManager {
	csrfManagerOnce.Do(func() {
		globalCSRFManager = NewCSRFManager()
	})
	return globalCSRFManager
}

// CSRFMiddleware provides CSRF protection
func CSRFMiddleware() gin.HandlerFunc {
	csrfManager := GetCSRFManager()

	return func(c *gin.Context) {
		// Skip CSRF check for GET, HEAD, OPTIONS requests
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip CSRF check for auth endpoints (login, refresh)
		if c.FullPath() == "/api/v1/auth/login" || c.FullPath() == "/api/v1/auth/refresh" {
			c.Next()
			return
		}

		// Get session ID (use user ID or generate from IP + User-Agent)
		sessionID := getSessionID(c)

		// Get CSRF token from header
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "CSRF_TOKEN_MISSING",
				"message":    "Token CSRF diperlukan untuk permintaan ini.",
			})
			c.Abort()
			return
		}

		// Validate CSRF token
		if !csrfManager.ValidateToken(sessionID, token) {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "CSRF_TOKEN_INVALID",
				"message":    "Token CSRF tidak valid atau telah kedaluwarsa.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CSRFTokenHandler generates and returns a CSRF token
func CSRFTokenHandler() gin.HandlerFunc {
	csrfManager := GetCSRFManager()

	return func(c *gin.Context) {
		sessionID := getSessionID(c)

		token, err := csrfManager.GenerateToken(sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":    false,
				"error_code": "CSRF_TOKEN_GENERATION_FAILED",
				"message":    "Gagal membuat token CSRF.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"csrf_token": token,
			},
		})
	}
}

// getSessionID generates a session ID for CSRF token management
func getSessionID(c *gin.Context) string {
	// Try to get user ID from context first
	if userID, exists := c.Get("user_id"); exists {
		return "user_" + string(rune(userID.(uint)))
	}

	// Fallback to IP + User-Agent hash
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	return base64.URLEncoding.EncodeToString([]byte(ip + userAgent))
}