# Задача для Flutter: Карточка товара с вариациями

## Описание задачи

Нужно создать экран карточки товара в Flutter приложении, где пользователь может:
1. Выбрать цвет товара
2. Выбрать размер товара  
3. Видеть цену, которая меняется в зависимости от выбранных параметров
4. Просматривать изображения товара

## Получение токена админа

**Сервер:** `159.89.99.252:8080`

**Эндпоинт для входа:** `POST /api/v1/auth/login`

**Данные админа:**
```json
{
  "email": "admin@mm.com",
  "password": "admin123"
}
```

**Пример запроса для получения токена:**
```bash
curl -X POST http://159.89.99.252:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@mm.com",
    "password": "admin123"
  }'
```

**Ответ сервера:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "name": "Admin",
      "email": "admin@mm.com",
      "role": "admin"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "message": "Login successful"
}
```

## API Эндпоинты для продуктов

### 1. Получение конкретного продукта (карточка товара)

**URL:** `GET /api/v1/products/:id`

**Заголовки:**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Пример запроса:**
```bash
GET http://159.89.99.252:8080/api/v1/products/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 2. Получение всех продуктов (каталог)

**URL:** `GET /api/v1/admin/products`

**Заголовки:**
```
Authorization: Bearer {admin_token}
Content-Type: application/json
```

**⚠️ ВАЖНО:** Этот эндпоинт возвращает **ВСЕ продукты независимо от магазина/владельца**. Используйте токен админа для доступа ко всем товарам в системе.

**Пример запроса:**
```bash
GET http://159.89.99.252:8080/api/v1/admin/products
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Параметры запроса:**
- `page` - номер страницы (по умолчанию 1)
- `limit` - количество товаров на странице (по умолчанию 20)
- `search` - поиск по названию или описанию
- `category` - фильтр по категории
- `in_stock` - только товары в наличии (true/false)
- `sort_by` - сортировка (name, price, created_at)
- `sort_order` - порядок сортировки (asc, desc)

**Пример с параметрами:**
```bash
GET http://159.89.99.252:8080/api/v1/admin/products?page=1&limit=20&search=футболка&in_stock=true&sort_by=price&sort_order=asc
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Структура ответа для каталога:**
```json
{
  "products": [
    {
      "id": "uuid",
      "name": "Название товара",
      "description": "Описание",
      "brand": "Бренд",
      "variations": [...],
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

## Структура JSON ответа

```json
{
  "product": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Футболка Nike Air Max",
    "description": "Удобная футболка из хлопка с технологией Dri-FIT",
    "brand": "Nike",
    "gender": "unisex",
    "categoryId": "456e7890-e89b-12d3-a456-426614174001",
    "isAvailable": true,
    "ownerId": "789e0123-e89b-12d3-a456-426614174002",
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:30:00Z",
    "variations": [
      {
        "id": "var-001",
        "sizes": ["S", "M", "L"],
        "colors": ["Красный"],
        "price": 1500.0,
        "originalPrice": 2000.0,
        "imageUrls": [
          "https://example.com/images/red-shirt-1.jpg",
          "https://example.com/images/red-shirt-2.jpg"
        ],
        "stockQuantity": 10,
        "isAvailable": true,
        "sku": "NKE-RED-SML"
      },
      {
        "id": "var-002", 
        "sizes": ["M", "L", "XL"],
        "colors": ["Синий"],
        "price": 1600.0,
        "originalPrice": 2100.0,
        "imageUrls": [
          "https://example.com/images/blue-shirt-1.jpg",
          "https://example.com/images/blue-shirt-2.jpg"
        ],
        "stockQuantity": 5,
        "isAvailable": true,
        "sku": "NKE-BLUE-MXL"
      },
      {
        "id": "var-003",
        "sizes": ["S", "M"],
        "colors": ["Черный"],
        "price": 1400.0,
        "originalPrice": 1800.0,
        "imageUrls": [
          "https://example.com/images/black-shirt-1.jpg"
        ],
        "stockQuantity": 0,
        "isAvailable": false,
        "sku": "NKE-BLACK-SM"
      }
    ]
  }
}
```

## Модели данных для Flutter

### Product Model
```dart
class Product {
  final String id;
  final String name;
  final String description;
  final String brand;
  final String gender;
  final String categoryId;
  final bool isAvailable;
  final String? ownerId;
  final DateTime createdAt;
  final DateTime updatedAt;
  final List<ProductVariation> variations;

  Product({
    required this.id,
    required this.name,
    required this.description,
    required this.brand,
    required this.gender,
    required this.categoryId,
    required this.isAvailable,
    this.ownerId,
    required this.createdAt,
    required this.updatedAt,
    required this.variations,
  });

  factory Product.fromJson(Map<String, dynamic> json) {
    return Product(
      id: json['id'],
      name: json['name'],
      description: json['description'],
      brand: json['brand'],
      gender: json['gender'],
      categoryId: json['categoryId'],
      isAvailable: json['isAvailable'],
      ownerId: json['ownerId'],
      createdAt: DateTime.parse(json['createdAt']),
      updatedAt: DateTime.parse(json['updatedAt']),
      variations: (json['variations'] as List)
          .map((v) => ProductVariation.fromJson(v))
          .toList(),
    );
  }
}
```

### ProductVariation Model
```dart
class ProductVariation {
  final String id;
  final List<String> sizes;
  final List<String> colors;
  final double price;
  final double? originalPrice;
  final List<String> imageUrls;
  final int stockQuantity;
  final bool isAvailable;
  final String sku;

  ProductVariation({
    required this.id,
    required this.sizes,
    required this.colors,
    required this.price,
    this.originalPrice,
    required this.imageUrls,
    required this.stockQuantity,
    required this.isAvailable,
    required this.sku,
  });

  factory ProductVariation.fromJson(Map<String, dynamic> json) {
    return ProductVariation(
      id: json['id'],
      sizes: List<String>.from(json['sizes']),
      colors: List<String>.from(json['colors']),
      price: (json['price'] as num).toDouble(),
      originalPrice: json['originalPrice'] != null 
          ? (json['originalPrice'] as num).toDouble() 
          : null,
      imageUrls: List<String>.from(json['imageUrls']),
      stockQuantity: json['stockQuantity'],
      isAvailable: json['isAvailable'],
      sku: json['sku'],
    );
  }
}
```

## Логика работы UI

### 1. Получение токена админа (с кэшированием)
```dart
class AuthService {
  static String? _cachedToken;
  static DateTime? _tokenExpiry;
  
  static Future<String> getAdminToken() async {
    // Проверяем, есть ли действующий токен в кэше
    if (_cachedToken != null && _tokenExpiry != null && DateTime.now().isBefore(_tokenExpiry!)) {
      return _cachedToken!;
    }
    
    // Получаем новый токен
    final response = await http.post(
      Uri.parse('http://159.89.99.252:8080/api/v1/auth/login'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode({
        'email': 'admin@mm.com',
        'password': 'admin123',
      }),
    );
    
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      _cachedToken = data['data']['token'];
      // Токен действует 24 часа (настраивается на сервере)
      _tokenExpiry = DateTime.now().add(Duration(hours: 23));
      return _cachedToken!;
    } else {
      throw Exception('Failed to get admin token: ${response.statusCode}');
    }
  }
  
  static void clearToken() {
    _cachedToken = null;
    _tokenExpiry = null;
  }
}
```

### 2. Получение всех продуктов (каталог)
```dart
Future<List<Product>> fetchAllProducts({
  int page = 1,
  int limit = 20,
  String? search,
  String? category,
  bool? inStock,
  String sortBy = 'created_at',
  String sortOrder = 'desc',
}) async {
  final token = await AuthService.getAdminToken();
  
  String url = 'http://159.89.99.252:8080/api/v1/admin/products?page=$page&limit=$limit&sort_by=$sortBy&sort_order=$sortOrder';
  
  if (search != null) url += '&search=$search';
  if (category != null) url += '&category=$category';
  if (inStock != null) url += '&in_stock=$inStock';
  
  final response = await http.get(
    Uri.parse(url),
    headers: {
      'Authorization': 'Bearer $token',
      'Content-Type': 'application/json',
    },
  );
  
  if (response.statusCode == 200) {
    final data = json.decode(response.body);
    final products = (data['products'] as List)
        .map((p) => Product.fromJson(p))
        .toList();
    return products;
  } else {
    throw Exception('Failed to load products: ${response.statusCode}');
  }
}
```

### 3. Получение конкретного продукта (карточка товара)
```dart
Future<Product> fetchProduct(String productId) async {
  final token = await AuthService.getAdminToken();
  
  final response = await http.get(
    Uri.parse('http://159.89.99.252:8080/api/v1/products/$productId'),
    headers: {
      'Authorization': 'Bearer $token',
      'Content-Type': 'application/json',
    },
  );
  
  if (response.statusCode == 200) {
    final data = json.decode(response.body);
    return Product.fromJson(data['product']);
  } else {
    throw Exception('Failed to load product: ${response.statusCode}');
  }
}
```

### 2. Группировка вариаций по цветам
```dart
Map<String, List<ProductVariation>> groupVariationsByColor(List<ProductVariation> variations) {
  Map<String, List<ProductVariation>> grouped = {};
  
  for (var variation in variations) {
    for (var color in variation.colors) {
      if (!grouped.containsKey(color)) {
        grouped[color] = [];
      }
      grouped[color]!.add(variation);
    }
  }
  
  return grouped;
}
```

### 3. Получение доступных размеров для выбранного цвета
```dart
List<String> getAvailableSizesForColor(String selectedColor, Map<String, List<ProductVariation>> groupedVariations) {
  List<String> sizes = [];
  
  if (groupedVariations.containsKey(selectedColor)) {
    for (var variation in groupedVariations[selectedColor]!) {
      if (variation.isAvailable) {
        sizes.addAll(variation.sizes);
      }
    }
  }
  
  return sizes.toSet().toList(); // Убираем дубликаты
}
```

### 4. Получение цены для выбранной комбинации
```dart
double? getPriceForSelection(String selectedColor, String selectedSize, Map<String, List<ProductVariation>> groupedVariations) {
  if (groupedVariations.containsKey(selectedColor)) {
    for (var variation in groupedVariations[selectedColor]!) {
      if (variation.isAvailable && variation.sizes.contains(selectedSize)) {
        return variation.price;
      }
    }
  }
  return null;
}
```

### 5. Получение изображений для выбранной комбинации
```dart
List<String> getImagesForSelection(String selectedColor, String selectedSize, Map<String, List<ProductVariation>> groupedVariations) {
  if (groupedVariations.containsKey(selectedColor)) {
    for (var variation in groupedVariations[selectedColor]!) {
      if (variation.isAvailable && variation.sizes.contains(selectedSize)) {
        return variation.imageUrls;
      }
    }
  }
  return [];
}
```

## Пример UI состояния

```dart
class ProductCardState extends State<ProductCard> {
  Product? product;
  String? selectedColor;
  String? selectedSize;
  double? currentPrice;
  List<String> currentImages = [];
  List<String> availableSizes = [];
  
  @override
  void initState() {
    super.initState();
    loadProduct();
  }
  
  void loadProduct() async {
    try {
      final productData = await fetchProduct(widget.productId);
      setState(() {
        product = productData;
        // Автоматически выбираем первый доступный цвет
        if (product!.variations.isNotEmpty) {
          selectedColor = product!.variations.first.colors.first;
          updateAvailableSizes();
        }
      });
    } catch (e) {
      // Обработка ошибки
    }
  }
  
  void onColorSelected(String color) {
    setState(() {
      selectedColor = color;
      selectedSize = null; // Сбрасываем размер при смене цвета
      updateAvailableSizes();
    });
  }
  
  void onSizeSelected(String size) {
    setState(() {
      selectedSize = size;
      updatePriceAndImages();
    });
  }
  
  void updateAvailableSizes() {
    if (selectedColor != null && product != null) {
      final grouped = groupVariationsByColor(product!.variations);
      availableSizes = getAvailableSizesForColor(selectedColor!, grouped);
    }
  }
  
  void updatePriceAndImages() {
    if (selectedColor != null && selectedSize != null && product != null) {
      final grouped = groupVariationsByColor(product!.variations);
      currentPrice = getPriceForSelection(selectedColor!, selectedSize!, grouped);
      currentImages = getImagesForSelection(selectedColor!, selectedSize!, grouped);
    }
  }
}
```

## Требования к UI

1. **Отображение основной информации:**
   - Название товара
   - Бренд
   - Описание
   - Рейтинг (если есть)

2. **Выбор цвета:**
   - Кнопки или чипы с названиями цветов
   - Визуальное выделение выбранного цвета
   - Показывать только доступные цвета

3. **Выбор размера:**
   - Кнопки с размерами
   - Показывать только размеры, доступные для выбранного цвета
   - Визуальное выделение выбранного размера

4. **Отображение цены:**
   - Текущая цена (жирным шрифтом)
   - Зачеркнутая оригинальная цена (если есть скидка)
   - Цена должна обновляться при выборе параметров

5. **Галерея изображений:**
   - Показывать изображения для выбранной комбинации
   - Возможность пролистывания
   - Увеличение по тапу

6. **Кнопка добавления в корзину:**
   - Активна только при выборе цвета и размера
   - Показывать количество на складе
   - Блокировать если товар недоступен

## Зависимости для Flutter

Добавьте в `pubspec.yaml`:

```yaml
dependencies:
  flutter:
    sdk: flutter
  http: ^1.1.0
  shared_preferences: ^2.2.2  # для кэширования токена
  cached_network_image: ^3.3.0  # для кэширования изображений
```

## Дополнительные соображения

1. **Обработка ошибок:**
   - Нет интернета
   - Товар не найден
   - Ошибка сервера (401 - неверный токен, 404 - товар не найден)
   - Автоматическое обновление токена при истечении

2. **Загрузка:**
   - Показывать индикатор загрузки
   - Skeleton screen для лучшего UX
   - Pull-to-refresh для обновления данных

3. **Кэширование:**
   - Сохранять токен админа локально
   - Кэшировать изображения товаров
   - Сохранять данные товара локально
   - Обновлять при необходимости

4. **Валидация:**
   - Проверять доступность выбранной комбинации
   - Показывать предупреждения о низком остатке
   - Валидировать токен перед запросами

5. **Безопасность:**
   - Не хранить пароль админа в коде
   - Использовать переменные окружения для конфиденциальных данных
   - Регулярно обновлять токен

## Пример использования в приложении

```dart
class ProductCatalogScreen extends StatefulWidget {
  @override
  _ProductCatalogScreenState createState() => _ProductCatalogScreenState();
}

class _ProductCatalogScreenState extends State<ProductCatalogScreen> {
  List<Product> products = [];
  bool isLoading = true;
  String? error;

  @override
  void initState() {
    super.initState();
    loadProducts();
  }

  Future<void> loadProducts() async {
    try {
      setState(() {
        isLoading = true;
        error = null;
      });
      
      final fetchedProducts = await fetchAllProducts(
        page: 1,
        limit: 20,
        inStock: true,
        sortBy: 'created_at',
        sortOrder: 'desc',
      );
      
      setState(() {
        products = fetchedProducts;
        isLoading = false;
      });
    } catch (e) {
      setState(() {
        error = e.toString();
        isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    if (isLoading) {
      return Center(child: CircularProgressIndicator());
    }
    
    if (error != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text('Ошибка: $error'),
            ElevatedButton(
              onPressed: loadProducts,
              child: Text('Повторить'),
            ),
          ],
        ),
      );
    }
    
    return ListView.builder(
      itemCount: products.length,
      itemBuilder: (context, index) {
        final product = products[index];
        return ListTile(
          title: Text(product.name),
          subtitle: Text(product.brand),
          onTap: () {
            Navigator.push(
              context,
              MaterialPageRoute(
                builder: (context) => ProductDetailScreen(productId: product.id),
              ),
            );
          },
        );
      },
    );
  }
}
```

Этот эндпоинт идеально подходит для создания карточки товара с выбором параметров!
