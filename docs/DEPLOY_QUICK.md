# 🚀 Быстрый деплой

## ⚠️ ПЕРВЫЙ РАЗ: Настройка Volume для Docker (если добавлен новый диск)

Если у вас подключен Volume (например, `/mnt/mm_shop_data`), перенесите Docker данные на него:

```bash
# 1. Остановите Docker и контейнеры
cd /root/mm-shop/release
docker compose -f docker-compose.release.yml down
sudo systemctl stop docker

# 2. Проверьте точку монтирования Volume
df -h | grep /mnt
# Должен быть виден ваш Volume (например, /mnt/mm_shop_data)

# 3. Перенесите данные Docker на Volume
VOLUME_PATH="/mnt/mm_shop_data"  # Замените на ваш путь
sudo mkdir -p $VOLUME_PATH/docker
sudo rsync -avxP /var/lib/docker/ $VOLUME_PATH/docker/

# 4. Создайте резервную копию старой директории и создайте симлинк
sudo mv /var/lib/docker /var/lib/docker.old
sudo ln -s $VOLUME_PATH/docker /var/lib/docker

# 5. Запустите Docker
sudo systemctl start docker

# 6. Проверьте, что все работает
docker ps
df -h

# 7. Проверьте использование диска
df -h
# Должно показать, что /mnt/mm_shop_data используется для Docker

# 8. Если все работает, удалите старую директорию (опционально, после проверки)
# sudo rm -rf /var/lib/docker.old
```

## ✅ После переноса Docker на Volume

Docker теперь использует Volume для всех операций (образы, контейнеры, build cache).
Теперь можно безопасно собирать образы без ошибок "no space left on device".

## На сервере выполните:

```bash
cd /root/mm-shop/release

# Если есть локальные изменения - откатываем их
git checkout .

# Обновляем код
git pull origin main

# Создайте .env.production файл (если еще не создан) с переменными окружения:
# LEMONSQUEEZY_API_KEY=your-api-key
# LEMONSQUEEZY_STORE_ID=your-store-id
# JWT_SECRET=your-jwt-secret
# PGADMIN_EMAIL=admin@mm.com (опционально, по умолчанию admin@mm.com)
# PGADMIN_PASSWORD=your-secure-password (опционально, по умолчанию admin123)
# POSTGRES_PASSWORD=your-postgres-password (опционально, по умолчанию muhammadjon)
# и другие необходимые переменные

# ⚠️ ВАЖНО: Если используете поддомен для PgAdmin (pgadmin.libiss.com):
# 1. Настройте DNS запись A для pgadmin.libiss.com на IP сервера (159.89.99.252)
# 2. Получите SSL сертификат для поддомена (см. инструкцию ниже)

# ОСТАНОВКА и удаление контейнеров для чистого билда
docker compose -f docker-compose.release.yml stop api admin pgadmin
docker compose -f docker-compose.release.yml rm -f api admin pgadmin

# Удаляем старые образы
docker rmi release-api release-admin 2>/dev/null || true

# ОЧИСТКА МЕСТА НА ДИСКЕ (если нехватка места)
# ⚠️ ВАЖНО: Выполняйте эти команды перед сборкой, если видите ошибку "no space left on device"

# 1. Проверяем использование диска
echo "=== Использование диска ==="
df -h
echo ""
echo "=== Использование Docker ==="
docker system df
echo ""
echo "=== Самые большие директории Docker ==="
du -sh /var/lib/docker/* 2>/dev/null | sort -h | tail -10

# 2. Очищаем build cache (освобождает ~400-500MB)
echo "Очищаем build cache..."
docker builder prune -af

# 3. Очищаем неиспользуемые образы, контейнеры, volumes
echo "Очищаем неиспользуемые ресурсы Docker..."
docker system prune -af --volumes

# 4. Очищаем старые образы (если есть)
echo "Очищаем dangling образы..."
docker images --filter "dangling=true" -q | xargs -r docker rmi 2>/dev/null || true

# 5. Очищаем логи Docker (могут занимать много места)
echo "Очищаем логи..."
journalctl --vacuum-time=7d 2>/dev/null || true
find /var/lib/docker/containers -name "*-json.log" -exec truncate -s 0 {} \; 2>/dev/null || true

# 6. Очищаем временные файлы системы
echo "Очищаем временные файлы..."
rm -rf /tmp/* /var/tmp/* 2>/dev/null || true
rm -rf /root/.cache/go-build 2>/dev/null || true

# 7. Проверяем свободное место после очистки
echo ""
echo "=== Свободное место после очистки ==="
df -h

# 8. Если все еще мало места (< 1GB свободно), проверьте другие директории
if [ $(df / | tail -1 | awk '{print $4}') -lt 1048576 ]; then
    echo "⚠️ ВНИМАНИЕ: Мало места на диске! Проверьте:"
    echo "  - Логи приложений: du -sh /var/log/*"
    echo "  - Старые бэкапы: find /root -name '*.dump' -o -name '*.sql'"
    echo "  - Большие файлы: find / -type f -size +100M 2>/dev/null | head -10"
fi

# ПЕРЕСБОРКА без кэша и запуск
docker compose -f docker-compose.release.yml build --no-cache api admin
docker compose -f docker-compose.release.yml up -d api admin pgadmin
```

## Проверка:

```bash
docker ps
docker logs mm-api-prod --tail 50 -f
docker logs mm-pgadmin-prod --tail 50 -f
```

## Доступ к сервисам:

- **API:** http://159.89.99.252:8080 или https://api.libiss.com
- **Admin Panel:** https://admin.libiss.com
- **PgAdmin (через поддомен):** https://pgadmin.libiss.com (рекомендуется)
- **PgAdmin (прямой доступ):** http://159.89.99.252:5050 (резервный)
  - Email: admin@mm.com (или значение из PGADMIN_EMAIL)
  - Password: admin123 (или значение из PGADMIN_PASSWORD)

## 🔒 Настройка SSL для поддомена PgAdmin (pgadmin.libiss.com):

Если вы хотите использовать поддомен для доступа к PgAdmin:

```bash
# 1. Убедитесь, что DNS запись A для pgadmin.libiss.com указывает на IP сервера (159.89.99.252)

# 2. Остановите nginx контейнер (освободить порт 80)
docker compose -f docker-compose.release.yml stop admin

# 3. Получите SSL сертификат для поддомена
sudo certbot certonly --standalone -d pgadmin.libiss.com

# 4. Если у вас уже есть сертификат для admin.libiss.com, можно добавить поддомен к существующему:
sudo certbot certonly --standalone -d admin.libiss.com -d api.libiss.com -d shop.libiss.com -d pgadmin.libiss.com

# 5. Запустите контейнеры заново
docker compose -f docker-compose.release.yml up -d admin pgadmin

# 6. Проверьте доступ: https://pgadmin.libiss.com
```

## Версия:

**1.2.10** - Добавлена поддержка поддомена для PgAdmin (pgadmin.libiss.com) через nginx reverse proxy

