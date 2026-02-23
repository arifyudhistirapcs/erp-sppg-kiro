package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort = errors.New("password harus minimal 8 karakter")
	ErrPasswordTooWeak  = errors.New("password harus mengandung huruf besar, huruf kecil, dan angka")
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePasswordStrength validates password strength
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrPasswordTooWeak
	}

	return nil
}

// GenerateRandomToken generates a random token for password reset, etc.
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SessionInfo represents session information
type SessionInfo struct {
	UserID       uint
	LastActivity time.Time
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// IsSessionExpired checks if a session has expired based on timeout
func IsSessionExpired(lastActivity time.Time, timeoutMinutes int) bool {
	timeout := time.Duration(timeoutMinutes) * time.Minute
	return time.Since(lastActivity) > timeout
}

// IsSessionValid checks if a session is still valid
func IsSessionValid(session *SessionInfo, timeoutMinutes int) bool {
	// Check if session has expired based on absolute expiry
	if time.Now().After(session.ExpiresAt) {
		return false
	}

	// Check if session has timed out due to inactivity
	if IsSessionExpired(session.LastActivity, timeoutMinutes) {
		return false
	}

	return true
}

// UpdateSessionActivity updates the last activity time of a session
func UpdateSessionActivity(session *SessionInfo) {
	session.LastActivity = time.Now()
}

// CreateSession creates a new session with expiry
func CreateSession(userID uint, expiryHours int) *SessionInfo {
	now := time.Now()
	return &SessionInfo{
		UserID:       userID,
		LastActivity: now,
		CreatedAt:    now,
		ExpiresAt:    now.Add(time.Duration(expiryHours) * time.Hour),
	}
}
