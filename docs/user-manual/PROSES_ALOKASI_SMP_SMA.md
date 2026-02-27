# Proses Alokasi Porsi untuk SMP dan SMA

## Gambaran Umum

Sekolah Menengah Pertama (SMP) dan Sekolah Menengah Atas (SMA) menggunakan satu jenis ukuran porsi standar (porsi besar) untuk semua siswa. Dokumen ini menjelaskan proses alokasi porsi untuk SMP dan SMA secara detail.

## Karakteristik Alokasi SMP/SMA

### Satu Jenis Porsi
- **Porsi Besar (Large)**: Untuk semua siswa SMP dan SMA
- **Karakteristik**: Porsi standar untuk remaja usia 13-18 tahun
- **Label di sistem**: "Besar" atau "Large"

### Data Referensi
Sistem menampilkan:
- Total jumlah siswa (`student_count`)
- Kategori sekolah (SMP atau SMA)

### Perbedaan dengan SD
- ❌ **TIDAK ADA** kolom porsi kecil
- ✓ **HANYA** satu kolom input untuk porsi besar
- ❌ **TIDAK BOLEH** mengalokasikan porsi kecil

## Proses Alokasi Langkah demi Langkah

### Langkah 1: Identifikasi Sekolah SMP/SMA

Dalam form alokasi menu, sekolah SMP/SMA ditandai dengan:
- **Kategori**: "SMP" atau "SMA"
- **Kolom Input**: Satu kolom (Besar saja)
- **Label**: "Besar" atau "Large"

**Contoh Tampilan:**
```
SMP Negeri 1 Jakarta
└─ Besar: [___] (250 siswa)

SMA Negeri 5 Bandung
└─ Besar: [___] (320 siswa)
```

### Langkah 2: Tentukan Jumlah Porsi

**Pertimbangan:**
- Total jumlah siswa
- Tingkat kehadiran rata-rata
- Buffer untuk cadangan (5-10%)
- Kebutuhan guru/staff (opsional)

**Rumus Sederhana:**
```
Porsi Besar = Total Siswa × Tingkat Kehadiran + Buffer
```

**Contoh Perhitungan:**
- Total siswa: 250 orang
- Tingkat kehadiran: 95% (238 siswa)
- Buffer 10%: 24 porsi
- **Total Porsi: 262 porsi**

**Input ke Sistem:**
```
Besar: [262]
```

### Langkah 3: Verifikasi Input

Pastikan:
- ✓ Nilai adalah angka positif
- ✓ Tidak ada input di kolom porsi kecil (kolom tidak ada)
- ✓ Total sesuai dengan perhitungan

### Langkah 4: Validasi Sistem

Sistem akan memvalidasi:
- ✓ Porsi besar > 0
- ✓ Tidak ada porsi kecil (sistem otomatis set ke 0)
- ✓ Total alokasi semua sekolah = total porsi menu

## Skenario Alokasi

### Skenario 1: SMP Ukuran Sedang

**Situasi**: SMP dengan jumlah siswa standar

**Data Sekolah:**
- Nama: SMP Negeri 3 Surabaya
- Total siswa: 180 orang
- Kategori: SMP

**Alokasi:**
```
Besar: [190]  ← 180 siswa × 95% + 10% buffer
Total: 190 porsi
```

**Hasil di Database:**
- Record 1: menu_item_id=X, school_id=3, portions=190, portion_size='large'

### Skenario 2: SMA Besar

**Situasi**: SMA dengan banyak siswa

**Data Sekolah:**
- Nama: SMA Negeri 1 Medan
- Total siswa: 450 orang
- Kategori: SMA

**Alokasi:**
```
Besar: [475]  ← 450 siswa × 95% + 10% buffer
Total: 475 porsi
```

**Hasil di Database:**
- Record 1: menu_item_id=X, school_id=1, portions=475, portion_size='large'

### Skenario 3: SMP Kecil

**Situasi**: SMP dengan jumlah siswa terbatas

**Data Sekolah:**
- Nama: SMP Negeri 12 Yogyakarta
- Total siswa: 90 orang
- Kategori: SMP

**Alokasi:**
```
Besar: [95]  ← 90 siswa × 95% + 5% buffer
Total: 95 porsi
```

**Hasil di Database:**
- Record 1: menu_item_id=X, school_id=12, portions=95, portion_size='large'

### Skenario 4: SMA dengan Kebutuhan Tambahan

**Situasi**: SMA yang juga melayani guru dan staff

**Data Sekolah:**
- Nama: SMA Negeri 8 Semarang
- Total siswa: 300 orang
- Guru dan staff: 30 orang
- Kategori: SMA

**Alokasi:**
```
Besar: [350]  ← (300 siswa + 30 staff) × 95% + 10% buffer
Total: 350 porsi
```

**Catatan**: Koordinasi dengan sekolah untuk konfirmasi kebutuhan staff.

## Validasi dan Error Handling

### Validasi yang Dilakukan Sistem

#### 1. Validasi Nilai Positif
```
❌ SALAH:
Besar: [0]
Error: "Sekolah harus memiliki minimal satu porsi"

✓ BENAR:
Besar: [150]
```

#### 2. Validasi Tidak Ada Porsi Kecil
```
Sistem otomatis memastikan:
portions_small = 0 untuk SMP/SMA

Jika ada upaya manual (via API):
❌ Error: "SMP/SMA schools cannot have small portions"
```

#### 3. Validasi Total Porsi
```
Total Porsi Menu: 1000

Alokasi:
- SMP Negeri 1: 262 (besar)
- SMP Negeri 2: 190 (besar)
- SMA Negeri 1: 475 (besar)
- SD Negeri 1: 73 (kecil + besar)
─────────────────────────────────
Total: 1000 ✓

Jika total ≠ 1000, sistem akan menampilkan error.
```

## Tips Praktis untuk Ahli Gizi

### 1. Gunakan Template Perhitungan

Buat spreadsheet sederhana untuk perhitungan cepat:

| Sekolah | Kategori | Total Siswa | Kehadiran | Buffer | Porsi Besar | Catatan |
|---------|----------|-------------|-----------|--------|-------------|---------|
| SMP N 1 | SMP      | 250         | 95%       | 10%    | 262         | -       |
| SMP N 2 | SMP      | 180         | 95%       | 10%    | 190         | -       |
| SMA N 1 | SMA      | 450         | 95%       | 10%    | 475         | -       |

### 2. Pertimbangkan Pola Kehadiran

**Kehadiran SMP/SMA biasanya lebih stabil:**
- **Senin-Kamis**: 95-98%
- **Jumat**: 92-95%
- **Menjelang ujian**: 98-100%
- **Setelah libur**: 90-93%

Sesuaikan buffer berdasarkan pola ini.

### 3. Koordinasi dengan Sekolah

**Informasi yang perlu dikonfirmasi:**
- Jumlah siswa aktif (tidak cuti/sakit)
- Acara khusus yang mempengaruhi kehadiran
- Kebutuhan tambahan untuk guru/staff
- Feedback tentang kecukupan porsi sebelumnya

### 4. Buffer yang Tepat

**Panduan Buffer:**
- **5%**: Untuk sekolah dengan kehadiran sangat stabil
- **7-8%**: Untuk kondisi normal
- **10%**: Untuk sekolah dengan kehadiran fluktuatif
- **12-15%**: Untuk hari-hari khusus (ujian, acara)

### 5. Monitor dan Sesuaikan

**Tracking yang perlu dilakukan:**
- Catat sisa porsi per sekolah
- Identifikasi pola kelebihan/kekurangan
- Bandingkan dengan data kehadiran aktual
- Sesuaikan alokasi untuk menu berikutnya

## Perbedaan SMP vs SMA

### Karakteristik SMP
- **Usia**: 13-15 tahun
- **Kebutuhan Kalori**: 2000-2400 kkal/hari
- **Porsi**: Standar besar
- **Kehadiran**: Biasanya lebih stabil

### Karakteristik SMA
- **Usia**: 16-18 tahun
- **Kebutuhan Kalori**: 2200-2600 kkal/hari
- **Porsi**: Standar besar (sama dengan SMP)
- **Kehadiran**: Dapat fluktuatif (kegiatan ekstrakurikuler)

**Catatan**: Meskipun kebutuhan kalori berbeda, ukuran porsi tetap sama. Perbedaan kalori disesuaikan melalui komposisi menu, bukan ukuran porsi.

## Integrasi dengan Proses Lain

### Dengan Menu Planning
1. Ahli gizi membuat menu dengan total porsi
2. Sistem menampilkan daftar sekolah SMP/SMA dengan satu kolom input
3. Ahli gizi mengalokasikan porsi besar
4. Sistem menyimpan satu record per sekolah dengan portion_size='large'

### Dengan KDS Cooking
1. Dapur menerima order dengan format:
   - "SMP Negeri 1: Besar (262)"
   - "SMA Negeri 5: Besar (475)"
2. Dapur menyiapkan porsi standar besar
3. Tidak ada perbedaan ukuran untuk SMP/SMA

### Dengan KDS Packing
1. Tim packing melihat detail per sekolah:
   ```
   SMP Negeri 1
   └─ Porsi Besar: 262 porsi
   
   SMA Negeri 5
   └─ Porsi Besar: 475 porsi
   ```
2. Packing dilakukan dengan kemasan standar
3. Label sekolah ditempelkan pada setiap paket

### Dengan Logistics
1. Driver menerima manifest dengan detail porsi per sekolah
2. Sekolah menandatangani penerimaan
3. Sistem mencatat delivery dengan portion_size='large'

## Troubleshooting Khusus SMP/SMA

### Masalah 1: Jumlah Siswa Tidak Akurat

**Gejala**: Data siswa di sistem tidak sesuai dengan realita

**Solusi**:
1. Hubungi admin sekolah untuk update data
2. Gunakan data manual sementara
3. Catat perbedaan untuk update database
4. Verifikasi dengan data dapodik

### Masalah 2: Porsi Terlalu Banyak/Sedikit

**Gejala**: Feedback dari sekolah tentang ketidaksesuaian porsi

**Solusi**:
1. Review data kehadiran aktual
2. Sesuaikan buffer (naik/turun 5%)
3. Koordinasi dengan kepala sekolah
4. Pertimbangkan faktor musiman (ujian, libur)
5. Update alokasi untuk menu berikutnya

### Masalah 3: Sistem Menolak Input

**Gejala**: Error saat menyimpan alokasi

**Kemungkinan Penyebab dan Solusi**:

| Error | Penyebab | Solusi |
|-------|----------|--------|
| "Sekolah harus memiliki minimal satu porsi" | Input 0 atau kosong | Isi dengan nilai > 0 |
| "Total alokasi tidak sesuai" | Jumlah total ≠ total porsi | Sesuaikan alokasi |
| "Nilai tidak valid" | Input bukan angka | Gunakan angka bulat positif |

### Masalah 4: Data Tidak Tersimpan

**Gejala**: Setelah simpan, data tidak muncul atau hilang

**Solusi**:
1. Periksa koneksi internet
2. Refresh browser dan coba lagi
3. Periksa console browser untuk error
4. Verifikasi tidak ada validasi yang gagal
5. Hubungi admin IT jika masalah berlanjut

## Checklist Alokasi SMP/SMA

Gunakan checklist ini setiap kali mengalokasikan porsi untuk SMP/SMA:

- [ ] Data jumlah siswa sudah diverifikasi
- [ ] Tingkat kehadiran sudah dipertimbangkan
- [ ] Buffer sudah ditambahkan (5-10%)
- [ ] Kebutuhan tambahan (guru/staff) sudah dipertimbangkan
- [ ] Porsi besar sudah diinput (nilai > 0)
- [ ] Tidak ada input di kolom porsi kecil (kolom tidak ada)
- [ ] Total porsi sekolah sudah dihitung
- [ ] Total alokasi semua sekolah = total porsi menu
- [ ] Validasi sistem menunjukkan status OK (hijau)
- [ ] Data sudah disimpan
- [ ] Konfirmasi penyimpanan diterima

## Contoh Kasus Lengkap

### Kasus: Menu Ayam Goreng untuk 3 SMP dan 2 SMA

**Data Menu:**
- Resep: Ayam Goreng Bumbu Kuning
- Tanggal: 20 Januari 2024
- Total Porsi: 1500

**Data Sekolah:**

| Sekolah | Kategori | Total Siswa |
|---------|----------|-------------|
| SMP N 1 | SMP      | 250         |
| SMP N 2 | SMP      | 180         |
| SMP N 3 | SMP      | 150         |
| SMA N 1 | SMA      | 450         |
| SMA N 2 | SMA      | 320         |

**Perhitungan Alokasi (dengan kehadiran 95% dan buffer 10%):**

| Sekolah | Perhitungan | Porsi Besar |
|---------|-------------|-------------|
| SMP N 1 | 250 × 0.95 × 1.10 | 262 |
| SMP N 2 | 180 × 0.95 × 1.10 | 190 |
| SMP N 3 | 150 × 0.95 × 1.10 | 158 |
| SMA N 1 | 450 × 0.95 × 1.10 | 475 |
| SMA N 2 | 320 × 0.95 × 1.10 | 338 |
| **Total** | | **1423** |

**Catatan**: Total 1423 < 1500, sisa 77 porsi untuk sekolah lain (SD).

**Input ke Sistem:**
```
SMP Negeri 1
└─ Besar: [262]

SMP Negeri 2
└─ Besar: [190]

SMP Negeri 3
└─ Besar: [158]

SMA Negeri 1
└─ Besar: [475]

SMA Negeri 2
└─ Besar: [338]
```

**Hasil di Database:**
- 5 allocation records (satu per sekolah)
- Semua dengan portion_size='large'
- Total portions: 1423 untuk 5 sekolah SMP/SMA

## Perbandingan dengan Alokasi SD

| Aspek | SD | SMP/SMA |
|-------|----|----|
| Jumlah Jenis Porsi | 2 (kecil & besar) | 1 (besar saja) |
| Kolom Input | 2 kolom | 1 kolom |
| Data Referensi | Siswa per tingkat kelas | Total siswa |
| Database Records | 0-2 per sekolah | 1 per sekolah |
| Kompleksitas | Lebih tinggi | Lebih sederhana |
| Validasi | Lebih kompleks | Lebih sederhana |

## Best Practices

### 1. Konsistensi Buffer
Gunakan persentase buffer yang konsisten untuk semua SMP/SMA dalam satu menu, kecuali ada alasan khusus.

### 2. Dokumentasi
Catat alasan jika menggunakan buffer di luar range normal (5-10%):
- Acara khusus
- Feedback sebelumnya
- Permintaan sekolah
- Kondisi cuaca/musim

### 3. Komunikasi Proaktif
Hubungi sekolah sebelum alokasi jika:
- Ada perubahan signifikan dari alokasi sebelumnya
- Sekolah baru ditambahkan
- Ada feedback negatif sebelumnya

### 4. Review Berkala
Setiap bulan, review:
- Akurasi alokasi vs kehadiran aktual
- Pola sisa porsi per sekolah
- Feedback dari sekolah
- Penyesuaian yang perlu dilakukan

### 5. Koordinasi Tim
Pastikan koordinasi dengan:
- Tim dapur (kapasitas produksi)
- Tim packing (kapasitas packing)
- Tim logistics (kapasitas delivery)
- Admin sekolah (konfirmasi data)

## FAQ Khusus SMP/SMA

**Q: Mengapa SMP dan SMA menggunakan porsi yang sama?**
A: Porsi besar dirancang untuk memenuhi kebutuhan nutrisi remaja usia 13-18 tahun. Perbedaan kebutuhan kalori disesuaikan melalui komposisi menu, bukan ukuran porsi.

**Q: Apakah bisa mengalokasikan porsi kecil untuk SMP/SMA?**
A: Tidak. Sistem akan menolak alokasi porsi kecil untuk SMP/SMA karena tidak sesuai dengan kebutuhan nutrisi remaja.

**Q: Bagaimana jika SMP/SMA meminta porsi lebih besar?**
A: Koordinasi dengan tim nutrisi untuk evaluasi. Jika perlu, sesuaikan komposisi menu atau tambahkan buffer, bukan mengubah ukuran porsi standar.

**Q: Apakah guru dan staff dihitung dalam alokasi?**
A: Tergantung kebijakan. Jika sekolah meminta, tambahkan jumlah guru/staff dalam perhitungan. Catat dalam dokumentasi.

**Q: Bagaimana menangani sekolah dengan kehadiran sangat fluktuatif?**
A: Gunakan buffer lebih tinggi (12-15%) dan monitor ketat. Koordinasi dengan sekolah untuk prediksi kehadiran yang lebih akurat.

---

**Versi Dokumen**: 1.0  
**Terakhir Diperbarui**: 2024  
**Untuk**: Ahli Gizi
