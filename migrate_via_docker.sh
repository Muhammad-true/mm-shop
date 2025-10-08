#!/bin/bash
# Скрипт для выполнения миграции через Docker

echo "🔧 Выполнение миграции order_items через Docker..."

# Выполняем SQL команды через docker exec
docker exec mm-postgres-prod psql -U mm_user -d mm_shop_prod -c "
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS variation_id uuid;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS size text;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS color text;
ALTER TABLE order_items ADD CONSTRAINT IF NOT EXISTS fk_order_items_variation_id FOREIGN KEY (variation_id) REFERENCES product_variations(id);
CREATE INDEX IF NOT EXISTS idx_order_items_variation_id ON order_items(variation_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
"

echo "✅ Миграция завершена!"

# Проверяем структуру таблицы
echo "📋 Проверка структуры order_items:"
docker exec mm-postgres-prod psql -U mm_user -d mm_shop_prod -c "\d order_items"

echo "📊 Количество записей в order_items:"
docker exec mm-postgres-prod psql -U mm_user -d mm_shop_prod -c "SELECT COUNT(*) as total_order_items FROM order_items;"
