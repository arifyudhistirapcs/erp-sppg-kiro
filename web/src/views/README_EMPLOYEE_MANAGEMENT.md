# Employee Management Module

## Overview

The Employee Management module provides comprehensive functionality for managing employee data and user accounts in the ERP SPPG system. This module implements the requirements specified in Requirements 14.1-14.6.

## Components

### 1. EmployeeListView.vue
- **Purpose**: Main listing page for employee management
- **Features**:
  - Search by NIK, name, or email
  - Filter by status (active/inactive) and position
  - Statistics dashboard showing employee counts
  - Table view with sorting and pagination
  - Create, edit, view, and deactivate employees
  - Auto-generated credentials display for new employees

### 2. EmployeeFormView.vue
- **Purpose**: Form for creating and editing employees
- **Features**:
  - Separate cards for personal and job information
  - Validation for NIK (16 digits), email format, phone number
  - Position and role selection
  - Date picker for join date
  - Auto-generated login credentials for new employees
  - Unique NIK and email validation

### 3. employeeService.js
- **Purpose**: API service for employee operations
- **Methods**:
  - `getEmployees(params)` - Fetch employees with filters
  - `getEmployeeById(id)` - Get single employee
  - `createEmployee(data)` - Create new employee
  - `updateEmployee(id, data)` - Update employee
  - `deactivateEmployee(id)` - Deactivate employee
  - `getEmployeeStats()` - Get employee statistics

## Key Features Implemented

### ✅ Unique NIK and Email Validation (Req 14.2)
- Frontend validation with proper error messages
- Backend validation prevents duplicates
- NIK format validation (16 digits)

### ✅ Auto-generated Login Credentials (Req 14.3)
- System generates secure random password
- Credentials displayed once after creation
- User account automatically created with employee

### ✅ Employee Deactivation (Req 14.5)
- Soft delete - preserves historical data
- Prevents login but maintains audit trail
- Toggle between active/inactive status

### ✅ Search and Filter Functionality
- Search by NIK, name, or email
- Filter by status and position
- Real-time search with debouncing
- Pagination support

### ✅ Role-Based Access Control
- Only Kepala SPPG and Akuntan can manage employees
- Different permissions for create/edit operations
- Secure API endpoints with authentication

## Data Validation

### Frontend Validation
- **NIK**: Required, exactly 16 digits, numeric only
- **Full Name**: Required, minimum 2 characters
- **Email**: Required, valid email format, unique
- **Phone**: Required, Indonesian mobile format (08xxxxxxxxxx)
- **Position**: Required selection from predefined list
- **Role**: Required selection from system roles
- **Join Date**: Required date selection

### Backend Validation
- Unique constraints on NIK and email
- Password hashing with secure algorithms
- Transaction-based operations for data consistency

## Position and Role Mapping

| Position | System Role | Description |
|----------|-------------|-------------|
| Kepala SPPG | kepala_sppg | Full system access |
| Akuntan | akuntan | Financial and HR access |
| Ahli Gizi | ahli_gizi | Recipe and menu management |
| Pengadaan | pengadaan | Supply chain management |
| Chef | chef | Kitchen operations |
| Packing | packing | Packing operations |
| Driver | driver | Delivery operations |
| Asisten Lapangan | asisten | Field assistance |

## API Endpoints

```
GET    /api/v1/employees              - List employees with filters
POST   /api/v1/employees              - Create new employee
GET    /api/v1/employees/:id          - Get employee by ID
PUT    /api/v1/employees/:id          - Update employee
DELETE /api/v1/employees/:id          - Deactivate employee
GET    /api/v1/employees/stats        - Get employee statistics
```

## Testing

### Unit Tests
- **employeeService.test.js**: Tests all API service methods
- **EmployeeListView.test.js**: Tests Vue component functionality

### Test Coverage
- Service layer: 100% method coverage
- Component layer: Core functionality tested
- Validation logic: All validation rules tested

## Usage Examples

### Creating a New Employee
1. Navigate to `/employees`
2. Click "Tambah Karyawan" button
3. Fill in required information
4. System validates unique NIK and email
5. Auto-generated credentials displayed
6. Employee account created and activated

### Searching Employees
```javascript
// Search by name
searchText.value = 'John Doe'
handleSearch()

// Filter by status
filterStatus.value = 'active'
handleSearch()

// Filter by position
filterPosition.value = 'Chef'
handleSearch()
```

### Deactivating Employee
```javascript
// Deactivate employee (soft delete)
await employeeService.deactivateEmployee(employeeId)
// Employee marked as inactive, login disabled
// Historical data preserved for audit trail
```

## Security Considerations

1. **Password Security**: Auto-generated passwords are cryptographically secure
2. **Data Validation**: All inputs validated on both client and server
3. **Access Control**: Role-based permissions enforced
4. **Audit Trail**: All employee changes logged
5. **Soft Delete**: Deactivation preserves data for compliance

## Future Enhancements

1. **Bulk Operations**: Import/export employee data
2. **Photo Upload**: Employee profile pictures
3. **Advanced Search**: Search by multiple criteria
4. **Employee Self-Service**: Allow employees to update their own info
5. **Integration**: Connect with attendance and payroll systems