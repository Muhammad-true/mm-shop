# API Документация для программы клиента (Мобильное приложение)

## Обзор

Эта документация описывает API эндпоинты для мобильного приложения клиента, которые позволяют:
- Получать список магазинов клиента с информацией о бонусах
- Просматривать детальную информацию о бонусах в конкретном магазине
- Просматривать историю изменений бонусов

## Базовый URL

```
https://your-domain.com/api/v1
```

## Аутентификация

Все эндпоинты требуют авторизации через JWT токен.

**Заголовок запроса:**
```
Authorization: Bearer <ваш_jwt_токен>
```

Токен получается при регистрации или входе через эндпоинты:
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

## Эндпоинты

### 1. Получить список магазинов клиента с бонусами

Возвращает список всех магазинов, в которых клиент зарегистрирован, с информацией о бонусах.

**Endpoint:** `GET /shops/my`

**Тип:** Защищенный (требует авторизации)

**Query параметры:** Нет

**Пример запроса:**

```bash
curl -X GET https://your-domain.com/api/v1/shops/my \
  -H "Authorization: Bearer ваш_jwt_токен"
```

**Успешный ответ (200 OK):**

```json
{
  "success": true,
  "message": "Список магазинов получен",
  "data": [
    {
      "id": "789e4567-e89b-12d3-a456-426614174001",
      "shopId": "123e4567-e89b-12d3-a456-426614174000",
      "phone": "+992901234567",
      "qrCode": "SHOP1-CLIENT-001",
      "bonusAmount": 250,
      "firstBonusDate": "2024-01-15T10:30:00Z",
      "shop": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Мой Магазин",
        "inn": "123456789"
      },
      "createdAt": "2024-01-15T10:30:00Z",
      "updatedAt": "2024-01-15T11:45:00Z"
    },
    {
      "id": "789e4567-e89b-12d3-a456-426614174002",
      "shopId": "456e4567-e89b-12d3-a456-426614174000",
      "phone": "+992901234567",
      "qrCode": "SHOP2-CLIENT-001",
      "bonusAmount": 75,
      "firstBonusDate": "2024-01-20T14:20:00Z",
      "shop": {
        "id": "456e4567-e89b-12d3-a456-426614174000",
        "name": "Другой Магазин",
        "inn": "987654321"
      },
      "createdAt": "2024-01-20T14:20:00Z",
      "updatedAt": "2024-01-22T09:15:00Z"
    }
  ]
}
```

**Ошибки:**

**401 Unauthorized** - Пользователь не авторизован:
```json
{
  "success": false,
  "error": {
    "code": "AUTH_REQUIRED",
    "message": "Пользователь не авторизован"
  }
}
```

**400 Bad Request** - У пользователя не указан номер телефона:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "У пользователя не указан номер телефона"
  }
}
```

**Пустой список:** Если клиент еще не зарегистрирован ни в одном магазине, вернется пустой массив:
```json
{
  "success": true,
  "message": "Список магазинов получен",
  "data": []
}
```

---

### 2. Получить информацию о бонусах в конкретном магазине

Возвращает детальную информацию о бонусах клиента в указанном магазине.

**Endpoint:** `GET /shops/:id/bonus`

**Тип:** Защищенный (требует авторизации)

**URL параметры:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `id` | string (UUID) | ID магазина |

**Пример запроса:**

```bash
curl -X GET https://your-domain.com/api/v1/shops/123e4567-e89b-12d3-a456-426614174000/bonus \
  -H "Authorization: Bearer ваш_jwt_токен"
```

**Успешный ответ (200 OK):**

```json
{
  "success": true,
  "message": "Информация о бонусах получена",
  "data": {
    "id": "789e4567-e89b-12d3-a456-426614174001",
    "shopId": "123e4567-e89b-12d3-a456-426614174000",
    "phone": "+992901234567",
    "qrCode": "SHOP1-CLIENT-001",
    "bonusAmount": 250,
    "firstBonusDate": "2024-01-15T10:30:00Z",
    "shop": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Мой Магазин",
      "inn": "123456789"
    },
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T11:45:00Z"
  }
}
```

**Ошибки:**

**404 Not Found** - Клиент не найден в этом магазине:
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Клиент не найден в этом магазине"
  }
}
```

**400 Bad Request** - Неверный формат shop_id:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Неверный формат shop_id"
  }
}
```

---

### 3. Получить историю изменений бонусов

Возвращает историю всех изменений бонусов клиента в указанном магазине с поддержкой пагинации.

**Endpoint:** `GET /shops/:id/bonus/history`

**Тип:** Защищенный (требует авторизации)

**URL параметры:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `id` | string (UUID) | ID магазина |

**Query параметры:**

| Параметр | Тип | Обязательный | По умолчанию | Описание |
|----------|-----|--------------|--------------|-----------|
| `page` | integer | Нет | 1 | Номер страницы |
| `limit` | integer | Нет | 20 | Количество записей на странице (максимум 100) |

**Пример запроса:**

```bash
curl -X GET "https://your-domain.com/api/v1/shops/123e4567-e89b-12d3-a456-426614174000/bonus/history?page=1&limit=20" \
  -H "Authorization: Bearer ваш_jwt_токен"
```

**Успешный ответ (200 OK):**

```json
{
  "success": true,
  "message": "История бонусов получена",
  "data": {
    "history": [
      {
        "id": "abc12345-e89b-12d3-a456-426614174001",
        "shopClientId": "789e4567-e89b-12d3-a456-426614174001",
        "previousAmount": 220,
        "newAmount": 250,
        "changeAmount": 30,
        "createdAt": "2024-01-15T11:45:00Z"
      },
      {
        "id": "abc12345-e89b-12d3-a456-426614174002",
        "shopClientId": "789e4567-e89b-12d3-a456-426614174001",
        "previousAmount": 150,
        "newAmount": 220,
        "changeAmount": 70,
        "createdAt": "2024-01-15T11:00:00Z"
      },
      {
        "id": "abc12345-e89b-12d3-a456-426614174003",
        "shopClientId": "789e4567-e89b-12d3-a456-426614174001",
        "previousAmount": 100,
        "newAmount": 150,
        "changeAmount": 50,
        "createdAt": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 3,
    "page": 1,
    "limit": 20
  }
}
```

**Структура записи истории:**

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | string (UUID) | ID записи истории |
| `shopClientId` | string (UUID) | ID связи клиента с магазином |
| `previousAmount` | integer | Количество бонусов до изменения |
| `newAmount` | integer | Количество бонусов после изменения |
| `changeAmount` | integer | Изменение (положительное = начисление, отрицательное = списание) |
| `createdAt` | string (ISO 8601) | Время изменения |

**Ошибки:**

**404 Not Found** - Клиент не найден в этом магазине:
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Клиент не найден в этом магазине"
  }
}
```

---

## Автоматическое связывание с магазинами

**Важная особенность:** Когда клиент регистрируется в приложении по номеру телефона, система автоматически находит все магазины, где этот номер телефона был зарегистрирован ранее (даже до регистрации в приложении), и связывает их с аккаунтом пользователя.

**Процесс:**

1. Магазин регистрирует клиента по номеру телефона через POS систему
2. Клиент еще не зарегистрирован в мобильном приложении
3. Клиент регистрируется в приложении с тем же номером телефона
4. При первом запросе `GET /shops/my` система автоматически:
   - Находит все записи `ShopClient` с этим номером телефона
   - Связывает их с аккаунтом пользователя (устанавливает `userId`)
   - Возвращает список всех магазинов с бонусами

## Пример интеграции (React Native / Flutter)

### React Native (JavaScript)

```javascript
// Конфигурация API
const API_BASE_URL = 'https://your-domain.com/api/v1';
const getAuthHeaders = () => ({
  'Authorization': `Bearer ${getStoredToken()}`,
  'Content-Type': 'application/json'
});

// Получить список магазинов клиента
async function getMyShops() {
  try {
    const response = await fetch(`${API_BASE_URL}/shops/my`, {
      method: 'GET',
      headers: getAuthHeaders()
    });
    
    const data = await response.json();
    
    if (data.success) {
      return data.data;
    } else {
      throw new Error(data.error.message);
    }
  } catch (error) {
    console.error('Ошибка получения магазинов:', error);
    throw error;
  }
}

// Получить информацию о бонусах в магазине
async function getShopBonusInfo(shopId) {
  try {
    const response = await fetch(`${API_BASE_URL}/shops/${shopId}/bonus`, {
      method: 'GET',
      headers: getAuthHeaders()
    });
    
    const data = await response.json();
    
    if (data.success) {
      return data.data;
    } else {
      throw new Error(data.error.message);
    }
  } catch (error) {
    console.error('Ошибка получения информации о бонусах:', error);
    throw error;
  }
}

// Получить историю бонусов
async function getBonusHistory(shopId, page = 1, limit = 20) {
  try {
    const response = await fetch(
      `${API_BASE_URL}/shops/${shopId}/bonus/history?page=${page}&limit=${limit}`,
      {
        method: 'GET',
        headers: getAuthHeaders()
      }
    );
    
    const data = await response.json();
    
    if (data.success) {
      return data.data;
    } else {
      throw new Error(data.error.message);
    }
  } catch (error) {
    console.error('Ошибка получения истории бонусов:', error);
    throw error;
  }
}

// Использование
async function loadMyShops() {
  try {
    const shops = await getMyShops();
    console.log('Мои магазины:', shops);
    
    if (shops.length > 0) {
      const firstShop = shops[0];
      
      // Получаем детальную информацию о бонусах
      const bonusInfo = await getShopBonusInfo(firstShop.shopId);
      console.log('Информация о бонусах:', bonusInfo);
      
      // Получаем историю бонусов
      const history = await getBonusHistory(firstShop.shopId);
      console.log('История бонусов:', history);
    }
  } catch (error) {
    console.error('Ошибка:', error);
  }
}
```

### Flutter (Dart)

```dart
import 'package:http/http.dart' as http;
import 'dart:convert';

class BonusService {
  static const String baseUrl = 'https://your-domain.com/api/v1';
  
  String? _token;
  
  void setToken(String token) {
    _token = token;
  }
  
  Map<String, String> get _headers => {
    'Authorization': 'Bearer $_token',
    'Content-Type': 'application/json',
  };
  
  // Получить список магазинов клиента
  Future<List<ShopClient>> getMyShops() async {
    final response = await http.get(
      Uri.parse('$baseUrl/shops/my'),
      headers: _headers,
    );
    
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      if (data['success'] == true) {
        return (data['data'] as List)
            .map((item) => ShopClient.fromJson(item))
            .toList();
      } else {
        throw Exception(data['error']['message']);
      }
    } else {
      throw Exception('Ошибка получения магазинов: ${response.statusCode}');
    }
  }
  
  // Получить информацию о бонусах в магазине
  Future<ShopClient> getShopBonusInfo(String shopId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/shops/$shopId/bonus'),
      headers: _headers,
    );
    
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      if (data['success'] == true) {
        return ShopClient.fromJson(data['data']);
      } else {
        throw Exception(data['error']['message']);
      }
    } else {
      throw Exception('Ошибка получения информации о бонусах: ${response.statusCode}');
    }
  }
  
  // Получить историю бонусов
  Future<BonusHistoryResponse> getBonusHistory(
    String shopId, {
    int page = 1,
    int limit = 20,
  }) async {
    final response = await http.get(
      Uri.parse('$baseUrl/shops/$shopId/bonus/history?page=$page&limit=$limit'),
      headers: _headers,
    );
    
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      if (data['success'] == true) {
        return BonusHistoryResponse.fromJson(data['data']);
      } else {
        throw Exception(data['error']['message']);
      }
    } else {
      throw Exception('Ошибка получения истории бонусов: ${response.statusCode}');
    }
  }
}
```

## Рекомендации по использованию

1. **Кэширование:** Кэшируйте список магазинов на устройстве и обновляйте при необходимости
2. **Обновление данных:** Реализуйте механизм периодического обновления информации о бонусах
3. **Обработка ошибок:** Всегда обрабатывайте ошибки сети и авторизации
4. **Пагинация:** При отображении истории бонусов используйте пагинацию для больших списков
5. **Офлайн режим:** Сохраняйте данные локально для работы в офлайн режиме

## Отображение данных в UI

### Экран "Мои магазины"

Покажите список магазинов с:
- Названием магазина (`shop.name`)
- Текущим количеством бонусов (`bonusAmount`)

**При нажатии на магазин** - открывается bottom sheet (нижняя панель) с:
- QR кодом (`qrCode`) - для сканирования в магазине
- Текущим количеством бонусов (`bonusAmount`)
- Датой первого получения бонусов (`firstBonusDate`)
- Кнопкой "История бонусов"

### Bottom Sheet деталей магазина

При открытии bottom sheet покажите:
- QR код (большой размер для сканирования)
- Текущее количество бонусов
- Дата первого получения бонусов
- Кнопку "История бонусов"

### Экран истории бонусов

Покажите список изменений с:
- Датой и временем (`createdAt`)
- Предыдущим количеством (`previousAmount`)
- Новым количеством (`newAmount`)
- Изменением (`changeAmount`) - выделите цветом (зеленый для начисления, красный для списания)

## Вопросы и поддержка

При возникновении вопросов или проблем обращайтесь к администратору системы.

