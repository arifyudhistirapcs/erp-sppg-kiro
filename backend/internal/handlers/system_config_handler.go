package handlers

import (
	"net/http"

	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SystemConfigHandler handles system configuration HTTP requests
type SystemConfigHandler struct {
	systemConfigService *services.SystemConfigService
}

// NewSystemConfigHandler creates a new system configuration handler
func NewSystemConfigHandler(db *gorm.DB) *SystemConfigHandler {
	return &SystemConfigHandler{
		systemConfigService: services.NewSystemConfigService(db),
	}
}

// SetConfigRequest represents a request to set configuration
type SetConfigRequest struct {
	Key      string `json:"key" binding:"required"`
	Value    string `json:"value" binding:"required"`
	DataType string `json:"data_type" binding:"required,oneof=string int float bool json"`
	Category string `json:"category" binding:"required"`
}

// GetAllConfigs retrieves all system configurations
func (h *SystemConfigHandler) GetAllConfigs(c *gin.Context) {
	category := c.Query("category")
	
	configs, err := h.systemConfigService.ListConfigs(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengambil konfigurasi sistem",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": configs,
	})
}

// GetConfig retrieves a specific configuration by key
func (h *SystemConfigHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Key konfigurasi wajib diisi",
		})
		return
	}

	config, err := h.systemConfigService.GetConfig(key)
	if err != nil {
		if err == services.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Konfigurasi tidak ditemukan",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengambil konfigurasi",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": config,
	})
}

// SetConfig creates or updates a configuration
func (h *SystemConfigHandler) SetConfig(c *gin.Context) {
	var req SetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Data tidak valid",
			"message": err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User tidak terautentikasi",
		})
		return
	}

	updatedBy := userID.(uint)

	err := h.systemConfigService.SetConfig(req.Key, req.Value, req.DataType, req.Category, updatedBy)
	if err != nil {
		if err == services.ErrInvalidConfigType {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Tipe konfigurasi tidak valid",
			})
			return
		}
		if err == services.ErrInvalidConfigValue {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nilai konfigurasi tidak valid untuk tipe yang dipilih",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menyimpan konfigurasi",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Konfigurasi berhasil disimpan",
	})
}

// DeleteConfig deletes a configuration
func (h *SystemConfigHandler) DeleteConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Key konfigurasi wajib diisi",
		})
		return
	}

	err := h.systemConfigService.DeleteConfig(key)
	if err != nil {
		if err == services.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Konfigurasi tidak ditemukan",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus konfigurasi",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Konfigurasi berhasil dihapus",
	})
}

// GetConfigsByCategory retrieves configurations grouped by category
func (h *SystemConfigHandler) GetConfigsByCategory(c *gin.Context) {
	configs, err := h.systemConfigService.ListConfigs("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengambil konfigurasi sistem",
			"message": err.Error(),
		})
		return
	}

	// Group configs by category
	configsByCategory := make(map[string][]interface{})
	for _, config := range configs {
		if configsByCategory[config.Category] == nil {
			configsByCategory[config.Category] = []interface{}{}
		}
		configsByCategory[config.Category] = append(configsByCategory[config.Category], config)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": configsByCategory,
	})
}

// SetMultipleConfigs sets multiple configurations at once
func (h *SystemConfigHandler) SetMultipleConfigs(c *gin.Context) {
	var configs []SetConfigRequest
	if err := c.ShouldBindJSON(&configs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Data tidak valid",
			"message": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User tidak terautentikasi",
		})
		return
	}

	updatedBy := userID.(uint)

	// Set each configuration
	for _, config := range configs {
		err := h.systemConfigService.SetConfig(config.Key, config.Value, config.DataType, config.Category, updatedBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal menyimpan konfigurasi: " + config.Key,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Semua konfigurasi berhasil disimpan",
	})
}

// InitializeDefaultConfigs initializes default system configurations
func (h *SystemConfigHandler) InitializeDefaultConfigs(c *gin.Context) {
	err := h.systemConfigService.InitializeDefaultConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menginisialisasi konfigurasi default",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Konfigurasi default berhasil diinisialisasi",
	})
}