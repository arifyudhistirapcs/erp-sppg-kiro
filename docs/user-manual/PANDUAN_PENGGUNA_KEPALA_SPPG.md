# Panduan Pengguna - Kepala SPPG

## Selamat Datang di Sistem ERP SPPG

Panduan ini akan membantu Anda sebagai Kepala SPPG untuk menggunakan sistem ERP SPPG secara efektif dalam mengelola operasional program pemenuhan gizi.

## Daftar Isi

1. [Login dan Dashboard](#login-dan-dashboard)
2. [Monitoring Operasional Harian](#monitoring-operasional-harian)
3. [Persetujuan Purchase Order](#persetujuan-purchase-order)
4. [Laporan Keuangan](#laporan-keuangan)
5. [Manajemen Pengguna](#manajemen-pengguna)
6. [Audit Trail](#audit-trail)
7. [Konfigurasi Sistem](#konfigurasi-sistem)

## Login dan Dashboard

### Cara Login
1. Buka browser dan akses `https://erp-sppg.example.com`
2. Masukkan NIK atau Email Anda
3. Masukkan Password
4. Klik tombol "Masuk"

### Dashboard Kepala SPPG
Setelah login, Anda akan melihat dashboard dengan informasi:

#### Panel Produksi Harian
- **Status Menu**: Menu yang sedang dimasak hari ini
- **Progress Memasak**: Resep yang sedang dalam proses
- **Status Packing**: Kemasan yang sudah siap kirim
- **Jumlah Porsi**: Total porsi yang diproduksi

#### Panel Pengiriman
- **Pengiriman Selesai**: Jumlah sekolah yang sudah menerima makanan
- **Pengiriman Pending**: Sekolah yang belum menerima
- **Status Driver**: Aktivitas driver di lapangan
- **Waktu Pengiriman**: Estimasi selesai pengiriman

#### Panel Stok Kritis
- **Bahan Baku Menipis**: Daftar bahan yang perlu segera dipesan
- **Hari Tersisa**: Estimasi habis stok
- **Rekomendasi**: Saran pemesanan

#### KPI (Key Performance Indicators)
- **Tingkat Ketersediaan Stok**: Persentase stok yang tersedia
- **Tingkat Penyelesaian Pengiriman**: Persentase pengiriman tepat waktu
- **Efisiensi Produksi**: Rasio produksi vs target

### Navigasi Menu
Menu utama tersedia di sidebar kiri:
- **Dashboard**: Halaman utama
- **Produksi**: Menu planning dan resep
- **Pengadaan**: Supplier dan purchase order
- **Inventori**: Stok bahan baku
- **Logistik**: Pengiriman dan tracking
- **Keuangan**: Laporan dan cash flow
- **SDM**: Karyawan dan absensi
- **Laporan**: Berbagai laporan sistem
- **Pengaturan**: Konfigurasi sistem

## Monitoring Operasional Harian

### Memantau Produksi
1. Klik menu **"Produksi"** → **"Kitchen Display"**
2. Lihat status memasak:
   - **Belum Dimulai**: Resep yang belum dimasak
   - **Sedang Dimasak**: Resep yang sedang dalam proses
   - **Siap Packing**: Resep yang sudah selesai dimasak

### Memantau Packing
1. Klik menu **"Produksi"** → **"Packing Display"**
2. Lihat alokasi per sekolah:
   - **Nama Sekolah**: Tujuan pengiriman
   - **Jumlah Porsi**: Porsi yang harus dikemas
   - **Menu**: Jenis makanan untuk sekolah tersebut
   - **Status**: Belum dikemas / Sedang dikemas / Siap kirim

### Memantau Pengiriman
1. Klik menu **"Logistik"** → **"Tugas Pengiriman"**
2. Filter berdasarkan tanggal atau driver
3. Lihat status pengiriman:
   - **Pending**: Belum berangkat
   - **Dalam Perjalanan**: Sedang mengirim
   - **Selesai**: Sudah diterima sekolah

### Melihat Detail Pengiriman
1. Klik pada tugas pengiriman tertentu
2. Lihat informasi:
   - **Sekolah Tujuan**: Nama dan alamat
   - **Driver**: Nama driver yang bertugas
   - **Waktu Berangkat**: Jam mulai pengiriman
   - **Bukti Kirim**: Foto dan tanda tangan penerima
   - **Ompreng**: Jumlah wadah yang diantar dan diambil

## Persetujuan Purchase Order

### Melihat PO yang Perlu Disetujui
1. Klik menu **"Pengadaan"** → **"Purchase Order"**
2. Filter status **"Pending Approval"**
3. Lihat daftar PO yang menunggu persetujuan

### Meninjau Detail PO
1. Klik pada nomor PO yang ingin ditinjau
2. Periksa informasi:
   - **Supplier**: Nama dan kontak supplier
   - **Tanggal Order**: Kapan PO dibuat
   - **Tanggal Kirim**: Kapan barang diharapkan tiba
   - **Daftar Barang**: Item yang dipesan dengan harga
   - **Total Nilai**: Jumlah keseluruhan PO

### Menyetujui atau Menolak PO
1. Setelah meninjau detail PO:
   - **Untuk Menyetujui**: Klik tombol **"Setujui"**
   - **Untuk Menolak**: Klik tombol **"Tolak"** dan berikan alasan

2. Konfirmasi keputusan Anda
3. Sistem akan mengirim notifikasi ke staff pengadaan dan supplier

### Tips Persetujuan PO
- Periksa ketersediaan budget
- Pastikan supplier terpercaya (lihat rating performa)
- Cek kesesuaian harga dengan harga pasar
- Verifikasi kebutuhan barang dengan menu planning

## Laporan Keuangan

### Mengakses Laporan Keuangan
1. Klik menu **"Keuangan"** → **"Laporan Keuangan"**
2. Pilih periode laporan (harian, mingguan, bulanan, atau custom)
3. Klik **"Generate Laporan"**

### Jenis Laporan yang Tersedia

#### Laporan Arus Kas
- **Pemasukan**: Sumber dana program
- **Pengeluaran**: Belanja bahan baku, gaji, operasional
- **Saldo**: Sisa dana yang tersedia
- **Trend**: Grafik pergerakan kas

#### Laporan Budget vs Aktual
- **Budget Allocated**: Anggaran yang dialokasikan
- **Actual Spending**: Pengeluaran aktual
- **Variance**: Selisih budget vs aktual
- **Absorption Rate**: Tingkat penyerapan anggaran

#### Laporan Pengeluaran per Kategori
- **Bahan Baku**: Pembelian ingredients
- **Gaji Karyawan**: Biaya SDM
- **Utilitas**: Listrik, air, gas
- **Operasional**: Biaya operasional lainnya

### Mengekspor Laporan
1. Setelah laporan ditampilkan, klik **"Export"**
2. Pilih format:
   - **PDF**: Untuk presentasi atau arsip
   - **Excel**: Untuk analisis lebih lanjut
3. File akan diunduh otomatis

### Analisis Laporan Keuangan
- **Monitor Trend**: Perhatikan pola pengeluaran bulanan
- **Budget Control**: Pastikan tidak melebihi anggaran
- **Cost Efficiency**: Identifikasi area penghematan
- **Variance Analysis**: Analisis penyebab selisih budget

## Manajemen Pengguna

### Melihat Daftar Karyawan
1. Klik menu **"SDM"** → **"Karyawan"**
2. Lihat daftar semua karyawan dengan informasi:
   - **NIK**: Nomor Induk Karyawan
   - **Nama Lengkap**: Nama karyawan
   - **Posisi**: Jabatan dalam organisasi
   - **Role**: Hak akses dalam sistem
   - **Status**: Aktif atau non-aktif

### Menambah Karyawan Baru
1. Klik tombol **"Tambah Karyawan"**
2. Isi form dengan lengkap:
   - **NIK**: Nomor Induk Karyawan (unik)
   - **Nama Lengkap**: Nama sesuai KTP
   - **Email**: Email untuk login
   - **No. Telepon**: Nomor yang bisa dihubungi
   - **Posisi**: Jabatan karyawan
   - **Role**: Pilih sesuai tanggung jawab
3. Klik **"Simpan"**
4. Sistem akan generate username dan password otomatis

### Mengubah Role Karyawan
1. Klik pada nama karyawan yang ingin diubah
2. Klik tombol **"Edit"**
3. Ubah field **"Role"** sesuai kebutuhan
4. Klik **"Simpan"**

### Menonaktifkan Karyawan
1. Klik pada karyawan yang akan dinonaktifkan
2. Klik tombol **"Nonaktifkan"**
3. Konfirmasi keputusan
4. Karyawan tidak bisa login tapi data tetap tersimpan

## Audit Trail

### Mengakses Audit Trail
1. Klik menu **"Laporan"** → **"Audit Trail"**
2. Sistem akan menampilkan semua aktivitas pengguna

### Filter Audit Trail
Anda dapat memfilter berdasarkan:
- **Tanggal**: Pilih rentang waktu
- **Pengguna**: Pilih karyawan tertentu
- **Aksi**: Jenis aktivitas (create, update, delete)
- **Modul**: Area sistem (resep, PO, inventori, dll)

### Informasi dalam Audit Trail
- **Timestamp**: Kapan aktivitas dilakukan
- **Pengguna**: Siapa yang melakukan
- **Aksi**: Apa yang dilakukan
- **Modul**: Di bagian mana sistem
- **Detail**: Perubahan yang dilakukan
- **IP Address**: Dari mana akses dilakukan

### Menggunakan Audit Trail untuk Investigasi
1. **Tracking Perubahan**: Lihat siapa yang mengubah data penting
2. **Security Monitoring**: Deteksi aktivitas mencurigakan
3. **Compliance**: Bukti untuk audit eksternal
4. **Troubleshooting**: Identifikasi penyebab masalah

## Konfigurasi Sistem

### Mengakses Pengaturan Sistem
1. Klik menu **"Pengaturan"** → **"Konfigurasi Sistem"**
2. Anda akan melihat berbagai parameter yang bisa diatur

### Parameter yang Dapat Dikonfigurasi

#### Pengaturan Inventori
- **Minimum Stock Threshold**: Batas minimum stok per bahan
- **Alert Timing**: Kapan alert stok menipis dikirim
- **FIFO/FEFO Setting**: Metode pengelolaan stok

#### Pengaturan Nutrisi
- **Minimum Calories**: Kalori minimum per porsi
- **Minimum Protein**: Protein minimum per porsi
- **Vitamin Requirements**: Kebutuhan vitamin harian

#### Pengaturan Keamanan
- **Session Timeout**: Berapa lama user bisa idle
- **Password Policy**: Aturan password yang kuat
- **Login Attempts**: Maksimal percobaan login

#### Pengaturan Backup
- **Backup Schedule**: Jadwal backup otomatis
- **Retention Period**: Berapa lama backup disimpan
- **Backup Location**: Lokasi penyimpanan backup

### Mengubah Konfigurasi
1. Klik pada parameter yang ingin diubah
2. Masukkan nilai baru
3. Klik **"Simpan"**
4. Perubahan akan berlaku segera

### Pengaturan Wi-Fi untuk Absensi
1. Klik **"Pengaturan Wi-Fi"**
2. Tambah jaringan Wi-Fi kantor:
   - **SSID**: Nama jaringan Wi-Fi
   - **BSSID**: MAC address access point
   - **Lokasi**: Deskripsi lokasi
3. Karyawan hanya bisa absen jika terhubung ke Wi-Fi yang terdaftar

## Tips dan Best Practices

### Monitoring Harian
- Cek dashboard setiap pagi untuk overview operasional
- Perhatikan alert stok menipis
- Monitor progress produksi dan pengiriman
- Review KPI untuk identifikasi masalah

### Manajemen Keuangan
- Review laporan keuangan mingguan
- Monitor budget absorption rate
- Analisis variance untuk kontrol biaya
- Export laporan untuk dokumentasi

### Keamanan Sistem
- Ganti password secara berkala
- Logout setelah selesai menggunakan sistem
- Jangan share akun dengan orang lain
- Laporkan aktivitas mencurigakan

### Backup dan Recovery
- Pastikan backup berjalan setiap hari
- Test restore procedure secara berkala
- Simpan backup di lokasi yang aman
- Dokumentasikan prosedur recovery

## Troubleshooting

### Masalah Login
**Gejala**: Tidak bisa login ke sistem
**Solusi**:
1. Pastikan NIK/Email dan password benar
2. Cek koneksi internet
3. Clear browser cache
4. Hubungi admin sistem jika masih bermasalah

### Dashboard Tidak Update
**Gejala**: Data di dashboard tidak real-time
**Solusi**:
1. Refresh halaman browser
2. Cek koneksi internet
3. Logout dan login kembali
4. Hubungi tim teknis jika masalah berlanjut

### Laporan Tidak Muncul
**Gejala**: Laporan kosong atau error
**Solusi**:
1. Periksa filter tanggal
2. Pastikan ada data di periode tersebut
3. Coba periode yang berbeda
4. Hubungi admin jika error berlanjut

### Notifikasi Tidak Muncul
**Gejala**: Tidak menerima alert atau notifikasi
**Solusi**:
1. Cek pengaturan browser untuk notifikasi
2. Pastikan tidak ada ad-blocker yang memblokir
3. Refresh halaman
4. Hubungi admin untuk konfigurasi

## Kontak Dukungan

### Tim Teknis
- **Email**: support@erp-sppg.com
- **Telepon**: 021-xxx-xxxx
- **WhatsApp**: 08xx-xxxx-xxxx

### Jam Operasional Dukungan
- **Senin - Jumat**: 08:00 - 17:00 WIB
- **Sabtu**: 08:00 - 12:00 WIB
- **Emergency**: 24/7 (untuk masalah kritis)

### Pelatihan Tambahan
Jika membutuhkan pelatihan tambahan atau refresher, hubungi tim training:
- **Email**: training@erp-sppg.com
- **Telepon**: 021-xxx-xxxx

---

**Catatan**: Panduan ini akan terus diperbarui seiring dengan pengembangan sistem. Pastikan Anda selalu menggunakan versi terbaru dari panduan ini.