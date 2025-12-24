package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/mm-api/mm-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShopController обрабатывает запросы магазинов
type ShopController struct{}

// GetShopInfo возвращает информацию о магазине по ID
func (sc *ShopController) GetShopInfo(c *gin.Context) {
	shopIDParam := c.Param("id")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid shop ID",
		})
		return
	}

	var shop models.Shop
	var shopUser models.User // Для обратной совместимости
	var useLegacyMode bool

	// Пробуем найти в новой таблице shops
	if err := database.DB.Preload("Owner.Role").Where("id = ?", shopID).First(&shop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Обратная совместимость: пробуем найти в старой таблице users
			if err := database.DB.Preload("Role").Where("id = ? AND role_id IN (SELECT id FROM roles WHERE name = 'shop_owner')", shopID).First(&shopUser).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Shop not found",
				})
				return
			}
			if shopUser.Role == nil || shopUser.Role.Name != "shop_owner" {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Shop not found",
				})
				return
			}
			useLegacyMode = true
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
			return
		}
	}

	var productsCount int64
	var subscribersCount int64
	var shopInfo models.ShopInfo
	var email, phone, avatar string
	var createdAt time.Time

	if useLegacyMode {
		// Обратная совместимость: используем данные из User
		shopInfo = models.ShopInfo{
			ID:   shopUser.ID,
			Name: shopUser.Name,
			INN:  shopUser.INN,
		}
		email = shopUser.Email
		phone = shopUser.Phone
		avatar = shopUser.Avatar
		createdAt = shopUser.CreatedAt
		// Подсчитываем количество товаров (старый способ)
		database.DB.Model(&models.Product{}).Where("owner_id = ?", shopUser.ID).Count(&productsCount)
		// Подсчитываем количество подписчиков (старый способ - shop_id = user_id)
		database.DB.Model(&models.ShopSubscription{}).Where("shop_id = ?", shopUser.ID).Count(&subscribersCount)
	} else {
		// Новый способ: используем Shop
		shopInfo = shop.ToShopInfo()
		email = shop.Email
		phone = shop.Phone
		avatar = shop.Logo
		createdAt = shop.CreatedAt
		// Подсчитываем количество товаров (новый способ)
		database.DB.Model(&models.Product{}).Where("shop_id = ? OR owner_id = ?", shop.ID, shop.OwnerID).Count(&productsCount)
		// Подсчитываем количество подписчиков
		database.DB.Model(&models.ShopSubscription{}).Where("shop_id = ?", shop.ID).Count(&subscribersCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"shop": gin.H{
				"id":              shopInfo.ID,
				"name":            shopInfo.Name,
				"inn":             shopInfo.INN,
				"email":           email,
				"phone":           phone,
				"avatar":          avatar,
				"productsCount":   productsCount,
				"subscribersCount": subscribersCount,
				"createdAt":       createdAt,
			},
		},
	})
}

// GetShopProducts возвращает товары магазина по ID с фильтрацией
func (sc *ShopController) GetShopProducts(c *gin.Context) {
	shopIDParam := c.Param("id")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid shop ID",
		})
		return
	}

	var shop models.Shop
	var shopUser models.User // Для обратной совместимости
	var useLegacyMode bool
	var shopInfo models.ShopInfo

	// Пробуем найти в новой таблице shops
	if err := database.DB.Preload("Owner.Role").Where("id = ?", shopID).First(&shop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Обратная совместимость: пробуем найти в старой таблице users
			if err := database.DB.Preload("Role").Where("id = ? AND role_id IN (SELECT id FROM roles WHERE name = 'shop_owner')", shopID).First(&shopUser).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Shop not found",
				})
				return
			}
			if shopUser.Role == nil || shopUser.Role.Name != "shop_owner" {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Shop not found",
				})
				return
			}
			useLegacyMode = true
			shopInfo = models.ShopInfo{
				ID:   shopUser.ID,
				Name: shopUser.Name,
				INN:  shopUser.INN,
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
			return
		}
	} else {
		shopInfo = shop.ToShopInfo()
	}

	var products []models.Product
	var query *gorm.DB

	if useLegacyMode {
		// Обратная совместимость: используем owner_id
		query = database.DB.Model(&models.Product{}).Where("owner_id = ?", shopID)
	} else {
		// Новый способ: используем shop_id (или owner_id для обратной совместимости)
		query = database.DB.Model(&models.Product{}).Where("shop_id = ? OR owner_id = ?", shopID, shop.OwnerID)
	}

	// Фильтрация по категории
	if categoryID := c.Query("category"); categoryID != "" {
		if categoryUUID, err := uuid.Parse(categoryID); err == nil {
			query = query.Where("category_id = ?", categoryUUID)
		}
	}

	// Поиск по названию
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Фильтрация по наличию на складе
	if inStock := c.Query("in_stock"); inStock == "true" {
		query = query.Joins("JOIN product_variations ON products.id = product_variations.product_id").
			Where("product_variations.stock_quantity > 0").
			Distinct("products.id")
	}

	// Фильтрация по полу
	if gender := c.Query("gender"); gender != "" {
		query = query.Where("gender = ?", gender)
	}

	// Фильтрация по бренду
	if brand := c.Query("brand"); brand != "" {
		query = query.Where("brand ILIKE ?", "%"+brand+"%")
	}

	// Фильтрация по городу
	if cityID := c.Query("city_id"); cityID != "" {
		if cityUUID, err := uuid.Parse(cityID); err == nil {
			// Фильтруем по city_id в продукте или через shop.city_id
			query = query.Where("city_id = ? OR shop_id IN (SELECT id FROM shops WHERE city_id = ?)", cityUUID, cityUUID)
		}
	}

	// Сортировка
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	allowedSorts := map[string]bool{
		"name":       true,
		"created_at": true,
		"price":      false, // Для цены нужен JOIN с вариациями
	}

	if allowedSorts[sortBy] {
		if sortOrder == "asc" {
			query = query.Order("products." + sortBy + " ASC")
		} else {
			query = query.Order("products." + sortBy + " DESC")
		}
	}

	// Пагинация
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Получаем общее количество
	var total int64
	query.Count(&total)

	// Получаем продукты с загрузкой связей
	preloadQuery := query.Offset(offset).Limit(limit).
		Preload("Variations").
		Preload("Category")
	
	if useLegacyMode {
		preloadQuery = preloadQuery.Preload("Owner.Role")
	} else {
		preloadQuery = preloadQuery.Preload("Shop").Preload("Owner.Role")
	}

	if err := preloadQuery.Find(&products).Error; err != nil {
		log.Printf("❌ Ошибка получения товаров магазина: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch shop products",
		})
		return
	}

	// Преобразуем в response
	productResponses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = product.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"products": productResponses,
			"shop":     shopInfo,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	})
}

// SubscribeToShop подписывает пользователя на магазин
func (sc *ShopController) SubscribeToShop(c *gin.Context) {
	shopIDParam := c.Param("id")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid shop ID",
		})
		return
	}

	// Получаем текущего пользователя
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	user := currentUser.(models.User)

	// Проверяем, что магазин существует
	var shop models.Shop
	var shopUser models.User // Для обратной совместимости
	var useLegacyMode bool

	// Пробуем найти в новой таблице shops
	if err := database.DB.Preload("Owner.Role").Where("id = ?", shopID).First(&shop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Обратная совместимость: пробуем найти в старой таблице users
			if err := database.DB.Preload("Role").Where("id = ? AND role_id IN (SELECT id FROM roles WHERE name = 'shop_owner')", shopID).First(&shopUser).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Shop not found",
				})
				return
			}
			if shopUser.Role == nil || shopUser.Role.Name != "shop_owner" {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Shop not found",
				})
				return
			}
			useLegacyMode = true
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
			return
		}
	}

	// Проверяем, что пользователь не подписывается сам на свой магазин
	if useLegacyMode {
		if user.ID == shopID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot subscribe to your own shop",
			})
			return
		}
	} else {
		if user.ID == shop.OwnerID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot subscribe to your own shop",
			})
			return
		}
	}

	// Проверяем, не подписан ли уже
	log.Printf("🔍 [SubscribeToShop] Проверяем существующую подписку: userID=%s, shopID=%s", user.ID, shopID)
	var existingSubscription models.ShopSubscription
	if err := database.DB.Where("user_id = ? AND shop_id = ?", user.ID, shopID).First(&existingSubscription).Error; err == nil {
		log.Printf("✅ [SubscribeToShop] Подписка уже существует: subscriptionID=%s", existingSubscription.ID)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Already subscribed",
			"data":    existingSubscription.ToResponse(),
		})
		return
	} else {
		if err == gorm.ErrRecordNotFound {
			log.Printf("ℹ️ [SubscribeToShop] Подписка не найдена, создаем новую")
		} else {
			log.Printf("⚠️ [SubscribeToShop] Ошибка при проверке существующей подписки: %v", err)
		}
	}

	// Создаем подписку
	subscription := models.ShopSubscription{
		UserID: user.ID,
		ShopID: shopID,
	}

	log.Printf("📝 [SubscribeToShop] Создаем подписку: userID=%s, shopID=%s", user.ID, shopID)
	if err := database.DB.Create(&subscription).Error; err != nil {
		log.Printf("❌ [SubscribeToShop] Ошибка создания подписки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create subscription",
		})
		return
	}

	log.Printf("✅ [SubscribeToShop] Подписка успешно создана: subscriptionID=%s", subscription.ID)

	// Проверяем, что подписка действительно сохранилась
	var verifySubscription models.ShopSubscription
	if err := database.DB.Where("user_id = ? AND shop_id = ?", user.ID, shopID).First(&verifySubscription).Error; err != nil {
		log.Printf("⚠️ [SubscribeToShop] Предупреждение: не удалось проверить сохранение подписки: %v", err)
	} else {
		log.Printf("✅ [SubscribeToShop] Подписка подтверждена в БД: ID=%s", verifySubscription.ID)
	}

	// Загружаем информацию о магазине
	if useLegacyMode {
		// Обратная совместимость: загружаем User как Shop
		var shopUserForSub models.User
		database.DB.Preload("Role").First(&shopUserForSub, "id = ?", shopID)
		// Создаем временный Shop из User для совместимости
		subscription.Shop = models.Shop{
			ID:   shopUserForSub.ID,
			Name: shopUserForSub.Name,
			INN:  shopUserForSub.INN,
		}
	} else {
		database.DB.First(&subscription.Shop, "id = ?", shopID)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscribed successfully",
		"data":    subscription.ToResponse(),
	})
}

// UnsubscribeFromShop отписывает пользователя от магазина
func (sc *ShopController) UnsubscribeFromShop(c *gin.Context) {
	shopIDParam := c.Param("id")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid shop ID",
		})
		return
	}

	// Получаем текущего пользователя
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	user := currentUser.(models.User)

	// Удаляем подписку
	result := database.DB.Where("user_id = ? AND shop_id = ?", user.ID, shopID).Delete(&models.ShopSubscription{})
	if result.Error != nil {
		log.Printf("❌ Ошибка удаления подписки: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unsubscribe",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Subscription not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Unsubscribed successfully",
	})
}

// GetShopSubscribers возвращает список подписчиков магазина (только для владельца магазина)
func (sc *ShopController) GetShopSubscribers(c *gin.Context) {
	shopIDParam := c.Param("id")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid shop ID",
		})
		return
	}

	// Получаем текущего пользователя
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	user := currentUser.(models.User)

	// Проверяем, что магазин существует
	var shop models.Shop
	var shopUser models.User // Для обратной совместимости
	var useLegacyMode bool

	// Пробуем найти в новой таблице shops
	if err := database.DB.Preload("Owner.Role").Where("id = ?", shopID).First(&shop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Обратная совместимость: пробуем найти в старой таблице users
			if err := database.DB.Preload("Role").Where("id = ? AND role_id IN (SELECT id FROM roles WHERE name = 'shop_owner')", shopID).First(&shopUser).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Shop not found",
				})
				return
			}
			if shopUser.Role == nil || shopUser.Role.Name != "shop_owner" {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Shop not found",
				})
				return
			}
			useLegacyMode = true
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
			return
		}
	}

	// Проверяем, что текущий пользователь - владелец магазина или админ
	isOwner := false
	if useLegacyMode {
		isOwner = user.ID == shopID
	} else {
		isOwner = user.ID == shop.OwnerID
	}

	if !isOwner && (user.Role == nil || (user.Role.Name != "admin" && user.Role.Name != "super_admin")) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// Пагинация
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var subscriptions []models.ShopSubscription
	var total int64

	query := database.DB.Where("shop_id = ?", shopID).Preload("User")

	// Подсчитываем общее количество
	query.Model(&models.ShopSubscription{}).Count(&total)

	// Получаем подписки
	if err := query.Offset(offset).Limit(limit).Find(&subscriptions).Error; err != nil {
		log.Printf("❌ Ошибка получения подписчиков: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch subscribers",
		})
		return
	}

	// Преобразуем в response
	subscriberResponses := make([]models.ShopSubscriptionResponse, len(subscriptions))
	for i, sub := range subscriptions {
		subscriberResponses[i] = sub.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subscribers": subscriberResponses,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	})
}

// CheckSubscription проверяет, подписан ли пользователь на магазин
func (sc *ShopController) CheckSubscription(c *gin.Context) {
	shopIDParam := c.Param("id")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		log.Printf("❌ [CheckSubscription] Неверный shop ID: %s", shopIDParam)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid shop ID",
		})
		return
	}

	log.Printf("🔍 [CheckSubscription] Начало проверки подписки для shopID=%s", shopID)

	var userID uuid.UUID
	var userFound bool

	// Пробуем получить пользователя из контекста (если был middleware)
	currentUser, exists := c.Get("user")
	if exists {
		user := currentUser.(models.User)
		userID = user.ID
		userFound = true
		log.Printf("✅ [CheckSubscription] Пользователь найден через middleware: userID=%s, email=%s", userID, user.Email)
	} else {
		log.Printf("⚠️ [CheckSubscription] Пользователь не найден в контексте, проверяем токен вручную")
		
		// Если пользователя нет в контексте, пробуем опционально проверить токен
		authHeader := c.GetHeader("Authorization")
		log.Printf("🔍 [CheckSubscription] Authorization заголовок: %s", func() string {
			if authHeader == "" {
				return "ОТСУТСТВУЕТ"
			}
			if len(authHeader) > 50 {
				return authHeader[:50] + "..."
			}
			return authHeader
		}())
		
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != authHeader {
				log.Printf("🔍 [CheckSubscription] Токен извлечен из заголовка (длина: %d)", len(tokenString))
				
				// Валидируем токен
				claims, err := utils.ValidateJWT(tokenString)
				if err != nil {
					log.Printf("❌ [CheckSubscription] Ошибка валидации токена: %v", err)
				} else {
					log.Printf("✅ [CheckSubscription] Токен валиден, UserID из claims: %s", claims.UserID)
					
					parsedUserID, err := uuid.Parse(claims.UserID)
					if err != nil {
						log.Printf("❌ [CheckSubscription] Ошибка парсинга UserID из токена: %v, UserID из claims: %s", err, claims.UserID)
					} else {
						log.Printf("🔍 [CheckSubscription] Парсинг UserID успешен: %s", parsedUserID)
						
						// Проверяем, что пользователь существует и активен
						var user models.User
						if err := database.DB.Preload("Role").First(&user, "id = ? AND is_active = ?", parsedUserID, true).Error; err != nil {
							log.Printf("❌ [CheckSubscription] Пользователь не найден в БД или неактивен: userID=%s, ошибка: %v", parsedUserID, err)
						} else {
							userID = user.ID
							userFound = true
							log.Printf("✅ [CheckSubscription] Пользователь найден через токен: userID=%s, email=%s", userID, user.Email)
						}
					}
				}
			} else {
				log.Printf("⚠️ [CheckSubscription] Неверный формат заголовка Authorization (нет 'Bearer ')")
			}
		} else {
			log.Printf("⚠️ [CheckSubscription] Заголовок Authorization отсутствует")
		}
	}

	// Если пользователь не найден, возвращаем false
	if !userFound {
		log.Printf("❌ [CheckSubscription] Пользователь не найден, возвращаем isSubscribed=false")
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"isSubscribed": false,
		})
		return
	}

	// Проверяем подписку
	log.Printf("🔍 [CheckSubscription] Проверяем подписку в БД: userID=%s, shopID=%s", userID, shopID)
	var subscription models.ShopSubscription
	err = database.DB.Where("user_id = ? AND shop_id = ?", userID, shopID).First(&subscription).Error
	isSubscribed := err == nil
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("ℹ️ [CheckSubscription] Подписка не найдена в БД: userID=%s, shopID=%s", userID, shopID)
		} else {
			log.Printf("❌ [CheckSubscription] Ошибка при проверке подписки в БД: %v", err)
		}
	} else {
		log.Printf("✅ [CheckSubscription] Подписка найдена: subscriptionID=%s, userID=%s, shopID=%s", subscription.ID, userID, shopID)
	}
	
	log.Printf("📊 [CheckSubscription] Результат проверки: userID=%s, shopID=%s, isSubscribed=%v", userID, shopID, isSubscribed)

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"isSubscribed": isSubscribed,
	})
}

// GetShops возвращает список всех магазинов с информацией о подписке пользователя
func (sc *ShopController) GetShops(c *gin.Context) {
	log.Printf("🛍️ [GetShops] Начало получения списка магазинов")

	var userID *uuid.UUID
	var userFound bool

	// Пробуем получить пользователя из контекста (если был middleware)
	currentUser, exists := c.Get("user")
	if exists {
		user := currentUser.(models.User)
		userID = &user.ID
		userFound = true
		log.Printf("✅ [GetShops] Пользователь найден через middleware: userID=%s", user.ID)
	} else {
		// Если пользователя нет в контексте, пробуем опционально проверить токен
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != authHeader {
				claims, err := utils.ValidateJWT(tokenString)
				if err == nil {
					parsedUserID, err := uuid.Parse(claims.UserID)
					if err == nil {
						var user models.User
						if err := database.DB.Preload("Role").First(&user, "id = ? AND is_active = ?", parsedUserID, true).Error; err == nil {
							userID = &user.ID
							userFound = true
							log.Printf("✅ [GetShops] Пользователь найден через токен: userID=%s", user.ID)
						}
					}
				}
			}
		}
	}

	// Получаем список подписок пользователя (если пользователь найден)
	var subscribedShopIDs map[uuid.UUID]bool
	if userFound && userID != nil {
		var subscriptions []models.ShopSubscription
		if err := database.DB.Where("user_id = ?", *userID).Find(&subscriptions).Error; err == nil {
			subscribedShopIDs = make(map[uuid.UUID]bool)
			for _, sub := range subscriptions {
				subscribedShopIDs[sub.ShopID] = true
			}
			log.Printf("📋 [GetShops] Найдено подписок: %d", len(subscribedShopIDs))
		}
	}

	// Пагинация
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Получаем магазины из новой таблицы shops
	var shops []models.Shop
	var total int64

	query := database.DB.Model(&models.Shop{}).Where("is_active = ?", true)

	// Фильтрация по городу
	if cityID := c.Query("city_id"); cityID != "" {
		if cityUUID, err := uuid.Parse(cityID); err == nil {
			query = query.Where("city_id = ?", cityUUID)
		}
	}

	// Поиск по названию
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Подсчитываем общее количество
	query.Count(&total)

	// Получаем магазины с загрузкой связей
	if err := query.Preload("Owner.Role").
		Preload("City").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&shops).Error; err != nil {
		log.Printf("❌ [GetShops] Ошибка получения магазинов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch shops",
		})
		return
	}

	// Также получаем магазины из legacy таблицы users (shop_owner)
	var legacyShops []models.User
	var legacyTotal int64

	legacyQuery := database.DB.Model(&models.User{}).
		Joins("JOIN roles ON users.role_id = roles.id").
		Where("roles.name = ? AND users.is_active = ?", "shop_owner", true)

	// Применяем те же фильтры
	if search := c.Query("search"); search != "" {
		legacyQuery = legacyQuery.Where("users.name ILIKE ?", "%"+search+"%")
	}

	legacyQuery.Count(&legacyTotal)

	if err := legacyQuery.Preload("Role").
		Order("users.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&legacyShops).Error; err == nil {
		log.Printf("📋 [GetShops] Найдено legacy магазинов: %d", len(legacyShops))
	}

	// Формируем ответ
	shopResponses := make([]gin.H, 0)

	// Обрабатываем новые магазины
	for _, shop := range shops {
		var productsCount int64
		var subscribersCount int64

		database.DB.Model(&models.Product{}).Where("shop_id = ? OR owner_id = ?", shop.ID, shop.OwnerID).Count(&productsCount)
		database.DB.Model(&models.ShopSubscription{}).Where("shop_id = ?", shop.ID).Count(&subscribersCount)

		isSubscribed := false
		if subscribedShopIDs != nil {
			isSubscribed = subscribedShopIDs[shop.ID]
		}

		// Получаем информацию о лицензии магазина (подписке магазина на платформу)
		// Ищем последнюю лицензию для магазина (даже если она неактивна или истекла)
		var license models.License
		var licenseInfo gin.H
		err := database.DB.Where("shop_id = ?", shop.ID).
			Order("created_at DESC").
			First(&license).Error
		
		if err == nil {
			log.Printf("✅ [GetShops] Найдена лицензия для магазина %s: licenseKey=%s, status=%s", shop.ID, license.LicenseKey, license.SubscriptionStatus)
			// Вычисляем оставшиеся дни
			var daysRemaining *int
			if license.ExpiresAt != nil {
				days := int(time.Until(*license.ExpiresAt).Hours() / 24)
				if days > 0 {
					daysRemaining = &days
				} else {
					zero := 0
					daysRemaining = &zero
				}
			}

			// Получаем план подписки для отображения цены
			var subscriptionPlan models.SubscriptionPlan
			planPrice := license.PaymentAmount
			planCurrency := license.PaymentCurrency
			if license.SubscriptionType != "" {
				if err := database.DB.Where("subscription_type = ? AND is_active = ?", license.SubscriptionType, true).
					Order("sort_order ASC").
					First(&subscriptionPlan).Error; err == nil {
					// Используем цену из плана, если она есть
					if subscriptionPlan.Price > 0 {
						planPrice = subscriptionPlan.Price
						planCurrency = subscriptionPlan.Currency
					}
				}
			}

			licenseInfo = gin.H{
				"licenseKey":       license.LicenseKey,
				"activatedAt":      license.ActivatedAt,
				"expiresAt":        license.ExpiresAt,
				"daysRemaining":    daysRemaining,
				"price":            planPrice,
				"currency":         planCurrency,
				"subscriptionType": license.SubscriptionType,
				"subscriptionStatus": license.SubscriptionStatus,
				"isValid":          license.IsValid(),
				"isExpired":        license.IsExpired(),
			}
		}

		shopResponse := gin.H{
			"id":               shop.ID,
			"name":             shop.Name,
			"inn":              shop.INN,
			"description":      shop.Description,
			"logo":             shop.Logo,
			"email":            shop.Email,
			"phone":            shop.Phone,
			"address":          shop.Address,
			"rating":           shop.Rating,
			"isActive":         shop.IsActive,
			"ownerId":          shop.OwnerID,
			"productsCount":    productsCount,
			"subscribersCount": subscribersCount,
			"isSubscribed":     isSubscribed,
			"createdAt":        shop.CreatedAt,
		}

		// Добавляем информацию о лицензии, если она есть
		if licenseInfo != nil {
			shopResponse["license"] = licenseInfo
		}

		if shop.City != nil {
			shopResponse["city"] = gin.H{
				"id":   shop.City.ID,
				"name": shop.City.Name,
			}
		}

		shopResponses = append(shopResponses, shopResponse)
	}

	// Обрабатываем legacy магазины
	for _, shopUser := range legacyShops {
		// Проверяем, не добавлен ли уже этот магазин (может быть дубликат)
		isDuplicate := false
		for _, shop := range shops {
			if shop.OwnerID == shopUser.ID {
				isDuplicate = true
				break
			}
		}

		if isDuplicate {
			continue
		}

		var productsCount int64
		var subscribersCount int64

		database.DB.Model(&models.Product{}).Where("owner_id = ?", shopUser.ID).Count(&productsCount)
		database.DB.Model(&models.ShopSubscription{}).Where("shop_id = ?", shopUser.ID).Count(&subscribersCount)

		isSubscribed := false
		if subscribedShopIDs != nil {
			isSubscribed = subscribedShopIDs[shopUser.ID]
		}

		// Получаем информацию о лицензии магазина (для legacy магазинов shop_id = user_id)
		var license models.License
		var licenseInfo gin.H
		if err := database.DB.Where("shop_id = ?", shopUser.ID).
			Order("created_at DESC").
			First(&license).Error; err == nil {
			// Вычисляем оставшиеся дни
			var daysRemaining *int
			if license.ExpiresAt != nil {
				days := int(time.Until(*license.ExpiresAt).Hours() / 24)
				if days > 0 {
					daysRemaining = &days
				} else {
					zero := 0
					daysRemaining = &zero
				}
			}

			// Получаем план подписки для отображения цены
			var subscriptionPlan models.SubscriptionPlan
			planPrice := license.PaymentAmount
			planCurrency := license.PaymentCurrency
			if license.SubscriptionType != "" {
				if err := database.DB.Where("subscription_type = ? AND is_active = ?", license.SubscriptionType, true).
					Order("sort_order ASC").
					First(&subscriptionPlan).Error; err == nil {
					// Используем цену из плана, если она есть
					if subscriptionPlan.Price > 0 {
						planPrice = subscriptionPlan.Price
						planCurrency = subscriptionPlan.Currency
					}
				}
			}

			licenseInfo = gin.H{
				"licenseKey":       license.LicenseKey,
				"activatedAt":      license.ActivatedAt,
				"expiresAt":        license.ExpiresAt,
				"daysRemaining":    daysRemaining,
				"price":            planPrice,
				"currency":         planCurrency,
				"subscriptionType": license.SubscriptionType,
				"subscriptionStatus": license.SubscriptionStatus,
				"isValid":          license.IsValid(),
				"isExpired":        license.IsExpired(),
			}
		}

		shopResponse := gin.H{
			"id":               shopUser.ID,
			"name":             shopUser.Name,
			"inn":              shopUser.INN,
			"description":      "",
			"logo":             shopUser.Avatar,
			"email":            shopUser.Email,
			"phone":            shopUser.Phone,
			"address":          "",
			"rating":           0,
			"isActive":         shopUser.IsActive,
			"ownerId":          shopUser.ID,
			"productsCount":    productsCount,
			"subscribersCount": subscribersCount,
			"isSubscribed":     isSubscribed,
			"createdAt":        shopUser.CreatedAt,
		}

		// Добавляем информацию о лицензии, если она есть
		if licenseInfo != nil {
			shopResponse["license"] = licenseInfo
		}

		shopResponses = append(shopResponses, shopResponse)
	}

	// Обновляем total с учетом legacy магазинов
	total = total + legacyTotal

	log.Printf("✅ [GetShops] Возвращаем %d магазинов (страница %d, лимит %d)", len(shopResponses), page, limit)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"shops": shopResponses,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	})
}

// GetShopsWithLicenses возвращает список всех магазинов с информацией о лицензиях (админ)
func (sc *ShopController) GetShopsWithLicenses(c *gin.Context) {
	log.Printf("🛍️ [GetShopsWithLicenses] Начало получения списка магазинов с лицензиями")

	// Пагинация
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Получаем магазины из новой таблицы shops
	var shops []models.Shop
	var total int64

	query := database.DB.Model(&models.Shop{}).Where("is_active = ?", true)

	// Фильтрация по наличию лицензии
	if hasLicense := c.Query("hasLicense"); hasLicense != "" {
		if hasLicense == "true" {
			query = query.Joins("INNER JOIN licenses ON licenses.shop_id = shops.id AND licenses.subscription_status = 'active'")
		} else if hasLicense == "false" {
			query = query.Where("NOT EXISTS (SELECT 1 FROM licenses WHERE licenses.shop_id = shops.id AND licenses.subscription_status = 'active')")
		}
	}

	// Поиск по названию
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Подсчитываем общее количество
	query.Count(&total)

	// Получаем магазины с загрузкой связей
	if err := query.Preload("Owner.Role").
		Preload("City").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&shops).Error; err != nil {
		log.Printf("❌ [GetShopsWithLicenses] Ошибка получения магазинов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch shops",
		})
		return
	}

	// Также получаем магазины из legacy таблицы users (shop_owner)
	var legacyShops []models.User
	var legacyTotal int64

	legacyQuery := database.DB.Model(&models.User{}).
		Joins("JOIN roles ON users.role_id = roles.id").
		Where("roles.name = ? AND users.is_active = ?", "shop_owner", true)

	if search := c.Query("search"); search != "" {
		legacyQuery = legacyQuery.Where("users.name ILIKE ?", "%"+search+"%")
	}

	legacyQuery.Count(&legacyTotal)

	if err := legacyQuery.
		Preload("Role").
		Order("users.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&legacyShops).Error; err == nil {
		log.Printf("📋 [GetShopsWithLicenses] Найдено legacy магазинов: %d", len(legacyShops))
	}

	// Формируем ответ
	shopResponses := make([]gin.H, 0)

	// Обрабатываем новые магазины
	for _, shop := range shops {
		var productsCount int64
		var subscribersCount int64

		database.DB.Model(&models.Product{}).Where("shop_id = ? OR owner_id = ?", shop.ID, shop.OwnerID).Count(&productsCount)
		database.DB.Model(&models.ShopSubscription{}).Where("shop_id = ?", shop.ID).Count(&subscribersCount)

		// Получаем информацию о лицензии магазина (последняя активная или последняя вообще)
		var license models.License
		var licenseInfo gin.H
		
		// Сначала ищем активную лицензию
		err := database.DB.Where("shop_id = ? AND subscription_status = ?", shop.ID, models.SubscriptionStatusActive).
			Order("created_at DESC").
			First(&license).Error
		
		// Если активной нет, берем последнюю любую
		if err != nil {
			err = database.DB.Where("shop_id = ?", shop.ID).
				Order("created_at DESC").
				First(&license).Error
		}

		if err == nil {
			// Вычисляем оставшиеся дни
			var daysRemaining *int
			if license.ExpiresAt != nil {
				days := int(time.Until(*license.ExpiresAt).Hours() / 24)
				if days > 0 {
					daysRemaining = &days
				} else {
					zero := 0
					daysRemaining = &zero
				}
			}

			// Получаем план подписки для отображения цены
			var subscriptionPlan models.SubscriptionPlan
			planPrice := license.PaymentAmount
			planCurrency := license.PaymentCurrency
			if license.SubscriptionType != "" {
				if err := database.DB.Where("subscription_type = ? AND is_active = ?", license.SubscriptionType, true).
					Order("sort_order ASC").
					First(&subscriptionPlan).Error; err == nil {
					if subscriptionPlan.Price > 0 {
						planPrice = subscriptionPlan.Price
						planCurrency = subscriptionPlan.Currency
					}
				}
			}

			licenseInfo = gin.H{
				"id":                license.ID,
				"licenseKey":        license.LicenseKey,
				"activatedAt":       license.ActivatedAt,
				"expiresAt":         license.ExpiresAt,
				"daysRemaining":     daysRemaining,
				"price":              planPrice,
				"currency":           planCurrency,
				"subscriptionType":   license.SubscriptionType,
				"subscriptionStatus": license.SubscriptionStatus,
				"isValid":           license.IsValid(),
				"isExpired":         license.IsExpired(),
				"paymentProvider":   license.PaymentProvider,
			}
		}

		shopResponse := gin.H{
			"id":               shop.ID,
			"name":             shop.Name,
			"description":      shop.Description,
			"email":            shop.Email,
			"phone":            shop.Phone,
			"logo":             shop.Logo,
			"rating":           shop.Rating,
			"isActive":         shop.IsActive,
			"productsCount":    productsCount,
			"subscribersCount": subscribersCount,
			"owner": gin.H{
				"id":    shop.OwnerID,
				"name":  shop.Owner.Name,
				"email": shop.Owner.Email,
			},
			"hasLicense": licenseInfo != nil,
			"createdAt":  shop.CreatedAt,
		}

		// Добавляем информацию о лицензии, если она есть
		if licenseInfo != nil {
			shopResponse["license"] = licenseInfo
		}

		shopResponses = append(shopResponses, shopResponse)
	}

	// Обрабатываем legacy магазины
	for _, shopUser := range legacyShops {
		var productsCount int64
		var subscribersCount int64

		database.DB.Model(&models.Product{}).Where("owner_id = ?", shopUser.ID).Count(&productsCount)
		database.DB.Model(&models.ShopSubscription{}).Where("shop_id = ?", shopUser.ID).Count(&subscribersCount)

		// Получаем информацию о лицензии
		var license models.License
		var licenseInfo gin.H
		
		err := database.DB.Where("shop_id = ? AND subscription_status = ?", shopUser.ID, models.SubscriptionStatusActive).
			Order("created_at DESC").
			First(&license).Error
		
		if err != nil {
			err = database.DB.Where("shop_id = ?", shopUser.ID).
				Order("created_at DESC").
				First(&license).Error
		}

		if err == nil {
			var daysRemaining *int
			if license.ExpiresAt != nil {
				days := int(time.Until(*license.ExpiresAt).Hours() / 24)
				if days > 0 {
					daysRemaining = &days
				} else {
					zero := 0
					daysRemaining = &zero
				}
			}

			var subscriptionPlan models.SubscriptionPlan
			planPrice := license.PaymentAmount
			planCurrency := license.PaymentCurrency
			if license.SubscriptionType != "" {
				if err := database.DB.Where("subscription_type = ? AND is_active = ?", license.SubscriptionType, true).
					Order("sort_order ASC").
					First(&subscriptionPlan).Error; err == nil {
					if subscriptionPlan.Price > 0 {
						planPrice = subscriptionPlan.Price
						planCurrency = subscriptionPlan.Currency
					}
				}
			}

			licenseInfo = gin.H{
				"id":                license.ID,
				"licenseKey":        license.LicenseKey,
				"activatedAt":       license.ActivatedAt,
				"expiresAt":         license.ExpiresAt,
				"daysRemaining":     daysRemaining,
				"price":              planPrice,
				"currency":           planCurrency,
				"subscriptionType":   license.SubscriptionType,
				"subscriptionStatus": license.SubscriptionStatus,
				"isValid":           license.IsValid(),
				"isExpired":         license.IsExpired(),
				"paymentProvider":   license.PaymentProvider,
			}
		}

		shopResponse := gin.H{
			"id":               shopUser.ID,
			"name":             shopUser.Name,
			"description":      "",
			"email":            shopUser.Email,
			"phone":            shopUser.Phone,
			"logo":             shopUser.Avatar,
			"rating":           0.0,
			"isActive":         shopUser.IsActive,
			"productsCount":    productsCount,
			"subscribersCount": subscribersCount,
			"owner": gin.H{
				"id":    shopUser.ID,
				"name":  shopUser.Name,
				"email": shopUser.Email,
			},
			"hasLicense": licenseInfo != nil,
			"createdAt":  shopUser.CreatedAt,
		}

		if licenseInfo != nil {
			shopResponse["license"] = licenseInfo
		}

		shopResponses = append(shopResponses, shopResponse)
	}

	// Общее количество магазинов
	totalShops := total + legacyTotal

	pagination := gin.H{
		"page":       page,
		"limit":      limit,
		"total":      totalShops,
		"totalPages": (totalShops + int64(limit) - 1) / int64(limit),
	}

	log.Printf("✅ [GetShopsWithLicenses] Возвращаем %d магазинов (страница %d, лимит %d)", len(shopResponses), page, limit)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"shops":      shopResponses,
			"pagination": pagination,
		},
	})
}

