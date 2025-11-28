# 🚀 Инструкция по деплою на сервер

## ✅ Что уже сделано

- ✅ Все изменения закоммичены и запушены в Git
- ✅ Версия 1.2.0 готова к деплою
- ✅ Файлы для production созданы

---

## 📋 На сервере выполните:

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
docker-compose -f docker-compose.release.yml up -d --build api
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
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/version
```

### 4. Проверить админку

Откройте в браузере: `http://159.89.99.252:3000`

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

