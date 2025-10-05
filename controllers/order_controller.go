package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"gorm.io/gorm"
)

type OrderController struct{}

// CreateOrder - создать заказ (для авторизованного пользователя)
func (oc *OrderController) CreateOrder(c *gin.Context) {
	// Достаём текущего пользователя
	userIDValue, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не найден",
		))
		return
	}

	type createItem struct {
		ProductID   string  `json:"product_id" binding:"required,uuid"`
		VariationID string  `json:"variation_id" binding:"required,uuid"`
		Quantity    int     `json:"quantity" binding:"required,gt=0"`
		Price       float64 `json:"price" binding:"required,gt=0"`
		SKU         string  `json:"sku"`
		Name        string  `json:"name"`
		ImageURL    string  `json:"image_url"`
	}

	var req struct {
		RecipientName  string       `json:"recipient_name" binding:"required"`
		Phone          string       `json:"phone" binding:"required"`
		ShippingAddr   string       `json:"shipping_addr" binding:"required"`
		DesiredAt      *time.Time   `json:"desired_at"`
		DesiredDate    string       `json:"desired_date"` // YYYY-MM-DD
		DesiredTime    string       `json:"desired_time"` // HH:mm
		PaymentMethod  string       `json:"payment_method" binding:"required,oneof=cash card"`
		ShippingMethod string       `json:"shipping_method"`
		ItemsSubtotal  float64      `json:"items_subtotal"`
		DeliveryFee    float64      `json:"delivery_fee"`
		TotalAmount    float64      `json:"total_amount"`
		Currency       string       `json:"currency"`
		Notes          string       `json:"notes"`
		Items          []createItem `json:"items" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверные данные запроса",
		))
		return
	}

	// Если пришли desired_date + desired_time — склеиваем в desired_at (UTC)
	if req.DesiredAt == nil && req.DesiredDate != "" && req.DesiredTime != "" {
		if t, err := time.Parse("2006-01-02 15:04", req.DesiredDate+" "+req.DesiredTime); err == nil {
			tt := t.UTC()
			req.DesiredAt = &tt
		}
	}

	// Пересчёт на сервере
	var subtotal float64
	for _, it := range req.Items {
		subtotal += it.Price * float64(it.Quantity)
	}
	delivery := 10.0
	if subtotal >= 200.0 {
		delivery = 0.0
	}
	total := subtotal + delivery
	currency := req.Currency
	if currency == "" {
		currency = "TJS"
	}

	currentUserID := userIDValue.(uuid.UUID)
	shippingMethod := req.ShippingMethod
	if strings.TrimSpace(shippingMethod) == "" {
		shippingMethod = "courier"
	}

	var createdOrder models.Order
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Создаём адрес (гостевой, простой, из одной строки shipping_addr)
		addr := models.Address{
			UserID:    currentUserID,
			Street:    req.ShippingAddr,
			City:      "",
			State:     "",
			ZipCode:   "",
			Country:   "",
			Label:     "Другое",
			IsDefault: false,
		}
		if err := tx.Create(&addr).Error; err != nil {
			return err
		}
		order := models.Order{
			UserID:         currentUserID,
			Status:         models.OrderStatusPending,
			ItemsSubtotal:  subtotal,
			DeliveryFee:    delivery,
			TotalAmount:    total,
			Currency:       currency,
			AddressID:      &addr.ID,
			ShippingAddr:   req.ShippingAddr,
			PaymentMethod:  req.PaymentMethod,
			ShippingMethod: shippingMethod,
			PaymentStatus:  "pending",
			RecipientName:  req.RecipientName,
			Phone:          req.Phone,
			DesiredAt:      req.DesiredAt,
			Notes:          req.Notes,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Создаём позиции заказа
		for _, it := range req.Items {
			pid, err := uuid.Parse(it.ProductID)
			if err != nil {
				return err
			}
			vid, err := uuid.Parse(it.VariationID)
			if err != nil {
				return err
			}
			item := models.OrderItem{
				OrderID:     order.ID,
				ProductID:   pid,
				VariationID: vid,
				Quantity:    it.Quantity,
				Price:       it.Price,
				SKU:         it.SKU,
				Name:        it.Name,
				ImageURL:    it.ImageURL,
				Total:       it.Price * float64(it.Quantity),
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		createdOrder = order
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при создании заказа",
		))
		return
	}

	// Возвращаем заказ с позициями
	if err := database.DB.Preload("OrderItems").Preload("OrderItems.Product").Preload("OrderItems.Variation").First(&createdOrder, "id = ?", createdOrder.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении созданного заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    createdOrder.ToResponse(),
		Message: "Заказ создан успешно",
	})
}

// GetMyOrders - список заказов текущего пользователя
func (oc *OrderController) GetMyOrders(c *gin.Context) {
	userIDValue, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не найден",
		))
		return
	}
	currentUserID := userIDValue.(uuid.UUID)

	status := c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Order{}).Where("user_id = ?", currentUserID)
	if status != "" {
		if status == "active" {
			query = query.Where("status NOT IN ?", []models.OrderStatus{models.OrderStatusCompleted, models.OrderStatusCancelled})
		} else if status == "completed" {
			query = query.Where("status IN ?", []models.OrderStatus{models.OrderStatusCompleted, models.OrderStatusCancelled})
		} else {
			query = query.Where("status = ?", status)
		}
	}

	var total int64
	query.Count(&total)

	var orders []models.Order
	if err := query.Preload("OrderItems").Preload("OrderItems.Product").Preload("OrderItems.Variation").Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении заказов",
		))
		return
	}

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

// GetMyOrder - детали заказа текущего пользователя
func (oc *OrderController) GetMyOrder(c *gin.Context) {
	userIDValue, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не найден",
		))
		return
	}
	currentUserID := userIDValue.(uuid.UUID)

	orderID := c.Param("id")
	var order models.Order
	err := database.DB.Preload("OrderItems").Preload("OrderItems.Product").Preload("OrderItems.Variation").First(&order, "id = ? AND user_id = ?", orderID, currentUserID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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

// CancelMyOrder - отменить заказ текущего пользователя
func (oc *OrderController) CancelMyOrder(c *gin.Context) {
	userIDValue, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не найден",
		))
		return
	}
	currentUserID := userIDValue.(uuid.UUID)

	orderID := c.Param("id")
	var order models.Order
	if err := database.DB.First(&order, "id = ? AND user_id = ?", orderID, currentUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	// Обновляем статус
	now := time.Now()
	order.Status = models.OrderStatusCancelled
	order.CancelledAt = &now

	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при отмене заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    order.ToResponse(),
		Message: "Заказ отменен",
	})
}

// CreateGuestOrder - создать заказ для гостя (без авторизации)
func (oc *OrderController) CreateGuestOrder(c *gin.Context) {
	type createItem struct {
		ProductID   string  `json:"product_id" binding:"required,uuid"`
		VariationID string  `json:"variation_id" binding:"required,uuid"`
		Quantity    int     `json:"quantity" binding:"required,gt=0"`
		Price       float64 `json:"price" binding:"required,gt=0"`
		SKU         string  `json:"sku"`
		Name        string  `json:"name"`
		ImageURL    string  `json:"image_url"`
	}

	var req struct {
		GuestName      string       `json:"guest_name" binding:"required"`
		GuestPhone     string       `json:"guest_phone" binding:"required"`
		ShippingAddr   string       `json:"shipping_addr" binding:"required"`
		DesiredAt      *time.Time   `json:"desired_at"`
		DesiredDate    string       `json:"desired_date"`
		DesiredTime    string       `json:"desired_time"`
		PaymentMethod  string       `json:"payment_method" binding:"required,oneof=cash card"`
		ShippingMethod string       `json:"shipping_method"`
		ItemsSubtotal  float64      `json:"items_subtotal"`
		DeliveryFee    float64      `json:"delivery_fee"`
		TotalAmount    float64      `json:"total_amount"`
		Currency       string       `json:"currency"`
		Notes          string       `json:"notes"`
		Items          []createItem `json:"items" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверные данные запроса",
		))
		return
	}

	if req.DesiredAt == nil && req.DesiredDate != "" && req.DesiredTime != "" {
		if t, err := time.Parse("2006-01-02 15:04", req.DesiredDate+" "+req.DesiredTime); err == nil {
			tt := t.UTC()
			req.DesiredAt = &tt
		}
	}

	// Пересчёт на сервере
	var subtotal float64
	for _, it := range req.Items {
		subtotal += it.Price * float64(it.Quantity)
	}
	delivery := 10.0
	if subtotal >= 200.0 {
		delivery = 0.0
	}
	total := subtotal + delivery
	currency := req.Currency
	if currency == "" {
		currency = "TJS"
	}

	// Создаем или находим пользователя по номеру телефона
	var user models.User
	err := database.DB.Where("phone = ?", req.GuestPhone).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Создаем нового пользователя автоматически
		user = models.User{
			Name:     req.GuestName,
			Email:    "guest_" + uuid.New().String() + "@temp.local", // Временный email
			Phone:    req.GuestPhone,
			Password: "auto_password_" + uuid.New().String(), // Автоматический пароль
			IsGuest:  true,
			IsActive: true,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Ошибка при создании пользователя",
			))
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при поиске пользователя",
		))
		return
	} else {
		// Пользователь уже существует - обновляем имя если нужно
		if user.Name != req.GuestName {
			user.Name = req.GuestName
			if err := database.DB.Save(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
					models.ErrInternalError,
					"Ошибка при обновлении пользователя",
				))
				return
			}
		}
	}

	var createdOrder models.Order
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Создаём адрес для гостя из строки shipping_addr
		addr := models.Address{
			UserID:    user.ID,
			Street:    req.ShippingAddr,
			City:      "",
			State:     "",
			ZipCode:   "",
			Country:   "",
			Label:     "Другое",
			IsDefault: false,
		}
		if err := tx.Create(&addr).Error; err != nil {
			return err
		}
		shippingMethod := req.ShippingMethod
		if strings.TrimSpace(shippingMethod) == "" {
			shippingMethod = "courier"
		}

		order := models.Order{
			UserID:         user.ID,
			Status:         models.OrderStatusPending,
			ItemsSubtotal:  subtotal,
			DeliveryFee:    delivery,
			TotalAmount:    total,
			Currency:       currency,
			AddressID:      &addr.ID,
			ShippingAddr:   req.ShippingAddr,
			PaymentMethod:  req.PaymentMethod,
			ShippingMethod: shippingMethod,
			PaymentStatus:  "pending",
			RecipientName:  req.GuestName,
			Phone:          req.GuestPhone,
			DesiredAt:      req.DesiredAt,
			Notes:          req.Notes,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Создаём позиции заказа
		for _, it := range req.Items {
			pid, err := uuid.Parse(it.ProductID)
			if err != nil {
				return err
			}
			vid, err := uuid.Parse(it.VariationID)
			if err != nil {
				return err
			}
			item := models.OrderItem{
				OrderID:     order.ID,
				ProductID:   pid,
				VariationID: vid,
				Quantity:    it.Quantity,
				Price:       it.Price,
				SKU:         it.SKU,
				Name:        it.Name,
				ImageURL:    it.ImageURL,
				Total:       it.Price * float64(it.Quantity),
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		createdOrder = order
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при создании заказа",
		))
		return
	}

	// Возвращаем заказ с позициями
	if err := database.DB.Preload("OrderItems").Preload("OrderItems.Product").Preload("OrderItems.Variation").First(&createdOrder, "id = ?", createdOrder.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении созданного заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    createdOrder.ToResponse(),
		Message: "Гостевой заказ создан успешно",
	})
}

// GetGuestOrder - получить заказ гостя по email/телефону
func (oc *OrderController) GetGuestOrder(c *gin.Context) {
	email := c.Query("email")
	phone := c.Query("phone")

	if email == "" && phone == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Необходимо указать email или телефон",
		))
		return
	}

	var orders []models.Order
	query := database.DB.Preload("OrderItems").Preload("OrderItems.Product").Preload("OrderItems.Variation").Joins("JOIN users ON orders.user_id = users.id").Where("users.is_guest = true")

	if email != "" {
		query = query.Where("users.email = ?", email)
	}
	if phone != "" {
		query = query.Where("users.phone = ?", phone)
	}

	if err := query.Order("orders.created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении заказов",
		))
		return
	}

	var orderResponses []models.OrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, order.ToResponse())
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    orderResponses,
		Message: "Заказы гостя получены успешно",
	})
}

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
		Preload("OrderItems.Product").Preload("OrderItems.Variation").
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
		Preload("OrderItems.Product").Preload("OrderItems.Variation").
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
		Preload("OrderItems.Product").Preload("OrderItems.Variation").
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
		Preload("OrderItems.Product").Preload("OrderItems.Variation")

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
		Preload("OrderItems.Product").Preload("OrderItems.Variation").
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
