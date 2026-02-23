# Recipe & Menu Planning Module - Implementation Guide

## Overview

This module implements the Recipe Management and Menu Planning features for the ERP SPPG web application. It allows Ahli Gizi (Nutritionists) to manage recipes with nutritional information and create weekly menu plans that meet nutritional standards.

## Components Implemented

### 1. Recipe Management

#### RecipeListView.vue
- **Location**: `web/src/views/RecipeListView.vue`
- **Features**:
  - Display recipes in a table with search and filter by category
  - Show nutritional information (calories, protein, carbs, fat) per recipe
  - View, edit, and delete recipes
  - View recipe version history
  - Real-time nutrition calculation when adding/editing ingredients

#### RecipeFormModal.vue
- **Location**: `web/src/components/RecipeFormModal.vue`
- **Features**:
  - Create and edit recipes with ingredient selection
  - Real-time nutrition calculation as ingredients are added
  - Validation against minimum nutritional standards (600 kcal, 15g protein per portion)
  - Visual feedback for nutrition validation status
  - Support for multiple ingredients with quantities

#### RecipeHistoryModal.vue
- **Location**: `web/src/components/RecipeHistoryModal.vue`
- **Features**:
  - Display recipe version history in timeline format
  - Show changes between versions
  - Highlight current active version

### 2. Menu Planning

#### MenuPlanningView.vue
- **Location**: `web/src/views/MenuPlanningView.vue`
- **Features**:
  - Weekly calendar view (Monday to Sunday)
  - Drag-and-drop recipes between days
  - Add recipes to specific days via dropdown selection
  - Real-time daily nutrition totals and validation
  - Visual indicators for days meeting/not meeting nutrition standards
  - Approve menu button (for Ahli Gizi role)
  - Duplicate previous week's menu functionality
  - Week navigation (previous/next/current week)

## Services

### recipeService.js
- **Location**: `web/src/services/recipeService.js`
- **API Endpoints**:
  - `GET /recipes` - Get all recipes with filters
  - `POST /recipes` - Create new recipe
  - `GET /recipes/:id` - Get single recipe
  - `PUT /recipes/:id` - Update recipe
  - `DELETE /recipes/:id` - Delete recipe
  - `GET /recipes/:id/nutrition` - Get recipe nutrition info
  - `GET /recipes/:id/history` - Get recipe version history
  - `GET /ingredients` - Get all ingredients

### menuPlanningService.js
- **Location**: `web/src/services/menuPlanningService.js`
- **API Endpoints**:
  - `GET /menu-plans` - Get all menu plans
  - `POST /menu-plans` - Create new menu plan
  - `GET /menu-plans/:id` - Get single menu plan
  - `PUT /menu-plans/:id` - Update menu plan
  - `POST /menu-plans/:id/approve` - Approve menu plan
  - `GET /menu-plans/current-week` - Get current week menu

## Routing

Routes added to `web/src/router/index.js`:

```javascript
{
  path: 'recipes',
  name: 'recipes',
  component: () => import('@/views/RecipeListView.vue'),
  meta: { 
    requiresAuth: true,
    roles: ['kepala_sppg', 'ahli_gizi'],
    title: 'Manajemen Resep'
  }
},
{
  path: 'menu-planning',
  name: 'menu-planning',
  component: () => import('@/views/MenuPlanningView.vue'),
  meta: { 
    requiresAuth: true,
    roles: ['kepala_sppg', 'ahli_gizi'],
    title: 'Perencanaan Menu'
  }
}
```

## Navigation

Menu items are already configured in `MainLayout.vue` under the "Resep & Menu" section:
- Manajemen Resep
- Perencanaan Menu

Access is restricted to users with roles: `kepala_sppg` and `ahli_gizi`.

## Key Features

### Real-time Nutrition Calculation
- Automatically calculates total nutrition values when ingredients are added or quantities changed
- Shows per-portion nutrition values
- Validates against minimum standards (600 kcal, 15g protein per portion)

### Menu Planning Workflow
1. Create a new weekly menu plan
2. Add recipes to each day (drag-drop or dropdown)
3. View daily nutrition totals
4. Ensure all days meet minimum standards
5. Approve menu (only Ahli Gizi or Kepala SPPG)
6. Approved menus become available to Kitchen Display System

### Duplicate Previous Week
- Quickly create a new week's menu by duplicating the previous week
- Automatically adjusts dates by +7 days
- Saves time in menu planning

## Nutritional Standards

Minimum standards per portion (configurable):
- **Calories**: 600 kcal
- **Protein**: 15g

These values are validated both at the recipe level and daily menu level.

## UI/UX Features

### Recipe Management
- Search by recipe name
- Filter by category (Makanan Pokok, Lauk Pauk, Sayuran, Buah, Minuman)
- Color-coded category tags
- Nutrition summary in table
- Active/inactive status indicators

### Menu Planning
- Visual weekly calendar layout
- Today's date highlighted with blue border
- Color-coded validation status (green = valid, orange = warning)
- Drag-and-drop support for easy menu rearrangement
- Real-time nutrition updates
- Responsive design for different screen sizes

## Dependencies

Added to `package.json`:
- `dayjs`: ^1.11.10 - For date manipulation and formatting

## Installation

```bash
cd web
npm install
```

## Running the Application

```bash
npm run dev
```

The application will be available at `http://localhost:5173` (or the configured port).

## Backend Requirements

The following backend API endpoints must be implemented and available:

### Recipe Endpoints
- GET /api/v1/recipes
- POST /api/v1/recipes
- GET /api/v1/recipes/:id
- PUT /api/v1/recipes/:id
- DELETE /api/v1/recipes/:id
- GET /api/v1/recipes/:id/nutrition
- GET /api/v1/recipes/:id/history
- GET /api/v1/ingredients

### Menu Planning Endpoints
- GET /api/v1/menu-plans
- POST /api/v1/menu-plans
- GET /api/v1/menu-plans/:id
- PUT /api/v1/menu-plans/:id
- POST /api/v1/menu-plans/:id/approve
- GET /api/v1/menu-plans/current-week

## Data Models

### Recipe
```javascript
{
  id: number,
  name: string,
  category: string,
  serving_size: number,
  instructions: string,
  total_calories: number,
  total_protein: number,
  total_carbs: number,
  total_fat: number,
  version: number,
  is_active: boolean,
  recipe_ingredients: [
    {
      ingredient_id: number,
      quantity: number,
      ingredient: {
        name: string,
        unit: string,
        calories_per_100g: number,
        protein_per_100g: number,
        carbs_per_100g: number,
        fat_per_100g: number
      }
    }
  ]
}
```

### Menu Plan
```javascript
{
  id: number,
  week_start: string, // YYYY-MM-DD
  week_end: string,   // YYYY-MM-DD
  status: string,     // 'draft' | 'approved'
  approved_by: number,
  approved_at: string,
  menu_items: [
    {
      id: number,
      date: string,     // YYYY-MM-DD
      recipe_id: number,
      portions: number,
      recipe: Recipe
    }
  ]
}
```

## Language

All UI text is in professional Bahasa Indonesia as per requirements:
- Form labels and buttons
- Error messages and validation
- Table headers and content
- Navigation menu items

## Future Enhancements

Potential improvements for future iterations:
1. Bulk recipe import from Excel
2. Recipe photo upload
3. Print menu plan as PDF
4. Nutrition analysis charts
5. Recipe recommendations based on available ingredients
6. Cost calculation per recipe
7. Allergen tracking
8. Recipe rating and feedback system

## Testing

To test the implementation:
1. Login as user with role `ahli_gizi` or `kepala_sppg`
2. Navigate to "Resep & Menu" > "Manajemen Resep"
3. Create a new recipe with ingredients
4. Verify nutrition calculation updates in real-time
5. Navigate to "Perencanaan Menu"
6. Create a new weekly menu
7. Add recipes to different days
8. Verify daily nutrition totals
9. Test drag-and-drop functionality
10. Approve the menu

## Troubleshooting

### Recipes not loading
- Check API endpoint is accessible
- Verify authentication token is valid
- Check browser console for errors

### Nutrition calculation not updating
- Ensure ingredient data includes nutrition values per 100g
- Check that quantities are numeric values
- Verify calculation logic in RecipeFormModal.vue

### Menu plan not saving
- Ensure menu plan is created first
- Check that recipes have valid IDs
- Verify backend accepts the payload format

## Support

For issues or questions, refer to:
- Design document: `.kiro/specs/erp-sppg-system/design.md`
- Requirements: `.kiro/specs/erp-sppg-system/requirements.md`
- Tasks: `.kiro/specs/erp-sppg-system/tasks.md`
