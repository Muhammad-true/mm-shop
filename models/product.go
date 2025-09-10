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
	OwnerID     *uuid.UUID `json:"ownerId" gorm:"type:uuid"` // Владелец товара (магазина) - временно nullable
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	// Связи
	Category   *Category          `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Owner      *User              `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
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
	ImageURLs     []string  `json:"imageUrls" gorm:"serializer:json"` // Множественные фото
	StockQuantity int       `json:"stockQuantity" gorm:"default:0"`
	IsAvailable   bool      `json:"isAvailable" gorm:"default:true"`
	SKU           string    `json:"sku"`
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
	ImageURLs     []string `json:"imageUrls"` // Множественные фото
	StockQuantity int      `json:"stockQuantity" binding:"gte=0"`
	SKU           string   `json:"sku"`
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
	ImageURLs     []string  `json:"imageUrls"` // Множественные фото
	StockQuantity int       `json:"stockQuantity"`
	IsAvailable   bool      `json:"isAvailable"`
	SKU           string    `json:"sku"`
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
			ImageURLs:     v.ImageURLs, // Assuming the first image is the main one for response
			StockQuantity: v.StockQuantity,
			IsAvailable:   v.IsAvailable,
			SKU:           v.SKU,
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

	// Добавляем данные владельца, если они загружены
	if p.Owner != nil && p.Owner.ID != uuid.Nil {
		ownerResponse := p.Owner.ToResponse()
		response.Owner = &ownerResponse
	}

	return response
}

// ToResponseWithFavorite преобразует Product в ProductResponse с указанием избранного
func (p *Product) ToResponseWithFavorite(isFavorite bool) ProductResponse {
	response := p.ToResponse()
	response.IsFavorite = isFavorite
	return response
}
