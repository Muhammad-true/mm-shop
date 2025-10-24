# Тестирование запросов для отслеживания заказов

## SQL запрос, который использует GetActiveOrder

```sql
-- Это эквивалент Go запроса в GetActiveOrder
SELECT o.*
FROM orders o
WHERE o.user_id = 'USER_UUID'  -- Замените на реальный UUID
  AND o.status NOT IN ('completed', 'cancelled')
ORDER BY o.created_at DESC
LIMIT 1;
```

## Проверка данных

### 1. Посмотреть все заказы пользователя
```sql
SELECT id, user_id, status, created_at, total_amount
FROM orders
WHERE user_id = 'USER_UUID'
ORDER BY created_at DESC;
```

### 2. Посмотреть активные заказы всех пользователей
```sql
SELECT 
    u.name,
    u.phone,
    o.id as order_id,
    o.status,
    o.total_amount,
    o.created_at
FROM orders o
JOIN users u ON o.user_id = u.id
WHERE o.status NOT IN ('completed', 'cancelled')
ORDER BY o.created_at DESC;
```

### 3. Статистика по статусам
```sql
SELECT 
    status,
    COUNT(*) as count,
    SUM(total_amount) as total_sum
FROM orders
GROUP BY status
ORDER BY count DESC;
```

## Тестирование через API

### Curl команда для проверки эндпоинта
```bash
# Получить активный заказ
curl -X GET "http://159.89.99.252/api/v1/orders/active" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json"
```

### Ожидаемые ответы

**Если есть активный заказ:**
```json
{
  "success": true,
  "message": "Активный заказ получен",
  "data": {
    "order": { ... },
    "tracking": { ... }
  }
}
```

**Если нет активных заказов:**
```json
{
  "success": true,
  "message": "Нет активных заказов",
  "data": null
}
```

## Проверка производительности

### EXPLAIN запроса
```sql
EXPLAIN ANALYZE
SELECT o.*
FROM orders o
WHERE o.user_id = 'USER_UUID'
  AND o.status NOT IN ('completed', 'cancelled')
ORDER BY o.created_at DESC
LIMIT 1;
```

Ожидается использование индекса `idx_orders_user_status_created` для быстрого поиска.

