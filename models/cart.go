package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CartItem представляет элемент в корзине пользователя
type CartItem struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	Quantity  int       `json:"quantity" gorm:"not null;default:1"`
	Size      string    `json:"size"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Связи
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
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
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
	Size      string    `json:"size"`
	Color     string    `json:"color"`
}

// CartItemUpdateRequest представляет запрос на обновление элемента корзины
type CartItemUpdateRequest struct {
	Quantity int    `json:"quantity" binding:"required,gt=0"`
	Size     string `json:"size"`
	Color    string `json:"color"`
}

// CartItemResponse представляет ответ с элементом корзины
type CartItemResponse struct {
	ID        uuid.UUID       `json:"id"`
	Quantity  int             `json:"quantity"`
	Size      string          `json:"size"`
	Color     string          `json:"color"`
	Subtotal  float64         `json:"subtotal"`
	Product   ProductResponse `json:"product"`
	CreatedAt time.Time       `json:"created_at"`
}

// ToResponse преобразует CartItem в CartItemResponse
func (c *CartItem) ToResponse() CartItemResponse {
	// Находим подходящую вариацию по размеру и цвету
	var subtotal float64
	for _, variation := range c.Product.Variations {
		// Проверяем, есть ли нужный размер и цвет в вариации
		sizeMatch := false
		colorMatch := false

		for _, size := range variation.Sizes {
			if size == c.Size {
				sizeMatch = true
				break
			}
		}

		for _, color := range variation.Colors {
			if color == c.Color {
				colorMatch = true
				break
			}
		}

		if sizeMatch && colorMatch {
			subtotal = variation.Price * float64(c.Quantity)
			break
		}
	}

	return CartItemResponse{
		ID:        c.ID,
		Quantity:  c.Quantity,
		Size:      c.Size,
		Color:     c.Color,
		Subtotal:  subtotal,
		Product:   c.Product.ToResponse(),
		CreatedAt: c.CreatedAt,
	}
}

// CartResponse представляет ответ с содержимым корзины
type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	TotalPrice float64            `json:"total_price"`
}
