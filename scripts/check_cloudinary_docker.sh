#!/bin/sh

# Скрипт для проверки Cloudinary конфигурации внутри Docker контейнера
# Использование: docker compose -f docker-compose.release.yml exec api sh /app/scripts/check_cloudinary_docker.sh

echo "🔍 Проверка конфигурации Cloudinary в Docker контейнере..."
echo ""

# Функция для маскировки строк
mask_string() {
    str="$1"
    if [ -z "$str" ]; then
        echo "(не настроен)"
    elif [ ${#str} -le 4 ]; then
        echo "***"
    else
        prefix=$(echo "$str" | cut -c1-4)
        suffix=$(echo "$str" | rev | cut -c1-4 | rev)
        echo "${prefix}***${suffix}"
    fi
}

# Читаем переменные окружения
USE_CLOUDINARY="${USE_CLOUDINARY:-}"
CLOUDINARY_CLOUD_NAME="${CLOUDINARY_CLOUD_NAME:-}"
CLOUDINARY_API_KEY="${CLOUDINARY_API_KEY:-}"
CLOUDINARY_API_SECRET="${CLOUDINARY_API_SECRET:-}"
CLOUDINARY_UPLOAD_PRESET="${CLOUDINARY_UPLOAD_PRESET:-}"
CLOUDINARY_REMOVE_BACKGROUND="${CLOUDINARY_REMOVE_BACKGROUND:-}"

# Нормализуем булевы значения
if [ "$USE_CLOUDINARY" = "true" ] || [ "$USE_CLOUDINARY" = "1" ]; then
    USE_CLOUDINARY="true"
else
    USE_CLOUDINARY="false"
fi

if [ "$CLOUDINARY_REMOVE_BACKGROUND" = "true" ] || [ "$CLOUDINARY_REMOVE_BACKGROUND" = "1" ]; then
    CLOUDINARY_REMOVE_BACKGROUND="true"
else
    CLOUDINARY_REMOVE_BACKGROUND="false"
fi

echo "📋 Текущая конфигурация:"
echo "   USE_CLOUDINARY: $USE_CLOUDINARY"
echo "   CLOUDINARY_CLOUD_NAME: $(mask_string "$CLOUDINARY_CLOUD_NAME")"
echo "   CLOUDINARY_API_KEY: $(mask_string "$CLOUDINARY_API_KEY")"
echo "   CLOUDINARY_API_SECRET: $(mask_string "$CLOUDINARY_API_SECRET")"
echo "   CLOUDINARY_UPLOAD_PRESET: $CLOUDINARY_UPLOAD_PRESET"
echo "   CLOUDINARY_REMOVE_BACKGROUND: $CLOUDINARY_REMOVE_BACKGROUND"
echo ""

# Проверка
ALL_OK=true

if [ "$USE_CLOUDINARY" != "true" ]; then
    echo "❌ Cloudinary отключен (USE_CLOUDINARY=false)"
    echo "   → Изображения будут обрабатываться локально"
    ALL_OK=false
else
    echo "✅ Cloudinary включен"
fi

if [ -z "$CLOUDINARY_CLOUD_NAME" ]; then
    echo "❌ CLOUDINARY_CLOUD_NAME не настроен"
    ALL_OK=false
else
    echo "✅ Cloud Name: $CLOUDINARY_CLOUD_NAME"
fi

if [ -z "$CLOUDINARY_API_KEY" ]; then
    echo "❌ CLOUDINARY_API_KEY не настроен"
    ALL_OK=false
else
    echo "✅ API Key настроен"
fi

if [ -z "$CLOUDINARY_API_SECRET" ]; then
    echo "❌ CLOUDINARY_API_SECRET не настроен"
    ALL_OK=false
else
    echo "✅ API Secret настроен"
fi

if [ -z "$CLOUDINARY_UPLOAD_PRESET" ]; then
    echo "❌ CLOUDINARY_UPLOAD_PRESET не настроен"
    ALL_OK=false
else
    echo "✅ Upload Preset: $CLOUDINARY_UPLOAD_PRESET"
fi

if [ "$CLOUDINARY_REMOVE_BACKGROUND" = "true" ]; then
    echo "✅ Удаление фона включено"
    echo "   ⚠️  ВАЖНО: Убедитесь, что в Upload Preset настроена трансформация:"
    echo "      e_background_removal:fineedges_y"
    if [ -n "$CLOUDINARY_UPLOAD_PRESET" ]; then
        echo "   → https://console.cloudinary.com/settings/upload_presets/$CLOUDINARY_UPLOAD_PRESET"
    fi
else
    echo "ℹ️  Удаление фона отключено"
    echo "   → Фон НЕ будет удаляться"
fi

echo ""
echo "📝 Инструкции:"
echo "   1. Создайте файл .env.production в директории release/ на сервере"
echo "   2. Добавьте переменные Cloudinary в файл"
echo "   3. Перезапустите контейнер: docker compose -f docker-compose.release.yml restart api"
echo ""

if [ "$ALL_OK" = true ] && [ "$USE_CLOUDINARY" = "true" ]; then
    echo "✅ Конфигурация Cloudinary корректна!"
    echo "   → Изображения будут загружаться в Cloudinary"
    if [ "$CLOUDINARY_REMOVE_BACKGROUND" = "true" ]; then
        echo "   → Удаление фона включено (проверьте preset!)"
    fi
    exit 0
else
    echo "⚠️  Есть проблемы с конфигурацией"
    echo "   → Исправьте ошибки выше и перезапустите сервер"
    exit 1
fi

