#!/bin/bash

# Скрипт проверки конфигурации Cloudinary
# Использование: ./scripts/check_cloudinary_config.sh [путь_к_env_файлу]

ENV_FILE="${1:-env.development}"

echo "🔍 Проверка конфигурации Cloudinary..."
echo "📁 Файл конфигурации: $ENV_FILE"
echo ""

# Проверяем, существует ли файл
if [ ! -f "$ENV_FILE" ]; then
    echo "❌ Файл $ENV_FILE не найден!"
    echo "   Использование: $0 [путь_к_env_файлу]"
    exit 1
fi

# Функция для маскировки строк
mask_string() {
    local str="$1"
    if [ -z "$str" ]; then
        echo "(не настроен)"
    elif [ ${#str} -le 4 ]; then
        echo "***"
    else
        echo "${str:0:4}***${str: -4}"
    fi
}

# Читаем переменные из файла (игнорируем комментарии и пустые строки)
USE_CLOUDINARY=$(grep -E "^USE_CLOUDINARY=" "$ENV_FILE" | cut -d'=' -f2 | sed 's/#.*$//' | tr -d ' ' | head -1)
CLOUDINARY_CLOUD_NAME=$(grep -E "^CLOUDINARY_CLOUD_NAME=" "$ENV_FILE" | cut -d'=' -f2 | sed 's/#.*$//' | tr -d ' ' | head -1)
CLOUDINARY_API_KEY=$(grep -E "^CLOUDINARY_API_KEY=" "$ENV_FILE" | cut -d'=' -f2 | sed 's/#.*$//' | tr -d ' ' | head -1)
CLOUDINARY_API_SECRET=$(grep -E "^CLOUDINARY_API_SECRET=" "$ENV_FILE" | cut -d'=' -f2 | sed 's/#.*$//' | tr -d ' ' | head -1)
CLOUDINARY_UPLOAD_PRESET=$(grep -E "^CLOUDINARY_UPLOAD_PRESET=" "$ENV_FILE" | cut -d'=' -f2 | sed 's/#.*$//' | tr -d ' ' | head -1)
CLOUDINARY_REMOVE_BACKGROUND=$(grep -E "^CLOUDINARY_REMOVE_BACKGROUND=" "$ENV_FILE" | cut -d'=' -f2 | sed 's/#.*$//' | tr -d ' ' | head -1)

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
    echo "   → Проверьте настройки preset в Cloudinary Dashboard:"
    if [ -n "$CLOUDINARY_UPLOAD_PRESET" ]; then
        echo "      https://console.cloudinary.com/settings/upload_presets/$CLOUDINARY_UPLOAD_PRESET"
    else
        echo "      https://console.cloudinary.com/settings/upload_presets"
    fi
else
    echo "ℹ️  Удаление фона отключено"
    echo "   → Фон НЕ будет удаляться"
fi

echo ""
echo "📝 Инструкции:"
echo "   1. Проверьте, что все переменные настроены в $ENV_FILE"
echo "   2. Если удаление фона включено, проверьте настройки Upload Preset:"
echo "      - Откройте Cloudinary Dashboard"
echo "      - Settings → Upload → Upload Presets"
echo "      - Найдите ваш preset и проверьте 'Incoming Transformation'"
echo "      - Должна быть цепочка с e_background_removal:fineedges_y"
echo "   3. Перезапустите сервер после изменения конфигурации"
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

