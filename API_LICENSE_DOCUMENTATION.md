# API Документация: Система лицензий и подписок

## Версия API: 1.4.0

## Обзор

Система лицензий позволяет магазинам подписываться на использование приложения. Процесс включает:
1. Регистрацию магазина (создание пользователя shop_owner + магазина)
2. Выбор плана подписки и оплату
3. Создание лицензии после оплаты
4. Активацию лицензии в Flutter приложении

---

## Публичные эндпоинты (для сайта и Flutter приложения)

### 1. Регистрация магазина

**POST** `/api/v1/shop-registration/register`

Регистрирует новый магазин. Создает пользователя с ролью `shop_owner` и магазин.

#### Запрос:
```json
{
  "name": "Иван Иванов",
  "email": "ivan@example.com",
  "password": "password123",
  "phone": "+998901234567",
  "shopName": "Мой магазин",
  "inn": "123456789",
  "description": "Описание магазина",
  "address": "Улица, дом",
  "cityId": "uuid-города" // опционально
}
```

#### Ответ (201 Created):
```json
{
  "success": true,
  "message": "Shop registered successfully",
  "data": {
    "user": {
      "id": "uuid",
      "name": "Иван Иванов",
      "email": "ivan@example.com",
      "phone": "+998901234567",
      "role": "shop_owner"
    },
    "shop": {
      "id": "uuid",
      "name": "Мой магазин",
      "inn": "123456789",
      "description": "Описание магазина",
      "address": "Улица, дом",
      "cityId": "uuid-города"
    },
    "token": "jwt-token" // Токен для автоматического входа
  }
}
```

#### Ошибки:
- `400 Bad Request` - Неверные данные запроса
- `409 Conflict` - Пользователь с таким email уже существует
- `500 Internal Server Error` - Ошибка сервера

---

### 2. Получить планы подписки

**GET** `/api/v1/subscriptions/plans`

Возвращает список всех активных планов подписки.

#### Ответ (200 OK):
```json
{
  "success": true,
  "data": {
    "plans": [
      {
        "id": "uuid",
        "name": "Месячная подписка",
        "description": "Доступ к приложению на 1 месяц",
        "subscriptionType": "monthly",
        "price": 29.99,
        "currency": "USD",
        "durationMonths": 1,
        "isActive": true,
        "features": "{\"products\": true, \"orders\": true, \"analytics\": true}",
        "sortOrder": 1,
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
      },
      {
        "id": "uuid",
        "name": "Годовая подписка",
        "description": "Доступ к приложению на 1 год (экономия 20%)",
        "subscriptionType": "yearly",
        "price": 299.99,
        "currency": "USD",
        "durationMonths": 12,
        "isActive": true,
        "features": "{\"products\": true, \"orders\": true, \"analytics\": true, \"priority_support\": true}",
        "sortOrder": 2,
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
      },
      {
        "id": "uuid",
        "name": "Пожизненная подписка",
        "description": "Пожизненный доступ к приложению",
        "subscriptionType": "lifetime",
        "price": 999.99,
        "currency": "USD",
        "durationMonths": 0,
        "isActive": true,
        "features": "{\"products\": true, \"orders\": true, \"analytics\": true, \"priority_support\": true, \"lifetime_updates\": true}",
        "sortOrder": 3,
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

### 3. Получить информацию о плане подписки

**GET** `/api/v1/subscriptions/plans/:id`

Возвращает информацию о конкретном плане подписки.

#### Параметры:
- `id` (path) - UUID плана подписки

#### Ответ (200 OK):
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Месячная подписка",
    "description": "Доступ к приложению на 1 месяц",
    "subscriptionType": "monthly",
    "price": 29.99,
    "currency": "USD",
    "durationMonths": 1,
    "isActive": true,
    "features": "{\"products\": true, \"orders\": true, \"analytics\": true}",
    "sortOrder": 1,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

---

### 4. Подписка магазина (создание лицензии после оплаты)

**POST** `/api/v1/shop-registration/subscribe`

Создает лицензию для магазина после успешной оплаты. Вызывается после обработки платежа на сайте.

#### Запрос:
```json
{
  "shopId": "uuid-магазина",
  "subscriptionPlanId": "uuid-плана",
  "paymentProvider": "stripe", // stripe, paypal, etc.
  "paymentTransactionId": "txn_123456789",
  "paymentAmount": 29.99,
  "paymentCurrency": "USD", // опционально, по умолчанию из плана
  "autoRenew": false
}
```

#### Ответ (201 Created):
```json
{
  "success": true,
  "message": "Subscription created successfully",
  "data": {
    "id": "uuid",
    "licenseKey": "XXXX-XXXX-XXXX-XXXX",
    "shopId": "uuid",
    "activationType": "payment",
    "subscriptionType": "monthly",
    "subscriptionStatus": "active",
    "activatedAt": "2024-01-01T00:00:00Z",
    "expiresAt": "2024-02-01T00:00:00Z",
    "lastPaymentDate": "2024-01-01T00:00:00Z",
    "nextPaymentDate": "2024-02-01T00:00:00Z",
    "paymentProvider": "stripe",
    "paymentTransactionId": "txn_123456789",
    "paymentAmount": 29.99,
    "paymentCurrency": "USD",
    "userId": "uuid",
    "isActive": true,
    "autoRenew": false,
    "isValid": true,
    "isExpired": false,
    "daysRemaining": 30,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

#### Ошибки:
- `400 Bad Request` - Неверные данные запроса
- `404 Not Found` - Магазин или план подписки не найден
- `409 Conflict` - У магазина уже есть активная лицензия
- `500 Internal Server Error` - Ошибка сервера

**Важно:** После успешного создания лицензии, `licenseKey` нужно отобразить пользователю для активации в Flutter приложении.

---

## Эндпоинты для Flutter приложения

### 5. Проверка лицензии

**POST** `/api/v1/licenses/check`

Проверяет статус лицензии по ключу. Используется при первом запуске приложения.

#### Запрос:
```json
{
  "licenseKey": "XXXX-XXXX-XXXX-XXXX",
  "deviceId": "unique-device-id",
  "deviceInfo": {
    "platform": "android",
    "model": "Samsung Galaxy S21",
    "manufacturer": "Samsung",
    "osVersion": "Android 12",
    "appVersion": "1.0.0"
  }
}
```

**Поля:**
- `licenseKey` (required) - Ключ лицензии
- `deviceId` (required) - Уникальный ID устройства (например, Android ID или iOS IdentifierForVendor)
- `deviceInfo` (optional) - Информация о железе для дополнительной проверки

#### Ответ (200 OK):
```json
{
  "success": true,
  "data": {
    "isValid": true,
    "isExpired": false,
    "subscriptionStatus": "active",
    "subscriptionType": "monthly",
    "expiresAt": "2024-02-01T00:00:00Z",
    "daysRemaining": 30,
    "deviceMatch": true
  }
}
```

**Если лицензия активирована на другом устройстве:**
```json
{
  "success": true,
  "data": {
    "isValid": false,
    "isExpired": false,
    "subscriptionStatus": "active",
    "subscriptionType": "monthly",
    "expiresAt": "2024-02-01T00:00:00Z",
    "daysRemaining": 30,
    "deviceMatch": false,
    "error": "License is activated on a different device"
  }
}
```

#### Ошибки:
- `400 Bad Request` - Неверные данные запроса
- `404 Not Found` - Лицензия не найдена
- `500 Internal Server Error` - Ошибка сервера

---

### 6. Активация лицензии

**POST** `/api/v1/licenses/activate`

Активирует лицензию для магазина. Привязывает лицензию к конкретному магазину.

#### Запрос:
```json
{
  "licenseKey": "XXXX-XXXX-XXXX-XXXX",
  "shopId": "uuid-магазина",
  "deviceId": "unique-device-id",
  "deviceInfo": {
    "platform": "android",
    "model": "Samsung Galaxy S21",
    "manufacturer": "Samsung",
    "osVersion": "Android 12",
    "appVersion": "1.0.0"
  }
}
```

**Поля:**
- `licenseKey` (required) - Ключ лицензии
- `shopId` (required) - UUID магазина
- `deviceId` (required) - Уникальный ID устройства
- `deviceInfo` (required) - Информация о железе (платформа, модель, производитель, версия ОС и т.д.)

#### Ответ (200 OK):
```json
{
  "success": true,
  "message": "License activated successfully",
  "data": {
    "id": "uuid",
    "licenseKey": "XXXX-XXXX-XXXX-XXXX",
    "shopId": "uuid",
    "activationType": "payment",
    "subscriptionType": "monthly",
    "subscriptionStatus": "active",
    "activatedAt": "2024-01-01T00:00:00Z",
    "expiresAt": "2024-02-01T00:00:00Z",
    "isValid": true,
    "isExpired": false,
    "daysRemaining": 30,
    "shop": {
      "id": "uuid",
      "name": "Мой магазин",
      "inn": "123456789"
    }
  }
}
```

#### Ошибки:
- `400 Bad Request` - Неверные данные запроса
- `403 Forbidden` - Лицензия уже активирована на другом устройстве
- `404 Not Found` - Лицензия или магазин не найдены
- `500 Internal Server Error` - Ошибка сервера

**Если лицензия уже активирована на этом устройстве:**
```json
{
  "success": true,
  "message": "License already activated on this device",
  "data": {
    "id": "uuid",
    "licenseKey": "XXXX-XXXX-XXXX-XXXX",
    "deviceId": "unique-device-id",
    ...
  }
}
```

**Если лицензия активирована на другом устройстве:**
```json
{
  "success": false,
  "error": "License is already activated on a different device",
  "data": {
    "deviceId": "other-device-id"
  }
}
```

---

## Админские эндпоинты (требуют авторизации)

### 7. Список всех лицензий

**GET** `/api/v1/admin/licenses`

Возвращает список всех лицензий с фильтрами.

#### Заголовки:
```
Authorization: Bearer <admin-token>
```

#### Query параметры:
- `shopId` (optional) - Фильтр по ID магазина
- `status` (optional) - Фильтр по статусу (active, expired, cancelled, pending)

#### Пример:
```
GET /api/v1/admin/licenses?shopId=uuid&status=active
```

#### Ответ (200 OK):
```json
{
  "success": true,
  "data": {
    "licenses": [
      {
        "id": "uuid",
        "licenseKey": "XXXX-XXXX-XXXX-XXXX",
        "shopId": "uuid",
        "subscriptionType": "monthly",
        "subscriptionStatus": "active",
        "expiresAt": "2024-02-01T00:00:00Z",
        "isValid": true,
        "daysRemaining": 30,
        "shop": {
          "id": "uuid",
          "name": "Мой магазин"
        }
      }
    ]
  }
}
```

---

### 8. Получить информацию о лицензии

**GET** `/api/v1/admin/licenses/:id`

Возвращает полную информацию о лицензии.

#### Параметры:
- `id` (path) - UUID лицензии

#### Ответ (200 OK):
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "licenseKey": "XXXX-XXXX-XXXX-XXXX",
    "shopId": "uuid",
    "subscriptionType": "monthly",
    "subscriptionStatus": "active",
    "expiresAt": "2024-02-01T00:00:00Z",
    "paymentAmount": 29.99,
    "isValid": true,
    "daysRemaining": 30,
    "shop": {
      "id": "uuid",
      "name": "Мой магазин"
    }
  }
}
```

---

### 9. Создать лицензию

**POST** `/api/v1/admin/licenses`

Создает новую лицензию вручную (для админов).

#### Запрос:
```json
{
  "shopId": "uuid-магазина", // опционально
  "subscriptionType": "monthly", // monthly, yearly, lifetime
  "activationType": "manual", // manual, payment
  "paymentAmount": 29.99,
  "paymentCurrency": "USD",
  "paymentProvider": "manual",
  "paymentTransactionId": "",
  "autoRenew": false,
  "notes": "Создано вручную"
}
```

#### Ответ (201 Created):
```json
{
  "success": true,
  "message": "License created successfully",
  "data": {
    "id": "uuid",
    "licenseKey": "XXXX-XXXX-XXXX-XXXX",
    "shopId": "uuid",
    "subscriptionType": "monthly",
    "subscriptionStatus": "pending",
    "isValid": false,
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

---

### 10. Генерация лицензии для магазина

**POST** `/api/v1/admin/licenses/shops/:shopId/generate`

Генерирует лицензию для конкретного магазина (после оплаты или вручную).

#### Параметры:
- `shopId` (path) - UUID магазина

#### Запрос:
```json
{
  "subscriptionType": "monthly",
  "paymentAmount": 29.99,
  "paymentCurrency": "USD",
  "paymentProvider": "stripe",
  "paymentTransactionId": "txn_123456789",
  "autoRenew": false,
  "notes": "Оплачено через Stripe"
}
```

#### Ответ (201 Created):
```json
{
  "success": true,
  "message": "License generated successfully",
  "data": {
    "id": "uuid",
    "licenseKey": "XXXX-XXXX-XXXX-XXXX",
    "shopId": "uuid",
    "subscriptionType": "monthly",
    "subscriptionStatus": "active",
    "expiresAt": "2024-02-01T00:00:00Z",
    "isValid": true,
    "daysRemaining": 30
  }
}
```

---

### 11. Обновить лицензию

**PUT** `/api/v1/admin/licenses/:id`

Обновляет статус лицензии.

#### Параметры:
- `id` (path) - UUID лицензии

#### Запрос:
```json
{
  "subscriptionStatus": "active", // active, expired, cancelled, pending
  "isActive": true,
  "autoRenew": true,
  "notes": "Обновлено вручную"
}
```

#### Ответ (200 OK):
```json
{
  "success": true,
  "message": "License updated successfully",
  "data": {
    "id": "uuid",
    "subscriptionStatus": "active",
    "isActive": true,
    "autoRenew": true
  }
}
```

---

## Типы данных

### SubscriptionType
- `monthly` - Месячная подписка
- `yearly` - Годовая подписка
- `lifetime` - Пожизненная подписка

### SubscriptionStatus
- `active` - Активна
- `expired` - Истекла
- `cancelled` - Отменена
- `pending` - Ожидает оплаты

### ActivationType
- `manual` - Ручная активация (админом)
- `payment` - Активация через оплату

---

## Процесс работы

### Для сайта:

1. **Регистрация магазина:**
   ```
   POST /api/v1/shop-registration/register
   ```
   - Создается пользователь с ролью `shop_owner`
   - Создается магазин
   - Возвращается токен для входа

2. **Выбор плана подписки:**
   ```
   GET /api/v1/subscriptions/plans
   ```
   - Пользователь выбирает план
   - Происходит оплата через платежную систему

3. **Создание лицензии после оплаты:**
   ```
   POST /api/v1/shop-registration/subscribe
   ```
   - Создается лицензия с ключом
   - Ключ отображается пользователю

### Для Flutter приложения:

1. **При первом запуске:**
   - Запрашивается ключ лицензии у пользователя

2. **Проверка ключа:**
   ```
   POST /api/v1/licenses/check
   ```
   - Проверяется валидность ключа

3. **Активация лицензии:**
   ```
   POST /api/v1/licenses/activate
   ```
   - Лицензия привязывается к магазину
   - Приложение готово к работе

---

## База данных

### Таблица `licenses`
- `id` (UUID, PK)
- `license_key` (VARCHAR, UNIQUE) - Уникальный ключ лицензии
- `shop_id` (UUID, FK) - ID магазина
- `user_id` (UUID, FK) - ID владельца магазина
- `activation_type` (ENUM) - manual, payment
- `subscription_type` (ENUM) - monthly, yearly, lifetime
- `subscription_status` (ENUM) - active, expired, cancelled, pending
- `activated_at` (DATETIME)
- `expires_at` (DATETIME)
- `last_payment_date` (DATETIME)
- `next_payment_date` (DATETIME)
- `payment_provider` (VARCHAR)
- `payment_transaction_id` (VARCHAR)
- `payment_amount` (DECIMAL)
- `payment_currency` (VARCHAR)
- `is_active` (BOOLEAN)
- `auto_renew` (BOOLEAN)
- `notes` (TEXT)
- `created_at` (DATETIME)
- `updated_at` (DATETIME)

### Таблица `subscription_plans`
- `id` (UUID, PK)
- `name` (VARCHAR)
- `description` (TEXT)
- `subscription_type` (ENUM)
- `price` (DECIMAL)
- `currency` (VARCHAR)
- `duration_months` (INT)
- `is_active` (BOOLEAN)
- `features` (TEXT, JSON)
- `sort_order` (INT)
- `created_at` (DATETIME)
- `updated_at` (DATETIME)

---

## Миграции

Миграции выполняются автоматически при запуске приложения через GORM AutoMigrate.

Таблицы создаются автоматически:
- `licenses`
- `subscription_plans`

Дефолтные планы подписки создаются автоматически при первом запуске.

---

## Примеры использования

### Пример: Регистрация магазина и подписка

```bash
# 1. Регистрация магазина
curl -X POST https://api.example.com/api/v1/shop-registration/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Иван Иванов",
    "email": "ivan@example.com",
    "password": "password123",
    "phone": "+998901234567",
    "shopName": "Мой магазин",
    "inn": "123456789",
    "description": "Описание магазина",
    "address": "Улица, дом"
  }'

# 2. Получение планов подписки
curl -X GET https://api.example.com/api/v1/subscriptions/plans

# 3. Подписка (после оплаты)
curl -X POST https://api.example.com/api/v1/shop-registration/subscribe \
  -H "Content-Type: application/json" \
  -d '{
    "shopId": "uuid-магазина",
    "subscriptionPlanId": "uuid-плана",
    "paymentProvider": "stripe",
    "paymentTransactionId": "txn_123456789",
    "paymentAmount": 29.99,
    "paymentCurrency": "USD",
    "autoRenew": false
  }'
```

### Пример: Активация в Flutter приложении

```bash
# 1. Проверка ключа
curl -X POST https://api.example.com/api/v1/licenses/check \
  -H "Content-Type: application/json" \
  -d '{
    "licenseKey": "XXXX-XXXX-XXXX-XXXX"
  }'

# 2. Активация
curl -X POST https://api.example.com/api/v1/licenses/activate \
  -H "Content-Type: application/json" \
  -d '{
    "licenseKey": "XXXX-XXXX-XXXX-XXXX",
    "shopId": "uuid-магазина"
  }'
```

---

## Версия
1.4.0

## Дата обновления
2024-12-01

