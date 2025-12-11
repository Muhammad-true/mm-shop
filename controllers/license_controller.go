package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LicenseController обрабатывает запросы лицензий
type LicenseController struct{}

// CheckLicense проверяет статус лицензии (публичный эндпоинт)
func (lc *LicenseController) CheckLicense(c *gin.Context) {
	var req models.LicenseCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	var license models.License
	if err := database.DB.Where("license_key = ?", req.LicenseKey).First(&license).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "License not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Возвращаем только публичную информацию
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"isValid":           license.IsValid(),
			"isExpired":         license.IsExpired(),
			"subscriptionStatus": license.SubscriptionStatus,
			"subscriptionType":  license.SubscriptionType,
			"expiresAt":         license.ExpiresAt,
			"daysRemaining":     license.ToResponse().DaysRemaining,
		},
	})
}

// ActivateLicense активирует лицензию для магазина (публичный эндпоинт)
func (lc *LicenseController) ActivateLicense(c *gin.Context) {
	var req models.LicenseActivationRequest
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
	if err := database.DB.First(&shop, shopID).Error; err != nil {
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

	// Находим лицензию
	var license models.License
	if err := database.DB.Where("license_key = ?", req.LicenseKey).First(&license).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "License not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Проверяем, не активирована ли уже лицензия
	if license.ShopID != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "License already activated",
		})
		return
	}

	// Проверяем валидность лицензии
	if !license.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "License is not active",
		})
		return
	}

	if license.SubscriptionStatus != models.SubscriptionStatusActive && license.SubscriptionStatus != models.SubscriptionStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "License is not available for activation",
		})
		return
	}

	// Активируем лицензию
	now := time.Now()
	license.ShopID = &shopID
	license.UserID = &shop.OwnerID
	license.ActivatedAt = &now
	license.SubscriptionStatus = models.SubscriptionStatusActive

	// Вычисляем дату окончания
	if license.ExpiresAt == nil {
		license.ExpiresAt = license.CalculateExpirationDate(now)
	}

	if err := database.DB.Save(&license).Error; err != nil {
		log.Printf("❌ Ошибка активации лицензии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to activate license",
		})
		return
	}

	// Загружаем связанные данные
	database.DB.Preload("Shop").Preload("User").First(&license, license.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "License activated successfully",
		"data":    license.ToResponse(),
	})
}

// GetLicenses возвращает список всех лицензий (админ)
func (lc *LicenseController) GetLicenses(c *gin.Context) {
	var licenses []models.License
	query := database.DB.Preload("Shop").Preload("User")

	// Фильтры
	if shopID := c.Query("shopId"); shopID != "" {
		if parsedID, err := uuid.Parse(shopID); err == nil {
			query = query.Where("shop_id = ?", parsedID)
		}
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("subscription_status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&licenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch licenses",
		})
		return
	}

	responses := make([]models.LicenseResponse, len(licenses))
	for i, license := range licenses {
		responses[i] = license.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"licenses": responses,
		},
	})
}

// GetLicense возвращает информацию о лицензии по ID (админ)
func (lc *LicenseController) GetLicense(c *gin.Context) {
	licenseIDParam := c.Param("id")
	licenseID, err := uuid.Parse(licenseIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid license ID",
		})
		return
	}

	var license models.License
	if err := database.DB.Preload("Shop").Preload("User").First(&license, licenseID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "License not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Database error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    license.ToResponse(),
	})
}

// CreateLicense создает новую лицензию (админ)
func (lc *LicenseController) CreateLicense(c *gin.Context) {
	var req struct {
		ShopID            *string                `json:"shopId"`
		SubscriptionType  models.SubscriptionType `json:"subscriptionType" binding:"required"`
		ActivationType    models.ActivationType   `json:"activationType"`
		PaymentAmount     float64                 `json:"paymentAmount"`
		PaymentCurrency   string                  `json:"paymentCurrency"`
		PaymentProvider   string                  `json:"paymentProvider"`
		PaymentTransactionID string              `json:"paymentTransactionId"`
		AutoRenew         bool                    `json:"autoRenew"`
		Notes             string                  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	license := models.License{
		SubscriptionType:     req.SubscriptionType,
		ActivationType:       req.ActivationType,
		SubscriptionStatus:   models.SubscriptionStatusPending,
		PaymentAmount:        req.PaymentAmount,
		PaymentCurrency:      req.PaymentCurrency,
		PaymentProvider:      req.PaymentProvider,
		PaymentTransactionID: req.PaymentTransactionID,
		AutoRenew:            req.AutoRenew,
		Notes:                req.Notes,
		IsActive:             true,
	}

	// Если передан ShopID, привязываем к магазину
	if req.ShopID != nil {
		shopID, err := uuid.Parse(*req.ShopID)
		if err == nil {
			var shop models.Shop
			if err := database.DB.First(&shop, shopID).Error; err == nil {
				license.ShopID = &shopID
				license.UserID = &shop.OwnerID
				now := time.Now()
				license.ActivatedAt = &now
				license.SubscriptionStatus = models.SubscriptionStatusActive
				license.ExpiresAt = license.CalculateExpirationDate(now)
			}
		}
	}

	// Если есть оплата, обновляем статус
	if req.PaymentAmount > 0 && req.PaymentTransactionID != "" {
		now := time.Now()
		license.SubscriptionStatus = models.SubscriptionStatusActive
		license.LastPaymentDate = &now
		license.NextPaymentDate = license.CalculateExpirationDate(now)
		if license.ExpiresAt == nil {
			license.ExpiresAt = license.NextPaymentDate
		}
	}

	if err := database.DB.Create(&license).Error; err != nil {
		log.Printf("❌ Ошибка создания лицензии: %v", err)
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
		"message": "License created successfully",
		"data":    license.ToResponse(),
	})
}

// GenerateLicenseForShop генерирует лицензию для магазина (админ)
func (lc *LicenseController) GenerateLicenseForShop(c *gin.Context) {
	shopIDParam := c.Param("shopId")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid shop ID",
		})
		return
	}

	var req struct {
		SubscriptionType models.SubscriptionType `json:"subscriptionType" binding:"required"`
		PaymentAmount    float64                 `json:"paymentAmount"`
		PaymentCurrency  string                  `json:"paymentCurrency"`
		PaymentProvider  string                  `json:"paymentProvider"`
		PaymentTransactionID string              `json:"paymentTransactionId"`
		AutoRenew        bool                    `json:"autoRenew"`
		Notes            string                  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Проверяем существование магазина
	var shop models.Shop
	if err := database.DB.First(&shop, shopID).Error; err != nil {
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

	// Создаем лицензию
	now := time.Now()
	license := models.License{
		ShopID:                &shopID,
		UserID:                &shop.OwnerID,
		SubscriptionType:      req.SubscriptionType,
		ActivationType:        models.ActivationTypePayment,
		SubscriptionStatus:    models.SubscriptionStatusActive,
		ActivatedAt:           &now,
		PaymentAmount:         req.PaymentAmount,
		PaymentCurrency:       req.PaymentCurrency,
		PaymentProvider:       req.PaymentProvider,
		PaymentTransactionID:  req.PaymentTransactionID,
		AutoRenew:             req.AutoRenew,
		Notes:                 req.Notes,
		IsActive:              true,
	}

	// Вычисляем дату окончания
	license.ExpiresAt = license.CalculateExpirationDate(now)
	license.NextPaymentDate = license.ExpiresAt

	if req.PaymentAmount > 0 {
		license.LastPaymentDate = &now
	}

	if err := database.DB.Create(&license).Error; err != nil {
		log.Printf("❌ Ошибка создания лицензии для магазина: %v", err)
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
		"message": "License generated successfully",
		"data":    license.ToResponse(),
	})
}

// UpdateLicense обновляет лицензию (админ)
func (lc *LicenseController) UpdateLicense(c *gin.Context) {
	licenseIDParam := c.Param("id")
	licenseID, err := uuid.Parse(licenseIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid license ID",
		})
		return
	}

	var license models.License
	if err := database.DB.First(&license, licenseID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "License not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Database error",
			})
		}
		return
	}

	var req struct {
		SubscriptionStatus *models.SubscriptionStatus `json:"subscriptionStatus"`
		IsActive           *bool                      `json:"isActive"`
		AutoRenew          *bool                      `json:"autoRenew"`
		Notes              string                     `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if req.SubscriptionStatus != nil {
		license.SubscriptionStatus = *req.SubscriptionStatus
	}
	if req.IsActive != nil {
		license.IsActive = *req.IsActive
	}
	if req.AutoRenew != nil {
		license.AutoRenew = *req.AutoRenew
	}
	if req.Notes != "" {
		license.Notes = req.Notes
	}

	if err := database.DB.Save(&license).Error; err != nil {
		log.Printf("❌ Ошибка обновления лицензии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update license",
		})
		return
	}

	// Загружаем связанные данные
	database.DB.Preload("Shop").Preload("User").First(&license, license.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "License updated successfully",
		"data":    license.ToResponse(),
	})
}

