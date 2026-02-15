package models

import (
	"fmt"
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
	ID                uuid.UUID `json:"id"`
	ShopClientID      uuid.UUID `json:"shopClientId"`
	PreviousAmount    int       `json:"previousAmount"`    // В дирамах
	NewAmount         int       `json:"newAmount"`         // В дирамах
	ChangeAmount      int       `json:"changeAmount"`      // В дирамах
	PreviousFormatted string    `json:"previousFormatted"` // Форматированная строка: "X с Y д"
	NewFormatted      string    `json:"newFormatted"`      // Форматированная строка: "X с Y д"
	ChangeFormatted   string    `json:"changeFormatted"`   // Форматированная строка: "+X с Y д" или "-X с Y д"
	CreatedAt         time.Time `json:"createdAt"`
}

// ToResponse преобразует BonusHistory в BonusHistoryResponse
func (bh *BonusHistory) ToResponse() BonusHistoryResponse {
	// Форматируем предыдущее количество
	prevSom := bh.PreviousAmount / 100
	prevDiram := bh.PreviousAmount % 100
	var prevFormatted string
	if prevDiram > 0 {
		prevFormatted = fmt.Sprintf("%d с %d д", prevSom, prevDiram)
	} else {
		prevFormatted = fmt.Sprintf("%d с", prevSom)
	}

	// Форматируем новое количество
	newSom := bh.NewAmount / 100
	newDiram := bh.NewAmount % 100
	var newFormatted string
	if newDiram > 0 {
		newFormatted = fmt.Sprintf("%d с %d д", newSom, newDiram)
	} else {
		newFormatted = fmt.Sprintf("%d с", newSom)
	}

	// Форматируем изменение
	changeAbs := bh.ChangeAmount
	if changeAbs < 0 {
		changeAbs = -changeAbs
	}
	changeSom := changeAbs / 100
	changeDiram := changeAbs % 100
	var changeFormatted string
	if changeDiram > 0 {
		if bh.ChangeAmount >= 0 {
			changeFormatted = fmt.Sprintf("+%d с %d д", changeSom, changeDiram)
		} else {
			changeFormatted = fmt.Sprintf("-%d с %d д", changeSom, changeDiram)
		}
	} else {
		if bh.ChangeAmount >= 0 {
			changeFormatted = fmt.Sprintf("+%d с", changeSom)
		} else {
			changeFormatted = fmt.Sprintf("-%d с", changeSom)
		}
	}

	return BonusHistoryResponse{
		ID:                bh.ID,
		ShopClientID:      bh.ShopClientID,
		PreviousAmount:    bh.PreviousAmount,
		NewAmount:         bh.NewAmount,
		ChangeAmount:      bh.ChangeAmount,
		PreviousFormatted: prevFormatted,
		NewFormatted:      newFormatted,
		ChangeFormatted:   changeFormatted,
		CreatedAt:         bh.CreatedAt,
	}
}

