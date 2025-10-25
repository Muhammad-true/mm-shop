#!/bin/bash

echo "🔄 Перезапуск сервера MM Shop..."

# Переходим в директорию проекта
cd /root/mm-shop/release

# Останавливаем все контейнеры
echo "⏹️ Останавливаем контейнеры..."
docker compose -f docker-compose.release.yml down

# Очищаем неиспользуемые образы и контейнеры
echo "🧹 Очищаем Docker кэш..."
docker system prune -f

# Получаем последние изменения
echo "📥 Получаем последние изменения из репозитория..."
git pull origin main

# Пересобираем и запускаем контейнеры
echo "🔨 Пересобираем и запускаем контейнеры..."
docker compose -f docker-compose.release.yml up -d --build

# Проверяем статус
echo "✅ Проверяем статус контейнеров..."
docker compose -f docker-compose.release.yml ps

echo "🎉 Перезапуск завершен!"
