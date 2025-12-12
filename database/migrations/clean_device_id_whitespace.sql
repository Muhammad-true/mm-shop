-- Миграция: Очистка лишних пробелов и переносов строк из device_id
-- Дата: 2024-12-01
-- Описание: Удаляет пробелы и переносы строк из поля device_id в таблице licenses

-- Обновляем все записи, удаляя пробелы и переносы строк из device_id
UPDATE licenses 
SET device_id = TRIM(REGEXP_REPLACE(device_id, E'[\\n\\r\\t]+', '', 'g'))
WHERE device_id IS NOT NULL 
  AND device_id != TRIM(REGEXP_REPLACE(device_id, E'[\\n\\r\\t]+', '', 'g'));

-- Для PostgreSQL можно использовать более простой вариант:
-- UPDATE licenses 
-- SET device_id = TRIM(device_id)
-- WHERE device_id IS NOT NULL 
--   AND device_id != TRIM(device_id);

-- Проверяем результат
SELECT 
    id, 
    license_key, 
    shop_id, 
    device_id,
    LENGTH(device_id) as device_id_length,
    LENGTH(TRIM(device_id)) as trimmed_length
FROM licenses 
WHERE device_id IS NOT NULL;

