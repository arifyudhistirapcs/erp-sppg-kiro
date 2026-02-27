package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrMenuPlanNotFound       = errors.New("rencana menu tidak ditemukan")
	ErrMenuPlanAlreadyApproved = errors.New("rencana menu sudah disetujui")
	ErrMenuPlanValidation     = errors.New("validasi rencana menu gagal")
	ErrDailyNutritionInsufficient = errors.New("nutrisi harian tidak memenuhi standar")
)

// MenuPlanningService handles menu planning business logic
type MenuPlanningService struct {
	db            *gorm.DB
	recipeService *RecipeService
}

// NewMenuPlanningService creates a new menu planning service
func NewMenuPlanningService(db *gorm.DB) *MenuPlanningService {
	return &MenuPlanningService{
		db:            db,
		recipeService: NewRecipeService(db),
	}
}

// CreateWeeklyPlan creates a new weekly menu plan
func (s *MenuPlanningService) CreateWeeklyPlan(weekStart time.Time, menuItems []models.MenuItem, userID uint) (*models.MenuPlan, error) {
	// Calculate week end (6 days after start)
	weekEnd := weekStart.AddDate(0, 0, 6)

	// Create menu plan
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "draft",
		CreatedBy: userID,
	}

	// Validate daily nutrition for each day
	if err := s.validateWeeklyNutrition(menuItems); err != nil {
		return nil, err
	}

	// Create menu plan in transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create menu plan
		if err := tx.Create(menuPlan).Error; err != nil {
			return err
		}

		// Create menu items
		for i := range menuItems {
			menuItems[i].MenuPlanID = menuPlan.ID
		}
		if err := tx.Create(&menuItems).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load relationships
	s.db.Preload("MenuItems.Recipe").
		Preload("Creator").
		First(menuPlan, menuPlan.ID)

	return menuPlan, nil
}

// CreateEmptyMenuPlan creates an empty menu plan without menu items
func (s *MenuPlanningService) CreateEmptyMenuPlan(menuPlan *models.MenuPlan) error {
	if err := s.db.Create(menuPlan).Error; err != nil {
		return err
	}

	// Load relationships
	s.db.Preload("Creator").First(menuPlan, menuPlan.ID)

	return nil
}

// GetMenuPlanByID retrieves a menu plan by ID
func (s *MenuPlanningService) GetMenuPlanByID(id uint) (*models.MenuPlan, error) {
	var menuPlan models.MenuPlan
	err := s.db.Preload("MenuItems.Recipe").
		Preload("MenuItems.SchoolAllocations.School").
		Preload("Creator").
		Preload("Approver").
		First(&menuPlan, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuPlanNotFound
		}
		return nil, err
	}

	return &menuPlan, nil
}

// GetAllMenuPlans retrieves all menu plans
func (s *MenuPlanningService) GetAllMenuPlans() ([]models.MenuPlan, error) {
	var menuPlans []models.MenuPlan
	err := s.db.
		Preload("MenuItems.Recipe").
		Preload("MenuItems.SchoolAllocations").
		Preload("MenuItems.SchoolAllocations.School").
		Preload("Creator").
		Preload("Approver").
		Order("week_start DESC").
		Find(&menuPlans).Error
	return menuPlans, err
}

// GetCurrentWeekMenuPlan retrieves the menu plan for the current week
func (s *MenuPlanningService) GetCurrentWeekMenuPlan() (*models.MenuPlan, error) {
	now := time.Now()
	// Get start of current week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}
	weekStart := now.AddDate(0, 0, -(weekday - 1))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	var menuPlan models.MenuPlan
	err := s.db.Preload("MenuItems.Recipe").
		Preload("MenuItems.SchoolAllocations.School").
		Preload("Creator").
		Preload("Approver").
		Where("week_start = ? AND status = ?", weekStart, "approved").
		First(&menuPlan).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuPlanNotFound
		}
		return nil, err
	}

	return &menuPlan, nil
}

// ApproveMenu approves a menu plan
func (s *MenuPlanningService) ApproveMenu(id uint, approverID uint) error {
	// Get menu plan
	menuPlan, err := s.GetMenuPlanByID(id)
	if err != nil {
		return err
	}

	// Check if already approved
	if menuPlan.Status == "approved" {
		return ErrMenuPlanAlreadyApproved
	}

	// Note: We don't validate nutrition here because not all days need to be filled
	// The frontend will show warnings for empty days or insufficient nutrition

	// Update status
	now := time.Now()
	return s.db.Model(&models.MenuPlan{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "approved",
		"approved_by": approverID,
		"approved_at": now,
		"updated_at":  now,
	}).Error
}

// UpdateMenuPlan updates an existing menu plan (only if not approved)
func (s *MenuPlanningService) UpdateMenuPlan(id uint, menuItems []models.MenuItem) error {
	// Get existing menu plan
	menuPlan, err := s.GetMenuPlanByID(id)
	if err != nil {
		return err
	}

	// Check if already approved
	if menuPlan.Status == "approved" {
		return ErrMenuPlanAlreadyApproved
	}

	// Validate daily nutrition
	if err := s.validateWeeklyNutrition(menuItems); err != nil {
		return err
	}

	// Update menu plan in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete old menu items
		if err := tx.Where("menu_plan_id = ?", id).Delete(&models.MenuItem{}).Error; err != nil {
			return err
		}

		// Create new menu items
		for i := range menuItems {
			menuItems[i].MenuPlanID = id
		}
		if err := tx.Create(&menuItems).Error; err != nil {
			return err
		}

		// Update menu plan timestamp
		if err := tx.Model(&models.MenuPlan{}).Where("id = ?", id).Update("updated_at", time.Now()).Error; err != nil {
			return err
		}

		return nil
	})
}

// DuplicateMenuPlan duplicates a previous menu plan as a template
func (s *MenuPlanningService) DuplicateMenuPlan(sourceID uint, newWeekStart time.Time, userID uint) (*models.MenuPlan, error) {
	// Get source menu plan
	sourceMenuPlan, err := s.GetMenuPlanByID(sourceID)
	if err != nil {
		return nil, err
	}

	// Calculate week end
	weekEnd := newWeekStart.AddDate(0, 0, 6)

	// Create new menu plan
	newMenuPlan := &models.MenuPlan{
		WeekStart: newWeekStart,
		WeekEnd:   weekEnd,
		Status:    "draft",
		CreatedBy: userID,
	}

	// Duplicate menu items with adjusted dates
	var newMenuItems []models.MenuItem
	daysDiff := int(newWeekStart.Sub(sourceMenuPlan.WeekStart).Hours() / 24)

	for _, item := range sourceMenuPlan.MenuItems {
		newItem := models.MenuItem{
			Date:     item.Date.AddDate(0, 0, daysDiff),
			RecipeID: item.RecipeID,
			Portions: item.Portions,
		}
		newMenuItems = append(newMenuItems, newItem)
	}

	// Create in transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create menu plan
		if err := tx.Create(newMenuPlan).Error; err != nil {
			return err
		}

		// Create menu items
		for i := range newMenuItems {
			newMenuItems[i].MenuPlanID = newMenuPlan.ID
		}
		if err := tx.Create(&newMenuItems).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load relationships
	s.db.Preload("MenuItems.Recipe").
		Preload("Creator").
		First(newMenuPlan, newMenuPlan.ID)

	return newMenuPlan, nil
}

// DailyNutrition represents aggregated nutrition for a day
type DailyNutrition struct {
	Date          time.Time
	TotalCalories float64
	TotalProtein  float64
	TotalCarbs    float64
	TotalFat      float64
	TotalPortions int
}

// CalculateDailyNutrition calculates aggregated nutrition for each day in a menu plan
func (s *MenuPlanningService) CalculateDailyNutrition(menuPlanID uint) ([]DailyNutrition, error) {
	menuPlan, err := s.GetMenuPlanByID(menuPlanID)
	if err != nil {
		return nil, err
	}

	// Group menu items by date
	dailyMap := make(map[string]*DailyNutrition)

	for _, item := range menuPlan.MenuItems {
		dateKey := item.Date.Format("2006-01-02")

		if _, exists := dailyMap[dateKey]; !exists {
			dailyMap[dateKey] = &DailyNutrition{
				Date: item.Date,
			}
		}

		// Add recipe nutrition scaled by portions
		// Since serving_size is removed, nutrition is per menu, so just multiply by portions
		dailyMap[dateKey].TotalCalories += item.Recipe.TotalCalories * float64(item.Portions)
		dailyMap[dateKey].TotalProtein += item.Recipe.TotalProtein * float64(item.Portions)
		dailyMap[dateKey].TotalCarbs += item.Recipe.TotalCarbs * float64(item.Portions)
		dailyMap[dateKey].TotalFat += item.Recipe.TotalFat * float64(item.Portions)
		dailyMap[dateKey].TotalPortions += item.Portions
	}

	// Convert map to slice
	var dailyNutrition []DailyNutrition
	for _, dn := range dailyMap {
		dailyNutrition = append(dailyNutrition, *dn)
	}

	return dailyNutrition, nil
}

// IngredientRequirement represents total ingredient requirements
type IngredientRequirement struct {
	IngredientID   uint
	IngredientName string
	Unit           string
	TotalQuantity  float64
}

// CalculateIngredientRequirements calculates total ingredient requirements for procurement
func (s *MenuPlanningService) CalculateIngredientRequirements(menuPlanID uint) ([]IngredientRequirement, error) {
	menuPlan, err := s.GetMenuPlanByID(menuPlanID)
	if err != nil {
		return nil, err
	}

	// Aggregate semi-finished goods requirements
	sfGoodsMap := make(map[uint]*IngredientRequirement)

	for _, item := range menuPlan.MenuItems {
		// No need to calculate scaling factor since serving_size is removed
		// Nutrition is per menu, so just multiply by portions
		for _, recipeItem := range item.Recipe.RecipeItems {
			sfGoodsID := recipeItem.SemiFinishedGoodsID

			if _, exists := sfGoodsMap[sfGoodsID]; !exists {
				sfGoodsMap[sfGoodsID] = &IngredientRequirement{
					IngredientID:   sfGoodsID,
					IngredientName: recipeItem.SemiFinishedGoods.Name,
					Unit:           recipeItem.SemiFinishedGoods.Unit,
					TotalQuantity:  0,
				}
			}

			// Multiply by portions directly (no scaling factor needed)
			sfGoodsMap[sfGoodsID].TotalQuantity += recipeItem.Quantity * float64(item.Portions)
		}
	}

	// Convert map to slice
	var requirements []IngredientRequirement
	for _, req := range sfGoodsMap {
		requirements = append(requirements, *req)
	}

	return requirements, nil
}

// validateWeeklyNutrition validates that each day meets minimum nutritional standards
func (s *MenuPlanningService) validateWeeklyNutrition(menuItems []models.MenuItem) error {
	// Group by date
	dailyMap := make(map[string]struct {
		totalCalories float64
		totalProtein  float64
		totalPortions int
	})

	for _, item := range menuItems {
		dateKey := item.Date.Format("2006-01-02")

		// Get recipe if not preloaded
		var recipe models.Recipe
		if item.Recipe.ID == 0 {
			if err := s.db.First(&recipe, item.RecipeID).Error; err != nil {
				return err
			}
		} else {
			recipe = item.Recipe
		}

		daily := dailyMap[dateKey]
		// Since serving_size is removed, nutrition is per menu
		daily.totalCalories += recipe.TotalCalories * float64(item.Portions)
		daily.totalProtein += recipe.TotalProtein * float64(item.Portions)
		daily.totalPortions += item.Portions
		dailyMap[dateKey] = daily
	}

	// Validate each day
	standards := DefaultNutritionStandards()
	for dateKey, daily := range dailyMap {
		if daily.totalPortions == 0 {
			continue
		}

		caloriesPerPortion := daily.totalCalories / float64(daily.totalPortions)
		proteinPerPortion := daily.totalProtein / float64(daily.totalPortions)

		if caloriesPerPortion < standards.MinCalories || proteinPerPortion < standards.MinProtein {
			return fmt.Errorf("nutrisi harian tidak memenuhi standar untuk tanggal %s: kalori=%.2f (min %.2f), protein=%.2f (min %.2f) per porsi", 
				dateKey, caloriesPerPortion, standards.MinCalories, proteinPerPortion, standards.MinProtein)
		}
	}

	return nil
}
// SchoolAllocationInput represents input for school allocation validation
type SchoolAllocationInput struct {
	SchoolID uint `json:"school_id" validate:"required"`
	Portions int  `json:"portions" validate:"required,gt=0"`
}

// PortionSizeAllocationInput represents input for school allocation with portion sizes
// Used for portion size differentiation feature (Requirements 3, 4)
type PortionSizeAllocationInput struct {
	SchoolID       uint `json:"school_id" validate:"required"`
	PortionsSmall  int  `json:"portions_small" validate:"gte=0"`
	PortionsLarge  int  `json:"portions_large" validate:"gte=0"`
}


// ValidateSchoolAllocations validates that school allocations meet business rules
// Returns an error if any validation rule is violated:
// - Allocations array must not be empty (Requirement 7)
// - No duplicate school IDs (Requirement 8)
// - All portion counts must be positive (Requirement 9)
// - Sum of allocated portions must equal total portions (Requirement 2)
func (s *MenuPlanningService) ValidateSchoolAllocations(
	totalPortions int,
	allocations []SchoolAllocationInput,
) error {
	// Check if allocations exist (Requirement 7.1)
	if len(allocations) == 0 {
		return errors.New("at least one school allocation is required")
	}

	// Check for duplicate schools and validate portion counts (Requirements 8, 9)
	schoolSet := make(map[uint]bool)
	sum := 0

	for _, alloc := range allocations {
		// Check for duplicates (Requirement 8.1)
		if schoolSet[alloc.SchoolID] {
			return fmt.Errorf("duplicate allocation for school_id %d", alloc.SchoolID)
		}
		schoolSet[alloc.SchoolID] = true

		// Validate portion count (Requirement 9.1)
		if alloc.Portions <= 0 {
			return fmt.Errorf("portions must be positive for school_id %d", alloc.SchoolID)
		}

		sum += alloc.Portions
	}

	// Validate sum equals total (Requirement 2.1, 2.2)
	if sum != totalPortions {
		return fmt.Errorf("sum of allocated portions (%d) does not equal total portions (%d)", sum, totalPortions)
	}

	return nil
}

// ValidatePortionSizeAllocations validates that portion size allocations meet business rules
// Returns (true, "") if all validations pass, or (false, error_message) if any validation fails
//
// Validation rules (from Requirements 3):
// - Allocations array must not be empty (Requirement 3.1)
// - Sum of all portions_small + portions_large must equal total_portions (Requirement 3.1, 3.2)
// - SMP/SMA schools cannot have portions_small > 0 (Requirement 3.3)
// - SD schools can have portions_small >= 0 and portions_large >= 0 (Requirement 3.4)
// - At least one of portions_small or portions_large must be > 0 for each school (Requirement 3.5)
// - Both portions_small and portions_large must be non-negative (Requirement 3.6)
// - No duplicate school IDs allowed
func (s *MenuPlanningService) ValidatePortionSizeAllocations(
	allocations []PortionSizeAllocationInput,
	totalPortions int,
) (bool, string) {
	// Check for empty allocations
	if len(allocations) == 0 {
		return false, "at least one school allocation is required"
	}

	// Validate total_portions is positive
	if totalPortions <= 0 {
		return false, "total portions must be positive"
	}

	// Track schools to detect duplicates and accumulate totals
	schoolSet := make(map[uint]bool)
	totalSmall := 0
	totalLarge := 0

	for _, alloc := range allocations {
		// Check for duplicate schools
		if schoolSet[alloc.SchoolID] {
			return false, fmt.Sprintf("duplicate allocation for school_id %d", alloc.SchoolID)
		}
		schoolSet[alloc.SchoolID] = true

		// Validate non-negative portions (Requirement 3.6)
		if alloc.PortionsSmall < 0 {
			return false, fmt.Sprintf("small portions cannot be negative for school_id %d", alloc.SchoolID)
		}
		if alloc.PortionsLarge < 0 {
			return false, fmt.Sprintf("large portions cannot be negative for school_id %d", alloc.SchoolID)
		}

		// At least one portion type must be positive (Requirement 3.5)
		if alloc.PortionsSmall == 0 && alloc.PortionsLarge == 0 {
			return false, fmt.Sprintf("school must have at least one portion: school_id %d", alloc.SchoolID)
		}

		// Fetch school to check category (Requirement 3.3, 12)
		var school models.School
		if err := s.db.First(&school, alloc.SchoolID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, fmt.Sprintf("school not found: school_id %d", alloc.SchoolID)
			}
			return false, fmt.Sprintf("error fetching school: %v", err)
		}

		// Validate SMP/SMA schools cannot have small portions (Requirement 3.3, 12)
		if school.Category == "SMP" && alloc.PortionsSmall > 0 {
			return false, "SMP schools cannot have small portions"
		}
		if school.Category == "SMA" && alloc.PortionsSmall > 0 {
			return false, "SMA schools cannot have small portions"
		}

		// Accumulate totals
		totalSmall += alloc.PortionsSmall
		totalLarge += alloc.PortionsLarge
	}

	// Validate sum equals total (Requirements 3.1, 3.2)
	totalAllocated := totalSmall + totalLarge
	if totalAllocated != totalPortions {
		return false, fmt.Sprintf("sum of allocated portions (%d) does not equal total portions (%d)", totalAllocated, totalPortions)
	}

	// All validations passed
	return true, ""
}


// MenuItemInput represents input for creating a menu item with allocations
type MenuItemInput struct {
	Date              time.Time                     `json:"date" validate:"required"`
	RecipeID          uint                          `json:"recipe_id" validate:"required"`
	Portions          int                           `json:"portions" validate:"required,gt=0"`
	SchoolAllocations []PortionSizeAllocationInput  `json:"school_allocations" validate:"required,dive"`
}

// CreateMenuItemWithAllocations creates a menu item and its school allocations
// This method:
// 1. Validates allocations using ValidatePortionSizeAllocations
// 2. Verifies all school IDs exist in the database
// 3. Creates the menu item and separate allocation records for small and large portions in a single transaction
// 4. Handles transaction rollback on any errors
// 5. Loads relationships (Recipe, SchoolAllocations.School) before returning
// Returns the created menu item with all relationships loaded
func (s *MenuPlanningService) CreateMenuItemWithAllocations(
	menuPlanID uint,
	input MenuItemInput,
) (*models.MenuItem, error) {
	// Validate allocations (Requirements 3, 4)
	isValid, errMsg := s.ValidatePortionSizeAllocations(input.SchoolAllocations, input.Portions)
	if !isValid {
		return nil, fmt.Errorf("%s", errMsg)
	}

	// Verify all schools exist and validate portion size compatibility (Requirement 1.4, 12)
	for _, alloc := range input.SchoolAllocations {
		var school models.School
		if err := s.db.First(&school, alloc.SchoolID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("school_id %d not found", alloc.SchoolID)
			}
			return nil, err
		}

		// Validate portion size compatibility with school category (Requirement 12)
		if (school.Category == "SMP" || school.Category == "SMA") && alloc.PortionsSmall > 0 {
			return nil, fmt.Errorf("%s schools cannot have small portions", school.Category)
		}
	}

	// Create menu item and allocations in transaction (Requirements 3.1, 3.2, 4)
	var menuItem models.MenuItem
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create menu item
		menuItem = models.MenuItem{
			MenuPlanID: menuPlanID,
			Date:       input.Date,
			RecipeID:   input.RecipeID,
			Portions:   input.Portions,
		}
		if err := tx.Create(&menuItem).Error; err != nil {
			return err
		}

		// Create separate allocation records for small and large portions (Requirement 4)
		for _, alloc := range input.SchoolAllocations {
			// Create small portion allocation if portions_small > 0 (Requirement 4.1)
			if alloc.PortionsSmall > 0 {
				smallAllocation := models.MenuItemSchoolAllocation{
					MenuItemID:  menuItem.ID,
					SchoolID:    alloc.SchoolID,
					Portions:    alloc.PortionsSmall,
					PortionSize: "small",
					Date:        input.Date,
				}
				if err := tx.Create(&smallAllocation).Error; err != nil {
					return err
				}
			}

			// Create large portion allocation if portions_large > 0 (Requirement 4.2)
			if alloc.PortionsLarge > 0 {
				largeAllocation := models.MenuItemSchoolAllocation{
					MenuItemID:  menuItem.ID,
					SchoolID:    alloc.SchoolID,
					Portions:    alloc.PortionsLarge,
					PortionSize: "large",
					Date:        input.Date,
				}
				if err := tx.Create(&largeAllocation).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load relationships (Requirement 2.4)
	err = s.db.Preload("Recipe").
		Preload("SchoolAllocations.School").
		First(&menuItem, menuItem.ID).Error

	if err != nil {
		return nil, err
	}

	return &menuItem, nil
}

// UpdateMenuItemWithAllocations updates a menu item and replaces its school allocations
// This method:
// 1. Validates the menu item exists
// 2. Validates new allocations using ValidateSchoolAllocations
// 3. Verifies all school IDs exist in the database
// 4. Deletes existing allocations for the menu item
// 5. Creates new allocations in a transaction
// 6. Handles transaction rollback on errors
// 7. Loads relationships (Recipe, SchoolAllocations.School) before returning
// Returns the updated menu item with all relationships loaded
func (s *MenuPlanningService) UpdateMenuItemWithAllocations(
	menuItemID uint,
	input MenuItemInput,
) (*models.MenuItem, error) {
	// Verify menu item exists
	var existingMenuItem models.MenuItem
	if err := s.db.First(&existingMenuItem, menuItemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("menu item with ID %d not found", menuItemID)
		}
		return nil, err
	}

	// Validate allocations (Requirements 5.2, 5.5)
	isValid, errMsg := s.ValidatePortionSizeAllocations(input.SchoolAllocations, input.Portions)
	if !isValid {
		return nil, fmt.Errorf("%s", errMsg)
	}

	// Verify all schools exist and validate portion size compatibility (Requirement 1.4, 12)
	for _, alloc := range input.SchoolAllocations {
		var school models.School
		if err := s.db.First(&school, alloc.SchoolID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("school_id %d not found", alloc.SchoolID)
			}
			return nil, err
		}

		// Validate portion size compatibility with school category (Requirement 12)
		if (school.Category == "SMP" || school.Category == "SMA") && alloc.PortionsSmall > 0 {
			return nil, fmt.Errorf("%s schools cannot have small portions", school.Category)
		}
	}

	// Update menu item and replace allocations in transaction (Requirements 5.1, 5.3, 5.4)
	var menuItem models.MenuItem
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Update menu item fields
		menuItem = existingMenuItem
		menuItem.Date = input.Date
		menuItem.RecipeID = input.RecipeID
		menuItem.Portions = input.Portions

		if err := tx.Save(&menuItem).Error; err != nil {
			return err
		}

		// Delete existing allocations (Requirement 5.1)
		if err := tx.Where("menu_item_id = ?", menuItemID).Delete(&models.MenuItemSchoolAllocation{}).Error; err != nil {
			return err
		}

		// Create new allocations with portion sizes (Requirements 5.3)
		for _, alloc := range input.SchoolAllocations {
			// Create small portion allocation if portions_small > 0
			if alloc.PortionsSmall > 0 {
				smallAllocation := models.MenuItemSchoolAllocation{
					MenuItemID:  menuItem.ID,
					SchoolID:    alloc.SchoolID,
					Portions:    alloc.PortionsSmall,
					PortionSize: "small",
					Date:        input.Date,
				}
				if err := tx.Create(&smallAllocation).Error; err != nil {
					return err
				}
			}

			// Create large portion allocation if portions_large > 0
			if alloc.PortionsLarge > 0 {
				largeAllocation := models.MenuItemSchoolAllocation{
					MenuItemID:  menuItem.ID,
					SchoolID:    alloc.SchoolID,
					Portions:    alloc.PortionsLarge,
					PortionSize: "large",
					Date:        input.Date,
				}
				if err := tx.Create(&largeAllocation).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load relationships (Requirement 5.3)
	err = s.db.Preload("Recipe").
		Preload("SchoolAllocations.School").
		First(&menuItem, menuItem.ID).Error

	if err != nil {
		return nil, err
	}

	return &menuItem, nil
}

// GetMenuItemWithAllocations retrieves a menu item with its school allocations
// This method:
// 1. Queries the menu item by ID
// 2. Preloads SchoolAllocations relationship
// 3. Preloads School relationship for each allocation
// 4. Orders allocations by school name alphabetically
// Returns the menu item with all relationships loaded, or an error if not found
func (s *MenuPlanningService) GetMenuItemWithAllocations(menuItemID uint) (*models.MenuItem, error) {
	var menuItem models.MenuItem

	// Query menu item with preloaded allocations and schools, ordered by school name
	// Requirements 4.2, 4.3, 4.4
	err := s.db.
		Preload("Recipe").
		Preload("SchoolAllocations", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN schools ON schools.id = menu_item_school_allocations.school_id").
				Order("schools.name ASC")
		}).
		Preload("SchoolAllocations.School").
		First(&menuItem, menuItemID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("menu item with ID %d not found", menuItemID)
		}
		return nil, err
	}

	return &menuItem, nil
}

// GetAllocationsByDate retrieves all school allocations for a specific date
// This method:
// 1. Queries all allocations for the specified date
// 2. Preloads MenuItem relationship for each allocation
// 3. Preloads School relationship for each allocation
// 4. Orders allocations by school name alphabetically
// Returns all allocations with relationships loaded, or an error if query fails
// Requirements: 4.1, 4.3, 4.4
func (s *MenuPlanningService) GetAllocationsByDate(date time.Time) ([]models.MenuItemSchoolAllocation, error) {
	var allocations []models.MenuItemSchoolAllocation

	// Query allocations for the date with preloaded relationships, ordered by school name
	err := s.db.
		Preload("MenuItem").
		Preload("MenuItem.Recipe").
		Preload("School").
		Joins("JOIN schools ON schools.id = menu_item_school_allocations.school_id").
		Where("menu_item_school_allocations.date = ?", date).
		Order("schools.name ASC").
		Find(&allocations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve allocations for date %s: %w", date.Format("2006-01-02"), err)
	}

	return allocations, nil
}



// DeleteMenuItem deletes a menu item and its school allocations
// This method:
// 1. Verifies the menu item exists and belongs to the specified menu plan
// 2. Checks if the menu plan is not approved (cannot delete from approved plans)
// 3. Deletes the menu item (allocations are cascade deleted by database constraint)
// Returns an error if the menu item doesn't exist or the menu plan is approved
func (s *MenuPlanningService) DeleteMenuItem(menuPlanID uint, menuItemID uint) error {
	// Get menu item to verify it exists and belongs to the menu plan
	var menuItem models.MenuItem
	if err := s.db.First(&menuItem, menuItemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("menu item with ID %d not found", menuItemID)
		}
		return err
	}

	// Verify the menu item belongs to the specified menu plan
	if menuItem.MenuPlanID != menuPlanID {
		return fmt.Errorf("menu item with ID %d not found in menu plan %d", menuItemID, menuPlanID)
	}

	// Get menu plan to check if it's approved
	var menuPlan models.MenuPlan
	if err := s.db.First(&menuPlan, menuPlanID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuPlanNotFound
		}
		return err
	}

	// Check if menu plan is approved
	if menuPlan.Status == "approved" {
		return ErrMenuPlanAlreadyApproved
	}

	// Delete in transaction to ensure atomicity
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete allocations first
		if err := tx.Where("menu_item_id = ?", menuItemID).Delete(&models.MenuItemSchoolAllocation{}).Error; err != nil {
			return err
		}

		// Delete menu item
		if err := tx.Delete(&menuItem).Error; err != nil {
			return err
		}

		return nil
	})
}

// SchoolAllocationDisplay represents a school allocation with portion size breakdown
// Used for displaying allocations grouped by school (Requirement 8)
type SchoolAllocationDisplay struct {
	SchoolID         uint   `json:"school_id"`
	SchoolName       string `json:"school_name"`
	SchoolCategory   string `json:"school_category"`
	PortionSizeType  string `json:"portion_size_type"`  // 'small', 'large', or 'mixed'
	PortionsSmall    int    `json:"portions_small"`
	PortionsLarge    int    `json:"portions_large"`
	TotalPortions    int    `json:"total_portions"`
}

// GetSchoolAllocationsWithPortionSizes retrieves allocations for a menu item grouped by school
// This method implements Requirement 8: Retrieve Allocations Grouped by School
//
// Algorithm (from Design Document):
// 1. Fetch all allocations for the menu item
// 2. Group allocations by school_id
// 3. For SD schools with multiple records (small + large), combine into single display record
// 4. For SMP/SMA schools, display single allocation with large portions only
// 5. Return allocations ordered alphabetically by school name
//
// Preconditions:
// - menu_item_id is a positive integer
// - Menu item exists in database
//
// Postconditions (Requirements 8.1-8.5):
// - Returns array of allocations grouped by school (8.1)
// - SD schools with multiple records are combined into single display record (8.2)
// - SMP/SMA schools display single allocation with portions_large only (8.3)
// - Allocations are ordered alphabetically by school name (8.4)
// - Each allocation includes school category in response (8.5)
func (s *MenuPlanningService) GetSchoolAllocationsWithPortionSizes(menuItemID uint) ([]SchoolAllocationDisplay, error) {
	// Step 1: Fetch all allocations for menu item
	var rawAllocations []models.MenuItemSchoolAllocation
	err := s.db.
		Preload("School").
		Where("menu_item_id = ?", menuItemID).
		Find(&rawAllocations).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve allocations for menu item %d: %w", menuItemID, err)
	}

	// Step 2: Group allocations by school
	schoolMap := make(map[uint]*SchoolAllocationDisplay)

	for _, allocation := range rawAllocations {
		schoolID := allocation.SchoolID

		// Initialize school entry if not exists
		if _, exists := schoolMap[schoolID]; !exists {
			// Determine portion size type based on school category (Requirement 1)
			portionSizeType := "large"
			if allocation.School.Category == "SD" {
				portionSizeType = "mixed"
			}

			schoolMap[schoolID] = &SchoolAllocationDisplay{
				SchoolID:        schoolID,
				SchoolName:      allocation.School.Name,
				SchoolCategory:  allocation.School.Category,
				PortionSizeType: portionSizeType,
				PortionsSmall:   0,
				PortionsLarge:   0,
				TotalPortions:   0,
			}
		}

		// Accumulate portions by size (Requirements 8.2, 8.3)
		if allocation.PortionSize == "small" {
			schoolMap[schoolID].PortionsSmall += allocation.Portions
		} else if allocation.PortionSize == "large" {
			schoolMap[schoolID].PortionsLarge += allocation.Portions
		}

		schoolMap[schoolID].TotalPortions += allocation.Portions
	}

	// Step 3: Convert map to sorted array (Requirement 8.4)
	result := make([]SchoolAllocationDisplay, 0, len(schoolMap))
	for _, display := range schoolMap {
		result = append(result, *display)
	}

	// Sort by school name alphabetically (Requirement 8.4)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].SchoolName > result[j].SchoolName {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}
