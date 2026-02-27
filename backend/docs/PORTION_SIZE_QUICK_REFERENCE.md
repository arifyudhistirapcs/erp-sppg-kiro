# Portion Size Differentiation - Quick Reference

## Overview

Quick reference guide for the Portion Size Differentiation feature.

## School Types & Portion Rules

| School Type | Category | Small Portions | Large Portions | Records Created |
|-------------|----------|----------------|----------------|-----------------|
| Elementary (SD) | SD | ✅ Grades 1-3 | ✅ Grades 4-6 | 1 or 2 (based on input) |
| Junior High (SMP) | SMP | ❌ Not allowed | ✅ All students | 1 |
| Senior High (SMA) | SMA | ❌ Not allowed | ✅ All students | 1 |

## Validation Rules

✅ **Must Pass:**
1. Sum of all portions = total portions
2. SMP/SMA schools: `portions_small` must be 0
3. Each school: at least one portion type > 0
4. All portions: non-negative integers
5. No duplicate schools in request

❌ **Will Fail:**
- Sum mismatch: `portions_small + portions_large ≠ total`
- SMP/SMA with small portions: `portions_small > 0`
- Zero portions: `portions_small = 0 AND portions_large = 0`
- Negative values: `portions_small < 0 OR portions_large < 0`
- Duplicate school IDs in same request

## Request Format

### Create/Update Menu Item

```json
{
  "date": "2024-01-15",
  "recipe_id": 5,
  "portions": 500,
  "school_allocations": [
    {
      "school_id": 1,           // SD school
      "portions_small": 150,    // Grades 1-3
      "portions_large": 200     // Grades 4-6
    },
    {
      "school_id": 2,           // SMP school
      "portions_small": 0,      // Must be 0
      "portions_large": 150     // All students
    }
  ]
}
```

## Response Format

### Get Menu Item

```json
{
  "school_allocations": [
    {
      "school_id": 1,
      "school_name": "SD Negeri 1",
      "school_category": "SD",
      "portion_size_type": "mixed",
      "portions_small": 150,
      "portions_large": 200,
      "total_portions": 350
    }
  ]
}
```

## Common Error Messages

| Error | Cause | Fix |
|-------|-------|-----|
| `sum of allocated portions (X) does not equal total portions (Y)` | Math doesn't add up | Adjust allocations to match total |
| `SMP schools cannot have small portions` | SMP has portions_small > 0 | Set portions_small to 0 |
| `SMA schools cannot have small portions` | SMA has portions_small > 0 | Set portions_small to 0 |
| `school must have at least one portion: school_id X` | Both portion types are 0 | Set at least one > 0 |
| `duplicate allocation for school_id X` | Same school twice | Remove duplicate |

## API Endpoints

### Menu Planning

```
POST   /api/menu-plans/:id/items           Create menu item
GET    /api/menu-plans/:id/items/:item_id  Get menu item
PUT    /api/menu-plans/:id/items/:item_id  Update menu item
DELETE /api/menu-plans/:id/items/:item_id  Delete menu item
```

### KDS (Kitchen Display System)

```
GET /api/v1/kds/cooking/today?date=YYYY-MM-DD  Get cooking menu
GET /api/v1/kds/packing/today?date=YYYY-MM-DD  Get packing allocations
PUT /api/v1/kds/cooking/:recipe_id/status      Update cooking status
PUT /api/v1/kds/packing/:school_id/status      Update packing status
```

## Database Schema

### menu_item_school_allocations

```sql
CREATE TABLE menu_item_school_allocations (
  id              SERIAL PRIMARY KEY,
  menu_item_id    INTEGER NOT NULL,
  school_id       INTEGER NOT NULL,
  portions        INTEGER NOT NULL,
  portion_size    VARCHAR(10) NOT NULL CHECK (portion_size IN ('small', 'large')),
  date            DATE NOT NULL,
  FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE CASCADE,
  FOREIGN KEY (school_id) REFERENCES schools(id)
);

CREATE INDEX idx_menu_item_school_portion 
  ON menu_item_school_allocations(menu_item_id, school_id, portion_size);
```

## Example Calculations

### Example 1: SD School Only
```
Total: 350 portions
SD School 1:
  - Small (Grades 1-3): 150
  - Large (Grades 4-6): 200
  - Sum: 150 + 200 = 350 ✅
```

### Example 2: Multiple Schools
```
Total: 800 portions
SD School 1:  150 small + 200 large = 350
SMP School 2: 0 small + 200 large = 200
SD School 3:  100 small + 150 large = 250
Sum: 350 + 200 + 250 = 800 ✅
```

### Example 3: Invalid - Sum Mismatch
```
Total: 500 portions
SD School 1: 150 small + 200 large = 350
Sum: 350 ≠ 500 ❌
Error: "sum of allocated portions (350) does not equal total portions (500)"
```

### Example 4: Invalid - SMP with Small Portions
```
Total: 200 portions
SMP School 2: 50 small + 150 large = 200
Error: "SMP schools cannot have small portions" ❌
```

## UI Display Guidelines

### SD Schools (Mixed Portions)
```
SD Negeri 1
├─ Small (Grades 1-3): 150 portions
└─ Large (Grades 4-6): 200 portions
   Total: 350 portions
```

### SMP/SMA Schools (Large Only)
```
SMP Negeri 1
└─ Large: 150 portions
   Total: 150 portions
```

## Testing Checklist

- [ ] Create SD school with both portion sizes
- [ ] Create SMP school with large only
- [ ] Create SMA school with large only
- [ ] Test sum validation (must equal total)
- [ ] Test SMP cannot have small portions
- [ ] Test SMA cannot have small portions
- [ ] Test at least one portion required
- [ ] Test negative portions rejected
- [ ] Test duplicate school rejected
- [ ] Update existing allocations
- [ ] View in cooking menu
- [ ] View in packing allocations
- [ ] Delete menu item (cascade)

## Postman Collection

Import `Portion_Size_Differentiation_API.postman_collection.json` for 17 pre-configured test requests.

## Related Documentation

- [Full API Documentation](./API_PORTION_SIZE_DIFFERENTIATION.md)
- [Requirements](../../.kiro/specs/portion-size-differentiation/requirements.md)
- [Design Document](../../.kiro/specs/portion-size-differentiation/design.md)
- [Database Migrations](../migrations/)
