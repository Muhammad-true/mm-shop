package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product представляет товар
type Product struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"`
	Name        string     `json:"name" gorm:"not null"`
	Description string     `json:"description"`
	Gender      string     `json:"gender" gorm:"not null"` // 'male', 'female', 'unisex'
	CategoryID  uuid.UUID  `json:"categoryId" gorm:"type:uuid;not null"`
	Brand       string     `json:"brand"`
	IsAvailable bool       `json:"isAvailable" gorm:"default:true"`
	OwnerID     *uuid.UUID `json:"ownerId" gorm:"type:uuid"` // DEPRECATED: Используйте ShopID. Оставлено для обратной совместимости
	ShopID      *uuid.UUID `json:"shopId" gorm:"type:uuid;index"` // ID магазина
	CityID      *uuid.UUID `json:"cityId" gorm:"type:uuid;index"` // ID города (для быстрой фильтрации)
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	// Связи
	Category   *Category          `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Owner      *User              `json:"owner,omitempty" gorm:"foreignKey:OwnerID"` // DEPRECATED
	Shop       *Shop              `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
	City       *City              `json:"city,omitempty" gorm:"foreignKey:CityID"`
	Variations []ProductVariation `json:"variations,omitempty" gorm:"foreignKey:ProductID"`
}

// ProductVariation представляет вариацию товара (размеры + цвета + цена + фото)
type ProductVariation struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	ProductID     uuid.UUID `json:"productId" gorm:"type:uuid;not null"`
	Sizes         []string  `json:"sizes" gorm:"serializer:json"`  // Множественные размеры
	Colors        []string  `json:"colors" gorm:"serializer:json"` // Множественные цвета
	Price         float64   `json:"price" gorm:"not null"`
	OriginalPrice *float64  `json:"originalPrice"`
	Discount      int       `json:"discount" gorm:"default:0"`        // Скидка в процентах (0-100%), например: 15 = 15%
	ImageURLs     []string  `json:"imageUrls" gorm:"serializer:json"` // Множественные фото
	StockQuantity int       `json:"stockQuantity" gorm:"default:0"`
	IsAvailable   bool      `json:"isAvailable" gorm:"default:true"`
	SKU           string    `json:"sku"`
	Barcode       string    `json:"barcode" gorm:"index"` // Штрих-код (EAN-13, UPC, Code128 и т.д.)
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`

	// Связи
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// BeforeCreate устанавливает UUID перед созданием для вариации
func (pv *ProductVariation) BeforeCreate(tx *gorm.DB) error {
	if pv.ID == uuid.Nil {
		pv.ID = uuid.New()
	}
	return nil
}

// ProductRequest представляет запрос на создание/обновление продукта
type ProductRequest struct {
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
	Gender      string                    `json:"gender" binding:"required,oneof=male female unisex"`
	CategoryID  uuid.UUID                 `json:"categoryId" binding:"required"`
	Brand       string                    `json:"brand"`
	Variations  []ProductVariationRequest `json:"variations" binding:"required,min=1"`
	// OwnerID не включаем в запрос - будет автоматически устанавливаться из токена
}

// ProductVariationRequest представляет запрос на создание/обновление вариации
type ProductVariationRequest struct {
	Sizes         []string `json:"sizes" binding:"required,min=1"`
	Colors        []string `json:"colors" binding:"required,min=1"`
	Price         float64  `json:"price" binding:"required,gt=0"`
	OriginalPrice *float64 `json:"originalPrice"`
	Discount      int      `json:"discount" binding:"gte=0,lte=100"` // Скидка в процентах 0-100%, например: 15 = 15%
	ImageURLs     []string `json:"imageUrls"`                        // Множественные фото
	StockQuantity int      `json:"stockQuantity" binding:"gte=0"`
	SKU           string   `json:"sku"`
	Barcode       string   `json:"barcode"` // Штрих-код (EAN-13, UPC, Code128 и т.д.)
}

// ShopInfo представляет информацию о магазине
type ShopInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	INN  string    `json:"inn"`
}

// ProductResponse представляет ответ с информацией о продукте
type ProductResponse struct {
	ID          uuid.UUID                  `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	IsFavorite  bool                       `json:"isFavorite"` // Вычисляется для конкретного пользователя
	Gender      string                     `json:"gender"`
	CategoryID  uuid.UUID                  `json:"categoryId"`
	Category    *CategoryResponse          `json:"category,omitempty"`
	Brand       string                     `json:"brand"`
	IsAvailable bool                       `json:"isAvailable"`
	OwnerID     *uuid.UUID                 `json:"ownerId"`
	Owner       *UserResponse              `json:"owner,omitempty"`
	Shop        *ShopInfo                  `json:"shop,omitempty"` // Информация о магазине (имя и ИНН)
	Variations  []ProductVariationResponse `json:"variations"`
	CreatedAt   time.Time                  `json:"createdAt"`
	UpdatedAt   time.Time                  `json:"updatedAt"`
}

// ProductVariationResponse представляет ответ с информацией о вариации
type ProductVariationResponse struct {
	ID            uuid.UUID `json:"id"`
	Sizes         []string  `json:"sizes"`
	Colors        []string  `json:"colors"`
	Price         float64   `json:"price"`
	OriginalPrice *float64  `json:"originalPrice"`
	Discount      int       `json:"discount"`  // Скидка в процентах (0-100%), например: 15 = 15%
	ImageURLs     []string  `json:"imageUrls"` // Множественные фото
	StockQuantity int       `json:"stockQuantity"`
	IsAvailable   bool      `json:"isAvailable"`
	SKU           string    `json:"sku"`
	Barcode       string    `json:"barcode"` // Штрих-код
}

// ToResponse преобразует ProductVariation в ProductVariationResponse
func (pv *ProductVariation) ToResponse() ProductVariationResponse {
	return ProductVariationResponse{
		ID:            pv.ID,
		Sizes:         pv.Sizes,
		Colors:        pv.Colors,
		Price:         pv.Price,
		OriginalPrice: pv.OriginalPrice,
		Discount:      pv.Discount,
		ImageURLs:     pv.ImageURLs,
		StockQuantity: pv.StockQuantity,
		IsAvailable:   pv.IsAvailable,
		SKU:           pv.SKU,
		Barcode:       pv.Barcode,
	}
}

// ToResponse преобразует Product в ProductResponse
func (p *Product) ToResponse() ProductResponse {
	variations := make([]ProductVariationResponse, len(p.Variations))
	for i, v := range p.Variations {
		variations[i] = ProductVariationResponse{
			ID:            v.ID,
			Sizes:         v.Sizes,
			Colors:        v.Colors,
			Price:         v.Price,
			OriginalPrice: v.OriginalPrice,
			Discount:      v.Discount,
			ImageURLs:     v.ImageURLs, // Assuming the first image is the main one for response
			StockQuantity: v.StockQuantity,
			IsAvailable:   v.IsAvailable,
			SKU:           v.SKU,
			Barcode:       v.Barcode,
		}
	}

	response := ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		IsFavorite:  false, // Будет устанавливаться в контроллере
		Gender:      p.Gender,
		CategoryID:  p.CategoryID,
		Brand:       p.Brand,
		IsAvailable: p.IsAvailable,
		OwnerID:     p.OwnerID,
		Variations:  variations,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	// Добавляем данные категории, если они загружены
	if p.Category != nil && p.Category.ID != uuid.Nil {
		categoryResponse := p.Category.ToResponse()
		response.Category = &categoryResponse
	}

	// Добавляем данные владельца (для обратной совместимости)
	if p.Owner != nil && p.Owner.ID != uuid.Nil {
		ownerResponse := p.Owner.ToResponse()
		response.Owner = &ownerResponse
	}

	// Добавляем информацию о магазине (приоритет Shop, затем Owner для обратной совместимости)
	if p.Shop != nil && p.Shop.ID != uuid.Nil {
		// Используем новую таблицу Shop
		response.Shop = &ShopInfo{
			ID:   p.Shop.ID,
			Name: p.Shop.Name,
			INN:  p.Shop.INN,
		}
	} else if p.Owner != nil && p.Owner.ID != uuid.Nil {
		// Обратная совместимость: используем Owner если Shop не загружен
		response.Shop = &ShopInfo{
			ID:   p.Owner.ID,
			Name: p.Owner.Name,
			INN:  p.Owner.INN,
		}
	}

	return response
}

// ToResponseWithFavorite преобразует Product в ProductResponse с указанием избранного
func (p *Product) ToResponseWithFavorite(isFavorite bool) ProductResponse {
	response := p.ToResponse()
	response.IsFavorite = isFavorite
	return response
}

// ProductWithVariation представляет результат JOIN запроса между продуктами и вариациями
type ProductWithVariation struct {
	ProductID     uuid.UUID `json:"productId"`     // p.id
	Name          string    `json:"name"`          // p.name
	Description   string    `json:"description"`   // p.description
	Brand         string    `json:"brand"`         // p.brand
	Sizes         []string  `json:"sizes"`         // pv.sizes
	Colors        []string  `json:"colors"`        // pv.colors
	Price         float64   `json:"price"`         // pv.price
	OriginalPrice *float64  `json:"originalPrice"` // pv.original_price
	ImageURLs     []string  `json:"imageUrls"`     // pv.image_urls
	StockQuantity int64     `json:"stockQuantity"` // pv.stock_quantity
	SKU           string    `json:"sku"`           // pv.sku
}
