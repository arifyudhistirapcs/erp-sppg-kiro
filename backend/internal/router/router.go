package router

import (
	"log"

	firebase "firebase.google.com/go/v4"
	"github.com/erp-sppg/backend/internal/cache"
	"github.com/erp-sppg/backend/internal/config"
	"github.com/erp-sppg/backend/internal/handlers"
	"github.com/erp-sppg/backend/internal/middleware"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, firebaseApp *firebase.App, cfg *config.Config, cacheService *cache.CacheService) *gin.Engine {
	r := gin.Default()

	// Security middleware (applied to all routes)
	if cfg.EnableHTTPS {
		r.Use(middleware.HTTPSRedirect())
	}
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.UserAgentValidation())
	r.Use(middleware.RequestSizeLimit(cfg.MaxRequestSize))
	r.Use(middleware.InputSanitization())

	// Rate limiting middleware
	if cfg.EnableRateLimit {
		r.Use(middleware.APIRateLimitMiddleware())
	}

	// CORS middleware
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "ERP SPPG Backend is running",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// CSRF token endpoint (public)
		if cfg.EnableCSRFProtection {
			v1.GET("/csrf-token", middleware.CSRFTokenHandler())
		}

		// Auth routes (public, with stricter rate limiting)
		authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
		auth := v1.Group("/auth")
		if cfg.EnableRateLimit {
			auth.Use(middleware.AuthRateLimitMiddleware())
		}
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes (require JWT authentication)
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(cfg.JWTSecret))
		protected.Use(middleware.SessionTimeoutMiddleware(cfg.SessionTimeoutMinutes))
		protected.Use(middleware.AuditTrail(db))
		if cfg.EnableCSRFProtection {
			protected.Use(middleware.CSRFMiddleware())
		}
		// Apply cache invalidation middleware for data modifications
		if cacheService != nil {
			protected.Use(middleware.CacheInvalidationMiddleware(cacheService))
		}
		{
			// Auth protected routes
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.GetMe)

			// Recipe routes
			recipeHandler := handlers.NewRecipeHandler(db)
			recipes := protected.Group("/recipes")
			// Apply caching for recipe GET requests
			if cacheService != nil {
				recipes.Use(middleware.ConditionalCacheMiddleware(cacheService, middleware.CacheForReadOnlyOperations, cache.LongCacheDuration))
			}
			{
				recipes.GET("", recipeHandler.GetAllRecipes)
				recipes.POST("", recipeHandler.CreateRecipe)
				recipes.GET("/:id", recipeHandler.GetRecipe)
				recipes.PUT("/:id", recipeHandler.UpdateRecipe)
				recipes.DELETE("/:id", recipeHandler.DeleteRecipe)
				recipes.GET("/:id/nutrition", recipeHandler.GetRecipeNutrition)
				recipes.GET("/:id/history", recipeHandler.GetRecipeHistory)
			}

			// Ingredient routes
			ingredients := protected.Group("/ingredients")
			if cacheService != nil {
				ingredients.Use(middleware.ConditionalCacheMiddleware(cacheService, middleware.CacheForReadOnlyOperations, cache.LongCacheDuration))
			}
			{
				ingredients.GET("", recipeHandler.GetAllIngredients)
				ingredients.POST("", recipeHandler.CreateIngredient)
			}

			// Semi-Finished Goods routes
			semiFinishedHandler := handlers.NewSemiFinishedHandler(db)
			semiFinished := protected.Group("/semi-finished")
			{
				semiFinished.GET("", semiFinishedHandler.GetAllSemiFinishedGoods)
				semiFinished.POST("", semiFinishedHandler.CreateSemiFinishedGoods)
				semiFinished.GET("/:id", semiFinishedHandler.GetSemiFinishedGoods)
				semiFinished.PUT("/:id", semiFinishedHandler.UpdateSemiFinishedGoods)
				semiFinished.DELETE("/:id", semiFinishedHandler.DeleteSemiFinishedGoods)
				semiFinished.POST("/:id/produce", semiFinishedHandler.ProduceSemiFinishedGoods)
				semiFinished.GET("/inventory", semiFinishedHandler.GetSemiFinishedInventory)
			}

			// Menu Planning routes
			menuPlanningHandler := handlers.NewMenuPlanningHandler(db)
			menuPlans := protected.Group("/menu-plans")
			{
				menuPlans.GET("", menuPlanningHandler.GetAllMenuPlans)
				menuPlans.POST("", menuPlanningHandler.CreateMenuPlan)
				menuPlans.GET("/current-week", menuPlanningHandler.GetCurrentWeekMenuPlan)
				menuPlans.GET("/:id", menuPlanningHandler.GetMenuPlan)
				menuPlans.PUT("/:id", menuPlanningHandler.UpdateMenuPlan)
				menuPlans.POST("/:id/approve", menuPlanningHandler.ApproveMenuPlan)
				menuPlans.POST("/:id/duplicate", menuPlanningHandler.DuplicateMenuPlan)
				menuPlans.GET("/:id/daily-nutrition", menuPlanningHandler.GetDailyNutrition)
				menuPlans.GET("/:id/ingredient-requirements", menuPlanningHandler.GetIngredientRequirements)
			}

			// KDS routes
			kdsService, err := services.NewKDSService(db, firebaseApp)
			if err != nil {
				panic("Failed to initialize KDS service: " + err.Error())
			}
			packingAllocationService, err := services.NewPackingAllocationService(db, firebaseApp)
			if err != nil {
				panic("Failed to initialize Packing Allocation service: " + err.Error())
			}
			kdsHandler := handlers.NewKDSHandler(kdsService, packingAllocationService)
			kds := protected.Group("/kds")
			{
				// Cooking routes
				kds.GET("/cooking/today", kdsHandler.GetCookingToday)
				kds.PUT("/cooking/:recipe_id/status", kdsHandler.UpdateCookingStatus)
				kds.POST("/cooking/sync", kdsHandler.SyncCookingToFirebase)

				// Packing routes
				kds.GET("/packing/today", kdsHandler.GetPackingToday)
				kds.PUT("/packing/:school_id/status", kdsHandler.UpdatePackingStatus)
				kds.POST("/packing/sync", kdsHandler.SyncPackingToFirebase)
			}

			// Supply Chain routes
			supplyChainHandler := handlers.NewSupplyChainHandler(db)
			
			// Supplier routes
			suppliers := protected.Group("/suppliers")
			// Apply caching for supplier GET requests
			if cacheService != nil {
				suppliers.Use(middleware.ConditionalCacheMiddleware(cacheService, middleware.CacheForReadOnlyOperations, cache.LongCacheDuration))
			}
			{
				suppliers.GET("", supplyChainHandler.GetAllSuppliers)
				suppliers.POST("", supplyChainHandler.CreateSupplier)
				suppliers.GET("/:id", supplyChainHandler.GetSupplier)
				suppliers.PUT("/:id", supplyChainHandler.UpdateSupplier)
				suppliers.GET("/:id/performance", supplyChainHandler.GetSupplierPerformance)
			}

			// Purchase Order routes
			purchaseOrders := protected.Group("/purchase-orders")
			{
				purchaseOrders.GET("", supplyChainHandler.GetAllPurchaseOrders)
				purchaseOrders.POST("", supplyChainHandler.CreatePurchaseOrder)
				purchaseOrders.GET("/:id", supplyChainHandler.GetPurchaseOrder)
				purchaseOrders.PUT("/:id", supplyChainHandler.UpdatePurchaseOrder)
				purchaseOrders.POST("/:id/approve", supplyChainHandler.ApprovePurchaseOrder)
			}

			// Goods Receipt routes
			goodsReceipts := protected.Group("/goods-receipts")
			{
				goodsReceipts.GET("", supplyChainHandler.GetAllGoodsReceipts)
				goodsReceipts.POST("", supplyChainHandler.CreateGoodsReceipt)
				goodsReceipts.GET("/:id", supplyChainHandler.GetGoodsReceipt)
				goodsReceipts.POST("/:id/upload-invoice", supplyChainHandler.UploadInvoicePhoto)
			}

			// Inventory routes
			inventory := protected.Group("/inventory")
			// Apply inventory caching middleware
			if cacheService != nil {
				inventory.Use(middleware.InventoryCacheMiddleware(cacheService))
			}
			{
				inventory.GET("", supplyChainHandler.GetInventory)
				inventory.GET("/alerts", supplyChainHandler.GetInventoryAlerts)
				inventory.GET("/movements", supplyChainHandler.GetInventoryMovements)
				inventory.POST("/initialize", supplyChainHandler.InitializeInventory)
			}

			// Logistics routes
			logisticsHandler := handlers.NewLogisticsHandler(db)
			
			// School routes
			schools := protected.Group("/schools")
			{
				schools.GET("", logisticsHandler.GetAllSchools)
				schools.POST("", logisticsHandler.CreateSchool)
				schools.GET("/:id", logisticsHandler.GetSchool)
				schools.PUT("/:id", logisticsHandler.UpdateSchool)
			}

			// Delivery Task routes
			deliveryTasks := protected.Group("/delivery-tasks")
			{
				deliveryTasks.GET("", logisticsHandler.GetAllDeliveryTasks)
				deliveryTasks.POST("", logisticsHandler.CreateDeliveryTask)
				deliveryTasks.GET("/driver/:driver_id/today", logisticsHandler.GetDriverTasksToday)
				deliveryTasks.GET("/:id", logisticsHandler.GetDeliveryTask)
				deliveryTasks.PUT("/:id", logisticsHandler.UpdateDeliveryTask)
				deliveryTasks.PUT("/:id/status", logisticsHandler.UpdateDeliveryTaskStatus)
				deliveryTasks.DELETE("/:id", logisticsHandler.DeleteDeliveryTask)
			}

			// e-POD routes
			epod := protected.Group("/epod")
			{
				epod.POST("", logisticsHandler.CreateEPOD)
				epod.POST("/:id/upload-photo", logisticsHandler.UploadEPODPhoto)
				epod.POST("/:id/upload-signature", logisticsHandler.UploadEPODSignature)
			}

			// Ompreng Tracking routes
			ompreng := protected.Group("/ompreng")
			{
				ompreng.GET("/tracking", logisticsHandler.GetOmprengTracking)
				ompreng.POST("/drop-off", logisticsHandler.RecordOmprengDropOff)
				ompreng.POST("/pick-up", logisticsHandler.RecordOmprengPickUp)
				ompreng.GET("/reports", logisticsHandler.GetOmprengReports)
			}

			// HRM routes
			authService := services.NewAuthService(db, cfg.JWTSecret)
			hrmHandler := handlers.NewHRMHandler(db, authService)
			
			// Employee routes
			employees := protected.Group("/employees")
			{
				employees.GET("", hrmHandler.GetEmployees)
				employees.POST("", hrmHandler.CreateEmployee)
				employees.GET("/stats", hrmHandler.GetEmployeeStats)
				employees.GET("/:id", hrmHandler.GetEmployeeByID)
				employees.PUT("/:id", hrmHandler.UpdateEmployee)
				employees.POST("/:id/deactivate", hrmHandler.DeactivateEmployee)
			}

			// Attendance routes
			attendance := protected.Group("/attendance")
			{
				attendance.POST("/check-in", hrmHandler.CheckIn)
				attendance.POST("/check-out", hrmHandler.CheckOut)
				attendance.POST("/validate-wifi", hrmHandler.ValidateWiFi)
				attendance.GET("/today", hrmHandler.GetTodayAttendance)
				attendance.GET("/report", hrmHandler.GetAttendanceReport)
				attendance.GET("/by-date-range", hrmHandler.GetAttendanceByDateRange)
				attendance.GET("/export/excel", hrmHandler.ExportAttendanceReport)
				attendance.GET("/export/pdf", hrmHandler.ExportAttendanceReport)
				attendance.GET("/stats", hrmHandler.GetAttendanceStats)
			}

			// Wi-Fi Configuration routes
			wifiConfig := protected.Group("/wifi-config")
			{
				wifiConfig.GET("", hrmHandler.GetWiFiConfigs)
				wifiConfig.POST("", hrmHandler.CreateWiFiConfig)
				wifiConfig.PUT("/:id", hrmHandler.UpdateWiFiConfig)
				wifiConfig.DELETE("/:id", hrmHandler.DeleteWiFiConfig)
			}

			// System Configuration routes (admin only with IP whitelist)
			systemConfigHandler := handlers.NewSystemConfigHandler(db)
			systemConfig := protected.Group("/system-config")
			if len(cfg.AdminWhitelistIPs) > 0 {
				systemConfig.Use(middleware.IPWhitelist(cfg.AdminWhitelistIPs))
			}
			{
				systemConfig.GET("", systemConfigHandler.GetAllConfigs)
				systemConfig.GET("/by-category", systemConfigHandler.GetConfigsByCategory)
				systemConfig.GET("/:key", systemConfigHandler.GetConfig)
				systemConfig.POST("", systemConfigHandler.SetConfig)
				systemConfig.POST("/bulk", systemConfigHandler.SetMultipleConfigs)
				systemConfig.POST("/initialize-defaults", systemConfigHandler.InitializeDefaultConfigs)
				systemConfig.DELETE("/:key", systemConfigHandler.DeleteConfig)
			}

			// Financial routes
			financialHandler := handlers.NewFinancialHandler(db)
			
			// Asset routes
			assets := protected.Group("/assets")
			{
				assets.GET("", financialHandler.GetAllAssets)
				assets.POST("", financialHandler.CreateAsset)
				assets.GET("/report", financialHandler.GetAssetReport)
				assets.GET("/:id", financialHandler.GetAsset)
				assets.PUT("/:id", financialHandler.UpdateAsset)
				assets.DELETE("/:id", financialHandler.DeleteAsset)
				assets.POST("/:id/maintenance", financialHandler.AddMaintenance)
				assets.GET("/:id/depreciation-schedule", financialHandler.GetDepreciationSchedule)
			}

			// Cash Flow routes
			cashFlow := protected.Group("/cash-flow")
			{
				cashFlow.GET("", financialHandler.GetAllCashFlow)
				cashFlow.POST("", financialHandler.CreateCashFlow)
				cashFlow.GET("/summary", financialHandler.GetCashFlowSummary)
			}

			// Financial Report routes
			financialReports := protected.Group("/financial-reports")
			{
				financialReports.GET("", financialHandler.GetFinancialReport)
				financialReports.POST("/export", financialHandler.ExportFinancialReport)
			}

			// Dashboard routes (works with or without Firebase)
			dashboardHandler, err := handlers.NewDashboardHandler(db, firebaseApp)
			if err != nil {
				log.Printf("Warning: Dashboard handler initialization failed: %v. Using dummy data mode.", err)
			}
			dashboard := protected.Group("/dashboard")
			// Apply dashboard caching middleware
			if cacheService != nil {
				dashboard.Use(middleware.DashboardCacheMiddleware(cacheService))
			}
			{
				dashboard.GET("/kepala-sppg", dashboardHandler.GetKepalaSSPGDashboard)
				dashboard.GET("/kepala-yayasan", dashboardHandler.GetKepalaYayasanDashboard)
				dashboard.GET("/kpi", dashboardHandler.GetKPIs)
				dashboard.POST("/sync", dashboardHandler.SyncDashboardToFirebase)
				dashboard.POST("/export", dashboardHandler.ExportDashboard)
			}

			// Notification routes
			notificationHandler, err := handlers.NewNotificationHandler(db, firebaseApp)
			if err != nil {
				panic("Failed to initialize Notification handler: " + err.Error())
			}
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notificationHandler.GetNotifications)
				notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
				notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
				notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
				notifications.DELETE("/:id", notificationHandler.DeleteNotification)
			}

			// Audit Trail routes
			auditHandler := handlers.NewAuditHandler(db)
			auditTrail := protected.Group("/audit-trail")
			{
				auditTrail.GET("", auditHandler.GetAuditTrail)
				auditTrail.GET("/stats", auditHandler.GetAuditStats)
			}
		}
	}

	return r
}
