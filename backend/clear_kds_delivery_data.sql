-- Clear KDS and Delivery Data (PostgreSQL)
-- This script removes all KDS, delivery, and monitoring data while preserving master data

-- Clear delivery-related tables (CASCADE will handle foreign key constraints)
TRUNCATE TABLE delivery_menu_items CASCADE;
TRUNCATE TABLE delivery_tasks CASCADE;
TRUNCATE TABLE electronic_pods CASCADE;
TRUNCATE TABLE status_transitions CASCADE;
TRUNCATE TABLE delivery_records CASCADE;

-- Try to clear ompreng tables if they exist
DO $$ 
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'ompreng_trackings') THEN
        TRUNCATE TABLE ompreng_trackings CASCADE;
    END IF;
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'ompreng_cleanings') THEN
        TRUNCATE TABLE ompreng_cleanings CASCADE;
    END IF;
END $$;

-- Display confirmation
SELECT 'KDS and Delivery data cleared successfully!' AS status;
SELECT 'Master data (schools, recipes, ingredients, users) preserved.' AS note;
