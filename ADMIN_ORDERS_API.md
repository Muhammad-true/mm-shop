# 📦 API для управления заказами в админ панели

## Обзор

Расширенный API для владельцев магазина (админов) для управления заказами с полной информацией о клиентах, поиском, фильтрацией и быстрыми действиями.

---

## 🔐 Авторизация

Все эндпоинты требуют:
- Bearer Token администратора
- Роль: `admin` или `super_admin`

---

## 📋 Эндпоинты

### 1. Получить все заказы с фильтрами

**GET** `/api/v1/admin/orders/`

#### Query параметры:

| Параметр | Тип | Описание | Пример |
|----------|-----|----------|--------|
| `page` | number | Номер страницы (default: 1) | `?page=1` |
| `limit` | number | Количество на странице (default: 20) | `?limit=50` |
| `status` | string | Фильтр по статусу | `?status=pending` |
| `search` | string | Поиск по имени, телефону, номеру заказа | `?search=Иван` |
| `order_number` | string | Поиск по номеру заказа (первые 8 символов) | `?order_number=a1b2c3d4` |
| `phone` | string | Поиск по телефону | `?phone=+992927` |
| `date_from` | string | Дата от (YYYY-MM-DD) | `?date_from=2025-10-01` |
| `date_to` | string | Дата до (YYYY-MM-DD) | `?date_to=2025-10-31` |

#### Примеры запросов:

```bash
# Все заказы (первая страница)
GET /api/v1/admin/orders/

# Заказы со статусом "pending"
GET /api/v1/admin/orders/?status=pending

# Поиск по телефону
GET /api/v1/admin/orders/?phone=+992927781020

# Поиск по имени
GET /api/v1/admin/orders/?search=Иван

# Поиск по номеру заказа
GET /api/v1/admin/orders/?order_number=a1b2c3d4

# Фильтр по дате
GET /api/v1/admin/orders/?date_from=2025-10-01&date_to=2025-10-31

# Комбинированный запрос
GET /api/v1/admin/orders/?status=pending&page=1&limit=50
```

#### Пример ответа:

```json
{
  "success": true,
  "message": "Заказы получены успешно",
  "data": {
    "orders": [
      {
        "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "order_number": "a1b2c3d4",
        "status": "pending",
        "total_amount": 350.50,
        "items_subtotal": 340.50,
        "delivery_fee": 10.00,
        "currency": "TJS",
        "shipping_address": "ул. Ленина, д. 10, кв. 5",
        "payment_method": "cash",
        "shipping_method": "courier",
        "payment_status": "pending",
        "recipient_name": "Иван Иванов",
        "phone": "+992927781020",
        "notes": "Позвоните за 10 минут до доставки",
        "desired_at": "2025-10-25T14:00:00Z",
        "confirmed_at": null,
        "cancelled_at": null,
        "created_at": "2025-10-24T10:00:00Z",
        "updated_at": "2025-10-24T10:00:00Z",
        "order_items": [
          {
            "id": "uuid",
            "quantity": 2,
            "price": 150.0,
            "size": "M",
            "color": "Красный",
            "variation_id": "uuid",
            "subtotal": 300.0,
            "variation": {
              "id": "uuid",
              "sku": "SKU123",
              "price": 150.0,
              "size": "M",
              "color": "Красный",
              "stock": 50
            }
          }
        ],
        "user": {
          "id": "user-uuid",
          "name": "Иван Иванов",
          "phone": "+992927781020",
          "email": "ivan@example.com",
          "is_guest": false
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "totalPages": 3
    },
    "stats": {
      "pending": 12,
      "confirmed": 8,
      "preparing": 5,
      "inDelivery": 3,
      "delivered": 10,
      "completed": 5,
      "cancelled": 2
    }
  }
}
```

---

### 2. Подтвердить заказ

**POST** `/api/v1/admin/orders/:id/confirm`

Быстрый метод для подтверждения заказа. Устанавливает статус `confirmed` и время подтверждения.

#### Пример запроса:

```bash
POST /api/v1/admin/orders/a1b2c3d4-e5f6-7890-abcd-ef1234567890/confirm
Authorization: Bearer {admin_token}
```

**Body:** Не требуется

#### Пример ответа:

```json
{
  "success": true,
  "message": "Заказ подтвержден",
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "order_number": "a1b2c3d4",
    "status": "confirmed",
    "confirmed_at": "2025-10-24T15:30:00Z",
    ...
  }
}
```

---

### 3. Отклонить заказ

**POST** `/api/v1/admin/orders/:id/reject`

Быстрый метод для отклонения заказа. Устанавливает статус `cancelled` и время отмены.

#### Пример запроса:

```bash
POST /api/v1/admin/orders/a1b2c3d4-e5f6-7890-abcd-ef1234567890/reject
Authorization: Bearer {admin_token}
```

**Body:** Не требуется

#### Пример ответа:

```json
{
  "success": true,
  "message": "Заказ отклонен",
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "order_number": "a1b2c3d4",
    "status": "cancelled",
    "cancelled_at": "2025-10-24T15:30:00Z",
    ...
  }
}
```

---

### 4. Обновить статус заказа (универсальный)

**PUT** `/api/v1/admin/orders/:id/status`

Универсальный метод для изменения статуса заказа на любой.

#### Body:

```json
{
  "status": "preparing"
}
```

**Доступные статусы:**
- `pending` - Ожидает подтверждения
- `confirmed` - Подтвержден
- `preparing` - Готовится
- `inDelivery` - В доставке
- `delivered` - Доставлен
- `completed` - Завершен
- `cancelled` - Отменен

#### Пример запроса:

```bash
PUT /api/v1/admin/orders/a1b2c3d4-e5f6-7890-abcd-ef1234567890/status
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "status": "inDelivery"
}
```

---

### 5. Получить один заказ

**GET** `/api/v1/admin/orders/:id`

Получить детальную информацию об одном заказе.

#### Пример запроса:

```bash
GET /api/v1/admin/orders/a1b2c3d4-e5f6-7890-abcd-ef1234567890
Authorization: Bearer {admin_token}
```

---

## 📊 Статистика

В ответе на список заказов всегда включается статистика по всем статусам:

```json
"stats": {
  "pending": 12,      // Ожидают подтверждения
  "confirmed": 8,     // Подтверждены
  "preparing": 5,     // Готовятся
  "inDelivery": 3,    // В доставке
  "delivered": 10,    // Доставлены
  "completed": 5,     // Завершены
  "cancelled": 2      // Отменены
}
```

---

## 🔍 Примеры использования

### JavaScript (для админ панели)

```javascript
// Получить все заказы
async function getOrders(page = 1, filters = {}) {
  const params = new URLSearchParams({
    page: page,
    limit: 20,
    ...filters
  });
  
  const response = await fetch(`/api/v1/admin/orders/?${params}`, {
    headers: {
      'Authorization': `Bearer ${adminToken}`,
      'Content-Type': 'application/json'
    }
  });
  
  return await response.json();
}

// Поиск по телефону
const orders = await getOrders(1, { phone: '+992927781020' });

// Фильтр по статусу
const pendingOrders = await getOrders(1, { status: 'pending' });

// Общий поиск
const searchResults = await getOrders(1, { search: 'Иван' });

// Подтвердить заказ
async function confirmOrder(orderId) {
  const response = await fetch(`/api/v1/admin/orders/${orderId}/confirm`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${adminToken}`,
      'Content-Type': 'application/json'
    }
  });
  
  return await response.json();
}

// Отклонить заказ
async function rejectOrder(orderId) {
  const response = await fetch(`/api/v1/admin/orders/${orderId}/reject`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${adminToken}`,
      'Content-Type': 'application/json'
    }
  });
  
  return await response.json();
}

// Изменить статус
async function updateOrderStatus(orderId, status) {
  const response = await fetch(`/api/v1/admin/orders/${orderId}/status`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${adminToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ status })
  });
  
  return await response.json();
}
```

---

## 🎯 Ключевые особенности

✅ **Полная информация о клиенте** - имя, телефон, email, тип пользователя  
✅ **Мощный поиск** - по номеру заказа, имени, телефону  
✅ **Гибкие фильтры** - по статусу, датам  
✅ **Статистика** - количество заказов по каждому статусу  
✅ **Быстрые действия** - подтвердить/отклонить одним запросом  
✅ **Короткий номер заказа** - первые 8 символов UUID для удобства  
✅ **Пагинация** - для работы с большим количеством заказов  
✅ **Информация о товарах** - полные данные о позициях заказа  

---

## 🔒 Безопасность

- Все эндпоинты требуют авторизации
- Доступ только для ролей `admin` и `super_admin`
- Проверка прав выполняется middleware

