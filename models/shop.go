package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Shop представляет магазин в системе
type Shop struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"`
	Name        string     `json:"name" gorm:"not null"`
	INN         string     `json:"inn" gorm:"index"`
	Description string     `json:"description"`
	Logo        string     `json:"logo"`        // URL логотипа магазина
	Email       string     `json:"email"`       // Email магазина
	Phone       string     `json:"phone"`       // Телефон магазина
	Address     string     `json:"address"`     // Адрес магазина
	Rating      float64    `json:"rating" gorm:"default:0"` // Рейтинг магазина
	IsActive    bool       `json:"isActive" gorm:"default:true"`
	OwnerID     uuid.UUID  `json:"ownerId" gorm:"type:uuid;not null;index"` // ID владельца (User с ролью shop_owner)
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	// Связи
	Owner        *User              `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Products     []Product          `json:"products,omitempty" gorm:"foreignKey:ShopID"`
	Subscriptions []ShopSubscription `json:"subscriptions,omitempty" gorm:"foreignKey:ShopID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (s *Shop) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// ToShopInfo преобразует Shop в ShopInfo (для обратной совместимости)
func (s *Shop) ToShopInfo() ShopInfo {
	return ShopInfo{
		ID:   s.ID,
		Name: s.Name,
		INN:  s.INN,
	}
}

// ShopResponse представляет ответ с информацией о магазине
type ShopResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	INN             string    `json:"inn"`
	Description     string    `json:"description"`
	Logo            string    `json:"logo"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Address         string    `json:"address"`
	Rating          float64   `json:"rating"`
	IsActive        bool      `json:"isActive"`
	OwnerID         uuid.UUID `json:"ownerId"`
	ProductsCount   int64     `json:"productsCount"`
	SubscribersCount int64    `json:"subscribersCount"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ToResponse преобразует Shop в ShopResponse
func (s *Shop) ToResponse() ShopResponse {
	return ShopResponse{
		ID:        s.ID,
		Name:      s.Name,
		INN:       s.INN,
		Description: s.Description,
		Logo:      s.Logo,
		Email:     s.Email,
		Phone:     s.Phone,
		Address:   s.Address,
		Rating:    s.Rating,
		IsActive:  s.IsActive,
		OwnerID:   s.OwnerID,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

