# Requirements Document: Sistem ERP SPPG

## Introduction

Sistem ERP SPPG (Satuan Pelayanan Pemenuhan Gizi) adalah platform manajemen operasional terintegrasi untuk mengelola produksi, distribusi, dan pelaporan program pemenuhan gizi. Sistem ini dirancang untuk meningkatkan transparansi, akuntabilitas, dan efisiensi operasional dari perencanaan menu hingga pengiriman makanan ke sekolah-sekolah penerima manfaat.

Sistem terdiri dari 3 komponen utama:
- **erp-sppg-web**: Aplikasi web desktop untuk staff kantor
- **erp-sppg-pwa**: Progressive Web App untuk driver dan absensi karyawan
- **erp-sppg-be**: Backend API dengan Golang, PostgreSQL/Cloud SQL, dan Firebase real-time listener

## Glossary

- **SPPG**: Satuan Pelayanan Pemenuhan Gizi - unit organisasi yang mengelola program gizi
- **System**: Sistem ERP SPPG secara keseluruhan (backend, web, dan PWA)
- **Backend**: Server aplikasi yang mengelola logika bisnis dan database
- **Web_App**: Aplikasi web desktop untuk staff kantor
- **PWA_App**: Progressive Web App untuk perangkat mobile
- **User**: Pengguna sistem dengan role tertentu
- **BoM**: Bill of Materials - daftar bahan baku dan takaran untuk satu resep
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

### Requirement 2: Manajemen Master Resep dan Bill of Materials

**User Story:** Sebagai Ahli Gizi, saya ingin mengelola database resep dengan rincian bahan baku dan informasi gizi agar saya dapat menyusun menu yang memenuhi standar kecukupan gizi.

#### Acceptance Criteria

1. WHEN an Ahli Gizi creates a new recipe, THE System SHALL store the recipe name, ingredients list, quantities, and cooking instructions
2. WHEN an Ahli Gizi inputs ingredients for a recipe, THE System SHALL automatically calculate total nutritional values (calories, protein, carbohydrates, fat, vitamins)
3. THE System SHALL validate that each recipe meets minimum nutritional standards before allowing it to be saved
4. WHEN an Ahli Gizi updates a recipe, THE System SHALL recalculate nutritional values automatically
5. THE System SHALL maintain version history for each recipe modification
6. WHEN an Ahli Gizi searches for recipes, THE System SHALL provide filtering by nutritional content, ingredients, and recipe category

### Requirement 3: Penyusunan Menu Mingguan

**User Story:** Sebagai Ahli Gizi, saya ingin menyusun rencana menu mingguan yang terintegrasi dengan database nutrisi (BoM) agar standar kecukupan gizi terjaga dan menjadi instruksi produksi yang akurat bagi tim dapur.

#### Acceptance Criteria

1. WHEN an Ahli Gizi creates a weekly menu plan, THE System SHALL allow selection of recipes from the Master BoM database for each day
2. WHEN an Ahli Gizi assigns recipes to specific days, THE System SHALL calculate total nutritional values for each day
3. THE System SHALL validate that daily menu plans meet minimum nutritional requirements before allowing approval
4. WHEN an Ahli Gizi approves a weekly menu, THE System SHALL make it available to the Kitchen Display System
5. WHEN a weekly menu is approved, THE System SHALL automatically calculate total ingredient requirements for procurement planning
6. THE System SHALL allow Ahli Gizi to duplicate previous weekly menus as templates for faster planning

### Requirement 4: Kitchen Display System untuk Tim Masak

**User Story:** Sebagai tukang masak, saya ingin melihat menu harian dan kebutuhan bahan baku (BoM) di monitor dapur agar saya bisa memasak dengan takaran gizi yang tepat.

#### Acceptance Criteria

1. WHEN the current day has an approved menu, THE KDS SHALL display the list of recipes to be prepared
2. WHEN a Chef views a recipe on KDS, THE System SHALL display ingredient quantities from the BoM and cooking instructions
3. WHEN a Chef starts cooking a recipe, THE System SHALL update the recipe status to "Sedang Dimasak" and record the start time
4. WHEN a Chef marks a recipe as complete, THE System SHALL update the status to "Siap Packing" and notify the packing team
5. THE KDS SHALL update recipe status in real-time using Firebase listeners
6. WHEN a recipe status changes to "Sedang Dimasak", THE System SHALL automatically deduct ingredient quantities from inventory

### Requirement 5: Kitchen Display System untuk Tim Packing

**User Story:** Sebagai tim packing, saya ingin melihat rincian jumlah porsi per sekolah di layar agar tidak ada kesalahan jumlah dan jenis menu saat pengemasan.

#### Acceptance Criteria

1. WHEN recipes have status "Siap Packing", THE KDS SHALL display them on the packing display screen
2. WHEN the packing team views the display, THE System SHALL show the allocation of portions per school with school names and portion counts
3. WHEN the packing team views a school allocation, THE System SHALL display which menu items are assigned to that school
4. WHEN the packing team completes packing for a school, THE System SHALL allow marking the allocation as "Siap Kirim"
5. WHEN all school allocations are marked "Siap Kirim", THE System SHALL notify the logistics team
6. THE System SHALL update packing status in real-time across all connected displays

### Requirement 6: Manajemen Supplier

**User Story:** Sebagai staff Pengadaan, saya ingin mengelola database supplier dengan histori transaksi agar saya dapat memilih vendor terbaik berdasarkan performa sebelumnya.

#### Acceptance Criteria

1. WHEN a Pengadaan staff creates a new supplier, THE System SHALL store supplier name, contact information, address, and product categories
2. THE System SHALL maintain transaction history for each supplier including order dates, amounts, and delivery performance
3. WHEN a Pengadaan staff views a supplier profile, THE System SHALL display performance metrics including on-time delivery rate and quality ratings
4. THE System SHALL allow Pengadaan staff to mark suppliers as active or inactive
5. WHEN creating a Purchase Order, THE System SHALL only display active suppliers
6. THE System SHALL allow Pengadaan staff to add notes and ratings after each transaction with a supplier

### Requirement 7: Pembuatan dan Pengelolaan Purchase Order

**User Story:** Sebagai staff Pengadaan, saya ingin membuat Purchase Order digital ke supplier agar proses pemesanan tercatat secara transparan dan dapat dilacak.

#### Acceptance Criteria

1. WHEN a Pengadaan staff creates a PO, THE System SHALL allow selection of supplier, items, quantities, and expected delivery date
2. WHEN a PO is created, THE System SHALL generate a unique PO number and store the PO with status "Pending"
3. THE System SHALL allow Pengadaan staff to submit PO for approval by Kepala SPPG
4. WHEN Kepala SPPG approves a PO, THE System SHALL update status to "Approved" and send notification to the supplier contact
5. WHEN a PO is approved, THE System SHALL create an expected receipt record for warehouse verification
6. THE System SHALL allow tracking PO status from creation through delivery completion

### Requirement 8: Penerimaan Barang dan Verifikasi Stok

**User Story:** Sebagai petugas gudang, saya ingin mencatat barang masuk dari supplier dengan foto nota agar stok tercatat secara transparan dan akuntabel.

#### Acceptance Criteria

1. WHEN goods arrive from a supplier, THE System SHALL allow warehouse staff to create a GRN linked to the corresponding PO
2. WHEN creating a GRN, THE System SHALL require warehouse staff to upload a photo of the supplier invoice
3. WHEN warehouse staff inputs received quantities, THE System SHALL compare them with PO quantities and flag discrepancies
4. WHEN a GRN is completed, THE System SHALL automatically update inventory quantities for all received items
5. THE System SHALL record the receipt date, time, and receiving staff member in the GRN
6. WHEN inventory is updated, THE System SHALL apply FIFO or FEFO method based on item category and expiration dates

### Requirement 9: Kontrol Stok dan Alert Stok Menipis

**User Story:** Sebagai staff Pengadaan, saya ingin menerima notifikasi ketika stok bahan baku menipis agar saya dapat melakukan pemesanan ulang tepat waktu.

#### Acceptance Criteria

1. THE System SHALL maintain real-time inventory levels for all ingredients and materials
2. WHEN an ingredient quantity falls below the defined minimum threshold, THE System SHALL generate a low stock alert
3. WHEN a low stock alert is generated, THE System SHALL notify Pengadaan staff through the system dashboard
4. THE System SHALL allow Pengadaan staff to configure minimum stock thresholds for each ingredient
5. WHEN viewing inventory, THE System SHALL display current quantity, minimum threshold, and days of supply remaining
6. THE System SHALL provide inventory reports showing stock movements (in, out, adjustments) for any date range

### Requirement 10: Master Data Sekolah Penerima

**User Story:** Sebagai staff Logistik, saya ingin mengelola database sekolah penerima dengan informasi lengkap agar pengiriman dapat direncanakan dengan akurat.

#### Acceptance Criteria

1. WHEN a logistics staff creates a school record, THE System SHALL store school name, address, GPS coordinates, contact person, and number of students
2. THE System SHALL validate that GPS coordinates are in valid format before saving
3. WHEN a logistics staff updates school information, THE System SHALL maintain a change history
4. THE System SHALL allow logistics staff to mark schools as active or inactive for delivery planning
5. WHEN planning deliveries, THE System SHALL only include active schools in the distribution list
6. THE System SHALL allow logistics staff to add delivery notes or special instructions for each school

### Requirement 11: Aplikasi PWA untuk Driver - Daftar Tugas Pengiriman

**User Story:** Sebagai driver, saya ingin melihat daftar sekolah tujuan di HP saya agar saya mengetahui rute pengiriman hari ini.

#### Acceptance Criteria

1. WHEN a Driver logs into the PWA_App, THE System SHALL display the list of assigned delivery tasks for the current day
2. WHEN a Driver views a delivery task, THE System SHALL display school name, address, GPS coordinates, number of portions, and menu items
3. THE System SHALL order delivery tasks by optimized route sequence
4. WHEN a Driver is offline, THE PWA_App SHALL display cached delivery tasks from the last sync
5. WHEN a Driver comes back online, THE PWA_App SHALL sync any offline changes to the Backend
6. THE System SHALL allow logistics staff to assign delivery tasks to specific drivers through the Web_App

### Requirement 12: Electronic Proof of Delivery dengan Geotagging

**User Story:** Sebagai driver, saya ingin melakukan konfirmasi penerimaan dengan foto dan tanda tangan digital agar bukti kirim terekam secara real-time.

#### Acceptance Criteria

1. WHEN a Driver arrives at a school, THE PWA_App SHALL automatically capture GPS coordinates (geotagging)
2. WHEN a Driver confirms delivery, THE System SHALL require input of ompreng quantities dropped off and picked up
3. WHEN a Driver confirms delivery, THE System SHALL require a photo of the handover moment
4. WHEN a Driver confirms delivery, THE System SHALL provide a digital signature capture interface for the school representative
5. WHEN a Driver completes the e-POD, THE System SHALL update delivery status to "Selesai" and timestamp the completion
6. WHEN e-POD is completed, THE System SHALL sync the proof (photo, signature, GPS, timestamp) to the Backend immediately if online, or queue for sync when connection is restored

### Requirement 13: Pelacakan Aset Ompreng

**User Story:** Sebagai staff Logistik, saya ingin melacak sirkulasi wadah makanan (ompreng) agar tidak ada kehilangan aset dan dapat direncanakan kebutuhan penambahan.

#### Acceptance Criteria

1. WHEN a Driver records ompreng drop-off at a school, THE System SHALL increment the ompreng count at that school location
2. WHEN a Driver records ompreng pick-up from a school, THE System SHALL decrement the ompreng count at that school location
3. THE System SHALL maintain a global inventory of total ompreng assets including those in circulation and at the central kitchen
4. WHEN viewing ompreng tracking, THE System SHALL display current location and quantity of ompreng at each school
5. WHEN ompreng counts show discrepancies, THE System SHALL flag schools with missing ompreng for follow-up
6. THE System SHALL generate reports showing ompreng circulation patterns and loss rates over time

### Requirement 14: Master Data Karyawan

**User Story:** Sebagai staff HRM, saya ingin mengelola profil lengkap karyawan SPPG agar data kepegawaian tersimpan secara terstruktur.

#### Acceptance Criteria

1. WHEN an HRM staff creates an employee record, THE System SHALL store NIK, full name, email, phone number, position, and assigned role
2. THE System SHALL validate that NIK and email are unique across all employee records
3. WHEN an HRM staff creates an employee, THE System SHALL automatically generate login credentials for that employee
4. THE System SHALL allow HRM staff to update employee information and maintain change history
5. THE System SHALL allow HRM staff to deactivate employee accounts when staff leave the organization
6. WHEN an employee account is deactivated, THE System SHALL prevent login but retain historical data for audit purposes

### Requirement 15: Absensi Karyawan dengan Validasi Wi-Fi

**User Story:** Sebagai karyawan, saya ingin absen melalui HP saat sudah sampai di kantor dengan syarat terhubung Wi-Fi kantor agar proses kehadiran jadi praktis namun tetap disiplin.

#### Acceptance Criteria

1. WHEN a karyawan opens the attendance feature in PWA_App, THE System SHALL verify the device is connected to the authorized office Wi-Fi network
2. THE System SHALL validate Wi-Fi connection by checking SSID and BSSID against registered office network identifiers
3. WHEN Wi-Fi validation fails, THE System SHALL prevent check-in and display an error message in Indonesian
4. WHEN Wi-Fi validation succeeds and karyawan checks in, THE System SHALL record attendance with timestamp and employee ID
5. WHEN Wi-Fi validation succeeds and karyawan checks out, THE System SHALL record check-out time and calculate total work hours
6. THE System SHALL allow HRM staff to configure authorized Wi-Fi network identifiers (SSID and BSSID) through the Web_App

### Requirement 16: Manajemen Aset Dapur

**User Story:** Sebagai Akuntan, saya ingin mengelola inventaris alat masak dan aset dapur agar nilai aset tercatat dan dapat dilaporkan.

#### Acceptance Criteria

1. WHEN an Akuntan creates an asset record, THE System SHALL store asset name, category, purchase date, purchase price, and current condition
2. THE System SHALL assign a unique asset ID to each kitchen asset
3. THE System SHALL allow Akuntan to record asset maintenance activities and associated costs
4. THE System SHALL calculate asset depreciation based on purchase date and configured depreciation rates
5. WHEN viewing asset inventory, THE System SHALL display current book value and accumulated depreciation
6. THE System SHALL generate asset reports showing total asset value by category and depreciation schedules

### Requirement 17: Pencatatan Arus Kas Operasional

**User Story:** Sebagai Akuntan, saya ingin mencatat semua transaksi keuangan operasional agar arus kas terpantau secara akurat.

#### Acceptance Criteria

1. WHEN a financial transaction occurs (ingredient purchase, salary payment, utility bill), THE System SHALL create a cash flow entry
2. WHEN creating a cash flow entry, THE System SHALL require transaction date, amount, category, and description
3. THE System SHALL categorize transactions into predefined accounts (Bahan Baku, Gaji, Utilitas, Operasional Lainnya)
4. WHEN a GRN is completed, THE System SHALL automatically create a cash flow entry for the ingredient purchase
5. THE System SHALL allow Akuntan to manually record cash flow entries for non-automated transactions
6. THE System SHALL maintain running balance calculations for each account category

### Requirement 18: Laporan Keuangan Otomatis

**User Story:** Sebagai Akuntan, saya ingin laporan arus kas otomatis yang mencakup belanja bahan dan gaji agar laporan ke Yayasan/Badan Gizi bisa selesai tepat waktu.

#### Acceptance Criteria

1. WHEN an Akuntan requests a financial report, THE System SHALL allow filtering by date range (daily, weekly, monthly, custom)
2. WHEN generating a report, THE System SHALL aggregate all cash flow entries within the selected period
3. THE System SHALL generate reports showing income, expenses by category, and net cash flow
4. THE System SHALL provide budget vs actual comparison when budget targets are configured
5. THE System SHALL allow exporting financial reports in PDF and Excel formats
6. WHEN exporting reports, THE System SHALL include summary tables, charts, and detailed transaction listings

### Requirement 19: Dashboard Monitoring untuk Kepala SPPG

**User Story:** Sebagai Kepala SPPG, saya ingin melihat status operasional harian secara real-time agar saya dapat mengidentifikasi masalah dan mengambil keputusan cepat.

#### Acceptance Criteria

1. WHEN Kepala SPPG accesses the dashboard, THE System SHALL display production milestones for the current day (menu status, cooking progress, packing status)
2. WHEN Kepala SPPG views the dashboard, THE System SHALL display real-time delivery status showing completed and pending deliveries
3. WHEN Kepala SPPG views the dashboard, THE System SHALL highlight critical stock items below minimum threshold
4. THE System SHALL update dashboard metrics in real-time using Firebase listeners
5. WHEN Kepala SPPG clicks on a metric, THE System SHALL provide drill-down details for that operational area
6. THE System SHALL display key performance indicators including total portions prepared, delivery completion rate, and stock availability percentage

### Requirement 20: Dashboard Monitoring untuk Kepala Yayasan

**User Story:** Sebagai Kepala Yayasan, saya ingin melihat grafik penyerapan anggaran dan capaian gizi secara real-time untuk memastikan program berjalan efisien.

#### Acceptance Criteria

1. WHEN Kepala Yayasan accesses the dashboard, THE System SHALL display budget absorption rate comparing actual spending to allocated budget
2. WHEN Kepala Yayasan views the dashboard, THE System SHALL display cumulative nutrition distribution metrics (total portions distributed, schools served, students reached)
3. WHEN Kepala Yayasan views the dashboard, THE System SHALL display supplier performance metrics including on-time delivery rates and quality scores
4. THE System SHALL provide trend charts showing budget spending and distribution volumes over time
5. WHEN Kepala Yayasan selects a time period, THE System SHALL update all dashboard metrics for that period
6. THE System SHALL allow exporting dashboard data and charts for presentation purposes

### Requirement 21: Audit Trail untuk Semua Aktivitas Pengguna

**User Story:** Sebagai Kepala SPPG, saya ingin melihat riwayat aktivitas pengguna dalam sistem agar saya dapat memastikan akuntabilitas dan melacak perubahan data penting.

#### Acceptance Criteria

1. WHEN any User performs a create, update, or delete action, THE System SHALL record the action in the Audit_Trail
2. WHEN recording an audit entry, THE System SHALL capture User ID, timestamp, action type, affected entity, old values, and new values
3. THE System SHALL allow authorized users to search audit trail by date range, user, action type, or entity
4. WHEN viewing audit trail, THE System SHALL display entries in reverse chronological order with clear descriptions in Indonesian
5. THE System SHALL retain audit trail data for a minimum of 2 years
6. THE System SHALL prevent modification or deletion of audit trail entries by any user including administrators

### Requirement 22: Sinkronisasi Real-Time dengan Firebase

**User Story:** Sebagai pengguna sistem, saya ingin melihat perubahan data secara real-time tanpa perlu refresh halaman agar informasi yang saya lihat selalu terkini.

#### Acceptance Criteria

1. WHEN data changes in the Backend, THE System SHALL push updates to connected clients using Firebase real-time listeners
2. THE System SHALL implement real-time updates for KDS displays, delivery status, inventory levels, and dashboard metrics
3. WHEN a client receives a real-time update, THE System SHALL update the UI without requiring page refresh
4. WHEN a client loses connection, THE System SHALL attempt to reconnect automatically
5. WHEN a client reconnects after being offline, THE System SHALL sync any missed updates
6. THE System SHALL handle concurrent updates from multiple users without data conflicts

### Requirement 23: Kemampuan Offline untuk PWA Driver

**User Story:** Sebagai Driver, saya ingin tetap dapat melihat tugas pengiriman dan mencatat bukti kirim meskipun sinyal internet lemah agar pekerjaan tidak terhambat.

#### Acceptance Criteria

1. WHEN the PWA_App detects no internet connection, THE System SHALL allow Driver to continue viewing cached delivery tasks
2. WHEN a Driver is offline, THE PWA_App SHALL allow recording e-POD data (photos, signatures, GPS, ompreng counts) locally
3. WHEN the PWA_App regains internet connection, THE System SHALL automatically sync all offline-recorded data to the Backend
4. THE System SHALL indicate offline status clearly in the PWA_App interface
5. WHEN syncing offline data, THE System SHALL handle conflicts if the same delivery was updated from another source
6. THE System SHALL cache essential data (delivery tasks, school information) for offline access when the app is loaded while online

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

1. THE System SHALL encrypt all passwords using industry-standard hashing algorithms before storage
2. THE System SHALL transmit all data between clients and Backend using HTTPS/TLS encryption
3. THE System SHALL implement session timeout after 30 minutes of inactivity
4. WHEN a User session expires, THE System SHALL require re-authentication before allowing further access
5. THE System SHALL validate and sanitize all user inputs to prevent SQL injection and XSS attacks
6. THE System SHALL implement rate limiting on authentication endpoints to prevent brute force attacks

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

1. WHEN a low stock alert is triggered, THE System SHALL send notification to Pengadaan staff
2. WHEN a PO requires approval, THE System SHALL send notification to Kepala SPPG
3. WHEN packing is complete and ready for delivery, THE System SHALL send notification to assigned Driver
4. WHEN a delivery is completed, THE System SHALL send notification to logistics staff
5. THE System SHALL display unread notifications count in the user interface
6. WHEN a User clicks a notification, THE System SHALL navigate to the relevant screen or record

### Requirement 29: Konfigurasi Sistem dan Parameter

**User Story:** Sebagai administrator sistem, saya ingin mengkonfigurasi parameter operasional agar sistem dapat disesuaikan dengan kebutuhan SPPG tanpa perlu perubahan kode.

#### Acceptance Criteria

1. THE System SHALL allow administrators to configure minimum stock thresholds for each ingredient
2. THE System SHALL allow administrators to configure authorized Wi-Fi networks (SSID and BSSID) for attendance
3. THE System SHALL allow administrators to configure nutritional minimum standards for menu validation
4. THE System SHALL allow administrators to configure session timeout duration
5. THE System SHALL allow administrators to configure backup schedule and retention period
6. WHEN configuration changes are saved, THE System SHALL apply them immediately without requiring system restart

### Requirement 30: Validasi Data dan Penanganan Error

**User Story:** Sebagai pengguna sistem, saya ingin sistem memberikan pesan error yang jelas ketika saya melakukan kesalahan input agar saya dapat memperbaikinya dengan mudah.

#### Acceptance Criteria

1. WHEN a User submits a form with missing required fields, THE System SHALL display validation errors in Indonesian indicating which fields are required
2. WHEN a User inputs data in incorrect format, THE System SHALL display format requirements and examples
3. WHEN a system error occurs, THE System SHALL display a user-friendly error message and log technical details for administrators
4. THE System SHALL validate email addresses, phone numbers, and NIK formats before accepting input
5. WHEN a User attempts an action that violates business rules, THE System SHALL prevent the action and explain the constraint
6. THE System SHALL provide inline validation feedback as users type in form fields where possible
