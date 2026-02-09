# Промпт для Flutter Админ Панели

## Общее описание

Необходимо создать Flutter приложение админ-панели для управления интернет-магазином MM Shop. Приложение должно поддерживать три роли: `super_admin`, `admin`, `shop_owner`. Все запросы требуют JWT токен в заголовке `Authorization: Bearer <token>`.

## Базовый URL API

- **Development**: `http://localhost:8080` (или пустая строка для same-origin)
- **Production**: `https://api.libiss.com`

## Структура ответов API

Все ответы имеют стандартный формат:
```json
{
  "success": true/false,
  "data": { ... },
  "message": "опциональное сообщение"
}
```

При ошибках:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Описание ошибки"
  }
}
```

---

## 1. АУТЕНТИФИКАЦИЯ

### 1.1 Вход в систему
**Эндпоинт:** `POST /api/v1/auth/login`  
**Токен:** Не требуется  
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
    "token": "jwt-token",
    "user": {
      "id": "uuid",
      "name": "Имя",
      "email": "email@example.com",
      "phone": "+992927781020",
      "role": {
        "name": "admin|super_admin|shop_owner",
        "displayName": "Администратор"
      }
    }
  }
}
```
**Когда используется:** При входе в приложение. Сохранить токен и роль пользователя в secure storage.

### 1.2 Получение профиля текущего пользователя
**Эндпоинт:** `GET /api/v1/users/profile`  
**Токен:** Обязателен  
**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Имя",
    "email": "email@example.com",
    "phone": "+992927781020",
    "role": { ... },
    "addresses": [ ... ],
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```
**Когда используется:** При инициализации приложения для проверки валидности токена и получения актуальной информации о пользователе.

---

## 2. ДАШБОРД

### 2.1 Получение статистики (для админа)
**Эндпоинт:** `GET /api/v1/products/` (для товаров)  
**Эндпоинт:** `GET /api/v1/admin/users/` (для пользователей)  
**Эндпоинт:** `GET /api/v1/admin/orders/` (для заказов)  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** При открытии дашборда для отображения счетчиков (товары, пользователи, заказы, доход).

**Ответ товаров:**
```json
{
  "success": true,
  "products": [ ... ]
}
```

**Ответ пользователей:**
```json
{
  "success": true,
  "data": {
    "users": [ ... ]
  }
}
```

**Ответ заказов:**
```json
{
  "success": true,
  "data": {
    "orders": [ ... ],
    "pagination": { ... }
  }
}
```

### 2.2 Получение статистики (для владельца магазина)
**Эндпоинт:** `GET /api/v1/products/` (товары, фильтруются по ownerId на клиенте)  
**Эндпоинт:** `GET /api/v1/shops/:shopId/subscribers` (подписчики)  
**Эндпоинт:** `GET /api/v1/shop/orders/` (заказы магазина)  
**Токен:** Обязателен  
**Роль:** `shop_owner`  
**Когда используется:** При открытии дашборда владельцем магазина.

**Ответ подписчиков:**
```json
{
  "success": true,
  "data": {
    "subscribers": [ ... ]
  }
}
```

### 2.3 Уведомления
**Эндпоинт:** `GET /api/v1/notifications/?limit=10&isRead=false`  
**Эндпоинт:** `GET /api/v1/notifications/unread-count`  
**Токен:** Обязателен  
**Когда используется:** На дашборде для отображения списка непрочитанных уведомлений и счетчика.

**Ответ списка:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "title": "Заголовок",
      "body": "Текст уведомления",
      "type": "order|promotion|system|reminder",
      "isRead": false,
      "timestamp": "2024-01-01T00:00:00Z",
      "actionUrl": "/orders?orderId=uuid"
    }
  ]
}
```

**Ответ счетчика:**
```json
{
  "success": true,
  "data": {
    "unreadCount": 5
  }
}
```

### 2.4 Отметить уведомление как прочитанное
**Эндпоинт:** `PUT /api/v1/notifications/:id/read`  
**Токен:** Обязателен  
**Когда используется:** При клике на уведомление.

### 2.5 Отметить все уведомления как прочитанные
**Эндпоинт:** `PUT /api/v1/notifications/read-all`  
**Токен:** Обязателен  
**Когда используется:** При нажатии кнопки "Отметить все прочитанными" на дашборде.

---

## 3. ТОВАРЫ

### 3.1 Список товаров
**Эндпоинт:** `GET /api/v1/products/`  
**Токен:** Обязателен  
**Роль:** Все роли  
**Параметры запроса (опционально):**
- `limit` (int)
- `offset` (int)
- `categoryId` (uuid)
- `search` (string)

**Когда используется:** 
- На вкладке "Товары" для отображения списка
- Для админа: все товары
- Для shop_owner: фильтровать по `ownerId` на клиенте

**Ответ:**
```json
{
  "success": true,
  "products": [
    {
      "id": "uuid",
      "name": "Название товара",
      "brand": "Бренд",
      "category": { "id": "uuid", "name": "Категория" },
      "gender": "male|female|unisex",
      "variations": [ ... ],
      "ownerId": "uuid"
    }
  ]
}
```

### 3.2 Детали товара
**Эндпоинт:** `GET /api/v1/products/:id`  
**Токен:** Обязателен  
**Когда используется:** При просмотре деталей товара или вариаций.

**Ответ:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Название",
    "variations": [
      {
        "id": "uuid",
        "sizes": ["S", "M", "L"],
        "colors": ["Красный", "Синий"],
        "price": 1000,
        "discount": 10,
        "stockQuantity": 50,
        "sku": "SKU123",
        "barcode": "123456789",
        "imageUrls": ["/images/variations/..."]
      }
    ]
  }
}
```

### 3.3 Создание товара
**Эндпоинт:** `POST /api/v1/shop/products/`  
**Токен:** Обязателен  
**Роль:** `shop_owner` или `admin`  
**Тело запроса:**
```json
{
  "name": "Название товара",
  "brand": "Бренд",
  "categoryId": "uuid",
  "gender": "unisex",
  "description": "Описание",
  "variations": [
    {
      "sizes": ["S", "M"],
      "colors": ["Красный"],
      "price": 1000,
      "discount": 0,
      "stockQuantity": 10,
      "sku": "SKU123",
      "barcode": "123456789",
      "imageUrls": []
    }
  ]
}
```
**Когда используется:** При создании нового товара через форму.

### 3.4 Обновление товара
**Эндпоинт:** `PUT /api/v1/shop/products/:id/`  
**Токен:** Обязателен  
**Роль:** `shop_owner` или `admin`  
**Тело запроса:** Аналогично созданию  
**Когда используется:** При редактировании товара.

### 3.5 Удаление товара
**Эндпоинт:** `DELETE /api/v1/shop/products/:id/`  
**Токен:** Обязателен  
**Роль:** `shop_owner` или `admin`  
**Когда используется:** При удалении товара из списка.

---

## 4. КАТЕГОРИИ

### 4.1 Список категорий
**Эндпоинт:** `GET /api/v1/categories/`  
**Токен:** Обязателен  
**Когда используется:** 
- На вкладке "Категории" для отображения списка
- В форме создания/редактирования товара для выбора категории

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Название",
      "description": "Описание",
      "iconUrl": "/images/categories/...",
      "parent": { "id": "uuid", "name": "Родительская" },
      "children": [ ... ],
      "sortOrder": 0,
      "isActive": true
    }
  ]
}
```

### 4.2 Создание категории
**Эндпоинт:** `POST /api/v1/admin/categories/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Тело запроса (FormData):**
- `name` (string, обязательное)
- `description` (string)
- `parentId` (uuid, опционально)
- `icon` (file, PNG, опционально)
- `sortOrder` (int, опционально)
- `isActive` (bool, опционально)

**Когда используется:** При создании новой категории.

### 4.3 Обновление категории
**Эндпоинт:** `PUT /api/v1/admin/categories/:id/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Тело запроса:** Аналогично созданию  
**Когда используется:** При редактировании категории.

### 4.4 Удаление категории
**Эндпоинт:** `DELETE /api/v1/admin/categories/:id/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** При удалении категории.

---

## 5. ПОЛЬЗОВАТЕЛИ

### 5.1 Список пользователей
**Эндпоинт:** `GET /api/v1/admin/users/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** На вкладке "Пользователи" для отображения списка.

**Ответ:**
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "uuid",
        "name": "Имя",
        "email": "email@example.com",
        "phone": "+992...",
        "role": { "name": "user", "displayName": "Пользователь" },
        "isActive": true,
        "createdAt": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 5.2 Создание пользователя
**Эндпоинт:** `POST /api/v1/admin/users/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Тело запроса:**
```json
{
  "name": "Имя",
  "email": "email@example.com",
  "password": "password123",
  "phone": "+992...",
  "roleId": "uuid",
  "shop": {
    "name": "Название магазина",
    "inn": "123456789",
    "description": "Описание",
    "address": "Адрес",
    "cityId": "uuid"
  }
}
```
**Когда используется:** При создании нового пользователя. Если роль `shop_owner`, можно передать данные магазина.

### 5.3 Обновление пользователя
**Эндпоинт:** `PUT /api/v1/admin/users/:id/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Тело запроса:** Аналогично созданию (все поля опциональны)  
**Когда используется:** При редактировании пользователя.

### 5.4 Удаление пользователя
**Эндпоинт:** `DELETE /api/v1/admin/users/:id/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** При удалении пользователя.

### 5.5 Детали пользователя
**Эндпоинт:** `GET /api/v1/admin/users/:id`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** При просмотре деталей пользователя.

---

## 6. РОЛИ

### 6.1 Список ролей
**Эндпоинт:** `GET /api/v1/admin/roles/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** На вкладке "Роли" для отображения списка и в форме создания/редактирования пользователя.

**Ответ:**
```json
{
  "success": true,
  "data": {
    "roles": [
      {
        "id": "uuid",
        "name": "admin",
        "displayName": "Администратор",
        "description": "Описание",
        "permissions": "{}",
        "isActive": true,
        "userCount": 5
      }
    ]
  }
}
```

### 6.2 Создание роли
**Эндпоинт:** `POST /api/v1/admin/roles/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Тело запроса:**
```json
{
  "name": "moderator",
  "displayName": "Модератор",
  "description": "Описание",
  "permissions": "{\"dashboard\": true, \"users\": true}"
}
```
**Когда используется:** При создании новой роли.

### 6.3 Обновление роли
**Эндпоинт:** `PUT /api/v1/admin/roles/:id/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Тело запроса:** Аналогично созданию  
**Когда используется:** При редактировании роли.

### 6.4 Удаление роли
**Эндпоинт:** `DELETE /api/v1/admin/roles/:id/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** При удалении роли (системные роли нельзя удалить).

---

## 7. ЗАКАЗЫ

### 7.1 Список заказов (админ)
**Эндпоинт:** `GET /api/v1/admin/orders/`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Параметры запроса:**
- `page` (int, по умолчанию 1)
- `limit` (int, по умолчанию 20)
- `status` (string, опционально: pending, confirmed, preparing, inDelivery, delivered, completed, cancelled)
- `search` (string, поиск по телефону или имени)
- `dateFrom` (date, опционально)
- `dateTo` (date, опционально)

**Когда используется:** На вкладке "Заказы" для админа.

**Ответ:**
```json
{
  "success": true,
  "data": {
    "orders": [
      {
        "id": "uuid",
        "order_number": "ORD-001",
        "user_id": "uuid",
        "user": { "name": "Имя", "is_guest": false },
        "shop_owner": { "name": "Магазин", "phone": "+992..." },
        "phone": "+992...",
        "recipient_name": "Имя получателя",
        "status": "pending",
        "total_amount": 1000,
        "order_items": [ ... ],
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "pages": 5
    },
    "stats": { ... }
  }
}
```

### 7.2 Список заказов (владелец магазина)
**Эндпоинт:** `GET /api/v1/shop/orders/`  
**Токен:** Обязателен  
**Роль:** `shop_owner`  
**Параметры запроса:** Аналогично админскому эндпоинту  
**Когда используется:** На вкладке "Заказы" для владельца магазина (только заказы его магазина).

### 7.3 Детали заказа
**Эндпоинт:** `GET /api/v1/admin/orders/:id` или `GET /api/v1/shop/orders/:id`  
**Токен:** Обязателен  
**Когда используется:** При просмотре деталей заказа.

### 7.4 Обновление статуса заказа
**Эндпоинт:** `PUT /api/v1/admin/orders/:id/status` или `PUT /api/v1/shop/orders/:id/status`  
**Токен:** Обязателен  
**Тело запроса:**
```json
{
  "status": "confirmed|preparing|inDelivery|delivered|completed|cancelled"
}
```
**Когда используется:** При изменении статуса заказа.

---

## 8. МАГАЗИНЫ И ЛИЦЕНЗИИ

### 8.1 Список магазинов с лицензиями
**Эндпоинт:** `GET /api/v1/admin/shops`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Параметры запроса:**
- `page` (int, по умолчанию 1)
- `limit` (int, по умолчанию 50)
- `search` (string, поиск по названию)
- `hasLicense` (bool, фильтр по наличию лицензии)

**Когда используется:** На вкладке "Магазины" для управления магазинами и их лицензиями.

**Ответ:**
```json
{
  "success": true,
  "data": {
    "shops": [
      {
        "id": "uuid",
        "name": "Название магазина",
        "email": "shop@example.com",
        "owner": { "name": "Владелец", "email": "owner@example.com" },
        "productsCount": 100,
        "subscribersCount": 50,
        "hasLicense": true,
        "license": {
          "id": "uuid",
          "licenseKey": "XXXX-XXXX-XXXX-XXXX",
          "subscriptionType": "monthly|yearly|lifetime|trial",
          "subscriptionStatus": "active|expired|cancelled|pending",
          "activatedAt": "2024-01-01T00:00:00Z",
          "expiresAt": "2024-02-01T00:00:00Z",
          "daysRemaining": 20,
          "isValid": true,
          "isExpired": false
        }
      }
    ],
    "pagination": { ... }
  }
}
```

### 8.2 Создание лицензии для магазина
**Эндпоинт:** `POST /api/v1/admin/licenses/shops/:shopId/generate`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Тело запроса:**
```json
{
  "subscriptionType": "monthly|yearly|lifetime",
  "paymentAmount": 0,
  "paymentCurrency": "USD",
  "paymentProvider": "manual",
  "paymentTransactionId": "",
  "autoRenew": false,
  "notes": "Примечания"
}
```
**Когда используется:** При создании лицензии для магазина без лицензии.

### 8.3 Продление лицензии
**Эндпоинт:** `PUT /api/v1/admin/licenses/:licenseId/extend`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Тело запроса:**
```json
{
  "months": 1
}
```
**Когда используется:** При продлении существующей лицензии.

### 8.4 Просмотр лицензии
**Эндпоинт:** `GET /api/v1/admin/licenses/:licenseId`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** При просмотре деталей лицензии.

### 8.5 Удаление лицензии
**Эндпоинт:** `DELETE /api/v1/admin/licenses/:licenseId`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** При удалении лицензии.

---

## 9. ОБНОВЛЕНИЯ

### 9.1 Список обновлений
**Эндпоинт:** `GET /api/v1/admin/updates`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Когда используется:** На вкладке "Обновления" для просмотра истории обновлений.

**Ответ:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "platform": "server|windows|android",
      "version": "1.0.0",
      "fileName": "update.zip",
      "fileUrl": "/uploads/updates/...",
      "fileSize": 1024000,
      "checksumSha256": "hash",
      "releaseNotes": "Описание изменений",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 9.2 Загрузка обновления
**Эндпоинт:** `POST /api/v1/admin/updates/upload`  
**Токен:** Обязателен  
**Роль:** `super_admin` или `admin`  
**Тело запроса (FormData):**
- `platform` (string: server, windows, android)
- `version` (string, обязательное)
- `releaseNotes` (string)
- `file` (file: .zip для server, .exe для windows, .apk для android)

**Когда используется:** При загрузке нового обновления через форму.

---

## 10. ЗАГРУЗКА ИЗОБРАЖЕНИЙ

### 10.1 Загрузка изображения
**Эндпоинт:** `POST /api/v1/upload/image`  
**Токен:** Обязателен  
**Тело запроса (FormData):**
- `file` (file, обязательное)
- `folder` (string, опционально: products, variations, categories, users)

**Когда используется:** 
- При загрузке изображений для товаров/вариаций
- При загрузке иконки категории
- При загрузке аватара пользователя

**Ответ:**
```json
{
  "success": true,
  "data": {
    "url": "/images/variations/filename.jpg",
    "filename": "filename.jpg"
  }
}
```

### 10.2 Удаление изображения
**Эндпоинт:** `DELETE /api/v1/upload/image/:filename`  
**Токен:** Обязателен  
**Когда используется:** При удалении загруженного изображения.

---

## 11. HEALTH CHECK

### 11.1 Проверка доступности API
**Эндпоинт:** `GET /api/health`  
**Токен:** Не требуется  
**Когда используется:** При проверке подключения к API (кнопка "Проверить подключение" на форме входа).

**Ответ:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

---

## ВАЖНЫЕ ЗАМЕЧАНИЯ

1. **Авторизация:** Все запросы (кроме login и health) требуют JWT токен в заголовке `Authorization: Bearer <token>`.

2. **Роли и доступ:**
   - `super_admin` и `admin`: полный доступ ко всем функциям
   - `shop_owner`: доступ только к своим товарам, заказам своего магазина, подписчикам

3. **Фильтрация данных:**
   - Для `shop_owner` товары фильтруются на клиенте по `ownerId`
   - Заказы для `shop_owner` возвращаются уже отфильтрованными с сервера

4. **Пагинация:** Большинство списков поддерживают пагинацию через параметры `page` и `limit`.

5. **Обработка ошибок:** При 401 Unauthorized нужно перенаправить на экран входа и очистить токен.

6. **Изображения:** URL изображений могут быть относительными (`/images/...`) или абсолютными. Нужно обрабатывать оба случая.

7. **Формат дат:** Все даты в формате ISO 8601 (`2024-01-01T00:00:00Z`).

8. **Валидация:** Проверять обязательные поля перед отправкой запросов.

9. **Кэширование:** Токен и роль пользователя хранить в secure storage (flutter_secure_storage).

10. **Обновление токена:** При истечении токена (401) нужно перенаправить на экран входа.

