# Requirements Document

## Introduction

The Logistics Monitoring Process feature is a sub-module under the Logistics module that provides comprehensive tracking of menu delivery and ompreng (food container) management throughout their complete lifecycle. This system tracks 15 distinct stages from cooking through delivery to cleaning, enabling real-time visibility for all stakeholders including kitchen staff, packing staff, drivers, cleaning staff, and administrators.

## Glossary

- **Monitoring_System**: The logistics monitoring process module that tracks and displays delivery and ompreng lifecycle status
- **Menu_Item**: A food item prepared for delivery to a school on a specific date
- **Ompreng**: Reusable food containers used for menu delivery
- **KDS_Cooking**: Kitchen Display System module for cooking operations
- **KDS_Packing**: Kitchen Display System module for packing operations
- **KDS_Cleaning**: Kitchen Display System module for ompreng cleaning operations
- **Delivery_Record**: A record tracking a menu delivery to a specific school
- **School**: An educational institution receiving menu deliveries
- **Driver**: Personnel responsible for delivering menus and collecting ompreng
- **Cleaning_Staff**: Personnel with "kebersihan" role responsible for cleaning ompreng
- **Status_Transition**: A change from one lifecycle stage to another
- **Timeline_View**: Visual representation of delivery progress with status indicators
- **Activity_Log**: Chronological record of status changes with timestamps
- **SPPG**: Central facility where ompreng are returned and cleaned

## Requirements

### Requirement 1: Display Monitoring Dashboard

**User Story:** As a logistics manager, I want to view a monitoring dashboard for deliveries on a specific date, so that I can track the progress of all menu deliveries and ompreng management.

#### Acceptance Criteria

1. THE Monitoring_System SHALL display a list of all Delivery_Records for a selected date
2. WHEN a Delivery_Record is selected, THE Monitoring_System SHALL display the Timeline_View showing all 15 lifecycle stages
3. THE Monitoring_System SHALL display School information including name, portion count, contact information, and address
4. THE Monitoring_System SHALL display Driver information including name, vehicle type, and contact information
5. THE Monitoring_System SHALL display the Activity_Log with timestamps for each Status_Transition

### Requirement 2: Track Menu Delivery Lifecycle

**User Story:** As a logistics coordinator, I want to track menu items through the delivery lifecycle, so that I can ensure timely delivery to schools.

#### Acceptance Criteria

1. WHEN a Menu_Item enters cooking stage, THE Monitoring_System SHALL record status "Sedang dimasak"
2. WHEN cooking is completed, THE Monitoring_System SHALL update status to "Selesai dimasak"
3. WHEN a Menu_Item is ready for packing, THE Monitoring_System SHALL update status to "Siap dipacking"
4. WHEN packing is completed, THE Monitoring_System SHALL update status to "Selesai dipacking"
5. WHEN a Menu_Item is ready for delivery, THE Monitoring_System SHALL update status to "Siap dikirim"
6. WHEN a Driver begins delivery, THE Monitoring_System SHALL update status to "Diperjalanan"
7. WHEN a Driver arrives at School, THE Monitoring_System SHALL update status to "Sudah sampai sekolah"
8. WHEN School personnel confirm receipt, THE Monitoring_System SHALL update status to "Sudah diterima pihak sekolah"

### Requirement 3: Track Ompreng Collection Lifecycle

**User Story:** As a logistics coordinator, I want to track ompreng collection from schools, so that I can ensure containers are returned for cleaning and reuse.

#### Acceptance Criteria

1. WHEN a Driver is assigned to collect Ompreng, THE Monitoring_System SHALL record status "Driver ditugaskan mengambil ompreng"
2. WHEN a Driver begins traveling to School, THE Monitoring_System SHALL update status to "Driver menuju sekolah"
3. WHEN a Driver arrives at School for collection, THE Monitoring_System SHALL update status to "Driver sampai di sekolah"
4. WHEN Ompreng are loaded for return, THE Monitoring_System SHALL update status to "Ompreng telah diambil"
5. WHEN Ompreng arrive at SPPG, THE Monitoring_System SHALL update status to "Ompreng sampai di SPPG"

### Requirement 4: Track Ompreng Cleaning Lifecycle

**User Story:** As a cleaning supervisor, I want to track ompreng through the cleaning process, so that I can ensure containers are properly cleaned and ready for reuse.

#### Acceptance Criteria

1. WHEN Ompreng enter the cleaning process, THE Monitoring_System SHALL record status "Ompreng proses pencucian"
2. WHEN cleaning is completed, THE Monitoring_System SHALL update status to "Ompreng selesai dicuci"
3. THE Monitoring_System SHALL record the timestamp for each cleaning Status_Transition
4. THE Monitoring_System SHALL associate cleaning records with the original Delivery_Record

### Requirement 5: Integrate with KDS Cooking Module

**User Story:** As a kitchen staff member, I want my cooking status updates to automatically reflect in the monitoring system, so that logistics can track progress without manual data entry.

#### Acceptance Criteria

1. WHEN KDS_Cooking records a Menu_Item as started, THE Monitoring_System SHALL update to status "Sedang dimasak"
2. WHEN KDS_Cooking records a Menu_Item as completed, THE Monitoring_System SHALL update to status "Selesai dimasak"
3. THE Monitoring_System SHALL retrieve Menu_Item details from KDS_Cooking including quantity and school assignment
4. FOR ALL status updates from KDS_Cooking, THE Monitoring_System SHALL record the exact timestamp

### Requirement 6: Integrate with KDS Packing Module

**User Story:** As a packing staff member, I want my packing status updates to automatically reflect in the monitoring system, so that delivery teams know when items are ready.

#### Acceptance Criteria

1. WHEN KDS_Packing marks a Menu_Item as ready, THE Monitoring_System SHALL update to status "Siap dipacking"
2. WHEN KDS_Packing marks a Menu_Item as packed, THE Monitoring_System SHALL update to status "Selesai dipacking"
3. THE Monitoring_System SHALL retrieve packing completion data from KDS_Packing
4. FOR ALL status updates from KDS_Packing, THE Monitoring_System SHALL record the exact timestamp

### Requirement 7: Create KDS Cleaning Module

**User Story:** As a cleaning staff member, I want a dedicated interface to view and update ompreng cleaning status, so that I can efficiently manage the cleaning workflow.

#### Acceptance Criteria

1. THE KDS_Cleaning SHALL display a list of Ompreng with status "Ompreng sampai di SPPG"
2. WHEN Cleaning_Staff begins cleaning, THE KDS_Cleaning SHALL allow updating status to "Ompreng proses pencucian"
3. WHEN Cleaning_Staff completes cleaning, THE KDS_Cleaning SHALL allow updating status to "Ompreng selesai dicuci"
4. THE KDS_Cleaning SHALL display Ompreng details including associated School and delivery date
5. THE KDS_Cleaning SHALL record the Cleaning_Staff member who performed each Status_Transition

### Requirement 8: Implement Kebersihan Role

**User Story:** As a system administrator, I want to create a "kebersihan" role for cleaning staff, so that I can control access to cleaning functions.

#### Acceptance Criteria

1. THE Monitoring_System SHALL support a user role named "kebersihan"
2. WHEN a user has role "kebersihan", THE Monitoring_System SHALL grant access to KDS_Cleaning only
3. WHEN a user has role "kebersihan", THE Monitoring_System SHALL deny access to other KDS modules
4. THE Monitoring_System SHALL allow administrators to assign role "kebersihan" to user accounts

### Requirement 9: Record Status Timestamps

**User Story:** As a logistics analyst, I want to see timestamps for each status change, so that I can analyze delivery performance and identify bottlenecks.

#### Acceptance Criteria

1. WHEN any Status_Transition occurs, THE Monitoring_System SHALL record the exact timestamp
2. THE Monitoring_System SHALL display timestamps in the Activity_Log in chronological order
3. THE Monitoring_System SHALL calculate elapsed time between consecutive Status_Transitions
4. THE Monitoring_System SHALL display timestamps in local timezone format

### Requirement 10: Display Visual Status Indicators

**User Story:** As a logistics coordinator, I want to see visual indicators for each status, so that I can quickly identify completed and in-progress stages.

#### Acceptance Criteria

1. WHEN a lifecycle stage is completed, THE Monitoring_System SHALL display a completed indicator
2. WHEN a lifecycle stage is in progress, THE Monitoring_System SHALL display an in-progress indicator
3. WHEN a lifecycle stage is pending, THE Monitoring_System SHALL display a pending indicator
4. THE Timeline_View SHALL display all 15 stages in sequential order with appropriate indicators

### Requirement 11: Associate Deliveries with Schools and Drivers

**User Story:** As a logistics manager, I want to see which driver is assigned to each school delivery, so that I can coordinate delivery schedules and handle issues.

#### Acceptance Criteria

1. THE Monitoring_System SHALL associate each Delivery_Record with exactly one School
2. THE Monitoring_System SHALL associate each Delivery_Record with exactly one Driver
3. THE Monitoring_System SHALL display School contact information for communication purposes
4. THE Monitoring_System SHALL display Driver contact information for coordination purposes
5. WHEN a Driver is reassigned, THE Monitoring_System SHALL update the Delivery_Record association

### Requirement 12: Support Multiple Daily Deliveries

**User Story:** As a logistics coordinator, I want to track multiple deliveries to different schools on the same date, so that I can manage the complete daily delivery schedule.

#### Acceptance Criteria

1. THE Monitoring_System SHALL support multiple Delivery_Records for a single date
2. THE Monitoring_System SHALL distinguish between Delivery_Records for different Schools
3. THE Monitoring_System SHALL allow filtering Delivery_Records by date
4. THE Monitoring_System SHALL allow filtering Delivery_Records by School
5. THE Monitoring_System SHALL allow filtering Delivery_Records by status

### Requirement 13: Track Driver Location Updates

**User Story:** As a logistics coordinator, I want to track driver location changes during delivery and collection, so that I can provide accurate delivery estimates to schools.

#### Acceptance Criteria

1. WHEN a Driver updates location to "in transit", THE Monitoring_System SHALL record status "Diperjalanan"
2. WHEN a Driver updates location to "arrived at school", THE Monitoring_System SHALL record status "Sudah sampai sekolah"
3. WHEN a Driver updates location to "heading to school" for collection, THE Monitoring_System SHALL record status "Driver menuju sekolah"
4. WHEN a Driver updates location to "arrived at school" for collection, THE Monitoring_System SHALL record status "Driver sampai di sekolah"
5. THE Monitoring_System SHALL record timestamp for each Driver location update

### Requirement 14: Validate Status Transition Sequence

**User Story:** As a system administrator, I want to ensure status transitions follow the correct sequence, so that data integrity is maintained and invalid states are prevented.

#### Acceptance Criteria

1. WHEN a Status_Transition is requested, THE Monitoring_System SHALL verify the current status allows the transition
2. IF a Status_Transition violates the sequence, THEN THE Monitoring_System SHALL reject the transition and return an error message
3. THE Monitoring_System SHALL enforce that delivery stages (1-8) occur before collection stages (9-13)
4. THE Monitoring_System SHALL enforce that collection stages (9-13) occur before cleaning stages (14-15)
5. THE Monitoring_System SHALL allow administrators to manually override status for error correction

### Requirement 15: Display Delivery Summary Statistics

**User Story:** As a logistics manager, I want to see summary statistics for daily deliveries, so that I can quickly assess overall progress and identify delays.

#### Acceptance Criteria

1. THE Monitoring_System SHALL display total count of Delivery_Records for the selected date
2. THE Monitoring_System SHALL display count of Delivery_Records at each lifecycle stage
3. THE Monitoring_System SHALL display count of completed deliveries (status "Sudah diterima pihak sekolah")
4. THE Monitoring_System SHALL display count of Ompreng in cleaning process
5. THE Monitoring_System SHALL display count of Ompreng with completed cleaning

