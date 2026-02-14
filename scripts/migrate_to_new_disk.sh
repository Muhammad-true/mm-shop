#!/bin/bash
# Скрипт для миграции данных на новый диск
# Использование: sudo ./scripts/migrate_to_new_disk.sh /mnt/mm_shop_data

set -e

NEW_DISK_PATH="${1:-/mnt/mm_shop_data}"
OLD_IMAGES_PATH="./images"
OLD_UPDATES_PATH="./updates"
OLD_LIBISS_PATH="./libiss_pos"
OLD_POSTGRES_PATH="./postgres_data"
OLD_REDIS_PATH="./redis_data"

echo "🔄 Миграция данных на новый диск"
echo "📂 Новый путь: $NEW_DISK_PATH"
echo ""

# Проверяем, запущен ли скрипт от root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Ошибка: скрипт должен быть запущен с правами root (sudo)"
    exit 1
fi

# Проверяем существование новой директории
if [ ! -d "$NEW_DISK_PATH" ]; then
    echo "❌ Ошибка: директория $NEW_DISK_PATH не существует"
    echo "💡 Сначала запустите: sudo ./scripts/setup_new_disk.sh $NEW_DISK_PATH"
    exit 1
fi

# Функция для безопасного копирования
migrate_directory() {
    local SOURCE="$1"
    local DEST="$2"
    local NAME="$3"
    
    if [ -d "$SOURCE" ] && [ "$(ls -A $SOURCE 2>/dev/null)" ]; then
        echo "📦 Миграция $NAME..."
        echo "   Из: $SOURCE"
        echo "   В:  $DEST"
        
        # Создаем директорию назначения
        mkdir -p "$DEST"
        
        # Копируем с сохранением прав
        rsync -av --progress "$SOURCE/" "$DEST/"
        
        echo "✅ $NAME мигрирован"
        echo ""
    else
        echo "⏭️  $NAME: исходная директория пуста или не существует, пропускаем"
        echo ""
    fi
}

# Останавливаем контейнеры перед миграцией
echo "🛑 Остановка контейнеров..."
cd "$(dirname "$0")/.."
if [ -f "docker-compose.release.yml" ]; then
    docker-compose -f docker-compose.release.yml down
    echo "✅ Контейнеры остановлены"
    echo ""
fi

# Мигрируем данные
echo "🚀 Начало миграции данных..."
echo ""

# Изображения
migrate_directory "$OLD_IMAGES_PATH" "$NEW_DISK_PATH/images" "Изображения"

# Обновления
migrate_directory "$OLD_UPDATES_PATH" "$NEW_DISK_PATH/updates" "Обновления"

# Libiss POS
migrate_directory "$OLD_LIBISS_PATH" "$NEW_DISK_PATH/libiss_pos" "Libiss POS"

# PostgreSQL (если есть локальная БД)
if [ -d "$OLD_POSTGRES_PATH" ] && [ "$(ls -A $OLD_POSTGRES_PATH 2>/dev/null)" ]; then
    echo "⚠️  ВНИМАНИЕ: Обнаружена локальная БД PostgreSQL"
    echo "   Рекомендуется сделать дамп БД перед миграцией:"
    echo "   docker exec mm-postgres-prod pg_dump -U mm_user mm_shop_prod > backup.sql"
    echo ""
    read -p "Продолжить миграцию БД? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        migrate_directory "$OLD_POSTGRES_PATH" "$NEW_DISK_PATH/postgres_data" "PostgreSQL"
    else
        echo "⏭️  Миграция БД пропущена"
        echo ""
    fi
fi

# Redis (обычно не нужно мигрировать, но на всякий случай)
if [ -d "$OLD_REDIS_PATH" ] && [ "$(ls -A $OLD_REDIS_PATH 2>/dev/null)" ]; then
    migrate_directory "$OLD_REDIS_PATH" "$NEW_DISK_PATH/redis_data" "Redis"
fi

echo "✅ Миграция завершена!"
echo ""
echo "📊 Проверка размеров:"
du -sh "$NEW_DISK_PATH"/* 2>/dev/null || true
echo ""
echo "💡 Следующий шаг:"
echo "   1. Проверьте, что все данные скопированы"
echo "   2. Запустите контейнеры: docker-compose -f docker-compose.release.yml up -d"
echo "   3. Проверьте логи: docker-compose -f docker-compose.release.yml logs -f"

