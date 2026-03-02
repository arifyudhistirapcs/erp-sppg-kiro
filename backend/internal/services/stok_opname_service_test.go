package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupStokOpnameTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.StokOpnameForm{},
		&models.StokOpnameItem{},
		&models.Ingredient{},
		&models.InventoryItem{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// createStokOpnameTestUser is a helper function to create a test user with all required fields
func createStokOpnameTestUser(db *gorm.DB, nik, fullName, email, role string) *models.User {
	user := &models.User{
		NIK:      nik,
		FullName: fullName,
		Email:    email,
		Role:     role,
	}
	db.Create(user)
	return user
}

func TestCreateForm(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create a test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Test creating a form
	notes := "Test stok opname form"
	form, err := service.CreateForm(user.ID, notes)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, form)
	assert.NotEmpty(t, form.FormNumber)
	assert.Equal(t, user.ID, form.CreatedBy)
	assert.Equal(t, "pending", form.Status)
	assert.Equal(t, notes, form.Notes)
	assert.False(t, form.IsProcessed)
	assert.NotZero(t, form.CreatedAt)

	// Verify form number format: SO-YYYYMMDD-NNNN
	assert.Regexp(t, `^SO-\d{8}-\d{4}$`, form.FormNumber)

	// Verify the form was saved to database
	var savedForm models.StokOpnameForm
	err = db.First(&savedForm, form.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, form.FormNumber, savedForm.FormNumber)
}

func TestCreateForm_FormNumberGeneration(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create a test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create multiple forms on the same day
	form1, err := service.CreateForm(user.ID, "Form 1")
	assert.NoError(t, err)
	assert.NotNil(t, form1)

	form2, err := service.CreateForm(user.ID, "Form 2")
	assert.NoError(t, err)
	assert.NotNil(t, form2)

	form3, err := service.CreateForm(user.ID, "Form 3")
	assert.NoError(t, err)
	assert.NotNil(t, form3)

	// Verify form numbers are sequential
	today := time.Now().Format("20060102")
	assert.Equal(t, "SO-"+today+"-0001", form1.FormNumber)
	assert.Equal(t, "SO-"+today+"-0002", form2.FormNumber)
	assert.Equal(t, "SO-"+today+"-0003", form3.FormNumber)

	// Verify all forms are unique
	assert.NotEqual(t, form1.FormNumber, form2.FormNumber)
	assert.NotEqual(t, form2.FormNumber, form3.FormNumber)
	assert.NotEqual(t, form1.FormNumber, form3.FormNumber)
}

func TestCreateForm_AuditTrail(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create a test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Record time before creation
	beforeCreate := time.Now()

	// Create form
	form, err := service.CreateForm(user.ID, "Test form")
	assert.NoError(t, err)

	// Record time after creation
	afterCreate := time.Now()

	// Verify audit trail fields
	assert.Equal(t, user.ID, form.CreatedBy, "CreatedBy should match user ID")
	assert.True(t, form.CreatedAt.After(beforeCreate) || form.CreatedAt.Equal(beforeCreate), "CreatedAt should be after or equal to beforeCreate")
	assert.True(t, form.CreatedAt.Before(afterCreate) || form.CreatedAt.Equal(afterCreate), "CreatedAt should be before or equal to afterCreate")
	assert.Equal(t, "pending", form.Status, "Initial status should be pending")
}

func TestGetForm(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create a form
	form, err := service.CreateForm(creator.ID, "Test form with items")
	assert.NoError(t, err)

	// Add items to the form
	item := &models.StokOpnameItem{
		FormID:        form.ID,
		IngredientID:  ingredient.ID,
		SystemStock:   100.0,
		PhysicalCount: 95.0,
		Difference:    -5.0,
		ItemNotes:     "Test item",
	}
	db.Create(item)

	// Update form with approver
	db.Model(form).Updates(map[string]interface{}{
		"approved_by": approver.ID,
		"status":      "approved",
	})

	// Test GetForm
	retrievedForm, err := service.GetForm(form.ID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, retrievedForm)
	assert.Equal(t, form.ID, retrievedForm.ID)
	assert.Equal(t, form.FormNumber, retrievedForm.FormNumber)
	
	// Verify Creator relationship is loaded
	assert.NotNil(t, retrievedForm.Creator)
	assert.Equal(t, creator.ID, retrievedForm.Creator.ID)
	assert.Equal(t, creator.FullName, retrievedForm.Creator.FullName)
	
	// Verify Approver relationship is loaded
	assert.NotNil(t, retrievedForm.Approver)
	assert.Equal(t, approver.ID, retrievedForm.Approver.ID)
	assert.Equal(t, approver.FullName, retrievedForm.Approver.FullName)
	
	// Verify Items are loaded
	assert.Len(t, retrievedForm.Items, 1)
	assert.Equal(t, item.ID, retrievedForm.Items[0].ID)
	assert.Equal(t, 100.0, retrievedForm.Items[0].SystemStock)
	assert.Equal(t, 95.0, retrievedForm.Items[0].PhysicalCount)
	assert.Equal(t, -5.0, retrievedForm.Items[0].Difference)
	
	// Verify Ingredient relationship is loaded for each item
	assert.NotNil(t, retrievedForm.Items[0].Ingredient)
	assert.Equal(t, ingredient.ID, retrievedForm.Items[0].Ingredient.ID)
	assert.Equal(t, ingredient.Name, retrievedForm.Items[0].Ingredient.Name)
}

func TestGetForm_NotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Try to get a non-existent form
	form, err := service.GetForm(999)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, form)
	assert.Equal(t, ErrFormNotFound, err)
}

func TestGetAllForms_NoFilters(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test users
	user1 := createStokOpnameTestUser(db, "1111111111", "User One", "user1@example.com", "staff")
	user2 := createStokOpnameTestUser(db, "2222222222", "User Two", "user2@example.com", "staff")

	// Create multiple forms
	form1, _ := service.CreateForm(user1.ID, "Form 1")
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	form2, _ := service.CreateForm(user2.ID, "Form 2")
	time.Sleep(10 * time.Millisecond)
	form3, _ := service.CreateForm(user1.ID, "Form 3")

	// Get all forms without filters
	filters := FormFilters{
		Page:     1,
		PageSize: 20,
	}
	forms, totalCount, err := service.GetAllForms(filters)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, forms, 3)
	assert.Equal(t, 3, totalCount)
	
	// Verify sorting (newest first - created_at DESC)
	assert.Equal(t, form3.ID, forms[0].ID)
	assert.Equal(t, form2.ID, forms[1].ID)
	assert.Equal(t, form1.ID, forms[2].ID)
	
	// Verify Creator is preloaded
	assert.NotNil(t, forms[0].Creator)
	assert.Equal(t, user1.FullName, forms[0].Creator.FullName)
}

func TestGetAllForms_StatusFilter(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create forms with different statuses
	form1, _ := service.CreateForm(user.ID, "Pending form")
	form2, _ := service.CreateForm(user.ID, "Approved form")
	form3, _ := service.CreateForm(user.ID, "Rejected form")

	// Update statuses
	db.Model(form2).Update("status", "approved")
	db.Model(form3).Update("status", "rejected")

	// Filter by pending status
	filters := FormFilters{
		Status:   "pending",
		Page:     1,
		PageSize: 20,
	}
	forms, totalCount, err := service.GetAllForms(filters)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, forms, 1)
	assert.Equal(t, 1, totalCount)
	assert.Equal(t, form1.ID, forms[0].ID)
	assert.Equal(t, "pending", forms[0].Status)

	// Filter by approved status
	filters.Status = "approved"
	forms, totalCount, err = service.GetAllForms(filters)
	assert.NoError(t, err)
	assert.Len(t, forms, 1)
	assert.Equal(t, 1, totalCount)
	assert.Equal(t, form2.ID, forms[0].ID)
	assert.Equal(t, "approved", forms[0].Status)
}

func TestGetAllForms_CreatorFilter(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test users
	user1 := createStokOpnameTestUser(db, "1111111111", "User One", "user1@example.com", "staff")
	user2 := createStokOpnameTestUser(db, "2222222222", "User Two", "user2@example.com", "staff")

	// Create forms by different users
	_, _ = service.CreateForm(user1.ID, "Form by user 1")
	_, _ = service.CreateForm(user2.ID, "Form by user 2")
	_, _ = service.CreateForm(user1.ID, "Another form by user 1")

	// Filter by user1
	filters := FormFilters{
		CreatedBy: &user1.ID,
		Page:      1,
		PageSize:  20,
	}
	forms, totalCount, err := service.GetAllForms(filters)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, forms, 2)
	assert.Equal(t, 2, totalCount)
	
	// Verify both forms are by user1
	for _, form := range forms {
		assert.Equal(t, user1.ID, form.CreatedBy)
	}
}

func TestGetAllForms_DateRangeFilter(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create forms with different dates
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	form1, _ := service.CreateForm(user.ID, "Form 1")
	db.Model(form1).Update("created_at", yesterday)

	form2, _ := service.CreateForm(user.ID, "Form 2")
	db.Model(form2).Update("created_at", now)

	// Don't create form3 or update it - we only want 2 forms in the date range

	// Filter by date range (yesterday to now)
	startDate := yesterday.Add(-1 * time.Hour)
	endDate := now.Add(1 * time.Hour)
	filters := FormFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
		Page:      1,
		PageSize:  20,
	}
	forms, totalCount, err := service.GetAllForms(filters)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, forms, 2)
	assert.Equal(t, 2, totalCount)
}

func TestGetAllForms_SearchText(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test users with distinct names
	user1 := createStokOpnameTestUser(db, "1111111111", "John Doe", "john@example.com", "staff")
	user2 := createStokOpnameTestUser(db, "2222222222", "Jane Smith", "jane@example.com", "staff")

	// Create forms with different notes
	form1, _ := service.CreateForm(user1.ID, "Form with special keyword")
	form2, _ := service.CreateForm(user2.ID, "Regular form")
	_, _ = service.CreateForm(user1.ID, "Another regular form")

	// Search by notes keyword
	filters := FormFilters{
		SearchText: "special",
		Page:       1,
		PageSize:   20,
	}
	forms, totalCount, err := service.GetAllForms(filters)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, forms, 1)
	assert.Equal(t, 1, totalCount)
	assert.Equal(t, form1.ID, forms[0].ID)

	// Search by creator name
	filters.SearchText = "Jane"
	forms, totalCount, err = service.GetAllForms(filters)
	assert.NoError(t, err)
	assert.Len(t, forms, 1)
	assert.Equal(t, 1, totalCount)
	assert.Equal(t, form2.ID, forms[0].ID)
}

func TestGetAllForms_Pagination(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create 25 forms
	for i := 1; i <= 25; i++ {
		service.CreateForm(user.ID, fmt.Sprintf("Form %d", i))
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Get first page (20 items)
	filters := FormFilters{
		Page:     1,
		PageSize: 20,
	}
	forms, totalCount, err := service.GetAllForms(filters)

	// Assertions for first page
	assert.NoError(t, err)
	assert.Len(t, forms, 20)
	assert.Equal(t, 25, totalCount)

	// Get second page (5 items)
	filters.Page = 2
	forms, totalCount, err = service.GetAllForms(filters)

	// Assertions for second page
	assert.NoError(t, err)
	assert.Len(t, forms, 5)
	assert.Equal(t, 25, totalCount)
}

func TestGetAllForms_DefaultPagination(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create 5 forms
	for i := 1; i <= 5; i++ {
		service.CreateForm(user.ID, fmt.Sprintf("Form %d", i))
	}

	// Get forms with default pagination (page 0 or negative should default to 1)
	filters := FormFilters{
		Page:     0,
		PageSize: 0,
	}
	forms, totalCount, err := service.GetAllForms(filters)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, forms, 5)
	assert.Equal(t, 5, totalCount)
}

func TestUpdateFormNotes_Success(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Original notes")
	assert.NoError(t, err)

	// Update form notes
	newNotes := "Updated notes for testing"
	err = service.UpdateFormNotes(form.ID, newNotes)

	// Assertions
	assert.NoError(t, err)

	// Verify notes were updated in database
	var updatedForm models.StokOpnameForm
	db.First(&updatedForm, form.ID)
	assert.Equal(t, newNotes, updatedForm.Notes)
	assert.Equal(t, "pending", updatedForm.Status)
}

func TestUpdateFormNotes_FormNotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Try to update notes for non-existent form
	err := service.UpdateFormNotes(999, "Some notes")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotFound, err)
}

func TestUpdateFormNotes_FormNotPending(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create a form and approve it
	form, err := service.CreateForm(creator.ID, "Original notes")
	assert.NoError(t, err)

	// Update form status to approved
	db.Model(form).Updates(map[string]interface{}{
		"status":      "approved",
		"approved_by": approver.ID,
	})

	// Try to update notes for approved form
	err = service.UpdateFormNotes(form.ID, "New notes")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotPending, err)

	// Verify notes were NOT updated
	var unchangedForm models.StokOpnameForm
	db.First(&unchangedForm, form.ID)
	assert.Equal(t, "Original notes", unchangedForm.Notes)
}

func TestUpdateFormNotes_RejectedForm(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create a form and reject it
	form, err := service.CreateForm(creator.ID, "Original notes")
	assert.NoError(t, err)

	// Update form status to rejected
	db.Model(form).Updates(map[string]interface{}{
		"status":           "rejected",
		"approved_by":      approver.ID,
		"rejection_reason": "Invalid data",
	})

	// Try to update notes for rejected form
	err = service.UpdateFormNotes(form.ID, "New notes")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotPending, err)

	// Verify notes were NOT updated
	var unchangedForm models.StokOpnameForm
	db.First(&unchangedForm, form.ID)
	assert.Equal(t, "Original notes", unchangedForm.Notes)
}

func TestUpdateFormNotes_EmptyNotes(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create a pending form with notes
	form, err := service.CreateForm(user.ID, "Original notes")
	assert.NoError(t, err)

	// Update form notes to empty string
	err = service.UpdateFormNotes(form.ID, "")

	// Assertions
	assert.NoError(t, err)

	// Verify notes were updated to empty string
	var updatedForm models.StokOpnameForm
	db.First(&updatedForm, form.ID)
	assert.Equal(t, "", updatedForm.Notes)
}

func TestDeleteForm_Success(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Form to be deleted")
	assert.NoError(t, err)

	// Add items to the form
	item1 := &models.StokOpnameItem{
		FormID:        form.ID,
		IngredientID:  ingredient.ID,
		SystemStock:   100.0,
		PhysicalCount: 95.0,
		Difference:    -5.0,
		ItemNotes:     "Test item 1",
	}
	db.Create(item1)

	item2 := &models.StokOpnameItem{
		FormID:        form.ID,
		IngredientID:  ingredient.ID,
		SystemStock:   50.0,
		PhysicalCount: 48.0,
		Difference:    -2.0,
		ItemNotes:     "Test item 2",
	}
	db.Create(item2)

	// Verify form and items exist
	var formCount int64
	db.Model(&models.StokOpnameForm{}).Where("id = ?", form.ID).Count(&formCount)
	assert.Equal(t, int64(1), formCount)

	var itemCount int64
	db.Model(&models.StokOpnameItem{}).Where("form_id = ?", form.ID).Count(&itemCount)
	assert.Equal(t, int64(2), itemCount)

	// Delete the form
	err = service.DeleteForm(form.ID)

	// Assertions
	assert.NoError(t, err)

	// Verify form was deleted
	db.Model(&models.StokOpnameForm{}).Where("id = ?", form.ID).Count(&formCount)
	assert.Equal(t, int64(0), formCount)

	// Verify items were cascade deleted
	db.Model(&models.StokOpnameItem{}).Where("form_id = ?", form.ID).Count(&itemCount)
	assert.Equal(t, int64(0), itemCount)
}

func TestDeleteForm_FormNotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Try to delete a non-existent form
	err := service.DeleteForm(999)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotFound, err)
}

func TestDeleteForm_FormNotPending(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create a form and approve it
	form, err := service.CreateForm(creator.ID, "Approved form")
	assert.NoError(t, err)

	// Update form status to approved
	db.Model(form).Updates(map[string]interface{}{
		"status":      "approved",
		"approved_by": approver.ID,
	})

	// Try to delete approved form
	err = service.DeleteForm(form.ID)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotPending, err)

	// Verify form was NOT deleted
	var formCount int64
	db.Model(&models.StokOpnameForm{}).Where("id = ?", form.ID).Count(&formCount)
	assert.Equal(t, int64(1), formCount)
}

func TestDeleteForm_RejectedForm(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create a form and reject it
	form, err := service.CreateForm(creator.ID, "Rejected form")
	assert.NoError(t, err)

	// Update form status to rejected
	db.Model(form).Updates(map[string]interface{}{
		"status":           "rejected",
		"approved_by":      approver.ID,
		"rejection_reason": "Invalid data",
	})

	// Try to delete rejected form
	err = service.DeleteForm(form.ID)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotPending, err)

	// Verify form was NOT deleted
	var formCount int64
	db.Model(&models.StokOpnameForm{}).Where("id = ?", form.ID).Count(&formCount)
	assert.Equal(t, int64(1), formCount)
}

func TestDeleteForm_CascadeDeleteItems(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create test ingredients
	ingredient1 := &models.Ingredient{
		Name:     "Ingredient 1",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient1)

	ingredient2 := &models.Ingredient{
		Name:     "Ingredient 2",
		Unit:     "liter",
		Category: "test",
	}
	db.Create(ingredient2)

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Form with multiple items")
	assert.NoError(t, err)

	// Add multiple items to the form
	items := []models.StokOpnameItem{
		{
			FormID:        form.ID,
			IngredientID:  ingredient1.ID,
			SystemStock:   100.0,
			PhysicalCount: 95.0,
			Difference:    -5.0,
			ItemNotes:     "Item 1",
		},
		{
			FormID:        form.ID,
			IngredientID:  ingredient2.ID,
			SystemStock:   50.0,
			PhysicalCount: 52.0,
			Difference:    2.0,
			ItemNotes:     "Item 2",
		},
	}

	for _, item := range items {
		db.Create(&item)
	}

	// Verify items exist
	var itemCount int64
	db.Model(&models.StokOpnameItem{}).Where("form_id = ?", form.ID).Count(&itemCount)
	assert.Equal(t, int64(2), itemCount)

	// Delete the form
	err = service.DeleteForm(form.ID)
	assert.NoError(t, err)

	// Verify all items were cascade deleted
	db.Model(&models.StokOpnameItem{}).Where("form_id = ?", form.ID).Count(&itemCount)
	assert.Equal(t, int64(0), itemCount)

	// Verify ingredients still exist (should not be deleted)
	var ingredientCount int64
	db.Model(&models.Ingredient{}).Count(&ingredientCount)
	assert.Equal(t, int64(2), ingredientCount)
}

func TestAddItem_Success(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create inventory item with system stock
	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Test form")
	assert.NoError(t, err)

	// Add item to form
	physicalCount := 95.0
	itemNotes := "Test item notes"
	err = service.AddItem(form.ID, ingredient.ID, physicalCount, itemNotes)

	// Assertions
	assert.NoError(t, err)

	// Verify item was created in database
	var item models.StokOpnameItem
	err = db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient.ID).First(&item).Error
	assert.NoError(t, err)
	assert.Equal(t, form.ID, item.FormID)
	assert.Equal(t, ingredient.ID, item.IngredientID)
	assert.Equal(t, 100.0, item.SystemStock)
	assert.Equal(t, 95.0, item.PhysicalCount)
	assert.Equal(t, -5.0, item.Difference) // 95 - 100 = -5
	assert.Equal(t, itemNotes, item.ItemNotes)
}

func TestAddItem_FormNotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Try to add item to non-existent form
	err := service.AddItem(999, 1, 100.0, "notes")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotFound, err)
}

func TestAddItem_FormNotPending(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create inventory item
	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create a form and approve it
	form, err := service.CreateForm(creator.ID, "Approved form")
	assert.NoError(t, err)

	// Update form status to approved
	db.Model(form).Updates(map[string]interface{}{
		"status":      "approved",
		"approved_by": approver.ID,
	})

	// Try to add item to approved form
	err = service.AddItem(form.ID, ingredient.ID, 95.0, "notes")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotPending, err)

	// Verify item was NOT created
	var itemCount int64
	db.Model(&models.StokOpnameItem{}).Where("form_id = ?", form.ID).Count(&itemCount)
	assert.Equal(t, int64(0), itemCount)
}

func TestAddItem_DuplicateIngredient(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create inventory item
	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Test form")
	assert.NoError(t, err)

	// Add item to form (first time)
	err = service.AddItem(form.ID, ingredient.ID, 95.0, "First item")
	assert.NoError(t, err)

	// Try to add the same ingredient again
	err = service.AddItem(form.ID, ingredient.ID, 90.0, "Duplicate item")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrDuplicateIngredient, err)

	// Verify only one item exists
	var itemCount int64
	db.Model(&models.StokOpnameItem{}).Where("form_id = ?", form.ID).Count(&itemCount)
	assert.Equal(t, int64(1), itemCount)
}

func TestAddItem_SystemStockCapture(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create inventory item with specific system stock
	systemStock := 123.45
	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     systemStock,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Test form")
	assert.NoError(t, err)

	// Add item to form
	physicalCount := 120.0
	err = service.AddItem(form.ID, ingredient.ID, physicalCount, "")
	assert.NoError(t, err)

	// Verify system stock was captured correctly
	var item models.StokOpnameItem
	err = db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient.ID).First(&item).Error
	assert.NoError(t, err)
	assert.Equal(t, systemStock, item.SystemStock)
	assert.Equal(t, physicalCount, item.PhysicalCount)
}

func TestAddItem_DifferenceCalculation(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create test ingredients
	ingredient1 := &models.Ingredient{
		Name:     "Ingredient 1",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient1)

	ingredient2 := &models.Ingredient{
		Name:     "Ingredient 2",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient2)

	ingredient3 := &models.Ingredient{
		Name:     "Ingredient 3",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient3)

	// Create inventory items
	db.Create(&models.InventoryItem{
		IngredientID: ingredient1.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	})

	db.Create(&models.InventoryItem{
		IngredientID: ingredient2.ID,
		Quantity:     50.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	})

	db.Create(&models.InventoryItem{
		IngredientID: ingredient3.ID,
		Quantity:     75.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	})

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Test form")
	assert.NoError(t, err)

	// Test case 1: Physical count < System stock (negative difference)
	err = service.AddItem(form.ID, ingredient1.ID, 95.0, "")
	assert.NoError(t, err)

	var item1 models.StokOpnameItem
	db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient1.ID).First(&item1)
	assert.Equal(t, -5.0, item1.Difference) // 95 - 100 = -5

	// Test case 2: Physical count > System stock (positive difference)
	err = service.AddItem(form.ID, ingredient2.ID, 55.0, "")
	assert.NoError(t, err)

	var item2 models.StokOpnameItem
	db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient2.ID).First(&item2)
	assert.Equal(t, 5.0, item2.Difference) // 55 - 50 = 5

	// Test case 3: Physical count = System stock (zero difference)
	err = service.AddItem(form.ID, ingredient3.ID, 75.0, "")
	assert.NoError(t, err)

	var item3 models.StokOpnameItem
	db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient3.ID).First(&item3)
	assert.Equal(t, 0.0, item3.Difference) // 75 - 75 = 0
}

func TestAddItem_MultipleItems(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create multiple test ingredients
	ingredients := make([]*models.Ingredient, 5)
	for i := 0; i < 5; i++ {
		ingredient := &models.Ingredient{
			Name:     fmt.Sprintf("Ingredient %d", i+1),
			Unit:     "kg",
			Category: "test",
		}
		db.Create(ingredient)
		ingredients[i] = ingredient

		// Create inventory item for each ingredient
		db.Create(&models.InventoryItem{
			IngredientID: ingredient.ID,
			Quantity:     float64((i + 1) * 10),
			MinThreshold: 5.0,
			LastUpdated:  time.Now(),
		})
	}

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Test form with multiple items")
	assert.NoError(t, err)

	// Add all ingredients to the form
	for i, ingredient := range ingredients {
		err = service.AddItem(form.ID, ingredient.ID, float64((i+1)*10-2), fmt.Sprintf("Item %d notes", i+1))
		assert.NoError(t, err)
	}

	// Verify all items were added
	var itemCount int64
	db.Model(&models.StokOpnameItem{}).Where("form_id = ?", form.ID).Count(&itemCount)
	assert.Equal(t, int64(5), itemCount)

	// Verify each item has correct data
	var items []models.StokOpnameItem
	db.Where("form_id = ?", form.ID).Order("ingredient_id").Find(&items)
	assert.Len(t, items, 5)

	for i, item := range items {
		expectedSystemStock := float64((i + 1) * 10)
		expectedPhysicalCount := float64((i + 1) * 10 - 2)
		expectedDifference := expectedPhysicalCount - expectedSystemStock

		assert.Equal(t, ingredients[i].ID, item.IngredientID)
		assert.Equal(t, expectedSystemStock, item.SystemStock)
		assert.Equal(t, expectedPhysicalCount, item.PhysicalCount)
		assert.Equal(t, expectedDifference, item.Difference)
		assert.Equal(t, fmt.Sprintf("Item %d notes", i+1), item.ItemNotes)
	}
}


func TestUpdateItem_Success(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create inventory item with system stock
	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Test form")
	assert.NoError(t, err)

	// Add item to form
	err = service.AddItem(form.ID, ingredient.ID, 95.0, "Original notes")
	assert.NoError(t, err)

	// Get the item ID
	var item models.StokOpnameItem
	err = db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient.ID).First(&item).Error
	assert.NoError(t, err)

	// Update the item
	newPhysicalCount := 110.0
	newNotes := "Updated notes"
	err = service.UpdateItem(item.ID, newPhysicalCount, newNotes)

	// Assertions
	assert.NoError(t, err)

	// Verify item was updated in database
	var updatedItem models.StokOpnameItem
	err = db.First(&updatedItem, item.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, newPhysicalCount, updatedItem.PhysicalCount)
	assert.Equal(t, newNotes, updatedItem.ItemNotes)
	assert.Equal(t, 100.0, updatedItem.SystemStock) // System stock should not change
	assert.Equal(t, 10.0, updatedItem.Difference)   // 110 - 100 = 10
}

func TestUpdateItem_ItemNotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Try to update non-existent item
	err := service.UpdateItem(999, 100.0, "notes")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrItemNotFound, err)
}

func TestUpdateItem_FormNotPending(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create inventory item
	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create a form and add item
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	err = service.AddItem(form.ID, ingredient.ID, 95.0, "Original notes")
	assert.NoError(t, err)

	// Get the item ID
	var item models.StokOpnameItem
	err = db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient.ID).First(&item).Error
	assert.NoError(t, err)

	// Update form status to approved
	db.Model(form).Updates(map[string]interface{}{
		"status":      "approved",
		"approved_by": approver.ID,
	})

	// Try to update item in approved form
	err = service.UpdateItem(item.ID, 110.0, "Updated notes")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotPending, err)

	// Verify item was NOT updated
	var unchangedItem models.StokOpnameItem
	db.First(&unchangedItem, item.ID)
	assert.Equal(t, 95.0, unchangedItem.PhysicalCount)
	assert.Equal(t, "Original notes", unchangedItem.ItemNotes)
}

func TestUpdateItem_RejectedForm(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create inventory item
	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create a form and add item
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	err = service.AddItem(form.ID, ingredient.ID, 95.0, "Original notes")
	assert.NoError(t, err)

	// Get the item ID
	var item models.StokOpnameItem
	err = db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient.ID).First(&item).Error
	assert.NoError(t, err)

	// Update form status to rejected
	db.Model(form).Updates(map[string]interface{}{
		"status":           "rejected",
		"approved_by":      approver.ID,
		"rejection_reason": "Test rejection",
	})

	// Try to update item in rejected form
	err = service.UpdateItem(item.ID, 110.0, "Updated notes")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotPending, err)

	// Verify item was NOT updated
	var unchangedItem models.StokOpnameItem
	db.First(&unchangedItem, item.ID)
	assert.Equal(t, 95.0, unchangedItem.PhysicalCount)
	assert.Equal(t, "Original notes", unchangedItem.ItemNotes)
}

func TestUpdateItem_DifferenceRecalculation(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create inventory item with system stock
	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Test form")
	assert.NoError(t, err)

	// Add item to form
	err = service.AddItem(form.ID, ingredient.ID, 95.0, "Original notes")
	assert.NoError(t, err)

	// Get the item ID
	var item models.StokOpnameItem
	err = db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient.ID).First(&item).Error
	assert.NoError(t, err)

	// Test case 1: Update to higher physical count (positive difference)
	err = service.UpdateItem(item.ID, 120.0, "Higher count")
	assert.NoError(t, err)

	var updatedItem models.StokOpnameItem
	db.First(&updatedItem, item.ID)
	assert.Equal(t, 120.0, updatedItem.PhysicalCount)
	assert.Equal(t, 20.0, updatedItem.Difference) // 120 - 100 = 20

	// Test case 2: Update to lower physical count (negative difference)
	err = service.UpdateItem(item.ID, 80.0, "Lower count")
	assert.NoError(t, err)

	db.First(&updatedItem, item.ID)
	assert.Equal(t, 80.0, updatedItem.PhysicalCount)
	assert.Equal(t, -20.0, updatedItem.Difference) // 80 - 100 = -20

	// Test case 3: Update to match system stock (zero difference)
	err = service.UpdateItem(item.ID, 100.0, "Exact match")
	assert.NoError(t, err)

	db.First(&updatedItem, item.ID)
	assert.Equal(t, 100.0, updatedItem.PhysicalCount)
	assert.Equal(t, 0.0, updatedItem.Difference) // 100 - 100 = 0
}

func TestUpdateItem_EmptyNotes(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, &inventoryService, nil)

	// Create test user
	user := createStokOpnameTestUser(db, "1234567890", "Test User", "test@example.com", "staff")

	// Create test ingredient
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Unit:     "kg",
		Category: "test",
	}
	db.Create(ingredient)

	// Create inventory item with system stock
	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create a pending form
	form, err := service.CreateForm(user.ID, "Test form")
	assert.NoError(t, err)

	// Add item to form with notes
	err = service.AddItem(form.ID, ingredient.ID, 95.0, "Original notes")
	assert.NoError(t, err)

	// Get the item ID
	var item models.StokOpnameItem
	err = db.Where("form_id = ? AND ingredient_id = ?", form.ID, ingredient.ID).First(&item).Error
	assert.NoError(t, err)

	// Update the item with empty notes
	err = service.UpdateItem(item.ID, 110.0, "")
	assert.NoError(t, err)

	// Verify notes were cleared
	var updatedItem models.StokOpnameItem
	db.First(&updatedItem, item.ID)
	assert.Equal(t, "", updatedItem.ItemNotes)
	assert.Equal(t, 110.0, updatedItem.PhysicalCount)
}

func TestApproveForm_Success(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	
	// Need to migrate InventoryMovement model
	err := db.AutoMigrate(&models.InventoryMovement{})
	assert.NoError(t, err)
	
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, inventoryService, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create test ingredients
	ingredient1 := &models.Ingredient{
		Name:     "Test Ingredient 1",
		Category: "Test Category",
		Unit:     "kg",
	}
	db.Create(ingredient1)

	ingredient2 := &models.Ingredient{
		Name:     "Test Ingredient 2",
		Category: "Test Category",
		Unit:     "kg",
	}
	db.Create(ingredient2)

	// Create inventory items with initial stock
	inventoryItem1 := &models.InventoryItem{
		IngredientID: ingredient1.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem1)

	inventoryItem2 := &models.InventoryItem{
		IngredientID: ingredient2.ID,
		Quantity:     50.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem2)

	// Create a stok opname form
	form, err := service.CreateForm(creator.ID, "Test approval form")
	assert.NoError(t, err)

	// Add items with differences
	err = service.AddItem(form.ID, ingredient1.ID, 120.0, "Positive difference") // +20
	assert.NoError(t, err)

	err = service.AddItem(form.ID, ingredient2.ID, 40.0, "Negative difference") // -10
	assert.NoError(t, err)

	// Approve the form
	err = service.ApproveForm(form.ID, approver.ID)
	assert.NoError(t, err)

	// Verify form status updated
	var updatedForm models.StokOpnameForm
	err = db.First(&updatedForm, form.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "approved", updatedForm.Status)
	assert.Equal(t, approver.ID, *updatedForm.ApprovedBy)
	assert.NotNil(t, updatedForm.ApprovedAt)
	assert.True(t, updatedForm.IsProcessed)

	// Verify inventory stocks updated
	var updatedInventory1 models.InventoryItem
	err = db.Where("ingredient_id = ?", ingredient1.ID).First(&updatedInventory1).Error
	assert.NoError(t, err)
	assert.Equal(t, 120.0, updatedInventory1.Quantity)

	var updatedInventory2 models.InventoryItem
	err = db.Where("ingredient_id = ?", ingredient2.ID).First(&updatedInventory2).Error
	assert.NoError(t, err)
	assert.Equal(t, 40.0, updatedInventory2.Quantity)

	// Verify inventory movements created
	var movements []models.InventoryMovement
	err = db.Where("reference LIKE ?", "Stok Opname: %").Find(&movements).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, len(movements))

	// Check movement details
	for _, movement := range movements {
		assert.Equal(t, "adjustment", movement.MovementType)
		assert.Equal(t, approver.ID, movement.CreatedBy)
		assert.Contains(t, movement.Reference, form.FormNumber)
		
		if movement.IngredientID == ingredient1.ID {
			assert.Equal(t, 20.0, movement.Quantity) // absolute value of difference
		} else if movement.IngredientID == ingredient2.ID {
			assert.Equal(t, 10.0, movement.Quantity) // absolute value of difference
		}
	}
}

func TestApproveForm_FormNotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create approver
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Try to approve non-existent form
	err := service.ApproveForm(999, approver.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotFound, err)
}

func TestApproveForm_ApproverNotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create creator and form
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	// Try to approve with non-existent approver
	err = service.ApproveForm(form.ID, 999)
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestApproveForm_ApproverNotKepalaSSPG(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	nonApprover := createStokOpnameTestUser(db, "0987654321", "Non-Approver User", "nonapprover@example.com", "chef")

	// Create form
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	// Try to approve with non-kepala_sppg user
	err = service.ApproveForm(form.ID, nonApprover.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestApproveForm_CreatorCannotApproveOwnForm(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create user who is both creator and kepala_sppg
	creatorApprover := createStokOpnameTestUser(db, "1234567890", "Creator Approver", "creator@example.com", "kepala_sppg")

	// Create form
	form, err := service.CreateForm(creatorApprover.ID, "Test form")
	assert.NoError(t, err)

	// Try to approve own form
	err = service.ApproveForm(form.ID, creatorApprover.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestApproveForm_AlreadyProcessed(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	
	// Need to migrate InventoryMovement model
	err := db.AutoMigrate(&models.InventoryMovement{})
	assert.NoError(t, err)
	
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, inventoryService, nil)

	// Create users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create ingredient and inventory
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Category: "Test Category",
		Unit:     "kg",
	}
	db.Create(ingredient)

	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create form and add item
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	err = service.AddItem(form.ID, ingredient.ID, 120.0, "Test item")
	assert.NoError(t, err)

	// Approve the form first time
	err = service.ApproveForm(form.ID, approver.ID)
	assert.NoError(t, err)

	// Try to approve again
	err = service.ApproveForm(form.ID, approver.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrFormAlreadyProcessed, err)
}

func TestApproveForm_WithNoDifference(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	
	// Need to migrate InventoryMovement model
	err := db.AutoMigrate(&models.InventoryMovement{})
	assert.NoError(t, err)
	
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, inventoryService, nil)

	// Create users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create ingredient and inventory
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Category: "Test Category",
		Unit:     "kg",
	}
	db.Create(ingredient)

	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create form and add item with no difference
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	err = service.AddItem(form.ID, ingredient.ID, 100.0, "No difference") // Same as system stock
	assert.NoError(t, err)

	// Approve the form
	err = service.ApproveForm(form.ID, approver.ID)
	assert.NoError(t, err)

	// Verify form approved
	var updatedForm models.StokOpnameForm
	err = db.First(&updatedForm, form.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "approved", updatedForm.Status)
	assert.True(t, updatedForm.IsProcessed)

	// Verify no inventory movements created (since there's no difference)
	var movements []models.InventoryMovement
	err = db.Where("reference LIKE ?", "Stok Opname: %").Find(&movements).Error
	assert.NoError(t, err)
	assert.Equal(t, 0, len(movements))

	// Verify inventory stock unchanged
	var updatedInventory models.InventoryItem
	err = db.Where("ingredient_id = ?", ingredient.ID).First(&updatedInventory).Error
	assert.NoError(t, err)
	assert.Equal(t, 100.0, updatedInventory.Quantity)
}

func TestApproveForm_WithNewIngredient(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	
	// Need to migrate InventoryMovement model
	err := db.AutoMigrate(&models.InventoryMovement{})
	assert.NoError(t, err)
	
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, inventoryService, nil)

	// Create users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create ingredient WITHOUT inventory item (new ingredient)
	ingredient := &models.Ingredient{
		Name:     "New Ingredient",
		Category: "Test Category",
		Unit:     "kg",
	}
	db.Create(ingredient)

	// Create form and add item
	form, err := service.CreateForm(creator.ID, "Test form with new ingredient")
	assert.NoError(t, err)

	err = service.AddItem(form.ID, ingredient.ID, 50.0, "New ingredient")
	assert.NoError(t, err)

	// Approve the form
	err = service.ApproveForm(form.ID, approver.ID)
	assert.NoError(t, err)

	// Verify inventory item created with correct quantity
	var newInventory models.InventoryItem
	err = db.Where("ingredient_id = ?", ingredient.ID).First(&newInventory).Error
	assert.NoError(t, err)
	assert.Equal(t, 50.0, newInventory.Quantity)
	assert.Equal(t, 10.0, newInventory.MinThreshold) // default threshold

	// Verify inventory movement created
	var movements []models.InventoryMovement
	err = db.Where("reference LIKE ?", "Stok Opname: %").Find(&movements).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, len(movements))
	assert.Equal(t, "adjustment", movements[0].MovementType)
	assert.Equal(t, 50.0, movements[0].Quantity)
}

func TestRejectForm_Success(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create a stok opname form
	form, err := service.CreateForm(creator.ID, "Test rejection form")
	assert.NoError(t, err)

	// Reject the form
	reason := "Data tidak akurat, perlu penghitungan ulang"
	err = service.RejectForm(form.ID, approver.ID, reason)
	assert.NoError(t, err)

	// Verify form status updated
	var updatedForm models.StokOpnameForm
	err = db.First(&updatedForm, form.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "rejected", updatedForm.Status)
	assert.Equal(t, approver.ID, *updatedForm.ApprovedBy)
	assert.NotNil(t, updatedForm.ApprovedAt)
	assert.Equal(t, reason, updatedForm.RejectionReason)
	assert.False(t, updatedForm.IsProcessed) // Should NOT be processed on rejection
}

func TestRejectForm_FormNotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create approver
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Try to reject non-existent form
	err := service.RejectForm(999, approver.ID, "Test reason")
	assert.Error(t, err)
	assert.Equal(t, ErrFormNotFound, err)
}

func TestRejectForm_ApproverNotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create creator and form
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	// Try to reject with non-existent approver
	err = service.RejectForm(form.ID, 999, "Test reason")
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestRejectForm_ApproverNotKepalaSSPG(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	nonApprover := createStokOpnameTestUser(db, "0987654321", "Non-Approver User", "nonapprover@example.com", "chef")

	// Create form
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	// Try to reject with non-kepala_sppg user
	err = service.RejectForm(form.ID, nonApprover.ID, "Test reason")
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestRejectForm_CreatorCannotRejectOwnForm(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create user who is both creator and kepala_sppg
	creatorApprover := createStokOpnameTestUser(db, "1234567890", "Creator Approver", "creator@example.com", "kepala_sppg")

	// Create form
	form, err := service.CreateForm(creatorApprover.ID, "Test form")
	assert.NoError(t, err)

	// Try to reject own form
	err = service.RejectForm(form.ID, creatorApprover.ID, "Test reason")
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestRejectForm_WithReason(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create form
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	// Reject with detailed reason
	reason := "Penghitungan tidak sesuai dengan prosedur. Beberapa item tidak dihitung dengan benar. Mohon lakukan penghitungan ulang dengan lebih teliti."
	err = service.RejectForm(form.ID, approver.ID, reason)
	assert.NoError(t, err)

	// Verify rejection reason stored correctly
	var updatedForm models.StokOpnameForm
	err = db.First(&updatedForm, form.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, reason, updatedForm.RejectionReason)
}

func TestRejectForm_DoesNotProcessStockAdjustments(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	
	// Need to migrate InventoryMovement model
	err := db.AutoMigrate(&models.InventoryMovement{})
	assert.NoError(t, err)
	
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, inventoryService, nil)

	// Create users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create ingredient and inventory
	ingredient := &models.Ingredient{
		Name:     "Test Ingredient",
		Category: "Test Category",
		Unit:     "kg",
	}
	db.Create(ingredient)

	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		MinThreshold: 10.0,
		LastUpdated:  time.Now(),
	}
	db.Create(inventoryItem)

	// Create form and add item with difference
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	err = service.AddItem(form.ID, ingredient.ID, 120.0, "Test item") // +20 difference
	assert.NoError(t, err)

	// Reject the form
	err = service.RejectForm(form.ID, approver.ID, "Test rejection")
	assert.NoError(t, err)

	// Verify inventory stock NOT updated (should still be 100.0)
	var updatedInventory models.InventoryItem
	err = db.Where("ingredient_id = ?", ingredient.ID).First(&updatedInventory).Error
	assert.NoError(t, err)
	assert.Equal(t, 100.0, updatedInventory.Quantity) // Should NOT change

	// Verify NO inventory movements created
	var movements []models.InventoryMovement
	err = db.Where("reference LIKE ?", "Stok Opname: %").Find(&movements).Error
	assert.NoError(t, err)
	assert.Equal(t, 0, len(movements)) // Should be empty

	// Verify is_processed flag is false
	var updatedForm models.StokOpnameForm
	err = db.First(&updatedForm, form.ID).Error
	assert.NoError(t, err)
	assert.False(t, updatedForm.IsProcessed)
}

func TestRejectForm_AuditTrail(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create users
	creator := createStokOpnameTestUser(db, "1234567890", "Creator User", "creator@example.com", "staff")
	approver := createStokOpnameTestUser(db, "0987654321", "Approver User", "approver@example.com", "kepala_sppg")

	// Create form
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	// Record time before rejection
	beforeReject := time.Now()

	// Reject the form
	err = service.RejectForm(form.ID, approver.ID, "Test reason")
	assert.NoError(t, err)

	// Record time after rejection
	afterReject := time.Now()

	// Verify audit trail
	var updatedForm models.StokOpnameForm
	err = db.First(&updatedForm, form.ID).Error
	assert.NoError(t, err)

	// Verify approver ID recorded
	assert.NotNil(t, updatedForm.ApprovedBy)
	assert.Equal(t, approver.ID, *updatedForm.ApprovedBy)

	// Verify timestamp recorded and within expected range
	assert.NotNil(t, updatedForm.ApprovedAt)
	assert.True(t, updatedForm.ApprovedAt.After(beforeReject) || updatedForm.ApprovedAt.Equal(beforeReject))
	assert.True(t, updatedForm.ApprovedAt.Before(afterReject) || updatedForm.ApprovedAt.Equal(afterReject))

	// Verify updated_at timestamp updated
	assert.True(t, updatedForm.UpdatedAt.After(beforeReject) || updatedForm.UpdatedAt.Equal(beforeReject))
}

func TestExportForm_ExcelFormat(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, inventoryService, nil)

	// Create test user
	creator := createStokOpnameTestUser(db, "12345", "Test Creator", "creator@test.com", "chef")

	// Create form
	form, err := service.CreateForm(creator.ID, "Test form for export")
	assert.NoError(t, err)

	// Create ingredients and inventory items
	ingredient1 := &models.Ingredient{
		Name:        "Beras",
		Unit:        "kg",
		Category:    "Bahan Pokok",
		Description: "Beras putih",
	}
	db.Create(ingredient1)

	inventoryItem1 := &models.InventoryItem{
		IngredientID: ingredient1.ID,
		Quantity:     100.0,
		Unit:         "kg",
		Location:     "Gudang A",
	}
	db.Create(inventoryItem1)

	ingredient2 := &models.Ingredient{
		Name:        "Gula",
		Unit:        "kg",
		Category:    "Bahan Pokok",
		Description: "Gula pasir",
	}
	db.Create(ingredient2)

	inventoryItem2 := &models.InventoryItem{
		IngredientID: ingredient2.ID,
		Quantity:     50.0,
		Unit:         "kg",
		Location:     "Gudang A",
	}
	db.Create(inventoryItem2)

	// Add items to form
	err = service.AddItem(form.ID, ingredient1.ID, 95.0, "Kurang 5kg")
	assert.NoError(t, err)

	err = service.AddItem(form.ID, ingredient2.ID, 52.0, "Lebih 2kg")
	assert.NoError(t, err)

	// Export form as Excel
	exportData, err := service.ExportForm(form.ID, "excel", "Test Exporter")
	assert.NoError(t, err)
	assert.NotNil(t, exportData)
	assert.Greater(t, len(exportData), 0, "Export data should not be empty")

	// Verify it's a valid Excel file by checking the file signature
	// Excel files start with PK (ZIP format)
	assert.Equal(t, byte('P'), exportData[0])
	assert.Equal(t, byte('K'), exportData[1])
}

func TestExportForm_PDFFormat(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, inventoryService, nil)

	// Create test user
	creator := createStokOpnameTestUser(db, "12345", "Test Creator", "creator@test.com", "chef")

	// Create form
	form, err := service.CreateForm(creator.ID, "Test form for PDF export")
	assert.NoError(t, err)

	// Create ingredient and inventory item
	ingredient := &models.Ingredient{
		Name:        "Beras",
		Unit:        "kg",
		Category:    "Bahan Pokok",
		Description: "Beras putih",
	}
	db.Create(ingredient)

	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		Unit:         "kg",
		Location:     "Gudang A",
	}
	db.Create(inventoryItem)

	// Add item to form
	err = service.AddItem(form.ID, ingredient.ID, 95.0, "Kurang 5kg")
	assert.NoError(t, err)

	// Export form as PDF
	exportData, err := service.ExportForm(form.ID, "pdf", "Test Exporter")
	assert.NoError(t, err)
	assert.NotNil(t, exportData)
	assert.Greater(t, len(exportData), 0, "Export data should not be empty")

	// Verify it's a valid PDF file by checking the file signature
	// PDF files start with %PDF
	assert.Equal(t, byte('%'), exportData[0])
	assert.Equal(t, byte('P'), exportData[1])
	assert.Equal(t, byte('D'), exportData[2])
	assert.Equal(t, byte('F'), exportData[3])
}

func TestExportForm_InvalidFormat(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Create test user
	creator := createStokOpnameTestUser(db, "12345", "Test Creator", "creator@test.com", "chef")

	// Create form
	form, err := service.CreateForm(creator.ID, "Test form")
	assert.NoError(t, err)

	// Try to export with invalid format
	_, err = service.ExportForm(form.ID, "invalid", "Test Exporter")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "format tidak valid")
}

func TestExportForm_FormNotFound(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	service := NewStokOpnameService(db, nil, nil)

	// Try to export non-existent form
	_, err := service.ExportForm(999, "excel", "Test Exporter")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gagal mengambil form")
}

func TestExportForm_WithApprovalInfo(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, inventoryService, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "12345", "Test Creator", "creator@test.com", "chef")
	approver := createStokOpnameTestUser(db, "67890", "Test Approver", "approver@test.com", "kepala_sppg")

	// Create form
	form, err := service.CreateForm(creator.ID, "Test form with approval")
	assert.NoError(t, err)

	// Create ingredient and inventory item
	ingredient := &models.Ingredient{
		Name:        "Beras",
		Unit:        "kg",
		Category:    "Bahan Pokok",
		Description: "Beras putih",
	}
	db.Create(ingredient)

	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		Unit:         "kg",
		Location:     "Gudang A",
	}
	db.Create(inventoryItem)

	// Add item to form
	err = service.AddItem(form.ID, ingredient.ID, 95.0, "Kurang 5kg")
	assert.NoError(t, err)

	// Approve the form
	err = service.ApproveForm(form.ID, approver.ID)
	assert.NoError(t, err)

	// Export form
	exportData, err := service.ExportForm(form.ID, "excel", "Test Exporter")
	assert.NoError(t, err)
	assert.NotNil(t, exportData)
	assert.Greater(t, len(exportData), 0, "Export data should not be empty")
}

func TestExportForm_WithRejectionInfo(t *testing.T) {
	db := setupStokOpnameTestDB(t)
	inventoryService := NewInventoryService(db)
	service := NewStokOpnameService(db, inventoryService, nil)

	// Create test users
	creator := createStokOpnameTestUser(db, "12345", "Test Creator", "creator@test.com", "chef")
	approver := createStokOpnameTestUser(db, "67890", "Test Approver", "approver@test.com", "kepala_sppg")

	// Create form
	form, err := service.CreateForm(creator.ID, "Test form with rejection")
	assert.NoError(t, err)

	// Create ingredient and inventory item
	ingredient := &models.Ingredient{
		Name:        "Beras",
		Unit:        "kg",
		Category:    "Bahan Pokok",
		Description: "Beras putih",
	}
	db.Create(ingredient)

	inventoryItem := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		Unit:         "kg",
		Location:     "Gudang A",
	}
	db.Create(inventoryItem)

	// Add item to form
	err = service.AddItem(form.ID, ingredient.ID, 95.0, "Kurang 5kg")
	assert.NoError(t, err)

	// Reject the form
	err = service.RejectForm(form.ID, approver.ID, "Data tidak akurat")
	assert.NoError(t, err)

	// Export form
	exportData, err := service.ExportForm(form.ID, "excel", "Test Exporter")
	assert.NoError(t, err)
	assert.NotNil(t, exportData)
	assert.Greater(t, len(exportData), 0, "Export data should not be empty")
}
