# Panduan Pengguna - Driver

## Selamat Datang di Aplikasi ERP SPPG Mobile

Panduan ini akan membantu Anda sebagai Driver untuk menggunakan aplikasi mobile ERP SPPG dalam menjalankan tugas pengiriman makanan ke sekolah-sekolah.

## Daftar Isi

1. [Download dan Install Aplikasi](#download-dan-install-aplikasi)
2. [Login dan Dashboard](#login-dan-dashboard)
3. [Melihat Tugas Pengiriman](#melihat-tugas-pengiriman)
4. [Navigasi ke Sekolah](#navigasi-ke-sekolah)
5. [Electronic Proof of Delivery (e-POD)](#electronic-proof-of-delivery-e-pod)
6. [Tracking Ompreng](#tracking-ompreng)
7. [Absensi Mobile](#absensi-mobile)
8. [Mode Offline](#mode-offline)

## Download dan Install Aplikasi

### Android
1. Buka Google Play Store
2. Cari "ERP SPPG Driver"
3. Klik "Install"
4. Tunggu hingga download selesai
5. Buka aplikasi

### iOS
1. Buka App Store
2. Cari "ERP SPPG Driver"
3. Klik "Get"
4. Tunggu hingga download selesai
5. Buka aplikasi

### PWA (Progressive Web App)
Jika tidak ada di app store, Anda bisa menggunakan versi web:
1. Buka browser di HP
2. Akses `https://erp-sppg.example.com/pwa`
3. Klik menu browser → "Add to Home Screen"
4. Aplikasi akan muncul di home screen seperti app biasa

## Login dan Dashboard

### Cara Login
1. Buka aplikasi ERP SPPG
2. Masukkan NIK atau Email Anda
3. Masukkan Password
4. Klik tombol "Masuk"

### Dashboard Driver
Setelah login, Anda akan melihat:

#### Panel Tugas Hari Ini
- **Total Sekolah**: Jumlah sekolah yang harus dikunjungi
- **Selesai**: Sekolah yang sudah selesai dikirim
- **Pending**: Sekolah yang belum dikunjungi
- **Estimasi Selesai**: Perkiraan waktu selesai semua tugas

#### Panel Status
- **Status Aktif**: Sedang dalam perjalanan atau idle
- **Lokasi Terakhir**: GPS terakhir yang tercatat
- **Waktu Update**: Kapan terakhir update lokasi
- **Koneksi**: Status koneksi internet

#### Menu Navigasi
- **Tugas**: Daftar tugas pengiriman hari ini
- **e-POD**: Form bukti pengiriman
- **Absensi**: Check-in dan check-out
- **Profil**: Informasi akun Anda

## Melihat Tugas Pengiriman

### Mengakses Daftar Tugas
1. Klik tab **"Tugas"** di menu bawah
2. Anda akan melihat daftar sekolah yang harus dikunjungi hari ini

### Informasi Tugas
Setiap tugas menampilkan:
- **Nama Sekolah**: Sekolah tujuan
- **Alamat**: Alamat lengkap sekolah
- **Jarak**: Jarak dari lokasi Anda saat ini
- **Jumlah Porsi**: Berapa porsi yang harus diantar
- **Menu**: Jenis makanan yang diantar
- **Status**: Belum dikunjungi / Dalam perjalanan / Selesai
- **Urutan**: Urutan kunjungan yang optimal

### Filter dan Pencarian
- **Filter Status**: Tampilkan berdasarkan status (semua, pending, selesai)
- **Pencarian**: Cari sekolah berdasarkan nama
- **Urutkan**: Berdasarkan jarak, urutan, atau nama

### Detail Tugas
1. Klik pada tugas tertentu untuk melihat detail:
   - **Informasi Sekolah**: Nama, alamat, kontak person
   - **Koordinat GPS**: Latitude dan longitude
   - **Menu Detail**: Daftar lengkap makanan dan jumlahnya
   - **Catatan Khusus**: Instruksi khusus untuk sekolah tersebut
   - **History**: Riwayat pengiriman sebelumnya

## Navigasi ke Sekolah

### Memulai Navigasi
1. Dari daftar tugas, klik pada sekolah tujuan
2. Klik tombol **"Mulai Navigasi"**
3. Pilih aplikasi navigasi:
   - **Google Maps** (recommended)
   - **Waze**
   - **Maps bawaan HP**

### Update Status Perjalanan
1. Saat mulai berangkat, klik **"Mulai Perjalanan"**
2. Status akan berubah menjadi "Dalam Perjalanan"
3. GPS akan mulai tracking lokasi Anda
4. Sistem akan update lokasi setiap 30 detik

### Tiba di Sekolah
1. Saat tiba di sekolah, klik **"Tiba di Lokasi"**
2. Sistem akan verifikasi lokasi GPS
3. Jika lokasi sesuai, Anda bisa lanjut ke proses e-POD
4. Jika lokasi tidak sesuai, sistem akan memberikan peringatan

## Electronic Proof of Delivery (e-POD)

### Mengakses e-POD
1. Setelah tiba di sekolah, klik **"Buat e-POD"**
2. Atau dari tab **"e-POD"** di menu bawah
3. Pilih sekolah yang sedang dikunjungi

### Langkah 1: Verifikasi Lokasi
1. Sistem akan otomatis capture GPS coordinates
2. Pastikan akurasi GPS baik (< 10 meter)
3. Jika GPS tidak akurat, tunggu beberapa saat atau pindah ke area terbuka

### Langkah 2: Input Ompreng
1. **Ompreng Diantar**: Masukkan jumlah wadah yang Anda antar
2. **Ompreng Diambil**: Masukkan jumlah wadah kotor yang Anda ambil
3. **Saldo Ompreng**: Sistem akan hitung otomatis saldo di sekolah

### Langkah 3: Foto Bukti Pengiriman
1. Klik **"Ambil Foto"**
2. Arahkan kamera ke:
   - Makanan yang diantar
   - Proses serah terima
   - Penerima dari sekolah
3. Pastikan foto jelas dan tidak blur
4. Klik **"Gunakan Foto"** jika sudah sesuai

### Langkah 4: Tanda Tangan Digital
1. Minta penerima dari sekolah untuk tanda tangan
2. Berikan HP ke penerima
3. Penerima tanda tangan di layar HP
4. Jika salah, klik **"Hapus"** dan ulangi
5. Klik **"Simpan Tanda Tangan"**

### Langkah 5: Data Penerima
1. **Nama Penerima**: Nama orang yang menerima makanan
2. **Jabatan**: Posisi di sekolah (guru, kepala sekolah, dll)
3. **Waktu Penerimaan**: Otomatis tercatat
4. **Catatan**: Tambahkan catatan jika perlu

### Langkah 6: Submit e-POD
1. Review semua data yang sudah diinput
2. Pastikan semua field sudah terisi
3. Klik **"Submit e-POD"**
4. Jika online, data langsung terkirim ke server
5. Jika offline, data disimpan dan akan sync otomatis saat online

## Tracking Ompreng

### Pentingnya Tracking Ompreng
Ompreng adalah aset perusahaan yang harus dijaga. Setiap pengiriman harus dicatat dengan benar untuk menghindari kehilangan.

### Cara Tracking
1. **Saat Mengantar**: Catat jumlah ompreng bersih yang diantar
2. **Saat Mengambil**: Catat jumlah ompreng kotor yang diambil
3. **Saldo**: Sistem hitung otomatis berapa ompreng yang tersisa di sekolah

### Contoh Perhitungan
- Ompreng di sekolah kemarin: 20 buah
- Ompreng diantar hari ini: 50 buah
- Ompreng diambil hari ini: 30 buah
- **Saldo hari ini**: 20 + 50 - 30 = 40 buah

### Alert Ompreng Hilang
Jika ada selisih ompreng yang tidak wajar:
1. Sistem akan memberikan peringatan
2. Konfirmasi ulang jumlah dengan penerima
3. Jika memang hilang, buat catatan di e-POD
4. Laporkan ke supervisor

## Absensi Mobile

### Syarat Absensi
Untuk bisa absen, Anda harus:
1. Berada di area kantor/depot
2. Terhubung ke Wi-Fi kantor
3. Menggunakan HP yang sudah terdaftar

### Check-in (Masuk Kerja)
1. Pastikan HP terhubung Wi-Fi kantor
2. Buka aplikasi ERP SPPG
3. Klik tab **"Absensi"**
4. Klik tombol **"Check-in"**
5. Sistem akan verifikasi lokasi dan Wi-Fi
6. Jika berhasil, akan muncul konfirmasi

### Check-out (Pulang Kerja)
1. Kembali ke kantor/depot
2. Pastikan terhubung Wi-Fi kantor
3. Klik tombol **"Check-out"**
4. Sistem akan hitung total jam kerja
5. Konfirmasi check-out

### Troubleshooting Absensi
**Tidak bisa check-in/out**:
1. Pastikan terhubung Wi-Fi kantor yang benar
2. Cek apakah GPS aktif
3. Pastikan berada di area kantor
4. Restart aplikasi jika perlu
5. Hubungi admin jika masih bermasalah

## Mode Offline

### Kapan Mode Offline Aktif
Mode offline otomatis aktif ketika:
- Tidak ada koneksi internet
- Sinyal lemah atau tidak stabil
- Berada di area dengan coverage buruk

### Fitur yang Tersedia Offline
1. **Lihat Tugas**: Daftar tugas yang sudah di-download
2. **Navigasi**: Menggunakan GPS offline
3. **e-POD**: Bisa buat bukti pengiriman
4. **Foto**: Bisa ambil foto bukti
5. **Tanda Tangan**: Bisa capture tanda tangan

### Fitur yang Tidak Tersedia Offline
1. **Update Real-time**: Status tidak update ke server
2. **Download Tugas Baru**: Tidak bisa dapat tugas baru
3. **Sync Data**: Data tidak tersinkronisasi
4. **Notifikasi**: Tidak menerima notifikasi baru

### Sinkronisasi Data
Ketika koneksi kembali normal:
1. Aplikasi akan otomatis sync data
2. Semua e-POD offline akan dikirim ke server
3. Status tugas akan update
4. Foto dan tanda tangan akan diupload
5. Anda akan menerima konfirmasi sync berhasil

### Tips Mode Offline
1. **Download Tugas**: Selalu download tugas saat masih ada internet
2. **Charge HP**: Pastikan baterai cukup untuk seharian
3. **Simpan Data**: Jangan hapus aplikasi saat ada data offline
4. **Sync Rutin**: Sync data setiap kali ada koneksi

## Tips dan Best Practices

### Persiapan Sebelum Berangkat
1. **Check Tugas**: Lihat semua tugas hari ini
2. **Charge HP**: Pastikan baterai penuh
3. **Download Offline**: Download tugas untuk mode offline
4. **Cek Kendaraan**: Pastikan kendaraan siap
5. **Bawa Charger**: Bawa charger mobil untuk HP

### Selama Perjalanan
1. **Update Status**: Selalu update status perjalanan
2. **Foto Berkualitas**: Ambil foto yang jelas dan terang
3. **Tanda Tangan Jelas**: Pastikan tanda tangan terbaca
4. **Catat Ompreng**: Hitung ompreng dengan teliti
5. **Komunikasi**: Hubungi sekolah jika ada masalah

### Keamanan dan Keselamatan
1. **Jangan HP Saat Nyetir**: Gunakan hands-free atau berhenti dulu
2. **Parkir Aman**: Parkir di tempat yang aman
3. **Jaga Barang**: Jangan tinggalkan HP atau barang berharga
4. **Emergency Contact**: Simpan nomor emergency di HP
5. **Lapor Masalah**: Segera lapor jika ada masalah keamanan

## Troubleshooting

### Aplikasi Tidak Bisa Login
**Gejala**: Error saat login
**Solusi**:
1. Cek koneksi internet
2. Pastikan username/password benar
3. Restart aplikasi
4. Clear cache aplikasi
5. Hubungi admin jika masih bermasalah

### GPS Tidak Akurat
**Gejala**: Lokasi tidak sesuai kenyataan
**Solusi**:
1. Pindah ke area terbuka
2. Restart GPS di HP
3. Tunggu beberapa menit untuk fix GPS
4. Restart aplikasi
5. Cek setting lokasi di HP

### Foto Tidak Bisa Diambil
**Gejala**: Kamera tidak berfungsi
**Solusi**:
1. Cek permission kamera di setting HP
2. Restart aplikasi
3. Bersihkan lensa kamera
4. Cek storage HP (mungkin penuh)
5. Restart HP jika perlu

### Data Tidak Sync
**Gejala**: e-POD tidak terkirim ke server
**Solusi**:
1. Cek koneksi internet
2. Tunggu beberapa menit
3. Buka aplikasi dan tunggu sync otomatis
4. Manual sync dengan pull-to-refresh
5. Hubungi admin jika data hilang

### Aplikasi Lemot atau Hang
**Gejala**: Aplikasi berjalan lambat
**Solusi**:
1. Close aplikasi lain yang tidak perlu
2. Restart aplikasi
3. Clear cache aplikasi
4. Restart HP
5. Update aplikasi ke versi terbaru

## Kontak Dukungan

### Tim Support Driver
- **WhatsApp**: 08xx-xxxx-xxxx (24 jam)
- **Telepon**: 021-xxx-xxxx
- **Email**: driver-support@erp-sppg.com

### Emergency Contact
Untuk situasi darurat (kecelakaan, keamanan, dll):
- **Emergency Hotline**: 08xx-xxxx-xxxx (24 jam)
- **Supervisor**: 08xx-xxxx-xxxx
- **Kantor Pusat**: 021-xxx-xxxx

### Jam Operasional Support
- **Senin - Sabtu**: 06:00 - 18:00 WIB
- **Emergency**: 24/7 untuk masalah kritis
- **WhatsApp**: Respon dalam 15 menit saat jam kerja

### Pelatihan dan Refresher
Jika membutuhkan pelatihan ulang:
- **Email**: training@erp-sppg.com
- **Telepon**: 021-xxx-xxxx
- **Jadwal**: Setiap Jumat sore atau by appointment

---

**Catatan Penting**: 
- Selalu update aplikasi ke versi terbaru
- Backup data penting secara berkala
- Laporkan bug atau masalah ke tim support
- Ikuti pelatihan refresher secara berkala