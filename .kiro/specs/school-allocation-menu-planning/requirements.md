# Requirements Document

## Introduction

This feature enables users to allocate portions of menu items to specific schools when creating weekly menu plans. Currently, the menu planning system only tracks total portions per menu item per day without school-level breakdown. This enhancement will allow users to specify which schools receive which portions during menu plan creation, providing better visibility and control over school-specific allocations that will be used by the KDS packing system.

## Glossary

- **Menu_Planning_System**: The system component responsible for creating and managing weekly menu plans
- **Menu_Item**: A recipe assigned to a specific date in a menu plan with a total portion count
- **School_Allocation**: A record that assigns a specific number of portions of a menu item to a particular school
- **KDS_Packing_System**: The Kitchen Display System component that uses allocation data for packing operations
- **KDS_Cooking_View**: The Kitchen Display System interface used by kitchen staff to view recipes and preparation instructions
- **Weekly_Menu_Plan**: A collection of menu items organized by date for a specific week
- **Portion**: A single serving unit of a menu item
- **School**: An educational institution that receives meal deliveries

## Requirements

### Requirement 1: Allocate Portions to Schools

**User Story:** As a menu planner, I want to allocate portions of a menu item to specific schools, so that I can control which schools receive which meals on each day.

#### Acceptance Criteria

1. WHEN adding a menu item to a weekly menu plan, THE Menu_Planning_System SHALL allow the user to specify one or more School_Allocations
2. FOR EACH School_Allocation, THE Menu_Planning_System SHALL record the school identifier and the number of portions allocated
3. THE Menu_Planning_System SHALL allow multiple School_Allocations for a single Menu_Item
4. WHEN a user creates a School_Allocation, THE Menu_Planning_System SHALL validate that the school identifier exists in the schools table

### Requirement 2: Validate Total Portions

**User Story:** As a menu planner, I want the system to ensure my school allocations match the total portions, so that I don't create inconsistent data.

#### Acceptance Criteria

1. WHEN a user saves a Menu_Item with School_Allocations, THE Menu_Planning_System SHALL calculate the sum of all allocated portions
2. IF the sum of allocated portions does not equal the total portions for the Menu_Item, THEN THE Menu_Planning_System SHALL reject the save operation and return a validation error
3. THE Menu_Planning_System SHALL display the validation error message indicating the mismatch between total portions and allocated portions
4. WHEN the sum of allocated portions equals the total portions, THE Menu_Planning_System SHALL save the Menu_Item and all School_Allocations

### Requirement 3: Store School Allocation Data

**User Story:** As a system administrator, I want school allocation data to be persisted in the database, so that it can be retrieved and used by other system components.

#### Acceptance Criteria

1. THE Menu_Planning_System SHALL store School_Allocations in a database table that links menu items, schools, and portion counts
2. FOR EACH School_Allocation, THE Menu_Planning_System SHALL store the menu item identifier, school identifier, portion count, and date
3. WHEN a Menu_Item is deleted, THE Menu_Planning_System SHALL delete all associated School_Allocations
4. THE Menu_Planning_System SHALL ensure that School_Allocation records maintain referential integrity with menu items and schools tables

### Requirement 4: Retrieve School Allocations

**User Story:** As a KDS operator, I want to retrieve school allocation data for menu items, so that the packing system can use accurate school-specific portion counts.

#### Acceptance Criteria

1. WHEN the KDS_Packing_System requests allocation data for a specific date, THE Menu_Planning_System SHALL return all School_Allocations for that date
2. WHEN the KDS_Packing_System requests allocation data for a specific menu item, THE Menu_Planning_System SHALL return all School_Allocations for that menu item
3. FOR EACH School_Allocation returned, THE Menu_Planning_System SHALL include the school identifier, school name, and portion count
4. THE Menu_Planning_System SHALL return School_Allocations ordered by school name

### Requirement 5: Update School Allocations

**User Story:** As a menu planner, I want to modify school allocations after creating a menu item, so that I can adjust allocations when plans change.

#### Acceptance Criteria

1. WHEN a user edits a Menu_Item with existing School_Allocations, THE Menu_Planning_System SHALL display the current School_Allocations
2. THE Menu_Planning_System SHALL allow the user to add new School_Allocations to the Menu_Item
3. THE Menu_Planning_System SHALL allow the user to modify the portion count for existing School_Allocations
4. THE Menu_Planning_System SHALL allow the user to remove School_Allocations from the Menu_Item
5. WHEN saving updated School_Allocations, THE Menu_Planning_System SHALL validate that the sum of allocated portions equals the total portions

### Requirement 6: Display School Allocation Summary

**User Story:** As a menu planner, I want to see a summary of school allocations for each menu item, so that I can quickly verify the distribution of portions.

#### Acceptance Criteria

1. WHEN viewing a weekly menu plan, THE Menu_Planning_System SHALL display the total portions for each Menu_Item
2. FOR EACH Menu_Item with School_Allocations, THE Menu_Planning_System SHALL display a breakdown showing each school and its allocated portion count
3. THE Menu_Planning_System SHALL display the school name and portion count for each School_Allocation
4. WHEN a Menu_Item has no School_Allocations, THE Menu_Planning_System SHALL indicate that no schools have been allocated

### Requirement 7: Require School Allocations for Menu Items

**User Story:** As a menu planner, I want the system to require school allocations before saving menu items, so that all menu items have complete distribution information.

#### Acceptance Criteria

1. WHEN a user attempts to save a Menu_Item, THE Menu_Planning_System SHALL validate that at least one School_Allocation exists
2. IF no School_Allocations exist for the Menu_Item, THEN THE Menu_Planning_System SHALL reject the save operation and return a validation error
3. THE Menu_Planning_System SHALL display an error message indicating that school allocations are required before saving
4. WHEN at least one School_Allocation exists and the sum equals total portions, THE Menu_Planning_System SHALL save the Menu_Item

### Requirement 8: Prevent Duplicate School Allocations

**User Story:** As a menu planner, I want the system to prevent me from allocating portions to the same school multiple times for one menu item, so that I don't create duplicate or conflicting allocations.

#### Acceptance Criteria

1. WHEN a user creates a School_Allocation, THE Menu_Planning_System SHALL check if an allocation already exists for that school and menu item combination
2. IF a School_Allocation already exists for the school and menu item, THEN THE Menu_Planning_System SHALL reject the operation and return an error message
3. THE Menu_Planning_System SHALL display an error message indicating that the school has already been allocated portions for this menu item
4. THE Menu_Planning_System SHALL allow the user to modify the existing School_Allocation instead of creating a duplicate

### Requirement 9: Validate Portion Counts

**User Story:** As a menu planner, I want the system to ensure portion counts are valid numbers, so that I don't accidentally enter incorrect data.

#### Acceptance Criteria

1. WHEN a user enters a portion count for a School_Allocation, THE Menu_Planning_System SHALL validate that the value is a positive integer
2. IF the portion count is zero or negative, THEN THE Menu_Planning_System SHALL reject the input and display an error message
3. IF the portion count is not a valid integer, THEN THE Menu_Planning_System SHALL reject the input and display an error message
4. THE Menu_Planning_System SHALL validate portion counts before saving School_Allocations to the database

### Requirement 10: Display School Allocations in KDS Cooking View

**User Story:** As a kitchen staff member, I want to see which schools will receive each menu item in the cooking view, so that I know the distribution while preparing meals.

#### Acceptance Criteria

1. WHEN viewing menu items in the KDS_Cooking_View, THE KDS_Packing_System SHALL display the school allocation breakdown for each recipe
2. FOR EACH School_Allocation, THE KDS_Packing_System SHALL display the school name and portion count
3. THE KDS_Packing_System SHALL display the total portions as the sum of all School_Allocations for each menu item
4. THE KDS_Packing_System SHALL display School_Allocations ordered by school name in the cooking view

### Requirement 11: Display School Allocations in KDS Packing View

**User Story:** As a packing staff member, I want to see school-specific allocations in the packing view, so that I can pack the correct portions for each school.

#### Acceptance Criteria

1. WHEN viewing packing allocations, THE KDS_Packing_System SHALL display which menu items and portions are allocated to each school
2. FOR EACH school, THE KDS_Packing_System SHALL display the menu item name, recipe name, and portion count
3. THE KDS_Packing_System SHALL group allocations by school for easy packing workflow
4. THE KDS_Packing_System SHALL display schools in alphabetical order in the packing view
