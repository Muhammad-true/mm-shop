package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivationType тип активации лицензии
type ActivationType string

const (
	ActivationTypeManual  ActivationType = "manual"  // Ручная активация
	ActivationTypePayment ActivationType = "payment" // Активация через оплату
)

// SubscriptionType тип подписки
type SubscriptionType string

const (
	SubscriptionTypeMonthly  SubscriptionType = "monthly"  // Месячная подписка
	SubscriptionTypeYearly   SubscriptionType = "yearly"   // Годовая подписка
	SubscriptionTypeLifetime SubscriptionType = "lifetime" // Пожизненная подписка
)

// SubscriptionStatus статус подписки
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"    // Активна
	SubscriptionStatusExpired   SubscriptionStatus = "expired"   // Истекла
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled" // Отменена
	SubscriptionStatusPending   SubscriptionStatus = "pending"   // Ожидает оплаты
)

// License представляет лицензию для магазина
type License struct {
	ID                    uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;"`
	LicenseKey            string            `json:"licenseKey" gorm:"uniqueIndex;not null"` // Уникальный ключ лицензии
	ShopID                *uuid.UUID        `json:"shopId" gorm:"type:uuid;index"`          // ID магазина (может быть null до активации)
	ActivationType        ActivationType    `json:"activationType" gorm:"type:varchar(20);default:'manual'"` // Тип активации
	SubscriptionType      SubscriptionType  `json:"subscriptionType" gorm:"type:varchar(20);not null"`       // Тип подписки
	SubscriptionStatus    SubscriptionStatus `json:"subscriptionStatus" gorm:"type:varchar(20);default:'pending'"` // Статус подписки
	ActivatedAt           *time.Time        `json:"activatedAt"`                          // Дата активации
	ExpiresAt             *time.Time        `json:"expiresAt"`                            // Дата окончания
	LastPaymentDate       *time.Time        `json:"lastPaymentDate"`                      // Дата последней оплаты
	NextPaymentDate       *time.Time        `json:"nextPaymentDate"`                      // Дата следующей оплаты
	PaymentProvider       string            `json:"paymentProvider" gorm:"type:varchar(50)"` // Провайдер оплаты (stripe, paypal, etc.)
	PaymentTransactionID  string            `json:"paymentTransactionId" gorm:"type:varchar(255)"` // ID транзакции
	PaymentAmount         float64           `json:"paymentAmount" gorm:"type:decimal(10,2)"`       // Сумма оплаты
	PaymentCurrency       string            `json:"paymentCurrency" gorm:"type:varchar(3);default:'USD'"` // Валюта
	UserID                *uuid.UUID        `json:"userId" gorm:"type:uuid;index"`        // ID пользователя (владельца магазина)
	IsActive              bool              `json:"isActive" gorm:"default:true"`         // Активна ли лицензия
	AutoRenew             bool              `json:"autoRenew" gorm:"default:false"`       // Автопродление
	Notes                 string            `json:"notes" gorm:"type:text"`               // Заметки
	DeviceID              string            `json:"deviceId" gorm:"type:varchar(255);index"` // Уникальный ID устройства
	DeviceInfo            string            `json:"deviceInfo" gorm:"type:text"`          // JSON с информацией о железе
	DeviceFingerprint     string            `json:"deviceFingerprint" gorm:"type:varchar(255);index"` // Хеш для быстрой проверки
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`

	// Связи
	Shop *Shop `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate устанавливает UUID и генерирует ключ лицензии перед созданием
func (l *License) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	if l.LicenseKey == "" {
		l.LicenseKey = generateLicenseKey()
	}
	return nil
}

// generateLicenseKey генерирует уникальный ключ лицензии
func generateLicenseKey() string {
	// Формат: XXXX-XXXX-XXXX-XXXX (16 символов, разделенные дефисами)
	key := uuid.New().String()
	// Убираем дефисы
	key = key[:8] + key[9:13] + key[14:18] + key[19:23] + key[24:36]
	// Берем первые 16 символов и делаем верхний регистр
	key = key[:16]
	// Форматируем как XXXX-XXXX-XXXX-XXXX
	formatted := key[:4] + "-" + key[4:8] + "-" + key[8:12] + "-" + key[12:16]
	return formatted
}

// IsExpired проверяет, истекла ли лицензия
func (l *License) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false // Пожизненная лицензия
	}
	return time.Now().After(*l.ExpiresAt)
}

// IsValid проверяет, действительна ли лицензия
func (l *License) IsValid() bool {
	if !l.IsActive {
		return false
	}
	if l.SubscriptionStatus != SubscriptionStatusActive {
		return false
	}
	return !l.IsExpired()
}

// CalculateExpirationDate вычисляет дату окончания на основе типа подписки
func (l *License) CalculateExpirationDate(startDate time.Time) *time.Time {
	var expirationDate time.Time
	switch l.SubscriptionType {
	case SubscriptionTypeMonthly:
		expirationDate = startDate.AddDate(0, 1, 0) // +1 месяц
	case SubscriptionTypeYearly:
		expirationDate = startDate.AddDate(1, 0, 0) // +1 год
	case SubscriptionTypeLifetime:
		return nil // Пожизненная лицензия
	}
	return &expirationDate
}

// LicenseResponse представляет ответ с информацией о лицензии
type LicenseResponse struct {
	ID                    uuid.UUID         `json:"id"`
	LicenseKey            string            `json:"licenseKey"`
	ShopID                *uuid.UUID        `json:"shopId"`
	ActivationType        ActivationType    `json:"activationType"`
	SubscriptionType      SubscriptionType  `json:"subscriptionType"`
	SubscriptionStatus    SubscriptionStatus `json:"subscriptionStatus"`
	ActivatedAt           *time.Time        `json:"activatedAt"`
	ExpiresAt             *time.Time        `json:"expiresAt"`
	LastPaymentDate       *time.Time        `json:"lastPaymentDate"`
	NextPaymentDate       *time.Time        `json:"nextPaymentDate"`
	PaymentProvider       string            `json:"paymentProvider"`
	PaymentTransactionID  string            `json:"paymentTransactionId"`
	PaymentAmount         float64           `json:"paymentAmount"`
	PaymentCurrency       string            `json:"paymentCurrency"`
	UserID                *uuid.UUID        `json:"userId"`
	IsActive              bool              `json:"isActive"`
	AutoRenew             bool              `json:"autoRenew"`
	Notes                 string            `json:"notes"`
	DeviceID              string            `json:"deviceId"`              // ID устройства, на котором активирована лицензия
	DeviceInfo            string            `json:"deviceInfo"`            // JSON с информацией о железе
	IsValid               bool              `json:"isValid"`               // Вычисляемое поле
	IsExpired             bool              `json:"isExpired"`             // Вычисляемое поле
	DaysRemaining         *int              `json:"daysRemaining"`         // Оставшиеся дни (null для lifetime)
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`
	Shop                  *ShopResponse     `json:"shop,omitempty"`
}

// ToResponse преобразует License в LicenseResponse
func (l *License) ToResponse() LicenseResponse {
	var daysRemaining *int
	if l.ExpiresAt != nil {
		days := int(time.Until(*l.ExpiresAt).Hours() / 24)
		if days > 0 {
			daysRemaining = &days
		} else {
			zero := 0
			daysRemaining = &zero
		}
	}

	var shopResponse *ShopResponse
	if l.Shop != nil {
		resp := l.Shop.ToResponse()
		shopResponse = &resp
	}

	return LicenseResponse{
		ID:                   l.ID,
		LicenseKey:           l.LicenseKey,
		ShopID:               l.ShopID,
		ActivationType:       l.ActivationType,
		SubscriptionType:     l.SubscriptionType,
		SubscriptionStatus:   l.SubscriptionStatus,
		ActivatedAt:          l.ActivatedAt,
		ExpiresAt:            l.ExpiresAt,
		LastPaymentDate:      l.LastPaymentDate,
		NextPaymentDate:      l.NextPaymentDate,
		PaymentProvider:      l.PaymentProvider,
		PaymentTransactionID: l.PaymentTransactionID,
		PaymentAmount:        l.PaymentAmount,
		PaymentCurrency:      l.PaymentCurrency,
		UserID:               l.UserID,
		IsActive:             l.IsActive,
		AutoRenew:            l.AutoRenew,
		Notes:                l.Notes,
		IsValid:              l.IsValid(),
		IsExpired:            l.IsExpired(),
		DaysRemaining:        daysRemaining,
		CreatedAt:            l.CreatedAt,
		UpdatedAt:            l.UpdatedAt,
		Shop:                 shopResponse,
	}
}

// LicenseActivationRequest запрос на активацию лицензии
type LicenseActivationRequest struct {
	LicenseKey string                 `json:"licenseKey" binding:"required"`
	ShopID     string                 `json:"shopId" binding:"required"` // UUID магазина
	DeviceID   string                 `json:"deviceId" binding:"required"` // Уникальный ID устройства
	DeviceInfo map[string]interface{} `json:"deviceInfo" binding:"required"` // Информация о железе
}

// LicenseCheckRequest запрос на проверку лицензии
type LicenseCheckRequest struct {
	LicenseKey string                 `json:"licenseKey" binding:"required"`
	DeviceID   string                 `json:"deviceId" binding:"required"` // Уникальный ID устройства
	DeviceInfo map[string]interface{} `json:"deviceInfo"`                  // Опционально, для проверки
}

