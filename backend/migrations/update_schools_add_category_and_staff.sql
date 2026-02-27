-- Add new columns to schools table
ALTER TABLE schools ADD COLUMN IF NOT EXISTS category VARCHAR(10);
ALTER TABLE schools ADD COLUMN IF NOT EXISTS student_count_grade_1_3 INTEGER DEFAULT 0;
ALTER TABLE schools ADD COLUMN IF NOT EXISTS student_count_grade_4_6 INTEGER DEFAULT 0;
ALTER TABLE schools ADD COLUMN IF NOT EXISTS staff_count INTEGER DEFAULT 0;
ALTER TABLE schools ADD COLUMN IF NOT EXISTS npsn VARCHAR(50);
ALTER TABLE schools ADD COLUMN IF NOT EXISTS principal_name VARCHAR(255);
ALTER TABLE schools ADD COLUMN IF NOT EXISTS school_email VARCHAR(255);
ALTER TABLE schools ADD COLUMN IF NOT EXISTS school_phone VARCHAR(50);
ALTER TABLE schools ADD COLUMN IF NOT EXISTS committee_count INTEGER DEFAULT 0;
ALTER TABLE schools ADD COLUMN IF NOT EXISTS cooperation_letter_url VARCHAR(500);

-- Rename student_count to be more generic (will be used for SMP/SMA)
-- For SD, we'll use student_count_grade_1_3 and student_count_grade_4_6
-- For SMP/SMA, we'll use student_count

-- Add check constraint for category
ALTER TABLE schools ADD CONSTRAINT check_category CHECK (category IN ('SD', 'SMP', 'SMA'));

-- Add comment
COMMENT ON COLUMN schools.category IS 'Kategori sekolah: SD, SMP, atau SMA';
COMMENT ON COLUMN schools.student_count_grade_1_3 IS 'Jumlah siswa kelas 1-3 (khusus SD)';
COMMENT ON COLUMN schools.student_count_grade_4_6 IS 'Jumlah siswa kelas 4-6 (khusus SD)';
COMMENT ON COLUMN schools.staff_count IS 'Jumlah guru/karyawan';
COMMENT ON COLUMN schools.npsn IS 'Nomor Pokok Sekolah Nasional';
COMMENT ON COLUMN schools.principal_name IS 'Nama Kepala Sekolah';
COMMENT ON COLUMN schools.school_email IS 'Email Sekolah';
COMMENT ON COLUMN schools.school_phone IS 'Nomor Telepon Sekolah';
COMMENT ON COLUMN schools.committee_count IS 'Jumlah Anggota Komite';
COMMENT ON COLUMN schools.cooperation_letter_url IS 'URL Surat Perjanjian Kerjasama';
