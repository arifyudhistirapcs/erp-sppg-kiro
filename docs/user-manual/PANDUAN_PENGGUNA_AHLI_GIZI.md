# Panduan Pengguna - Ahli Gizi

## Selamat Datang di Sistem ERP SPPG

Panduan ini akan membantu Anda sebagai Ahli Gizi untuk menggunakan sistem ERP SPPG dalam mengelola resep, perencanaan menu, dan memastikan standar gizi terpenuhi.

## Daftar Isi

1. [Login dan Dashboard](#login-dan-dashboard)
2. [Manajemen Resep](#manajemen-resep)
3. [Perencanaan Menu Mingguan](#perencanaan-menu-mingguan)
4. [Monitoring Kitchen Display](#monitoring-kitchen-display)
5. [Laporan Nutrisi](#laporan-nutrisi)
6. [Tips dan Best Practices](#tips-dan-best-practices)

## Login dan Dashboard

### Cara Login
1. Buka browser dan akses `https://erp-sppg.example.com`
2. Masukkan NIK atau Email Anda
3. Masukkan Password
4. Klik tombol "Masuk"

### Dashboard Ahli Gizi
Dashboard Anda menampilkan:

#### Panel Resep
- **Total Resep**: Jumlah resep yang sudah dibuat
- **Resep Aktif**: Resep yang sedang digunakan
- **Resep Baru**: Resep yang baru ditambahkan minggu ini
- **Rata-rata Kalori**: Kalori rata-rata per porsi

#### Panel Menu Planning
- **Menu Minggu Ini**: Status menu yang sedang berjalan
- **Menu Minggu Depan**: Status perencanaan menu berikutnya
- **Compliance Rate**: Persentase menu yang memenuhi standar gizi
- **Variasi Menu**: Tingkat keragaman menu

#### Panel Nutrisi
- **Standar Terpenuhi**: Persentase menu yang memenuhi standar
- **Kalori Rata-rata**: Kalori rata-rata per hari
- **Protein Rata-rata**: Protein rata-rata per hari
- **Alert Nutrisi**: Peringatan jika ada menu yang tidak memenuhi standar

## Manajemen Resep

### Melihat Daftar Resep
1. Klik menu **"Produksi"** → **"Resep"**
2. Anda akan melihat daftar semua resep dengan informasi:
   - **Nama Resep**: Nama makanan
   - **Kategori**: Jenis makanan (makanan pokok, lauk, sayur, dll)
   - **Porsi**: Jumlah porsi yang dihasilkan
   - **Kalori**: Total kalori per porsi
   - **Status**: Aktif atau tidak aktif

### Mencari dan Filter Resep
1. **Pencarian**: Ketik nama resep di kotak pencarian
2. **Filter Kategori**: Pilih kategori makanan
3. **Filter Nutrisi**: Filter berdasarkan kalori, protein, dll
4. **Filter Status**: Tampilkan resep aktif atau semua resep

### Menambah Resep Baru

#### Langkah 1: Informasi Dasar
1. Klik tombol **"Tambah Resep"**
2. Isi informasi dasar:
   - **Nama Resep**: Nama makanan yang akan dibuat
   - **Kategori**: Pilih kategori yang sesuai
   - **Jumlah Porsi**: Berapa porsi yang dihasilkan
   - **Deskripsi**: Deskripsi singkat tentang makanan

#### Langkah 2: Tambah Bahan-bahan
1. Klik **"Tambah Bahan"**
2. Pilih bahan dari dropdown atau ketik untuk mencari
3. Masukkan jumlah bahan yang dibutuhkan
4. Pilih satuan (kg, gram, liter, ml, pcs, dll)
5. Ulangi untuk semua bahan yang diperlukan

#### Langkah 3: Instruksi Memasak
1. Tulis langkah-langkah memasak secara detail
2. Gunakan numbering untuk urutan yang jelas
3. Sertakan waktu memasak dan suhu jika perlu
4. Tambahkan tips khusus jika ada

#### Langkah 4: Review Nutrisi
Sistem akan otomatis menghitung:
- **Total Kalori**: Berdasarkan semua bahan
- **Protein**: Total protein per porsi
- **Karbohidrat**: Total karbohidrat per porsi
- **Lemak**: Total lemak per porsi
- **Vitamin**: Kandungan vitamin utama

#### Langkah 5: Validasi dan Simpan
1. Periksa apakah resep memenuhi standar gizi minimum:
   - Kalori minimal 600 kcal per porsi
   - Protein minimal 15 gram per porsi
   - Karbohidrat 45-65% dari total kalori
   - Lemak maksimal 30% dari total kalori

2. Jika ada yang tidak memenuhi standar, sistem akan memberikan peringatan
3. Anda bisa menyesuaikan bahan atau melanjutkan dengan catatan
4. Klik **"Simpan"** untuk menyimpan resep

### Mengedit Resep Existing
1. Klik pada nama resep yang ingin diedit
2. Klik tombol **"Edit"**
3. Lakukan perubahan yang diperlukan
4. Sistem akan membuat versi baru dari resep
5. Klik **"Simpan"** untuk menyimpan perubahan

### Melihat History Resep
1. Klik pada nama resep
2. Klik tab **"History"**
3. Anda akan melihat semua versi resep:
   - **Versi**: Nomor versi resep
   - **Tanggal**: Kapan versi dibuat
   - **Perubahan**: Apa yang diubah
   - **Dibuat oleh**: Siapa yang membuat perubahan

### Menonaktifkan Resep
1. Klik pada resep yang ingin dinonaktifkan
2. Klik tombol **"Nonaktifkan"**
3. Resep tidak akan muncul dalam menu planning
4. Data resep tetap tersimpan untuk referensi

## Perencanaan Menu Mingguan

### Membuat Menu Plan Baru
1. Klik menu **"Produksi"** → **"Menu Planning"**
2. Klik tombol **"Buat Menu Mingguan"**
3. Pilih minggu yang akan direncanakan
4. Sistem akan menampilkan kalender 7 hari

### Menambah Menu per Hari

#### Cara 1: Drag and Drop
1. Di panel kiri, lihat daftar resep yang tersedia
2. Drag resep yang diinginkan ke hari tertentu
3. Sistem akan otomatis menghitung nutrisi harian

#### Cara 2: Dropdown Selection
1. Klik pada hari tertentu
2. Klik **"Tambah Menu"**
3. Pilih resep dari dropdown
4. Masukkan jumlah porsi yang dibutuhkan
5. Klik **"Tambah"**

### Monitoring Nutrisi Harian
Untuk setiap hari, sistem menampilkan:
- **Total Kalori**: Kalori dari semua menu hari itu
- **Total Protein**: Protein dari semua menu
- **Total Karbohidrat**: Karbohidrat dari semua menu
- **Total Lemak**: Lemak dari semua menu
- **Status Compliance**: ✅ Memenuhi standar atau ❌ Tidak memenuhi

### Validasi Menu Mingguan
Sistem akan memvalidasi:
1. **Standar Gizi Harian**: Setiap hari harus memenuhi minimum nutrisi
2. **Variasi Menu**: Tidak boleh ada menu yang sama dalam 3 hari berturut-turut
3. **Ketersediaan Bahan**: Cek apakah bahan tersedia di inventori
4. **Budget Compliance**: Apakah total biaya sesuai budget

### Menyetujui Menu Mingguan
1. Setelah semua hari terisi dan validasi passed
2. Klik tombol **"Review Menu"**
3. Periksa summary nutrisi mingguan
4. Periksa total kebutuhan bahan baku
5. Klik **"Setujui Menu"** jika sudah sesuai

### Duplikasi Menu Minggu Sebelumnya
1. Klik **"Duplikasi Menu"**
2. Pilih minggu yang ingin diduplikasi
3. Sistem akan copy semua menu ke minggu baru
4. Anda bisa melakukan modifikasi sesuai kebutuhan
5. Jangan lupa validasi ulang sebelum approve

## Monitoring Kitchen Display

### Mengakses Kitchen Display
1. Klik menu **"Produksi"** → **"Kitchen Display"**
2. Pilih tab **"Cooking Display"** untuk monitor memasak
3. Pilih tab **"Packing Display"** untuk monitor packing

### Cooking Display
Menampilkan resep hari ini dengan status:
- **Belum Dimulai**: Resep yang belum mulai dimasak
- **Sedang Dimasak**: Resep yang sedang dalam proses
- **Selesai**: Resep yang sudah siap untuk packing

Informasi yang ditampilkan:
- **Nama Resep**: Makanan yang harus dibuat
- **Jumlah Porsi**: Berapa porsi yang harus dibuat
- **Bahan-bahan**: Daftar bahan dengan takaran
- **Instruksi**: Langkah-langkah memasak
- **Waktu Mulai**: Kapan mulai memasak
- **Estimasi Selesai**: Perkiraan waktu selesai

### Packing Display
Menampilkan alokasi packing per sekolah:
- **Nama Sekolah**: Tujuan pengiriman
- **Jumlah Porsi**: Porsi untuk sekolah tersebut
- **Menu Items**: Daftar makanan untuk sekolah
- **Status Packing**: Belum dikemas / Sedang dikemas / Siap kirim

### Real-time Updates
Display akan update otomatis ketika:
- Chef mulai memasak resep
- Resep selesai dimasak
- Tim packing mulai mengemas
- Packing selesai dan siap kirim

## Laporan Nutrisi

### Laporan Compliance Nutrisi
1. Klik menu **"Laporan"** → **"Laporan Nutrisi"**
2. Pilih periode (mingguan, bulanan, atau custom)
3. Sistem akan menampilkan:
   - **Compliance Rate**: Persentase menu yang memenuhi standar
   - **Average Nutrition**: Rata-rata nutrisi per hari
   - **Trend Analysis**: Trend nutrisi dari waktu ke waktu
   - **Non-compliant Days**: Hari-hari yang tidak memenuhi standar

### Laporan Variasi Menu
1. Pilih tab **"Variasi Menu"**
2. Lihat analisis:
   - **Menu Diversity Index**: Tingkat keragaman menu
   - **Repeat Frequency**: Seberapa sering menu diulang
   - **Popular Recipes**: Resep yang paling sering digunakan
   - **Underutilized Recipes**: Resep yang jarang digunakan

### Laporan Kebutuhan Bahan Baku
1. Pilih tab **"Kebutuhan Bahan"**
2. Sistem menampilkan:
   - **Weekly Requirements**: Kebutuhan bahan per minggu
   - **Monthly Projection**: Proyeksi kebutuhan bulanan
   - **Seasonal Trends**: Trend kebutuhan berdasarkan musim
   - **Cost Analysis**: Analisis biaya bahan baku

### Export Laporan
1. Setelah laporan ditampilkan, klik **"Export"**
2. Pilih format:
   - **PDF**: Untuk presentasi atau dokumentasi
   - **Excel**: Untuk analisis lebih lanjut
3. File akan diunduh otomatis

## Tips dan Best Practices

### Membuat Resep yang Baik
1. **Gunakan Bahan Lokal**: Prioritaskan bahan yang mudah didapat
2. **Perhatikan Musim**: Sesuaikan dengan ketersediaan bahan musiman
3. **Variasi Warna**: Buat menu yang menarik secara visual
4. **Tekstur Beragam**: Kombinasikan tekstur yang berbeda
5. **Rasa Seimbang**: Perhatikan keseimbangan rasa manis, asin, asam

### Perencanaan Menu Efektif
1. **Rencanakan 2 Minggu ke Depan**: Berikan waktu untuk procurement
2. **Pertimbangkan Hari Libur**: Sesuaikan dengan kalender sekolah
3. **Monitor Feedback**: Perhatikan respon dari sekolah dan siswa
4. **Rotasi Menu**: Buat siklus menu 4-6 minggu
5. **Backup Plan**: Siapkan menu alternatif jika ada masalah

### Optimasi Nutrisi
1. **Gunakan Kalkulator Nutrisi**: Manfaatkan fitur otomatis sistem
2. **Fortifikasi**: Tambahkan bahan yang kaya nutrisi
3. **Kombinasi Protein**: Gabungkan protein hewani dan nabati
4. **Sayuran Warna-warni**: Pastikan ada sayuran berbagai warna
5. **Hindari Pengolahan Berlebihan**: Pertahankan kandungan nutrisi

### Kolaborasi dengan Tim
1. **Komunikasi dengan Chef**: Pastikan resep bisa dieksekusi
2. **Koordinasi dengan Procurement**: Informasikan kebutuhan bahan
3. **Feedback dari Lapangan**: Dengarkan masukan dari driver dan sekolah
4. **Update Berkala**: Lakukan review dan update resep secara berkala

## Troubleshooting

### Resep Tidak Memenuhi Standar Gizi
**Gejala**: Sistem menampilkan warning nutrisi
**Solusi**:
1. Tambahkan bahan yang kaya nutrisi (telur, daging, kacang-kacangan)
2. Kurangi bahan yang tinggi lemak jenuh
3. Tambahkan sayuran hijau untuk vitamin
4. Sesuaikan porsi untuk mencapai kalori minimum

### Menu Planning Error
**Gejala**: Tidak bisa menyimpan menu mingguan
**Solusi**:
1. Pastikan semua hari sudah terisi menu
2. Cek apakah ada resep yang tidak aktif
3. Verifikasi ketersediaan bahan di inventori
4. Pastikan tidak ada konflik jadwal

### Kitchen Display Tidak Update
**Gejala**: Status memasak tidak berubah real-time
**Solusi**:
1. Refresh halaman browser
2. Cek koneksi internet
3. Pastikan Firebase connection aktif
4. Hubungi admin jika masalah berlanjut

### Laporan Kosong
**Gejala**: Laporan nutrisi tidak menampilkan data
**Solusi**:
1. Periksa filter tanggal
2. Pastikan ada menu yang sudah diapprove di periode tersebut
3. Coba periode yang berbeda
4. Clear browser cache

## Kontak Dukungan

### Tim Teknis
- **Email**: support@erp-sppg.com
- **Telepon**: 021-xxx-xxxx
- **WhatsApp**: 08xx-xxxx-xxxx

### Konsultasi Nutrisi
Jika membutuhkan konsultasi terkait standar gizi atau formulasi resep:
- **Email**: nutrition@erp-sppg.com
- **Telepon**: 021-xxx-xxxx

### Jam Operasional Dukungan
- **Senin - Jumat**: 08:00 - 17:00 WIB
- **Sabtu**: 08:00 - 12:00 WIB
- **Emergency**: 24/7 (untuk masalah kritis)

---

**Catatan**: Panduan ini akan terus diperbarui seiring dengan pengembangan sistem. Pastikan Anda selalu menggunakan versi terbaru dari panduan ini.