package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionPlan представляет план подписки
type SubscriptionPlan struct {
	ID                    uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;"`
	Name                  string          `json:"name" gorm:"not null"`                    // Название плана (например, "Базовый", "Премиум")
	Description           string          `json:"description" gorm:"type:text"`            // Описание плана
	SubscriptionType      SubscriptionType `json:"subscriptionType" gorm:"type:varchar(20);not null"` // Тип подписки
	Price                 float64         `json:"price" gorm:"type:decimal(10,2);not null"` // Цена
	Currency              string          `json:"currency" gorm:"type:varchar(3);default:'USD'"` // Валюта
	DurationMonths        int             `json:"durationMonths" gorm:"default:1"`         // Длительность в месяцах (для monthly = 1, yearly = 12, lifetime = 0)
	IsActive              bool            `json:"isActive" gorm:"default:true"`            // Активен ли план
	Features              string          `json:"features" gorm:"type:text"`               // JSON с описанием возможностей
	SortOrder             int             `json:"sortOrder" gorm:"default:0"`              // Порядок сортировки
	LemonSqueezyVariantID string          `json:"lemonsqueezyVariantId" gorm:"type:varchar(255);index"` // ID варианта продукта из Lemon Squeezy
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

// BeforeCreate устанавливает UUID перед созданием
func (sp *SubscriptionPlan) BeforeCreate(tx *gorm.DB) error {
	if sp.ID == uuid.Nil {
		sp.ID = uuid.New()
	}
	return nil
}

// SubscriptionPlanResponse представляет ответ с информацией о плане подписки
type SubscriptionPlanResponse struct {
	ID                    uuid.UUID       `json:"id"`
	Name                  string          `json:"name"`
	Description           string          `json:"description"`
	SubscriptionType      SubscriptionType `json:"subscriptionType"`
	Price                 float64         `json:"price"`
	Currency              string          `json:"currency"`
	DurationMonths        int             `json:"durationMonths"`
	IsActive              bool            `json:"isActive"`
	Features              string          `json:"features"`
	SortOrder             int             `json:"sortOrder"`
	LemonSqueezyVariantID string          `json:"lemonsqueezyVariantId"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

// ToResponse преобразует SubscriptionPlan в SubscriptionPlanResponse
func (sp *SubscriptionPlan) ToResponse() SubscriptionPlanResponse {
	return SubscriptionPlanResponse{
		ID:                    sp.ID,
		Name:                  sp.Name,
		Description:           sp.Description,
		SubscriptionType:      sp.SubscriptionType,
		Price:                 sp.Price,
		Currency:              sp.Currency,
		DurationMonths:        sp.DurationMonths,
		IsActive:              sp.IsActive,
		Features:              sp.Features,
		SortOrder:             sp.SortOrder,
		LemonSqueezyVariantID: sp.LemonSqueezyVariantID,
		CreatedAt:             sp.CreatedAt,
		UpdatedAt:             sp.UpdatedAt,
	}
}

