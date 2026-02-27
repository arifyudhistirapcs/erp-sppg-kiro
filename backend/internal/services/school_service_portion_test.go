package services

import (
	"testing"
)

func TestDetermineSchoolPortionType(t *testing.T) {
	// Create a school service instance (db can be nil for this test)
	service := &SchoolService{}

	tests := []struct {
		name     string
		category string
		expected string
	}{
		{
			name:     "SD school should return mixed",
			category: "SD",
			expected: "mixed",
		},
		{
			name:     "SMP school should return large",
			category: "SMP",
			expected: "large",
		},
		{
			name:     "SMA school should return large",
			category: "SMA",
			expected: "large",
		},
		{
			name:     "Empty category should return large",
			category: "",
			expected: "large",
		},
		{
			name:     "Lowercase sd should return large (case sensitive)",
			category: "sd",
			expected: "large",
		},
		{
			name:     "Unknown category should return large",
			category: "UNKNOWN",
			expected: "large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.DetermineSchoolPortionType(tt.category)
			if result != tt.expected {
				t.Errorf("DetermineSchoolPortionType(%s) = %s; want %s", tt.category, result, tt.expected)
			}
		})
	}
}
