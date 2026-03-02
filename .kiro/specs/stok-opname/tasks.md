# Implementation Plan: Stok Opname

## Overview

This implementation plan breaks down the Stok Opname feature into 8 phases following the design document's implementation approach. The feature enables warehouse staff to perform physical inventory counts and adjust system records through an approval workflow. Implementation uses Go with GORM for the backend and Vue 3 with Ant Design Vue for the frontend.

## Tasks

- [x] 1. Database and Models Setup
  - [x] 1.1 Create database migration for stok_opname_forms table
    - Add migration file with table schema including all columns and indexes
    - Include unique constraint on form_number
    - Include composite unique constraint on (form_id, ingredient_id) in items table
    - _Requirements: 2.2, 2.3, 2.4, 6.5, 6.6, 12.3, 13.1, 13.2, 13.3, 13.4_
  
  - [x] 1.2 Create database migration for stok_opname_items table
    - Add migration file with table schema including all columns and indexes
    - Add foreign key constraints to stok_opname_forms and ingredients
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_
  
  - [x] 1.3 Define StokOpnameForm and StokOpnameItem models in Go
    - Create structs in backend/internal/models/supply_chain.go
    - Add GORM tags for all fields
    - Define relationships (Creator, Approver, Items, Ingredient)
    - _Requirements: 2.2, 2.3, 2.4, 3.1, 3.2, 3.3, 3.4, 6.5, 6.6, 12.3_
  
  - [x] 1.4 Add models to AllModels() function for auto-migration
    - Register new models in the auto-migration list
    - Test database schema creation
    - _Requirements: 2.2, 3.1_
  
  - [ ]* 1.5 Write unit tests for model validation
    - Test GORM tag constraints
    - Test relationship loading
    - _Requirements: 2.2, 3.1_

- [x] 2. Service Layer - Form Management
  - [x] 2.1 Create StokOpnameService interface and implementation
    - Create backend/internal/services/stok_opname_service.go
    - Define interface with all required methods
    - Implement struct with dependencies (DB, InventoryService)
    - _Requirements: 2.1, 2.2, 2.3, 2.4_
  
  - [x] 2.2 Implement CreateForm method with form number generation
    - Generate form number in format SO-YYYYMMDD-NNNN
    - Set initial status to "pending"
    - Record creator user ID and timestamp
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 13.1, 13.2_
  
  - [ ]* 2.3 Write property test for CreateForm
    - **Property 1: Form Creation Audit Trail**
    - **Validates: Requirements 2.2, 2.3, 2.4, 13.1, 13.2**
  
  - [x] 2.4 Implement GetForm and GetAllForms methods
    - GetForm: Load form with all relationships (creator, approver, items, ingredients)
    - GetAllForms: Support pagination, filtering by status/date/creator, search by text
    - Implement sorting by created_at DESC
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 9.1, 11.1, 11.2, 11.3, 11.4_
  
  - [ ]* 2.5 Write property test for list sorting
    - **Property 13: List Sorting Order**
    - **Validates: Requirements 8.6**
  
  - [x] 2.6 Implement UpdateFormNotes method
    - Validate form is in pending status
    - Update notes field
    - _Requirements: 2.5, 4.3_
  
  - [x] 2.7 Implement DeleteForm method
    - Validate form is in pending status
    - Delete form and cascade delete all items
    - _Requirements: 10.1, 10.2, 10.3, 10.4_
  
  - [ ]* 2.8 Write unit tests for form CRUD operations
    - Test form creation with valid data
    - Test form retrieval with relationships
    - Test form deletion with cascade
    - Test validation errors for non-pending forms
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 4.4, 10.1, 10.2, 10.3, 10.4_

- [x] 3. Service Layer - Item Management
  - [x] 3.1 Implement AddItem method
    - Validate form is in pending status
    - Fetch current system stock for ingredient
    - Calculate difference (physical_count - system_stock)
    - Prevent duplicate ingredient in same form
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 4.2_
  
  - [ ]* 3.2 Write property test for item addition
    - **Property 2: Item Addition with System Stock Capture**
    - **Validates: Requirements 3.1, 3.2**
  
  - [ ]* 3.3 Write property test for difference calculation
    - **Property 3: Difference Calculation**
    - **Validates: Requirements 3.4, 3.5**
  
  - [x] 3.4 Implement UpdateItem method
    - Validate parent form is in pending status
    - Update physical_count and item_notes
    - Recalculate difference
    - _Requirements: 3.3, 3.6, 4.1_
  
  - [x] 3.5 Implement RemoveItem method
    - Validate parent form is in pending status
    - Delete item from database
    - _Requirements: 4.2_
  
  - [ ]* 3.6 Write unit tests for item management
    - Test adding item with system stock capture
    - Test updating item recalculates difference
    - Test removing item from form
    - Test duplicate ingredient prevention
    - Test validation errors for non-pending forms
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 4.1, 4.2_

- [x] 4. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Service Layer - Workflow Operations
  - [x] 5.1 Implement SubmitForApproval method
    - Validate form has at least one item
    - Validate all items have valid physical counts
    - Keep status as "pending" (approval changes status)
    - Send notification to Kepala_SPPG users
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_
  
  - [ ]* 5.2 Write property test for submission validation
    - **Property 6: Submission Validation**
    - **Validates: Requirements 5.2, 5.3, 5.4**
  
  - [x] 5.3 Implement ApproveForm method with stock adjustment
    - Validate approver has Kepala_SPPG role
    - Validate approver is different from creator
    - Update status to "approved"
    - Record approver ID and timestamp
    - Begin database transaction
    - For each item: create stock adjustment, update system stock, create inventory movement
    - Set is_processed flag to true
    - Commit transaction or rollback on error
    - _Requirements: 6.1, 6.2, 6.5, 6.6, 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 12.1, 12.2, 12.3, 13.3, 13.4, 14.1, 14.2_
  
  - [ ]* 5.4 Write property test for stock adjustment application
    - **Property 10: Stock Adjustment Application**
    - **Validates: Requirements 7.1, 7.2, 7.3, 7.4**
  
  - [ ]* 5.5 Write property test for movement type determination
    - **Property 11: Movement Type Determination**
    - **Validates: Requirements 7.5, 7.6, 7.7**
  
  - [ ]* 5.6 Write property test for transaction rollback
    - **Property 17: Transaction Rollback on Failure**
    - **Validates: Requirements 14.2**
  
  - [x] 5.7 Implement RejectForm method
    - Validate approver has Kepala_SPPG role
    - Update status to "rejected"
    - Record approver ID, timestamp, and rejection reason
    - _Requirements: 6.1, 6.3, 6.4, 6.5, 6.6, 13.3, 13.4_
  
  - [ ]* 5.8 Write property test for approval audit trail
    - **Property 9: Approval Audit Trail**
    - **Validates: Requirements 6.5, 6.6, 13.3, 13.4**
  
  - [ ]* 5.9 Write unit tests for workflow operations
    - Test successful approval with stock adjustment
    - Test rejection with reason
    - Test authorization checks (Kepala_SPPG only)
    - Test creator cannot approve own form
    - Test notification sending on submission
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 6. Service Layer - Additional Features
  - [x] 6.1 Implement duplicate processing prevention
    - Check is_processed flag before applying adjustments
    - Log error if duplicate processing attempted
    - Return appropriate error
    - _Requirements: 12.1, 12.2, 12.3, 12.4_
  
  - [ ]* 6.2 Write property test for duplicate processing prevention
    - **Property 16: Duplicate Processing Prevention**
    - **Validates: Requirements 12.1, 12.2, 12.4**
  
  - [x] 6.3 Implement ExportForm method
    - Support Excel and PDF formats
    - Include all form details: date, creator, status, items, counts, differences
    - Include export timestamp and exporter name
    - Use streaming for large forms
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_
  
  - [ ]* 6.4 Write unit tests for export functionality
    - Test Excel generation with complete data
    - Test PDF generation with complete data
    - Test invalid format handling
    - _Requirements: 15.1, 15.2, 15.3, 15.4_

- [x] 7. Handler Layer - HTTP Endpoints
  - [x] 7.1 Create StokOpnameHandler struct and initialization
    - Create backend/internal/handlers/stok_opname_handler.go
    - Define handler struct with service dependency
    - _Requirements: 2.1_
  
  - [x] 7.2 Implement form management endpoints
    - POST /api/stok-opname/forms - Create new form
    - GET /api/stok-opname/forms - List forms with filters
    - GET /api/stok-opname/forms/:id - Get form details
    - PUT /api/stok-opname/forms/:id/notes - Update form notes
    - DELETE /api/stok-opname/forms/:id - Delete pending form
    - Add request validation and error handling
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 4.3, 8.1, 9.1, 10.1, 10.2, 11.1, 11.2, 11.3, 11.4_
  
  - [x] 7.3 Implement item management endpoints
    - POST /api/stok-opname/forms/:id/items - Add item to form
    - PUT /api/stok-opname/items/:id - Update item
    - DELETE /api/stok-opname/items/:id - Remove item
    - Add request validation and error handling
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 4.1, 4.2_
  
  - [x] 7.4 Implement workflow endpoints
    - POST /api/stok-opname/forms/:id/submit - Submit for approval
    - POST /api/stok-opname/forms/:id/approve - Approve form (Kepala_SPPG only)
    - POST /api/stok-opname/forms/:id/reject - Reject form (Kepala_SPPG only)
    - Add authorization middleware for approval endpoints
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 6.1, 6.2, 6.3, 6.4_
  
  - [x] 7.5 Implement export endpoint
    - GET /api/stok-opname/forms/:id/export?format=excel|pdf
    - Set appropriate content-type headers
    - Handle file download response
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_
  
  - [ ]* 7.6 Write handler tests
    - Test all endpoints with valid requests
    - Test validation errors return HTTP 400
    - Test authorization errors return HTTP 403
    - Test not found errors return HTTP 404
    - Test conflict errors return HTTP 409
    - _Requirements: All requirements_

- [x] 8. Router Integration
  - [x] 8.1 Add stok opname routes to router
    - Register all endpoints in backend/internal/router/router.go
    - Apply authentication middleware to all routes
    - Apply Kepala_SPPG authorization to approval/rejection routes
    - _Requirements: 2.1, 6.1_
  
  - [ ]* 8.2 Test all endpoints with API client
    - Test complete workflow: create → add items → submit → approve
    - Test rejection workflow
    - Test authorization checks
    - Test concurrent access scenarios
    - _Requirements: All requirements_

- [x] 9. Checkpoint - Backend Complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 10. Frontend Service Layer
  - [x] 10.1 Create stokOpnameService.js
    - Create web/src/services/stokOpnameService.js
    - Implement all API client methods (createForm, getForms, getForm, updateFormNotes, deleteForm)
    - Implement item methods (addItem, updateItem, removeItem)
    - Implement workflow methods (submitForApproval, approveForm, rejectForm)
    - Implement export method with blob response handling
    - Add error handling and response transformation
    - _Requirements: 2.1, 2.2, 3.1, 4.1, 4.2, 4.3, 5.1, 6.2, 6.3, 8.1, 9.1, 10.1, 15.1_

- [x] 11. Frontend Components - Tab Navigation
  - [x] 11.1 Modify InventoryView.vue to add Stok Opname tab
    - Add "Stok Opname" tab to existing tabs array
    - Add tab content component slot for StokOpnameList
    - Ensure tab navigation works correctly
    - _Requirements: 1.1, 1.2, 1.3_

- [x] 12. Frontend Components - List View
  - [x] 12.1 Create StokOpnameList.vue component
    - Create web/src/components/StokOpnameList.vue
    - Implement table with columns: form number, date, creator, status, approver
    - Add "Create New Form" button
    - Add search input with debounce (300ms)
    - Add filter controls: status dropdown, date range picker
    - Add action buttons per row: View, Edit (pending only), Delete (pending only), Export
    - Implement pagination (20 items per page)
    - Load data on mount and when filters change
    - _Requirements: 1.3, 2.1, 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 11.1, 11.2, 11.3, 11.4, 15.1_
  
  - [ ]* 12.2 Write component tests for StokOpnameList
    - Test table rendering with data
    - Test search functionality
    - Test filter functionality
    - Test pagination
    - Test action button visibility based on status
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 11.1, 11.2, 11.3, 11.4_

- [x] 13. Frontend Components - Form View
  - [x] 13.1 Create StokOpnameForm.vue component
    - Create web/src/components/StokOpnameForm.vue
    - Display form header: form number (auto-generated), date, creator, status
    - Add form notes textarea
    - Add item list table with columns: ingredient, system stock, physical count, difference, notes, actions
    - Implement ingredient selector (searchable dropdown from inventory)
    - Auto-fill system stock when ingredient selected
    - Calculate and display difference when physical count entered
    - Add "Add Item" button
    - Add "Remove" button per item row
    - Add action buttons: "Save Draft", "Submit for Approval", "Cancel"
    - Implement validation: at least one item, all physical counts filled
    - Show validation errors
    - Disable editing if status is not pending
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4_
  
  - [ ]* 13.2 Write component tests for StokOpnameForm
    - Test form rendering in create mode
    - Test form rendering in edit mode
    - Test item addition and removal
    - Test difference calculation
    - Test validation on submit
    - Test form disabled when not pending
    - _Requirements: 2.1, 3.1, 3.2, 3.3, 3.4, 4.1, 4.2, 4.3, 4.4, 5.2, 5.3, 5.4_

- [x] 14. Frontend Components - Detail View
  - [x] 14.1 Create StokOpnameDetail.vue component
    - Create web/src/components/StokOpnameDetail.vue
    - Display form header: form number, date, creator, status
    - Display form notes if present
    - Display items table: ingredient, system stock, physical count, difference, item notes
    - Display approval information: approver name, approval date (if approved/rejected)
    - Display rejection reason if status is rejected
    - Add conditional action buttons based on status and user role:
      - Pending + Creator: "Edit", "Delete"
      - Pending + Kepala_SPPG: "Approve", "Reject"
      - Any status: "Export"
    - Implement approve modal (confirmation)
    - Implement reject modal (with reason textarea)
    - Implement delete confirmation modal
    - _Requirements: 4.4, 6.1, 6.2, 6.3, 6.4, 8.1, 8.2, 8.3, 8.4, 8.5, 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 10.1, 10.4, 15.1_
  
  - [ ]* 14.2 Write component tests for StokOpnameDetail
    - Test detail rendering with all data
    - Test action button visibility based on status and role
    - Test approve modal
    - Test reject modal with reason
    - Test delete confirmation
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 10.1, 10.4_

- [x] 15. Frontend Routing and Integration
  - [x] 15.1 Add routes for stok opname views
    - Add route for list view: /inventory/stok-opname
    - Add route for create form: /inventory/stok-opname/create
    - Add route for edit form: /inventory/stok-opname/:id/edit
    - Add route for detail view: /inventory/stok-opname/:id
    - Configure route guards for authentication
    - _Requirements: 1.1, 1.2, 1.3, 2.1, 9.1_
  
  - [x] 15.2 Implement navigation between views
    - List → Create: Click "Create New Form" button
    - List → Detail: Click row or "View" button
    - Detail → Edit: Click "Edit" button (pending only)
    - Form → List: Click "Cancel" or after successful save/submit
    - _Requirements: 1.3, 2.1, 9.1_

- [x] 16. Checkpoint - Frontend Complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 17. Integration Testing
  - [ ]* 17.1 End-to-end workflow testing
    - Test complete workflow: Create form → Add items → Submit → Approve → Verify stock adjusted
    - Test rejection workflow: Create → Submit → Reject → Verify form status
    - Test edit workflow: Create → Add items → Edit items → Submit
    - Test delete workflow: Create → Delete → Verify cascade deletion
    - _Requirements: All requirements_
  
  - [ ]* 17.2 Concurrent access testing
    - Test two users editing same form simultaneously
    - Test optimistic locking conflict handling
    - Verify conflict error message displayed
    - _Requirements: 14.3, 14.4_
  
  - [ ]* 17.3 Performance testing
    - Test form with 100+ items
    - Test list view with 1000+ forms
    - Test export with large forms (verify <5 seconds)
    - Test search and filter response time (verify <1 second)
    - _Requirements: 11.4, 15.5_
  
  - [ ]* 17.4 Export functionality testing
    - Test Excel export with complete data
    - Test PDF export with complete data
    - Verify all required fields present in export
    - Test export download in browser
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_
  
  - [ ]* 17.5 Authorization testing
    - Test non-Kepala_SPPG cannot approve/reject
    - Test creator cannot approve own form
    - Test user cannot edit/delete non-pending forms
    - Verify appropriate error messages
    - _Requirements: 4.4, 6.1, 10.4_

- [x] 18. Final Integration and Wiring
  - [x] 18.1 Verify all components integrated correctly
    - Test navigation flow through all views
    - Test data flow from frontend to backend
    - Test error handling and user feedback
    - Verify loading states and indicators
    - _Requirements: All requirements_
  
  - [x] 18.2 Verify notification system integration
    - Test notification sent to Kepala_SPPG on submission
    - Verify notification content includes form details
    - Test notification delivery mechanism
    - _Requirements: 5.5_
  
  - [x] 18.3 Verify audit trail logging
    - Check all form operations logged with user ID and timestamp
    - Check all stock adjustments logged with reference to form
    - Verify audit logs retained for 2 years
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6_
  
  - [x] 18.4 Verify error handling across all layers
    - Test validation errors display correctly
    - Test authorization errors display correctly
    - Test database errors handled gracefully
    - Test network errors handled gracefully
    - _Requirements: 5.4, 12.4, 14.2, 14.4_

- [x] 19. Final Checkpoint - Complete System Verification
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at key milestones
- Property tests validate universal correctness properties across all inputs
- Unit tests validate specific examples and edge cases
- Backend implementation (Phases 1-4) can proceed independently of frontend (Phases 5-8)
- Integration testing (Phase 9) requires both backend and frontend complete
