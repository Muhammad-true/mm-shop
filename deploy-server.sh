#!/bin/bash

# MM Shop - Server Deployment Script
# ==================================
# Скрипт для деплоя на сервер 159.89.99.252

set -e

echo "🚀 Начинаем деплой MM Shop на сервер..."

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ОШИБКА]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[УСПЕХ]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[ПРЕДУПРЕЖДЕНИЕ]${NC} $1"
}

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    error "Docker не установлен!"
fi

if ! command -v docker-compose &> /dev/null; then
    error "Docker Compose не установлен!"
fi

# Останавливаем старые контейнеры
log "🛑 Останавливаем старые контейнеры..."
docker-compose -f docker-compose.release.yml down --remove-orphans || true

# Удаляем старые образы
log "🗑️ Удаляем старые образы..."
docker image prune -f

# Создаем .env файл для продакшена
log "📝 Создаем .env файл для продакшена..."
cat > .env << EOF
# Production Environment Variables
POSTGRES_PASSWORD=muhammadjon
PGADMIN_EMAIL=admin@mm-api.com
PGADMIN_PASSWORD=admin123
CORS_ALLOWED_ORIGINS=http://159.89.99.252,http://localhost
EOF

# Собираем образы
log "🔨 Собираем образ API..."
docker build -f Dockerfile.api.release -t mm-api-prod .

log "🔨 Собираем образ админки..."
docker build -f Dockerfile.admin.release -t mm-admin-prod .

# Запускаем сервисы
log "🚀 Запускаем сервисы..."
docker-compose -f docker-compose.release.yml up -d

# Ждем запуска
log "⏳ Ждем запуска сервисов..."
sleep 30

# Проверяем статус
log "🔍 Проверяем статус сервисов..."
docker-compose -f docker-compose.release.yml ps

# Проверяем health checks
log "🏥 Проверяем health checks..."
for service in postgres redis api admin; do
    if docker-compose -f docker-compose.release.yml ps | grep -q "$service.*healthy"; then
        success "✅ $service работает корректно"
    else
        warning "⚠️ $service не прошел health check"
    fi
done

# Проверяем доступность API
log "🌐 Проверяем доступность API..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    success "✅ API доступен на порту 8080"
else
    error "❌ API недоступен на порту 8080"
fi

# Проверяем доступность админки
log "🌐 Проверяем доступность админки..."
if curl -f http://localhost/ > /dev/null 2>&1; then
    success "✅ Админка доступна на порту 80"
else
    error "❌ Админка недоступна на порту 80"
fi

# Показываем информацию о деплое
log "📊 Информация о деплое:"
echo "=================================="
echo "🌐 Админ панель: http://159.89.99.252"
echo "🔌 API: http://159.89.99.252:8080"
echo "🗄️ PgAdmin: http://159.89.99.252:8081"
echo "📊 PostgreSQL: localhost:5432"
echo "🔴 Redis: localhost:6379"
echo "=================================="

success "🎉 Деплой завершен успешно!"
success "Сервисы доступны по указанным адресам"

# Показываем логи для отладки
log "📋 Последние логи API:"
docker-compose -f docker-compose.release.yml logs --tail=10 api

log "📋 Последние логи админки:"
docker-compose -f docker-compose.release.yml logs --tail=10 admin
