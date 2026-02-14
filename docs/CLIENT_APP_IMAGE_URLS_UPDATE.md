# 🔄 Обновление клиентского приложения: Переход на imageUrlsByColor

## ⚠️ ВАЖНО: Изменение структуры данных

С версии API обновлена структура хранения изображений товаров. Теперь используется `imageUrlsByColor` вместо `imageUrls`.

## 📋 Что изменилось

### Старая структура (DEPRECATED):
```json
{
  "variation": {
    "imageUrls": [
      "/images/variations/photo1.jpg",
      "/images/variations/photo2.jpg"
    ]
  }
}
```

### Новая структура:
```json
{
  "variation": {
    "imageUrlsByColor": {
      "Черный": [
        "/images/variations/photo1.jpg",
        "/images/variations/photo2.jpg"
      ],
      "Белый": [
        "/images/variations/photo3.jpg"
      ]
    }
  }
}
```

## 🔧 Что нужно исправить в клиентском приложении

### 1. Отображение изображений товара

**Было:**
```javascript
// ❌ УСТАРЕЛО
const images = variation.imageUrls || [];
images.forEach(url => {
  // Отобразить изображение
});
```

**Стало:**
```javascript
// ✅ ПРАВИЛЬНО
const imageUrlsByColor = variation.imageUrlsByColor || {};
const selectedColor = userSelectedColor || variation.colors[0]; // Выбранный цвет
const images = imageUrlsByColor[selectedColor] || [];

images.forEach(url => {
  // Отобразить изображение
});
```

### 2. Отображение всех изображений (для галереи)

```javascript
// Получить все изображения из всех цветов
const imageUrlsByColor = variation.imageUrlsByColor || {};
const allImages = Object.values(imageUrlsByColor).flat();

// Или сгруппированные по цветам
Object.entries(imageUrlsByColor).forEach(([color, urls]) => {
  console.log(`Цвет ${color}:`, urls);
});
```

### 3. Выбор изображений по цвету

```javascript
function getImagesForColor(variation, color) {
  const imageUrlsByColor = variation.imageUrlsByColor || {};
  return imageUrlsByColor[color] || [];
}

// Использование
const blackImages = getImagesForColor(variation, 'Черный');
const whiteImages = getImagesForColor(variation, 'Белый');
```

### 4. Обратная совместимость

Если нужно поддерживать старые данные:

```javascript
function getVariationImages(variation) {
  // Сначала проверяем новую структуру
  if (variation.imageUrlsByColor && Object.keys(variation.imageUrlsByColor).length > 0) {
    // Используем новую структуру
    const selectedColor = userSelectedColor || variation.colors[0];
    return variation.imageUrlsByColor[selectedColor] || [];
  }
  
  // Fallback на старую структуру (для обратной совместимости)
  if (variation.imageUrls && variation.imageUrls.length > 0) {
    return variation.imageUrls;
  }
  
  return [];
}
```

### 5. Компонент выбора цвета с превью

```javascript
function renderColorSelector(variation) {
  const imageUrlsByColor = variation.imageUrlsByColor || {};
  
  return variation.colors.map(color => {
    const images = imageUrlsByColor[color] || [];
    const previewUrl = images[0]; // Первое изображение для превью
    
    return `
      <div class="color-option" data-color="${color}">
        <div class="color-preview">
          ${previewUrl ? `<img src="${previewUrl}" alt="${color}">` : ''}
        </div>
        <span>${color}</span>
      </div>
    `;
  }).join('');
}
```

### 6. Обработка изменения цвета

```javascript
function onColorChange(variation, selectedColor) {
  const imageUrlsByColor = variation.imageUrlsByColor || {};
  const images = imageUrlsByColor[selectedColor] || [];
  
  // Обновить отображаемые изображения
  updateImageGallery(images);
  
  // Обновить превью
  if (images.length > 0) {
    updateMainImage(images[0]);
  }
}
```

## 📱 Примеры для разных платформ

### React Native / React

```jsx
function ProductVariation({ variation, selectedColor, onColorChange }) {
  const imageUrlsByColor = variation.imageUrlsByColor || {};
  const images = imageUrlsByColor[selectedColor] || [];
  
  return (
    <View>
      {/* Селектор цвета */}
      <View style={styles.colorSelector}>
        {variation.colors.map(color => (
          <TouchableOpacity
            key={color}
            onPress={() => onColorChange(color)}
            style={[
              styles.colorOption,
              selectedColor === color && styles.colorOptionSelected
            ]}
          >
            {imageUrlsByColor[color]?.[0] && (
              <Image
                source={{ uri: imageUrlsByColor[color][0] }}
                style={styles.colorPreview}
              />
            )}
            <Text>{color}</Text>
          </TouchableOpacity>
        ))}
      </View>
      
      {/* Галерея изображений */}
      <ScrollView horizontal>
        {images.map((url, index) => (
          <Image
            key={index}
            source={{ uri: url }}
            style={styles.productImage}
          />
        ))}
      </ScrollView>
    </View>
  );
}
```

### Flutter / Dart

```dart
class ProductVariationWidget extends StatelessWidget {
  final Variation variation;
  final String selectedColor;
  final Function(String) onColorChange;
  
  @override
  Widget build(BuildContext context) {
    final imageUrlsByColor = variation.imageUrlsByColor ?? {};
    final images = imageUrlsByColor[selectedColor] ?? [];
    
    return Column(
      children: [
        // Селектор цвета
        Wrap(
          children: variation.colors.map((color) {
            final colorImages = imageUrlsByColor[color] ?? [];
            return GestureDetector(
              onTap: () => onColorChange(color),
              child: Container(
                decoration: BoxDecoration(
                  border: Border.all(
                    color: selectedColor == color ? Colors.blue : Colors.grey,
                  ),
                ),
                child: colorImages.isNotEmpty
                    ? Image.network(colorImages[0])
                    : Text(color),
              ),
            );
          }).toList(),
        ),
        
        // Галерея изображений
        ListView.builder(
          scrollDirection: Axis.horizontal,
          itemCount: images.length,
          itemBuilder: (context, index) {
            return Image.network(images[index]);
          },
        ),
      ],
    );
  }
}
```

### Swift (iOS)

```swift
struct ProductVariationView: View {
    let variation: ProductVariation
    @State var selectedColor: String
    
    var images: [String] {
        let imageUrlsByColor = variation.imageUrlsByColor ?? [:]
        return imageUrlsByColor[selectedColor] ?? []
    }
    
    var body: some View {
        VStack {
            // Селектор цвета
            ScrollView(.horizontal) {
                HStack {
                    ForEach(variation.colors, id: \.self) { color in
                        Button(action: { selectedColor = color }) {
                            VStack {
                                if let firstImage = variation.imageUrlsByColor?[color]?.first {
                                    AsyncImage(url: URL(string: firstImage))
                                        .frame(width: 50, height: 50)
                                }
                                Text(color)
                            }
                        }
                    }
                }
            }
            
            // Галерея изображений
            ScrollView(.horizontal) {
                HStack {
                    ForEach(images, id: \.self) { url in
                        AsyncImage(url: URL(string: url))
                            .frame(width: 200, height: 200)
                    }
                }
            }
        }
    }
}
```

### Kotlin (Android)

```kotlin
class ProductVariationAdapter(
    private val variation: ProductVariation,
    private val onColorSelected: (String) -> Unit
) : RecyclerView.Adapter<RecyclerView.ViewHolder>() {
    
    private var selectedColor: String = variation.colors.firstOrNull() ?: ""
    
    private val images: List<String>
        get() = variation.imageUrlsByColor?.get(selectedColor) ?: emptyList()
    
    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): RecyclerView.ViewHolder {
        // Создание ViewHolder для цвета или изображения
    }
    
    override fun onBindViewHolder(holder: RecyclerView.ViewHolder, position: Int) {
        // Привязка данных
    }
    
    fun selectColor(color: String) {
        selectedColor = color
        notifyDataSetChanged()
        onColorSelected(color)
    }
}
```

## ✅ Чеклист для обновления

- [ ] Найти все места, где используется `variation.imageUrls`
- [ ] Заменить на `variation.imageUrlsByColor[selectedColor]`
- [ ] Добавить логику выбора цвета
- [ ] Обновить компоненты отображения изображений
- [ ] Добавить превью для селектора цвета
- [ ] Протестировать с товарами, у которых есть изображения по цветам
- [ ] Протестировать обратную совместимость (если нужна)

## 📞 Вопросы?

Если возникли вопросы по обновлению, обратитесь к документации API:
- [API_ENDPOINTS.md](./API_ENDPOINTS.md)
- [API_CLIENT_APP.md](./API_CLIENT_APP.md)

