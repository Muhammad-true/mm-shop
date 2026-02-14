-- Миграция: Добавление поля image_urls_by_color в таблицу product_variations
-- Это поле хранит фото по цветам: цвет -> массив URL фото (максимум 2 фото на цвет)
-- Дата: 2025-01-XX

-- Добавляем поле image_urls_by_color типа JSONB
-- Если поле уже существует, команда не выполнится (безопасно)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'product_variations' 
        AND column_name = 'image_urls_by_color'
    ) THEN
        ALTER TABLE product_variations 
        ADD COLUMN image_urls_by_color JSONB DEFAULT '{}'::jsonb;
        
        -- Создаем индекс для быстрого поиска (опционально)
        CREATE INDEX IF NOT EXISTS idx_product_variations_image_urls_by_color 
        ON product_variations USING GIN (image_urls_by_color);
        
        RAISE NOTICE 'Поле image_urls_by_color успешно добавлено в product_variations';
    ELSE
        RAISE NOTICE 'Поле image_urls_by_color уже существует в product_variations';
    END IF;
END $$;

