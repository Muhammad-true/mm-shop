# 🚀 Инструкция по деплою на сервер

## ✅ Что уже сделано

- ✅ Все изменения закоммичены и запушены в Git
- ✅ Версия 1.2.0 готова к деплою
- ✅ Файлы для production созданы

---

## 📋 На сервере выполните:

### ⚠️ ВАЖНО: Перед деплоем настройте SSL!

**Если SSL еще не настроен:**

```bash
cd /root/mm-shop
chmod +x setup-ssl.sh
./setup-ssl.sh
```

### Вариант 1: Пересобрать все одной командой

```bash
cd /root/mm-shop
git pull origin main
docker-compose -f docker-compose.release.yml up -d --build
```

### Вариант 2: Пересобрать по отдельности (рекомендуется)

```bash
cd /root/mm-shop
git pull origin main
docker compose -f docker-compose.release.yml up -d --build api
docker-compose -f docker-compose.release.yml up -d --build admin
```

### Вариант 3: С полной очисткой (если что-то не работает)

```bash
cd /root/mm-shop
git pull origin main
docker compose -f docker-compose.release.yml down
docker compose -f docker-compose.release.yml build --no-cache
docker compose -f docker-compose.release.yml up -d
```

---

## 💾 Миграция на новый диск

Если нужно перенести данные на новый диск с большим объемом места:

📖 **Подробная инструкция:** [DISK_MIGRATION.md](./DISK_MIGRATION.md)

**Быстрый старт:**

```bash
# 1. Подготовка нового диска
sudo ./scripts/setup_new_disk.sh /mnt/mm_shop_data

# 2. Миграция данных (создайте резервную копию БД перед этим!)
sudo ./scripts/migrate_to_new_disk.sh /mnt/mm_shop_data

# 3. Запуск контейнеров
docker-compose -f docker-compose.release.yml up -d
```

**Важно:** Пути в `docker-compose.release.yml` уже настроены на `/mnt/mm_shop_data/`

---

## 🔒 Настройка HTTPS (ОБЯЗАТЕЛЬНО!)

### Проблема
Фронтенд работает на HTTPS, а API на HTTP, поэтому браузер блокирует запросы (mixed content).

### Решение: Настроить SSL сертификат

#### Вариант 1: Let's Encrypt (рекомендуется, бесплатный)

```bash
# 1. Установить certbot
sudo apt-get update
sudo apt-get install certbot

# 2. Получить сертификат (замените your-domain.com на ваш домен)
sudo certbot certonly --standalone -d your-domain.com

# 3. Сертификаты будут в:
# /etc/letsencrypt/live/your-domain.com/fullchain.pem
# /etc/letsencrypt/live/your-domain.com/privkey.pem

# 4. Обновить docker-compose.release.yml:
# Заменить volume:
#   - ./ssl:/etc/nginx/ssl:ro
# На:
#   - /etc/letsencrypt:/etc/letsencrypt:ro

# 5. Обновить nginx.production.conf:
# Заменить пути к сертификатам:
#   ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
#   ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
```

#### Вариант 2: Самоподписанный сертификат (для тестирования)

**Быстрый способ (используя скрипт):**

```bash
# Сделать скрипт исполняемым
chmod +x setup-ssl.sh

# Запустить скрипт (по умолчанию для 159.89.99.252)
./setup-ssl.sh

# Или указать свой домен/IP
./setup-ssl.sh your-domain.com
```

**Ручной способ:**

```bash
# 1. Создать директорию для сертификатов
mkdir -p ssl

# 2. Сгенерировать самоподписанный сертификат
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/key.pem \
  -out ssl/cert.pem \
  -subj "/C=RU/ST=State/L=City/O=Organization/CN=159.89.99.252"

# 3. Установить права
chmod 600 ssl/key.pem
chmod 644 ssl/cert.pem
```

**⚠️ Внимание:** Самоподписанный сертификат будет показывать предупреждение в браузере. Для продакшена используйте Let's Encrypt!

---

## 🔍 Проверка после деплоя

### 1. Проверить контейнеры

```bash
docker ps
```

Должны быть запущены:
- `mm-postgres-prod`
- `mm-redis-prod`
- `mm-api-prod`
- `mm-admin-prod`

### 2. Проверить логи

```bash
# Логи API
docker logs mm-api-prod --tail 50 -f

# Логи Admin
docker logs mm-admin-prod --tail 50 -f
```

### 3. Проверить API

```bash
# HTTP (должен редиректить на HTTPS)
curl -I http://localhost/health

# HTTPS
curl https://localhost/health
curl https://localhost/api/v1/version
```

### 4. Проверить админку

Откройте в браузере: `https://159.89.99.252` (или ваш домен)

**Важно:** Используйте HTTPS, не HTTP!

Нажмите **Ctrl+Shift+R** для очистки кэша!

---

## 📦 Что изменилось в версии 1.2.0

### Новые возможности:
- ✅ PNG иконки для категорий
- ✅ Подкатегории (многоуровневая иерархия)
- ✅ 3 типа размеров (Одежда, Обувь, Штаны)
- ✅ 16 цветов
- ✅ Исправлен дашборд
- ✅ Исправлены фильтры товаров
- ✅ Добавлены фильтры заказов

### Технические улучшения:
- ✅ Cache busting для админки (обновление версии)
- ✅ Модульная структура JS
- ✅ Исправлены все баги

---

## 🐛 Если что-то не работает

### Ошибка: "Cannot connect to database"

```bash
# Перезапустить PostgreSQL
docker-compose -f docker-compose.release.yml restart postgres
sleep 10
docker-compose -f docker-compose.release.yml restart api
```

### Ошибка SSL: "SSL certificate not found"

```bash
# Проверить наличие сертификатов
ls -la ssl/
# Должны быть: cert.pem и key.pem

# Если нет - создать самоподписанный (см. раздел "Настройка HTTPS")
mkdir -p ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/key.pem \
  -out ssl/cert.pem \
  -subj "/C=RU/ST=State/L=City/O=Organization/CN=159.89.99.252"

# Перезапустить админку
docker-compose -f docker-compose.release.yml restart admin
```

### Ошибка: "Mixed Content" в браузере

**Причина:** Фронтенд на HTTPS пытается обратиться к HTTP API.

**Решение:** Убедитесь, что:
1. SSL сертификат настроен (см. раздел "Настройка HTTPS")
2. Nginx слушает на порту 443 (HTTPS)
3. Все запросы идут через HTTPS

```bash
# Проверить, что nginx слушает на 443
docker exec mm-admin-prod netstat -tlnp | grep 443

# Проверить логи nginx
docker logs mm-admin-prod | grep -i ssl
```

### Админка показывает старую версию

```bash
# Полная пересборка админки
docker-compose -f docker-compose.release.yml down admin
docker rmi mm-shop-admin
docker-compose -f docker-compose.release.yml build --no-cache admin
docker-compose -f docker-compose.release.yml up -d admin
```

### Логи показывают ошибки

```bash
# Полная очистка и пересборка
docker-compose -f docker-compose.release.yml down
docker-compose -f docker-compose.release.yml build --no-cache
docker-compose -f docker-compose.release.yml up -d
```

---

## 📊 Версия

**BUILD_VERSION:** `1.2.0-20251030211000`

Проверить в браузере консоль (F12):
```javascript
window.BUILD_VERSION
// Должно быть: "1.2.0-20251030211000"
```

---

**Дата деплоя:** 30 октября 2025  
**Статус:** ✅ Готово к деплою

