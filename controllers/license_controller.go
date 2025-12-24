package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mm-api/mm-api/config"
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

	// Очищаем deviceId от лишних пробелов и переносов строк
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.LicenseKey = strings.TrimSpace(req.LicenseKey)

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

	// Проверяем, активирована ли лицензия
	if license.ShopID == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"isValid":            false,
				"isExpired":          false,
				"subscriptionStatus": license.SubscriptionStatus,
				"subscriptionType":   license.SubscriptionType,
				"isActivated":        false,
				"message":            "License not activated yet",
			},
		})
		return
	}

	// Проверяем соответствие устройства
	deviceMatch := false
	storedDeviceID := strings.TrimSpace(license.DeviceID)
	if storedDeviceID != "" {
		deviceMatch = storedDeviceID == req.DeviceID
		if !deviceMatch && req.DeviceInfo != nil {
			// Дополнительная проверка по fingerprint
			fingerprint := generateDeviceFingerprint(req.DeviceID, req.DeviceInfo)
			deviceMatch = license.DeviceFingerprint == fingerprint
		}
	}

	// Возвращаем информацию о лицензии
	response := gin.H{
		"isValid":            license.IsValid() && deviceMatch,
		"isExpired":          license.IsExpired(),
		"subscriptionStatus": license.SubscriptionStatus,
		"subscriptionType":   license.SubscriptionType,
		"expiresAt":          license.ExpiresAt,
		"daysRemaining":      license.ToResponse().DaysRemaining,
		"deviceMatch":        deviceMatch,
	}

	if !deviceMatch && license.DeviceID != "" {
		response["error"] = "License is activated on a different device"
		response["isValid"] = false
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
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

	// Очищаем deviceId от лишних пробелов и переносов строк
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.LicenseKey = strings.TrimSpace(req.LicenseKey)
	req.ShopID = strings.TrimSpace(req.ShopID)

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

	// Находим лицензию по license_key
	var license models.License
	if err := database.DB.Where("license_key = ?", req.LicenseKey).First(&license).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("❌ Лицензия не найдена по ключу: %s", req.LicenseKey)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "License not found",
			})
			return
		}
		log.Printf("❌ Ошибка БД при поиске лицензии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	log.Printf("🔍 Найдена лицензия: ID=%s, ShopID=%v, DeviceID='%s', Status=%s, IsActive=%v, SubscriptionType=%s",
		license.ID, license.ShopID, license.DeviceID, license.SubscriptionStatus, license.IsActive, license.SubscriptionType)

	// Если лицензия уже активирована для другого магазина, запрещаем активацию
	if license.ShopID != nil && *license.ShopID != shopID {
		log.Printf("❌ Лицензия уже активирована для другого магазина: %v (запрошен: %v)", license.ShopID, shopID)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "License is already activated for a different shop",
			"data": gin.H{
				"shopId": license.ShopID,
			},
		})
		return
	}

	// Проверяем device_id
	storedDeviceID := strings.TrimSpace(license.DeviceID)
	log.Printf("🔍 Проверка устройства: storedDeviceID='%s', reqDeviceID='%s'", storedDeviceID, req.DeviceID)

	// Если device_id пустой - можно активировать (записываем данные)
	if storedDeviceID == "" {
		log.Printf("✅ DeviceID пустой - можно активировать лицензию и записать данные устройства")
	} else if storedDeviceID == req.DeviceID {
		// Лицензия уже активирована на этом устройстве
		log.Printf("✅ Лицензия уже активирована на этом устройстве")
		database.DB.Preload("Shop").Preload("User").First(&license, license.ID)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "License already activated on this device",
			"data":    license.ToResponse(),
		})
		return
	} else {
		// device_id не пустой и не совпадает - это другой компьютер
		// Разрешаем переактивацию на новом устройстве (обновляем device_id)
		log.Printf("🔄 DeviceID не совпадает - это другой компьютер. Разрешаем переактивацию.")
		log.Printf("   Старое устройство: '%s' -> Новое устройство: '%s'", storedDeviceID, req.DeviceID)
	}

	// Проверяем валидность лицензии
	if !license.IsActive {
		log.Printf("❌ Лицензия неактивна: IsActive=%v", license.IsActive)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "License is not active",
		})
		return
	}

	if license.SubscriptionStatus != models.SubscriptionStatusActive && license.SubscriptionStatus != models.SubscriptionStatusPending {
		log.Printf("❌ Лицензия недоступна для активации: Status=%s (ожидается: active или pending)", license.SubscriptionStatus)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "License is not available for activation",
			"details": gin.H{
				"subscriptionStatus": license.SubscriptionStatus,
				"expectedStatus":     []string{"active", "pending"},
			},
		})
		return
	}

	log.Printf("✅ Лицензия прошла проверки, начинаем активацию...")

	// Генерируем fingerprint устройства
	deviceFingerprint := generateDeviceFingerprint(req.DeviceID, req.DeviceInfo)

	// Сохраняем информацию об устройстве в JSON
	deviceInfoJSON, err := json.Marshal(req.DeviceInfo)
	if err != nil {
		log.Printf("⚠️ Failed to marshal device info: %v", err)
		deviceInfoJSON = []byte("{}")
	}

	// Активируем или переактивируем лицензию
	now := time.Now()
	wasAlreadyActivated := license.ShopID != nil

	// Если лицензия уже была активирована, обновляем информацию об устройстве
	if !wasAlreadyActivated {
		license.ShopID = &shopID
		license.UserID = &shop.OwnerID
		license.ActivatedAt = &now
	}

	license.SubscriptionStatus = models.SubscriptionStatusActive
	license.DeviceID = req.DeviceID // Уже обрезан выше
	license.DeviceInfo = string(deviceInfoJSON)
	license.DeviceFingerprint = deviceFingerprint

	// Вычисляем дату окончания
	if license.ExpiresAt == nil {
		license.ExpiresAt = license.CalculateExpirationDate(now)
	}

	if err := database.DB.Save(&license).Error; err != nil {
		log.Printf("❌ Ошибка сохранения лицензии в БД: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to activate license",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Лицензия успешно сохранена в БД: ShopID=%v, DeviceID='%s', Status=%s",
		license.ShopID, license.DeviceID, license.SubscriptionStatus)

	// Загружаем связанные данные
	if err := database.DB.Preload("Shop").Preload("User").First(&license, license.ID).Error; err != nil {
		log.Printf("⚠️ Ошибка загрузки связанных данных: %v", err)
		// Продолжаем, даже если не удалось загрузить связанные данные
	}

	message := "License activated successfully"
	if wasAlreadyActivated {
		message = "License reactivated on new device successfully"
	}

	log.Printf("✅ Активация завершена успешно: %s", message)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    license.ToResponse(),
	})
}

// GetMyLicenses возвращает список лицензий текущего пользователя
// Автоматически синхронизирует подписки из Lemon Squeezy, если лицензий нет
func (lc *LicenseController) GetMyLicenses(c *gin.Context) {
	// Получаем пользователя из контекста
	userIDValue, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	userID := userIDValue.(uuid.UUID)

	// Получаем пользователя из БД
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Получаем все лицензии пользователя
	var licenses []models.License
	if err := database.DB.Where("user_id = ?", userID).
		Preload("Shop").
		Order("created_at DESC").
		Find(&licenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch licenses",
		})
		return
	}

	// Если лицензий нет, пытаемся синхронизировать из Lemon Squeezy
	if len(licenses) == 0 {
		log.Printf("🔄 [GetMyLicenses] Лицензий не найдено, пытаемся синхронизировать из Lemon Squeezy для пользователя: %s", user.Email)
		
		// Используем ShopRegistrationController для синхронизации
		shopRegistrationController := &ShopRegistrationController{}
		
		// Вызываем синхронизацию (без создания лицензий, только проверка)
		// Но лучше просто вызвать метод синхронизации напрямую
		// Для этого нужно получить подписки и создать лицензии
		cfg := config.GetConfig()
		if cfg.LemonSqueezyAPIKey != "" {
			subscriptions, err := shopRegistrationController.getLemonSqueezySubscriptionsByEmail(user.Email, cfg.LemonSqueezyAPIKey)
			if err == nil && len(subscriptions) > 0 {
				log.Printf("✅ [GetMyLicenses] Найдено подписок в Lemon Squeezy: %d, синхронизируем...", len(subscriptions))
				
				// Получаем магазины пользователя
				var shops []models.Shop
				if err := database.DB.Where("owner_id = ?", user.ID).Find(&shops).Error; err == nil && len(shops) > 0 {
					// Синхронизируем каждую подписку
					for _, sub := range subscriptions {
						variantID := shopRegistrationController.extractVariantIDFromSubscription(sub)
						if variantID == "" {
							continue
						}

						// Находим план подписки
						var plan models.SubscriptionPlan
						if err := database.DB.Where("lemonsqueezy_variant_id = ?", variantID).First(&plan).Error; err != nil {
							continue
						}

						// Используем первый магазин
						targetShop := &shops[0]

						// Проверяем, нет ли уже лицензии
						var existingLicense models.License
						if err := database.DB.Where("shop_id = ? AND subscription_status = ?", targetShop.ID, models.SubscriptionStatusActive).First(&existingLicense).Error; err == nil {
							if !existingLicense.IsExpired() {
								licenses = append(licenses, existingLicense)
								continue
							}
						}

						// Извлекаем данные о подписке
						var amount float64
						var transactionID string
						if attrs, ok := sub["attributes"].(map[string]interface{}); ok {
							if total, ok := attrs["total"].(float64); ok {
								amount = total
							}
						}
						if id, ok := sub["id"].(string); ok {
							transactionID = id
						}

						// Создаем лицензию
						now := time.Now()
						license := models.License{
							ShopID:              &targetShop.ID,
							UserID:              &user.ID,
							SubscriptionType:     plan.SubscriptionType,
							ActivationType:       models.ActivationTypePayment,
							SubscriptionStatus:   models.SubscriptionStatusActive,
							ActivatedAt:          &now,
							PaymentAmount:        amount,
							PaymentCurrency:      plan.Currency,
							PaymentProvider:      "lemonsqueezy",
							PaymentTransactionID: transactionID,
							LastPaymentDate:      &now,
							AutoRenew:            true,
							IsActive:             true,
						}

						license.ExpiresAt = license.CalculateExpirationDate(now)
						license.NextPaymentDate = license.ExpiresAt

						if err := database.DB.Create(&license).Error; err == nil {
							log.Printf("✅ [GetMyLicenses] Лицензия создана: %s", license.ID)
							// Загружаем Shop для лицензии
							database.DB.Preload("Shop").First(&license, license.ID)
							licenses = append(licenses, license)
						}
					}
				}
			}
		}
	}

	// Преобразуем в ответы
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
		ShopID               *string                 `json:"shopId"`
		SubscriptionType     models.SubscriptionType `json:"subscriptionType" binding:"required"`
		ActivationType       models.ActivationType   `json:"activationType"`
		PaymentAmount        float64                 `json:"paymentAmount"`
		PaymentCurrency      string                  `json:"paymentCurrency"`
		PaymentProvider      string                  `json:"paymentProvider"`
		PaymentTransactionID string                  `json:"paymentTransactionId"`
		AutoRenew            bool                    `json:"autoRenew"`
		Notes                string                  `json:"notes"`
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
		SubscriptionType     models.SubscriptionType `json:"subscriptionType" binding:"required"`
		PaymentAmount        float64                 `json:"paymentAmount"`
		PaymentCurrency      string                  `json:"paymentCurrency"`
		PaymentProvider      string                  `json:"paymentProvider"`
		PaymentTransactionID string                  `json:"paymentTransactionId"`
		AutoRenew            bool                    `json:"autoRenew"`
		Notes                string                  `json:"notes"`
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
		ShopID:               &shopID,
		UserID:               &shop.OwnerID,
		SubscriptionType:     req.SubscriptionType,
		ActivationType:       models.ActivationTypePayment,
		SubscriptionStatus:   models.SubscriptionStatusActive,
		ActivatedAt:          &now,
		PaymentAmount:        req.PaymentAmount,
		PaymentCurrency:      req.PaymentCurrency,
		PaymentProvider:      req.PaymentProvider,
		PaymentTransactionID: req.PaymentTransactionID,
		AutoRenew:            req.AutoRenew,
		Notes:                req.Notes,
		IsActive:             true,
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

// ExtendLicense продлевает лицензию (админ)
func (lc *LicenseController) ExtendLicense(c *gin.Context) {
	licenseIDParam := c.Param("id")
	licenseID, err := uuid.Parse(licenseIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid license ID",
		})
		return
	}

	var req struct {
		Months int `json:"months" binding:"required,min=1"` // Количество месяцев для продления
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
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

	// Определяем базовую дату для продления
	var baseDate time.Time
	if license.ExpiresAt != nil && license.ExpiresAt.After(time.Now()) {
		// Если лицензия еще не истекла, продлеваем от текущей даты окончания
		baseDate = *license.ExpiresAt
	} else {
		// Если лицензия истекла, продлеваем от текущей даты
		baseDate = time.Now()
	}

	// Вычисляем новую дату окончания
	newExpiresAt := baseDate.AddDate(0, req.Months, 0)
	license.ExpiresAt = &newExpiresAt
	license.NextPaymentDate = &newExpiresAt

	// Обновляем статус на активный, если был неактивен
	if license.SubscriptionStatus != models.SubscriptionStatusActive {
		license.SubscriptionStatus = models.SubscriptionStatusActive
	}
	license.IsActive = true

	// Обновляем дату последнего платежа
	now := time.Now()
	license.LastPaymentDate = &now

	if err := database.DB.Save(&license).Error; err != nil {
		log.Printf("❌ Ошибка продления лицензии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to extend license",
		})
		return
	}

	// Загружаем связанные данные
	database.DB.Preload("Shop").Preload("User").First(&license, license.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("License extended by %d months", req.Months),
		"data":    license.ToResponse(),
	})
}

// DeactivateLicense деактивирует лицензию для магазина (очищает device_id для возможности активации на новом устройстве)
func (lc *LicenseController) DeactivateLicense(c *gin.Context) {
	var req struct {
		LicenseKey string `json:"licenseKey" binding:"required"`
		ShopID     string `json:"shopId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Очищаем данные
	req.LicenseKey = strings.TrimSpace(req.LicenseKey)
	req.ShopID = strings.TrimSpace(req.ShopID)

	// Парсим ShopID
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid shop ID",
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

	// Проверяем, что лицензия принадлежит этому магазину
	if license.ShopID == nil || *license.ShopID != shopID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "License does not belong to this shop",
		})
		return
	}

	// Деактивируем устройство (очищаем device_id, но оставляем shop_id)
	license.DeviceID = ""
	license.DeviceInfo = ""
	license.DeviceFingerprint = ""

	if err := database.DB.Save(&license).Error; err != nil {
		log.Printf("❌ Ошибка деактивации лицензии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to deactivate license",
		})
		return
	}

	log.Printf("✅ Лицензия %s деактивирована для магазина %s (устройство очищено)", req.LicenseKey, shopID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "License deactivated successfully. You can now activate it on a new device.",
		"data":    license.ToResponse(),
	})
}

// generateDeviceFingerprint создает уникальный fingerprint устройства на основе DeviceID и DeviceInfo
func generateDeviceFingerprint(deviceID string, deviceInfo map[string]interface{}) string {
	// Создаем строку для хеширования
	var parts []string
	parts = append(parts, deviceID)

	// Сортируем ключи deviceInfo для консистентности
	if deviceInfo != nil {
		keys := make([]string, 0, len(deviceInfo))
		for k := range deviceInfo {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := deviceInfo[k]
			parts = append(parts, k+":"+toString(v))
		}
	}

	// Создаем хеш
	data := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// toString преобразует значение в строку
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int32, int64:
		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%v", val), " ", ""), "\n", ""))
	case float32, float64:
		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%.0f", val), " ", ""), "\n", ""))
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%v", val), " ", ""), "\n", ""))
	}
}

// DeleteLicense удаляет лицензию (админ)
func (lc *LicenseController) DeleteLicense(c *gin.Context) {
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

	// Сохраняем информацию о лицензии для ответа
	licenseInfo := license.ToResponse()

	// Удаляем лицензию
	if err := database.DB.Delete(&license).Error; err != nil {
		log.Printf("❌ Ошибка удаления лицензии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete license",
		})
		return
	}

	log.Printf("✅ Лицензия удалена: %s (shop_id: %s)", license.ID, license.ShopID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "License deleted successfully",
		"data":    licenseInfo,
	})
}
