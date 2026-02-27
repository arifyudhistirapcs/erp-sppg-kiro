# API Documentation: Portion Size Differentiation

## Overview

This document describes the API endpoints for the Portion Size Differentiation feature, which enables menu planners to allocate meals based on student age groups. The system distinguishes between:

- **Small portions**: For SD (elementary school) grades 1-3
- **Large portions**: For SD grades 4-6, SMP (junior high), and SMA (senior high) students

## Base URL

All endpoints are relative to: `/api/menu-plans/:id`

Where `:id` is the menu plan ID.

## Authentication

All endpoints require authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

---

## Endpoints

### 1. Create Menu Item with Portion Size Allocations

Creates a new menu item with school allocations that support portion size differentiation.

**Endpoint:** `POST /api/menu-plans/:id/items`

**URL Parameters:**
- `id` (required): Menu plan ID (integer)

#### Request Body

```json
{
  "date": "2024-01-15",
  "recipe_id": 5,
  "portions": 500,
  "school_allocations": [
    {
      "school_id": 1,
      "portions_small": 150,
      "portions_large": 200
    },
    {
      "school_id": 2,
      "portions_small": 0,
      "portions_large": 150
    }
  ]
}
```

#### Request Fields

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `date` | string | Yes | YYYY-MM-DD or ISO 8601 format | Date for the menu item |
| `recipe_id` | integer | Yes | Must exist in database | ID of the recipe to use |
| `portions` | integer | Yes | Must be > 0 | Total portions for this menu item |
| `school_allocations` | array | Yes | Min 1 item | Array of school allocations |

#### School Allocation Fields

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `school_id` | integer | Yes | Must exist in database | ID of the school |
| `portions_small` | integer | No | Must be >= 0 | Small portions (SD grades 1-3) |
| `portions_large` | integer | No | Must be >= 0 | Large portions (SD grades 4-6, SMP, SMA) |

#### Validation Rules

1. **Sum Validation**: The sum of all `portions_small` and `portions_large` across all schools must equal `portions`
2. **SMP/SMA Restriction**: SMP and SMA schools cannot have `portions_small` > 0 (must be 0)
3. **At Least One Portion**: Each school must have at least one of `portions_small` or `portions_large` > 0
4. **Non-Negative**: Both `portions_small` and `portions_large` must be >= 0
5. **No Duplicates**: Each `school_id` can only appear once in the allocations array

#### Success Response (201 Created)

```json
{
  "success": true,
  "data": {
    "id": 42,
    "menu_plan_id": 10,
    "date": "2024-01-15T00:00:00Z",
    "recipe_id": 5,
    "portions": 500,
    "recipe": {
      "id": 5,
      "name": "Nasi Goreng",
      "category": "main_course"
    },
    "school_allocations": [
      {
        "id": 101,
        "menu_item_id": 42,
        "school_id": 1,
        "school_name": "SD Negeri 1",
        "portions": 150,
        "date": "2024-01-15"
      },
      {
        "id": 102,
        "menu_item_id": 42,
        "school_id": 1,
        "school_name": "SD Negeri 1",
        "portions": 200,
        "date": "2024-01-15"
      },
      {
        "id": 103,
        "menu_item_id": 42,
        "school_id": 2,
        "school_name": "SMP Negeri 1",
        "portions": 150,
        "date": "2024-01-15"
      }
    ]
  }
}
```

**Note**: For SD schools with both small and large portions, two separate allocation records are created in the database (one for each portion size).

#### Error Responses

**400 Bad Request - Invalid ID**
```json
{
  "success": false,
  "error_code": "INVALID_ID",
  "message": "ID tidak valid"
}
```

**400 Bad Request - Validation Error (Sum Mismatch)**
```json
{
  "success": false,
  "error_code": "VALIDATION_ERROR",
  "message": "Validasi gagal",
  "details": {
    "field": "school_allocations",
    "error": "sum of allocated portions (450) does not equal total portions (500)"
  }
}
```

**400 Bad Request - SMP/SMA Cannot Have Small Portions**
```json
{
  "success": false,
  "error_code": "VALIDATION_ERROR",
  "message": "Validasi gagal",
  "details": {
    "field": "school_allocations",
    "error": "SMP schools cannot have small portions"
  }
}
```

**400 Bad Request - At Least One Portion Required**
```json
{
  "success": false,
  "error_code": "VALIDATION_ERROR",
  "message": "Validasi gagal",
  "details": {
    "field": "school_allocations",
    "error": "school must have at least one portion: school_id 1"
  }
}
```

**400 Bad Request - Duplicate School**
```json
{
  "success": false,
  "error_code": "VALIDATION_ERROR",
  "message": "Validasi gagal",
  "details": {
    "field": "school_allocations",
    "error": "duplicate allocation for school_id 1"
  }
}
```

**400 Bad Request - Invalid Date Format**
```json
{
  "success": false,
  "error_code": "INVALID_DATE",
  "message": "Format tanggal tidak valid (gunakan YYYY-MM-DD atau ISO format)"
}
```

**500 Internal Server Error**
```json
{
  "success": false,
  "error_code": "INTERNAL_ERROR",
  "message": "Terjadi kesalahan pada server"
}
```

---

### 2. Get Menu Item with Portion Size Breakdown

Retrieves a menu item with school allocations grouped by school, showing portion size breakdown.

**Endpoint:** `GET /api/menu-plans/:id/items/:item_id`

**URL Parameters:**
- `id` (required): Menu plan ID (integer)
- `item_id` (required): Menu item ID (integer)

#### Success Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": 42,
    "menu_plan_id": 10,
    "date": "2024-01-15T00:00:00Z",
    "recipe_id": 5,
    "portions": 500,
    "recipe": {
      "id": 5,
      "name": "Nasi Goreng"
    },
    "school_allocations": [
      {
        "school_id": 1,
        "school_name": "SD Negeri 1",
        "school_category": "SD",
        "portion_size_type": "mixed",
        "portions_small": 150,
        "portions_large": 200,
        "total_portions": 350
      },
      {
        "school_id": 2,
        "school_name": "SMP Negeri 1",
        "school_category": "SMP",
        "portion_size_type": "large",
        "portions_small": 0,
        "portions_large": 150,
        "total_portions": 150
      }
    ]
  }
}
```

#### Response Fields

**School Allocation Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `school_id` | integer | ID of the school |
| `school_name` | string | Name of the school |
| `school_category` | string | School category: "SD", "SMP", or "SMA" |
| `portion_size_type` | string | "mixed" for SD schools, "large" for SMP/SMA |
| `portions_small` | integer | Small portions (SD grades 1-3), 0 for SMP/SMA |
| `portions_large` | integer | Large portions (SD grades 4-6, SMP, SMA) |
| `total_portions` | integer | Sum of small and large portions |

**Note**: Allocations are grouped by school and sorted alphabetically by school name.

#### Error Responses

**400 Bad Request - Invalid ID**
```json
{
  "success": false,
  "error_code": "INVALID_ID",
  "message": "ID tidak valid"
}
```

**404 Not Found - Menu Item Not Found**
```json
{
  "success": false,
  "error_code": "NOT_FOUND",
  "message": "Item menu tidak ditemukan"
}
```

**404 Not Found - Menu Item Not in Menu Plan**
```json
{
  "success": false,
  "error_code": "NOT_FOUND",
  "message": "Item menu tidak ditemukan dalam menu plan yang ditentukan"
}
```

**500 Internal Server Error**
```json
{
  "success": false,
  "error_code": "INTERNAL_ERROR",
  "message": "Terjadi kesalahan pada server"
}
```

---

### 3. Update Menu Item with Portion Size Allocations

Updates an existing menu item and its school allocations with portion size support.

**Endpoint:** `PUT /api/menu-plans/:id/items/:item_id`

**URL Parameters:**
- `id` (required): Menu plan ID (integer)
- `item_id` (required): Menu item ID (integer)

#### Request Body

```json
{
  "date": "2024-01-15",
  "recipe_id": 5,
  "portions": 600,
  "school_allocations": [
    {
      "school_id": 1,
      "portions_small": 200,
      "portions_large": 250
    },
    {
      "school_id": 2,
      "portions_small": 0,
      "portions_large": 150
    }
  ]
}
```

#### Request Fields

Same as Create Menu Item endpoint (see above).

#### Validation Rules

Same as Create Menu Item endpoint (see above).

#### Success Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": 42,
    "menu_plan_id": 10,
    "date": "2024-01-15T00:00:00Z",
    "recipe_id": 5,
    "portions": 600,
    "recipe": {
      "id": 5,
      "name": "Nasi Goreng",
      "category": "main_course"
    },
    "school_allocations": [
      {
        "school_id": 1,
        "school_name": "SD Negeri 1",
        "school_category": "SD",
        "portion_size_type": "mixed",
        "portions_small": 200,
        "portions_large": 250,
        "total_portions": 450
      },
      {
        "school_id": 2,
        "school_name": "SMP Negeri 1",
        "school_category": "SMP",
        "portion_size_type": "large",
        "portions_small": 0,
        "portions_large": 150,
        "total_portions": 150
      }
    ]
  }
}
```

#### Error Responses

Same as Create Menu Item endpoint, plus:

**400 Bad Request - Invalid Menu Plan**
```json
{
  "success": false,
  "error_code": "INVALID_MENU_PLAN",
  "message": "Item tidak termasuk dalam menu plan yang ditentukan"
}
```

---

## Validation Error Reference

### Complete List of Validation Errors

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `sum of allocated portions (X) does not equal total portions (Y)` | Sum of all portions doesn't match total | Adjust allocations so sum equals total |
| `SMP schools cannot have small portions` | SMP school has portions_small > 0 | Set portions_small to 0 for SMP schools |
| `SMA schools cannot have small portions` | SMA school has portions_small > 0 | Set portions_small to 0 for SMA schools |
| `school must have at least one portion: school_id X` | Both portions_small and portions_large are 0 | Set at least one portion type > 0 |
| `small portions cannot be negative for school_id X` | portions_small < 0 | Use non-negative values |
| `large portions cannot be negative for school_id X` | portions_large < 0 | Use non-negative values |
| `duplicate allocation for school_id X` | Same school appears multiple times | Remove duplicate school entries |
| `at least one school allocation is required` | Empty school_allocations array | Add at least one school allocation |
| `school not found: school_id X` | School ID doesn't exist | Use valid school ID |

---

## Example Use Cases

### Example 1: SD School with Mixed Portions

An elementary school (SD) needs both small and large portions:

```json
{
  "date": "2024-01-15",
  "recipe_id": 5,
  "portions": 350,
  "school_allocations": [
    {
      "school_id": 1,
      "portions_small": 150,
      "portions_large": 200
    }
  ]
}
```

This creates **two allocation records** in the database:
- One record with 150 small portions (portion_size = 'small')
- One record with 200 large portions (portion_size = 'large')

### Example 2: SMP School with Large Portions Only

A junior high school (SMP) only needs large portions:

```json
{
  "date": "2024-01-15",
  "recipe_id": 5,
  "portions": 150,
  "school_allocations": [
    {
      "school_id": 2,
      "portions_small": 0,
      "portions_large": 150
    }
  ]
}
```

This creates **one allocation record** with 150 large portions (portion_size = 'large').

### Example 3: Multiple Schools with Different Portion Types

```json
{
  "date": "2024-01-15",
  "recipe_id": 5,
  "portions": 800,
  "school_allocations": [
    {
      "school_id": 1,
      "portions_small": 150,
      "portions_large": 200
    },
    {
      "school_id": 2,
      "portions_small": 0,
      "portions_large": 200
    },
    {
      "school_id": 3,
      "portions_small": 100,
      "portions_large": 150
    }
  ]
}
```

Total: 150 + 200 + 200 + 100 + 150 = 800 ✓

### Example 4: SD School with Only Large Portions

An SD school can also have only large portions (e.g., if only grades 4-6 are ordering):

```json
{
  "date": "2024-01-15",
  "recipe_id": 5,
  "portions": 200,
  "school_allocations": [
    {
      "school_id": 1,
      "portions_small": 0,
      "portions_large": 200
    }
  ]
}
```

This creates **one allocation record** with 200 large portions.

---

## Database Schema

### menu_item_school_allocations Table

The system stores portion size allocations in separate records:

| Column | Type | Description |
|--------|------|-------------|
| `id` | integer | Primary key |
| `menu_item_id` | integer | Foreign key to menu_items |
| `school_id` | integer | Foreign key to schools |
| `portions` | integer | Number of portions for this size |
| `portion_size` | varchar(10) | 'small' or 'large' |
| `date` | date | Date of the allocation |

**Key Points:**
- SD schools with both portion sizes have **2 records** (one small, one large)
- SMP/SMA schools have **1 record** (large only)
- The `portion_size` field has a CHECK constraint: `portion_size IN ('small', 'large')`
- Composite index on `(menu_item_id, school_id, portion_size)` for performance

---

## Migration Notes

### Backward Compatibility

Existing allocations without portion size data have been migrated with `portion_size = 'large'`. The system maintains full backward compatibility with existing data.

### Data Migration

All existing allocation records were updated with:
- `portion_size = 'large'` for all schools (SD, SMP, SMA)
- This ensures continuity with historical data

---

## Best Practices

1. **Always validate sum**: Ensure the sum of all portions equals the total before submitting
2. **Use 0 for unused portion types**: Set `portions_small: 0` for SMP/SMA schools
3. **Check school category**: Verify school type before allocating small portions
4. **Handle grouped responses**: When retrieving data, remember that SD schools may have both portion types combined
5. **Transaction safety**: All allocation operations are atomic - either all succeed or all fail

---

## Related Documentation

- [Requirements Document](../../.kiro/specs/portion-size-differentiation/requirements.md)
- [Design Document](../../.kiro/specs/portion-size-differentiation/design.md)
- [Database Migration Guide](../migrations/ROLLBACK_PROCEDURE.md)


---

## KDS (Kitchen Display System) Endpoints

### 4. Get Today's Cooking Menu

Retrieves today's cooking menu with portion size information for kitchen staff.

**Endpoint:** `GET /api/v1/kds/cooking/today`

**Query Parameters:**
- `date` (optional): Date in YYYY-MM-DD format. Defaults to today if not provided.

#### Success Response (200 OK)

```json
{
  "success": true,
  "data": [
    {
      "recipe_id": 5,
      "recipe_name": "Nasi Goreng",
      "recipe_category": "main_course",
      "total_portions": 500,
      "status": "pending",
      "school_allocations": [
        {
          "school_id": 1,
          "school_name": "SD Negeri 1",
          "school_category": "SD",
          "portion_size_type": "mixed",
          "portions_small": 150,
          "portions_large": 200,
          "total_portions": 350
        },
        {
          "school_id": 2,
          "school_name": "SMP Negeri 1",
          "school_category": "SMP",
          "portion_size_type": "large",
          "portions_small": 0,
          "portions_large": 150,
          "total_portions": 150
        }
      ]
    }
  ]
}
```

#### Response Fields

**Recipe Status Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `recipe_id` | integer | ID of the recipe |
| `recipe_name` | string | Name of the recipe |
| `recipe_category` | string | Category of the recipe |
| `total_portions` | integer | Total portions to cook for this recipe |
| `status` | string | Cooking status: "pending", "cooking", or "ready" |
| `school_allocations` | array | Array of school allocations with portion sizes |

**School Allocation Fields:** (same as Get Menu Item response)

#### Error Responses

**400 Bad Request - Invalid Date Format**
```json
{
  "success": false,
  "error_code": "INVALID_DATE_FORMAT",
  "message": "Invalid date format. Expected YYYY-MM-DD",
  "details": "parsing time \"invalid\" as \"2006-01-02\": cannot parse \"invalid\" as \"2006\""
}
```

**500 Internal Server Error**
```json
{
  "success": false,
  "error_code": "INTERNAL_ERROR",
  "message": "Gagal mengambil menu hari ini",
  "details": "error details"
}
```

---

### 5. Get Today's Packing Allocations

Retrieves today's packing allocations with portion size breakdown for packing staff. Only shows allocations for recipes that have been cooked (status = "ready").

**Endpoint:** `GET /api/v1/kds/packing/today`

**Query Parameters:**
- `date` (optional): Date in YYYY-MM-DD format. Defaults to today if not provided.

#### Success Response (200 OK)

```json
{
  "success": true,
  "data": [
    {
      "school_id": 1,
      "school_name": "SD Negeri 1",
      "school_category": "SD",
      "portion_size_type": "mixed",
      "portions_small": 150,
      "portions_large": 200,
      "total_portions": 350,
      "status": "pending",
      "menu_items": [
        {
          "recipe_id": 5,
          "recipe_name": "Nasi Goreng",
          "portions_small": 150,
          "portions_large": 200,
          "total_portions": 350
        }
      ]
    },
    {
      "school_id": 2,
      "school_name": "SMP Negeri 1",
      "school_category": "SMP",
      "portion_size_type": "large",
      "portions_small": 0,
      "portions_large": 150,
      "total_portions": 150,
      "status": "pending",
      "menu_items": [
        {
          "recipe_id": 5,
          "recipe_name": "Nasi Goreng",
          "portions_small": 0,
          "portions_large": 150,
          "total_portions": 150
        }
      ]
    }
  ]
}
```

#### Response Fields

**School Allocation Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `school_id` | integer | ID of the school |
| `school_name` | string | Name of the school |
| `school_category` | string | School category: "SD", "SMP", or "SMA" |
| `portion_size_type` | string | "mixed" for SD schools, "large" for SMP/SMA |
| `portions_small` | integer | Total small portions for this school (all recipes) |
| `portions_large` | integer | Total large portions for this school (all recipes) |
| `total_portions` | integer | Total portions for this school (all recipes) |
| `status` | string | Packing status: "pending", "packing", or "packed" |
| `menu_items` | array | Array of menu items with portion breakdown |

**Menu Item Summary Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `recipe_id` | integer | ID of the recipe |
| `recipe_name` | string | Name of the recipe |
| `portions_small` | integer | Small portions for this recipe |
| `portions_large` | integer | Large portions for this recipe |
| `total_portions` | integer | Total portions for this recipe |

**Note**: 
- Allocations are grouped by school and sorted alphabetically by school name
- Only recipes with status "ready" are included in packing view
- School-level portions are aggregated across all menu items

#### Error Responses

**400 Bad Request - Invalid Date Format**
```json
{
  "success": false,
  "error_code": "INVALID_DATE_FORMAT",
  "message": "Invalid date format. Expected YYYY-MM-DD",
  "details": "parsing time \"invalid\" as \"2006-01-02\": cannot parse \"invalid\" as \"2006\""
}
```

**500 Internal Server Error**
```json
{
  "success": false,
  "error_code": "INTERNAL_ERROR",
  "message": "Gagal mengambil alokasi packing hari ini",
  "details": "error details"
}
```

---

### 6. Update Cooking Status

Updates the cooking status of a recipe.

**Endpoint:** `PUT /api/v1/kds/cooking/:recipe_id/status`

**URL Parameters:**
- `recipe_id` (required): Recipe ID (integer)

#### Request Body

```json
{
  "status": "cooking"
}
```

#### Request Fields

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `status` | string | Yes | Must be one of: "pending", "cooking", "ready" | New cooking status |

#### Success Response (200 OK)

```json
{
  "success": true,
  "message": "Status resep berhasil diperbarui"
}
```

#### Error Responses

**400 Bad Request - Invalid Recipe ID**
```json
{
  "success": false,
  "error_code": "INVALID_RECIPE_ID",
  "message": "ID resep tidak valid"
}
```

**400 Bad Request - Validation Error**
```json
{
  "success": false,
  "error_code": "VALIDATION_ERROR",
  "message": "Data tidak valid",
  "details": "Key: 'status' Error:Field validation for 'status' failed on the 'oneof' tag"
}
```

**401 Unauthorized**
```json
{
  "success": false,
  "error_code": "UNAUTHORIZED",
  "message": "Pengguna tidak terautentikasi"
}
```

**500 Internal Server Error**
```json
{
  "success": false,
  "error_code": "UPDATE_FAILED",
  "message": "Gagal memperbarui status resep",
  "details": "error details"
}
```

---

### 7. Update Packing Status

Updates the packing status for a school.

**Endpoint:** `PUT /api/v1/kds/packing/:school_id/status`

**URL Parameters:**
- `school_id` (required): School ID (integer)

#### Request Body

```json
{
  "status": "packing"
}
```

#### Request Fields

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `status` | string | Yes | Must be one of: "pending", "packing", "packed" | New packing status |

#### Success Response (200 OK)

```json
{
  "success": true,
  "message": "Status packing berhasil diperbarui"
}
```

#### Error Responses

Same as Update Cooking Status endpoint.

---

## KDS Workflow

### Cooking View Workflow

1. Kitchen staff opens the cooking view
2. System calls `GET /api/v1/kds/cooking/today` to get today's menu
3. Response shows all recipes with portion size breakdown by school
4. Kitchen staff sees:
   - For SD schools: "SD Negeri 1: Small (150), Large (200)"
   - For SMP/SMA schools: "SMP Negeri 1: Large (150)"
5. As cooking progresses, staff updates status using `PUT /api/v1/kds/cooking/:recipe_id/status`
6. When status changes to "ready", the recipe becomes visible in packing view

### Packing View Workflow

1. Packing staff opens the packing view
2. System calls `GET /api/v1/kds/packing/today` to get ready recipes
3. Response shows schools grouped with portion size breakdown
4. Packing staff sees:
   - School name and category
   - Total small portions (for SD schools)
   - Total large portions (all schools)
   - Breakdown by recipe
5. Staff packs meals according to portion sizes
6. Updates status using `PUT /api/v1/kds/packing/:school_id/status`

### Real-time Updates

Both cooking and packing views use Firebase for real-time synchronization:
- Status changes are immediately reflected across all connected clients
- No manual refresh needed
- Ensures all staff see the same current state

---

## Integration Notes

### Frontend Integration

When displaying portion sizes in the UI:

1. **Check `portion_size_type`**:
   - If "mixed": Display both small and large portions
   - If "large": Display only large portions

2. **Label Format**:
   - Small portions: "Small (Grades 1-3): X"
   - Large portions (SD): "Large (Grades 4-6): Y"
   - Large portions (SMP/SMA): "Large: Y"

3. **Visual Indicators**:
   - Use different colors or icons for small vs large portions
   - Make it clear which portions are for which grade levels

### Backend Integration

When implementing new features that interact with allocations:

1. **Always group by school**: Multiple allocation records may exist for the same school
2. **Respect portion_size field**: Filter or aggregate based on portion_size when needed
3. **Maintain transaction safety**: Use database transactions for multi-record operations
4. **Validate school category**: Check school category before allowing small portions

---

## Performance Considerations

### Database Queries

The system uses optimized queries with:
- Composite index on `(menu_item_id, school_id, portion_size)`
- Preloading of related entities (School, Recipe)
- Efficient grouping and aggregation

### Caching

Consider implementing caching for:
- School category lookups (rarely change)
- Today's menu (cache until midnight)
- Recipe information (cache with TTL)

### Firebase Sync

- Real-time updates use Firebase Realtime Database
- Status changes trigger automatic sync
- Manual sync endpoints available if needed

---

## Testing Checklist

When testing the portion size differentiation feature:

- [ ] Create menu item with SD school (both portion sizes)
- [ ] Create menu item with SMP school (large only)
- [ ] Create menu item with SMA school (large only)
- [ ] Test sum validation (must equal total)
- [ ] Test SMP/SMA cannot have small portions
- [ ] Test at least one portion required
- [ ] Test negative portions rejected
- [ ] Test duplicate school rejected
- [ ] Update existing menu item allocations
- [ ] View menu item with portion breakdown
- [ ] View cooking menu with portion sizes
- [ ] View packing allocations with portion sizes
- [ ] Update cooking status
- [ ] Update packing status
- [ ] Verify real-time updates in Firebase
- [ ] Test with multiple schools
- [ ] Test alphabetical sorting
- [ ] Delete menu item (cascade delete allocations)
