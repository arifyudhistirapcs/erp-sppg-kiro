# FAQ - Sistem ERP SPPG

## Frequently Asked Questions

Dokumen ini berisi pertanyaan yang sering diajukan tentang penggunaan Sistem ERP SPPG beserta jawabannya.

## Daftar Isi

1. [Umum](#umum)
2. [Login dan Akses](#login-dan-akses)
3. [Resep dan Menu Planning](#resep-dan-menu-planning)
4. [Pengadaan dan Inventori](#pengadaan-dan-inventori)
5. [Pengiriman dan e-POD](#pengiriman-dan-e-pod)
6. [Laporan dan Keuangan](#laporan-dan-keuangan)
7. [Teknis dan Troubleshooting](#teknis-dan-troubleshooting)

## Umum

### Q: Apa itu Sistem ERP SPPG?
**A:** Sistem ERP SPPG adalah platform manajemen operasional terintegrasi untuk mengelola seluruh siklus program pemenuhan gizi, mulai dari perencanaan menu, pengadaan bahan baku, produksi, hingga distribusi ke sekolah-sekolah penerima manfaat.

### Q: Siapa saja yang bisa menggunakan sistem ini?
**A:** Sistem ini digunakan oleh:
- Kepala SPPG/Yayasan
- Akuntan
- Ahli Gizi
- Staff Pengadaan
- Chef/Tukang Masak
- Tim Packing
- Driver
- Asisten Lapangan

### Q: Apakah sistem ini bisa diakses dari HP?
**A:** Ya, tersedia aplikasi mobile (PWA) khusus untuk Driver dan fitur absensi karyawan. Untuk fungsi lainnya, bisa diakses melalui browser HP tapi lebih optimal menggunakan komputer/laptop.

### Q: Apakah data aman di sistem ini?
**A:** Ya, sistem menggunakan enkripsi HTTPS, backup otomatis harian, dan audit trail lengkap untuk memastikan keamanan dan integritas data.

## Login dan Akses

### Q: Bagaimana cara mendapatkan akun untuk login?
**A:** Akun dibuat oleh admin sistem (biasanya Kepala SPPG atau staff HRM). Setelah data karyawan diinput, sistem akan generate username dan password yang akan diberikan kepada Anda.

### Q: Lupa password, bagaimana cara reset?
**A:** Hubungi admin sistem atau Kepala SPPG untuk reset password. Untuk keamanan, reset password tidak bisa dilakukan sendiri.

### Q: Kenapa tidak bisa login padahal username/password sudah benar?
**A:** Kemungkinan penyebab:
- Akun sudah dinonaktifkan
- Terlalu banyak percobaan login yang salah (akun ter-lock)
- Masalah koneksi internet
- Browser cache perlu dibersihkan
Hubungi admin sistem untuk bantuan.

### Q: Berapa lama session login bertahan?
**A:** Session akan expired setelah 30 menit tidak ada aktivitas. Anda perlu login ulang setelah itu.

### Q: Bisa login dari beberapa device sekaligus?
**A:** Ya, bisa login dari beberapa device, tapi untuk keamanan disarankan logout dari device yang tidak digunakan.

## Resep dan Menu Planning

### Q: Siapa yang bisa membuat dan mengedit resep?
**A:** Hanya Ahli Gizi yang memiliki akses untuk membuat, mengedit, dan mengelola resep.

### Q: Bagaimana sistem menghitung nilai gizi resep?
**A:** Sistem otomatis menghitung berdasarkan database nutrisi bahan baku. Setiap bahan memiliki data kalori, protein, karbohidrat, lemak, dan vitamin per 100 gram.

### Q: Apa yang terjadi jika resep tidak memenuhi standar gizi?
**A:** Sistem akan memberikan peringatan, tapi resep masih bisa disimpan dengan catatan. Ahli Gizi perlu menyesuaikan komposisi bahan untuk memenuhi standar.

### Q: Bisa menggunakan resep lama untuk menu baru?
**A:** Ya, ada fitur duplikasi menu mingguan dan template resep yang memudahkan penggunaan ulang.

### Q: Bagaimana jika ada perubahan mendadak pada menu?
**A:** Menu yang sudah diapprove masih bisa diubah oleh Ahli Gizi, tapi perubahan akan tercatat di audit trail dan perlu koordinasi dengan tim produksi.

### Q: Apakah sistem bisa suggest menu berdasarkan bahan yang tersedia?
**A:** Saat ini belum ada fitur otomatis, tapi sistem menampilkan status ketersediaan bahan saat planning menu.

## Pengadaan dan Inventori

### Q: Siapa yang bisa membuat Purchase Order?
**A:** Staff Pengadaan yang membuat PO, tapi perlu approval dari Kepala SPPG sebelum dikirim ke supplier.

### Q: Bagaimana cara menambah supplier baru?
**A:** Staff Pengadaan bisa menambah supplier baru melalui menu Pengadaan → Supplier → Tambah Supplier.

### Q: Kenapa stok tidak update otomatis setelah terima barang?
**A:** Stok akan update otomatis setelah GRN (Goods Receipt Note) dibuat dan dikonfirmasi. Pastikan proses penerimaan barang sudah selesai.

### Q: Bagaimana cara setting alert stok menipis?
**A:** Kepala SPPG bisa mengatur minimum threshold untuk setiap bahan di menu Pengaturan → Konfigurasi Sistem → Pengaturan Inventori.

### Q: Apa itu FIFO dan FEFO dalam sistem?
**A:** 
- **FIFO (First In First Out)**: Barang yang masuk duluan digunakan duluan
- **FEFO (First Expired First Out)**: Barang yang expired duluan digunakan duluan
Sistem otomatis menerapkan metode ini untuk mengurangi waste.

### Q: Bisa track barang dari supplier mana yang digunakan?
**A:** Ya, sistem mencatat batch dan supplier untuk setiap penerimaan barang, sehingga bisa dilacak asal-usulnya.

## Pengiriman dan e-POD

### Q: Bagaimana cara driver mendapat tugas pengiriman?
**A:** Staff Logistik membuat dan assign tugas pengiriman ke driver melalui sistem. Driver akan melihat tugasnya di aplikasi mobile.

### Q: Apa yang harus dilakukan jika HP driver rusak saat pengiriman?
**A:** Driver bisa:
1. Pinjam HP rekan kerja dan login dengan akun sendiri
2. Catat manual bukti pengiriman dan input ke sistem setelah kembali
3. Hubungi supervisor untuk bantuan

### Q: Bagaimana jika sekolah menolak menerima makanan?
**A:** Driver harus:
1. Tanyakan alasan penolakan
2. Foto kondisi makanan
3. Hubungi supervisor
4. Buat catatan di e-POD
5. Bawa kembali makanan ke dapur

### Q: Apa yang terjadi jika ompreng hilang di sekolah?
**A:** 
1. Driver catat selisih di e-POD dengan keterangan
2. Sistem akan flag sekolah tersebut
3. Follow up oleh staff logistik
4. Jika perlu, lakukan penagihan ke sekolah

### Q: Bisa buat e-POD tanpa internet?
**A:** Ya, aplikasi mobile bisa bekerja offline. Data akan tersimpan di HP dan otomatis sync saat ada koneksi internet.

### Q: Bagaimana jika GPS tidak akurat saat buat e-POD?
**A:** 
1. Pindah ke area terbuka
2. Tunggu beberapa menit untuk GPS fix
3. Jika masih tidak akurat, buat catatan di e-POD
4. Laporkan ke supervisor

## Laporan dan Keuangan

### Q: Siapa yang bisa akses laporan keuangan?
**A:** Kepala SPPG, Kepala Yayasan, dan Akuntan memiliki akses penuh. Role lain hanya bisa melihat laporan yang relevan dengan tugasnya.

### Q: Seberapa sering laporan keuangan di-update?
**A:** Laporan real-time untuk transaksi harian. Laporan bulanan biasanya finalized di awal bulan berikutnya.

### Q: Bagaimana cara export laporan ke Excel?
**A:** Di setiap halaman laporan, ada tombol "Export" yang memungkinkan download dalam format PDF atau Excel.

### Q: Apakah bisa custom laporan sesuai kebutuhan?
**A:** Saat ini tersedia laporan standar. Untuk custom laporan, bisa request ke tim teknis dengan spesifikasi yang dibutuhkan.

### Q: Bagaimana cara melihat trend pengeluaran bulanan?
**A:** Di dashboard Kepala Yayasan atau menu Laporan Keuangan, tersedia grafik trend yang menampilkan pola pengeluaran dari waktu ke waktu.

### Q: Apa bedanya cash flow dan budget report?
**A:** 
- **Cash Flow**: Catatan actual pemasukan dan pengeluaran
- **Budget Report**: Perbandingan antara budget yang dialokasikan vs pengeluaran actual

## Teknis dan Troubleshooting

### Q: Browser apa yang direkomendasikan?
**A:** Chrome, Firefox, atau Safari versi terbaru. Hindari Internet Explorer karena tidak didukung.

### Q: Kenapa halaman loading lama?
**A:** Kemungkinan penyebab:
- Koneksi internet lambat
- Server sedang maintenance
- Browser cache penuh (coba clear cache)
- Terlalu banyak tab browser terbuka

### Q: Bagaimana cara clear browser cache?
**A:** 
- **Chrome**: Ctrl+Shift+Delete → pilih "Cached images and files"
- **Firefox**: Ctrl+Shift+Delete → pilih "Cache"
- **Safari**: Cmd+Option+E

### Q: Data yang sudah diinput hilang, bagaimana?
**A:** 
1. Cek di audit trail apakah ada yang menghapus
2. Hubungi admin sistem untuk cek backup
3. Jika memang hilang, bisa restore dari backup harian

### Q: Sistem error/crash, apa yang harus dilakukan?
**A:** 
1. Screenshot error message
2. Catat apa yang sedang dilakukan saat error
3. Refresh halaman atau restart browser
4. Jika masih error, hubungi tim teknis dengan screenshot

### Q: Bagaimana cara update password?
**A:** Saat ini belum ada fitur self-service change password. Hubungi admin sistem untuk update password.

### Q: Apakah sistem backup otomatis?
**A:** Ya, sistem melakukan backup database otomatis setiap hari jam 3 pagi. Backup disimpan selama 30 hari.

### Q: Bagaimana jika lupa logout dan HP/laptop hilang?
**A:** Segera hubungi admin sistem untuk force logout semua session akun Anda dan ganti password.

### Q: Bisa akses sistem dari luar negeri?
**A:** Ya, sistem bisa diakses dari mana saja selama ada koneksi internet. Tapi untuk keamanan, akses dari lokasi tidak biasa mungkin perlu verifikasi tambahan.

### Q: Kenapa notifikasi tidak muncul?
**A:** 
1. Cek setting browser untuk allow notifications
2. Pastikan tidak ada ad-blocker yang memblokir
3. Refresh halaman
4. Cek di menu notifikasi sistem

### Q: Bagaimana cara melaporkan bug atau request fitur baru?
**A:** Hubungi tim teknis melalui:
- Email: support@erp-sppg.com
- WhatsApp: 08xx-xxxx-xxxx
- Atau melalui supervisor Anda

## Kontak Support

### Tim Teknis
- **Email**: support@erp-sppg.com
- **WhatsApp**: 08xx-xxxx-xxxx
- **Telepon**: 021-xxx-xxxx

### Admin Sistem
- **Email**: admin@erp-sppg.com
- **Telepon**: 021-xxx-xxxx

### Jam Operasional Support
- **Senin - Jumat**: 08:00 - 17:00 WIB
- **Sabtu**: 08:00 - 12:00 WIB
- **Emergency**: 24/7 untuk masalah kritis

### Escalation
Jika masalah tidak terselesaikan:
1. **Level 1**: Tim Support
2. **Level 2**: Senior Technical
3. **Level 3**: IT Manager
4. **Level 4**: Management

---

**Catatan**: FAQ ini akan terus diperbarui berdasarkan pertanyaan yang masuk. Jika pertanyaan Anda tidak ada di sini, jangan ragu untuk menghubungi tim support.