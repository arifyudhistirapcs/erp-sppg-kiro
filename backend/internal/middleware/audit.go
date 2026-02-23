package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditTrail middleware records all create/update/delete actions
func AuditTrail(db *gorm.DB) gin.HandlerFunc {
	auditService := services.NewAuditTrailService(db)

	return func(c *gin.Context) {
		// Only audit CUD operations
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
			c.Next()
			return
		}

		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// Read request body for old/new values
		var requestBody map[string]interface{}
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			json.Unmarshal(bodyBytes, &requestBody)
		}

		// Determine action type
		action := ""
		switch method {
		case "POST":
			action = "create"
		case "PUT", "PATCH":
			action = "update"
		case "DELETE":
			action = "delete"
		}

		// Extract entity from path
		path := c.Request.URL.Path
		parts := strings.Split(strings.Trim(path, "/"), "/")
		entity := ""
		entityID := ""

		// Try to extract entity name from path (e.g., /api/v1/recipes -> recipes)
		if len(parts) >= 3 {
			entity = parts[2]
		}

		// Try to extract entity ID from path (e.g., /api/v1/recipes/123 -> 123)
		if len(parts) >= 4 {
			entityID = parts[3]
		}

		// Get client IP
		ipAddress := c.ClientIP()

		// Process request
		c.Next()

		// Only record if request was successful (2xx status)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// For updates, we don't have old value here (would need to fetch before update)
			// For creates, old value is nil
			// For deletes, new value is nil
			var oldValue, newValue interface{}

			switch action {
			case "create":
				newValue = requestBody
			case "update":
				newValue = requestBody
			case "delete":
				oldValue = requestBody
			}

			// Record audit trail (ignore errors to not affect main request)
			auditService.RecordAction(userID.(uint), action, entity, entityID, oldValue, newValue, ipAddress)
		}
	}
}
