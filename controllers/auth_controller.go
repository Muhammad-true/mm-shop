package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/mm-api/mm-api/services"
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

    // Проверяем, существует ли пользователь с таким телефоном
    var existingUser models.User
    if err := database.DB.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
        // Если это гость — конвертируем в обычного пользователя
        if existingUser.IsGuest {
            // Проверяем email, если передан: не занят ли другим
            if req.Email != "" {
                var emailOwner models.User
                if err := database.DB.Where("email = ?", req.Email).First(&emailOwner).Error; err == nil && emailOwner.ID != existingUser.ID {
                    c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
                        models.ErrUserAlreadyExists,
                        "User with this email already exists",
                    ))
                    return
                }
                existingUser.Email = req.Email
            }

            // Обновляем имя, пароль, статус гостя
            existingUser.Name = req.Name
            if err := existingUser.HashPassword(req.Password); err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
                    models.ErrInternalError,
                    "Failed to hash password",
                ))
                return
            }
            existingUser.IsGuest = false
            existingUser.IsActive = true

            if err := database.DB.Save(&existingUser).Error; err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
                    models.ErrInternalError,
                    "Failed to update user",
                ))
                return
            }

            // Убедимся, что у пользователя роль user
            var userRole models.Role
            if err := database.DB.Where("name = ?", "user").First(&userRole).Error; err == nil {
                existingUser.RoleID = &userRole.ID
                _ = database.DB.Save(&existingUser).Error
                database.DB.Preload("Role").First(&existingUser, existingUser.ID)
            }

            // Генерируем JWT токен
            roleName := "user"
            if existingUser.Role != nil {
                roleName = existingUser.Role.Name
            }
            token, err := utils.GenerateJWT(existingUser.ID, existingUser.Email, roleName)
            if err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
                    models.ErrInternalError,
                    "Failed to generate token",
                ))
                return
            }

            authResponse := models.AuthResponse{
                User:         existingUser.ToResponse(),
                Token:        token,
                RefreshToken: token,
            }

            c.JSON(http.StatusOK, models.SuccessResponse(
                authResponse,
                "Guest account upgraded and registered successfully",
            ))
            return
        }

        // Если это не гость — конфликт
        c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
            models.ErrUserAlreadyExists,
            "User with this phone already exists",
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
        Email:    req.Email, // может быть пустым
		Phone:    req.Phone,
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

    // Если email не указан, генерируем временный
    if user.Email == "" {
        user.Email = "user_" + uuid.New().String() + "@temp.local"
    } else {
        // Проверяем уникальность email
        var emailOwner models.User
        if err := database.DB.Where("email = ?", user.Email).First(&emailOwner).Error; err == nil {
            c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
                models.ErrUserAlreadyExists,
                "User with this email already exists",
            ))
            return
        }
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

	// Ищем пользователя по телефону
	var user models.User
	if err := database.DB.Preload("Addresses").Preload("Role").Where("phone = ? AND is_active = ?", req.Phone, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
				models.ErrInvalidCredentials,
				"Invalid phone or password",
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
			"Invalid phone or password",
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

	// Отправляем push-уведомления о непрочитанных уведомлениях при входе
	go sendUnreadNotificationsPush(user.ID)

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

// UploadAvatar загружает аватар пользователя
func (ac *AuthController) UploadAvatar(c *gin.Context) {
	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"Пользователь не авторизован",
		))
		return
	}

	userModel := user.(models.User)

	// Получаем файл из формы
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Файл не предоставлен",
			err.Error(),
		))
		return
	}
	defer file.Close()

	// Используем UploadController для загрузки
	uploadController := &UploadController{}
	
	// Создаем временный файл для чтения содержимого
	tempFile, err := os.CreateTemp("", "avatar-*.tmp")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка создания временного файла",
		))
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Копируем содержимое файла
	if _, err := io.Copy(tempFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка сохранения файла",
		))
		return
	}

	// Перемещаем указатель в начало файла
	tempFile.Seek(0, 0)

	// Загружаем изображение
	folder := "avatars"
	uploadDir := fmt.Sprintf("images/%s", folder)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка создания папки",
		))
		return
	}

	// Генерируем уникальное имя файла
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	ext = strings.ToLower(ext)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Сохраняем файл через метод UploadController
	contentType := header.Header.Get("Content-Type")
	finalFilename, _, err := uploadController.CompressAndSaveImage(tempFile, filePath, ext, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка сохранения изображения",
			err.Error(),
		))
		return
	}

	// Обновляем filename если формат изменился
	if finalFilename != filename {
		filename = finalFilename
	}

	// Формируем URL
	avatarURL := uploadController.GetImageURL(filename, folder)

	// Удаляем старый аватар, если он есть
	if userModel.Avatar != "" {
		oldFilename := filepath.Base(userModel.Avatar)
		oldPath := filepath.Join(uploadDir, oldFilename)
		if _, err := os.Stat(oldPath); err == nil {
			os.Remove(oldPath)
		}
	}

	// Обновляем аватар пользователя
	userModel.Avatar = avatarURL
	if err := database.DB.Save(&userModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка обновления аватара",
		))
		return
	}

	// Загружаем обновленные данные
	database.DB.Preload("Addresses").Preload("Role").First(&userModel, userModel.ID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		userModel.ToResponse(),
		"Аватар успешно загружен",
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

// ForgotPassword инициирует восстановление пароля по телефону
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req models.UserForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	// Заглушка: в проде отправляем SMS код или ссылку
	// Здесь подтверждаем, что телефон существует (но не раскрываем факт отсутствия)
	var user models.User
	_ = database.DB.Where("phone = ?", req.Phone).First(&user)

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{
		"status":  "pending",
		"message": "If this phone exists, a reset code was sent",
	}))
}

// DeleteAccount удаляет аккаунт текущего пользователя
func (ac *AuthController) DeleteAccount(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	userModel := user.(models.User)

	// Проверяем, что пользователь активен
	if !userModel.IsActive {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Account is already deleted",
		))
		return
	}

	// Мягкое удаление - деактивируем аккаунт
	userModel.IsActive = false
	if err := database.DB.Save(&userModel).Error; err != nil {
		log.Printf("❌ Ошибка удаления аккаунта пользователя %s: %v", userModel.ID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to delete account",
		))
		return
	}

	log.Printf("✅ Аккаунт пользователя %s успешно удален (деактивирован)", userModel.ID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		nil,
		"Account deleted successfully",
	))
}

// sendUnreadNotificationsPush отправляет push-уведомления о непрочитанных уведомлениях при входе
func sendUnreadNotificationsPush(userID uuid.UUID) {
	// Получаем все непрочитанные уведомления пользователя
	var unreadNotifications []models.Notification
	if err := database.DB.Where("user_id = ? AND is_read = ?", userID, false).
		Order("timestamp DESC").
		Limit(5). // Отправляем только последние 5 непрочитанных
		Find(&unreadNotifications).Error; err != nil {
		log.Printf("⚠️ Ошибка получения непрочитанных уведомлений для пользователя %s: %v", userID, err)
		return
	}

	if len(unreadNotifications) == 0 {
		return // Нет непрочитанных уведомлений
	}

	// Получаем все активные токены устройств пользователя
	var deviceTokens []models.DeviceToken
	if err := database.DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&deviceTokens).Error; err != nil {
		log.Printf("⚠️ Ошибка получения токенов устройств для пользователя %s: %v", userID, err)
		return
	}

	if len(deviceTokens) == 0 {
		return // Нет активных токенов
	}

	// Группируем токены по платформам
	var fcmTokens []string
	for _, token := range deviceTokens {
		if token.Platform == "android" || token.Platform == "web" || token.Platform == "ios" {
			fcmTokens = append(fcmTokens, token.Token)
		}
	}

	if len(fcmTokens) == 0 {
		return
	}

	// Отправляем push-уведомление о непрочитанных уведомлениях
	fcmService := services.GetFCMService()
	if fcmService != nil {
		title := "У вас есть непрочитанные уведомления"
		body := fmt.Sprintf("У вас %d непрочитанных уведомлений", len(unreadNotifications))
		actionURL := "/admin#dashboard" // Переход на дашборд с уведомлениями

		if err := fcmService.SendPushNotificationToMultiple(fcmTokens, title, body, actionURL); err != nil {
			log.Printf("❌ Ошибка отправки push-уведомления о непрочитанных уведомлениях: %v", err)
		} else {
			log.Printf("✅ Push-уведомление о непрочитанных уведомлениях отправлено пользователю %s", userID)
		}
	}
}
