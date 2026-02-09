#!/bin/bash

# Скрипт для настройки автоматической очистки старых обновлений
# Использование: ./setup_cleanup_timer.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="cleanup-old-updates"
TIMER_NAME="${SERVICE_NAME}.timer"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
TIMER_FILE="/etc/systemd/system/${TIMER_NAME}"

echo "🧹 Настройка автоматической очистки старых обновлений"
echo ""

# Проверка прав root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Ошибка: скрипт должен быть запущен от root"
    echo "   Используй: sudo ./setup_cleanup_timer.sh"
    exit 1
fi

# Создание файла сервиса
echo "📝 Создание systemd сервиса..."
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Cleanup Old Updates - Remove update files older than 10 days
After=network.target

[Service]
Type=oneshot
User=root
WorkingDirectory=/root/mm-shop/release
ExecStart=$SCRIPT_DIR/cleanup_old_updates.sh 10 ./updates
StandardOutput=journal
StandardError=journal
EOF

echo "   ✅ Файл сервиса создан: $SERVICE_FILE"

# Создание файла таймера
echo ""
echo "⏰ Создание systemd таймера..."
cat > "$TIMER_FILE" <<EOF
[Unit]
Description=Run cleanup old updates daily
Requires=${SERVICE_NAME}.service

[Timer]
# Запуск каждый день в 3:00 ночи
OnCalendar=daily
OnCalendar=03:00
Persistent=true

[Install]
WantedBy=timers.target
EOF

echo "   ✅ Файл таймера создан: $TIMER_FILE"

# Установка прав на скрипт
echo ""
echo "🔧 Настройка скрипта..."
chmod +x "$SCRIPT_DIR/cleanup_old_updates.sh"
echo "   ✅ Скрипт готов к запуску"

# Перезагрузка systemd
echo ""
echo "🔄 Перезагрузка systemd..."
systemctl daemon-reload
echo "   ✅ systemd перезагружен"

# Включение таймера
echo ""
echo "⚙️  Включение таймера..."
systemctl enable "$TIMER_NAME"
echo "   ✅ Таймер включен"

# Запуск таймера
echo ""
echo "▶️  Запуск таймера..."
systemctl start "$TIMER_NAME"
echo "   ✅ Таймер запущен"

# Проверка статуса
echo ""
echo "📊 Статус таймера:"
systemctl status "$TIMER_NAME" --no-pager -l

echo ""
echo "✅ Настройка завершена!"
echo ""
echo "📋 Полезные команды:"
echo "   Статус таймера: sudo systemctl status $TIMER_NAME"
echo "   Список таймеров: sudo systemctl list-timers"
echo "   Запустить вручную: sudo systemctl start $SERVICE_NAME"
echo "   Логи: sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo "⏰ Очистка будет выполняться каждый день в 3:00 ночи"

