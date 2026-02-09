# Быстрое решение проблемы Cloudflare

## Проблема
Cloudflare обрывает загрузку файлов на ~87% из-за таймаута 100 секунд.

## Решение
Загрузка файлов теперь идет **напрямую на IP сервера по HTTP**, обходя Cloudflare.

## Что изменилось
1. ✅ Загрузка файлов идет на `http://159.89.99.252:80` (прямой IP, HTTP)
2. ✅ Nginx настроен для обработки HTTP запросов по IP без редиректа на HTTPS
3. ✅ Host header `api.libiss.com` передается для правильной маршрутизации
4. ✅ Используется HTTP, так как SSL сертификат не выдан для IP адреса

## Проверка
1. Открой консоль браузера (F12 → Network)
2. Загрузи файл
3. Проверь, что запрос идет на `http://159.89.99.252:80` (не на `api.libiss.com`)

## Если не работает
1. **Проверь доступность IP**:
   ```bash
   curl -I -H "Host: api.libiss.com" http://159.89.99.252:80/api/health
   ```

2. **Проверь, что nginx слушает на порту 80**:
   ```bash
   docker exec -it mm-admin-prod netstat -tlnp | grep :80
   ```

3. **Проверь логи**:
   ```bash
   docker logs mm-api-prod | grep UploadUpdate
   docker logs mm-admin-prod | grep upload
   ```

## Подробности
См. `CLOUDFLARE_FIX.md` для полной документации.

