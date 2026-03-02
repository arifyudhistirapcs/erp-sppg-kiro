# Requirements Document: Stok Opname

## Introduction

Stok Opname (Physical Inventory Count) adalah fitur untuk melakukan penghitungan fisik stok di gudang/SPPG dan menyesuaikan catatan stok di sistem agar sesuai dengan kondisi aktual. Fitur ini memungkinkan pengguna untuk mencatat hasil penghitungan fisik untuk beberapa item inventory sekaligus, kemudian meminta persetujuan dari Kepala SPPG sebelum penyesuaian stok diterapkan ke sistem.

## Glossary

- **Stok_Opname_System**: Sistem yang mengelola proses penghitungan fisik stok dan penyesuaian stok
- **Stok_Opname_Form**: Formulir yang berisi daftar item inventory yang akan dihitung fisiknya
- **Inventory_Item**: Item bahan baku yang tersimpan di sistem inventory
- **Physical_Count**: Jumlah stok aktual hasil penghitungan fisik di gudang
- **System_Stock**: Jumlah stok yang tercatat di sistem
- **Stock_Adjustment**: Perubahan jumlah stok untuk menyesuaikan system stock dengan physical count
- **Kepala_SPPG**: Peran pengguna yang memiliki wewenang untuk menyetujui stok opname
- **Approval_Status**: Status persetujuan stok opname (pending, approved, rejected)
- **Inventory_Movement**: Catatan pergerakan stok masuk atau keluar
- **SPPG**: Satuan Pelayanan Pemenuhan Gizi (unit layanan katering makanan sekolah)

## Requirements

### Requirement 1: Stok Opname Tab Navigation

**User Story:** Sebagai pengguna sistem inventory, saya ingin mengakses fitur Stok Opname melalui tab khusus, sehingga saya dapat melakukan penghitungan fisik stok dengan mudah.

#### Acceptance Criteria

1. THE Stok_Opname_System SHALL menyediakan tab "Stok Opname" di modul Inventory
2. THE Stok_Opname_System SHALL menampilkan tab "Stok Opname" bersama dengan tab "Daftar Inventory", "Alert Stok Menipis", dan "Riwayat Pergerakan"
3. WHEN pengguna mengklik tab "Stok Opname", THE Stok_Opname_System SHALL menampilkan halaman daftar stok opname

### Requirement 2: Create Stok Opname Form

**User Story:** Sebagai staff gudang, saya ingin membuat formulir stok opname baru, sehingga saya dapat mencatat hasil penghitungan fisik untuk beberapa item sekaligus.

#### Acceptance Criteria

1. THE Stok_Opname_System SHALL menyediakan tombol untuk membuat Stok_Opname_Form baru
2. WHEN pengguna membuat Stok_Opname_Form baru, THE Stok_Opname_System SHALL menyimpan tanggal pembuatan formulir
3. WHEN pengguna membuat Stok_Opname_Form baru, THE Stok_Opname_System SHALL menyimpan identitas pengguna yang membuat formulir
4. THE Stok_Opname_System SHALL mengatur Approval_Status formulir baru menjadi "pending"
5. THE Stok_Opname_System SHALL memungkinkan pengguna menambahkan catatan atau keterangan pada Stok_Opname_Form

### Requirement 3: Add Multiple Items to Stok Opname

**User Story:** Sebagai staff gudang, saya ingin menambahkan beberapa item inventory ke dalam satu formulir stok opname, sehingga saya dapat melakukan penghitungan fisik secara batch.

#### Acceptance Criteria

1. THE Stok_Opname_System SHALL memungkinkan pengguna menambahkan satu atau lebih Inventory_Item ke dalam Stok_Opname_Form
2. WHEN pengguna menambahkan Inventory_Item, THE Stok_Opname_System SHALL menampilkan System_Stock saat ini untuk item tersebut
3. THE Stok_Opname_System SHALL memungkinkan pengguna memasukkan Physical_Count untuk setiap Inventory_Item
4. WHEN Physical_Count dimasukkan, THE Stok_Opname_System SHALL menghitung selisih antara Physical_Count dan System_Stock
5. THE Stok_Opname_System SHALL menampilkan selisih stok (positif atau negatif) untuk setiap item
6. THE Stok_Opname_System SHALL memungkinkan pengguna menambahkan catatan untuk setiap item dalam formulir

### Requirement 4: Edit Stok Opname Before Approval

**User Story:** Sebagai staff gudang, saya ingin mengubah data stok opname sebelum diajukan untuk persetujuan, sehingga saya dapat memperbaiki kesalahan input.

#### Acceptance Criteria

1. WHILE Approval_Status adalah "pending", THE Stok_Opname_System SHALL memungkinkan pengguna mengubah Physical_Count
2. WHILE Approval_Status adalah "pending", THE Stok_Opname_System SHALL memungkinkan pengguna menambah atau menghapus Inventory_Item dari formulir
3. WHILE Approval_Status adalah "pending", THE Stok_Opname_System SHALL memungkinkan pengguna mengubah catatan formulir
4. WHEN Approval_Status adalah "approved" atau "rejected", THE Stok_Opname_System SHALL mencegah pengeditan formulir

### Requirement 5: Submit Stok Opname for Approval

**User Story:** Sebagai staff gudang, saya ingin mengajukan stok opname untuk persetujuan Kepala SPPG, sehingga penyesuaian stok dapat diproses.

#### Acceptance Criteria

1. WHILE Approval_Status adalah "pending", THE Stok_Opname_System SHALL menyediakan tombol untuk mengajukan persetujuan
2. WHEN pengguna mengajukan Stok_Opname_Form untuk persetujuan, THE Stok_Opname_System SHALL memvalidasi bahwa minimal satu Inventory_Item telah ditambahkan
3. WHEN pengguna mengajukan Stok_Opname_Form untuk persetujuan, THE Stok_Opname_System SHALL memvalidasi bahwa semua item memiliki Physical_Count yang valid
4. IF Stok_Opname_Form tidak memiliki item atau Physical_Count tidak valid, THEN THE Stok_Opname_System SHALL menampilkan pesan error dan mencegah pengajuan
5. WHEN pengajuan berhasil, THE Stok_Opname_System SHALL mengirim notifikasi ke Kepala_SPPG

### Requirement 6: Approve or Reject Stok Opname

**User Story:** Sebagai Kepala SPPG, saya ingin menyetujui atau menolak stok opname yang diajukan, sehingga saya dapat mengontrol penyesuaian stok di sistem.

#### Acceptance Criteria

1. WHERE pengguna memiliki peran Kepala_SPPG, THE Stok_Opname_System SHALL menampilkan tombol "Approve" dan "Reject" untuk formulir dengan status "pending"
2. WHEN Kepala_SPPG menyetujui Stok_Opname_Form, THE Stok_Opname_System SHALL mengubah Approval_Status menjadi "approved"
3. WHEN Kepala_SPPG menolak Stok_Opname_Form, THE Stok_Opname_System SHALL mengubah Approval_Status menjadi "rejected"
4. WHEN Kepala_SPPG menolak formulir, THE Stok_Opname_System SHALL memungkinkan Kepala_SPPG memasukkan alasan penolakan
5. THE Stok_Opname_System SHALL menyimpan identitas Kepala_SPPG yang melakukan approval atau rejection
6. THE Stok_Opname_System SHALL menyimpan timestamp saat approval atau rejection dilakukan

### Requirement 7: Apply Stock Adjustments After Approval

**User Story:** Sebagai sistem, saya ingin menerapkan penyesuaian stok secara otomatis setelah stok opname disetujui, sehingga System_Stock sesuai dengan Physical_Count.

#### Acceptance Criteria

1. WHEN Stok_Opname_Form disetujui, THE Stok_Opname_System SHALL membuat Stock_Adjustment untuk setiap Inventory_Item dalam formulir
2. WHEN Stock_Adjustment dibuat, THE Stok_Opname_System SHALL mengupdate System_Stock menjadi sama dengan Physical_Count
3. WHEN Stock_Adjustment dibuat, THE Stok_Opname_System SHALL mencatat Inventory_Movement dengan tipe "adjustment"
4. THE Stok_Opname_System SHALL menyimpan referensi ke Stok_Opname_Form dalam Inventory_Movement
5. WHEN selisih stok positif, THE Stok_Opname_System SHALL mencatat Inventory_Movement sebagai "stock in"
6. WHEN selisih stok negatif, THE Stok_Opname_System SHALL mencatat Inventory_Movement sebagai "stock out"
7. THE Stok_Opname_System SHALL menyimpan jumlah selisih (absolute value) dalam Inventory_Movement

### Requirement 8: View Stok Opname History

**User Story:** Sebagai pengguna sistem, saya ingin melihat daftar semua stok opname yang pernah dibuat, sehingga saya dapat melacak riwayat penghitungan fisik stok.

#### Acceptance Criteria

1. THE Stok_Opname_System SHALL menampilkan daftar semua Stok_Opname_Form yang pernah dibuat
2. THE Stok_Opname_System SHALL menampilkan tanggal pembuatan untuk setiap formulir
3. THE Stok_Opname_System SHALL menampilkan Approval_Status untuk setiap formulir
4. THE Stok_Opname_System SHALL menampilkan nama pembuat formulir
5. WHERE formulir telah disetujui atau ditolak, THE Stok_Opname_System SHALL menampilkan nama Kepala_SPPG yang melakukan approval
6. THE Stok_Opname_System SHALL mengurutkan daftar berdasarkan tanggal pembuatan (terbaru di atas)

### Requirement 9: View Stok Opname Details

**User Story:** Sebagai pengguna sistem, saya ingin melihat detail lengkap dari stok opname, sehingga saya dapat mengetahui item-item yang dihitung dan selisihnya.

#### Acceptance Criteria

1. WHEN pengguna memilih Stok_Opname_Form dari daftar, THE Stok_Opname_System SHALL menampilkan detail formulir
2. THE Stok_Opname_System SHALL menampilkan semua Inventory_Item yang tercantum dalam formulir
3. THE Stok_Opname_System SHALL menampilkan System_Stock, Physical_Count, dan selisih untuk setiap item
4. THE Stok_Opname_System SHALL menampilkan catatan formulir jika ada
5. THE Stok_Opname_System SHALL menampilkan catatan per item jika ada
6. WHERE formulir ditolak, THE Stok_Opname_System SHALL menampilkan alasan penolakan

### Requirement 10: Delete Pending Stok Opname

**User Story:** Sebagai staff gudang, saya ingin menghapus stok opname yang masih pending, sehingga saya dapat membatalkan penghitungan yang tidak jadi digunakan.

#### Acceptance Criteria

1. WHILE Approval_Status adalah "pending", THE Stok_Opname_System SHALL menyediakan tombol untuk menghapus Stok_Opname_Form
2. WHEN pengguna menghapus formulir pending, THE Stok_Opname_System SHALL menampilkan konfirmasi penghapusan
3. WHEN penghapusan dikonfirmasi, THE Stok_Opname_System SHALL menghapus Stok_Opname_Form dan semua item di dalamnya
4. WHEN Approval_Status adalah "approved" atau "rejected", THE Stok_Opname_System SHALL mencegah penghapusan formulir

### Requirement 11: Search and Filter Stok Opname

**User Story:** Sebagai pengguna sistem, saya ingin mencari dan memfilter stok opname, sehingga saya dapat menemukan formulir tertentu dengan cepat.

#### Acceptance Criteria

1. THE Stok_Opname_System SHALL menyediakan field pencarian untuk mencari berdasarkan nama pembuat atau catatan
2. THE Stok_Opname_System SHALL menyediakan filter berdasarkan Approval_Status (pending, approved, rejected)
3. THE Stok_Opname_System SHALL menyediakan filter berdasarkan rentang tanggal pembuatan
4. WHEN pengguna menerapkan filter atau pencarian, THE Stok_Opname_System SHALL menampilkan hasil yang sesuai dalam waktu kurang dari 1 detik

### Requirement 12: Prevent Duplicate Stock Adjustments

**User Story:** Sebagai sistem, saya ingin mencegah penyesuaian stok ganda, sehingga data stok tetap akurat.

#### Acceptance Criteria

1. WHEN Stok_Opname_Form telah disetujui, THE Stok_Opname_System SHALL menandai formulir sebagai "processed"
2. IF Stok_Opname_Form sudah ditandai "processed", THEN THE Stok_Opname_System SHALL mencegah pembuatan Stock_Adjustment duplikat
3. THE Stok_Opname_System SHALL menyimpan flag "is_processed" pada setiap Stok_Opname_Form
4. WHEN sistem mendeteksi upaya pemrosesan ulang, THE Stok_Opname_System SHALL mencatat error log dan menolak operasi

### Requirement 13: Audit Trail for Stok Opname

**User Story:** Sebagai administrator, saya ingin melihat audit trail lengkap untuk setiap stok opname, sehingga saya dapat melacak siapa melakukan apa dan kapan.

#### Acceptance Criteria

1. THE Stok_Opname_System SHALL mencatat timestamp pembuatan Stok_Opname_Form
2. THE Stok_Opname_System SHALL mencatat user ID pembuat formulir
3. WHERE formulir disetujui atau ditolak, THE Stok_Opname_System SHALL mencatat timestamp approval atau rejection
4. WHERE formulir disetujui atau ditolak, THE Stok_Opname_System SHALL mencatat user ID Kepala_SPPG yang melakukan approval
5. THE Stok_Opname_System SHALL mencatat semua perubahan pada Physical_Count sebelum pengajuan approval
6. THE Stok_Opname_System SHALL menyimpan audit trail minimal selama 2 tahun

### Requirement 14: Handle Concurrent Stock Operations

**User Story:** Sebagai sistem, saya ingin menangani operasi stok yang bersamaan dengan aman, sehingga tidak terjadi race condition atau data corruption.

#### Acceptance Criteria

1. WHEN Stock_Adjustment diterapkan, THE Stok_Opname_System SHALL menggunakan database transaction untuk memastikan atomicity
2. IF transaction gagal, THEN THE Stok_Opname_System SHALL rollback semua perubahan dan menampilkan pesan error
3. WHEN multiple pengguna mengakses Inventory_Item yang sama, THE Stok_Opname_System SHALL menggunakan locking mechanism untuk mencegah race condition
4. IF terjadi conflict saat menyimpan, THEN THE Stok_Opname_System SHALL menampilkan pesan error dan meminta pengguna refresh data

### Requirement 15: Export Stok Opname Report

**User Story:** Sebagai Kepala SPPG, saya ingin mengekspor laporan stok opname, sehingga saya dapat melakukan analisis atau arsip di luar sistem.

#### Acceptance Criteria

1. THE Stok_Opname_System SHALL menyediakan tombol export untuk setiap Stok_Opname_Form
2. WHEN pengguna mengekspor laporan, THE Stok_Opname_System SHALL menghasilkan file dalam format Excel atau PDF
3. THE Stok_Opname_System SHALL menyertakan semua detail formulir dalam file export (tanggal, pembuat, status, items, counts, selisih)
4. THE Stok_Opname_System SHALL menyertakan timestamp export dan nama pengguna yang melakukan export
5. WHEN export selesai, THE Stok_Opname_System SHALL mengunduh file ke perangkat pengguna dalam waktu kurang dari 5 detik untuk formulir dengan maksimal 100 item
