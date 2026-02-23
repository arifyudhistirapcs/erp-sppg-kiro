package middleware

import (
	"net/http"
	"strings"

	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// JWTAuth middleware validates JWT token and sets user context
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"error_code": "UNAUTHORIZED",
				"message":    "Token autentikasi tidak ditemukan",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"error_code": "UNAUTHORIZED",
				"message":    "Format token tidak valid",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		authService := services.NewAuthService(nil, jwtSecret)
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"error_code": "UNAUTHORIZED",
				"message":    "Token tidak valid atau sudah kadaluarsa",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"error_code": "UNAUTHORIZED",
				"message":    "Autentikasi diperlukan",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		allowed := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "FORBIDDEN",
				"message":    "Anda tidak memiliki izin untuk mengakses resource ini",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionChecker defines permission checking logic
type PermissionChecker struct {
	permissions map[string][]string // feature -> allowed roles
}

// NewPermissionChecker creates a new permission checker with default permissions
func NewPermissionChecker() *PermissionChecker {
	pc := &PermissionChecker{
		permissions: make(map[string][]string),
	}

	// Define permissions based on requirements
	pc.permissions["dashboard_executive"] = []string{"kepala_sppg", "kepala_yayasan"}
	pc.permissions["financial_reports"] = []string{"kepala_sppg", "kepala_yayasan", "akuntan"}
	pc.permissions["menu_planning"] = []string{"kepala_sppg", "ahli_gizi"}
	pc.permissions["recipe_management"] = []string{"kepala_sppg", "ahli_gizi"}
	pc.permissions["kitchen_display"] = []string{"kepala_sppg", "ahli_gizi", "chef", "packing"}
	pc.permissions["procurement"] = []string{"kepala_sppg", "pengadaan"}
	pc.permissions["inventory"] = []string{"kepala_sppg", "akuntan", "pengadaan"}
	pc.permissions["delivery_tasks"] = []string{"kepala_sppg", "driver", "asisten_lapangan"}
	pc.permissions["attendance"] = []string{"kepala_sppg", "akuntan", "ahli_gizi", "pengadaan", "chef", "packing", "driver", "asisten_lapangan"}
	pc.permissions["hrm_management"] = []string{"kepala_sppg", "akuntan"}

	return pc
}

// CheckPermission checks if a role has permission for a feature
func (pc *PermissionChecker) CheckPermission(role, feature string) bool {
	allowedRoles, exists := pc.permissions[feature]
	if !exists {
		return false
	}

	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}

	return false
}

// RequirePermission middleware checks if user has permission for a feature
func RequirePermission(feature string) gin.HandlerFunc {
	pc := NewPermissionChecker()

	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"error_code": "UNAUTHORIZED",
				"message":    "Autentikasi diperlukan",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		if !pc.CheckPermission(role, feature) {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "FORBIDDEN",
				"message":    "Anda tidak memiliki izin untuk mengakses fitur ini",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
