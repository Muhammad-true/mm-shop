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
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;"`
	UserID         uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	Status         OrderStatus `json:"status" gorm:"default:pending"`
	TotalAmount    float64     `json:"total_amount" gorm:"not null"`
	ItemsSubtotal  float64     `json:"items_subtotal" gorm:"not null"`
	DeliveryFee    float64     `json:"delivery_fee" gorm:"not null;default:0"`
	Currency       string      `json:"currency" gorm:"not null;default:TJS"`
	AddressID      *uuid.UUID  `json:"address_id" gorm:"type:uuid"`
	ShippingAddr   string      `json:"shipping_address" gorm:"not null"`
	PaymentMethod  string      `json:"payment_method" gorm:"not null"`
	ShippingMethod string      `json:"shipping_method" gorm:"not null;default:courier"`
	PaymentStatus  string      `json:"payment_status" gorm:"not null;default:pending"`
	TransactionID  string      `json:"transaction_id"`
	RecipientName  string      `json:"recipient_name" gorm:"not null"`
	Phone          string      `json:"phone" gorm:"not null"`
	DesiredAt      *time.Time  `json:"desired_at" gorm:"type:timestamp with time zone"`
	ConfirmedAt    *time.Time  `json:"confirmed_at"`
	CancelledAt    *time.Time  `json:"cancelled_at"`
	Notes          string      `json:"notes"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`

	// Связи
	User       User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Address    *Address    `json:"address,omitempty" gorm:"foreignKey:AddressID"`
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
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	OrderID     uuid.UUID `json:"order_id" gorm:"type:uuid;not null"`
	VariationID uuid.UUID `json:"variation_id" gorm:"type:uuid;not null"`
	Quantity    int       `json:"quantity" gorm:"not null"`
	Price       float64   `json:"price" gorm:"not null"` // Цена на момент заказа
	Size        string    `json:"size"`
	Color       string    `json:"color"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name" gorm:"not null"`
	ImageURL    string    `json:"image_url"`
	Total       float64   `json:"total" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`

	// Связи
	Order     Order            `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Variation ProductVariation `json:"variation,omitempty" gorm:"foreignKey:VariationID"`
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
	ID             uuid.UUID           `json:"id"`
	Status         OrderStatus         `json:"status"`
	TotalAmount    float64             `json:"total_amount"`
	ItemsSubtotal  float64             `json:"items_subtotal"`
	DeliveryFee    float64             `json:"delivery_fee"`
	Currency       string              `json:"currency"`
	ShippingAddr   string              `json:"shipping_address"`
	PaymentMethod  string              `json:"payment_method"`
	ShippingMethod string              `json:"shipping_method"`
	PaymentStatus  string              `json:"payment_status"`
	RecipientName  string              `json:"recipient_name"`
	Phone          string              `json:"phone"`
	Notes          string              `json:"notes"`
	OrderItems     []OrderItemResponse `json:"order_items"`
	DesiredAt      *time.Time          `json:"desired_at"`
	ConfirmedAt    *time.Time          `json:"confirmed_at"`
	CancelledAt    *time.Time          `json:"cancelled_at"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// AdminOrderResponse представляет расширенный ответ для админ панели
type AdminOrderResponse struct {
	ID             uuid.UUID           `json:"id"`
	OrderNumber    string              `json:"order_number"` // Короткий номер для отображения
	Status         OrderStatus         `json:"status"`
	TotalAmount    float64             `json:"total_amount"`
	ItemsSubtotal  float64             `json:"items_subtotal"`
	DeliveryFee    float64             `json:"delivery_fee"`
	Currency       string              `json:"currency"`
	ShippingAddr   string              `json:"shipping_address"`
	PaymentMethod  string              `json:"payment_method"`
	ShippingMethod string              `json:"shipping_method"`
	PaymentStatus  string              `json:"payment_status"`
	RecipientName  string              `json:"recipient_name"`
	Phone          string              `json:"phone"`
	Notes          string              `json:"notes"`
	OrderItems     []OrderItemResponse `json:"order_items"`
	DesiredAt      *time.Time          `json:"desired_at"`
	ConfirmedAt    *time.Time          `json:"confirmed_at"`
	CancelledAt    *time.Time          `json:"cancelled_at"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	User           *UserBasicInfo      `json:"user,omitempty"` // Информация о пользователе
	ShopOwner      *UserBasicInfo      `json:"shop_owner,omitempty"` // Информация о владельце магазина
}

// UserBasicInfo базовая информация о пользователе для заказа
type UserBasicInfo struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Phone   string    `json:"phone"`
	Email   string    `json:"email"`
	IsGuest bool      `json:"is_guest"`
}

// OrderItemResponse представляет ответ с элементом заказа
type OrderItemResponse struct {
	ID          uuid.UUID                `json:"id"`
	Quantity    int                      `json:"quantity"`
	Price       float64                  `json:"price"`
	Size        string                   `json:"size"`
	Color       string                   `json:"color"`
	VariationID uuid.UUID                `json:"variation_id"`
	Variation   ProductVariationResponse `json:"variation"`
	Subtotal    float64                  `json:"subtotal"`
}

// ToResponse преобразует Order в OrderResponse
func (o *Order) ToResponse() OrderResponse {
	orderItems := make([]OrderItemResponse, len(o.OrderItems))
	for i, item := range o.OrderItems {
		orderItems[i] = item.ToResponse()
	}

	return OrderResponse{
		ID:             o.ID,
		Status:         o.Status,
		TotalAmount:    o.TotalAmount,
		ItemsSubtotal:  o.ItemsSubtotal,
		DeliveryFee:    o.DeliveryFee,
		Currency:       o.Currency,
		ShippingAddr:   o.ShippingAddr,
		PaymentMethod:  o.PaymentMethod,
		ShippingMethod: o.ShippingMethod,
		PaymentStatus:  o.PaymentStatus,
		RecipientName:  o.RecipientName,
		Phone:          o.Phone,
		Notes:          o.Notes,
		OrderItems:     orderItems,
		DesiredAt:      o.DesiredAt,
		ConfirmedAt:    o.ConfirmedAt,
		CancelledAt:    o.CancelledAt,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

// ToAdminResponse преобразует Order в AdminOrderResponse с информацией о пользователе
func (o *Order) ToAdminResponse() AdminOrderResponse {
	orderItems := make([]OrderItemResponse, len(o.OrderItems))
	for i, item := range o.OrderItems {
		orderItems[i] = item.ToResponse()
	}

	// Создаём короткий номер заказа из первых 8 символов UUID
	orderNumber := o.ID.String()[:8]

	response := AdminOrderResponse{
		ID:             o.ID,
		OrderNumber:    orderNumber,
		Status:         o.Status,
		TotalAmount:    o.TotalAmount,
		ItemsSubtotal:  o.ItemsSubtotal,
		DeliveryFee:    o.DeliveryFee,
		Currency:       o.Currency,
		ShippingAddr:   o.ShippingAddr,
		PaymentMethod:  o.PaymentMethod,
		ShippingMethod: o.ShippingMethod,
		PaymentStatus:  o.PaymentStatus,
		RecipientName:  o.RecipientName,
		Phone:          o.Phone,
		Notes:          o.Notes,
		OrderItems:     orderItems,
		DesiredAt:      o.DesiredAt,
		ConfirmedAt:    o.ConfirmedAt,
		CancelledAt:    o.CancelledAt,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}

	// Добавляем информацию о пользователе если она загружена
	if o.User.ID != uuid.Nil {
		response.User = &UserBasicInfo{
			ID:      o.User.ID,
			Name:    o.User.Name,
			Phone:   o.User.Phone,
			Email:   o.User.Email,
			IsGuest: o.User.IsGuest,
		}
	}

	// Добавляем информацию о владельце магазина из первого товара в заказе
	if len(o.OrderItems) > 0 && o.OrderItems[0].Variation.Product.Owner != nil && o.OrderItems[0].Variation.Product.Owner.ID != uuid.Nil {
		response.ShopOwner = &UserBasicInfo{
			ID:      o.OrderItems[0].Variation.Product.Owner.ID,
			Name:    o.OrderItems[0].Variation.Product.Owner.Name,
			Phone:   o.OrderItems[0].Variation.Product.Owner.Phone,
			Email:   o.OrderItems[0].Variation.Product.Owner.Email,
			IsGuest: o.OrderItems[0].Variation.Product.Owner.IsGuest,
		}
	}

	return response
}

// ToResponse преобразует OrderItem в OrderItemResponse
func (oi *OrderItem) ToResponse() OrderItemResponse {
	subtotal := oi.Price * float64(oi.Quantity)

	return OrderItemResponse{
		ID:          oi.ID,
		Quantity:    oi.Quantity,
		Price:       oi.Price,
		Size:        oi.Size,
		Color:       oi.Color,
		VariationID: oi.VariationID,
		Variation:   oi.Variation.ToResponse(),
		Subtotal:    subtotal,
	}
}
