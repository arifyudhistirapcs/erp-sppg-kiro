# Pre-Validation Check Implementation

## Overview
Task 3.3 adds a pre-validation check to the `deductInventory` method that validates ALL recipe items for sufficient stock BEFORE starting the database transaction.

## Implementation Details

### Location
File: `backend/internal/services/kds_service.go`
Method: `deductInventory` (starting at line ~530)

### Key Changes

1. **Pre-Validation Loop** (lines ~545-597)
   - Iterates through all recipe items BEFORE starting transaction
   - Calculates required quantities based on portion sizes
   - Checks stock availability for each item
   - Collects ALL insufficient items in a list

2. **Detailed Error Message**
   - Format: `"Stok tidak mencukupi untuk: [item1] (butuh X, tersedia Y), [item2] (butuh X, tersedia Y)"`
   - Lists ALL insufficient items, not just the first one
   - Provides both needed and available quantities in Indonesian

3. **HTTP Error Handling** (in `backend/internal/handlers/kds_handler.go`)
   - Returns HTTP 400 (Bad Request) for insufficient stock errors
   - Error code: `INSUFFICIENT_STOCK`
   - Distinguishes stock errors from other server errors

## Example Scenarios

### Scenario 1: Multiple Insufficient Items
**Input:**
- Recipe requires: 50 kg Nasi, 30 kg Ayam Goreng, 20 kg Sayur
- Available stock: 20 kg Nasi, 15 kg Ayam Goreng, 25 kg Sayur

**Output:**
```
Error: Stok tidak mencukupi untuk: Nasi (butuh 50.00, tersedia 20.00), Ayam Goreng (butuh 30.00, tersedia 15.00)
```

**Behavior:**
- Pre-validation detects BOTH insufficient items
- Transaction is NOT started
- No partial stock deductions occur
- Sayur is not mentioned (sufficient stock)

### Scenario 2: Single Insufficient Item
**Input:**
- Recipe requires: 10 kg Nasi, 15 kg Ayam Goreng
- Available stock: 5 kg Nasi, 20 kg Ayam Goreng

**Output:**
```
Error: Stok tidak mencukupi untuk: Nasi (butuh 10.00, tersedia 5.00)
```

**Behavior:**
- Pre-validation detects one insufficient item
- Transaction is NOT started
- No stock deductions occur

### Scenario 3: All Items Sufficient
**Input:**
- Recipe requires: 10 kg Nasi, 15 kg Ayam Goreng
- Available stock: 20 kg Nasi, 25 kg Ayam Goreng

**Output:**
```
Success: Stock deducted, inventory movements recorded
```

**Behavior:**
- Pre-validation passes
- Transaction starts
- Stock is deducted for all items
- Inventory movements are recorded

## Benefits

1. **Complete Validation**: Checks ALL items before starting transaction
2. **Better UX**: Users see all insufficient items at once, not one at a time
3. **Data Integrity**: No partial transactions or inconsistent state
4. **Clear Errors**: Detailed error messages in Indonesian with quantities
5. **Proper HTTP Codes**: Returns 400 for client errors, 500 for server errors

## Testing

The implementation can be verified by:
1. Running the property-based tests in `kds_stock_validation_bugfix_test.go`
2. Testing with multiple insufficient items
3. Verifying no stock deductions occur when validation fails
4. Checking error message format and content

## Requirements Satisfied

✅ 2.1: System validates stock before allowing cooking status
✅ 2.2: System returns detailed error when stock insufficient with information about which items are insufficient
