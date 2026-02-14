# ⚡ Быстрая проверка Cloudinary на сервере

## Шаг 1: Проверка файла .env.production

Файл `.env.production` уже есть на сервере в `/root/mm-shop/release/`.

### Проверка содержимого через SSH:

```bash
# Подключитесь к серверу
ssh root@159.89.99.252

# Проверьте содержимое файла (маскируя секреты)
cd ~/mm-shop/release
grep CLOUDINARY .env.production | sed 's/=.*/=***/'
```

### Или через FileZilla:

1. Откройте файл `.env.production` на сервере (правый клик → View/Edit)
2. Проверьте, что есть все переменные:
   ```
   USE_CLOUDINARY=true
   CLOUDINARY_CLOUD_NAME=ваш_cloud_name
   CLOUDINARY_API_KEY=ваш_api_key
   CLOUDINARY_API_SECRET=ваш_api_secret
   CLOUDINARY_UPLOAD_PRESET=ваш_preset_name
   CLOUDINARY_REMOVE_BACKGROUND=false
   ```

## Шаг 2: Проверка переменных в Docker контейнере

```bash
# Проверить переменные внутри контейнера
docker compose -f docker-compose.release.yml exec api printenv | grep CLOUDINARY
```

**Ожидаемый результат:**
```
USE_CLOUDINARY=true
CLOUDINARY_CLOUD_NAME=ваш_cloud_name
CLOUDINARY_API_KEY=ваш_api_key
CLOUDINARY_API_SECRET=ваш_api_secret
CLOUDINARY_UPLOAD_PRESET=ваш_preset_name
CLOUDINARY_REMOVE_BACKGROUND=false
```

## Шаг 3: Если переменных нет в контейнере

### Перезапустите контейнер:

```bash
cd ~/mm-shop/release
docker compose -f docker-compose.release.yml restart api
```

### Или полный перезапуск:

```bash
docker compose -f docker-compose.release.yml down
docker compose -f docker-compose.release.yml up -d
```

## Шаг 4: Проверка логов

После перезапуска проверьте логи:

```bash
# Смотреть логи в реальном времени
docker compose -f docker-compose.release.yml logs -f api

# Или проверьте последние логи
docker compose -f docker-compose.release.yml logs api | grep -i cloudinary
```

**При загрузке изображения вы должны увидеть:**
```
☁️ Обработка изображения товара через Cloudinary...
   ✅ Cloudinary включен
   ☁️  Cloud Name: ваш_cloud_name
   ⚙️  Upload Preset: ваш_preset
   🎨 Remove Background: true/false
```

## Шаг 5: Использование скрипта проверки

```bash
# Запустить скрипт проверки внутри контейнера
docker compose -f docker-compose.release.yml exec api sh /app/scripts/check_cloudinary_docker.sh
```

## Если что-то не работает

1. **Проверьте, что файл `.env.production` в правильной директории:**
   ```bash
   ls -la ~/mm-shop/release/.env.production
   ```

2. **Проверьте, что Docker Compose использует этот файл:**
   ```bash
   grep -A 2 "env_file" ~/mm-shop/release/docker-compose.release.yml
   ```
   
   Должно быть:
   ```yaml
   env_file:
     - .env.production
   ```

3. **Проверьте права доступа к файлу:**
   ```bash
   chmod 644 ~/mm-shop/release/.env.production
   ```

4. **Убедитесь, что файл не пустой:**
   ```bash
   wc -l ~/mm-shop/release/.env.production
   ```

## Быстрая команда для проверки всего сразу

```bash
cd ~/mm-shop/release && \
echo "=== Файл .env.production ===" && \
ls -la .env.production && \
echo "" && \
echo "=== Переменные в контейнере ===" && \
docker compose -f docker-compose.release.yml exec api printenv | grep CLOUDINARY && \
echo "" && \
echo "=== Последние логи ===" && \
docker compose -f docker-compose.release.yml logs api --tail=20 | grep -i cloudinary
```

