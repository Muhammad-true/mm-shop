# 🔍 Проверка .env.production и пересборка контейнеров

## Шаг 1: Проверка содержимого .env.production

### На сервере через SSH:

```bash
# Посмотреть содержимое файла (маскируя секреты)
cd ~/mm-shop/release
cat .env.production | sed 's/=.*/=***/' | grep -E "^[A-Z]"

# Или посмотреть только Cloudinary переменные
grep CLOUDINARY .env.production | sed 's/=.*/=***/'

# Или посмотреть весь файл (будьте осторожны - покажет секреты!)
cat .env.production
```

### Через FileZilla:

1. Подключитесь к серверу через SFTP
2. Перейдите в `/root/mm-shop/release/`
3. Правый клик на `.env.production` → View/Edit
4. Проверьте содержимое

## Шаг 2: Проверка переменных в Docker контейнере

```bash
# Проверить переменные внутри контейнера
docker compose -f docker-compose.release.yml exec api printenv | grep CLOUDINARY

# Или все переменные окружения
docker compose -f docker-compose.release.yml exec api printenv | sort
```

## Шаг 3: Пересборка контейнеров

### Вариант 1: Пересборка с кэшем (быстрее)

```bash
cd ~/mm-shop/release

# Остановить контейнеры
docker compose -f docker-compose.release.yml down

# Пересобрать и запустить
docker compose -f docker-compose.release.yml up -d --build
```

### Вариант 2: Пересборка без кэша (полная пересборка)

```bash
cd ~/mm-shop/release

# Остановить контейнеры
docker compose -f docker-compose.release.yml down

# Пересобрать БЕЗ кэша
docker compose -f docker-compose.release.yml build --no-cache api admin

# Запустить
docker compose -f docker-compose.release.yml up -d
```

### Вариант 3: Только перезапуск (если код не менялся)

```bash
cd ~/mm-shop/release

# Просто перезапустить контейнеры (загрузит новые переменные из .env.production)
docker compose -f docker-compose.release.yml restart api
```

## Шаг 4: Проверка после пересборки

```bash
# Проверить статус контейнеров
docker compose -f docker-compose.release.yml ps

# Проверить логи
docker compose -f docker-compose.release.yml logs -f api

# Проверить переменные Cloudinary в контейнере
docker compose -f docker-compose.release.yml exec api printenv | grep CLOUDINARY

# Использовать скрипт проверки
docker compose -f docker-compose.release.yml exec api sh /app/scripts/check_cloudinary_docker.sh
```

## Шаг 5: Проверка работы Cloudinary

После пересборки загрузите тестовое изображение через админ-панель и проверьте логи:

```bash
# Смотреть логи в реальном времени
docker compose -f docker-compose.release.yml logs -f api
```

**Ожидаемый вывод при загрузке изображения:**
```
☁️ Обработка изображения товара через Cloudinary...
   ✅ Cloudinary включен
   ☁️  Cloud Name: ваш_cloud_name
   ⚙️  Upload Preset: ваш_preset
   🎨 Remove Background: true/false
```

## Быстрая команда для всего сразу

```bash
cd ~/mm-shop/release && \
echo "=== Проверка .env.production ===" && \
grep CLOUDINARY .env.production | sed 's/=.*/=***/' && \
echo "" && \
echo "=== Пересборка контейнеров ===" && \
docker compose -f docker-compose.release.yml down && \
docker compose -f docker-compose.release.yml up -d --build && \
echo "" && \
echo "=== Ожидание запуска (10 секунд) ===" && \
sleep 10 && \
echo "=== Проверка переменных в контейнере ===" && \
docker compose -f docker-compose.release.yml exec api printenv | grep CLOUDINARY
```

## Если переменные не загрузились

1. **Проверьте, что файл `.env.production` в правильной директории:**
   ```bash
   ls -la ~/mm-shop/release/.env.production
   ```

2. **Проверьте, что Docker Compose использует этот файл:**
   ```bash
   grep -A 2 "env_file" ~/mm-shop/release/docker-compose.release.yml
   ```

3. **Проверьте формат файла (без пробелов вокруг =):**
   ```bash
   cat .env.production | grep CLOUDINARY
   ```
   
   Правильный формат:
   ```
   USE_CLOUDINARY=true
   CLOUDINARY_CLOUD_NAME=your_cloud_name
   ```
   
   Неправильный формат:
   ```
   USE_CLOUDINARY = true  # Пробелы вокруг =
   ```

4. **Перезапустите контейнер:**
   ```bash
   docker compose -f docker-compose.release.yml restart api
   ```

