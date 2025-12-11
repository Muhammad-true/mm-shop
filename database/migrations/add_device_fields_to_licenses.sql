-- Миграция: Добавление полей для защиты от повторного использования лицензий
-- Дата: 2024-12-01
-- Описание: Добавляет поля device_id, device_info и device_fingerprint в таблицу licenses

-- Добавляем поле device_id (уникальный ID устройства)
ALTER TABLE licenses 
ADD COLUMN IF NOT EXISTS device_id VARCHAR(255);

-- Добавляем индекс для device_id для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_licenses_device_id ON licenses(device_id);

-- Добавляем поле device_info (JSON с информацией о железе)
ALTER TABLE licenses 
ADD COLUMN IF NOT EXISTS device_info TEXT;

-- Добавляем поле device_fingerprint (хеш для быстрой проверки)
ALTER TABLE licenses 
ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(255);

-- Добавляем индекс для device_fingerprint
CREATE INDEX IF NOT EXISTS idx_licenses_device_fingerprint ON licenses(device_fingerprint);

-- Комментарии к полям
COMMENT ON COLUMN licenses.device_id IS 'Уникальный ID устройства, на котором активирована лицензия';
COMMENT ON COLUMN licenses.device_info IS 'JSON с информацией о железе устройства (платформа, модель, производитель, версия ОС и т.д.)';
COMMENT ON COLUMN licenses.device_fingerprint IS 'SHA256 хеш для быстрой проверки соответствия устройства';

