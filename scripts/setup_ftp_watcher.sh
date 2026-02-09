#!/bin/bash

# Скрипт для автоматической установки и настройки FTP Upload Watcher
# Использование: ./setup_ftp_watcher.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="ftp-upload-watcher"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

echo "🚀 Установка FTP Upload Watcher"
echo ""

# Проверка прав root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Ошибка: скрипт должен быть запущен от root"
    echo "   Используй: sudo ./setup_ftp_watcher.sh"
    exit 1
fi

# 1. Получение токена (если не установлен)
if [ -z "$API_TOKEN" ]; then
    echo "🔑 Получение API токена..."
    echo ""
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
    
    echo "✅ Токен получен"
else
    echo "✅ Используется токен из переменной окружения"
fi

# 2. Создание папок
echo ""
echo "📁 Создание папок..."
FTP_WATCH_DIR="${FTP_WATCH_DIR:-/var/ftp/uploads}"
PROCESSED_DIR="${PROCESSED_DIR:-/var/ftp/processed}"

mkdir -p "$FTP_WATCH_DIR"
mkdir -p "$PROCESSED_DIR"
chown root:root "$FTP_WATCH_DIR"
chown root:root "$PROCESSED_DIR"
chmod 755 "$FTP_WATCH_DIR"
chmod 755 "$PROCESSED_DIR"
echo "   ✅ $FTP_WATCH_DIR"
echo "   ✅ $PROCESSED_DIR"

# 3. Установка прав на скрипт
echo ""
echo "🔧 Настройка скрипта..."
chmod +x "$SCRIPT_DIR/ftp_upload_watcher.sh"
echo "   ✅ Скрипт готов к запуску"

# 4. Создание файла сервиса
echo ""
echo "📝 Создание systemd сервиса..."
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=FTP Upload Watcher - Automatic update uploader
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$SCRIPT_DIR
Environment="FTP_WATCH_DIR=$FTP_WATCH_DIR"
Environment="API_BASE_URL=${API_BASE_URL:-https://api.libiss.com/api/v1}"
Environment="API_TOKEN=$API_TOKEN"
Environment="PROCESSED_DIR=$PROCESSED_DIR"
Environment="CHECK_INTERVAL=${CHECK_INTERVAL:-30}"
ExecStart=$SCRIPT_DIR/ftp_upload_watcher.sh
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

echo "   ✅ Файл сервиса создан: $SERVICE_FILE"

# 5. Перезагрузка systemd
echo ""
echo "🔄 Перезагрузка systemd..."
systemctl daemon-reload
echo "   ✅ systemd перезагружен"

# 6. Включение автозапуска
echo ""
echo "⚙️  Включение автозапуска..."
systemctl enable "$SERVICE_NAME"
echo "   ✅ Автозапуск включен"

# 7. Запуск сервиса
echo ""
echo "▶️  Запуск сервиса..."
if systemctl is-active --quiet "$SERVICE_NAME"; then
    systemctl restart "$SERVICE_NAME"
    echo "   ✅ Сервис перезапущен"
else
    systemctl start "$SERVICE_NAME"
    echo "   ✅ Сервис запущен"
fi

# 8. Проверка статуса
echo ""
echo "📊 Статус сервиса:"
systemctl status "$SERVICE_NAME" --no-pager -l

echo ""
echo "✅ Установка завершена!"
echo ""
echo "📋 Полезные команды:"
echo "   Статус:     sudo systemctl status $SERVICE_NAME"
echo "   Логи:       sudo journalctl -u $SERVICE_NAME -f"
echo "   Остановить: sudo systemctl stop $SERVICE_NAME"
echo "   Запустить:  sudo systemctl start $SERVICE_NAME"
echo "   Перезапуск: sudo systemctl restart $SERVICE_NAME"
echo ""
echo "📁 Папка для загрузки файлов: $FTP_WATCH_DIR"
echo "📦 Папка обработанных файлов: $PROCESSED_DIR"
echo ""
echo "⚠️  ВАЖНО: Токен истекает через 24 часа!"
echo "   Для обновления токена запусти: ./update_token.sh"

