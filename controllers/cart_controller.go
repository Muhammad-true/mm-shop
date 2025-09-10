package controllers

import (
	"net/http"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/middleware"
	"github.com/mm-api/mm-api/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CartController обрабатывает запросы корзины
type CartController struct{}

// GetCart возвращает содержимое корзины пользователя
func (cc *CartController) GetCart(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	var cartItems []models.CartItem
	if err := database.DB.Preload("Product").Where("user_id = ?", user.ID).Find(&cartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch cart items",
		})
		return
	}

	// Формируем ответ
	items := make([]models.CartItemResponse, len(cartItems))
	var totalPrice float64
	var totalItems int

	for i, item := range cartItems {
		items[i] = item.ToResponse()
		totalPrice += items[i].Subtotal
		totalItems += item.Quantity
	}

	cart := models.CartResponse{
		Items:      items,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
	}

	c.JSON(http.StatusOK, gin.H{
		"cart": cart,
	})
}

// AddToCart добавляет товар в корзину
func (cc *CartController) AddToCart(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	var req models.CartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Проверяем, существует ли продукт
	var product models.Product
	if err := database.DB.First(&product, "id = ? AND in_stock = ?", req.ProductID, true).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found or out of stock",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
		}
		return
	}

	// Проверяем наличие товара в нужном количестве
	// Ищем подходящую вариацию по размеру и цвету
	var availableStock int
	for _, variation := range product.Variations {
		sizeMatch := false
		colorMatch := false

		for _, size := range variation.Sizes {
			if size == req.Size {
				sizeMatch = true
				break
			}
		}

		for _, color := range variation.Colors {
			if color == req.Color {
				colorMatch = true
				break
			}
		}

		if sizeMatch && colorMatch {
			availableStock = variation.StockQuantity
			break
		}
	}

	if availableStock < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Insufficient stock",
		})
		return
	}

	// Проверяем, есть ли уже такой товар в корзине
	var existingItem models.CartItem
	err := database.DB.Where("user_id = ? AND product_id = ? AND size = ? AND color = ?",
		user.ID, req.ProductID, req.Size, req.Color).First(&existingItem).Error

	if err == nil {
		// Товар уже есть в корзине, увеличиваем количество
		existingItem.Quantity += req.Quantity

		// Проверяем, не превышает ли общее количество доступный запас
		if availableStock < existingItem.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient stock for total quantity",
			})
			return
		}

		if err := database.DB.Save(&existingItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update cart item",
			})
			return
		}

		// Загружаем продукт для ответа
		database.DB.Preload("Product").First(&existingItem, existingItem.ID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Cart item updated successfully",
			"item":    existingItem.ToResponse(),
		})
		return
	}

	// Создаем новый элемент корзины
	cartItem := models.CartItem{
		UserID:    user.ID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Size:      req.Size,
		Color:     req.Color,
	}

	if err := database.DB.Create(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add item to cart",
		})
		return
	}

	// Загружаем продукт для ответа
	database.DB.Preload("Product").First(&cartItem, cartItem.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Item added to cart successfully",
		"item":    cartItem.ToResponse(),
	})
}

// UpdateCartItem обновляет элемент корзины
func (cc *CartController) UpdateCartItem(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	id := c.Param("id")
	itemID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cart item ID",
		})
		return
	}

	var req models.CartItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	var cartItem models.CartItem
	if err := database.DB.Preload("Product").Where("id = ? AND user_id = ?", itemID, user.ID).First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Cart item not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
		}
		return
	}

	// Проверяем доступность товара через вариации
	var availableStock int
	for _, variation := range cartItem.Product.Variations {
		sizeMatch := false
		colorMatch := false

		for _, size := range variation.Sizes {
			if size == req.Size {
				sizeMatch = true
				break
			}
		}

		for _, color := range variation.Colors {
			if color == req.Color {
				colorMatch = true
				break
			}
		}

		if sizeMatch && colorMatch {
			availableStock = variation.StockQuantity
			break
		}
	}

	if availableStock < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Insufficient stock",
		})
		return
	}

	// Обновляем элемент корзины
	cartItem.Quantity = req.Quantity
	cartItem.Size = req.Size
	cartItem.Color = req.Color

	if err := database.DB.Save(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update cart item",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart item updated successfully",
		"item":    cartItem.ToResponse(),
	})
}

// RemoveFromCart удаляет элемент из корзины
func (cc *CartController) RemoveFromCart(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	id := c.Param("id")
	itemID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cart item ID",
		})
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", itemID, user.ID).Delete(&models.CartItem{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove item from cart",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cart item not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item removed from cart successfully",
	})
}

// ClearCart очищает всю корзину пользователя
func (cc *CartController) ClearCart(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	if err := database.DB.Where("user_id = ?", user.ID).Delete(&models.CartItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart cleared successfully",
	})
}
