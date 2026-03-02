# Bugfix Requirements Document

## Introduction

Sistem KDS (Kitchen Display System) saat ini memungkinkan proses memasak dimulai tanpa memvalidasi ketersediaan stok komponen bahan yang dibutuhkan. Hal ini menyebabkan masalah operasional dimana pesanan dapat dimasak meskipun bahan tidak tersedia, dan stok tidak berkurang secara otomatis saat memasak dimulai. Bug ini berdampak pada akurasi inventori dan dapat menyebabkan kegagalan dalam memenuhi pesanan.

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN pengguna mengklik tombol "Mulai Masak" THEN sistem mengizinkan proses memasak dimulai tanpa memeriksa ketersediaan stok komponen bahan

1.2 WHEN status pesanan berubah menjadi "sedang dimasak" THEN sistem tidak mengurangi stok komponen bahan yang dibutuhkan

1.3 WHEN stok komponen bahan tidak mencukupi THEN sistem tetap mengizinkan proses memasak dimulai tanpa peringatan atau error

### Expected Behavior (Correct)

2.1 WHEN pengguna mengklik tombol "Mulai Masak" THEN sistem SHALL memvalidasi bahwa semua komponen bahan tersedia dalam jumlah yang mencukupi sesuai resep

2.2 WHEN stok komponen bahan tidak mencukupi THEN sistem SHALL mencegah proses memasak dimulai dan menampilkan pesan error yang menjelaskan komponen mana yang tidak tersedia

2.3 WHEN stok komponen bahan mencukupi dan status berubah menjadi "sedang dimasak" THEN sistem SHALL secara otomatis mengurangi stok komponen bahan sesuai dengan jumlah yang dibutuhkan dalam resep

### Unchanged Behavior (Regression Prevention)

3.1 WHEN pengguna mengklik tombol "Mulai Masak" dan semua stok mencukupi THEN sistem SHALL CONTINUE TO mengubah status pesanan menjadi "sedang dimasak" seperti biasa

3.2 WHEN proses memasak selesai atau dibatalkan THEN sistem SHALL CONTINUE TO memproses perubahan status tanpa mempengaruhi stok (stok hanya berkurang saat mulai memasak)

3.3 WHEN pengguna melihat daftar pesanan di KDS THEN sistem SHALL CONTINUE TO menampilkan informasi pesanan dengan benar tanpa perubahan tampilan

3.4 WHEN pesanan dalam status selain "sedang dimasak" THEN sistem SHALL CONTINUE TO tidak mempengaruhi stok komponen bahan
