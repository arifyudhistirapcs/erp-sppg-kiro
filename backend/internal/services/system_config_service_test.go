package services

import (
	"testing"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSystemConfigTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.SystemConfig{})
	assert.NoError(t, err)

	return db
}

func TestSystemConfigService_SetAndGetConfig(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	service := NewSystemConfigService(db)

	tests := []struct {
		name      string
		key       string
		value     string
		dataType  string
		category  string
		updatedBy uint
		wantErr   bool
	}{
		{
			name:      "Set string config",
			key:       "app_name",
			value:     "ERP SPPG",
			dataType:  "string",
			category:  "system",
			updatedBy: 1,
			wantErr:   false,
		},
		{
			name:      "Set int config",
			key:       "session_timeout",
			value:     "30",
			dataType:  "int",
			category:  "system",
			updatedBy: 1,
			wantErr:   false,
		},
		{
			name:      "Set float config",
			key:       "min_stock_threshold",
			value:     "10.5",
			dataType:  "float",
			category:  "inventory",
			updatedBy: 1,
			wantErr:   false,
		},
		{
			name:      "Set bool config",
			key:       "enable_notifications",
			value:     "true",
			dataType:  "bool",
			category:  "system",
			updatedBy: 1,
			wantErr:   false,
		},
		{
			name:      "Set json config",
			key:       "wifi_config",
			value:     `{"ssid":"Office","bssid":"00:11:22:33:44:55"}`,
			dataType:  "json",
			category:  "system",
			updatedBy: 1,
			wantErr:   false,
		},
		{
			name:      "Invalid data type",
			key:       "invalid",
			value:     "test",
			dataType:  "invalid_type",
			category:  "system",
			updatedBy: 1,
			wantErr:   true,
		},
		{
			name:      "Invalid int value",
			key:       "invalid_int",
			value:     "not_a_number",
			dataType:  "int",
			category:  "system",
			updatedBy: 1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SetConfig(tt.key, tt.value, tt.dataType, tt.category, tt.updatedBy)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify we can retrieve the config
			config, err := service.GetConfig(tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.key, config.Key)
			assert.Equal(t, tt.value, config.Value)
			assert.Equal(t, tt.dataType, config.DataType)
			assert.Equal(t, tt.category, config.Category)
		})
	}
}

func TestSystemConfigService_GetConfigValue(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	service := NewSystemConfigService(db)

	// Set up test configs
	service.SetConfig("string_val", "test", "string", "test", 1)
	service.SetConfig("int_val", "42", "int", "test", 1)
	service.SetConfig("float_val", "3.14", "float", "test", 1)
	service.SetConfig("bool_val", "true", "bool", "test", 1)

	tests := []struct {
		name      string
		key       string
		wantValue interface{}
		wantErr   bool
	}{
		{
			name:      "Get string value",
			key:       "string_val",
			wantValue: "test",
			wantErr:   false,
		},
		{
			name:      "Get int value",
			key:       "int_val",
			wantValue: 42,
			wantErr:   false,
		},
		{
			name:      "Get float value",
			key:       "float_val",
			wantValue: 3.14,
			wantErr:   false,
		},
		{
			name:      "Get bool value",
			key:       "bool_val",
			wantValue: true,
			wantErr:   false,
		},
		{
			name:      "Get non-existent config",
			key:       "non_existent",
			wantValue: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := service.GetConfigValue(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestSystemConfigService_TypedGetters(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	service := NewSystemConfigService(db)

	// Set up test configs
	service.SetConfig("string_val", "test", "string", "test", 1)
	service.SetConfig("int_val", "42", "int", "test", 1)
	service.SetConfig("float_val", "3.14", "float", "test", 1)
	service.SetConfig("bool_val", "true", "bool", "test", 1)

	// Test GetConfigString
	strVal := service.GetConfigString("string_val", "default")
	assert.Equal(t, "test", strVal)

	strDefault := service.GetConfigString("non_existent", "default")
	assert.Equal(t, "default", strDefault)

	// Test GetConfigInt
	intVal := service.GetConfigInt("int_val", 0)
	assert.Equal(t, 42, intVal)

	intDefault := service.GetConfigInt("non_existent", 99)
	assert.Equal(t, 99, intDefault)

	// Test GetConfigFloat
	floatVal := service.GetConfigFloat("float_val", 0.0)
	assert.Equal(t, 3.14, floatVal)

	floatDefault := service.GetConfigFloat("non_existent", 9.99)
	assert.Equal(t, 9.99, floatDefault)

	// Test GetConfigBool
	boolVal := service.GetConfigBool("bool_val", false)
	assert.Equal(t, true, boolVal)

	boolDefault := service.GetConfigBool("non_existent", false)
	assert.Equal(t, false, boolDefault)
}

func TestSystemConfigService_UpdateConfig(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	service := NewSystemConfigService(db)

	// Create initial config
	err := service.SetConfig("test_key", "initial", "string", "test", 1)
	assert.NoError(t, err)

	// Update the config
	err = service.SetConfig("test_key", "updated", "string", "test", 2)
	assert.NoError(t, err)

	// Verify update
	config, err := service.GetConfig("test_key")
	assert.NoError(t, err)
	assert.Equal(t, "updated", config.Value)
	assert.Equal(t, uint(2), config.UpdatedBy)
}

func TestSystemConfigService_ListConfigs(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	service := NewSystemConfigService(db)

	// Create multiple configs
	service.SetConfig("system_1", "val1", "string", "system", 1)
	service.SetConfig("system_2", "val2", "string", "system", 1)
	service.SetConfig("inventory_1", "val3", "string", "inventory", 1)

	// List all configs
	allConfigs, err := service.ListConfigs("")
	assert.NoError(t, err)
	assert.Len(t, allConfigs, 3)

	// List configs by category
	systemConfigs, err := service.ListConfigs("system")
	assert.NoError(t, err)
	assert.Len(t, systemConfigs, 2)

	inventoryConfigs, err := service.ListConfigs("inventory")
	assert.NoError(t, err)
	assert.Len(t, inventoryConfigs, 1)
}

func TestSystemConfigService_DeleteConfig(t *testing.T) {
	db := setupSystemConfigTestDB(t)
	service := NewSystemConfigService(db)

	// Create config
	err := service.SetConfig("test_key", "value", "string", "test", 1)
	assert.NoError(t, err)

	// Delete config
	err = service.DeleteConfig("test_key")
	assert.NoError(t, err)

	// Verify deletion
	_, err = service.GetConfig("test_key")
	assert.Error(t, err)
	assert.Equal(t, ErrConfigNotFound, err)

	// Try to delete non-existent config
	err = service.DeleteConfig("non_existent")
	assert.Error(t, err)
	assert.Equal(t, ErrConfigNotFound, err)
}
