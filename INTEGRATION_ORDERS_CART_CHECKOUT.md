## Интеграция Go API и Flutter: Корзина → Оформление → Заказы (статусы)

Документ описывает, как фронт (Flutter) взаимодействует с бэком (Go) для оформления заказов, расчёта доставки в сомони, и отслеживания статусов заказа.

### Валюта и доставка
- Валюта: `TJS`
- Стоимость доставки: `10 с`
- Бесплатная доставка: если сумма товаров (items_subtotal) ≥ `200 с`

Расчёт доставки выполняется на сервере (клиентские значения считаются черновыми).

---

## 1) Схема БД (минимум)

Таблица `orders` (основные колонки):
- `id uuid PK`
- `user_id uuid NOT NULL`
- `status text NOT NULL DEFAULT 'pending'` (pending|confirmed|preparing|inDelivery|delivered|completed|cancelled)
- `total_amount numeric NOT NULL`
- `items_subtotal numeric NOT NULL`
- `delivery_fee numeric NOT NULL DEFAULT 0`
- `currency text NOT NULL DEFAULT 'TJS'`
- `shipping_addr text NOT NULL`
- `payment_method text NOT NULL` (cash|card)
- `payment_status text NOT NULL DEFAULT 'pending'` (pending|paid|failed|refunded)
- `transaction_id text NULL`
- `recipient_name text NOT NULL`
- `phone text NOT NULL`
- `desired_at timestamptz NULL`
- `confirmed_at timestamptz NULL`
- `delivered_at timestamptz NULL`
- `cancelled_at timestamptz NULL`
- `created_at timestamptz NULL`
- `updated_at timestamptz NULL`

Таблица `order_items`:
- `id uuid PK`
- `order_id uuid NOT NULL FK -> orders(id) ON DELETE CASCADE`
- `product_id uuid NOT NULL`
- `quantity bigint NOT NULL`
- `price numeric NOT NULL`
- `size text NULL`
- `color text NULL`
- `name text NOT NULL` (снэпшот)
- `sku text NULL`
- `image_url text NULL`
- `total numeric NOT NULL` (price * quantity)
- `created_at timestamptz NULL`

Рекомендованные индексы: `orders(user_id, created_at DESC)`, `orders(status)`, `order_items(order_id)`.

---

## 2) Эндпоинты API

Все запросы под `/api/v1/…` принимают заголовок:
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

### 2.1 Создать заказ
POST `/api/v1/orders`

Request body (пример):
```json
{
  "recipient_name": "Иван Иванов",
  "phone": "+992 900 00-00-00",
  "shipping_addr": "г. Душанбе, ул. Рудаки, д.1, кв.10",
  "desired_at": "2025-09-24T16:30:00Z",
  "payment_method": "cash",
  "items_subtotal": 180.0,
  "delivery_fee": 10.0,
  "total_amount": 190.0,
  "currency": "TJS",
  "notes": "2 подъезд, код 1234",
  "items": [
    { "product_id": "uuid-1", "quantity": 2, "price": 50.0, "size": "M", "color": "Black", "sku": "SKU-1", "name": "Футболка", "image_url": "https://…" },
    { "product_id": "uuid-2", "quantity": 1, "price": 80.0, "size": "L", "color": "Blue",  "sku": "SKU-2", "name": "Худи",      "image_url": "https://…" }
  ]
}
```

Ответ (успех):
```json
{
  "id": "order-uuid",
  "status": "pending",
  "payment_status": "pending",
  "total_amount": 190.0,
  "items_subtotal": 180.0,
  "delivery_fee": 10.0,
  "currency": "TJS",
  "desired_at": "2025-09-24T16:30:00Z",
  "created_at": "2025-09-21T12:00:00Z",
  "items": [ { "product_id": "uuid-1", "quantity": 2, … }, { … } ]
}
```

Сервер обязан пересчитать `items_subtotal`, `delivery_fee` (10 с, либо 0 если subtotal ≥ 200 с) и `total_amount` на своей стороне.

### 2.2 Список заказов пользователя
GET `/api/v1/orders?status=active|completed`  
Ответ: массив заказов с последним статусом и позициями (или без, по вашему усмотрению).

### 2.3 Детали заказа
GET `/api/v1/orders/{order_id}`

### 2.4 Отмена заказа пользователем
POST `/api/v1/orders/{order_id}/cancel`

### 2.5 Обновление статуса (админ)
PATCH `/api/v1/admin/orders/{order_id}/status`
```json
{ "status": "confirmed|preparing|inDelivery|delivered|completed|cancelled" }
```
При смене статуса сервер может заполнять таймстемпы: `confirmed_at`, `delivered_at`, `cancelled_at`.

### 2.6 Адреса (опционально)
- GET `/api/v1/addresses`
- POST `/api/v1/addresses`
- PATCH `/api/v1/addresses/{id}`
- DELETE `/api/v1/addresses/{id}`
- PATCH `/api/v1/addresses/{id}/default`

---

## 3) Интеграция Flutter

### 3.1 Корзина (CartCubit)
- Считает `subtotal`, доставку и `grandTotal` (см. реализацию в `CartCubit`):
  - доставка: `10 с` или `0 с` при `subtotal ≥ 200`
  - итог: `grandTotal = subtotal + shipping`
- Отображение в `CartScreen`: товары, доставка, итог в сомони (`с`).

### 3.2 Оформление (CheckoutScreen)
- Двухшаговый Stepper:
  1) Контакты/адрес (автоподстановка из `SettingsRepository`, редактируется)
  2) Дата/время доставки + способ оплаты (cash|card) + сводка сумм
- При нажатии «Оформить»:
  - собираем массив `items` из локальной корзины
  - считаем `subtotal`, `delivery_fee`, `total_amount`
  - вызываем `OrdersCubit.createOrder(...)`
  - при успехе: очищаем корзину, показываем Snackbar, редиректим на `/` (главный экран)

Пример сборки тела (псевдокод):
```dart
final items = cartItems.map((it) => {
  'product_id': it.product.id,
  'quantity': it.quantity,
  'price': it.product.price,
  'size': it.selectedSize,
  'color': it.selectedColor,
  'sku': it.product.sku,
  'name': it.product.name,
  'image_url': it.product.imageUrls.isNotEmpty ? it.product.imageUrls.first : null,
}).toList();

final body = {
  'recipient_name': name,
  'phone': phone,
  'shipping_addr': address,
  'desired_at': desiredAt.toUtc().toIso8601String(),
  'payment_method': paymentMethod, // 'cash' | 'card'
  'items_subtotal': subtotal,
  'delivery_fee': shipping,
  'total_amount': grandTotal,
  'currency': 'TJS',
  'items': items,
};
```

### 3.3 Заказы и статусы
- `OrdersCubit.loadActiveOrders()` получает список активных заказов (status != completed/cancelled)
- Таймстемпы статусов (`confirmed_at`, `delivered_at`, `cancelled_at`) используются на экране трекинга для времени метки.

---

## 4) Интеграция Go (эскиз обработчика)

```go
// POST /api/v1/orders
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest // структура, соответствующая body
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil { /* 400 */ }

    // userID из JWT
    userID := r.Context().Value(ctxUserIDKey).(uuid.UUID)

    // пересчёт subtotal/fee/total на сервере
    subtotal := decimal.Zero
    for _, it := range req.Items { subtotal = subtotal.Add(it.Price.Mul(it.Qty)) }
    fee := decimal.NewFromInt(10)
    if subtotal.Cmp(decimal.NewFromInt(200)) >= 0 { fee = decimal.Zero }
    total := subtotal.Add(fee)

    // txn
    // 1) insert into orders(...)
    // 2) insert into order_items(...)
    // 3) select и вернуть заказ c items
}
```

Коды ошибок:
- 400 — валидация
- 401 — неавторизован
- 404 — продукт/заказ не найден
- 409 — конфликт статусов/повторная оплата
- 500 — внутренняя ошибка

---

## 5) Безопасность и аудит
- Все операции с заказами требуют авторизации.
- Сервер проверяет принадлежность заказа пользователю.
- Значения сумм всегда валидируются на бэке.
- При изменении статуса админом заполняются таймстемпы: `confirmed_at`, `delivered_at`, `cancelled_at`.

---

## 6) Чек‑лист интеграции
- [x] Эндпоинты orders: POST/GET/GET/{id}/cancel
- [x] Админ PATCH status
- [x] Расчёт доставки 10 с, бесплатно от 200 с
- [x] Валюта TJS
- [x] Flutter: Cart → Checkout (2 шага) → Orders
- [x] Проброс даты/времени и способа оплаты


