# 🚀 Быстрый деплой

## На сервере выполните:

```bash
cd /root/mm-shop/release

# Если есть локальные изменения - откатываем их
git checkout .

# Обновляем код
git pull origin main

# Пересобираем контейнеры
docker compose -f docker-compose.release.yml up -d --build api admin
```

## Проверка:

```bash
docker ps
docker logs mm-api-prod --tail 50 -f
```

## Версия:

**1.2.3** - Fixed image URLs to use relative paths

