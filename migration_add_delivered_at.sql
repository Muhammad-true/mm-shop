-- Миграция для изменения типа поля desired_at на timestamp with time zone
-- Выполнить: psql -d your_database -f migration_add_delivered_at.sql

-- Изменяем тип колонки desired_at на timestamp with time zone
-- Сначала проверяем, существует ли колонка
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'orders' AND column_name = 'desired_at') THEN
        -- Если колонка существует, изменяем её тип
        ALTER TABLE orders 
        ALTER COLUMN desired_at TYPE TIMESTAMP WITH TIME ZONE;
    ELSE
        -- Если колонки нет, создаем её
        ALTER TABLE orders 
        ADD COLUMN desired_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- Добавляем комментарий к колонке
COMMENT ON COLUMN orders.desired_at IS 'Желаемое время доставки заказа (указывает пользователь)';

-- Создаем индекс для быстрого поиска по желаемому времени доставки
CREATE INDEX IF NOT EXISTS idx_orders_desired_at ON orders(desired_at);
