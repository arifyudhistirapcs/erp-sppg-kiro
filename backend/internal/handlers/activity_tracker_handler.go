package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ActivityTrackerHandler handles HTTP requests for Activity Tracker
type ActivityTrackerHandler struct {
	service *services.ActivityTrackerService
}

// NewActivityTrackerHandler creates a new ActivityTrackerHandler instance
func NewActivityTrackerHandler(service *services.ActivityTrackerService) *ActivityTrackerHandler {
	return &ActivityTrackerHandler{
		service: service,
	}
}

// GetOrdersByDate retrieves all orders for a specific date with optional filters
// GET /api/activity-tracker/orders?date=2024-01-15&school_id=5&search=nasi
func (h *ActivityTrackerHandler) GetOrdersByDate(c *gin.Context) {
	// Parse date parameter
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "date parameter is required (format: YYYY-MM-DD)",
		})
		return
	}
	
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid date format, expected YYYY-MM-DD",
		})
		return
	}
	
	// Parse optional school_id parameter
	var schoolID *uint
	if schoolIDStr := c.Query("school_id"); schoolIDStr != "" {
		id, err := strconv.ParseUint(schoolIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid school_id parameter",
			})
			return
		}
		schoolIDUint := uint(id)
		schoolID = &schoolIDUint
	}
	
	// Parse optional search parameter
	search := c.Query("search")
	
	// Call service
	response, err := h.service.GetOrdersByDate(c.Request.Context(), date, schoolID, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch orders",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"orders":      response.Orders,
			"total_count": response.Summary.TotalOrders,
			"summary":     response.Summary,
		},
	})
}

// GetOrderDetails retrieves detailed information for a specific order
// GET /api/activity-tracker/orders/:id
func (h *ActivityTrackerHandler) GetOrderDetails(c *gin.Context) {
	// Parse order ID
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid order ID",
		})
		return
	}
	
	// Call service
	response, err := h.service.GetOrderDetails(c.Request.Context(), uint(orderID))
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "order not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch order details",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateOrderStatusRequest represents the request body for updating order status
type UpdateOrderStatusRequest struct {
	NewStatus string `json:"new_status" binding:"required"`
	Stage     int    `json:"stage" binding:"required,min=1,max=16"`
	Notes     string `json:"notes"`
}

// UpdateOrderStatus manually updates order status
// PUT /api/activity-tracker/orders/:id/status
func (h *ActivityTrackerHandler) UpdateOrderStatus(c *gin.Context) {
	// Parse order ID
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid order ID",
		})
		return
	}
	
	// Parse request body
	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request body",
		})
		return
	}
	
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "user not authenticated",
		})
		return
	}
	
	// Call service
	err = h.service.UpdateOrderStatus(c.Request.Context(), uint(orderID), req.NewStatus, req.Stage, userID.(uint), req.Notes)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "order not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to update order status",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "status updated successfully",
		"data": gin.H{
			"id":             uint(orderID),
			"current_status": req.NewStatus,
			"current_stage":  req.Stage,
			"updated_at":     time.Now(),
		},
	})
}

// AttachStageMedia attaches photo or video to a specific stage
// POST /api/activity-tracker/orders/:id/stages/:stage/media
func (h *ActivityTrackerHandler) AttachStageMedia(c *gin.Context) {
	// Parse order ID
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid order ID",
		})
		return
	}
	
	// Parse stage number
	stageStr := c.Param("stage")
	stage, err := strconv.Atoi(stageStr)
	if err != nil || stage < 1 || stage > 16 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid stage number (must be 1-16)",
		})
		return
	}
	
	// Get uploaded file
	file, err := c.FormFile("media")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "media file is required",
		})
		return
	}
	
	// Check file size limits (10MB for photos, 50MB for videos)
	const maxPhotoSize = 10 * 1024 * 1024  // 10MB
	const maxVideoSize = 50 * 1024 * 1024  // 50MB
	
	// Get media type
	mediaType := c.PostForm("media_type")
	if mediaType != "photo" && mediaType != "video" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "media_type must be 'photo' or 'video'",
		})
		return
	}
	
	// Validate file size based on media type
	if mediaType == "photo" && file.Size > maxPhotoSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"success": false,
			"error":   "photo file size exceeds 10MB limit",
		})
		return
	}
	
	if mediaType == "video" && file.Size > maxVideoSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"success": false,
			"error":   "video file size exceeds 50MB limit",
		})
		return
	}
	
	// TODO: Upload file to cloud storage (Firebase Storage or S3)
	// For now, we'll just return a mock URL
	mediaURL := "https://storage.example.com/" + file.Filename
	thumbnailURL := mediaURL // TODO: Generate actual thumbnail
	
	// Call service to attach media
	err = h.service.AttachStageMedia(c.Request.Context(), uint(orderID), stage, mediaURL, mediaType)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "order not found",
			})
			return
		}
		if err.Error() == "stage not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "stage not found or not yet reached",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to attach media",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "media attached successfully",
		"data": gin.H{
			"media_url":      mediaURL,
			"thumbnail_url":  thumbnailURL,
			"media_type":     mediaType,
		},
	})
}
