# Implementation Plan: Activity Tracker (Aktivitas Pelacakan)

## Overview

This implementation plan covers the development of the Activity Tracker (Aktivitas Pelacakan) module, a standalone monitoring component that provides Kepala SPPG with real-time visibility into the 16-stage lifecycle of menu orders from initiation through preparation, cooking, packing, delivery, ompreng collection, and cleaning. The implementation follows a backend-first approach, establishing data models and services before building the frontend vertical timeline visualization components.

## Tasks

- [x] 1. Set up backend data models and database migrations
  - Add `current_stage` field to DeliveryRecord model (integer, default 1)
  - Add `stage` field to StatusTransition model (integer, not null)
  - Add `media_url` and `media_type` fields to StatusTransition model
  - Create database migration for new fields
  - Create database indexes for performance optimization on delivery_date, current_status, and current_stage fields
  - Add validation for the 16 status values and stage numbers (1-16) in DeliveryRecord model
  - _Requirements: 1.1-1.8, 2.1-2.4, 3.1-3.3, 4.1-4.4, 5.1-5.5_

- [ ]* 1.1 Write property test for status transition recording
  - **Property 4: Stage Transition Recording**
  - **Validates: Requirements 2.1-5.5**

- [x] 2. Implement ActivityTrackerService core methods
  - [x] 2.1 Implement GetOrdersByDate method
    - Query orders by date with preloaded school, driver, and menu item relations
    - Support optional school_id filter and search query
    - Return order list with summary statistics (total count, status distribution)
    - _Requirements: 8.1-8.7, 9.1-9.8, 15.1-15.6_
  
  - [ ]* 2.2 Write property test for date filter accuracy
    - **Property 11: Date Filter Accuracy**
    - **Validates: Requirements 8.2_
  
  - [x] 2.3 Implement GetOrderDetails method
    - Query single order with all relations
    - Build timeline array with all 16 stages including status, timestamps, and media
    - Calculate is_completed flag for each stage based on current_stage
    - _Requirements: 1.1-1.8, 6.1-6.7, 7.1-7.7_
  
  - [ ]* 2.4 Write property test for timeline ordering
    - **Property 7: Chronological Timeline Ordering**
    - **Validates: Requirements 1.1_
  
  - [x] 2.5 Implement UpdateOrderStatus method
    - Create status transition record with timestamp, stage number, and user info
    - Update order current_status and current_stage fields
    - Validate status transition sequence and log warnings for skipped stages
    - Use database transaction for atomicity
    - _Requirements: 16.1-16.5_
  
  - [x] 2.6 Implement AttachStageMedia method
    - Accept photo or video file upload
    - Store media in cloud storage (Firebase Storage or S3)
    - Update StatusTransition record with media_url and media_type
    - Generate thumbnail for videos
    - _Requirements: 7.3-7.6_
  
  - [ ]* 2.7 Write unit tests for UpdateOrderStatus
    - Test valid sequential transitions
    - Test stage skip warning generation
    - Test manual correction audit trail
    - Test concurrent update handling
    - _Requirements: 16.1-16.5_

- [x] 3. Implement Firebase synchronization
  - [x] 3.1 Implement SyncOrderToFirebase method
    - Write order record to Firebase at /order_tracking/{date}/{order_id}
    - Write status transitions to Firebase at /status_transitions/{date}/{order_id}
    - Include media URLs in Firebase data
    - Handle Firebase connection errors gracefully (log and continue)
    - _Requirements: 10.1, 10.2, 10.3_
  
  - [x] 3.2 Add Firebase sync calls to UpdateOrderStatus
    - Call SyncOrderToFirebase after successful database update
    - Implement retry queue for failed Firebase writes
    - _Requirements: 10.1, 10.2_
  
  - [ ]* 3.3 Write unit tests for Firebase synchronization
    - Test successful sync with mock Firebase client
    - Test error handling when Firebase unavailable
    - Test retry queue functionality
    - _Requirements: 10.3, 10.4_

- [x] 4. Implement KDS and Logistics module integration
  - [x] 4.1 Implement HandleKDSStatusUpdate method
    - Map KDS Cooking statuses to Activity Tracker stages (cooking → order_dimasak/stage 2, ready → order_dikemas/stage 3, packing_completed → order_siap_diambil/stage 4)
    - Call UpdateOrderStatus with mapped status and stage number
    - _Requirements: 11.1-11.3_
  
  - [x] 4.2 Implement HandleLogisticsStatusUpdate method
    - Map Logistics statuses to Activity Tracker stages (driver_departed → pesanan_dalam_perjalanan/stage 5, driver_arrived → pesanan_sudah_tiba/stage 6, delivery_confirmed → pesanan_sudah_diterima/stage 7, etc.)
    - Handle ompreng collection stages (8-11)
    - Call UpdateOrderStatus with mapped status and stage number
    - _Requirements: 12.1-12.7_
  
  - [ ]* 4.3 Write property test for status mapping
    - **Property 13: KDS and Logistics Status Mapping Correctness**
    - **Validates: Requirements 11.1-12.7**
  
  - [ ]* 4.4 Write unit tests for HandleKDSStatusUpdate and HandleLogisticsStatusUpdate
    - Test each module status mapping
    - Test error handling for invalid statuses
    - _Requirements: 11.1-13.5_

- [x] 5. Checkpoint - Ensure backend tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Implement ActivityTrackerHandler API endpoints
  - [x] 6.1 Implement GET /api/activity-tracker/orders endpoint
    - Parse and validate date query parameter (YYYY-MM-DD format)
    - Parse optional school_id and search query parameters
    - Call GetOrdersByDate service method
    - Return JSON response with orders and summary
    - Handle errors (invalid date format, database errors)
    - _Requirements: 8.1-8.7, 9.1-9.8, 15.1-15.6_
  
  - [x] 6.2 Implement GET /api/activity-tracker/orders/:id endpoint
    - Parse and validate order ID path parameter
    - Call GetOrderDetails service method
    - Return JSON response with order details and vertical timeline data
    - Handle errors (order not found, database errors)
    - _Requirements: 1.1-1.8, 6.1-6.7, 7.1-7.7_
  
  - [x] 6.3 Implement PUT /api/activity-tracker/orders/:id/status endpoint
    - Parse and validate request body (new_status, stage, notes)
    - Verify user has Kepala_SPPG role
    - Call UpdateOrderStatus service method
    - Return JSON response with updated order
    - Handle errors (unauthorized, invalid status, concurrent update)
    - _Requirements: 14.1-14.4, 16.1-16.5_
  
  - [x] 6.4 Implement POST /api/activity-tracker/orders/:id/stages/:stage/media endpoint
    - Parse multipart form data (media file, media_type)
    - Validate file type (image or video)
    - Upload file to cloud storage
    - Call AttachStageMedia service method
    - Return JSON response with media URLs
    - Handle errors (invalid file type, upload failure)
    - _Requirements: 7.3-7.6_
  
  - [ ]* 6.5 Write unit tests for API handlers
    - Test request parsing and validation
    - Test authentication and authorization
    - Test error response formatting
    - Test file upload handling
    - _Requirements: 8.1-8.7, 14.1-14.4_

- [x] 7. Add API routes and middleware
  - Register Activity Tracker routes in main router under /api/activity-tracker
  - Add authentication middleware to all Activity Tracker routes
  - Add role-based authorization middleware (Kepala_SPPG or management roles)
  - Add file upload middleware for media endpoint
  - _Requirements: 14.1-14.4_

- [ ]* 7.1 Write property test for role-based access control
  - **Property 14: Role-Based Access Control**
  - **Validates: Requirements 14.1-14.4**

- [x] 8. Implement frontend ActivityTrackerListView component
  - [x] 8.1 Create ActivityTrackerListView.vue component structure
    - Add date picker for date filter with default to current date
    - Add school filter dropdown
    - Add search input box
    - Add order cards grid with loading state
    - Add empty state for no orders
    - Implement fetchOrders method to call GET /api/activity-tracker/orders
    - _Requirements: 8.1-8.7, 9.1-9.8, 15.1-15.6_
  
  - [x] 8.2 Implement order card component
    - Display menu photo thumbnail
    - Display menu name and school name
    - Display current stage status with color-coded badge
    - Display portion quantity
    - Add click handler to navigate to detail view
    - _Requirements: 9.1-9.8_
  
  - [x] 8.3 Add summary statistics display
    - Display total order count for selected date
    - Display status distribution (e.g., "5 sedang dimasak, 3 dalam perjalanan")
    - _Requirements: 15.4-15.5_
  
  - [ ]* 8.4 Write unit tests for ActivityTrackerListView
    - Test date filter interaction
    - Test school filter interaction
    - Test search functionality
    - Test empty state rendering
    - Test navigation to detail view
    - _Requirements: 8.1-8.7, 9.1-9.8, 15.1-15.6_

- [x] 9. Implement ActivityTrackerDetailView component
  - [x] 9.1 Create ActivityTrackerDetailView.vue component structure
    - Add order header section with photo, menu name, school, portions
    - Add back button to return to list view
    - Add VerticalTimeline component
    - Implement fetchOrderDetails method to call GET /api/activity-tracker/orders/:id
    - _Requirements: 1.1-1.8, 6.1-6.7_
  
  - [x] 9.2 Implement order header display
    - Display menu photo (large)
    - Display "Aktivitas Pelacakan" title
    - Display menu name
    - Display school name and delivery date
    - Display portion quantity
    - Display driver name and vehicle info
    - Display current order status badge
    - _Requirements: 6.1-6.7_
  
  - [ ]* 9.3 Write unit tests for ActivityTrackerDetailView
    - Test order header rendering
    - Test back button navigation
    - Test loading state
    - Test error handling
    - _Requirements: 1.1-1.8, 6.1-6.7_

- [x] 10. Implement VerticalTimeline component
  - [x] 10.1 Create VerticalTimeline.vue component with 16-stage display
    - Define all 16 stages with Indonesian labels and descriptions
    - Render all stages in sequential vertical order
    - Add connecting vertical lines between stages
    - Pass stage data to TimelineStage components
    - _Requirements: 1.1-1.8, 2.1-2.4, 3.1-3.3, 4.1-4.4, 5.1-5.5_
  
  - [ ]* 10.2 Write property test for stage display
    - **Property 1: Sequential Stage Display**
    - **Validates: Requirements 1.1**
  
  - [ ]* 10.3 Write unit tests for VerticalTimeline
    - Test all 16 stages rendered
    - Test vertical layout
    - Test connecting lines
    - _Requirements: 1.1-1.8_

- [x] 11. Implement TimelineStage component
  - [x] 11.1 Create TimelineStage.vue component structure
    - Display stage indicator circle (filled for completed, empty for pending, blue for in-progress)
    - Display stage title in Indonesian
    - Display stage description
    - Display timestamp in format "Rabu, 13:49 - Rabu, 13:50" for completed stages
    - Display photo thumbnail if media_type is "photo"
    - Display video play button if media_type is "video"
    - Add click handlers to open media in modal
    - _Requirements: 1.2-1.8, 7.1-7.7_
  
  - [x] 11.2 Implement status indicator logic
    - Completed: filled green circle
    - In-progress: filled blue circle
    - Pending: empty gray circle
    - Apply appropriate CSS classes based on stage status
    - _Requirements: 1.2-1.4_
  
  - [x] 11.3 Implement timestamp formatting
    - Format timestamps to Asia/Jakarta timezone (WIB)
    - Use format "Day, HH:MM - Day, HH:MM" (e.g., "Rabu, 13:49 - Rabu, 13:50")
    - Display only for completed stages
    - _Requirements: 1.6_
  
  - [x] 11.4 Implement media display
    - Show photo thumbnail with click to enlarge
    - Show video thumbnail with play button overlay
    - Open media viewer modal on click
    - Support fullscreen photo view
    - Support video playback in modal
    - _Requirements: 7.3-7.6_
  
  - [ ]* 11.5 Write property test for status indicator correctness
    - **Property 2: Status Indicator Correctness**
    - **Validates: Requirements 1.2-1.4**
  
  - [ ]* 11.6 Write unit tests for TimelineStage
    - Test stage rendering with various statuses
    - Test timestamp display and formatting
    - Test Indonesian label display
    - Test media thumbnail display
    - Test media modal interaction
    - _Requirements: 1.1-1.8, 7.1-7.7_

- [x] 12. Implement Firebase real-time updates in frontend
  - [x] 12.1 Add Firebase listener setup in ActivityTrackerListView
    - Initialize Firebase connection in mounted hook
    - Subscribe to /order_tracking/{date} path
    - Update order list when Firebase events received
    - _Requirements: 10.1, 10.2, 10.3_
  
  - [x] 12.2 Add Firebase listener setup in ActivityTrackerDetailView
    - Initialize Firebase connection in mounted hook
    - Subscribe to /order_tracking/{date}/{order_id} path
    - Subscribe to /status_transitions/{date}/{order_id} path
    - Update timeline when Firebase events received
    - _Requirements: 10.1, 10.2_
  
  - [x] 12.3 Implement handleStatusUpdate method
    - Update local order data when Firebase event received
    - Update timeline visualization within 2 seconds
    - Show notification badge for updates on other orders
    - _Requirements: 10.1, 10.2, 10.5_
  
  - [x] 12.4 Add Firebase connection error handling
    - Display banner notification when connection lost
    - Implement automatic reconnection every 5 seconds
    - Fall back to polling API every 30 seconds if Firebase unavailable
    - Remove banner when connection restored
    - _Requirements: 10.4_
  
  - [ ]* 12.5 Write unit tests for Firebase integration
    - Test listener setup and teardown
    - Test status update handling with mock Firebase events
    - Test connection error handling and reconnection
    - _Requirements: 10.1-10.5_

- [x] 13. Checkpoint - Ensure frontend tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 14. Implement error handling and edge cases
  - [x] 14.1 Add frontend error handling
    - Display toast notification for network errors
    - Implement automatic retry with exponential backoff (max 3 attempts)
    - Show loading states during operations
    - Display empty state when no orders found
    - Handle media upload errors
    - _Requirements: 8.1-8.7, 15.1-15.6_
  
  - [x] 14.2 Add backend error handling
    - Return appropriate HTTP status codes (400, 403, 404, 409, 413, 503)
    - Format error responses consistently
    - Log errors with context for monitoring
    - Implement database transaction rollback on errors
    - Handle file upload size limits
    - _Requirements: 14.3-14.4, 16.1-16.5_
  
  - [ ]* 14.3 Write unit tests for error scenarios
    - Test network request failures
    - Test unauthorized access
    - Test invalid date format
    - Test order not found
    - Test concurrent update conflicts
    - Test file upload failures
    - _Requirements: 14.3-14.4, 16.1-16.5_

- [x] 15. Add route configuration and navigation
  - Add Activity Tracker routes to Vue Router:
    - /activity-tracker (list view)
    - /activity-tracker/:id (detail view)
  - Add navigation link in main sidebar as standalone module
  - Add "Aktivitas Pelacakan" icon and label
  - Restrict route access to Kepala_SPPG and management roles
  - _Requirements: 14.1-14.4_

- [x] 16. Integration testing and validation
  - [x] 16.1 Test complete order lifecycle flow
    - Create test order record
    - Trigger status transitions through all 16 stages
    - Verify timeline updates correctly in UI
    - Verify vertical timeline displays all stages with correct status indicators
    - Verify timestamps display in correct format
    - _Requirements: 1.1-1.8, 2.1-5.5_
  
  - [x] 16.2 Test KDS and Logistics module integration
    - Trigger status updates from KDS Cooking module
    - Trigger status updates from Logistics/Delivery module
    - Trigger status updates from KDS Cleaning module
    - Verify Activity Tracker reflects updates within 2 seconds
    - Verify correct stage numbers assigned
    - _Requirements: 11.1-13.5_
  
  - [x] 16.3 Test real-time Firebase synchronization
    - Update order status via API
    - Verify Firebase receives update
    - Verify frontend receives real-time update in both list and detail views
    - Test with multiple concurrent users
    - _Requirements: 10.1-10.5_
  
  - [x] 16.4 Test media upload and display
    - Upload photo to a stage
    - Upload video to a stage
    - Verify media displays in timeline
    - Verify thumbnail generation
    - Verify media viewer modal functionality
    - _Requirements: 7.3-7.7_
  
  - [ ]* 16.5 Write integration tests
    - Test end-to-end flow from API to UI
    - Test Firebase real-time updates with test Firebase instance
    - Test role-based access control
    - Test media upload and retrieval
    - _Requirements: 10.1-10.5, 14.1-14.4_

- [x] 17. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Backend uses Go with Gin framework and GORM for database operations
- Frontend uses Vue 3 with Composition API and Ant Design Vue components
- Property tests use gopter (Go) and fast-check (JavaScript) libraries
- All property tests require minimum 100 iterations
- Firebase is used for real-time synchronization but PostgreSQL remains the source of truth
- The module is a standalone module (not sub-module of Dashboard or Logistics) with its own navigation entry
- The module integrates with existing KDS modules (Cooking, Packing, Cleaning) and Logistics/Delivery operations via status transition events
- UI follows vertical timeline design pattern similar to delivery tracking apps
- Media files (photos/videos) are stored in cloud storage (Firebase Storage or S3)
- Timestamps are displayed in Asia/Jakarta timezone (WIB) with format "Day, HH:MM - Day, HH:MM"
