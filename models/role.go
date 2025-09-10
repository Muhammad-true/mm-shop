package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role представляет роль пользователя в системе
type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"` // admin, shop_owner, user
	DisplayName string    `json:"displayName" gorm:"not null"`      // Администратор, Владелец магазина, Пользователь
	Description string    `json:"description"`                      // Описание роли
	Permissions string    `json:"permissions"`                      // JSON строка с правами доступа
	IsActive    bool      `json:"isActive" gorm:"default:true"`
	IsSystem    bool      `json:"isSystem" gorm:"default:false"` // Системная роль (нельзя удалить)
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Связи
	Users []User `json:"users,omitempty" gorm:"foreignKey:RoleID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// RoleResponse представляет ответ с информацией о роли
type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	Permissions string    `json:"permissions"`
	IsActive    bool      `json:"isActive"`
	IsSystem    bool      `json:"isSystem"`
	UserCount   int       `json:"userCount"` // Количество пользователей с этой ролью
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ToResponse преобразует Role в RoleResponse
func (r *Role) ToResponse() RoleResponse {
	return RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		Permissions: r.Permissions,
		IsActive:    r.IsActive,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// RoleCreateRequest представляет запрос на создание роли
type RoleCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"displayName" binding:"required"`
	Description string `json:"description"`
	Permissions string `json:"permissions"`
}

// RoleUpdateRequest представляет запрос на обновление роли
type RoleUpdateRequest struct {
	DisplayName *string `json:"displayName"`
	Description *string `json:"description"`
	Permissions *string `json:"permissions"`
	IsActive    *bool   `json:"isActive"`
}
