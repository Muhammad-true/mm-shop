package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserController struct{}

// GetUsers - получить всех пользователей (только для админов)
func (uc *UserController) GetUsers(c *gin.Context) {
	var users []models.User
	var total int64

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Подсчет общего количества
	database.DB.Model(&models.User{}).Count(&total)

	// Получение пользователей с пагинацией
	result := database.DB.Preload("Addresses").Preload("Role").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении пользователей",
		))
		return
	}

	// Преобразование в ответы
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data: gin.H{
			"users": userResponses,
			"pagination": models.PaginationInfo{
				Page:       page,
				Limit:      limit,
				Total:      int(total),
				TotalPages: int((total + int64(limit) - 1) / int64(limit)),
			},
		},
		Message: "Пользователи получены успешно",
	})
}

// GetUser - получить пользователя по ID (только для админов)
func (uc *UserController) GetUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	result := database.DB.Preload("Addresses").Preload("Role").
		First(&user, "id = ?", userID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Пользователь не найден",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении пользователя",
		))
		return
	}

	userResponse := user.ToResponse()

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    userResponse,
		Message: "Пользователь получен успешно",
	})
}

// UpdateUser - обновить пользователя (только для админов)
func (uc *UserController) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var updateRequest models.UserUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверные данные запроса",
		))
		return
	}

	var user models.User
	result := database.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Пользователь не найден",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при поиске пользователя",
		))
		return
	}

	// Обновление полей
	if updateRequest.Name != nil && *updateRequest.Name != "" {
		user.Name = *updateRequest.Name
	}
	if updateRequest.Phone != nil && *updateRequest.Phone != "" {
		user.Phone = *updateRequest.Phone
	}
	if updateRequest.DateOfBirth != nil {
		user.DateOfBirth = updateRequest.DateOfBirth
	}
	if updateRequest.Gender != nil && *updateRequest.Gender != "" {
		user.Gender = *updateRequest.Gender
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при обновлении пользователя",
		))
		return
	}

	userResponse := user.ToResponse()

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    userResponse,
		Message: "Пользователь обновлен успешно",
	})
}

// DeleteUser - удалить пользователя (только для админов)
func (uc *UserController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	result := database.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Пользователь не найден",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при поиске пользователя",
		))
		return
	}

	// Мягкое удаление - просто деактивируем
	user.IsActive = false
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при удалении пользователя",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Message: "Пользователь удален успешно",
	})
}

// CreateUser - создать нового пользователя (только для админов)
func (uc *UserController) CreateUser(c *gin.Context) {
	var req struct {
		Name     string  `json:"name" binding:"required"`
		Email    string  `json:"email" binding:"required,email"`
		Password string  `json:"password" binding:"required,min=6"`
		Phone    string  `json:"phone"`
		RoleID   *string `json:"roleId"`
		IsActive bool    `json:"isActive"`
		Shop     *struct {
			Name        string  `json:"name"`
			INN         string  `json:"inn"`
			Description string  `json:"description"`
			Address     string  `json:"address"`
			Email       string  `json:"email"`
			Phone       string  `json:"phone"`
			CityID      *string `json:"cityId"` // ID города
		} `json:"shop"` // Данные магазина (если роль shop_owner)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверные данные запроса",
			err.Error(),
		))
		return
	}

	// Проверяем, существует ли пользователь с таким email
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
			models.ErrUserAlreadyExists,
			"Пользователь с таким email уже существует",
		))
		return
	}

	// Создаем нового пользователя
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		IsActive: req.IsActive,
	}

	// Хешируем пароль
	if err := user.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при хешировании пароля",
		))
		return
	}

	// Устанавливаем роль, если указана
	var roleName string
	if req.RoleID != nil {
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
				models.ErrValidationError,
				"Неверный ID роли",
			))
			return
		}
		user.RoleID = &roleID
		
		// Получаем имя роли для проверки shop_owner
		var role models.Role
		if err := database.DB.Where("id = ?", roleID).First(&role).Error; err == nil {
			roleName = role.Name
		}
	}

	// Сохраняем в базу данных
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при создании пользователя",
		))
		return
	}

	// Если роль shop_owner и переданы данные магазина, создаем магазин
	if roleName == "shop_owner" && req.Shop != nil {
		shopName := req.Shop.Name
		if shopName == "" {
			shopName = user.Name // Используем имя пользователя, если название магазина не указано
		}
		
		var cityID *uuid.UUID
		if req.Shop.CityID != nil {
			if parsedCityID, err := uuid.Parse(*req.Shop.CityID); err == nil {
				cityID = &parsedCityID
			}
		}

		shop := models.Shop{
			ID:          user.ID, // Используем тот же ID для обратной совместимости
			Name:        shopName,
			INN:         req.Shop.INN,
			Description: req.Shop.Description,
			Address:     req.Shop.Address,
			Email:       req.Shop.Email,
			Phone:       req.Shop.Phone,
			IsActive:    user.IsActive,
			OwnerID:     user.ID,
			CityID:      cityID, // ID города
		}
		
		// Если email/phone не указаны для магазина, используем из пользователя
		if shop.Email == "" {
			shop.Email = user.Email
		}
		if shop.Phone == "" {
			shop.Phone = user.Phone
		}
		
		if err := database.DB.Create(&shop).Error; err != nil {
			log.Printf("⚠️ Ошибка создания магазина для пользователя %s: %v", user.ID, err)
			// Не прерываем создание пользователя, но логируем ошибку
		} else {
			log.Printf("✅ Магазин создан для пользователя %s: %s", user.ID, shop.ID)
		}
	}

	// Загружаем связанные данные для ответа
	database.DB.Preload("Role").Preload("Addresses").First(&user, user.ID)

	c.JSON(http.StatusCreated, models.StandardResponse{
		Success: true,
		Data:    user.ToResponse(),
		Message: "Пользователь создан успешно",
	})
}

// GetShopCustomers - получить клиентов владельца магазина (только обычные пользователи)
func (uc *UserController) GetShopCustomers(c *gin.Context) {
	var users []models.User
	var total int64

	// Получаем текущего пользователя
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не найден",
		))
		return
	}

	user := currentUser.(models.User)

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Показываем только обычных пользователей (клиентов)
	query := database.DB.Model(&models.User{}).Joins("JOIN roles ON users.role_id = roles.id").Where("roles.name = ?", "user")

	// Если владелец магазина - показываем только активных клиентов
	if user.Role != nil && user.Role.Name == "shop_owner" {
		query = query.Where("users.is_active = ?", true)
	}

	// Подсчет общего количества
	query.Count(&total)

	// Получение пользователей с пагинацией
	result := query.Preload("Addresses").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении клиентов",
		))
		return
	}

	// Преобразование в ответы
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data: gin.H{
			"customers": userResponses,
			"pagination": models.PaginationInfo{
				Page:       page,
				Limit:      limit,
				Total:      int(total),
				TotalPages: int((total + int64(limit) - 1) / int64(limit)),
			},
		},
		Message: "Клиенты получены успешно",
	})
}
