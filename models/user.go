package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User представляет пользователя системы
type User struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"`
	Name            string     `json:"name" gorm:"not null"`
	Email           string     `json:"email" gorm:"uniqueIndex;not null"`
	Password        string     `json:"-" gorm:"not null"` // "-" исключает из JSON
	Phone           string     `json:"phone"`
	Avatar          string     `json:"avatar"` // URL аватара
	DateOfBirth     *time.Time `json:"dateOfBirth"`
	Gender          string     `json:"gender"` // male/female/other
	IsEmailVerified bool       `json:"isEmailVerified" gorm:"default:false"`
	IsPhoneVerified bool       `json:"isPhoneVerified" gorm:"default:false"`
	IsActive        bool       `json:"isActive" gorm:"default:true"`
	IsGuest         bool       `json:"isGuest" gorm:"default:false"` // Флаг гостевого пользователя
	RoleID          *uuid.UUID `json:"roleId" gorm:"type:uuid"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`

	// Связи
	Role          *Role          `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Addresses     []Address      `json:"addresses,omitempty" gorm:"foreignKey:UserID"`
	Orders        []Order        `json:"orders,omitempty" gorm:"foreignKey:UserID"`
	CartItems     []CartItem     `json:"cartItems,omitempty" gorm:"foreignKey:UserID"`
	Favorites     []Favorite     `json:"favorites,omitempty" gorm:"foreignKey:UserID"`
	Notifications []Notification `json:"notifications,omitempty" gorm:"foreignKey:UserID"`
	// Settings      *UserSettings  `json:"settings,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate устанавливает UUID перед созданием
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// HashPassword хеширует пароль пользователя
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword проверяет пароль пользователя
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// UserRegisterRequest представляет запрос на регистрацию клиента (Flutter)
// Email опционален (для отправки чека), регистрация по телефону обязательна
type UserRegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required,min=8"`
}

// UserLoginRequest представляет запрос на вход по телефону
type UserLoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserForgotPasswordRequest запрос на восстановление пароля по телефону
type UserForgotPasswordRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// UserUpdateRequest представляет запрос на обновление профиля
type UserUpdateRequest struct {
	Name        *string    `json:"name"`
	Phone       *string    `json:"phone"`
	DateOfBirth *time.Time `json:"dateOfBirth"`
	Gender      *string    `json:"gender" binding:"omitempty,oneof=male female other"`
}

// UserResponse представляет ответ с информацией о пользователе
type UserResponse struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	Email           string            `json:"email"`
	Phone           string            `json:"phone"`
	Avatar          string            `json:"avatar"`
	DateOfBirth     *time.Time        `json:"dateOfBirth"`
	Gender          string            `json:"gender"`
	IsEmailVerified bool              `json:"isEmailVerified"`
	IsPhoneVerified bool              `json:"isPhoneVerified"`
	IsActive        bool              `json:"isActive"`
	IsGuest         bool              `json:"isGuest"`
	Role            *RoleResponse     `json:"role,omitempty"`
	Addresses       []AddressResponse `json:"addresses,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

// ToResponse преобразует User в UserResponse
func (u *User) ToResponse() UserResponse {
	addresses := make([]AddressResponse, len(u.Addresses))
	for i, addr := range u.Addresses {
		addresses[i] = addr.ToResponse()
	}

	var roleResponse *RoleResponse
	if u.Role != nil {
		role := u.Role.ToResponse()
		roleResponse = &role
	}

	return UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Phone:           u.Phone,
		Avatar:          u.Avatar,
		DateOfBirth:     u.DateOfBirth,
		Gender:          u.Gender,
		IsEmailVerified: u.IsEmailVerified,
		IsPhoneVerified: u.IsPhoneVerified,
		IsActive:        u.IsActive,
		IsGuest:         u.IsGuest,
		Role:            roleResponse,
		Addresses:       addresses,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}
