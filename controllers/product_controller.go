package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductController обрабатывает запросы продуктов
type ProductController struct{}

// GetProducts возвращает список продуктов с фильтрацией
func (pc *ProductController) GetProducts(c *gin.Context) {
	var products []models.Product
	query := database.DB.Model(&models.Product{})

	// Получаем текущего пользователя из контекста
	currentUser, exists := c.Get("user")
	if exists {
		user := currentUser.(models.User)
		// Фильтруем товары по OwnerID пользователя
		query = query.Where("owner_id = ?", user.ID)
		log.Printf("🔍 Фильтруем товары по OwnerID пользователя: %s (email: %s, role: %s)", user.ID, user.Email, user.Role)
	} else {
		log.Printf("⚠️ Пользователь не найден в контексте!")
	}

	// Фильтрация по категории
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// Фильтрация по наличию на складе
	if inStock := c.Query("in_stock"); inStock == "true" {
		query = query.Where("in_stock = ?", true)
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

	// Логируем SQL запрос для отладки
	log.Printf("🔍 SQL запрос: %v", query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]models.Product{})
	}))

	// Получаем продукты
	if err := query.Offset(offset).Limit(limit).Preload("Variations").Preload("Category").Find(&products).Error; err != nil {
		log.Printf("❌ Ошибка получения товаров: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch products",
		})
		return
	}

	log.Printf("📦 Загружено %d товаров", len(products))
	for i, product := range products {
		log.Printf("📦 Товар %d: ID=%s, CategoryID=%s, Category=%v",
			i+1, product.ID, product.CategoryID, product.Category)
		log.Printf("📦 Товар %d вариации: %+v", i+1, product.Variations)
		log.Printf("📦 Товар %d количество вариаций: %d", i+1, len(product.Variations))
	}

	// Преобразуем в response
	productResponses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = product.ToResponse()
		log.Printf("📦 Response %d: Category=%v", i+1, productResponses[i].Category)
	}

	c.JSON(http.StatusOK, gin.H{
		"products": productResponses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetProduct возвращает один продукт по ID
func (pc *ProductController) GetProduct(c *gin.Context) {
	id := c.Param("id")
	productID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	var product models.Product
	query := database.DB.Preload("Variations").Preload("Category").Where("id = ?", productID)

	// Получаем текущего пользователя из контекста
	currentUser, exists := c.Get("user")
	if exists {
		user := currentUser.(models.User)
		// Фильтруем товар по OwnerID пользователя
		query = query.Where("owner_id = ?", user.ID)
		log.Printf("🔍 Фильтруем товар по OwnerID пользователя: %s", user.ID)
	}

	if err := query.First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
		}
		return
	}

	log.Printf("📦 Получен товар: ID=%s, Name=%s", product.ID, product.Name)
	log.Printf("📦 Вариации товара: %+v", product.Variations)
	log.Printf("📦 Количество вариаций: %d", len(product.Variations))

	c.JSON(http.StatusOK, gin.H{
		"product": product.ToResponse(),
	})
}

// CreateProduct создает новый продукт (только для админов)
func (pc *ProductController) CreateProduct(c *gin.Context) {
	log.Printf("🛍️ Начало создания товара...")

	var req models.ProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Ошибка валидации JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ JSON валидация прошла успешно")
	log.Printf("📋 Данные товара: %+v", req)
	log.Printf("🎨 Количество вариаций: %d", len(req.Variations))

	for i, variation := range req.Variations {
		log.Printf("🎨 Вариация %d: %+v", i+1, variation)
	}

	// Получаем текущего пользователя из контекста
	currentUser, exists := c.Get("user")
	if !exists {
		log.Printf("❌ Пользователь не найден в контексте")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	user := currentUser.(models.User)
	log.Printf("👤 Создает товар пользователь: %s (ID: %s)", user.Name, user.ID)

	// Создаем продукт
	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Gender:      req.Gender,
		CategoryID:  req.CategoryID,
		Brand:       req.Brand,
		IsAvailable: true,
		OwnerID:     &user.ID, // Привязываем товар к текущему пользователю

	}

	log.Printf("🏷️ Создаем товар: %+v", product)

	// Начинаем транзакцию
	tx := database.DB.Begin()
	log.Printf("💾 Начинаем транзакцию")

	// Создаем продукт
	if err := tx.Create(&product).Error; err != nil {
		log.Printf("❌ Ошибка создания товара: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create product",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Товар создан с ID: %s", product.ID)

	// Создаем вариации
	for i, variationReq := range req.Variations {
		log.Printf("🎨 Создаем вариацию %d/%d", i+1, len(req.Variations))

		variation := models.ProductVariation{
			ProductID:     product.ID,
			Sizes:         variationReq.Sizes,
			Colors:        variationReq.Colors,
			Price:         variationReq.Price,
			OriginalPrice: variationReq.OriginalPrice,
			Discount:      variationReq.Discount,
			ImageURLs:     variationReq.ImageURLs,
			StockQuantity: variationReq.StockQuantity,
			IsAvailable:   variationReq.StockQuantity > 0,
			SKU:           variationReq.SKU,
			Barcode:       variationReq.Barcode,
		}

		log.Printf("🎨 Вариация %d: %+v", i+1, variation)

		if err := tx.Create(&variation).Error; err != nil {
			log.Printf("❌ Ошибка создания вариации %d: %v", i+1, err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create product variation",
				"details": err.Error(),
			})
			return
		}

		log.Printf("✅ Вариация %d создана", i+1)
	}

	// Подтверждаем транзакцию
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Ошибка подтверждения транзакции: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to commit transaction",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Транзакция подтверждена")

	// Загружаем продукт с вариациями для ответа
	var productWithVariations models.Product
	err := database.DB.Preload("Variations").First(&productWithVariations, product.ID).Error
	if err != nil {
		log.Printf("❌ Ошибка загрузки созданного товара: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load created product",
			"details": err.Error(),
		})
		return
	}

	log.Printf("🎉 Товар успешно создан: %s", productWithVariations.Name)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": productWithVariations.ToResponse(),
	})
}

// UpdateProduct обновляет продукт (только для админов)
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	productID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	var req models.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	var product models.Product
	query := database.DB.Where("id = ?", productID)

	// Получаем текущего пользователя из контекста
	currentUser, exists := c.Get("user")
	if exists {
		user := currentUser.(models.User)
		// Проверяем, что товар принадлежит текущему пользователю
		query = query.Where("owner_id = ?", user.ID)
		log.Printf("🔍 Проверяем права на обновление товара по OwnerID: %s", user.ID)
	}

	if err := query.First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
		}
		return
	}

	// Начинаем транзакцию
	tx := database.DB.Begin()

	// Обновляем основные поля продукта
	product.Name = req.Name
	product.Description = req.Description
	product.Gender = req.Gender
	product.CategoryID = req.CategoryID
	product.Brand = req.Brand

	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update product",
		})
		return
	}

	// Удаляем старые вариации
	if err := tx.Where("product_id = ?", productID).Delete(&models.ProductVariation{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete old variations",
		})
		return
	}

	// Создаем новые вариации
	for _, variationReq := range req.Variations {
		variation := models.ProductVariation{
			ProductID:     product.ID,
			Sizes:         variationReq.Sizes,
			Colors:        variationReq.Colors,
			Price:         variationReq.Price,
			OriginalPrice: variationReq.OriginalPrice,
			Discount:      variationReq.Discount,
			ImageURLs:     variationReq.ImageURLs,
			StockQuantity: variationReq.StockQuantity,
			IsAvailable:   variationReq.StockQuantity > 0,
			SKU:           variationReq.SKU,
			Barcode:       variationReq.Barcode,
		}

		if err := tx.Create(&variation).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create product variation",
			})
			return
		}
	}

	// Подтверждаем транзакцию
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	// Загружаем обновленный продукт с вариациями
	var updatedProduct models.Product
	if err := database.DB.Preload("Variations").First(&updatedProduct, productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load updated product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": updatedProduct.ToResponse(),
	})
}

// GetAllProducts возвращает все продукты (только для админов)
func (pc *ProductController) GetAllProducts(c *gin.Context) {
	var products []models.Product
	query := database.DB.Model(&models.Product{})

	// Фильтрация по категории
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// Фильтрация по наличию на складе
	if inStock := c.Query("in_stock"); inStock == "true" {
		query = query.Where("in_stock = ?", true)
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

	// Получаем продукты (все, без фильтрации по пользователю)
	if err := query.Offset(offset).Limit(limit).Preload("Variations").Preload("Category").Preload("Owner").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch products",
		})
		return
	}

	log.Printf("📦 Загружено %d товаров (все товары)", len(products))

	// Преобразуем в response
	productResponses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = product.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"products": productResponses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// DeleteProduct удаляет продукт (только для админов)
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	productID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	// Получаем текущего пользователя из контекста
	currentUser, exists := c.Get("user")
	if exists {
		user := currentUser.(models.User)
		log.Printf("🔍 Проверяем права на удаление товара по OwnerID: %s", user.ID)
	}

	// Начинаем транзакцию
	tx := database.DB.Begin()

	// Проверяем, что товар принадлежит текущему пользователю
	var product models.Product
	query := tx.Where("id = ?", productID)
	if exists {
		user := currentUser.(models.User)
		query = query.Where("owner_id = ?", user.ID)
	}

	if err := query.First(&product).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found or access denied",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
		}
		return
	}

	// Сначала удаляем все вариации товара
	if err := tx.Where("product_id = ?", productID).Delete(&models.ProductVariation{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete product variations",
		})
		return
	}

	// Затем удаляем сам товар
	result := tx.Delete(&models.Product{}, "id = ?", productID)
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete product",
		})
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	// Подтверждаем транзакцию
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}

// GetProductAdmin возвращает один продукт по ID (для админов, без проверки владельца)
func (pc *ProductController) GetProductAdmin(c *gin.Context) {
	id := c.Param("id")
	productID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	var product models.Product
	if err := database.DB.Preload("Variations").Preload("Category").Preload("Owner").First(&product, "id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
		}
		return
	}

	log.Printf("📦 Получен товар (админ): ID=%s, Name=%s, OwnerID=%s", product.ID, product.Name, product.OwnerID)

	c.JSON(http.StatusOK, gin.H{
		"product": product.ToResponse(),
	})
}

// GetProductsWithVariations возвращает продукты с вариациями через JOIN запрос
func (pc *ProductController) GetProductsWithVariations(c *gin.Context) {
	var productsWithVariations []models.ProductWithVariation

	// Строим SQL запрос с JOIN и конвертацией JSON-текста в массивы строк
	query := `
		WITH sizes_arr AS (
			SELECT pv.id AS variation_id,
			       COALESCE(ARRAY(SELECT json_array_elements_text(NULLIF(pv.sizes,'')::json)), ARRAY[]::text[]) AS sizes
			FROM public.product_variations pv
		),
		colors_arr AS (
			SELECT pv.id AS variation_id,
			       COALESCE(ARRAY(SELECT json_array_elements_text(NULLIF(pv.colors,'')::json)), ARRAY[]::text[]) AS colors
			FROM public.product_variations pv
		),
		images_arr AS (
			SELECT pv.id AS variation_id,
			       COALESCE(ARRAY(SELECT json_array_elements_text(NULLIF(pv.image_urls,'')::json)), ARRAY[]::text[]) AS image_urls
			FROM public.product_variations pv
		)
		SELECT
			pv.id AS product_id,
			p.name,
			p.description,
			p.brand,
			sz.sizes,
			cl.colors,
			pv.price,
			pv.original_price,
			im.image_urls,
			pv.stock_quantity,
			pv.sku
		FROM public.products AS p
		INNER JOIN public.product_variations pv ON p.id = pv.product_id
		LEFT JOIN sizes_arr sz ON sz.variation_id = pv.id
		LEFT JOIN colors_arr cl ON cl.variation_id = pv.id
		LEFT JOIN images_arr im ON im.variation_id = pv.id
		WHERE 1=1
		ORDER BY p.created_at DESC
	`

	// Всегда возвращаем все товары без фильтра по владельцу (независимо от токена)
	if err := database.DB.Raw(query).Scan(&productsWithVariations).Error; err != nil {
		log.Printf("❌ Ошибка выполнения JOIN запроса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch products with variations",
		})
		return
	}

	log.Printf("📦 Получено %d записей продуктов с вариациями", len(productsWithVariations))

	// Применяем дополнительные фильтры
	filteredProducts := productsWithVariations

	// Фильтрация по бренду
	if brand := c.Query("brand"); brand != "" {
		var temp []models.ProductWithVariation
		for _, product := range filteredProducts {
			if product.Brand == brand {
				temp = append(temp, product)
			}
		}
		filteredProducts = temp
	}

	// Фильтрация по цене
	if minPrice := c.Query("min_price"); minPrice != "" {
		if min, err := strconv.ParseFloat(minPrice, 64); err == nil {
			var temp []models.ProductWithVariation
			for _, product := range filteredProducts {
				if product.Price >= min {
					temp = append(temp, product)
				}
			}
			filteredProducts = temp
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if max, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			var temp []models.ProductWithVariation
			for _, product := range filteredProducts {
				if product.Price <= max {
					temp = append(temp, product)
				}
			}
			filteredProducts = temp
		}
	}

	// Поиск по названию или описанию
	if search := c.Query("search"); search != "" {
		var temp []models.ProductWithVariation
		searchLower := strings.ToLower(search)
		for _, product := range filteredProducts {
			if strings.Contains(strings.ToLower(product.Name), searchLower) ||
				strings.Contains(strings.ToLower(product.Description), searchLower) {
				temp = append(temp, product)
			}
		}
		filteredProducts = temp
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

	total := len(filteredProducts)
	offset := (page - 1) * limit

	// Применяем пагинацию
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var paginatedProducts []models.ProductWithVariation
	if start < total {
		paginatedProducts = filteredProducts[start:end]
	} else {
		paginatedProducts = []models.ProductWithVariation{}
	}

	log.Printf("📦 Возвращаем %d записей (страница %d из %d)", len(paginatedProducts), page, (total+limit-1)/limit)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    paginatedProducts,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
		"message": "Products with variations retrieved successfully",
	})
}
