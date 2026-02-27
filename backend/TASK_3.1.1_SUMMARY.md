# Task 3.1.1 Summary: Modify Request Payload to Accept portions_small and portions_large

## Changes Made

### 1. Updated SchoolAllocationInput Structure
**File:** `backend/internal/handlers/menu_planning_handler.go`

Changed from:
```go
type SchoolAllocationInput struct {
	SchoolID uint `json:"school_id" binding:"required"`
	Portions int  `json:"portions" binding:"required,gt=0"`
}
```

To:
```go
type SchoolAllocationInput struct {
	SchoolID       uint `json:"school_id" binding:"required"`
	PortionsSmall  int  `json:"portions_small" binding:"omitempty,gte=0"`
	PortionsLarge  int  `json:"portions_large" binding:"omitempty,gte=0"`
}
```

### 2. Updated CreateMenuItem Handler
**File:** `backend/internal/handlers/menu_planning_handler.go`

Modified the transformation logic to pass the new fields to the service layer:
```go
// Transform request to service input
var serviceAllocations []services.PortionSizeAllocationInput
for _, alloc := range req.SchoolAllocations {
	serviceAllocations = append(serviceAllocations, services.PortionSizeAllocationInput{
		SchoolID:       alloc.SchoolID,
		PortionsSmall:  alloc.PortionsSmall,
		PortionsLarge:  alloc.PortionsLarge,
	})
}
```

### 3. Updated UpdateMenuItem Handler
**File:** `backend/internal/handlers/menu_planning_handler.go`

Applied the same transformation logic to the update handler to maintain consistency.

### 4. Updated Handler Tests
**File:** `backend/internal/handlers/menu_planning_handler_test.go`

Updated all test cases to use the new field structure:
- `TestUpdateMenuItem_ValidUpdate`
- `TestUpdateMenuItem_InvalidSum`
- `TestUpdateMenuItem_NonExistentMenuItem`
- `TestGetMenuItem_Success`
- `TestGetMenuItem_WrongMenuPlan`

## API Request Format

### Before
```json
{
  "date": "2024-01-15",
  "recipe_id": 5,
  "portions": 500,
  "school_allocations": [
    {
      "school_id": 1,
      "portions": 350
    }
  ]
}
```

### After
```json
{
  "date": "2024-01-15",
  "recipe_id": 5,
  "portions": 500,
  "school_allocations": [
    {
      "school_id": 1,
      "portions_small": 150,
      "portions_large": 200
    },
    {
      "school_id": 2,
      "portions_small": 0,
      "portions_large": 150
    }
  ]
}
```

## Validation Rules

The handler now accepts:
- `portions_small`: Optional, must be >= 0 if provided
- `portions_large`: Optional, must be >= 0 if provided
- Both fields are passed to the service layer which performs business logic validation

## Testing

All handler tests pass successfully:
- ✓ TestUpdateMenuItem_ValidUpdate
- ✓ TestUpdateMenuItem_InvalidSum
- ✓ TestUpdateMenuItem_NonExistentMenuItem
- ✓ TestGetMenuItem_Success
- ✓ TestGetMenuItem_NotFound
- ✓ TestGetMenuItem_WrongMenuPlan

## Integration with Service Layer

The handler correctly transforms the request payload to `services.PortionSizeAllocationInput` which is already supported by the service layer (implemented in Phase 2).

## Next Steps

The following tasks in Phase 3 will build upon this change:
- Task 3.1.2: Update request validation to check portion size fields
- Task 3.1.3: Call ValidatePortionSizeAllocations before processing
- Task 3.1.4: Return detailed error messages for validation failures
- Task 3.1.5: Update API documentation with new request format
