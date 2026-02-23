package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "user@example.com", false},
		{"Valid email with subdomain", "user@mail.example.com", false},
		{"Valid email with plus", "user+tag@example.com", false},
		{"Invalid email - no @", "userexample.com", true},
		{"Invalid email - no domain", "user@", true},
		{"Invalid email - no TLD", "user@example", true},
		{"Empty email", "", true},
		{"Invalid email - spaces", "user @example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"Valid phone - 08xx", "081234567890", false},
		{"Valid phone - +62", "+6281234567890", false},
		{"Valid phone - 62", "6281234567890", false},
		{"Valid phone - landline", "02112345678", false},
		{"Valid phone with spaces", "0812 3456 7890", false},
		{"Valid phone with dashes", "0812-3456-7890", false},
		{"Invalid phone - too short", "0812345", true},
		{"Invalid phone - letters", "081234abcd", true},
		{"Empty phone", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePhone(tt.phone)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{"Normalize 08xx", "081234567890", "6281234567890"},
		{"Normalize +62", "+6281234567890", "6281234567890"},
		{"Already normalized", "6281234567890", "6281234567890"},
		{"With spaces", "0812 3456 7890", "6281234567890"},
		{"With dashes", "0812-3456-7890", "6281234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePhone(tt.phone)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateNIK(t *testing.T) {
	tests := []struct {
		name    string
		nik     string
		wantErr bool
	}{
		{"Valid NIK", "1234567890123456", false},
		{"Invalid NIK - too short", "123456789012345", true},
		{"Invalid NIK - too long", "12345678901234567", true},
		{"Invalid NIK - letters", "123456789012345a", true},
		{"Empty NIK", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNIK(tt.nik)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGPSCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		wantErr   bool
	}{
		{"Valid coordinates", -6.2088, 106.8456, false},
		{"Valid coordinates - North", 40.7128, -74.0060, false},
		{"Valid coordinates - boundary", 90.0, 180.0, false},
		{"Valid coordinates - boundary negative", -90.0, -180.0, false},
		{"Invalid latitude - too high", 91.0, 106.8456, true},
		{"Invalid latitude - too low", -91.0, 106.8456, true},
		{"Invalid longitude - too high", -6.2088, 181.0, true},
		{"Invalid longitude - too low", -6.2088, -181.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGPSCoordinates(tt.latitude, tt.longitude)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal text", "Hello World", "Hello World"},
		{"HTML tags", "<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"Special chars", "Test & <test>", "Test &amp; &lt;test&gt;"},
		{"Whitespace", "  test  ", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectSQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Normal text", "Hello World", false},
		{"SQL SELECT", "SELECT * FROM users", true},
		{"SQL UNION", "1' UNION SELECT * FROM users--", true},
		{"SQL DROP", "DROP TABLE users", true},
		{"SQL comment", "test--comment", true},
		{"OR equals", "1=1 OR 1=1", true},
		{"Safe input", "user@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectSQLInjection(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectXSS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Normal text", "Hello World", false},
		{"Script tag", "<script>alert('xss')</script>", true},
		{"JavaScript protocol", "javascript:alert('xss')", true},
		{"Event handler", "<img src=x onerror=alert('xss')>", true},
		{"Iframe tag", "<iframe src='evil.com'></iframe>", true},
		{"Safe HTML", "<p>Hello</p>", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectXSS(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateAndSanitize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Normal text", "Hello World", false},
		{"SQL injection attempt", "SELECT * FROM users", true},
		{"XSS attempt", "<script>alert('xss')</script>", true},
		{"Safe input with special chars", "Test & Co.", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateAndSanitize(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestIsValidString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid string", "test", true},
		{"Empty string", "", false},
		{"Whitespace only", "   ", false},
		{"String with content", "  test  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidPositiveNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected bool
	}{
		{"Positive number", 10.5, true},
		{"Zero", 0, false},
		{"Negative number", -5.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPositiveNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]string
		expected []string
	}{
		{
			name: "All fields present",
			fields: map[string]string{
				"name":  "John",
				"email": "john@example.com",
			},
			expected: []string{},
		},
		{
			name: "Missing field",
			fields: map[string]string{
				"name":  "John",
				"email": "",
			},
			expected: []string{"email"},
		},
		{
			name: "Multiple missing fields",
			fields: map[string]string{
				"name":  "",
				"email": "",
			},
			expected: []string{"name", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateRequired(tt.fields)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}
