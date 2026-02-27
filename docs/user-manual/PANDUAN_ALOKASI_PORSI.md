# Panduan Alokasi Porsi Berdasarkan Ukuran

## Pengenalan

Sistem ERP SPPG kini mendukung diferensiasi ukuran porsi untuk sekolah dasar (SD) berdasarkan kelompok usia siswa. Fitur ini memungkinkan ahli gizi untuk mengalokasikan porsi makanan dengan lebih akurat sesuai kebutuhan nutrisi siswa.

## Jenis Ukuran Porsi

### Porsi Kecil (Small)
- **Untuk**: Siswa SD kelas 1-3
- **Karakteristik**: Porsi yang disesuaikan untuk anak usia 6-9 tahun
- **Label di sistem**: "Kecil (Kelas 1-3)"

### Porsi Besar (Large)
- **Untuk**: Siswa SD kelas 4-6, SMP, dan SMA
- **Karakteristik**: Porsi standar untuk anak usia 10 tahun ke atas
- **Label di sistem**: "Besar (Kelas 4-6)" atau "Besar"

## Kategori Sekolah

### Sekolah SD (Mixed Portion)
- Memerlukan **dua jenis porsi**: kecil dan besar
- Sistem menampilkan dua kolom input untuk alokasi
- Jumlah siswa kelas 1-3 dan 4-6 ditampilkan sebagai referensi

### Sekolah SMP/SMA (Single Portion)
- Hanya memerlukan **porsi besar**
- Sistem menampilkan satu kolom input untuk alokasi
- Tidak dapat mengalokasikan porsi kecil

## Cara Mengalokasikan Porsi

### Langkah 1: Buat Menu Item Baru
1. Buka halaman **Menu Planning**
2. Klik tombol **"Tambah Menu Item"**
3. Pilih resep dan tanggal
4. Masukkan total porsi yang akan dibuat

### Langkah 2: Alokasi untuk Sekolah SD
1. Temukan sekolah SD dalam daftar
2. Lihat jumlah siswa kelas 1-3 dan 4-6 sebagai referensi
3. Masukkan jumlah porsi kecil di kolom **"Kecil (Kelas 1-3)"**
4. Masukkan jumlah porsi besar di kolom **"Besar (Kelas 4-6)"**
5. Sistem akan menghitung total otomatis

**Contoh:**
- SD Negeri 1 memiliki 80 siswa kelas 1-3 dan 100 siswa kelas 4-6
- Alokasi: 80 porsi kecil + 100 porsi besar = 180 total

### Langkah 3: Alokasi untuk Sekolah SMP/SMA
1. Temukan sekolah SMP/SMA dalam daftar
2. Lihat total jumlah siswa sebagai referensi
3. Masukkan jumlah porsi di kolom **"Besar"**
4. Kolom porsi kecil tidak tersedia untuk SMP/SMA

**Contoh:**
- SMP Negeri 1 memiliki 150 siswa
- Alokasi: 150 porsi besar

### Langkah 4: Validasi Total
1. Sistem menampilkan **total alokasi** secara real-time
2. Pastikan total alokasi = total porsi yang dibuat
3. Jika tidak sesuai, sistem menampilkan pesan error
4. Tombol **"Simpan"** hanya aktif jika validasi berhasil

### Langkah 5: Simpan Alokasi
1. Periksa kembali semua alokasi
2. Klik tombol **"Simpan"**
3. Sistem akan menyimpan alokasi untuk semua sekolah

## Validasi Sistem

### Validasi Otomatis
Sistem akan memvalidasi hal-hal berikut:

1. **Total Porsi**: Jumlah semua porsi kecil + besar harus sama dengan total porsi
2. **Porsi Non-Negatif**: Semua nilai harus ≥ 0
3. **Minimal Satu Porsi**: Setiap sekolah harus mendapat minimal 1 porsi
4. **Tipe Sekolah**: SMP/SMA tidak boleh memiliki porsi kecil

### Pesan Error Umum

| Pesan Error | Penyebab | Solusi |
|-------------|----------|--------|
| "Jumlah alokasi tidak sesuai dengan total porsi" | Total alokasi ≠ total porsi | Sesuaikan alokasi hingga total sama |
| "SMP/SMA tidak dapat memiliki porsi kecil" | Mengisi porsi kecil untuk SMP/SMA | Kosongkan kolom porsi kecil |
| "Sekolah harus memiliki minimal satu porsi" | Semua kolom bernilai 0 | Isi minimal satu kolom dengan nilai > 0 |
| "Nilai porsi tidak boleh negatif" | Memasukkan angka negatif | Gunakan angka positif atau 0 |

## Mengedit Alokasi Porsi

### Langkah Edit
1. Buka menu item yang ingin diedit
2. Klik tombol **"Edit"**
3. Sistem menampilkan alokasi saat ini
4. Ubah nilai porsi kecil atau besar sesuai kebutuhan
5. Validasi otomatis akan berjalan
6. Klik **"Simpan"** untuk menyimpan perubahan

### Catatan Penting
- Perubahan alokasi akan menghapus data lama dan membuat data baru
- Pastikan total porsi tetap sesuai setelah edit
- Perubahan akan langsung terlihat di KDS

## Melihat Statistik Porsi

Sistem menampilkan statistik real-time:

- **Total Porsi Kecil**: Jumlah semua porsi kecil di semua SD
- **Total Porsi Besar**: Jumlah semua porsi besar di semua sekolah
- **Persentase**: Distribusi porsi kecil vs besar
- **Jumlah Sekolah**: Berapa sekolah menerima setiap jenis porsi

## Tips dan Best Practices

### 1. Gunakan Jumlah Siswa sebagai Referensi
- Sistem menampilkan jumlah siswa per tingkat kelas
- Gunakan angka ini sebagai panduan alokasi
- Sesuaikan dengan kehadiran dan kebutuhan aktual

### 2. Pertimbangkan Buffer
- Tambahkan 5-10% buffer untuk cadangan
- Contoh: 80 siswa → alokasi 85-88 porsi

### 3. Periksa Total Sebelum Simpan
- Selalu verifikasi total alokasi = total porsi
- Gunakan indikator visual (hijau = valid, merah = error)

### 4. Koordinasi dengan Dapur
- Pastikan dapur memahami perbedaan ukuran porsi
- Komunikasikan perubahan alokasi yang signifikan

### 5. Monitor Feedback
- Perhatikan feedback dari sekolah tentang kecukupan porsi
- Sesuaikan alokasi berdasarkan kebutuhan aktual

## Integrasi dengan KDS

### KDS Cooking View
- Menampilkan breakdown porsi per sekolah
- Format: "SD Negeri 1: Kecil (80), Besar (100)"
- Dapur dapat melihat total porsi kecil dan besar yang harus disiapkan

### KDS Packing View
- Menampilkan detail packing per sekolah
- Memisahkan porsi kecil dan besar
- Memudahkan tim packing untuk menyiapkan jumlah yang tepat

## Troubleshooting

### Masalah: Tombol Simpan Tidak Aktif
**Penyebab**: Validasi gagal
**Solusi**: 
1. Periksa pesan error di layar
2. Pastikan total alokasi = total porsi
3. Pastikan tidak ada nilai negatif
4. Pastikan setiap sekolah memiliki minimal 1 porsi

### Masalah: Tidak Bisa Mengisi Porsi Kecil untuk SMP
**Penyebab**: SMP/SMA hanya menerima porsi besar
**Solusi**: Ini adalah perilaku yang benar. Hanya isi kolom porsi besar.

### Masalah: Data Tidak Tersimpan
**Penyebab**: Error server atau koneksi
**Solusi**:
1. Periksa koneksi internet
2. Refresh halaman dan coba lagi
3. Hubungi admin jika masalah berlanjut

## Pertanyaan Umum (FAQ)

**Q: Apakah semua SD harus memiliki porsi kecil dan besar?**
A: Tidak wajib. SD dapat memiliki hanya porsi kecil, hanya porsi besar, atau keduanya. Minimal satu jenis porsi harus > 0.

**Q: Bagaimana jika SD hanya memiliki kelas 1-3?**
A: Isi hanya kolom porsi kecil, kosongkan kolom porsi besar (atau isi 0).

**Q: Apakah porsi kecil dan besar berbeda secara fisik?**
A: Ya, dapur harus menyiapkan porsi dengan ukuran berbeda sesuai label. Koordinasi dengan kepala dapur untuk standar ukuran.

**Q: Bagaimana cara melihat history alokasi?**
A: Saat ini sistem menyimpan alokasi terbaru. Untuk history, hubungi admin untuk export data.

**Q: Apakah bisa mengubah alokasi setelah disimpan?**
A: Ya, gunakan fitur edit. Perubahan akan mengganti data lama dengan data baru.

## Kontak Dukungan

Jika mengalami masalah atau memiliki pertanyaan:
- **Email**: support@sppg.id
- **Telepon**: (021) XXX-XXXX
- **WhatsApp**: +62 XXX-XXXX-XXXX

---

**Versi Dokumen**: 1.0  
**Terakhir Diperbarui**: 2024  
**Untuk**: Ahli Gizi dan Menu Planner
