package utils

import (
	"errors"
	"html"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail      = errors.New("format email tidak valid")
	ErrInvalidPhone      = errors.New("format nomor telepon tidak valid")
	ErrInvalidNIK        = errors.New("format NIK tidak valid")
	ErrInvalidGPS        = errors.New("koordinat GPS tidak valid")
	ErrInvalidInput      = errors.New("input tidak valid")
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Phone validation regex (Indonesian format)
// Supports: 08xx, +628xx, 628xx, 021xxx (landline)
var phoneRegex = regexp.MustCompile(`^(\+62|62|0)[0-9]{8,13}$`)

// NIK validation regex (16 digits)
var nikRegex = regexp.MustCompile(`^[0-9]{16}$`)

// SQL injection patterns to detect
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|<script)`),
	regexp.MustCompile(`(?i)(--|;|\/\*|\*\/|xp_|sp_)`),
	regexp.MustCompile(`(?i)(\bor\b|\band\b).*=.*`),
}

// XSS patterns to detect
var xssPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
	regexp.MustCompile(`(?i)javascript:`),
	regexp.MustCompile(`(?i)on\w+\s*=`), // onclick, onload, etc.
	regexp.MustCompile(`(?i)<iframe[^>]*>`),
	regexp.MustCompile(`(?i)<object[^>]*>`),
	regexp.MustCompile(`(?i)<embed[^>]*>`),
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrInvalidEmail
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePhone validates Indonesian phone number format
func ValidatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	// Remove spaces and dashes
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	
	if phone == "" {
		return ErrInvalidPhone
	}
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}
	return nil
}

// NormalizePhone normalizes phone number to standard format (62xxx)
func NormalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	
	// Convert to 62xxx format
	if strings.HasPrefix(phone, "0") {
		phone = "62" + phone[1:]
	} else if strings.HasPrefix(phone, "+62") {
		phone = phone[1:]
	}
	
	return phone
}

// ValidateNIK validates Indonesian NIK (16 digits)
func ValidateNIK(nik string) error {
	nik = strings.TrimSpace(nik)
	if nik == "" {
		return ErrInvalidNIK
	}
	if !nikRegex.MatchString(nik) {
		return ErrInvalidNIK
	}
	return nil
}

// ValidateGPSCoordinates validates latitude and longitude
func ValidateGPSCoordinates(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 {
		return ErrInvalidGPS
	}
	if longitude < -180 || longitude > 180 {
		return ErrInvalidGPS
	}
	return nil
}

// SanitizeInput sanitizes user input to prevent SQL injection and XSS
func SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	// HTML escape to prevent XSS
	input = html.EscapeString(input)
	
	return input
}

// SanitizeHTML removes potentially dangerous HTML tags and attributes
func SanitizeHTML(input string) string {
	// Remove script tags
	input = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`).ReplaceAllString(input, "")
	
	// Remove iframe tags
	input = regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`).ReplaceAllString(input, "")
	
	// Remove object and embed tags
	input = regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`).ReplaceAllString(input, "")
	input = regexp.MustCompile(`(?i)<embed[^>]*>.*?</embed>`).ReplaceAllString(input, "")
	
	// Remove event handlers (onclick, onload, etc.)
	input = regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`).ReplaceAllString(input, "")
	input = regexp.MustCompile(`(?i)\s*on\w+\s*=\s*[^\s>]*`).ReplaceAllString(input, "")
	
	// Remove javascript: protocol
	input = regexp.MustCompile(`(?i)javascript:`).ReplaceAllString(input, "")
	
	return input
}

// DetectSQLInjection checks if input contains SQL injection patterns
func DetectSQLInjection(input string) bool {
	input = strings.ToLower(input)
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// DetectXSS checks if input contains XSS patterns
func DetectXSS(input string) bool {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// ValidateAndSanitize validates and sanitizes input, returns error if malicious content detected
func ValidateAndSanitize(input string) (string, error) {
	// Check for SQL injection
	if DetectSQLInjection(input) {
		return "", errors.New("input mengandung pola SQL injection yang tidak diizinkan")
	}
	
	// Check for XSS
	if DetectXSS(input) {
		return "", errors.New("input mengandung pola XSS yang tidak diizinkan")
	}
	
	// Sanitize the input
	sanitized := SanitizeInput(input)
	
	return sanitized, nil
}

// IsValidString checks if a string is not empty after trimming
func IsValidString(s string) bool {
	return strings.TrimSpace(s) != ""
}

// IsValidPositiveNumber checks if a number is positive
func IsValidPositiveNumber(n float64) bool {
	return n > 0
}

// IsValidNonNegativeNumber checks if a number is non-negative
func IsValidNonNegativeNumber(n float64) bool {
	return n >= 0
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}

// ValidateRequired checks if required fields are present
func ValidateRequired(fields map[string]string) []string {
	var missing []string
	for fieldName, fieldValue := range fields {
		if !IsValidString(fieldValue) {
			missing = append(missing, fieldName)
		}
	}
	return missing
}
