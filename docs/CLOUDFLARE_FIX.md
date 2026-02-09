# Решение проблемы Cloudflare для загрузки больших файлов

## Проблема

Cloudflare на бесплатном тарифе имеет ограничения:
- **Таймаут запросов: 100 секунд** — если сервер не отвечает в течение 100 секунд, Cloudflare обрывает соединение
- Ограничения на размер файлов
- Проксирование может замедлять передачу больших файлов

При загрузке файлов 30-60 МБ загрузка может занимать больше 100 секунд, особенно при медленном интернете, что приводит к обрыву соединения на ~87% прогресса.

## Решение: Обход Cloudflare через прямой IP

### Что было сделано

1. **Модифицирован `admin/js/updates.js`**:
   - Добавлена функция `getDirectUploadUrl()` для получения прямого URL к серверу
   - Загрузка файлов теперь идет напрямую на IP `159.89.99.252:443` (обход Cloudflare)
   - Добавлен `Host: api.libiss.com` header для правильной маршрутизации в nginx

### Как это работает

```javascript
// Обычный запрос (через Cloudflare):
https://api.libiss.com/api/v1/admin/updates/upload

// Прямой запрос (обход Cloudflare):
https://159.89.99.252:443/api/v1/admin/updates/upload
Host: api.libiss.com
```

Nginx на сервере получает запрос с правильным Host header и маршрутизирует его к нужному server block.

### Настройка nginx для прямого IP

Убедись, что в `nginx.production.conf` есть server block, который обрабатывает запросы по IP:

```nginx
server {
    listen 443 ssl http2;
    server_name api.libiss.com 159.89.99.252;  # Добавь IP в server_name
    
    # ... остальная конфигурация
}
```

Или добавь отдельный server block для прямого IP:

```nginx
server {
    listen 443 ssl http2 default_server;
    server_name 159.89.99.252;
    
    # Используем тот же SSL сертификат
    ssl_certificate /etc/letsencrypt/live/admin.libiss.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.libiss.com/privkey.pem;
    
    # Проксируем на тот же backend
    location / {
        proxy_pass http://mm-api-prod:8080;
        # ... остальные настройки как в api.libiss.com
    }
}
```

### Альтернативные решения

#### 1. Использовать HTTP вместо HTTPS (если SSL не настроен для IP)

Если SSL не настроен для прямого IP, можно использовать HTTP:

```javascript
const DIRECT_SERVER_PROTOCOL = 'http';
const DIRECT_SERVER_PORT = '80';
```

**⚠️ Внимание**: HTTP небезопасен, но для внутренних загрузок может быть приемлемо.

#### 2. Настроить Cloudflare Page Rules (только на платных тарифах)

На платных тарифах Cloudflare можно увеличить таймаут через Page Rules:
- URL Pattern: `api.libiss.com/api/v1/admin/updates/upload`
- Setting: `Edge Cache TTL` → `Bypass`
- Setting: `Cache Level` → `Bypass`

#### 3. Отключить Cloudflare для конкретного поддомена

В панели Cloudflare:
1. DNS → Records
2. Найди `api.libiss.com`
3. Отключи проксирование (серый облачко вместо оранжевого)

**⚠️ Внимание**: Это отключит защиту Cloudflare для всего API.

#### 4. Использовать отдельный поддомен без Cloudflare

Создай поддомен `upload.libiss.com` без Cloudflare проксирования:
- DNS Record: `upload.libiss.com` → `159.89.99.252` (серый облачко)
- Используй этот поддомен только для загрузки файлов

### Проверка работы

1. Открой консоль браузера (F12)
2. Попробуй загрузить файл
3. В Network tab проверь, что запрос идет на `159.89.99.252:443`
4. Проверь, что Host header установлен: `Host: api.libiss.com`

### Логи для диагностики

```bash
# Проверь логи nginx
docker logs mm-admin-prod | grep upload

# Проверь логи API
docker logs mm-api-prod | grep UploadUpdate

# Проверь доступность прямого IP
curl -I -H "Host: api.libiss.com" https://159.89.99.252:443/api/health
```

### Если SSL не работает на прямом IP

Если при обращении к `https://159.89.99.252:443` получаешь ошибку SSL:

1. **Вариант 1**: Используй HTTP (менее безопасно):
   ```javascript
   const DIRECT_SERVER_PROTOCOL = 'http';
   const DIRECT_SERVER_PORT = '80';
   ```

2. **Вариант 2**: Настрой SSL для IP в nginx (используй wildcard или self-signed сертификат)

3. **Вариант 3**: Используй поддомен без Cloudflare (см. выше)

### Резюме

✅ **Текущее решение**: Загрузка файлов идет напрямую на IP `159.89.99.252:443` с Host header `api.libiss.com`, что обходит ограничения Cloudflare.

✅ **Преимущества**:
- Нет таймаута 100 секунд
- Быстрее передача (без проксирования)
- Работает на бесплатном тарифе Cloudflare

⚠️ **Требования**:
- SSL должен быть настроен для прямого IP, или используй HTTP
- Nginx должен правильно обрабатывать запросы с Host header

