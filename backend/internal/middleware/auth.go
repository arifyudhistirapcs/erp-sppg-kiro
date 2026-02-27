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
	
	// Logistics monitoring permissions - all roles except kebersihan
	pc.permissions["monitoring"] = []string{"kepala_sppg", "kepala_yayasan", "akuntan", "ahli_gizi", "pengadaan", "chef", "packing", "driver", "asisten_lapangan"}
	
	// Cleaning module permissions - kebersihan role only (plus admin override)
	pc.permissions["cleaning"] = []string{"kebersihan", "kepala_sppg", "kepala_yayasan"}

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

// StatusCategoryAuthorizer defines authorization logic for status updates
type StatusCategoryAuthorizer struct {
	statusRoleMap map[string][]string // status -> allowed roles
}

// NewStatusCategoryAuthorizer creates a new status category authorizer
func NewStatusCategoryAuthorizer() *StatusCategoryAuthorizer {
	sca := &StatusCategoryAuthorizer{
		statusRoleMap: make(map[string][]string),
	}

	// Cooking statuses - chef role
	sca.statusRoleMap["sedang_dimasak"] = []string{"chef", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["selesai_dimasak"] = []string{"chef", "kepala_sppg", "kepala_yayasan"}

	// Packing statuses - packing role
	sca.statusRoleMap["siap_dipacking"] = []string{"packing", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["selesai_dipacking"] = []string{"packing", "kepala_sppg", "kepala_yayasan"}

	// Delivery statuses - driver role
	sca.statusRoleMap["siap_dikirim"] = []string{"driver", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["diperjalanan"] = []string{"driver", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["sudah_sampai_sekolah"] = []string{"driver", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["sudah_diterima_pihak_sekolah"] = []string{"driver", "kepala_sppg", "kepala_yayasan"}

	// Collection statuses - driver role
	sca.statusRoleMap["driver_ditugaskan_mengambil_ompreng"] = []string{"driver", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["driver_menuju_sekolah"] = []string{"driver", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["driver_sampai_di_sekolah"] = []string{"driver", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["ompreng_telah_diambil"] = []string{"driver", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["ompreng_sampai_di_sppg"] = []string{"driver", "kepala_sppg", "kepala_yayasan"}

	// Cleaning statuses - kebersihan role
	sca.statusRoleMap["ompreng_proses_pencucian"] = []string{"kebersihan", "kepala_sppg", "kepala_yayasan"}
	sca.statusRoleMap["ompreng_selesai_dicuci"] = []string{"kebersihan", "kepala_sppg", "kepala_yayasan"}

	return sca
}

// CheckStatusUpdatePermission checks if a role has permission to update to a specific status
func (sca *StatusCategoryAuthorizer) CheckStatusUpdatePermission(role, status string) bool {
	allowedRoles, exists := sca.statusRoleMap[status]
	if !exists {
		// If status not in map, deny by default
		return false
	}

	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}

	return false
}

// ValidateStatusUpdatePermission checks if a user role has permission to update to a specific status
// This is a helper function to be called from handlers
// Returns true if the user has permission, false otherwise
func ValidateStatusUpdatePermission(role, status string) bool {
	sca := NewStatusCategoryAuthorizer()
	return sca.CheckStatusUpdatePermission(role, status)
}
