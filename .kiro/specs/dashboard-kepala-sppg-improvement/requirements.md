# Requirements Document

## Introduction

Dashboard Kepala SPPG adalah antarmuka utama untuk kepala SPPG (Satuan Pelaksana Penyediaan Gizi) untuk memantau operasi harian, termasuk produksi makanan, pengiriman, dan keuangan. Peningkatan ini bertujuan untuk membuat dashboard lebih relevan dengan data terkini dan menambahkan metrik keuangan yang penting untuk pengambilan keputusan operasional.

## Glossary

- **Dashboard**: Halaman utama yang menampilkan ringkasan metrik operasional SPPG
- **KPI_Card**: Komponen kartu yang menampilkan Key Performance Indicator
- **Status_Produksi_Card**: Kartu yang menampilkan status tahapan produksi makanan
- **Status_Pengiriman_Card**: Kartu yang menampilkan status tahapan pengiriman dan pembersihan
- **Financial_Card**: Kartu yang menampilkan metrik keuangan SPPG
- **Activity_Stage**: Tahapan dalam siklus operasional dari persiapan hingga pembersihan (16 tahap)
- **Real_Time_Data**: Data yang diperbarui secara otomatis menggunakan Firebase listeners
- **SPPG**: Satuan Pelaksana Penyediaan Gizi
- **Backend_Service**: Layanan Go yang menyediakan data untuk dashboard
- **Frontend_Component**: Komponen Vue.js yang menampilkan data di browser

## Requirements

### Requirement 1: Enhanced KPI Cards with Real-Time Data

**User Story:** As a Kepala SPPG, I want to see updated KPI cards with more relevant real-time metrics, so that I can monitor daily operations effectively.

#### Acceptance Criteria

1. THE Dashboard SHALL display at least 4 KPI cards showing current operational metrics
2. WHEN operational data changes in the database, THE Dashboard SHALL update KPI card values within 5 seconds
3. THE Dashboard SHALL display "Porsi Disiapkan Hari Ini" showing total portions prepared today
4. THE Dashboard SHALL display "Tingkat Pengiriman Tepat Waktu" showing percentage of on-time deliveries today
5. THE Dashboard SHALL display "Ketersediaan Stok Kritis" showing count of ingredients below minimum threshold
6. THE Dashboard SHALL display "Efisiensi Operasional" showing percentage of orders completed on schedule today
7. WHEN a KPI card is clicked, THE Dashboard SHALL navigate to the relevant detail page
8. THE Dashboard SHALL display loading indicators while fetching KPI data

### Requirement 2: Production Status Card with 16-Stage Breakdown

**User Story:** As a Kepala SPPG, I want to see detailed production status broken down by activity stages, so that I can identify bottlenecks in the cooking process.

#### Acceptance Criteria

1. THE Status_Produksi_Card SHALL display order counts for stages 0 through 5
2. THE Status_Produksi_Card SHALL display stage 0 as "Order Disiapkan" with current order count
3. THE Status_Produksi_Card SHALL display stage 1 as "Sedang Dimasak" with current order count
4. THE Status_Produksi_Card SHALL display stage 2 as "Selesai Dimasak" with current order count
5. THE Status_Produksi_Card SHALL display stage 3 as "Siap Dipacking" with current order count
6. THE Status_Produksi_Card SHALL display stage 4 as "Selesai Dipacking" with current order count
7. THE Status_Produksi_Card SHALL display stage 5 as "Siap Dikirim" with current order count
8. WHEN production stage data changes, THE Status_Produksi_Card SHALL update within 5 seconds
9. THE Status_Produksi_Card SHALL display total orders in production pipeline
10. WHEN the Status_Produksi_Card is clicked, THE Dashboard SHALL navigate to production detail page
11. THE Status_Produksi_Card SHALL use color coding to indicate stage urgency (green for on-track, yellow for delayed, red for critical)

### Requirement 3: Delivery Status Card with 16-Stage Breakdown

**User Story:** As a Kepala SPPG, I want to see detailed delivery and cleaning status broken down by activity stages, so that I can track the complete delivery cycle including ompreng return and cleaning.

#### Acceptance Criteria

1. THE Status_Pengiriman_Card SHALL display order counts for stages 6 through 15
2. THE Status_Pengiriman_Card SHALL display stage 6 as "Diperjalanan" with current order count
3. THE Status_Pengiriman_Card SHALL display stage 7 as "Sudah Sampai Sekolah" with current order count
4. THE Status_Pengiriman_Card SHALL display stage 8 as "Sudah Diterima Pihak Sekolah" with current order count
5. THE Status_Pengiriman_Card SHALL display stage 9 as "Driver Menuju Lokasi Pengambilan" with current order count
6. THE Status_Pengiriman_Card SHALL display stage 10 as "Driver Tiba di Lokasi Pengambilan" with current order count
7. THE Status_Pengiriman_Card SHALL display stage 11 as "Driver Kembali ke SPPG" with current order count
8. THE Status_Pengiriman_Card SHALL display stage 12 as "Driver Tiba di SPPG" with current order count
9. THE Status_Pengiriman_Card SHALL display stage 13 as "Ompreng Siap Dicuci" with current order count
10. THE Status_Pengiriman_Card SHALL display stage 14 as "Ompreng Proses Pencucian" with current order count
11. THE Status_Pengiriman_Card SHALL display stage 15 as "Ompreng Selesai Dicuci" with current order count
12. WHEN delivery stage data changes, THE Status_Pengiriman_Card SHALL update within 5 seconds
13. THE Status_Pengiriman_Card SHALL display total orders in delivery pipeline
14. WHEN the Status_Pengiriman_Card is clicked, THE Dashboard SHALL navigate to delivery detail page
15. THE Status_Pengiriman_Card SHALL use color coding to indicate delivery status (green for on-time, yellow for delayed, red for critical)

### Requirement 4: Financial Overview Cards

**User Story:** As a Kepala SPPG, I want to see financial metrics on the dashboard, so that I can monitor budget utilization and operational costs.

#### Acceptance Criteria

1. THE Dashboard SHALL display a "Biaya Operasional Hari Ini" card showing total operational costs for the current day
2. THE Dashboard SHALL display a "Utilisasi Anggaran Bulan Ini" card showing percentage of monthly budget used
3. THE Dashboard SHALL display a "Biaya Per Porsi" card showing average cost per portion for the current month
4. THE Dashboard SHALL display a "Ringkasan Keuangan Bulanan" card showing total income, expenses, and balance for current month
5. WHEN financial data is updated, THE Financial_Card SHALL refresh within 10 seconds
6. THE Financial_Card SHALL display currency values in Indonesian Rupiah format (Rp X.XXX.XXX)
7. THE Financial_Card SHALL display percentage values with 2 decimal places
8. WHEN a Financial_Card is clicked, THE Dashboard SHALL navigate to financial detail page
9. IF budget utilization exceeds 80%, THEN THE "Utilisasi Anggaran Bulan Ini" card SHALL display a warning indicator
10. IF cost per portion exceeds target threshold, THEN THE "Biaya Per Porsi" card SHALL display a warning indicator

### Requirement 5: Backend API for Dashboard Data

**User Story:** As a system, I want to provide efficient API endpoints for dashboard data, so that the frontend can display real-time information.

#### Acceptance Criteria

1. THE Backend_Service SHALL provide an endpoint GET /api/dashboard/kpi returning current KPI metrics
2. THE Backend_Service SHALL provide an endpoint GET /api/dashboard/production-status returning order counts by production stages (0-5)
3. THE Backend_Service SHALL provide an endpoint GET /api/dashboard/delivery-status returning order counts by delivery stages (6-15)
4. THE Backend_Service SHALL provide an endpoint GET /api/dashboard/financial-overview returning financial metrics
5. WHEN a dashboard endpoint is called, THE Backend_Service SHALL respond within 500 milliseconds
6. THE Backend_Service SHALL calculate KPI metrics from current day's data only
7. THE Backend_Service SHALL calculate financial metrics from current month's data
8. THE Backend_Service SHALL aggregate order counts by activity stage from the orders table
9. IF database query fails, THEN THE Backend_Service SHALL return error response with status code 500
10. THE Backend_Service SHALL cache dashboard data for 30 seconds to reduce database load

### Requirement 6: Real-Time Data Synchronization

**User Story:** As a Kepala SPPG, I want dashboard data to update automatically without refreshing the page, so that I always see current information.

#### Acceptance Criteria

1. THE Frontend_Component SHALL establish Firebase listeners for order status changes
2. WHEN an order status changes, THE Frontend_Component SHALL update relevant dashboard cards
3. THE Frontend_Component SHALL poll backend API every 60 seconds for KPI updates
4. THE Frontend_Component SHALL poll backend API every 120 seconds for financial updates
5. WHEN network connection is lost, THE Frontend_Component SHALL display offline indicator
6. WHEN network connection is restored, THE Frontend_Component SHALL refresh all dashboard data
7. THE Frontend_Component SHALL display last update timestamp on each card
8. THE Frontend_Component SHALL handle concurrent data updates without UI flickering

### Requirement 7: Responsive Dashboard Layout

**User Story:** As a Kepala SPPG, I want the dashboard to work well on different screen sizes, so that I can monitor operations from any device.

#### Acceptance Criteria

1. THE Dashboard SHALL display cards in a responsive grid layout
2. WHEN viewport width is greater than 1200px, THE Dashboard SHALL display 4 cards per row
3. WHEN viewport width is between 768px and 1200px, THE Dashboard SHALL display 2 cards per row
4. WHEN viewport width is less than 768px, THE Dashboard SHALL display 1 card per row
5. THE Dashboard SHALL maintain card aspect ratio across different screen sizes
6. THE Dashboard SHALL display readable text on mobile devices (minimum 14px font size)
7. THE Dashboard SHALL support touch interactions on mobile devices

### Requirement 8: Error Handling and Loading States

**User Story:** As a Kepala SPPG, I want clear feedback when data is loading or errors occur, so that I understand the dashboard state.

#### Acceptance Criteria

1. WHEN dashboard data is loading, THE Dashboard SHALL display skeleton loading placeholders
2. IF API request fails, THEN THE Dashboard SHALL display error message on affected card
3. IF API request fails, THEN THE Dashboard SHALL provide retry button on affected card
4. WHEN retry button is clicked, THE Dashboard SHALL attempt to reload failed data
5. THE Dashboard SHALL display error messages in Indonesian language
6. IF no data is available for a metric, THEN THE Dashboard SHALL display "Data tidak tersedia" message
7. THE Dashboard SHALL log errors to console for debugging purposes
8. IF multiple API requests fail consecutively, THEN THE Dashboard SHALL display general error notification

## Notes

- Implementasi harus menggunakan existing backend services (dashboard_service.go) dan handlers (dashboard_handler.go)
- Frontend menggunakan Vue.js 3 dengan Composition API
- Real-time updates menggunakan Firebase Realtime Database listeners
- Semua teks UI harus dalam Bahasa Indonesia
- Financial data harus dihitung dari tabel transaksi yang ada
- Activity stages (0-15) sudah ada di sistem, tinggal agregasi data per stage
