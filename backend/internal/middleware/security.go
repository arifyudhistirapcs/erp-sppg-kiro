package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/erp-sppg/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Enforce HTTPS (only in production)
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.gstatic.com https://apis.google.com; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self' https://api.firebase.com https://*.firebaseio.com wss://*.firebaseio.com; " +
			"frame-src 'none'; " +
			"object-src 'none'"
		c.Header("Content-Security-Policy", csp)
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions Policy (formerly Feature Policy)
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(self), payment=()")

		c.Next()
	}
}

// InputSanitization sanitizes all input data to prevent injection attacks
func InputSanitization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for certain content types (file uploads, etc.)
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "multipart/form-data") ||
			strings.Contains(contentType, "application/octet-stream") {
			c.Next()
			return
		}

		// Get the raw body for JSON requests
		if strings.Contains(contentType, "application/json") {
			// For JSON requests, we'll validate in the handlers
			// since we need to parse the JSON first
			c.Next()
			return
		}

		// Sanitize form data
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.Form {
					for i, value := range values {
						// Detect and reject malicious input
						if utils.DetectSQLInjection(value) || utils.DetectXSS(value) {
							c.JSON(http.StatusBadRequest, gin.H{
								"success":    false,
								"error_code": "MALICIOUS_INPUT_DETECTED",
								"message":    "Input mengandung konten berbahaya yang tidak diizinkan.",
								"field":      key,
							})
							c.Abort()
							return
						}

						// Sanitize the value
						sanitized := utils.SanitizeInput(value)
						c.Request.Form[key][i] = sanitized
					}
				}
			}
		}

		// Sanitize query parameters
		query := c.Request.URL.Query()
		for key, values := range query {
			for i, value := range values {
				// Detect and reject malicious input
				if utils.DetectSQLInjection(value) || utils.DetectXSS(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"success":    false,
						"error_code": "MALICIOUS_INPUT_DETECTED",
						"message":    "Parameter query mengandung konten berbahaya yang tidak diizinkan.",
						"field":      key,
					})
					c.Abort()
					return
				}

				// Sanitize the value
				sanitized := utils.SanitizeInput(value)
				query[key][i] = sanitized
			}
		}
		c.Request.URL.RawQuery = query.Encode()

		c.Next()
	}
}

// HTTPSRedirect redirects HTTP requests to HTTPS in production
func HTTPSRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only redirect in production mode
		if gin.Mode() != gin.ReleaseMode {
			c.Next()
			return
		}

		// Check if request is already HTTPS
		if c.Request.Header.Get("X-Forwarded-Proto") == "https" ||
			c.Request.TLS != nil ||
			c.Request.Header.Get("X-Forwarded-Ssl") == "on" {
			c.Next()
			return
		}

		// Redirect to HTTPS
		httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
		c.Redirect(http.StatusMovedPermanently, httpsURL)
		c.Abort()
	}
}

// RequestSizeLimit limits the size of request bodies to prevent DoS attacks
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET requests
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Check Content-Length header
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success":    false,
				"error_code": "REQUEST_TOO_LARGE",
				"message":    "Ukuran permintaan terlalu besar.",
			})
			c.Abort()
			return
		}

		// Limit the request body reader
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}

// IPWhitelist restricts access to specific IP addresses (for admin endpoints)
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	allowedIPMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		allowedIPMap[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if IP is in whitelist
		if !allowedIPMap[clientIP] {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "IP_NOT_ALLOWED",
				"message":    "Akses ditolak dari alamat IP ini.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserAgentValidation validates User-Agent header to block suspicious requests
func UserAgentValidation() gin.HandlerFunc {
	// Common bot/scanner user agents to block
	blockedUserAgents := []string{
		"sqlmap",
		"nikto",
		"nmap",
		"masscan",
		"nessus",
		"openvas",
		"w3af",
		"skipfish",
		"burp",
		"owasp",
	}

	return func(c *gin.Context) {
		userAgent := strings.ToLower(c.GetHeader("User-Agent"))

		// Block empty user agents
		if userAgent == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_USER_AGENT",
				"message":    "User-Agent header diperlukan.",
			})
			c.Abort()
			return
		}

		// Check against blocked user agents
		for _, blocked := range blockedUserAgents {
			if strings.Contains(userAgent, blocked) {
				c.JSON(http.StatusForbidden, gin.H{
					"success":    false,
					"error_code": "BLOCKED_USER_AGENT",
					"message":    "User-Agent tidak diizinkan.",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequestLogging logs all requests for security monitoring
func RequestLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format for security monitoring
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.ClientIP,
			param.Method,
			param.StatusCode,
			param.Path,
			param.Latency,
		)
	})
}