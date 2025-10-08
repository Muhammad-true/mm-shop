-- Миграция для добавления variation_id в order_items
-- Выполнить на сервере для исправления ошибок 500

-- 1. Добавляем поле variation_id
ALTER TABLE order_items 
ADD COLUMN IF NOT EXISTS variation_id uuid;

-- 2. Добавляем внешний ключ для variation_id
ALTER TABLE order_items 
ADD CONSTRAINT IF NOT EXISTS fk_order_items_variation_id 
FOREIGN KEY (variation_id) REFERENCES product_variations(id);

-- 3. Добавляем индексы для производительности
CREATE INDEX IF NOT EXISTS idx_order_items_variation_id ON order_items(variation_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- 4. Проверяем, что поля size и color существуют
ALTER TABLE order_items 
ADD COLUMN IF NOT EXISTS size text;

ALTER TABLE order_items 
ADD COLUMN IF NOT EXISTS color text;

-- 5. Проверяем структуру таблицы
\d order_items;
