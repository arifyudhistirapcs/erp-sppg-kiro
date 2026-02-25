# Dashboard Real Data Implementation Guide

## Overview
This guide explains how the Kepala SPPG Dashboard retrieves real data from the database and how to verify it's working correctly.

## Changes Made

### 1. Enhanced Dashboard Service (`backend/internal/services/dashboard_service.go`)

#### Production Status Query
- **Fixed**: Changed from `DATE(menu_items.date) = DATE(?)` to date range comparison
- **Why**: Better database compatibility and timezone handling
- **Added**: Debug logging to track query results
- **Query**: Finds approved menu items for today from `menu_items` and `menu_plans` tables

#### Delivery Status Query
- **Fixed**: Improved date handling with explicit timezone
- **Added**: Debug logging for delivery task counts
- **Query**: Retrieves all delivery tasks for today with status breakdown

#### Critical Stock Query
- **Added**: Debug logging for low stock alerts
- **Query**: Uses `InventoryService.CheckLowStock()` to find items below threshold

#### Today's KPIs Calculation
- **Fixed**: Date range queries for better accuracy
- **Added**: Comprehensive debug logging for all KPI metrics
- **Queries**:
  - Portions prepared: Sum of portions from approved menu items
  - Delivery rate: Percentage of completed deliveries
  - Stock availability: Percentage of items above minimum threshold
  - On-time delivery: Calculated from completed deliveries

## Data Flow

```
Frontend Request
    ↓
GET /api/dashboard/kepala-sppg
    ↓
DashboardHandler.GetKepalaSSPGDashboard()
    ↓
DashboardService.GetKepalaSSPGDashboard()
    ↓
├── getProductionStatus() → menu_items + menu_plans tables + Firebase KDS
├── getDeliveryStatus() → delivery_tasks table
├── getCriticalStock() → inventory_items + ingredients tables
└── calculateTodayKPIs() → menu_items, delivery_tasks, inventory_items tables
    ↓
JSON Response to Frontend
```

## Database Tables Used

### Production Status
- **Tables**: `menu_items`, `menu_plans`
- **Conditions**: 
  - `menu_plans.status = 'approved'`
  - `menu_items.date` = today
- **Firebase**: `/kds/cooking/{date}` and `/kds/packing/{date}` for real-time status

### Delivery Status
- **Tables**: `delivery_tasks`, `users`, `schools`
- **Conditions**: `task_date` = today
- **Status values**: `pending`, `in_progress`, `completed`, `cancelled`

### Critical Stock
- **Tables**: `inventory_items`, `ingredients`
- **Conditions**: `quantity < min_threshold`

### Today's KPIs
- **Portions Prepared**: Sum from `menu_items` where date = today
- **Delivery Rate**: Count of completed vs total `delivery_tasks`
- **Stock Availability**: Count of `inventory_items` where `quantity >= min_threshold`
- **On-Time Delivery**: Calculated from completed deliveries (95% default)

## Verification Steps

### 1. Check if Database Has Seed Data

```bash
# Run the seed command to populate test data
cd backend
go run cmd/seed/main.go
```

This will create:
- 10 users (including drivers)
- 18 ingredients
- 16 semi-finished goods
- 6 recipes
- 6 suppliers
- Menu plan for current week
- Delivery tasks for today
- Inventory items

### 2. Check Backend Logs

When the dashboard API is called, you should see logs like:

```
Dashboard: Found X menu items for today
Dashboard: Found X delivery tasks for today
Dashboard: Found X critical stock items
Dashboard: Portions prepared today: X
Dashboard: Delivery rate: X% (X completed out of X)
Dashboard: Stock availability: X% (X items above threshold)
```

### 3. Verify Data in Database

```sql
-- Check menu items for today
SELECT COUNT(*) FROM menu_items mi
JOIN menu_plans mp ON mi.menu_plan_id = mp.id
WHERE mp.status = 'approved'
AND DATE(mi.date) = CURDATE();

-- Check delivery tasks for today
SELECT COUNT(*), status FROM delivery_tasks
WHERE DATE(task_date) = CURDATE()
GROUP BY status;

-- Check critical stock
SELECT COUNT(*) FROM inventory_items
WHERE quantity < min_threshold;

-- Check total inventory
SELECT COUNT(*) FROM inventory_items;
```

### 4. Test API Endpoint

```bash
# Get authentication token first
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"kepala.sppg@sppg.com","password":"password123"}'

# Use the token to access dashboard
curl -X GET http://localhost:8080/api/dashboard/kepala-sppg \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

Expected response structure:
```json
{
  "production_status": {
    "total_recipes": 5,
    "recipes_pending": 5,
    "recipes_cooking": 0,
    "recipes_ready": 0,
    "packing_pending": 0,
    "packing_in_progress": 0,
    "packing_ready": 0,
    "completion_rate": 0
  },
  "delivery_status": {
    "total_deliveries": 6,
    "deliveries_pending": 2,
    "deliveries_in_progress": 2,
    "deliveries_completed": 2,
    "completion_rate": 33.33
  },
  "critical_stock": [
    {
      "ingredient_id": 1,
      "ingredient_name": "Beras Putih",
      "current_stock": 45.5,
      "min_threshold": 50.0,
      "unit": "kg",
      "days_remaining": 3.2
    }
  ],
  "today_kpis": {
    "portions_prepared": 3500,
    "delivery_rate": 33.33,
    "stock_availability": 85.5,
    "on_time_delivery_rate": 95.0
  },
  "updated_at": "2026-02-25T10:30:00Z"
}
```

## Troubleshooting

### Dashboard Shows All Zeros

**Possible Causes:**
1. No seed data in database
2. Menu plan not created for current week
3. Delivery tasks not created for today
4. Date/timezone mismatch

**Solutions:**
1. Run seed command: `go run cmd/seed/main.go`
2. Check backend logs for query errors
3. Verify database has data for today's date
4. Check server timezone matches database timezone

### Production Status Shows All Pending

**Cause:** Firebase KDS data not available or not synced

**Solution:** This is expected if:
- Firebase is not configured
- KDS system hasn't been used yet
- No cooking/packing activities recorded

The system gracefully falls back to showing all recipes as "pending" when Firebase data is unavailable.

### Critical Stock Always Empty

**Possible Causes:**
1. All inventory items are above threshold
2. Inventory not initialized

**Solutions:**
1. Check inventory: `SELECT * FROM inventory_items WHERE quantity < min_threshold;`
2. Manually adjust some inventory quantities below threshold for testing
3. Run seed again to reset data

### Delivery Status Shows Wrong Date

**Cause:** Timezone mismatch between application and database

**Solution:**
1. Check server timezone: `date`
2. Check database timezone: `SELECT @@global.time_zone, @@session.time_zone;`
3. Ensure both use same timezone or UTC

## Production Considerations

### Performance Optimization

1. **Add Database Indexes:**
```sql
CREATE INDEX idx_menu_items_date ON menu_items(date);
CREATE INDEX idx_delivery_tasks_date_status ON delivery_tasks(task_date, status);
CREATE INDEX idx_inventory_threshold ON inventory_items(quantity, min_threshold);
```

2. **Enable Query Caching:**
- Dashboard data can be cached for 1-5 minutes
- Use Redis cache for frequently accessed data
- Implement cache invalidation on data updates

3. **Optimize Queries:**
- Use `EXPLAIN` to analyze query performance
- Consider materialized views for complex aggregations
- Batch related queries when possible

### Monitoring

1. **Add Metrics:**
- Query execution time
- Cache hit/miss rates
- API response time
- Error rates

2. **Set Up Alerts:**
- Slow query alerts (> 1 second)
- High error rate alerts
- Cache miss rate alerts

3. **Log Analysis:**
- Track dashboard access patterns
- Monitor data freshness
- Identify performance bottlenecks

## Next Steps

1. **Real-time Updates:**
   - Implement WebSocket for live dashboard updates
   - Sync Firebase KDS data automatically
   - Push notifications for critical events

2. **Historical Data:**
   - Add date range filters
   - Implement trend analysis
   - Create comparison views (today vs yesterday)

3. **Advanced Analytics:**
   - Predictive analytics for stock levels
   - Delivery route optimization metrics
   - Production efficiency trends

4. **Export Features:**
   - PDF report generation
   - Excel export with charts
   - Scheduled email reports

## Related Files

- `backend/internal/services/dashboard_service.go` - Main dashboard logic
- `backend/internal/handlers/dashboard_handler.go` - API endpoints
- `backend/internal/services/inventory_service.go` - Stock management
- `backend/internal/services/delivery_task_service.go` - Delivery operations
- `backend/cmd/seed/main.go` - Test data generation
- `web/src/views/DashboardKepalaSSPGView.vue` - Frontend component
- `web/src/services/dashboardService.js` - API client

## Support

For issues or questions:
1. Check backend logs for error messages
2. Verify database connectivity
3. Ensure seed data is present
4. Review this guide's troubleshooting section
5. Check related documentation files
