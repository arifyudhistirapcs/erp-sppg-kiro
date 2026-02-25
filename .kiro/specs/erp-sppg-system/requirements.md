# Requirements Document: Sistem ERP SPPG

## Introduction

Sistem ERP SPPG (Satuan Pelayanan Pemenuhan Gizi) adalah platform manajemen operasional terintegrasi untuk mengelola produksi, distribusi, dan pelaporan program pemenuhan gizi. Sistem ini dirancang untuk meningkatkan transparansi, akuntabilitas, dan efisiensi operasional dari perencanaan menu hingga pengiriman makanan ke sekolah-sekolah penerima manfaat.

Sistem mengimplementasikan arsitektur produksi dua tingkat:
1. **Tingkat 1**: Bahan baku (raw ingredients) → Barang setengah jadi (semi-finished goods)
2. **Tingkat 2**: Barang setengah jadi → Menu final (recipes)

Contoh alur produksi:
- Beras (bahan baku) → Nasi (barang setengah jadi)
- Ayam mentah (bahan baku) → Ayam Goreng (barang setengah jadi)
- Nasi + Ayam Goreng + Sambal (barang setengah jadi) → Paket Ayam Goreng (menu final)

Sistem terdiri dari 3 komponen utama:
- **erp-sppg-web**: Aplikasi web desktop untuk staff kantor
- **erp-sppg-pwa**: Progressive Web App untuk driver dan absensi karyawan
- **erp-sppg-be**: Backend API dengan Golang, PostgreSQL/Cloud SQL, Redis caching, dan Firebase real-time synchronization

## Glossary

- **SPPG**: Satuan Pelayanan Pemenuhan Gizi - unit organisasi yang mengelola program gizi
- **System**: Sistem ERP SPPG secara keseluruhan (backend, web, dan PWA)
- **Backend**: Server aplikasi yang mengelola logika bisnis dan database
- **Web_App**: Aplikasi web desktop untuk staff kantor
- **PWA_App**: Progressive Web App untuk perangkat mobile
- **User**: Pengguna sistem dengan role tertentu
- **Ingredient**: Bahan baku mentah yang digunakan untuk membuat barang setengah jadi
- **Semi_Finished_Goods**: Barang setengah jadi yang diproduksi dari bahan baku (contoh: Nasi, Ayam Goreng, Sambal)
- **Recipe**: Resep menu final yang terdiri dari kombinasi barang setengah jadi (contoh: Paket Ayam Goreng)
- **BoM**: Bill of Materials - daftar bahan dan takaran untuk satu resep (dalam sistem ini mengacu pada resep barang setengah jadi)
- **KDS**: Kitchen Display System - sistem tampilan untuk dapur dan packing
- **PO**: Purchase Order - dokumen pemesanan barang ke supplier
- **GRN**: Goods Receipt Note - dokumen penerimaan barang
- **e-POD**: Electronic Proof of Delivery - bukti pengiriman digital
- **Ompreng**: Wadah makanan yang digunakan untuk distribusi
- **FIFO**: First In First Out - metode pengelolaan stok
- **FEFO**: First Expired First Out - metode pengelolaan stok berdasarkan tanggal kadaluarsa
- **RBAC**: Role-Based Access Control - kontrol akses berbasis peran
- **Audit_Trail**: Catatan aktivitas pengguna dalam sistem
- **Geotagging**: Pencatatan koordinat GPS otomatis
- **SSID**: Service Set Identifier - nama jaringan Wi-Fi
- **BSSID**: Basic Service Set Identifier - alamat MAC access point Wi-Fi
- **CSRF**: Cross-Site Request Forgery - jenis serangan keamanan web
- **Redis**: In-memory data store untuk caching
- **Firebase**: Platform Google untuk real-time database dan sinkronisasi

## Requirements

### Requirement 1: Autentikasi dan Otorisasi Pengguna

**User Story:** Sebagai karyawan SPPG, saya ingin masuk ke sistem menggunakan akun pribadi saya agar saya hanya melihat fitur yang sesuai dengan pekerjaan saya dan aktivitas saya tercatat dalam sistem.

#### Acceptance Criteria

1. WHEN a User provides valid credentials (NIK or Email and Password), THE System SHALL authenticate the User and grant access
2. WHEN a User provides invalid credentials, THE System SHALL reject the login attempt and display an error message in Indonesian
3. WHEN a User successfully logs in, THE System SHALL redirect the User to the appropriate interface based on their role
4. THE System SHALL implement RBAC with eight distinct roles: Kepala SPPG/Yayasan, Akuntan, Ahli Gizi, Pengadaan, Chef, Packing, Driver, and Asisten Lapangan
5. WHEN a User attempts to access a feature, THE System SHALL verify the User has the required role permission before allowing access
6. WHEN a User performs any action in the System, THE System SHALL record the action in the Audit_Trail with timestamp, User ID, and action details

### Requirement 2: Manajemen Bahan Baku (Raw Ingredients)

**User Story:** Sebagai Ahli Gizi, saya ingin mengelola database bahan baku dengan informasi nutrisi agar saya dapat menghitung nilai gizi dari resep dan menu.

#### Acceptance Criteria

1. WHEN an Ahli Gizi creates a new ingredient, THE System SHALL store ingredient name, unit, and nutritional values per 100g (calories, protein, carbs, fat)
2. THE System SHALL generate unique ingredient codes automatically with format "B-XXXX"
3. WHEN an Ahli Gizi views ingredients, THE System SHALL display all ingredients with their nutritional information
4. THE System SHALL allow filtering and searching ingredients by name and category
5. THE System SHALL maintain ingredient data for use in semi-finished goods recipes and inventory tracking
6. THE System SHALL validate that ingredient codes are unique before saving

### Requirement 3: Manajemen Barang Setengah Jadi (Semi-Finished Goods)

**User Story:** Sebagai Ahli Gizi dan Chef, saya ingin mengelola barang setengah jadi dengan resep produksinya agar proses produksi lebih efisien dan terstruktur dalam dua tahap: bahan baku menjadi barang setengah jadi, lalu barang setengah jadi menjadi menu final.

#### Acceptance Criteria

1. WHEN an Ahli Gizi creates semi-finished goods, THE System SHALL store name, unit, category, description, and nutritional values per 100g
2. WHEN creating semi-finished goods, THE System SHALL require a production recipe that specifies raw ingredients needed and yield amount
3. WHEN a Chef produces semi-finished goods, THE System SHALL deduct raw ingredient quantities from inventory based on the recipe
4. WHEN semi-finished goods are produced, THE System SHALL add the produced quantity to semi-finished inventory
5. THE System SHALL track semi-finished inventory levels with minimum thresholds
6. THE System SHALL calculate nutritional values of semi-finished goods based on ingredient composition
7. WHEN semi-finished goods stock falls below minimum threshold, THE System SHALL generate low stock alerts

### Requirement 4: Manajemen Master Resep Menu Final

**User Story:** Sebagai Ahli Gizi, saya ingin mengelola resep menu final yang terdiri dari barang setengah jadi agar saya dapat menyusun paket menu lengkap dengan perhitungan gizi yang akurat.

#### Acceptance Criteria

1. WHEN an Ahli Gizi creates a recipe, THE System SHALL allow selection of semi-finished goods as recipe components
2. WHEN an Ahli Gizi inputs semi-finished goods for a recipe, THE System SHALL automatically calculate total nutritional values from the semi-finished goods composition
3. THE System SHALL validate that each recipe meets minimum nutritional standards before allowing it to be saved
4. WHEN an Ahli Gizi updates a recipe, THE System SHALL recalculate nutritional values automatically
5. THE System SHALL maintain version history for each recipe modification with change descriptions
6. WHEN an Ahli Gizi searches for recipes, THE System SHALL provide filtering by nutritional content, category, and active status
7. THE System SHALL store recipe serving size, instructions, and creator information

### Requirement 5: Penyusunan Menu Mingguan

**User Story:** Sebagai Ahli Gizi, saya ingin menyusun rencana menu mingguan yang terintegrasi dengan database resep agar standar kecukupan gizi terjaga dan menjadi instruksi produksi yang akurat bagi tim dapur.

#### Acceptance Criteria

1. WHEN an Ahli Gizi creates a weekly menu plan, THE System SHALL allow selection of recipes for each day with specified portion quantities
2. WHEN an Ahli Gizi assigns recipes to specific days, THE System SHALL calculate total nutritional values for each day
3. THE System SHALL validate that daily menu plans meet minimum nutritional requirements before allowing approval
4. WHEN Kepala SPPG approves a weekly menu, THE System SHALL update status to "approved" and record approver and approval timestamp
5. WHEN a weekly menu is approved, THE System SHALL automatically calculate total semi-finished goods and raw ingredient requirements for procurement planning
6. THE System SHALL allow Ahli Gizi to duplicate previous weekly menus as templates for faster planning
7. THE System SHALL provide daily nutrition breakdown and ingredient requirements reports for approved menu plans

### Requirement 6: Kitchen Display System untuk Tim Masak

**User Story:** Sebagai tukang masak, saya ingin melihat menu harian dan status produksi di monitor dapur agar saya bisa memasak dengan takaran yang tepat dan melacak progress produksi.

#### Acceptance Criteria

1. WHEN the current day has an approved menu, THE KDS SHALL display the list of recipes to be prepared with portion quantities
2. WHEN a Chef views a recipe on KDS, THE System SHALL display semi-finished goods requirements and cooking instructions
3. WHEN a Chef starts cooking a recipe, THE System SHALL update the recipe status to "cooking" and record the start time
4. WHEN a Chef marks a recipe as complete, THE System SHALL update the status to "ready" and notify the packing team
5. THE KDS SHALL update recipe status in real-time using Firebase listeners
6. WHEN a recipe status changes to "cooking", THE System SHALL automatically deduct semi-finished goods quantities from inventory
7. THE System SHALL provide manual sync functionality to push cooking status updates to Firebase

### Requirement 7: Kitchen Display System untuk Tim Packing

**User Story:** Sebagai tim packing, saya ingin melihat rincian alokasi per sekolah di layar agar tidak ada kesalahan jumlah dan jenis menu saat pengemasan.

#### Acceptance Criteria

1. WHEN recipes have status "ready", THE KDS SHALL display them on the packing display screen
2. WHEN the packing team views the display, THE System SHALL show the allocation of portions per school with school names and portion counts
3. WHEN the packing team views a school allocation, THE System SHALL display which menu items are assigned to that school
4. WHEN the packing team completes packing for a school, THE System SHALL allow marking the allocation as "ready" for delivery
5. WHEN all school allocations are marked "ready", THE System SHALL notify the logistics team
6. THE System SHALL update packing status in real-time across all connected displays using Firebase
7. THE System SHALL provide manual sync functionality to push packing allocation updates to Firebase

### Requirement 8: Manajemen Supplier

**User Story:** Sebagai staff Pengadaan, saya ingin mengelola database supplier dengan histori transaksi dan metrik performa agar saya dapat memilih vendor terbaik berdasarkan performa sebelumnya.

#### Acceptance Criteria

1. WHEN a Pengadaan staff creates a new supplier, THE System SHALL store supplier name, contact person, phone number, email, address, and product category
2. THE System SHALL maintain transaction history for each supplier including order dates, amounts, and delivery performance
3. WHEN a Pengadaan staff views a supplier profile, THE System SHALL display performance metrics including on-time delivery rate and quality ratings (1-5 scale)
4. THE System SHALL allow Pengadaan staff to mark suppliers as active or inactive
5. WHEN creating a Purchase Order, THE System SHALL only display active suppliers
6. THE System SHALL calculate and update supplier performance metrics automatically based on delivery completion and quality ratings

### Requirement 9: Pembuatan dan Pengelolaan Purchase Order

**User Story:** Sebagai staff Pengadaan, saya ingin membuat Purchase Order digital ke supplier agar proses pemesanan tercatat secara transparan dan dapat dilacak.

#### Acceptance Criteria

1. WHEN a Pengadaan staff creates a PO, THE System SHALL allow selection of supplier, raw ingredients, quantities, unit prices, and expected delivery date
2. WHEN a PO is created, THE System SHALL generate a unique PO number automatically and store the PO with status "pending"
3. WHEN a PO is created, THE System SHALL calculate total amount automatically from line items (quantity × unit price)
4. THE System SHALL allow Pengadaan staff to submit PO for approval by Kepala SPPG
5. WHEN Kepala SPPG approves a PO, THE System SHALL update status to "approved", record approver and approval timestamp
6. WHEN a PO is approved, THE System SHALL create an expected receipt record for warehouse verification
7. THE System SHALL allow tracking PO status from creation through delivery completion (pending, approved, received, cancelled)
8. THE System SHALL allow updating PO details only when status is "pending"

### Requirement 10: Penerimaan Barang dan Verifikasi Stok

**User Story:** Sebagai petugas gudang, saya ingin mencatat barang masuk dari supplier dengan foto nota agar stok tercatat secara transparan dan akuntabel.

#### Acceptance Criteria

1. WHEN goods arrive from a supplier, THE System SHALL allow warehouse staff to create a GRN linked to the corresponding PO
2. WHEN creating a GRN, THE System SHALL require warehouse staff to upload a photo of the supplier invoice to cloud storage
3. WHEN warehouse staff inputs received quantities, THE System SHALL compare them with PO quantities and flag discrepancies
4. WHEN a GRN is completed, THE System SHALL automatically update inventory quantities for all received raw ingredients
5. WHEN a GRN includes items with expiration dates, THE System SHALL store expiry date for FEFO inventory management
6. THE System SHALL record the receipt date, time, and receiving staff member in the GRN
7. WHEN inventory is updated from GRN, THE System SHALL apply FIFO or FEFO method based on item category and expiration dates
8. WHEN a GRN is completed, THE System SHALL automatically create a cash flow entry for the ingredient purchase
9. THE System SHALL generate unique GRN numbers automatically for tracking

### Requirement 11: Kontrol Stok dan Alert Stok Menipis

**User Story:** Sebagai staff Pengadaan, saya ingin menerima notifikasi ketika stok bahan baku atau barang setengah jadi menipis agar saya dapat melakukan pemesanan ulang tepat waktu.

#### Acceptance Criteria

1. THE System SHALL maintain real-time inventory levels for all raw ingredients and semi-finished goods
2. WHEN an ingredient or semi-finished goods quantity falls below the defined minimum threshold, THE System SHALL generate a low stock alert
3. WHEN a low stock alert is generated, THE System SHALL create a notification for Pengadaan staff through the notification system
4. THE System SHALL allow Pengadaan staff to configure minimum stock thresholds for each ingredient and semi-finished goods
5. WHEN viewing inventory, THE System SHALL display current quantity, minimum threshold, and highlight items below threshold
6. THE System SHALL provide inventory reports showing stock movements (in, out, adjustment) for any date range with filtering by ingredient
7. THE System SHALL track inventory movements with reference to source transactions (GRN number, recipe ID, production log)
8. THE System SHALL allow manual inventory initialization for new ingredients and semi-finished goods

### Requirement 12: Master Data Sekolah Penerima

**User Story:** Sebagai staff Logistik, saya ingin mengelola database sekolah penerima dengan informasi lengkap agar pengiriman dapat direncanakan dengan akurat.

#### Acceptance Criteria

1. WHEN a logistics staff creates a school record, THE System SHALL store school name, address, GPS coordinates (latitude, longitude), contact person, phone number, and student count
2. THE System SHALL validate that GPS coordinates are in valid format (latitude: -90 to 90, longitude: -180 to 180) before saving
3. WHEN a logistics staff updates school information, THE System SHALL maintain a change history through audit trail
4. THE System SHALL allow logistics staff to mark schools as active or inactive for delivery planning
5. WHEN planning deliveries, THE System SHALL only include active schools in the distribution list
6. THE System SHALL display school information including location coordinates for route planning

### Requirement 13: Aplikasi PWA untuk Driver - Daftar Tugas Pengiriman

**User Story:** Sebagai driver, saya ingin melihat daftar sekolah tujuan di HP saya agar saya mengetahui rute pengiriman hari ini.

#### Acceptance Criteria

1. WHEN a Driver logs into the PWA_App, THE System SHALL display the list of assigned delivery tasks for the current day
2. WHEN a Driver views a delivery task, THE System SHALL display school name, address, GPS coordinates, number of portions, and menu items with quantities
3. THE System SHALL order delivery tasks by route sequence number for optimized routing
4. WHEN a Driver is offline, THE PWA_App SHALL display cached delivery tasks from the last sync
5. WHEN a Driver comes back online, THE PWA_App SHALL sync any offline changes to the Backend
6. THE System SHALL allow logistics staff to assign delivery tasks to specific drivers through the Web_App with task date and route order
7. THE System SHALL allow updating delivery task status (pending, in_progress, completed, cancelled)
8. THE System SHALL allow filtering delivery tasks by driver, date, and status

### Requirement 14: Electronic Proof of Delivery dengan Geotagging

**User Story:** Sebagai driver, saya ingin melakukan konfirmasi penerimaan dengan foto dan tanda tangan digital agar bukti kirim terekam secara real-time.

#### Acceptance Criteria

1. WHEN a Driver arrives at a school, THE PWA_App SHALL automatically capture GPS coordinates (latitude, longitude) for geotagging
2. WHEN a Driver confirms delivery, THE System SHALL require input of ompreng quantities dropped off and picked up
3. WHEN a Driver confirms delivery, THE System SHALL require uploading a photo of the handover moment to cloud storage
4. WHEN a Driver confirms delivery, THE System SHALL provide a digital signature capture interface for the school representative with recipient name
5. WHEN a Driver completes the e-POD, THE System SHALL update delivery task status to "completed" and timestamp the completion
6. WHEN e-POD is completed, THE System SHALL sync the proof (photo URL, signature URL, GPS, timestamp, ompreng counts) to the Backend immediately if online, or queue for sync when connection is restored
7. THE System SHALL link e-POD uniquely to one delivery task
8. WHEN e-POD is completed, THE System SHALL automatically update ompreng tracking records for the school

### Requirement 15: Pelacakan Aset Ompreng

**User Story:** Sebagai staff Logistik, saya ingin melacak sirkulasi wadah makanan (ompreng) agar tidak ada kehilangan aset dan dapat direncanakan kebutuhan penambahan.

#### Acceptance Criteria

1. WHEN a Driver records ompreng drop-off at a school through e-POD, THE System SHALL create an ompreng tracking record and increment the ompreng balance at that school location
2. WHEN a Driver records ompreng pick-up from a school through e-POD, THE System SHALL create an ompreng tracking record and decrement the ompreng balance at that school location
3. THE System SHALL maintain a global ompreng inventory tracking total owned, at kitchen, in circulation, and missing ompreng
4. WHEN viewing ompreng tracking, THE System SHALL display current balance at each school with date-wise drop-off and pick-up history
5. WHEN ompreng counts show discrepancies, THE System SHALL calculate missing ompreng (total owned - at kitchen - in circulation)
6. THE System SHALL generate reports showing ompreng circulation patterns by school and date range
7. THE System SHALL record the driver who performed each ompreng transaction for accountability

### Requirement 16: Master Data Karyawan

**User Story:** Sebagai staff HRM, saya ingin mengelola profil lengkap karyawan SPPG agar data kepegawaian tersimpan secara terstruktur.

#### Acceptance Criteria

1. WHEN an HRM staff creates an employee record, THE System SHALL store NIK, full name, email, phone number, position, join date, and link to user account
2. THE System SHALL validate that NIK and email are unique across all employee records
3. WHEN an HRM staff creates an employee, THE System SHALL require linking to an existing user account with assigned role
4. THE System SHALL allow HRM staff to update employee information and maintain change history through audit trail
5. THE System SHALL allow HRM staff to deactivate employee accounts when staff leave the organization
6. WHEN an employee account is deactivated, THE System SHALL prevent login but retain historical data for audit purposes
7. THE System SHALL provide employee statistics including total employees, active employees, and employees by position

### Requirement 17: Absensi Karyawan dengan Validasi Wi-Fi

**User Story:** Sebagai karyawan, saya ingin absen melalui HP saat sudah sampai di kantor dengan syarat terhubung Wi-Fi kantor agar proses kehadiran jadi praktis namun tetap disiplin.

#### Acceptance Criteria

1. WHEN a karyawan opens the attendance feature in PWA_App, THE System SHALL verify the device is connected to the authorized office Wi-Fi network
2. THE System SHALL validate Wi-Fi connection by checking SSID and BSSID against registered office network identifiers
3. WHEN Wi-Fi validation fails, THE System SHALL prevent check-in and display an error message in Indonesian
4. WHEN Wi-Fi validation succeeds and karyawan checks in, THE System SHALL record attendance with check-in timestamp, employee ID, SSID, and BSSID
5. WHEN Wi-Fi validation succeeds and karyawan checks out, THE System SHALL record check-out time and calculate total work hours
6. THE System SHALL allow HRM staff to configure authorized Wi-Fi network identifiers (SSID, BSSID, location) through the Web_App
7. THE System SHALL provide attendance reports by date range with filtering by employee
8. THE System SHALL allow exporting attendance reports in Excel and PDF formats
9. THE System SHALL provide attendance statistics including total present, average work hours, and attendance rate
10. THE System SHALL retrieve today's attendance records and attendance by specific date

### Requirement 18: Manajemen Aset Dapur

**User Story:** Sebagai Akuntan, saya ingin mengelola inventaris alat masak dan aset dapur agar nilai aset tercatat dan dapat dilaporkan.

#### Acceptance Criteria

1. WHEN an Akuntan creates an asset record, THE System SHALL store unique asset code, name, category, purchase date, purchase price, current value, depreciation rate, condition (good/fair/poor), and location
2. THE System SHALL validate that asset codes are unique
3. THE System SHALL allow Akuntan to record asset maintenance activities with maintenance date, description, cost, and performer name
4. THE System SHALL calculate asset depreciation based on purchase date and configured annual depreciation rate percentage
5. WHEN viewing asset inventory, THE System SHALL display current book value and accumulated depreciation
6. THE System SHALL generate asset reports showing total asset value by category and depreciation schedules
7. THE System SHALL provide depreciation schedule calculation for individual assets showing year-by-year depreciation

### Requirement 19: Pencatatan Arus Kas Operasional

**User Story:** Sebagai Akuntan, saya ingin mencatat semua transaksi keuangan operasional agar arus kas terpantau secara akurat.

#### Acceptance Criteria

1. WHEN a financial transaction occurs, THE System SHALL create a cash flow entry with unique transaction ID, date, category, type (income/expense), amount, description, and reference
2. THE System SHALL categorize transactions into predefined accounts (bahan_baku, gaji, utilitas, operasional, lainnya)
3. WHEN a GRN is completed, THE System SHALL automatically create a cash flow entry for the ingredient purchase with reference to GRN number
4. THE System SHALL allow Akuntan to manually create cash flow entries for non-automated transactions
5. THE System SHALL record the user who created each cash flow entry for audit purposes
6. THE System SHALL provide cash flow summary reports showing total income, total expenses by category, and net cash flow for a date range

### Requirement 20: Laporan Keuangan Otomatis

**User Story:** Sebagai Akuntan, saya ingin laporan arus kas otomatis yang mencakup belanja bahan dan gaji agar laporan ke Yayasan/Badan Gizi bisa selesai tepat waktu.

#### Acceptance Criteria

1. WHEN an Akuntan requests a financial report, THE System SHALL allow filtering by date range (start date and end date)
2. WHEN generating a report, THE System SHALL aggregate all cash flow entries within the selected period
3. THE System SHALL generate reports showing income, expenses by category, and net cash flow
4. THE System SHALL provide budget vs actual comparison when budget targets are configured for the period
5. THE System SHALL allow exporting financial reports in PDF and Excel formats with formatted tables and charts
6. WHEN exporting reports, THE System SHALL include summary tables showing totals by category and detailed transaction listings
7. THE System SHALL allow filtering financial reports by transaction category

### Requirement 21: Dashboard Monitoring untuk Kepala SPPG

**User Story:** Sebagai Kepala SPPG, saya ingin melihat status operasional harian secara real-time agar saya dapat mengidentifikasi masalah dan mengambil keputusan cepat.

#### Acceptance Criteria

1. WHEN Kepala SPPG accesses the dashboard, THE System SHALL display production milestones for the current day (menu status, cooking progress, packing status)
2. WHEN Kepala SPPG views the dashboard, THE System SHALL display real-time delivery status showing completed and pending deliveries
3. WHEN Kepala SPPG views the dashboard, THE System SHALL highlight critical stock items (raw ingredients and semi-finished goods) below minimum threshold
4. THE System SHALL update dashboard metrics in real-time using Firebase listeners
5. WHEN Kepala SPPG clicks on a metric, THE System SHALL provide drill-down details for that operational area
6. THE System SHALL display key performance indicators including total portions prepared, delivery completion rate, and stock availability percentage
7. THE System SHALL allow manual sync of dashboard data to Firebase
8. THE System SHALL allow exporting dashboard data for reporting purposes

### Requirement 22: Dashboard Monitoring untuk Kepala Yayasan

**User Story:** Sebagai Kepala Yayasan, saya ingin melihat grafik penyerapan anggaran dan capaian gizi secara real-time untuk memastikan program berjalan efisien.

#### Acceptance Criteria

1. WHEN Kepala Yayasan accesses the dashboard, THE System SHALL display budget absorption rate comparing actual spending to allocated budget
2. WHEN Kepala Yayasan views the dashboard, THE System SHALL display cumulative nutrition distribution metrics (total portions distributed, schools served, students reached)
3. WHEN Kepala Yayasan views the dashboard, THE System SHALL display supplier performance metrics including on-time delivery rates and quality scores
4. THE System SHALL provide trend charts showing budget spending and distribution volumes over time
5. WHEN Kepala Yayasan selects a time period, THE System SHALL update all dashboard metrics for that period
6. THE System SHALL allow exporting dashboard data and charts for presentation purposes
7. THE System SHALL display key performance indicators relevant to foundation oversight

### Requirement 23: Audit Trail untuk Semua Aktivitas Pengguna

**User Story:** Sebagai Kepala SPPG, saya ingin melihat riwayat aktivitas pengguna dalam sistem agar saya dapat memastikan akuntabilitas dan melacak perubahan data penting.

#### Acceptance Criteria

1. WHEN any User performs a create, update, or delete action, THE System SHALL record the action in the Audit_Trail automatically through middleware
2. WHEN recording an audit entry, THE System SHALL capture User ID, timestamp, action type, affected entity (table/resource name), entity ID, old values, new values, and IP address
3. THE System SHALL allow authorized users to search audit trail by date range, user, action type, or entity
4. WHEN viewing audit trail, THE System SHALL display entries in reverse chronological order with clear descriptions in Indonesian
5. THE System SHALL retain audit trail data for a minimum of 2 years
6. THE System SHALL prevent modification or deletion of audit trail entries by any user including administrators
7. THE System SHALL provide audit statistics showing activity counts by user, action type, and time period

### Requirement 24: Sinkronisasi Real-Time dengan Firebase

**User Story:** Sebagai pengguna sistem, saya ingin melihat perubahan data secara real-time tanpa perlu refresh halaman agar informasi yang saya lihat selalu terkini.

#### Acceptance Criteria

1. WHEN data changes in the Backend, THE System SHALL push updates to connected clients using Firebase Realtime Database
2. THE System SHALL implement real-time updates for KDS displays (cooking and packing status), delivery status, inventory levels, dashboard metrics, and notifications
3. WHEN a client receives a real-time update, THE System SHALL update the UI without requiring page refresh
4. WHEN a client loses connection, THE System SHALL attempt to reconnect automatically
5. WHEN a client reconnects after being offline, THE System SHALL sync any missed updates
6. THE System SHALL handle concurrent updates from multiple users using server-wins conflict resolution strategy
7. THE System SHALL include updated_at timestamp with all Firebase updates for tracking
8. THE System SHALL organize Firebase data by logical paths (e.g., /kds/cooking/{date}/{recipe_id}, /notifications/{user_id}/{notification_id})

### Requirement 25: Kemampuan Offline untuk PWA Driver

**User Story:** Sebagai Driver, saya ingin tetap dapat melihat tugas pengiriman dan mencatat bukti kirim meskipun sinyal internet lemah agar pekerjaan tidak terhambat.

#### Acceptance Criteria

1. WHEN the PWA_App detects no internet connection, THE System SHALL allow Driver to continue viewing cached delivery tasks
2. WHEN a Driver is offline, THE PWA_App SHALL allow recording e-POD data (photos, signatures, GPS coordinates, ompreng counts, recipient name) locally
3. WHEN the PWA_App regains internet connection, THE System SHALL automatically sync all offline-recorded data to the Backend
4. THE System SHALL indicate offline status clearly in the PWA_App interface
5. WHEN syncing offline data, THE System SHALL handle conflicts using server-wins strategy if the same delivery was updated from another source
6. THE System SHALL cache essential data (delivery tasks, school information, menu items) for offline access when the app is loaded while online
7. THE System SHALL track sync status for offline-captured data and retry failed syncs automatically

### Requirement 24: Antarmuka Pengguna dalam Bahasa Indonesia

**User Story:** Sebagai pengguna sistem, saya ingin semua label, pesan, dan instruksi dalam Bahasa Indonesia agar mudah dipahami oleh seluruh tim SPPG.

#### Acceptance Criteria

1. THE System SHALL display all user interface labels, buttons, and menu items in professional Indonesian language
2. THE System SHALL display all error messages, validation messages, and notifications in Indonesian
3. THE System SHALL display all reports, exports, and printed documents with Indonesian labels and formatting
4. THE System SHALL use Indonesian date and number formatting conventions (DD/MM/YYYY, thousand separator)
5. THE System SHALL provide Indonesian language help text and tooltips for complex features
6. WHEN displaying technical terms without direct Indonesian translation, THE System SHALL use widely understood terms with explanations where needed

### Requirement 25: Keamanan dan Enkripsi Data

**User Story:** Sebagai Kepala SPPG, saya ingin data sistem dilindungi dengan enkripsi dan kontrol akses yang ketat agar informasi sensitif organisasi aman.

#### Acceptance Criteria

1. THE System SHALL encrypt all passwords using bcrypt hashing algorithm before storage
2. THE System SHALL transmit all data between clients and Backend using HTTPS/TLS encryption
3. THE System SHALL implement session timeout after configured minutes of inactivity (default 30 minutes)
4. WHEN a User session expires, THE System SHALL require re-authentication before allowing further access
5. THE System SHALL validate and sanitize all user inputs to prevent SQL injection and XSS attacks through middleware
6. THE System SHALL implement rate limiting on authentication endpoints to prevent brute force attacks
7. THE System SHALL implement API rate limiting on all endpoints to prevent abuse
8. THE System SHALL implement CSRF protection with token validation for state-changing operations
9. THE System SHALL validate User-Agent headers to block suspicious requests
10. THE System SHALL enforce request size limits to prevent denial of service attacks
11. THE System SHALL implement IP whitelist for sensitive administrative endpoints (system configuration)
12. THE System SHALL add security headers (X-Frame-Options, X-Content-Type-Options, etc.) to all responses
13. THE System SHALL redirect HTTP requests to HTTPS when HTTPS is enabled

### Requirement 26: Backup dan Recovery Data

**User Story:** Sebagai Kepala SPPG, saya ingin sistem memiliki backup otomatis agar data tidak hilang jika terjadi kegagalan sistem.

#### Acceptance Criteria

1. THE System SHALL perform automated daily backups of the PostgreSQL database
2. THE System SHALL retain backup copies for a minimum of 30 days
3. THE System SHALL store backups in a separate geographic location from the primary database
4. THE System SHALL verify backup integrity after each backup operation
5. THE System SHALL provide a documented recovery procedure for restoring from backups
6. WHEN a backup operation fails, THE System SHALL alert system administrators immediately

### Requirement 27: Ekspor Data dan Laporan

**User Story:** Sebagai Akuntan, saya ingin mengekspor data dan laporan dalam format Excel dan PDF agar dapat dibagikan kepada pihak eksternal.

#### Acceptance Criteria

1. WHEN a User requests to export data, THE System SHALL provide options for PDF and Excel formats
2. WHEN exporting to Excel, THE System SHALL format data in tables with appropriate column headers in Indonesian
3. WHEN exporting to PDF, THE System SHALL include organization header, report title, date range, and page numbers
4. THE System SHALL allow exporting financial reports, inventory reports, delivery reports, and attendance reports
5. WHEN generating exports, THE System SHALL apply the same filters and date ranges selected in the UI
6. THE System SHALL generate export files within 30 seconds for datasets up to 10,000 records

### Requirement 28: Notifikasi dan Alert Sistem

**User Story:** Sebagai pengguna sistem, saya ingin menerima notifikasi penting agar saya tidak melewatkan informasi kritis yang memerlukan tindakan.

#### Acceptance Criteria

1. WHEN a low stock alert is triggered, THE System SHALL create a notification for Pengadaan staff with ingredient name, current quantity, and minimum threshold
2. WHEN a PO requires approval, THE System SHALL create a notification for Kepala SPPG with PO number, supplier name, and total amount
3. WHEN packing is complete and ready for delivery, THE System SHALL create a notification for assigned Driver with date and school count
4. WHEN a delivery is completed, THE System SHALL create a notification for logistics staff with school name and driver name
5. THE System SHALL push notifications to Firebase for real-time delivery to connected clients
6. THE System SHALL display unread notifications count in the user interface
7. WHEN a User clicks a notification, THE System SHALL navigate to the relevant screen using the notification link
8. THE System SHALL allow users to mark individual notifications as read or mark all as read
9. THE System SHALL allow users to delete individual notifications
10. THE System SHALL store notification type, title, message, link, read status, and creation timestamp

### Requirement 29: Konfigurasi Sistem dan Parameter

**User Story:** Sebagai administrator sistem, saya ingin mengkonfigurasi parameter operasional agar sistem dapat disesuaikan dengan kebutuhan SPPG tanpa perlu perubahan kode.

#### Acceptance Criteria

1. THE System SHALL allow administrators to configure system parameters with key-value pairs, data type (string, int, float, bool, json), and category
2. THE System SHALL validate configuration values match the specified data type before saving
3. THE System SHALL allow administrators to configure minimum stock thresholds for ingredients and semi-finished goods
4. THE System SHALL allow administrators to configure authorized Wi-Fi networks (SSID and BSSID) for attendance
5. THE System SHALL allow administrators to configure nutritional minimum standards for menu validation
6. THE System SHALL allow administrators to configure session timeout duration
7. WHEN configuration changes are saved, THE System SHALL apply them immediately without requiring system restart
8. THE System SHALL record who updated each configuration and when for audit purposes
9. THE System SHALL provide configuration retrieval by key or by category
10. THE System SHALL allow bulk configuration updates for multiple parameters at once
11. THE System SHALL provide default configuration initialization for new system deployments
12. THE System SHALL allow deleting configuration entries when no longer needed

### Requirement 30: Validasi Data dan Penanganan Error

**User Story:** Sebagai pengguna sistem, saya ingin sistem memberikan pesan error yang jelas ketika saya melakukan kesalahan input agar saya dapat memperbaikinya dengan mudah.

#### Acceptance Criteria

1. WHEN a User submits a form with missing required fields, THE System SHALL display validation errors in Indonesian indicating which fields are required
2. WHEN a User inputs data in incorrect format, THE System SHALL display format requirements and examples
3. WHEN a system error occurs, THE System SHALL display a user-friendly error message in Indonesian and log technical details for administrators
4. THE System SHALL validate email addresses, phone numbers, and NIK formats before accepting input
5. WHEN a User attempts an action that violates business rules, THE System SHALL prevent the action and explain the constraint in Indonesian
6. THE System SHALL provide inline validation feedback as users type in form fields where possible
7. THE System SHALL return structured error responses with error_code, message, and optional details for API requests
8. THE System SHALL validate numeric ranges (e.g., latitude -90 to 90, longitude -180 to 180, percentages 0 to 100)
9. THE System SHALL validate enum values (e.g., user roles, status values) against allowed options
10. THE System SHALL validate uniqueness constraints (e.g., NIK, email, asset codes, PO numbers) before saving

### Requirement 31: Performa dan Caching

**User Story:** Sebagai pengguna sistem, saya ingin sistem merespon dengan cepat agar pekerjaan saya tidak terhambat oleh loading yang lama.

#### Acceptance Criteria

1. THE System SHALL implement Redis caching for frequently accessed data to improve response times
2. THE System SHALL cache read-only operations (GET requests) for recipes, ingredients, suppliers, and schools with appropriate cache duration
3. THE System SHALL implement long-duration caching (e.g., 1 hour) for master data that changes infrequently
4. THE System SHALL implement short-duration caching for inventory data to balance freshness and performance
5. THE System SHALL implement dashboard-specific caching with appropriate invalidation strategies
6. WHEN data is modified (create, update, delete), THE System SHALL automatically invalidate related cache entries through middleware
7. THE System SHALL implement database query optimization with proper indexing on frequently queried fields
8. THE System SHALL monitor database query performance and log slow queries for optimization
9. THE System SHALL use database connection pooling for efficient resource utilization
10. THE System SHALL implement conditional caching that only caches appropriate request types (read-only operations)

### Requirement 32: Backup dan Recovery Data

**User Story:** Sebagai Kepala SPPG, saya ingin sistem memiliki backup otomatis agar data tidak hilang jika terjadi kegagalan sistem.

#### Acceptance Criteria

1. THE System SHALL perform automated daily backups of the PostgreSQL database
2. THE System SHALL retain backup copies for a minimum of 30 days
3. THE System SHALL store backups in a separate geographic location from the primary database
4. THE System SHALL verify backup integrity after each backup operation
5. THE System SHALL provide a documented recovery procedure for restoring from backups
6. WHEN a backup operation fails, THE System SHALL alert system administrators immediately

