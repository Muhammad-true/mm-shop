# Руководство по пересборке Docker контейнеров

## Правильная последовательность пересборки

### 1. Остановить и удалить текущие контейнеры

```bash
cd ~/mm-shop/release
docker compose -f docker-compose.release.yml down
```

**Примечание:** Используйте `docker compose` (без дефиса) - это встроенный плагин Docker CLI.

Эта команда:
- Останавливает все контейнеры
- Удаляет контейнеры
- Удаляет сети (но сохраняет volumes с данными)

### 2. Пересобрать образы (с очисткой кэша)

```bash
docker compose -f docker-compose.release.yml build --no-cache
```

Или без очистки кэша (быстрее):
```bash
docker compose -f docker-compose.release.yml build
```

### 3. Запустить контейнеры заново

```bash
docker compose -f docker-compose.release.yml up -d
```

### 4. Проверить статус

```bash
docker compose -f docker-compose.release.yml ps
# или
docker ps
```

---

## Быстрый способ (одной командой)

```bash
cd ~/mm-shop/release
docker compose -f docker-compose.release.yml up -d --build --force-recreate
```

Эта команда:
- `--build` - пересобирает образы
- `--force-recreate` - принудительно пересоздает контейнеры
- `-d` - запускает в фоновом режиме

---

## Очистка старых образов

После пересборки старые образы остаются в системе. Для их удаления:

### Удалить неиспользуемые образы:
```bash
docker image prune -a
```

### Удалить все неиспользуемые ресурсы (образы, контейнеры, сети, volumes):
```bash
docker system prune -a --volumes
```

**⚠️ ВНИМАНИЕ:** Последняя команда удалит ВСЕ неиспользуемые volumes, включая данные, если они не примонтированы!

### Безопасная очистка (только образы):
```bash
# Показать неиспользуемые образы
docker images

# Удалить конкретный образ
docker rmi <IMAGE_ID>

# Удалить все образы без тегов
docker image prune
```

---

## Полная пересборка с нуля

Если нужно полностью пересобрать все с нуля:

```bash
cd ~/mm-shop/release

# 1. Остановить и удалить все
docker compose -f docker-compose.release.yml down

# 2. Удалить volumes (ОСТОРОЖНО! Это удалит данные БД!)
# docker compose -f docker-compose.release.yml down -v

# 3. Пересобрать без кэша
docker compose -f docker-compose.release.yml build --no-cache

# 4. Запустить
docker compose -f docker-compose.release.yml up -d

# 5. Проверить логи
docker compose -f docker-compose.release.yml logs -f
```

---

## Проверка после пересборки

### Проверить статус контейнеров:
```bash
docker ps
```

### Проверить логи:
```bash
# Все контейнеры
docker compose -f docker-compose.release.yml logs

# Конкретный контейнер
docker compose -f docker-compose.release.yml logs api
docker compose -f docker-compose.release.yml logs admin

# Следить за логами в реальном времени
docker compose -f docker-compose.release.yml logs -f api
```

### Проверить здоровье контейнеров:
```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
```

---

## Устранение проблем

### Если контейнер не запускается:

1. Проверить логи:
```bash
docker compose -f docker-compose.release.yml logs <service_name>
```

2. Проверить конфигурацию:
```bash
docker compose -f docker-compose.release.yml config
```

3. Пересоздать конкретный сервис:
```bash
docker compose -f docker-compose.release.yml up -d --force-recreate <service_name>
```

### Если порты заняты:

```bash
# Найти процесс, использующий порт
sudo lsof -i :8080
sudo lsof -i :80
sudo lsof -i :443

# Или
sudo netstat -tulpn | grep :8080
```

### Если нужно удалить конкретный контейнер:

```bash
# Остановить
docker stop <CONTAINER_ID>

# Удалить
docker rm <CONTAINER_ID>

# Или принудительно
docker rm -f <CONTAINER_ID>
```

---

## Рекомендуемый workflow для обновления

```bash
# 1. Перейти в директорию
cd ~/mm-shop/release

# 2. Получить последние изменения (если используете git)
git pull

# 3. Остановить контейнеры
docker compose -f docker-compose.release.yml down

# 4. Пересобрать образы
docker compose -f docker-compose.release.yml build

# 5. Запустить
docker compose -f docker-compose.release.yml up -d

# 6. Проверить статус
docker compose -f docker-compose.release.yml ps

# 7. Проверить логи
docker compose -f docker-compose.release.yml logs -f --tail=50
```

---

## Полезные команды

### Просмотр использования ресурсов:
```bash
docker stats
```

### Просмотр всех образов:
```bash
docker images
```

### Просмотр всех контейнеров (включая остановленные):
```bash
docker ps -a
```

### Войти в контейнер:
```bash
docker exec -it <CONTAINER_NAME> /bin/sh
# или для alpine
docker exec -it <CONTAINER_NAME> /bin/ash
```

### Копировать файл из контейнера:
```bash
docker cp <CONTAINER_NAME>:/path/to/file /host/path
```

### Копировать файл в контейнер:
```bash
docker cp /host/path <CONTAINER_NAME>:/path/to/file
```

