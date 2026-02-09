# Быстрое решение проблемы Cloudflare

## Проблема
Cloudflare обрывает загрузку файлов на ~87% из-за таймаута 100 секунд.

## Решение
Загрузка файлов теперь идет **напрямую на IP сервера**, обходя Cloudflare.

## Что изменилось
1. ✅ Загрузка файлов идет на `https://159.89.99.252:443` (прямой IP)
2. ✅ Nginx настроен для обработки запросов по IP
3. ✅ Host header `api.libiss.com` передается для правильной маршрутизации

## Проверка
1. Открой консоль браузера (F12 → Network)
2. Загрузи файл
3. Проверь, что запрос идет на `159.89.99.252:443`

## Если не работает
1. **Ошибка SSL**: Используй HTTP вместо HTTPS (менее безопасно, но работает)
   - В `admin/js/updates.js` измени:
     ```javascript
     const DIRECT_SERVER_PROTOCOL = 'http';
     const DIRECT_SERVER_PORT = '80';
     ```

2. **Проверь доступность IP**:
   ```bash
   curl -I -H "Host: api.libiss.com" https://159.89.99.252:443/api/health
   ```

3. **Проверь логи**:
   ```bash
   docker logs mm-api-prod | grep UploadUpdate
   docker logs mm-admin-prod | grep upload
   ```

## Подробности
См. `CLOUDFLARE_FIX.md` для полной документации.

