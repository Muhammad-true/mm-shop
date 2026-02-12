package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/mm-api/mm-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShopCustomerController обрабатывает запросы для работы с клиентами магазинов и бонусами
type ShopCustomerController struct{}

// RegisterOrUpdateCustomer регистрирует или обновляет клиента магазина
// Публичный эндпоинт с проверкой shop_id
// POST /api/v1/shop/customers/register
func (scc *ShopCustomerController) RegisterOrUpdateCustomer(c *gin.Context) {
	var req struct {
		ShopID      string `json:"shopId" binding:"required"`      // ID магазина
		Phone       string `json:"phone" binding:"required"`        // Номер телефона клиента
		QRCode      string `json:"qrCode"`                          // QR код (обязателен при первом создании)
		BonusAmount int    `json:"bonusAmount" binding:"required"`   // Новое количество бонусов
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверные данные запроса",
			err.Error(),
		))
		return
	}

	// Парсим shop_id
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверный формат shop_id",
		))
		return
	}

	// Проверяем, что магазин существует
	var shop models.Shop
	if err := database.DB.Where("id = ? AND is_active = ?", shopID, true).First(&shop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Магазин не найден",
			))
			return
		}
		log.Printf("❌ [RegisterOrUpdateCustomer] Ошибка при поиске магазина: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка базы данных",
		))
		return
	}

	// Нормализуем номер телефона
	normalizedPhone := utils.NormalizePhone(req.Phone)

	// Ищем существующего клиента
	var shopClient models.ShopClient
	isNew := false
	err = database.DB.Where("shop_id = ? AND phone = ?", shopID, normalizedPhone).First(&shopClient).Error

	if err == gorm.ErrRecordNotFound {
		// Новый клиент - создаем запись
		isNew = true
		if req.QRCode == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
				models.ErrValidationError,
				"QR код обязателен при первой регистрации клиента",
			))
			return
		}

		// Проверяем уникальность QR кода в этом магазине
		var existingQR models.ShopClient
		if err := database.DB.Where("shop_id = ? AND qr_code = ?", shopID, req.QRCode).First(&existingQR).Error; err == nil {
			c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
				models.ErrValidationError,
				"QR код уже используется другим клиентом в этом магазине",
			))
			return
		}

		shopClient = models.ShopClient{
			ShopID:      shopID,
			Phone:       normalizedPhone,
			QRCode:      req.QRCode,
			BonusAmount: req.BonusAmount,
		}

		// Если бонусы начисляются впервые, устанавливаем дату
		if req.BonusAmount > 0 {
			now := time.Now()
			shopClient.FirstBonusDate = &now
		}

		if err := database.DB.Create(&shopClient).Error; err != nil {
			log.Printf("❌ [RegisterOrUpdateCustomer] Ошибка создания клиента: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Ошибка при создании клиента",
			))
			return
		}

		log.Printf("✅ [RegisterOrUpdateCustomer] Создан новый клиент: shopID=%s, phone=%s, qrCode=%s", shopID, normalizedPhone, req.QRCode)
	} else if err != nil {
		log.Printf("❌ [RegisterOrUpdateCustomer] Ошибка при поиске клиента: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка базы данных",
		))
		return
	} else {
		// Существующий клиент - обновляем бонусы
		previousAmount := shopClient.BonusAmount
		changeAmount := req.BonusAmount - previousAmount

		// Если бонусы начисляются впервые, устанавливаем дату
		if shopClient.FirstBonusDate == nil && req.BonusAmount > 0 {
			now := time.Now()
			shopClient.FirstBonusDate = &now
		}

		shopClient.BonusAmount = req.BonusAmount
		if err := database.DB.Save(&shopClient).Error; err != nil {
			log.Printf("❌ [RegisterOrUpdateCustomer] Ошибка обновления клиента: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Ошибка при обновлении клиента",
			))
			return
		}

		// Создаем запись в истории бонусов, если количество изменилось
		if changeAmount != 0 {
			bonusHistory := models.BonusHistory{
				ShopClientID:  shopClient.ID,
				PreviousAmount: previousAmount,
				NewAmount:     req.BonusAmount,
				ChangeAmount:  changeAmount,
			}
			if err := database.DB.Create(&bonusHistory).Error; err != nil {
				log.Printf("⚠️ [RegisterOrUpdateCustomer] Ошибка создания истории бонусов: %v", err)
				// Не прерываем выполнение, только логируем
			}
		}

		log.Printf("✅ [RegisterOrUpdateCustomer] Обновлен клиент: shopID=%s, phone=%s, бонусы: %d -> %d", shopID, normalizedPhone, previousAmount, req.BonusAmount)
	}

	// Загружаем информацию о магазине для ответа
	database.DB.Preload("Shop").First(&shopClient, shopClient.ID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		shopClient.ToResponse(),
		func() string {
			if isNew {
				return "Клиент успешно зарегистрирован"
			}
			return "Бонусы клиента успешно обновлены"
		}(),
	))
}

// GetMyShops возвращает список магазинов клиента с информацией о бонусах
// GET /api/v1/shops/my
func (scc *ShopCustomerController) GetMyShops(c *gin.Context) {
	// Получаем текущего пользователя
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не авторизован",
		))
		return
	}

	user := currentUser.(models.User)

	// Нормализуем номер телефона пользователя
	normalizedPhone := utils.NormalizePhone(user.Phone)
	if normalizedPhone == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"У пользователя не указан номер телефона",
		))
		return
	}

	// Находим все связи клиента с магазинами по номеру телефона
	var shopClients []models.ShopClient
	if err := database.DB.Preload("Shop").
		Where("phone = ?", normalizedPhone).
		Order("created_at DESC").
		Find(&shopClients).Error; err != nil {
		log.Printf("❌ [GetMyShops] Ошибка получения магазинов клиента: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении магазинов",
		))
		return
	}

	// Обновляем user_id для всех найденных клиентов, если он еще не установлен
	for i := range shopClients {
		if shopClients[i].UserID == nil {
			shopClients[i].UserID = &user.ID
			database.DB.Save(&shopClients[i])
		}
	}

	// Преобразуем в ответы
	responses := make([]models.ShopClientResponse, len(shopClients))
	for i, sc := range shopClients {
		responses[i] = sc.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		responses,
		"Список магазинов получен",
	))
}

// GetShopBonusInfo возвращает информацию о бонусах клиента в конкретном магазине
// GET /api/v1/shops/:id/bonus
func (scc *ShopCustomerController) GetShopBonusInfo(c *gin.Context) {
	shopIDParam := c.Param("id")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверный формат shop_id",
		))
		return
	}

	// Получаем текущего пользователя
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не авторизован",
		))
		return
	}

	user := currentUser.(models.User)

	// Нормализуем номер телефона
	normalizedPhone := utils.NormalizePhone(user.Phone)
	if normalizedPhone == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"У пользователя не указан номер телефона",
		))
		return
	}

	// Находим клиента в этом магазине
	var shopClient models.ShopClient
	if err := database.DB.Preload("Shop").
		Where("shop_id = ? AND phone = ?", shopID, normalizedPhone).
		First(&shopClient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Клиент не найден в этом магазине",
			))
			return
		}
		log.Printf("❌ [GetShopBonusInfo] Ошибка получения информации: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка базы данных",
		))
		return
	}

	// Обновляем user_id, если он еще не установлен
	if shopClient.UserID == nil {
		shopClient.UserID = &user.ID
		database.DB.Save(&shopClient)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		shopClient.ToResponse(),
		"Информация о бонусах получена",
	))
}

// GetBonusHistory возвращает историю изменений бонусов клиента в магазине
// GET /api/v1/shops/:id/bonus/history
func (scc *ShopCustomerController) GetBonusHistory(c *gin.Context) {
	shopIDParam := c.Param("id")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверный формат shop_id",
		))
		return
	}

	// Получаем текущего пользователя
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не авторизован",
		))
		return
	}

	user := currentUser.(models.User)

	// Нормализуем номер телефона
	normalizedPhone := utils.NormalizePhone(user.Phone)
	if normalizedPhone == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"У пользователя не указан номер телефона",
		))
		return
	}

	// Находим клиента в этом магазине
	var shopClient models.ShopClient
	if err := database.DB.Where("shop_id = ? AND phone = ?", shopID, normalizedPhone).First(&shopClient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Клиент не найден в этом магазине",
			))
			return
		}
		log.Printf("❌ [GetBonusHistory] Ошибка получения клиента: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка базы данных",
		))
		return
	}

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Получаем историю бонусов
	var history []models.BonusHistory
	var total int64

	database.DB.Model(&models.BonusHistory{}).
		Where("shop_client_id = ?", shopClient.ID).
		Count(&total)

	if err := database.DB.Where("shop_client_id = ?", shopClient.ID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&history).Error; err != nil {
		log.Printf("❌ [GetBonusHistory] Ошибка получения истории: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении истории бонусов",
		))
		return
	}

	// Преобразуем в ответы
	responses := make([]models.BonusHistoryResponse, len(history))
	for i, h := range history {
		responses[i] = h.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		gin.H{
			"history": responses,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
		"История бонусов получена",
	))
}

