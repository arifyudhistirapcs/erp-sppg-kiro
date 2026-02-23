package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SessionManager manages user sessions and enforces timeout
type SessionManager struct {
	sessions       map[uint]*SessionData
	mu             sync.RWMutex
	timeoutMinutes int
}

// SessionData represents session information
type SessionData struct {
	UserID       uint
	LastActivity time.Time
	CreatedAt    time.Time
}

// NewSessionManager creates a new session manager
func NewSessionManager(timeoutMinutes int) *SessionManager {
	sm := &SessionManager{
		sessions:       make(map[uint]*SessionData),
		timeoutMinutes: timeoutMinutes,
	}

	// Start cleanup goroutine
	go sm.cleanup()

	return sm
}

// cleanup removes expired sessions
func (sm *SessionManager) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		timeout := time.Duration(sm.timeoutMinutes) * time.Minute

		for userID, session := range sm.sessions {
			if now.Sub(session.LastActivity) > timeout {
				delete(sm.sessions, userID)
			}
		}
		sm.mu.Unlock()
	}
}

// UpdateActivity updates the last activity time for a user session
func (sm *SessionManager) UpdateActivity(userID uint) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[userID]
	if !exists {
		sm.sessions[userID] = &SessionData{
			UserID:       userID,
			LastActivity: time.Now(),
			CreatedAt:    time.Now(),
		}
	} else {
		session.LastActivity = time.Now()
	}
}

// IsSessionValid checks if a user session is still valid
func (sm *SessionManager) IsSessionValid(userID uint) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return false
	}

	timeout := time.Duration(sm.timeoutMinutes) * time.Minute
	return time.Since(session.LastActivity) <= timeout
}

// InvalidateSession removes a user session
func (sm *SessionManager) InvalidateSession(userID uint) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, userID)
}

// GetSessionInfo retrieves session information for a user
func (sm *SessionManager) GetSessionInfo(userID uint) (*SessionData, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[userID]
	return session, exists
}

// Global session manager instance
var globalSessionManager *SessionManager
var sessionManagerOnce sync.Once

// GetSessionManager returns the global session manager instance
func GetSessionManager(timeoutMinutes int) *SessionManager {
	sessionManagerOnce.Do(func() {
		globalSessionManager = NewSessionManager(timeoutMinutes)
	})
	return globalSessionManager
}

// SessionTimeoutMiddleware enforces session timeout
func SessionTimeoutMiddleware(timeoutMinutes int) gin.HandlerFunc {
	sessionManager := GetSessionManager(timeoutMinutes)

	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			// No user ID means not authenticated, skip session check
			c.Next()
			return
		}

		userID := userIDInterface.(uint)

		// Check if session is valid
		if !sessionManager.IsSessionValid(userID) {
			// Session has expired
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"error_code": "SESSION_EXPIRED",
				"message":    "Sesi Anda telah berakhir. Silakan login kembali.",
			})
			c.Abort()
			return
		}

		// Update last activity time
		sessionManager.UpdateActivity(userID)

		c.Next()
	}
}

// LogoutHandler invalidates a user session
func LogoutHandler(sessionManager *SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, exists := c.Get("user_id")
		if exists {
			userID := userIDInterface.(uint)
			sessionManager.InvalidateSession(userID)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Logout berhasil",
		})
	}
}
