package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Server
	Port    string
	GinMode string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// Firebase
	FirebaseCredentialsPath string
	FirebaseDatabaseURL     string
	StorageBucket           string

	// CORS
	AllowedOrigins []string

	// Session
	SessionTimeoutMinutes int

	// Security
	EnableHTTPS           bool
	MaxRequestSize        int64
	EnableRateLimit       bool
	AuthRateLimit         int
	APIRateLimit          int
	RateLimitWindow       int // minutes
	AdminWhitelistIPs     []string
	EnableCSRFProtection  bool

	// Redis Cache
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	EnableCache   bool
}

func Load() *Config {
	jwtExpiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	sessionTimeout, _ := strconv.Atoi(getEnv("SESSION_TIMEOUT_MINUTES", "30"))
	maxRequestSize, _ := strconv.ParseInt(getEnv("MAX_REQUEST_SIZE", "10485760"), 10, 64) // 10MB default
	authRateLimit, _ := strconv.Atoi(getEnv("AUTH_RATE_LIMIT", "5"))
	apiRateLimit, _ := strconv.Atoi(getEnv("API_RATE_LIMIT", "100"))
	rateLimitWindow, _ := strconv.Atoi(getEnv("RATE_LIMIT_WINDOW", "1"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	allowedOrigins := strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174"), ",")
	adminWhitelistIPs := strings.Split(getEnv("ADMIN_WHITELIST_IPS", ""), ",")
	if len(adminWhitelistIPs) == 1 && adminWhitelistIPs[0] == "" {
		adminWhitelistIPs = []string{}
	}

	return &Config{
		Port:                    getEnv("PORT", "8080"),
		GinMode:                 getEnv("GIN_MODE", "debug"),
		DBHost:                  getEnv("DB_HOST", "localhost"),
		DBPort:                  getEnv("DB_PORT", "5432"),
		DBUser:                  getEnv("DB_USER", "postgres"),
		DBPassword:              getEnv("DB_PASSWORD", "postgres"),
		DBName:                  getEnv("DB_NAME", "erp_sppg"),
		DBSSLMode:               getEnv("DB_SSLMODE", "disable"),
		JWTSecret:               getEnv("JWT_SECRET", "change-this-secret"),
		JWTExpiryHours:          jwtExpiryHours,
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "./firebase-credentials.json"),
		FirebaseDatabaseURL:     getEnv("FIREBASE_DATABASE_URL", ""),
		StorageBucket:           getEnv("STORAGE_BUCKET", ""),
		AllowedOrigins:          allowedOrigins,
		SessionTimeoutMinutes:   sessionTimeout,
		EnableHTTPS:             getEnv("ENABLE_HTTPS", "false") == "true",
		MaxRequestSize:          maxRequestSize,
		EnableRateLimit:         getEnv("ENABLE_RATE_LIMIT", "true") == "true",
		AuthRateLimit:           authRateLimit,
		APIRateLimit:            apiRateLimit,
		RateLimitWindow:         rateLimitWindow,
		AdminWhitelistIPs:       adminWhitelistIPs,
		EnableCSRFProtection:    getEnv("ENABLE_CSRF_PROTECTION", "true") == "true",
		RedisHost:               getEnv("REDIS_HOST", "localhost"),
		RedisPort:               getEnv("REDIS_PORT", "6379"),
		RedisPassword:           getEnv("REDIS_PASSWORD", ""),
		RedisDB:                 redisDB,
		EnableCache:             getEnv("ENABLE_CACHE", "true") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
