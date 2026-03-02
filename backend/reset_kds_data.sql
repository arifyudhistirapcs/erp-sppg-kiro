-- Reset KDS cooking and packing status to start fresh

-- Reset cooking status for all menu items to 'pending'
UPDATE menu_items 
SET cooking_status = 'pending'
WHERE cooking_status IN ('cooking', 'ready');

-- Reset packing status for all school allocations to 'pending'
UPDATE school_allocations 
SET packing_status = 'pending'
WHERE packing_status IN ('packing', 'ready');

-- Verify reset
SELECT 
    cooking_status, 
    COUNT(*) as count 
FROM menu_items 
GROUP BY cooking_status;

SELECT 
    packing_status, 
    COUNT(*) as count 
FROM school_allocations 
GROUP BY packing_status;
