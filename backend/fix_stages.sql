-- Fix current_stage based on current_status for existing delivery records

UPDATE delivery_records 
SET current_stage = CASE current_status
    WHEN 'order_disiapkan' THEN 1
    WHEN 'sedang_dimasak' THEN 2
    WHEN 'selesai_dimasak' THEN 3
    WHEN 'siap_dipacking' THEN 4
    WHEN 'selesai_dipacking' THEN 5
    WHEN 'siap_dikirim' THEN 6
    WHEN 'diperjalanan' THEN 7
    WHEN 'sudah_sampai_sekolah' THEN 8
    WHEN 'sudah_diterima_pihak_sekolah' THEN 9
    WHEN 'driver_menuju_lokasi_pengambilan' THEN 10
    WHEN 'driver_tiba_di_lokasi_pengambilan' THEN 11
    WHEN 'driver_kembali_ke_sppg' THEN 12
    WHEN 'driver_tiba_di_sppg' THEN 13
    WHEN 'ompreng_siap_dicuci' THEN 14
    WHEN 'ompreng_proses_pencucian' THEN 15
    WHEN 'ompreng_selesai_dicuci' THEN 16
    ELSE current_stage
END;

-- Also update status_transitions table to have correct stage numbers
UPDATE status_transitions 
SET stage = CASE to_status
    WHEN 'order_disiapkan' THEN 1
    WHEN 'sedang_dimasak' THEN 2
    WHEN 'selesai_dimasak' THEN 3
    WHEN 'siap_dipacking' THEN 4
    WHEN 'selesai_dipacking' THEN 5
    WHEN 'siap_dikirim' THEN 6
    WHEN 'diperjalanan' THEN 7
    WHEN 'sudah_sampai_sekolah' THEN 8
    WHEN 'sudah_diterima_pihak_sekolah' THEN 9
    WHEN 'driver_menuju_lokasi_pengambilan' THEN 10
    WHEN 'driver_tiba_di_lokasi_pengambilan' THEN 11
    WHEN 'driver_kembali_ke_sppg' THEN 12
    WHEN 'driver_tiba_di_sppg' THEN 13
    WHEN 'ompreng_siap_dicuci' THEN 14
    WHEN 'ompreng_proses_pencucian' THEN 15
    WHEN 'ompreng_selesai_dicuci' THEN 16
    ELSE stage
END;
