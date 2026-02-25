# Implementation Plan: School Allocation Menu Planning

## Overview

This implementation plan follows a backend-first approach, building the database layer, models, services, and API endpoints before moving to frontend components. The plan ensures incremental validation through property-based tests and unit tests at each stage, with checkpoints to verify functionality before proceeding.

## Tasks

- [x] 1. Set up database schema and migrations
  - [x] 1.1 Create database migration for menu_item_school_allocations table
    - Add migration file with table creation SQL
    - Include all columns: id, menu_item_id, school_id, portions, date, created_at, updated_at
    - Add foreign key constraints (menu_item_id CASCADE, school_id RESTRICT)
    - Add UNIQUE constraint on (menu_item_id, school_id)
    - Add CHECK constraint for portions > 0
    - Add indexes on menu_item_id, school_id, and date
    - _Requirements: 3.1, 3.2, 3.4, 8.1_
  
  - [ ]* 1.2 Write property test for database constraints
    - **Property 5: Cascade Delete Behavior**
    - **Validates: Requirements 3.3**
  
  - [x] 1.3 Run migration and verify schema
    - Execute migration against development database
    - Verify table structure and constraints
    - _Requirements: 3.1_

- [x] 2. Implement Go models and validation
  - [x] 2.1 Create MenuItemSchoolAllocation model
    - Define struct with GORM tags
    - Add relationships to MenuItem and School
    - Include JSON serialization tags
    - _Requirements: 3.2_
  
  - [x] 2.2 Update MenuItem model to include SchoolAllocations relationship
    - Add SchoolAllocations field with GORM relationship
    - Update JSON serialization
    - _Requirements: 1.3_
  
  - [x] 2.3 Implement ValidateSchoolAllocations service method
    - Check for empty allocations array
    - Validate sum equals total portions
    - Check for duplicate school IDs
    - Validate positive portion counts
    - Return descriptive error messages
    - _Requirements: 2.1, 2.2, 7.1, 8.1, 9.1_
  
  - [ ]* 2.4 Write property test for allocation sum validation
    - **Property 3: Allocation Sum Validation**
    - **Validates: Requirements 2.2**
  
  - [ ]* 2.5 Write property test for empty allocation rejection
    - **Property 11: Empty Allocation Rejection**
    - **Validates: Requirements 7.1**
  
  - [ ]* 2.6 Write property test for duplicate school prevention
    - **Property 12: Duplicate School Prevention**
    - **Validates: Requirements 8.2**
  
  - [ ]* 2.7 Write property test for positive portion validation
    - **Property 13: Positive Portion Validation**
    - **Validates: Requirements 9.1**
  
  - [ ]* 2.8 Write unit tests for ValidateSchoolAllocations
    - Test specific error messages for each validation rule
    - Test edge cases (sum off by one, exactly matching)
    - Test boundary values (portions = 1, large numbers)
    - _Requirements: 2.2, 7.1, 8.1, 9.1_

- [x] 3. Checkpoint - Verify validation logic
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Implement service layer for menu item creation
  - [x] 4.1 Implement CreateMenuItemWithAllocations service method
    - Validate allocations using ValidateSchoolAllocations
    - Verify all school IDs exist in database
    - Create menu item and allocations in transaction
    - Handle transaction rollback on errors
    - Load relationships before returning
    - _Requirements: 1.1, 1.2, 1.4, 2.4, 3.1, 3.2_
  
  - [ ]* 4.2 Write property test for invalid school rejection
    - **Property 2: Invalid School Rejection**
    - **Validates: Requirements 1.4**
  
  - [ ]* 4.3 Write property test for valid allocation round trip
    - **Property 4: Valid Allocation Round Trip**
    - **Validates: Requirements 2.4**
  
  - [ ]* 4.4 Write property test for allocation persistence and multiplicity
    - **Property 1: Allocation Persistence and Multiplicity**
    - **Validates: Requirements 1.2, 1.3**
  
  - [ ]* 4.5 Write unit tests for CreateMenuItemWithAllocations
    - Test transaction rollback on partial failure
    - Test error handling for non-existent schools
    - Test successful creation with multiple allocations
    - _Requirements: 1.1, 1.4, 2.4_

- [x] 5. Implement service layer for menu item updates
  - [x] 5.1 Implement UpdateMenuItemWithAllocations service method
    - Delete existing allocations for menu item
    - Validate new allocations
    - Create new allocations in transaction
    - Handle transaction rollback on errors
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_
  
  - [ ]* 5.2 Write property test for allocation update persistence
    - **Property 10: Allocation Update Persistence**
    - **Validates: Requirements 5.3**
  
  - [ ]* 5.3 Write unit tests for UpdateMenuItemWithAllocations
    - Test replacing allocations with new set
    - Test validation during update
    - Test transaction rollback
    - _Requirements: 5.2, 5.3, 5.5_

- [x] 6. Implement service layer for allocation retrieval
  - [x] 6.1 Implement GetMenuItemWithAllocations service method
    - Query menu item with preloaded allocations
    - Preload school relationships
    - Order allocations by school name
    - _Requirements: 4.2, 4.3, 4.4_
  
  - [x] 6.2 Implement GetAllocationsByDate service method
    - Query allocations for specific date
    - Preload menu item and school relationships
    - Order by school name
    - _Requirements: 4.1, 4.3, 4.4_
  
  - [ ]* 6.3 Write property test for date-based allocation retrieval
    - **Property 6: Date-Based Allocation Retrieval**
    - **Validates: Requirements 4.1**
  
  - [ ]* 6.4 Write property test for menu item allocation retrieval
    - **Property 7: Menu Item Allocation Retrieval**
    - **Validates: Requirements 4.2**
  
  - [ ]* 6.5 Write property test for allocation response completeness
    - **Property 8: Allocation Response Completeness**
    - **Validates: Requirements 4.3**
  
  - [ ]* 6.6 Write property test for alphabetical school ordering
    - **Property 9: Alphabetical School Ordering**
    - **Validates: Requirements 4.4, 11.4**
  
  - [ ]* 6.7 Write unit tests for retrieval methods
    - Test filtering by date
    - Test filtering by menu item
    - Test ordering by school name
    - _Requirements: 4.1, 4.2, 4.4_

- [x] 7. Checkpoint - Verify service layer
  - Ensure all tests pass, ask the user if questions arise.

- [x] 8. Implement API handlers for menu item operations
  - [x] 8.1 Create request/response DTOs for school allocations
    - Define SchoolAllocationInput struct for requests
    - Define SchoolAllocationResponse struct for responses
    - Add validation tags
    - _Requirements: 1.1, 4.3_
  
  - [x] 8.2 Implement POST /api/v1/menu-plans/{id}/items handler
    - Parse request body with school_allocations
    - Call CreateMenuItemWithAllocations service
    - Return 201 Created with allocation data
    - Handle validation errors with 400 Bad Request
    - Return structured error responses
    - _Requirements: 1.1, 1.2, 2.2, 2.3_
  
  - [x] 8.3 Implement PUT /api/v1/menu-plans/{id}/items/{item_id} handler
    - Parse request body with school_allocations
    - Call UpdateMenuItemWithAllocations service
    - Return 200 OK with updated allocation data
    - Handle validation errors with 400 Bad Request
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_
  
  - [x] 8.4 Implement GET /api/v1/menu-plans/{id}/items/{item_id} handler
    - Call GetMenuItemWithAllocations service
    - Return menu item with school_allocations array
    - Include school names in response
    - _Requirements: 4.2, 4.3, 4.4, 6.1, 6.2_
  
  - [ ]* 8.5 Write unit tests for API handlers
    - Test request parsing and validation
    - Test HTTP status codes for each scenario
    - Test response structure and format
    - Test error response format
    - _Requirements: 2.3, 4.3_

- [x] 9. Update KDS service for cooking view
  - [x] 9.1 Modify GetTodayMenu method to include school allocations
    - Preload SchoolAllocations relationship
    - Preload School relationship for allocation data
    - Transform allocations to response format
    - Sort allocations by school name
    - _Requirements: 10.1, 10.2, 10.3, 10.4_
  
  - [x] 9.2 Update RecipeStatus response struct
    - Add SchoolAllocations field
    - Include school_id, school_name, and portions
    - _Requirements: 10.2_
  
  - [ ]* 9.3 Write unit tests for KDS cooking view
    - Test allocation data in response
    - Test school name ordering
    - Test total portions calculation
    - _Requirements: 10.1, 10.3, 10.4_

- [x] 10. Update KDS service for packing view
  - [x] 10.1 Verify packing view includes school allocations
    - Review existing GetPackingAllocations method
    - Ensure school allocations are included in response
    - Verify school ordering is alphabetical
    - _Requirements: 11.1, 11.2, 11.3, 11.4_
  
  - [ ]* 10.2 Write unit tests for KDS packing view
    - Test school grouping
    - Test allocation data per school
    - Test alphabetical school ordering
    - _Requirements: 11.1, 11.2, 11.4_

- [x] 11. Checkpoint - Verify backend integration
  - Ensure all tests pass, ask the user if questions arise.

- [x] 12. Implement frontend menu item form component
  - [x] 12.1 Create SchoolAllocationInput component
    - Create Vue component for school allocation inputs
    - Display list of all schools with portion input fields
    - Implement real-time sum calculation
    - Show allocation summary (allocated / total)
    - Display validation errors
    - Highlight when sum matches total portions
    - _Requirements: 1.1, 2.1, 2.3, 6.2, 6.3_
  
  - [x] 12.2 Integrate SchoolAllocationInput into MenuItemForm
    - Add school allocations section to form
    - Bind allocation data to form model
    - Disable submit button when validation fails
    - Transform allocations to API format on submit
    - _Requirements: 1.1, 2.2, 7.3_
  
  - [ ]* 12.3 Write unit tests for SchoolAllocationInput component
    - Test sum calculation
    - Test validation error display
    - Test form submission with valid/invalid data
    - Test user interaction flows
    - _Requirements: 2.1, 2.3, 7.3_

- [x] 13. Implement frontend menu plan view updates
  - [x] 13.1 Update menu item display to show allocation summary
    - Display total portions for each menu item
    - Show school breakdown with school names and portions
    - Handle menu items without allocations
    - _Requirements: 6.1, 6.2, 6.3, 6.4_
  
  - [x] 13.2 Update menu item edit flow
    - Load existing allocations when editing
    - Populate allocation inputs with current values
    - Allow modification of allocations
    - _Requirements: 5.1_
  
  - [ ]* 13.3 Write unit tests for menu plan view
    - Test allocation summary display
    - Test edit flow with existing allocations
    - Test display when no allocations exist
    - _Requirements: 6.1, 6.2, 6.4_

- [x] 14. Update KDS frontend views
  - [x] 14.1 Update KDS cooking view to display school allocations
    - Add school allocation breakdown to recipe cards
    - Display school name and portions for each allocation
    - Show allocations ordered by school name
    - _Requirements: 10.1, 10.2, 10.4_
  
  - [x] 14.2 Verify KDS packing view displays school allocations
    - Review existing packing view implementation
    - Ensure school allocations are displayed correctly
    - Verify alphabetical school ordering
    - _Requirements: 11.1, 11.2, 11.4_
  
  - [ ]* 14.3 Write unit tests for KDS views
    - Test cooking view allocation display
    - Test packing view school grouping
    - Test school name ordering
    - _Requirements: 10.1, 10.4, 11.1, 11.4_

- [x] 15. Integration and end-to-end testing
  - [x] 15.1 Test complete workflow: create menu item with allocations
    - Create menu plan
    - Add menu item with school allocations
    - Verify allocations saved correctly
    - Retrieve menu item and verify allocations
    - _Requirements: 1.1, 1.2, 2.4, 3.1, 3.2_
  
  - [x] 15.2 Test complete workflow: update menu item allocations
    - Edit existing menu item
    - Modify school allocations
    - Verify updated allocations saved correctly
    - _Requirements: 5.1, 5.2, 5.3, 5.5_
  
  - [x] 15.3 Test complete workflow: delete menu item with allocations
    - Create menu item with allocations
    - Delete menu item
    - Verify allocations are cascade deleted
    - _Requirements: 3.3_
  
  - [x] 15.4 Test KDS integration: cooking view displays allocations
    - Create menu items with allocations
    - View cooking view for the date
    - Verify school allocations displayed correctly
    - _Requirements: 10.1, 10.2, 10.3, 10.4_
  
  - [x] 15.5 Test KDS integration: packing view displays allocations
    - Create menu items with allocations
    - View packing view for the date
    - Verify school-grouped allocations displayed correctly
    - _Requirements: 11.1, 11.2, 11.3, 11.4_
  
  - [ ]* 15.6 Write integration tests for complete workflows
    - Test create → retrieve → update → delete flow
    - Test concurrent operations
    - Test error scenarios
    - _Requirements: 1.1, 2.4, 3.3, 5.5_

- [x] 16. Final checkpoint - Verify all functionality
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at key milestones
- Property tests validate universal correctness properties across randomized inputs
- Unit tests validate specific examples, edge cases, and error conditions
- Backend implementation is completed before frontend to ensure API stability
- All database operations use transactions to ensure data consistency
- Validation is implemented at multiple layers (client, API, service, database)
