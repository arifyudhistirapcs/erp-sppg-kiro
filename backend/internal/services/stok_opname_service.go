package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var (
	ErrFormNotFound           = errors.New("form stok opname tidak ditemukan")
	ErrFormNotPending         = errors.New("hanya form dengan status pending yang dapat diubah")
	ErrFormAlreadyProcessed   = errors.New("form ini sudah diproses sebelumnya")
	ErrEmptyForm              = errors.New("form harus memiliki minimal satu item")
	ErrInvalidPhysicalCount   = errors.New("semua item harus memiliki physical count yang valid")
	ErrUnauthorized           = errors.New("anda tidak memiliki akses untuk operasi ini")
	ErrDuplicateIngredient    = errors.New("ingredient sudah ada dalam form ini")
	ErrItemNotFound           = errors.New("item tidak ditemukan")
)

// StokOpnameService handles stok opname business logic
type StokOpnameService interface {
	// Form Management
	CreateForm(userID uint, notes string) (*models.StokOpnameForm, error)
	GetForm(formID uint) (*models.StokOpnameForm, error)
	GetAllForms(filters FormFilters) ([]models.StokOpnameForm, int, error)
	UpdateFormNotes(formID uint, notes string) error
	DeleteForm(formID uint) error

	// Item Management
	AddItem(formID uint, ingredientID uint, physicalCount float64, notes string) error
	UpdateItem(itemID uint, physicalCount float64, notes string) error
	RemoveItem(itemID uint) error

	// Workflow
	SubmitForApproval(formID uint) error
	ApproveForm(formID uint, approverID uint) error
	RejectForm(formID uint, approverID uint, reason string) error

	// Reporting
	ExportForm(formID uint, format string, exporterName string) ([]byte, error)
}

// FormFilters defines filters for querying stok opname forms
type FormFilters struct {
	Status     string
	CreatedBy  *uint
	StartDate  *time.Time
	EndDate    *time.Time
	SearchText string
	Page       int
	PageSize   int
}

// stokOpnameServiceImpl is the concrete implementation of StokOpnameService
type stokOpnameServiceImpl struct {
	db                  *gorm.DB
	inventoryService    *InventoryService
	notificationService *NotificationService
}

// NewStokOpnameService creates a new stok opname service
func NewStokOpnameService(db *gorm.DB, inventoryService *InventoryService, notificationService *NotificationService) StokOpnameService {
	return &stokOpnameServiceImpl{
		db:                  db,
		inventoryService:    inventoryService,
		notificationService: notificationService,
	}
}

// CreateForm creates a new stok opname form
func (s *stokOpnameServiceImpl) CreateForm(userID uint, notes string) (*models.StokOpnameForm, error) {
	// Generate form number
	formNumber, err := s.generateFormNumber()
	if err != nil {
		return nil, err
	}

	// Create form with initial values
	form := &models.StokOpnameForm{
		FormNumber: formNumber,
		CreatedBy:  userID,
		CreatedAt:  time.Now(),
		Status:     "pending",
		Notes:      notes,
		IsProcessed: false,
	}

	// Save to database
	if err := s.db.Create(form).Error; err != nil {
		return nil, err
	}

	return form, nil
}

// generateFormNumber generates a unique form number in format SO-YYYYMMDD-NNNN
func (s *stokOpnameServiceImpl) generateFormNumber() (string, error) {
	// Format: SO-YYYYMMDD-NNNN
	now := time.Now()
	datePrefix := now.Format("20060102")
	
	// Count forms created today
	var count int64
	s.db.Model(&models.StokOpnameForm{}).
		Where("form_number LIKE ?", fmt.Sprintf("SO-%s-%%", datePrefix)).
		Count(&count)
	
	// Generate form number with 4-digit padding
	formNumber := fmt.Sprintf("SO-%s-%04d", datePrefix, count+1)
	
	// Check if it already exists (race condition protection)
	var existing models.StokOpnameForm
	err := s.db.Where("form_number = ?", formNumber).First(&existing).Error
	if err == nil {
		// If exists, try with incremented number
		formNumber = fmt.Sprintf("SO-%s-%04d", datePrefix, count+2)
	}
	
	return formNumber, nil
}

// GetForm retrieves a stok opname form by ID with all relationships
func (s *stokOpnameServiceImpl) GetForm(formID uint) (*models.StokOpnameForm, error) {
	var form models.StokOpnameForm
	
	// Load form with all relationships:
	// - Creator (user who created the form)
	// - Approver (user who approved/rejected the form)
	// - Items (all items in the form)
	// - For each Item, preload Ingredient
	err := s.db.
		Preload("Creator").
		Preload("Approver").
		Preload("Items.Ingredient").
		First(&form, formID).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFormNotFound
		}
		return nil, err
	}
	
	return &form, nil
}

// GetAllForms retrieves all stok opname forms with filtering, pagination, and sorting
func (s *stokOpnameServiceImpl) GetAllForms(filters FormFilters) ([]models.StokOpnameForm, int, error) {
	var forms []models.StokOpnameForm
	var totalCount int64
	
	// Start building query
	query := s.db.Model(&models.StokOpnameForm{})
	
	// Apply status filter
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	
	// Apply creator filter
	if filters.CreatedBy != nil {
		query = query.Where("created_by = ?", *filters.CreatedBy)
	}
	
	// Apply date range filters
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}
	
	// Apply search text filter (search in form notes and creator name)
	if filters.SearchText != "" {
		searchPattern := "%" + filters.SearchText + "%"
		query = query.Where(
			s.db.Where("notes LIKE ?", searchPattern).
				Or("EXISTS (SELECT 1 FROM users WHERE users.id = stok_opname_forms.created_by AND users.full_name LIKE ?)", searchPattern),
		)
	}
	
	// Get total count before pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}
	
	// Set default pagination values
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 20 // default 20 items per page
	}
	
	// Calculate offset
	offset := (page - 1) * pageSize
	
	// Apply sorting (newest first), pagination, and preload relationships
	err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Preload("Creator").
		Preload("Approver").
		Find(&forms).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return forms, int(totalCount), nil
}

// UpdateFormNotes updates the notes of a pending stok opname form
func (s *stokOpnameServiceImpl) UpdateFormNotes(formID uint, notes string) error {
	// Retrieve the form by ID
	var form models.StokOpnameForm
	if err := s.db.First(&form, formID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFormNotFound
		}
		return err
	}
	
	// Check if form status is "pending"
	if form.Status != "pending" {
		return ErrFormNotPending
	}
	
	// Update the notes field
	form.Notes = notes
	
	// Save to database
	if err := s.db.Save(&form).Error; err != nil {
		return err
	}
	
	return nil
}

// DeleteForm deletes a pending stok opname form and all its items
func (s *stokOpnameServiceImpl) DeleteForm(formID uint) error {
	// Retrieve the form by ID
	var form models.StokOpnameForm
	if err := s.db.First(&form, formID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFormNotFound
		}
		return err
	}
	
	// Check if form status is "pending"
	if form.Status != "pending" {
		return ErrFormNotPending
	}
	
	// Delete the form from database
	// GORM will cascade delete all items due to the foreign key relationship
	// We use Select("Items") to explicitly cascade delete the associated items
	if err := s.db.Select("Items").Delete(&form).Error; err != nil {
		return err
	}
	
	return nil
}

// AddItem adds an inventory item to a stok opname form
func (s *stokOpnameServiceImpl) AddItem(formID uint, ingredientID uint, physicalCount float64, notes string) error {
	// 1. Retrieve the form by ID
	var form models.StokOpnameForm
	if err := s.db.First(&form, formID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFormNotFound
		}
		return err
	}
	
	// 2. Check if form status is "pending"
	if form.Status != "pending" {
		return ErrFormNotPending
	}
	
	// 3. Check if ingredient already exists in this form
	var existingItem models.StokOpnameItem
	err := s.db.Where("form_id = ? AND ingredient_id = ?", formID, ingredientID).First(&existingItem).Error
	if err == nil {
		// Item found, this is a duplicate
		return ErrDuplicateIngredient
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other error occurred
		return err
	}
	
	// 4. Fetch the current system stock for the ingredient from inventory_items table
	inventoryItem, err := (*s.inventoryService).GetInventoryItem(ingredientID)
	if err != nil {
		return err
	}
	
	systemStock := inventoryItem.Quantity
	
	// 5. Calculate difference = physical_count - system_stock
	difference := physicalCount - systemStock
	
	// 6. Create a new StokOpnameItem
	item := &models.StokOpnameItem{
		FormID:        formID,
		IngredientID:  ingredientID,
		SystemStock:   systemStock,
		PhysicalCount: physicalCount,
		Difference:    difference,
		ItemNotes:     notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// 7. Save to database
	if err := s.db.Create(item).Error; err != nil {
		return err
	}
	
	return nil
}

// UpdateItem updates the physical count and notes of a stok opname item
func (s *stokOpnameServiceImpl) UpdateItem(itemID uint, physicalCount float64, notes string) error {
	// 1. Retrieve the item by ID with its parent form preloaded
	var item models.StokOpnameItem
	if err := s.db.Preload("Form").First(&item, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrItemNotFound
		}
		return err
	}
	
	// 2. Check if parent form status is "pending"
	if item.Form.Status != "pending" {
		return ErrFormNotPending
	}
	
	// 3. Update physical_count and item_notes
	item.PhysicalCount = physicalCount
	item.ItemNotes = notes
	
	// 4. Recalculate difference = physical_count - system_stock
	item.Difference = physicalCount - item.SystemStock
	
	// 5. Update timestamp and save to database
	item.UpdatedAt = time.Now()
	if err := s.db.Save(&item).Error; err != nil {
		return err
	}
	
	return nil
}

// RemoveItem removes an item from a stok opname form
func (s *stokOpnameServiceImpl) RemoveItem(itemID uint) error {
	// 1. Retrieve the item by ID with its parent form preloaded
	var item models.StokOpnameItem
	if err := s.db.Preload("Form").First(&item, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrItemNotFound
		}
		return err
	}
	
	// 2. Check if parent form status is "pending"
	if item.Form.Status != "pending" {
		return ErrFormNotPending
	}
	
	// 3. Delete the item from database
	if err := s.db.Delete(&item).Error; err != nil {
		return err
	}
	
	return nil
}

// SubmitForApproval submits a stok opname form for approval
func (s *stokOpnameServiceImpl) SubmitForApproval(formID uint) error {
	// Retrieve the form with items preloaded
	var form models.StokOpnameForm
	if err := s.db.Preload("Items").First(&form, formID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrFormNotFound
		}
		return fmt.Errorf("gagal mengambil form: %w", err)
	}

	// Validate form has at least one item
	if len(form.Items) == 0 {
		return ErrEmptyForm
	}

	// Validate all items have valid physical counts (>= 0)
	for _, item := range form.Items {
		if item.PhysicalCount < 0 {
			return ErrInvalidPhysicalCount
		}
	}

	// Status remains "pending" - no status change needed
	// Only approval/rejection actions change the status

	// Send notification to all users with Kepala_SPPG role
	var kepalaUsers []models.User
	if err := s.db.Where("role = ? AND is_active = ?", "kepala_sppg", true).Find(&kepalaUsers).Error; err != nil {
		return fmt.Errorf("gagal mengambil daftar Kepala SPPG: %w", err)
	}

	// Send notification to each Kepala SPPG user
	ctx := context.Background()
	for _, user := range kepalaUsers {
		notification := &models.Notification{
			UserID:  user.ID,
			Type:    "stok_opname_approval",
			Title:   "Permintaan Persetujuan Stok Opname",
			Message: fmt.Sprintf("Form stok opname %s memerlukan persetujuan Anda", form.FormNumber),
			Link:    fmt.Sprintf("/inventory/stok-opname/%d", form.ID),
		}

		if err := s.notificationService.CreateNotification(ctx, notification); err != nil {
			// Log error but don't fail the submission
			fmt.Printf("Peringatan: gagal mengirim notifikasi ke user %d: %v\n", user.ID, err)
		}
	}

	return nil
}

// ApproveForm approves a stok opname form and applies stock adjustments
func (s *stokOpnameServiceImpl) ApproveForm(formID uint, approverID uint) error {
	// 1. Retrieve the form with items and creator preloaded
	var form models.StokOpnameForm
	if err := s.db.Preload("Items").Preload("Creator").First(&form, formID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFormNotFound
		}
		return err
	}

	// 2. Retrieve the approver user by approverID
	var approver models.User
	if err := s.db.First(&approver, approverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUnauthorized
		}
		return err
	}

	// 3. Validate approver has "kepala_sppg" role
	if approver.Role != "kepala_sppg" {
		return ErrUnauthorized
	}

	// 4. Check if form is already processed
	if form.IsProcessed {
		return ErrFormAlreadyProcessed
	}

	// 5. Begin database transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 6. Update form: status="approved", approved_by=approverID, approved_at=now, is_processed=true
	now := time.Now()
	if err := tx.Model(&models.StokOpnameForm{}).Where("id = ?", formID).Updates(map[string]interface{}{
		"status":       "approved",
		"approved_by":  approverID,
		"approved_at":  now,
		"is_processed": true,
		"updated_at":   now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 7. For each item: update inventory stock and create inventory movement
	for _, item := range form.Items {
		// Calculate absolute value of difference for movement quantity
		var quantity float64

		if item.Difference > 0 {
			// Physical count is higher than system stock
			quantity = item.Difference
		} else if item.Difference < 0 {
			// Physical count is lower than system stock
			quantity = -item.Difference // absolute value
		} else {
			// No difference, skip this item
			continue
		}

		// Get current inventory item to know current stock
		var inventoryItem models.InventoryItem
		err := tx.Where("ingredient_id = ?", item.IngredientID).First(&inventoryItem).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return err
		}

		// Create inventory movement with type "adjustment"
		reference := fmt.Sprintf("Stok Opname: %s", form.FormNumber)
		notes := fmt.Sprintf("Penyesuaian stok dari %s ke %s. Selisih: %s",
			fmt.Sprintf("%.2f", item.SystemStock),
			fmt.Sprintf("%.2f", item.PhysicalCount),
			fmt.Sprintf("%.2f", item.Difference))

		if item.ItemNotes != "" {
			notes += fmt.Sprintf(". Catatan: %s", item.ItemNotes)
		}

		// Update stock directly to physical count
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new inventory item
			newItem := models.InventoryItem{
				IngredientID: item.IngredientID,
				Quantity:     item.PhysicalCount,
				MinThreshold: 10, // default threshold
				LastUpdated:  now,
			}
			if err := tx.Create(&newItem).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// Update existing inventory item
			if err := tx.Model(&models.InventoryItem{}).Where("id = ?", inventoryItem.ID).Updates(map[string]interface{}{
				"quantity":     item.PhysicalCount,
				"last_updated": now,
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		// Create inventory movement record
		movement := models.InventoryMovement{
			IngredientID: item.IngredientID,
			MovementType: "adjustment",
			Quantity:     quantity, // absolute value of difference
			Reference:    reference,
			MovementDate: now,
			CreatedBy:    approverID,
			Notes:        notes,
		}
		if err := tx.Create(&movement).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 9. Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// RejectForm rejects a stok opname form with a reason
func (s *stokOpnameServiceImpl) RejectForm(formID uint, approverID uint, reason string) error {
	// 1. Retrieve the form with creator preloaded
	var form models.StokOpnameForm
	if err := s.db.Preload("Creator").First(&form, formID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFormNotFound
		}
		return err
	}

	// 2. Retrieve the approver user by approverID
	var approver models.User
	if err := s.db.First(&approver, approverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUnauthorized
		}
		return err
	}

	// 3. Validate approver has "kepala_sppg" role
	if approver.Role != "kepala_sppg" {
		return ErrUnauthorized
	}

	// 4. Validate approver is different from creator
	if approverID == form.CreatedBy {
		return ErrUnauthorized
	}

	// 5. Update form: status="rejected", approved_by=approverID, approved_at=now, rejection_reason=reason
	now := time.Now()
	if err := s.db.Model(&models.StokOpnameForm{}).Where("id = ?", formID).Updates(map[string]interface{}{
		"status":           "rejected",
		"approved_by":      approverID,
		"approved_at":      now,
		"rejection_reason": reason,
		"updated_at":       now,
	}).Error; err != nil {
		return err
	}

	// 6. Return nil on success
	return nil
}

// ExportForm exports a stok opname form to Excel or PDF format
func (s *stokOpnameServiceImpl) ExportForm(formID uint, format string, exporterName string) ([]byte, error) {
	// Validate format
	if format != "excel" && format != "pdf" {
		return nil, fmt.Errorf("format tidak valid: gunakan 'excel' atau 'pdf'")
	}

	// Retrieve form with all relationships
	form, err := s.GetForm(formID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil form: %w", err)
	}

	// Prepare export data
	var buffer *bytes.Buffer
	if format == "excel" {
		buffer, err = s.generateExcelExport(form, exporterName)
	} else {
		buffer, err = s.generatePDFExport(form, exporterName)
	}

	if err != nil {
		return nil, fmt.Errorf("gagal membuat file export: %w", err)
	}

	return buffer.Bytes(), nil
}

func (s *stokOpnameServiceImpl) generateExcelExport(form *models.StokOpnameForm, exporterName string) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Stok Opname"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 20)
	f.SetColWidth(sheetName, "B", "B", 25)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 15)
	f.SetColWidth(sheetName, "F", "F", 30)

	currentRow := 1

	// Title style
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})

	// Header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})

	// Table header style
	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Border:    []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	// Data style
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Title
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "LAPORAN STOK OPNAME")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), titleStyle)
	currentRow += 2

	// Form header information
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Nomor Form:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), form.FormNumber)
	currentRow++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Tanggal:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), form.CreatedAt.Format("02/01/2006 15:04"))
	currentRow++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Dibuat Oleh:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
	creatorName := "Unknown"
	if form.Creator.ID != 0 {
		creatorName = form.Creator.FullName
	}
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), creatorName)
	currentRow++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Status:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), form.Status)
	currentRow++

	// Approval information if approved or rejected
	if form.Status == "approved" || form.Status == "rejected" {
		approverName := "Unknown"
		if form.Approver != nil && form.Approver.ID != 0 {
			approverName = form.Approver.FullName
		}
		
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Disetujui/Ditolak Oleh:")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), approverName)
		currentRow++

		if form.ApprovedAt != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Tanggal Persetujuan:")
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), form.ApprovedAt.Format("02/01/2006 15:04"))
			currentRow++
		}
	}

	// Rejection reason if rejected
	if form.Status == "rejected" && form.RejectionReason != "" {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Alasan Penolakan:")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), form.RejectionReason)
		currentRow++
	}

	// Form notes if present
	if form.Notes != "" {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Catatan:")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), form.Notes)
		currentRow++
	}

	currentRow += 2 // Add spacing

	// Items table headers
	headers := []string{"Nama Bahan", "Stok Sistem", "Jumlah Fisik", "Selisih", "Catatan"}
	for i, header := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s%d", col, currentRow)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, tableHeaderStyle)
	}
	currentRow++

	// Items data
	for _, item := range form.Items {
		ingredientName := "Unknown"
		if item.Ingredient.ID != 0 {
			ingredientName = item.Ingredient.Name
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), ingredientName)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), item.SystemStock)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), item.PhysicalCount)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), item.Difference)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), item.ItemNotes)

		// Apply data style to all cells in the row
		for i := 0; i < len(headers); i++ {
			col, _ := excelize.ColumnNumberToName(i + 1)
			cell := fmt.Sprintf("%s%d", col, currentRow)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}

		currentRow++
	}

	currentRow += 2 // Add spacing

	// Export metadata
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Diekspor Oleh:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), exporterName)
	currentRow++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), "Tanggal Ekspor:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), headerStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), time.Now().Format("02/01/2006 15:04"))

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (s *stokOpnameServiceImpl) generatePDFExport(form *models.StokOpnameForm, exporterName string) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 10)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "LAPORAN STOK OPNAME")
	pdf.Ln(12)

	// Form header information
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(50, 6, "Nomor Form:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, form.FormNumber)
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(50, 6, "Tanggal:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, form.CreatedAt.Format("02/01/2006 15:04"))
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(50, 6, "Dibuat Oleh:")
	pdf.SetFont("Arial", "", 10)
	creatorName := "Unknown"
	if form.Creator.ID != 0 {
		creatorName = form.Creator.FullName
	}
	pdf.Cell(0, 6, creatorName)
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(50, 6, "Status:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, form.Status)
	pdf.Ln(6)

	// Approval information if approved or rejected
	if form.Status == "approved" || form.Status == "rejected" {
		approverName := "Unknown"
		if form.Approver != nil && form.Approver.ID != 0 {
			approverName = form.Approver.FullName
		}

		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(50, 6, "Disetujui/Ditolak Oleh:")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 6, approverName)
		pdf.Ln(6)

		if form.ApprovedAt != nil {
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(50, 6, "Tanggal Persetujuan:")
			pdf.SetFont("Arial", "", 10)
			pdf.Cell(0, 6, form.ApprovedAt.Format("02/01/2006 15:04"))
			pdf.Ln(6)
		}
	}

	// Rejection reason if rejected
	if form.Status == "rejected" && form.RejectionReason != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(50, 6, "Alasan Penolakan:")
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 6, form.RejectionReason, "", "", false)
		pdf.Ln(2)
	}

	// Form notes if present
	if form.Notes != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(50, 6, "Catatan:")
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 6, form.Notes, "", "", false)
		pdf.Ln(2)
	}

	pdf.Ln(6)

	// Items table
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 200, 200)
	
	// Table headers
	pdf.CellFormat(70, 7, "Nama Bahan", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Stok Sistem", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Jumlah Fisik", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Selisih", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 7, "Catatan", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)

	for i, item := range form.Items {
		// Alternate row colors
		if i%2 == 0 {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		ingredientName := "Unknown"
		if item.Ingredient.ID != 0 {
			ingredientName = item.Ingredient.Name
		}

		pdf.CellFormat(70, 6, ingredientName, "1", 0, "L", true, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%.2f", item.SystemStock), "1", 0, "R", true, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%.2f", item.PhysicalCount), "1", 0, "R", true, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%.2f", item.Difference), "1", 0, "R", true, 0, "")
		pdf.CellFormat(45, 6, item.ItemNotes, "1", 0, "L", true, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(10)

	// Export metadata
	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(0, 5, fmt.Sprintf("Diekspor oleh: %s | Tanggal: %s", 
		exporterName, 
		time.Now().Format("02/01/2006 15:04")))

	// Write to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}
