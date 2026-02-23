package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID       int    `json:"id"`
	NIK      string `json:"nik"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

func main() {
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Demo login endpoint
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		log.Printf("Login attempt: identifier=%s, password=%s", req.Identifier, req.Password)

		// Demo users
		demoUsers := map[string]User{
			"kepala.sppg@sppg.com": {ID: 1, NIK: "12345678901", Email: "kepala.sppg@sppg.com", FullName: "Kepala SPPG", Role: "kepala_sppg"},
			"12345678901":          {ID: 1, NIK: "12345678901", Email: "kepala.sppg@sppg.com", FullName: "Kepala SPPG", Role: "kepala_sppg"},
			"ahli.gizi@sppg.com":   {ID: 2, NIK: "12345678902", Email: "ahli.gizi@sppg.com", FullName: "Ahli Gizi", Role: "ahli_gizi"},
			"12345678902":          {ID: 2, NIK: "12345678902", Email: "ahli.gizi@sppg.com", FullName: "Ahli Gizi", Role: "ahli_gizi"},
			"chef@sppg.com":        {ID: 3, NIK: "12345678903", Email: "chef@sppg.com", FullName: "Chef Utama", Role: "chef"},
			"12345678903":          {ID: 3, NIK: "12345678903", Email: "chef@sppg.com", FullName: "Chef Utama", Role: "chef"},
		}

		user, exists := demoUsers[req.Identifier]
		if !exists || req.Password != "admin123" {
			log.Printf("Login failed: user not found or wrong password")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "NIK/Email atau password salah"})
			return
		}

		// Generate demo token
		token := "demo-jwt-token-" + user.Role + "-" + time.Now().Format("20060102150405")

		log.Printf("Login successful: user=%s, role=%s", user.FullName, user.Role)

		c.JSON(http.StatusOK, LoginResponse{
			Token: token,
			User:  user,
		})
	})

	// Demo me endpoint
	r.GET("/api/v1/auth/me", func(c *gin.Context) {
		// For demo, just return a default user
		c.JSON(http.StatusOK, User{
			ID: 1, NIK: "12345678901", Email: "kepala.sppg@sppg.com", 
			FullName: "Kepala SPPG", Role: "kepala_sppg",
		})
	})

	log.Println("Demo server starting on port 8080")
	log.Fatal(r.Run(":8080"))
}
