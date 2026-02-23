# Implementation Plan: Sistem ERP SPPG

## Overview

Sistem ERP SPPG akan diimplementasikan dengan arsitektur 3-tier:
- **Backend**: Golang dengan Gin framework, PostgreSQL, Firebase Admin SDK
- **Web App**: Vue 3 dengan JavaScript, Pinia state management, Ant Design Vue
- **PWA**: Vue 3 dengan PWA plugin, IndexedDB untuk offline storage

Implementasi akan dilakukan secara modular, dimulai dari infrastruktur dasar, kemudian modul-modul inti, dan terakhir integrasi serta testing menyeluruh.

## Tasks

- [x] 1. Setup Project Infrastructure
  - Inisialisasi repository dengan struktur folder untuk backend, web, dan PWA
  - Setup Golang project dengan Go modules, konfigurasi database PostgreSQL
  - Setup Vue 3 projects untuk web dan PWA dengan Vite
  - Konfigurasi Firebase project dan credentials
  - Setup environment variables dan configuration management
  - _Requirements: Semua modul memerlukan infrastruktur dasar_

- [x] 2. Implement Database Schema and Migrations
  - [x] 2.1 Create database migration files untuk semua tabel
    - Buat migration files menggunakan golang-migrate atau GORM AutoMigrate
    - Definisikan schema untuk User, AuditTrail, Recipe, Ingredient, MenuPlan, Supplier, PurchaseOrder, Inventory, School, DeliveryTask, Employee, Attendance, Asset, CashFlow
    - Tambahkan indexes untuk kolom yang sering di-query (user_id, date, status)
    - _Requirements: 1.1-1.6, 2.1-2.6, 3.1-3.6, 6.1-6.6, 7.1-7.6, 8.1-8.6, 9.1-9.6, 10.1-10.6, 14.1-14.6, 16.1-16.6, 17.1-17.6_

  - [x] 2.2 Implement GORM models untuk semua entities
    - Buat struct definitions dengan GORM tags untuk User, Recipe, Ingredient, dll
    - Definisikan relationships (hasMany, belongsTo, manyToMany)
    - Tambahkan validation tags
    - _Requirements: Semua data models_

- [x] 3. Implement Authentication & Authorization Module (Backend)
  - [x] 3.1 Implement User authentication service
    - Buat AuthService dengan methods: Login, Logout, RefreshToken, ValidateToken
    - Implement password hashing menggunakan bcrypt
    - Generate JWT tokens dengan claims (user_id, role, exp)
    - _Requirements: 1.1, 1.2, 25.1_

  - [x] 3.2 Write property test untuk authentication
    - **Property 1: Authentication Success for Valid Credentials**
    - **Property 2: Authentication Rejection for Invalid Credentials**
    - **Validates: Requirements 1.1, 1.2**

  - [x] 3.3 Implement RBAC middleware
    - Buat middleware untuk validate JWT token pada setiap request
    - Implement permission checking berdasarkan role dan feature
    - Return 401 untuk unauthenticated, 403 untuk unauthorized
    - _Requirements: 1.4, 1.5_

  - [x] 3.4 Write property test untuk RBAC
    - **Property 3: Role-Based Access Control**
    - **Validates: Requirements 1.5**

  - [x] 3.5 Implement Audit Trail service
    - Buat AuditTrailService dengan method RecordAction
    - Implement middleware untuk auto-record semua create/update/delete actions
    - Store user_id, timestamp, action, entity, old_value, new_value
    - _Requirements: 1.6, 21.1, 21.2_

  - [x] 3.6 Write property test untuk audit trail
    - **Property 4: Audit Trail Completeness**
    - **Validates: Requirements 1.6, 21.1, 21.2**

  - [x] 3.7 Implement authentication API endpoints
    - POST /api/v1/auth/login - accept NIK/Email + password
    - POST /api/v1/auth/logout - invalidate token
    - POST /api/v1/auth/refresh - refresh JWT token
    - GET /api/v1/auth/me - get current user info
    - _Requirements: 1.1, 1.2, 1.3_


- [x] 4. Implement Recipe & Menu Planning Module (Backend)
  - [x] 4.1 Implement Recipe service dengan nutrition calculation
    - Buat RecipeService dengan CRUD methods
    - Implement CalculateNutrition function yang sum ingredient nutrition values
    - Implement ValidateNutrition function untuk check minimum standards
    - Maintain recipe version history pada setiap update
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [ ]* 4.2 Write property tests untuk recipe nutrition
    - **Property 5: Nutritional Calculation Accuracy**
    - **Property 6: Recipe Nutrition Recalculation**
    - **Validates: Requirements 2.2, 2.4**

  - [x] 4.3 Implement Menu Planning service
    - Buat MenuPlanningService dengan methods: CreateWeeklyPlan, ApproveMenu
    - Implement CalculateDailyNutrition untuk aggregate recipe nutrition
    - Implement CalculateIngredientRequirements untuk procurement
    - Allow duplicate previous menus sebagai template
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_

  - [ ]* 4.4 Write property tests untuk menu planning
    - **Property 7: Daily Menu Nutrition Aggregation**
    - **Property 8: Ingredient Requirements Calculation**
    - **Validates: Requirements 3.2, 3.5**

  - [x] 4.5 Implement Recipe & Menu API endpoints
    - GET/POST/PUT/DELETE /api/v1/recipes
    - GET /api/v1/recipes/:id/nutrition
    - GET /api/v1/recipes/:id/history
    - GET/POST/PUT /api/v1/menu-plans
    - POST /api/v1/menu-plans/:id/approve
    - GET /api/v1/menu-plans/current-week
    - _Requirements: 2.1-2.6, 3.1-3.6_

- [x] 5. Implement Kitchen Display System Module (Backend)
  - [x] 5.1 Implement KDS service dengan Firebase integration
    - Buat KDSService dengan methods: GetTodayMenu, UpdateRecipeStatus, GetPackingAllocations
    - Implement Firebase sync untuk push real-time updates ke /kds/cooking dan /kds/packing
    - Trigger inventory deduction saat status berubah ke "Sedang Dimasak"
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

  - [ ]* 5.2 Write property test untuk inventory deduction
    - **Property 9: Inventory Deduction on Cooking Start**
    - **Validates: Requirements 4.6**

  - [x] 5.3 Implement Packing Allocation service
    - Buat PackingAllocationService untuk calculate portions per school
    - Map menu items ke schools berdasarkan student count
    - Track packing completion status
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

  - [x] 5.4 Implement KDS API endpoints
    - GET /api/v1/kds/cooking/today
    - PUT /api/v1/kds/cooking/:recipe_id/status
    - GET /api/v1/kds/packing/today
    - PUT /api/v1/kds/packing/:school_id/status
    - _Requirements: 4.1-4.6, 5.1-5.6_

- [x] 6. Implement Supply Chain & Inventory Module (Backend)
  - [x] 6.1 Implement Supplier service
    - Buat SupplierService dengan CRUD methods
    - Track transaction history dan performance metrics
    - Calculate on-time delivery rate dan quality ratings
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

  - [x] 6.2 Implement Purchase Order service
    - Buat PurchaseOrderService dengan methods: CreatePO, ApprovePO, TrackPO
    - Generate unique PO numbers
    - Implement approval workflow dengan notifications
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

  - [x] 6.3 Implement Goods Receipt service dengan photo upload
    - Buat GoodsReceiptService dengan method CreateGRN
    - Integrate dengan Cloud Storage untuk upload invoice photos
    - Compare received vs ordered quantities dan flag discrepancies
    - Trigger inventory update dan cash flow entry
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

  - [ ]* 6.4 Write property tests untuk inventory management
    - **Property 10: Inventory Update on Goods Receipt**
    - **Property 11: FIFO/FEFO Inventory Method**
    - **Property 23: Automatic Cash Flow Entry from GRN**
    - **Validates: Requirements 8.4, 8.6, 17.4**

  - [x] 6.5 Implement Inventory service dengan FIFO/FEFO
    - Buat InventoryService dengan methods: UpdateStock, CheckLowStock, GetMovements
    - Implement FIFO/FEFO logic berdasarkan expiry dates
    - Generate low stock alerts saat quantity < threshold
    - Maintain real-time inventory levels
    - _Requirements: 8.6, 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

  - [ ]* 6.6 Write property tests untuk stock alerts
    - **Property 12: Real-Time Inventory Maintenance**
    - **Property 13: Low Stock Alert Generation**
    - **Validates: Requirements 9.1, 9.2, 9.3**

  - [x] 6.7 Implement Supply Chain API endpoints
    - GET/POST/PUT /api/v1/suppliers
    - GET /api/v1/suppliers/:id/performance
    - GET/POST/PUT /api/v1/purchase-orders
    - POST /api/v1/purchase-orders/:id/approve
    - POST /api/v1/goods-receipts
    - POST /api/v1/goods-receipts/:id/upload-invoice
    - GET /api/v1/inventory
    - GET /api/v1/inventory/alerts
    - GET /api/v1/inventory/movements
    - _Requirements: 6.1-6.6, 7.1-7.6, 8.1-8.6, 9.1-9.6_


- [x] 7. Implement Logistics & Distribution Module (Backend)
  - [x] 7.1 Implement School service dengan GPS validation
    - Buat SchoolService dengan CRUD methods
    - Validate GPS coordinates (lat: -90 to 90, lng: -180 to 180)
    - Maintain change history untuk school updates
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_

  - [ ]* 7.2 Write property test untuk GPS validation
    - **Property 14: GPS Coordinate Validation**
    - **Validates: Requirements 10.2**

  - [x] 7.3 Implement Delivery Task service
    - Buat DeliveryTaskService dengan methods: CreateTasks, AssignDriver, TrackStatus
    - Optimize route sequence (optional: integrate Google Maps API)
    - Allow logistics staff assign tasks ke drivers
    - _Requirements: 11.1, 11.2, 11.3, 11.6_

  - [x] 7.4 Implement e-POD service dengan geotagging
    - Buat ePODService dengan method CreateProof
    - Auto-capture GPS coordinates dari device
    - Upload photo dan signature ke Cloud Storage
    - Update delivery status ke "Selesai" dengan timestamp
    - Support offline queue dan sync saat online
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6_

  - [ ]* 7.5 Write property tests untuk delivery tracking
    - **Property 15: Automatic Geotagging on Delivery**
    - **Property 16: Delivery Status Update on e-POD Completion**
    - **Validates: Requirements 12.1, 12.5**

  - [x] 7.6 Implement Ompreng Tracking service
    - Buat OmprengTrackingService dengan methods: RecordDropOff, RecordPickUp, GetBalance
    - Maintain school-level ompreng balance (cumulative drop-off - pick-up)
    - Maintain global ompreng inventory (kitchen + circulation)
    - Flag schools dengan missing ompreng
    - Generate circulation reports
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6_

  - [ ]* 7.7 Write property tests untuk ompreng tracking
    - **Property 17: Ompreng Balance Conservation**
    - **Property 18: Global Ompreng Inventory Conservation**
    - **Validates: Requirements 13.1, 13.2, 13.3**

  - [x] 7.8 Implement Logistics API endpoints
    - GET/POST/PUT /api/v1/schools
    - GET/POST /api/v1/delivery-tasks
    - GET /api/v1/delivery-tasks/driver/:driver_id/today
    - PUT /api/v1/delivery-tasks/:id/status
    - POST /api/v1/epod
    - POST /api/v1/epod/upload-photo
    - POST /api/v1/epod/upload-signature
    - GET/POST /api/v1/ompreng/tracking
    - POST /api/v1/ompreng/drop-off
    - POST /api/v1/ompreng/pick-up
    - GET /api/v1/ompreng/reports
    - _Requirements: 10.1-10.6, 11.1-11.6, 12.1-12.6, 13.1-13.6_

- [x] 8. Checkpoint - Backend Core Modules Complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 9. Implement Human Resource Management Module (Backend)
  - [x] 9.1 Implement Employee service
    - Buat EmployeeService dengan CRUD methods
    - Validate unique NIK dan email
    - Auto-generate login credentials saat create employee
    - Maintain change history
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.6_

  - [ ]* 9.2 Write property test untuk employee uniqueness
    - **Property 19: Unique Employee Identifiers**
    - **Validates: Requirements 14.2**

  - [x] 9.3 Implement Attendance service dengan Wi-Fi validation
    - Buat AttendanceService dengan methods: CheckIn, CheckOut, ValidateWiFi
    - Validate SSID dan BSSID terhadap authorized networks
    - Calculate work hours (check-out - check-in)
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5, 15.6_

  - [ ]* 9.4 Write property tests untuk attendance
    - **Property 20: Wi-Fi Validation for Attendance**
    - **Property 21: Work Hours Calculation**
    - **Validates: Requirements 15.1, 15.2, 15.3, 15.5**

  - [x] 9.5 Implement HRM API endpoints
    - GET/POST/PUT/DELETE /api/v1/employees
    - POST /api/v1/attendance/check-in
    - POST /api/v1/attendance/check-out
    - POST /api/v1/attendance/validate-wifi
    - GET /api/v1/attendance/report
    - GET/POST /api/v1/wifi-config
    - _Requirements: 14.1-14.6, 15.1-15.6_

- [x] 10. Implement Financial & Asset Management Module (Backend)
  - [x] 10.1 Implement Asset service dengan depreciation
    - Buat AssetService dengan CRUD methods
    - Calculate depreciation berdasarkan purchase date dan rate
    - Calculate current book value (purchase price - accumulated depreciation)
    - Record maintenance activities
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.6_

  - [ ]* 10.2 Write property test untuk asset depreciation
    - **Property 22: Asset Depreciation Calculation**
    - **Validates: Requirements 16.4, 16.5**

  - [x] 10.3 Implement Cash Flow service
    - Buat CashFlowService dengan methods: CreateEntry, GetBalance, GenerateReport
    - Auto-create entry saat GRN completed
    - Categorize transactions (Bahan Baku, Gaji, Utilitas, dll)
    - Calculate running balance per category
    - _Requirements: 17.1, 17.2, 17.3, 17.4, 17.5, 17.6_

  - [ ]* 10.4 Write property test untuk cash flow
    - **Property 24: Running Balance Accuracy**
    - **Validates: Requirements 17.6**

  - [x] 10.5 Implement Financial Report service dengan export
    - Buat FinancialReportService dengan methods: GenerateReport, ExportPDF, ExportExcel
    - Aggregate cash flow entries by date range
    - Generate budget vs actual comparison
    - Include summary tables dan charts
    - _Requirements: 18.1, 18.2, 18.3, 18.4, 18.5, 18.6_

  - [ ]* 10.6 Write property test untuk report aggregation
    - **Property 25: Financial Report Aggregation**
    - **Validates: Requirements 18.1, 18.2**

  - [x] 10.7 Implement Financial API endpoints
    - GET/POST/PUT /api/v1/assets
    - POST /api/v1/assets/:id/maintenance
    - GET/POST /api/v1/cash-flow
    - GET /api/v1/cash-flow/summary
    - GET /api/v1/financial-reports
    - POST /api/v1/financial-reports/export
    - _Requirements: 16.1-16.6, 17.1-17.6, 18.1-18.6_


- [x] 11. Implement Executive Dashboard Module (Backend)
  - [x] 11.1 Implement Dashboard service dengan Firebase sync
    - Buat DashboardService dengan methods: GetKepalaSSPGDashboard, GetKepalaYayasanDashboard
    - Aggregate data dari production, delivery, inventory, financial modules
    - Calculate KPIs (portions prepared, delivery rate, stock availability, budget absorption)
    - Push updates ke Firebase /dashboard path untuk real-time sync
    - _Requirements: 19.1, 19.2, 19.3, 19.4, 19.5, 19.6, 20.1, 20.2, 20.3, 20.4, 20.5, 20.6_

  - [x] 11.2 Implement Dashboard API endpoints
    - GET /api/v1/dashboard/kepala-sppg
    - GET /api/v1/dashboard/kepala-yayasan
    - GET /api/v1/dashboard/kpi
    - POST /api/v1/dashboard/export
    - _Requirements: 19.1-19.6, 20.1-20.6_

- [x] 12. Implement Real-Time Sync & Notification Services (Backend)
  - [x] 12.1 Implement Firebase real-time sync service
    - Setup Firebase Admin SDK
    - Implement methods untuk push updates ke Firebase paths
    - Handle connection state dan reconnection
    - Implement conflict resolution untuk concurrent updates
    - _Requirements: 22.1, 22.2, 22.3, 22.4, 22.5, 22.6_

  - [ ]* 12.2 Write property test untuk real-time sync
    - **Property 26: Real-Time Data Push**
    - **Validates: Requirements 22.1, 22.2**

  - [x] 12.3 Implement Notification service
    - Buat NotificationService dengan method SendNotification
    - Send notifications untuk low stock, PO approval, packing complete, delivery complete
    - Store notifications di database dengan read/unread status
    - _Requirements: 28.1, 28.2, 28.3, 28.4, 28.5, 28.6_

  - [x] 12.4 Implement Notification API endpoints
    - GET /api/v1/notifications
    - PUT /api/v1/notifications/:id/read
    - _Requirements: 28.1-28.6_

- [x] 13. Implement System Configuration & Utilities (Backend)
  - [x] 13.1 Implement System Config service
    - Buat SystemConfigService dengan methods: GetConfig, SetConfig
    - Support config types: string, int, float, bool, json
    - Apply config changes immediately tanpa restart
    - _Requirements: 29.1, 29.2, 29.3, 29.4, 29.5, 29.6_

  - [x] 13.2 Implement Export service untuk PDF dan Excel
    - Buat ExportService dengan methods: ExportToPDF, ExportToExcel
    - Format data dengan headers dalam Bahasa Indonesia
    - Include organization header, title, date range, page numbers
    - Generate files dalam 30 detik untuk 10,000 records
    - _Requirements: 27.1, 27.2, 27.3, 27.4, 27.5, 27.6_

  - [ ]* 13.3 Write property test untuk export completeness
    - **Property 30: Export Data Completeness**
    - **Validates: Requirements 27.5**

  - [x] 13.3 Implement validation utilities
    - Buat validation functions untuk email, phone, NIK, GPS coordinates
    - Implement input sanitization untuk prevent SQL injection dan XSS
    - _Requirements: 30.1, 30.2, 30.4, 30.5, 30.6_

  - [x] 13.4 Implement Security utilities
    - Implement password hashing dengan bcrypt
    - Implement session timeout enforcement
    - Implement rate limiting middleware
    - _Requirements: 25.1, 25.2, 25.3, 25.4, 25.5, 25.6_

  - [ ]* 13.5 Write property tests untuk security
    - **Property 28: Password Hashing**
    - **Property 29: Session Timeout Enforcement**
    - **Validates: Requirements 25.1, 25.3, 25.4**

- [x] 14. Checkpoint - Backend Complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 15. Implement Web App - Authentication & Layout (Vue.js)
  - [x] 15.1 Setup Vue 3 project dengan Vite dan dependencies
    - Initialize Vue 3 project dengan Vite
    - Install dependencies: vue-router, pinia, axios, ant-design-vue, firebase
    - Setup folder structure: components, views, stores, services, utils
    - Configure environment variables untuk API base URL dan Firebase config
    - _Requirements: Infrastructure_

  - [x] 15.2 Implement authentication store dan service
    - Buat auth store dengan Pinia (state: user, token, isAuthenticated)
    - Buat auth service dengan methods: login, logout, refreshToken
    - Store JWT token di localStorage
    - Implement axios interceptor untuk attach token ke requests
    - _Requirements: 1.1, 1.2, 1.3_

  - [x] 15.3 Implement Login page
    - Buat LoginView component dengan form (NIK/Email, Password)
    - Validate input dan show error messages dalam Bahasa Indonesia
    - Redirect ke dashboard sesuai role setelah login sukses
    - _Requirements: 1.1, 1.2, 1.3, 24.1, 24.2_

  - [x] 15.4 Implement main layout dengan navigation
    - Buat MainLayout component dengan sidebar navigation
    - Show menu items berdasarkan user role (RBAC)
    - Include header dengan user info dan logout button
    - Include notification bell dengan unread count
    - _Requirements: 1.4, 1.5, 28.5_

  - [x] 15.5 Implement route guards untuk authentication
    - Setup vue-router dengan route guards
    - Redirect ke login jika tidak authenticated
    - Check role permissions untuk protected routes
    - _Requirements: 1.5_


- [x] 16. Implement Web App - Recipe & Menu Planning Module
  - [x] 16.1 Implement Recipe management pages
    - Buat RecipeListView dengan table, search, dan filter
    - Buat RecipeFormView untuk create/edit recipe dengan ingredient selection
    - Show calculated nutrition values real-time saat input ingredients
    - Show validation errors jika nutrition tidak meet standards
    - Show recipe version history
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

  - [x] 16.2 Implement Menu Planning pages
    - Buat MenuPlanningView dengan weekly calendar
    - Allow drag-drop recipes ke days atau select dari dropdown
    - Show daily nutrition totals dan validation status
    - Implement approve menu button (untuk Ahli Gizi)
    - Allow duplicate previous week menu
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_

- [x] 17. Implement Web App - Kitchen Display System
  - [x] 17.1 Implement KDS Cooking Display
    - Buat KDSCookingView dengan real-time Firebase listener
    - Show list of recipes untuk hari ini dengan status
    - Show ingredient quantities dan cooking instructions per recipe
    - Implement buttons untuk update status (Mulai Masak, Selesai)
    - Auto-refresh saat ada perubahan dari Firebase
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 22.1, 22.2, 22.3_

  - [x] 17.2 Implement KDS Packing Display
    - Buat KDSPackingView dengan real-time Firebase listener
    - Show list of schools dengan portion allocations
    - Show menu items per school
    - Implement button untuk mark "Siap Kirim"
    - Show notification saat semua schools ready
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 22.1, 22.2, 22.3_

- [x] 18. Implement Web App - Supply Chain & Inventory Module
  - [x] 18.1 Implement Supplier management pages
    - Buat SupplierListView dengan table dan search
    - Buat SupplierFormView untuk create/edit supplier
    - Show supplier performance metrics (on-time delivery, quality rating)
    - Show transaction history
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

  - [x] 18.2 Implement Purchase Order pages
    - Buat POListView dengan table, filter by status
    - Buat POFormView untuk create PO dengan supplier dan item selection
    - Show PO approval workflow (submit untuk approval, approve button untuk Kepala SPPG)
    - Track PO status dari creation sampai delivery
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

  - [x] 18.3 Implement Goods Receipt pages
    - Buat GRNFormView untuk create GRN linked ke PO
    - Implement photo upload untuk invoice
    - Show comparison received vs ordered quantities dengan highlight discrepancies
    - Auto-update inventory setelah GRN completed
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_

  - [x] 18.4 Implement Inventory pages
    - Buat InventoryListView dengan table showing current stock, min threshold, days of supply
    - Highlight items dengan low stock (red color)
    - Show stock movement history dengan filter by date range
    - Show alerts untuk low stock items
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

- [x] 19. Implement Web App - Logistics & Distribution Module
  - [x] 19.1 Implement School management pages
    - Buat SchoolListView dengan table dan search
    - Buat SchoolFormView untuk create/edit school dengan GPS input
    - Validate GPS coordinates format
    - Show change history
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_

  - [x] 19.2 Implement Delivery Task management pages
    - Buat DeliveryTaskListView dengan table, filter by date dan driver
    - Buat DeliveryTaskFormView untuk create tasks dan assign ke driver
    - Show optimized route sequence
    - Track delivery status real-time
    - _Requirements: 11.1, 11.2, 11.3, 11.6_

  - [x] 19.3 Implement Ompreng Tracking pages
    - Buat OmprengTrackingView dengan table showing balance per school
    - Highlight schools dengan missing ompreng
    - Show global ompreng inventory (kitchen + circulation)
    - Generate circulation reports dengan export
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6_

- [x] 20. Implement Web App - HRM Module
  - [x] 20.1 Implement Employee management pages
    - Buat EmployeeListView dengan table dan search
    - Buat EmployeeFormView untuk create/edit employee
    - Validate unique NIK dan email
    - Show auto-generated credentials saat create
    - Allow deactivate employee
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.6_

  - [x] 20.2 Implement Attendance report pages
    - Buat AttendanceReportView dengan table, filter by date range dan employee
    - Show check-in, check-out, work hours
    - Export report ke Excel/PDF
    - _Requirements: 15.4, 15.5, 27.1, 27.2, 27.3, 27.4_

  - [x] 20.3 Implement Wi-Fi configuration page
    - Buat WiFiConfigView untuk manage authorized networks
    - Allow add/edit/delete SSID dan BSSID
    - _Requirements: 15.6_

- [x] 21. Implement Web App - Financial & Asset Module
  - [x] 21.1 Implement Asset management pages
    - Buat AssetListView dengan table showing assets, book value, depreciation
    - Buat AssetFormView untuk create/edit asset
    - Show maintenance history dan allow add maintenance records
    - Generate asset reports dengan export
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.6_

  - [x] 21.2 Implement Cash Flow pages
    - Buat CashFlowListView dengan table, filter by date range dan category
    - Buat CashFlowFormView untuk manual entry
    - Show running balance per category
    - _Requirements: 17.1, 17.2, 17.3, 17.5, 17.6_

  - [x] 21.3 Implement Financial Report pages
    - Buat FinancialReportView dengan date range selector
    - Show summary tables (income, expenses by category, net cash flow)
    - Show charts (trend over time, category breakdown)
    - Show budget vs actual comparison
    - Implement export ke PDF dan Excel
    - _Requirements: 18.1, 18.2, 18.3, 18.4, 18.5, 18.6, 27.1, 27.2, 27.3_


- [x] 22. Implement Web App - Executive Dashboard Module
  - [x] 22.1 Implement Kepala SPPG Dashboard
    - Buat DashboardKepalaSSPGView dengan real-time Firebase listener
    - Show production milestones (menu status, cooking progress, packing status)
    - Show delivery status (completed vs pending) dengan real-time updates
    - Show critical stock items dengan red highlights
    - Show KPIs (total portions prepared, delivery completion rate, stock availability %)
    - Implement drill-down ke detail pages saat click metrics
    - _Requirements: 19.1, 19.2, 19.3, 19.4, 19.5, 19.6, 22.1, 22.2, 22.3_

  - [x] 22.2 Implement Kepala Yayasan Dashboard
    - Buat DashboardKepalaYayasanView dengan real-time Firebase listener
    - Show budget absorption rate dengan progress bar
    - Show cumulative nutrition distribution metrics (portions, schools, students)
    - Show supplier performance metrics
    - Show trend charts (budget spending, distribution volumes over time)
    - Allow select time period dan update all metrics
    - Implement export dashboard data dan charts
    - _Requirements: 20.1, 20.2, 20.3, 20.4, 20.5, 20.6, 22.1, 22.2, 22.3_

- [x] 23. Implement Web App - Audit Trail & System Config
  - [x] 23.1 Implement Audit Trail page
    - Buat AuditTrailView dengan table showing all user actions
    - Implement search dan filter by date range, user, action type, entity
    - Show entries dalam reverse chronological order
    - Display clear descriptions dalam Bahasa Indonesia
    - _Requirements: 21.1, 21.2, 21.3, 21.4, 21.5, 21.6_

  - [x] 23.2 Implement System Configuration page
    - Buat SystemConfigView untuk manage system parameters
    - Allow configure min stock thresholds, nutrition standards, session timeout, backup schedule
    - Apply changes immediately tanpa restart
    - _Requirements: 29.1, 29.2, 29.3, 29.4, 29.5, 29.6_

- [x] 24. Checkpoint - Web App Complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 25. Implement PWA - Setup & Authentication
  - [x] 25.1 Setup Vue 3 PWA project
    - Initialize Vue 3 project dengan Vite dan PWA plugin
    - Install dependencies: vue-router, pinia, axios, vant (mobile UI), firebase, dexie (IndexedDB)
    - Configure service worker dengan Workbox
    - Setup offline caching strategy
    - _Requirements: Infrastructure, 23.1, 23.2, 23.3, 23.4, 23.5, 23.6_

  - [x] 25.2 Implement PWA authentication
    - Buat auth store dan service (similar to web app)
    - Buat LoginView untuk PWA dengan mobile-friendly UI
    - Store credentials securely
    - _Requirements: 1.1, 1.2, 1.3_

  - [x] 25.3 Implement PWA main layout
    - Buat MainLayout dengan bottom navigation (untuk Driver: Tugas, e-POD, Profil)
    - Show offline indicator saat no connection
    - _Requirements: 23.4_

- [x] 26. Implement PWA - Delivery Tasks Module
  - [x] 26.1 Implement Delivery Tasks list
    - Buat DeliveryTasksView showing assigned tasks untuk hari ini
    - Show school name, address, GPS, portions, menu items
    - Order by route sequence
    - Cache tasks di IndexedDB untuk offline access
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 23.1, 23.6_

  - [x] 26.2 Implement Delivery Task detail dengan navigation
    - Buat DeliveryTaskDetailView showing full task info
    - Implement button untuk open GPS navigation (Google Maps)
    - Show status dan allow start delivery
    - _Requirements: 11.2_

- [x] 27. Implement PWA - Electronic Proof of Delivery
  - [x] 27.1 Implement e-POD form dengan geotagging
    - Buat ePODFormView dengan auto-capture GPS coordinates
    - Show GPS coordinates dan accuracy
    - Input fields untuk ompreng drop-off dan pick-up quantities
    - _Requirements: 12.1, 12.2_

  - [x] 27.2 Implement photo capture untuk e-POD
    - Implement camera access menggunakan MediaDevices API
    - Allow capture photo of handover moment
    - Store photo locally jika offline
    - _Requirements: 12.3_

  - [x] 27.3 Implement digital signature capture
    - Implement signature pad untuk school representative
    - Allow clear dan re-sign
    - Store signature locally jika offline
    - _Requirements: 12.4_

  - [x] 27.4 Implement e-POD submission dan sync
    - Submit e-POD dengan photo, signature, GPS, timestamp
    - Update delivery status ke "Selesai"
    - If offline, store di IndexedDB dan queue untuk sync
    - If online, upload immediately ke backend
    - Show sync status (pending, syncing, synced)
    - _Requirements: 12.5, 12.6, 23.2, 23.3_

  - [ ]* 27.5 Write property test untuk offline sync
    - **Property 27: Offline Data Sync Completeness**
    - **Validates: Requirements 23.3**

- [x] 28. Implement PWA - Attendance Module
  - [x] 28.1 Implement Attendance check-in/out
    - Buat AttendanceView dengan check-in dan check-out buttons
    - Validate Wi-Fi connection (SSID dan BSSID) sebelum allow check-in
    - Show error message dalam Bahasa Indonesia jika Wi-Fi invalid
    - Record timestamp dan employee ID
    - Calculate work hours saat check-out
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_

  - [x] 28.2 Implement Wi-Fi detection
    - Implement Wi-Fi detection menggunakan browser APIs (limited support)
    - Fallback: allow manual SSID input atau GPS-based validation
    - _Requirements: 15.1, 15.2_

- [x] 29. Implement PWA - Offline Sync Service
  - [x] 29.1 Implement IndexedDB storage service
    - Buat IndexedDB schema untuk cache delivery tasks, schools, e-POD data
    - Implement methods: saveTask, getTasks, saveePOD, getPendingePODs
    - _Requirements: 23.1, 23.2, 23.6_

  - [x] 29.2 Implement sync service
    - Buat SyncService dengan method syncPendingData
    - Detect online/offline status
    - Auto-sync saat connection restored
    - Handle sync conflicts (server data wins)
    - Show sync progress dan status
    - _Requirements: 11.5, 23.3, 23.5_

- [x] 30. Checkpoint - PWA Complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 31. Integration Testing & End-to-End Workflows
  - [x] 31.1 Write integration tests untuk production workflow
    - Test complete flow: Menu planning → Cooking → Packing → Delivery
    - Verify data consistency across modules
    - _Requirements: 2.1-5.6_

  - [x] 31.2 Write integration tests untuk procurement workflow
    - Test complete flow: PO creation → Approval → GRN → Inventory update → Cash flow entry
    - Verify automatic triggers dan data propagation
    - _Requirements: 7.1-9.6, 17.4_

  - [x]* 31.3 Write integration tests untuk offline-online cycle
    - Test PWA offline data capture → Sync → Backend verification
    - Test conflict resolution
    - _Requirements: 23.1-23.6_

- [x] 32. Security Hardening & Performance Optimization
  - [x] 32.1 Implement security measures
    - Add rate limiting ke authentication endpoints
    - Implement CSRF protection
    - Add input sanitization dan validation
    - Setup HTTPS/TLS untuk all communications
    - _Requirements: 25.1, 25.2, 25.3, 25.4, 25.5, 25.6_

  - [x] 32.2 Optimize database queries
    - Add indexes untuk frequently queried columns
    - Optimize N+1 queries dengan eager loading
    - Implement connection pooling
    - _Requirements: Performance_

  - [x] 32.3 Implement caching strategy
    - Setup Redis untuk cache frequently accessed data
    - Cache dashboard metrics (refresh every 5 minutes)
    - Implement cache invalidation on updates
    - _Requirements: Performance_

- [x] 33. Deployment & Documentation
  - [x] 33.1 Setup deployment infrastructure
    - Configure Cloud SQL dengan HA dan replica
    - Setup load balancer untuk API servers
    - Configure CDN untuk static assets
    - Setup backup automation
    - _Requirements: 26.1, 26.2, 26.3, 26.4, 26.5, 26.6_

  - [x] 33.2 Create deployment documentation
    - Document environment setup
    - Document database migration procedures
    - Document backup dan recovery procedures
    - Document monitoring dan alerting setup
    - _Requirements: 26.5_

  - [x] 33.3 Create user documentation dalam Bahasa Indonesia
    - Create user manual untuk setiap role
    - Create video tutorials untuk key workflows
    - Create FAQ document
    - _Requirements: 24.1-24.6_

- [x] 34. Final Checkpoint - System Complete
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional property-based tests and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Property tests validate universal correctness properties across randomized inputs
- Unit tests validate specific examples and edge cases
- Integration tests validate end-to-end workflows
- All UI text must be in professional Bahasa Indonesia
- Real-time updates use Firebase for push notifications
- PWA must work offline with automatic sync when online
- Security is critical - implement RBAC, audit trail, encryption throughout
