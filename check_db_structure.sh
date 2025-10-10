#!/bin/bash
# Скрипт для проверки структуры БД

echo "🔍 Проверка структуры таблицы order_items..."
echo ""

echo "📋 Структура таблицы order_items:"
docker exec mm-postgres-prod psql -U mm_user -d mm_shop_prod -c "\d order_items"

echo ""
echo "📊 Количество записей в order_items:"
docker exec mm-postgres-prod psql -U mm_user -d mm_shop_prod -c "SELECT COUNT(*) as total_order_items FROM order_items;"

echo ""
echo "🔗 Индексы таблицы order_items:"
docker exec mm-postgres-prod psql -U mm_user -d mm_shop_prod -c "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'order_items';"

echo ""
echo "🔑 Внешние ключи таблицы order_items:"
docker exec mm-postgres-prod psql -U mm_user -d mm_shop_prod -c "
SELECT 
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY' 
AND tc.table_name = 'order_items';"

echo ""
echo "✅ Проверка завершена!"

