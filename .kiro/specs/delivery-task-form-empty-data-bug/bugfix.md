# Bugfix Requirements Document

## Introduction

The delivery task creation form displays "Tidak ada driver yang tersedia" (No drivers available) and "Tidak ada order yang siap kirim" (No orders ready for delivery) despite the backend database containing 2 active drivers and 2 delivery records with status "selesai_dipacking". A check script has verified that the data exists in the database, but the frontend form is not receiving or displaying this data when a date is selected.

This bug prevents users from creating delivery tasks even when drivers and packed orders are available, blocking the core delivery workflow.

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN a user selects a delivery date in the form THEN the system displays "Tidak ada order yang siap kirim" despite delivery records with status "selesai_dipacking" existing in the database for that date

1.2 WHEN a user selects a delivery date in the form THEN the system displays "Tidak ada driver yang tersedia" despite active drivers with role "driver" existing in the database

1.3 WHEN the frontend calls `/delivery-tasks/ready-orders` endpoint with a valid date parameter THEN the response does not contain the expected orders data in the format the frontend expects

1.4 WHEN the frontend calls `/delivery-tasks/available-drivers` endpoint with a valid date parameter THEN the response does not contain the expected drivers data in the format the frontend expects

### Expected Behavior (Correct)

2.1 WHEN a user selects a delivery date in the form AND there are delivery records with status "selesai_dipacking" for that date THEN the system SHALL display those orders in the "Order/Menu yang Siap Kirim" dropdown with school name, menu item name, and portions

2.2 WHEN a user selects a delivery date in the form AND there are active drivers not assigned on that date THEN the system SHALL display those drivers in the "Driver" dropdown with their full names

2.3 WHEN the frontend calls `/delivery-tasks/ready-orders` endpoint with a valid date parameter THEN the backend SHALL return a response with structure `{ success: true, orders: [...] }` where orders is an array of ReadyOrderResponse objects

2.4 WHEN the frontend calls `/delivery-tasks/available-drivers` endpoint with a valid date parameter THEN the backend SHALL return a response with structure `{ success: true, drivers: [...] }` where drivers is an array of AvailableDriverResponse objects

2.5 WHEN the API responses are received by the frontend THEN the system SHALL correctly extract the data from `response.data.orders` and `response.data.drivers` and populate the respective dropdown options

### Unchanged Behavior (Regression Prevention)

3.1 WHEN a user selects a delivery date with no packed orders THEN the system SHALL CONTINUE TO display the warning message "Tidak ada order yang siap kirim" with appropriate description

3.2 WHEN a user selects a delivery date with no available drivers THEN the system SHALL CONTINUE TO display the warning message "Tidak ada driver yang tersedia" with appropriate description

3.3 WHEN a user has not yet selected a delivery date THEN the system SHALL CONTINUE TO display the info message "Pilih tanggal pengiriman terlebih dahulu" and keep the order dropdown disabled

3.4 WHEN a user selects an order from the dropdown THEN the system SHALL CONTINUE TO display the order details (school name, portions, menu item, status) in the information panel below the dropdown

3.5 WHEN the form is submitted with valid data THEN the system SHALL CONTINUE TO create the delivery task successfully and refresh the task list

3.6 WHEN API calls fail due to network or server errors THEN the system SHALL CONTINUE TO display appropriate error messages to the user
