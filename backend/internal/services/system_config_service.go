package services

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrConfigNotFound    = errors.New("konfigurasi tidak ditemukan")
	ErrInvalidConfigType = errors.New("tipe konfigurasi tidak valid")
	ErrInvalidConfigValue = errors.New("nilai konfigurasi tidak valid")
)

// SystemConfigService handles system configuration operations
type SystemConfigService struct {
	db *gorm.DB
}

// NewSystemConfigService creates a new system configuration service
func NewSystemConfigService(db *gorm.DB) *SystemConfigService {
	return &SystemConfigService{db: db}
}

// GetConfig retrieves a configuration value by key
func (s *SystemConfigService) GetConfig(key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	result := s.db.Where("key = ?", key).First(&config)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrConfigNotFound
		}
		return nil, result.Error
	}
	return &config, nil
}

// GetConfigValue retrieves and parses a configuration value by key
func (s *SystemConfigService) GetConfigValue(key string) (interface{}, error) {
	config, err := s.GetConfig(key)
	if err != nil {
		return nil, err
	}

	return s.parseConfigValue(config)
}

// GetConfigString retrieves a string configuration value
func (s *SystemConfigService) GetConfigString(key string, defaultValue string) string {
	value, err := s.GetConfigValue(key)
	if err != nil {
		return defaultValue
	}
	if str, ok := value.(string); ok {
		return str
	}
	return defaultValue
}

// GetConfigInt retrieves an integer configuration value
func (s *SystemConfigService) GetConfigInt(key string, defaultValue int) int {
	value, err := s.GetConfigValue(key)
	if err != nil {
		return defaultValue
	}
	if intVal, ok := value.(int); ok {
		return intVal
	}
	return defaultValue
}

// GetConfigFloat retrieves a float configuration value
func (s *SystemConfigService) GetConfigFloat(key string, defaultValue float64) float64 {
	value, err := s.GetConfigValue(key)
	if err != nil {
		return defaultValue
	}
	if floatVal, ok := value.(float64); ok {
		return floatVal
	}
	return defaultValue
}

// GetConfigBool retrieves a boolean configuration value
func (s *SystemConfigService) GetConfigBool(key string, defaultValue bool) bool {
	value, err := s.GetConfigValue(key)
	if err != nil {
		return defaultValue
	}
	if boolVal, ok := value.(bool); ok {
		return boolVal
	}
	return defaultValue
}

// GetConfigJSON retrieves a JSON configuration value
func (s *SystemConfigService) GetConfigJSON(key string) (map[string]interface{}, error) {
	value, err := s.GetConfigValue(key)
	if err != nil {
		return nil, err
	}
	if jsonVal, ok := value.(map[string]interface{}); ok {
		return jsonVal, nil
	}
	return nil, ErrInvalidConfigType
}

// SetConfig creates or updates a configuration value
func (s *SystemConfigService) SetConfig(key, value, dataType, category string, updatedBy uint) error {
	// Validate data type
	if !isValidDataType(dataType) {
		return ErrInvalidConfigType
	}

	// Validate value matches data type
	if err := validateConfigValue(value, dataType); err != nil {
		return err
	}

	// Check if config exists
	var config models.SystemConfig
	result := s.db.Where("key = ?", key).First(&config)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create new config
			config = models.SystemConfig{
				Key:       key,
				Value:     value,
				DataType:  dataType,
				Category:  category,
				UpdatedBy: updatedBy,
			}
			return s.db.Create(&config).Error
		}
		return result.Error
	}

	// Update existing config
	config.Value = value
	config.DataType = dataType
	config.Category = category
	config.UpdatedBy = updatedBy
	return s.db.Save(&config).Error
}

// DeleteConfig deletes a configuration by key
func (s *SystemConfigService) DeleteConfig(key string) error {
	result := s.db.Where("key = ?", key).Delete(&models.SystemConfig{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConfigNotFound
	}
	return nil
}

// ListConfigs retrieves all configurations, optionally filtered by category
func (s *SystemConfigService) ListConfigs(category string) ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	query := s.db

	if category != "" {
		query = query.Where("category = ?", category)
	}

	result := query.Order("category, key").Find(&configs)
	return configs, result.Error
}

// parseConfigValue parses the stored string value based on data type
func (s *SystemConfigService) parseConfigValue(config *models.SystemConfig) (interface{}, error) {
	switch config.DataType {
	case "string":
		return config.Value, nil
	case "int":
		intVal, err := strconv.Atoi(config.Value)
		if err != nil {
			return nil, ErrInvalidConfigValue
		}
		return intVal, nil
	case "float":
		floatVal, err := strconv.ParseFloat(config.Value, 64)
		if err != nil {
			return nil, ErrInvalidConfigValue
		}
		return floatVal, nil
	case "bool":
		boolVal, err := strconv.ParseBool(config.Value)
		if err != nil {
			return nil, ErrInvalidConfigValue
		}
		return boolVal, nil
	case "json":
		var jsonVal map[string]interface{}
		err := json.Unmarshal([]byte(config.Value), &jsonVal)
		if err != nil {
			return nil, ErrInvalidConfigValue
		}
		return jsonVal, nil
	default:
		return nil, ErrInvalidConfigType
	}
}

// InitializeDefaultConfigs sets up default system configurations
func (s *SystemConfigService) InitializeDefaultConfigs() error {
	defaultConfigs := []struct {
		key      string
		value    string
		dataType string
		category string
	}{
		// Inventory settings
		{"inventory_min_stock_days", "7", "int", "inventory"},
		{"inventory_low_stock_percentage", "20", "int", "inventory"},
		{"inventory_stock_method", "FEFO", "string", "inventory"},
		{"inventory_auto_reorder", "false", "bool", "inventory"},
		
		// Nutrition standards
		{"nutrition_min_calories", "600", "int", "nutrition"},
		{"nutrition_min_protein", "15.0", "float", "nutrition"},
		{"nutrition_min_carbs", "80.0", "float", "nutrition"},
		{"nutrition_strict_validation", "true", "bool", "nutrition"},
		
		// Security settings
		{"security_session_timeout", "30", "int", "security"},
		{"security_max_login_attempts", "5", "int", "security"},
		{"security_lockout_duration", "15", "int", "security"},
		{"security_strong_password", "true", "bool", "security"},
		
		// System operations
		{"system_backup_schedule", "daily", "string", "system"},
		{"system_backup_retention", "30", "int", "system"},
		{"system_audit_retention", "365", "int", "system"},
		{"system_email_notifications", "true", "bool", "system"},
	}

	for _, config := range defaultConfigs {
		// Check if config already exists
		_, err := s.GetConfig(config.key)
		if err == nil {
			// Config already exists, skip
			continue
		}
		if err != ErrConfigNotFound {
			// Some other error occurred
			return err
		}

		// Create the config (using system user ID 1 as default)
		err = s.SetConfig(config.key, config.value, config.dataType, config.category, 1)
		if err != nil {
			return err
		}
	}

	return nil
}

// isValidDataType checks if the data type is supported
func isValidDataType(dataType string) bool {
	validTypes := []string{"string", "int", "float", "bool", "json"}
	for _, t := range validTypes {
		if dataType == t {
			return true
		}
	}
	return false
}

// validateConfigValue validates that the value matches the data type
func validateConfigValue(value, dataType string) error {
	switch dataType {
	case "string":
		return nil
	case "int":
		_, err := strconv.Atoi(value)
		if err != nil {
			return ErrInvalidConfigValue
		}
	case "float":
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrInvalidConfigValue
		}
	case "bool":
		_, err := strconv.ParseBool(value)
		if err != nil {
			return ErrInvalidConfigValue
		}
	case "json":
		var jsonVal map[string]interface{}
		err := json.Unmarshal([]byte(value), &jsonVal)
		if err != nil {
			return ErrInvalidConfigValue
		}
	default:
		return ErrInvalidConfigType
	}
	return nil
}
