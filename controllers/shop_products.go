package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"

	"github.com/gin-gonic/gin"
)

// GetShopProducts возвращает товары только для владельца магазина (фильтрует по owner_id)
func (pc *ProductController) GetShopProducts(c *gin.Context) {
	var products []models.Product
	query := database.DB.Model(&models.Product{})

	// Получаем текущего пользователя из контекста
	currentUser, exists := c.Get("user")
	if !exists {
		log.Printf("❌ Пользователь не найден в контексте!")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	user := currentUser.(models.User)
	log.Printf("🏪 Владелец магазина %s (ID: %s, email: %s) запрашивает свои товары", user.Name, user.ID, user.Email)

	// Фильтруем товары по ShopID или OwnerID (обратная совместимость)
	// Сначала пробуем найти shop для этого пользователя
	var shop models.Shop
	if err := database.DB.Where("owner_id = ?", user.ID).First(&shop).Error; err == nil {
		// Пользователь владеет shop - фильтруем по shop_id
		query = query.Where("shop_id = ? OR owner_id = ?", shop.ID, user.ID)
		log.Printf("🔍 Фильтруем товары по ShopID: %s", shop.ID)
	} else {
		// Обратная совместимость: фильтруем по owner_id
		query = query.Where("owner_id = ?", user.ID)
		log.Printf("🔍 Фильтруем товары по OwnerID: %s", user.ID)
	}

	// Фильтрация по категории
	if category := c.Query("category"); category != "" {
		query = query.Where("category_id = ?", category)
	}

	// Фильтрация по полу (gender) - для раздела "Детям": male, female, unisex
	if gender := c.Query("gender"); gender != "" {
		query = query.Where("gender = ?", gender)
	}

	// Поиск по названию
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Сортировка
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	allowedSorts := map[string]bool{
		"name":       true,
		"price":      true,
		"created_at": true,
	}

	if allowedSorts[sortBy] {
		if sortOrder == "asc" {
			query = query.Order(sortBy + " ASC")
		} else {
			query = query.Order(sortBy + " DESC")
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

	// Получаем продукты с предзагрузкой связей
	if err := query.Offset(offset).Limit(limit).Preload("Variations").Preload("Category").Preload("Owner.Role").Find(&products).Error; err != nil {
		log.Printf("❌ Ошибка получения товаров владельца магазина: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch shop products",
		})
		return
	}

	log.Printf("🏪 Загружено %d товаров для владельца магазина %s", len(products), user.Email)

	// Преобразуем в ответ
	var responseProducts []models.ProductResponse
	for _, product := range products {
		responseProducts = append(responseProducts, product.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"products": responseProducts,
			"total":    total,
			"page":     page,
			"limit":    limit,
		},
		"message": "Shop products loaded successfully",
	})
}
