package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

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

	var shop models.User
	// Загружаем пользователя с ролью shop_owner
	if err := database.DB.Preload("Role").Where("id = ? AND role_id IN (SELECT id FROM roles WHERE name = 'shop_owner')", shopID).First(&shop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Shop not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
		}
		return
	}

	// Проверяем, что это действительно магазин
	if shop.Role == nil || shop.Role.Name != "shop_owner" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Shop not found",
		})
		return
	}

	shopInfo := models.ShopInfo{
		ID:   shop.ID,
		Name: shop.Name,
		INN:  shop.INN,
	}

	// Подсчитываем количество товаров
	var productsCount int64
	database.DB.Model(&models.Product{}).Where("owner_id = ?", shop.ID).Count(&productsCount)

	// Подсчитываем количество подписчиков
	var subscribersCount int64
	database.DB.Model(&models.ShopSubscription{}).Where("shop_id = ?", shop.ID).Count(&subscribersCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"shop": gin.H{
				"id":            shopInfo.ID,
				"name":          shopInfo.Name,
				"inn":           shopInfo.INN,
				"email":         shop.Email,
				"phone":         shop.Phone,
				"avatar":        shop.Avatar,
				"productsCount": productsCount,
				"subscribersCount": subscribersCount,
				"createdAt":     shop.CreatedAt,
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

	// Проверяем, что магазин существует и является shop_owner
	var shop models.User
	if err := database.DB.Preload("Role").Where("id = ?", shopID).First(&shop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Shop not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
		}
		return
	}

	if shop.Role == nil || shop.Role.Name != "shop_owner" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Shop not found",
		})
		return
	}

	var products []models.Product
	query := database.DB.Model(&models.Product{}).Where("owner_id = ?", shopID)

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
	if err := query.Offset(offset).Limit(limit).
		Preload("Variations").
		Preload("Category").
		Preload("Owner.Role").
		Find(&products).Error; err != nil {
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
			"shop": models.ShopInfo{
				ID:   shop.ID,
				Name: shop.Name,
				INN:  shop.INN,
			},
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

	// Проверяем, что магазин существует и является shop_owner
	var shop models.User
	if err := database.DB.Preload("Role").Where("id = ?", shopID).First(&shop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Shop not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
		}
		return
	}

	if shop.Role == nil || shop.Role.Name != "shop_owner" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Shop not found",
		})
		return
	}

	// Проверяем, что пользователь не подписывается сам на себя
	if user.ID == shopID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot subscribe to your own shop",
		})
		return
	}

	// Проверяем, не подписан ли уже
	var existingSubscription models.ShopSubscription
	if err := database.DB.Where("user_id = ? AND shop_id = ?", user.ID, shopID).First(&existingSubscription).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Already subscribed",
			"data":    existingSubscription.ToResponse(),
		})
		return
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
	database.DB.Preload("Role").First(&subscription.Shop, "id = ?", shopID)

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

	// Проверяем, что текущий пользователь - владелец магазина или админ
	if user.ID != shopID && (user.Role == nil || (user.Role.Name != "admin" && user.Role.Name != "super_admin")) {
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid shop ID",
		})
		return
	}

	var userID uuid.UUID
	var userFound bool

	// Пробуем получить пользователя из контекста (если был middleware)
	currentUser, exists := c.Get("user")
	if exists {
		user := currentUser.(models.User)
		userID = user.ID
		userFound = true
		log.Printf("🔍 [CheckSubscription] Пользователь найден через middleware: %s", userID)
	} else {
		// Если пользователя нет в контексте, пробуем опционально проверить токен
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != authHeader {
				// Валидируем токен
				claims, err := utils.ValidateJWT(tokenString)
				if err == nil {
					parsedUserID, err := uuid.Parse(claims.UserID)
					if err == nil {
						// Проверяем, что пользователь существует и активен
						var user models.User
						if err := database.DB.Preload("Role").First(&user, "id = ? AND is_active = ?", parsedUserID, true).Error; err == nil {
							userID = user.ID
							userFound = true
							log.Printf("🔍 [CheckSubscription] Пользователь найден через токен: %s", userID)
						}
					}
				}
			}
		}
	}

	// Если пользователь не найден, возвращаем false
	if !userFound {
		log.Printf("🔍 [CheckSubscription] Пользователь не найден, возвращаем false")
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"isSubscribed": false,
		})
		return
	}

	// Проверяем подписку
	var subscription models.ShopSubscription
	isSubscribed := database.DB.Where("user_id = ? AND shop_id = ?", userID, shopID).First(&subscription).Error == nil
	log.Printf("🔍 [CheckSubscription] Проверка подписки: userID=%s, shopID=%s, isSubscribed=%v", userID, shopID, isSubscribed)

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"isSubscribed": isSubscribed,
	})
}

