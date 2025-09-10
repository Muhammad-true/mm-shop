package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Address представляет адрес пользователя
type Address struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Street    string    `json:"street" gorm:"not null"`
	City      string    `json:"city" gorm:"not null"`
	State     string    `json:"state" gorm:"not null"`
	ZipCode   string    `json:"zipCode" gorm:"not null"`
	Country   string    `json:"country" gorm:"not null"`
	Apartment string    `json:"apartment"`
	Building  string    `json:"building"`
	Entrance  string    `json:"entrance"`
	Floor     string    `json:"floor"`
	Intercom  string    `json:"intercom"`
	IsDefault bool      `json:"isDefault" gorm:"default:false"`
	Label     string    `json:"label" gorm:"default:'Другое'"` // Дом/Работа/Другое
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Связи
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (a *Address) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// AddressRequest представляет запрос на создание/обновление адреса
type AddressRequest struct {
	Street    string `json:"street" binding:"required"`
	City      string `json:"city" binding:"required"`
	State     string `json:"state" binding:"required"`
	ZipCode   string `json:"zipCode" binding:"required"`
	Country   string `json:"country" binding:"required"`
	Apartment string `json:"apartment"`
	Building  string `json:"building"`
	Entrance  string `json:"entrance"`
	Floor     string `json:"floor"`
	Intercom  string `json:"intercom"`
	IsDefault bool   `json:"isDefault"`
	Label     string `json:"label" binding:"oneof=Дом Работа Другое"`
}

// AddressResponse представляет ответ с адресом
type AddressResponse struct {
	ID        uuid.UUID `json:"id"`
	Street    string    `json:"street"`
	City      string    `json:"city"`
	State     string    `json:"state"`
	ZipCode   string    `json:"zipCode"`
	Country   string    `json:"country"`
	Apartment string    `json:"apartment"`
	Building  string    `json:"building"`
	Entrance  string    `json:"entrance"`
	Floor     string    `json:"floor"`
	Intercom  string    `json:"intercom"`
	IsDefault bool      `json:"isDefault"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToResponse преобразует Address в AddressResponse
func (a *Address) ToResponse() AddressResponse {
	return AddressResponse{
		ID:        a.ID,
		Street:    a.Street,
		City:      a.City,
		State:     a.State,
		ZipCode:   a.ZipCode,
		Country:   a.Country,
		Apartment: a.Apartment,
		Building:  a.Building,
		Entrance:  a.Entrance,
		Floor:     a.Floor,
		Intercom:  a.Intercom,
		IsDefault: a.IsDefault,
		Label:     a.Label,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// FullAddress возвращает полный адрес строкой
func (a *Address) FullAddress() string {
	address := a.Country + ", " + a.State + ", " + a.City + ", " + a.Street
	if a.Building != "" {
		address += ", д. " + a.Building
	}
	if a.Apartment != "" {
		address += ", кв. " + a.Apartment
	}
	return address
}
