-- Диагностика проблем с созданием заказов
-- Выполнить на сервере для проверки структуры БД

-- 1. Проверяем структуру таблицы order_items
\d order_items;

-- 2. Проверяем структуру таблицы orders
\d orders;

-- 3. Проверяем структуру таблицы addresses
\d addresses;

-- 4. Проверяем структуру таблицы product_variations
\d product_variations;

-- 5. Проверяем существующие индексы
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename IN ('order_items', 'orders', 'addresses', 'product_variations')
ORDER BY tablename, indexname;

-- 6. Проверяем внешние ключи
SELECT
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' 
AND tc.table_name IN ('order_items', 'orders', 'addresses');
