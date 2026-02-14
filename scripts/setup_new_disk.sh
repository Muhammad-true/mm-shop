#!/bin/bash
# Скрипт для настройки нового диска для MM Shop
# Использование: sudo ./scripts/setup_new_disk.sh /mnt/mm_shop_data

set -e

DISK_PATH="${1:-/mnt/mm_shop_data}"

echo "🚀 Настройка нового диска для MM Shop"
echo "📂 Путь: $DISK_PATH"
echo ""

# Проверяем, запущен ли скрипт от root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Ошибка: скрипт должен быть запущен с правами root (sudo)"
    exit 1
fi

# Создаем основную директорию
echo "📁 Создание директорий..."
mkdir -p "$DISK_PATH"/{postgres_data,redis_data,images,logs,updates,libiss_pos}

# Создаем поддиректории для изображений
mkdir -p "$DISK_PATH"/images/{variations,products,uploads,categories,avatars,shop_logos}

# Устанавливаем правильные права доступа
echo "🔐 Установка прав доступа..."

# PostgreSQL требует права для пользователя postgres (обычно UID 999)
if id "postgres" &>/dev/null; then
    POSTGRES_UID=$(id -u postgres)
    POSTGRES_GID=$(id -g postgres)
    chown -R $POSTGRES_UID:$POSTGRES_GID "$DISK_PATH/postgres_data"
    echo "✅ Права для PostgreSQL установлены (UID: $POSTGRES_UID, GID: $POSTGRES_GID)"
else
    # Если пользователя postgres нет, используем стандартные права
    chown -R 999:999 "$DISK_PATH/postgres_data"
    echo "✅ Права для PostgreSQL установлены (999:999)"
fi

# Redis требует права для пользователя redis (обычно UID 999)
chown -R 999:999 "$DISK_PATH/redis_data"

# Для остальных директорий - права для текущего пользователя и docker группы
# Docker обычно работает от имени root в контейнере, но лучше установить широкие права
chmod -R 755 "$DISK_PATH/images"
chmod -R 755 "$DISK_PATH/logs"
chmod -R 755 "$DISK_PATH/updates"
chmod -R 755 "$DISK_PATH/libiss_pos"

# Если есть группа docker, добавляем её
if getent group docker > /dev/null 2>&1; then
    chgrp -R docker "$DISK_PATH/images" "$DISK_PATH/logs" "$DISK_PATH/updates" "$DISK_PATH/libiss_pos"
    chmod -R g+w "$DISK_PATH/images" "$DISK_PATH/logs" "$DISK_PATH/updates" "$DISK_PATH/libiss_pos"
    echo "✅ Права для группы docker установлены"
fi

echo ""
echo "✅ Диск настроен успешно!"
echo ""
echo "📊 Структура директорий:"
ls -lah "$DISK_PATH"
echo ""
echo "💡 Следующий шаг: запустите скрипт миграции данных (если нужно)"
echo "   sudo ./scripts/migrate_to_new_disk.sh $DISK_PATH"

