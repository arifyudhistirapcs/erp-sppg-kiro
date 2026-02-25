package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParseDateParameter(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		queryParam  string
		expectError bool
		description string
	}{
		{
			name:        "Valid date format",
			queryParam:  "2024-01-15",
			expectError: false,
			description: "Should accept valid YYYY-MM-DD format",
		},
		{
			name:        "Missing date parameter",
			queryParam:  "",
			expectError: false,
			description: "Should default to current date when parameter is missing",
		},
		{
			name:        "Invalid format - wrong separator",
			queryParam:  "2024/01/15",
			expectError: true,
			description: "Should reject date with wrong separator",
		},
		{
			name:        "Invalid format - missing leading zeros",
			queryParam:  "2024-1-5",
			expectError: true,
			description: "Should reject date without leading zeros",
		},
		{
			name:        "Invalid format - wrong order",
			queryParam:  "01-15-2024",
			expectError: true,
			description: "Should reject date in MM-DD-YYYY format",
		},
		{
			name:        "Invalid date - February 30",
			queryParam:  "2024-02-30",
			expectError: true,
			description: "Should reject invalid date like February 30",
		},
		{
			name:        "Invalid date - month 13",
			queryParam:  "2024-13-01",
			expectError: true,
			description: "Should reject invalid month",
		},
		{
			name:        "Invalid date - day 32",
			queryParam:  "2024-01-32",
			expectError: true,
			description: "Should reject invalid day",
		},
		{
			name:        "Valid leap year date",
			queryParam:  "2024-02-29",
			expectError: false,
			description: "Should accept valid leap year date",
		},
		{
			name:        "Invalid non-leap year date",
			queryParam:  "2023-02-29",
			expectError: true,
			description: "Should reject February 29 in non-leap year",
		},
		{
			name:        "Future date",
			queryParam:  "2025-12-31",
			expectError: false,
			description: "Should accept future dates",
		},
		{
			name:        "Past date",
			queryParam:  "2020-01-01",
			expectError: false,
			description: "Should accept past dates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			// Create a test request with query parameter
			req := httptest.NewRequest(http.MethodGet, "/?date="+tt.queryParam, nil)
			c.Request = req

			// Call the function
			date, err := parseDateParameter(c)

			// Check error expectation
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				
				// For valid dates, verify the date is normalized to start of day
				if tt.queryParam != "" {
					assert.Equal(t, 0, date.Hour(), "Hour should be 0")
					assert.Equal(t, 0, date.Minute(), "Minute should be 0")
					assert.Equal(t, 0, date.Second(), "Second should be 0")
					assert.Equal(t, 0, date.Nanosecond(), "Nanosecond should be 0")
				}
			}
		})
	}
}

func TestParseDateParameterDefaultsToToday(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test context without date parameter
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request = req

	// Call the function
	date, err := parseDateParameter(c)

	// Should not error
	assert.NoError(t, err)

	// Should return today's date
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	expectedDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	assert.Equal(t, expectedDate.Year(), date.Year(), "Year should match current year")
	assert.Equal(t, expectedDate.Month(), date.Month(), "Month should match current month")
	assert.Equal(t, expectedDate.Day(), date.Day(), "Day should match current day")
	assert.Equal(t, 0, date.Hour(), "Hour should be 0")
	assert.Equal(t, 0, date.Minute(), "Minute should be 0")
	assert.Equal(t, 0, date.Second(), "Second should be 0")
}

func TestParseDateParameterTimezone(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test context with a valid date
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/?date=2024-01-15", nil)
	c.Request = req

	// Call the function
	date, err := parseDateParameter(c)

	// Should not error
	assert.NoError(t, err)

	// Verify timezone is Asia/Jakarta
	loc, _ := time.LoadLocation("Asia/Jakarta")
	assert.Equal(t, loc.String(), date.Location().String(), "Timezone should be Asia/Jakarta")
}
