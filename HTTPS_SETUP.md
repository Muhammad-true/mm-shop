# 🔒 Настройка HTTPS для MM Shop

## Проблема

Фронтенд работает на HTTPS, а API на HTTP, поэтому браузер блокирует запросы (mixed content error).

## Решение

Настроить HTTPS для всего приложения через Nginx reverse proxy.

---

## 🚀 Быстрый старт (самоподписанный сертификат)

### 1. Создать SSL сертификат

```bash
cd /root/mm-shop
chmod +x setup-ssl.sh
./setup-ssl.sh
```

### 2. Пересобрать и запустить

```bash
docker-compose -f docker-compose.release.yml up -d --build admin
```

### 3. Проверить

Откройте в браузере: `https://159.89.99.252`

**⚠️ Браузер покажет предупреждение о самоподписанном сертификате - это нормально для тестирования.**

---

## 🌐 Продакшен (Let's Encrypt)

### 1. Установить certbot

```bash
sudo apt-get update
sudo apt-get install certbot
```

### 2. Получить сертификат

```bash
# Остановить nginx в контейнере (освободить порт 80)
docker-compose -f docker-compose.release.yml stop admin

# Получить сертификат
sudo certbot certonly --standalone -d your-domain.com

# Сертификаты будут в:
# /etc/letsencrypt/live/your-domain.com/fullchain.pem
# /etc/letsencrypt/live/your-domain.com/privkey.pem
```

### 3. Обновить конфигурацию

**docker-compose.release.yml:**
```yaml
admin:
  volumes:
    - /etc/letsencrypt:/etc/letsencrypt:ro  # Вместо ./ssl:/etc/nginx/ssl:ro
```

**nginx.production.conf:**
```nginx
ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
```

### 4. Запустить

```bash
docker-compose -f docker-compose.release.yml up -d --build admin
```

### 5. Настроить автообновление сертификата

```bash
# Добавить в crontab
sudo crontab -e

# Добавить строку (обновление каждые 2 месяца)
0 0 1 */2 * certbot renew --quiet && docker-compose -f /root/mm-shop/docker-compose.release.yml restart admin
```

---

## 📝 Что изменилось

1. **nginx.production.conf:**
   - Добавлен HTTP сервер с редиректом на HTTPS
   - Добавлен HTTPS сервер на порту 443
   - Настроены SSL параметры безопасности
   - Добавлен заголовок Strict-Transport-Security

2. **docker-compose.release.yml:**
   - Проброс портов 80 и 443
   - Монтирование SSL сертификатов

3. **Dockerfile.admin.release:**
   - Открыт порт 443

---

## ✅ Проверка

### Проверить HTTPS работает:

```bash
# Проверить порты
docker exec mm-admin-prod netstat -tlnp | grep -E '80|443'

# Проверить SSL
curl -k https://localhost/health
curl -k https://localhost/api/v1/version

# Проверить редирект с HTTP на HTTPS
curl -I http://localhost/health
# Должен вернуть: HTTP/1.1 301 Moved Permanently
```

### Проверить в браузере:

1. Откройте `https://159.89.99.252` (или ваш домен)
2. Проверьте консоль браузера (F12) - не должно быть ошибок mixed content
3. Все запросы к API должны идти через HTTPS

---

## 🐛 Решение проблем

### Ошибка: "SSL certificate not found"

```bash
# Проверить наличие сертификатов
ls -la ssl/
# Должны быть: cert.pem и key.pem

# Если нет - создать
./setup-ssl.sh
```

### Ошибка: "Port 443 already in use"

```bash
# Найти процесс, использующий порт 443
sudo lsof -i :443

# Остановить его или изменить порт в docker-compose
```

### Ошибка: "Mixed Content" в браузере

Убедитесь, что:
1. Открываете сайт через HTTPS (не HTTP)
2. SSL сертификат настроен правильно
3. Nginx слушает на порту 443

---

## 📚 Дополнительная информация

- [Let's Encrypt документация](https://letsencrypt.org/docs/)
- [Nginx SSL настройки](https://nginx.org/en/docs/http/configuring_https_servers.html)
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)

