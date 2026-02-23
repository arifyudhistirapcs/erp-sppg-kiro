# Supply Chain & Inventory Module

## Overview

Modul Supply Chain & Inventory mengelola seluruh proses pengadaan barang, penerimaan barang, dan pengelolaan stok inventory untuk operasional SPPG.

## Features

### 1. Manajemen Supplier (`SupplierListView.vue`)

**Fitur:**
- Daftar supplier dengan pencarian dan filter
- Form tambah/edit supplier dengan validasi
- Metrik performa supplier:
  - Pengiriman tepat waktu (%)
  - Rating kualitas (1-5 bintang)
- Riwayat transaksi per supplier
- Status aktif/tidak aktif

**Role Access:** Kepala SPPG, Pengadaan

**API Endpoints:**
- `GET /api/v1/suppliers` - List suppliers
- `POST /api/v1/suppliers` - Create supplier
- `PUT /api/v1/suppliers/:id` - Update supplier
- `DELETE /api/v1/suppliers/:id` - Delete supplier
- `GET /api/v1/suppliers/:id/performance` - Get performance metrics

**Key Components:**
- Tabel supplier dengan sorting
- Modal form dengan validasi email dan telepon
- Progress bar untuk on-time delivery rate
- Rating bintang untuk quality rating
- Tabel riwayat transaksi

### 2. Purchase Order (`PurchaseOrderListView.vue`)

**Fitur:**
- Daftar PO dengan filter status (Pending, Disetujui, Diterima, Dibatalkan)
- Form buat PO dengan:
  - Pilih supplier dari daftar aktif
  - Tambah multiple items dengan ingredient selection
  - Input quantity dan harga satuan
  - Kalkulasi subtotal dan total otomatis
- Workflow approval:
  - Staff Pengadaan membuat PO (status: Pending)
  - Kepala SPPG menyetujui PO (status: Approved)
- Tracking status PO dari creation sampai delivery
- Detail PO dengan informasi lengkap

**Role Access:** 
- Kepala SPPG (full access + approval)
- Pengadaan (create, edit pending PO)

**API Endpoints:**
- `GET /api/v1/purchase-orders` - List POs
- `POST /api/v1/purchase-orders` - Create PO
- `PUT /api/v1/purchase-orders/:id` - Update PO
- `POST /api/v1/purchase-orders/:id/approve` - Approve PO (Kepala SPPG only)

**Key Components:**
- Status tags dengan warna (Pending=orange, Approved=blue, Received=green)
- Dynamic item table dengan add/remove rows
- Input number dengan currency formatter
- Date picker dengan disabled past dates
- Approval button (conditional rendering based on role)

### 3. Penerimaan Barang / GRN (`GoodsReceiptView.vue`)

**Fitur:**
- Form penerimaan barang linked ke PO yang sudah disetujui
- Upload foto invoice/nota (wajib):
  - Preview sebelum upload
  - Validasi file type (image only)
  - Max size 5MB
- Tabel item dengan:
  - Jumlah yang dipesan (dari PO)
  - Input jumlah yang diterima
  - Input tanggal kadaluarsa (opsional)
  - Highlight discrepancy (selisih) dengan warna
- Alert jika ada perbedaan antara ordered vs received
- Auto-update inventory setelah GRN completed
- Catatan tambahan

**Role Access:** Kepala SPPG, Pengadaan

**API Endpoints:**
- `POST /api/v1/goods-receipts` - Create GRN
- `POST /api/v1/goods-receipts/:id/upload-invoice` - Upload invoice photo
- `GET /api/v1/goods-receipts` - List GRNs
- `GET /api/v1/goods-receipts/:id` - Get GRN detail

**Key Components:**
- Upload component dengan preview
- Comparison table (ordered vs received)
- Discrepancy tags (red for less, green for more)
- Date picker untuk expiry date
- Auto-calculation of differences

**Business Logic:**
- Saat GRN completed, inventory otomatis ter-update
- Cash flow entry otomatis dibuat untuk pembelian
- FIFO/FEFO method diterapkan berdasarkan expiry date

### 4. Manajemen Inventory (`InventoryView.vue`)

**Fitur:**
- **Tab 1: Daftar Inventory**
  - Tabel inventory dengan stok saat ini, batas minimum, status
  - Highlight baris merah untuk stok menipis
  - Perkiraan hari sampai habis
  - Filter by stock level (Low, Normal, High)
  - Pencarian by nama bahan
  - Link ke riwayat pergerakan per item

- **Tab 2: Alert Stok Menipis**
  - List item dengan stok di bawah threshold
  - Informasi: stok saat ini, batas minimum, perkiraan habis
  - Quick action: Buat PO langsung untuk item tersebut
  - Alert count di dashboard card

- **Tab 3: Riwayat Pergerakan**
  - Filter by bahan, tanggal, tipe pergerakan
  - Tipe: Masuk (green), Keluar (red), Penyesuaian (blue)
  - Referensi (GRN number, recipe ID, dll)
  - Catatan per movement

**Role Access:** Kepala SPPG, Pengadaan, Akuntan

**API Endpoints:**
- `GET /api/v1/inventory` - List inventory items
- `GET /api/v1/inventory/:id` - Get item detail
- `GET /api/v1/inventory/alerts` - Get low stock alerts
- `GET /api/v1/inventory/movements` - Get movement history

**Key Components:**
- Statistics cards (low stock count, total items, last update)
- Tabbed interface untuk different views
- Row highlighting untuk low stock items
- Status tags dengan warna
- Date range picker untuk filter movements
- Quick action buttons

## Data Flow

### Purchase Order Flow
```
1. Pengadaan creates PO (status: pending)
2. Kepala SPPG approves PO (status: approved)
3. Supplier delivers goods
4. Warehouse staff creates GRN
5. GRN completed → Inventory updated (status: received)
6. Cash flow entry auto-created
```

### Inventory Update Flow
```
1. GRN completed → Inventory IN (+quantity)
2. Cooking starts → Inventory OUT (-quantity based on BoM)
3. Manual adjustment → Inventory ADJUSTMENT (±quantity)
```

### Low Stock Alert Flow
```
1. Inventory quantity < min_threshold
2. Alert generated automatically
3. Notification sent to Pengadaan staff
4. Pengadaan creates PO to restock
```

## UI/UX Guidelines

### Bahasa Indonesia
- Semua label, button, message dalam Bahasa Indonesia profesional
- Format tanggal: DD/MM/YYYY atau "1 Januari 2024"
- Format currency: Rp 1.000.000 (tanpa desimal untuk IDR)

### Color Coding
- **Red**: Stok menipis, discrepancy negative, keluar
- **Orange**: Pending status, perlu perhatian
- **Green**: Stok aman, approved, masuk
- **Blue**: Adjustment, informational

### Validation
- Required fields marked dengan *
- Inline validation saat user typing
- Error messages jelas dan actionable
- Confirmation dialog untuk delete actions

### Responsive Design
- Table dengan horizontal scroll untuk banyak kolom
- Modal width disesuaikan dengan content (600-1000px)
- Form layout menggunakan grid system (a-row, a-col)

## Integration Points

### With Other Modules
- **Recipe Module**: Ingredient list untuk PO item selection
- **KDS Module**: Inventory deduction saat cooking starts
- **Financial Module**: Auto cash flow entry dari GRN
- **Dashboard Module**: Low stock alerts, inventory metrics

### With Backend
- Real-time inventory updates via API
- File upload untuk invoice photos (multipart/form-data)
- Pagination untuk large datasets
- Search dan filter via query parameters

## Testing Checklist

### Supplier Management
- [ ] Create supplier dengan data valid
- [ ] Validate email format
- [ ] View supplier performance metrics
- [ ] Edit supplier information
- [ ] Delete supplier (with confirmation)
- [ ] Search supplier by name
- [ ] Filter by active/inactive status

### Purchase Order
- [ ] Create PO dengan multiple items
- [ ] Calculate subtotal dan total correctly
- [ ] Submit PO for approval
- [ ] Approve PO (as Kepala SPPG)
- [ ] Edit pending PO
- [ ] Cannot edit approved PO
- [ ] View PO detail
- [ ] Filter by status

### Goods Receipt
- [ ] Select approved PO
- [ ] Upload invoice photo (valid image)
- [ ] Reject non-image files
- [ ] Input received quantities
- [ ] Highlight discrepancies
- [ ] Input expiry dates
- [ ] Submit GRN
- [ ] Verify inventory updated after GRN

### Inventory
- [ ] View inventory list
- [ ] Low stock items highlighted in red
- [ ] Filter by stock level
- [ ] Search by ingredient name
- [ ] View movement history per item
- [ ] Low stock alerts displayed
- [ ] Quick create PO from alert
- [ ] Filter movements by date range
- [ ] Filter movements by type

## Future Enhancements

1. **Barcode Scanning**: Scan barcode untuk quick item selection
2. **Supplier Rating**: Allow users to rate suppliers after delivery
3. **Price History**: Track price changes over time per ingredient
4. **Automated Reordering**: Auto-create PO when stock reaches threshold
5. **Batch/Lot Tracking**: Track inventory by batch numbers
6. **Expiry Alerts**: Notify before items expire
7. **Supplier Portal**: Allow suppliers to view POs and update delivery status
8. **Mobile GRN**: Mobile app untuk warehouse staff to record GRN on-site

## Troubleshooting

### Common Issues

**Issue**: Inventory tidak update setelah GRN
- **Solution**: Check backend logs, verify GRN status is "completed"

**Issue**: Upload foto gagal
- **Solution**: Check file size (<5MB), format (image only), network connection

**Issue**: PO tidak bisa diapprove
- **Solution**: Verify user role is "kepala_sppg", check PO status is "pending"

**Issue**: Low stock alert tidak muncul
- **Solution**: Verify min_threshold configured, check inventory quantity

## API Service Files

- `web/src/services/supplierService.js` - Supplier API calls
- `web/src/services/purchaseOrderService.js` - PO API calls
- `web/src/services/goodsReceiptService.js` - GRN API calls
- `web/src/services/inventoryService.js` - Inventory API calls

## Component Files

- `web/src/views/SupplierListView.vue` - Supplier management
- `web/src/views/PurchaseOrderListView.vue` - PO management
- `web/src/views/GoodsReceiptView.vue` - GRN management
- `web/src/views/InventoryView.vue` - Inventory management

## Routes

```javascript
{
  path: '/suppliers',
  name: 'suppliers',
  component: SupplierListView,
  meta: { roles: ['kepala_sppg', 'pengadaan'] }
}

{
  path: '/purchase-orders',
  name: 'purchase-orders',
  component: PurchaseOrderListView,
  meta: { roles: ['kepala_sppg', 'pengadaan'] }
}

{
  path: '/goods-receipts',
  name: 'goods-receipts',
  component: GoodsReceiptView,
  meta: { roles: ['kepala_sppg', 'pengadaan'] }
}

{
  path: '/inventory',
  name: 'inventory',
  component: InventoryView,
  meta: { roles: ['kepala_sppg', 'pengadaan', 'akuntan'] }
}
```
