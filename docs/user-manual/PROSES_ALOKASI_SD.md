# Proses Alokasi Porsi untuk Sekolah Dasar (SD)

## Gambaran Umum

Sekolah Dasar (SD) memerlukan dua jenis ukuran porsi karena perbedaan kebutuhan nutrisi antara siswa kelas rendah (1-3) dan kelas tinggi (4-6). Dokumen ini menjelaskan proses alokasi porsi untuk SD secara detail.

## Karakteristik Alokasi SD

### Dua Jenis Porsi
1. **Porsi Kecil**: Untuk siswa kelas 1-3 (usia 6-9 tahun)
2. **Porsi Besar**: Untuk siswa kelas 4-6 (usia 10-12 tahun)

### Data Referensi
Sistem menampilkan:
- Jumlah siswa kelas 1-3 (`student_count_grade_1_3`)
- Jumlah siswa kelas 4-6 (`student_count_grade_4_6`)
- Total siswa sekolah

## Proses Alokasi Langkah demi Langkah

### Langkah 1: Identifikasi Sekolah SD

Dalam form alokasi menu, sekolah SD ditandai dengan:
- **Kategori**: "SD" atau "Sekolah Dasar"
- **Kolom Input**: Dua kolom (Kecil dan Besar)
- **Label**: "Kecil (Kelas 1-3)" dan "Besar (Kelas 4-6)"

**Contoh Tampilan:**
```
SD Negeri 1 Jakarta
├─ Kecil (Kelas 1-3): [___] (80 siswa)
└─ Besar (Kelas 4-6): [___] (100 siswa)
```

### Langkah 2: Tentukan Jumlah Porsi Kecil

**Pertimbangan:**
- Jumlah siswa kelas 1-3
- Tingkat kehadiran rata-rata
- Buffer untuk cadangan (5-10%)

**Rumus Sederhana:**
```
Porsi Kecil = Siswa Kelas 1-3 × Tingkat Kehadiran + Buffer
```

**Contoh Perhitungan:**
- Siswa kelas 1-3: 80 orang
- Tingkat kehadiran: 95% (76 siswa)
- Buffer 10%: 8 porsi
- **Total Porsi Kecil: 84 porsi**

**Input ke Sistem:**
```
Kecil (Kelas 1-3): [84]
```

### Langkah 3: Tentukan Jumlah Porsi Besar

**Pertimbangan:**
- Jumlah siswa kelas 4-6
- Tingkat kehadiran rata-rata
- Buffer untuk cadangan (5-10%)

**Rumus Sederhana:**
```
Porsi Besar = Siswa Kelas 4-6 × Tingkat Kehadiran + Buffer
```

**Contoh Perhitungan:**
- Siswa kelas 4-6: 100 orang
- Tingkat kehadiran: 95% (95 siswa)
- Buffer 10%: 10 porsi
- **Total Porsi Besar: 105 porsi**

**Input ke Sistem:**
```
Besar (Kelas 4-6): [105]
```

### Langkah 4: Verifikasi Total

**Total untuk SD Negeri 1:**
```
Porsi Kecil:  84
Porsi Besar: 105
─────────────────
Total:       189 porsi
```

Sistem akan menampilkan total ini secara otomatis dan menambahkannya ke total alokasi keseluruhan.

### Langkah 5: Validasi Sistem

Sistem akan memvalidasi:
- ✓ Porsi kecil ≥ 0
- ✓ Porsi besar ≥ 0
- ✓ Minimal satu jenis porsi > 0
- ✓ Total alokasi semua sekolah = total porsi menu

## Skenario Alokasi

### Skenario 1: SD dengan Kedua Jenis Porsi

**Situasi**: SD lengkap dengan kelas 1-6

**Data Sekolah:**
- Nama: SD Negeri 5 Bandung
- Siswa kelas 1-3: 120 orang
- Siswa kelas 4-6: 150 orang

**Alokasi:**
```
Kecil (Kelas 1-3): [125]  ← 120 siswa + 5 buffer
Besar (Kelas 4-6): [160]  ← 150 siswa + 10 buffer
Total: 285 porsi
```

**Hasil di Database:**
- Record 1: menu_item_id=X, school_id=5, portions=125, portion_size='small'
- Record 2: menu_item_id=X, school_id=5, portions=160, portion_size='large'

### Skenario 2: SD Hanya Kelas Rendah (1-3)

**Situasi**: SD kecil yang hanya memiliki kelas 1-3

**Data Sekolah:**
- Nama: SD Negeri 12 Surabaya
- Siswa kelas 1-3: 60 orang
- Siswa kelas 4-6: 0 orang

**Alokasi:**
```
Kecil (Kelas 1-3): [65]   ← 60 siswa + 5 buffer
Besar (Kelas 4-6): [0]    ← Tidak ada siswa
Total: 65 porsi
```

**Hasil di Database:**
- Record 1: menu_item_id=X, school_id=12, portions=65, portion_size='small'
- (Tidak ada record untuk portion_size='large' karena 0)

### Skenario 3: SD Hanya Kelas Tinggi (4-6)

**Situasi**: SD yang sedang transisi, hanya memiliki kelas 4-6

**Data Sekolah:**
- Nama: SD Negeri 8 Medan
- Siswa kelas 1-3: 0 orang
- Siswa kelas 4-6: 90 orang

**Alokasi:**
```
Kecil (Kelas 1-3): [0]    ← Tidak ada siswa
Besar (Kelas 4-6): [95]   ← 90 siswa + 5 buffer
Total: 95 porsi
```

**Hasil di Database:**
- Record 1: menu_item_id=X, school_id=8, portions=95, portion_size='large'
- (Tidak ada record untuk portion_size='small' karena 0)

### Skenario 4: SD dengan Distribusi Tidak Merata

**Situasi**: SD dengan lebih banyak siswa kelas rendah

**Data Sekolah:**
- Nama: SD Negeri 3 Yogyakarta
- Siswa kelas 1-3: 180 orang
- Siswa kelas 4-6: 80 orang

**Alokasi:**
```
Kecil (Kelas 1-3): [190]  ← 180 siswa + 10 buffer
Besar (Kelas 4-6): [85]   ← 80 siswa + 5 buffer
Total: 275 porsi
```

**Catatan**: Distribusi tidak harus seimbang, sesuaikan dengan jumlah siswa aktual.

## Validasi dan Error Handling

### Validasi yang Dilakukan Sistem

#### 1. Validasi Nilai Non-Negatif
```
❌ SALAH:
Kecil (Kelas 1-3): [-10]
Besar (Kelas 4-6): [100]

✓ BENAR:
Kecil (Kelas 1-3): [0]
Besar (Kelas 4-6): [100]
```

#### 2. Validasi Minimal Satu Porsi
```
❌ SALAH:
Kecil (Kelas 1-3): [0]
Besar (Kelas 4-6): [0]
Error: "Sekolah harus memiliki minimal satu porsi"

✓ BENAR:
Kecil (Kelas 1-3): [0]
Besar (Kelas 4-6): [50]
```

#### 3. Validasi Total Porsi
```
Total Porsi Menu: 500

Alokasi:
- SD Negeri 1: 189 (84 kecil + 105 besar)
- SD Negeri 2: 156 (70 kecil + 86 besar)
- SMP Negeri 1: 155 (besar)
─────────────────────────────────────────
Total: 500 ✓

Jika total ≠ 500, sistem akan menampilkan error.
```

## Tips Praktis untuk Ahli Gizi

### 1. Gunakan Template Perhitungan

Buat spreadsheet sederhana untuk perhitungan cepat:

| Sekolah | Siswa 1-3 | Kehadiran | Buffer | Porsi Kecil | Siswa 4-6 | Kehadiran | Buffer | Porsi Besar | Total |
|---------|-----------|-----------|--------|-------------|-----------|-----------|--------|-------------|-------|
| SD N 1  | 80        | 95%       | 10%    | 84          | 100       | 95%       | 10%    | 105         | 189   |
| SD N 2  | 65        | 95%       | 10%    | 68          | 80        | 95%       | 10%    | 88          | 156   |

### 2. Pertimbangkan Pola Kehadiran

- **Senin**: Kehadiran biasanya lebih rendah (90-92%)
- **Selasa-Kamis**: Kehadiran optimal (95-98%)
- **Jumat**: Kehadiran sedang (92-95%)

Sesuaikan buffer berdasarkan hari.

### 3. Komunikasi dengan Sekolah

Sebelum alokasi:
- Konfirmasi jumlah siswa aktif per tingkat kelas
- Tanyakan ada tidaknya acara khusus yang mempengaruhi kehadiran
- Catat feedback tentang kecukupan porsi sebelumnya

### 4. Monitor dan Sesuaikan

Setelah distribusi:
- Catat sisa porsi per sekolah
- Identifikasi pola kelebihan/kekurangan
- Sesuaikan alokasi untuk menu berikutnya

### 5. Dokumentasi

Simpan catatan:
- Alokasi per sekolah per hari
- Feedback dari sekolah
- Penyesuaian yang dilakukan
- Alasan perubahan alokasi

## Integrasi dengan Proses Lain

### Dengan Menu Planning
1. Ahli gizi membuat menu dengan total porsi
2. Sistem menampilkan daftar sekolah SD dengan dua kolom input
3. Ahli gizi mengalokasikan porsi kecil dan besar
4. Sistem menyimpan dua record per SD (jika keduanya > 0)

### Dengan KDS Cooking
1. Dapur menerima order dengan breakdown:
   - "SD Negeri 1: Kecil (84), Besar (105)"
2. Dapur menyiapkan porsi sesuai ukuran
3. Total porsi kecil dan besar ditampilkan terpisah

### Dengan KDS Packing
1. Tim packing melihat detail per sekolah:
   ```
   SD Negeri 1
   ├─ Porsi Kecil (Kelas 1-3): 84 porsi
   └─ Porsi Besar (Kelas 4-6): 105 porsi
   ```
2. Packing dilakukan dengan label berbeda untuk setiap ukuran
3. Sekolah menerima dua jenis kemasan

### Dengan Logistics
1. Driver menerima manifest dengan detail porsi
2. Sekolah menandatangani penerimaan untuk kedua jenis porsi
3. Sistem mencatat delivery per portion_size

## Troubleshooting Khusus SD

### Masalah 1: Jumlah Siswa Tidak Akurat

**Gejala**: Data siswa di sistem tidak sesuai dengan realita

**Solusi**:
1. Hubungi admin sekolah untuk update data
2. Gunakan data manual sementara
3. Catat perbedaan untuk update database

### Masalah 2: Porsi Kecil Terlalu Banyak/Sedikit

**Gejala**: Feedback dari sekolah tentang ketidaksesuaian porsi

**Solusi**:
1. Review data kehadiran aktual
2. Sesuaikan buffer (naik/turun 5%)
3. Koordinasi dengan kepala sekolah
4. Update alokasi untuk menu berikutnya

### Masalah 3: Tidak Bisa Membedakan Porsi di Lapangan

**Gejala**: Sekolah kesulitan membedakan porsi kecil dan besar

**Solusi**:
1. Gunakan label warna berbeda (misal: biru untuk kecil, merah untuk besar)
2. Gunakan kemasan berbeda ukuran
3. Tambahkan stiker "Kelas 1-3" dan "Kelas 4-6"
4. Koordinasi dengan tim packing untuk standarisasi

### Masalah 4: Data Tidak Tersimpan dengan Benar

**Gejala**: Setelah simpan, data porsi tidak muncul atau salah

**Solusi**:
1. Periksa koneksi internet
2. Refresh browser dan coba lagi
3. Periksa console browser untuk error
4. Hubungi admin IT jika masalah berlanjut

## Checklist Alokasi SD

Gunakan checklist ini setiap kali mengalokasikan porsi untuk SD:

- [ ] Data jumlah siswa kelas 1-3 sudah diverifikasi
- [ ] Data jumlah siswa kelas 4-6 sudah diverifikasi
- [ ] Tingkat kehadiran sudah dipertimbangkan
- [ ] Buffer sudah ditambahkan (5-10%)
- [ ] Porsi kecil sudah diinput (atau 0 jika tidak ada)
- [ ] Porsi besar sudah diinput (atau 0 jika tidak ada)
- [ ] Minimal satu jenis porsi > 0
- [ ] Total porsi SD sudah dihitung
- [ ] Total alokasi semua sekolah = total porsi menu
- [ ] Validasi sistem menunjukkan status OK (hijau)
- [ ] Data sudah disimpan
- [ ] Konfirmasi penyimpanan diterima

## Contoh Kasus Lengkap

### Kasus: Menu Nasi Goreng untuk 5 SD

**Data Menu:**
- Resep: Nasi Goreng Spesial
- Tanggal: 15 Januari 2024
- Total Porsi: 1000

**Data Sekolah:**

| Sekolah | Siswa 1-3 | Siswa 4-6 | Total |
|---------|-----------|-----------|-------|
| SD N 1  | 80        | 100       | 180   |
| SD N 2  | 65        | 80        | 145   |
| SD N 3  | 120       | 150       | 270   |
| SD N 4  | 0         | 90        | 90    |
| SD N 5  | 60        | 0         | 60    |

**Perhitungan Alokasi (dengan kehadiran 95% dan buffer 10%):**

| Sekolah | Porsi Kecil | Porsi Besar | Total |
|---------|-------------|-------------|-------|
| SD N 1  | 84          | 105         | 189   |
| SD N 2  | 68          | 88          | 156   |
| SD N 3  | 126         | 158         | 284   |
| SD N 4  | 0           | 95          | 95    |
| SD N 5  | 63          | 0           | 63    |
| **Total** | **341**   | **446**     | **787** |

**Catatan**: Total 787 < 1000, sisa 213 porsi untuk sekolah lain (SMP/SMA).

**Input ke Sistem:**
```
SD Negeri 1
├─ Kecil (Kelas 1-3): [84]
└─ Besar (Kelas 4-6): [105]

SD Negeri 2
├─ Kecil (Kelas 1-3): [68]
└─ Besar (Kelas 4-6): [88]

SD Negeri 3
├─ Kecil (Kelas 1-3): [126]
└─ Besar (Kelas 4-6): [158]

SD Negeri 4
├─ Kecil (Kelas 1-3): [0]
└─ Besar (Kelas 4-6): [95]

SD Negeri 5
├─ Kecil (Kelas 1-3): [63]
└─ Besar (Kelas 4-6): [0]
```

**Hasil di Database:**
- 9 allocation records (SD N 1, 2, 3 masing-masing 2 records; SD N 4 dan 5 masing-masing 1 record)
- Total portions: 787 untuk 5 SD

---

**Versi Dokumen**: 1.0  
**Terakhir Diperbarui**: 2024  
**Untuk**: Ahli Gizi
