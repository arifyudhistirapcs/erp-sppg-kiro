# Screenshots Placeholder - Portion Size Differentiation UI

## Instruksi untuk Tim Dokumentasi

Dokumen ini berisi placeholder untuk screenshot UI yang perlu diambil dan ditambahkan ke dokumentasi pengguna. Setiap screenshot harus menunjukkan fitur diferensiasi ukuran porsi dengan jelas.

## Screenshot yang Diperlukan

### 1. Menu Planning - Form Alokasi SD
**File**: `menu-planning-sd-allocation.png`

**Deskripsi**: Screenshot form alokasi menu item yang menampilkan sekolah SD dengan dua kolom input (porsi kecil dan porsi besar).

**Elemen yang harus terlihat**:
- Nama sekolah dengan kategori "SD"
- Kolom input "Kecil (Kelas 1-3)" dengan jumlah siswa
- Kolom input "Besar (Kelas 4-6)" dengan jumlah siswa
- Total porsi untuk sekolah tersebut
- Indikator validasi (jika ada)

**Contoh konten**:
```
SD Negeri 1 Jakarta
├─ Kecil (Kelas 1-3): [84] (80 siswa)
└─ Besar (Kelas 4-6): [105] (100 siswa)
Total: 189 porsi
```

**Lokasi penyimpanan**: `docs/user-manual/images/menu-planning-sd-allocation.png`

---

### 2. Menu Planning - Form Alokasi SMP/SMA
**File**: `menu-planning-smp-sma-allocation.png`

**Deskripsi**: Screenshot form alokasi menu item yang menampilkan sekolah SMP/SMA dengan satu kolom input (porsi besar saja).

**Elemen yang harus terlihat**:
- Nama sekolah dengan kategori "SMP" atau "SMA"
- Kolom input "Besar" dengan total jumlah siswa
- Total porsi untuk sekolah tersebut
- Tidak ada kolom porsi kecil

**Contoh konten**:
```
SMP Negeri 1 Jakarta
└─ Besar: [262] (250 siswa)
Total: 262 porsi

SMA Negeri 5 Bandung
└─ Besar: [475] (450 siswa)
Total: 475 porsi
```

**Lokasi penyimpanan**: `docs/user-manual/images/menu-planning-smp-sma-allocation.png`

---

### 3. Menu Planning - Validasi Real-time
**File**: `menu-planning-validation.png`

**Deskripsi**: Screenshot yang menampilkan validasi real-time saat pengguna mengisi alokasi.

**Elemen yang harus terlihat**:
- Total porsi menu (target)
- Total alokasi saat ini (running total)
- Indikator status validasi (hijau untuk valid, merah untuk error)
- Pesan error jika ada
- Status tombol "Simpan" (enabled/disabled)

**Contoh konten**:
```
Total Porsi Menu: 1000
Total Alokasi: 987
Status: ❌ Kurang 13 porsi
[Simpan] (disabled)
```

**Lokasi penyimpanan**: `docs/user-manual/images/menu-planning-validation.png`

---

### 4. Menu Planning - Statistik Porsi
**File**: `menu-planning-statistics.png`

**Deskripsi**: Screenshot panel statistik yang menampilkan breakdown porsi kecil vs besar.

**Elemen yang harus terlihat**:
- Total porsi kecil (semua SD)
- Total porsi besar (semua sekolah)
- Persentase distribusi
- Jumlah sekolah per jenis porsi
- Grafik atau visualisasi (jika ada)

**Contoh konten**:
```
Statistik Alokasi Porsi
├─ Total Porsi Kecil: 341 (34%)
├─ Total Porsi Besar: 659 (66%)
├─ Sekolah dengan Porsi Kecil: 5 SD
└─ Sekolah dengan Porsi Besar: 8 (5 SD + 3 SMP/SMA)
```

**Lokasi penyimpanan**: `docs/user-manual/images/menu-planning-statistics.png`

---

### 5. Menu Planning - Form Edit
**File**: `menu-planning-edit-form.png`

**Deskripsi**: Screenshot form edit menu item yang menampilkan alokasi porsi yang sudah ada.

**Elemen yang harus terlihat**:
- Data menu item (resep, tanggal, total porsi)
- Alokasi saat ini untuk setiap sekolah
- Kolom input yang dapat diedit
- Tombol "Simpan Perubahan"

**Contoh konten**:
```
Edit Menu Item: Nasi Goreng - 15 Jan 2024
Total Porsi: 1000

SD Negeri 1
├─ Kecil (Kelas 1-3): [84] → [90]
└─ Besar (Kelas 4-6): [105] → [110]

[Simpan Perubahan]
```

**Lokasi penyimpanan**: `docs/user-manual/images/menu-planning-edit-form.png`

---

### 6. KDS Cooking View - Breakdown Porsi
**File**: `kds-cooking-portion-breakdown.png`

**Deskripsi**: Screenshot KDS Cooking View yang menampilkan breakdown porsi per sekolah.

**Elemen yang harus terlihat**:
- Nama resep
- Daftar sekolah dengan alokasi
- Label porsi kecil dan besar untuk SD
- Label porsi besar untuk SMP/SMA
- Total porsi kecil dan besar

**Contoh konten**:
```
Nasi Goreng Spesial
Total: 1000 porsi (341 kecil + 659 besar)

Alokasi:
├─ SD Negeri 1: Kecil (84), Besar (105)
├─ SD Negeri 2: Kecil (68), Besar (88)
├─ SMP Negeri 1: Besar (262)
└─ SMA Negeri 5: Besar (475)
```

**Lokasi penyimpanan**: `docs/user-manual/images/kds-cooking-portion-breakdown.png`

---

### 7. KDS Packing View - Detail Sekolah
**File**: `kds-packing-school-detail.png`

**Deskripsi**: Screenshot KDS Packing View yang menampilkan detail packing per sekolah dengan porsi size.

**Elemen yang harus terlihat**:
- Nama sekolah
- Breakdown porsi kecil dan besar (untuk SD)
- Porsi besar saja (untuk SMP/SMA)
- Checkbox atau status packing
- Label yang jelas untuk setiap jenis porsi

**Contoh konten**:
```
Packing List - 15 Jan 2024

☐ SD Negeri 1 Jakarta
  ├─ Porsi Kecil (Kelas 1-3): 84 porsi
  └─ Porsi Besar (Kelas 4-6): 105 porsi

☐ SMP Negeri 1 Jakarta
  └─ Porsi Besar: 262 porsi
```

**Lokasi penyimpanan**: `docs/user-manual/images/kds-packing-school-detail.png`

---

### 8. Error Message - Validasi Gagal
**File**: `error-validation-failed.png`

**Deskripsi**: Screenshot pesan error saat validasi gagal.

**Elemen yang harus terlihat**:
- Pesan error yang jelas
- Indikator visual (icon error, warna merah)
- Petunjuk untuk memperbaiki error
- Lokasi error (field mana yang bermasalah)

**Contoh konten**:
```
❌ Error: Jumlah alokasi tidak sesuai dengan total porsi
Total porsi: 1000
Total alokasi: 987
Kekurangan: 13 porsi

Silakan sesuaikan alokasi untuk sekolah-sekolah berikut.
```

**Lokasi penyimpanan**: `docs/user-manual/images/error-validation-failed.png`

---

### 9. Error Message - SMP/SMA Porsi Kecil
**File**: `error-smp-sma-small-portion.png`

**Deskripsi**: Screenshot pesan error saat mencoba mengalokasikan porsi kecil untuk SMP/SMA.

**Elemen yang harus terlihat**:
- Pesan error spesifik
- Nama sekolah yang bermasalah
- Petunjuk perbaikan

**Contoh konten**:
```
❌ Error: SMP/SMA tidak dapat memiliki porsi kecil
Sekolah: SMP Negeri 1 Jakarta

SMP dan SMA hanya dapat menerima porsi besar.
Silakan kosongkan kolom porsi kecil atau isi dengan 0.
```

**Lokasi penyimpanan**: `docs/user-manual/images/error-smp-sma-small-portion.png`

---

### 10. Success Message - Alokasi Tersimpan
**File**: `success-allocation-saved.png`

**Deskripsi**: Screenshot pesan sukses setelah alokasi berhasil disimpan.

**Elemen yang harus terlihat**:
- Pesan sukses yang jelas
- Indikator visual (icon success, warna hijau)
- Ringkasan data yang disimpan
- Opsi untuk melihat atau edit

**Contoh konten**:
```
✓ Berhasil menyimpan alokasi menu item
Menu: Nasi Goreng Spesial
Tanggal: 15 Januari 2024
Total Porsi: 1000
Sekolah: 8 sekolah

[Lihat Detail] [Edit]
```

**Lokasi penyimpanan**: `docs/user-manual/images/success-allocation-saved.png`

---

## Panduan Pengambilan Screenshot

### Persiapan
1. Gunakan data dummy yang realistis
2. Pastikan UI dalam kondisi bersih (tidak ada error console)
3. Gunakan resolusi standar (1920x1080 atau 1366x768)
4. Pastikan semua elemen UI terlihat jelas

### Teknis
1. **Format**: PNG (untuk kualitas terbaik)
2. **Resolusi**: Minimal 1366x768
3. **Crop**: Fokus pada area yang relevan, tidak perlu full screen
4. **Annotasi**: Tambahkan arrow atau highlight jika perlu untuk menekankan elemen penting

### Konten
1. Gunakan nama sekolah yang realistis (SD/SMP/SMA Negeri)
2. Gunakan angka yang masuk akal (sesuai dengan jumlah siswa tipikal)
3. Pastikan semua label dalam Bahasa Indonesia
4. Tampilkan berbagai skenario (valid, error, success)

### Organisasi File
```
docs/user-manual/images/
├── menu-planning-sd-allocation.png
├── menu-planning-smp-sma-allocation.png
├── menu-planning-validation.png
├── menu-planning-statistics.png
├── menu-planning-edit-form.png
├── kds-cooking-portion-breakdown.png
├── kds-packing-school-detail.png
├── error-validation-failed.png
├── error-smp-sma-small-portion.png
└── success-allocation-saved.png
```

## Integrasi dengan Dokumentasi

Setelah screenshot diambil, update dokumen berikut dengan menambahkan gambar:

### 1. PANDUAN_ALOKASI_PORSI.md
Tambahkan screenshot di bagian:
- "Cara Mengalokasikan Porsi" → screenshot 1, 2
- "Validasi Sistem" → screenshot 3, 8, 9
- "Mengedit Alokasi Porsi" → screenshot 5
- "Melihat Statistik Porsi" → screenshot 4

### 2. PROSES_ALOKASI_SD.md
Tambahkan screenshot di bagian:
- "Langkah 1: Identifikasi Sekolah SD" → screenshot 1
- "Langkah 4: Verifikasi Total" → screenshot 3
- "Validasi dan Error Handling" → screenshot 8

### 3. PROSES_ALOKASI_SMP_SMA.md
Tambahkan screenshot di bagian:
- "Langkah 1: Identifikasi Sekolah SMP/SMA" → screenshot 2
- "Validasi dan Error Handling" → screenshot 9

### 4. FAQ (akan dibuat di task berikutnya)
Tambahkan screenshot untuk ilustrasi jawaban FAQ

## Syntax Markdown untuk Menambahkan Gambar

```markdown
![Deskripsi gambar](images/nama-file.png)

Atau dengan caption:

<figure>
  <img src="images/nama-file.png" alt="Deskripsi gambar">
  <figcaption>Caption gambar</figcaption>
</figure>
```

## Checklist Pengambilan Screenshot

- [ ] Buat folder `docs/user-manual/images/`
- [ ] Siapkan environment dengan data dummy
- [ ] Ambil screenshot 1: Menu Planning - Form Alokasi SD
- [ ] Ambil screenshot 2: Menu Planning - Form Alokasi SMP/SMA
- [ ] Ambil screenshot 3: Menu Planning - Validasi Real-time
- [ ] Ambil screenshot 4: Menu Planning - Statistik Porsi
- [ ] Ambil screenshot 5: Menu Planning - Form Edit
- [ ] Ambil screenshot 6: KDS Cooking View - Breakdown Porsi
- [ ] Ambil screenshot 7: KDS Packing View - Detail Sekolah
- [ ] Ambil screenshot 8: Error Message - Validasi Gagal
- [ ] Ambil screenshot 9: Error Message - SMP/SMA Porsi Kecil
- [ ] Ambil screenshot 10: Success Message - Alokasi Tersimpan
- [ ] Crop dan optimize semua gambar
- [ ] Simpan dengan nama file yang sesuai
- [ ] Update PANDUAN_ALOKASI_PORSI.md dengan gambar
- [ ] Update PROSES_ALOKASI_SD.md dengan gambar
- [ ] Update PROSES_ALOKASI_SMP_SMA.md dengan gambar
- [ ] Verifikasi semua gambar terlihat dengan baik
- [ ] Commit dan push ke repository

## Catatan

- Screenshot ini adalah placeholder dan harus diganti dengan screenshot aktual dari aplikasi yang sudah berjalan
- Koordinasi dengan tim frontend untuk memastikan UI sudah final sebelum mengambil screenshot
- Jika ada perubahan UI di masa depan, screenshot harus diupdate
- Pertimbangkan untuk membuat video tutorial sebagai pelengkap screenshot

---

**Status**: Placeholder - Menunggu pengambilan screenshot aktual  
**Terakhir Diperbarui**: 2024  
**PIC**: Tim Dokumentasi
