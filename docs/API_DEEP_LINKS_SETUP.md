# Настройка API для Deep Links и Sharing

## Обзор

Приложение Libiss поддерживает deep links для прямого перехода к товарам и магазинам с QR-кодами. При переходе по ссылке пользователь попадает напрямую на нужный экран.

## Форматы ссылок

### 1. Ссылка на товар
```
https://shop.libiss.com/product-details/{productId}
```

**Пример:**
```
https://shop.libiss.com/product-details/12345
```

**Поведение:**
- Открывает экран деталей товара
- Работает на вебе и в мобильном приложении (через App Links/Universal Links)
- Если приложение не установлено, открывается веб-версия

### 2. Ссылка на магазин с QR-кодом
```
https://shop.libiss.com/my-shops/{shopId}
```

**Пример:**
```
https://shop.libiss.com/my-shops/67890
```

**Поведение:**
- Требует авторизации пользователя
- Если пользователь не авторизован:
  - Показывается экран входа/регистрации
  - После успешной авторизации автоматически открывается экран магазина с QR-кодом
- Если пользователь авторизован:
  - Сразу открывается экран магазина с QR-кодом

## Что нужно сделать на API

### 1. Настроить редиректы на сервере

Для корректной работы deep links на вебе необходимо настроить редиректы:

#### Nginx конфигурация (пример)
```nginx
# Редирект для товаров
location /product-details/ {
    try_files $uri $uri/ /index.html;
}

# Редирект для магазинов
location /my-shops/ {
    try_files $uri $uri/ /index.html;
}
```

#### Apache конфигурация (пример)
```apache
<IfModule mod_rewrite.c>
  RewriteEngine On
  RewriteBase /
  RewriteRule ^product-details/.*$ /index.html [L]
  RewriteRule ^my-shops/.*$ /index.html [L]
</IfModule>
```

### 2. API для получения данных товара

**Endpoint:** `GET /api/v1/products/{productId}/public`

**Примечание:** Это публичный endpoint, не требует авторизации и не фильтрует по владельцу. Используется для deep links и sharing.

**Ответ должен содержать:**
```json
{
  "id": "12345",
  "name": "Офисная рубашка H&M",
  "description": "Описание товара",
  "brand": "H&M",
  "price": 2999.00,
  "images": [
    "https://example.com/image1.jpg",
    "https://example.com/image2.jpg"
  ],
  "variations": [
    {
      "id": "var1",
      "colors": ["Белый", "Синий"],
      "sizes": ["M", "L", "XL"],
      "price": 2999.00,
      "stockQuantity": 10,
      "isAvailable": true
    }
  ],
  "categoryId": "cat1",
  "gender": "male"
}
```

### 3. API для получения данных магазина с QR-кодом

**Endpoint:** `GET /api/client-shops/{shopId}/bonus-info`

**Требования:**
- Требует авторизации (Bearer token в заголовке)
- Возвращает данные магазина и QR-код пользователя

**Ответ должен содержать:**
```json
{
  "shopId": "67890",
  "shop": {
    "id": "67890",
    "name": "Название магазина",
    "avatarUrl": "https://example.com/shop-logo.jpg"
  },
  "qrCode": "USER_SHOP_QR_CODE_STRING",
  "bonusAmount": 150.50
}
```

**Ошибки:**
- `401 Unauthorized` - если пользователь не авторизован
- `404 Not Found` - если магазин не найден или не связан с пользователем

### 4. API для sharing товаров

При sharing товара используется ссылка формата:
```
https://shop.libiss.com/product-details/{productId}
```

**Рекомендации:**
- Убедитесь, что все товары имеют уникальный ID
- Ссылки должны быть доступны без авторизации (для товаров)
- Для магазинов ссылки требуют авторизации

### 5. Open Graph мета-теги для веба (опционально)

Для лучшего отображения ссылок в социальных сетях добавьте мета-теги:

```html
<!-- Для товара -->
<meta property="og:title" content="Офисная рубашка H&M">
<meta property="og:description" content="Описание товара">
<meta property="og:image" content="https://example.com/product-image.jpg">
<meta property="og:url" content="https://shop.libiss.com/product-details/12345">
<meta property="og:type" content="product">

<!-- Для магазина -->
<meta property="og:title" content="Магазин - Название">
<meta property="og:description" content="Ваш QR-код для получения бонусов">
<meta property="og:image" content="https://example.com/shop-logo.jpg">
<meta property="og:url" content="https://shop.libiss.com/my-shops/67890">
```

## Проверка работы

### Тестирование на вебе:
1. Откройте `https://shop.libiss.com/product-details/12345` в браузере
2. Должен открыться экран деталей товара

### Тестирование deep links:
1. Установите приложение на устройство
2. Откройте ссылку `https://shop.libiss.com/product-details/12345` в браузере
3. Должно появиться предложение открыть в приложении
4. После выбора приложения должен открыться экран товара

### Тестирование магазина с авторизацией:
1. Откройте `https://shop.libiss.com/my-shops/67890` без авторизации
2. Должен появиться экран входа
3. После входа должен автоматически открыться экран магазина с QR-кодом

## Дополнительные рекомендации

1. **Кэширование:** Кэшируйте данные товаров для быстрой загрузки
2. **Аналитика:** Отслеживайте переходы по deep links для аналитики
3. **Безопасность:** Проверяйте права доступа для магазинов (пользователь должен иметь доступ только к своим магазинам)
4. **Ошибки:** Обрабатывайте случаи, когда товар/магазин не найден (404)

## Примеры использования

### Sharing товара из приложения:
```dart
final productUrl = 'https://shop.libiss.com/product-details/${product.id}';
await Share.share('Посмотри этот товар: $productUrl');
```

### Открытие магазина из приложения:
```dart
final shopUrl = 'https://shop.libiss.com/my-shops/${shopId}';
// При переходе по ссылке автоматически проверится авторизация
```

## Поддержка

Если возникли вопросы по настройке, обратитесь к документации:
- [Android App Links](https://developer.android.com/training/app-links)
- [iOS Universal Links](https://developer.apple.com/documentation/xcode/allowing-apps-and-websites-to-link-to-your-content)
- [Flutter Deep Links](https://docs.flutter.dev/development/ui/navigation/deep-linking)

