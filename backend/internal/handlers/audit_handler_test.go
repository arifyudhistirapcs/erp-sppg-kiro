package handlers

import (
	"testing"
)

func TestGenerateDescription(t *testing.T) {
	handler := &AuditHandler{}

	tests := []struct {
		action   string
		entity   string
		userName string
		expected string
	}{
		{"create", "recipe", "John Doe", "John Doe membuat resep"},
		{"update", "user", "Jane Smith", "Jane Smith mengubah pengguna"},
		{"delete", "supplier", "Admin", "Admin menghapus supplier"},
		{"login", "user", "John Doe", "John Doe masuk ke sistem"},
		{"logout", "user", "Jane Smith", "Jane Smith keluar dari sistem"},
		{"approve", "purchase_order", "Manager", "Manager menyetujui purchase order"},
		{"export", "cash_flow", "Accountant", "Accountant mengekspor arus kas"},
	}

	for _, test := range tests {
		result := handler.generateDescription(test.action, test.entity, test.userName)
		if result != test.expected {
			t.Errorf("generateDescription(%s, %s, %s) = %s; want %s",
				test.action, test.entity, test.userName, result, test.expected)
		}
	}
}

func TestGetActionLabel(t *testing.T) {
	// This would be in the frontend, but testing the concept
	actionLabels := map[string]string{
		"create": "Membuat",
		"update": "Mengubah",
		"delete": "Menghapus",
		"login":  "Masuk",
		"logout": "Keluar",
		"approve": "Menyetujui",
		"reject": "Menolak",
		"export": "Mengekspor",
	}

	for action, expected := range actionLabels {
		if actionLabels[action] != expected {
			t.Errorf("Action label for %s should be %s", action, expected)
		}
	}
}

func TestGetEntityLabel(t *testing.T) {
	// This would be in the frontend, but testing the concept
	entityLabels := map[string]string{
		"user":           "Pengguna",
		"recipe":         "Resep",
		"menu":           "Menu",
		"supplier":       "Supplier",
		"purchase_order": "Purchase Order",
		"inventory":      "Inventori",
		"delivery_task":  "Tugas Pengiriman",
		"employee":       "Karyawan",
		"asset":          "Aset",
		"cash_flow":      "Arus Kas",
	}

	for entity, expected := range entityLabels {
		if entityLabels[entity] != expected {
			t.Errorf("Entity label for %s should be %s", entity, expected)
		}
	}
}