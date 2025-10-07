# MM Shop API Documentation

## 🚀 Общая информация

**Базовый URL:** `http://159.89.99.252:8080`  
**Версия API:** `v1`  
**Формат данных:** `JSON`

## 🔐 Аутентификация

Все защищенные эндпоинты требуют токен в заголовке:
```
Authorization: Bearer {your_token}
```

### Получение токена

#### 1. Вход для зарегистрированных пользователей:
```bash
POST /api/v1/auth/login
```

**Данные для входа админа:**
```json
{
  "email": "admin@mm.com",
  "password": "admin123"
}
```

#### 2. Быстрый вход по номеру телефона:
```bash
POST /api/v1/auth/guest-token
```

**Данные для быстрого входа:**
```json
{
  "name": "Имя пользователя",
  "phone": "+992901234567"
}
```

**Ответ (для обоих методов):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "name": "Имя пользователя",
      "email": "user@example.com",
      "role": "user",
      "isGuest": false
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "message": "Login successful"
}
```

**Особенности быстрого входа:**
- Если пользователь с таким номером телефона не существует - создается автоматически
- Если пользователь уже существует - входит в свой аккаунт
- Email генерируется автоматически (временный)
- Пароль генерируется автоматически
- Токен действует 24 часа
- Пользователь может использовать все функции системы

---

## 📦 Продукты

### 1. Получить все продукты (каталог) - АДМИН
```bash
GET /api/v1/admin/products
```

**Заголовки:** `Authorization: Bearer {admin_token}`

**Параметры запроса:**
- `page` - номер страницы (по умолчанию 1)
- `limit` - количество товаров на странице (по умолчанию 20)
- `search` - поиск по названию или описанию
- `category` - фильтр по категории
- `in_stock` - только товары в наличии (true/false)
- `sort_by` - сортировка (name, price, created_at)
- `sort_order` - порядок сортировки (asc, desc)

**Пример:**
```bash
GET /api/v1/admin/products?page=1&limit=20&search=футболка&in_stock=true
```

**Ответ:**
```json
{
  "products": [
    {
      "id": "uuid",
      "name": "Название товара",
      "description": "Описание",
      "brand": "Бренд",
      "gender": "unisex",
      "categoryId": "uuid",
      "isAvailable": true,
      "ownerId": "uuid",
      "createdAt": "2024-01-15T10:30:00Z",
      "updatedAt": "2024-01-15T10:30:00Z",
      "variations": [
        {
          "id": "uuid",
          "sizes": ["S", "M", "L"],
          "colors": ["Красный"],
          "price": 1500.0,
          "originalPrice": 2000.0,
          "imageUrls": ["url1", "url2"],
          "stockQuantity": 10,
          "isAvailable": true,
          "sku": "SKU123"
        }
      ],
      "owner": {
        "id": "uuid",
        "name": "Название магазина",
        "email": "shop@example.com"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "pages": 5
  }
}
```

### 2. Получить конкретный продукт (карточка товара)
```bash
GET /api/v1/products/:id
```

**Заголовки:** `Authorization: Bearer {token}`

**Ответ:**
```json
{
  "product": {
    "id": "uuid",
    "name": "Название товара",
    "description": "Описание",
    "brand": "Бренд",
    "gender": "unisex",
    "categoryId": "uuid",
    "isAvailable": true,
    "ownerId": "uuid",
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:30:00Z",
    "variations": [
      {
        "id": "uuid",
        "sizes": ["S", "M", "L"],
        "colors": ["Красный"],
        "price": 1500.0,
        "originalPrice": 2000.0,
        "imageUrls": ["url1", "url2"],
        "stockQuantity": 10,
        "isAvailable": true,
        "sku": "SKU123"
      }
    ]
  }
}
```

### 3. Получить продукты с вариациями (JOIN запрос)
```bash
GET /api/v1/products/with-variations
```

**Заголовки:** `Authorization: Bearer {token}`

**Параметры запроса:**
- `page` - номер страницы
- `limit` - количество записей
- `brand` - фильтр по бренду
- `min_price` - минимальная цена
- `max_price` - максимальная цена
- `search` - поиск по названию

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "productId": "uuid",
      "name": "Название товара",
      "description": "Описание",
      "brand": "Бренд",
      "sizes": ["S", "M", "L"],
      "colors": ["Красный"],
      "price": 1500.0,
      "originalPrice": 2000.0,
      "imageUrls": ["url1", "url2"],
      "stockQuantity": 10,
      "sku": "SKU123"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

### 4. Создать продукт (для владельцев магазинов)
```bash
POST /api/v1/shop/products
```

**Заголовки:** `Authorization: Bearer {token}`

**Тело запроса:**
```json
{
  "name": "Название товара",
  "description": "Описание",
  "gender": "unisex",
  "categoryId": "uuid",
  "brand": "Бренд",
  "variations": [
    {
      "sizes": ["S", "M", "L"],
      "colors": ["Красный"],
      "price": 1500.0,
      "originalPrice": 2000.0,
      "imageUrls": ["url1", "url2"],
      "stockQuantity": 10,
      "sku": "SKU123"
    }
  ]
}
```

### 5. Обновить продукт
```bash
PUT /api/v1/shop/products/:id
```

### 6. Удалить продукт
```bash
DELETE /api/v1/shop/products/:id
```

---

## 🏷️ Категории

### 1. Получить все категории
```bash
GET /api/v1/categories
```

**Ответ:**
```json
[
  {
    "id": "uuid",
    "name": "Одежда",
    "description": "Описание категории",
    "parentId": null,
    "isActive": true,
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:30:00Z"
  }
]
```

### 2. Получить конкретную категорию
```bash
GET /api/v1/categories/:id
```

### 3. Получить продукты категории
```bash
GET /api/v1/categories/:id/products
```

### 4. Создать категорию (только супер админ)
```bash
POST /api/v1/admin/categories
```

### 5. Обновить категорию
```bash
PUT /api/v1/admin/categories/:id
```

### 6. Удалить категорию
```bash
DELETE /api/v1/admin/categories/:id
```

---

## 👥 Пользователи

### 1. Регистрация
```bash
POST /api/v1/auth/register
```

**Тело запроса:**
```json
{
  "name": "Имя пользователя",
  "email": "user@example.com",
  "password": "password123"
}
```

### 2. Получить профиль
```bash
GET /api/v1/users/profile
```

### 3. Обновить профиль
```bash
PUT /api/v1/users/profile
```

### 4. Получить всех пользователей (админ)
```bash
GET /api/v1/admin/users
```

### 5. Создать пользователя (админ)
```bash
POST /api/v1/admin/users
```

### 6. Обновить пользователя (админ)
```bash
PUT /api/v1/admin/users/:id
```

### 7. Удалить пользователя (админ)
```bash
DELETE /api/v1/admin/users/:id
```

---

## 🛒 Корзина

### 1. Получить корзину
```bash
GET /api/v1/cart
```

**Ответ:**
```json
{
  "cart": {
    "items": [
      {
        "id": "uuid",
        "quantity": 2,
        "subtotal": 3000.0,
        "product": {
          "id": "uuid",
          "name": "Название товара",
          "description": "Описание",
          "brand": "Бренд",
          "variations": [...]
        },
        "variation": {
          "id": "uuid",
          "sizes": ["S", "M", "L"],
          "colors": ["Красный"],
          "price": 1500.0,
          "originalPrice": 2000.0,
          "imageUrls": ["url1", "url2"],
          "stockQuantity": 10,
          "isAvailable": true,
          "sku": "SKU123"
        },
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total_items": 2,
    "total_price": 3000.0
  }
}
```

### 2. Добавить в корзину
```bash
POST /api/v1/cart/items
```

**Тело запроса:**
```json
{
  "variation_id": "uuid",
  "quantity": 2
}
```

**Описание полей:**
- `variation_id` - ID конкретной вариации продукта (обязательно)
- `quantity` - количество товара (обязательно, больше 0)

### 3. Обновить товар в корзине
```bash
PUT /api/v1/cart/items/:id
```

**Тело запроса:**
```json
{
  "quantity": 3
}
```

**Описание полей:**
- `quantity` - новое количество товара (обязательно, больше 0)

### 4. Удалить из корзины
```bash
DELETE /api/v1/cart/items/:id
```

### 5. Очистить корзину
```bash
DELETE /api/v1/cart
```

---

## ❤️ Избранное

### 1. Получить избранное
```bash
GET /api/v1/favorites
```

### 2. Добавить в избранное
```bash
POST /api/v1/favorites/:productId
```

### 3. Удалить из избранного
```bash
DELETE /api/v1/favorites/:productId
```

### 4. Синхронизировать избранное
```bash
GET /api/v1/favorites/sync
```

### 5. Проверить, в избранном ли товар
```bash
GET /api/v1/favorites/:productId/check
```

---

## 📍 Адреса

### 1. Получить адреса пользователя
```bash
GET /api/v1/users/addresses
```

### 2. Создать адрес
```bash
POST /api/v1/users/addresses
```

### 3. Обновить адрес
```bash
PUT /api/v1/users/addresses/:id
```

### 4. Удалить адрес
```bash
DELETE /api/v1/users/addresses/:id
```

### 5. Установить адрес по умолчанию
```bash
PUT /api/v1/users/addresses/:id/default
```

---

## 🔔 Уведомления

### 1. Получить уведомления
```bash
GET /api/v1/notifications
```

### 2. Отметить как прочитанное
```bash
PUT /api/v1/notifications/:id/read
```

### 3. Отметить все как прочитанные
```bash
PUT /api/v1/notifications/read-all
```

### 4. Удалить уведомление
```bash
DELETE /api/v1/notifications/:id
```

### 5. Получить количество непрочитанных
```bash
GET /api/v1/notifications/unread-count
```

### 6. Создать уведомление (админ)
```bash
POST /api/v1/admin/notifications
```

---

## ⚙️ Настройки

### 1. Получить настройки
```bash
GET /api/v1/settings
```

### 2. Обновить настройки
```bash
PUT /api/v1/settings
```

### 3. Сбросить настройки
```bash
POST /api/v1/settings/reset
```

---

## 🛡️ Роли

### 1. Получить все роли (админ)
```bash
GET /api/v1/admin/roles
```

### 2. Получить конкретную роль (админ)
```bash
GET /api/v1/admin/roles/:id
```

### 3. Создать роль (админ)
```bash
POST /api/v1/admin/roles
```

### 4. Обновить роль (админ)
```bash
PUT /api/v1/admin/roles/:id
```

### 5. Удалить роль (админ)
```bash
DELETE /api/v1/admin/roles/:id
```

---

## 📦 Заказы

### 1. Создать заказ (авторизованный пользователь)
```bash
POST /api/v1/orders
```

**Заголовки:** `Authorization: Bearer {token}`

**Тело запроса:**
```json
{
  "recipient_name": "Имя получателя",
  "phone": "+992901234567",
  "shipping_addr": "Полный адрес доставки",
  "payment_method": "cash",
  "shipping_method": "courier",
  "currency": "TJS",
  "notes": "Комментарий к заказу",
  "desired_date": "2024-12-25",
  "desired_time": "14:30",
  "items": [
    {
      "variation_id": "uuid",
      "quantity": 2,
      "price": 1500.0,
      "size": "L",
      "color": "Красный",
      "sku": "SKU123",
      "name": "Название товара",
      "image_url": "https://example.com/image.jpg"
    }
  ]
}
```

**Описание полей:**
- `recipient_name` - имя получателя (обязательно)
- `phone` - телефон получателя (обязательно)
- `shipping_addr` - адрес доставки (обязательно)
- `payment_method` - способ оплаты: "cash" или "card" (обязательно)
- `shipping_method` - способ доставки (по умолчанию "courier")
- `currency` - валюта (по умолчанию "TJS")
- `notes` - комментарий к заказу
- `desired_date` - желаемая дата доставки (YYYY-MM-DD)
- `desired_time` - желаемое время доставки (HH:mm)
- `items` - массив товаров (обязательно, минимум 1)

**Автоматический расчет:**
- Доставка: 10 TJS (бесплатно от 200 TJS)
- Итоговая сумма пересчитывается на сервере

### 2. Создать гостевой заказ (без авторизации)
```bash
POST /api/v1/guest-orders
```

**Тело запроса:**
```json
{
  "guest_name": "Имя гостя",
  "guest_phone": "+992901234567",
  "shipping_addr": "Полный адрес доставки",
  "payment_method": "cash",
  "shipping_method": "courier",
  "currency": "TJS",
  "notes": "Комментарий к заказу",
  "desired_date": "2024-12-25",
  "desired_time": "14:30",
  "items": [
    {
      "variation_id": "uuid",
      "quantity": 2,
      "price": 1500.0,
      "size": "L",
      "color": "Красный",
      "sku": "SKU123",
      "name": "Название товара",
      "image_url": "https://example.com/image.jpg"
    }
  ]
}
```

**Дополнительные поля для гостей:**
- `guest_name` - имя гостя (обязательно)
- `guest_phone` - телефон гостя (обязательно)

### 3. Получить мои заказы
```bash
GET /api/v1/orders
```

**Заголовки:** `Authorization: Bearer {token}`

### 4. Получить конкретный заказ
```bash
GET /api/v1/orders/:id
```

**Заголовки:** `Authorization: Bearer {token}`

### 5. Отменить заказ
```bash
POST /api/v1/orders/:id/cancel
```

**Заголовки:** `Authorization: Bearer {token}`

### 6. Получить гостевой заказ
```bash
GET /api/v1/guest-orders
```

### 7. Получить заказы (админ)
```bash
GET /api/v1/admin/orders
```

### 8. Получить конкретный заказ (админ)
```bash
GET /api/v1/admin/orders/:id
```

### 9. Обновить статус заказа (админ)
```bash
PUT /api/v1/admin/orders/:id/status
```

### 10. Получить заказы магазина
```bash
GET /api/v1/shop/orders
```

### 11. Получить заказ магазина
```bash
GET /api/v1/shop/orders/:id
```

### 12. Обновить статус заказа магазина
```bash
PUT /api/v1/shop/orders/:id/status
```

---

## 👥 Клиенты магазина

### 1. Получить клиентов магазина
```bash
GET /api/v1/shop/customers
```

### 2. Получить заказы клиента
```bash
GET /api/v1/shop/customers/:id/orders
```

---

## 📤 Загрузка файлов

### 1. Загрузить изображение
```bash
POST /api/v1/upload/image
```

**Content-Type:** `multipart/form-data`

**Параметры:**
- `file` - файл изображения

### 2. Удалить изображение
```bash
DELETE /api/v1/upload/image/:filename
```

---

## 🖼️ Работа с изображениями

### 1. Исправить URL изображений
```bash
GET /api/v1/images/fix-urls
```

### 2. Получить URL изображения
```bash
GET /api/v1/images/url/:filename
```

---

## 🔧 Системные эндпоинты

### 1. Проверка здоровья
```bash
GET /health
```

**Ответ:**
```json
{
  "status": "ok",
  "message": "MM API is running",
  "version": "1.1.0"
}
```

### 2. Информация о версии
```bash
GET /version
```

**Ответ:**
```json
{
  "version": "1.1.0",
  "name": "MM API",
  "build": "development"
}
```

### 3. Главная страница
```bash
GET /
```

**Ответ:**
```json
{
  "message": "Welcome to MM API",
  "version": "1.1.0",
  "docs": "/api/v1/docs",
  "health": "/health"
}
```

### 4. Диагностика БД (админ)
```bash
GET /api/v1/admin/debug/db
```

---

## 📊 Коды ответов

- `200` - Успешно
- `201` - Создано
- `400` - Неверный запрос
- `401` - Не авторизован
- `403` - Доступ запрещен
- `404` - Не найдено
- `409` - Конфликт
- `500` - Внутренняя ошибка сервера

---

## 🔒 Права доступа

### Публичные эндпоинты (без токена):
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/categories/*`

### Пользовательские эндпоинты (требуют токен):
- `GET /api/v1/products/*`
- `GET /api/v1/users/*`
- `GET /api/v1/cart/*`
- `GET /api/v1/favorites/*`
- `GET /api/v1/notifications/*`
- `GET /api/v1/settings/*`

### Админские эндпоинты (требуют роль admin/super_admin):
- `GET /api/v1/admin/*`

### Эндпоинты владельцев магазинов (требуют роль admin/shop_owner):
- `GET /api/v1/shop/*`

---

## 📝 Примеры использования

### Получение всех продуктов с вариациями для Flutter:
```bash
curl -X GET "http://159.89.99.252:8080/api/v1/admin/products?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json"
```

### Получение конкретного продукта:
```bash
curl -X GET "http://159.89.99.252:8080/api/v1/products/PRODUCT_ID" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

### Быстрый вход по номеру телефона:
```bash
curl -X POST "http://159.89.99.252:8080/api/v1/auth/guest-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Имя пользователя",
    "phone": "+992901234567"
  }'
```

### Добавление товара в корзину:
```bash
curl -X POST "http://159.89.99.252:8080/api/v1/cart/items" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "PRODUCT_ID",
    "variation_id": "VARIATION_ID",
    "quantity": 2
  }'
```

### Создание заказа:
```bash
curl -X POST "http://159.89.99.252:8080/api/v1/orders" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient_name": "Ахмад Алиев",
    "phone": "+992901234567",
    "shipping_addr": "ул. Рудаки 123, Душанбе",
    "payment_method": "cash",
    "currency": "TJS",
    "notes": "Позвонить за час до доставки",
    "items": [
      {
        "variation_id": "VARIATION_ID",
        "quantity": 1,
        "price": 200.0,
        "size": "L",
        "color": "Черный",
        "sku": "SKU-123",
        "name": "Название товара",
        "image_url": "https://example.com/image.jpg"
      }
    ]
  }'
```

### Создание гостевого заказа:
```bash
curl -X POST "http://159.89.99.252:8080/api/v1/guest-orders" \
  -H "Content-Type: application/json" \
  -d '{
    "guest_name": "Ахмад Алиев",
    "guest_phone": "+992901234567",
    "shipping_addr": "ул. Рудаки 123, Душанбе",
    "payment_method": "cash",
    "items": [
      {
        "variation_id": "VARIATION_ID",
        "quantity": 1,
        "price": 200.0,
        "size": "L",
        "color": "Черный",
        "sku": "SKU-123",
        "name": "Название товара"
      }
    ]
  }'
```

---

## 🚀 Быстрый старт для Flutter

1. **Получите токен админа:**
```dart
final response = await http.post(
  Uri.parse('http://159.89.99.252:8080/api/v1/auth/login'),
  headers: {'Content-Type': 'application/json'},
  body: json.encode({
    'email': 'admin@mm.com',
    'password': 'admin123',
  }),
);
```

**Или быстрый вход по номеру телефона:**
```dart
final response = await http.post(
  Uri.parse('http://159.89.99.252:8080/api/v1/auth/guest-token'),
  headers: {'Content-Type': 'application/json'},
  body: json.encode({
    'name': 'Имя пользователя',
    'phone': '+992901234567',
  }),
);
```

2. **Получите все продукты:**
```dart
final response = await http.get(
  Uri.parse('http://159.89.99.252:8080/api/v1/admin/products'),
  headers: {
    'Authorization': 'Bearer $token',
    'Content-Type': 'application/json',
  },
);
```

3. **Получите конкретный продукт:**
```dart
final response = await http.get(
  Uri.parse('http://159.89.99.252:8080/api/v1/products/$productId'),
  headers: {
    'Authorization': 'Bearer $token',
    'Content-Type': 'application/json',
  },
);
```

4. **Добавьте товар в корзину:**
```dart
final response = await http.post(
  Uri.parse('http://159.89.99.252:8080/api/v1/cart/items'),
  headers: {
    'Authorization': 'Bearer $token',
    'Content-Type': 'application/json',
  },
  body: json.encode({
    'product_id': productId,
    'variation_id': variationId,
    'quantity': 2,
  }),
);
```

5. **Создайте заказ:**
```dart
final response = await http.post(
  Uri.parse('http://159.89.99.252:8080/api/v1/orders'),
  headers: {
    'Authorization': 'Bearer $token',
    'Content-Type': 'application/json',
  },
  body: json.encode({
    'recipient_name': 'Ахмад Алиев',
    'phone': '+992901234567',
    'shipping_addr': 'ул. Рудаки 123, Душанбе',
    'payment_method': 'cash',
    'currency': 'TJS',
    'items': [
      {
        'product_id': productId,
        'quantity': 1,
        'price': 200.0,
        'sku': 'SKU-123',
        'name': 'Название товара',
      }
    ],
  }),
);
```

6. **Создайте гостевой заказ:**
```dart
final response = await http.post(
  Uri.parse('http://159.89.99.252:8080/api/v1/guest-orders'),
  headers: {
    'Content-Type': 'application/json',
  },
  body: json.encode({
    'guest_name': 'Ахмад Алиев',
    'guest_phone': '+992901234567',
    'shipping_addr': 'ул. Рудаки 123, Душанбе',
    'payment_method': 'cash',
    'items': [
      {
        'product_id': productId,
        'quantity': 1,
        'price': 200.0,
        'sku': 'SKU-123',
        'name': 'Название товара',
      }
    ],
  }),
);
```

---

## 🎯 **Интеграция с Flutter приложением**

### **Логика работы с заказами:**

#### **1. Локальная корзина на клиенте:**
- Пользователь собирает товары в локальной корзине
- Корзина хранится в локальном хранилище (SharedPreferences/Hive)
- Расчет доставки: 10 TJS (бесплатно от 200 TJS)

#### **2. Быстрое оформление заказа:**
```dart
// 1. Получить быстрый токен (если нужен)
final token = await getQuickToken(name, phone);

// 2. Создать заказ из локальной корзины
final order = await createOrderFromLocalCart(
  recipientName: name,
  phone: phone,
  shippingAddr: address,
  paymentMethod: 'cash',
  items: localCartItems, // Из локальной корзины
);
```

#### **3. Создание заказа без токена (гостевой заказ):**
```dart
// Прямо из локальной корзины без авторизации
final order = await createGuestOrder(
  guestName: name,
  guestPhone: phone,
  shippingAddr: address,
  paymentMethod: 'cash',
  items: localCartItems,
);
```

#### **4. Структура товара в заказе:**
```dart
class OrderItem {
  final String productId;
  final String variationId;  // ✅ Новое поле!
  final int quantity;
  final double price;
  final String sku;
  final String name;
  final String? imageUrl;
}
```

### **Flutter Cubit примеры:**

#### **CartCubit (локальная корзина):**
```dart
class CartCubit extends Cubit<CartState> {
  // Добавить в локальную корзину
  void addToLocalCart(Product product, ProductVariation variation, int quantity) {
    final cartItem = CartItem(
      productId: product.id,
      variationId: variation.id,  // ✅ Используем VariationID
      quantity: quantity,
      price: variation.price,
      name: product.name,
      sku: variation.sku,
    );
    // Сохранить в локальное хранилище
  }

  // Расчет доставки
  double calculateDeliveryFee(double subtotal) {
    return subtotal >= 200.0 ? 0.0 : 10.0;
  }

  // Итоговая сумма
  double calculateTotal(double subtotal) {
    return subtotal + calculateDeliveryFee(subtotal);
  }
}
```

#### **OrdersCubit (заказы):**
```dart
class OrdersCubit extends Cubit<OrdersState> {
  // Создать заказ из локальной корзины
  Future<void> createOrderFromLocalCart({
    required String recipientName,
    required String phone,
    required String shippingAddr,
    required String paymentMethod,
    required List<CartItem> localCartItems,
  }) async {
    final items = localCartItems.map((item) => {
      'variation_id': item.variationId,  // ✅ Новое поле!
      'quantity': item.quantity,
      'price': item.price,
      'size': item.size,
      'color': item.color,
      'sku': item.sku,
      'name': item.name,
      'image_url': item.imageUrl,
    }).toList();

    final subtotal = localCartItems.fold(0.0, (sum, item) => sum + (item.price * item.quantity));
    final deliveryFee = subtotal >= 200.0 ? 0.0 : 10.0;
    final total = subtotal + deliveryFee;

    final response = await http.post(
      Uri.parse('$baseUrl/api/v1/guest-orders'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode({
        'guest_name': recipientName,
        'guest_phone': phone,
        'shipping_addr': shippingAddr,
        'payment_method': paymentMethod,
        'items_subtotal': subtotal,
        'delivery_fee': deliveryFee,
        'total_amount': total,
        'currency': 'TJS',
        'items': items,
      }),
    );

    if (response.statusCode == 200) {
      // Очистить локальную корзину
      clearLocalCart();
      // Показать успех
      emit(OrdersSuccess());
    }
  }
}
```

### **Экран оформления заказа (CheckoutScreen):**
```dart
class CheckoutScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return BlocBuilder<CartCubit, CartState>(
      builder: (context, cartState) {
        final cartItems = cartState.cartItems;
        final subtotal = cartState.subtotal;
        final deliveryFee = cartState.deliveryFee;
        final total = cartState.total;

        return Column(
          children: [
            // Товары из локальной корзины
            ...cartItems.map((item) => CartItemWidget(item)),
            
            // Расчеты
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Сумма товаров:'),
                Text('${subtotal.toStringAsFixed(0)} с'),
              ],
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Доставка:'),
                Text(deliveryFee == 0 ? 'Бесплатно' : '${deliveryFee.toStringAsFixed(0)} с'),
              ],
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Итого:', style: TextStyle(fontWeight: FontWeight.bold)),
                Text('${total.toStringAsFixed(0)} с', style: TextStyle(fontWeight: FontWeight.bold)),
              ],
            ),

            // Кнопка оформления
            ElevatedButton(
              onPressed: () => _createOrder(context),
              child: Text('Оформить заказ'),
            ),
          ],
        );
      },
    );
  }

  void _createOrder(BuildContext context) {
    context.read<OrdersCubit>().createOrderFromLocalCart(
      recipientName: _nameController.text,
      phone: _phoneController.text,
      shippingAddr: _addressController.text,
      paymentMethod: _selectedPaymentMethod,
      localCartItems: context.read<CartCubit>().state.cartItems,
    );
  }
}
```

### **Преимущества новой логики:**
- ✅ **Простота**: Корзина работает локально, заказ создается одним запросом
- ✅ **Скорость**: Нет необходимости в токенах для добавления в корзину
- ✅ **Надежность**: Точные данные через VariationID
- ✅ **UX**: Пользователь может оформить заказ без регистрации
- ✅ **Гибкость**: Позже можно привязать заказы к аккаунту

---

*Документация актуальна на версию API 1.1.0*
