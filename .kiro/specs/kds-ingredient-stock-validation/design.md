# KDS Ingredient Stock Validation Bugfix Design

## Overview

Sistem KDS saat ini memiliki bug kritis dimana proses memasak dapat dimulai tanpa validasi ketersediaan stok komponen bahan (semi-finished goods). Fungsi `deductInventory` sudah ada tetapi dinonaktifkan (commented out) dalam method `UpdateRecipeStatus`. Bug ini menyebabkan:
1. Pesanan dapat dimasak meskipun stok tidak mencukupi
2. Stok tidak berkurang secara otomatis saat memasak dimulai
3. Ketidakakuratan data inventori

Pendekatan perbaikan adalah mengaktifkan kembali validasi stok dengan perbaikan pada logika perhitungan kebutuhan bahan berdasarkan portion size (small/large) dan memastikan error handling yang tepat untuk mencegah proses memasak saat stok tidak mencukupi.

## Glossary

- **Bug_Condition (C)**: Kondisi yang memicu bug - ketika status berubah ke "cooking" tanpa validasi stok semi-finished goods
- **Property (P)**: Perilaku yang diharapkan - sistem harus memvalidasi stok sebelum mengizinkan status "cooking" dan mengurangi stok secara otomatis
- **Preservation**: Perilaku existing yang harus tetap tidak berubah - proses memasak normal saat stok mencukupi, tampilan KDS, dan status transitions lainnya
- **UpdateRecipeStatus**: Method di `backend/internal/services/kds_service.go` yang mengubah status resep di KDS
- **deductInventory**: Method yang mengurangi stok semi-finished goods (saat ini dinonaktifkan)
- **SemiFinishedInventory**: Tabel yang menyimpan stok komponen bahan setengah jadi
- **RecipeItem**: Item dalam resep yang menentukan komponen bahan yang dibutuhkan dengan quantity per portion size
- **MenuItemSchoolAllocation**: Alokasi porsi untuk sekolah dengan portion_size (small/large)

## Bug Details

### Fault Condition

Bug terjadi ketika pengguna mengklik tombol "Mulai Masak" dan sistem mengubah status resep menjadi "cooking". Method `UpdateRecipeStatus` memiliki kode validasi dan pengurangan stok yang dinonaktifkan (baris 311-319 di kds_service.go):

```go
// TEMPORARILY DISABLED - Skip inventory deduction for now
/*
if status == "cooking" {
    err = s.deductInventory(ctx, &menuItem.Recipe, userID)
    if err != nil {
        return fmt.Errorf("failed to deduct inventory: %w", err)
    }
}
*/
```

**Formal Specification:**
```
FUNCTION isBugCondition(input)
  INPUT: input of type StatusUpdateRequest
  OUTPUT: boolean
  
  RETURN input.status == "cooking"
         AND hasRecipeItems(input.recipeID)
         AND NOT stockValidationPerformed()
         AND NOT stockDeductionPerformed()
END FUNCTION
```

### Examples

- **Contoh 1**: Chef mengklik "Mulai Masak" untuk Paket Ayam Goreng yang membutuhkan 50 porsi Nasi (small) dan 30 porsi Ayam Goreng (large). Stok Nasi hanya 20 porsi. Sistem tetap mengizinkan proses memasak dimulai tanpa error. Expected: Sistem menolak dengan pesan "Stok tidak mencukupi: Nasi (butuh 50, tersedia 20)"

- **Contoh 2**: Chef mengklik "Mulai Masak" untuk menu dengan 100 porsi. Semua komponen tersedia. Proses memasak dimulai tetapi stok tidak berkurang. Expected: Stok berkurang otomatis sesuai kebutuhan resep

- **Contoh 3**: Chef mengklik "Mulai Masak" untuk menu SD yang memiliki mixed portion sizes (40 small, 60 large). Sistem tidak menghitung kebutuhan stok berdasarkan portion size. Expected: Sistem menghitung kebutuhan: (40 × quantity_per_portion_small) + (60 × quantity_per_portion_large)

- **Edge Case**: Chef mengklik "Mulai Masak" untuk resep tanpa RecipeItems (resep kosong). Expected: Sistem menolak dengan pesan error yang jelas

## Expected Behavior

### Preservation Requirements

**Unchanged Behaviors:**
- Proses memasak normal saat semua stok mencukupi harus tetap berjalan seperti biasa
- Perubahan status ke "ready" dan "pending" tidak boleh terpengaruh
- Tampilan daftar menu di KDS harus tetap sama
- Integrasi dengan monitoring system (delivery records) harus tetap berfungsi
- Firebase sync untuk status updates harus tetap berjalan
- Mouse clicks dan interaksi UI lainnya harus tetap berfungsi

**Scope:**
Semua inputs yang TIDAK melibatkan perubahan status ke "cooking" harus sepenuhnya tidak terpengaruh oleh fix ini. Ini termasuk:
- Status updates ke "ready" atau "pending"
- Operasi GET untuk mengambil menu hari ini
- Operasi sync manual ke Firebase
- Operasi packing dan delivery

## Hypothesized Root Cause

Berdasarkan analisis kode, penyebab bug yang paling mungkin adalah:

1. **Intentional Temporary Disable**: Kode validasi dan pengurangan stok sengaja dinonaktifkan dengan comment "TEMPORARILY DISABLED" kemungkinan karena:
   - Bug atau issue dalam implementasi awal yang belum diperbaiki
   - Kebutuhan untuk testing tanpa validasi stok
   - Incomplete implementation yang ditunda

2. **Incomplete Portion Size Calculation**: Method `deductInventory` yang ada menggunakan `ri.Quantity` (deprecated field) tanpa memperhitungkan:
   - `quantity_per_portion_small` dan `quantity_per_portion_large` dari RecipeItem
   - Actual portion allocations dari MenuItemSchoolAllocation
   - Mixed portion sizes untuk sekolah SD

3. **Missing Validation Logic**: Tidak ada pre-check untuk memvalidasi stok sebelum memulai transaction, sehingga:
   - Error baru terdeteksi di tengah proses deduction
   - Tidak ada informasi lengkap tentang semua komponen yang kurang stok
   - User experience buruk karena partial error messages

4. **Transaction Rollback Issues**: Implementasi existing menggunakan transaction tetapi error handling mungkin tidak optimal untuk concurrent requests

## Correctness Properties

Property 1: Fault Condition - Stock Validation Before Cooking

_For any_ status update request where status is "cooking" and the recipe has recipe items, the fixed UpdateRecipeStatus function SHALL validate that all semi-finished goods have sufficient stock based on portion size calculations (small/large) before allowing the status change, and SHALL return a detailed error message listing all insufficient items if validation fails.

**Validates: Requirements 2.1, 2.2**

Property 2: Fault Condition - Automatic Stock Deduction

_For any_ status update request where status is "cooking" and stock validation passes, the fixed UpdateRecipeStatus function SHALL automatically deduct the calculated quantities from SemiFinishedInventory and record the movements in InventoryMovement table, ensuring atomic transaction completion.

**Validates: Requirements 2.3**

Property 3: Preservation - Non-Cooking Status Updates

_For any_ status update request where status is NOT "cooking" (e.g., "ready", "pending"), the fixed UpdateRecipeStatus function SHALL produce exactly the same behavior as before, without performing any stock validation or deduction, preserving all existing functionality for non-cooking status transitions.

**Validates: Requirements 3.1, 3.2, 3.4**

Property 4: Preservation - KDS Display and UI Interactions

_For any_ GET request to retrieve cooking menu or any UI interaction that does not involve status updates to "cooking", the system SHALL produce exactly the same behavior as before, preserving the display of menu items, school allocations, and all UI interactions.

**Validates: Requirements 3.3**

## Fix Implementation

### Changes Required

Assuming our root cause analysis is correct:

**File**: `backend/internal/services/kds_service.go`

**Function**: `UpdateRecipeStatus` (line 283)

**Specific Changes**:

1. **Uncomment and Fix Stock Validation**: Aktifkan kembali blok kode yang dinonaktifkan (baris 311-319)
   - Hapus comment markers `/*` dan `*/`
   - Perbaiki error message untuk lebih informatif

2. **Fix deductInventory Method**: Perbaiki method `deductInventory` (line 528) untuk menghitung kebutuhan berdasarkan portion size
   - Hitung total kebutuhan dari MenuItemSchoolAllocation
   - Gunakan `quantity_per_portion_small` dan `quantity_per_portion_large`
   - Formula: `totalNeeded = (smallPortions × quantity_per_portion_small) + (largePortions × quantity_per_portion_large)`

3. **Add Pre-Validation Check**: Tambahkan validasi awal sebelum memulai transaction
   - Check semua komponen sekaligus
   - Kumpulkan semua item yang kurang stok
   - Return error dengan daftar lengkap komponen yang kurang

4. **Improve Error Messages**: Perbaiki error messages untuk lebih user-friendly
   - Format: "Stok tidak mencukupi untuk: Nasi (butuh 50.00 kg, tersedia 20.00 kg), Ayam Goreng (butuh 30.00 kg, tersedia 15.00 kg)"
   - Gunakan bahasa Indonesia sesuai dengan sistem

5. **Add Logging**: Tambahkan logging untuk debugging dan audit trail
   - Log setiap stock validation attempt
   - Log setiap stock deduction
   - Log errors dengan context lengkap

### API Changes

**No API signature changes required**. Endpoint tetap sama:
- `PUT /api/v1/kds/cooking/:recipe_id/status`

**Response changes**:
- Error response akan lebih detail saat stok tidak mencukupi
- Error code baru: `INSUFFICIENT_STOCK`

**Example Error Response**:
```json
{
  "success": false,
  "error_code": "INSUFFICIENT_STOCK",
  "message": "Stok tidak mencukupi untuk memulai memasak",
  "details": "Stok tidak mencukupi untuk: Nasi (butuh 50.00 kg, tersedia 20.00 kg), Ayam Goreng (butuh 30.00 kg, tersedia 15.00 kg)"
}
```

### Database Schema Changes

**No schema changes required**. Existing tables sudah mencukupi:
- `semi_finished_inventories` - untuk tracking stok
- `inventory_movements` - untuk audit trail
- `recipe_items` - sudah memiliki `quantity_per_portion_small` dan `quantity_per_portion_large`
- `menu_item_school_allocations` - sudah memiliki `portion_size` dan `portions`

### Error Handling

**Error Scenarios**:

1. **Insufficient Stock**: Return HTTP 400 dengan error_code `INSUFFICIENT_STOCK`
   - Message: Detail komponen yang kurang dengan jumlah
   - Action: User harus menambah stok atau mengurangi porsi

2. **Missing Inventory Record**: Return HTTP 500 dengan error_code `INVENTORY_NOT_FOUND`
   - Message: "Data inventori tidak ditemukan untuk [nama komponen]"
   - Action: Admin harus setup inventory record

3. **Transaction Failure**: Return HTTP 500 dengan error_code `TRANSACTION_FAILED`
   - Message: "Gagal memperbarui stok, silakan coba lagi"
   - Action: Retry atau contact support

4. **Empty Recipe**: Return HTTP 400 dengan error_code `INVALID_RECIPE`
   - Message: "Resep tidak memiliki komponen bahan"
   - Action: Admin harus melengkapi resep

## Testing Strategy

### Validation Approach

Testing strategy mengikuti pendekatan dua fase: pertama, surface counterexamples yang mendemonstrasikan bug pada kode yang belum diperbaiki, kemudian verifikasi bahwa fix bekerja dengan benar dan mempertahankan perilaku existing.

### Exploratory Fault Condition Checking

**Goal**: Surface counterexamples yang mendemonstrasikan bug SEBELUM mengimplementasikan fix. Konfirmasi atau refute root cause analysis. Jika refute, kita perlu re-hypothesize.

**Test Plan**: Tulis tests yang mensimulasikan klik tombol "Mulai Masak" dengan berbagai kondisi stok. Jalankan tests pada kode UNFIXED untuk mengamati failures dan memahami root cause.

**Test Cases**:
1. **Insufficient Stock Test**: Simulate status update ke "cooking" dengan stok tidak mencukupi (akan berhasil pada unfixed code, seharusnya gagal)
2. **Zero Stock Test**: Simulate status update ke "cooking" dengan stok 0 (akan berhasil pada unfixed code, seharusnya gagal)
3. **Mixed Portion Size Test**: Simulate status update untuk menu SD dengan mixed portions (akan berhasil pada unfixed code tanpa perhitungan yang benar)
4. **Stock Not Deducted Test**: Verify stok tidak berkurang setelah status "cooking" (akan pass pada unfixed code, menunjukkan bug)

**Expected Counterexamples**:
- Status berubah ke "cooking" meskipun stok tidak mencukupi
- Stok tidak berkurang setelah status berubah ke "cooking"
- Perhitungan kebutuhan stok tidak memperhitungkan portion size
- Possible causes: kode validasi dinonaktifkan, perhitungan portion size salah, transaction tidak di-commit

### Fix Checking

**Goal**: Verifikasi bahwa untuk semua inputs dimana bug condition berlaku, fixed function menghasilkan expected behavior.

**Pseudocode:**
```
FOR ALL input WHERE isBugCondition(input) DO
  result := UpdateRecipeStatus_fixed(input)
  ASSERT stockValidationPerformed(result)
  ASSERT stockDeductionPerformed(result) OR errorReturned(result)
  IF errorReturned(result) THEN
    ASSERT errorMessage contains detailed stock information
  END IF
END FOR
```

**Test Cases**:
1. **Sufficient Stock Success**: Verify status berubah ke "cooking" dan stok berkurang dengan benar
2. **Insufficient Stock Rejection**: Verify status tidak berubah dan error message detail
3. **Portion Size Calculation**: Verify perhitungan kebutuhan stok benar untuk mixed portions
4. **Transaction Atomicity**: Verify semua atau tidak ada perubahan stok (no partial updates)

### Preservation Checking

**Goal**: Verifikasi bahwa untuk semua inputs dimana bug condition TIDAK berlaku, fixed function menghasilkan hasil yang sama dengan original function.

**Pseudocode:**
```
FOR ALL input WHERE NOT isBugCondition(input) DO
  ASSERT UpdateRecipeStatus_original(input) = UpdateRecipeStatus_fixed(input)
END FOR
```

**Testing Approach**: Property-based testing direkomendasikan untuk preservation checking karena:
- Menghasilkan banyak test cases secara otomatis across input domain
- Menangkap edge cases yang mungkin terlewat oleh manual unit tests
- Memberikan jaminan kuat bahwa behavior tidak berubah untuk semua non-buggy inputs

**Test Plan**: Observe behavior pada UNFIXED code terlebih dahulu untuk status updates non-cooking, kemudian tulis property-based tests yang menangkap behavior tersebut.

**Test Cases**:
1. **Status "ready" Preservation**: Observe bahwa status update ke "ready" bekerja dengan benar pada unfixed code, kemudian verify tetap bekerja setelah fix
2. **Status "pending" Preservation**: Observe bahwa status update ke "pending" bekerja dengan benar pada unfixed code, kemudian verify tetap bekerja setelah fix
3. **GET Menu Preservation**: Observe bahwa GET /api/v1/kds/cooking/today bekerja dengan benar pada unfixed code, kemudian verify tetap bekerja setelah fix
4. **Firebase Sync Preservation**: Observe bahwa Firebase sync bekerja dengan benar pada unfixed code, kemudian verify tetap bekerja setelah fix

### Unit Tests

- Test `deductInventory` dengan berbagai kombinasi portion sizes
- Test stock validation logic dengan edge cases (zero stock, exact match, overflow)
- Test error message formatting untuk berbagai skenario insufficient stock
- Test transaction rollback saat terjadi error di tengah proses

### Property-Based Tests

- Generate random menu configurations dengan berbagai portion allocations dan verify stock calculation correct
- Generate random stock levels dan verify validation logic works across all scenarios
- Generate random status update sequences dan verify preservation of non-cooking updates
- Test concurrent status updates untuk verify transaction isolation

### Integration Tests

- Test full workflow: create menu → allocate portions → start cooking dengan sufficient stock
- Test full workflow: create menu → allocate portions → start cooking dengan insufficient stock (should fail)
- Test workflow dengan mixed portion sizes untuk sekolah SD
- Test workflow dengan multiple concurrent cooking requests
