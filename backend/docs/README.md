# Backend Documentation

This directory contains technical documentation for the backend API and system components.

## Available Documentation

### API Documentation

- **[Portion Size Differentiation API](./API_PORTION_SIZE_DIFFERENTIATION.md)** - Complete API documentation for the portion size differentiation feature, including:
  - Request/response formats
  - Validation rules and error messages
  - Example use cases
  - Database schema details
  - Best practices
- **[Portion Size Quick Reference](./PORTION_SIZE_QUICK_REFERENCE.md)** - Quick reference guide with:
  - School types and portion rules
  - Validation rules summary
  - Common error messages
  - Example calculations
  - Testing checklist

### Postman Collections

- **[Portion Size Differentiation API Collection](./Portion_Size_Differentiation_API.postman_collection.json)** - Postman collection with 17 pre-configured requests for testing the portion size differentiation API:
  - Create menu items with various portion size combinations
  - Get and update menu items
  - Test all validation error scenarios
  - Delete menu items
  - KDS cooking view endpoints
  - KDS packing view endpoints
  - Status update endpoints

### System Documentation

- **[KDS Data Consistency Verification](./KDS_DATA_CONSISTENCY_VERIFICATION.md)** - Documentation for KDS (Kitchen Display System) data consistency checks

## Quick Start

### Using the Postman Collection

1. Import the collection into Postman:
   - Open Postman
   - Click "Import" button
   - Select `Portion_Size_Differentiation_API.postman_collection.json`

2. Configure environment variables:
   - `base_url`: Your API base URL (default: `http://localhost:8080/api`)
   - `jwt_token`: Your authentication token
   - `menu_plan_id`: ID of an existing menu plan
   - `sd_school_id`: ID of an SD (elementary) school
   - `smp_school_id`: ID of an SMP (junior high) school
   - `recipe_id`: ID of an existing recipe
   - `test_date`: Date for testing KDS endpoints (default: `2024-01-15`)

3. Run the requests in order to test the complete workflow

### API Overview

The Portion Size Differentiation feature enables menu planners to allocate meals based on student age groups:

- **Small portions**: For SD (elementary school) grades 1-3
- **Large portions**: For SD grades 4-6, SMP (junior high), and SMA (senior high) students

**Key Endpoints:**
- `POST /api/menu-plans/:id/items` - Create menu item with portion size allocations
- `GET /api/menu-plans/:id/items/:item_id` - Get menu item with portion size breakdown
- `PUT /api/menu-plans/:id/items/:item_id` - Update menu item allocations
- `DELETE /api/menu-plans/:id/items/:item_id` - Delete menu item
- `GET /api/v1/kds/cooking/today` - Get today's cooking menu with portion sizes
- `GET /api/v1/kds/packing/today` - Get today's packing allocations with portion sizes
- `PUT /api/v1/kds/cooking/:recipe_id/status` - Update cooking status
- `PUT /api/v1/kds/packing/:school_id/status` - Update packing status

## Related Documentation

- [Database Migrations](../migrations/) - Database schema changes and migration scripts
- [Spec Documents](../../.kiro/specs/portion-size-differentiation/) - Feature requirements, design, and tasks

## Contributing

When adding new features or endpoints, please:

1. Create comprehensive API documentation following the format in `API_PORTION_SIZE_DIFFERENTIATION.md`
2. Include request/response examples with realistic data
3. Document all validation rules and error messages
4. Create a Postman collection with test cases
5. Update this README with links to new documentation
