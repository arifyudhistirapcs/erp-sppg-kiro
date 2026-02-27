# Requirements Document: Portion Size Differentiation

## Introduction

This feature introduces portion size differentiation for elementary schools (SD) in the menu planning system to accommodate different nutritional needs based on student age groups. Currently, the system treats all schools uniformly without distinguishing between younger and older students. This enhancement enables menu planners to allocate portions based on two size categories: small portions for SD grades 1-3 and large portions for SD grades 4-6, SMP, and SMA students.

The feature extends the existing school allocation system by adding portion size tracking to allocation records, modifying the menu planning UI to support size-aware allocation input, and updating KDS displays to show portion size information for kitchen and packing staff.

## Glossary

- **Portion_Size**: A classification of meal serving size, either 'small' (for younger students) or 'large' (for older students)
- **SD_School**: Elementary school (Sekolah Dasar) serving students in grades 1-6
- **SMP_School**: Junior high school (Sekolah Menengah Pertama) serving students in grades 7-9
- **SMA_School**: Senior high school (Sekolah Menengah Atas) serving students in grades 10-12
- **Mixed_Portion_School**: An SD school that requires both small and large portion allocations
- **Single_Portion_School**: An SMP or SMA school that requires only large portion allocations
- **School_Allocation**: A record linking a menu item to a school with specific portion counts and sizes
- **Menu_Item**: A recipe assigned to a specific date in a menu plan with total portion count
- **Allocation_Input**: User input specifying portions_small and portions_large for a school

## Requirements

### Requirement 1: Determine School Portion Size Type

**User Story:** As a system, I need to determine the portion size type for each school based on its category, so that the correct allocation options are presented to users.

#### Acceptance Criteria

1. WHEN the system processes a school with category = 'SD', THE system SHALL classify it as a Mixed_Portion_School requiring both small and large portion allocations
2. WHEN the system processes a school with category = 'SMP', THE system SHALL classify it as a Single_Portion_School requiring only large portion allocations
3. WHEN the system processes a school with category = 'SMA', THE system SHALL classify it as a Single_Portion_School requiring only large portion allocations
4. THE system SHALL store the portion size type determination logic in a reusable function accessible to all components

### Requirement 2: Add Portion Size Field to Allocations

**User Story:** As a system administrator, I need the database to track portion sizes for each allocation, so that the system can distinguish between small and large portion allocations.

#### Acceptance Criteria

1. THE system SHALL add a `portion_size` field to the `menu_item_school_allocations` table with values 'small' or 'large'
2. THE system SHALL make the `portion_size` field mandatory (NOT NULL) for all allocation records
3. THE system SHALL create a database index on the `portion_size` field for query performance
4. THE system SHALL ensure existing allocation records are migrated with appropriate portion_size values based on school category

### Requirement 3: Validate Portion Size Allocations

**User Story:** As a menu planner, I want the system to validate my portion size allocations, so that I don't create inconsistent or invalid data.

#### Acceptance Criteria

1. WHEN a user submits allocations for a menu item, THE system SHALL validate that the sum of all portions_small and portions_large equals the total_portions
2. IF the sum does not match, THE system SHALL reject the submission and return an error message indicating the mismatch
3. WHEN a user allocates portions to an SMP or SMA school, THE system SHALL reject any allocation with portions_small > 0
4. WHEN a user allocates portions to an SD school, THE system SHALL accept allocations with portions_small >= 0 and portions_large >= 0
5. THE system SHALL require that at least one of portions_small or portions_large is greater than zero for each school allocation
6. THE system SHALL validate that both portions_small and portions_large are non-negative integers

### Requirement 4: Create Separate Allocation Records by Portion Size

**User Story:** As a system, I need to create separate database records for small and large portions, so that each portion size can be tracked independently.

#### Acceptance Criteria

1. WHEN creating allocations for an SD school with portions_small > 0, THE system SHALL create an allocation record with portion_size = 'small'
2. WHEN creating allocations for an SD school with portions_large > 0, THE system SHALL create an allocation record with portion_size = 'large'
3. WHEN creating allocations for an SD school with both portion sizes > 0, THE system SHALL create exactly two allocation records (one small, one large)
4. WHEN creating allocations for an SMP or SMA school, THE system SHALL create exactly one allocation record with portion_size = 'large'
5. THE system SHALL ensure all allocation records for the same menu item and school share the same date value

### Requirement 5: Display Portion Size Options in Menu Planning UI

**User Story:** As a menu planner, I want to see portion size input fields for each school, so that I can allocate small and large portions appropriately.

#### Acceptance Criteria

1. WHEN viewing the menu item allocation form, THE system SHALL display each school with its category (SD, SMP, or SMA)
2. FOR each SD school, THE system SHALL display two input fields: one for small portions (grades 1-3) and one for large portions (grades 4-6)
3. FOR each SMP or SMA school, THE system SHALL display only one input field for large portions
4. THE system SHALL label the small portion field with "Small (Grades 1-3)" or equivalent text
5. THE system SHALL label the large portion field with "Large (Grades 4-6)" for SD schools or "Large" for SMP/SMA schools
6. THE system SHALL display real-time calculation of total allocated portions as users enter values

### Requirement 6: Validate Allocations Before Submission

**User Story:** As a menu planner, I want real-time validation feedback, so that I know if my allocations are correct before submitting.

#### Acceptance Criteria

1. WHEN the sum of allocated portions does not equal total portions, THE system SHALL display an error message indicating the difference
2. WHEN the sum of allocated portions equals total portions, THE system SHALL display a success indicator
3. WHEN a user enters portions_small for an SMP or SMA school, THE system SHALL display an error message immediately
4. THE system SHALL disable the submit button when validation errors exist
5. THE system SHALL enable the submit button only when all validations pass

### Requirement 7: Store Portion Size Allocations

**User Story:** As a system, I need to persist portion size allocations in the database, so that they can be retrieved and used by other system components.

#### Acceptance Criteria

1. WHEN a user saves a menu item with portion size allocations, THE system SHALL create allocation records with the correct portion_size values
2. THE system SHALL use a database transaction to ensure all allocations are saved atomically
3. IF any allocation fails to save, THE system SHALL roll back the entire transaction and return an error
4. THE system SHALL maintain referential integrity between menu items, schools, and allocations
5. WHEN a menu item is deleted, THE system SHALL cascade delete all associated allocation records

### Requirement 8: Retrieve Allocations Grouped by School

**User Story:** As a KDS operator, I want to see allocations grouped by school with portion size breakdown, so that I can understand the distribution of small and large portions.

#### Acceptance Criteria

1. WHEN retrieving allocations for a menu item, THE system SHALL group allocation records by school_id
2. FOR each SD school with multiple allocation records, THE system SHALL combine them into a single display record showing both portions_small and portions_large
3. FOR each SMP or SMA school, THE system SHALL display a single allocation with portions_large only
4. THE system SHALL return allocations ordered alphabetically by school name
5. THE system SHALL include school category in the response for each allocation

### Requirement 9: Display Portion Sizes in KDS Cooking View

**User Story:** As a kitchen staff member, I want to see portion size breakdown for each school, so that I know how many small and large portions to prepare.

#### Acceptance Criteria

1. WHEN viewing menu items in the KDS cooking view, THE system SHALL display school allocations with portion size information
2. FOR each SD school allocation, THE system SHALL display both small and large portion counts with appropriate labels
3. FOR each SMP or SMA school allocation, THE system SHALL display only the large portion count
4. THE system SHALL display the total portions as the sum of all small and large portions
5. THE system SHALL clearly label small portions as "Small (Grades 1-3)" and large portions as "Large (Grades 4-6)" or "Large"

### Requirement 10: Display Portion Sizes in KDS Packing View

**User Story:** As a packing staff member, I want to see portion size information for each school, so that I can pack the correct quantities of small and large portions.

#### Acceptance Criteria

1. WHEN viewing the KDS packing view, THE system SHALL display allocations grouped by school with portion size breakdown
2. FOR each SD school, THE system SHALL show separate counts for small portions and large portions
3. FOR each SMP or SMA school, THE system SHALL show only large portion count
4. THE system SHALL display schools in alphabetical order
5. THE system SHALL include visual indicators or labels to distinguish between small and large portions

### Requirement 11: Update Existing Allocations with Portion Sizes

**User Story:** As a menu planner, I want to edit existing menu items and update portion size allocations, so that I can adjust plans when needed.

#### Acceptance Criteria

1. WHEN editing a menu item with existing allocations, THE system SHALL display current portion size allocations for each school
2. THE system SHALL allow users to modify portions_small and portions_large values
3. WHEN saving updated allocations, THE system SHALL delete old allocation records and create new ones with updated values
4. THE system SHALL validate updated allocations using the same rules as creation
5. THE system SHALL use a transaction to ensure atomic updates (all or nothing)

### Requirement 12: Prevent Invalid Portion Size Combinations

**User Story:** As a system, I need to prevent invalid portion size combinations, so that data integrity is maintained.

#### Acceptance Criteria

1. WHEN a user attempts to allocate portions_small to an SMP school, THE system SHALL reject the request with error message "SMP schools cannot have small portions"
2. WHEN a user attempts to allocate portions_small to an SMA school, THE system SHALL reject the request with error message "SMA schools cannot have small portions"
3. WHEN a user submits an allocation with both portions_small = 0 and portions_large = 0, THE system SHALL reject the request with error message "School must have at least one portion"
4. THE system SHALL enforce these validations at both the API level and database level

### Requirement 13: Display Student Count Context

**User Story:** As a menu planner, I want to see student counts for each grade level, so that I can make informed decisions about portion allocations.

#### Acceptance Criteria

1. WHEN viewing the allocation form for an SD school, THE system SHALL display student_count_grade_1_3 next to the small portions field
2. WHEN viewing the allocation form for an SD school, THE system SHALL display student_count_grade_4_6 next to the large portions field
3. WHEN viewing the allocation form for an SMP or SMA school, THE system SHALL display total student_count next to the large portions field
4. THE system SHALL format student counts as "(X students)" or equivalent readable format
5. THE system SHALL retrieve student counts from the schools table in real-time

### Requirement 14: Maintain Backward Compatibility

**User Story:** As a system administrator, I want the system to handle existing allocations without portion size data, so that the migration is smooth.

#### Acceptance Criteria

1. WHEN the system encounters an allocation record without a portion_size value, THE system SHALL infer the portion size based on school category
2. FOR existing allocations to SD schools, THE system SHALL default to portion_size = 'large' during migration
3. FOR existing allocations to SMP or SMA schools, THE system SHALL set portion_size = 'large' during migration
4. THE system SHALL provide a migration script to update all existing allocation records with appropriate portion_size values
5. AFTER migration, THE system SHALL enforce the NOT NULL constraint on the portion_size field

### Requirement 15: Report Portion Size Statistics

**User Story:** As a menu planner, I want to see summary statistics of portion size allocations, so that I can verify the distribution is correct.

#### Acceptance Criteria

1. WHEN viewing a menu item summary, THE system SHALL display total small portions allocated across all SD schools
2. WHEN viewing a menu item summary, THE system SHALL display total large portions allocated across all schools
3. THE system SHALL display the percentage of small vs large portions
4. THE system SHALL display the number of schools receiving each portion size type
5. THE system SHALL update these statistics in real-time as users modify allocations
