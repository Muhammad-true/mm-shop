package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mm-api/mm-api/config"
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
		PaymentProvider     string                      `json:"paymentProvider"`     // lemonsqueezy, stripe, paypal, etc. (по умолчанию lemonsqueezy)
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

	// Используем lemonsqueezy по умолчанию, если не указан провайдер
	paymentProvider := req.PaymentProvider
	if paymentProvider == "" {
		paymentProvider = "lemonsqueezy"
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
		PaymentProvider:       paymentProvider,
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

// HandleLemonSqueezyWebhook обрабатывает webhook от Lemon Squeezy для подтверждения платежей
func (src *ShopRegistrationController) HandleLemonSqueezyWebhook(c *gin.Context) {
	// Читаем тело запроса
	bodyBytes, _ := c.GetRawData()
	var webhookData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &webhookData); err != nil {
		log.Printf("❌ [LemonSqueezyWebhook] Ошибка парсинга JSON: %v", err)
		log.Printf("📥 [LemonSqueezyWebhook] Raw body: %s", string(bodyBytes))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid JSON",
		})
		return
	}

	// Логируем полученный webhook для отладки (в JSON формате для читаемости)
	webhookJSON, _ := json.MarshalIndent(webhookData, "", "  ")
	log.Printf("📥 [LemonSqueezyWebhook] Получен webhook:\n%s", string(webhookJSON))

	// Извлекаем тип события
	eventName := ""
	if meta, ok := webhookData["meta"].(map[string]interface{}); ok {
		if name, ok := meta["event_name"].(string); ok {
			eventName = name
		}
	}

	log.Printf("📋 [LemonSqueezyWebhook] Тип события: %s", eventName)

	// Обрабатываем только события, связанные с оплатой
	switch eventName {
	case "order_created", "subscription_created", "subscription_payment_success":
		// Извлекаем данные о заказе/подписке
		data, ok := webhookData["data"].(map[string]interface{})
		if !ok {
			log.Printf("❌ [LemonSqueezyWebhook] Некорректная структура данных")
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid data structure",
			})
			return
		}

		// Извлекаем ID транзакции
		var transactionID string
		if id, ok := data["id"].(string); ok {
			transactionID = id
		}
		if attributes, ok := data["attributes"].(map[string]interface{}); ok {
			if transactionID == "" {
				if id, ok := attributes["order_id"].(string); ok {
					transactionID = id
				} else if id, ok := attributes["subscription_id"].(string); ok {
					transactionID = id
				} else if id, ok := attributes["id"].(string); ok {
					transactionID = id
				}
			}
		}

		// Извлекаем сумму оплаты
		var amount float64
		if attributes, ok := data["attributes"].(map[string]interface{}); ok {
			if total, ok := attributes["total"].(float64); ok {
				amount = total
			} else if total, ok := attributes["total"].(int); ok {
				amount = float64(total) / 100.0 // Если в центах
			} else if total, ok := attributes["total"].(int64); ok {
				amount = float64(total) / 100.0 // Если в центах
			}
		}

		// Извлекаем variant_id для определения плана подписки
		// Вариант 1: из attributes.variant_id
		// Вариант 2: из relationships.variant.data.id
		// Вариант 3: из order_items (для order_created)
		var variantID string
		if attributes, ok := data["attributes"].(map[string]interface{}); ok {
			// Прямой variant_id в attributes
			if variant, ok := attributes["variant_id"].(string); ok {
				variantID = variant
			} else if variant, ok := attributes["variant_id"].(float64); ok {
				variantID = fmt.Sprintf("%.0f", variant) // Конвертируем число в строку
			} else if variant, ok := attributes["variant_id"].(int); ok {
				variantID = fmt.Sprintf("%d", variant)
			} else if variant, ok := attributes["variant_id"].(int64); ok {
				variantID = fmt.Sprintf("%d", variant)
			}
			// variant_id из first_order_item
			if variantID == "" {
				if firstOrderItem, ok := attributes["first_order_item"].(map[string]interface{}); ok {
					if variant, ok := firstOrderItem["variant_id"].(string); ok {
						variantID = variant
					}
				}
			}
		}
		// Из relationships
		if variantID == "" {
			if relationships, ok := data["relationships"].(map[string]interface{}); ok {
				if variant, ok := relationships["variant"].(map[string]interface{}); ok {
					if variantData, ok := variant["data"].(map[string]interface{}); ok {
						if id, ok := variantData["id"].(string); ok {
							variantID = id
						}
					}
				}
			}
		}
		// Из included (для order_created с order_items)
		if variantID == "" {
			if included, ok := webhookData["included"].([]interface{}); ok {
				for _, item := range included {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if itemType, ok := itemMap["type"].(string); ok && itemType == "order-items" {
							if itemAttrs, ok := itemMap["attributes"].(map[string]interface{}); ok {
								if variant, ok := itemAttrs["variant_id"].(string); ok {
									variantID = variant
									break
								}
							}
						}
					}
				}
			}
		}

		log.Printf("💰 [LemonSqueezyWebhook] TransactionID: %s, Amount: %.2f, VariantID: %s", transactionID, amount, variantID)

		// Находим план подписки по variant_id
		if variantID != "" {
			log.Printf("🔍 [LemonSqueezyWebhook] Ищем план подписки по variant_id: %s", variantID)
			var plan models.SubscriptionPlan
			if err := database.DB.Where("lemonsqueezy_variant_id = ?", variantID).First(&plan).Error; err == nil {
				log.Printf("✅ [LemonSqueezyWebhook] Найден план подписки: %s (ID: %s)", plan.Name, plan.ID)

				// Извлекаем shop_id из custom данных
				// В Lemon Squeezy custom данные передаются через checkout_data.custom
				var shopID *uuid.UUID
				
				// Вариант 1: из attributes.custom (для order)
				if attributes, ok := data["attributes"].(map[string]interface{}); ok {
					if custom, ok := attributes["custom"].(map[string]interface{}); ok {
						if shopIDStr, ok := custom["shop_id"].(string); ok {
							if parsedID, err := uuid.Parse(shopIDStr); err == nil {
								shopID = &parsedID
								log.Printf("✅ [LemonSqueezyWebhook] Найден shop_id из attributes.custom: %s", shopIDStr)
							}
						}
					}
					// Вариант 2: из attributes.checkout_data.custom
					if shopID == nil {
						if checkoutData, ok := attributes["checkout_data"].(map[string]interface{}); ok {
							if custom, ok := checkoutData["custom"].(map[string]interface{}); ok {
								if shopIDStr, ok := custom["shop_id"].(string); ok {
									if parsedID, err := uuid.Parse(shopIDStr); err == nil {
										shopID = &parsedID
										log.Printf("✅ [LemonSqueezyWebhook] Найден shop_id из checkout_data.custom: %s", shopIDStr)
									}
								}
							}
						}
					}
					// Вариант 3: из relationships.checkout (для order_created)
					if shopID == nil {
						if relationships, ok := data["relationships"].(map[string]interface{}); ok {
							if checkout, ok := relationships["checkout"].(map[string]interface{}); ok {
								if checkoutData, ok := checkout["data"].(map[string]interface{}); ok {
									if checkoutID, ok := checkoutData["id"].(string); ok {
										// Нужно найти checkout в included
										if included, ok := webhookData["included"].([]interface{}); ok {
											for _, item := range included {
												if itemMap, ok := item.(map[string]interface{}); ok {
													if itemID, ok := itemMap["id"].(string); ok && itemID == checkoutID {
														if itemAttrs, ok := itemMap["attributes"].(map[string]interface{}); ok {
															if custom, ok := itemAttrs["custom"].(map[string]interface{}); ok {
																if shopIDStr, ok := custom["shop_id"].(string); ok {
																	if parsedID, err := uuid.Parse(shopIDStr); err == nil {
																		shopID = &parsedID
																		log.Printf("✅ [LemonSqueezyWebhook] Найден shop_id из checkout.custom: %s", shopIDStr)
																		break
																	}
																}
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
					// Вариант 4: ищем по email покупателя и синхронизируем подписки из Lemon Squeezy
					if shopID == nil {
						if customerEmail, ok := attributes["user_email"].(string); ok {
							log.Printf("🔍 [LemonSqueezyWebhook] shop_id не найден, проверяем подписки по email: %s", customerEmail)
							
							// Сначала пробуем найти магазин по email
							var shop models.Shop
							if err := database.DB.Where("email = ?", customerEmail).First(&shop).Error; err == nil {
								shopID = &shop.ID
								log.Printf("✅ [LemonSqueezyWebhook] Найден shop_id по email: %s (shop: %s)", customerEmail, shopID.String())
							} else {
								// Если магазин не найден, ищем пользователя и синхронизируем его подписки
								var user models.User
								if err := database.DB.Where("email = ?", customerEmail).First(&user).Error; err == nil {
									log.Printf("✅ [LemonSqueezyWebhook] Найден пользователь по email: %s (ID: %s)", customerEmail, user.ID)
									// Синхронизируем подписки из Lemon Squeezy
									if syncedShopID := src.syncUserSubscriptionsFromLemonSqueezy(&user, variantID, transactionID, amount, plan); syncedShopID != nil {
										shopID = syncedShopID
										log.Printf("✅ [LemonSqueezyWebhook] Подписки синхронизированы, shop_id: %s", shopID.String())
									}
								} else {
									log.Printf("⚠️ [LemonSqueezyWebhook] Пользователь не найден по email: %s", customerEmail)
								}
							}
						}
					}
				}

				if shopID != nil {
					log.Printf("✅ [LemonSqueezyWebhook] Найден shop_id: %s", shopID.String())
					
					// Проверяем существование магазина
					var shop models.Shop
					if err := database.DB.First(&shop, shopID).Error; err != nil {
						log.Printf("❌ [LemonSqueezyWebhook] Магазин не найден в БД: %s, ошибка: %v", shopID.String(), err)
						c.JSON(http.StatusNotFound, gin.H{
							"success": false,
							"error":   "Shop not found",
						})
						return
					}
					log.Printf("✅ [LemonSqueezyWebhook] Магазин найден: %s (Owner: %s)", shop.Name, shop.OwnerID.String())
					
					// Проверяем, нет ли уже активной лицензии
					var existingLicense models.License
					if err := database.DB.Where("shop_id = ? AND subscription_status = ?", shopID, models.SubscriptionStatusActive).First(&existingLicense).Error; err == nil {
						if !existingLicense.IsExpired() {
							log.Printf("ℹ️ [LemonSqueezyWebhook] У магазина уже есть активная лицензия: %s", existingLicense.ID)
							c.JSON(http.StatusOK, gin.H{
								"success": true,
								"message": "License already exists",
							})
							return
						} else {
							log.Printf("ℹ️ [LemonSqueezyWebhook] Существующая лицензия истекла, создаем новую")
						}
					} else {
						log.Printf("ℹ️ [LemonSqueezyWebhook] Активная лицензия не найдена, создаем новую")
					}

					// Создаем лицензию
					now := time.Now()
					license := models.License{
						ShopID:                shopID,
						SubscriptionType:       plan.SubscriptionType,
						ActivationType:         models.ActivationTypePayment,
						SubscriptionStatus:     models.SubscriptionStatusActive,
						ActivatedAt:            &now,
						PaymentAmount:          amount,
						PaymentCurrency:        plan.Currency,
						PaymentProvider:        "lemonsqueezy",
						PaymentTransactionID:   transactionID,
						LastPaymentDate:         &now,
						AutoRenew:              true, // Lemon Squeezy обычно поддерживает автопродление
						IsActive:               true,
					}

					// Вычисляем дату окончания
					license.ExpiresAt = license.CalculateExpirationDate(now)
					license.NextPaymentDate = license.ExpiresAt

					// Получаем UserID из магазина (shop уже получен выше)
					license.UserID = &shop.OwnerID

					log.Printf("🔄 [LemonSqueezyWebhook] Создаем лицензию для shop_id: %s, plan: %s, amount: %.2f", shopID.String(), plan.Name, amount)
					if err := database.DB.Create(&license).Error; err != nil {
						log.Printf("❌ [LemonSqueezyWebhook] Ошибка создания лицензии: %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{
							"success": false,
							"error":   "Failed to create license",
							"details": err.Error(),
						})
						return
					}

					log.Printf("✅ [LemonSqueezyWebhook] Лицензия создана успешно: %s (ExpiresAt: %v)", license.ID, license.ExpiresAt)
					c.JSON(http.StatusOK, gin.H{
						"success": true,
						"message": "License created successfully",
						"data":    license.ToResponse(),
					})
					return
				} else {
					log.Printf("⚠️ [LemonSqueezyWebhook] Не удалось определить shop_id из webhook данных")
				}
			} else {
				log.Printf("❌ [LemonSqueezyWebhook] План подписки не найден для variant_id: %s (ошибка: %v)", variantID, err)
				// Логируем все доступные планы для отладки
				var allPlans []models.SubscriptionPlan
				database.DB.Find(&allPlans)
				log.Printf("📋 [LemonSqueezyWebhook] Доступные планы в БД:")
				for _, p := range allPlans {
					log.Printf("   - %s: variant_id=%s", p.Name, p.LemonSqueezyVariantID)
				}
			}
		} else {
			log.Printf("❌ [LemonSqueezyWebhook] variant_id не найден в данных webhook")
			log.Printf("📋 [LemonSqueezyWebhook] Структура data для отладки:")
			if dataJSON, err := json.MarshalIndent(data, "", "  "); err == nil {
				log.Printf("%s", string(dataJSON))
			}
		}

	case "subscription_cancelled", "subscription_payment_failed":
		log.Printf("⚠️ [LemonSqueezyWebhook] Получено событие отмены/ошибки: %s", eventName)
		// Можно добавить логику для обработки отмены подписки
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Event processed",
		})
		return

	default:
		log.Printf("ℹ️ [LemonSqueezyWebhook] Необработанное событие: %s", eventName)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Event received",
		})
		return
	}

	// Если дошли сюда, значит не удалось обработать
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook received but not processed",
	})
}

// syncUserSubscriptionsFromLemonSqueezy синхронизирует подписки пользователя из Lemon Squeezy
// Возвращает shop_id для которого была создана лицензия, или nil если не удалось
func (src *ShopRegistrationController) syncUserSubscriptionsFromLemonSqueezy(
	user *models.User,
	variantID string,
	transactionID string,
	amount float64,
	plan models.SubscriptionPlan,
) *uuid.UUID {
	log.Printf("🔄 [LemonSqueezySync] Начинаем синхронизацию подписок для пользователя: %s", user.Email)

	cfg := config.GetConfig()
	if cfg.LemonSqueezyAPIKey == "" {
		log.Printf("❌ [LemonSqueezySync] Lemon Squeezy API key не настроен")
		return nil
	}

	// Получаем все магазины пользователя
	var shops []models.Shop
	if err := database.DB.Where("owner_id = ?", user.ID).Find(&shops).Error; err != nil {
		log.Printf("❌ [LemonSqueezySync] Ошибка получения магазинов: %v", err)
		return nil
	}

	if len(shops) == 0 {
		log.Printf("⚠️ [LemonSqueezySync] У пользователя нет магазинов")
		return nil
	}

	log.Printf("📦 [LemonSqueezySync] Найдено магазинов: %d", len(shops))

	// Получаем подписки пользователя из Lemon Squeezy API
	subscriptions, err := src.getLemonSqueezySubscriptionsByEmail(user.Email, cfg.LemonSqueezyAPIKey)
	if err != nil {
		log.Printf("❌ [LemonSqueezySync] Ошибка получения подписок из Lemon Squeezy: %v", err)
		return nil
	}

	log.Printf("📋 [LemonSqueezySync] Найдено подписок в Lemon Squeezy: %d", len(subscriptions))

	// Ищем активную подписку с нужным variant_id
	for _, sub := range subscriptions {
		subVariantID := src.extractVariantIDFromSubscription(sub)
		if subVariantID == variantID || (variantID == "" && subVariantID != "") {
			log.Printf("✅ [LemonSqueezySync] Найдена подписка с variant_id: %s", subVariantID)

			// Находим план подписки
			var subscriptionPlan models.SubscriptionPlan
			if err := database.DB.Where("lemonsqueezy_variant_id = ?", subVariantID).First(&subscriptionPlan).Error; err != nil {
				log.Printf("⚠️ [LemonSqueezySync] План подписки не найден для variant_id: %s", subVariantID)
				continue
			}

			// Определяем для какого магазина создавать лицензию
			// Используем первый магазин пользователя
			targetShop := &shops[0]

			// Проверяем, нет ли уже активной лицензии
			var existingLicense models.License
			if err := database.DB.Where("shop_id = ? AND subscription_status = ?", targetShop.ID, models.SubscriptionStatusActive).First(&existingLicense).Error; err == nil {
				if !existingLicense.IsExpired() {
					log.Printf("ℹ️ [LemonSqueezySync] У магазина уже есть активная лицензия: %s", existingLicense.ID)
					return &targetShop.ID
				}
			}

			// Создаем лицензию
			now := time.Now()
			license := models.License{
				ShopID:              &targetShop.ID,
				UserID:              &user.ID,
				SubscriptionType:     subscriptionPlan.SubscriptionType,
				ActivationType:       models.ActivationTypePayment,
				SubscriptionStatus:   models.SubscriptionStatusActive,
				ActivatedAt:          &now,
				PaymentAmount:        amount,
				PaymentCurrency:      subscriptionPlan.Currency,
				PaymentProvider:      "lemonsqueezy",
				PaymentTransactionID: transactionID,
				LastPaymentDate:      &now,
				AutoRenew:            true,
				IsActive:             true,
			}

			// Вычисляем дату окончания
			license.ExpiresAt = license.CalculateExpirationDate(now)
			license.NextPaymentDate = license.ExpiresAt

			if err := database.DB.Create(&license).Error; err != nil {
				log.Printf("❌ [LemonSqueezySync] Ошибка создания лицензии: %v", err)
				continue
			}

			log.Printf("✅ [LemonSqueezySync] Лицензия создана успешно: %s для shop_id: %s", license.ID, targetShop.ID)
			return &targetShop.ID
		}
	}

	log.Printf("⚠️ [LemonSqueezySync] Активная подписка с нужным variant_id не найдена")
	return nil
}

// getLemonSqueezySubscriptionsByEmail получает подписки пользователя из Lemon Squeezy API по email
func (src *ShopRegistrationController) getLemonSqueezySubscriptionsByEmail(email, apiKey string) ([]map[string]interface{}, error) {
	// Сначала находим customer по email
	customerID, err := src.findLemonSqueezyCustomerByEmail(email, apiKey)
	if err != nil || customerID == "" {
		log.Printf("⚠️ [LemonSqueezyAPI] Customer не найден по email: %s", email)
		return nil, err
	}

	// Получаем подписки customer
	apiURL := fmt.Sprintf("https://api.lemonsqueezy.com/v1/subscriptions?filter[customer_id]=%s", customerID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/vnd.api+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [LemonSqueezyAPI] Ошибка API: %d, %+v", resp.StatusCode, response)
		return nil, fmt.Errorf("Lemon Squeezy API error: %d", resp.StatusCode)
	}

	// Извлекаем подписки из ответа
	var subscriptions []map[string]interface{}
	if data, ok := response["data"].([]interface{}); ok {
		for _, item := range data {
			if subMap, ok := item.(map[string]interface{}); ok {
				// Проверяем статус подписки (только активные)
				if attrs, ok := subMap["attributes"].(map[string]interface{}); ok {
					if status, ok := attrs["status"].(string); ok {
						if status == "active" || status == "on_trial" {
							subscriptions = append(subscriptions, subMap)
						}
					}
				}
			}
		}
	}

	return subscriptions, nil
}

// findLemonSqueezyCustomerByEmail находит customer ID в Lemon Squeezy по email
func (src *ShopRegistrationController) findLemonSqueezyCustomerByEmail(email, apiKey string) (string, error) {
	apiURL := fmt.Sprintf("https://api.lemonsqueezy.com/v1/customers?filter[email]=%s", email)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/vnd.api+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Lemon Squeezy API error: %d", resp.StatusCode)
	}

	// Извлекаем customer ID из ответа
	if data, ok := response["data"].([]interface{}); ok && len(data) > 0 {
		if customer, ok := data[0].(map[string]interface{}); ok {
			if id, ok := customer["id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("customer not found")
}

// extractVariantIDFromSubscription извлекает variant_id из подписки Lemon Squeezy
func (src *ShopRegistrationController) extractVariantIDFromSubscription(subscription map[string]interface{}) string {
	// Вариант 1: из attributes.variant_id
	if attrs, ok := subscription["attributes"].(map[string]interface{}); ok {
		if variantID, ok := attrs["variant_id"].(string); ok {
			return variantID
		}
		if variantID, ok := attrs["variant_id"].(float64); ok {
			return fmt.Sprintf("%.0f", variantID)
		}
	}

	// Вариант 2: из relationships.variant.data.id
	if relationships, ok := subscription["relationships"].(map[string]interface{}); ok {
		if variant, ok := relationships["variant"].(map[string]interface{}); ok {
			if variantData, ok := variant["data"].(map[string]interface{}); ok {
				if id, ok := variantData["id"].(string); ok {
					return id
				}
			}
		}
	}

	return ""
}

// SyncUserSubscriptions синхронизирует подписки текущего пользователя из Lemon Squeezy
func (src *ShopRegistrationController) SyncUserSubscriptions(c *gin.Context) {
	// Получаем пользователя из контекста (должен быть установлен middleware.AuthRequired)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// Получаем пользователя из БД
	var user models.User
	if err := database.DB.First(&user, userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	log.Printf("🔄 [SyncSubscriptions] Начинаем синхронизацию подписок для пользователя: %s", user.Email)

	cfg := config.GetConfig()
	if cfg.LemonSqueezyAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Lemon Squeezy API key not configured",
		})
		return
	}

	// Получаем подписки пользователя из Lemon Squeezy
	subscriptions, err := src.getLemonSqueezySubscriptionsByEmail(user.Email, cfg.LemonSqueezyAPIKey)
	if err != nil {
		log.Printf("❌ [SyncSubscriptions] Ошибка получения подписок: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get subscriptions from Lemon Squeezy",
			"details": err.Error(),
		})
		return
	}

	if len(subscriptions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No active subscriptions found",
			"data":    []interface{}{},
		})
		return
	}

	// Получаем все магазины пользователя
	var shops []models.Shop
	if err := database.DB.Where("owner_id = ?", user.ID).Find(&shops).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get user shops",
		})
		return
	}

	if len(shops) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User has no shops",
		})
		return
	}

	// Синхронизируем каждую подписку
	var syncedLicenses []models.License
	for _, sub := range subscriptions {
		variantID := src.extractVariantIDFromSubscription(sub)
		if variantID == "" {
			log.Printf("⚠️ [SyncSubscriptions] Не удалось извлечь variant_id из подписки")
			continue
		}

		// Находим план подписки
		var plan models.SubscriptionPlan
		if err := database.DB.Where("lemonsqueezy_variant_id = ?", variantID).First(&plan).Error; err != nil {
			log.Printf("⚠️ [SyncSubscriptions] План подписки не найден для variant_id: %s", variantID)
			continue
		}

		// Используем первый магазин пользователя
		targetShop := &shops[0]

		// Проверяем, нет ли уже активной лицензии
		var existingLicense models.License
		if err := database.DB.Where("shop_id = ? AND subscription_status = ?", targetShop.ID, models.SubscriptionStatusActive).First(&existingLicense).Error; err == nil {
			if !existingLicense.IsExpired() {
				log.Printf("ℹ️ [SyncSubscriptions] У магазина уже есть активная лицензия: %s", existingLicense.ID)
				syncedLicenses = append(syncedLicenses, existingLicense)
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

		if err := database.DB.Create(&license).Error; err != nil {
			log.Printf("❌ [SyncSubscriptions] Ошибка создания лицензии: %v", err)
			continue
		}

		log.Printf("✅ [SyncSubscriptions] Лицензия создана: %s для shop_id: %s", license.ID, targetShop.ID)
		syncedLicenses = append(syncedLicenses, license)
	}

	// Преобразуем лицензии в ответ
	var licensesResponse []interface{}
	for _, license := range syncedLicenses {
		licensesResponse = append(licensesResponse, license.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Synced %d subscription(s)", len(syncedLicenses)),
		"data":    licensesResponse,
	})
}

