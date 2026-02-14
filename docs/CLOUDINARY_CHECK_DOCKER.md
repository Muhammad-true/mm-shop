# 🔍 Проверка конфигурации Cloudinary в Docker

## Быстрая проверка

### Вариант 1: Проверка внутри Docker контейнера

```bash
# Проверить переменные окружения в контейнере
docker compose -f docker-compose.release.yml exec api printenv | grep CLOUDINARY
```

### Вариант 2: Использование скрипта внутри контейнера

```bash
# Запустить скрипт проверки внутри контейнера
docker compose -f docker-compose.release.yml exec api bash -c "cd /app && ./scripts/check_cloudinary_config.sh"
```

### Вариант 3: Проверка файла .env.production на сервере

```bash
# Проверить, существует ли файл
ls -la ~/mm-shop/release/.env.production

# Посмотреть содержимое (маскируя секреты)
grep CLOUDINARY ~/mm-shop/release/.env.production | sed 's/=.*/=***/'
```

## Где находятся конфигурационные файлы

### На сервере (production)

1. **Файл `.env.production`** (в директории `~/mm-shop/release/`)
   - Используется Docker Compose (`env_file: - .env.production`)
   - Этот файл обычно НЕ в Git (добавлен в .gitignore)
   - Создается вручную на сервере

2. **Переменные в `docker-compose.release.yml`**
   - Некоторые переменные могут быть заданы напрямую в секции `environment:`

3. **Переменные окружения системы**
   - Могут быть заданы через systemd, systemd service или экспортированы в shell

## Создание/редактирование .env.production

```bash
cd ~/mm-shop/release

# Создать файл (если его нет)
nano .env.production

# Добавить переменные Cloudinary:
USE_CLOUDINARY=true
CLOUDINARY_CLOUD_NAME=ваш_cloud_name
CLOUDINARY_API_KEY=ваш_api_key
CLOUDINARY_API_SECRET=ваш_api_secret
CLOUDINARY_UPLOAD_PRESET=ваш_preset_name
CLOUDINARY_REMOVE_BACKGROUND=false

# Сохранить (Ctrl+O, Enter, Ctrl+X в nano)
```

## Проверка после изменения

После изменения `.env.production` нужно перезапустить контейнер:

```bash
cd ~/mm-shop/release
docker compose -f docker-compose.release.yml restart api
```

Или полный перезапуск:

```bash
docker compose -f docker-compose.release.yml down
docker compose -f docker-compose.release.yml up -d
```

## Проверка логов

После перезапуска проверьте логи, чтобы убедиться, что Cloudinary используется:

```bash
# Смотреть логи в реальном времени
docker compose -f docker-compose.release.yml logs -f api

# Искать упоминания Cloudinary
docker compose -f docker-compose.release.yml logs api | grep -i cloudinary
```

При загрузке изображения вы должны увидеть:
```
☁️ Обработка изображения товара через Cloudinary...
   ✅ Cloudinary включен
   ☁️  Cloud Name: ваш_cloud_name
   ⚙️  Upload Preset: ваш_preset
   🎨 Remove Background: true/false
```

## Устранение проблем

### Проблема: Файл .env.production не найден

**Решение:**
```bash
cd ~/mm-shop/release
touch .env.production
nano .env.production
# Добавьте переменные Cloudinary
```

### Проблема: Переменные не загружаются в контейнер

**Проверка:**
```bash
# Проверить, что файл указан в docker-compose.release.yml
grep -A 2 "env_file" docker-compose.release.yml

# Должно быть:
#   env_file:
#     - .env.production
```

**Решение:**
1. Убедитесь, что файл `.env.production` находится в той же директории, что и `docker-compose.release.yml`
2. Перезапустите контейнер: `docker compose -f docker-compose.release.yml restart api`

### Проблема: Переменные есть, но Cloudinary не используется

**Проверка логов:**
```bash
docker compose -f docker-compose.release.yml logs api | grep -i "cloudinary\|локально"
```

Если видите "Обработка изображения товара локально" → проверьте:
1. `USE_CLOUDINARY=true` в `.env.production`
2. Все обязательные переменные заполнены
3. Контейнер перезапущен после изменения

## Полезные команды

```bash
# Проверить все переменные окружения в контейнере
docker compose -f docker-compose.release.yml exec api printenv

# Проверить только Cloudinary переменные
docker compose -f docker-compose.release.yml exec api printenv | grep CLOUDINARY

# Проверить конфигурацию через скрипт
docker compose -f docker-compose.release.yml exec api bash -c "cd /app && ./scripts/check_cloudinary_config.sh"

# Посмотреть, какой файл используется
docker compose -f docker-compose.release.yml config | grep -A 5 "env_file"
```

