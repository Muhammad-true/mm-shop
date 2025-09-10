package controllers

import (
	"net/http"
	"strconv"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderController struct{}

// GetOrders - получить все заказы (только для админов)
func (oc *OrderController) GetOrders(c *gin.Context) {
	var orders []models.Order
	var total int64

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Подсчет общего количества
	database.DB.Model(&models.Order{}).Count(&total)

	// Получение заказов с пагинацией
	result := database.DB.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении заказов",
		))
		return
	}

	// Преобразование в ответы
	var orderResponses []models.OrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, order.ToResponse())
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data: gin.H{
			"orders": orderResponses,
			"pagination": models.PaginationInfo{
				Page:       page,
				Limit:      limit,
				Total:      int(total),
				TotalPages: int((total + int64(limit) - 1) / int64(limit)),
			},
		},
		Message: "Заказы получены успешно",
	})
}

// GetOrder - получить заказ по ID (только для админов)
func (oc *OrderController) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	var order models.Order
	result := database.DB.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		First(&order, "id = ?", orderID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Заказ не найден",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    order.ToResponse(),
		Message: "Заказ получен успешно",
	})
}

// UpdateOrderStatus - обновить статус заказа (только для админов)
func (oc *OrderController) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var updateRequest struct {
		Status string `json:"status" binding:"required,oneof=pending processing shipped delivered cancelled"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверные данные запроса",
		))
		return
	}

	var order models.Order
	result := database.DB.First(&order, "id = ?", orderID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Заказ не найден",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при поиске заказа",
		))
		return
	}

	order.Status = models.OrderStatus(updateRequest.Status)
	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при обновлении заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    order.ToResponse(),
		Message: "Статус заказа обновлен успешно",
	})
}

// DeleteOrder - удалить заказ (только для админов)
func (oc *OrderController) DeleteOrder(c *gin.Context) {
	orderID := c.Param("id")

	var order models.Order
	result := database.DB.First(&order, "id = ?", orderID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Заказ не найден",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при поиске заказа",
		))
		return
	}

	// Удаляем заказ и связанные элементы
	if err := database.DB.Delete(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при удалении заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Message: "Заказ удален успешно",
	})
}

// GetShopOrders - получить заказы клиентов владельца магазина
func (oc *OrderController) GetShopOrders(c *gin.Context) {
	var orders []models.Order
	var total int64

	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не найден",
		))
		return
	}

	currentUser := user.(models.User)

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Если админ - показываем все заказы, если владелец магазина - только заказы клиентов
	query := database.DB.Model(&models.Order{})
	if currentUser.Role != nil && currentUser.Role.Name == "shop_owner" {
		// Для владельца магазина показываем только заказы обычных пользователей
		query = query.Joins("JOIN users ON orders.user_id = users.id").
			Joins("JOIN roles ON users.role_id = roles.id").
			Where("roles.name = ?", "user")
	}

	// Подсчет общего количества
	query.Count(&total)

	// Получение заказов с пагинацией
	result := query.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении заказов",
		))
		return
	}

	// Преобразование в ответы
	var orderResponses []models.OrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, order.ToResponse())
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data: gin.H{
			"orders": orderResponses,
			"pagination": models.PaginationInfo{
				Page:       page,
				Limit:      limit,
				Total:      int(total),
				TotalPages: int((total + int64(limit) - 1) / int64(limit)),
			},
		},
		Message: "Заказы получены успешно",
	})
}

// GetShopOrder - получить заказ по ID для владельца магазина
func (oc *OrderController) GetShopOrder(c *gin.Context) {
	orderID := c.Param("id")

	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не найден",
		))
		return
	}

	currentUser := user.(models.User)

	var order models.Order
	query := database.DB.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Product")

	// Если владелец магазина - проверяем, что заказ от обычного пользователя
	if currentUser.Role != nil && currentUser.Role.Name == "shop_owner" {
		query = query.Joins("JOIN users ON orders.user_id = users.id").
			Joins("JOIN roles ON users.role_id = roles.id").
			Where("roles.name = ? AND orders.id = ?", "user", orderID)
	} else {
		query = query.Where("orders.id = ?", orderID)
	}

	result := query.First(&order)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Заказ не найден",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    order.ToResponse(),
		Message: "Заказ получен успешно",
	})
}

// GetCustomerOrders - получить заказы конкретного клиента для владельца магазина
func (oc *OrderController) GetCustomerOrders(c *gin.Context) {
	customerID := c.Param("id")

	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не найден",
		))
		return
	}

	currentUser := user.(models.User)

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var orders []models.Order
	var total int64

	query := database.DB.Model(&models.Order{}).Where("user_id = ?", customerID)

	// Если владелец магазина - проверяем, что клиент обычный пользователь
	if currentUser.Role != nil && currentUser.Role.Name == "shop_owner" {
		query = query.Joins("JOIN users ON orders.user_id = users.id").
			Joins("JOIN roles ON users.role_id = roles.id").
			Where("roles.name = ?", "user")
	}

	// Подсчет общего количества
	query.Count(&total)

	// Получение заказов с пагинацией
	result := query.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&orders)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении заказов клиента",
		))
		return
	}

	// Преобразование в ответы
	var orderResponses []models.OrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, order.ToResponse())
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data: gin.H{
			"orders": orderResponses,
			"pagination": models.PaginationInfo{
				Page:       page,
				Limit:      limit,
				Total:      int(total),
				TotalPages: int((total + int64(limit) - 1) / int64(limit)),
			},
		},
		Message: "Заказы клиента получены успешно",
	})
}
