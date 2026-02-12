package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BonusHistory представляет историю изменений бонусов клиента в магазине
type BonusHistory struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	ShopClientID uuid.UUID `json:"shopClientId" gorm:"type:uuid;not null;index"`
	PreviousAmount int     `json:"previousAmount"` // Предыдущее количество бонусов
	NewAmount      int     `json:"newAmount"`      // Новое количество бонусов
	ChangeAmount   int     `json:"changeAmount"`   // Изменение (может быть положительным или отрицательным)
	CreatedAt      time.Time `json:"createdAt"`

	// Связи
	ShopClient ShopClient `json:"shopClient,omitempty" gorm:"foreignKey:ShopClientID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (bh *BonusHistory) BeforeCreate(tx *gorm.DB) error {
	if bh.ID == uuid.Nil {
		bh.ID = uuid.New()
	}
	return nil
}

// BonusHistoryResponse представляет ответ с информацией об истории бонусов
type BonusHistoryResponse struct {
	ID            uuid.UUID `json:"id"`
	ShopClientID  uuid.UUID `json:"shopClientId"`
	PreviousAmount int      `json:"previousAmount"`
	NewAmount      int      `json:"newAmount"`
	ChangeAmount   int      `json:"changeAmount"`
	CreatedAt      time.Time `json:"createdAt"`
}

// ToResponse преобразует BonusHistory в BonusHistoryResponse
func (bh *BonusHistory) ToResponse() BonusHistoryResponse {
	return BonusHistoryResponse{
		ID:            bh.ID,
		ShopClientID:  bh.ShopClientID,
		PreviousAmount: bh.PreviousAmount,
		NewAmount:     bh.NewAmount,
		ChangeAmount:  bh.ChangeAmount,
		CreatedAt:     bh.CreatedAt,
	}
}

