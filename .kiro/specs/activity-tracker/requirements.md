# Requirements Document

## Introduction

The Activity Tracker (Aktivitas Pelacakan) is a standalone monitoring module that enables Kepala SPPG (Head of SPPG) to monitor and track the complete lifecycle of menu orders from initiation through preparation, packing, delivery, and ompreng (food container) return and cleaning across 16 distinct stages. This module provides real-time visibility into the operational workflow with a vertical timeline interface showing each stage's status and timestamp, integrating with existing KDS modules and logistics operations.

## Glossary

- **Activity_Tracker**: Standalone module (Aktivitas Pelacakan) for tracking complete menu order lifecycle from initiation to completion
- **Kepala_SPPG**: Head of SPPG (Satuan Pendidikan Pelaksana Gizi) - the primary user role
- **Menu_Order**: A complete menu order instance for a specific school on a specific date, tracked from initiation through all 16 stages
- **Menu**: Food items prepared for delivery to schools
- **Ompreng**: Food containers used for menu delivery that must be returned and cleaned
- **Stage**: One of 16 distinct phases in the menu order lifecycle (initiation, preparation, cooking, packing, delivery, collection, cleaning)
- **Status_Transition**: A change from one stage to another with associated timestamp
- **Vertical_Timeline**: Visual representation showing all 16 stages in vertical layout with status indicators and timestamps
- **Activity_Entry**: Individual stage entry showing status, description, timestamp, and optional media (photo/video)
- **KDS_Module**: Kitchen Display System modules (Cooking, Packing, Cleaning) that trigger status updates
- **Firebase**: Real-time database service used for status synchronization
- **School_Entity**: Educational institution receiving menu deliveries
- **Driver_Entity**: Personnel responsible for menu delivery and ompreng collection

## Requirements

### Requirement 1: Display Vertical Timeline Visualization

**User Story:** As Kepala SPPG, I want to see a vertical timeline of all 16 lifecycle stages with timestamps, so that I can quickly understand the current progress and history of any menu order.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL display all 16 stages in sequential vertical order for each Menu_Order
2. WHEN a Stage is completed, THE Activity_Tracker SHALL display the Stage with a completed status indicator (filled circle/checkmark)
3. WHEN a Stage is in progress, THE Activity_Tracker SHALL display the Stage with an in-progress status indicator (filled circle)
4. WHEN a Stage is pending, THE Activity_Tracker SHALL display the Stage with a pending status indicator (empty circle)
5. THE Activity_Tracker SHALL display stage names in Indonesian language matching the 16 defined stages
6. THE Activity_Tracker SHALL display timestamp for each completed stage in format "Day, HH:MM - Day, HH:MM" (e.g., "Rabu, 13:49 - Rabu, 13:50")
7. THE Activity_Tracker SHALL display stage description below the stage name
8. THE Activity_Tracker SHALL connect stages with vertical lines to show progression

### Requirement 2: Track Menu Order Initiation and Preparation Stages

**User Story:** As Kepala SPPG, I want to track menu order from initiation through preparation and cooking, so that I can monitor the complete order lifecycle from the beginning.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL track Stage 1 "Order sedang disiapkan" (Order being prepared) - Menu order is initiated and being prepared for cooking
2. THE Activity_Tracker SHALL track Stage 2 "Order sedang dimasak" (Order being cooked) - Menu is actively being cooked
3. THE Activity_Tracker SHALL track Stage 3 "Order sedang dikemas" (Order being packed) - Menu cooking completed and being packed
4. THE Activity_Tracker SHALL track Stage 4 "Order siap diambil" (Order ready for pickup) - Menu packing completed and ready for delivery

### Requirement 3: Track Delivery Transit Stages

**User Story:** As Kepala SPPG, I want to track menu delivery from departure through arrival at schools, so that I can monitor delivery operations in real-time.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL track Stage 5 "Pesanan dalam perjalanan" (Order in transit) - Driver departed with menu to school
2. THE Activity_Tracker SHALL track Stage 6 "Pesanan sudah tiba" (Order arrived) - Driver arrived at school with menu
3. THE Activity_Tracker SHALL track Stage 7 "Pesanan sudah diterima" (Order received) - Menu received and confirmed by school

### Requirement 4: Track Ompreng Collection Stages

**User Story:** As Kepala SPPG, I want to track ompreng collection from schools back to SPPG, so that I can monitor container return operations and ensure all containers are accounted for.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL track Stage 8 "Driver menuju lokasi pengambilan" (Driver heading to pickup location) - Driver assigned and traveling to school to collect ompreng
2. THE Activity_Tracker SHALL track Stage 9 "Driver sudah tiba di lokasi" (Driver arrived at location) - Driver arrived at school for ompreng collection
3. THE Activity_Tracker SHALL track Stage 10 "Driver dalam perjalanan kembali" (Driver returning) - Driver in transit back to SPPG with ompreng
4. THE Activity_Tracker SHALL track Stage 11 "Driver sudah tiba di SPPG" (Driver arrived at SPPG) - Driver returned to SPPG with ompreng

### Requirement 5: Track Ompreng Cleaning Stages

**User Story:** As Kepala SPPG, I want to track ompreng cleaning operations through completion, so that I can monitor container hygiene management and ensure readiness for next use.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL track Stage 12 "Ompreng siap dicuci" (Ompreng ready for cleaning) - Ompreng received and queued for cleaning
2. THE Activity_Tracker SHALL track Stage 13 "Ompreng sedang dicuci" (Ompreng being cleaned) - Cleaning in progress
3. THE Activity_Tracker SHALL track Stage 14 "Ompreng selesai dicuci" (Ompreng cleaning completed) - Cleaning completed and containers ready for reuse
4. THE Activity_Tracker SHALL track Stage 15 "Ompreng siap digunakan kembali" (Ompreng ready for reuse) - Containers sanitized and stored for next order
5. THE Activity_Tracker SHALL track Stage 16 "Order selesai" (Order completed) - Complete order lifecycle finished

### Requirement 6: Display Order Header Information

**User Story:** As Kepala SPPG, I want to see order summary at the top of the timeline, so that I can identify which order I'm viewing.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL display "Aktivitas Pelacakan" as the page title
2. THE Activity_Tracker SHALL display menu item photo/thumbnail at the top
3. THE Activity_Tracker SHALL display menu item name
4. THE Activity_Tracker SHALL display school name and delivery date
5. THE Activity_Tracker SHALL display portion quantity (e.g., "150 porsi")
6. THE Activity_Tracker SHALL display driver name and vehicle information
7. THE Activity_Tracker SHALL display current order status prominently

### Requirement 7: Display Stage Details with Media Support

**User Story:** As Kepala SPPG, I want to see detailed information for each stage including photos or videos when available, so that I have visual confirmation of progress.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL display stage title in Indonesian
2. THE Activity_Tracker SHALL display stage description text
3. THE Activity_Tracker SHALL display photo thumbnail when photo is attached to stage
4. THE Activity_Tracker SHALL display video play button when video is attached to stage
5. THE Activity_Tracker SHALL allow clicking photo to view full size
6. THE Activity_Tracker SHALL allow clicking video to play
7. THE Activity_Tracker SHALL display "Lapor MBG" (Report to MBG) button when applicable for certain stages

### Requirement 8: Filter and Search Orders

**User Story:** As Kepala SPPG, I want to filter and search orders by date and school, so that I can quickly find specific orders to monitor.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL provide a date filter control on the order list page
2. WHEN Kepala_SPPG selects a date, THE Activity_Tracker SHALL display only Menu_Orders for that date
3. THE Activity_Tracker SHALL display Menu_Orders for the current date by default
4. THE Activity_Tracker SHALL allow selection of past and future dates
5. THE Activity_Tracker SHALL provide a school filter dropdown
6. THE Activity_Tracker SHALL provide a search box to search by menu name or school name
7. THE Activity_Tracker SHALL display order count for selected filters

### Requirement 9: Display Order List with Status Summary

**User Story:** As Kepala SPPG, I want to see a list of all orders with their current status, so that I can select which order to view in detail.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL display a list of all Menu_Orders for selected date
2. THE Activity_Tracker SHALL display menu photo thumbnail for each order in the list
3. THE Activity_Tracker SHALL display menu name for each order
4. THE Activity_Tracker SHALL display school name for each order
5. THE Activity_Tracker SHALL display current stage status for each order
6. THE Activity_Tracker SHALL display portion quantity for each order
7. THE Activity_Tracker SHALL allow clicking an order to view its detailed timeline
8. THE Activity_Tracker SHALL display status badge with color coding (in-progress: blue, completed: green, pending: gray)

### Requirement 10: Receive Real-Time Status Updates

**User Story:** As Kepala SPPG, I want to see status updates in real-time, so that I always have current information without refreshing.

#### Acceptance Criteria

1. WHEN a Status_Transition occurs in Firebase, THE Activity_Tracker SHALL update the Vertical_Timeline within 2 seconds
2. WHEN a Status_Transition occurs in Firebase, THE Activity_Tracker SHALL update the order list within 2 seconds
3. THE Activity_Tracker SHALL maintain a persistent connection to Firebase for real-time updates
4. IF the Firebase connection is lost, THEN THE Activity_Tracker SHALL attempt to reconnect automatically
5. THE Activity_Tracker SHALL display a notification badge when new status updates occur while viewing a different order

### Requirement 11: Integrate with KDS Cooking Module

**User Story:** As Kepala SPPG, I want cooking status updates to automatically reflect in the Activity Tracker, so that I don't need manual data entry.

#### Acceptance Criteria

1. WHEN KDS_Module Cooking starts cooking, THE Activity_Tracker SHALL transition to Stage 2 "Order sedang dimasak"
2. WHEN KDS_Module Cooking completes cooking, THE Activity_Tracker SHALL transition to Stage 3 "Order sedang dikemas"
3. WHEN KDS_Module Packing completes packing, THE Activity_Tracker SHALL transition to Stage 4 "Order siap diambil"

### Requirement 12: Integrate with Logistics/Delivery Module

**User Story:** As Kepala SPPG, I want delivery status updates to automatically reflect in the Activity Tracker, so that I have accurate delivery progress.

#### Acceptance Criteria

1. WHEN driver departs for delivery, THE Activity_Tracker SHALL transition to Stage 5 "Pesanan dalam perjalanan"
2. WHEN driver arrives at school, THE Activity_Tracker SHALL transition to Stage 6 "Pesanan sudah tiba"
3. WHEN school confirms receipt, THE Activity_Tracker SHALL transition to Stage 7 "Pesanan sudah diterima"
4. WHEN driver is assigned to collect ompreng, THE Activity_Tracker SHALL transition to Stage 8 "Driver menuju lokasi pengambilan"
5. WHEN driver arrives at school for collection, THE Activity_Tracker SHALL transition to Stage 9 "Driver sudah tiba di lokasi"
6. WHEN driver departs with ompreng, THE Activity_Tracker SHALL transition to Stage 10 "Driver dalam perjalanan kembali"
7. WHEN driver arrives at SPPG, THE Activity_Tracker SHALL transition to Stage 11 "Driver sudah tiba di SPPG"

### Requirement 13: Integrate with KDS Cleaning Module

**User Story:** As Kepala SPPG, I want cleaning status updates to automatically reflect in the Activity Tracker, so that I can monitor ompreng hygiene operations.

#### Acceptance Criteria

1. WHEN ompreng is queued for cleaning, THE Activity_Tracker SHALL transition to Stage 12 "Ompreng siap dicuci"
2. WHEN KDS_Module Cleaning starts cleaning, THE Activity_Tracker SHALL transition to Stage 13 "Ompreng sedang dicuci"
3. WHEN KDS_Module Cleaning completes cleaning, THE Activity_Tracker SHALL transition to Stage 14 "Ompreng selesai dicuci"
4. WHEN ompreng is stored for reuse, THE Activity_Tracker SHALL transition to Stage 15 "Ompreng siap digunakan kembali"
5. WHEN all stages complete, THE Activity_Tracker SHALL transition to Stage 16 "Order selesai"

### Requirement 14: Control Access by Role

**User Story:** As a system administrator, I want to restrict Activity Tracker access to authorized roles, so that sensitive operational data is protected.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL allow access to users with Kepala_SPPG role
2. WHERE a user has management role permissions, THE Activity_Tracker SHALL allow view-only access
3. IF a user does not have authorized role, THEN THE Activity_Tracker SHALL deny access and display an authorization error
4. THE Activity_Tracker SHALL be accessible as a standalone module in the main navigation menu

### Requirement 15: Display Multiple Orders with Status Overview

**User Story:** As Kepala SPPG, I want to see all orders for a selected date with status overview, so that I can monitor multiple schools simultaneously.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL display all Menu_Orders for the selected date in a list view
2. THE Activity_Tracker SHALL display each Menu_Order as a card with photo, name, school, and current status
3. THE Activity_Tracker SHALL allow Kepala_SPPG to click any order to view its detailed vertical timeline
4. THE Activity_Tracker SHALL display a summary count of total Menu_Orders for the selected date
5. THE Activity_Tracker SHALL display status distribution (e.g., "5 sedang dimasak, 3 dalam perjalanan, 2 selesai")
6. THE Activity_Tracker SHALL provide a back button from detail view to return to order list

### Requirement 16: Handle Stage Transition Validation

**User Story:** As Kepala SPPG, I want the system to ensure stages progress logically, so that data integrity is maintained.

#### Acceptance Criteria

1. THE Activity_Tracker SHALL record Status_Transitions in sequential stage order (1→2→3...→16)
2. IF a Status_Transition skips a stage, THEN THE Activity_Tracker SHALL log a validation warning
3. THE Activity_Tracker SHALL allow manual correction of Status_Transitions by Kepala_SPPG with proper authorization
4. WHEN a Status_Transition is corrected, THE Activity_Tracker SHALL record the correction with a correction indicator and notes
5. THE Activity_Tracker SHALL prevent backward transitions (e.g., from Stage 5 back to Stage 3) unless explicitly authorized
