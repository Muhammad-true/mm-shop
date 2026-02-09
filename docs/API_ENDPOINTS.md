# MM Shop API - Описание эндпоинтов

**Версия API:** v1.3.3  
**Базовый URL:** `/api/v1`

## Общие заголовки

Все защищенные эндпоинты требуют заголовок:
```
Authorization: Bearer <token>
```

## Формат ответа

### Успешный ответ
```json
{
  "success": true,
  "data": { ... },
  "message": "Success message"
}
```

### Ошибка
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": "Additional details"
  }
}
```

---

## 🔓 Публичные эндпоинты (без аутентификации)

### Информация

#### `GET /health`
Проверка работоспособности API

**Ответ:**
```json
{
  "status": "ok",
  "message": "MM API is running",
  "version": "1.3.3"
}
```

#### `GET /version`
Информация о версии API

**Ответ:**
```json
{
  "version": "1.3.3",
  "name": "MM API",
  "build": "development",
  "changes": "..."
}
```

---

### Аутентификация

#### `POST /auth/register`
Регистрация нового пользователя

**Тело запроса:**
```json
{
  "name": "Иван Иванов",
  "email": "ivan@example.com",
  "phone": "+992927781020",
  "password": "password123"
}
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "user": { ... },
    "token": "jwt_token",
    "refreshToken": "jwt_token"
  }
}
```

#### `POST /auth/login`
Вход пользователя

**Тело запроса:**
```json
{
  "phone": "+992927781020",
  "password": "password123"
}
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "user": { ... },
    "token": "jwt_token",
    "refreshToken": "jwt_token"
  }
}
```

#### `POST /auth/guest-token`
Создание токена для гостевого пользователя

**Тело запроса:**
```json
{
  "name": "Гость",
  "phone": "+992927781020"
}
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "user": { ... },
    "token": "jwt_token",
    "refreshToken": "jwt_token"
  }
}
```

#### `POST /auth/refresh`
Обновление JWT токена

**Тело запроса:**
```json
{
  "refreshToken": "jwt_token"
}
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "token": "new_jwt_token",
    "refreshToken": "new_jwt_token"
  }
}
```

#### `POST /auth/forgot-password`
Восстановление пароля

**Тело запроса:**
```json
{
  "phone": "+992927781020"
}
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "status": "pending",
    "message": "If this phone exists, a reset code was sent"
  }
}
```

---

### Категории (публичные)

#### `GET /categories`
Получить список всех категорий

**Параметры запроса:**
- `limit` (int) - лимит результатов
- `offset` (int) - смещение

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Одежда",
      "description": "Описание",
      "imageUrl": "/images/categories/...",
      "parentId": null,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### `GET /categories/:id`
Получить категорию по ID

**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Одежда",
    "description": "Описание",
    "iconUrl": "/images/categories/...",
    "parentId": null,
    "subcategories": [ ... ],
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

#### `GET /categories/:id/products`
Получить товары категории

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `gender` (string) - фильтр по полу: "boy", "girl", "unisex"
- `search` (string) - поиск

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Товар",
      "description": "Описание",
      "categoryId": "uuid",
      "ownerId": "uuid",
      "shop": {
        "id": "uuid",
        "name": "Название магазина",
        "inn": "123456789"
      },
      "variations": [ ... ],
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### Города (публичные)

#### `GET /cities`
Получить список всех активных городов

**Ответ:**
```json
{
  "success": true,
  "data": {
    "cities": [
      {
        "id": "uuid",
        "name": "Душанбе",
        "latitude": 38.5598,
        "longitude": 68.7870,
        "isActive": true,
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
      },
      {
        "id": "uuid",
        "name": "Канибадам",
        "latitude": 40.2833,
        "longitude": 70.4167,
        "isActive": true,
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

#### `GET /cities/:id`
Получить информацию о городе по ID

**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Душанбе",
    "latitude": 38.5598,
    "longitude": 68.7870,
    "isActive": true,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

#### `POST /cities/find-by-location`
Найти ближайший город по координатам

**Тело запроса:**
```json
{
  "latitude": 38.5598,
  "longitude": 68.7870
}
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "city": {
      "id": "uuid",
      "name": "Душанбе",
      "latitude": 38.5598,
      "longitude": 68.7870,
      "isActive": true,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    },
    "distance": 0.5
  }
}
```

**Примечание:** `distance` - расстояние в километрах до ближайшего города.

---

### Магазины (публичные)

#### `GET /shops/:id`
Получить информацию о магазине

**Ответ:**
```json
{
  "success": true,
  "data": {
    "shop": {
      "id": "uuid",
      "name": "Название магазина",
      "inn": "123456789",
      "email": "shop@example.com",
      "phone": "+992927781020",
      "avatar": "/images/users/...",
      "productsCount": 100,
      "subscribersCount": 50,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### `GET /shops/:id/products`
Получить товары магазина

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `gender` (string) - "boy", "girl", "unisex"
- `search` (string)
- `categoryId` (uuid)

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Товар",
      "shop": { ... },
      "variations": [ ... ]
    }
  ]
}
```

#### `GET /shops/:id/subscription/check`
Проверить подписку на магазин (требует аутентификации)

**Ответ:**
```json
{
  "success": true,
  "data": {
    "isSubscribed": true
  }
}
```

---

### Продукты (публичные, требуют аутентификации)

#### `GET /products`
Получить список товаров

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `categoryId` (uuid)
- `gender` (string) - "boy", "girl", "unisex"
- `search` (string)
- `minPrice` (float)
- `maxPrice` (float)

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Товар",
      "description": "Описание",
      "categoryId": "uuid",
      "ownerId": "uuid",
      "shop": {
        "id": "uuid",
        "name": "Магазин",
        "inn": "123456789"
      },
      "variations": [
        {
          "id": "uuid",
          "sizes": ["S", "M", "L"],
          "colors": ["Красный", "Синий"],
          "price": 1000.00,
          "originalPrice": 1200.00,
          "discount": 15,
          "imageUrls": ["/images/variations/..."],
          "stockQuantity": 10,
          "isAvailable": true,
          "sku": "SKU123",
          "barcode": "1234567890123"
        }
      ],
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### `GET /products/:id`
Получить товар по ID

**Ответ:** Аналогично элементу массива из `GET /products`

#### `GET /products/featured`
Получить рекомендуемые товары (аналогично `GET /products`)

#### `GET /products/search`
Поиск товаров (аналогично `GET /products`)

#### `GET /products/with-variations`
Получить товары с вариациями (JOIN запрос)

---

### Админские продукты (публичные)

#### `GET /admin/allproducts`
Получить все товары (для админ панели)

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `gender` (string)
- `search` (string)

**Ответ:** Аналогично `GET /products`

#### `GET /admin/products/:id`
Получить товар (админ версия)

---

## 🔒 Защищенные эндпоинты (требуют аутентификации)

### Пользователи

#### `GET /users/profile`
Получить профиль текущего пользователя

**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Иван Иванов",
    "email": "ivan@example.com",
    "phone": "+992927781020",
    "role": { ... },
    "avatar": "/images/users/...",
    "inn": "123456789",
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

#### `PUT /users/profile`
Обновить профиль

**Тело запроса:**
```json
{
  "name": "Новое имя",
  "email": "new@example.com",
  "phone": "+992927781020"
}
```

**Ответ:** Обновленный профиль пользователя

---

### Адреса

#### `GET /users/addresses`
Получить адреса пользователя

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "userId": "uuid",
      "street": "Улица",
      "city": "Город",
      "postalCode": "123456",
      "country": "Таджикистан",
      "isDefault": true,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### `POST /users/addresses`
Создать адрес

**Тело запроса:**
```json
{
  "street": "Улица",
  "city": "Город",
  "postalCode": "123456",
  "country": "Таджикистан",
  "isDefault": false
}
```

#### `PUT /users/addresses/:id`
Обновить адрес

#### `DELETE /users/addresses/:id`
Удалить адрес

#### `PUT /users/addresses/:id/default`
Установить адрес по умолчанию

---

### Корзина

#### `GET /cart`
Получить корзину пользователя

**Ответ:**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "uuid",
        "userId": "uuid",
        "productVariationId": "uuid",
        "quantity": 2,
        "variation": { ... },
        "product": { ... },
        "createdAt": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 2000.00
  }
}
```

#### `POST /cart/items`
Добавить товар в корзину

**Тело запроса:**
```json
{
  "productVariationId": "uuid",
  "quantity": 2
}
```

#### `PUT /cart/items/:id`
Обновить количество товара в корзине

**Тело запроса:**
```json
{
  "quantity": 3
}
```

#### `DELETE /cart/items/:id`
Удалить товар из корзины

#### `DELETE /cart`
Очистить корзину

---

### Избранное

#### `GET /favorites`
Получить избранные товары

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "userId": "uuid",
      "productId": "uuid",
      "product": { ... },
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### `POST /favorites/:productId`
Добавить товар в избранное

#### `DELETE /favorites/:productId`
Удалить товар из избранного

#### `GET /favorites/sync`
Синхронизировать избранное

#### `GET /favorites/:productId/check`
Проверить, есть ли товар в избранном

**Ответ:**
```json
{
  "success": true,
  "data": {
    "isFavorite": true
  }
}
```

---

### Заказы

#### `POST /orders`
Создать заказ

**Тело запроса:**
```json
{
  "items": [
    {
      "productVariationId": "uuid",
      "quantity": 2
    }
  ],
  "addressId": "uuid",
  "paymentMethod": "cash",
  "notes": "Комментарий"
}
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "status": "pending",
    "totalAmount": 2000.00,
    "items": [ ... ],
    "address": { ... },
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

#### `GET /orders`
Получить заказы пользователя

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `status` (string)

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "status": "completed",
      "totalAmount": 2000.00,
      "items": [ ... ],
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### `GET /orders/active`
Получить активный заказ для отслеживания

#### `GET /orders/:id`
Получить заказ по ID

#### `POST /orders/:id/cancel`
Отменить заказ

---

### Гостевые заказы (публичные)

#### `POST /guest-orders`
Создать гостевой заказ

**Тело запроса:**
```json
{
  "name": "Имя",
  "phone": "+992927781020",
  "items": [ ... ],
  "address": { ... },
  "paymentMethod": "cash"
}
```

#### `GET /guest-orders`
Получить гостевой заказ

**Параметры запроса:**
- `orderId` (uuid)
- `phone` (string)

---

### Уведомления

#### `GET /notifications`
Получить уведомления пользователя

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `isRead` (bool)

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "userId": "uuid",
      "title": "Новый заказ",
      "body": "Получен новый заказ",
      "type": "order",
      "timestamp": "2024-01-01T00:00:00Z",
      "isRead": false,
      "actionUrl": "/admin#orders?orderId=uuid",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### `PUT /notifications/:id/read`
Отметить уведомление как прочитанное

#### `PUT /notifications/read-all`
Отметить все уведомления как прочитанные

#### `DELETE /notifications/:id`
Удалить уведомление

#### `GET /notifications/unread-count`
Получить количество непрочитанных уведомлений

**Ответ:**
```json
{
  "success": true,
  "data": {
    "unreadCount": 5
  }
}
```

---

### Токены устройств (для push-уведомлений)

#### `POST /device-tokens`
Зарегистрировать токен устройства

**Тело запроса:**
```json
{
  "token": "fcm_token",
  "platform": "web",
  "deviceId": "device_id"
}
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "token": "fcm_token",
    "platform": "web",
    "deviceId": "device_id",
    "isActive": true,
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

#### `DELETE /device-tokens/:token`
Удалить токен устройства

#### `GET /device-tokens`
Получить токены устройства пользователя

---

### Настройки

#### `GET /settings`
Получить настройки пользователя

**Ответ:**
```json
{
  "success": true,
  "data": {
    "userId": "uuid",
    "language": "ru",
    "theme": "system",
    "notificationsEnabled": true,
    "emailNotifications": true,
    "pushNotifications": true
  }
}
```

#### `PUT /settings`
Обновить настройки

**Тело запроса:**
```json
{
  "language": "ru",
  "theme": "dark",
  "notificationsEnabled": true
}
```

#### `POST /settings/reset`
Сбросить настройки на значения по умолчанию

---

### Подписки на магазины

#### `POST /shops/:id/subscribe`
Подписаться на магазин

**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "shopId": "uuid",
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

#### `DELETE /shops/:id/subscribe`
Отписаться от магазина

#### `GET /shops/:id/subscribers`
Получить список подписчиков магазина (для владельца)

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "userId": "uuid",
      "shopId": "uuid",
      "user": { ... },
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

## 👑 Админские эндпоинты (требуют роль admin или super_admin)

### Пользователи

#### `GET /admin/users`
Получить список пользователей

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `role` (string)
- `search` (string)

#### `POST /admin/users`
Создать пользователя

#### `GET /admin/users/:id`
Получить пользователя по ID

#### `PUT /admin/users/:id`
Обновить пользователя

#### `DELETE /admin/users/:id`
Удалить пользователя

---

### Роли

#### `GET /admin/roles`
Получить список ролей

#### `GET /admin/roles/:id`
Получить роль по ID

#### `POST /admin/roles`
Создать роль

#### `PUT /admin/roles/:id`
Обновить роль

#### `DELETE /admin/roles/:id`
Удалить роль

---

### Уведомления

#### `POST /admin/notifications`
Создать уведомление

**Тело запроса:**
```json
{
  "userId": "uuid",
  "title": "Заголовок",
  "body": "Текст",
  "type": "order",
  "actionUrl": "/admin#orders",
  "data": {}
}
```

---

### Категории

#### `POST /admin/categories`
Создать категорию (только super_admin)

**Тело запроса:**
```json
{
  "name": "Новая категория",
  "description": "Описание",
  "parentId": null,
  "imageUrl": "/images/categories/..."
}
```

#### `PUT /admin/categories/:id`
Обновить категорию

#### `DELETE /admin/categories/:id`
Удалить категорию

---

### Заказы

#### `GET /admin/orders`
Получить все заказы

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `status` (string)
- `search` (string)
- `dateFrom` (date)
- `dateTo` (date)

#### `GET /admin/orders/:id`
Получить заказ по ID

#### `PUT /admin/orders/:id/status`
Обновить статус заказа

**Тело запроса:**
```json
{
  "status": "completed"
}
```

#### `POST /admin/orders/:id/confirm`
Подтвердить заказ

#### `POST /admin/orders/:id/reject`
Отклонить заказ

---

### Продукты

#### `GET /admin/products`
Получить все товары (аналогично `/admin/allproducts`)

---

## 🏪 Эндпоинты для владельцев магазинов (требуют роль shop_owner или admin)

### Товары

#### `GET /shop/products`
Получить товары владельца магазина

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `search` (string)
- `categoryId` (uuid)

#### `POST /shop/products`
Создать товар

**Тело запроса:**
```json
{
  "name": "Новый товар",
  "description": "Описание",
  "categoryId": "uuid",
  "gender": "unisex",
  "variations": [
    {
      "sizes": ["S", "M"],
      "colors": ["Красный"],
      "price": 1000.00,
      "originalPrice": 1200.00,
      "discount": 15,
      "imageUrls": [],
      "stockQuantity": 10,
      "sku": "SKU123",
      "barcode": "1234567890123"
    }
  ]
}
```

#### `PUT /shop/products/:id`
Обновить товар

#### `DELETE /shop/products/:id`
Удалить товар

---

### Заказы

#### `GET /shop/orders`
Получить заказы клиентов владельца

**Параметры запроса:**
- `limit` (int)
- `offset` (int)
- `status` (string)

#### `GET /shop/orders/:id`
Получить заказ клиента

#### `PUT /shop/orders/:id/status`
Обновить статус заказа

---

### Клиенты

#### `GET /shop/customers`
Получить список клиентов владельца

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Имя клиента",
      "email": "client@example.com",
      "phone": "+992927781020",
      "ordersCount": 5,
      "totalSpent": 10000.00
    }
  ]
}
```

#### `GET /shop/customers/:id/orders`
Получить заказы клиента

---

## 🔄 Обновления приложения

### Публичные эндпоинты (без аутентификации)

#### `GET /updates/latest`
Получить последнее активное обновление для платформы

**Параметры запроса:**
- `platform` (string, обязательный) - платформа: `server`, `windows`, `android`

**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "platform": "android",
    "version": "1.0.0",
    "fileName": "android_1.0.0_abc12345.apk",
    "fileUrl": "/updates/android/android_1.0.0_abc12345.apk",
    "fileSize": 15728640,
    "checksumSha256": "abc123def456...",
    "releaseNotes": "Описание изменений",
    "isActive": true,
    "createdAt": "2024-01-01T12:00:00Z",
    "updatedAt": "2024-01-01T12:00:00Z"
  }
}
```

**Ошибки:**
- `400 Bad Request` - если параметр `platform` не указан
- `404 Not Found` - если обновление для указанной платформы не найдено

**Пример использования:**
```
GET /api/v1/updates/latest?platform=android
GET /api/v1/updates/latest?platform=windows
GET /api/v1/updates/latest?platform=server
```

**Скачивание файла обновления:**
После получения информации об обновлении, файл можно скачать по URL из поля `fileUrl`:
```
GET /updates/{platform}/{filename}
```

Например:
```
GET /updates/android/android_1.0.0_abc12345.apk
GET /updates/windows/windows_1.0.0_xyz67890.exe
GET /updates/server/server_1.0.0_def12345.zip
```

**Поддерживаемые платформы и форматы:**
- `server` - Node.js сервер (файл `.zip`)
- `windows` - Flutter Windows приложение (файл `.exe`)
- `android` - Flutter Android приложение (файл `.apk`)

---

### Админские эндпоинты (требуют роль admin или super_admin)

#### `GET /admin/updates`
Получить список всех обновлений

**Параметры запроса:**
- `platform` (string, опционально) - фильтр по платформе: `server`, `windows`, `android`

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "platform": "android",
      "version": "1.0.0",
      "fileName": "android_1.0.0_abc12345.apk",
      "filePath": "updates/android/android_1.0.0_abc12345.apk",
      "fileUrl": "/updates/android/android_1.0.0_abc12345.apk",
      "fileSize": 15728640,
      "checksumSha256": "abc123def456...",
      "releaseNotes": "Описание изменений",
      "isActive": true,
      "createdAt": "2024-01-01T12:00:00Z",
      "updatedAt": "2024-01-01T12:00:00Z"
    }
  ]
}
```

#### `POST /admin/updates/upload`
Загрузить новое обновление

**Формат:** `multipart/form-data`

**Параметры:**
- `platform` (string, обязательный) - платформа: `server`, `windows`, `android`
- `version` (string, обязательный) - версия обновления (например: `1.0.0`)
- `file` (file, обязательный) - файл обновления:
  - `.zip` для `server`
  - `.exe` для `windows`
  - `.apk` для `android`
- `releaseNotes` (string, опционально) - описание изменений

**Ответ:**
```json
{
  "success": true,
  "message": "Update uploaded successfully",
  "data": {
    "id": "uuid",
    "platform": "android",
    "version": "1.0.0",
    "fileName": "android_1.0.0_abc12345.apk",
    "fileUrl": "/updates/android/android_1.0.0_abc12345.apk",
    "fileSize": 15728640,
    "checksumSha256": "abc123def456...",
    "releaseNotes": "Описание изменений",
    "isActive": true,
    "createdAt": "2024-01-01T12:00:00Z",
    "updatedAt": "2024-01-01T12:00:00Z"
  }
}
```

**Ошибки при загрузке:**
- `400 Bad Request` - если `platform` или `version` не указаны
  ```json
  {
    "success": false,
    "error": "platform and version are required"
  }
  ```
- `400 Bad Request` - если указана неверная платформа
  ```json
  {
    "success": false,
    "error": "invalid platform (allowed: server, windows, android)"
  }
  ```
- `400 Bad Request` - если файл не указан
  ```json
  {
    "success": false,
    "error": "file is required",
    "details": "error details"
  }
  ```
- `400 Bad Request` - если расширение файла не поддерживается
  ```json
  {
    "success": false,
    "error": "unsupported extension .pdf (allowed: [.zip .exe .apk])"
  }
  ```
- `500 Internal Server Error` - если не удалось создать директорию для файлов
  ```json
  {
    "success": false,
    "error": "failed to create updates directory",
    "details": "error details"
  }
  ```
- `500 Internal Server Error` - если не удалось сохранить файл
  ```json
  {
    "success": false,
    "error": "failed to save file",
    "details": "error details"
  }
  ```
- `500 Internal Server Error` - если не удалось сохранить метаданные в БД
  ```json
  {
    "success": false,
    "error": "failed to save update metadata",
    "details": "error details"
  }
  ```

**Примечания:**
- При загрузке нового обновления автоматически вычисляется SHA256 хеш файла для проверки целостности
- Файл сохраняется в директории `/app/updates/{platform}/` с уникальным именем
- Все загруженные обновления по умолчанию помечаются как активные (`isActive: true`)

---

## 📤 Загрузка файлов

#### `POST /upload/image`
Загрузить изображение

**Формат:** `multipart/form-data`

**Параметры:**
- `file` (file) - файл изображения
- `folder` (string) - папка: "products", "variations", "categories", "users"

**Ответ:**
```json
{
  "success": true,
  "data": {
    "url": "/images/products/filename.jpg",
    "filename": "filename.jpg"
  }
}
```

#### `DELETE /upload/image/:filename`
Удалить изображение

---

## 🖼️ Работа с изображениями

#### `GET /images/fix-urls`
Исправить URL изображений (админ)

#### `GET /images/url/:filename`
Получить URL изображения

---

## Коды ошибок

- `VALIDATION_ERROR` - Ошибка валидации данных
- `AUTH_REQUIRED` - Требуется аутентификация
- `AUTH_INVALID` - Невалидный токен
- `INVALID_CREDENTIALS` - Неверные учетные данные
- `NOT_FOUND` - Ресурс не найден
- `FORBIDDEN` - Доступ запрещен
- `INTERNAL_ERROR` - Внутренняя ошибка сервера

---

## Статусы заказов

- `pending` - Ожидает подтверждения
- `confirmed` - Подтвержден
- `processing` - В обработке
- `shipped` - Отправлен
- `delivered` - Доставлен
- `completed` - Завершен
- `cancelled` - Отменен
- `rejected` - Отклонен

---

## Типы уведомлений

- `order` - Уведомление о заказе
- `promotion` - Промо-уведомление
- `system` - Системное уведомление
- `reminder` - Напоминание

---

**Последнее обновление:** 2024-01-01  
**Версия API:** 1.3.3

