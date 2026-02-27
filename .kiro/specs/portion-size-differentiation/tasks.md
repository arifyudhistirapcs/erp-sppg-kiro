# Tasks: Portion Size Differentiation

## Phase 1: Database Schema Updates

### Task 1.1: Add portion_size field to menu_item_school_allocations table
- [x] 1.1.1 Create database migration to add `portion_size` VARCHAR(10) field
- [x] 1.1.2 Add CHECK constraint to ensure portion_size IN ('small', 'large')
- [x] 1.1.3 Create index on portion_size field for query performance
- [x] 1.1.4 Test migration on development database

### Task 1.2: Migrate existing allocation records
- [x] 1.2.1 Write migration script to set portion_size = 'large' for all existing records
- [x] 1.2.2 Verify all existing records have portion_size values
- [x] 1.2.3 Add NOT NULL constraint to portion_size field
- [x] 1.2.4 Test rollback procedure for migration

### Task 1.3: Update database indexes
- [x] 1.3.1 Add composite index on (menu_item_id, school_id, portion_size)
- [x] 1.3.2 Verify query performance with EXPLAIN on common queries
- [x] 1.3.3 Document index usage in database schema documentation

## Phase 2: Backend Model and Service Updates

### Task 2.1: Update MenuItemSchoolAllocation model
- [x] 2.1.1 Add PortionSize field to MenuItemSchoolAllocation struct
- [x] 2.1.2 Add validation tag for portion_size field (oneof=small large)
- [x] 2.1.3 Update JSON tags for API serialization
- [x] 2.1.4 Add unit tests for model validation

### Task 2.2: Implement DetermineSchoolPortionType function
- [x] 2.2.1 Create function in school service to determine portion type based on category
- [x] 2.2.2 Return 'mixed' for SD schools, 'large' for SMP/SMA schools
- [x] 2.2.3 Add unit tests for all school categories
- [x] 2.2.4 Document function behavior and return values

### Task 2.3: Implement ValidatePortionSizeAllocations function
- [x] 2.3.1 Create validation function in menu planning service
- [x] 2.3.2 Validate sum of portions_small + portions_large equals total_portions
- [x] 2.3.3 Validate SMP/SMA schools have portions_small = 0
- [x] 2.3.4 Validate at least one portion type > 0 for each school
- [x] 2.3.5 Validate non-negative portion counts
- [x] 2.3.6 Add comprehensive unit tests for all validation scenarios
- [x] 2.3.7 Add property-based tests for validation logic

### Task 2.4: Update CreateMenuItemWithAllocations function
- [x] 2.4.1 Modify function to accept portions_small and portions_large for each school
- [x] 2.4.2 Create separate allocation records for small and large portions
- [x] 2.4.3 For SD schools: create up to 2 records (small and large if both > 0)
- [x] 2.4.4 For SMP/SMA schools: create 1 record with portion_size = 'large'
- [x] 2.4.5 Ensure transaction atomicity for all allocation creations
- [x] 2.4.6 Add unit tests for SD school dual allocation
- [x] 2.4.7 Add unit tests for SMP/SMA school single allocation
- [x] 2.4.8 Add integration tests for complete workflow

### Task 2.5: Implement GetSchoolAllocationsWithPortionSizes function
- [x] 2.5.1 Create function to retrieve allocations grouped by school
- [x] 2.5.2 Combine multiple allocation records for same school into single display object
- [x] 2.5.3 Populate portions_small and portions_large fields correctly
- [x] 2.5.4 Sort results alphabetically by school name
- [x] 2.5.5 Include school category in response
- [x] 2.5.6 Add unit tests for grouping logic
- [x] 2.5.7 Add integration tests with database queries

### Task 2.6: Update UpdateMenuItemAllocations function
- [x] 2.6.1 Modify function to handle portion size updates
- [x] 2.6.2 Delete existing allocation records for the menu item
- [x] 2.6.3 Create new allocation records with updated portion sizes
- [x] 2.6.4 Use transaction to ensure atomic update
- [x] 2.6.5 Add unit tests for update scenarios
- [x] 2.6.6 Add rollback tests for failed updates

## Phase 3: API Endpoint Updates

### Task 3.1: Update CreateMenuItem API endpoint
- [x] 3.1.1 Modify request payload to accept portions_small and portions_large
- [x] 3.1.2 Update request validation to check portion size fields
- [x] 3.1.3 Call ValidatePortionSizeAllocations before processing
- [x] 3.1.4 Return detailed error messages for validation failures
- [x] 3.1.5 Update API documentation with new request format
- [x] 3.1.6 Add API integration tests for valid requests
- [x] 3.1.7 Add API integration tests for invalid requests

### Task 3.2: Update GetMenuItem API endpoint
- [x] 3.2.1 Modify response to include portions_small and portions_large
- [x] 3.2.2 Call GetSchoolAllocationsWithPortionSizes for data retrieval
- [x] 3.2.3 Format response with grouped allocations
- [x] 3.2.4 Update API documentation with new response format
- [x] 3.2.5 Add API integration tests for response structure

### Task 3.3: Update UpdateMenuItem API endpoint
- [x] 3.3.1 Modify request payload to accept portion size updates
- [x] 3.3.2 Validate updated allocations using same rules as creation
- [x] 3.3.3 Call UpdateMenuItemAllocations service function
- [x] 3.3.4 Return updated menu item with portion sizes
- [x] 3.3.5 Add API integration tests for update scenarios

### Task 3.4: Update KDS Cooking API endpoint
- [x] 3.4.1 Modify GetTodayMenu to include portion size information
- [x] 3.4.2 Group allocations by school with portion size breakdown
- [x] 3.4.3 Include labels for small and large portions in response
- [x] 3.4.4 Update API documentation
- [x] 3.4.5 Add integration tests for KDS cooking view

### Task 3.5: Update KDS Packing API endpoint
- [x] 3.5.1 Modify GetPackingAllocations to include portion sizes
- [x] 3.5.2 Display portions_small and portions_large for each school
- [x] 3.5.3 Update response format with portion size labels
- [x] 3.5.4 Update API documentation
- [x] 3.5.5 Add integration tests for KDS packing view

## Phase 4: Frontend Menu Planning UI Updates

### Task 4.1: Update MenuItemForm component
- [x] 4.1.1 Add portions_small and portions_large input fields for each school
- [x] 4.1.2 Conditionally display fields based on school category
- [x] 4.1.3 Show only large portions field for SMP/SMA schools
- [x] 4.1.4 Show both fields for SD schools with appropriate labels
- [x] 4.1.5 Display student count context next to input fields
- [x] 4.1.6 Add unit tests for component rendering

### Task 4.2: Implement real-time validation in UI
- [x] 4.2.1 Calculate sum of all portions as user types
- [x] 4.2.2 Display running total vs target total portions
- [x] 4.2.3 Show error message when sum doesn't match
- [x] 4.2.4 Show success indicator when sum matches
- [x] 4.2.5 Disable submit button when validation fails
- [x] 4.2.6 Add unit tests for validation logic

### Task 4.3: Update allocation display in menu plan view
- [x] 4.3.1 Show portion size breakdown for each school in summary
- [x] 4.3.2 Display "Small: X, Large: Y" format for SD schools
- [x] 4.3.3 Display "Large: X" format for SMP/SMA schools
- [x] 4.3.4 Add visual indicators for portion sizes (icons or colors)
- [x] 4.3.5 Add unit tests for display formatting

### Task 4.4: Implement portion size statistics display
- [x] 4.4.1 Calculate total small portions across all SD schools
- [x] 4.4.2 Calculate total large portions across all schools
- [x] 4.4.3 Display percentage breakdown of small vs large
- [x] 4.4.4 Show count of schools by portion size type
- [x] 4.4.5 Update statistics in real-time as allocations change
- [x] 4.4.6 Add unit tests for statistics calculations

### Task 4.5: Update menu item edit form
- [x] 4.5.1 Load existing portion size allocations when editing
- [x] 4.5.2 Pre-populate portions_small and portions_large fields
- [x] 4.5.3 Allow modification of portion sizes
- [x] 4.5.4 Validate changes before submission
- [x] 4.5.5 Add unit tests for edit functionality

## Phase 5: Frontend KDS UI Updates

### Task 5.1: Update KDS Cooking View component
- [x] 5.1.1 Display portion size breakdown for each recipe
- [x] 5.1.2 Show school allocations with small and large portion labels
- [x] 5.1.3 Format display as "School Name: Small (X), Large (Y)"
- [x] 5.1.4 Add visual distinction between portion sizes
- [x] 5.1.5 Update component unit tests

### Task 5.2: Update KDS Packing View component
- [x] 5.2.1 Display portion sizes for each school allocation
- [x] 5.2.2 Show separate rows or columns for small and large portions
- [x] 5.2.3 Add labels indicating grade levels (1-3 vs 4-6)
- [x] 5.2.4 Ensure clear visual hierarchy for portion sizes
- [x] 5.2.5 Update component unit tests

### Task 5.3: Update Firebase sync for KDS views
- [x] 5.3.1 Include portion_size field in Firebase data structure
- [x] 5.3.2 Update real-time listeners to handle portion size data
- [x] 5.3.3 Test real-time updates with portion size changes
- [x] 5.3.4 Verify data consistency across all KDS views

## Phase 6: Testing and Validation

### Task 6.1: Backend unit tests
- [x] 6.1.1 Test DetermineSchoolPortionType for all categories
- [x] 6.1.2 Test ValidatePortionSizeAllocations with valid inputs
- [x] 6.1.3 Test ValidatePortionSizeAllocations with invalid inputs
- [x] 6.1.4 Test CreateMenuItemWithAllocations for SD schools
- [x] 6.1.5 Test CreateMenuItemWithAllocations for SMP/SMA schools
- [x] 6.1.6 Test GetSchoolAllocationsWithPortionSizes grouping logic
- [x] 6.1.7 Test transaction rollback scenarios

### Task 6.2: Backend integration tests
- [x] 6.2.1 Test complete workflow: create → retrieve → update → delete
- [x] 6.2.2 Test concurrent allocation creation
- [x] 6.2.3 Test database constraint enforcement
- [x] 6.2.4 Test cascade delete behavior
- [x] 6.2.5 Test query performance with large datasets

### Task 6.3: Backend property-based tests
- [x] 6.3.1 Property: Allocation sum always equals total portions
- [x] 6.3.2 Property: SD schools always have 0-2 allocation records
- [x] 6.3.3 Property: SMP/SMA schools always have exactly 1 allocation record
- [x] 6.3.4 Property: Retrieved allocations match created allocations
- [x] 6.3.5 Property: Alphabetical ordering is maintained

### Task 6.4: Frontend unit tests
- [x] 6.4.1 Test MenuItemForm rendering for SD schools
- [x] 6.4.2 Test MenuItemForm rendering for SMP/SMA schools
- [x] 6.4.3 Test real-time validation calculations
- [x] 6.4.4 Test error message display
- [x] 6.4.5 Test submit button enable/disable logic
- [x] 6.4.6 Test statistics calculations

### Task 6.5: End-to-end tests
- [x] 6.5.1 Test creating menu item with mixed portion sizes
- [x] 6.5.2 Test editing existing menu item allocations
- [x] 6.5.3 Test viewing allocations in KDS cooking view
- [x] 6.5.4 Test viewing allocations in KDS packing view
- [x] 6.5.5 Test validation error scenarios
- [x] 6.5.6 Test real-time updates across multiple clients

## Phase 7: Documentation and Deployment

### Task 7.1: Update API documentation
- [x] 7.1.1 Document new request payload format with portion sizes
- [x] 7.1.2 Document new response format with portion size breakdown
- [x] 7.1.3 Document validation rules and error messages
- [x] 7.1.4 Add example requests and responses
- [ ] 7.1.5 Update Postman collection with new endpoints

### Task 7.2: Update user documentation
- [x] 7.2.1 Create user guide for portion size allocation
- [x] 7.2.2 Document SD school allocation process (two portion sizes)
- [x] 7.2.3 Document SMP/SMA school allocation process (one portion size)
- [x] 7.2.4 Add screenshots of updated UI
- [x] 7.2.5 Create FAQ for common questions

### Task 7.3: Database migration planning
- [x] 7.3.1 Create migration runbook with step-by-step instructions
- [x] 7.3.2 Document rollback procedure
- [x] 7.3.3 Estimate migration time for production database
- [x] 7.3.4 Plan maintenance window for migration
- [x] 7.3.5 Create backup and restore procedures

### Task 7.4: Deployment preparation
- [-] 7.4.1 Test migration on staging environment
- [ ] 7.4.2 Verify all existing allocations are migrated correctly
- [ ] 7.4.3 Test backward compatibility with existing data
- [ ] 7.4.4 Prepare deployment checklist
- [ ] 7.4.5 Schedule deployment with stakeholders

### Task 7.5: Post-deployment validation
- [ ] 7.5.1 Verify database migration completed successfully
- [ ] 7.5.2 Test creating new menu items with portion sizes
- [ ] 7.5.3 Test editing existing menu items
- [ ] 7.5.4 Verify KDS views display portion sizes correctly
- [ ] 7.5.5 Monitor error logs for any issues
- [ ] 7.5.6 Collect user feedback on new feature
