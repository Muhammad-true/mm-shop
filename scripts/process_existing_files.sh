#!/bin/bash

# Скрипт для обработки уже загруженных файлов из папки FTP
# Использование: ./process_existing_files.sh

# Конфигурация
FTP_WATCH_DIR="${FTP_WATCH_DIR:-/var/ftp/uploads}"
API_BASE_URL="${API_BASE_URL:-https://api.libiss.com/api/v1}"
API_TOKEN="${API_TOKEN:-}"
PROCESSED_DIR="${PROCESSED_DIR:-/var/ftp/processed}"

# Проверка обязательных параметров
if [ -z "$API_TOKEN" ]; then
    echo "❌ Ошибка: API_TOKEN не установлен"
    echo "   Установите переменную окружения: export API_TOKEN=your_token"
    echo "   Или используйте скрипт: ./get_api_token.sh"
    exit 1
fi

# Создание папок
mkdir -p "$FTP_WATCH_DIR"
mkdir -p "$PROCESSED_DIR"

echo "🔍 Поиск файлов в папке: $FTP_WATCH_DIR"
echo "   API URL: $API_BASE_URL"
echo ""

# Функция для парсинга имени файла
parse_filename() {
    local filename="$1"
    local basename=$(basename "$filename")
    local ext="${basename##*.}"
    local name="${basename%.*}"

    # Определяем платформу по расширению
    case "$ext" in
        apk) platform="android" ;;
        exe) platform="windows" ;;
        zip) platform="server" ;;
        *) return 1 ;;
    esac

    # Извлекаем версию (формат: platform_version или platform-version)
    version=$(echo "$name" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    
    if [ -z "$version" ]; then
        # Пытаемся извлечь из стандартного формата
        version=$(echo "$name" | sed -n "s/.*${platform}[_-]\([0-9]\+\.[0-9]\+\.[0-9]\+\).*/\1/p")
    fi

    if [ -z "$version" ]; then
        return 1
    fi

    echo "$platform|$version"
}

# Функция для обработки файла
process_file() {
    local filepath="$1"
    local filename=$(basename "$filepath")

    echo "🔍 Обработка файла: $filename"

    # Парсим имя файла
    local info=$(parse_filename "$filename")
    if [ $? -ne 0 ]; then
        echo "❌ Не удалось определить платформу и версию из имени файла: $filename"
        echo "   Ожидаемый формат: android_1.0.0.apk, windows_1.2.0.exe, server_2.0.0.zip"
        return 1
    fi

    local platform=$(echo "$info" | cut -d'|' -f1)
    local version=$(echo "$info" | cut -d'|' -f2)

    echo "   Платформа: $platform, Версия: $version"

    # Загружаем через API
    echo "📤 Загрузка файла на сервер..."
    response=$(curl -s -w "\n%{http_code}" -X POST \
        "$API_BASE_URL/admin/updates/upload" \
        -H "Authorization: Bearer $API_TOKEN" \
        -F "platform=$platform" \
        -F "version=$version" \
        -F "releaseNotes=Автоматическая загрузка через FTP: $filename" \
        -F "file=@$filepath" \
        --max-time 1800)

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" != "201" ]; then
        echo "❌ Ошибка загрузки (код $http_code): $body"
        return 1
    fi

    echo "✅ Файл успешно загружен"

    # Перемещаем в папку обработанных
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local processed_file="$PROCESSED_DIR/${filename%.*}_${timestamp}.${filename##*.}"
    mv "$filepath" "$processed_file"
    echo "📦 Файл перемещен в обработанные: $(basename "$processed_file")"
    echo ""

    return 0
}

# Обрабатываем все файлы
processed=0
total=0

# Используем find вместо glob для совместимости
while IFS= read -r -d '' file; do
    if [ -f "$file" ]; then
        ((total++))
        if process_file "$file"; then
            ((processed++))
        fi
    fi
done < <(find "$FTP_WATCH_DIR" -maxdepth 1 -type f \( -name "*.apk" -o -name "*.exe" -o -name "*.zip" \) -print0 2>/dev/null)

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $total -eq 0 ]; then
    echo "ℹ️  Файлы не найдены в папке: $FTP_WATCH_DIR"
else
    echo "✅ Обработка завершена:"
    echo "   Найдено файлов: $total"
    echo "   Успешно обработано: $processed"
    if [ $processed -lt $total ]; then
        echo "   Ошибок: $((total - processed))"
    fi
fi
echo ""

