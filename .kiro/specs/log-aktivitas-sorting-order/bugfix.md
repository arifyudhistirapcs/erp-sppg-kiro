# Bugfix Requirements Document

## Introduction

Log aktivitas pada halaman monitoring deliveries saat ini menampilkan data dalam urutan ascending (dari yang terlama ke yang terbaru), yang menyulitkan pengguna untuk melihat aktivitas terkini. Bug ini terjadi pada halaman monitoring deliveries di URL `/logistics/monitoring/deliveries/{id}`. Perbaikan ini akan memastikan log aktivitas ditampilkan dalam urutan descending berdasarkan timestamp, sehingga aktivitas terbaru muncul di bagian atas tabel.

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN log aktivitas ditampilkan pada halaman monitoring deliveries THEN the system mengurutkan data dari waktu terlama ke waktu terbaru (ascending order)

1.2 WHEN pengguna membuka halaman monitoring deliveries THEN aktivitas terlama muncul di bagian atas tabel dan aktivitas terbaru muncul di bagian bawah

### Expected Behavior (Correct)

2.1 WHEN log aktivitas ditampilkan pada halaman monitoring deliveries THEN the system SHALL mengurutkan data dari waktu terbaru ke waktu terlama (descending order by timestamp)

2.2 WHEN pengguna membuka halaman monitoring deliveries THEN aktivitas terbaru SHALL muncul di bagian atas tabel dan aktivitas terlama muncul di bagian bawah

### Unchanged Behavior (Regression Prevention)

3.1 WHEN log aktivitas ditampilkan THEN the system SHALL CONTINUE TO menampilkan semua kolom yang ada (Waktu, Status Awal, Status Baru, Pengguna, Durasi, Catatan)

3.2 WHEN log aktivitas ditampilkan THEN the system SHALL CONTINUE TO menampilkan semua record aktivitas yang tersedia tanpa ada data yang hilang

3.3 WHEN log aktivitas ditampilkan THEN the system SHALL CONTINUE TO memformat waktu dalam format yang sama (DD MMM YYYY, HH:mm WIB)

3.4 WHEN pengguna berinteraksi dengan fitur lain pada halaman monitoring deliveries THEN the system SHALL CONTINUE TO berfungsi dengan normal
