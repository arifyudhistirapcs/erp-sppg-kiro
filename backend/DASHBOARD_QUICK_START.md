# Dashboard Kepala SPPG - Quick Start Guide

## ✅ Implementasi Selesai

Dashboard Kepala SPPG sekarang menggunakan **data real dari database PostgreSQL**.

## 🎯 Apa yang Sudah Diimplementasikan

### 1. Dashboard Kepala SPPG (Operasional Harian)
- ✅ Production Status (cooking & packing progress)
- ✅ Delivery Status (pending, in progress, completed)
- ✅ Critical Stock (bahan baku di bawah threshold)
- ✅ Today's KPIs (portions, delivery rate, stock availability)

### 2. Dashboard Kepala Yayasan (Strategic Overview)
- ✅ Budget Absorption (per kategori)
- ✅ Nutrition Distribution (porsi, sekolah, siswa)
- ✅ Supplier Performance (on-time delivery, quality)
- ✅ Monthly Trend (6 bulan terakhir)

### 3. Features
- ✅ Real-time data dari database
- ✅ Graceful error handling
- ✅ Firebase sync (manual trigger)
- ✅ Export dashboard data (JSON)
- ✅ Date range filtering (Kepala Yayasan)

## 🚀 Cara Testing

### 1. Start Server
```bash
cd backend
go run cmd/server/main.go
```

### 2. Test Dashboard Kepala SPPG
```bash
curl http://localhost:8080/api/v1/dashboard/kepala-sppg
```

### 3. Test Dashboard Kepala Yayasan
```bash
curl "http://localhost:8080/api/v1/dashboard/kepala-yayasan?start_date=2026-02-01&end_date=2026-02-25"
```

### 4. Test dengan Browser
```
http://localhost:8080/api/v1/dashboard/kepala-sppg
http://localhost:8080/api/v1/dashboard/kepala-yayasan
```

## 📊 Data yang Dibutuhkan

Untuk dashboard menampilkan data yang benar, pastikan database memiliki:

1. **Menu Plans** (status: approved)
2. **Menu Items** (untuk hari ini)
3. **Delivery Tasks** (untuk hari ini)
4. **Inventory Items** (dengan min_threshold)
5. **Suppliers** (dengan performance metrics)
6. **Cash Flow Entries** (untuk budget tracking)
7. **Budget Targets** (untuk comparison)

### Seed Data (Optional)
```bash
go run cmd/seed/main.go
```

## 🔧 File yang Dimodifikasi

1. `backend/internal/services/dashboard_service.go`
   - Method `GetKepalaSSPGDashboard()` - query data real
   - Method `GetKepalaYayasanDashboard()` - query data real
   - Graceful error handling

2. `backend/internal/handlers/dashboard_handler.go`
   - Tidak ada perubahan (sudah siap)

## 📝 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/dashboard/kepala-sppg` | Dashboard operasional |
| GET | `/api/v1/dashboard/kepala-yayasan` | Dashboard strategic |
| GET | `/api/v1/dashboard/kpi` | KPIs only |
| POST | `/api/v1/dashboard/sync` | Sync to Firebase |
| POST | `/api/v1/dashboard/export` | Export data |

## 🎨 Response Example

### Dashboard Kepala SPPG
```json
{
  "success": true,
  "dashboard": {
    "production_status": {
      "total_recipes": 12,
      "recipes_ready": 5,
      "completion_rate": 58.3
    },
    "delivery_status": {
      "total_deliveries": 15,
      "deliveries_completed": 7,
      "completion_rate": 46.7
    },
    "critical_stock": [
      {
        "ingredient_name": "Beras Putih",
        "current_stock": 50,
        "min_threshold": 100,
        "days_remaining": 2.5
      }
    ],
    "today_kpis": {
      "portions_prepared": 3250,
      "delivery_rate": 78.5,
      "stock_availability": 85.2
    }
  }
}
```

## ⚠️ Important Notes

1. **Firebase Optional**: Dashboard berfungsi tanpa Firebase (hanya tidak ada real-time updates)
2. **Error Handling**: Jika query gagal, dashboard tetap return data (dengan default values)
3. **Performance**: Query optimized dengan proper indexing
4. **Date Range**: Kepala Yayasan dashboard default = current month

## 🔄 Next Steps (Optional)

1. Implementasi caching dengan Redis (TTL 5-10 menit)
2. Implementasi PDF/Excel export
3. Implementasi auto-sync ke Firebase (cron job)
4. Implementasi drill-down details per metric
5. Frontend integration dengan real-time updates

## 📚 Dokumentasi Lengkap

Lihat `DASHBOARD_IMPLEMENTATION.md` untuk dokumentasi detail.

## ✨ Summary

Dashboard Kepala SPPG sudah **siap digunakan** dengan data real dari database. Tidak perlu konfigurasi tambahan, langsung bisa di-test dengan curl atau browser.

