# Requirements Document

## Introduction

The Tugas Pengambilan (Pickup Task Management) feature extends the existing delivery management system to handle the collection phase of the food delivery workflow. After drivers complete food deliveries to schools (Stage 9: sudah_diterima_pihak_sekolah), they must return to collect ompreng (food containers). This feature enables dispatchers to create pickup tasks, assign drivers, plan multi-school pickup routes, and track pickup progress through stages 10-13 until containers are returned to SPPG.

## Glossary

- **Pickup_Task**: A work assignment for a driver to collect ompreng from one or more schools
- **Ompreng**: Food containers that must be collected from schools after delivery
- **Delivery_Record**: A record of a food delivery to a school, tracking status through stages 1-16
- **Stage**: A specific status in the delivery workflow (stages 10-13 for pickup phase)
- **SPPG**: The central facility where food is prepared and containers are cleaned
- **Driver**: A person assigned to perform pickup tasks
- **School**: A location where food was delivered and ompreng must be collected
- **Route_Order**: The sequence in which schools should be visited during a pickup task
- **Dispatcher**: A user who creates and manages pickup tasks
- **Pickup_Form**: The user interface for creating and managing pickup tasks

## Requirements

### Requirement 1: Display Eligible Orders for Pickup

**User Story:** As a dispatcher, I want to see all orders that are ready for pickup, so that I can create pickup tasks for completed deliveries.

#### Acceptance Criteria

1. THE Pickup_Form SHALL display Delivery_Records with status "Sudah Diterima" (Stage 9)
2. WHEN a Delivery_Record transitions to Stage 9, THE System SHALL make it available in the Pickup_Form
3. THE Pickup_Form SHALL display school name, school address, and GPS coordinates for each eligible Delivery_Record
4. THE Pickup_Form SHALL exclude Delivery_Records that are already assigned to an active Pickup_Task
5. THE Pickup_Form SHALL refresh the list of eligible Delivery_Records when the page loads

### Requirement 2: Create Pickup Task with Driver Assignment

**User Story:** As a dispatcher, I want to create a pickup task and assign a driver, so that ompreng can be collected from schools.

#### Acceptance Criteria

1. WHEN a dispatcher selects one or more eligible Delivery_Records, THE Pickup_Form SHALL enable pickup task creation
2. THE Pickup_Form SHALL provide a driver selection interface with available drivers
3. WHEN a dispatcher submits a pickup task, THE System SHALL create a Pickup_Task record with the assigned driver
4. WHEN a Pickup_Task is created, THE System SHALL associate all selected Delivery_Records with the Pickup_Task
5. WHEN a Pickup_Task is created, THE System SHALL transition associated Delivery_Records to Stage 10 (driver_menuju_lokasi_pengambilan)
6. THE System SHALL prevent creating a Pickup_Task without at least one selected Delivery_Record
7. THE System SHALL prevent creating a Pickup_Task without an assigned driver

### Requirement 3: Define Pickup Route Order

**User Story:** As a dispatcher, I want to specify the order in which schools should be visited, so that drivers follow an efficient pickup route.

#### Acceptance Criteria

1. WHEN multiple Delivery_Records are selected for a Pickup_Task, THE Pickup_Form SHALL provide a route ordering interface
2. THE Pickup_Form SHALL allow the dispatcher to reorder schools by dragging or using sequence controls
3. WHEN a Pickup_Task is created, THE System SHALL store the Route_Order for each associated Delivery_Record
4. THE System SHALL assign Route_Order values starting from 1 for the first school in sequence
5. THE Pickup_Form SHALL display the current Route_Order for each selected school

### Requirement 4: Track Pickup Progress Through Stages

**User Story:** As a dispatcher, I want to track pickup task progress through stages 10-13, so that I can monitor driver activities and container collection status.

#### Acceptance Criteria

1. WHEN a driver arrives at a pickup location, THE System SHALL transition the Delivery_Record to Stage 11 (driver_tiba_di_lokasi_pengambilan)
2. WHEN a driver departs from a pickup location, THE System SHALL transition the Delivery_Record to Stage 12 (driver_kembali_ke_sppg)
3. WHEN a driver arrives at SPPG with collected ompreng, THE System SHALL transition the Delivery_Record to Stage 13 (driver_tiba_di_sppg)
4. THE System SHALL record the timestamp for each stage transition
5. THE System SHALL allow stage transitions only in sequential order (10 → 11 → 12 → 13)
6. WHEN all Delivery_Records in a Pickup_Task reach Stage 13, THE System SHALL mark the Pickup_Task as completed

### Requirement 5: Display Pickup Task Information

**User Story:** As a dispatcher, I want to view pickup task details, so that I can monitor active pickups and verify task information.

#### Acceptance Criteria

1. THE Pickup_Form SHALL display a list of active Pickup_Tasks
2. THE Pickup_Form SHALL display the assigned driver name for each Pickup_Task
3. THE Pickup_Form SHALL display the number of schools in each Pickup_Task
4. THE Pickup_Form SHALL display the current stage for each Delivery_Record in a Pickup_Task
5. WHEN a dispatcher selects a Pickup_Task, THE Pickup_Form SHALL display all associated schools in Route_Order sequence
6. THE Pickup_Form SHALL display school name, address, GPS coordinates, and current stage for each school in a Pickup_Task

### Requirement 6: Integrate Pickup Form with Delivery Task Page

**User Story:** As a dispatcher, I want to access the pickup task form on the same page as delivery tasks, so that I can manage both delivery and pickup operations efficiently.

#### Acceptance Criteria

1. THE System SHALL display the Pickup_Form on the same page as the existing delivery task form
2. THE Pickup_Form SHALL be visually distinct from the delivery task form
3. THE System SHALL allow simultaneous viewing of both delivery and pickup task forms
4. THE Pickup_Form SHALL maintain its state independently from the delivery task form

### Requirement 7: Validate Pickup Task Data

**User Story:** As a system administrator, I want pickup task data to be validated, so that data integrity is maintained.

#### Acceptance Criteria

1. WHEN creating a Pickup_Task, THE System SHALL verify that all selected Delivery_Records are in Stage 9
2. WHEN creating a Pickup_Task, THE System SHALL verify that the assigned driver exists in the system
3. WHEN creating a Pickup_Task, THE System SHALL verify that Route_Order values are unique within the task
4. IF a Delivery_Record is already assigned to an active Pickup_Task, THEN THE System SHALL prevent it from being assigned to another Pickup_Task
5. IF a stage transition is invalid, THEN THE System SHALL reject the transition and return an error message

### Requirement 8: Handle Multi-School Pickup Scenarios

**User Story:** As a dispatcher, I want to assign multiple schools to a single pickup task, so that drivers can collect ompreng from multiple locations in one trip.

#### Acceptance Criteria

1. THE System SHALL support Pickup_Tasks with one or more Delivery_Records
2. WHEN a Pickup_Task contains multiple schools, THE System SHALL track each school's stage independently
3. THE Pickup_Form SHALL display progress for each school within a multi-school Pickup_Task
4. THE System SHALL allow stage transitions for individual schools within a Pickup_Task
5. WHEN a driver completes pickup at one school in a multi-school task, THE System SHALL allow progression to the next school based on Route_Order

### Requirement 9: Persist Pickup Task Data

**User Story:** As a system administrator, I want pickup task data to be persisted in the database, so that task information is retained and can be retrieved.

#### Acceptance Criteria

1. WHEN a Pickup_Task is created, THE System SHALL store the task in the database with a unique identifier
2. THE System SHALL store the relationship between Pickup_Task and Delivery_Records in the database
3. THE System SHALL store Route_Order values for each Delivery_Record in a Pickup_Task
4. THE System SHALL store stage transition timestamps in the database
5. THE System SHALL store the assigned driver identifier with each Pickup_Task
6. WHEN the Pickup_Form loads, THE System SHALL retrieve Pickup_Task data from the database

### Requirement 10: Display School Location Information

**User Story:** As a driver, I want to see school location details including GPS coordinates, so that I can navigate to pickup locations.

#### Acceptance Criteria

1. THE Pickup_Form SHALL display school name for each pickup location
2. THE Pickup_Form SHALL display school address for each pickup location
3. THE Pickup_Form SHALL display GPS coordinates (latitude and longitude) for each pickup location
4. THE System SHALL retrieve school location information from existing Delivery_Record data
5. THE Pickup_Form SHALL display schools in Route_Order sequence when showing a Pickup_Task

### Requirement 11: Update Status Per Route in Pickup Task Detail

**User Story:** As a dispatcher, I want to update the status of each school route independently in the pickup task detail view, so that I can track progress for each school separately as drivers complete pickups at different times.

#### Acceptance Criteria

1. THE Pickup_Form SHALL display an "Aksi" column in the Detail Rute Pengambilan section for each school route
2. THE Pickup_Form SHALL provide a status dropdown in the "Aksi" column for each school route
3. THE Status_Dropdown SHALL contain three options: "Sudah Tiba" (Stage 11), "Dalam Perjalanan Kembali" (Stage 12), and "Selesai" (Stage 13)
4. WHEN a dispatcher selects a status from the dropdown, THE System SHALL transition the corresponding Delivery_Record to the selected stage
5. THE System SHALL enforce sequential stage progression (Stage 11 → 12 → 13) and prevent skipping stages
6. IF a dispatcher attempts to skip a stage, THEN THE System SHALL reject the transition and display an error message
7. THE Status_Dropdown SHALL display only valid next stages based on the current stage of each Delivery_Record
8. WHEN a Delivery_Record is in Stage 10, THE Status_Dropdown SHALL only show "Sudah Tiba" (Stage 11) as available
9. WHEN a Delivery_Record is in Stage 11, THE Status_Dropdown SHALL only show "Dalam Perjalanan Kembali" (Stage 12) as available
10. WHEN a Delivery_Record is in Stage 12, THE Status_Dropdown SHALL only show "Selesai" (Stage 13) as available
11. WHEN a Delivery_Record reaches Stage 13, THE Status_Dropdown SHALL be disabled or hidden
12. THE System SHALL update the stage display in the Detail Rute Pengambilan immediately after a successful status transition
13. THE System SHALL allow independent status updates for each school route within the same Pickup_Task
