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
	PushNotifications    bool       `json:"pushNotifications" gorm:"default:true"`
	SelectedCityID       *uuid.UUID `json:"selectedCityId" gorm:"type:uuid;index"` // Выбранный город пользователя
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`

	// Связи
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	City *City `json:"city,omitempty" gorm:"foreignKey:SelectedCityID"`
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
	Language             string  `json:"language" binding:"oneof=ru en"`
	Theme                string  `json:"theme" binding:"oneof=system light dark"`
	NotificationsEnabled *bool   `json:"notificationsEnabled"`
	EmailNotifications   *bool   `json:"emailNotifications"`
	PushNotifications    *bool   `json:"pushNotifications"`
	SelectedCityID       *string `json:"selectedCityId"` // ID выбранного города
}

// SettingsResponse представляет ответ с настройками
type SettingsResponse struct {
	Language             string        `json:"language"`
	Theme                string        `json:"theme"`
	NotificationsEnabled bool          `json:"notificationsEnabled"`
	EmailNotifications   bool          `json:"emailNotifications"`
	PushNotifications    bool          `json:"pushNotifications"`
	SelectedCityID       *uuid.UUID    `json:"selectedCityId"`
	City                 *CityResponse `json:"city,omitempty"` // Информация о выбранном городе
	UpdatedAt            time.Time     `json:"updatedAt"`
}

// ToResponse преобразует UserSettings в SettingsResponse
func (s *UserSettings) ToResponse() SettingsResponse {
	response := SettingsResponse{
		Language:             s.Language,
		Theme:                s.Theme,
		NotificationsEnabled: s.NotificationsEnabled,
		EmailNotifications:   s.EmailNotifications,
		PushNotifications:    s.PushNotifications,
		SelectedCityID:       s.SelectedCityID,
		UpdatedAt:            s.UpdatedAt,
	}
	
	// Если есть информация о городе, добавляем её
	if s.City != nil {
		cityResp := s.City.ToResponse()
		response.City = &cityResp
	}
	
	return response
}
