# Implementation Plan: Pickup Task Management

## Overview

This implementation plan breaks down the Pickup Task Management feature into sequential, incremental tasks. Each task builds on previous work, with testing integrated throughout to catch errors early. The implementation follows five phases: database setup, service layer, API layer, frontend components, and integration.

## Tasks

- [x] 1. Set up database schema and core models
  - [x] 1.1 Create database migration for pickup_tasks table
    - Create migration file `backend/migrations/YYYYMMDD_create_pickup_tasks_table.sql`
    - Add pickup_tasks table with columns: id, task_date, driver_id, status, created_at, updated_at
    - Add indexes for task_date, driver_id, and status
    - Add status check constraint (active, completed, cancelled)
    - _Requirements: 9.1, 9.5_
  
  - [x] 1.2 Extend delivery_records table with pickup fields
    - Add pickup_task_id column (nullable, foreign key to pickup_tasks)
    - Add route_order column (integer, default 0)
    - Add index for pickup_task_id
    - Add check constraint ensuring route_order > 0 when pickup_task_id is set
    - _Requirements: 9.2, 9.3_
  
  - [x] 1.3 Add PickupTask model to Go codebase
    - Add PickupTask struct to `backend/internal/models/logistics.go`
    - Define fields: ID, TaskDate, DriverID, Status, CreatedAt, UpdatedAt
    - Add GORM tags for database mapping
    - Add JSON tags for API serialization
    - Define relationships: Driver (belongs to User), DeliveryRecords (has many)
    - _Requirements: 9.1, 9.5_
  
  - [x] 1.4 Extend DeliveryRecord model with pickup fields
    - Add PickupTaskID field (*uint, nullable) to DeliveryRecord struct
    - Add RouteOrder field (int, default 0)
    - Update GORM tags and JSON tags
    - _Requirements: 9.2, 9.3_
  
  - [x] 1.5 Run database migration
    - Execute migration script against development database
    - Verify tables and columns created correctly
    - Test foreign key constraints
    - _Requirements: 9.1, 9.2_

- [x] 2. Implement PickupTaskService core functionality
  - [x] 2.1 Create PickupTaskService struct and constructor
    - Create file `backend/internal/services/pickup_task_service.go`
    - Define PickupTaskService struct with db and activityTrackerService dependencies
    - Implement NewPickupTaskService constructor
    - _Requirements: 2.3, 2.4_
  
  - [x] 2.2 Implement GetEligibleOrders method
    - Query delivery_records WHERE current_stage = 9 AND pickup_task_id IS NULL
    - Join with schools table to get school information
    - Return EligibleOrderResponse with school name, address, GPS coordinates, ompreng count
    - _Requirements: 1.1, 1.2, 1.3, 1.4_
  
  - [ ]* 2.3 Write property test for GetEligibleOrders
    - **Property 1: Eligible Orders Stage Filter**
    - **Validates: Requirements 1.1, 1.4**
    - Test that all returned records have current_stage = 9 and pickup_task_id IS NULL
  
  - [x] 2.4 Implement GetAvailableDrivers method
    - Query users table WHERE role = 'driver'
    - Return AvailableDriverResponse with driver ID, name, phone number
    - _Requirements: 2.2_
  
  - [x] 2.5 Implement CreatePickupTask method with transaction
    - Begin database transaction
    - Validate all delivery records are at stage 9
    - Validate driver exists and has role 'driver'
    - Validate route_order values are unique
    - Create pickup_task record
    - Update delivery_records with pickup_task_id and route_order
    - Call ActivityTrackerService to transition records to stage 10
    - Commit transaction or rollback on error
    - _Requirements: 2.3, 2.4, 2.5, 7.1, 7.2, 7.3, 7.4_
  
  - [ ]* 2.6 Write property tests for CreatePickupTask
    - **Property 4: Pickup Task Creation Persistence**
    - **Validates: Requirements 2.3, 2.4**
    - **Property 5: Pickup Task Creation Stage Transition**
    - **Validates: Requirements 2.5**
    - **Property 14: Stage 9 Validation on Creation**
    - **Validates: Requirements 7.1**
  
  - [ ]* 2.7 Write property tests for input validation
    - **Property 6: Pickup Task Input Validation**
    - **Validates: Requirements 2.6, 2.7**
    - **Property 15: Driver Existence Validation**
    - **Validates: Requirements 7.2**
    - **Property 16: Route Order Uniqueness Validation**
    - **Validates: Requirements 7.3**
    - **Property 17: No Double Assignment**
    - **Validates: Requirements 7.4**
  
  - [x] 2.8 Implement GetPickupTaskByID method
    - Query pickup_tasks by ID with eager loading
    - Load associated driver and delivery_records
    - Load school information for each delivery record
    - Sort delivery records by route_order
    - _Requirements: 5.5, 5.6, 10.1, 10.2, 10.3_
  
  - [ ]* 2.9 Write property test for GetPickupTaskByID
    - **Property 13: Pickup Task Detail Route Ordering**
    - **Validates: Requirements 5.5, 5.6**
  
  - [x] 2.10 Implement GetActivePickupTasks method
    - Query pickup_tasks WHERE status = 'active'
    - Filter by date and driver_id if provided
    - Load driver information and delivery record count
    - Return summary information for each task
    - _Requirements: 5.1, 5.2, 5.3, 5.4_
  
  - [ ]* 2.11 Write property test for GetActivePickupTasks
    - **Property 12: Active Pickup Tasks Filter**
    - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**
  
  - [x] 2.12 Implement UpdatePickupTaskStatus method
    - Update pickup_task status field
    - Validate status is one of: active, completed, cancelled
    - Update updated_at timestamp
    - _Requirements: 4.6_
  
  - [x] 2.13 Implement CancelPickupTask method
    - Set pickup_task status to 'cancelled'
    - Optionally clear pickup_task_id from associated delivery_records
    - Log cancellation event
    - _Requirements: 7.5_
  
  - [x] 2.14 Implement UpdateDeliveryRecordStage method
    - Validate delivery record belongs to specified pickup task
    - Validate current stage is between 10 and 12 (stage 13 is final)
    - Validate new stage is exactly current_stage + 1 (no skipping)
    - Validate stage-status mapping is correct
    - Call ActivityTrackerService to transition to new stage
    - Update delivery record current_stage and current_status
    - Check if all delivery records in pickup task are at stage 13
    - If all at stage 13, automatically update pickup task status to 'completed'
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.12, 11.13_
  
  - [ ]* 2.15 Write property tests for UpdateDeliveryRecordStage
    - **Property 23: Individual Delivery Record Stage Update**
    - **Validates: Requirements 11.4, 11.13**
    - **Property 24: Sequential Stage Enforcement Per Route**
    - **Validates: Requirements 11.5, 11.6**
    - **Property 27: UI Reflects Stage Update Immediately**
    - **Validates: Requirements 11.12**

- [x] 3. Checkpoint - Verify service layer implementation
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Implement API handlers and routes
  - [x] 4.1 Create PickupTaskHandler struct and constructor
    - Create file `backend/internal/handlers/pickup_task_handler.go`
    - Define PickupTaskHandler struct with service dependency
    - Implement NewPickupTaskHandler constructor
    - _Requirements: 2.1_
  
  - [x] 4.2 Implement GetEligibleOrders HTTP handler
    - Parse optional date query parameter
    - Call service.GetEligibleOrders
    - Return JSON response with eligible_orders array
    - Handle errors with appropriate HTTP status codes
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_
  
  - [x] 4.3 Implement GetAvailableDrivers HTTP handler
    - Parse optional date query parameter
    - Call service.GetAvailableDrivers
    - Return JSON response with available_drivers array
    - _Requirements: 2.2_
  
  - [x] 4.4 Implement CreatePickupTask HTTP handler
    - Parse request body (task_date, driver_id, delivery_records with route_order)
    - Validate request structure
    - Call service.CreatePickupTask
    - Return 201 Created with pickup_task details
    - Handle validation errors with 400 Bad Request
    - Handle conflicts with 409 Conflict
    - _Requirements: 2.1, 2.3, 2.4, 2.5, 2.6, 2.7_
  
  - [x] 4.5 Implement GetAllPickupTasks HTTP handler
    - Parse query parameters: date, driver_id, status
    - Call service.GetActivePickupTasks
    - Return JSON response with pickup_tasks array
    - _Requirements: 5.1, 5.2, 5.3, 5.4_
  
  - [x] 4.6 Implement GetPickupTask HTTP handler
    - Parse pickup task ID from URL parameter
    - Call service.GetPickupTaskByID
    - Return detailed pickup task with delivery records
    - Return 404 if task not found
    - _Requirements: 5.5, 5.6, 10.1, 10.2, 10.3, 10.5_
  
  - [x] 4.7 Implement UpdatePickupTaskStatus HTTP handler
    - Parse pickup task ID and status from request
    - Call service.UpdatePickupTaskStatus
    - Return updated pickup task
    - _Requirements: 4.6_
  
  - [x] 4.8 Implement CancelPickupTask HTTP handler
    - Parse pickup task ID from URL parameter
    - Call service.CancelPickupTask
    - Return success message
    - _Requirements: 7.5_
  
  - [x] 4.9 Implement UpdateDeliveryRecordStage HTTP handler
    - Parse pickup task ID and delivery record ID from URL parameters
    - Parse request body (stage, status)
    - Validate request structure
    - Call service.UpdateDeliveryRecordStage
    - Return updated delivery record with current stage
    - Handle validation errors with 400 Bad Request
    - Handle not found errors with 404 Not Found
    - Handle invalid stage transition with 409 Conflict
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6_
  
  - [x] 4.10 Register UpdateDeliveryRecordStage route in router
    - Edit `backend/internal/router/router.go`
    - Add route: PUT /api/v1/pickup-tasks/:id/delivery-records/:delivery_record_id/stage
    - Apply authentication middleware
    - Apply role-based authorization (kepala_sppg, kepala_yayasan, asisten_lapangan, driver)
    - _Requirements: 11.1_
  
  - [ ]* 4.11 Write integration tests for UpdateDeliveryRecordStage endpoint
    - Test PUT /api/v1/pickup-tasks/:id/delivery-records/:delivery_record_id/stage
    - Test successful stage update (10→11, 11→12, 12→13)
    - Test error scenarios (invalid stage transition, delivery record not in pickup task, stage 13 update attempt)

- [x] 5. Checkpoint - Verify API layer implementation
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Implement frontend API client
  - [x] 6.1 Create pickup tasks API client module
    - Create file `web/src/api/pickupTasks.js`
    - Implement getEligibleOrders(date) method
    - Implement getAvailableDrivers(date) method
    - Implement createPickupTask(data) method
    - Implement getPickupTasks(params) method
    - Implement getPickupTask(id) method
    - Implement updatePickupTaskStatus(id, status) method
    - Implement cancelPickupTask(id) method
    - _Requirements: 1.5, 2.1, 5.1_
  
  - [x] 6.2 Add updateDeliveryRecordStage method to API client
    - Edit `web/src/api/pickupTasks.js`
    - Implement updateDeliveryRecordStage(pickupTaskId, deliveryRecordId, data) method
    - Method should call PUT /api/v1/pickup-tasks/:id/delivery-records/:delivery_record_id/stage
    - _Requirements: 11.1, 11.2_

- [ ] 7. Implement PickupTaskForm component
  - [x] 7.1 Create PickupTaskForm component structure
    - Create file `web/src/components/PickupTaskForm.vue`
    - Set up component template with card layout
    - Define component data properties: eligibleOrders, selectedOrders, selectedDriver, availableDrivers
    - _Requirements: 1.1, 2.1, 3.1_
  
  - [x] 7.2 Implement eligible orders display
    - Add table to display eligible orders
    - Show columns: school name, address, GPS coordinates, ompreng count
    - Add row selection functionality
    - Load eligible orders on component mount
    - _Requirements: 1.1, 1.3, 1.5, 10.1, 10.2, 10.3_
  
  - [x] 7.3 Implement driver selection dropdown
    - Add select dropdown for driver selection
    - Load available drivers on component mount
    - Display driver name and phone number
    - _Requirements: 2.2_
  
  - [x] 7.4 Implement route order management
    - Install vuedraggable package: `npm install vuedraggable`
    - Add draggable list for selected orders
    - Display schools in current route order
    - Allow reordering via drag-and-drop
    - Update route_order values when order changes
    - _Requirements: 3.1, 3.2, 3.5_
  
  - [x] 7.5 Implement form submission
    - Add submit button with validation
    - Disable button if no orders selected or no driver selected
    - Call createPickupTask API on submit
    - Show success message on successful creation
    - Show error message on failure
    - Clear form after successful submission
    - _Requirements: 2.1, 2.3, 2.4, 2.6, 2.7_
  
  - [x] 7.6 Add loading states and error handling
    - Show loading spinner while fetching data
    - Show loading spinner during form submission
    - Display error messages for API failures
    - Handle network errors gracefully
    - _Requirements: 1.5, 2.1_

- [x] 8. Implement PickupTaskList component
  - [x] 8.1 Create PickupTaskList component structure
    - Create file `web/src/components/PickupTaskList.vue`
    - Set up component template with card layout
    - Define component props: date (optional)
    - Define component data properties: pickupTasks, loading
    - _Requirements: 5.1_
  
  - [x] 8.2 Implement pickup tasks table
    - Add table to display active pickup tasks
    - Show columns: task ID, driver name, school count, status, created date
    - Load pickup tasks on component mount
    - Add refresh button
    - _Requirements: 5.1, 5.2, 5.3_
  
  - [x] 8.3 Implement expandable rows for task details
    - Add expandable row functionality
    - Show delivery records for each task when expanded
    - Display school name, address, GPS coordinates, current stage for each record
    - Sort records by route_order
    - _Requirements: 5.4, 5.5, 5.6, 8.3, 10.1, 10.2, 10.3, 10.5_
  
  - [x] 8.4 Add stage progress indicators
    - Display current stage for each delivery record
    - Use color coding for different stages (10=blue, 11=yellow, 12=orange, 13=green)
    - Show stage name in Indonesian
    - _Requirements: 4.1, 4.2, 4.3, 5.4, 8.2_
  
  - [x] 8.5 Add Aksi column with status dropdown for each route
    - Add "Aksi" column to delivery records table in expanded rows
    - Add status dropdown (a-select) for each delivery record
    - Dropdown should show only valid next stage based on current stage
    - Disable dropdown when delivery record is at stage 13 (final stage)
    - _Requirements: 11.7, 11.8, 11.9, 11.10, 11.11_
  
  - [x] 8.6 Implement getAvailableStages() method
    - Create method that returns available next stages based on current stage
    - Stage 10 → returns [{ value: 11, label: 'Sudah Tiba (Stage 11)' }]
    - Stage 11 → returns [{ value: 12, label: 'Dalam Perjalanan Kembali (Stage 12)' }]
    - Stage 12 → returns [{ value: 13, label: 'Selesai (Stage 13)' }]
    - Stage 13 → returns [] (empty, dropdown disabled)
    - _Requirements: 11.7, 11.8, 11.9, 11.10_
  
  - [x] 8.7 Implement handleStageUpdate() method
    - Create method to handle stage update when dropdown value changes
    - Map stage to status: 11→'driver_tiba_di_lokasi_pengambilan', 12→'driver_kembali_ke_sppg', 13→'driver_tiba_di_sppg'
    - Call pickupTasksAPI.updateDeliveryRecordStage with pickup task ID, delivery record ID, stage, and status
    - Show success message on successful update
    - Show error message on failure (including invalid stage transition errors)
    - Refresh pickup tasks list after successful update
    - Emit 'stage-updated' event
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.12_
  
  - [ ]* 8.8 Write property tests for dropdown behavior
    - **Property 25: Dropdown Shows Only Valid Next Stage**
    - **Validates: Requirements 11.7, 11.8, 11.9, 11.10**
    - **Property 26: Dropdown Disabled at Final Stage**
    - **Validates: Requirements 11.11**
  
  - [ ]* 8.9 Write component tests for PickupTaskList
    - Test component rendering with mock data
    - Test expandable row functionality
    - Test stage indicator display
    - Test dropdown rendering and stage options
    - Test handleStageUpdate method with valid and invalid transitions

- [ ] 9. Integrate pickup components into DeliveryTaskListView
  - [x] 9.1 Import pickup components
    - Edit `web/src/views/DeliveryTaskListView.vue`
    - Import PickupTaskForm and PickupTaskList components
    - Register components in component options
    - _Requirements: 6.1_
  
  - [x] 9.2 Add pickup task section to page layout
    - Add divider after existing delivery task section
    - Add page header "Manajemen Tugas Pengambilan"
    - Add tabs for "Buat Tugas Pengambilan" and "Daftar Tugas Aktif"
    - Place PickupTaskForm in first tab
    - Place PickupTaskList in second tab
    - _Requirements: 6.1, 6.2, 6.3_
  
  - [x] 9.3 Implement event handling
    - Listen for task-created event from PickupTaskForm
    - Refresh PickupTaskList when new task is created
    - Show success notification on task creation
    - _Requirements: 2.1, 5.1_
  
  - [x] 9.4 Ensure independent state management
    - Verify pickup form state is independent from delivery form
    - Verify both sections can be used simultaneously
    - Test switching between tabs maintains state
    - _Requirements: 6.4_

- [ ] 10. Implement stage transition validation
  - [x] 10.1 Add stage transition validation to ActivityTrackerService
    - Edit `backend/internal/services/activity_tracker_service.go`
    - Add validation for stages 10→11→12→13 sequence
    - Reject invalid transitions (e.g., 10→13, 11→10)
    - Return descriptive error messages
    - _Requirements: 4.5, 7.5_
  
  - [ ]* 10.2 Write property tests for stage transitions
    - **Property 9: Sequential Stage Transitions**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.5**
    - **Property 10: Stage Transition Timestamp Recording**
    - **Validates: Requirements 4.4**
    - **Property 18: Invalid Stage Transition Rejection**
    - **Validates: Requirements 7.5**
  
  - [x] 10.3 Implement automatic pickup task completion
    - Add logic to check if all delivery records in a pickup task are at stage 13
    - Automatically update pickup task status to 'completed'
    - Trigger on any stage transition to stage 13
    - _Requirements: 4.6_
  
  - [ ]* 10.4 Write property test for pickup task completion
    - **Property 11: Pickup Task Completion**
    - **Validates: Requirements 4.6**

- [ ] 11. Implement multi-school pickup support
  - [x] 11.1 Verify independent stage tracking
    - Test that transitioning one delivery record doesn't affect others
    - Verify route_order is preserved across stage transitions
    - Test multi-school pickup task with different stages per school
    - _Requirements: 8.1, 8.2, 8.4_
  
  - [ ]* 11.2 Write property tests for multi-school scenarios
    - **Property 19: Variable Pickup Task Size Support**
    - **Validates: Requirements 8.1**
    - **Property 20: Independent Stage Tracking**
    - **Validates: Requirements 8.2, 8.4**
    - **Property 21: Multi-School Progress Tracking**
    - **Validates: Requirements 8.3**

- [ ] 12. Implement data persistence verification
  - [ ]* 12.1 Write property test for round-trip persistence
    - **Property 22: Pickup Task Round Trip**
    - **Validates: Requirements 9.1, 9.6**
    - Test creating a pickup task and retrieving it returns matching data
  
  - [ ]* 12.2 Write property tests for route order persistence
    - **Property 7: Route Order Persistence**
    - **Validates: Requirements 3.3**
    - **Property 8: Route Order Starts at One**
    - **Validates: Requirements 3.4**

- [ ] 13. Final checkpoint - Integration testing
  - [x] 13.1 Test complete pickup workflow end-to-end
    - Create delivery records at stage 9
    - Create pickup task with multiple schools
    - Verify stage transitions work correctly
    - Verify pickup task completes when all schools reach stage 13
    - _Requirements: All_
  
  - [x] 13.2 Test error scenarios
    - Test creating pickup task with invalid data
    - Test double assignment prevention
    - Test invalid stage transitions
    - Verify error messages are clear and helpful
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_
  
  - [x] 13.3 Test authorization and permissions
    - Verify role-based access control works
    - Test that drivers can only view their own tasks
    - Test that dispatchers can create and manage all tasks
    - _Requirements: 2.1, 5.1_
  
  - [x] 13.4 Verify UI/UX functionality
    - Test form validation and error display
    - Test drag-and-drop route ordering
    - Test table sorting and filtering
    - Test responsive layout on different screen sizes
    - _Requirements: 1.1, 2.1, 3.1, 5.1, 6.1_

- [x] 14. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP delivery
- Each task references specific requirements for traceability
- Property tests validate universal correctness properties from the design document
- Unit tests validate specific examples and edge cases
- The implementation follows a bottom-up approach: database → service → API → frontend
- Checkpoints ensure incremental validation and provide opportunities for user feedback
- All stage transitions must go through ActivityTrackerService to maintain consistency
- Route ordering uses drag-and-drop for better UX but can fall back to manual input
- Authorization is enforced at the API layer with role-based middleware
