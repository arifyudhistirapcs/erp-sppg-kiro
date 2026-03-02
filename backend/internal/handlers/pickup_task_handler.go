package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PickupTaskHandler handles pickup task endpoints
type PickupTaskHandler struct {
	service *services.PickupTaskService
}

// NewPickupTaskHandler creates a new pickup task handler
func NewPickupTaskHandler(service *services.PickupTaskService) *PickupTaskHandler {
	return &PickupTaskHandler{
		service: service,
	}
}

// GetEligibleOrders handles GET /api/v1/pickup-tasks/eligible-orders
// Returns delivery records that are ready for pickup (Stage 9)
func (h *PickupTaskHandler) GetEligibleOrders(c *gin.Context) {
	// Parse optional date query parameter
	dateStr := c.Query("date")
	var date time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(400, gin.H{
				"error": gin.H{
					"code":    "INVALID_DATE",
					"message": "Invalid date format. Use YYYY-MM-DD",
				},
			})
			return
		}
		date = parsedDate
	}

	// Call service to get eligible orders
	eligibleOrders, err := h.service.GetEligibleOrders(date)
	if err != nil {
		c.JSON(500, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch eligible orders",
				"details": err.Error(),
			},
		})
		return
	}

	// Return JSON response
	c.JSON(200, gin.H{
		"eligible_orders": eligibleOrders,
	})
}

// GetAvailableDrivers handles GET /api/v1/pickup-tasks/available-drivers
// Returns drivers available for pickup task assignment
func (h *PickupTaskHandler) GetAvailableDrivers(c *gin.Context) {
	// Parse optional date query parameter
	dateStr := c.Query("date")
	var date time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(400, gin.H{
				"error": gin.H{
					"code":    "INVALID_DATE",
					"message": "Invalid date format. Use YYYY-MM-DD",
				},
			})
			return
		}
		date = parsedDate
	}

	// Call service to get available drivers
	availableDrivers, err := h.service.GetAvailableDrivers(date)
	if err != nil {
		c.JSON(500, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch available drivers",
				"details": err.Error(),
			},
		})
		return
	}

	// Return JSON response
	c.JSON(200, gin.H{
		"available_drivers": availableDrivers,
	})
}

// CreatePickupTask handles POST /api/v1/pickup-tasks
// Creates a new pickup task with driver assignment and route order
func (h *PickupTaskHandler) CreatePickupTask(c *gin.Context) {
	var req services.CreatePickupTaskRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate request structure
	if req.DriverID == 0 {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Driver ID is required",
			},
		})
		return
	}

	if len(req.DeliveryRecords) == 0 {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "At least one delivery record is required",
			},
		})
		return
	}

	// Call service to create pickup task
	pickupTask, err := h.service.CreatePickupTask(req)
	if err != nil {
		// Determine appropriate status code based on error
		statusCode := 500
		errorCode := "INTERNAL_ERROR"
		
		// Check for specific error types
		errMsg := err.Error()
		if contains(errMsg, "not found") || contains(errMsg, "not a driver") {
			statusCode = 404
			errorCode = "NOT_FOUND"
		} else if contains(errMsg, "not at stage 9") || contains(errMsg, "duplicate route_order") {
			statusCode = 400
			errorCode = "VALIDATION_ERROR"
		} else if contains(errMsg, "already assigned") {
			statusCode = 409
			errorCode = "ALREADY_ASSIGNED"
		}

		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	// Return 201 Created with pickup task details
	c.JSON(201, gin.H{
		"pickup_task": pickupTask,
	})
}

// GetAllPickupTasks handles GET /api/v1/pickup-tasks
// Returns all pickup tasks with optional filters
func (h *PickupTaskHandler) GetAllPickupTasks(c *gin.Context) {
	// Parse query parameters
	dateStr := c.Query("date")
	driverIDStr := c.Query("driver_id")
	statusFilter := c.Query("status")

	var date time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(400, gin.H{
				"error": gin.H{
					"code":    "INVALID_DATE",
					"message": "Invalid date format. Use YYYY-MM-DD",
				},
			})
			return
		}
		date = parsedDate
	}

	var driverID *uint
	if driverIDStr != "" {
		var id uint
		if _, err := fmt.Sscanf(driverIDStr, "%d", &id); err != nil {
			c.JSON(400, gin.H{
				"error": gin.H{
					"code":    "INVALID_DRIVER_ID",
					"message": "Invalid driver ID format",
				},
			})
			return
		}
		driverID = &id
	}

	// Call service to get pickup tasks
	pickupTasks, err := h.service.GetActivePickupTasks(date, driverID, statusFilter)
	if err != nil {
		c.JSON(500, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch pickup tasks",
				"details": err.Error(),
			},
		})
		return
	}

	// Filter by status if provided
	if statusFilter != "" {
		filtered := make([]models.PickupTask, 0)
		for _, task := range pickupTasks {
			if task.Status == statusFilter {
				filtered = append(filtered, task)
			}
		}
		pickupTasks = filtered
	}

	// Add computed fields for each pickup task
	type PickupTaskResponse struct {
		models.PickupTask
		SchoolCount    int `json:"school_count"`
		CompletedCount int `json:"completed_count"`
	}

	response := make([]PickupTaskResponse, len(pickupTasks))
	for i, task := range pickupTasks {
		schoolCount := len(task.DeliveryRecords)
		completedCount := 0
		for _, dr := range task.DeliveryRecords {
			if dr.CurrentStage == 13 {
				completedCount++
			}
		}

		response[i] = PickupTaskResponse{
			PickupTask:     task,
			SchoolCount:    schoolCount,
			CompletedCount: completedCount,
		}
	}

	// Return JSON response
	c.JSON(200, gin.H{
		"pickup_tasks": response,
	})
}

// GetPickupTask handles GET /api/v1/pickup-tasks/:id
// Returns detailed information about a specific pickup task
func (h *PickupTaskHandler) GetPickupTask(c *gin.Context) {
	// Parse pickup task ID from URL parameter
	idStr := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid pickup task ID format",
			},
		})
		return
	}

	// Call service to get pickup task by ID
	pickupTask, err := h.service.GetPickupTaskByID(id)
	if err != nil {
		statusCode := 500
		errorCode := "INTERNAL_ERROR"
		
		if contains(err.Error(), "not found") {
			statusCode = 404
			errorCode = "NOT_FOUND"
		}

		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	// Return detailed pickup task
	c.JSON(200, gin.H{
		"pickup_task": pickupTask,
	})
}

// UpdatePickupTaskStatus handles PUT /api/v1/pickup-tasks/:id/status
// Updates the status of a pickup task
func (h *PickupTaskHandler) UpdatePickupTaskStatus(c *gin.Context) {
	// Parse pickup task ID from URL parameter
	idStr := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid pickup task ID format",
			},
		})
		return
	}

	// Parse request body
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate status value
	validStatuses := map[string]bool{
		"active":    true,
		"completed": true,
		"cancelled": true,
	}
	if !validStatuses[req.Status] {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "INVALID_STATUS",
				"message": "Status must be one of: active, completed, cancelled",
			},
		})
		return
	}

	// Call service to update status
	err := h.service.UpdatePickupTaskStatus(id, req.Status)
	if err != nil {
		statusCode := 500
		errorCode := "INTERNAL_ERROR"
		
		if contains(err.Error(), "not found") {
			statusCode = 404
			errorCode = "NOT_FOUND"
		}

		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	// Get updated pickup task
	pickupTask, err := h.service.GetPickupTaskByID(id)
	if err != nil {
		c.JSON(500, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Status updated but failed to fetch updated task",
			},
		})
		return
	}

	// Return success response
	c.JSON(200, gin.H{
		"message":     "Pickup task status updated successfully",
		"pickup_task": pickupTask,
	})
}

// CancelPickupTask handles DELETE /api/v1/pickup-tasks/:id
// Cancels a pickup task (soft delete by setting status to cancelled)
func (h *PickupTaskHandler) CancelPickupTask(c *gin.Context) {
	// Parse pickup task ID from URL parameter
	idStr := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid pickup task ID format",
			},
		})
		return
	}

	// Get user ID from context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User ID not found in context",
			},
		})
		return
	}

	// Call service to cancel pickup task
	err := h.service.CancelPickupTask(id, userID.(uint))
	if err != nil {
		statusCode := 500
		errorCode := "INTERNAL_ERROR"
		
		if contains(err.Error(), "not found") {
			statusCode = 404
			errorCode = "NOT_FOUND"
		}

		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	// Return success message
	c.JSON(200, gin.H{
		"message": "Pickup task cancelled successfully",
	})
}
// UpdateDeliveryRecordStage handles PUT /api/v1/pickup-tasks/:pickup_task_id/delivery-records/:delivery_record_id/stage
// Updates the stage of an individual delivery record within a pickup task
func (h *PickupTaskHandler) UpdateDeliveryRecordStage(c *gin.Context) {
	// Parse pickup task ID from URL parameter
	pickupTaskIDStr := c.Param("id")
	var pickupTaskID uint
	if _, err := fmt.Sscanf(pickupTaskIDStr, "%d", &pickupTaskID); err != nil {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid pickup task ID format",
			},
		})
		return
	}

	// Parse delivery record ID from URL parameter
	deliveryRecordIDStr := c.Param("delivery_record_id")
	var deliveryRecordID uint
	if _, err := fmt.Sscanf(deliveryRecordIDStr, "%d", &deliveryRecordID); err != nil {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid delivery record ID format",
			},
		})
		return
	}

	// Parse request body
	var req struct {
		Stage  int    `json:"stage" binding:"required"`
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User ID not found in context",
			},
		})
		return
	}

	// Call service to update delivery record stage
	deliveryRecord, err := h.service.UpdateDeliveryRecordStage(pickupTaskID, deliveryRecordID, req.Stage, req.Status, userID.(uint))
	if err != nil {
		// Determine appropriate status code based on error
		statusCode := 500
		errorCode := "INTERNAL_ERROR"

		errMsg := err.Error()
		if contains(errMsg, "not found") {
			statusCode = 404
			errorCode = "DELIVERY_RECORD_NOT_FOUND"
		} else if contains(errMsg, "not part of pickup task") {
			statusCode = 400
			errorCode = "DELIVERY_RECORD_NOT_IN_TASK"
		} else if contains(errMsg, "cannot skip stages") || contains(errMsg, "cannot update stage") || contains(errMsg, "invalid stage") || contains(errMsg, "invalid status") {
			statusCode = 400
			errorCode = "INVALID_STAGE_TRANSITION"
		}

		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	// Return success response
	c.JSON(200, gin.H{
		"message":         "Delivery record stage updated successfully",
		"delivery_record": deliveryRecord,
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
