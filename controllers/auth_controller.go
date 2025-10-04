package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/mm-api/mm-api/utils"
	"gorm.io/gorm"
)

// AuthController обрабатывает запросы аутентификации
type AuthController struct{}

// Register регистрирует нового пользователя
func (ac *AuthController) Register(c *gin.Context) {
	var req models.UserRegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	// Проверяем, существует ли пользователь с таким email
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
			models.ErrUserAlreadyExists,
			"User with this email already exists",
		))
		return
	}

	// Получаем роль "user" по умолчанию
	var defaultRole models.Role
	if err := database.DB.Where("name = ?", "user").First(&defaultRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to find default user role",
		))
		return
	}

	// Создаем нового пользователя
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		IsActive: true,
		RoleID:   &defaultRole.ID, // Присваиваем роль пользователя по умолчанию
	}

	// Хешируем пароль
	if err := user.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to hash password",
		))
		return
	}

	// Сохраняем в базу данных
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to create user",
		))
		return
	}

	// Получаем роль пользователя
	roleName := "user" // По умолчанию
	if user.Role != nil {
		roleName = user.Role.Name
	}

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user.ID, user.Email, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to generate token",
		))
		return
	}

	// Создаем настройки по умолчанию для пользователя
	settings := models.UserSettings{
		UserID:               user.ID,
		Language:             "ru",
		Theme:                "system",
		NotificationsEnabled: true,
		EmailNotifications:   true,
		PushNotifications:    true,
	}
	database.DB.Create(&settings)

	authResponse := models.AuthResponse{
		User:         user.ToResponse(),
		Token:        token,
		RefreshToken: token, // TODO: Реализовать отдельные refresh токены
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(
		authResponse,
		"User registered successfully",
	))
}

// Login выполняет вход пользователя
func (ac *AuthController) Login(c *gin.Context) {
	var req models.UserLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	// Ищем пользователя по email
	var user models.User
	if err := database.DB.Preload("Addresses").Preload("Role").Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
				models.ErrInvalidCredentials,
				"Invalid email or password",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
		}
		return
	}

	// Проверяем пароль
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrInvalidCredentials,
			"Invalid email or password",
		))
		return
	}

	// Получаем роль пользователя
	roleName := "user" // По умолчанию
	if user.Role != nil {
		roleName = user.Role.Name
	}

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user.ID, user.Email, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to generate token",
		))
		return
	}

	authResponse := models.AuthResponse{
		User:         user.ToResponse(),
		Token:        token,
		RefreshToken: token, // TODO: Реализовать отдельные refresh токены
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		authResponse,
		"Login successful",
	))
}

// CreateGuestToken создает токен для гостевого пользователя или входит в существующий аккаунт
func (ac *AuthController) CreateGuestToken(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	// Получаем роль "user" для гостей
	var userRole models.Role
	if err := database.DB.Where("name = ?", "user").First(&userRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to find user role",
		))
		return
	}

	// Ищем пользователя по номеру телефона
	var user models.User
	err := database.DB.Where("phone = ?", req.Phone).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Создаем нового пользователя автоматически
		user = models.User{
			Name:     req.Name,
			Email:    "guest_" + uuid.New().String() + "@temp.local", // Временный email
			Phone:    req.Phone,
			Password: "auto_password_" + uuid.New().String(), // Автоматический пароль
			IsGuest:  true,
			IsActive: true,
			RoleID:   &userRole.ID,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to create user",
			))
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Database error",
		))
		return
	} else {
		// Пользователь уже существует - обновляем имя если нужно
		if user.Name != req.Name {
			user.Name = req.Name
			if err := database.DB.Save(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
					models.ErrInternalError,
					"Failed to update user",
				))
				return
			}
		}
	}

	// Загружаем роль для ответа
	database.DB.Preload("Role").First(&user, user.ID)

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user.ID, user.Email, "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to generate token",
		))
		return
	}

	authResponse := models.AuthResponse{
		User:         user.ToResponse(),
		Token:        token,
		RefreshToken: token, // TODO: Реализовать отдельные refresh токены
	}

	// Определяем сообщение в зависимости от того, новый это пользователь или существующий
	message := "User logged in successfully"
	if err == gorm.ErrRecordNotFound {
		message = "User created and logged in successfully"
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		authResponse,
		message,
	))
}

// Profile возвращает профиль текущего пользователя
func (ac *AuthController) Profile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	userModel := user.(models.User)

	// Загружаем связанные данные
	database.DB.Preload("Addresses").First(&userModel, userModel.ID)

	c.JSON(http.StatusOK, models.SuccessResponse(userModel.ToResponse()))
}

// UpdateProfile обновляет профиль пользователя
func (ac *AuthController) UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	userModel := user.(models.User)

	// Обновляем только переданные поля
	if req.Name != nil {
		userModel.Name = *req.Name
	}

	if req.Phone != nil {
		userModel.Phone = *req.Phone
	}

	if req.DateOfBirth != nil {
		userModel.DateOfBirth = req.DateOfBirth
	}

	if req.Gender != nil {
		userModel.Gender = *req.Gender
	}

	if err := database.DB.Save(&userModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to update profile",
		))
		return
	}

	// Загружаем обновленные данные
	database.DB.Preload("Addresses").First(&userModel, userModel.ID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		userModel.ToResponse(),
		"Profile updated successfully",
	))
}

// RefreshToken обновляет JWT токен
func (ac *AuthController) RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	newToken, err := utils.RefreshJWT(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthInvalid,
			"Failed to refresh token",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{
		"token":        newToken,
		"refreshToken": newToken, // TODO: Генерировать отдельный refresh токен
	}))
}
