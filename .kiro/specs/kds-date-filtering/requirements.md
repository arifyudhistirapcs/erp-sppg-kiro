# Requirements Document

## Introduction

Fitur ini menambahkan kemampuan filter tanggal pada Kitchen Display System (KDS) untuk tampilan dapur (cooking) dan packing. Saat ini, KDS hanya menampilkan data hari ini tanpa kemampuan melihat data historis, dan terdapat ketidaksesuaian data antara perencanaan mingguan dengan yang ditampilkan di KDS. Fitur ini akan memungkinkan pengguna untuk melihat data historis dan memastikan konsistensi data dengan perencanaan mingguan.

## Glossary

- **KDS**: Kitchen Display System - sistem tampilan untuk dapur dan packing
- **Cooking_Display**: Tampilan KDS untuk area dapur/memasak
- **Packing_Display**: Tampilan KDS untuk area packing
- **Menu_Items_Table**: Tabel database yang menyimpan data perencanaan mingguan
- **Date_Filter**: Komponen UI untuk memilih tanggal yang ingin ditampilkan
- **Historical_Data**: Data KDS dari tanggal sebelumnya
- **Backend_API**: API endpoint yang menyediakan data KDS
- **Query_Parameter**: Parameter tanggal yang dikirim dalam request API

## Requirements

### Requirement 1: Default Display Current Date

**User Story:** As a kitchen staff member, I want to see today's data by default when I open the KDS display, so that I can immediately start working on current orders without additional steps.

#### Acceptance Criteria

1. WHEN the Cooking_Display is loaded without a date parameter, THE Backend_API SHALL return data for the current date
2. WHEN the Packing_Display is loaded without a date parameter, THE Backend_API SHALL return data for the current date
3. THE Cooking_Display SHALL display the current date prominently in the UI
4. THE Packing_Display SHALL display the current date prominently in the UI

### Requirement 2: Date Filter Parameter Support

**User Story:** As a kitchen manager, I want to filter KDS data by specific dates, so that I can review historical data and verify past operations.

#### Acceptance Criteria

1. THE Backend_API SHALL accept an optional date Query_Parameter in format YYYY-MM-DD
2. WHEN a valid date Query_Parameter is provided, THE Backend_API SHALL return data for the specified date
3. WHEN an invalid date format is provided, THE Backend_API SHALL return an error message with status code 400
4. THE Backend_API SHALL support date queries for both cooking and packing endpoints
5. WHEN a future date is provided, THE Backend_API SHALL return an empty dataset or planned data if available

### Requirement 3: Historical Data Access

**User Story:** As a kitchen manager, I want to access historical KDS data, so that I can analyze past performance and resolve discrepancies.

#### Acceptance Criteria

1. THE Date_Filter SHALL allow selection of any past date
2. WHEN a historical date is selected, THE Cooking_Display SHALL update to show data from that date
3. WHEN a historical date is selected, THE Packing_Display SHALL update to show data from that date
4. THE Date_Filter SHALL provide a "Today" quick action button to return to current date
5. THE Date_Filter SHALL persist the selected date during the user session

### Requirement 4: Data Consistency with Planning

**User Story:** As a kitchen manager, I want KDS data to match the weekly planning data, so that I can trust the accuracy of the displays and avoid confusion.

#### Acceptance Criteria

1. WHEN data is retrieved for a specific date, THE Backend_API SHALL query the Menu_Items_Table using the same date
2. THE Cooking_Display SHALL show menu items that match the Menu_Items_Table for the selected date
3. THE Packing_Display SHALL show packing allocations that match the Menu_Items_Table for the selected date
4. IF there is no data in Menu_Items_Table for the selected date, THE Backend_API SHALL return an empty dataset with status code 200
5. THE Backend_API SHALL use consistent timezone handling across all date queries

### Requirement 5: UI Date Filter Component

**User Story:** As a kitchen staff member, I want an intuitive date picker in the KDS interface, so that I can easily switch between different dates.

#### Acceptance Criteria

1. THE Date_Filter SHALL be visible on both Cooking_Display and Packing_Display
2. THE Date_Filter SHALL display the currently selected date
3. WHEN the Date_Filter is clicked, THE Date_Filter SHALL open a calendar picker
4. THE Date_Filter SHALL support keyboard navigation for accessibility
5. WHEN a date is selected from the calendar, THE Date_Filter SHALL trigger a data refresh with the new date
6. THE Date_Filter SHALL provide visual feedback during data loading

### Requirement 6: API Endpoint Modifications

**User Story:** As a backend developer, I want to modify existing KDS endpoints to support date filtering, so that the system can serve both current and historical data.

#### Acceptance Criteria

1. THE Backend_API SHALL modify GET /api/v1/kds/cooking/today to accept optional date query parameter
2. THE Backend_API SHALL modify GET /api/v1/kds/packing/today to accept optional date query parameter
3. WHEN no date parameter is provided, THE Backend_API SHALL maintain backward compatibility by defaulting to current date
4. THE Backend_API SHALL validate date format before processing queries
5. THE Backend_API SHALL return appropriate error messages for invalid date inputs

### Requirement 7: Service Layer Date Handling

**User Story:** As a backend developer, I want service layer methods to handle date parameters, so that business logic correctly retrieves date-specific data.

#### Acceptance Criteria

1. THE kds_service.GetTodayMenu() SHALL be modified to accept a date parameter
2. THE packing_allocation_service.GetPackingAllocations() SHALL be modified to accept a date parameter
3. WHEN a date parameter is provided, THE services SHALL query database tables filtered by that date
4. THE services SHALL handle timezone conversion consistently
5. THE services SHALL return empty results for dates with no data rather than errors

### Requirement 8: Error Handling and Validation

**User Story:** As a user, I want clear error messages when something goes wrong with date filtering, so that I can understand and correct the issue.

#### Acceptance Criteria

1. WHEN an invalid date format is provided, THE Backend_API SHALL return error message "Invalid date format. Expected YYYY-MM-DD"
2. WHEN a date query fails, THE Backend_API SHALL log the error and return status code 500
3. THE Date_Filter SHALL display user-friendly error messages in the UI
4. WHEN network errors occur, THE Date_Filter SHALL show a retry option
5. THE Backend_API SHALL validate that date strings can be parsed before executing database queries

## Notes

- Implementasi ini akan meningkatkan transparansi operasional dapur
- Fitur ini memungkinkan audit dan verifikasi data historis
- Konsistensi data dengan perencanaan mingguan akan meningkatkan kepercayaan pengguna terhadap sistem
- Perlu mempertimbangkan performa query untuk data historis dalam jumlah besar
