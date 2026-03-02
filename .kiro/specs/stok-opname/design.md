# Design Document: Stok Opname

## Overview

The Stok Opname (Physical Inventory Count) feature enables warehouse staff to perform physical stock counts and adjust system records to match actual inventory levels. The feature supports batch processing of multiple inventory items within a single form, implements an approval workflow requiring Kepala SPPG authorization, and automatically applies stock adjustments upon approval.

### Key Design Goals

1. **Batch Processing**: Allow multiple inventory items to be counted and adjusted in a single operation
2. **Approval Workflow**: Implement a two-stage process (submission → approval) to ensure oversight
3. **Automatic Adjustment**: Apply stock changes atomically after approval to maintain data integrity
4. **Audit Trail**: Track all actions, changes, and approvals for compliance and troubleshooting
5. **Concurrent Safety**: Handle simultaneous operations without data corruption or race conditions

### Technology Stack

- **Backend**: Go with GORM ORM, SQLite database
- **Frontend**: Vue 3 with Ant Design Vue components
- **API**: RESTful endpoints with JSON payloads
- **Export**: Excel/PDF generation for reports

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Vue 3)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Inventory    │  │ Stok Opname  │  │ Stok Opname  │      │
│  │ View (Tabs)  │  │ List View    │  │ Form View    │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                  │                  │              │
│         └──────────────────┴──────────────────┘              │
│                            │                                 │
└────────────────────────────┼─────────────────────────────────┘
                             │ HTTP/JSON
┌────────────────────────────┼─────────────────────────────────┐
│                            │                                 │
│  ┌─────────────────────────▼──────────────────────────┐     │
│  │         StokOpnameHandler (HTTP Layer)             │     │
│  └─────────────────────────┬──────────────────────────┘     │
│                            │                                 │
│  ┌─────────────────────────▼──────────────────────────┐     │
│  │         StokOpnameService (Business Logic)         │     │
│  │  - Create/Edit/Delete Forms                        │     │
│  │  - Submit for Approval                             │     │
│  │  - Approve/Reject                                  │     │
│  │  - Apply Stock Adjustments                         │     │
│  └─────────────────────────┬──────────────────────────┘     │
│                            │                                 │
│  ┌─────────────────────────▼──────────────────────────┐     │
│  │         InventoryService (Stock Updates)           │     │
│  │  - UpdateStockWithTx                               │     │
│  │  - Create InventoryMovement                        │     │
│  └─────────────────────────┬──────────────────────────┘     │
│                            │                                 │
│  ┌─────────────────────────▼──────────────────────────┐     │
│  │              Database (SQLite + GORM)              │     │
│  │  - stok_opname_forms                               │     │
│  │  - stok_opname_items                               │     │
│  │  - inventory_items                                 │     │
│  │  - inventory_movements                             │     │
│  └────────────────────────────────────────────────────┘     │
│                      Backend (Go)                            │
└──────────────────────────────────────────────────────────────┘
```

### Data Flow

#### Creating and Submitting Stok Opname

```
User → Create Form → Add Items → Enter Physical Counts → Submit
  ↓                    ↓              ↓                    ↓
  DB                   DB             DB                   Notification
  (pending)         (items)        (counts)              (to Kepala SPPG)
```

#### Approval and Stock Adjustment

```
Kepala SPPG → Review Form → Approve
                              ↓
                    ┌─────────┴─────────┐
                    │  Transaction      │
                    │  - Update Status  │
                    │  - Adjust Stock   │
                    │  - Log Movements  │
                    │  - Set Processed  │
                    └───────────────────┘
```

## Components and Interfaces

### Backend Components

#### 1. Models (Data Structures)

**StokOpnameForm**
```go
type StokOpnameForm struct {
    ID              uint       `gorm:"primaryKey" json:"id"`
    FormNumber      string     `gorm:"uniqueIndex;size:50;not null" json:"form_number"`
    CreatedBy       uint       `gorm:"not null;index" json:"created_by"`
    CreatedAt       time.Time  `gorm:"index;not null" json:"created_at"`
    Status          string     `gorm:"size:20;not null;index" json:"status"` // pending, approved, rejected
    Notes           string     `gorm:"type:text" json:"notes"`
    ApprovedBy      *uint      `gorm:"index" json:"approved_by"`
    ApprovedAt      *time.Time `json:"approved_at"`
    RejectionReason string     `gorm:"type:text" json:"rejection_reason"`
    IsProcessed     bool       `gorm:"default:false;index" json:"is_processed"`
    UpdatedAt       time.Time  `json:"updated_at"`
    
    Creator         User                `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
    Approver        *User               `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
    Items           []StokOpnameItem    `gorm:"foreignKey:FormID" json:"items,omitempty"`
}
```

**StokOpnameItem**
```go
type StokOpnameItem struct {
    ID            uint      `gorm:"primaryKey" json:"id"`
    FormID        uint      `gorm:"index;not null" json:"form_id"`
    IngredientID  uint      `gorm:"index;not null" json:"ingredient_id"`
    SystemStock   float64   `gorm:"not null" json:"system_stock"`
    PhysicalCount float64   `gorm:"not null" json:"physical_count"`
    Difference    float64   `gorm:"not null" json:"difference"` // physical_count - system_stock
    ItemNotes     string    `gorm:"type:text" json:"item_notes"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    
    Form          StokOpnameForm `gorm:"foreignKey:FormID" json:"form,omitempty"`
    Ingredient    Ingredient     `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}
```

#### 2. Service Layer

**StokOpnameService Interface**
```go
type StokOpnameService interface {
    // Form Management
    CreateForm(userID uint, notes string) (*StokOpnameForm, error)
    GetForm(formID uint) (*StokOpnameForm, error)
    GetAllForms(filters FormFilters) ([]StokOpnameForm, int, error)
    UpdateFormNotes(formID uint, notes string) error
    DeleteForm(formID uint) error
    
    // Item Management
    AddItem(formID uint, ingredientID uint, physicalCount float64, notes string) error
    UpdateItem(itemID uint, physicalCount float64, notes string) error
    RemoveItem(itemID uint) error
    
    // Workflow
    SubmitForApproval(formID uint) error
    ApproveForm(formID uint, approverID uint) error
    RejectForm(formID uint, approverID uint, reason string) error
    
    // Reporting
    ExportForm(formID uint, format string) ([]byte, error)
}

type FormFilters struct {
    Status      string
    CreatedBy   *uint
    StartDate   *time.Time
    EndDate     *time.Time
    SearchText  string
    Page        int
    PageSize    int
}
```

#### 3. Handler Layer

**HTTP Endpoints**
```go
// Form endpoints
POST   /api/stok-opname/forms              // Create new form
GET    /api/stok-opname/forms              // List all forms (with filters)
GET    /api/stok-opname/forms/:id          // Get form details
PUT    /api/stok-opname/forms/:id/notes    // Update form notes
DELETE /api/stok-opname/forms/:id          // Delete pending form

// Item endpoints
POST   /api/stok-opname/forms/:id/items    // Add item to form
PUT    /api/stok-opname/items/:id          // Update item
DELETE /api/stok-opname/items/:id          // Remove item from form

// Workflow endpoints
POST   /api/stok-opname/forms/:id/submit   // Submit for approval
POST   /api/stok-opname/forms/:id/approve  // Approve form
POST   /api/stok-opname/forms/:id/reject   // Reject form

// Export endpoint
GET    /api/stok-opname/forms/:id/export   // Export form (query param: format=excel|pdf)
```

### Frontend Components

#### 1. InventoryView.vue (Modified)

Add new tab "Stok Opname" to existing tabs:
- Daftar Inventory
- Alert Stok Menipis
- Riwayat Pergerakan
- **Stok Opname** (NEW)

#### 2. StokOpnameList.vue (New Component)

Displays list of all stok opname forms with:
- Search and filter controls (status, date range, creator)
- Table showing: form number, date, creator, status, approver
- Action buttons: View, Edit (pending only), Delete (pending only), Export
- Create new form button

#### 3. StokOpnameForm.vue (New Component)

Form for creating/editing stok opname:
- Form header: form number, date, creator, status
- Notes field for form-level comments
- Item list with:
  - Ingredient selector (searchable dropdown)
  - System stock (read-only, auto-filled)
  - Physical count input
  - Difference calculation (auto-calculated)
  - Item notes field
  - Remove button
- Add item button
- Action buttons: Save Draft, Submit for Approval, Cancel
- Validation: at least one item, all physical counts filled

#### 4. StokOpnameDetail.vue (New Component)

Read-only view of stok opname form:
- All form and item details
- Approval information (if approved/rejected)
- Rejection reason (if rejected)
- Action buttons based on status and user role:
  - Pending + Creator: Edit, Delete
  - Pending + Kepala SPPG: Approve, Reject
  - Any status: Export

### Frontend Services

**stokOpnameService.js**
```javascript
export default {
  // Form operations
  createForm(data) {
    return api.post('/api/stok-opname/forms', data)
  },
  
  getForms(params) {
    return api.get('/api/stok-opname/forms', { params })
  },
  
  getForm(id) {
    return api.get(`/api/stok-opname/forms/${id}`)
  },
  
  updateFormNotes(id, notes) {
    return api.put(`/api/stok-opname/forms/${id}/notes`, { notes })
  },
  
  deleteForm(id) {
    return api.delete(`/api/stok-opname/forms/${id}`)
  },
  
  // Item operations
  addItem(formId, data) {
    return api.post(`/api/stok-opname/forms/${formId}/items`, data)
  },
  
  updateItem(itemId, data) {
    return api.put(`/api/stok-opname/items/${itemId}`, data)
  },
  
  removeItem(itemId) {
    return api.delete(`/api/stok-opname/items/${itemId}`)
  },
  
  // Workflow operations
  submitForApproval(formId) {
    return api.post(`/api/stok-opname/forms/${formId}/submit`)
  },
  
  approveForm(formId) {
    return api.post(`/api/stok-opname/forms/${formId}/approve`)
  },
  
  rejectForm(formId, reason) {
    return api.post(`/api/stok-opname/forms/${formId}/reject`, { reason })
  },
  
  // Export
  exportForm(formId, format) {
    return api.get(`/api/stok-opname/forms/${formId}/export`, {
      params: { format },
      responseType: 'blob'
    })
  }
}
```

## Data Models

### Database Schema

#### stok_opname_forms Table

| Column           | Type         | Constraints                    | Description                          |
|------------------|--------------|--------------------------------|--------------------------------------|
| id               | INTEGER      | PRIMARY KEY, AUTO_INCREMENT    | Unique identifier                    |
| form_number      | VARCHAR(50)  | UNIQUE, NOT NULL, INDEX        | Auto-generated form number           |
| created_by       | INTEGER      | NOT NULL, INDEX, FK(users.id)  | User who created the form            |
| created_at       | DATETIME     | NOT NULL, INDEX                | Form creation timestamp              |
| status           | VARCHAR(20)  | NOT NULL, INDEX                | pending/approved/rejected            |
| notes            | TEXT         |                                | Form-level notes                     |
| approved_by      | INTEGER      | INDEX, FK(users.id), NULLABLE  | User who approved/rejected           |
| approved_at      | DATETIME     | NULLABLE                       | Approval/rejection timestamp         |
| rejection_reason | TEXT         |                                | Reason for rejection                 |
| is_processed     | BOOLEAN      | DEFAULT FALSE, INDEX           | Flag to prevent duplicate processing |
| updated_at       | DATETIME     | NOT NULL                       | Last update timestamp                |

**Indexes:**
- `idx_form_number` (UNIQUE) on `form_number`
- `idx_created_by` on `created_by`
- `idx_created_at` on `created_at`
- `idx_status` on `status`
- `idx_approved_by` on `approved_by`
- `idx_is_processed` on `is_processed`

#### stok_opname_items Table

| Column         | Type         | Constraints                              | Description                     |
|----------------|--------------|------------------------------------------|---------------------------------|
| id             | INTEGER      | PRIMARY KEY, AUTO_INCREMENT              | Unique identifier               |
| form_id        | INTEGER      | NOT NULL, INDEX, FK(stok_opname_forms.id)| Parent form reference           |
| ingredient_id  | INTEGER      | NOT NULL, INDEX, FK(ingredients.id)      | Ingredient being counted        |
| system_stock   | REAL         | NOT NULL                                 | Stock level at time of creation |
| physical_count | REAL         | NOT NULL                                 | Actual counted quantity         |
| difference     | REAL         | NOT NULL                                 | physical_count - system_stock   |
| item_notes     | TEXT         |                                          | Item-specific notes             |
| created_at     | DATETIME     | NOT NULL                                 | Item creation timestamp         |
| updated_at     | DATETIME     | NOT NULL                                 | Last update timestamp           |

**Indexes:**
- `idx_form_id` on `form_id`
- `idx_ingredient_id` on `ingredient_id`
- Composite index on `(form_id, ingredient_id)` to prevent duplicates

**Constraints:**
- UNIQUE constraint on `(form_id, ingredient_id)` - one ingredient per form

### Relationships

```
users (1) ──────< (N) stok_opname_forms [created_by]
users (1) ──────< (N) stok_opname_forms [approved_by]

stok_opname_forms (1) ──────< (N) stok_opname_items

ingredients (1) ──────< (N) stok_opname_items

stok_opname_forms (1) ──────< (N) inventory_movements [reference]
```

### Form Number Generation

Format: `SO-YYYYMMDD-NNNN`
- `SO`: Stok Opname prefix
- `YYYYMMDD`: Date of creation
- `NNNN`: Sequential number for the day (padded to 4 digits)

Example: `SO-20240115-0001`

## Correctness Properties


*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property Reflection

After analyzing all acceptance criteria, I identified the following redundancies:

**Redundant Properties:**
- 13.1, 13.2 are duplicates of 2.2, 2.3 (creation audit trail)
- 13.3, 13.4 are duplicates of 6.5, 6.6 (approval audit trail)
- 7.5, 7.6, 7.7 can be combined into a single comprehensive property about movement type determination
- 8.2, 8.3, 8.4, 8.5 can be combined into a single property about list display completeness
- 9.2, 9.3, 9.4, 9.5, 9.6 can be combined into a single property about detail display completeness
- 4.1, 4.2, 4.3 can be combined into a single property about pending form editability
- 10.4 is the inverse of 4.4 (non-pending forms are immutable)

**Combined Properties:**
- Stock adjustment properties (7.1, 7.2, 7.3, 7.4) form a comprehensive approval workflow property
- Validation properties (5.2, 5.3, 5.4) form a comprehensive submission validation property

### Property 1: Form Creation Audit Trail

*For any* newly created stok opname form, the system should record the creation timestamp, creator user ID, and set the initial status to "pending".

**Validates: Requirements 2.2, 2.3, 2.4, 13.1, 13.2**

### Property 2: Item Addition with System Stock Capture

*For any* inventory item added to a stok opname form, the system should capture and store the current system stock at the time of addition.

**Validates: Requirements 3.1, 3.2**

### Property 3: Difference Calculation

*For any* stok opname item with a physical count entered, the difference should equal physical_count minus system_stock, and this value should be stored and displayed correctly (preserving sign for positive/negative differences).

**Validates: Requirements 3.4, 3.5**

### Property 4: Pending Form Mutability

*For any* stok opname form with status "pending", users should be able to modify physical counts, add or remove items, and update form notes.

**Validates: Requirements 4.1, 4.2, 4.3**

### Property 5: Approved/Rejected Form Immutability

*For any* stok opname form with status "approved" or "rejected", the system should prevent all editing and deletion operations.

**Validates: Requirements 4.4, 10.4**

### Property 6: Submission Validation

*For any* stok opname form submission attempt, the system should validate that: (1) at least one item exists in the form, and (2) all items have valid physical counts. If validation fails, the submission should be rejected with an appropriate error message.

**Validates: Requirements 5.2, 5.3, 5.4**

### Property 7: Approval Notification

*For any* successfully submitted stok opname form, the system should send a notification to users with the Kepala_SPPG role.

**Validates: Requirements 5.5**

### Property 8: Approval Status Transition

*For any* approval action by Kepala_SPPG, the form status should change to "approved", and for any rejection action, the status should change to "rejected".

**Validates: Requirements 6.2, 6.3**

### Property 9: Approval Audit Trail

*For any* approved or rejected stok opname form, the system should record the approver's user ID and the approval/rejection timestamp.

**Validates: Requirements 6.5, 6.6, 13.3, 13.4**

### Property 10: Stock Adjustment Application

*For any* approved stok opname form, the system should create stock adjustments for all items in the form, updating each ingredient's system stock to match its physical count, and create corresponding inventory movements with type "adjustment" that reference the form.

**Validates: Requirements 7.1, 7.2, 7.3, 7.4**

### Property 11: Movement Type Determination

*For any* stock adjustment, if the difference is positive (physical > system), the inventory movement should be recorded as "in" with quantity equal to the absolute difference; if negative (physical < system), it should be recorded as "out" with quantity equal to the absolute difference.

**Validates: Requirements 7.5, 7.6, 7.7**

### Property 12: List Display Completeness

*For any* stok opname form in the list view, the system should display: creation date, approval status, creator name, and (if approved/rejected) approver name.

**Validates: Requirements 8.2, 8.3, 8.4, 8.5**

### Property 13: List Sorting Order

*For any* retrieval of the stok opname form list, forms should be sorted by creation date in descending order (newest first).

**Validates: Requirements 8.6**

### Property 14: Detail Display Completeness

*For any* stok opname form detail view, the system should display: all items with their system stock, physical count, and difference; form notes (if present); item notes (if present); and rejection reason (if status is rejected).

**Validates: Requirements 9.2, 9.3, 9.4, 9.5, 9.6**

### Property 15: Cascade Deletion

*For any* pending stok opname form that is deleted, all associated items should also be deleted from the database.

**Validates: Requirements 10.3**

### Property 16: Duplicate Processing Prevention

*For any* stok opname form that has been marked as processed (is_processed = true), any subsequent attempt to apply stock adjustments should be rejected and logged as an error.

**Validates: Requirements 12.1, 12.2, 12.4**

### Property 17: Transaction Rollback on Failure

*For any* stock adjustment transaction that encounters an error, all changes within that transaction should be rolled back, leaving the database in its pre-transaction state.

**Validates: Requirements 14.2**

### Property 18: Concurrent Save Conflict Handling

*For any* save operation that encounters a database conflict (e.g., optimistic locking failure), the system should return an error response instructing the user to refresh their data.

**Validates: Requirements 14.4**

### Property 19: Export Content Completeness

*For any* exported stok opname report, the generated file should contain: form date, creator name, status, all items with their system stock/physical count/difference, export timestamp, and exporter name.

**Validates: Requirements 15.2, 15.3, 15.4**

## Error Handling

### Validation Errors

**Form Submission Validation**
- Empty form (no items): Return HTTP 400 with message "Form harus memiliki minimal satu item"
- Invalid physical counts (negative, null, non-numeric): Return HTTP 400 with message "Semua item harus memiliki physical count yang valid"
- Form not in pending status: Return HTTP 400 with message "Hanya form dengan status pending yang dapat diajukan"

**Authorization Errors**
- Non-Kepala SPPG attempting approval: Return HTTP 403 with message "Hanya Kepala SPPG yang dapat menyetujui stok opname"
- User attempting to edit non-pending form: Return HTTP 403 with message "Form yang sudah disetujui/ditolak tidak dapat diubah"
- User attempting to delete non-pending form: Return HTTP 403 with message "Form yang sudah disetujui/ditolak tidak dapat dihapus"

**Business Logic Errors**
- Duplicate processing attempt: Return HTTP 409 with message "Form ini sudah diproses sebelumnya"
- Ingredient not found: Return HTTP 404 with message "Ingredient tidak ditemukan"
- Form not found: Return HTTP 404 with message "Form stok opname tidak ditemukan"
- Duplicate ingredient in form: Return HTTP 409 with message "Ingredient sudah ada dalam form ini"

### Database Errors

**Transaction Failures**
- Deadlock detected: Retry transaction up to 3 times with exponential backoff
- Transaction timeout: Return HTTP 500 with message "Operasi timeout, silakan coba lagi"
- Constraint violation: Return HTTP 409 with appropriate message based on constraint

**Concurrent Access**
- Optimistic locking failure: Return HTTP 409 with message "Data telah diubah oleh pengguna lain. Silakan refresh halaman."
- Database connection failure: Return HTTP 503 with message "Layanan sementara tidak tersedia"

### Export Errors

**File Generation Failures**
- Template not found: Return HTTP 500 with message "Template export tidak ditemukan"
- Data too large: Return HTTP 413 with message "Data terlalu besar untuk diekspor"
- Invalid format requested: Return HTTP 400 with message "Format export tidak valid (gunakan excel atau pdf)"

### Error Logging

All errors should be logged with:
- Timestamp
- User ID (if authenticated)
- Request details (endpoint, method, parameters)
- Error type and message
- Stack trace (for server errors)

Critical errors (transaction failures, data corruption risks) should trigger alerts to system administrators.

## Testing Strategy

### Dual Testing Approach

The testing strategy employs both unit tests and property-based tests to ensure comprehensive coverage:

**Unit Tests** focus on:
- Specific examples of form creation, submission, and approval workflows
- Edge cases like empty forms, invalid inputs, and boundary conditions
- Integration points between services (StokOpnameService ↔ InventoryService)
- Error handling for specific scenarios (duplicate processing, authorization failures)
- UI component rendering and user interactions

**Property-Based Tests** focus on:
- Universal properties that hold across all inputs (e.g., "for any approved form, stock adjustments are applied")
- Comprehensive input coverage through randomization (random forms, items, counts)
- Invariants that must be maintained (e.g., "difference always equals physical_count - system_stock")
- Round-trip properties (e.g., "create form → retrieve form → data matches")

Together, unit tests catch concrete bugs in specific scenarios, while property tests verify general correctness across the input space.

### Property-Based Testing Configuration

**Framework**: Use `gopter` for Go backend property tests

**Configuration**:
- Minimum 100 iterations per property test
- Each test tagged with comment referencing design property
- Tag format: `// Feature: stok-opname, Property {number}: {property_text}`

**Example Property Test Structure**:
```go
// Feature: stok-opname, Property 3: Difference Calculation
func TestProperty_DifferenceCalculation(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("difference equals physical_count minus system_stock", 
        prop.ForAll(
            func(systemStock, physicalCount float64) bool {
                item := &StokOpnameItem{
                    SystemStock:   systemStock,
                    PhysicalCount: physicalCount,
                }
                item.Difference = physicalCount - systemStock
                
                return item.Difference == (physicalCount - systemStock)
            },
            gen.Float64Range(0, 10000),
            gen.Float64Range(0, 10000),
        ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Test Coverage Requirements

**Backend Service Tests**:
- StokOpnameService: All CRUD operations, workflow transitions, validation logic
- Stock adjustment application with transaction handling
- Concurrent access scenarios
- Authorization checks

**Backend Handler Tests**:
- HTTP endpoint responses (status codes, JSON structure)
- Request validation
- Error response formatting

**Frontend Component Tests**:
- StokOpnameList: Rendering, filtering, sorting
- StokOpnameForm: Item addition/removal, validation, submission
- StokOpnameDetail: Display of all form data, conditional action buttons

**Integration Tests**:
- End-to-end workflow: Create → Add Items → Submit → Approve → Verify Stock Adjusted
- Export functionality with actual file generation
- Notification delivery to Kepala SPPG

**Property Tests** (minimum 100 iterations each):
- Property 1: Form Creation Audit Trail
- Property 3: Difference Calculation
- Property 6: Submission Validation
- Property 10: Stock Adjustment Application
- Property 11: Movement Type Determination
- Property 13: List Sorting Order
- Property 16: Duplicate Processing Prevention
- Property 17: Transaction Rollback on Failure

### Manual Testing Scenarios

**Approval Workflow**:
1. Create form as staff, add items, submit
2. Login as Kepala SPPG, verify notification received
3. Approve form, verify stock updated in inventory
4. Verify form cannot be edited after approval

**Concurrent Operations**:
1. Two users open same form simultaneously
2. Both attempt to edit and save
3. Verify one succeeds, other receives conflict error

**Export Functionality**:
1. Create form with 50+ items
2. Export as Excel and PDF
3. Verify all data present in exported files
4. Verify export completes within 5 seconds

## Implementation Approach

### Phase 1: Database and Models (Backend)

1. Create database migration for `stok_opname_forms` and `stok_opname_items` tables
2. Define Go structs in `backend/internal/models/supply_chain.go`
3. Add models to `AllModels()` function for auto-migration
4. Test database schema creation

### Phase 2: Service Layer (Backend)

1. Create `backend/internal/services/stok_opname_service.go`
2. Implement form CRUD operations
3. Implement item management operations
4. Implement workflow methods (submit, approve, reject)
5. Implement stock adjustment logic with transaction handling
6. Add duplicate processing prevention
7. Write unit tests and property tests for service layer

### Phase 3: Handler Layer (Backend)

1. Create `backend/internal/handlers/stok_opname_handler.go`
2. Implement HTTP endpoints for all operations
3. Add request validation and error handling
4. Add authorization checks (role-based access)
5. Implement export endpoint with Excel/PDF generation
6. Write handler tests

### Phase 4: Router Integration (Backend)

1. Add routes to `backend/internal/router/router.go`
2. Apply authentication middleware
3. Apply authorization middleware for approval endpoints
4. Test all endpoints with Postman/curl

### Phase 5: Frontend Service (Web)

1. Create `web/src/services/stokOpnameService.js`
2. Implement API client methods for all endpoints
3. Add error handling and response transformation

### Phase 6: Frontend Components (Web)

1. Modify `web/src/views/InventoryView.vue` to add "Stok Opname" tab
2. Create `web/src/components/StokOpnameList.vue`
3. Create `web/src/components/StokOpnameForm.vue`
4. Create `web/src/components/StokOpnameDetail.vue`
5. Add routing for stok opname views
6. Implement search, filter, and export UI
7. Write component tests

### Phase 7: Integration and Testing

1. End-to-end testing of complete workflow
2. Performance testing with large forms (100+ items)
3. Concurrent access testing
4. Export functionality testing
5. Bug fixes and refinements

### Phase 8: Documentation and Deployment

1. Update API documentation
2. Create user manual for stok opname feature
3. Database migration for production
4. Deploy to staging environment
5. User acceptance testing
6. Deploy to production

## Security Considerations

### Authentication and Authorization

- All endpoints require authentication (JWT token)
- Approval/rejection endpoints restricted to Kepala_SPPG role
- Users can only edit/delete their own pending forms (unless admin)
- Form creator and approver must be different users

### Data Validation

- Server-side validation for all inputs
- Sanitize text inputs to prevent XSS
- Validate numeric inputs (physical counts must be non-negative)
- Prevent SQL injection through parameterized queries (GORM handles this)

### Audit Trail

- Log all form creations, submissions, approvals, rejections
- Log all stock adjustments with reference to source form
- Retain audit logs for minimum 2 years
- Include user ID and timestamp in all audit records

### Concurrent Access Protection

- Use database transactions for stock adjustments
- Implement optimistic locking for form updates
- Handle deadlocks with retry logic
- Return clear error messages for conflicts

## Performance Considerations

### Database Optimization

**Indexes**:
- `stok_opname_forms`: form_number (unique), created_by, created_at, status, approved_by, is_processed
- `stok_opname_items`: form_id, ingredient_id, composite (form_id, ingredient_id)

**Query Optimization**:
- Use eager loading for related data (creator, approver, items, ingredients)
- Paginate list views (default 20 items per page)
- Add database query timeout (5 seconds)

### Caching Strategy

- Cache ingredient list for dropdown (TTL: 1 hour)
- Cache user list for creator/approver display (TTL: 30 minutes)
- Invalidate form cache on create/update/delete

### Export Performance

- Generate exports asynchronously for large forms (>50 items)
- Use streaming for Excel generation to reduce memory usage
- Implement export queue if needed for high concurrency
- Set timeout of 30 seconds for export generation

### Frontend Optimization

- Lazy load stok opname tab content
- Debounce search input (300ms delay)
- Virtualize large item lists (>100 items)
- Show loading indicators for all async operations

## Monitoring and Observability

### Metrics to Track

- Form creation rate (per day/week/month)
- Average time from creation to approval
- Approval vs rejection rate
- Average number of items per form
- Stock adjustment volume (total quantity adjusted)
- Export request frequency
- API endpoint response times
- Database query performance

### Alerts

- Failed stock adjustment transactions
- Duplicate processing attempts
- Export generation failures
- Slow database queries (>1 second)
- High error rate on any endpoint (>5% of requests)

### Logging

- Log all workflow state transitions (pending → approved/rejected)
- Log all stock adjustments with before/after values
- Log all validation failures with details
- Log all authorization failures
- Use structured logging (JSON format) for easy parsing

## Future Enhancements

### Potential Improvements

1. **Batch Approval**: Allow Kepala SPPG to approve multiple forms at once
2. **Scheduled Stok Opname**: Automatically create forms on a schedule (e.g., monthly)
3. **Mobile App**: Dedicated mobile app for warehouse staff to perform counts
4. **Barcode Scanning**: Integrate barcode scanner for faster item identification
5. **Photo Attachment**: Allow attaching photos of physical inventory
6. **Variance Analysis**: Dashboard showing trends in stock discrepancies
7. **Approval Delegation**: Allow Kepala SPPG to delegate approval authority
8. **Multi-level Approval**: Require multiple approvals for large adjustments
9. **Integration with Purchase Orders**: Suggest PO creation for low stock items found during opname
10. **Real-time Collaboration**: Multiple users can work on same form simultaneously

