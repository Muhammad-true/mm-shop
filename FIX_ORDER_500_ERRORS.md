# 🔧 Исправление ошибок 500 при создании заказов

## 🚨 Проблема
Сервер возвращает ошибку 500 при создании заказов через эндпоинты:
- `POST /api/v1/orders/` (авторизованные пользователи)
- `POST /api/v1/guest-orders/` (гости)

## 🔍 Причина
Проблема в том, что код был обновлен для использования `variation_id` в `OrderItem`, но база данных не была обновлена.

## ✅ Решение

### 1. Выполнить миграцию на сервере:
```bash
# Подключиться к серверу
ssh root@159.89.99.252

# Перейти в папку проекта
cd mm-shop

# Выполнить миграцию
psql -h localhost -U postgres -d mm_shop -f migration_add_variation_id.sql
```

### 2. Альтернативно - выполнить SQL команды вручную:
```sql
-- Подключиться к PostgreSQL
psql -h localhost -U postgres -d mm_shop

-- Выполнить команды:
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS variation_id uuid;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS size text;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS color text;

-- Добавить внешний ключ
ALTER TABLE order_items 
ADD CONSTRAINT IF NOT EXISTS fk_order_items_variation_id 
FOREIGN KEY (variation_id) REFERENCES product_variations(id);

-- Добавить индексы
CREATE INDEX IF NOT EXISTS idx_order_items_variation_id ON order_items(variation_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
```

### 3. Перезапустить API сервер:
```bash
# В папке mm-shop
docker-compose -f docker-compose.release.yml restart api
```

### 4. Проверить логи:
```bash
docker-compose -f docker-compose.release.yml logs -f api
```

## 🧪 Тестирование

После исправления протестировать:

```bash
# Создание гостевого заказа
curl -X POST "http://159.89.99.252:8080/api/v1/guest-orders/" \
  -H "Content-Type: application/json" \
  -d '{
    "guest_name": "Тест Тестов",
    "guest_phone": "+992900000001",
    "shipping_addr": "Тестовый адрес",
    "payment_method": "cash",
    "items": [{
      "variation_id": "REAL_VARIATION_ID",
      "quantity": 1,
      "price": 100.0,
      "size": "M",
      "color": "Белый",
      "sku": "TEST-001",
      "name": "Тестовый товар"
    }]
  }'
```

## 📋 Диагностика

Если проблема остается, выполнить диагностику:
```bash
psql -h localhost -U postgres -d mm_shop -f debug_order_creation.sql
```

## 🎯 Ожидаемый результат
После исправления:
- ✅ `POST /api/v1/orders/` работает
- ✅ `POST /api/v1/guest-orders/` работает
- ✅ Заказы создаются в базе данных
- ✅ Логи показывают успешное создание
