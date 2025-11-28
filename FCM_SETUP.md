# Настройка Firebase Cloud Messaging (FCM) для Push-уведомлений

## Описание

Система поддерживает отправку push-уведомлений через Firebase Cloud Messaging (FCM) для Android, iOS и Web приложений.

## Настройка

### 1. Создание Firebase проекта

1. Перейдите на [Firebase Console](https://console.firebase.google.com/)
2. Создайте новый проект или выберите существующий
3. В настройках проекта перейдите в "Project Settings" → "Cloud Messaging"
4. Скопируйте **Server Key** (Legacy Server Key)

### 2. Настройка переменной окружения

Добавьте в файл `env.development` или установите переменную окружения:

```bash
FCM_SERVER_KEY=your_fcm_server_key_here
```

Для production добавьте в `docker-compose.release.yml`:

```yaml
environment:
  - FCM_SERVER_KEY=${FCM_SERVER_KEY}
```

### 3. Регистрация токенов устройств

При первом входе пользователя в приложение (Flutter/Web), необходимо зарегистрировать токен устройства:

```javascript
// Пример для Web
POST /api/v1/device-tokens/
Headers: {
  "Authorization": "Bearer <user_token>"
}
Body: {
  "token": "fcm_token_from_firebase",
  "platform": "web", // или "android", "ios"
  "deviceId": "unique_device_id" // опционально
}
```

### 4. Как это работает

1. **При создании заказа:**
   - Система создает уведомление в БД
   - Отправляет push-уведомление на все активные устройства владельца магазина

2. **При входе пользователя:**
   - Если пользователь не заходил долгое время (3-10 дней)
   - Система проверяет непрочитанные уведомления
   - Отправляет push-уведомление о наличии непрочитанных уведомлений

3. **Deep Linking:**
   - При клике на push-уведомление открывается нужная страница
   - Если токен истек (24 часа), пользователь перенаправляется на логин
   - После логина автоматически открывается нужная страница

## API Endpoints

### Регистрация токена устройства
```
POST /api/v1/device-tokens/
Authorization: Bearer <token>
Body: {
  "token": "fcm_token",
  "platform": "android|ios|web",
  "deviceId": "optional_device_id"
}
```

### Получение токенов пользователя
```
GET /api/v1/device-tokens/
Authorization: Bearer <token>
```

### Удаление токена
```
DELETE /api/v1/device-tokens/:token
Authorization: Bearer <token>
```

## Формат уведомлений

Push-уведомления содержат:
- **title**: Заголовок уведомления
- **body**: Текст уведомления
- **action_url**: Deep link для перехода (например: `/admin#orders?orderId=123`)
- **data**: Дополнительные данные

## Безопасность

- FCM Server Key хранится в переменных окружения
- Токены устройств привязаны к пользователям
- Токены автоматически деактивируются при удалении
- JWT токены истекают через 24 часа

## Отладка

Логи отправки push-уведомлений:
- ✅ Успешная отправка
- ⚠️ Предупреждения (нет токенов, FCM не настроен)
- ❌ Ошибки отправки

## Примечания

- Если `FCM_SERVER_KEY` не настроен, система продолжит работать, но push-уведомления не будут отправляться
- Уведомления всегда сохраняются в БД, даже если push не отправлен
- Пользователь может просмотреть все уведомления в дашборде

