package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShopSubscription представляет подписку пользователя на магазин
type ShopSubscription struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	UserID    uuid.UUID `json:"userId" gorm:"type:uuid;not null;index"`
	ShopID    uuid.UUID `json:"shopId" gorm:"type:uuid;not null;index"` // ID пользователя с ролью shop_owner
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Связи
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Shop User `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (ss *ShopSubscription) BeforeCreate(tx *gorm.DB) error {
	if ss.ID == uuid.Nil {
		ss.ID = uuid.New()
	}
	return nil
}

// ShopSubscriptionResponse представляет ответ с информацией о подписке
type ShopSubscriptionResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"userId"`
	ShopID    uuid.UUID  `json:"shopId"`
	Shop      *ShopInfo  `json:"shop,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// ToResponse преобразует ShopSubscription в ShopSubscriptionResponse
func (ss *ShopSubscription) ToResponse() ShopSubscriptionResponse {
	response := ShopSubscriptionResponse{
		ID:        ss.ID,
		UserID:    ss.UserID,
		ShopID:    ss.ShopID,
		CreatedAt: ss.CreatedAt,
	}
	
	// Добавляем информацию о магазине, если она загружена
	if ss.Shop.ID != uuid.Nil {
		response.Shop = &ShopInfo{
			ID:   ss.Shop.ID,
			Name: ss.Shop.Name,
			INN:  ss.Shop.INN,
		}
	}
	
	return response
}

