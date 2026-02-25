# Dashboard Quick Fix Reference

## 🚀 Quick Start (3 Steps)

### 1. Seed Database
```bash
cd backend
go run cmd/seed/main.go
```

### 2. Verify Data
```bash
./scripts/verify_dashboard_data.sh
```

### 3. Start Server
```bash
go run cmd/server/main.go
```

## ✅ What Was Fixed

| Issue | Solution |
|-------|----------|
| Date queries not working | Changed to date range comparison |
| Timezone mismatches | Explicit timezone handling |
| No visibility into queries | Added debug logging |
| Silent failures | Better error handling |

## 📊 Expected Dashboard Data

### Production Status
```json
{
  "total_recipes": 5,        // From menu_items table
  "recipes_pending": 5,      // From Firebase or default
  "recipes_cooking": 0,      // From Firebase
  "recipes_ready": 0,        // From Firebase
  "completion_rate": 0       // Calculated
}
```

### Delivery Status
```json
{
  "total_deliveries": 6,     // From delivery_tasks table
  "deliveries_pending": 2,   // Count by status
  "deliveries_in_progress": 2,
  "deliveries_completed": 2,
  "completion_rate": 33.33   // Calculated
}
```

### Critical Stock
```json
[
  {
    "ingredient_name": "Beras Putih",
    "current_stock": 45.5,
    "min_threshold": 50.0,
    "unit": "kg",
    "days_remaining": 3.2
  }
]
```

### Today's KPIs
```json
{
  "portions_prepared": 3500,    // Sum from menu_items
  "delivery_rate": 33.33,       // % completed deliveries
  "stock_availability": 85.5,   // % items above threshold
  "on_time_delivery_rate": 95.0 // Default (needs tracking)
}
```

## 🔍 Debug Logs to Look For

When dashboard API is called, you should see:
```
Dashboard: Found X menu items for today
Dashboard: Found X delivery tasks for today
Dashboard: Found X critical stock items
Dashboard: Portions prepared today: X
Dashboard: Delivery rate: X% (X completed out of X)
Dashboard: Stock availability: X% (X items above threshold)
```

## ⚠️ Common Issues

### All Zeros in Dashboard

**Cause:** No data for today

**Fix:**
```bash
go run cmd/seed/main.go
```

### No Menu Items

**Cause:** Seed creates data for weekdays only

**Fix:** Run seed on Monday-Friday, or create manual menu plan

### Firebase Errors

**Cause:** Firebase not configured

**Fix:** This is OK! System works without Firebase. All recipes show as "pending".

## 🧪 Test API Directly

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"kepala.sppg@sppg.com","password":"password123"}' \
  | jq -r '.token')

# 2. Get Dashboard
curl -X GET http://localhost:8080/api/dashboard/kepala-sppg \
  -H "Authorization: Bearer $TOKEN" | jq
```

## 📁 Files Changed

- ✅ `backend/internal/services/dashboard_service.go` - Fixed queries
- ✅ `backend/DASHBOARD_REAL_DATA_GUIDE.md` - Full guide
- ✅ `backend/DASHBOARD_FIX_SUMMARY.md` - Detailed summary
- ✅ `backend/scripts/verify_dashboard_data.sh` - Verification script

## 📚 Documentation

| File | Purpose |
|------|---------|
| `DASHBOARD_QUICK_FIX.md` | This file - quick reference |
| `DASHBOARD_FIX_SUMMARY.md` | Detailed changes and testing |
| `DASHBOARD_REAL_DATA_GUIDE.md` | Complete implementation guide |
| `DASHBOARD_IMPLEMENTATION.md` | Original implementation docs |
| `DASHBOARD_QUICK_START.md` | Quick start guide |

## ✨ Success Criteria

- [ ] Seed script runs successfully
- [ ] Verification script shows data
- [ ] Backend logs show query results
- [ ] API returns non-zero values
- [ ] Frontend displays real data

## 🆘 Need Help?

1. Check `DASHBOARD_FIX_SUMMARY.md` for troubleshooting
2. Review `DASHBOARD_REAL_DATA_GUIDE.md` for details
3. Run verification script: `./scripts/verify_dashboard_data.sh`
4. Check backend logs for errors
5. Verify database has data for TODAY

---

**Status:** ✅ Ready for Testing

**Last Updated:** 2026-02-25
