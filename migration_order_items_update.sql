-- Миграция для обновления таблицы order_items
-- Заменяем product_id на variation_id и добавляем size/color

-- 1. Добавляем новые поля
ALTER TABLE order_items 
ADD COLUMN variation_id uuid;

ALTER TABLE order_items 
ADD COLUMN size text;

ALTER TABLE order_items 
ADD COLUMN color text;

-- 2. Добавляем внешний ключ для variation_id
ALTER TABLE order_items 
ADD CONSTRAINT fk_order_items_variation_id 
FOREIGN KEY (variation_id) REFERENCES product_variations(id);

-- 3. Переносим данные (если есть)
-- UPDATE order_items 
-- SET variation_id = (
--   SELECT pv.id 
--   FROM product_variations pv 
--   WHERE pv.product_id = order_items.product_id 
--   LIMIT 1
-- );

-- 4. Удаляем старые поля (выполнить после проверки данных)
-- ALTER TABLE order_items DROP COLUMN product_id;
