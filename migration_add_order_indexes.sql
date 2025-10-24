-- Добавление индексов для оптимизации запросов к заказам

-- Индекс для быстрого поиска заказов пользователя по статусу
CREATE INDEX IF NOT EXISTS idx_orders_user_status_created 
ON orders(user_id, status, created_at DESC);

-- Индекс для быстрого поиска активных заказов
CREATE INDEX IF NOT EXISTS idx_orders_status_created 
ON orders(status, created_at DESC);

-- Индекс для order_items по order_id (если еще нет)
CREATE INDEX IF NOT EXISTS idx_order_items_order_id 
ON order_items(order_id);

-- Проверим созданные индексы
SELECT 
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename IN ('orders', 'order_items')
ORDER BY tablename, indexname;

