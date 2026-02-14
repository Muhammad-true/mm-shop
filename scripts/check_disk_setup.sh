#!/bin/bash
# Скрипт для проверки настройки диска
# Использование: sudo ./scripts/check_disk_setup.sh

set -e

DISK_PATH="/mnt/mm_shop_data"

echo "🔍 Проверка настройки диска для MM Shop"
echo "📂 Путь: $DISK_PATH"
echo ""

# Проверяем, смонтирован ли диск
if mountpoint -q "$DISK_PATH"; then
    echo "✅ Диск смонтирован в $DISK_PATH"
else
    echo "❌ Диск НЕ смонтирован в $DISK_PATH"
    echo "💡 Выполните: sudo mount /dev/sda /mnt/mm_shop_data"
    exit 1
fi

# Проверяем размер
echo ""
echo "📊 Информация о диске:"
df -h "$DISK_PATH"

# Проверяем, есть ли в /etc/fstab
echo ""
echo "📝 Проверка /etc/fstab:"
if grep -q "$DISK_PATH" /etc/fstab; then
    echo "✅ Диск добавлен в /etc/fstab (будет монтироваться автоматически)"
    grep "$DISK_PATH" /etc/fstab
else
    echo "⚠️  Диск НЕ добавлен в /etc/fstab"
    echo "💡 Добавьте строку в /etc/fstab для автоматического монтирования:"
    UUID=$(blkid -s UUID -o value /dev/sda)
    echo "   UUID=$UUID $DISK_PATH ext4 defaults,nofail,discard 0 2"
fi

# Проверяем структуру директорий
echo ""
echo "📁 Проверка структуры директорий:"
REQUIRED_DIRS=(
    "postgres_data"
    "redis_data"
    "images"
    "images/variations"
    "images/products"
    "images/uploads"
    "images/categories"
    "images/avatars"
    "images/shop_logos"
    "logs"
    "updates"
    "libiss_pos"
)

MISSING_DIRS=()
for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$DISK_PATH/$dir" ]; then
        echo "✅ $dir"
    else
        echo "❌ $dir - отсутствует"
        MISSING_DIRS+=("$dir")
    fi
done

if [ ${#MISSING_DIRS[@]} -gt 0 ]; then
    echo ""
    echo "⚠️  Отсутствуют директории. Запустите скрипт настройки:"
    echo "   sudo ./scripts/setup_new_disk.sh $DISK_PATH"
else
    echo ""
    echo "✅ Все необходимые директории созданы"
fi

# Проверяем права доступа
echo ""
echo "🔐 Проверка прав доступа:"
ls -ld "$DISK_PATH" | awk '{print "   " $1 " " $3 " " $4 " " $9}'

# Проверяем использование места
echo ""
echo "💾 Использование места:"
if [ "$(ls -A $DISK_PATH 2>/dev/null)" ]; then
    du -sh "$DISK_PATH"/* 2>/dev/null | head -10
else
    echo "   Диск пуст (готов к использованию)"
fi

echo ""
echo "✅ Проверка завершена!"

