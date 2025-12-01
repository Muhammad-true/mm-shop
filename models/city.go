package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// City представляет город в системе
type City struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Name      string    `json:"name" gorm:"not null;index"`
	Latitude  float64   `json:"latitude" gorm:"not null"`  // Широта
	Longitude float64   `json:"longitude" gorm:"not null"` // Долгота
	IsActive  bool      `json:"isActive" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Связи
	Shops    []Shop    `json:"shops,omitempty" gorm:"foreignKey:CityID"`
	Products []Product `json:"products,omitempty" gorm:"foreignKey:CityID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (c *City) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CityResponse представляет ответ с информацией о городе
type CityResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToResponse преобразует City в CityResponse
func (c *City) ToResponse() CityResponse {
	return CityResponse{
		ID:        c.ID,
		Name:      c.Name,
		Latitude:  c.Latitude,
		Longitude: c.Longitude,
		IsActive:  c.IsActive,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

