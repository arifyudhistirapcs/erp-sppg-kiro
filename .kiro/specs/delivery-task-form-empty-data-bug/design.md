# Delivery Task Form Empty Data Bug - Bugfix Design

## Overview

The delivery task creation form displays "Tidak ada driver yang tersedia" (No drivers available) and "Tidak ada order yang siap kirim" (No orders ready for delivery) despite the backend database containing active drivers and delivery records with status "selesai_dipacking". The root cause is an overly restrictive query condition in the `GetReadyOrders` service method that filters out delivery records where `driver_id IS NULL`. This prevents orders that may have been previously assigned to a driver from appearing in the dropdown, even though they should be available for creating new delivery tasks.

The fix involves removing or adjusting the `driver_id IS NULL` condition to allow orders with status "selesai_dipacking" to appear regardless of whether they have a driver assigned, since the delivery task creation workflow should be able to assign or reassign drivers to packed orders.

## Glossary

- **Bug_Condition (C)**: The condition that triggers the bug - when the form is loaded with a valid date that has packed orders and available drivers, but the dropdowns show "No data"
- **Property (P)**: The desired behavior - dropdowns should display all orders with status "selesai_dipacking" and all active drivers not currently assigned on that date
- **Preservation**: Existing form behavior for date selection, order details display, and task creation must remain unchanged
- **GetReadyOrders**: The service method in `backend/internal/services/delivery_task_service.go` that retrieves delivery records ready for delivery
- **GetAvailableDrivers**: The service method in `backend/internal/services/delivery_task_service.go` that retrieves drivers available for assignment
- **selesai_dipacking**: Database status value indicating a delivery record has completed packing and is ready for delivery assignment
- **driver_id**: Foreign key field in delivery_records table that references the assigned driver (nullable)

## Bug Details

### Fault Condition

The bug manifests when a user selects a delivery date in the form and the backend queries return empty arrays despite valid data existing in the database. The `GetReadyOrders` function includes a condition `WHERE("delivery_records.driver_id IS NULL")` that filters out any delivery records that already have a driver assigned, even though these records may still need to be available for delivery task creation.

**Formal Specification:**
```
FUNCTION isBugCondition(input)
  INPUT: input of type { date: Date, deliveryRecords: DeliveryRecord[], drivers: User[] }
  OUTPUT: boolean
  
  RETURN input.date IS valid
         AND EXISTS deliveryRecord IN database WHERE (
           deliveryRecord.delivery_date = input.date
           AND deliveryRecord.current_status = "selesai_dipacking"
         )
         AND EXISTS driver IN database WHERE (
           driver.role = "driver"
           AND driver.is_active = true
         )
         AND (readyOrders.length = 0 OR availableDrivers.length = 0)
END FUNCTION
```

### Examples

- **Example 1**: User selects date "01/03/2026", database has 2 delivery records with status "selesai_dipacking" and driver_id = 1, but GetReadyOrders returns empty array because of the `driver_id IS NULL` filter
- **Example 2**: User selects date "01/03/2026", database has Driver 1 (ID=1) and Driver 2 (ID=2) with role "driver" and is_active=true, GetAvailableDrivers correctly returns both drivers
- **Example 3**: User selects date "01/03/2026", database has 1 delivery record with status "selesai_dipacking" and driver_id = NULL, GetReadyOrders correctly returns this record
- **Edge case**: User selects a date with no delivery records at all - form should correctly display "Tidak ada order yang siap kirim" (this behavior should be preserved)

## Expected Behavior

### Preservation Requirements

**Unchanged Behaviors:**
- Date selection triggering API calls must continue to work exactly as before
- Order details display panel showing school name, portions, menu item, and status must remain unchanged
- Form submission creating delivery tasks must continue to work as before
- Error handling for network failures must remain unchanged
- Loading states for dropdowns must remain unchanged
- Warning messages for truly empty data (no records in database) must continue to display correctly

**Scope:**
All inputs that do NOT involve the specific query conditions for ready orders should be completely unaffected by this fix. This includes:
- Form validation logic
- Date picker behavior
- Dropdown UI rendering
- Task list refresh after creation
- User authentication and authorization

## Hypothesized Root Cause

Based on the bug description and code analysis, the most likely issues are:

1. **Overly Restrictive Query Filter**: The `GetReadyOrders` method includes `Where("delivery_records.driver_id IS NULL")` which filters out delivery records that already have a driver assigned. This is problematic because:
   - Delivery records may have a driver_id set during a previous workflow step
   - The delivery task creation form should allow creating tasks for orders regardless of whether they have a preliminary driver assignment
   - The business logic may require reassigning or confirming driver assignments through the delivery task creation process

2. **Misunderstanding of Workflow**: The query assumes that only orders without a driver should appear in the form, but the actual workflow may be:
   - Orders get packed (status = "selesai_dipacking")
   - Orders may have a preliminary driver assignment
   - Delivery tasks are created to formalize the driver assignment and schedule
   - The form should show all packed orders, not just unassigned ones

3. **Data State Mismatch**: The database may have delivery records where driver_id is populated during packing or an earlier stage, causing the NULL filter to exclude all records

4. **Missing Business Logic**: The query may need additional conditions to determine which orders are truly "ready" for delivery task creation, rather than simply filtering by driver_id

## Correctness Properties

Property 1: Fault Condition - Display All Packed Orders

_For any_ date selection where delivery records exist with status "selesai_dipacking", the fixed GetReadyOrders function SHALL return all such records regardless of whether driver_id is NULL or populated, allowing the form to display all packed orders available for delivery task creation.

**Validates: Requirements 2.1, 2.3, 2.5**

Property 2: Preservation - Empty Data Handling

_For any_ date selection where NO delivery records exist with status "selesai_dipacking" OR NO active drivers exist, the fixed code SHALL produce exactly the same behavior as the original code, preserving the display of appropriate warning messages "Tidak ada order yang siap kirim" or "Tidak ada driver yang tersedia".

**Validates: Requirements 3.1, 3.2, 3.3**

## Fix Implementation

### Changes Required

Assuming our root cause analysis is correct:

**File**: `backend/internal/services/delivery_task_service.go`

**Function**: `GetReadyOrders` (Line 348)

**Specific Changes**:
1. **Remove driver_id Filter**: Remove the line `Where("delivery_records.driver_id IS NULL")` from the query
   - This allows all delivery records with status "selesai_dipacking" to be returned
   - The form will display all packed orders regardless of preliminary driver assignment

2. **Alternative Approach (if needed)**: If the business logic requires filtering, replace the NULL check with a more appropriate condition such as:
   - Check if a delivery_task already exists for this delivery_record
   - Check if the delivery_record is in a specific stage that indicates it needs task creation
   - Add a flag field to delivery_records indicating availability for task creation

3. **Update Query Logic**: Modify the WHERE clause to:
   ```go
   Where("DATE(delivery_records.delivery_date) = DATE(?)", date).
   Where("delivery_records.current_status = ?", "selesai_dipacking")
   // Remove: Where("delivery_records.driver_id IS NULL")
   ```

4. **Consider Adding Join**: If delivery tasks should not be created for records that already have a task, add a LEFT JOIN to check:
   ```go
   Joins("LEFT JOIN delivery_tasks ON delivery_records.id = delivery_tasks.delivery_record_id").
   Where("delivery_tasks.id IS NULL")
   ```

5. **Verify GetAvailableDrivers**: Ensure the GetAvailableDrivers query is working correctly (it appears correct based on code review, but should be tested)

## Testing Strategy

### Validation Approach

The testing strategy follows a two-phase approach: first, surface counterexamples that demonstrate the bug on unfixed code, then verify the fix works correctly and preserves existing behavior.

### Exploratory Fault Condition Checking

**Goal**: Surface counterexamples that demonstrate the bug BEFORE implementing the fix. Confirm or refute the root cause analysis. If we refute, we will need to re-hypothesize.

**Test Plan**: Write tests that query the database directly and call the GetReadyOrders service method with dates that have delivery records with status "selesai_dipacking" and non-NULL driver_id. Run these tests on the UNFIXED code to observe failures and confirm the root cause.

**Test Cases**:
1. **Packed Order with Driver Test**: Create a delivery record with status "selesai_dipacking" and driver_id = 1, call GetReadyOrders for that date (will return empty array on unfixed code, confirming the bug)
2. **Packed Order without Driver Test**: Create a delivery record with status "selesai_dipacking" and driver_id = NULL, call GetReadyOrders for that date (should return the record even on unfixed code)
3. **Multiple Orders Mixed Test**: Create 2 delivery records, one with driver_id = 1 and one with driver_id = NULL, call GetReadyOrders (will return only 1 record on unfixed code)
4. **Available Drivers Test**: Create 2 active drivers, call GetAvailableDrivers for a date with no tasks (should return both drivers even on unfixed code)

**Expected Counterexamples**:
- GetReadyOrders returns empty array when delivery records exist with non-NULL driver_id
- Possible causes: the `driver_id IS NULL` filter is excluding valid records

### Fix Checking

**Goal**: Verify that for all inputs where the bug condition holds, the fixed function produces the expected behavior.

**Pseudocode:**
```
FOR ALL date WHERE hasPackedOrders(date) AND hasActiveDrivers() DO
  orders := GetReadyOrders_fixed(date)
  drivers := GetAvailableDrivers_fixed(date)
  ASSERT orders.length > 0
  ASSERT drivers.length > 0
  ASSERT ALL order IN orders HAVE order.current_status = "selesai_dipacking"
END FOR
```

### Preservation Checking

**Goal**: Verify that for all inputs where the bug condition does NOT hold, the fixed function produces the same result as the original function.

**Pseudocode:**
```
FOR ALL date WHERE NOT hasPackedOrders(date) OR NOT hasActiveDrivers() DO
  ASSERT GetReadyOrders_original(date) = GetReadyOrders_fixed(date)
  ASSERT GetAvailableDrivers_original(date) = GetAvailableDrivers_fixed(date)
END FOR
```

**Testing Approach**: Property-based testing is recommended for preservation checking because:
- It generates many test cases automatically across the input domain
- It catches edge cases that manual unit tests might miss
- It provides strong guarantees that behavior is unchanged for all non-buggy inputs

**Test Plan**: Observe behavior on UNFIXED code first for dates with no data, then write property-based tests capturing that behavior.

**Test Cases**:
1. **Empty Date Preservation**: Observe that dates with no delivery records return empty arrays on unfixed code, then write test to verify this continues after fix
2. **Different Status Preservation**: Observe that delivery records with status other than "selesai_dipacking" are not returned on unfixed code, then write test to verify this continues after fix
3. **Inactive Driver Preservation**: Observe that drivers with is_active = false are not returned on unfixed code, then write test to verify this continues after fix

### Unit Tests

- Test GetReadyOrders with various delivery record configurations (with/without driver_id, different statuses, different dates)
- Test GetAvailableDrivers with various driver configurations (active/inactive, assigned/unassigned)
- Test edge cases (no data, invalid dates, multiple records on same date)
- Test that the response structure matches the expected format

### Property-Based Tests

- Generate random delivery records with various driver_id values and verify all "selesai_dipacking" records are returned
- Generate random driver configurations and verify all active unassigned drivers are returned
- Test that response data structure is consistent across many scenarios

### Integration Tests

- Test full form flow: select date, verify dropdowns populate, select order and driver, create task
- Test with real database data matching the user's screenshot scenario
- Test that created delivery tasks correctly reference the selected order and driver
- Test that form displays appropriate messages when data is truly empty
