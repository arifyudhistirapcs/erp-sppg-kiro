package services

import (
	"errors"
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

// GetMenuPlanByID retrieves a menu plan by ID
func (s *MenuPlanningService) GetMenuPlanByID(id uint) (*models.MenuPlan, error) {
	var menuPlan models.MenuPlan
	err := s.db.Preload("MenuItems.Recipe.RecipeIngredients.Ingredient").
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
	err := s.db.Preload("MenuItems.Recipe").
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
	err := s.db.Preload("MenuItems.Recipe.RecipeIngredients.Ingredient").
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

	// Validate nutrition before approval
	if err := s.validateWeeklyNutrition(menuPlan.MenuItems); err != nil {
		return err
	}

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
		dailyMap[dateKey].TotalCalories += item.Recipe.TotalCalories * float64(item.Portions) / float64(item.Recipe.ServingSize)
		dailyMap[dateKey].TotalProtein += item.Recipe.TotalProtein * float64(item.Portions) / float64(item.Recipe.ServingSize)
		dailyMap[dateKey].TotalCarbs += item.Recipe.TotalCarbs * float64(item.Portions) / float64(item.Recipe.ServingSize)
		dailyMap[dateKey].TotalFat += item.Recipe.TotalFat * float64(item.Portions) / float64(item.Recipe.ServingSize)
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
		// Calculate scaling factor based on portions
		scaleFactor := float64(item.Portions) / float64(item.Recipe.ServingSize)

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

			sfGoodsMap[sfGoodsID].TotalQuantity += recipeItem.Quantity * scaleFactor
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
			if err := s.db.Preload("RecipeIngredients.Ingredient").First(&recipe, item.RecipeID).Error; err != nil {
				return err
			}
		} else {
			recipe = item.Recipe
		}

		daily := dailyMap[dateKey]
		daily.totalCalories += recipe.TotalCalories * float64(item.Portions) / float64(recipe.ServingSize)
		daily.totalProtein += recipe.TotalProtein * float64(item.Portions) / float64(recipe.ServingSize)
		daily.totalPortions += item.Portions
		dailyMap[dateKey] = daily
	}

	// Validate each day
	standards := DefaultNutritionStandards()
	for _, daily := range dailyMap {
		if daily.totalPortions == 0 {
			continue
		}

		caloriesPerPortion := daily.totalCalories / float64(daily.totalPortions)
		proteinPerPortion := daily.totalProtein / float64(daily.totalPortions)

		if caloriesPerPortion < standards.MinCalories || proteinPerPortion < standards.MinProtein {
			return ErrDailyNutritionInsufficient
		}
	}

	return nil
}
