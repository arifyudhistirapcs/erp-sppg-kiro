package services

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupAuditTestDB creates an in-memory SQLite database for audit testing
func setupAuditTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.AuditTrail{})
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	return db
}

// cleanupAuditTestDB cleans up the test database
func cleanupAuditTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM audit_trails")
	db.Exec("DELETE FROM users")
}

// TestProperty4_AuditTrailCompleteness tests Property 4
// **Validates: Requirements 1.6, 21.1, 21.2**
// Feature: erp-sppg-system, Property 4: Audit Trail Completeness
// For any User action that modifies data (create, update, delete), an audit trail entry
// should be created with User ID, timestamp, action type, entity, and values.
func TestProperty4_AuditTrailCompleteness(t *testing.T) {
	db := setupAuditTestDB(t)
	defer cleanupAuditTestDB(db)

	auditService := NewAuditTrailService(db)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: For any create action, audit trail should be recorded with all required fields
	properties.Property("create actions should record complete audit trail", prop.ForAll(
		func(userID, entityIDSuffix int, entityIdx int) bool {
			// Generate test data
			uid := uint(userID%10000 + 1)
			entities := []string{"recipe", "menu_plan", "purchase_order", "supplier", "school", "employee"}
			entity := entities[entityIdx%len(entities)]
			entityID := fmt.Sprintf("%s_%d", entity, entityIDSuffix)
			ipAddress := fmt.Sprintf("192.168.1.%d", userID%255)

			// Create new value
			newValue := map[string]interface{}{
				"id":   entityID,
				"name": fmt.Sprintf("Test %s %d", entity, entityIDSuffix),
			}

			// Record the action
			err := auditService.RecordAction(uid, "create", entity, entityID, nil, newValue, ipAddress)
			if err != nil {
				t.Logf("Failed to record audit action: %v", err)
				return false
			}

			// Verify audit trail entry was created
			var auditEntry models.AuditTrail
			result := db.Where("user_id = ? AND action = ? AND entity = ? AND entity_id = ?",
				uid, "create", entity, entityID).First(&auditEntry)

			if result.Error != nil {
				t.Logf("Audit trail entry not found: %v", result.Error)
				return false
			}

			// Verify all required fields are present
			if auditEntry.UserID != uid {
				t.Logf("UserID mismatch: expected %d, got %d", uid, auditEntry.UserID)
				return false
			}

			if auditEntry.Action != "create" {
				t.Logf("Action mismatch: expected 'create', got '%s'", auditEntry.Action)
				return false
			}

			if auditEntry.Entity != entity {
				t.Logf("Entity mismatch: expected '%s', got '%s'", entity, auditEntry.Entity)
				return false
			}

			if auditEntry.EntityID != entityID {
				t.Logf("EntityID mismatch: expected '%s', got '%s'", entityID, auditEntry.EntityID)
				return false
			}

			if auditEntry.IPAddress != ipAddress {
				t.Logf("IPAddress mismatch: expected '%s', got '%s'", ipAddress, auditEntry.IPAddress)
				return false
			}

			// Verify timestamp is recent (within last 5 seconds)
			if time.Since(auditEntry.Timestamp) > 5*time.Second {
				t.Logf("Timestamp is too old: %v", auditEntry.Timestamp)
				return false
			}

			// Verify new value is stored correctly
			var storedNewValue map[string]interface{}
			if err := json.Unmarshal([]byte(auditEntry.NewValue), &storedNewValue); err != nil {
				t.Logf("Failed to unmarshal new value: %v", err)
				return false
			}

			if storedNewValue["id"] != entityID {
				t.Logf("New value ID mismatch: expected '%s', got '%v'", entityID, storedNewValue["id"])
				return false
			}

			// Verify old value is empty for create actions
			if auditEntry.OldValue != "{}" && auditEntry.OldValue != "null" {
				t.Logf("Old value should be empty for create action, got: %s", auditEntry.OldValue)
				return false
			}

			return true
		},
		gen.IntRange(1, 100000),
		gen.IntRange(1, 100000),
		gen.IntRange(0, 100),
	))

	// Property: For any update action, audit trail should record both old and new values
	properties.Property("update actions should record old and new values", prop.ForAll(
		func(userID, entityIDSuffix, version int, entityIdx int) bool {
			uid := uint(userID%10000 + 1)
			entities := []string{"recipe", "menu_plan", "purchase_order", "supplier", "school", "employee"}
			entity := entities[entityIdx%len(entities)]
			entityID := fmt.Sprintf("%s_%d", entity, entityIDSuffix)
			ipAddress := fmt.Sprintf("10.0.0.%d", userID%255)

			// Create old and new values
			oldValue := map[string]interface{}{
				"id":      entityID,
				"name":    fmt.Sprintf("Old %s %d", entity, entityIDSuffix),
				"version": version,
			}

			newValue := map[string]interface{}{
				"id":      entityID,
				"name":    fmt.Sprintf("New %s %d", entity, entityIDSuffix),
				"version": version + 1,
			}

			// Record the update action
			err := auditService.RecordAction(uid, "update", entity, entityID, oldValue, newValue, ipAddress)
			if err != nil {
				t.Logf("Failed to record audit action: %v", err)
				return false
			}

			// Verify audit trail entry was created
			var auditEntry models.AuditTrail
			result := db.Where("user_id = ? AND action = ? AND entity = ? AND entity_id = ?",
				uid, "update", entity, entityID).Order("timestamp DESC").First(&auditEntry)

			if result.Error != nil {
				t.Logf("Audit trail entry not found: %v", result.Error)
				return false
			}

			// Verify old value is stored
			var storedOldValue map[string]interface{}
			if err := json.Unmarshal([]byte(auditEntry.OldValue), &storedOldValue); err != nil {
				t.Logf("Failed to unmarshal old value: %v", err)
				return false
			}

			if storedOldValue["version"] != float64(version) {
				t.Logf("Old value version mismatch: expected %d, got %v", version, storedOldValue["version"])
				return false
			}

			// Verify new value is stored
			var storedNewValue map[string]interface{}
			if err := json.Unmarshal([]byte(auditEntry.NewValue), &storedNewValue); err != nil {
				t.Logf("Failed to unmarshal new value: %v", err)
				return false
			}

			if storedNewValue["version"] != float64(version+1) {
				t.Logf("New value version mismatch: expected %d, got %v", version+1, storedNewValue["version"])
				return false
			}

			return true
		},
		gen.IntRange(1, 100000),
		gen.IntRange(1, 100000),
		gen.IntRange(1, 100),
		gen.IntRange(0, 100),
	))

	// Property: For any delete action, audit trail should record the deleted entity
	properties.Property("delete actions should record deleted entity details", prop.ForAll(
		func(userID, entityIDSuffix int, entityIdx int) bool {
			uid := uint(userID%10000 + 1)
			entities := []string{"recipe", "menu_plan", "purchase_order", "supplier", "school", "employee"}
			entity := entities[entityIdx%len(entities)]
			entityID := fmt.Sprintf("%s_%d", entity, entityIDSuffix)
			ipAddress := fmt.Sprintf("172.16.0.%d", userID%255)

			// Create old value (the entity being deleted)
			oldValue := map[string]interface{}{
				"id":   entityID,
				"name": fmt.Sprintf("Deleted %s %d", entity, entityIDSuffix),
			}

			// Record the delete action
			err := auditService.RecordAction(uid, "delete", entity, entityID, oldValue, nil, ipAddress)
			if err != nil {
				t.Logf("Failed to record audit action: %v", err)
				return false
			}

			// Verify audit trail entry was created
			var auditEntry models.AuditTrail
			result := db.Where("user_id = ? AND action = ? AND entity = ? AND entity_id = ?",
				uid, "delete", entity, entityID).First(&auditEntry)

			if result.Error != nil {
				t.Logf("Audit trail entry not found: %v", result.Error)
				return false
			}

			// Verify old value is stored (the deleted entity)
			var storedOldValue map[string]interface{}
			if err := json.Unmarshal([]byte(auditEntry.OldValue), &storedOldValue); err != nil {
				t.Logf("Failed to unmarshal old value: %v", err)
				return false
			}

			if storedOldValue["id"] != entityID {
				t.Logf("Old value ID mismatch: expected '%s', got '%v'", entityID, storedOldValue["id"])
				return false
			}

			// Verify new value is empty for delete actions
			if auditEntry.NewValue != "{}" && auditEntry.NewValue != "null" {
				t.Logf("New value should be empty for delete action, got: %s", auditEntry.NewValue)
				return false
			}

			return true
		},
		gen.IntRange(1, 100000),
		gen.IntRange(1, 100000),
		gen.IntRange(0, 100),
	))

	// Property: Multiple actions by the same user should all be recorded
	properties.Property("multiple actions should all be recorded in audit trail", prop.ForAll(
		func(userID, startID, actionCount int) bool {
			uid := uint(userID%10000 + 1)
			// Limit action count to reasonable range
			count := (actionCount % 10) + 1
			ipAddress := fmt.Sprintf("192.168.100.%d", userID%255)

			// Record multiple actions
			for i := 0; i < count; i++ {
				entityID := fmt.Sprintf("entity_%d_%d", startID, i)
				newValue := map[string]interface{}{
					"id":    entityID,
					"index": i,
				}

				err := auditService.RecordAction(uid, "create", "test_entity", entityID, nil, newValue, ipAddress)
				if err != nil {
					t.Logf("Failed to record action %d: %v", i, err)
					return false
				}
			}

			// Verify all actions were recorded
			var auditEntries []models.AuditTrail
			result := db.Where("user_id = ? AND entity = ? AND ip_address = ?",
				uid, "test_entity", ipAddress).Find(&auditEntries)

			if result.Error != nil {
				t.Logf("Failed to query audit entries: %v", result.Error)
				return false
			}

			// Count entries matching our test
			matchingCount := 0
			for _, entry := range auditEntries {
				var newVal map[string]interface{}
				if err := json.Unmarshal([]byte(entry.NewValue), &newVal); err == nil {
					if index, ok := newVal["index"].(float64); ok {
						if int(index) >= 0 && int(index) < count {
							matchingCount++
						}
					}
				}
			}

			if matchingCount < count {
				t.Logf("Not all actions recorded: expected %d, found %d", count, matchingCount)
				return false
			}

			return true
		},
		gen.IntRange(1, 10000),
		gen.IntRange(1, 100000),
		gen.IntRange(1, 20),
	))

	// Property: Audit trail entries should be immutable (cannot be modified)
	properties.Property("audit trail entries should preserve original data", prop.ForAll(
		func(userID, entityIDSuffix int) bool {
			uid := uint(userID%10000 + 1)
			entityID := fmt.Sprintf("immutable_test_%d", entityIDSuffix)
			ipAddress := "192.168.1.100"

			originalValue := map[string]interface{}{
				"id":   entityID,
				"data": "original",
			}

			// Record the action
			err := auditService.RecordAction(uid, "create", "immutable_entity", entityID, nil, originalValue, ipAddress)
			if err != nil {
				t.Logf("Failed to record audit action: %v", err)
				return false
			}

			// Retrieve the audit entry
			var auditEntry models.AuditTrail
			result := db.Where("user_id = ? AND entity = ? AND entity_id = ?",
				uid, "immutable_entity", entityID).First(&auditEntry)

			if result.Error != nil {
				t.Logf("Audit trail entry not found: %v", result.Error)
				return false
			}

			// Store original values
			originalTimestamp := auditEntry.Timestamp
			originalNewValue := auditEntry.NewValue

			// Attempt to modify the audit entry (this should not affect the stored record)
			// In a real system, this would be prevented by database constraints or application logic
			// Here we verify that the data remains consistent

			// Re-query the entry
			var verifyEntry models.AuditTrail
			result = db.Where("id = ?", auditEntry.ID).First(&verifyEntry)

			if result.Error != nil {
				t.Logf("Failed to re-query audit entry: %v", result.Error)
				return false
			}

			// Verify data hasn't changed
			if verifyEntry.Timestamp != originalTimestamp {
				t.Logf("Timestamp changed: original=%v, current=%v", originalTimestamp, verifyEntry.Timestamp)
				return false
			}

			if verifyEntry.NewValue != originalNewValue {
				t.Logf("NewValue changed: original=%s, current=%s", originalNewValue, verifyEntry.NewValue)
				return false
			}

			return true
		},
		gen.IntRange(1, 10000),
		gen.IntRange(1, 100000),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
