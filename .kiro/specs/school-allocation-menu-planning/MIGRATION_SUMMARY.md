# Database Migration Summary - Task 1.1

## Overview
Created database migration for `menu_item_school_allocations` table to support school-level portion allocation in menu planning.

## Changes Made

### 1. Model Definition (`backend/internal/models/recipe.go`)

#### Created MenuItemSchoolAllocation Model
```go
type MenuItemSchoolAllocation struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    MenuItemID uint      `gorm:"index;not null" json:"menu_item_id"`
    SchoolID   uint      `gorm:"index;not null" json:"school_id"`
    Portions   int       `gorm:"not null;check:portions > 0" json:"portions" validate:"required,gt=0"`
    Date       time.Time `gorm:"index;not null" json:"date"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
    
    // Relationships
    MenuItem   MenuItem  `gorm:"foreignKey:MenuItemID;constraint:OnDelete:CASCADE" json:"menu_item,omitempty"`
    School     School    `gorm:"foreignKey:SchoolID;constraint:OnDelete:RESTRICT" json:"school,omitempty"`
}
```

**Key Features:**
- Primary key: `id` (auto-increment)
- Foreign keys:
  - `menu_item_id` → `menu_items(id)` with CASCADE delete
  - `school_id` → `schools(id)` with RESTRICT delete
- CHECK constraint: `portions > 0`
- Indexes on: `menu_item_id`, `school_id`, `date`
- Timestamps: `created_at`, `updated_at`

#### Updated MenuItem Model
Added relationship to school allocations:
```go
type MenuItem struct {
    // ... existing fields ...
    SchoolAllocations []MenuItemSchoolAllocation `gorm:"foreignKey:MenuItemID" json:"school_allocations,omitempty"`
}
```

### 2. Model Registration (`backend/internal/models/models.go`)

Added `&MenuItemSchoolAllocation{}` to the `AllModels()` function to ensure it's included in automatic migrations.

### 3. Database Indexes (`backend/internal/database/migrate.go`)

Added the following indexes for query optimization:

1. **UNIQUE Index** - Prevents duplicate allocations:
   ```sql
   CREATE UNIQUE INDEX idx_menu_item_school_allocation_unique 
   ON menu_item_school_allocations(menu_item_id, school_id)
   ```

2. **Menu Item Index** - Optimizes queries by menu item:
   ```sql
   CREATE INDEX idx_menu_item_school_allocation_menu_item 
   ON menu_item_school_allocations(menu_item_id)
   ```

3. **School Index** - Optimizes queries by school:
   ```sql
   CREATE INDEX idx_menu_item_school_allocation_school 
   ON menu_item_school_allocations(school_id)
   ```

4. **Date Index** - Optimizes date-based queries:
   ```sql
   CREATE INDEX idx_menu_item_school_allocation_date 
   ON menu_item_school_allocations(date)
   ```

## Requirements Satisfied

✅ **Requirement 3.1** - Store school allocations in database table linking menu items, schools, and portion counts
✅ **Requirement 3.2** - Store menu item identifier, school identifier, portion count, and date for each allocation
✅ **Requirement 3.4** - Maintain referential integrity with menu items and schools tables
✅ **Requirement 8.1** - Prevent duplicate allocations through UNIQUE constraint

## Database Schema

The migration will create the following table structure:

```sql
CREATE TABLE menu_item_school_allocations (
    id SERIAL PRIMARY KEY,
    menu_item_id INTEGER NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    school_id INTEGER NOT NULL REFERENCES schools(id) ON DELETE RESTRICT,
    portions INTEGER NOT NULL CHECK (portions > 0),
    date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(menu_item_id, school_id)
);

CREATE INDEX idx_menu_item_school_allocation_menu_item ON menu_item_school_allocations(menu_item_id);
CREATE INDEX idx_menu_item_school_allocation_school ON menu_item_school_allocations(school_id);
CREATE INDEX idx_menu_item_school_allocation_date ON menu_item_school_allocations(date);
```

## Migration Execution

The migration will run automatically when the server starts via GORM's AutoMigrate feature in `database.Migrate()`.

To apply the migration:
```bash
cd backend
go run cmd/server/main.go
```

The migration system will:
1. Create the `menu_item_school_allocations` table
2. Add foreign key constraints
3. Add CHECK constraint for positive portions
4. Create all indexes including the UNIQUE constraint

## Verification

All code changes have been verified:
- ✅ Go code compiles successfully
- ✅ No diagnostic errors in models or migration files
- ✅ Model properly registered in AllModels()
- ✅ Relationships correctly defined with CASCADE and RESTRICT constraints
- ✅ All required indexes created

## Next Steps

The migration is ready for execution. Subsequent tasks will:
- Task 1.2: Implement service layer validation logic
- Task 1.3: Create API endpoints for school allocations
- Task 1.4: Add frontend UI components
