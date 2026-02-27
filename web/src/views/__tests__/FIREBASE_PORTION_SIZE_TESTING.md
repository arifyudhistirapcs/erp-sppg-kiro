# Firebase Real-time Updates - Portion Size Testing Guide

## Overview
This document describes how to manually test that Firebase real-time updates correctly handle portion size data changes in the KDS views.

## Test Scenarios

### Test 1: SD School Portion Size Update in KDS Cooking View

**Setup:**
1. Create a menu item with an SD school allocation:
   - portions_small: 50
   - portions_large: 75
   - total_portions: 125

**Test Steps:**
1. Open KDS Cooking View
2. Verify the initial portion sizes are displayed correctly
3. Update the Firebase data at `/kds/cooking/{date}/{recipe_id}` to change:
   - portions_small: 60
   - portions_large: 85
   - total_portions: 145
4. Observe the KDS Cooking View

**Expected Result:**
- The view should automatically update to show the new portion sizes
- Small portions should show 60
- Large portions should show 85
- Total should show 145

### Test 2: SMP School Portion Size Update in KDS Cooking View

**Setup:**
1. Create a menu item with an SMP school allocation:
   - portions_small: 0
   - portions_large: 100
   - total_portions: 100

**Test Steps:**
1. Open KDS Cooking View
2. Verify the initial portion sizes are displayed correctly
3. Update the Firebase data to change:
   - portions_large: 120
   - total_portions: 120
4. Observe the KDS Cooking View

**Expected Result:**
- The view should automatically update to show the new portion sizes
- Large portions should show 120
- Total should show 120
- Small portions should remain 0

### Test 3: Multiple Schools Portion Size Update in KDS Cooking View

**Setup:**
1. Create a menu item with multiple school allocations:
   - SD Negeri 1: portions_small=50, portions_large=75
   - SMP Negeri 1: portions_small=0, portions_large=100

**Test Steps:**
1. Open KDS Cooking View
2. Update Firebase data to change both schools' portions
3. Observe the view

**Expected Result:**
- Both schools should update simultaneously
- All portion sizes should reflect the new values

### Test 4: SD School Portion Size Update in KDS Packing View

**Setup:**
1. Create packing allocations for an SD school:
   - portions_small: 50
   - portions_large: 75
   - total_portions: 125

**Test Steps:**
1. Open KDS Packing View
2. Verify the initial portion sizes in the portion breakdown cards
3. Update Firebase data at `/kds/packing/{date}/{school_id}` to change:
   - portions_small: 60
   - portions_large: 85
   - total_portions: 145
4. Observe the KDS Packing View

**Expected Result:**
- The portion breakdown cards should update automatically
- Small portion card should show 60
- Large portion card should show 85
- Total should show 145

### Test 5: SMP School Portion Size Update in KDS Packing View

**Setup:**
1. Create packing allocations for an SMP school:
   - portions_small: 0
   - portions_large: 100
   - total_portions: 100

**Test Steps:**
1. Open KDS Packing View
2. Update Firebase data to change:
   - portions_large: 120
   - total_portions: 120
3. Observe the view

**Expected Result:**
- The single large portion card should update to show 120
- No small portion card should be displayed

### Test 6: Preserve Portion Sizes When Status Changes

**Setup:**
1. Create a menu item with portion size data

**Test Steps:**
1. Open KDS Cooking View
2. Update only the status field in Firebase (not school_allocations)
3. Observe the view

**Expected Result:**
- Status should update
- Portion sizes should remain unchanged
- No data loss should occur

### Test 7: Zero Values in Portion Sizes

**Setup:**
1. Create an SD school allocation with both portion sizes > 0

**Test Steps:**
1. Open KDS Cooking View
2. Update Firebase to set portions_small to 0
3. Observe the view

**Expected Result:**
- Small portion display should show 0 or be hidden
- Large portion should show the updated value
- Total should be correct

## Firebase Listener Implementation

The Firebase listeners in both KDS views use the following logic to update portion size data:

```javascript
// KDS Cooking View
recipes.value = recipes.value.map(recipe => {
  const firebaseRecipe = firebaseRecipes.find(fr => fr.recipe_id === recipe.recipe_id)
  if (firebaseRecipe) {
    return {
      ...recipe,
      status: firebaseRecipe.status,
      start_time: firebaseRecipe.start_time,
      school_allocations: firebaseRecipe.school_allocations || recipe.school_allocations
    }
  }
  return recipe
})

// KDS Packing View
schools.value = schools.value.map(school => {
  const firebaseSchool = firebaseSchools.find(fs => fs.school_id === school.school_id)
  if (firebaseSchool) {
    return {
      ...school,
      status: firebaseSchool.status,
      portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
      portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
      portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
      total_portions: firebaseSchool.total_portions || school.total_portions
    }
  }
  return school
})
```

## Key Points to Verify

1. **Portion Size Fields**: Verify that `portions_small`, `portions_large`, and `total_portions` are correctly updated
2. **Portion Size Type**: Verify that `portion_size_type` field is maintained ('mixed' for SD, 'large' for SMP/SMA)
3. **School Category**: Verify that school category determines the display format
4. **Real-time Updates**: Verify that updates happen without page refresh
5. **Data Preservation**: Verify that unrelated fields are not affected by updates
6. **Zero Handling**: Verify that zero values are handled correctly (not treated as undefined)

## Testing Tools

- Firebase Console: Use to manually update data
- Browser DevTools: Monitor console logs for Firebase listener activity
- Vue DevTools: Inspect component state changes

## Success Criteria

All test scenarios should pass with:
- Immediate visual updates (no page refresh required)
- Correct portion size values displayed
- No data loss or corruption
- Proper handling of edge cases (zero values, missing fields)
