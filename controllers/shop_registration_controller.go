package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/mm-api/mm-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShopRegistrationController обрабатывает публичную регистрацию магазинов
type ShopRegistrationController struct{}

// RegisterShop регистрирует новый магазин (публичный эндпоинт для сайта)
// Создает пользователя с ролью shop_owner и магазин
func (src *ShopRegistrationController) RegisterShop(c *gin.Context) {
	var req struct {
		// Данные пользователя
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Phone    string `json:"phone" binding:"required"`

		// Данные магазина
		ShopName    string  `json:"shopName" binding:"required"`
		INN         string  `json:"inn" binding:"required"`
		Description string  `json:"description"`
		Address     string  `json:"address"`
		CityID      *string `json:"cityId"` // ID города (опционально)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Проверяем, существует ли пользователь с таким email
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "User with this email already exists",
		})
		return
	}

	// Получаем роль shop_owner
	var shopOwnerRole models.Role
	if err := database.DB.Where("name = ?", "shop_owner").First(&shopOwnerRole).Error; err != nil {
		log.Printf("❌ Failed to find shop_owner role: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Shop owner role not found",
		})
		return
	}

	// Создаем пользователя
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		IsActive: true,
		RoleID:   &shopOwnerRole.ID,
	}

	// Хешируем пароль
	if err := user.HashPassword(req.Password); err != nil {
		log.Printf("❌ Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process password",
		})
		return
	}

	// Сохраняем пользователя
	if err := database.DB.Create(&user).Error; err != nil {
		log.Printf("❌ Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create user",
		})
		return
	}

	// Парсим CityID если передан
	var cityID *uuid.UUID
	if req.CityID != nil {
		if parsedCityID, err := uuid.Parse(*req.CityID); err == nil {
			// Проверяем существование города
			var city models.City
			if err := database.DB.First(&city, parsedCityID).Error; err == nil {
				cityID = &parsedCityID
			}
		}
	}

	// Создаем магазин
	shop := models.Shop{
		Name:        req.ShopName,
		INN:         req.INN,
		Description: req.Description,
		Address:     req.Address,
		Email:       req.Email,
		Phone:       req.Phone,
		IsActive:    true,
		OwnerID:     user.ID,
		CityID:      cityID,
	}

	if err := database.DB.Create(&shop).Error; err != nil {
		log.Printf("❌ Failed to create shop: %v", err)
		// Удаляем пользователя, если не удалось создать магазин
		database.DB.Delete(&user)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create shop",
		})
		return
	}

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user.ID, user.Email, "shop_owner")
	if err != nil {
		log.Printf("⚠️ Failed to generate token: %v", err)
		// Не прерываем, токен можно получить через login
	}

	// Загружаем связанные данные
	database.DB.Preload("Role").Preload("Shops").First(&user, user.ID)
	database.DB.Preload("City").First(&shop, shop.ID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Shop registered successfully",
		"data": gin.H{
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"phone": user.Phone,
				"role":  "shop_owner",
			},
			"shop": gin.H{
				"id":          shop.ID,
				"name":        shop.Name,
				"inn":         shop.INN,
				"description": shop.Description,
				"address":     shop.Address,
				"cityId":      shop.CityID,
			},
			"token": token, // Токен для автоматического входа
		},
	})
}

// SubscribeShop создает подписку (лицензию) для магазина (публичный эндпоинт для сайта)
// Вызывается после успешной оплаты
func (src *ShopRegistrationController) SubscribeShop(c *gin.Context) {
	var req struct {
		ShopID              string                      `json:"shopId" binding:"required"`
		SubscriptionPlanID  string                      `json:"subscriptionPlanId" binding:"required"`
		PaymentProvider     string                      `json:"paymentProvider" binding:"required"`     // stripe, paypal, etc.
		PaymentTransactionID string                     `json:"paymentTransactionId" binding:"required"` // ID транзакции от платежной системы
		PaymentAmount       float64                     `json:"paymentAmount" binding:"required"`
		PaymentCurrency     string                      `json:"paymentCurrency"` // По умолчанию USD
		AutoRenew           bool                        `json:"autoRenew"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Парсим ShopID
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid shop ID",
		})
		return
	}

	// Проверяем существование магазина
	var shop models.Shop
	if err := database.DB.Preload("Owner").First(&shop, shopID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Shop not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Получаем план подписки
	planID, err := uuid.Parse(req.SubscriptionPlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid subscription plan ID",
		})
		return
	}

	var plan models.SubscriptionPlan
	if err := database.DB.First(&plan, planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Subscription plan not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Проверяем, нет ли уже активной лицензии для этого магазина
	var existingLicense models.License
	if err := database.DB.Where("shop_id = ? AND subscription_status = ?", shopID, models.SubscriptionStatusActive).First(&existingLicense).Error; err == nil {
		// Если есть активная лицензия, проверяем, не истекла ли она
		if !existingLicense.IsExpired() {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Shop already has an active license",
				"data": gin.H{
					"licenseId": existingLicense.ID,
					"expiresAt": existingLicense.ExpiresAt,
				},
			})
			return
		}
	}

	// Создаем лицензию
	now := time.Now()
	currency := req.PaymentCurrency
	if currency == "" {
		currency = plan.Currency
	}

	license := models.License{
		ShopID:                &shopID,
		UserID:                &shop.OwnerID,
		SubscriptionType:      plan.SubscriptionType,
		ActivationType:        models.ActivationTypePayment,
		SubscriptionStatus:    models.SubscriptionStatusActive,
		ActivatedAt:           &now,
		PaymentAmount:         req.PaymentAmount,
		PaymentCurrency:       currency,
		PaymentProvider:       req.PaymentProvider,
		PaymentTransactionID:  req.PaymentTransactionID,
		LastPaymentDate:       &now,
		AutoRenew:             req.AutoRenew,
		IsActive:              true,
	}

	// Вычисляем дату окончания
	license.ExpiresAt = license.CalculateExpirationDate(now)
	license.NextPaymentDate = license.ExpiresAt

	if err := database.DB.Create(&license).Error; err != nil {
		log.Printf("❌ Failed to create license: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create license",
		})
		return
	}

	// Загружаем связанные данные
	database.DB.Preload("Shop").Preload("User").First(&license, license.ID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Subscription created successfully",
		"data":    license.ToResponse(),
	})
}

