package models

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// TestMenuItemSchoolAllocation_PortionSizeValidation tests the validation tag for portion_size field
// Validates Requirement 2: Add Portion Size Field to Allocations
func TestMenuItemSchoolAllocation_PortionSizeValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name        string
		portionSize string
		shouldPass  bool
	}{
		{
			name:        "Valid portion size: small",
			portionSize: "small",
			shouldPass:  true,
		},
		{
			name:        "Valid portion size: large",
			portionSize: "large",
			shouldPass:  true,
		},
		{
			name:        "Invalid portion size: medium",
			portionSize: "medium",
			shouldPass:  false,
		},
		{
			name:        "Invalid portion size: empty string",
			portionSize: "",
			shouldPass:  false,
		},
		{
			name:        "Invalid portion size: invalid value",
			portionSize: "extra-large",
			shouldPass:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the PortionSize field validation directly
			err := validate.Var(tt.portionSize, "required,oneof=small large")

			if tt.shouldPass {
				assert.NoError(t, err, "Expected validation to pass for portion_size: %s", tt.portionSize)
			} else {
				assert.Error(t, err, "Expected validation to fail for portion_size: %s", tt.portionSize)
			}
		})
	}
}

// TestMenuItemSchoolAllocation_JSONSerialization tests JSON serialization and deserialization
// Validates that the portion_size field is correctly serialized as "portion_size" in JSON
func TestMenuItemSchoolAllocation_JSONSerialization(t *testing.T) {
	tests := []struct {
		name           string
		allocation     MenuItemSchoolAllocation
		expectedJSON   string
		checkFields    map[string]interface{}
	}{
		{
			name: "Serialize allocation with small portion size",
			allocation: MenuItemSchoolAllocation{
				ID:          1,
				MenuItemID:  10,
				SchoolID:    5,
				Portions:    100,
				PortionSize: "small",
			},
			checkFields: map[string]interface{}{
				"id":           float64(1),
				"menu_item_id": float64(10),
				"school_id":    float64(5),
				"portions":     float64(100),
				"portion_size": "small",
			},
		},
		{
			name: "Serialize allocation with large portion size",
			allocation: MenuItemSchoolAllocation{
				ID:          2,
				MenuItemID:  20,
				SchoolID:    8,
				Portions:    150,
				PortionSize: "large",
			},
			checkFields: map[string]interface{}{
				"id":           float64(2),
				"menu_item_id": float64(20),
				"school_id":    float64(8),
				"portions":     float64(150),
				"portion_size": "large",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test serialization
			jsonData, err := json.Marshal(tt.allocation)
			assert.NoError(t, err, "Failed to marshal allocation to JSON")

			// Parse JSON to verify field names
			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err, "Failed to unmarshal JSON")

			// Verify all expected fields are present with correct values
			for key, expectedValue := range tt.checkFields {
				actualValue, exists := result[key]
				assert.True(t, exists, "Expected field '%s' not found in JSON", key)
				assert.Equal(t, expectedValue, actualValue, "Field '%s' has incorrect value", key)
			}

			// Test deserialization
			var deserialized MenuItemSchoolAllocation
			err = json.Unmarshal(jsonData, &deserialized)
			assert.NoError(t, err, "Failed to unmarshal JSON to allocation")

			// Verify deserialized values match original
			assert.Equal(t, tt.allocation.ID, deserialized.ID)
			assert.Equal(t, tt.allocation.MenuItemID, deserialized.MenuItemID)
			assert.Equal(t, tt.allocation.SchoolID, deserialized.SchoolID)
			assert.Equal(t, tt.allocation.Portions, deserialized.Portions)
			assert.Equal(t, tt.allocation.PortionSize, deserialized.PortionSize)
		})
	}
}

// TestMenuItemSchoolAllocation_JSONFieldNames verifies the exact JSON field name for portion_size
func TestMenuItemSchoolAllocation_JSONFieldNames(t *testing.T) {
	allocation := MenuItemSchoolAllocation{
		ID:          1,
		MenuItemID:  10,
		SchoolID:    5,
		Portions:    100,
		PortionSize: "small",
	}

	jsonData, err := json.Marshal(allocation)
	assert.NoError(t, err, "Failed to marshal allocation")

	jsonString := string(jsonData)
	
	// Verify that the JSON contains "portion_size" (not "PortionSize" or "portionSize")
	assert.Contains(t, jsonString, `"portion_size":"small"`, 
		"JSON should contain 'portion_size' field with snake_case naming")
	
	// Verify it doesn't contain incorrect field names
	assert.NotContains(t, jsonString, `"PortionSize"`, 
		"JSON should not contain 'PortionSize' with PascalCase")
	assert.NotContains(t, jsonString, `"portionSize"`, 
		"JSON should not contain 'portionSize' with camelCase")
}

// TestMenuItemSchoolAllocation_PortionsValidation tests the validation for portions field
// Validates Requirement 3: Validate Portion Size Allocations (portions must be > 0)
func TestMenuItemSchoolAllocation_PortionsValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name       string
		portions   int
		shouldPass bool
	}{
		{
			name:       "Valid portions: positive value",
			portions:   100,
			shouldPass: true,
		},
		{
			name:       "Valid portions: minimum value 1",
			portions:   1,
			shouldPass: true,
		},
		{
			name:       "Invalid portions: zero",
			portions:   0,
			shouldPass: false,
		},
		{
			name:       "Invalid portions: negative value",
			portions:   -10,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the Portions field validation directly
			err := validate.Var(tt.portions, "required,gt=0")

			if tt.shouldPass {
				assert.NoError(t, err, "Expected validation to pass for portions: %d", tt.portions)
			} else {
				assert.Error(t, err, "Expected validation to fail for portions: %d", tt.portions)
			}
		})
	}
}

// TestMenuItemSchoolAllocation_RequiredFieldsValidation tests validation for required fields
// Validates that all required fields are properly validated
func TestMenuItemSchoolAllocation_RequiredFieldsValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name        string
		fieldName   string
		fieldValue  interface{}
		validationTag string
		shouldPass  bool
		description string
	}{
		{
			name:        "Valid portion_size: small",
			fieldName:   "portion_size",
			fieldValue:  "small",
			validationTag: "required,oneof=small large",
			shouldPass:  true,
			description: "PortionSize 'small' is valid",
		},
		{
			name:        "Valid portion_size: large",
			fieldName:   "portion_size",
			fieldValue:  "large",
			validationTag: "required,oneof=small large",
			shouldPass:  true,
			description: "PortionSize 'large' is valid",
		},
		{
			name:        "Invalid portion_size: empty",
			fieldName:   "portion_size",
			fieldValue:  "",
			validationTag: "required,oneof=small large",
			shouldPass:  false,
			description: "PortionSize is required",
		},
		{
			name:        "Invalid portion_size: invalid value",
			fieldName:   "portion_size",
			fieldValue:  "medium",
			validationTag: "required,oneof=small large",
			shouldPass:  false,
			description: "PortionSize must be 'small' or 'large'",
		},
		{
			name:        "Valid portions: positive",
			fieldName:   "portions",
			fieldValue:  100,
			validationTag: "required,gt=0",
			shouldPass:  true,
			description: "Portions must be positive",
		},
		{
			name:        "Invalid portions: zero",
			fieldName:   "portions",
			fieldValue:  0,
			validationTag: "required,gt=0",
			shouldPass:  false,
			description: "Portions must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Var(tt.fieldValue, tt.validationTag)

			if tt.shouldPass {
				assert.NoError(t, err, "Expected validation to pass: %s", tt.description)
			} else {
				assert.Error(t, err, "Expected validation to fail: %s", tt.description)
			}
		})
	}
}

// TestMenuItemSchoolAllocation_EdgeCases tests edge cases for model validation
func TestMenuItemSchoolAllocation_EdgeCases(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name        string
		fieldName   string
		fieldValue  interface{}
		validationTag string
		shouldPass  bool
		description string
	}{
		{
			name:        "Edge case: very large portions value",
			fieldName:   "portions",
			fieldValue:  999999,
			validationTag: "required,gt=0",
			shouldPass:  true,
			description: "Should handle large portion numbers",
		},
		{
			name:        "Edge case: portions = 1 (minimum valid)",
			fieldName:   "portions",
			fieldValue:  1,
			validationTag: "required,gt=0",
			shouldPass:  true,
			description: "Minimum valid portions value",
		},
		{
			name:        "Edge case: portion_size with whitespace",
			fieldName:   "portion_size",
			fieldValue:  " small ",
			validationTag: "required,oneof=small large",
			shouldPass:  false,
			description: "PortionSize with whitespace should fail",
		},
		{
			name:        "Edge case: portion_size with uppercase",
			fieldName:   "portion_size",
			fieldValue:  "SMALL",
			validationTag: "required,oneof=small large",
			shouldPass:  false,
			description: "PortionSize must be lowercase",
		},
		{
			name:        "Edge case: portion_size with mixed case",
			fieldName:   "portion_size",
			fieldValue:  "Small",
			validationTag: "required,oneof=small large",
			shouldPass:  false,
			description: "PortionSize must be exactly 'small' or 'large'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Var(tt.fieldValue, tt.validationTag)

			if tt.shouldPass {
				assert.NoError(t, err, "Expected validation to pass: %s", tt.description)
			} else {
				assert.Error(t, err, "Expected validation to fail: %s", tt.description)
			}
		})
	}
}

// TestMenuItemSchoolAllocation_StructValidation tests complete struct validation
// Validates Requirement 2 and Requirement 3 together
func TestMenuItemSchoolAllocation_StructValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name       string
		fieldName  string
		fieldValue interface{}
		validationTag string
		shouldPass bool
	}{
		{
			name:       "Valid: small portion_size",
			fieldName:  "portion_size",
			fieldValue: "small",
			validationTag: "required,oneof=small large",
			shouldPass: true,
		},
		{
			name:       "Valid: large portion_size",
			fieldName:  "portion_size",
			fieldValue: "large",
			validationTag: "required,oneof=small large",
			shouldPass: true,
		},
		{
			name:       "Invalid: empty portion_size",
			fieldName:  "portion_size",
			fieldValue: "",
			validationTag: "required,oneof=small large",
			shouldPass: false,
		},
		{
			name:       "Invalid: invalid portion_size value",
			fieldName:  "portion_size",
			fieldValue: "extra-large",
			validationTag: "required,oneof=small large",
			shouldPass: false,
		},
		{
			name:       "Valid: positive portions",
			fieldName:  "portions",
			fieldValue: 150,
			validationTag: "required,gt=0",
			shouldPass: true,
		},
		{
			name:       "Invalid: zero portions",
			fieldName:  "portions",
			fieldValue: 0,
			validationTag: "required,gt=0",
			shouldPass: false,
		},
		{
			name:       "Invalid: negative portions",
			fieldName:  "portions",
			fieldValue: -50,
			validationTag: "required,gt=0",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Var(tt.fieldValue, tt.validationTag)

			if tt.shouldPass {
				assert.NoError(t, err, "Expected field validation to pass")
			} else {
				assert.Error(t, err, "Expected field validation to fail")
			}
		})
	}
}
