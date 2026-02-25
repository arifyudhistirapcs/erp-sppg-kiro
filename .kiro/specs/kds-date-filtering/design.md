# Design Document: KDS Date Filtering

## Overview

This feature adds date filtering capability to the Kitchen Display System (KDS) for both cooking and packing displays. Currently, the KDS only shows today's data without the ability to view historical records, and there are data inconsistencies between weekly planning and what's displayed in the KDS.

The solution introduces an optional date query parameter to existing KDS endpoints, allowing users to view historical data while maintaining backward compatibility. The system will default to today's date when no parameter is provided, ensuring existing integrations continue to work without modification.

### Key Design Goals

- Maintain backward compatibility with existing API consumers
- Ensure data consistency with the menu_items table
- Provide intuitive date selection UI for kitchen staff
- Handle timezone conversions consistently across the system
- Return empty results (not errors) for dates with no data

### Technology Stack

- **Backend**: Go (Gin framework), GORM for database access
- **Frontend**: Vue 3 with Ant Design Vue components
- **Database**: PostgreSQL with timezone-aware date handling
- **Real-time**: Firebase Realtime Database for status updates

## Architecture

### System Components

The date filtering feature integrates into the existing KDS architecture with minimal changes:

```
┌─────────────────┐
│   Vue Frontend  │
│  (Date Picker)  │
└────────┬────────┘
         │ HTTP GET with ?date=YYYY-MM-DD
         ▼
┌─────────────────┐
│  KDS Handler    │
│  (Parse & Val)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  KDS Service    │
│  (Query Logic)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   PostgreSQL    │
│  (menu_items)   │
└─────────────────┘
```

### Data Flow

1. **User Interaction**: User selects a date from the date picker component
2. **API Request**: Frontend sends GET request with date query parameter
3. **Validation**: Handler validates date format (YYYY-MM-DD)
4. **Service Layer**: Service queries database with the specified date
5. **Response**: Handler returns data or empty array for dates with no data
6. **UI Update**: Frontend displays the filtered data

### Backward Compatibility Strategy

To maintain backward compatibility:
- Date parameter is optional in all endpoints
- When omitted, system defaults to current date
- Existing API consumers continue to work without changes
- Response format remains unchanged

## Components and Interfaces

### Backend Components

#### 1. KDS Handler Modifications

**File**: `backend/internal/handlers/kds_handler.go`

**New Functions**:
```go
// parseDateParameter extracts and validates date from query parameter
func parseDateParameter(c *gin.Context) (time.Time, error)
```

**Modified Functions**:
- `GetCookingToday(c *gin.Context)` - Add date parameter parsing
- `GetPackingToday(c *gin.Context)` - Add date parameter parsing

**Validation Logic**:
- Accept date in YYYY-MM-DD format
- Return 400 Bad Request for invalid formats
- Default to current date if parameter is missing
- Accept future dates (return planned data or empty)

#### 2. KDS Service Modifications

**File**: `backend/internal/services/kds_service.go`

**Modified Functions**:
```go
// GetTodayMenu retrieves menu for specified date
func (s *KDSService) GetTodayMenu(ctx context.Context, date time.Time) ([]RecipeStatus, error)

// SyncTodayMenuToFirebase syncs menu for specified date
func (s *KDSService) SyncTodayMenuToFirebase(ctx context.Context, date time.Time) error
```

**Query Changes**:
- Replace `time.Now()` with provided date parameter
- Use `DATE(menu_items.date) = DATE(?)` for date comparison
- Maintain timezone consistency using `time.Truncate(24 * time.Hour)`

#### 3. Packing Allocation Service Modifications

**File**: `backend/internal/services/packing_allocation_service.go`

**Modified Functions**:
```go
// GetPackingAllocations retrieves allocations for specified date
func (s *PackingAllocationService) GetPackingAllocations(ctx context.Context, date time.Time) ([]SchoolAllocation, error)

// CalculatePackingAllocations calculates allocations for specified date
func (s *PackingAllocationService) CalculatePackingAllocations(ctx context.Context, date time.Time) ([]SchoolAllocation, error)
```

**Query Changes**:
- Replace `time.Now()` with provided date parameter
- Calculate startOfDay and endOfDay based on provided date
- Query delivery_tasks table with date range filter

### Frontend Components

#### 1. Date Picker Component

**New Component**: `web/src/components/KDSDatePicker.vue`

**Props**:
```javascript
{
  modelValue: Date,        // Selected date
  loading: Boolean,        // Loading state
  disabled: Boolean        // Disabled state
}
```

**Events**:
```javascript
{
  'update:modelValue': Date,  // Date selection change
  'change': Date              // Date change confirmed
}
```

**Features**:
- Calendar picker for date selection
- "Today" quick action button
- Keyboard navigation support (arrow keys, Enter, Escape)
- Visual feedback during data loading
- Displays currently selected date prominently
- Persists selected date in session storage

#### 2. KDS Service Modifications

**File**: `web/src/services/kdsService.js`

**Modified Functions**:
```javascript
// Get cooking menu for specified date
export const getCookingToday = async (date = null) => {
  const params = date ? { date: formatDate(date) } : {}
  const response = await api.get('/kds/cooking/today', { params })
  return response.data
}

// Get packing allocations for specified date
export const getPackingToday = async (date = null) => {
  const params = date ? { date: formatDate(date) } : {}
  const response = await api.get('/kds/packing/today', { params })
  return response.data
}

// Format date to YYYY-MM-DD
const formatDate = (date) => {
  return date.toISOString().split('T')[0]
}
```

#### 3. View Modifications

**Files**: 
- `web/src/views/KDSCookingView.vue`
- `web/src/views/KDSPackingView.vue`

**Changes**:
- Add KDSDatePicker component to page header
- Add selectedDate reactive state
- Pass selectedDate to API calls
- Update Firebase listener path based on selected date
- Display selected date prominently in UI
- Handle empty data states gracefully

## Data Models

### Existing Models (No Changes Required)

The feature leverages existing database models without modifications:

#### MenuItem Model
```go
type MenuItem struct {
    ID         uint      `gorm:"primaryKey"`
    MenuPlanID uint      `gorm:"not null"`
    RecipeID   uint      `gorm:"not null"`
    Date       time.Time `gorm:"type:date;not null"`
    Portions   int       `gorm:"not null"`
    Recipe     Recipe    `gorm:"foreignKey:RecipeID"`
    MenuPlan   MenuPlan  `gorm:"foreignKey:MenuPlanID"`
}
```

#### DeliveryTask Model
```go
type DeliveryTask struct {
    ID         uint      `gorm:"primaryKey"`
    SchoolID   uint      `gorm:"not null"`
    TaskDate   time.Time `gorm:"type:timestamp;not null"`
    Status     string    `gorm:"type:varchar(50);not null"`
    School     School    `gorm:"foreignKey:SchoolID"`
    MenuItems  []MenuItem `gorm:"many2many:delivery_task_menu_items"`
}
```

### API Request/Response Formats

#### Request Format

**Query Parameter**:
```
GET /api/v1/kds/cooking/today?date=2024-01-15
GET /api/v1/kds/packing/today?date=2024-01-15
```

**Parameter Specification**:
- Name: `date`
- Type: String
- Format: YYYY-MM-DD (ISO 8601 date format)
- Required: No (defaults to current date)
- Example: `2024-01-15`

#### Response Format (Unchanged)

**Success Response**:
```json
{
  "success": true,
  "data": [...]
}
```

**Error Response**:
```json
{
  "success": false,
  "error_code": "INVALID_DATE_FORMAT",
  "message": "Invalid date format. Expected YYYY-MM-DD",
  "details": "parsing time \"2024-1-5\": month out of range"
}
```

### Timezone Handling

**Strategy**: Use consistent timezone handling across all components

- **Database**: Store dates in UTC
- **API Layer**: Accept dates in YYYY-MM-DD format (timezone-agnostic)
- **Service Layer**: Convert to local timezone (Asia/Jakarta) for date comparisons
- **Frontend**: Display dates in user's local timezone

**Implementation**:
```go
// Normalize date to start of day in local timezone
func normalizeDate(date time.Time) time.Time {
    loc, _ := time.LoadLocation("Asia/Jakarta")
    return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
}
```


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Valid date acceptance and filtering

*For any* valid date in YYYY-MM-DD format, when provided to the cooking or packing endpoints, the API should accept the request and return only data where the date field matches the requested date.

**Validates: Requirements 2.1, 2.2, 7.3**

### Property 2: Invalid date rejection

*For any* string that does not conform to YYYY-MM-DD format (including malformed dates, wrong separators, invalid month/day values), the API should return HTTP 400 status with error message "Invalid date format. Expected YYYY-MM-DD" without executing database queries.

**Validates: Requirements 2.3, 6.4, 6.5, 8.1, 8.5**

### Property 3: Cooking endpoint date filtering

*For any* valid date parameter, the cooking endpoint should return recipe data that matches menu_items records for that specific date from approved menu plans.

**Validates: Requirements 4.2, 7.3**

### Property 4: Packing endpoint date filtering

*For any* valid date parameter, the packing endpoint should return packing allocations that match delivery_tasks scheduled for that specific date.

**Validates: Requirements 4.3, 7.3**

### Property 5: Timezone consistency

*For any* date query, the system should apply the same timezone conversion (Asia/Jakarta) consistently across all database queries, ensuring that a date like "2024-01-15" always refers to the same 24-hour period regardless of which endpoint or service processes it.

**Validates: Requirements 4.5, 7.4**

### Property 6: Future date handling

*For any* future date (date > current date), the API should return HTTP 200 with either an empty array (if no planned data exists) or planned data (if menu_items exist for that date), never returning an error status.

**Validates: Requirements 2.5**

### Property 7: Backward compatibility

*For any* request to cooking or packing endpoints without a date parameter, the API should return the same data as if the current date was explicitly provided, maintaining compatibility with existing API consumers.

**Validates: Requirements 6.3**

### Property 8: Empty data handling

*For any* date where no menu_items or delivery_tasks exist, the API should return HTTP 200 with an empty array, not an error status or null value.

**Validates: Requirements 4.4, 7.5**

## Error Handling

### Validation Errors

**Invalid Date Format**:
- **Trigger**: Date parameter doesn't match YYYY-MM-DD pattern
- **Response**: HTTP 400
- **Error Code**: `INVALID_DATE_FORMAT`
- **Message**: "Invalid date format. Expected YYYY-MM-DD"
- **Example**: `?date=2024-1-5` or `?date=01/15/2024`

**Unparseable Date**:
- **Trigger**: Date string matches format but represents invalid date
- **Response**: HTTP 400
- **Error Code**: `INVALID_DATE`
- **Message**: "Invalid date value"
- **Example**: `?date=2024-02-30` or `?date=2024-13-01`

### Database Errors

**Query Failure**:
- **Trigger**: Database connection issues or query execution errors
- **Response**: HTTP 500
- **Error Code**: `INTERNAL_ERROR`
- **Message**: "Gagal mengambil data" (Failed to retrieve data)
- **Action**: Log full error details, return generic message to client

### Frontend Error Handling

**Network Errors**:
- Display user-friendly message: "Gagal memuat data. Silakan coba lagi."
- Show retry button
- Maintain last successfully loaded data in UI

**Empty Data**:
- Display empty state with message: "Tidak ada data untuk tanggal ini"
- Show date picker to allow selecting different date
- Not treated as an error condition

**Loading States**:
- Show spinner during data fetch
- Disable date picker during loading
- Provide visual feedback that operation is in progress

### Error Recovery

**Automatic Retry**:
- Frontend retries failed requests once after 1 second delay
- If retry fails, show manual retry button

**Fallback Behavior**:
- If date parameter fails validation, frontend falls back to current date
- If Firebase listener disconnects, show warning but keep last known data

**Logging**:
- Backend logs all validation errors with request details
- Backend logs all database errors with full stack trace
- Frontend logs network errors to console for debugging

## Testing Strategy

### Dual Testing Approach

This feature will be validated using both unit tests and property-based tests:

- **Unit tests**: Verify specific examples, edge cases, and error conditions
- **Property tests**: Verify universal properties across all inputs
- Both approaches are complementary and necessary for comprehensive coverage

### Property-Based Testing

**Framework**: We will use [gopter](https://github.com/leanovate/gopter) for Go property-based testing.

**Configuration**:
- Each property test will run minimum 100 iterations
- Each test will be tagged with a comment referencing the design property
- Tag format: `// Feature: kds-date-filtering, Property {number}: {property_text}`

**Property Test Coverage**:

1. **Property 1 - Valid date acceptance**: Generate random valid dates, verify API accepts them and filters correctly
2. **Property 2 - Invalid date rejection**: Generate random invalid date strings, verify all return 400 errors
3. **Property 3 - Cooking endpoint filtering**: Generate random dates with test data, verify returned recipes match date
4. **Property 4 - Packing endpoint filtering**: Generate random dates with test data, verify returned allocations match date
5. **Property 5 - Timezone consistency**: Generate random dates, verify same date produces same results across multiple calls
6. **Property 6 - Future date handling**: Generate random future dates, verify returns 200 with empty or planned data
7. **Property 7 - Backward compatibility**: Test multiple requests without date parameter, verify returns current date data
8. **Property 8 - Empty data handling**: Generate random dates with no data, verify returns 200 with empty array

### Unit Testing

**Backend Unit Tests**:

- Date parameter parsing with valid formats
- Date parameter parsing with invalid formats
- Default date behavior when parameter omitted
- Timezone conversion correctness
- Database query construction with date filter
- Error response formatting
- Edge cases: leap years, month boundaries, year boundaries

**Frontend Unit Tests**:

- Date picker component rendering
- Date selection event handling
- "Today" button functionality
- Session storage persistence
- API call with date parameter
- Error message display
- Loading state management
- Empty data state display

**Integration Tests**:

- End-to-end flow: select date → API call → data display
- Firebase listener updates with date changes
- Multiple date selections in sequence
- Network error recovery
- Backward compatibility with existing code

### Test Data Strategy

**Test Database**:
- Seed database with menu_items for known dates
- Include dates with no data for empty result testing
- Include future dates with planned data
- Include past dates with historical data

**Date Ranges**:
- Current date (today)
- Past dates (last 30 days)
- Future dates (next 30 days)
- Edge cases: leap year dates, year boundaries, month boundaries
- Invalid dates: February 30, month 13, negative years

### Manual Testing Checklist

- [ ] Date picker displays correctly on both cooking and packing views
- [ ] Selecting a date updates the displayed data
- [ ] "Today" button returns to current date
- [ ] Selected date persists during session
- [ ] Empty dates show appropriate empty state
- [ ] Invalid dates show error message
- [ ] Loading spinner appears during data fetch
- [ ] Firebase real-time updates work with selected date
- [ ] Keyboard navigation works in date picker
- [ ] Mobile responsive design works correctly

