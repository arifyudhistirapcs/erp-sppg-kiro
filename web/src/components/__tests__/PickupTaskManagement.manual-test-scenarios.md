# Pickup Task Management - Manual UI/UX Test Scenarios

This document provides comprehensive manual test scenarios for verifying the UI/UX functionality of the Pickup Task Management feature.

## Test Environment Setup

1. Ensure the backend server is running
2. Ensure the frontend development server is running
3. Login with appropriate credentials (kepala_sppg, kepala_yayasan, or asisten_lapangan)
4. Navigate to the Delivery Task List View page
5. Scroll down to the "Manajemen Tugas Pengambilan" section

## 1. Form Validation Tests

### Test 1.1: Required Field Validation - No Orders Selected
**Objective**: Verify that the form prevents submission when no orders are selected

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select a driver from the dropdown
3. Attempt to click "Buat Tugas Pengambilan" button

**Expected Result**:
- Submit button should be disabled (grayed out)
- No API call should be made

**Requirements**: 1.1, 2.6

---

### Test 1.2: Required Field Validation - No Driver Selected
**Objective**: Verify that the form prevents submission when no driver is selected

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select one or more orders from the eligible orders table
3. Do NOT select a driver
4. Attempt to click "Buat Tugas Pengambilan" button

**Expected Result**:
- Submit button should be disabled (grayed out)
- No API call should be made

**Requirements**: 2.1, 2.7

---

### Test 1.3: Successful Form Submission
**Objective**: Verify that the form submits successfully when all required fields are filled

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select 2-3 orders from the eligible orders table
3. Select a driver from the dropdown
4. Click "Buat Tugas Pengambilan" button

**Expected Result**:
- Success message appears: "Tugas pengambilan berhasil dibuat"
- Form is reset (all selections cleared)
- Eligible orders table is refreshed
- User is automatically switched to "Daftar Tugas Aktif" tab

**Requirements**: 2.1, 2.3, 2.4, 2.5

---

### Test 1.4: Error Display - API Failure
**Objective**: Verify that error messages are displayed when API calls fail

**Steps**:
1. Disconnect from network or stop backend server
2. Navigate to "Buat Tugas Pengambilan" tab
3. Observe the page loading

**Expected Result**:
- Error message appears: "Gagal memuat data order yang siap diambil"
- Error message appears: "Gagal memuat data driver yang tersedia"
- Tables show empty state or loading indicators

**Requirements**: 1.5, 2.1

---

## 2. Drag-and-Drop Route Ordering Tests

### Test 2.1: Display Route Order
**Objective**: Verify that selected orders display with route numbers

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select 3 orders from the eligible orders table
3. Observe the "Urutan Rute Pengambilan" section

**Expected Result**:
- Section appears with title showing count: "Urutan Rute Pengambilan (3 sekolah)"
- Info alert displays: "Seret untuk mengatur urutan pengambilan"
- Each order displays with a blue tag showing "Rute 1", "Rute 2", "Rute 3"
- Orders show school name, address, GPS coordinates, and ompreng count
- Drag handle icon (≡) appears on the left of each order card

**Requirements**: 3.1, 3.5

---

### Test 2.2: Drag and Drop Reordering
**Objective**: Verify that orders can be reordered via drag and drop

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select 3 orders (Order A, Order B, Order C)
3. Note the initial order
4. Click and hold the drag handle (≡) on Order C
5. Drag Order C to the first position
6. Release the mouse button

**Expected Result**:
- Order C moves to position 1 (displays "Rute 1")
- Order A moves to position 2 (displays "Rute 2")
- Order B moves to position 3 (displays "Rute 3")
- Route numbers update automatically
- Visual feedback during drag (cursor changes, card follows mouse)

**Requirements**: 3.1, 3.2

---

### Test 2.3: Remove Order from Route
**Objective**: Verify that orders can be removed from the route

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select 3 orders
3. Click the red delete button (trash icon) on the second order

**Expected Result**:
- Order is removed from the route list
- Remaining orders are renumbered (1, 2)
- Count updates: "Urutan Rute Pengambilan (2 sekolah)"
- Order becomes available for selection again in the eligible orders table

**Requirements**: 3.1, 3.2

---

### Test 2.4: Route Order Persistence in Submission
**Objective**: Verify that the route order is preserved when creating a pickup task

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select 3 orders
3. Reorder them via drag and drop to: Order C, Order A, Order B
4. Select a driver
5. Click "Buat Tugas Pengambilan"
6. Navigate to "Daftar Tugas Aktif" tab
7. Expand the newly created task

**Expected Result**:
- Task is created successfully
- In the expanded view, schools appear in the order: Order C (Rute 1), Order A (Rute 2), Order B (Rute 3)
- Route order matches the order set during creation

**Requirements**: 3.3, 3.4

---

## 3. Table Sorting and Filtering Tests

### Test 3.1: Eligible Orders Table Display
**Objective**: Verify that eligible orders are displayed correctly in the table

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Observe the eligible orders table

**Expected Result**:
- Table displays with columns: Sekolah, Koordinat GPS, Jumlah Ompreng, Tanggal Pengiriman
- Each row shows:
  - School name (bold) and address (gray text)
  - GPS coordinates with 6 decimal places and a map icon button
  - Ompreng count in a blue tag (e.g., "15 wadah")
  - Delivery date in Indonesian format (DD/MM/YYYY)
- Checkboxes appear for row selection
- Pagination controls appear if more than 10 orders

**Requirements**: 1.1, 1.3, 10.1, 10.2, 10.3

---

### Test 3.2: Row Selection in Eligible Orders
**Objective**: Verify that multiple orders can be selected

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Click checkboxes to select 3 different orders
3. Observe the selection state

**Expected Result**:
- Selected rows are highlighted
- Checkboxes show checked state
- Selected orders appear in the "Urutan Rute Pengambilan" section below
- Selection count is accurate

**Requirements**: 2.1

---

### Test 3.3: Pickup Tasks Table Display
**Objective**: Verify that active pickup tasks are displayed correctly

**Steps**:
1. Navigate to "Daftar Tugas Aktif" tab
2. Observe the pickup tasks table

**Expected Result**:
- Table displays with columns: ID Tugas, Driver, Jumlah Sekolah, Progress, Status, Dibuat
- Each row shows:
  - Task ID number
  - Driver name (bold) and phone number (gray text)
  - School count (number)
  - Progress bar with percentage and "X / Y sekolah selesai" text
  - Status tag (blue for "Aktif", green for "Selesai")
  - Creation timestamp in Indonesian format
- Expand icon (▶) appears for rows with delivery records
- Refresh button appears in the card header

**Requirements**: 5.1, 5.2, 5.3, 5.4

---

### Test 3.4: Expandable Row Functionality
**Objective**: Verify that pickup task rows can be expanded to show details

**Steps**:
1. Navigate to "Daftar Tugas Aktif" tab
2. Click the expand icon (▶) on a pickup task row

**Expected Result**:
- Row expands to show "Detail Rute Pengambilan" section
- Nested table displays with columns: Urutan, Sekolah, Koordinat GPS, Stage, Jumlah Ompreng
- Each delivery record shows:
  - Route order in a blue tag (e.g., "Rute 1")
  - School name (bold) and address (gray text)
  - GPS coordinates with map icon button
  - Stage tag with color coding (blue/orange/purple/green) and Indonesian text
  - Ompreng count in a blue tag
- Records are sorted by route_order (ascending)
- Expand icon changes to collapse icon (▼)

**Requirements**: 5.5, 5.6, 8.3, 10.1, 10.2, 10.3

---

### Test 3.5: Stage Color Coding
**Objective**: Verify that stage indicators use correct colors

**Steps**:
1. Navigate to "Daftar Tugas Aktif" tab
2. Expand a pickup task with multiple schools at different stages
3. Observe the stage tags

**Expected Result**:
- Stage 10 (Menuju Lokasi): Blue tag
- Stage 11 (Tiba di Sekolah): Orange tag
- Stage 12 (Kembali ke SPPG): Purple tag
- Stage 13 (Tiba di SPPG): Green tag
- Stage text is in Indonesian

**Requirements**: 4.1, 4.2, 4.3, 5.4

---

### Test 3.6: Progress Calculation
**Objective**: Verify that progress is calculated correctly

**Steps**:
1. Navigate to "Daftar Tugas Aktif" tab
2. Find a task with 3 schools where 1 is at stage 13
3. Observe the progress bar

**Expected Result**:
- Progress bar shows 33% (1/3)
- Text shows "1 / 3 sekolah selesai"
- Progress bar color is blue (active status)
- When all schools reach stage 13, progress shows 100% and turns green

**Requirements**: 4.6, 8.2, 8.3

---

### Test 3.7: Refresh Functionality
**Objective**: Verify that the refresh button updates the data

**Steps**:
1. Navigate to "Daftar Tugas Aktif" tab
2. Note the current tasks displayed
3. Click the "Refresh" button in the card header
4. Observe the loading state and updated data

**Expected Result**:
- Loading spinner appears on the button
- Table shows loading state
- Data is refreshed from the server
- Any changes in task status or progress are reflected
- Loading state clears when complete

**Requirements**: 5.1

---

## 4. Responsive Layout Tests

### Test 4.1: Desktop View (1920x1080)
**Objective**: Verify that the layout works correctly on large desktop screens

**Steps**:
1. Set browser window to 1920x1080 resolution
2. Navigate to the Pickup Task Management section
3. Observe the layout

**Expected Result**:
- All tables display with full width
- All columns are visible without horizontal scrolling
- Cards have appropriate spacing
- Text is readable and not cramped
- Drag-and-drop area is spacious and easy to use
- No layout overflow or broken elements

**Requirements**: 6.1, 6.2

---

### Test 4.2: Laptop View (1366x768)
**Objective**: Verify that the layout adapts to standard laptop screens

**Steps**:
1. Set browser window to 1366x768 resolution
2. Navigate to the Pickup Task Management section
3. Test all functionality

**Expected Result**:
- Layout adjusts to smaller width
- Tables may show horizontal scroll if needed
- All functionality remains accessible
- Text remains readable
- Buttons and interactive elements are properly sized
- No overlapping elements

**Requirements**: 6.1, 6.2

---

### Test 4.3: Tablet View (768x1024)
**Objective**: Verify that the layout works on tablet devices

**Steps**:
1. Set browser window to 768x1024 resolution (or use browser dev tools device emulation)
2. Navigate to the Pickup Task Management section
3. Test all functionality

**Expected Result**:
- Layout stacks vertically where appropriate
- Tables are scrollable horizontally if needed
- Touch targets (buttons, checkboxes) are appropriately sized
- Drag-and-drop still works with touch gestures
- Text remains readable
- No content is cut off or hidden

**Requirements**: 6.1, 6.2

---

### Test 4.4: Mobile View (375x667)
**Objective**: Verify that the layout is usable on mobile devices

**Steps**:
1. Set browser window to 375x667 resolution (iPhone SE size)
2. Navigate to the Pickup Task Management section
3. Test basic functionality

**Expected Result**:
- Layout is fully responsive and stacks vertically
- Tables show horizontal scroll with visible scroll indicators
- All interactive elements are touch-friendly (minimum 44x44px)
- Text is readable without zooming
- Forms are usable
- Drag-and-drop may be replaced with alternative controls if needed
- No horizontal page scroll (only table scroll)

**Requirements**: 6.1, 6.2

---

### Test 4.5: Browser Zoom Test
**Objective**: Verify that the layout works at different zoom levels

**Steps**:
1. Set browser to 100% zoom
2. Navigate to the Pickup Task Management section
3. Test at 75%, 100%, 125%, and 150% zoom levels

**Expected Result**:
- Layout remains functional at all zoom levels
- Text remains readable
- No overlapping elements
- Scrollbars appear when needed
- Interactive elements remain clickable
- No broken layouts

**Requirements**: 6.1

---

## 5. Integration and User Flow Tests

### Test 5.1: Complete Pickup Task Creation Flow
**Objective**: Verify the entire user flow from viewing eligible orders to creating a task

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Wait for eligible orders to load
3. Select 3 orders from different schools
4. Reorder them using drag-and-drop
5. Select a driver
6. Click "Buat Tugas Pengambilan"
7. Verify success message
8. Check "Daftar Tugas Aktif" tab

**Expected Result**:
- All steps complete without errors
- Success message appears
- Form resets after submission
- New task appears in the active tasks list
- Task shows correct driver, school count, and route order
- User experience is smooth and intuitive

**Requirements**: 1.1, 2.1, 3.1, 5.1, 6.1

---

### Test 5.2: Tab Switching and State Persistence
**Objective**: Verify that tab switching works correctly and state is maintained

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select 2 orders and a driver (do not submit)
3. Switch to "Daftar Tugas Aktif" tab
4. Switch back to "Buat Tugas Pengambilan" tab

**Expected Result**:
- Tab switching is smooth and immediate
- Form state is preserved (selections remain)
- No data loss occurs
- Both tabs function independently

**Requirements**: 6.3, 6.4

---

### Test 5.3: GPS Coordinates and Maps Integration
**Objective**: Verify that GPS coordinates are displayed and map links work

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Observe GPS coordinates in the eligible orders table
3. Click the map icon button next to coordinates
4. Repeat for expanded delivery records in "Daftar Tugas Aktif" tab

**Expected Result**:
- GPS coordinates display with 6 decimal places
- Format: "-6.208800, 106.845600"
- Map icon button is visible and clickable
- Clicking opens Google Maps in a new tab
- Map shows the correct location
- User can return to the application easily

**Requirements**: 1.3, 10.1, 10.2, 10.3

---

### Test 5.4: Driver Selection and Availability
**Objective**: Verify that driver selection works correctly

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Click the driver dropdown
3. Observe the available drivers
4. Search for a driver by name
5. Select a driver

**Expected Result**:
- Dropdown shows all available drivers
- Each driver shows name and phone number
- Search/filter functionality works
- Selected driver is highlighted
- Driver information is clear and readable
- If no drivers available, warning message appears

**Requirements**: 2.2

---

### Test 5.5: Error Recovery
**Objective**: Verify that users can recover from errors

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select orders and driver
3. Disconnect network
4. Click "Buat Tugas Pengambilan"
5. Observe error message
6. Reconnect network
7. Click "Buat Tugas Pengambilan" again

**Expected Result**:
- Error message appears clearly
- Form data is preserved
- User can retry after fixing the issue
- Successful submission after network restoration
- No data corruption or loss

**Requirements**: 2.1

---

### Test 5.6: Reset Functionality
**Objective**: Verify that the reset button clears the form

**Steps**:
1. Navigate to "Buat Tugas Pengambilan" tab
2. Select 3 orders
3. Reorder them
4. Select a driver
5. Click "Reset" button

**Expected Result**:
- All selections are cleared
- Route order section disappears
- Driver selection is cleared
- Form returns to initial state
- Eligible orders table remains loaded
- No errors occur

**Requirements**: 2.1

---

## 6. Accessibility Tests

### Test 6.1: Keyboard Navigation
**Objective**: Verify that all functionality is accessible via keyboard

**Steps**:
1. Navigate to the Pickup Task Management section
2. Use Tab key to navigate through all interactive elements
3. Use Enter/Space to activate buttons and checkboxes
4. Use arrow keys in dropdowns

**Expected Result**:
- All interactive elements are reachable via Tab
- Focus indicators are visible
- Tab order is logical (top to bottom, left to right)
- All actions can be performed with keyboard
- No keyboard traps

**Requirements**: 1.1, 2.1, 3.1, 5.1

---

### Test 6.2: Screen Reader Compatibility
**Objective**: Verify that the interface works with screen readers

**Steps**:
1. Enable a screen reader (NVDA, JAWS, or VoiceOver)
2. Navigate through the Pickup Task Management section
3. Listen to announcements for all elements

**Expected Result**:
- All text content is announced
- Form labels are associated with inputs
- Button purposes are clear
- Table structure is announced
- Status messages are announced
- Error messages are announced

**Requirements**: 1.1, 2.1, 5.1

---

### Test 6.3: Color Contrast
**Objective**: Verify that text has sufficient color contrast

**Steps**:
1. Use browser dev tools or a contrast checker
2. Check contrast ratios for all text elements
3. Check contrast for interactive elements

**Expected Result**:
- Normal text has at least 4.5:1 contrast ratio
- Large text has at least 3:1 contrast ratio
- Interactive elements have sufficient contrast
- Status indicators don't rely solely on color

**Requirements**: 1.1, 5.1

---

## Test Summary Checklist

- [ ] All form validation tests pass
- [ ] Drag-and-drop functionality works correctly
- [ ] Table sorting and filtering work as expected
- [ ] Responsive layout works on all screen sizes
- [ ] Integration flows complete successfully
- [ ] Error handling and recovery work properly
- [ ] Accessibility requirements are met
- [ ] GPS and maps integration functions correctly
- [ ] All requirements (1.1, 2.1, 3.1, 5.1, 6.1) are verified

## Notes for Testers

1. Test on multiple browsers (Chrome, Firefox, Safari, Edge)
2. Test with different user roles if applicable
3. Document any bugs or issues found
4. Take screenshots of any visual issues
5. Note any performance issues or slow loading
6. Verify that all Indonesian text is correct and appropriate
7. Check that date/time formats match Indonesian locale
8. Ensure all error messages are helpful and actionable
