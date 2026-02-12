-- Добавление уникальных индексов для таблицы shop_clients
-- Уникальность по паре shop_id + phone (один клиент с одним номером телефона в одном магазине)
-- Уникальность по паре shop_id + qr_code (один QR код в одном магазине)

-- Создаем уникальный индекс для shop_id + phone
CREATE UNIQUE INDEX IF NOT EXISTS idx_shop_clients_shop_phone_unique 
ON shop_clients(shop_id, phone);

-- Создаем уникальный индекс для shop_id + qr_code
CREATE UNIQUE INDEX IF NOT EXISTS idx_shop_clients_shop_qr_unique 
ON shop_clients(shop_id, qr_code);

