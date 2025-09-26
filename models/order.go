package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrderStatus представляет статус заказа
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusPreparing  OrderStatus = "preparing"
	OrderStatusInDelivery OrderStatus = "inDelivery"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// Order представляет заказ пользователя
type Order struct {
	ID            uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;"`
	UserID        uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	Status        OrderStatus `json:"status" gorm:"default:pending"`
	TotalAmount   float64     `json:"total_amount" gorm:"not null"`
	ItemsSubtotal float64     `json:"items_subtotal" gorm:"not null"`
	DeliveryFee   float64     `json:"delivery_fee" gorm:"not null;default:0"`
	Currency      string      `json:"currency" gorm:"not null;default:TJS"`
	ShippingAddr  string      `json:"shipping_address" gorm:"not null"`
	PaymentMethod string      `json:"payment_method" gorm:"not null"`
	PaymentStatus string      `json:"payment_status" gorm:"not null;default:pending"`
	TransactionID string      `json:"transaction_id"`
	RecipientName string      `json:"recipient_name" gorm:"not null"`
	Phone         string      `json:"phone" gorm:"not null"`
	DesiredAt     *time.Time  `json:"desired_at"`
	ConfirmedAt   *time.Time  `json:"confirmed_at"`
	DeliveredAt   *time.Time  `json:"delivered_at"`
	CancelledAt   *time.Time  `json:"cancelled_at"`
	Notes         string      `json:"notes"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`

	// Связи
	User       User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OrderItems []OrderItem `json:"order_items,omitempty" gorm:"foreignKey:OrderID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// OrderItem представляет элемент заказа
type OrderItem struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	OrderID   uuid.UUID `json:"order_id" gorm:"type:uuid;not null"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	Quantity  int       `json:"quantity" gorm:"not null"`
	Price     float64   `json:"price" gorm:"not null"` // Цена на момент заказа
	Size      string    `json:"size"`
	Color     string    `json:"color"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name" gorm:"not null"`
	ImageURL  string    `json:"image_url"`
	Total     float64   `json:"total" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`

	// Связи
	Order   Order   `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return nil
}

// OrderRequest представляет запрос на создание заказа
type OrderRequest struct {
	ShippingAddr  string `json:"shipping_address" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	Notes         string `json:"notes"`
}

// OrderResponse представляет ответ с информацией о заказе
type OrderResponse struct {
	ID            uuid.UUID           `json:"id"`
	Status        OrderStatus         `json:"status"`
	TotalAmount   float64             `json:"total_amount"`
	ShippingAddr  string              `json:"shipping_address"`
	PaymentMethod string              `json:"payment_method"`
	Notes         string              `json:"notes"`
	OrderItems    []OrderItemResponse `json:"order_items"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// OrderItemResponse представляет ответ с элементом заказа
type OrderItemResponse struct {
	ID       uuid.UUID       `json:"id"`
	Quantity int             `json:"quantity"`
	Price    float64         `json:"price"`
	Size     string          `json:"size"`
	Color    string          `json:"color"`
	Subtotal float64         `json:"subtotal"`
	Product  ProductResponse `json:"product"`
}

// ToResponse преобразует Order в OrderResponse
func (o *Order) ToResponse() OrderResponse {
	orderItems := make([]OrderItemResponse, len(o.OrderItems))
	for i, item := range o.OrderItems {
		orderItems[i] = item.ToResponse()
	}

	return OrderResponse{
		ID:            o.ID,
		Status:        o.Status,
		TotalAmount:   o.TotalAmount,
		ShippingAddr:  o.ShippingAddr,
		PaymentMethod: o.PaymentMethod,
		Notes:         o.Notes,
		OrderItems:    orderItems,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
	}
}

// ToResponse преобразует OrderItem в OrderItemResponse
func (oi *OrderItem) ToResponse() OrderItemResponse {
	subtotal := oi.Price * float64(oi.Quantity)

	return OrderItemResponse{
		ID:       oi.ID,
		Quantity: oi.Quantity,
		Price:    oi.Price,
		Size:     oi.Size,
		Color:    oi.Color,
		Subtotal: subtotal,
		Product:  oi.Product.ToResponse(),
	}
}
