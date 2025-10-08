-- Миграция для исправления order_items с правильными параметрами подключения
-- Выполнить на сервере: psql -h mm-postgres-prod -U mm_user -d mm_shop_prod -f migration_fix_order_items.sql

-- 1. Добавляем поле variation_id
ALTER TABLE order_items 
ADD COLUMN IF NOT EXISTS variation_id uuid;

-- 2. Добавляем поля size и color
ALTER TABLE order_items 
ADD COLUMN IF NOT EXISTS size text;

ALTER TABLE order_items 
ADD COLUMN IF NOT EXISTS color text;

-- 3. Добавляем внешний ключ для variation_id
ALTER TABLE order_items 
ADD CONSTRAINT IF NOT EXISTS fk_order_items_variation_id 
FOREIGN KEY (variation_id) REFERENCES product_variations(id);

-- 4. Добавляем индексы для производительности
CREATE INDEX IF NOT EXISTS idx_order_items_variation_id ON order_items(variation_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- 5. Проверяем структуру таблицы
\d order_items;

-- 6. Показываем количество записей
SELECT COUNT(*) as total_order_items FROM order_items;
