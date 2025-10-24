-- Тестовый SQL запрос для проверки получения активного заказа

-- Проверим структуру таблицы orders
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'orders'
ORDER BY ordinal_position;

-- Получить активный заказ для пользователя (пример)
-- Замените 'USER_UUID_HERE' на реальный UUID пользователя
SELECT 
    o.id,
    o.user_id,
    o.status,
    o.total_amount,
    o.shipping_addr,
    o.recipient_name,
    o.phone,
    o.desired_at,
    o.created_at,
    o.updated_at
FROM orders o
WHERE o.user_id = 'USER_UUID_HERE'
  AND o.status NOT IN ('completed', 'cancelled')
ORDER BY o.created_at DESC
LIMIT 1;

-- Посмотреть все статусы заказов в системе
SELECT status, COUNT(*) as count
FROM orders
GROUP BY status;

-- Проверить есть ли активные заказы у всех пользователей
SELECT 
    u.id as user_id,
    u.name,
    u.phone,
    o.id as order_id,
    o.status,
    o.created_at
FROM users u
LEFT JOIN orders o ON u.id = o.user_id 
    AND o.status NOT IN ('completed', 'cancelled')
WHERE o.id IS NOT NULL
ORDER BY o.created_at DESC;

