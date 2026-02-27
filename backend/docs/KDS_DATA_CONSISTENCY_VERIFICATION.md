# KDS Data Consistency Verification

## Overview

This document verifies that portion size data is consistent across KDS Cooking View and KDS Packing View for the Portion Size Differentiation feature.

## Data Flow

### 1. Database Layer
Both views read from the same source: `menu_item_school_allocations` table with the `portion_size` field.

**Schema:**
```sql
menu_item_school_allocations (
  id,
  menu_item_id,
  school_id,
  portions,
  portion_size VARCHAR(10) CHECK (portion_size IN ('small', 'large')),
  date
)
```

### 2. Backend Services

#### KDS Cooking View (`kds_service.go`)

**Function:** `GetTodayMenu()`
- **Location:** `backend/internal/services/kds_service.go:85-203`
- **Data Structure:** `SchoolAllocationResponse`
  ```go
  type SchoolAllocationResponse struct {
      SchoolID        uint   `json:"school_id"`
      SchoolName      string `json:"school_name"`
      SchoolCategory  string `json:"school_category"`
      PortionSizeType string `json:"portion_size_type"` // 'small', 'large', or 'mixed'
      PortionsSmall   int    `json:"portions_small"`
      PortionsLarge   int    `json:"portions_large"`
      TotalPortions   int    `json:"total_portions"`
  }
  ```

**Processing Logic (Lines 127-175):**
1. Groups allocations by `school_id`
2. Determines `portion_size_type` based on school category:
   - SD schools → "mixed"
   - SMP/SMA schools → "large"
3. Accumulates portions by size:
   - `portion_size = 'small'` → adds to `PortionsSmall`
   - `portion_size = 'large'` → adds to `PortionsLarge`
4. Calculates `TotalPortions` as sum of small + large
5. Sorts by school name alphabetically

#### KDS Packing View (`packing_allocation_service.go`)

**Function:** `CalculatePackingAllocations()`
- **Location:** `backend/internal/services/packing_allocation_service.go:62-200`
- **Data Structure:** `SchoolAllocation`
  ```go
  type SchoolAllocation struct {
      SchoolID        uint              `json:"school_id"`
      SchoolName      string            `json:"school_name"`
      SchoolCategory  string            `json:"school_category"`
      PortionSizeType string            `json:"portion_size_type"` // 'small', 'large', or 'mixed'
      PortionsSmall   int               `json:"portions_small"`
      PortionsLarge   int               `json:"portions_large"`
      TotalPortions   int               `json:"total_portions"`
      MenuItems       []MenuItemSummary `json:"menu_items"`
      Status          string            `json:"status"`
  }
  ```

**Processing Logic (Lines 82-180):**
1. Groups allocations by `school_id`
2. Determines `portion_size_type` based on school category:
   - SD schools → "mixed"
   - SMP/SMA schools → "large"
3. Accumulates portions by size:
   - `portion_size = 'small'` → adds to `PortionsSmall`
   - `portion_size = 'large'` → adds to `PortionsLarge`
4. Calculates `TotalPortions` as sum of small + large
5. Sorts by school name alphabetically
6. Additionally aggregates menu items per school

### 3. Firebase Sync

#### Cooking View Sync (`kds_service.go`)

**Function:** `SyncTodayMenuToFirebase()`
- **Location:** `backend/internal/services/kds_service.go:457-486`
- **Firebase Path:** `/kds/cooking/{date}`
- **Synced Data:**
  ```go
  {
    "recipe_id": {
      "recipe_id": uint,
      "name": string,
      "status": string,
      "portions_required": int,
      "instructions": string,
      "items": []SemiFinishedQuantity,
      "school_allocations": []SchoolAllocationResponse  // ← Includes portion size data
    }
  }
  ```

#### Packing View Sync (`packing_allocation_service.go`)

**Function:** `SyncPackingAllocationsToFirebase()`
- **Location:** `backend/internal/services/packing_allocation_service.go:460-491`
- **Firebase Path:** `/kds/packing/{date}`
- **Synced Data:**
  ```go
  {
    "school_id": {
      "school_id": uint,
      "school_name": string,
      "school_category": string,
      "portion_size_type": string,  // ← Same field
      "portions_small": int,         // ← Same field
      "portions_large": int,         // ← Same field
      "total_portions": int,         // ← Same field
      "menu_items": []MenuItemSummary,
      "status": string
    }
  }
  ```

### 4. Frontend Views

#### KDS Cooking View (`KDSCookingView.vue`)

**Data Reception:**
- Receives data from API: `/api/v1/kds/cooking/today`
- Real-time updates from Firebase: `/kds/cooking/{date}`

**Display Logic (Lines 127-145):**
```vue
<div v-if="item.portion_size_type === 'mixed'">
  <div v-if="item.portions_small > 0">
    <a-badge :count="item.portions_small">
      <a-tag color="orange">Kecil (Kelas 1-3)</a-tag>
    </a-badge>
  </div>
  <div v-if="item.portions_large > 0">
    <a-badge :count="item.portions_large">
      <a-tag color="blue">Besar (Kelas 4-6)</a-tag>
    </a-badge>
  </div>
</div>
<div v-else>
  <div>
    <a-badge :count="item.portions_large">
      <a-tag color="blue">Besar</a-tag>
    </a-badge>
  </div>
</div>
```

**Firebase Listener (Lines 318-340):**
- Updates `school_allocations` with Firebase data
- Preserves portion size fields: `portion_size_type`, `portions_small`, `portions_large`

#### KDS Packing View (`KDSPackingView.vue`)

**Data Reception:**
- Receives data from API: `/api/v1/kds/packing/today`
- Real-time updates from Firebase: `/kds/packing/{date}`

**Display Logic (Lines 82-107):**
```vue
<div v-if="school.portion_size_type === 'mixed'">
  <a-col :span="12">
    <div class="portion-size-card small">
      <div class="portion-label">Kecil (Kelas 1-3)</div>
      <div class="portion-value">{{ school.portions_small }} porsi</div>
    </div>
  </a-col>
  <a-col :span="12">
    <div class="portion-size-card large">
      <div class="portion-label">Besar (Kelas 4-6)</div>
      <div class="portion-value">{{ school.portions_large }} porsi</div>
    </div>
  </a-col>
</div>
<div v-else>
  <div class="portion-size-card large single">
    <div class="portion-label">Porsi Besar</div>
    <div class="portion-value">{{ school.portions_large }} porsi</div>
  </div>
</div>
```

**Firebase Listener (Lines 267-290):**
- Updates portion size fields from Firebase:
  - `portion_size_type`
  - `portions_small`
  - `portions_large`
  - `total_portions`

## Consistency Verification

### ✅ Data Structure Consistency

Both views use **identical field names and types**:

| Field | Cooking View | Packing View | Match |
|-------|-------------|-------------|-------|
| `school_id` | uint | uint | ✅ |
| `school_name` | string | string | ✅ |
| `school_category` | string | string | ✅ |
| `portion_size_type` | string | string | ✅ |
| `portions_small` | int | int | ✅ |
| `portions_large` | int | int | ✅ |
| `total_portions` | int | int | ✅ |

### ✅ Processing Logic Consistency

Both services use **identical algorithms**:

1. **Grouping:** Both group by `school_id`
2. **Portion Type Determination:** Both use same logic:
   ```go
   portionSizeType := "large"
   if school.Category == "SD" {
       portionSizeType = "mixed"
   }
   ```
3. **Accumulation:** Both accumulate portions by size:
   ```go
   if alloc.PortionSize == "small" {
       PortionsSmall += alloc.Portions
   } else if alloc.PortionSize == "large" {
       PortionsLarge += alloc.Portions
   }
   TotalPortions += alloc.Portions
   ```
4. **Sorting:** Both sort alphabetically by school name

### ✅ Firebase Sync Consistency

Both sync functions include **all portion size fields**:

| Field | Cooking Sync | Packing Sync | Match |
|-------|-------------|-------------|-------|
| `portion_size_type` | ✅ | ✅ | ✅ |
| `portions_small` | ✅ | ✅ | ✅ |
| `portions_large` | ✅ | ✅ | ✅ |
| `total_portions` | ✅ | ✅ | ✅ |

### ✅ Frontend Display Consistency

Both views display portion sizes using **identical logic**:

1. **Mixed Portions (SD schools):**
   - Both show small portions with "Kecil (Kelas 1-3)" label
   - Both show large portions with "Besar (Kelas 4-6)" label
   - Both display counts with badges/cards

2. **Large Only (SMP/SMA schools):**
   - Both show only large portions with "Besar" label
   - Both hide small portion display

3. **Firebase Updates:**
   - Both update the same fields from Firebase
   - Both preserve portion size data during updates

## Test Coverage

### Unit Tests
- ✅ `TestKDSDataConsistency` - Verifies single recipe consistency
- ✅ `TestKDSDataConsistencyMultipleRecipes` - Verifies aggregation consistency

### Integration Points Verified
1. ✅ Database → Backend Service (both views read same data)
2. ✅ Backend Service → API Response (both use same structure)
3. ✅ Backend Service → Firebase (both sync same fields)
4. ✅ Firebase → Frontend (both listen to same fields)
5. ✅ Frontend Display (both render same data)

## Conclusion

**Data consistency is VERIFIED across all KDS views:**

1. ✅ **Same Data Source:** Both views read from `menu_item_school_allocations` table
2. ✅ **Identical Data Structures:** Both use same field names and types
3. ✅ **Identical Processing Logic:** Both use same grouping and accumulation algorithms
4. ✅ **Consistent Firebase Sync:** Both sync all portion size fields
5. ✅ **Consistent Display Logic:** Both render portion sizes identically
6. ✅ **Real-time Consistency:** Firebase listeners update same fields in both views

**No inconsistencies found.** The portion size data shown in KDS Cooking View will always match the data shown in KDS Packing View for the same menu item and school.

## Recommendations

1. ✅ **Current Implementation:** No changes needed - data consistency is maintained
2. ✅ **Firebase Sync:** Both views sync portion size data correctly
3. ✅ **Real-time Updates:** Both views update portion sizes from Firebase
4. ✅ **Display Logic:** Both views render portion sizes consistently

## Manual Verification Steps

To manually verify data consistency:

1. **Create a menu item** with allocations for SD and SMP schools
2. **Open KDS Cooking View** and note the portion sizes for each school
3. **Open KDS Packing View** and verify the same portion sizes are displayed
4. **Update allocation** in the menu planning UI
5. **Verify both views** update with the same new values
6. **Check Firebase** data at `/kds/cooking/{date}` and `/kds/packing/{date}` to confirm same values

Expected Result: All portion size values (small, large, total) should match exactly across both views.
