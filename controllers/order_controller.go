package controllers

import (
	"log"
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
		VariationID string  `json:"variation_id" binding:"required,uuid"`
		Quantity    int     `json:"quantity" binding:"required,gt=0"`
		Price       float64 `json:"price" binding:"required,gt=0"`
		Size        string  `json:"size"`
		Color       string  `json:"color"`
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
			vid, err := uuid.Parse(it.VariationID)
			if err != nil {
				return err
			}
			item := models.OrderItem{
				OrderID:     order.ID,
				VariationID: vid,
				Quantity:    it.Quantity,
				Price:       it.Price,
				Size:        it.Size,
				Color:       it.Color,
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
		log.Printf("❌ Ошибка при создании заказа: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при создании заказа: "+err.Error(),
		))
		return
	}

	// Возвращаем заказ с позициями
	log.Printf("🔍 Загружаем заказ с ID: %s", createdOrder.ID)

	// Сначала попробуем без Preload для диагностики
	if err := database.DB.First(&createdOrder, "id = ?", createdOrder.ID).Error; err != nil {
		log.Printf("❌ Ошибка при получении созданного заказа: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении созданного заказа: "+err.Error(),
		))
		return
	}

	// Теперь загружаем OrderItems отдельно
	var orderItems []models.OrderItem
	if err := database.DB.Where("order_id = ?", createdOrder.ID).Find(&orderItems).Error; err != nil {
		log.Printf("❌ Ошибка при получении OrderItems: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении OrderItems: "+err.Error(),
		))
		return
	}
	createdOrder.OrderItems = orderItems

	log.Printf("✅ Заказ успешно загружен с %d позициями", len(orderItems))

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
	if err := query.Preload("OrderItems").Preload("OrderItems.Variation").Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
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
	err := database.DB.Preload("OrderItems").Preload("OrderItems.Variation").First(&order, "id = ? AND user_id = ?", orderID, currentUserID).Error
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

// GetActiveOrder - получить все активные заказы пользователя для отслеживания
func (oc *OrderController) GetActiveOrder(c *gin.Context) {
	userIDValue, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не найден",
		))
		return
	}
	currentUserID := userIDValue.(uuid.UUID)

	// Получаем все активные заказы (не завершённые и не отменённые)
	var orders []models.Order
	err := database.DB.
		Preload("OrderItems").
		Preload("OrderItems.Variation").
		Where("user_id = ? AND status NOT IN ?", currentUserID, []models.OrderStatus{
			models.OrderStatusCompleted,
			models.OrderStatusCancelled,
		}).
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении заказов",
		))
		return
	}

	// Если нет активных заказов
	if len(orders) == 0 {
		c.JSON(http.StatusOK, models.StandardResponse{
			Success: true,
			Data:    []interface{}{},
			Message: "Нет активных заказов",
		})
		return
	}

	// Формируем ответ для каждого заказа с информацией о статусе
	var activeOrders []gin.H
	for _, order := range orders {
		activeOrders = append(activeOrders, gin.H{
			"order": order.ToResponse(),
			"tracking": gin.H{
				"current_status": order.Status,
				"status_history": []gin.H{
					{
						"status":      "pending",
						"label":       "Ожидает подтверждения",
						"completed":   order.Status != models.OrderStatusPending,
						"is_current":  order.Status == models.OrderStatusPending,
						"icon":        "clock",
						"description": "Ваш заказ принят и ожидает подтверждения",
					},
					{
						"status":      "confirmed",
						"label":       "Подтвержден",
						"completed":   order.Status != models.OrderStatusPending && order.Status != models.OrderStatusConfirmed,
						"is_current":  order.Status == models.OrderStatusConfirmed,
						"icon":        "check-circle",
						"description": "Заказ подтвержден и готовится к отправке",
					},
					{
						"status":      "preparing",
						"label":       "Готовится",
						"completed":   order.Status == models.OrderStatusInDelivery || order.Status == models.OrderStatusDelivered,
						"is_current":  order.Status == models.OrderStatusPreparing,
						"icon":        "package",
						"description": "Ваш заказ готовится к отправке",
					},
					{
						"status":      "inDelivery",
						"label":       "В доставке",
						"completed":   order.Status == models.OrderStatusDelivered,
						"is_current":  order.Status == models.OrderStatusInDelivery,
						"icon":        "truck",
						"description": "Курьер в пути к вам",
					},
					{
						"status":      "delivered",
						"label":       "Доставлен",
						"completed":   false,
						"is_current":  order.Status == models.OrderStatusDelivered,
						"icon":        "check-double",
						"description": "Заказ доставлен",
					},
				},
				"estimated_delivery": order.DesiredAt,
				"recipient_name":     order.RecipientName,
				"phone":              order.Phone,
				"shipping_address":   order.ShippingAddr,
			},
		})
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    activeOrders,
		Message: "Активные заказы получены",
	})
}

// CreateGuestOrder - создать заказ для гостя (без авторизации)
func (oc *OrderController) CreateGuestOrder(c *gin.Context) {
	type createItem struct {
		VariationID string  `json:"variation_id" binding:"required,uuid"`
		Quantity    int     `json:"quantity" binding:"required,gt=0"`
		Price       float64 `json:"price" binding:"required,gt=0"`
		Size        string  `json:"size"`
		Color       string  `json:"color"`
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
			vid, err := uuid.Parse(it.VariationID)
			if err != nil {
				return err
			}
			item := models.OrderItem{
				OrderID:     order.ID,
				VariationID: vid,
				Quantity:    it.Quantity,
				Price:       it.Price,
				Size:        it.Size,
				Color:       it.Color,
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
		log.Printf("❌ Ошибка при создании заказа: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при создании заказа: "+err.Error(),
		))
		return
	}

	// Возвращаем заказ с позициями
	log.Printf("🔍 Загружаем заказ с ID: %s", createdOrder.ID)

	// Сначала попробуем без Preload для диагностики
	if err := database.DB.First(&createdOrder, "id = ?", createdOrder.ID).Error; err != nil {
		log.Printf("❌ Ошибка при получении созданного заказа: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении созданного заказа: "+err.Error(),
		))
		return
	}

	// Теперь загружаем OrderItems отдельно
	var orderItems []models.OrderItem
	if err := database.DB.Where("order_id = ?", createdOrder.ID).Find(&orderItems).Error; err != nil {
		log.Printf("❌ Ошибка при получении OrderItems: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении OrderItems: "+err.Error(),
		))
		return
	}
	createdOrder.OrderItems = orderItems

	log.Printf("✅ Заказ успешно загружен с %d позициями", len(orderItems))

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
	query := database.DB.Preload("OrderItems").Preload("OrderItems.Variation").Joins("JOIN users ON orders.user_id = users.id").Where("users.is_guest = true")

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

// GetAdminOrders - получить все заказы для админ панели с расширенной информацией
func (oc *OrderController) GetAdminOrders(c *gin.Context) {
	var orders []models.Order
	var total int64

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Параметры фильтрации
	status := c.Query("status")            // pending, confirmed, preparing, inDelivery, delivered, completed, cancelled
	search := c.Query("search")            // Поиск по номеру заказа, имени, телефону
	orderNumber := c.Query("order_number") // Поиск по номеру заказа
	phone := c.Query("phone")              // Поиск по телефону
	dateFrom := c.Query("date_from")       // Фильтр по дате от (YYYY-MM-DD)
	dateTo := c.Query("date_to")           // Фильтр по дате до (YYYY-MM-DD)

	// Начинаем строить запрос
	query := database.DB.Model(&models.Order{})

	// Фильтр по статусу
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Поиск по номеру заказа (первые 8 символов UUID)
	if orderNumber != "" {
		query = query.Where("CAST(id AS TEXT) LIKE ?", orderNumber+"%")
	}

	// Поиск по телефону
	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}

	// Общий поиск (по имени получателя, телефону или номеру заказа)
	if search != "" {
		query = query.Where(
			"recipient_name ILIKE ? OR phone LIKE ? OR CAST(id AS TEXT) LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%",
		)
	}

	// Фильтр по дате создания
	if dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			// Добавляем 1 день чтобы включить весь день dateTo
			t = t.Add(24 * time.Hour)
			query = query.Where("created_at < ?", t)
		}
	}

	// Подсчет общего количества с учетом фильтров
	query.Count(&total)

	// Получение заказов с пагинацией
	result := query.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Variation").
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

	// Преобразование в расширенные ответы для админ панели
	var orderResponses []models.AdminOrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, order.ToAdminResponse())
	}

	// Статистика по статусам
	var stats []struct {
		Status string
		Count  int64
	}
	database.DB.Model(&models.Order{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&stats)

	statusStats := make(map[string]int64)
	for _, stat := range stats {
		statusStats[stat.Status] = stat.Count
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
			"stats": statusStats,
		},
		Message: "Заказы получены успешно",
	})
}

// GetOrders - получить все заказы (только для админов) - старый метод для обратной совместимости
func (oc *OrderController) GetOrders(c *gin.Context) {
	// Перенаправляем на новый метод
	oc.GetAdminOrders(c)
}

// GetOrder - получить заказ по ID (только для админов)
func (oc *OrderController) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	var order models.Order
	result := database.DB.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Variation").
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
		Status string `json:"status" binding:"required,oneof=pending confirmed preparing inDelivery delivered completed cancelled"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверные данные запроса",
		))
		return
	}

	var order models.Order
	result := database.DB.Preload("User").Preload("OrderItems").Preload("OrderItems.Variation").First(&order, "id = ?", orderID)
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

	now := time.Now()
	order.Status = models.OrderStatus(updateRequest.Status)

	// Устанавливаем время подтверждения или отмены
	if updateRequest.Status == "confirmed" && order.ConfirmedAt == nil {
		order.ConfirmedAt = &now
	}
	if updateRequest.Status == "cancelled" && order.CancelledAt == nil {
		order.CancelledAt = &now
	}

	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при обновлении заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    order.ToAdminResponse(),
		Message: "Статус заказа обновлен успешно",
	})
}

// ConfirmOrder - подтвердить заказ (быстрый метод для админ панели)
func (oc *OrderController) ConfirmOrder(c *gin.Context) {
	orderID := c.Param("id")

	var order models.Order
	result := database.DB.Preload("User").Preload("OrderItems").Preload("OrderItems.Variation").First(&order, "id = ?", orderID)
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

	now := time.Now()
	order.Status = models.OrderStatusConfirmed
	order.ConfirmedAt = &now

	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при подтверждении заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    order.ToAdminResponse(),
		Message: "Заказ подтвержден",
	})
}

// RejectOrder - отклонить заказ (быстрый метод для админ панели)
func (oc *OrderController) RejectOrder(c *gin.Context) {
	orderID := c.Param("id")

	var order models.Order
	result := database.DB.Preload("User").Preload("OrderItems").Preload("OrderItems.Variation").First(&order, "id = ?", orderID)
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

	now := time.Now()
	order.Status = models.OrderStatusCancelled
	order.CancelledAt = &now

	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при отклонении заказа",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    order.ToAdminResponse(),
		Message: "Заказ отклонен",
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
		Preload("OrderItems.Variation").
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
		Preload("OrderItems.Variation")

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
		Preload("OrderItems.Variation").
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
