# Design Document: Logistics Monitoring Process

## Overview

The Logistics Monitoring Process is a comprehensive tracking system that monitors the complete lifecycle of menu deliveries and ompreng (food container) management through 15 distinct stages. This system integrates with existing KDS Cooking and KDS Packing modules while introducing a new KDS Cleaning module for the cleaning workflow.

The system provides real-time visibility for all stakeholders including kitchen staff, packing staff, drivers, cleaning staff, and administrators. It tracks deliveries from the cooking stage through delivery to schools, ompreng collection, and finally the cleaning process at SPPG.

### Key Features

- Real-time status tracking across 15 lifecycle stages
- Integration with existing KDS Cooking and KDS Packing modules
- New KDS Cleaning module for ompreng cleaning workflow
- Timeline visualization showing delivery progress
- Activity log with timestamps for all status transitions
- Role-based access control with new "kebersihan" role
- Dashboard with summary statistics and filtering capabilities
- Firebase real-time synchronization for live updates

### Technology Stack

- **Backend**: Go with Gorm ORM, PostgreSQL database
- **Frontend**: Vue.js 3 with Ant Design Vue components
- **Real-time**: Firebase Realtime Database
- **Authentication**: JWT-based with role-based access control

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Monitoring  │  │ KDS Cleaning │  │  KDS Cooking │      │
│  │  Dashboard   │  │    Module    │  │  & Packing   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      API Layer (Go)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Monitoring  │  │   Cleaning   │  │     KDS      │      │
│  │   Handler    │  │   Handler    │  │   Handler    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                ┌───────────┴───────────┐
                ▼                       ▼
┌──────────────────────────┐  ┌──────────────────────┐
│   PostgreSQL Database    │  │  Firebase Realtime   │
│  - Delivery Records      │  │  - Live Status       │
│  - Status Transitions    │  │  - Real-time Sync    │
│  - Ompreng Tracking      │  │                      │
└──────────────────────────┘  └──────────────────────┘
```

### Data Flow

1. **Cooking Stage**: KDS Cooking module updates status → triggers monitoring system update
2. **Packing Stage**: KDS Packing module updates status → triggers monitoring system update
3. **Delivery Stage**: Driver app/interface updates location → monitoring system records transition
4. **Collection Stage**: Driver updates collection status → monitoring system tracks ompreng return
5. **Cleaning Stage**: KDS Cleaning module updates cleaning status → monitoring system completes lifecycle

All status updates are synchronized to Firebase for real-time display across all connected clients.

## Components and Interfaces

### Backend Components

#### 1. Database Models

**DeliveryRecord** (New)
```go
type DeliveryRecord struct {
    ID              uint      `gorm:"primaryKey" json:"id"`
    DeliveryDate    time.Time `gorm:"index;not null" json:"delivery_date"`
    SchoolID        uint      `gorm:"index;not null" json:"school_id"`
    DriverID        uint      `gorm:"index;not null" json:"driver_id"`
    MenuItemID      uint      `gorm:"index;not null" json:"menu_item_id"`
    Portions        int       `gorm:"not null" json:"portions"`
    CurrentStatus   string    `gorm:"size:50;not null;index" json:"current_status"`
    OmprengCount    int       `gorm:"not null" json:"ompreng_count"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    School          School    `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
    Driver          User      `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
    MenuItem        MenuItem  `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
}
```

**StatusTransition** (New)
```go
type StatusTransition struct {
    ID               uint           `gorm:"primaryKey" json:"id"`
    DeliveryRecordID uint           `gorm:"index;not null" json:"delivery_record_id"`
    FromStatus       string         `gorm:"size:50" json:"from_status"`
    ToStatus         string         `gorm:"size:50;not null" json:"to_status"`
    TransitionedAt   time.Time      `gorm:"index;not null" json:"transitioned_at"`
    TransitionedBy   uint           `gorm:"index;not null" json:"transitioned_by"`
    Notes            string         `gorm:"type:text" json:"notes"`
    DeliveryRecord   DeliveryRecord `gorm:"foreignKey:DeliveryRecordID" json:"delivery_record,omitempty"`
    User             User           `gorm:"foreignKey:TransitionedBy" json:"user,omitempty"`
}
```

**OmprengCleaning** (New)
```go
type OmprengCleaning struct {
    ID               uint           `gorm:"primaryKey" json:"id"`
    DeliveryRecordID uint           `gorm:"index;not null" json:"delivery_record_id"`
    OmprengCount     int            `gorm:"not null" json:"ompreng_count"`
    CleaningStatus   string         `gorm:"size:30;not null" json:"cleaning_status"`
    StartedAt        *time.Time     `json:"started_at"`
    CompletedAt      *time.Time     `json:"completed_at"`
    CleanedBy        *uint          `gorm:"index" json:"cleaned_by"`
    CreatedAt        time.Time      `json:"created_at"`
    UpdatedAt        time.Time      `json:"updated_at"`
    DeliveryRecord   DeliveryRecord `gorm:"foreignKey:DeliveryRecordID" json:"delivery_record,omitempty"`
    Cleaner          User           `gorm:"foreignKey:CleanedBy" json:"cleaner,omitempty"`
}
```

**User Model Update**
```go
// Add "kebersihan" to the role validation
Role string `gorm:"size:50;not null;index" json:"role" validate:"required,oneof=kepala_sppg kepala_yayasan akuntan ahli_gizi pengadaan chef packing driver asisten_lapangan kebersihan"`
```

#### 2. Service Layer

**MonitoringService** (New)
```go
type MonitoringService struct {
    db          *gorm.DB
    firebaseApp *firebase.App
    dbClient    *db.Client
}

// Core methods
func (s *MonitoringService) GetDeliveryRecords(ctx context.Context, date time.Time, filters map[string]interface{}) ([]DeliveryRecord, error)
func (s *MonitoringService) GetDeliveryRecordDetail(ctx context.Context, recordID uint) (*DeliveryRecordDetail, error)
func (s *MonitoringService) UpdateDeliveryStatus(ctx context.Context, recordID uint, newStatus string, userID uint, notes string) error
func (s *MonitoringService) GetActivityLog(ctx context.Context, recordID uint) ([]StatusTransition, error)
func (s *MonitoringService) GetDailySummary(ctx context.Context, date time.Time) (*DailySummary, error)
func (s *MonitoringService) ValidateStatusTransition(currentStatus, newStatus string) error
```

**CleaningService** (New)
```go
type CleaningService struct {
    db          *gorm.DB
    firebaseApp *firebase.App
    dbClient    *db.Client
}

// Core methods
func (s *CleaningService) GetPendingOmpreng(ctx context.Context) ([]OmprengCleaning, error)
func (s *CleaningService) StartCleaning(ctx context.Context, cleaningID uint, userID uint) error
func (s *CleaningService) CompleteCleaning(ctx context.Context, cleaningID uint, userID uint) error
func (s *CleaningService) SyncToFirebase(ctx context.Context, cleaning *OmprengCleaning) error
```

#### 3. API Handlers

**MonitoringHandler** (New)
```go
type MonitoringHandler struct {
    monitoringService *services.MonitoringService
}

// Endpoints
func (h *MonitoringHandler) GetDeliveryRecords(c *gin.Context)      // GET /api/monitoring/deliveries
func (h *MonitoringHandler) GetDeliveryDetail(c *gin.Context)       // GET /api/monitoring/deliveries/:id
func (h *MonitoringHandler) UpdateStatus(c *gin.Context)            // PUT /api/monitoring/deliveries/:id/status
func (h *MonitoringHandler) GetActivityLog(c *gin.Context)          // GET /api/monitoring/deliveries/:id/activity
func (h *MonitoringHandler) GetDailySummary(c *gin.Context)         // GET /api/monitoring/summary
```

**CleaningHandler** (New)
```go
type CleaningHandler struct {
    cleaningService *services.CleaningService
}

// Endpoints
func (h *CleaningHandler) GetPendingOmpreng(c *gin.Context)         // GET /api/cleaning/pending
func (h *CleaningHandler) StartCleaning(c *gin.Context)             // POST /api/cleaning/:id/start
func (h *CleaningHandler) CompleteCleaning(c *gin.Context)          // POST /api/cleaning/:id/complete
```

### Frontend Components

#### 1. Monitoring Dashboard View

**MonitoringDashboardView.vue**
- Date picker for selecting delivery date
- Summary statistics cards (total deliveries, completed, in-progress, etc.)
- Filterable list of delivery records
- Status indicators for each delivery
- Quick actions for status updates

#### 2. Delivery Detail View

**DeliveryDetailView.vue**
- School information display
- Driver information display
- Timeline visualization with 15 stages
- Activity log with timestamps
- Status update controls (role-based)

#### 3. KDS Cleaning View

**KDSCleaningView.vue**
- List of ompreng awaiting cleaning
- Ompreng details (school, delivery date, count)
- Start cleaning button
- Complete cleaning button
- Real-time status updates via Firebase

#### 4. Timeline Component

**DeliveryTimeline.vue**
- Visual representation of 15 lifecycle stages
- Color-coded status indicators (completed, in-progress, pending)
- Timestamps for completed stages
- Responsive design for mobile and desktop

#### 5. Activity Log Component

**ActivityLogTable.vue**
- Chronological list of status transitions
- User information for each transition
- Timestamp display in local timezone
- Elapsed time calculations between stages

## Data Models

### Status Lifecycle

The system tracks 15 distinct statuses in sequential order:

**Delivery Stages (1-8)**
1. `sedang_dimasak` - Being cooked
2. `selesai_dimasak` - Cooking completed
3. `siap_dipacking` - Ready for packing
4. `selesai_dipacking` - Packing completed
5. `siap_dikirim` - Ready for delivery
6. `diperjalanan` - In transit to school
7. `sudah_sampai_sekolah` - Arrived at school
8. `sudah_diterima_pihak_sekolah` - Received by school

**Collection Stages (9-13)**
9. `driver_ditugaskan_mengambil_ompreng` - Driver assigned for collection
10. `driver_menuju_sekolah` - Driver heading to school
11. `driver_sampai_di_sekolah` - Driver arrived at school
12. `ompreng_telah_diambil` - Ompreng collected
13. `ompreng_sampai_di_sppg` - Ompreng arrived at SPPG

**Cleaning Stages (14-15)**
14. `ompreng_proses_pencucian` - Ompreng being cleaned
15. `ompreng_selesai_dicuci` - Ompreng cleaning completed

### Status Transition Rules

```go
var statusTransitionRules = map[string][]string{
    "sedang_dimasak":                      {"selesai_dimasak"},
    "selesai_dimasak":                     {"siap_dipacking"},
    "siap_dipacking":                      {"selesai_dipacking"},
    "selesai_dipacking":                   {"siap_dikirim"},
    "siap_dikirim":                        {"diperjalanan"},
    "diperjalanan":                        {"sudah_sampai_sekolah"},
    "sudah_sampai_sekolah":                {"sudah_diterima_pihak_sekolah"},
    "sudah_diterima_pihak_sekolah":        {"driver_ditugaskan_mengambil_ompreng"},
    "driver_ditugaskan_mengambil_ompreng": {"driver_menuju_sekolah"},
    "driver_menuju_sekolah":               {"driver_sampai_di_sekolah"},
    "driver_sampai_di_sekolah":            {"ompreng_telah_diambil"},
    "ompreng_telah_diambil":               {"ompreng_sampai_di_sppg"},
    "ompreng_sampai_di_sppg":              {"ompreng_proses_pencucian"},
    "ompreng_proses_pencucian":            {"ompreng_selesai_dicuci"},
    "ompreng_selesai_dicuci":              {}, // Final state
}
```

### Database Schema

```sql
-- Delivery Records Table
CREATE TABLE delivery_records (
    id SERIAL PRIMARY KEY,
    delivery_date DATE NOT NULL,
    school_id INTEGER NOT NULL REFERENCES schools(id),
    driver_id INTEGER NOT NULL REFERENCES users(id),
    menu_item_id INTEGER NOT NULL REFERENCES menu_items(id),
    portions INTEGER NOT NULL,
    current_status VARCHAR(50) NOT NULL,
    ompreng_count INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_records_date ON delivery_records(delivery_date);
CREATE INDEX idx_delivery_records_school ON delivery_records(school_id);
CREATE INDEX idx_delivery_records_status ON delivery_records(current_status);

-- Status Transitions Table
CREATE TABLE status_transitions (
    id SERIAL PRIMARY KEY,
    delivery_record_id INTEGER NOT NULL REFERENCES delivery_records(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    transitioned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    transitioned_by INTEGER NOT NULL REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_status_transitions_record ON status_transitions(delivery_record_id);
CREATE INDEX idx_status_transitions_time ON status_transitions(transitioned_at);

-- Ompreng Cleaning Table
CREATE TABLE ompreng_cleanings (
    id SERIAL PRIMARY KEY,
    delivery_record_id INTEGER NOT NULL REFERENCES delivery_records(id),
    ompreng_count INTEGER NOT NULL,
    cleaning_status VARCHAR(30) NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    cleaned_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ompreng_cleanings_record ON ompreng_cleanings(delivery_record_id);
CREATE INDEX idx_ompreng_cleanings_status ON ompreng_cleanings(cleaning_status);
```

### Firebase Data Structure

```json
{
  "monitoring": {
    "deliveries": {
      "2024-01-15": {
        "record_123": {
          "id": 123,
          "school_name": "SD Negeri 1",
          "driver_name": "John Doe",
          "current_status": "diperjalanan",
          "portions": 150,
          "ompreng_count": 15,
          "last_updated": 1705308000000
        }
      }
    }
  },
  "cleaning": {
    "pending": {
      "cleaning_456": {
        "id": 456,
        "delivery_record_id": 123,
        "school_name": "SD Negeri 1",
        "ompreng_count": 15,
        "status": "pending",
        "arrived_at": 1705308000000
      }
    }
  }
}
```

## Integration Design

### KDS Cooking Integration

The monitoring system listens for status updates from the KDS Cooking module:

```go
// In KDS Service - after status update
func (s *KDSService) UpdateRecipeStatus(ctx context.Context, recipeID uint, status string, userID uint) error {
    // ... existing code ...
    
    // Trigger monitoring system update
    if status == "cooking" {
        monitoringService.UpdateDeliveryStatus(ctx, deliveryRecordID, "sedang_dimasak", userID, "")
    } else if status == "ready" {
        monitoringService.UpdateDeliveryStatus(ctx, deliveryRecordID, "selesai_dimasak", userID, "")
    }
    
    return nil
}
```

### KDS Packing Integration

Similar integration pattern for packing status updates:

```go
// In Packing Service - after packing status update
func (s *PackingService) UpdatePackingStatus(ctx context.Context, allocationID uint, status string, userID uint) error {
    // ... existing code ...
    
    // Trigger monitoring system update
    if status == "ready_for_packing" {
        monitoringService.UpdateDeliveryStatus(ctx, deliveryRecordID, "siap_dipacking", userID, "")
    } else if status == "packed" {
        monitoringService.UpdateDeliveryStatus(ctx, deliveryRecordID, "selesai_dipacking", userID, "")
    }
    
    return nil
}
```

### Driver App Integration

Driver status updates flow through the monitoring API:

```javascript
// Driver app - update location status
async function updateDriverLocation(deliveryRecordId, newStatus) {
    const response = await fetch(`/api/monitoring/deliveries/${deliveryRecordId}/status`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
            status: newStatus,
            notes: 'Driver location update'
        })
    });
    return response.json();
}
```

### Real-time Synchronization

All status updates are synchronized to Firebase for real-time display:

```go
func (s *MonitoringService) syncToFirebase(ctx context.Context, record *DeliveryRecord) error {
    dateStr := record.DeliveryDate.Format("2006-01-02")
    firebasePath := fmt.Sprintf("/monitoring/deliveries/%s/record_%d", dateStr, record.ID)
    
    data := map[string]interface{}{
        "id":             record.ID,
        "school_name":    record.School.Name,
        "driver_name":    record.Driver.FullName,
        "current_status": record.CurrentStatus,
        "portions":       record.Portions,
        "ompreng_count":  record.OmprengCount,
        "last_updated":   time.Now().Unix(),
    }
    
    return s.dbClient.NewRef(firebasePath).Set(ctx, data)
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property Reflection

After analyzing all acceptance criteria, I've identified several areas where properties can be consolidated:

**Redundancy Analysis:**
- Properties 2.1-2.8 (delivery status updates) can be combined into a single property about status transition recording
- Properties 3.1-3.5 (collection status updates) are similar to 2.x and can be combined
- Properties 5.4, 6.4, and 9.1 all test timestamp recording and can be consolidated
- Properties 8.2 and 8.3 test opposite sides of role-based access and can be combined
- Properties 10.1-10.3 test UI indicators and can be combined into one comprehensive property
- Properties 11.3 and 11.4 test contact information display and can be combined
- Properties 12.3-12.5 test filtering and can be combined into one property
- Properties 13.1-13.5 test driver location updates and can be combined
- Properties 14.3 and 14.4 test stage sequence enforcement and can be combined
- Properties 15.1-15.5 test various counts and can be combined into one property about count accuracy

**Consolidated Properties:**
After reflection, the 60+ acceptance criteria consolidate into approximately 25 unique, non-redundant properties that provide comprehensive validation coverage.

### Property 1: Date-based Delivery Record Retrieval

*For any* date and set of delivery records, querying for records on that date should return all and only the records with that delivery date.

**Validates: Requirements 1.1, 12.1**

### Property 2: Timeline Completeness

*For any* delivery record, the timeline view should contain all 15 lifecycle stages in sequential order.

**Validates: Requirements 1.2, 10.4**

### Property 3: Required Information Display

*For any* delivery record, the rendered output should include school name, portions, contact information, address, driver name, vehicle type, and driver contact information.

**Validates: Requirements 1.3, 1.4, 11.3, 11.4**

### Property 4: Activity Log Completeness

*For any* delivery record with status transitions, the activity log should contain all transitions with timestamps in chronological order.

**Validates: Requirements 1.5, 9.2**

### Property 5: Status Transition Recording

*For any* valid status transition, the system should update the current status and create a status transition record with timestamp and user attribution.

**Validates: Requirements 2.1-2.8, 3.1-3.5, 13.1-13.5**

### Property 6: Cleaning Status Recording

*For any* ompreng cleaning operation (start or complete), the system should update the cleaning status and record the timestamp and cleaning staff member.

**Validates: Requirements 4.1, 4.2, 4.3, 7.5**

### Property 7: Cleaning Record Association

*For any* ompreng cleaning record, it should be associated with exactly one delivery record, and that delivery record should be retrievable.

**Validates: Requirements 4.4**

### Property 8: KDS Cooking Integration

*For any* cooking status update in KDS Cooking (started or completed), the monitoring system should update the corresponding delivery record status and record the timestamp.

**Validates: Requirements 5.1, 5.2, 5.4**

### Property 9: Menu Item Data Retrieval

*For any* delivery record created from KDS Cooking, it should include menu item details such as quantity and school assignments.

**Validates: Requirements 5.3**

### Property 10: KDS Packing Integration

*For any* packing status update in KDS Packing (ready or packed), the monitoring system should update the corresponding delivery record status and record the timestamp.

**Validates: Requirements 6.1, 6.2, 6.4**

### Property 11: Packing Data Retrieval

*For any* delivery record updated from KDS Packing, it should include packing completion data.

**Validates: Requirements 6.3**

### Property 12: Cleaning Queue Filtering

*For any* set of ompreng cleaning records, the KDS Cleaning interface should display only those with status "ompreng_sampai_di_sppg".

**Validates: Requirements 7.1**

### Property 13: Cleaning Status Updates

*For any* ompreng cleaning record, cleaning staff should be able to update status from "ompreng_sampai_di_sppg" to "ompreng_proses_pencucian" and then to "ompreng_selesai_dicuci".

**Validates: Requirements 7.2, 7.3**

### Property 14: Cleaning Record Details

*For any* ompreng cleaning record displayed in KDS Cleaning, it should include associated school name and delivery date.

**Validates: Requirements 7.4**

### Property 15: Role-Based Access Control

*For any* user with role "kebersihan", they should have access to KDS Cleaning endpoints and be denied access to other KDS module endpoints.

**Validates: Requirements 8.2, 8.3**

### Property 16: Timestamp Recording Universality

*For any* status transition in the system, it should have an exact timestamp recorded.

**Validates: Requirements 9.1**

### Property 17: Elapsed Time Calculation

*For any* two consecutive status transitions, the system should correctly calculate the elapsed time between them.

**Validates: Requirements 9.3**

### Property 18: Timezone Display

*For any* timestamp displayed to users, it should be converted to the local timezone (Asia/Jakarta).

**Validates: Requirements 9.4**

### Property 19: Status Indicator Rendering

*For any* lifecycle stage in the timeline, it should display the appropriate indicator (completed, in-progress, or pending) based on the current status.

**Validates: Requirements 10.1, 10.2, 10.3**

### Property 20: Delivery Record Associations

*For any* delivery record, it should be associated with exactly one school and exactly one driver, and both should be retrievable.

**Validates: Requirements 11.1, 11.2**

### Property 21: Driver Reassignment

*For any* delivery record, when the driver is reassigned, the driver association should be updated to the new driver.

**Validates: Requirements 11.5**

### Property 22: School Distinction

*For any* set of delivery records on the same date, records for different schools should be distinguishable by school ID.

**Validates: Requirements 12.2**

### Property 23: Multi-criteria Filtering

*For any* set of delivery records, filtering by date, school, or status should return only records matching the filter criteria.

**Validates: Requirements 12.3, 12.4, 12.5**

### Property 24: Status Transition Validation

*For any* status transition request, if the transition is not allowed by the transition rules, the system should reject it and return an error.

**Validates: Requirements 14.1, 14.2**

### Property 25: Stage Sequence Enforcement

*For any* delivery record, delivery stages (1-8) must be completed before collection stages (9-13), and collection stages must be completed before cleaning stages (14-15).

**Validates: Requirements 14.3, 14.4**

### Property 26: Summary Statistics Accuracy

*For any* date, the summary statistics (total count, count by stage, completed count, cleaning counts) should match the actual count of delivery records meeting each criterion.

**Validates: Requirements 15.1, 15.2, 15.3, 15.4, 15.5**

## Error Handling

### Status Transition Errors

**Invalid Transition Error**
```go
type InvalidTransitionError struct {
    CurrentStatus string
    RequestedStatus string
    AllowedStatuses []string
}

func (e *InvalidTransitionError) Error() string {
    return fmt.Sprintf("cannot transition from %s to %s. Allowed transitions: %v",
        e.CurrentStatus, e.RequestedStatus, e.AllowedStatuses)
}
```

**Error Response Format**
```json
{
    "error": "invalid_transition",
    "message": "Cannot transition from diperjalanan to ompreng_proses_pencucian",
    "current_status": "diperjalanan",
    "requested_status": "ompreng_proses_pencucian",
    "allowed_statuses": ["sudah_sampai_sekolah"]
}
```

### Access Control Errors

**Unauthorized Access Error**
```go
type UnauthorizedAccessError struct {
    UserRole string
    RequiredRole string
    Resource string
}

func (e *UnauthorizedAccessError) Error() string {
    return fmt.Sprintf("user with role %s cannot access %s (requires %s)",
        e.UserRole, e.Resource, e.RequiredRole)
}
```

### Data Validation Errors

**Missing Required Field Error**
- Delivery record without school ID
- Delivery record without driver ID
- Status transition without user ID
- Cleaning record without ompreng count

**Invalid Data Error**
- Negative portions count
- Negative ompreng count
- Invalid status value
- Future delivery date

### Firebase Synchronization Errors

**Sync Failure Handling**
```go
func (s *MonitoringService) UpdateDeliveryStatus(ctx context.Context, recordID uint, newStatus string, userID uint, notes string) error {
    // Update database first
    err := s.updateDatabase(ctx, recordID, newStatus, userID, notes)
    if err != nil {
        return err
    }
    
    // Attempt Firebase sync (non-blocking)
    go func() {
        syncErr := s.syncToFirebase(ctx, record)
        if syncErr != nil {
            log.Printf("Firebase sync failed for record %d: %v", recordID, syncErr)
            // Queue for retry
            s.queueForRetry(recordID)
        }
    }()
    
    return nil
}
```

### Retry Strategy

**Exponential Backoff for Firebase Sync**
- Initial retry: 1 second
- Second retry: 2 seconds
- Third retry: 4 seconds
- Maximum retries: 5
- After max retries, log error and alert administrators

## Testing Strategy

### Dual Testing Approach

The testing strategy employs both unit tests and property-based tests to ensure comprehensive coverage:

**Unit Tests** focus on:
- Specific examples of status transitions
- Edge cases (e.g., empty delivery lists, missing data)
- Error conditions (e.g., invalid transitions, unauthorized access)
- Integration points between modules
- Firebase synchronization logic

**Property-Based Tests** focus on:
- Universal properties that hold for all inputs
- Status transition validation across all possible states
- Data integrity across random delivery records
- Filtering and querying with random criteria
- Timestamp and calculation correctness

### Property-Based Testing Configuration

**Testing Library**: Use `gopter` for Go property-based testing

**Test Configuration**:
- Minimum 100 iterations per property test
- Each test tagged with feature name and property number
- Tag format: `Feature: logistics-monitoring-process, Property {number}: {property_text}`

**Example Property Test Structure**:
```go
// Feature: logistics-monitoring-process, Property 1: Date-based Delivery Record Retrieval
func TestProperty_DateBasedRetrieval(t *testing.T) {
    parameters := gopter.DefaultTestParameters()
    parameters.MinSuccessfulTests = 100
    
    properties := gopter.NewProperties(parameters)
    
    properties.Property("querying by date returns only records for that date", 
        prop.ForAll(
            func(records []DeliveryRecord, queryDate time.Time) bool {
                // Create records in database
                // Query by date
                // Verify all returned records match date
                // Verify no records for other dates are returned
                return true
            },
            genDeliveryRecords(),
            genDate(),
        ))
    
    properties.TestingRun(t)
}
```

### Unit Test Coverage

**Status Transition Tests**:
- Test each valid transition in the 15-stage lifecycle
- Test rejection of invalid transitions
- Test administrator override capability
- Test timestamp recording for each transition

**Integration Tests**:
- Test KDS Cooking status updates trigger monitoring updates
- Test KDS Packing status updates trigger monitoring updates
- Test driver location updates flow through the system
- Test cleaning module updates propagate correctly

**Role-Based Access Tests**:
- Test kebersihan role can access cleaning endpoints
- Test kebersihan role cannot access cooking/packing endpoints
- Test other roles can access appropriate endpoints
- Test administrator role assignment

**Firebase Sync Tests**:
- Test successful synchronization
- Test retry logic on failure
- Test queue mechanism for failed syncs
- Test data consistency after sync

**Filtering Tests**:
- Test date filtering with various date ranges
- Test school filtering with multiple schools
- Test status filtering with different statuses
- Test combined filters

**Summary Statistics Tests**:
- Test count accuracy with various record sets
- Test stage-specific counts
- Test completed delivery counts
- Test cleaning status counts

### Integration Test Scenarios

**End-to-End Delivery Lifecycle**:
1. Create menu item in KDS Cooking
2. Update to cooking status
3. Complete cooking
4. Move to packing
5. Complete packing
6. Assign driver and mark ready for delivery
7. Update driver location through delivery
8. Confirm school receipt
9. Assign driver for collection
10. Track collection process
11. Mark ompreng arrival at SPPG
12. Start cleaning in KDS Cleaning
13. Complete cleaning
14. Verify all status transitions recorded
15. Verify timestamps present
16. Verify Firebase sync occurred

**Multi-School Delivery Day**:
1. Create delivery records for 10 different schools
2. Progress each through different stages
3. Verify filtering by school works
4. Verify filtering by status works
5. Verify summary statistics are accurate
6. Verify timeline displays correctly for each

**Error Recovery Scenarios**:
1. Attempt invalid status transition
2. Verify error response
3. Verify status unchanged
4. Attempt transition with unauthorized user
5. Verify access denied
6. Simulate Firebase connection failure
7. Verify retry mechanism activates
8. Verify eventual consistency

### Performance Testing

**Load Testing**:
- Test with 100+ delivery records per day
- Test with 50+ concurrent status updates
- Test Firebase sync performance with high volume
- Test query performance with large datasets

**Response Time Targets**:
- Delivery record list query: < 500ms
- Status update: < 200ms
- Activity log retrieval: < 300ms
- Summary statistics: < 400ms
- Firebase sync: < 1000ms (async)

## API Endpoints

### Monitoring Endpoints

**GET /api/monitoring/deliveries**
- Description: Get list of delivery records with optional filters
- Query Parameters:
  - `date` (required): Delivery date (YYYY-MM-DD)
  - `school_id` (optional): Filter by school
  - `status` (optional): Filter by current status
  - `driver_id` (optional): Filter by driver
- Response: Array of delivery record summaries
- Authentication: Required (all roles except kebersihan)

**GET /api/monitoring/deliveries/:id**
- Description: Get detailed information for a specific delivery record
- Path Parameters:
  - `id`: Delivery record ID
- Response: Detailed delivery record with school, driver, and menu item info
- Authentication: Required (all roles except kebersihan)

**PUT /api/monitoring/deliveries/:id/status**
- Description: Update delivery status
- Path Parameters:
  - `id`: Delivery record ID
- Request Body:
  ```json
  {
    "status": "diperjalanan",
    "notes": "Optional notes"
  }
  ```
- Response: Updated delivery record
- Authentication: Required (role-based: driver for delivery stages, admin for override)

**GET /api/monitoring/deliveries/:id/activity**
- Description: Get activity log (status transition history)
- Path Parameters:
  - `id`: Delivery record ID
- Response: Array of status transitions with timestamps and user info
- Authentication: Required (all roles except kebersihan)

**GET /api/monitoring/summary**
- Description: Get daily summary statistics
- Query Parameters:
  - `date` (required): Date for summary (YYYY-MM-DD)
- Response: Summary statistics object
- Authentication: Required (all roles except kebersihan)

### Cleaning Endpoints

**GET /api/cleaning/pending**
- Description: Get list of ompreng awaiting cleaning
- Response: Array of ompreng cleaning records with status "ompreng_sampai_di_sppg"
- Authentication: Required (kebersihan role)

**POST /api/cleaning/:id/start**
- Description: Start cleaning process for ompreng
- Path Parameters:
  - `id`: Ompreng cleaning record ID
- Response: Updated cleaning record
- Authentication: Required (kebersihan role)

**POST /api/cleaning/:id/complete**
- Description: Mark cleaning as completed
- Path Parameters:
  - `id`: Ompreng cleaning record ID
- Response: Updated cleaning record
- Authentication: Required (kebersihan role)

**GET /api/cleaning/history**
- Description: Get cleaning history
- Query Parameters:
  - `date` (optional): Filter by date
  - `school_id` (optional): Filter by school
- Response: Array of completed cleaning records
- Authentication: Required (kebersihan role, admin)

### Request/Response Examples

**Get Delivery Records**
```http
GET /api/monitoring/deliveries?date=2024-01-15&status=diperjalanan
Authorization: Bearer <token>

Response 200:
{
  "success": true,
  "data": [
    {
      "id": 123,
      "delivery_date": "2024-01-15",
      "school": {
        "id": 45,
        "name": "SD Negeri 1",
        "address": "Jl. Pendidikan No. 1"
      },
      "driver": {
        "id": 78,
        "full_name": "John Doe",
        "phone_number": "081234567890"
      },
      "current_status": "diperjalanan",
      "portions": 150,
      "ompreng_count": 15,
      "last_updated": "2024-01-15T10:30:00Z"
    }
  ]
}
```

**Update Delivery Status**
```http
PUT /api/monitoring/deliveries/123/status
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "sudah_sampai_sekolah",
  "notes": "Arrived at school gate"
}

Response 200:
{
  "success": true,
  "data": {
    "id": 123,
    "current_status": "sudah_sampai_sekolah",
    "updated_at": "2024-01-15T11:00:00Z"
  }
}

Response 400 (Invalid Transition):
{
  "success": false,
  "error": "invalid_transition",
  "message": "Cannot transition from diperjalanan to ompreng_proses_pencucian",
  "current_status": "diperjalanan",
  "allowed_statuses": ["sudah_sampai_sekolah"]
}
```

**Get Activity Log**
```http
GET /api/monitoring/deliveries/123/activity
Authorization: Bearer <token>

Response 200:
{
  "success": true,
  "data": [
    {
      "id": 1,
      "from_status": null,
      "to_status": "sedang_dimasak",
      "transitioned_at": "2024-01-15T08:00:00Z",
      "transitioned_by": {
        "id": 10,
        "full_name": "Chef Ahmad",
        "role": "chef"
      },
      "notes": ""
    },
    {
      "id": 2,
      "from_status": "sedang_dimasak",
      "to_status": "selesai_dimasak",
      "transitioned_at": "2024-01-15T09:30:00Z",
      "transitioned_by": {
        "id": 10,
        "full_name": "Chef Ahmad",
        "role": "chef"
      },
      "notes": "Cooking completed"
    }
  ]
}
```

**Get Daily Summary**
```http
GET /api/monitoring/summary?date=2024-01-15
Authorization: Bearer <token>

Response 200:
{
  "success": true,
  "data": {
    "date": "2024-01-15",
    "total_deliveries": 25,
    "by_status": {
      "sedang_dimasak": 2,
      "selesai_dimasak": 3,
      "siap_dipacking": 1,
      "selesai_dipacking": 2,
      "siap_dikirim": 1,
      "diperjalanan": 5,
      "sudah_sampai_sekolah": 3,
      "sudah_diterima_pihak_sekolah": 8
    },
    "completed_deliveries": 8,
    "ompreng_in_cleaning": 4,
    "ompreng_cleaned": 6
  }
}
```

**Get Pending Ompreng (Cleaning)**
```http
GET /api/cleaning/pending
Authorization: Bearer <token>

Response 200:
{
  "success": true,
  "data": [
    {
      "id": 456,
      "delivery_record_id": 123,
      "school": {
        "id": 45,
        "name": "SD Negeri 1"
      },
      "delivery_date": "2024-01-15",
      "ompreng_count": 15,
      "cleaning_status": "pending",
      "arrived_at": "2024-01-15T14:00:00Z"
    }
  ]
}
```

**Start Cleaning**
```http
POST /api/cleaning/456/start
Authorization: Bearer <token>

Response 200:
{
  "success": true,
  "data": {
    "id": 456,
    "cleaning_status": "in_progress",
    "started_at": "2024-01-15T15:00:00Z",
    "cleaned_by": {
      "id": 90,
      "full_name": "Cleaning Staff 1",
      "role": "kebersihan"
    }
  }
}
```

## Security Considerations

### Authentication and Authorization

**JWT Token Validation**
- All endpoints require valid JWT token
- Token expiration: 8 hours
- Refresh token mechanism for extended sessions

**Role-Based Access Control**
```go
var endpointPermissions = map[string][]string{
    "/api/monitoring/*":        {"kepala_sppg", "kepala_yayasan", "akuntan", "chef", "packing", "driver", "asisten_lapangan"},
    "/api/cleaning/*":          {"kebersihan", "kepala_sppg", "kepala_yayasan"},
    "/api/monitoring/*/status": {"driver", "chef", "packing", "kepala_sppg", "kepala_yayasan"}, // Context-dependent
}
```

**Status Update Authorization**
- Cooking statuses: chef role only
- Packing statuses: packing role only
- Delivery statuses: driver role only
- Collection statuses: driver role only
- Cleaning statuses: kebersihan role only
- Override capability: kepala_sppg, kepala_yayasan roles

### Data Validation

**Input Sanitization**
- Validate all status values against allowed list
- Sanitize notes field to prevent XSS
- Validate date formats
- Validate numeric fields (portions, ompreng_count) are positive

**SQL Injection Prevention**
- Use parameterized queries via Gorm
- Never concatenate user input into queries
- Validate all ID parameters are numeric

### Audit Trail

**Comprehensive Logging**
- Log all status transitions with user ID
- Log all failed authorization attempts
- Log all invalid transition attempts
- Log all administrator overrides

**Audit Trail Format**
```go
type AuditTrail struct {
    UserID    uint
    Timestamp time.Time
    Action    string // "status_update", "access_denied", "invalid_transition", "override"
    Entity    string // "delivery_record", "ompreng_cleaning"
    EntityID  string
    OldValue  string
    NewValue  string
    IPAddress string
}
```

## Deployment Considerations

### Database Migration

**Migration Script**
```sql
-- Create delivery_records table
CREATE TABLE delivery_records (
    id SERIAL PRIMARY KEY,
    delivery_date DATE NOT NULL,
    school_id INTEGER NOT NULL REFERENCES schools(id),
    driver_id INTEGER NOT NULL REFERENCES users(id),
    menu_item_id INTEGER NOT NULL REFERENCES menu_items(id),
    portions INTEGER NOT NULL CHECK (portions > 0),
    current_status VARCHAR(50) NOT NULL,
    ompreng_count INTEGER NOT NULL CHECK (ompreng_count >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_records_date ON delivery_records(delivery_date);
CREATE INDEX idx_delivery_records_school ON delivery_records(school_id);
CREATE INDEX idx_delivery_records_driver ON delivery_records(driver_id);
CREATE INDEX idx_delivery_records_status ON delivery_records(current_status);

-- Create status_transitions table
CREATE TABLE status_transitions (
    id SERIAL PRIMARY KEY,
    delivery_record_id INTEGER NOT NULL REFERENCES delivery_records(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    transitioned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    transitioned_by INTEGER NOT NULL REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_status_transitions_record ON status_transitions(delivery_record_id);
CREATE INDEX idx_status_transitions_time ON status_transitions(transitioned_at);
CREATE INDEX idx_status_transitions_user ON status_transitions(transitioned_by);

-- Create ompreng_cleanings table
CREATE TABLE ompreng_cleanings (
    id SERIAL PRIMARY KEY,
    delivery_record_id INTEGER NOT NULL REFERENCES delivery_records(id),
    ompreng_count INTEGER NOT NULL CHECK (ompreng_count > 0),
    cleaning_status VARCHAR(30) NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    cleaned_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ompreng_cleanings_record ON ompreng_cleanings(delivery_record_id);
CREATE INDEX idx_ompreng_cleanings_status ON ompreng_cleanings(cleaning_status);
CREATE INDEX idx_ompreng_cleanings_cleaner ON ompreng_cleanings(cleaned_by);

-- Update users table to add kebersihan role
-- This is a schema change to the role validation constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check 
    CHECK (role IN ('kepala_sppg', 'kepala_yayasan', 'akuntan', 'ahli_gizi', 'pengadaan', 'chef', 'packing', 'driver', 'asisten_lapangan', 'kebersihan'));
```

### Firebase Setup

**Firebase Realtime Database Rules**
```json
{
  "rules": {
    "monitoring": {
      "deliveries": {
        "$date": {
          ".read": "auth != null",
          ".write": "auth != null && (auth.token.role == 'driver' || auth.token.role == 'chef' || auth.token.role == 'packing' || auth.token.role == 'kepala_sppg' || auth.token.role == 'kepala_yayasan')"
        }
      }
    },
    "cleaning": {
      "pending": {
        ".read": "auth != null && (auth.token.role == 'kebersihan' || auth.token.role == 'kepala_sppg' || auth.token.role == 'kepala_yayasan')",
        ".write": "auth != null && (auth.token.role == 'kebersihan' || auth.token.role == 'kepala_sppg' || auth.token.role == 'kepala_yayasan')"
      }
    }
  }
}
```

### Environment Configuration

**Backend Environment Variables**
```env
# Existing variables
DATABASE_URL=postgresql://user:pass@localhost:5432/erp_sppg
FIREBASE_CREDENTIALS_PATH=./firebase-credentials.json
JWT_SECRET=your-secret-key

# New variables for monitoring
MONITORING_SYNC_RETRY_MAX=5
MONITORING_SYNC_RETRY_DELAY=1s
MONITORING_FIREBASE_PATH=/monitoring
CLEANING_FIREBASE_PATH=/cleaning
```

### Monitoring and Alerting

**Health Checks**
- Database connection health
- Firebase connection health
- API endpoint availability
- Status transition processing rate

**Alerts**
- Failed Firebase sync after max retries
- High rate of invalid transition attempts
- Unusual delay between status transitions
- Cleaning backlog exceeding threshold

**Metrics to Track**
- Average time per lifecycle stage
- Daily delivery completion rate
- Ompreng cleaning throughput
- API response times
- Error rates by endpoint

## Future Enhancements

### Phase 2 Enhancements

1. **GPS Tracking Integration**
   - Real-time driver location on map
   - Estimated arrival time calculations
   - Route optimization suggestions

2. **Push Notifications**
   - Notify schools when delivery is en route
   - Notify cleaning staff when ompreng arrive
   - Notify administrators of delays

3. **Analytics Dashboard**
   - Historical performance metrics
   - Bottleneck identification
   - Driver performance analytics
   - School delivery patterns

4. **Mobile App**
   - Dedicated driver mobile app
   - Offline capability with sync
   - Photo capture for proof of delivery
   - Digital signature collection

5. **Automated Status Updates**
   - Geofencing for automatic arrival detection
   - QR code scanning for status updates
   - RFID tracking for ompreng

6. **Predictive Analytics**
   - Predict delivery delays based on historical data
   - Optimize cleaning schedules
   - Forecast ompreng inventory needs

## Conclusion

This design provides a comprehensive solution for tracking the complete lifecycle of menu deliveries and ompreng management through 15 distinct stages. The system integrates seamlessly with existing KDS Cooking and KDS Packing modules while introducing a new KDS Cleaning module for the cleaning workflow.

Key design decisions include:
- Real-time synchronization via Firebase for live updates
- Strict status transition validation to maintain data integrity
- Role-based access control with new kebersihan role
- Comprehensive audit trail for accountability
- Dual testing approach with both unit and property-based tests

The implementation follows established patterns in the codebase, uses existing infrastructure (PostgreSQL, Firebase, Go/Gorm, Vue.js), and provides a solid foundation for future enhancements.
