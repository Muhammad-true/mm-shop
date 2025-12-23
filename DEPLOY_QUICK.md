# 🚀 Быстрый деплой

## На сервере выполните:

```bash
cd /root/mm-shop/release

# Если есть локальные изменения - откатываем их
git checkout .

# Обновляем код
git pull origin main

# Создайте .env.production файл (если еще не создан) с переменными окружения:
# LEMONSQUEEZY_API_KEY=your-api-key
# LEMONSQUEEZY_STORE_ID=your-store-id
# JWT_SECRET=your-jwt-secret
# и другие необходимые переменные

# ОСТАНОВКА и удаление контейнеров для чистого билда
docker compose -f docker-compose.release.yml stop api admin
docker compose -f docker-compose.release.yml rm -f api admin

# Удаляем старые образы
docker rmi release-api release-admin 2>/dev/null || true

# ПЕРЕСБОРКА без кэша и запуск
docker compose -f docker-compose.release.yml build --no-cache api admin
docker compose -f docker-compose.release.yml up -d api admin
```

## Проверка:

```bash
docker ps
docker logs mm-api-prod --tail 50 -f
```

## Версия:

**1.2.8** - Admin panel now served directly from API on port 8080

