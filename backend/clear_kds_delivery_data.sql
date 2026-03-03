-- Clear KDS and Delivery Data (PostgreSQL)
-- This script removes all KDS, delivery, and monitoring data while preserving master data

-- Clear ompreng tables first (before delivery_records due to foreign keys)
DO $$ 
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'ompreng_cleanings') THEN
        TRUNCATE TABLE ompreng_cleanings CASCADE;
    END IF;
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'ompreng_trackings') THEN
        TRUNCATE TABLE ompreng_trackings CASCADE;
    END IF;
END $$;

-- Clear status transitions (references delivery_records)
TRUNCATE TABLE status_transitions CASCADE;

-- Clear delivery-related tables (CASCADE will handle foreign key constraints)
TRUNCATE TABLE delivery_menu_items CASCADE;
TRUNCATE TABLE delivery_tasks CASCADE;
TRUNCATE TABLE electronic_pods CASCADE;
TRUNCATE TABLE delivery_records CASCADE;

-- Display confirmation
SELECT 'KDS and Delivery data cleared successfully!' AS status;
SELECT 'Master data (schools, recipes, ingredients, users) preserved.' AS note;
