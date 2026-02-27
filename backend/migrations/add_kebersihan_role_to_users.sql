-- Migration: Add kebersihan role to users table
-- Date: 2026-02-26
-- Feature: Logistics Monitoring Process
-- Requirements: 8.1, 8.4

-- Add CHECK constraint for role validation including kebersihan
-- First, drop the existing constraint if it exists
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_user_role;

-- Add new constraint with kebersihan role included
ALTER TABLE users ADD CONSTRAINT check_user_role 
    CHECK (role IN (
        'kepala_sppg', 
        'kepala_yayasan', 
        'akuntan', 
        'ahli_gizi', 
        'pengadaan', 
        'chef', 
        'packing', 
        'driver', 
        'asisten_lapangan', 
        'kebersihan'
    ));

-- Add comment for documentation
COMMENT ON CONSTRAINT check_user_role ON users IS 'Validates that user role is one of the allowed values including kebersihan for cleaning staff';
