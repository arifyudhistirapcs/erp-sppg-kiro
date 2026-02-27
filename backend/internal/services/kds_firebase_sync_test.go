package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSchoolAllocationResponseStructure verifies that SchoolAllocationResponse
// includes all necessary portion size fields for Firebase sync
func TestSchoolAllocationResponseStructure(t *testing.T) {
	// Create a sample SchoolAllocationResponse
	allocation := SchoolAllocationResponse{
		SchoolID:        1,
		SchoolName:      "SD Negeri 1",
		SchoolCategory:  "SD",
		PortionSizeType: "mixed",
		PortionsSmall:   150,
		PortionsLarge:   200,
		TotalPortions:   350,
	}

	// Verify all fields are present
	assert.Equal(t, uint(1), allocation.SchoolID)
	assert.Equal(t, "SD Negeri 1", allocation.SchoolName)
	assert.Equal(t, "SD", allocation.SchoolCategory)
	assert.Equal(t, "mixed", allocation.PortionSizeType)
	assert.Equal(t, 150, allocation.PortionsSmall)
	assert.Equal(t, 200, allocation.PortionsLarge)
	assert.Equal(t, 350, allocation.TotalPortions)
}

// TestSchoolAllocationResponseForSMPSchool verifies that SMP schools
// have correct portion size type and zero small portions
func TestSchoolAllocationResponseForSMPSchool(t *testing.T) {
	allocation := SchoolAllocationResponse{
		SchoolID:        2,
		SchoolName:      "SMP Negeri 1",
		SchoolCategory:  "SMP",
		PortionSizeType: "large",
		PortionsSmall:   0,
		PortionsLarge:   300,
		TotalPortions:   300,
	}

	assert.Equal(t, "large", allocation.PortionSizeType)
	assert.Equal(t, 0, allocation.PortionsSmall)
	assert.Equal(t, 300, allocation.PortionsLarge)
	assert.Equal(t, 300, allocation.TotalPortions)
}
