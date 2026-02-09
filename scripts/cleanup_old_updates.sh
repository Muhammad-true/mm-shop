#!/bin/bash

# Скрипт для очистки старых обновлений
# Удаляет файлы обновлений старше указанного количества дней
# Использование: ./cleanup_old_updates.sh [дней] [путь_к_папке]

set -e

# Параметры по умолчанию
DAYS_OLD="${1:-10}"  # По умолчанию 10 дней
UPDATES_DIR="${2:-./updates}"  # По умолчанию ./updates

echo "🧹 Очистка старых обновлений"
echo "   Папка: $UPDATES_DIR"
echo "   Удаляем файлы старше: $DAYS_OLD дней"
echo ""

# Проверка существования папки
if [ ! -d "$UPDATES_DIR" ]; then
    echo "❌ Папка не найдена: $UPDATES_DIR"
    exit 1
fi

# Счетчики
total_size=0
deleted_count=0

# Функция для удаления файлов в папке
cleanup_folder() {
    local folder="$1"
    local folder_name=$(basename "$folder")
    
    if [ ! -d "$folder" ]; then
        return
    fi
    
    echo "📂 Обработка папки: $folder_name"
    
    # Находим и удаляем файлы старше указанного количества дней
    while IFS= read -r -d '' file; do
        if [ -f "$file" ]; then
            file_size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
            file_name=$(basename "$file")
            
            echo "   🗑️  Удаление: $file_name ($(numfmt --to=iec-i --suffix=B $file_size 2>/dev/null || echo "${file_size} bytes"))"
            
            rm -f "$file"
            total_size=$((total_size + file_size))
            deleted_count=$((deleted_count + 1))
        fi
    done < <(find "$folder" -type f -mtime +$DAYS_OLD -print0 2>/dev/null)
}

# Обрабатываем все подпапки (android, windows, server)
for platform_dir in "$UPDATES_DIR"/*; do
    if [ -d "$platform_dir" ]; then
        cleanup_folder "$platform_dir"
    fi
done

# Выводим результат
echo ""
if [ $deleted_count -eq 0 ]; then
    echo "✅ Старых файлов не найдено"
else
    size_mb=$(echo "scale=2; $total_size / 1024 / 1024" | bc 2>/dev/null || echo "0")
    echo "✅ Очистка завершена:"
    echo "   Удалено файлов: $deleted_count"
    echo "   Освобождено места: ${size_mb} MB"
fi

echo ""

