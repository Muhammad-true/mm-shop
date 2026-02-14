package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"gorm.io/gorm"
)

// PosController обрабатывает запросы от POS программы владельца магазина
type PosController struct{}

// BulkUploadRequest представляет запрос на массовую загрузку товаров со склада
type BulkUploadRequest struct {
	Products []BulkProductItem `json:"products" binding:"required"`
}

// BulkProductItem представляет товар для массовой загрузки
type BulkProductItem struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Brand       string                 `json:"brand"`
	CategoryID  uuid.UUID              `json:"categoryId" binding:"required"`
	Gender      string                 `json:"gender" binding:"required"` // 'male', 'female', 'unisex'
	Variations  []BulkVariationItem    `json:"variations" binding:"required,min=1"`
}

// BulkVariationItem представляет вариацию для массовой загрузки
type BulkVariationItem struct {
	Sizes            []string           `json:"sizes" binding:"required"`
	Colors           []string           `json:"colors" binding:"required"`
	Price            float64            `json:"price" binding:"required,gt=0"`
	OriginalPrice    *float64           `json:"originalPrice"`
	Discount         int                `json:"discount"`
	StockQuantity    int                `json:"stockQuantity" binding:"required,min=0"` // Количество на складе
	SKU              string             `json:"sku"`
	Barcode          string             `json:"barcode"`
	ImageURLs        []string           `json:"imageUrls"`
	ImageURLsByColor map[string][]string `json:"imageUrlsByColor"`
}

// BulkUploadResponse представляет ответ на массовую загрузку
type BulkUploadResponse struct {
	Success      bool     `json:"success"`
	TotalItems   int      `json:"totalItems"`
	Created      int      `json:"created"`
	Updated      int      `json:"updated"`
	Failed       int      `json:"failed"`
	Errors       []string `json:"errors,omitempty"`
	ProductIDs   []string `json:"productIds,omitempty"`
}

// SaleSyncRequest представляет запрос на синхронизацию продаж
type SaleSyncRequest struct {
	Sales []SaleItem `json:"sales" binding:"required,min=1"`
}

// SaleItem представляет одну продажу
type SaleItem struct {
	VariationID uuid.UUID `json:"variationId" binding:"required"`
	Quantity    int       `json:"quantity" binding:"required,gt=0"`
	Size        string    `json:"size"`
	Color       string    `json:"color"`
	Price       float64   `json:"price" binding:"required,gt=0"`
	SaleDate    time.Time `json:"saleDate"`
}

// SaleSyncResponse представляет ответ на синхронизацию продаж
type SaleSyncResponse struct {
	Success      bool     `json:"success"`
	TotalSales   int      `json:"totalSales"`
	Processed    int      `json:"processed"`
	Failed       int      `json:"failed"`
	Errors       []string `json:"errors,omitempty"`
	UpdatedStock []StockUpdate `json:"updatedStock,omitempty"`
}

// StockUpdate представляет обновление количества
type StockUpdate struct {
	VariationID uuid.UUID `json:"variationId"`
	OldQuantity int       `json:"oldQuantity"`
	NewQuantity int       `json:"newQuantity"`
}

// StockInfo представляет информацию о количестве товара
type StockInfo struct {
	VariationID uuid.UUID `json:"variationId"`
	ProductID   uuid.UUID `json:"productId"`
	ProductName string    `json:"productName"`
	SKU         string    `json:"sku"`
	Barcode     string    `json:"barcode"`
	Sizes       []string  `json:"sizes"`
	Colors      []string  `json:"colors"`
	StockQuantity int     `json:"stockQuantity"`
	IsAvailable   bool    `json:"isAvailable"`
}

// BulkUploadProducts обрабатывает массовую загрузку товаров со склада
// POST /api/v1/pos/products/bulk-upload
func (pc *PosController) BulkUploadProducts(c *gin.Context) {
	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	userModel := user.(models.User)

	// Получаем shop для этого пользователя
	var shop models.Shop
	if err := database.DB.Where("owner_id = ?", userModel.ID).First(&shop).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Shop not found for this user",
		})
		return
	}

	var req BulkUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	if len(req.Products) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No products provided",
		})
		return
	}

	response := BulkUploadResponse{
		TotalItems: len(req.Products),
		Errors:     []string{},
		ProductIDs: []string{},
	}

	// Начинаем транзакцию
	tx := database.DB.Begin()

	for i, productItem := range req.Products {
		// Проверяем существование категории
		var category models.Category
		if err := tx.First(&category, "id = ?", productItem.CategoryID).Error; err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Product %d: category not found", i+1))
			continue
		}

		// Ищем существующий товар по названию и бренду (или по штрих-коду)
		var existingProduct models.Product
		query := tx.Where("shop_id = ? AND name = ?", shop.ID, productItem.Name)
		if productItem.Brand != "" {
			query = query.Where("brand = ?", productItem.Brand)
		}

		err := query.First(&existingProduct).Error

		if err == gorm.ErrRecordNotFound {
			// Создаем новый товар
			product := models.Product{
				Name:        productItem.Name,
				Description: productItem.Description,
				Brand:       productItem.Brand,
				CategoryID:  productItem.CategoryID,
				Gender:      productItem.Gender,
				ShopID:      &shop.ID,
				CityID:      shop.CityID,
				IsAvailable: true,
			}

			if err := tx.Create(&product).Error; err != nil {
				tx.Rollback()
				response.Failed++
				response.Errors = append(response.Errors, fmt.Sprintf("Product %d: failed to create - %v", i+1, err))
				continue
			}

			// Создаем вариации
			for _, variationItem := range productItem.Variations {
				imageURLsByColor := variationItem.ImageURLsByColor
				if imageURLsByColor == nil {
					imageURLsByColor = make(map[string][]string)
				}

				variation := models.ProductVariation{
					ProductID:        product.ID,
					Sizes:            variationItem.Sizes,
					Colors:           variationItem.Colors,
					Price:            variationItem.Price,
					OriginalPrice:    variationItem.OriginalPrice,
					Discount:         variationItem.Discount,
					ImageURLs:        variationItem.ImageURLs,
					ImageURLsByColor: imageURLsByColor,
					StockQuantity:    variationItem.StockQuantity,
					IsAvailable:      variationItem.StockQuantity > 0,
					SKU:              variationItem.SKU,
					Barcode:          variationItem.Barcode,
				}

				if err := tx.Create(&variation).Error; err != nil {
					log.Printf("❌ Failed to create variation for product %s: %v", product.Name, err)
					response.Errors = append(response.Errors, fmt.Sprintf("Product %d: failed to create variation - %v", i+1, err))
				}
			}

			response.Created++
			response.ProductIDs = append(response.ProductIDs, product.ID.String())
			log.Printf("✅ Created product: %s (ID: %s)", product.Name, product.ID)

		} else if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Product %d: database error - %v", i+1, err))
			continue
		} else {
			// Обновляем существующий товар
			existingProduct.Name = productItem.Name
			existingProduct.Description = productItem.Description
			existingProduct.Brand = productItem.Brand
			existingProduct.CategoryID = productItem.CategoryID
			existingProduct.Gender = productItem.Gender

			if err := tx.Save(&existingProduct).Error; err != nil {
				response.Failed++
				response.Errors = append(response.Errors, fmt.Sprintf("Product %d: failed to update - %v", i+1, err))
				continue
			}

			// Обновляем или создаем вариации
			for _, variationItem := range productItem.Variations {
				// Ищем вариацию по штрих-коду или по размерам/цветам
				var existingVariation models.ProductVariation
				var findErr error

				if variationItem.Barcode != "" {
					findErr = tx.Where("product_id = ? AND barcode = ?", existingProduct.ID, variationItem.Barcode).First(&existingVariation).Error
				} else {
					// Ищем по размерам и цветам (точное совпадение)
					findErr = tx.Where("product_id = ?", existingProduct.ID).
						Where("sizes = ? AND colors = ?", variationItem.Sizes, variationItem.Colors).
						First(&existingVariation).Error
				}

				imageURLsByColor := variationItem.ImageURLsByColor
				if imageURLsByColor == nil {
					imageURLsByColor = make(map[string][]string)
				}

				if findErr == gorm.ErrRecordNotFound {
					// Создаем новую вариацию
					variation := models.ProductVariation{
						ProductID:        existingProduct.ID,
						Sizes:            variationItem.Sizes,
						Colors:           variationItem.Colors,
						Price:            variationItem.Price,
						OriginalPrice:    variationItem.OriginalPrice,
						Discount:         variationItem.Discount,
						ImageURLs:        variationItem.ImageURLs,
						ImageURLsByColor: imageURLsByColor,
						StockQuantity:    variationItem.StockQuantity,
						IsAvailable:      variationItem.StockQuantity > 0,
						SKU:              variationItem.SKU,
						Barcode:          variationItem.Barcode,
					}

					if err := tx.Create(&variation).Error; err != nil {
						log.Printf("❌ Failed to create variation: %v", err)
						response.Errors = append(response.Errors, fmt.Sprintf("Product %d: failed to create variation - %v", i+1, err))
					}
				} else if findErr != nil {
					response.Errors = append(response.Errors, fmt.Sprintf("Product %d: error finding variation - %v", i+1, findErr))
				} else {
					// Обновляем существующую вариацию (обновляем количество со склада)
					existingVariation.Sizes = variationItem.Sizes
					existingVariation.Colors = variationItem.Colors
					existingVariation.Price = variationItem.Price
					existingVariation.OriginalPrice = variationItem.OriginalPrice
					existingVariation.Discount = variationItem.Discount
					existingVariation.StockQuantity = variationItem.StockQuantity // Обновляем количество со склада
					existingVariation.IsAvailable = variationItem.StockQuantity > 0
					existingVariation.SKU = variationItem.SKU
					existingVariation.Barcode = variationItem.Barcode

					// Обновляем фото только если они предоставлены
					if len(variationItem.ImageURLs) > 0 {
						existingVariation.ImageURLs = variationItem.ImageURLs
					}
					if len(imageURLsByColor) > 0 {
						existingVariation.ImageURLsByColor = imageURLsByColor
					}

					if err := tx.Save(&existingVariation).Error; err != nil {
						log.Printf("❌ Failed to update variation: %v", err)
						response.Errors = append(response.Errors, fmt.Sprintf("Product %d: failed to update variation - %v", i+1, err))
					}
				}
			}

			response.Updated++
			response.ProductIDs = append(response.ProductIDs, existingProduct.ID.String())
			log.Printf("✅ Updated product: %s (ID: %s)", existingProduct.Name, existingProduct.ID)
		}
	}

	if response.Failed > 0 && response.Created == 0 && response.Updated == 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "All products failed to process",
			"details": response,
		})
		return
	}

	// Коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to commit transaction",
			"details": err.Error(),
		})
		return
	}

	response.Success = true
	c.JSON(http.StatusOK, response)
}

// SyncSales обрабатывает синхронизацию продаж из POS программы
// POST /api/v1/pos/sales/sync
func (pc *PosController) SyncSales(c *gin.Context) {
	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	userModel := user.(models.User)

	// Получаем shop для этого пользователя
	var shop models.Shop
	if err := database.DB.Where("owner_id = ?", userModel.ID).First(&shop).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Shop not found for this user",
		})
		return
	}

	var req SaleSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	response := SaleSyncResponse{
		TotalSales: len(req.Sales),
		Errors:     []string{},
		UpdatedStock: []StockUpdate{},
	}

	// Начинаем транзакцию
	tx := database.DB.Begin()

	for i, sale := range req.Sales {
		// Получаем вариацию
		var variation models.ProductVariation
		if err := tx.Preload("Product").Where("id = ?", sale.VariationID).First(&variation).Error; err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Sale %d: variation not found", i+1))
			continue
		}

		// Проверяем, что товар принадлежит этому магазину
		if variation.Product.ShopID == nil || *variation.Product.ShopID != shop.ID {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Sale %d: variation does not belong to this shop", i+1))
			continue
		}

		// Сохраняем старое количество
		oldQuantity := variation.StockQuantity

		// Уменьшаем количество
		if variation.StockQuantity < sale.Quantity {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Sale %d: insufficient stock (available: %d, requested: %d)", i+1, variation.StockQuantity, sale.Quantity))
			continue
		}

		variation.StockQuantity -= sale.Quantity
		variation.IsAvailable = variation.StockQuantity > 0

		// Обновляем вариацию
		if err := tx.Save(&variation).Error; err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Sale %d: failed to update stock - %v", i+1, err))
			continue
		}

		// Создаем запись о продаже (если нужна история)
		saleDate := sale.SaleDate
		if saleDate.IsZero() {
			saleDate = time.Now()
		}

		// TODO: Можно создать таблицу Sales для истории продаж
		// saleRecord := models.Sale{
		// 	VariationID: variation.ID,
		// 	ShopID:      shop.ID,
		// 	Quantity:    sale.Quantity,
		// 	Price:       sale.Price,
		// 	Size:        sale.Size,
		// 	Color:       sale.Color,
		// 	SaleDate:    saleDate,
		// }
		// tx.Create(&saleRecord)

		response.Processed++
		response.UpdatedStock = append(response.UpdatedStock, StockUpdate{
			VariationID: variation.ID,
			OldQuantity: oldQuantity,
			NewQuantity: variation.StockQuantity,
		})

		log.Printf("✅ Sale processed: Variation %s, Quantity: %d, New Stock: %d", variation.ID, sale.Quantity, variation.StockQuantity)
	}

	if response.Processed == 0 && response.Failed > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "All sales failed to process",
			"details": response,
		})
		return
	}

	// Коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to commit transaction",
			"details": err.Error(),
		})
		return
	}

	response.Success = true
	c.JSON(http.StatusOK, response)
}

// GetStockInfo возвращает информацию о количестве товаров
// GET /api/v1/pos/products/stock
func (pc *PosController) GetStockInfo(c *gin.Context) {
	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	userModel := user.(models.User)

	// Получаем shop для этого пользователя
	var shop models.Shop
	if err := database.DB.Where("owner_id = ?", userModel.ID).First(&shop).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Shop not found for this user",
		})
		return
	}

	// Получаем все вариации товаров этого магазина
	var variations []models.ProductVariation
	query := database.DB.Preload("Product").
		Joins("JOIN products ON products.id = product_variations.product_id").
		Where("products.shop_id = ?", shop.ID)

	// Фильтрация по штрих-коду
	if barcode := c.Query("barcode"); barcode != "" {
		query = query.Where("product_variations.barcode = ?", barcode)
	}

	// Фильтрация по SKU
	if sku := c.Query("sku"); sku != "" {
		query = query.Where("product_variations.sku = ?", sku)
	}

	// Фильтрация по наличию на складе
	if inStock := c.Query("in_stock"); inStock == "true" {
		query = query.Where("product_variations.stock_quantity > 0")
	}

	if err := query.Find(&variations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch stock info",
			"details": err.Error(),
		})
		return
	}

	// Формируем ответ
	stockInfo := make([]StockInfo, len(variations))
	for i, v := range variations {
		stockInfo[i] = StockInfo{
			VariationID:  v.ID,
			ProductID:    v.ProductID,
			ProductName:  v.Product.Name,
			SKU:          v.SKU,
			Barcode:      v.Barcode,
			Sizes:        v.Sizes,
			Colors:       v.Colors,
			StockQuantity: v.StockQuantity,
			IsAvailable:   v.IsAvailable,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stockInfo,
		"count":   len(stockInfo),
	})
}

// UpdateStock обновляет количество конкретной вариации
// PUT /api/v1/pos/products/:variationId/stock
func (pc *PosController) UpdateStock(c *gin.Context) {
	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	userModel := user.(models.User)

	// Получаем shop для этого пользователя
	var shop models.Shop
	if err := database.DB.Where("owner_id = ?", userModel.ID).First(&shop).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Shop not found for this user",
		})
		return
	}

	variationID := c.Param("variationId")
	variationUUID, err := uuid.Parse(variationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid variation ID",
		})
		return
	}

	var req struct {
		StockQuantity int `json:"stockQuantity" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Получаем вариацию
	var variation models.ProductVariation
	if err := database.DB.Preload("Product").Where("id = ?", variationUUID).First(&variation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Variation not found",
		})
		return
	}

	// Проверяем, что товар принадлежит этому магазину
	if variation.Product.ShopID == nil || *variation.Product.ShopID != shop.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Variation does not belong to this shop",
		})
		return
	}

	// Обновляем количество
	oldQuantity := variation.StockQuantity
	variation.StockQuantity = req.StockQuantity
	variation.IsAvailable = req.StockQuantity > 0

	if err := database.DB.Save(&variation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update stock",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": StockUpdate{
			VariationID: variation.ID,
			OldQuantity: oldQuantity,
			NewQuantity: variation.StockQuantity,
		},
	})
}

