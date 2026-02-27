# Implementation Plan: Logistics Monitoring Process

## Overview

This implementation plan breaks down the Logistics Monitoring Process feature into discrete, actionable coding tasks. The feature tracks menu deliveries and ompreng (food container) management through 15 lifecycle stages, integrating with existing KDS Cooking and KDS Packing modules while introducing a new KDS Cleaning module.

The implementation follows a logical sequence: database setup → backend models and services → API endpoints → frontend components → integration → testing. Each task builds incrementally on previous work to ensure continuous validation.

## Tasks

- [x] 1. Database schema and migrations
  - [x] 1.1 Create database migration for delivery_records table
    - Create migration file with delivery_records table schema
    - Include indexes for delivery_date, school_id, driver_id, and current_status
    - Add foreign key constraints to schools, users (driver), and menu_items tables
    - Add check constraints for positive portions and non-negative ompreng_count
    - _Requirements: 1.1, 11.1, 11.2, 12.1_
  
  - [x] 1.2 Create database migration for status_transitions table
    - Create migration file with status_transitions table schema
    - Include indexes for delivery_record_id, transitioned_at, and transitioned_by
    - Add foreign key constraint to delivery_records with CASCADE delete
    - Add foreign key constraint to users table for transitioned_by
    - _Requirements: 1.5, 9.1, 9.2_
  
  - [x] 1.3 Create database migration for ompreng_cleanings table
    - Create migration file with ompreng_cleanings table schema
    - Include indexes for delivery_record_id, cleaning_status, and cleaned_by
    - Add foreign key constraints to delivery_records and users tables
    - Add check constraint for positive ompreng_count
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 7.1_

  - [x] 1.4 Update users table to add kebersihan role
    - Create migration to modify role constraint on users table
    - Add "kebersihan" to the list of allowed roles
    - _Requirements: 8.1, 8.4_

- [x] 2. Backend models implementation
  - [x] 2.1 Create DeliveryRecord model
    - Define DeliveryRecord struct with all fields and Gorm tags
    - Include associations for School, Driver, and MenuItem
    - Add JSON tags for API responses
    - _Requirements: 1.1, 1.3, 1.4, 11.1, 11.2_
  
  - [x] 2.2 Create StatusTransition model
    - Define StatusTransition struct with all fields and Gorm tags
    - Include associations for DeliveryRecord and User
    - Add JSON tags for API responses
    - _Requirements: 1.5, 9.1_
  
  - [x] 2.3 Create OmprengCleaning model
    - Define OmprengCleaning struct with all fields and Gorm tags
    - Include associations for DeliveryRecord and Cleaner (User)
    - Add JSON tags for API responses
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  
  - [ ]* 2.4 Write property test for database models
    - **Property 20: Delivery Record Associations**
    - **Validates: Requirements 11.1, 11.2**
    - Test that any delivery record is associated with exactly one school and one driver

- [x] 3. Status transition validation logic
  - [x] 3.1 Implement status transition rules map
    - Define statusTransitionRules map with all 15 stages
    - Map each status to its allowed next statuses
    - Include empty array for final state (ompreng_selesai_dicuci)
    - _Requirements: 14.1, 14.2_
  
  - [x] 3.2 Implement ValidateStatusTransition function
    - Check if requested transition is allowed from current status
    - Return InvalidTransitionError if not allowed
    - Return allowed statuses in error for user guidance
    - _Requirements: 14.1, 14.2_
  
  - [x] 3.3 Implement stage sequence validation
    - Validate delivery stages (1-8) occur before collection stages (9-13)
    - Validate collection stages occur before cleaning stages (14-15)
    - Return error if sequence is violated
    - _Requirements: 14.3, 14.4_
  
  - [ ]* 3.4 Write property test for status transition validation
    - **Property 24: Status Transition Validation**
    - **Validates: Requirements 14.1, 14.2**
    - Test that invalid transitions are rejected with appropriate error

  - [ ]* 3.5 Write property test for stage sequence enforcement
    - **Property 25: Stage Sequence Enforcement**
    - **Validates: Requirements 14.3, 14.4**
    - Test that delivery stages must complete before collection, collection before cleaning

- [x] 4. MonitoringService implementation
  - [x] 4.1 Create MonitoringService struct and constructor
    - Define MonitoringService with db, firebaseApp, and dbClient fields
    - Implement NewMonitoringService constructor
    - _Requirements: 1.1_
  
  - [x] 4.2 Implement GetDeliveryRecords method
    - Query delivery_records by date with optional filters (school, status, driver)
    - Preload School, Driver, and MenuItem associations
    - Return array of delivery records
    - _Requirements: 1.1, 12.1, 12.2, 12.3, 12.4, 12.5_
  
  - [x] 4.3 Implement GetDeliveryRecordDetail method
    - Query single delivery record by ID
    - Preload all associations (School, Driver, MenuItem)
    - Return detailed delivery record with all related data
    - _Requirements: 1.2, 1.3, 1.4_
  
  - [x] 4.4 Implement UpdateDeliveryStatus method
    - Validate status transition using ValidateStatusTransition
    - Update delivery record current_status in database
    - Create StatusTransition record with timestamp and user
    - Trigger Firebase sync asynchronously
    - _Requirements: 2.1-2.8, 3.1-3.5, 9.1, 13.1-13.5_
  
  - [x] 4.5 Implement GetActivityLog method
    - Query status_transitions for delivery record
    - Order by transitioned_at chronologically
    - Preload User association for each transition
    - Calculate elapsed time between consecutive transitions
    - _Requirements: 1.5, 9.2, 9.3_
  
  - [x] 4.6 Implement GetDailySummary method
    - Count total delivery records for date
    - Count records by each status
    - Count completed deliveries (sudah_diterima_pihak_sekolah)
    - Count ompreng in cleaning and cleaned
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_
  
  - [x] 4.7 Implement syncToFirebase method
    - Format delivery record data for Firebase
    - Write to Firebase path: /monitoring/deliveries/{date}/record_{id}
    - Handle errors with retry queue mechanism
    - _Requirements: 1.1_
  
  - [ ]* 4.8 Write property test for date-based retrieval
    - **Property 1: Date-based Delivery Record Retrieval**
    - **Validates: Requirements 1.1, 12.1**
    - Test that querying by date returns only records for that date

  - [ ]* 4.9 Write property test for multi-criteria filtering
    - **Property 23: Multi-criteria Filtering**
    - **Validates: Requirements 12.3, 12.4, 12.5**
    - Test that filtering by date, school, or status returns only matching records
  
  - [ ]* 4.10 Write property test for summary statistics accuracy
    - **Property 26: Summary Statistics Accuracy**
    - **Validates: Requirements 15.1, 15.2, 15.3, 15.4, 15.5**
    - Test that summary counts match actual record counts for all criteria

- [x] 5. CleaningService implementation
  - [x] 5.1 Create CleaningService struct and constructor
    - Define CleaningService with db, firebaseApp, and dbClient fields
    - Implement NewCleaningService constructor
    - _Requirements: 7.1_
  
  - [x] 5.2 Implement GetPendingOmpreng method
    - Query ompreng_cleanings with status "pending" or delivery records with status "ompreng_sampai_di_sppg"
    - Preload DeliveryRecord with School association
    - Return array of pending ompreng cleaning records
    - _Requirements: 7.1, 7.4_
  
  - [x] 5.3 Implement StartCleaning method
    - Update ompreng_cleaning status to "in_progress"
    - Set started_at timestamp
    - Set cleaned_by to current user ID
    - Update corresponding delivery record status to "ompreng_proses_pencucian"
    - Sync to Firebase
    - _Requirements: 4.1, 7.2, 7.5_
  
  - [x] 5.4 Implement CompleteCleaning method
    - Update ompreng_cleaning status to "completed"
    - Set completed_at timestamp
    - Update corresponding delivery record status to "ompreng_selesai_dicuci"
    - Sync to Firebase
    - _Requirements: 4.2, 7.3, 7.5_
  
  - [x] 5.5 Implement SyncToFirebase method for cleaning
    - Format cleaning record data for Firebase
    - Write to Firebase path: /cleaning/pending/{cleaning_id}
    - Handle errors with retry mechanism
    - _Requirements: 7.1_
  
  - [ ]* 5.6 Write property test for cleaning status updates
    - **Property 13: Cleaning Status Updates**
    - **Validates: Requirements 7.2, 7.3**
    - Test that cleaning staff can update status through the workflow
  
  - [ ]* 5.7 Write property test for cleaning record association
    - **Property 7: Cleaning Record Association**
    - **Validates: Requirements 4.4**
    - Test that cleaning records are associated with exactly one delivery record

- [x] 6. Checkpoint - Ensure backend services compile and basic tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. API handlers and endpoints
  - [x] 7.1 Create MonitoringHandler struct
    - Define MonitoringHandler with monitoringService field
    - Implement NewMonitoringHandler constructor
    - _Requirements: 1.1_
  
  - [x] 7.2 Implement GetDeliveryRecords endpoint handler
    - Parse date query parameter (required)
    - Parse optional filters (school_id, status, driver_id)
    - Call monitoringService.GetDeliveryRecords
    - Return JSON response with delivery records array
    - Handle errors with appropriate HTTP status codes
    - _Requirements: 1.1, 12.3, 12.4, 12.5_
  
  - [x] 7.3 Implement GetDeliveryDetail endpoint handler
    - Parse delivery record ID from path parameter
    - Call monitoringService.GetDeliveryRecordDetail
    - Return JSON response with detailed record
    - Handle not found and other errors
    - _Requirements: 1.2, 1.3, 1.4_
  
  - [x] 7.4 Implement UpdateStatus endpoint handler
    - Parse delivery record ID from path parameter
    - Parse status and notes from request body
    - Get user ID from JWT context
    - Validate user role has permission for status update
    - Call monitoringService.UpdateDeliveryStatus
    - Return updated record or error response
    - _Requirements: 2.1-2.8, 3.1-3.5, 13.1-13.5_
  
  - [x] 7.5 Implement GetActivityLog endpoint handler
    - Parse delivery record ID from path parameter
    - Call monitoringService.GetActivityLog
    - Return JSON response with activity log array
    - _Requirements: 1.5, 9.2_
  
  - [x] 7.6 Implement GetDailySummary endpoint handler
    - Parse date query parameter (required)
    - Call monitoringService.GetDailySummary
    - Return JSON response with summary statistics
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_
  
  - [x] 7.7 Create CleaningHandler struct
    - Define CleaningHandler with cleaningService field
    - Implement NewCleaningHandler constructor
    - _Requirements: 7.1_
  
  - [x] 7.8 Implement GetPendingOmpreng endpoint handler
    - Call cleaningService.GetPendingOmpreng
    - Return JSON response with pending ompreng array
    - _Requirements: 7.1, 7.4_
  
  - [x] 7.9 Implement StartCleaning endpoint handler
    - Parse cleaning record ID from path parameter
    - Get user ID from JWT context
    - Validate user has kebersihan role
    - Call cleaningService.StartCleaning
    - Return updated cleaning record
    - _Requirements: 7.2, 7.5_

  - [x] 7.10 Implement CompleteCleaning endpoint handler
    - Parse cleaning record ID from path parameter
    - Get user ID from JWT context
    - Validate user has kebersihan role
    - Call cleaningService.CompleteCleaning
    - Return updated cleaning record
    - _Requirements: 7.3, 7.5_
  
  - [x] 7.11 Register monitoring routes in router
    - Add GET /api/monitoring/deliveries route
    - Add GET /api/monitoring/deliveries/:id route
    - Add PUT /api/monitoring/deliveries/:id/status route
    - Add GET /api/monitoring/deliveries/:id/activity route
    - Add GET /api/monitoring/summary route
    - Apply authentication middleware to all routes
    - Apply role-based authorization middleware (exclude kebersihan)
    - _Requirements: 1.1, 8.2, 8.3_
  
  - [x] 7.12 Register cleaning routes in router
    - Add GET /api/cleaning/pending route
    - Add POST /api/cleaning/:id/start route
    - Add POST /api/cleaning/:id/complete route
    - Apply authentication middleware to all routes
    - Apply kebersihan role authorization middleware
    - _Requirements: 7.1, 7.2, 7.3, 8.2_

- [x] 8. Role-based access control implementation
  - [x] 8.1 Update authorization middleware for kebersihan role
    - Add kebersihan role to role validation
    - Define endpoint permissions map with kebersihan access rules
    - Implement logic to grant kebersihan access to cleaning endpoints only
    - Implement logic to deny kebersihan access to other KDS modules
    - _Requirements: 8.1, 8.2, 8.3_
  
  - [x] 8.2 Implement context-dependent status update authorization
    - Check user role matches status category (chef for cooking, driver for delivery, etc.)
    - Allow kepala_sppg and kepala_yayasan to override any status
    - Return 403 Forbidden if user lacks permission
    - _Requirements: 8.2_
  
  - [ ]* 8.3 Write property test for role-based access control
    - **Property 15: Role-Based Access Control**
    - **Validates: Requirements 8.2, 8.3**
    - Test that kebersihan role has access to cleaning endpoints and denied access to others

- [x] 9. Checkpoint - Ensure API endpoints work correctly
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 10. Frontend monitoring dashboard view
  - [x] 10.1 Create MonitoringDashboardView.vue component
    - Create Vue component with template, script, and style sections
    - Add date picker for selecting delivery date
    - Add summary statistics cards section
    - Add delivery records list section with filters
    - _Requirements: 1.1, 15.1, 15.2, 15.3, 15.4, 15.5_

  - [x] 10.2 Implement date selection and data fetching
    - Add reactive date state with default to today
    - Implement fetchDeliveryRecords API call
    - Implement fetchDailySummary API call
    - Call both APIs when date changes
    - Display loading states during fetch
    - _Requirements: 1.1, 12.3_
  
  - [x] 10.3 Implement summary statistics display
    - Create cards for total deliveries, completed, in-progress
    - Display counts by status using summary data
    - Display ompreng cleaning statistics
    - Use Ant Design Vue Card and Statistic components
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_
  
  - [x] 10.4 Implement delivery records list with filtering
    - Display delivery records in Ant Design Vue Table
    - Add filter dropdowns for school, status, driver
    - Implement client-side filtering logic
    - Display school name, driver name, status, portions
    - Add click handler to navigate to detail view
    - _Requirements: 1.1, 12.2, 12.4, 12.5_
  
  - [x] 10.5 Implement status indicators in list
    - Add status badge component with color coding
    - Display completed indicator (green) for finished stages
    - Display in-progress indicator (blue) for current stage
    - Display pending indicator (gray) for future stages
    - _Requirements: 10.1, 10.2, 10.3_
  
  - [ ]* 10.6 Write unit tests for MonitoringDashboardView
    - Test date selection triggers data fetch
    - Test filtering logic works correctly
    - Test summary statistics display
    - Test navigation to detail view

- [ ] 11. Frontend delivery detail view
  - [x] 11.1 Create DeliveryDetailView.vue component
    - Create Vue component with template, script, and style sections
    - Add sections for school info, driver info, timeline, activity log
    - Add back button to return to dashboard
    - _Requirements: 1.2, 1.3, 1.4, 1.5_
  
  - [x] 11.2 Implement delivery record detail fetching
    - Get delivery record ID from route params
    - Implement fetchDeliveryDetail API call
    - Implement fetchActivityLog API call
    - Display loading state during fetch
    - Handle not found error
    - _Requirements: 1.2, 1.5_
  
  - [x] 11.3 Implement school information display
    - Display school name, address, contact information
    - Display portion count for this delivery
    - Use Ant Design Vue Descriptions component
    - _Requirements: 1.3, 11.3_

  - [x] 11.4 Implement driver information display
    - Display driver name, vehicle type, contact information
    - Use Ant Design Vue Descriptions component
    - _Requirements: 1.4, 11.4_
  
  - [x] 11.5 Integrate DeliveryTimeline component
    - Import and use DeliveryTimeline component
    - Pass current status and activity log as props
    - Display timeline showing all 15 stages
    - _Requirements: 1.2, 10.4_
  
  - [x] 11.6 Integrate ActivityLogTable component
    - Import and use ActivityLogTable component
    - Pass activity log data as prop
    - Display chronological list of status transitions
    - _Requirements: 1.5, 9.2_
  
  - [ ]* 11.7 Write unit tests for DeliveryDetailView
    - Test data fetching on mount
    - Test school and driver info display
    - Test timeline and activity log integration
    - Test error handling for not found

- [ ] 12. Frontend timeline component
  - [x] 12.1 Create DeliveryTimeline.vue component
    - Create Vue component accepting currentStatus and activityLog props
    - Define all 15 lifecycle stages in order
    - Add template for timeline visualization
    - _Requirements: 1.2, 10.4_
  
  - [x] 12.2 Implement timeline stage rendering
    - Map each of 15 stages to timeline items
    - Display stage name and description
    - Use Ant Design Vue Timeline component
    - _Requirements: 1.2, 10.4_
  
  - [x] 12.3 Implement status indicator logic
    - Determine if stage is completed, in-progress, or pending
    - Apply appropriate color and icon for each state
    - Completed: green checkmark icon
    - In-progress: blue loading icon
    - Pending: gray circle icon
    - _Requirements: 10.1, 10.2, 10.3_
  
  - [x] 12.4 Display timestamps for completed stages
    - Extract timestamp from activity log for each completed stage
    - Format timestamp in local timezone (Asia/Jakarta)
    - Display timestamp below stage name
    - _Requirements: 9.1, 9.4_
  
  - [ ]* 12.5 Write property test for timeline completeness
    - **Property 2: Timeline Completeness**
    - **Validates: Requirements 1.2, 10.4**
    - Test that timeline contains all 15 stages in sequential order
  
  - [ ]* 12.6 Write property test for status indicator rendering
    - **Property 19: Status Indicator Rendering**
    - **Validates: Requirements 10.1, 10.2, 10.3**
    - Test that each stage displays appropriate indicator based on current status

- [ ] 13. Frontend activity log component
  - [x] 13.1 Create ActivityLogTable.vue component
    - Create Vue component accepting activityLog prop
    - Add template for table display
    - Use Ant Design Vue Table component
    - _Requirements: 1.5, 9.2_
  
  - [x] 13.2 Implement activity log table columns
    - Add column for timestamp (formatted in local timezone)
    - Add column for from_status
    - Add column for to_status
    - Add column for user name and role
    - Add column for notes
    - _Requirements: 1.5, 9.1, 9.4_
  
  - [x] 13.3 Implement elapsed time calculation
    - Calculate time difference between consecutive transitions
    - Display elapsed time in human-readable format (e.g., "2h 30m")
    - Add elapsed time column to table
    - _Requirements: 9.3_
  
  - [x] 13.4 Implement chronological sorting
    - Sort activity log by transitioned_at in ascending order
    - Ensure most recent transition is at bottom
    - _Requirements: 9.2_
  
  - [ ]* 13.5 Write property test for activity log completeness
    - **Property 4: Activity Log Completeness**
    - **Validates: Requirements 1.5, 9.2**
    - Test that activity log contains all transitions in chronological order
  
  - [ ]* 13.6 Write property test for elapsed time calculation
    - **Property 17: Elapsed Time Calculation**
    - **Validates: Requirements 9.3**
    - Test that elapsed time between consecutive transitions is calculated correctly

- [ ] 14. Frontend KDS Cleaning view
  - [x] 14.1 Create KDSCleaningView.vue component
    - Create Vue component with template, script, and style sections
    - Add sections for pending ompreng list and action buttons
    - _Requirements: 7.1_
  
  - [x] 14.2 Implement pending ompreng list fetching
    - Implement fetchPendingOmpreng API call
    - Display loading state during fetch
    - Refresh list periodically or on Firebase update
    - _Requirements: 7.1_
  
  - [x] 14.3 Implement pending ompreng list display
    - Display ompreng in Ant Design Vue Table or List
    - Show school name, delivery date, ompreng count
    - Show current cleaning status
    - _Requirements: 7.1, 7.4_
  
  - [x] 14.4 Implement start cleaning action
    - Add "Start Cleaning" button for each pending ompreng
    - Call startCleaning API endpoint
    - Update local state and refresh list
    - Show success notification
    - _Requirements: 7.2_

  - [x] 14.5 Implement complete cleaning action
    - Add "Complete Cleaning" button for in-progress ompreng
    - Call completeCleaning API endpoint
    - Update local state and refresh list
    - Show success notification
    - _Requirements: 7.3_
  
  - [x] 14.6 Implement Firebase real-time updates
    - Subscribe to Firebase /cleaning/pending path
    - Update component state when Firebase data changes
    - Unsubscribe on component unmount
    - _Requirements: 7.1_
  
  - [ ]* 14.7 Write property test for cleaning queue filtering
    - **Property 12: Cleaning Queue Filtering**
    - **Validates: Requirements 7.1**
    - Test that only ompreng with status "ompreng_sampai_di_sppg" are displayed
  
  - [ ]* 14.8 Write unit tests for KDSCleaningView
    - Test pending ompreng list display
    - Test start cleaning action
    - Test complete cleaning action
    - Test Firebase real-time updates

- [ ] 15. Checkpoint - Ensure frontend components render correctly
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 16. Integration with KDS Cooking module
  - [x] 16.1 Locate KDS Cooking service status update method
    - Find the method that updates recipe/menu item cooking status
    - Identify where "cooking" and "ready" statuses are set
    - _Requirements: 5.1, 5.2_
  
  - [x] 16.2 Add monitoring system trigger for cooking started
    - After setting status to "cooking", call monitoringService.UpdateDeliveryStatus
    - Pass status "sedang_dimasak" with user ID
    - Handle errors gracefully (log but don't block cooking workflow)
    - _Requirements: 5.1, 5.4_
  
  - [x] 16.3 Add monitoring system trigger for cooking completed
    - After setting status to "ready", call monitoringService.UpdateDeliveryStatus
    - Pass status "selesai_dimasak" with user ID
    - Handle errors gracefully
    - _Requirements: 5.2, 5.4_
  
  - [x] 16.4 Ensure menu item details are available
    - Verify delivery record creation includes menu item ID
    - Verify quantity and school assignment are captured
    - _Requirements: 5.3_
  
  - [ ]* 16.5 Write property test for KDS Cooking integration
    - **Property 8: KDS Cooking Integration**
    - **Validates: Requirements 5.1, 5.2, 5.4**
    - Test that cooking status updates trigger monitoring system updates with timestamps

- [ ] 17. Integration with KDS Packing module
  - [x] 17.1 Locate KDS Packing service status update method
    - Find the method that updates packing status
    - Identify where "ready_for_packing" and "packed" statuses are set
    - _Requirements: 6.1, 6.2_
  
  - [x] 17.2 Add monitoring system trigger for packing ready
    - After setting status to "ready_for_packing", call monitoringService.UpdateDeliveryStatus
    - Pass status "siap_dipacking" with user ID
    - Handle errors gracefully
    - _Requirements: 6.1, 6.4_
  
  - [x] 17.3 Add monitoring system trigger for packing completed
    - After setting status to "packed", call monitoringService.UpdateDeliveryStatus
    - Pass status "selesai_dipacking" with user ID
    - Handle errors gracefully
    - _Requirements: 6.2, 6.4_
  
  - [x] 17.4 Ensure packing completion data is available
    - Verify delivery record includes packing completion timestamp
    - Verify packing staff attribution is captured
    - _Requirements: 6.3_
  
  - [ ]* 17.5 Write property test for KDS Packing integration
    - **Property 10: KDS Packing Integration**
    - **Validates: Requirements 6.1, 6.2, 6.4**
    - Test that packing status updates trigger monitoring system updates with timestamps

- [ ] 18. Firebase real-time synchronization
  - [ ] 18.1 Implement Firebase sync for delivery records
    - Format delivery record data for Firebase structure
    - Write to /monitoring/deliveries/{date}/record_{id} path
    - Include school_name, driver_name, current_status, portions, ompreng_count, last_updated
    - _Requirements: 1.1_
  
  - [ ] 18.2 Implement Firebase sync for cleaning records
    - Format cleaning record data for Firebase structure
    - Write to /cleaning/pending/{cleaning_id} path
    - Include delivery_record_id, school_name, ompreng_count, status, arrived_at
    - _Requirements: 7.1_
  
  - [x] 18.3 Implement retry mechanism for failed syncs
    - Create retry queue for failed Firebase writes
    - Implement exponential backoff (1s, 2s, 4s, 8s, 16s)
    - Set maximum retry attempts to 5
    - Log errors after max retries for admin alerting
    - _Requirements: 1.1_
  
  - [ ] 18.4 Implement Firebase listeners in frontend
    - Subscribe to /monitoring/deliveries/{date} in MonitoringDashboardView
    - Subscribe to /cleaning/pending in KDSCleaningView
    - Update component state when Firebase data changes
    - Handle connection errors gracefully
    - _Requirements: 1.1, 7.1_
  
  - [ ]* 18.5 Write unit tests for Firebase synchronization
    - Test successful sync writes data correctly
    - Test retry mechanism activates on failure
    - Test max retries logs error
    - Test frontend listeners update state

- [ ] 19. Frontend routing and navigation
  - [x] 19.1 Add monitoring dashboard route
    - Add route for /logistics/monitoring in router configuration
    - Set component to MonitoringDashboardView
    - Add route to navigation menu under Logistics section
    - Restrict access to all roles except kebersihan
    - _Requirements: 1.1, 8.3_
  
  - [x] 19.2 Add delivery detail route
    - Add route for /logistics/monitoring/deliveries/:id
    - Set component to DeliveryDetailView
    - Configure route to accept delivery record ID parameter
    - Restrict access to all roles except kebersihan
    - _Requirements: 1.2, 8.3_
  
  - [x] 19.3 Add KDS Cleaning route
    - Add route for /kds/cleaning in router configuration
    - Set component to KDSCleaningView
    - Add route to navigation menu in KDS section
    - Restrict access to kebersihan role only
    - _Requirements: 7.1, 8.2_
  
  - [x] 19.4 Update navigation menu with role-based visibility
    - Show monitoring dashboard link to all roles except kebersihan
    - Show KDS Cleaning link only to kebersihan role
    - Hide other KDS module links from kebersihan role
    - _Requirements: 8.2, 8.3_

- [ ] 20. Error handling and user feedback
  - [ ] 20.1 Implement error handling for invalid transitions
    - Display error message when status transition is rejected
    - Show current status and allowed next statuses
    - Use Ant Design Vue notification component
    - _Requirements: 14.1, 14.2_
  
  - [ ] 20.2 Implement error handling for unauthorized access
    - Display 403 Forbidden message when user lacks permission
    - Redirect to appropriate page based on user role
    - Log unauthorized access attempts
    - _Requirements: 8.2, 8.3_
  
  - [ ] 20.3 Implement loading states for all async operations
    - Show loading spinner during API calls
    - Disable action buttons during processing
    - Show skeleton loaders for data tables
    - _Requirements: 1.1, 7.1_
  
  - [ ] 20.4 Implement success notifications
    - Show success message after status update
    - Show success message after cleaning action
    - Auto-dismiss notifications after 3 seconds
    - _Requirements: 2.1-2.8, 7.2, 7.3_
  
  - [ ] 20.5 Implement validation error messages
    - Display validation errors for missing required fields
    - Display validation errors for invalid data (negative counts, etc.)
    - Show field-level error messages in forms
    - _Requirements: 1.1, 7.1_

- [ ] 21. Timestamp formatting and timezone handling
  - [ ] 21.1 Implement timezone conversion utility
    - Create utility function to convert UTC to Asia/Jakarta timezone
    - Format timestamps in readable format (e.g., "15 Jan 2024, 10:30 WIB")
    - _Requirements: 9.4_
  
  - [ ] 21.2 Apply timezone formatting to activity log
    - Use timezone utility in ActivityLogTable component
    - Display all timestamps in local timezone
    - _Requirements: 9.4_
  
  - [ ] 21.3 Apply timezone formatting to timeline
    - Use timezone utility in DeliveryTimeline component
    - Display stage completion timestamps in local timezone
    - _Requirements: 9.4_
  
  - [ ] 21.4 Apply timezone formatting to cleaning view
    - Use timezone utility in KDSCleaningView component
    - Display arrival and completion timestamps in local timezone
    - _Requirements: 9.4_
  
  - [ ]* 21.5 Write property test for timezone display
    - **Property 18: Timezone Display**
    - **Validates: Requirements 9.4**
    - Test that all displayed timestamps are converted to Asia/Jakarta timezone

- [ ] 22. Integration testing and end-to-end scenarios
  - [ ]* 22.1 Write integration test for complete delivery lifecycle
    - Test full workflow from cooking through delivery to cleaning
    - Verify all 15 status transitions are recorded
    - Verify timestamps are present for each transition
    - Verify Firebase sync occurs at each stage
    - _Requirements: 2.1-2.8, 3.1-3.5, 4.1-4.3, 9.1_
  
  - [ ]* 22.2 Write integration test for multi-school delivery day
    - Create delivery records for 10 different schools
    - Progress each through different stages
    - Verify filtering by school works correctly
    - Verify filtering by status works correctly
    - Verify summary statistics are accurate
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 15.1-15.5_
  
  - [ ]* 22.3 Write integration test for error recovery scenarios
    - Test invalid status transition is rejected
    - Verify error response format
    - Verify status remains unchanged after rejection
    - Test unauthorized access is denied
    - Test Firebase connection failure triggers retry
    - _Requirements: 14.1, 14.2, 8.2, 8.3_
  
  - [ ]* 22.4 Write integration test for role-based access
    - Test kebersihan role can access cleaning endpoints
    - Test kebersihan role cannot access monitoring endpoints
    - Test other roles can access monitoring endpoints
    - Test other roles cannot access cleaning endpoints
    - _Requirements: 8.1, 8.2, 8.3_

- [ ] 23. Performance optimization and testing
  - [ ] 23.1 Add database indexes for query optimization
    - Verify indexes exist on delivery_date, school_id, driver_id, current_status
    - Verify indexes exist on transitioned_at, delivery_record_id
    - Verify indexes exist on cleaning_status, cleaned_by
    - _Requirements: 1.1, 12.3, 12.4, 12.5_
  
  - [ ] 23.2 Implement pagination for large result sets
    - Add pagination to delivery records list API
    - Add pagination to activity log API
    - Add pagination to cleaning history API
    - Use limit and offset query parameters
    - _Requirements: 1.1, 1.5, 7.1_
  
  - [ ] 23.3 Optimize Firebase sync to be non-blocking
    - Ensure Firebase sync runs asynchronously
    - Don't block API response on Firebase completion
    - Log sync errors without failing the request
    - _Requirements: 1.1_
  
  - [ ]* 23.4 Write performance tests for API endpoints
    - Test delivery records query with 100+ records completes in <500ms
    - Test status update completes in <200ms
    - Test activity log retrieval completes in <300ms
    - Test summary statistics completes in <400ms
    - _Requirements: 1.1, 1.5, 15.1-15.5_

- [ ] 24. Documentation and code comments
  - [ ] 24.1 Add code comments to backend services
    - Document MonitoringService methods with purpose and parameters
    - Document CleaningService methods with purpose and parameters
    - Document status transition validation logic
    - _Requirements: All_
  
  - [ ] 24.2 Add code comments to API handlers
    - Document each endpoint with purpose, parameters, and responses
    - Document authentication and authorization requirements
    - Document error responses
    - _Requirements: All_
  
  - [ ] 24.3 Add code comments to frontend components
    - Document component props and emitted events
    - Document complex logic and calculations
    - Document Firebase integration points
    - _Requirements: All_
  
  - [ ] 24.4 Add JSDoc/TSDoc comments for utility functions
    - Document timezone conversion utility
    - Document elapsed time calculation utility
    - Document status indicator logic
    - _Requirements: 9.3, 9.4, 10.1-10.3_

- [ ] 25. Final checkpoint and integration verification
  - [ ] 25.1 Run all unit tests and property-based tests
    - Ensure all backend tests pass
    - Ensure all frontend tests pass
    - Verify test coverage meets requirements
    - _Requirements: All_
  
  - [ ] 25.2 Verify database migrations run successfully
    - Test migrations on clean database
    - Verify all tables and indexes are created
    - Verify foreign key constraints work correctly
    - _Requirements: 1.1, 4.4, 11.1, 11.2_

  - [ ] 25.3 Verify role-based access control works end-to-end
    - Test kebersihan user can access cleaning module
    - Test kebersihan user cannot access other modules
    - Test other roles can access monitoring dashboard
    - Test status update authorization works correctly
    - _Requirements: 8.1, 8.2, 8.3_
  
  - [ ] 25.4 Verify Firebase real-time synchronization works
    - Test delivery status updates sync to Firebase
    - Test cleaning status updates sync to Firebase
    - Test frontend receives real-time updates
    - Test retry mechanism works on connection failure
    - _Requirements: 1.1, 7.1_
  
  - [ ] 25.5 Verify integration with KDS modules works
    - Test cooking status updates trigger monitoring updates
    - Test packing status updates trigger monitoring updates
    - Test status transitions are recorded with correct timestamps
    - _Requirements: 5.1, 5.2, 6.1, 6.2_
  
  - [ ] 25.6 Perform manual testing of complete workflow
    - Create test delivery record
    - Progress through all 15 stages manually
    - Verify timeline displays correctly at each stage
    - Verify activity log records all transitions
    - Verify summary statistics update correctly
    - Verify cleaning workflow completes successfully
    - _Requirements: All_
  
  - [x] 25.7 Final checkpoint - Ensure all tests pass
    - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional testing tasks and can be skipped for faster MVP delivery
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at key milestones
- Property tests validate universal correctness properties across all inputs
- Unit tests validate specific examples, edge cases, and integration points
- The implementation follows a logical sequence: database → backend → API → frontend → integration
- Firebase synchronization is implemented asynchronously to avoid blocking API responses
- Role-based access control is enforced at both API and frontend routing levels
- All timestamps are stored in UTC and converted to Asia/Jakarta timezone for display
- Status transition validation ensures data integrity throughout the lifecycle
