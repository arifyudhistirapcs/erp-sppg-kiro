# Integration Tests - ERP SPPG System

This directory contains comprehensive integration tests for the ERP SPPG system, covering critical business workflows and data consistency across modules.

## Test Suites

### 1. Production Workflow Test (`production_workflow_test.go`)
Tests the complete production cycle from menu planning to delivery:
- **Menu Planning → Cooking → Packing → Delivery**
- Verifies data consistency across Recipe, Menu Planning, KDS, and Logistics modules
- Tests inventory deduction during cooking process
- Validates ompreng tracking throughout delivery
- Includes error handling scenarios

**Key Test Cases:**
- `TestCompleteProductionWorkflow`: End-to-end production cycle
- `TestWorkflowDataIntegrity`: Data consistency validation
- `TestWorkflowErrorHandling`: Error scenario handling

### 2. Procurement Workflow Test (`procurement_workflow_test.go`)
Tests the complete procurement cycle with automatic triggers:
- **PO Creation → Approval → GRN → Inventory Update → Cash Flow Entry**
- Verifies automatic data propagation between Supply Chain and Financial modules
- Tests FIFO/FEFO inventory management
- Validates low stock alert generation and resolution
- Includes quantity discrepancy handling

**Key Test Cases:**
- `TestCompleteProcurementWorkflow`: End-to-end procurement cycle
- `TestProcurementWorkflowWithDiscrepancies`: Quantity mismatch handling
- `TestProcurementWorkflowErrorHandling`: Authorization and validation errors
- `TestLowStockAlertGeneration`: Stock alert lifecycle
- `TestFIFOInventoryMethod`: Inventory method validation

### 3. Offline-Online Cycle Test (`offline_online_cycle_test.go`)
Tests PWA offline data capture and synchronization:
- **Offline Data Capture → Sync → Backend Verification**
- Tests e-POD offline capture and sync
- Validates attendance offline-online cycle
- Tests ompreng tracking sync
- Includes conflict resolution and partial sync failure handling

**Key Test Cases:**
- `TestCompleteOfflineOnlineCycle`: Full offline-online workflow
- `TestConflictResolution`: Data conflict handling
- `TestPartialSyncFailure`: Partial sync error recovery
- `TestDataIntegrityDuringSync`: Data consistency during sync
- `TestSyncStatusTracking`: Sync metadata handling

## Running the Tests

```bash
# Run all integration tests
go test ./tests/integration/... -v

# Run specific test suite
go test ./tests/integration/production_workflow_test.go -v
go test ./tests/integration/procurement_workflow_test.go -v
go test ./tests/integration/offline_online_cycle_test.go -v

# Run with coverage
go test ./tests/integration/... -v -cover
```

## Test Data Setup

Each test suite includes:
- **Master Data Setup**: Creates required entities (users, ingredients, recipes, schools, suppliers)
- **Authentication**: Sets up test users with appropriate roles
- **Database Cleanup**: Ensures clean state between tests
- **Mock Data**: Realistic test data reflecting actual business scenarios

## Validation Coverage

The integration tests validate:

### Data Consistency
- Cross-module data integrity
- Automatic trigger execution
- Referential integrity maintenance
- Transaction rollback on failures

### Business Logic
- Role-based access control
- Workflow state transitions
- Calculation accuracy (nutrition, inventory, financial)
- Alert generation and resolution

### System Behavior
- Real-time data synchronization
- Offline capability and sync
- Error handling and recovery
- Conflict resolution strategies

## Requirements Validation

These tests validate the following requirements:
- **Requirements 2.1-5.6**: Production workflow (Menu → Cooking → Packing → Delivery)
- **Requirements 7.1-9.6, 17.4**: Procurement workflow (PO → Approval → GRN → Inventory → Cash flow)
- **Requirements 23.1-23.6**: PWA offline-online cycle functionality

## Notes

- Tests use in-memory database for isolation
- Firebase integration is mocked for testing
- File uploads are simulated with test URLs
- All tests include comprehensive assertions for data validation
- Error scenarios are tested to ensure system robustness