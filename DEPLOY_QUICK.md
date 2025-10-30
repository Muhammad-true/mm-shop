# 🚀 Быстрый деплой

## На сервере выполните:

```bash
cd /root/mm-shop
git pull origin main
docker-compose -f docker-compose.release.yml up -d --build api
docker-compose -f docker-compose.release.yml up -d --build admin
```

## Проверка:

```bash
docker ps
docker logs mm-api-prod --tail 50 -f
```

## Версия:

**1.2.0** - Categories icons, subcategories, enhanced variations

