-- Добавление поля discount в таблицу product_variations
-- Дата: 2025-10-23
-- Описание: Добавляет поле discount (скидка в процентах 0-100%) к вариациям товаров
-- Например: discount = 15 означает скидку 15%

ALTER TABLE product_variations 
ADD COLUMN IF NOT EXISTS discount INTEGER DEFAULT 0 CHECK (discount >= 0 AND discount <= 100);

-- Комментарий к полю
COMMENT ON COLUMN product_variations.discount IS 'Скидка в процентах (0-100%), например: 15 = 15%';

