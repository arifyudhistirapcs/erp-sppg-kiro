package services

import (
	"testing"
)

// TestKDSErrorCreation tests that KDSError can be created and formatted correctly
func TestKDSErrorCreation(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		message        string
		details        string
		expectedError  string
	}{
		{
			name:          "Error with details",
			code:          ErrCodeInsufficientStock,
			message:       "Stok tidak mencukupi",
			details:       "Nasi (butuh 50.00, tersedia 20.00)",
			expectedError: "Stok tidak mencukupi: Nasi (butuh 50.00, tersedia 20.00)",
		},
		{
			name:          "Error without details",
			code:          ErrCodeInvalidRecipe,
			message:       "Resep tidak valid",
			details:       "",
			expectedError: "Resep tidak valid",
		},
		{
			name:          "Inventory not found error",
			code:          ErrCodeInventoryNotFound,
			message:       "Data inventori tidak ditemukan",
			details:       "Komponen bahan dengan ID 123 tidak ditemukan",
			expectedError: "Data inventori tidak ditemukan: Komponen bahan dengan ID 123 tidak ditemukan",
		},
		{
			name:          "Transaction failed error",
			code:          ErrCodeTransactionFailed,
			message:       "Gagal memperbarui stok",
			details:       "Database connection lost",
			expectedError: "Gagal memperbarui stok: Database connection lost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewKDSError(tt.code, tt.message, tt.details)
			
			if err.Code != tt.code {
				t.Errorf("Expected code %s, got %s", tt.code, err.Code)
			}
			
			if err.Message != tt.message {
				t.Errorf("Expected message %s, got %s", tt.message, err.Message)
			}
			
			if err.Details != tt.details {
				t.Errorf("Expected details %s, got %s", tt.details, err.Details)
			}
			
			if err.Error() != tt.expectedError {
				t.Errorf("Expected error string %s, got %s", tt.expectedError, err.Error())
			}
		})
	}
}

// TestErrorCodeConstants tests that error code constants are defined correctly
func TestErrorCodeConstants(t *testing.T) {
	expectedCodes := map[string]string{
		"INSUFFICIENT_STOCK":   ErrCodeInsufficientStock,
		"INVENTORY_NOT_FOUND":  ErrCodeInventoryNotFound,
		"TRANSACTION_FAILED":   ErrCodeTransactionFailed,
		"INVALID_RECIPE":       ErrCodeInvalidRecipe,
	}

	for expected, actual := range expectedCodes {
		if actual != expected {
			t.Errorf("Expected error code %s, got %s", expected, actual)
		}
	}
}
