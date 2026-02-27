# FAQ - Alokasi Porsi Berdasarkan Ukuran

## Pertanyaan Umum tentang Fitur Diferensiasi Ukuran Porsi

### Konsep Dasar

#### Q1: Apa itu diferensiasi ukuran porsi?
**A**: Diferensiasi ukuran porsi adalah fitur yang memungkinkan sistem membedakan antara porsi kecil (untuk siswa SD kelas 1-3) dan porsi besar (untuk siswa SD kelas 4-6, SMP, dan SMA). Fitur ini membantu memastikan setiap siswa mendapat porsi yang sesuai dengan kebutuhan nutrisi mereka berdasarkan usia.

#### Q2: Mengapa perlu membedakan ukuran porsi?
**A**: Kebutuhan nutrisi anak usia 6-9 tahun (SD kelas 1-3) berbeda dengan anak usia 10 tahun ke atas. Dengan membedakan ukuran porsi, kita dapat:
- Mengurangi food waste (porsi terlalu besar untuk anak kecil)
- Memastikan kecukupan nutrisi (porsi tidak terlalu kecil untuk anak besar)
- Mengoptimalkan biaya operasional
- Meningkatkan kepuasan siswa dan sekolah

#### Q3: Apakah semua sekolah menggunakan dua jenis porsi?
**A**: Tidak. Hanya sekolah SD yang menggunakan dua jenis porsi (kecil dan besar). Sekolah SMP dan SMA hanya menggunakan satu jenis porsi (besar).

---

### Alokasi untuk Sekolah SD

#### Q4: Apakah semua SD harus memiliki porsi kecil dan besar?
**A**: Tidak wajib. SD dapat memiliki:
- Hanya porsi kecil (jika hanya ada kelas 1-3)
- Hanya porsi besar (jika hanya ada kelas 4-6)
- Kedua jenis porsi (jika ada kelas 1-6)

Yang penting adalah minimal satu jenis porsi harus lebih dari 0.

#### Q5: Bagaimana jika SD hanya memiliki kelas 1-3?
**A**: Isi hanya kolom porsi kecil dengan jumlah yang sesuai, dan kosongkan kolom porsi besar (atau isi dengan 0). Sistem akan menyimpan hanya satu record dengan portion_size='small'.

#### Q6: Bagaimana jika SD hanya memiliki kelas 4-6?
**A**: Isi hanya kolom porsi besar dengan jumlah yang sesuai, dan kosongkan kolom porsi kecil (atau isi dengan 0). Sistem akan menyimpan hanya satu record dengan portion_size='large'.

#### Q7: Apakah distribusi porsi kecil dan besar harus seimbang?
**A**: Tidak harus. Distribusi disesuaikan dengan jumlah siswa aktual di setiap tingkat kelas. Jika SD memiliki lebih banyak siswa kelas 1-3, maka porsi kecil akan lebih banyak, dan sebaliknya.

#### Q8: Bagaimana cara menghitung jumlah porsi kecil dan besar untuk SD?
**A**: Gunakan rumus sederhana:
```
Porsi Kecil = Siswa Kelas 1-3 × Tingkat Kehadiran + Buffer
Porsi Besar = Siswa Kelas 4-6 × Tingkat Kehadiran + Buffer
```
Contoh: 80 siswa kelas 1-3 × 95% + 10% buffer = 84 porsi kecil

---

### Alokasi untuk Sekolah SMP/SMA

#### Q9: Mengapa SMP dan SMA tidak memiliki porsi kecil?
**A**: Siswa SMP dan SMA (usia 13-18 tahun) memiliki kebutuhan nutrisi yang lebih tinggi dan konsisten. Mereka semua memerlukan porsi besar, sehingga tidak perlu diferensiasi ukuran.

#### Q10: Apakah bisa mengalokasikan porsi kecil untuk SMP/SMA?
**A**: Tidak. Sistem akan menolak alokasi porsi kecil untuk SMP/SMA dengan pesan error "SMP/SMA schools cannot have small portions". Ini adalah validasi yang disengaja untuk menjaga konsistensi data.

#### Q11: Bagaimana jika SMP/SMA meminta porsi lebih besar?
**A**: Jika ada permintaan khusus:
1. Koordinasi dengan tim nutrisi untuk evaluasi
2. Pertimbangkan menambah buffer (10-15%)
3. Jika perlu, sesuaikan komposisi menu
4. Catat permintaan dan alasan dalam dokumentasi

Ukuran porsi standar tidak diubah, tetapi jumlah alokasi dapat disesuaikan.

#### Q12: Apakah porsi besar untuk SMP sama dengan porsi besar untuk SMA?
**A**: Ya, ukuran porsi besar adalah sama untuk SD kelas 4-6, SMP, dan SMA. Perbedaan kebutuhan kalori disesuaikan melalui komposisi menu (jenis dan jumlah bahan), bukan ukuran porsi.

---

### Validasi dan Error

#### Q13: Mengapa tombol "Simpan" tidak aktif?
**A**: Tombol "Simpan" hanya aktif jika semua validasi berhasil:
- Total alokasi = total porsi menu
- Semua nilai porsi ≥ 0
- Setiap sekolah memiliki minimal 1 porsi
- Tidak ada porsi kecil untuk SMP/SMA

Periksa pesan error di layar untuk mengetahui masalahnya.

#### Q14: Apa arti error "Jumlah alokasi tidak sesuai dengan total porsi"?
**A**: Error ini muncul ketika jumlah total semua porsi yang dialokasikan tidak sama dengan total porsi menu yang dibuat. 

Contoh:
- Total porsi menu: 1000
- Total alokasi: 987
- Error: Kurang 13 porsi

Solusi: Sesuaikan alokasi untuk beberapa sekolah hingga total = 1000.

#### Q15: Apa arti error "Sekolah harus memiliki minimal satu porsi"?
**A**: Error ini muncul ketika semua kolom porsi untuk satu sekolah bernilai 0. Setiap sekolah yang dialokasikan harus menerima minimal 1 porsi (kecil atau besar).

Solusi: Isi minimal satu kolom dengan nilai > 0, atau hapus sekolah dari alokasi jika memang tidak perlu.

#### Q16: Mengapa sistem menolak nilai negatif?
**A**: Nilai negatif tidak masuk akal dalam konteks alokasi porsi. Sistem hanya menerima angka 0 atau positif. Jika tidak sengaja memasukkan nilai negatif, ganti dengan 0 atau angka positif.

---

### Proses dan Workflow

#### Q17: Bagaimana cara mengedit alokasi yang sudah disimpan?
**A**: 
1. Buka menu item yang ingin diedit
2. Klik tombol "Edit"
3. Sistem menampilkan alokasi saat ini
4. Ubah nilai porsi sesuai kebutuhan
5. Validasi otomatis akan berjalan
6. Klik "Simpan" untuk menyimpan perubahan

Perubahan akan menghapus data lama dan membuat data baru.

#### Q18: Apakah perubahan alokasi langsung terlihat di KDS?
**A**: Ya, perubahan alokasi akan langsung terlihat di KDS Cooking View dan KDS Packing View setelah disimpan. Sistem menggunakan real-time sync melalui Firebase.

#### Q19: Bagaimana cara melihat history alokasi?
**A**: Saat ini sistem menyimpan alokasi terbaru saja. Untuk melihat history atau audit trail, hubungi admin untuk export data dari database. Fitur history mungkin ditambahkan di versi mendatang.

#### Q20: Apakah bisa mengalokasikan porsi untuk beberapa tanggal sekaligus?
**A**: Saat ini alokasi dilakukan per menu item (per tanggal). Untuk mengalokasikan beberapa tanggal, Anda perlu membuat menu item untuk setiap tanggal. Fitur bulk allocation mungkin ditambahkan di masa depan.

---

### Data dan Referensi

#### Q21: Dari mana data jumlah siswa per tingkat kelas berasal?
**A**: Data jumlah siswa diambil dari tabel `schools` di database, field:
- `student_count_grade_1_3`: Jumlah siswa SD kelas 1-3
- `student_count_grade_4_6`: Jumlah siswa SD kelas 4-6
- `student_count`: Total siswa untuk SMP/SMA

Data ini harus diupdate secara berkala oleh admin.

#### Q22: Bagaimana jika data jumlah siswa tidak akurat?
**A**: 
1. Hubungi admin sekolah untuk konfirmasi data terbaru
2. Gunakan data manual sementara untuk alokasi
3. Catat perbedaan dalam dokumentasi
4. Minta admin sistem untuk update database
5. Verifikasi dengan data dapodik jika perlu

#### Q23: Apakah sistem otomatis mengalokasikan porsi berdasarkan jumlah siswa?
**A**: Tidak. Sistem hanya menampilkan jumlah siswa sebagai referensi. Ahli gizi harus menentukan alokasi secara manual dengan mempertimbangkan:
- Jumlah siswa
- Tingkat kehadiran
- Buffer untuk cadangan
- Feedback sebelumnya
- Kondisi khusus (acara, cuaca, dll)

---

### KDS dan Operasional

#### Q24: Bagaimana dapur mengetahui perbedaan porsi kecil dan besar?
**A**: KDS Cooking View menampilkan breakdown porsi per sekolah dengan label jelas:
- "SD Negeri 1: Kecil (84), Besar (105)"
- Total porsi kecil dan besar ditampilkan terpisah

Dapur harus menyiapkan porsi dengan ukuran berbeda sesuai label.

#### Q25: Bagaimana tim packing membedakan porsi kecil dan besar?
**A**: 
1. KDS Packing View menampilkan detail per sekolah dengan jenis porsi
2. Gunakan kemasan atau label berbeda untuk setiap ukuran
3. Rekomendasi: 
   - Label warna berbeda (misal: biru untuk kecil, merah untuk besar)
   - Stiker "Kelas 1-3" dan "Kelas 4-6"
   - Kemasan ukuran berbeda jika memungkinkan

#### Q26: Bagaimana sekolah mengetahui porsi mana untuk kelas mana?
**A**: 
1. Setiap paket harus diberi label yang jelas
2. Koordinasi dengan kepala sekolah tentang sistem labeling
3. Sertakan instruksi distribusi jika perlu
4. Lakukan sosialisasi ke sekolah tentang sistem baru

#### Q27: Apakah driver perlu tahu tentang perbedaan porsi?
**A**: Driver perlu tahu bahwa ada dua jenis porsi untuk SD, tetapi tidak perlu detail teknis. Yang penting:
- Manifest menunjukkan total paket per sekolah
- Sekolah menandatangani penerimaan untuk semua paket
- Jika ada komplain, catat dan laporkan

---

### Statistik dan Reporting

#### Q28: Apa itu statistik porsi dan bagaimana cara membacanya?
**A**: Statistik porsi menampilkan:
- **Total Porsi Kecil**: Jumlah semua porsi kecil di semua SD
- **Total Porsi Besar**: Jumlah semua porsi besar di semua sekolah
- **Persentase**: Distribusi porsi kecil vs besar (misal: 34% kecil, 66% besar)
- **Jumlah Sekolah**: Berapa sekolah menerima setiap jenis porsi

Gunakan statistik ini untuk verifikasi distribusi masuk akal.

#### Q29: Berapa persentase ideal untuk porsi kecil vs besar?
**A**: Tidak ada persentase ideal yang tetap. Persentase tergantung pada:
- Jumlah SD vs SMP/SMA yang dilayani
- Distribusi siswa kelas 1-3 vs 4-6 di SD
- Kebijakan alokasi regional

Yang penting adalah persentase konsisten dengan data siswa aktual.

#### Q30: Bagaimana cara export data alokasi untuk reporting?
**A**: Saat ini export dilakukan oleh admin melalui database query. Untuk kebutuhan reporting:
1. Hubungi admin IT
2. Tentukan periode dan format yang diinginkan
3. Admin akan export data dalam format Excel/CSV
4. Fitur export self-service mungkin ditambahkan di masa depan

---

### Troubleshooting

#### Q31: Data tidak tersimpan setelah klik "Simpan"
**Kemungkinan penyebab dan solusi**:
1. **Koneksi internet terputus**: Periksa koneksi dan coba lagi
2. **Validasi gagal**: Periksa pesan error di layar
3. **Session timeout**: Refresh halaman dan login ulang
4. **Error server**: Periksa console browser, hubungi admin IT

#### Q32: Alokasi tersimpan tetapi tidak muncul di KDS
**Kemungkinan penyebab dan solusi**:
1. **Cache browser**: Refresh halaman KDS (Ctrl+F5)
2. **Firebase sync delay**: Tunggu beberapa detik dan refresh
3. **Tanggal tidak sesuai**: Pastikan tanggal menu sesuai dengan filter KDS
4. **Error sync**: Periksa console browser, hubungi admin IT

#### Q33: Jumlah porsi di KDS tidak sesuai dengan yang dialokasikan
**Kemungkinan penyebab dan solusi**:
1. **Data belum tersync**: Refresh halaman
2. **Edit setelah sync**: Pastikan edit sudah disimpan
3. **Bug sistem**: Catat detail masalah dan hubungi admin IT
4. **Multiple edit**: Pastikan tidak ada user lain yang edit bersamaan

#### Q34: Sistem lambat saat mengalokasikan banyak sekolah
**Solusi**:
1. Alokasikan sekolah secara bertahap (misal: 10 sekolah per batch)
2. Simpan secara berkala
3. Gunakan koneksi internet yang stabil
4. Tutup tab browser lain yang tidak perlu
5. Jika masalah berlanjut, hubungi admin IT untuk optimasi

---

### Best Practices

#### Q35: Berapa buffer yang ideal untuk alokasi?
**A**: Rekomendasi buffer:
- **5%**: Untuk sekolah dengan kehadiran sangat stabil (>98%)
- **7-8%**: Untuk kondisi normal (95-98% kehadiran)
- **10%**: Untuk sekolah dengan kehadiran fluktuatif (90-95%)
- **12-15%**: Untuk hari khusus (ujian, acara, cuaca buruk)

Sesuaikan berdasarkan pengalaman dan feedback.

#### Q36: Kapan waktu terbaik untuk melakukan alokasi?
**A**: Rekomendasi:
- **H-2 atau H-3**: Untuk perencanaan normal
- **H-1**: Untuk adjustment berdasarkan konfirmasi sekolah
- **H-0 pagi**: Hanya untuk emergency adjustment

Hindari alokasi terlalu jauh dari hari H karena data kehadiran bisa berubah.

#### Q37: Bagaimana cara menangani feedback dari sekolah?
**A**: 
1. **Porsi terlalu banyak**: Kurangi buffer untuk alokasi berikutnya
2. **Porsi terlalu sedikit**: Tambah buffer atau verifikasi data siswa
3. **Ukuran tidak sesuai**: Koordinasi dengan dapur untuk standarisasi
4. **Distribusi salah**: Review proses labeling dan packing

Catat semua feedback dalam dokumentasi untuk perbaikan berkelanjutan.

#### Q38: Bagaimana cara melatih user baru?
**A**: 
1. Berikan akses ke dokumentasi (PANDUAN_ALOKASI_PORSI.md)
2. Lakukan demo dengan data dummy
3. Biarkan user mencoba dengan supervisi
4. Berikan checklist untuk diikuti
5. Review hasil alokasi pertama sebelum produksi
6. Berikan feedback konstruktif

#### Q39: Bagaimana cara memastikan konsistensi alokasi antar ahli gizi?
**A**: 
1. Gunakan template perhitungan yang sama
2. Dokumentasikan keputusan dan alasan
3. Review alokasi secara berkala dalam tim
4. Buat standar buffer untuk setiap kondisi
5. Gunakan data historis sebagai referensi
6. Lakukan kalibrasi antar ahli gizi

#### Q40: Apa yang harus dilakukan jika terjadi perubahan mendadak?
**A**: 
**Skenario 1: Sekolah tiba-tiba tidak bisa menerima**
1. Edit alokasi dan redistribute porsi ke sekolah lain
2. Informasikan ke dapur dan packing
3. Update manifest driver

**Skenario 2: Sekolah minta tambahan porsi**
1. Cek ketersediaan porsi (buffer atau dari sekolah lain)
2. Edit alokasi jika memungkinkan
3. Koordinasi dengan dapur untuk produksi tambahan jika perlu
4. Catat untuk evaluasi buffer di masa depan

**Skenario 3: Data siswa berubah signifikan**
1. Verifikasi data baru dengan sekolah
2. Update database melalui admin
3. Sesuaikan alokasi untuk menu berikutnya
4. Informasikan perubahan ke tim terkait

---

### Teknis dan Sistem

#### Q41: Bagaimana sistem menyimpan data alokasi di database?
**A**: Untuk setiap alokasi, sistem membuat record di tabel `menu_item_school_allocations`:
- **SD dengan kedua porsi**: 2 records (satu small, satu large)
- **SD dengan satu porsi**: 1 record (small atau large)
- **SMP/SMA**: 1 record (large saja)

Setiap record memiliki field `portion_size` yang bernilai 'small' atau 'large'.

#### Q42: Apakah ada API untuk alokasi porsi?
**A**: Ya, API endpoint yang relevan:
- `POST /api/menu-items`: Create menu item dengan alokasi
- `GET /api/menu-items/:id`: Get menu item dengan breakdown porsi
- `PUT /api/menu-items/:id`: Update alokasi
- `GET /api/kds/cooking`: Get data untuk KDS Cooking View
- `GET /api/kds/packing`: Get data untuk KDS Packing View

Dokumentasi lengkap ada di `backend/docs/API_PORTION_SIZE_DIFFERENTIATION.md`.

#### Q43: Bagaimana cara rollback jika ada masalah setelah deployment?
**A**: Prosedur rollback:
1. Admin menjalankan rollback migration script
2. Field `portion_size` akan dihapus dari database
3. Aplikasi akan kembali ke versi sebelumnya
4. Data alokasi akan kembali ke format lama (tanpa diferensiasi)

Detail prosedur ada di `backend/migrations/ROLLBACK_PROCEDURE.md`.

#### Q44: Apakah fitur ini kompatibel dengan data lama?
**A**: Ya, sistem dirancang backward compatible:
- Data alokasi lama otomatis dimigrate dengan `portion_size='large'`
- Sistem dapat membaca data lama dan baru
- Tidak ada data yang hilang selama migrasi

#### Q45: Bagaimana cara backup data sebelum deployment?
**A**: Prosedur backup:
1. Admin menjalankan backup script
2. Database di-export ke file SQL
3. File disimpan di lokasi aman dengan timestamp
4. Verifikasi backup dapat di-restore
5. Baru lakukan deployment

Detail prosedur ada di deployment documentation.

---

### Kontak dan Dukungan

#### Q46: Siapa yang harus dihubungi jika ada masalah?
**A**: 
- **Masalah alokasi/workflow**: Kepala Ahli Gizi atau Supervisor
- **Masalah teknis/sistem**: Admin IT atau Tim Support
- **Masalah data sekolah**: Admin Sekolah atau Data Entry
- **Masalah operasional**: Kepala SPPG atau Koordinator

#### Q47: Bagaimana cara melaporkan bug atau request fitur?
**A**: 
1. Catat detail masalah atau request:
   - Apa yang terjadi
   - Apa yang diharapkan
   - Langkah untuk reproduce (jika bug)
   - Screenshot jika perlu
2. Hubungi admin IT melalui:
   - Email: support@sppg.id
   - WhatsApp: +62 XXX-XXXX-XXXX
   - Ticketing system (jika ada)
3. Admin akan follow up dan memberikan update

#### Q48: Apakah ada training untuk fitur baru ini?
**A**: Ya, training tersedia:
- **Training online**: Video tutorial dan dokumentasi
- **Training offline**: Workshop untuk ahli gizi
- **On-the-job training**: Pendampingan saat implementasi awal
- **Refresher training**: Berkala untuk update fitur

Hubungi koordinator training untuk jadwal.

#### Q49: Di mana bisa mendapatkan dokumentasi lengkap?
**A**: Dokumentasi tersedia di:
- **User Manual**: `docs/user-manual/PANDUAN_ALOKASI_PORSI.md`
- **Proses SD**: `docs/user-manual/PROSES_ALOKASI_SD.md`
- **Proses SMP/SMA**: `docs/user-manual/PROSES_ALOKASI_SMP_SMA.md`
- **FAQ**: `docs/user-manual/FAQ_ALOKASI_PORSI.md` (dokumen ini)
- **API Documentation**: `backend/docs/API_PORTION_SIZE_DIFFERENTIATION.md`

#### Q50: Apakah dokumentasi akan diupdate?
**A**: Ya, dokumentasi akan diupdate secara berkala:
- Saat ada perubahan fitur
- Saat ada feedback dari user
- Saat ada FAQ baru yang sering ditanyakan
- Saat ada best practice baru

Periksa tanggal "Terakhir Diperbarui" di setiap dokumen.

---

## Pertanyaan Tambahan?

Jika pertanyaan Anda tidak terjawab dalam FAQ ini, silakan:

1. **Periksa dokumentasi lain**:
   - PANDUAN_ALOKASI_PORSI.md
   - PROSES_ALOKASI_SD.md
   - PROSES_ALOKASI_SMP_SMA.md

2. **Hubungi support**:
   - Email: support@sppg.id
   - Telepon: (021) XXX-XXXX
   - WhatsApp: +62 XXX-XXXX-XXXX

3. **Submit pertanyaan untuk FAQ**:
   - Kirim pertanyaan Anda ke tim dokumentasi
   - Pertanyaan yang sering ditanyakan akan ditambahkan ke FAQ

---

**Versi Dokumen**: 1.0  
**Terakhir Diperbarui**: 2024  
**Total Pertanyaan**: 50  
**Untuk**: Semua pengguna sistem (Ahli Gizi, Admin, Operator KDS)
