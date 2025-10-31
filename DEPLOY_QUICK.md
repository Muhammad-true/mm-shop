# 🚀 Быстрый деплой

## На сервере выполните:

```bash
cd /root/mm-shop/release

# Если есть локальные изменения - откатываем их
git checkout .

# Обновляем код
git pull origin main

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

**1.2.6** - Fixed dashboard array checks and image URLs

