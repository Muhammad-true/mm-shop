# MM Shop - Backend API

## Локальная разработка

### Быстрый старт

1. **Запустить локальную среду:**
   ```bash
   docker-compose up -d
   ```

2. **Доступ к сервисам:**
   - **API:** http://localhost:8080
   - **Админ панель:** http://localhost:3000
   - **PgAdmin (БД):** http://localhost:5050
   - **Health:** http://localhost:8080/health

3. **Остановить:**
   ```bash
   docker-compose down
   ```

### Данные для входа в админку:
- Email: admin@mm.com  
- Пароль: admin123

**PgAdmin (управление БД):**
- URL: http://localhost:5050
- Email: admin@mm.com
- Password: admin123

### Доступ к БД

**Прямое подключение:**
- Host: localhost
- Port: 5432
- Database: mm_shop_dev
- User: mm_user
- Password: dev_password

### Переменные окружения

Файл `env.development` используется для локальной разработки.

## 🚀 Деплой на продакшн сервер

### На сервере выполнить:

```bash
# Перейти в директорию проекта
cd /root/mm-shop

# Получить последние изменения из Git
git pull origin main

# Остановить старые контейнеры
docker-compose -f docker-compose.release.yml -f docker-compose.release.override.yml down

# Пересобрать БЕЗ кэша (чтобы получить новые файлы)
docker-compose -f docker-compose.release.yml -f docker-compose.release.override.yml build --no-cache admin api

# Запустить новые контейнеры
docker-compose -f docker-compose.release.yml -f docker-compose.release.override.yml up -d
```

### Или одной командой:

```bash
cd /root/mm-shop && \
git pull origin main && \
docker-compose -f docker-compose.release.yml -f docker-compose.release.override.yml down && \
docker-compose -f docker-compose.release.yml -f docker-compose.release.override.yml up -d --build
```

### Или раздельно (API и Admin):

```bash
cd /root/mm-shop
git pull origin main
docker-compose -f docker-compose.release.yml up -d --build api
docker-compose -f docker-compose.release.yml up -d --build admin
```

## Полезные команды

**Посмотреть логи (prod):**
```bash
docker logs mm-api-prod --tail 50 -f
docker logs mm-admin-prod --tail 50 -f
```

**Посмотреть логи (dev):**
```bash
docker logs mm-api-dev --tail 50 -f
docker logs mm-admin-dev --tail 50 -f
```

**Войти в контейнер:**
```bash
docker exec -it mm-api-prod sh
docker exec -it mm-admin-prod sh
```

**Проверить статус контейнеров:**
```bash
docker ps
```

**Пересоздать БД:**
```bash
docker-compose down -v
docker-compose up -d
```

**Подключиться к pgAdmin:**
1. Откройте http://localhost:5050
2. Войдите: admin@mm.com / admin123
3. Добавьте новый сервер:
   - Name: MM Shop Dev
   - Host: postgres (важно!)
   - Port: 5432
   - Database: mm_shop_dev
   - Username: mm_user
   - Password: dev_password

