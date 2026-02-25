# Implementation Plan: KDS Date Filtering

## Overview

This implementation plan breaks down the KDS Date Filtering feature into discrete coding tasks. The approach follows an incremental strategy: backend first (API and services), then frontend (components and views), followed by integration and testing. Each task builds on previous work, ensuring no orphaned code.

## Tasks

- [x] 1. Backend: Add date parameter parsing and validation to KDS handlers
  - Modify `backend/internal/handlers/kds_handler.go`
  - Create `parseDateParameter()` function to extract and validate date from query parameter
  - Add date format validation (YYYY-MM-DD)
  - Return 400 Bad Request for invalid formats with appropriate error messages
  - Default to current date if parameter is missing
  - _Requirements: 2.1, 2.3, 6.1, 6.2, 6.3, 6.4, 6.5, 8.1, 8.5_

- [ ]* 1.1 Write property test for date parameter validation
  - **Property 2: Invalid date rejection**
  - **Validates: Requirements 2.3, 6.4, 6.5, 8.1, 8.5**
  - Generate random invalid date strings and verify all return 400 errors
  - Use gopter framework with minimum 100 iterations

- [x] 2. Backend: Modify KDS service to accept date parameters
  - Modify `backend/internal/services/kds_service.go`
  - Update `GetTodayMenu()` to accept `date time.Time` parameter
  - Update `SyncTodayMenuToFirebase()` to accept `date time.Time` parameter
  - Replace `time.Now()` calls with provided date parameter
  - Use `DATE(menu_items.date) = DATE(?)` for date comparison in queries
  - Implement timezone normalization using `normalizeDate()` helper function
  - _Requirements: 4.1, 4.2, 4.5, 7.1, 7.3, 7.4_

- [ ]* 2.1 Write property test for cooking endpoint date filtering
  - **Property 3: Cooking endpoint date filtering**
  - **Validates: Requirements 4.2, 7.3**
  - Generate random dates with test data and verify returned recipes match the requested date
  - Use gopter framework with minimum 100 iterations

- [ ]* 2.2 Write property test for timezone consistency
  - **Property 5: Timezone consistency**
  - **Validates: Requirements 4.5, 7.4**
  - Generate random dates and verify same date produces same results across multiple calls
  - Use gopter framework with minimum 100 iterations

- [x] 3. Backend: Modify packing allocation service to accept date parameters
  - Modify `backend/internal/services/packing_allocation_service.go`
  - Update `GetPackingAllocations()` to accept `date time.Time` parameter
  - Update `CalculatePackingAllocations()` to accept `date time.Time` parameter
  - Replace `time.Now()` calls with provided date parameter
  - Calculate startOfDay and endOfDay based on provided date
  - Query delivery_tasks table with date range filter
  - _Requirements: 4.3, 4.5, 7.2, 7.3, 7.4_

- [ ]* 3.1 Write property test for packing endpoint date filtering
  - **Property 4: Packing endpoint date filtering**
  - **Validates: Requirements 4.3, 7.3**
  - Generate random dates with test data and verify returned allocations match the requested date
  - Use gopter framework with minimum 100 iterations

- [x] 4. Backend: Wire handler to service layer with date parameter
  - Update `GetCookingToday()` handler to call `parseDateParameter()`
  - Pass parsed date to `GetTodayMenu()` service method
  - Update `GetPackingToday()` handler to call `parseDateParameter()`
  - Pass parsed date to `GetPackingAllocations()` service method
  - Handle validation errors and return appropriate HTTP responses
  - _Requirements: 2.2, 6.1, 6.2, 8.2_

- [ ]* 4.1 Write property test for valid date acceptance
  - **Property 1: Valid date acceptance and filtering**
  - **Validates: Requirements 2.1, 2.2, 7.3**
  - Generate random valid dates and verify API accepts them and filters correctly
  - Use gopter framework with minimum 100 iterations

- [ ]* 4.2 Write property test for future date handling
  - **Property 6: Future date handling**
  - **Validates: Requirements 2.5**
  - Generate random future dates and verify returns 200 with empty or planned data
  - Use gopter framework with minimum 100 iterations

- [ ]* 4.3 Write property test for backward compatibility
  - **Property 7: Backward compatibility**
  - **Validates: Requirements 6.3**
  - Test multiple requests without date parameter and verify returns current date data
  - Use gopter framework with minimum 100 iterations

- [ ]* 4.4 Write property test for empty data handling
  - **Property 8: Empty data handling**
  - **Validates: Requirements 4.4, 7.5**
  - Generate random dates with no data and verify returns 200 with empty array
  - Use gopter framework with minimum 100 iterations

- [ ]* 4.5 Write unit tests for backend handlers and services
  - Test date parameter parsing with valid and invalid formats
  - Test default date behavior when parameter omitted
  - Test timezone conversion correctness
  - Test error response formatting
  - Test edge cases: leap years, month boundaries, year boundaries

- [x] 5. Checkpoint - Ensure backend tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Frontend: Create KDSDatePicker component
  - Create new file `web/src/components/KDSDatePicker.vue`
  - Implement date picker using Ant Design Vue DatePicker component
  - Add props: modelValue (Date), loading (Boolean), disabled (Boolean)
  - Emit events: 'update:modelValue', 'change'
  - Add "Today" quick action button
  - Implement keyboard navigation support (arrow keys, Enter, Escape)
  - Add visual feedback during loading state
  - Display currently selected date prominently
  - Implement session storage persistence for selected date
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 3.4, 3.5_

- [ ]* 6.1 Write unit tests for KDSDatePicker component
  - Test component rendering
  - Test date selection event handling
  - Test "Today" button functionality
  - Test session storage persistence
  - Test loading state management
  - Test keyboard navigation

- [x] 7. Frontend: Modify KDS service to support date parameters
  - Modify `web/src/services/kdsService.js`
  - Update `getCookingToday()` to accept optional date parameter
  - Update `getPackingToday()` to accept optional date parameter
  - Create `formatDate()` helper function to format date as YYYY-MM-DD
  - Add date parameter to API request query string when provided
  - Maintain backward compatibility when date is null/undefined
  - _Requirements: 2.1, 2.2, 6.1, 6.2, 6.3_

- [ ]* 7.1 Write unit tests for KDS service modifications
  - Test API call with date parameter
  - Test API call without date parameter (backward compatibility)
  - Test date formatting function
  - Test error handling for network failures

- [x] 8. Frontend: Integrate date picker into KDSCookingView
  - Modify `web/src/views/KDSCookingView.vue`
  - Import and add KDSDatePicker component to page header
  - Add selectedDate reactive state (default to current date)
  - Pass selectedDate to `getCookingToday()` API call
  - Update Firebase listener path based on selected date
  - Display selected date prominently in UI header
  - Handle empty data states with appropriate message
  - Add error handling with retry option
  - _Requirements: 1.3, 3.2, 3.3, 5.1, 8.3, 8.4_

- [x] 9. Frontend: Integrate date picker into KDSPackingView
  - Modify `web/src/views/KDSPackingView.vue`
  - Import and add KDSDatePicker component to page header
  - Add selectedDate reactive state (default to current date)
  - Pass selectedDate to `getPackingToday()` API call
  - Update Firebase listener path based on selected date
  - Display selected date prominently in UI header
  - Handle empty data states with appropriate message
  - Add error handling with retry option
  - _Requirements: 1.4, 3.2, 3.3, 5.1, 8.3, 8.4_

- [ ]* 9.1 Write unit tests for view modifications
  - Test date picker integration in both views
  - Test API call triggering on date change
  - Test empty data state display
  - Test error message display
  - Test loading state management

- [x] 10. Checkpoint - Ensure frontend tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ]* 11. Integration: Write end-to-end integration tests
  - Test complete flow: select date → API call → data display
  - Test Firebase listener updates with date changes
  - Test multiple date selections in sequence
  - Test network error recovery
  - Test backward compatibility with existing code
  - Verify data consistency between cooking and packing views

- [x] 12. Final checkpoint - Verify all requirements and run full test suite
  - Ensure all property-based tests pass (minimum 100 iterations each)
  - Ensure all unit tests pass
  - Ensure all integration tests pass
  - Verify backward compatibility with existing API consumers
  - Ask the user if questions arise or if manual testing is needed

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Backend tasks (1-5) should be completed before frontend tasks (6-10)
- Property-based tests use gopter framework with minimum 100 iterations
- All date handling uses Asia/Jakarta timezone consistently
- Backward compatibility is maintained throughout - existing API consumers continue to work
- Empty data returns HTTP 200 with empty array, not error status
