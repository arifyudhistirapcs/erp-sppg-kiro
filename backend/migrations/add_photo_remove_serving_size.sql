-- Migration: Add photo_url and remove serving_size from recipes table
-- Date: 2026-02-25

-- Add photo_url column to recipes table
ALTER TABLE recipes ADD COLUMN IF NOT EXISTS photo_url VARCHAR(500);

-- Remove serving_size column from recipes table
ALTER TABLE recipes DROP COLUMN IF EXISTS serving_size;

-- Add photo_url column to recipe_versions table (for historical data)
ALTER TABLE recipe_versions ADD COLUMN IF NOT EXISTS photo_url VARCHAR(500);

-- Remove serving_size column from recipe_versions table
ALTER TABLE recipe_versions DROP COLUMN IF EXISTS serving_size;

-- Add index on photo_url for faster queries
CREATE INDEX IF NOT EXISTS idx_recipes_photo_url ON recipes(photo_url);
