package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceToken представляет токен устройства для push-уведомлений
type DeviceToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	UserID    uuid.UUID `json:"userId" gorm:"type:uuid;not null;index"`
	Token     string    `json:"token" gorm:"not null;uniqueIndex;size:500"` // FCM токен или APNS токен
	Platform  string    `json:"platform" gorm:"not null"`                   // "ios", "android", "web"
	DeviceID  string    `json:"deviceId" gorm:"size:255"`                   // Уникальный ID устройства
	IsActive  bool      `json:"isActive" gorm:"default:true"`               // Активен ли токен
	LastUsed  time.Time `json:"lastUsed"`                                   // Последнее использование
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Связи
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (dt *DeviceToken) BeforeCreate(tx *gorm.DB) error {
	if dt.ID == uuid.Nil {
		dt.ID = uuid.New()
	}
	if dt.LastUsed.IsZero() {
		dt.LastUsed = time.Now()
	}
	return nil
}

// DeviceTokenRequest представляет запрос на регистрацию токена устройства
type DeviceTokenRequest struct {
	Token    string `json:"token" binding:"required"`    // FCM/APNS токен
	Platform string `json:"platform" binding:"required"` // "ios", "android", "web"
	DeviceID string `json:"deviceId"`                    // Опциональный ID устройства
}

