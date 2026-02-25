# Dashboard Implementation - Data Real dari Database

## Perubahan yang Dilakukan

Dashboard Kepala SPPG dan Kepala Yayasan sekarang menggunakan **data real dari database** PostgreSQL, bukan lagi dummy data.

### File yang Dimodifikasi

1. **backend/internal/services/dashboard_service.go**
   - Method `GetKepalaSSPGDashboard()` - sekarang query data real dari database
   - Method `GetKepalaYayasanDashboard()` - sekarang query data real dari database
   - Menambahkan error handling graceful (jika query gagal, gunakan default values)
   - Firebase sync sekarang memvalidasi client availability

### Data yang Ditampilkan

#### Dashboard Kepala SPPG (Operasional Harian)

**Production Status:**
- Total resep hari ini (dari `menu_items` dengan status approved)
- Status cooking (pending, cooking, ready) - dari Firebase `/kds/cooking/{date}`
- Status packing (pending, packing, ready) - dari Firebase `/kds/packing/{date}`
- Completion rate (persentase resep yang sudah ready)

**Delivery Status:**
- Total delivery tasks hari ini
- Status delivery (pending, in_progress, completed)
- Completion rate delivery

**Critical Stock:**
- Bahan baku yang stoknya di bawah minimum threshold
- Menampilkan: nama bahan, stok saat ini, minimum threshold, unit, estimasi hari tersisa

**Today's KPIs:**
- Total porsi yang disiapkan hari ini
- Delivery completion rate
- Stock availability percentage
- On-time delivery rate

#### Dashboard Kepala Yayasan (Strategic Overview)

**Budget Absorption:**
- Total budget vs actual spending
- Breakdown per kategori (bahan_baku, gaji, utilitas, operasional)
- Absorption rate per kategori

**Nutrition Distribution:**
- Total porsi yang didistribusikan
- Jumlah sekolah yang dilayani
- Jumlah siswa yang dijangkau
- Rata-rata porsi per sekolah

**Supplier Performance:**
- Total suppliers (active vs inactive)
- Average on-time delivery rate
- Average quality rating

**Monthly Trend:**
- Trend bulanan: porsi didistribusikan, budget spent, sekolah dilayani
- Data untuk periode yang dipilih

## API Endpoints

### 1. Get Dashboard Kepala SPPG
```bash
GET /api/v1/dashboard/kepala-sppg
```

**Response:**
```json
{
  "success": true,
  "dashboard": {
    "production_status": {
      "total_recipes": 12,
      "recipes_pending": 2,
      "recipes_cooking": 5,
      "recipes_ready": 5,
      "packing_pending": 2,
      "packing_in_progress": 3,
      "packing_ready": 7,
      "completion_rate": 58.3
    },
    "delivery_status": {
      "total_deliveries": 15,
      "deliveries_pending": 3,
      "deliveries_in_progress": 5,
      "deliveries_completed": 7,
      "completion_rate": 46.7
    },
    "critical_stock": [
      {
        "ingredient_id": 1,
        "ingredient_name": "Beras Putih",
        "current_stock": 50,
        "min_threshold": 100,
        "unit": "kg",
        "days_remaining": 2.5
      }
    ],
    "today_kpis": {
      "portions_prepared": 3250,
      "delivery_rate": 78.5,
      "stock_availability": 85.2,
      "on_time_delivery_rate": 92.3
    },
    "updated_at": "2026-02-25T10:30:00Z"
  }
}
```

### 2. Get Dashboard Kepala Yayasan
```bash
GET /api/v1/dashboard/kepala-yayasan?start_date=2026-02-01&end_date=2026-02-25
```

**Query Parameters:**
- `start_date` (optional): Format YYYY-MM-DD, default = first day of current month
- `end_date` (optional): Format YYYY-MM-DD, default = today

**Response:**
```json
{
  "success": true,
  "dashboard": {
    "budget_absorption": {
      "total_budget": 5000000000,
      "total_spent": 3750000000,
      "absorption_rate": 75.0,
      "category_breakdown": [
        {
          "category": "bahan_baku",
          "budget": 3000000000,
          "spent": 2400000000,
          "absorption_rate": 80.0
        }
      ]
    },
    "nutrition_distribution": {
      "total_portions_distributed": 45000,
      "schools_served": 15,
      "students_reached": 3250,
      "average_portions_per_school": 3000
    },
    "supplier_performance": {
      "total_suppliers": 12,
      "active_suppliers": 10,
      "avg_on_time_delivery": 88.5,
      "avg_quality_rating": 4.2
    },
    "monthly_trend": [
      {
        "month": "Januari",
        "year": 2026,
        "portions_distributed": 42000,
        "budget_spent": 350000000,
        "schools_served": 14
      }
    ],
    "updated_at": "2026-02-25T10:30:00Z"
  },
  "start_date": "2026-02-01",
  "end_date": "2026-02-25"
}
```

### 3. Get KPIs Only
```bash
GET /api/v1/dashboard/kpi
```

### 4. Sync Dashboard to Firebase
```bash
POST /api/v1/dashboard/sync?type=kepala_sppg
POST /api/v1/dashboard/sync?type=kepala_yayasan&start_date=2026-02-01&end_date=2026-02-25
```

### 5. Export Dashboard
```bash
POST /api/v1/dashboard/export
Content-Type: application/json

{
  "type": "kepala_sppg",
  "format": "json"
}
```

## Testing Dashboard

### Prerequisites

Pastikan database sudah memiliki data:
1. Users dengan role yang sesuai
2. Menu plans yang approved
3. Menu items untuk hari ini
4. Delivery tasks
5. Inventory items dengan threshold
6. Suppliers
7. Cash flow entries
8. Budget targets

### Test dengan cURL

```bash
# 1. Test Dashboard Kepala SPPG
curl -X GET http://localhost:8080/api/v1/dashboard/kepala-sppg \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 2. Test Dashboard Kepala Yayasan
curl -X GET "http://localhost:8080/api/v1/dashboard/kepala-yayasan?start_date=2026-02-01&end_date=2026-02-25" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 3. Test KPIs
curl -X GET http://localhost:8080/api/v1/dashboard/kpi \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 4. Test Sync to Firebase (jika Firebase configured)
curl -X POST "http://localhost:8080/api/v1/dashboard/sync?type=kepala_sppg" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 5. Test Export
curl -X POST http://localhost:8080/api/v1/dashboard/export \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "kepala_sppg",
    "format": "json"
  }'
```

### Test dengan Postman

1. Import collection dari dokumentasi API
2. Set environment variable untuk JWT token
3. Test setiap endpoint di atas

## Error Handling

Dashboard service menggunakan **graceful error handling**:

- Jika query production status gagal → gunakan default values (0)
- Jika query delivery status gagal → gunakan default values (0)
- Jika query critical stock gagal → return empty array
- Jika query KPIs gagal → gunakan default values (0)
- Jika Firebase tidak tersedia → log warning, lanjutkan dengan data dari database

Ini memastikan dashboard tetap bisa ditampilkan meskipun ada error di beberapa bagian.

## Firebase Real-time Updates

Jika Firebase configured, dashboard akan:

1. **Production Status** - membaca dari `/kds/cooking/{date}` dan `/kds/packing/{date}`
2. **Manual Sync** - bisa trigger sync manual via endpoint `/api/v1/dashboard/sync`
3. **Auto Sync** - bisa diimplementasikan dengan cron job atau event trigger

### Firebase Data Structure

```
/dashboard
  /kepala_sppg
    - production_status: {...}
    - delivery_status: {...}
    - critical_stock: [...]
    - today_kpis: {...}
    - updated_at: timestamp
    
  /kepala_yayasan
    - budget_absorption: {...}
    - nutrition_distribution: {...}
    - supplier_performance: {...}
    - monthly_trend: [...]
    - updated_at: timestamp
```

## Performance Considerations

1. **Caching**: Dashboard data bisa di-cache dengan Redis (TTL 5-10 menit)
2. **Indexing**: Pastikan index database sudah optimal untuk query dashboard
3. **Pagination**: Monthly trend bisa di-limit untuk performa
4. **Background Jobs**: Sync ke Firebase bisa dilakukan di background

## Next Steps

1. ✅ Implementasi query data real dari database
2. ✅ Error handling graceful
3. ✅ Firebase sync validation
4. 🔲 Implementasi caching dengan Redis
5. 🔲 Implementasi PDF/Excel export
6. 🔲 Implementasi auto-sync ke Firebase (cron job)
7. 🔲 Implementasi drill-down details per metric
8. 🔲 Implementasi real-time updates di frontend

## Troubleshooting

### Dashboard menampilkan data kosong

**Penyebab:**
- Database belum ada data seed
- Menu plan belum approved
- Delivery tasks belum dibuat

**Solusi:**
```bash
# Run seed data
go run cmd/seed/main.go
```

### Firebase sync error

**Penyebab:**
- Firebase credentials tidak valid
- Firebase database URL tidak configured

**Solusi:**
- Check `firebase-credentials.json`
- Check environment variable `FIREBASE_DATABASE_URL`
- Dashboard tetap berfungsi tanpa Firebase (hanya tidak ada real-time updates)

### Query lambat

**Penyebab:**
- Database belum ada index
- Data terlalu banyak tanpa pagination

**Solusi:**
- Run migration untuk create indexes
- Implementasi caching dengan Redis
- Limit date range untuk query

