package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType представляет тип уведомления
type NotificationType string

const (
	NotificationTypeOrder     NotificationType = "order"
	NotificationTypePromotion NotificationType = "promotion"
	NotificationTypeSystem    NotificationType = "system"
	NotificationTypeReminder  NotificationType = "reminder"
)

// Notification представляет уведомление пользователя
type Notification struct {
	ID        uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;"`
	UserID    uuid.UUID        `json:"userId" gorm:"type:uuid;not null"`
	Title     string           `json:"title" gorm:"not null"`
	Body      string           `json:"body" gorm:"not null"`
	Type      NotificationType `json:"type" gorm:"not null"`
	Timestamp time.Time        `json:"timestamp"`
	IsRead    bool             `json:"isRead" gorm:"default:false"`
	Data      string           `json:"data" gorm:"type:text"` // JSON строка
	ImageURL  string           `json:"imageUrl"`
	ActionURL string           `json:"actionUrl"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`

	// Связи
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate устанавливает UUID и timestamp перед созданием
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.Timestamp.IsZero() {
		n.Timestamp = time.Now()
	}
	return nil
}

// NotificationRequest представляет запрос на создание уведомления
type NotificationRequest struct {
	UserID    uuid.UUID              `json:"userId" binding:"required"`
	Title     string                 `json:"title" binding:"required"`
	Body      string                 `json:"body" binding:"required"`
	Type      NotificationType       `json:"type" binding:"required,oneof=order promotion system reminder"`
	Data      map[string]interface{} `json:"data"`
	ImageURL  string                 `json:"imageUrl"`
	ActionURL string                 `json:"actionUrl"`
}

// NotificationResponse представляет ответ с уведомлением
type NotificationResponse struct {
	ID        uuid.UUID              `json:"id"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Type      NotificationType       `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	IsRead    bool                   `json:"isRead"`
	Data      map[string]interface{} `json:"data,omitempty"`
	ImageURL  string                 `json:"imageUrl,omitempty"`
	ActionURL string                 `json:"actionUrl,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
}

// ToResponse преобразует Notification в NotificationResponse
func (n *Notification) ToResponse() NotificationResponse {
	response := NotificationResponse{
		ID:        n.ID,
		Title:     n.Title,
		Body:      n.Body,
		Type:      n.Type,
		Timestamp: n.Timestamp,
		IsRead:    n.IsRead,
		ImageURL:  n.ImageURL,
		ActionURL: n.ActionURL,
		CreatedAt: n.CreatedAt,
	}

	// Парсим JSON данные если есть
	if n.Data != "" {
		// Здесь можно добавить парсинг JSON, пока оставляем nil
		response.Data = nil
	}

	return response
}
