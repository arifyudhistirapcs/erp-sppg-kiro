package handlers

import (
	"net/http"

	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService  *services.AuthService
	auditService *services.AuditTrailService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService:  services.NewAuthService(db, jwtSecret),
		auditService: services.NewAuditTrailService(db),
	}
}

// LoginRequest represents login request body
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // NIK or Email
	Password   string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	User    *struct {
		ID       uint   `json:"id"`
		NIK      string `json:"nik"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	} `json:"user,omitempty"`
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details": []gin.H{
				{
					"field":   "identifier",
					"message": "NIK atau Email dan Password harus diisi",
				},
			},
		})
		return
	}

	// Authenticate user
	user, token, err := h.authService.Login(req.Identifier, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"error_code": "INVALID_CREDENTIALS",
				"message":    "NIK/Email atau password salah",
			})
			return
		}

		if err == services.ErrUserInactive {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "USER_INACTIVE",
				"message":    "Akun Anda tidak aktif. Silakan hubungi administrator",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Record login in audit trail
	ipAddress := c.ClientIP()
	h.auditService.RecordLogin(user.ID, ipAddress)

	// Return success response
	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "Login berhasil",
		Token:   token,
		User: &struct {
			ID       uint   `json:"id"`
			NIK      string `json:"nik"`
			Email    string `json:"email"`
			FullName string `json:"full_name"`
			Role     string `json:"role"`
		}{
			ID:       user.ID,
			NIK:      user.NIK,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
		},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user ID from context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "Autentikasi diperlukan",
		})
		return
	}

	// Record logout in audit trail
	ipAddress := c.ClientIP()
	h.auditService.RecordLogout(userID.(uint), ipAddress)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout berhasil",
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "Token tidak ditemukan",
		})
		return
	}

	// Extract token
	var tokenString string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "Format token tidak valid",
		})
		return
	}

	// Refresh token
	newToken, err := h.authService.RefreshToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "INVALID_TOKEN",
			"message":    "Token tidak valid atau sudah kadaluarsa",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token berhasil diperbarui",
		"token":   newToken,
	})
}

// GetMe returns current user information
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Get user ID from context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "Autentikasi diperlukan",
		})
		return
	}

	// Get user details
	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "USER_NOT_FOUND",
				"message":    "Pengguna tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":           user.ID,
			"nik":          user.NIK,
			"email":        user.Email,
			"full_name":    user.FullName,
			"phone_number": user.PhoneNumber,
			"role":         user.Role,
			"is_active":    user.IsActive,
		},
	})
}
