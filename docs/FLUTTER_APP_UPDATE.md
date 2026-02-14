# 📱 Обновление Flutter приложения: imageUrlsByColor

## ⚠️ ВАЖНО: Изменение структуры данных

API обновлен для использования `imageUrlsByColor` вместо `imageUrls`. Это позволяет хранить изображения отдельно для каждого цвета товара.

## 🔧 Что нужно изменить

### 1. Обновить модель ProductVariation

**Было:**
```dart
class ProductVariation {
  final String id;
  final List<String> imageUrls; // ❌ УСТАРЕЛО
  
  ProductVariation({
    required this.id,
    this.imageUrls = const [],
  });
  
  factory ProductVariation.fromJson(Map<String, dynamic> json) {
    return ProductVariation(
      id: json['id'],
      imageUrls: List<String>.from(json['imageUrls'] ?? []),
    );
  }
}
```

**Стало:**
```dart
class ProductVariation {
  final String id;
  final Map<String, List<String>> imageUrlsByColor; // ✅ НОВОЕ
  
  ProductVariation({
    required this.id,
    this.imageUrlsByColor = const {},
  });
  
  factory ProductVariation.fromJson(Map<String, dynamic> json) {
    // Преобразуем Map<String, dynamic> в Map<String, List<String>>
    final imageUrlsByColorMap = json['imageUrlsByColor'] as Map<String, dynamic>? ?? {};
    final imageUrlsByColor = imageUrlsByColorMap.map(
      (key, value) => MapEntry(
        key,
        List<String>.from(value as List? ?? []),
      ),
    );
    
    return ProductVariation(
      id: json['id'],
      imageUrlsByColor: imageUrlsByColor,
    );
  }
  
  // Метод для получения изображений по цвету
  List<String> getImagesForColor(String color) {
    return imageUrlsByColor[color] ?? [];
  }
  
  // Метод для получения всех изображений
  List<String> getAllImages() {
    return imageUrlsByColor.values.expand((list) => list).toList();
  }
  
  // Метод для получения первого изображения (для превью)
  String? getFirstImage([String? color]) {
    if (color != null && imageUrlsByColor.containsKey(color)) {
      final images = imageUrlsByColor[color]!;
      return images.isNotEmpty ? images.first : null;
    }
    
    // Если цвет не указан, берем первое изображение из первого цвета
    if (imageUrlsByColor.isNotEmpty) {
      final firstColorImages = imageUrlsByColor.values.first;
      return firstColorImages.isNotEmpty ? firstColorImages.first : null;
    }
    
    return null;
  }
}
```

### 2. Обновить виджет отображения товара

**Было:**
```dart
// ❌ УСТАРЕЛО
Widget buildProductImages(ProductVariation variation) {
  return ListView.builder(
    scrollDirection: Axis.horizontal,
    itemCount: variation.imageUrls.length,
    itemBuilder: (context, index) {
      return Image.network(variation.imageUrls[index]);
    },
  );
}
```

**Стало:**
```dart
// ✅ ПРАВИЛЬНО
Widget buildProductImages(ProductVariation variation, String selectedColor) {
  final images = variation.getImagesForColor(selectedColor);
  
  if (images.isEmpty) {
    return const Center(
      child: Text('Нет изображений для выбранного цвета'),
    );
  }
  
  return ListView.builder(
    scrollDirection: Axis.horizontal,
    itemCount: images.length,
    itemBuilder: (context, index) {
      return Image.network(images[index]);
    },
  );
}
```

### 3. Создать виджет выбора цвета

```dart
class ColorSelectorWidget extends StatelessWidget {
  final ProductVariation variation;
  final String selectedColor;
  final Function(String) onColorSelected;
  
  const ColorSelectorWidget({
    Key? key,
    required this.variation,
    required this.selectedColor,
    required this.onColorSelected,
  }) : super(key: key);
  
  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: variation.colors.map((color) {
        final images = variation.getImagesForColor(color);
        final isSelected = color == selectedColor;
        
        return GestureDetector(
          onTap: () => onColorSelected(color),
          child: Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              border: Border.all(
                color: isSelected ? Colors.blue : Colors.grey,
                width: isSelected ? 2 : 1,
              ),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Превью изображения для цвета
                if (images.isNotEmpty)
                  ClipRRect(
                    borderRadius: BorderRadius.circular(4),
                    child: Image.network(
                      images.first,
                      width: 50,
                      height: 50,
                      fit: BoxFit.cover,
                      errorBuilder: (context, error, stackTrace) {
                        return Container(
                          width: 50,
                          height: 50,
                          color: Colors.grey[300],
                          child: const Icon(Icons.image_not_supported),
                        );
                      },
                    ),
                  )
                else
                  Container(
                    width: 50,
                    height: 50,
                    color: Colors.grey[300],
                    child: const Icon(Icons.image_not_supported),
                  ),
                const SizedBox(height: 4),
                Text(
                  color,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                  ),
                ),
              ],
            ),
          ),
        );
      }).toList(),
    );
  }
}
```

### 4. Обновить экран деталей товара

```dart
class ProductDetailScreen extends StatefulWidget {
  final Product product;
  
  const ProductDetailScreen({Key? key, required this.product}) : super(key: key);
  
  @override
  State<ProductDetailScreen> createState() => _ProductDetailScreenState();
}

class _ProductDetailScreenState extends State<ProductDetailScreen> {
  String? selectedColor;
  ProductVariation? selectedVariation;
  
  @override
  void initState() {
    super.initState();
    // Выбираем первую вариацию и первый цвет по умолчанию
    if (widget.product.variations.isNotEmpty) {
      selectedVariation = widget.product.variations.first;
      if (selectedVariation!.colors.isNotEmpty) {
        selectedColor = selectedVariation!.colors.first;
      }
    }
  }
  
  @override
  Widget build(BuildContext context) {
    if (selectedVariation == null || selectedColor == null) {
      return const Scaffold(
        body: Center(child: Text('Нет доступных вариаций')),
      );
    }
    
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.product.name),
      ),
      body: Column(
        children: [
          // Галерея изображений
          Expanded(
            flex: 2,
            child: buildProductImages(selectedVariation!, selectedColor!),
          ),
          
          // Селектор цвета
          Padding(
            padding: const EdgeInsets.all(16),
            child: ColorSelectorWidget(
              variation: selectedVariation!,
              selectedColor: selectedColor!,
              onColorSelected: (color) {
                setState(() {
                  selectedColor = color;
                });
              },
            ),
          ),
          
          // Информация о товаре
          Expanded(
            flex: 1,
            child: buildProductInfo(),
          ),
        ],
      ),
    );
  }
  
  Widget buildProductImages(ProductVariation variation, String color) {
    final images = variation.getImagesForColor(color);
    
    if (images.isEmpty) {
      return const Center(
        child: Text('Нет изображений для выбранного цвета'),
      );
    }
    
    return PageView.builder(
      itemCount: images.length,
      itemBuilder: (context, index) {
        return Image.network(
          images[index],
          fit: BoxFit.contain,
          errorBuilder: (context, error, stackTrace) {
            return const Center(
              child: Icon(Icons.error, size: 50),
            );
          },
        );
      },
    );
  }
  
  Widget buildProductInfo() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            widget.product.name,
            style: const TextStyle(
              fontSize: 24,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Цена: ${selectedVariation!.price} ₽',
            style: const TextStyle(
              fontSize: 20,
              color: Colors.green,
            ),
          ),
          if (selectedVariation!.discount > 0)
            Text(
              'Скидка: ${selectedVariation!.discount}%',
              style: const TextStyle(
                fontSize: 16,
                color: Colors.red,
              ),
            ),
        ],
      ),
    );
  }
}
```

### 5. Обновить список товаров (превью)

```dart
class ProductCard extends StatelessWidget {
  final Product product;
  
  const ProductCard({Key? key, required this.product}) : super(key: key);
  
  @override
  Widget build(BuildContext context) {
    // Получаем первое изображение для превью
    String? previewImage;
    if (product.variations.isNotEmpty) {
      final firstVariation = product.variations.first;
      if (firstVariation.colors.isNotEmpty) {
        final firstColor = firstVariation.colors.first;
        final images = firstVariation.getImagesForColor(firstColor);
        if (images.isNotEmpty) {
          previewImage = images.first;
        }
      }
    }
    
    return Card(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Превью изображение
          if (previewImage != null)
            Image.network(
              previewImage,
              height: 200,
              width: double.infinity,
              fit: BoxFit.cover,
              errorBuilder: (context, error, stackTrace) {
                return Container(
                  height: 200,
                  color: Colors.grey[300],
                  child: const Icon(Icons.image_not_supported),
                );
              },
            )
          else
            Container(
              height: 200,
              color: Colors.grey[300],
              child: const Icon(Icons.image_not_supported),
            ),
          
          // Информация о товаре
          Padding(
            padding: const EdgeInsets.all(8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  product.name,
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (product.variations.isNotEmpty)
                  Text(
                    '${product.variations.first.price} ₽',
                    style: const TextStyle(
                      fontSize: 18,
                      color: Colors.green,
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
```

### 6. Обратная совместимость (если нужна)

```dart
class ProductVariation {
  // ... остальные поля ...
  
  // Метод для обратной совместимости
  List<String> get imageUrls {
    // Если есть imageUrlsByColor, возвращаем все изображения
    if (imageUrlsByColor.isNotEmpty) {
      return getAllImages();
    }
    // Иначе возвращаем пустой список
    return [];
  }
  
  // Или можно добавить fallback в fromJson
  factory ProductVariation.fromJson(Map<String, dynamic> json) {
    Map<String, List<String>> imageUrlsByColor = {};
    
    // Сначала пробуем новую структуру
    if (json['imageUrlsByColor'] != null) {
      final imageUrlsByColorMap = json['imageUrlsByColor'] as Map<String, dynamic>;
      imageUrlsByColor = imageUrlsByColorMap.map(
        (key, value) => MapEntry(
          key,
          List<String>.from(value as List? ?? []),
        ),
      );
    }
    
    // Fallback на старую структуру (для обратной совместимости)
    if (imageUrlsByColor.isEmpty && json['imageUrls'] != null) {
      final oldImageUrls = List<String>.from(json['imageUrls'] ?? []);
      if (oldImageUrls.isNotEmpty) {
        // Если есть цвета, используем первый цвет
        final colors = List<String>.from(json['colors'] ?? []);
        final color = colors.isNotEmpty ? colors.first : 'default';
        imageUrlsByColor[color] = oldImageUrls;
      }
    }
    
    return ProductVariation(
      id: json['id'],
      imageUrlsByColor: imageUrlsByColor,
      // ... остальные поля ...
    );
  }
}
```

## ✅ Чеклист для обновления

- [ ] Обновить модель `ProductVariation` - добавить `imageUrlsByColor`
- [ ] Обновить `fromJson` для парсинга новой структуры
- [ ] Добавить методы `getImagesForColor()`, `getAllImages()`, `getFirstImage()`
- [ ] Обновить виджеты отображения изображений
- [ ] Создать виджет выбора цвета `ColorSelectorWidget`
- [ ] Обновить экран деталей товара для работы с цветами
- [ ] Обновить карточки товаров в списке
- [ ] Протестировать с товарами, у которых есть изображения по цветам
- [ ] Добавить обработку ошибок загрузки изображений
- [ ] Обновить кэширование изображений (если используется)

## 📝 Пример полной модели

```dart
class ProductVariation {
  final String id;
  final String productId;
  final List<String> sizes;
  final List<String> colors;
  final double price;
  final double? originalPrice;
  final int discount;
  final Map<String, List<String>> imageUrlsByColor;
  final int stockQuantity;
  final bool isAvailable;
  final String sku;
  final String barcode;
  
  ProductVariation({
    required this.id,
    required this.productId,
    this.sizes = const [],
    this.colors = const [],
    required this.price,
    this.originalPrice,
    this.discount = 0,
    this.imageUrlsByColor = const {},
    this.stockQuantity = 0,
    this.isAvailable = true,
    this.sku = '',
    this.barcode = '',
  });
  
  factory ProductVariation.fromJson(Map<String, dynamic> json) {
    // Парсинг imageUrlsByColor
    final imageUrlsByColorMap = json['imageUrlsByColor'] as Map<String, dynamic>? ?? {};
    final imageUrlsByColor = imageUrlsByColorMap.map(
      (key, value) => MapEntry(
        key,
        List<String>.from(value as List? ?? []),
      ),
    );
    
    return ProductVariation(
      id: json['id'] ?? '',
      productId: json['productId'] ?? '',
      sizes: List<String>.from(json['sizes'] ?? []),
      colors: List<String>.from(json['colors'] ?? []),
      price: (json['price'] ?? 0).toDouble(),
      originalPrice: json['originalPrice'] != null 
          ? (json['originalPrice'] as num).toDouble() 
          : null,
      discount: json['discount'] ?? 0,
      imageUrlsByColor: imageUrlsByColor,
      stockQuantity: json['stockQuantity'] ?? 0,
      isAvailable: json['isAvailable'] ?? true,
      sku: json['sku'] ?? '',
      barcode: json['barcode'] ?? '',
    );
  }
  
  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'productId': productId,
      'sizes': sizes,
      'colors': colors,
      'price': price,
      'originalPrice': originalPrice,
      'discount': discount,
      'imageUrlsByColor': imageUrlsByColor,
      'stockQuantity': stockQuantity,
      'isAvailable': isAvailable,
      'sku': sku,
      'barcode': barcode,
    };
  }
  
  // Методы для работы с изображениями
  List<String> getImagesForColor(String color) {
    return imageUrlsByColor[color] ?? [];
  }
  
  List<String> getAllImages() {
    return imageUrlsByColor.values.expand((list) => list).toList();
  }
  
  String? getFirstImage([String? color]) {
    if (color != null && imageUrlsByColor.containsKey(color)) {
      final images = imageUrlsByColor[color]!;
      return images.isNotEmpty ? images.first : null;
    }
    
    if (imageUrlsByColor.isNotEmpty) {
      final firstColorImages = imageUrlsByColor.values.first;
      return firstColorImages.isNotEmpty ? firstColorImages.first : null;
    }
    
    return null;
  }
  
  bool hasImagesForColor(String color) {
    final images = imageUrlsByColor[color];
    return images != null && images.isNotEmpty;
  }
}
```

## 🚀 Готово!

После обновления ваше Flutter приложение будет корректно работать с новой структурой данных `imageUrlsByColor`.

