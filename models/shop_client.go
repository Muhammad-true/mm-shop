package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShopClient представляет связь между магазином и клиентом
// Клиент может не быть зарегистрирован в системе (User), но иметь связь с магазином
type ShopClient struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"`
	ShopID        uuid.UUID  `json:"shopId" gorm:"type:uuid;not null;index"`
	Phone         string     `json:"phone" gorm:"not null;index"` // Номер телефона клиента
	QRCode        string     `json:"qrCode" gorm:"not null"`      // QR код клиента в этом магазине (уникальный для каждого магазина)
	BonusAmount   int        `json:"bonusAmount" gorm:"default:0"` // Текущее количество бонусов
	FirstBonusDate *time.Time `json:"firstBonusDate"`              // Дата первого получения бонусов
	UserID        *uuid.UUID `json:"userId" gorm:"type:uuid;index"` // ID пользователя, если он зарегистрирован в системе
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`

	// Связи
	Shop Shop `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (sc *ShopClient) BeforeCreate(tx *gorm.DB) error {
	if sc.ID == uuid.Nil {
		sc.ID = uuid.New()
	}
	return nil
}

// ShopClientResponse представляет ответ с информацией о клиенте магазина
type ShopClientResponse struct {
	ID            uuid.UUID   `json:"id"`
	ShopID        uuid.UUID   `json:"shopId"`
	Phone         string      `json:"phone"`
	QRCode        string      `json:"qrCode"`
	BonusAmount   int         `json:"bonusAmount"`
	FirstBonusDate *time.Time `json:"firstBonusDate"`
	Shop          *ShopInfo   `json:"shop,omitempty"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

// ToResponse преобразует ShopClient в ShopClientResponse
func (sc *ShopClient) ToResponse() ShopClientResponse {
	response := ShopClientResponse{
		ID:            sc.ID,
		ShopID:        sc.ShopID,
		Phone:         sc.Phone,
		QRCode:        sc.QRCode,
		BonusAmount:   sc.BonusAmount,
		FirstBonusDate: sc.FirstBonusDate,
		CreatedAt:     sc.CreatedAt,
		UpdatedAt:     sc.UpdatedAt,
	}

	// Добавляем информацию о магазине, если она загружена
	if sc.Shop.ID != uuid.Nil {
		shopInfo := sc.Shop.ToShopInfo()
		response.Shop = &shopInfo
	}

	return response
}

