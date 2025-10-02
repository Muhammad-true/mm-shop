package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CartItem представляет элемент в корзине пользователя
type CartItem struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	ProductID   uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	VariationID uuid.UUID `json:"variation_id" gorm:"type:uuid;not null"`
	Quantity    int       `json:"quantity" gorm:"not null;default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Связи
	User      User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Product   Product          `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Variation ProductVariation `json:"variation,omitempty" gorm:"foreignKey:VariationID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (c *CartItem) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CartItemRequest представляет запрос на добавление в корзину
type CartItemRequest struct {
	ProductID   uuid.UUID `json:"product_id" binding:"required"`
	VariationID uuid.UUID `json:"variation_id" binding:"required"`
	Quantity    int       `json:"quantity" binding:"required,gt=0"`
}

// CartItemUpdateRequest представляет запрос на обновление элемента корзины
type CartItemUpdateRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

// CartItemResponse представляет ответ с элементом корзины
type CartItemResponse struct {
	ID        uuid.UUID                `json:"id"`
	Quantity  int                      `json:"quantity"`
	Subtotal  float64                  `json:"subtotal"`
	Product   ProductResponse          `json:"product"`
	Variation ProductVariationResponse `json:"variation"`
	CreatedAt time.Time                `json:"created_at"`
}

// ToResponse преобразует CartItem в CartItemResponse
func (c *CartItem) ToResponse() CartItemResponse {
	// Теперь у нас есть прямая связь с вариацией - намного проще!
	subtotal := c.Variation.Price * float64(c.Quantity)

	// Преобразуем вариацию в response
	variationResponse := ProductVariationResponse{
		ID:            c.Variation.ID,
		Sizes:         c.Variation.Sizes,
		Colors:        c.Variation.Colors,
		Price:         c.Variation.Price,
		OriginalPrice: c.Variation.OriginalPrice,
		ImageURLs:     c.Variation.ImageURLs,
		StockQuantity: c.Variation.StockQuantity,
		IsAvailable:   c.Variation.IsAvailable,
		SKU:           c.Variation.SKU,
	}

	return CartItemResponse{
		ID:        c.ID,
		Quantity:  c.Quantity,
		Subtotal:  subtotal,
		Product:   c.Product.ToResponse(),
		Variation: variationResponse,
		CreatedAt: c.CreatedAt,
	}
}

// CartResponse представляет ответ с содержимым корзины
type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	TotalPrice float64            `json:"total_price"`
}
