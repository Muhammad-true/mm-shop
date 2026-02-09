#!/bin/bash

# Скрипт для обновления токена в systemd сервисе
# Использование: ./update_token.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="ftp-upload-watcher"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

echo "🔑 Обновление API токена для FTP Upload Watcher"
echo ""

# Проверка прав root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Ошибка: скрипт должен быть запущен от root"
    echo "   Используй: sudo ./update_token.sh"
    exit 1
fi

# Проверка существования сервиса
if [ ! -f "$SERVICE_FILE" ]; then
    echo "❌ Ошибка: сервис не установлен"
    echo "   Сначала запусти: sudo ./setup_ftp_watcher.sh"
    exit 1
fi

# Получение нового токена
echo "Введите данные для получения нового токена:"
read -p "Телефон: " PHONE
read -sp "Пароль: " PASSWORD
echo ""

API_BASE_URL="${API_BASE_URL:-https://api.libiss.com/api/v1}"
RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"password\": \"$PASSWORD\"
  }")

API_TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$API_TOKEN" ]; then
    echo "❌ Ошибка получения токена"
    echo "Ответ сервера:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    exit 1
fi

echo "✅ Новый токен получен"

# Обновление файла сервиса
echo ""
echo "📝 Обновление файла сервиса..."

# Читаем текущий файл и заменяем токен
sed -i "s|Environment=\"API_TOKEN=.*\"|Environment=\"API_TOKEN=$API_TOKEN\"|" "$SERVICE_FILE"

echo "   ✅ Токен обновлен в $SERVICE_FILE"

# Перезагрузка systemd
echo ""
echo "🔄 Перезагрузка systemd..."
systemctl daemon-reload
echo "   ✅ systemd перезагружен"

# Перезапуск сервиса
echo ""
echo "▶️  Перезапуск сервиса..."
systemctl restart "$SERVICE_NAME"
echo "   ✅ Сервис перезапущен"

# Проверка статуса
echo ""
echo "📊 Статус сервиса:"
systemctl status "$SERVICE_NAME" --no-pager -l

echo ""
echo "✅ Токен успешно обновлен!"

