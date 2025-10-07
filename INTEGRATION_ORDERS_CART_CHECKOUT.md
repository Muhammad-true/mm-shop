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
- `variation_id uuid NOT NULL` ✅ **Новое поле!**
- `quantity bigint NOT NULL`
- `price numeric NOT NULL`
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
    { "variation_id": "var-uuid-1", "quantity": 2, "price": 50.0, "size": "M", "color": "Black", "sku": "SKU-1", "name": "Футболка", "image_url": "https://…" },
    { "variation_id": "var-uuid-2", "quantity": 1, "price": 80.0, "size": "L", "color": "Blue", "sku": "SKU-2", "name": "Худи", "image_url": "https://…" }
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
  'product_id': it.product.id,      // ✅ Основной продукт
  'variation_id': it.variation.id,  // ✅ Конкретная вариация
  'quantity': it.quantity,
  'price': it.variation.price,      // ✅ Цена из вариации
  'sku': it.variation.sku,          // ✅ SKU из вариации
  'name': it.product.name,
  'image_url': it.variation.imageUrls.isNotEmpty ? it.variation.imageUrls.first : null,
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
- [x] **НОВОЕ**: VariationID вместо Size/Color в заказах
- [x] **НОВОЕ**: Гостевые заказы без авторизации
- [x] **НОВОЕ**: Быстрый токен по номеру телефона

---

## 7) 🚀 **Новая логика для Flutter (2024)**

### **7.1 Локальная корзина:**
```dart
class LocalCartItem {
  final String variationId;  // ✅ Основное поле!
  final int quantity;
  final double price;
  final String size;         // ✅ Размер товара
  final String color;        // ✅ Цвет товара
  final String name;
  final String sku;
  final String? imageUrl;
  
  // Размеры и цвета теперь берутся из вариации
  final List<String> sizes;
  final List<String> colors;
}
```

### **7.2 Добавление в корзину:**
```dart
void addToCart(Product product, ProductVariation variation, int quantity) {
  final cartItem = LocalCartItem(
    variationId: variation.id,  // ✅ Связываем с конкретной вариацией
    quantity: quantity,
    price: variation.price,     // ✅ Цена из вариации
    size: variation.size,       // ✅ Размер из вариации
    color: variation.color,     // ✅ Цвет из вариации
    name: product.name,
    sku: variation.sku,
    imageUrl: variation.imageUrls.isNotEmpty ? variation.imageUrls.first : null,
    sizes: variation.sizes,     // ✅ Размеры из вариации
    colors: variation.colors,   // ✅ Цвета из вариации
  );
  
  // Сохранить в локальное хранилище (SharedPreferences/Hive)
  _saveToLocalStorage(cartItem);
}
```

### **7.3 Оформление заказа:**
```dart
Future<Order> createOrder({
  required String recipientName,
  required String phone,
  required String shippingAddr,
  required String paymentMethod,
}) async {
  final localCartItems = _getLocalCartItems();
  
  // Расчет сумм
  final subtotal = localCartItems.fold(0.0, (sum, item) => sum + (item.price * item.quantity));
  final deliveryFee = subtotal >= 200.0 ? 0.0 : 10.0;
  final total = subtotal + deliveryFee;
  
  // Подготовка товаров для API
  final items = localCartItems.map((item) => {
    'variation_id': item.variationId,  // ✅ Основное поле!
    'quantity': item.quantity,
    'price': item.price,
    'size': item.size,
    'color': item.color,
    'sku': item.sku,
    'name': item.name,
    'image_url': item.imageUrl,
  }).toList();
  
  // Создание гостевого заказа (без токена)
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
    _clearLocalCart();
    return Order.fromJson(json.decode(response.body));
  }
  
  throw Exception('Ошибка создания заказа');
}
```

### **7.4 Альтернатива: Заказ с токеном:**
```dart
Future<Order> createOrderWithToken({
  required String recipientName,
  required String phone,
  required String shippingAddr,
  required String paymentMethod,
}) async {
  // 1. Получить быстрый токен
  final token = await _getQuickToken(recipientName, phone);
  
  // 2. Создать заказ с токеном (обычный эндпоинт)
  final response = await http.post(
    Uri.parse('$baseUrl/api/v1/orders'),
    headers: {
      'Authorization': 'Bearer $token',
      'Content-Type': 'application/json',
    },
    body: json.encode({...}), // Те же данные
  );
  
  return Order.fromJson(json.decode(response.body));
}

Future<String> _getQuickToken(String name, String phone) async {
  final response = await http.post(
    Uri.parse('$baseUrl/api/v1/auth/guest-token'),
    headers: {'Content-Type': 'application/json'},
    body: json.encode({
      'name': name,
      'phone': phone,
    }),
  );
  
  final data = json.decode(response.body);
  return data['data']['token'];
}
```

### **7.5 Преимущества новой логики:**
- ✅ **Простота**: Корзина работает полностью локально
- ✅ **Скорость**: Нет задержек на добавление в корзину
- ✅ **Надежность**: Точные данные через VariationID
- ✅ **UX**: Оформление заказа без регистрации
- ✅ **Гибкость**: Можно использовать с токеном или без

### **7.6 Миграция с старой логики:**
1. Заменить `size` и `color` на `variation_id`
2. Обновить структуры данных в корзине
3. Изменить логику добавления в корзину
4. Обновить экраны оформления заказа
5. Тестировать с новыми API эндпоинтами


