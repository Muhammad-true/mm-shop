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

**Ответ:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "name": "Администратор",
      "email": "admin@mm.com",
      "role": "admin"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "message": "Login successful"
}
```

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
  "product_id": "uuid",
  "variation_id": "uuid",
  "quantity": 2
}
```

**Описание полей:**
- `product_id` - ID продукта (обязательно)
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

### 1. Получить заказы (админ)
```bash
GET /api/v1/admin/orders
```

### 2. Получить конкретный заказ (админ)
```bash
GET /api/v1/admin/orders/:id
```

### 3. Обновить статус заказа (админ)
```bash
PUT /api/v1/admin/orders/:id/status
```

### 4. Получить заказы магазина
```bash
GET /api/v1/shop/orders
```

### 5. Получить заказ магазина
```bash
GET /api/v1/shop/orders/:id
```

### 6. Обновить статус заказа магазина
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

---

*Документация актуальна на версию API 1.1.0*
