package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserSettings представляет настройки пользователя
type UserSettings struct {
	ID                   uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	UserID               uuid.UUID `json:"userId" gorm:"type:uuid;not null;uniqueIndex"`
	Language             string    `json:"language" gorm:"default:'ru'"`  // ru/en
	Theme                string    `json:"theme" gorm:"default:'system'"` // system/light/dark
	NotificationsEnabled bool      `json:"notificationsEnabled" gorm:"default:true"`
	EmailNotifications   bool      `json:"emailNotifications" gorm:"default:true"`
	PushNotifications    bool      `json:"pushNotifications" gorm:"default:true"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`

	// Связи
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (s *UserSettings) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// SettingsRequest представляет запрос на обновление настроек
type SettingsRequest struct {
	Language             string `json:"language" binding:"oneof=ru en"`
	Theme                string `json:"theme" binding:"oneof=system light dark"`
	NotificationsEnabled *bool  `json:"notificationsEnabled"`
	EmailNotifications   *bool  `json:"emailNotifications"`
	PushNotifications    *bool  `json:"pushNotifications"`
}

// SettingsResponse представляет ответ с настройками
type SettingsResponse struct {
	Language             string    `json:"language"`
	Theme                string    `json:"theme"`
	NotificationsEnabled bool      `json:"notificationsEnabled"`
	EmailNotifications   bool      `json:"emailNotifications"`
	PushNotifications    bool      `json:"pushNotifications"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// ToResponse преобразует UserSettings в SettingsResponse
func (s *UserSettings) ToResponse() SettingsResponse {
	return SettingsResponse{
		Language:             s.Language,
		Theme:                s.Theme,
		NotificationsEnabled: s.NotificationsEnabled,
		EmailNotifications:   s.EmailNotifications,
		PushNotifications:    s.PushNotifications,
		UpdatedAt:            s.UpdatedAt,
	}
}
