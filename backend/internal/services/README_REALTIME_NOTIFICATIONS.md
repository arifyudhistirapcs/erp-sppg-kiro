# Real-Time Sync & Notification Services

## Overview

Implementasi layanan sinkronisasi real-time dengan Firebase dan sistem notifikasi untuk ERP SPPG.

## Firebase Sync Service

### Fitur
- Push data ke Firebase Realtime Database
- Update dengan timestamp otomatis
- Hapus data dari Firebase
- Ambil data dari Firebase
- Update field spesifik
- Resolusi konflik (server data wins)

### Penggunaan

```go
// Inisialisasi service
firebaseSyncService, err := services.NewFirebaseSyncService(firebaseApp)

// Push update ke path tertentu
err = firebaseSyncService.PushUpdate(ctx, "/path/to/data", data)

// Push update dengan timestamp
err = firebaseSyncService.PushUpdateWithTimestamp(ctx, "/path/to/data", data)

// Update KDS cooking status
err = firebaseSyncService.PushKDSCookingUpdate(ctx, "2024-01-15", recipeID, cookingData)

// Update KDS packing status
err = firebaseSyncService.PushKDSPackingUpdate(ctx, "2024-01-15", schoolID, packingData)

// Update dashboard
err = firebaseSyncService.PushDashboardUpdate(ctx, "kepala_sppg", dashboardData)

// Update inventory
err = firebaseSyncService.PushInventoryUpdate(ctx, ingredientID, inventoryData)

// Update delivery status
err = firebaseSyncService.PushDeliveryUpdate(ctx, taskID, deliveryData)
```

### Firebase Path Structure

```
/kds
  /cooking
    /{date}
      /{recipe_id}
        - name: string
        - status: string
        - start_time: timestamp
        - portions_required: number
        - updated_at: timestamp
  /packing
    /{date}
      /{school_id}
        - school_name: string
        - portions: number
        - menu_items: array
        - status: string
        - updated_at: timestamp

/dashboard
  /kepala_sppg
    - production_status: object
    - delivery_status: object
    - critical_stock: array
    - updated_at: timestamp
  /kepala_yayasan
    - budget_absorption: number
    - total_portions_distributed: number
    - schools_served: number
    - supplier_performance: object
    - updated_at: timestamp

/inventory
  /{ingredient_id}
    - quantity: number
    - min_threshold: number
    - last_updated: timestamp
    - updated_at: timestamp

/delivery
  /{task_id}
    - status: string
    - school_name: string
    - driver_name: string
    - updated_at: timestamp

/notifications
  /{user_id}
    /{notification_id}
      - id: number
      - type: string
      - title: string
      - message: string
      - is_read: boolean
      - link: string
      - created_at: timestamp
```

## Notification Service

### Fitur
- Buat notifikasi baru
- Kirim notifikasi untuk berbagai event:
  - Low stock alert
  - PO approval request
  - Packing complete
  - Delivery complete
- Ambil notifikasi user dengan pagination
- Hitung notifikasi yang belum dibaca
- Tandai notifikasi sebagai dibaca
- Tandai semua notifikasi sebagai dibaca
- Hapus notifikasi
- Sinkronisasi otomatis dengan Firebase

### Tipe Notifikasi

- `low_stock` - Peringatan stok menipis
- `po_approval` - Permintaan persetujuan PO
- `packing_complete` - Packing selesai
- `delivery_complete` - Pengiriman selesai

### Penggunaan

```go
// Inisialisasi service
notificationService, err := services.NewNotificationService(db, firebaseApp)

// Kirim notifikasi low stock
err = notificationService.SendLowStockNotification(
    ctx, 
    userID, 
    "Beras", 
    50.0,  // current quantity
    100.0, // min threshold
)

// Kirim notifikasi PO approval
err = notificationService.SendPOApprovalNotification(
    ctx,
    userID,
    "PO-2024-001",
    "PT Supplier ABC",
    5000000.0,
)

// Kirim notifikasi packing complete
err = notificationService.SendPackingCompleteNotification(
    ctx,
    userID,
    "2024-01-15",
    25, // total schools
)

// Kirim notifikasi delivery complete
err = notificationService.SendDeliveryCompleteNotification(
    ctx,
    userID,
    "SDN 01 Jakarta",
    "Budi Santoso",
)

// Ambil notifikasi user
notifications, total, err := notificationService.GetUserNotifications(userID, 20, 0)

// Hitung notifikasi yang belum dibaca
count, err := notificationService.GetUnreadCount(userID)

// Tandai sebagai dibaca
err = notificationService.MarkAsRead(ctx, notificationID, userID)

// Tandai semua sebagai dibaca
err = notificationService.MarkAllAsRead(ctx, userID)

// Hapus notifikasi
err = notificationService.DeleteNotification(ctx, notificationID, userID)
```

## API Endpoints

### Notification Endpoints

```
GET    /api/v1/notifications              - Ambil notifikasi user (dengan pagination)
GET    /api/v1/notifications/unread-count - Hitung notifikasi yang belum dibaca
PUT    /api/v1/notifications/:id/read     - Tandai notifikasi sebagai dibaca
PUT    /api/v1/notifications/read-all     - Tandai semua notifikasi sebagai dibaca
DELETE /api/v1/notifications/:id          - Hapus notifikasi
```

### Request Examples

#### Get Notifications
```bash
GET /api/v1/notifications?page=1&limit=20
Authorization: Bearer <jwt_token>
```

Response:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "user_id": 5,
      "type": "low_stock",
      "title": "Peringatan Stok Menipis",
      "message": "Stok Beras menipis. Jumlah saat ini: 50.00, batas minimum: 100.00",
      "is_read": false,
      "link": "/inventory",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1
  }
}
```

#### Get Unread Count
```bash
GET /api/v1/notifications/unread-count
Authorization: Bearer <jwt_token>
```

Response:
```json
{
  "success": true,
  "data": {
    "unread_count": 5
  }
}
```

#### Mark as Read
```bash
PUT /api/v1/notifications/1/read
Authorization: Bearer <jwt_token>
```

Response:
```json
{
  "success": true,
  "message": "Notifikasi berhasil ditandai sebagai dibaca"
}
```

## Integration dengan Modul Lain

### Inventory Service
Ketika stok menipis, kirim notifikasi ke staff Pengadaan:

```go
if item.Quantity < item.MinThreshold {
    // Get all Pengadaan staff
    var pengadaanUsers []models.User
    db.Where("role = ? AND is_active = ?", "pengadaan", true).Find(&pengadaanUsers)
    
    // Send notification to each
    for _, user := range pengadaanUsers {
        notificationService.SendLowStockNotification(
            ctx,
            user.ID,
            ingredient.Name,
            item.Quantity,
            item.MinThreshold,
        )
    }
}
```

### Purchase Order Service
Ketika PO dibuat, kirim notifikasi ke Kepala SPPG untuk approval:

```go
// Get Kepala SPPG
var kepalaSSPG models.User
db.Where("role = ? AND is_active = ?", "kepala_sppg", true).First(&kepalaSSPG)

// Send notification
notificationService.SendPOApprovalNotification(
    ctx,
    kepalaSSPG.ID,
    po.PONumber,
    supplier.Name,
    po.TotalAmount,
)
```

### KDS Service
Ketika packing selesai, kirim notifikasi ke driver:

```go
// Get all active drivers
var drivers []models.User
db.Where("role = ? AND is_active = ?", "driver", true).Find(&drivers)

// Send notification to each
for _, driver := range drivers {
    notificationService.SendPackingCompleteNotification(
        ctx,
        driver.ID,
        date,
        totalSchools,
    )
}
```

### Delivery Service
Ketika pengiriman selesai, kirim notifikasi ke staff logistik:

```go
// Get logistics staff
var logisticsStaff []models.User
db.Where("role IN ? AND is_active = ?", []string{"kepala_sppg", "asisten_lapangan"}, true).Find(&logisticsStaff)

// Send notification to each
for _, staff := range logisticsStaff {
    notificationService.SendDeliveryCompleteNotification(
        ctx,
        staff.ID,
        school.Name,
        driver.FullName,
    )
}
```

## Conflict Resolution

Service menggunakan strategi "server wins" untuk resolusi konflik:
- Ketika terjadi konflik antara data client dan server
- Data dari server selalu digunakan
- Client akan menerima update terbaru dari Firebase

## Error Handling

Semua error dikembalikan dalam format Indonesian:
- "gagal menginisialisasi Firebase Database client"
- "gagal mengirim update ke Firebase"
- "gagal membuat notifikasi"
- "notifikasi tidak ditemukan"

## Testing

Unit tests tersedia di `notification_service_test.go`:

```bash
go test -v ./internal/services -run TestNotification
```

Tests mencakup:
- Create notification
- Get user notifications
- Get unread count
- Mark as read
- Delete notification
- Notification types

## Requirements Validation

Implementasi ini memenuhi requirements:
- **22.1-22.6**: Real-time sync dengan Firebase
- **28.1-28.6**: Sistem notifikasi
- Semua error messages dalam Bahasa Indonesia profesional
- Firebase Admin SDK sudah diinisialisasi
- Push updates untuk real-time sync
- Handle connection state dan reconnection (di client side)
- Conflict resolution (server wins)
- CRUD operations untuk notifikasi
- Store notifications dengan read/unread status
- API endpoints terdaftar di router
