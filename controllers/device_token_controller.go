package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"gorm.io/gorm"
)

// DeviceTokenController обрабатывает запросы токенов устройств
type DeviceTokenController struct{}

// RegisterDeviceToken регистрирует токен устройства для push-уведомлений
func (dtc *DeviceTokenController) RegisterDeviceToken(c *gin.Context) {
	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not authenticated",
		))
		return
	}

	currentUser := user.(models.User)

	var req models.DeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	// Проверяем платформу
	validPlatforms := map[string]bool{
		"ios":     true,
		"android": true,
		"web":     true,
	}
	if !validPlatforms[req.Platform] {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid platform. Must be: ios, android, or web",
		))
		return
	}

	// Проверяем, существует ли уже токен для этого устройства
	var existingToken models.DeviceToken
	err := database.DB.Where("token = ?", req.Token).First(&existingToken).Error

	if err == nil {
		// Токен уже существует - обновляем его
		existingToken.UserID = currentUser.ID
		existingToken.Platform = req.Platform
		existingToken.DeviceID = req.DeviceID
		existingToken.IsActive = true
		existingToken.LastUsed = time.Now()

		if err := database.DB.Save(&existingToken).Error; err != nil {
			log.Printf("❌ Ошибка обновления токена устройства: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to update device token",
			))
			return
		}

		log.Printf("✅ Токен устройства обновлен для пользователя %s", currentUser.ID)
		c.JSON(http.StatusOK, models.SuccessResponse(
			existingToken,
			"Device token updated successfully",
		))
		return
	}

	if err != gorm.ErrRecordNotFound {
		log.Printf("❌ Ошибка проверки токена устройства: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Database error",
		))
		return
	}

	// Создаем новый токен
	deviceToken := models.DeviceToken{
		UserID:   currentUser.ID,
		Token:    req.Token,
		Platform: req.Platform,
		DeviceID: req.DeviceID,
		IsActive: true,
		LastUsed: time.Now(),
	}

	if err := database.DB.Create(&deviceToken).Error; err != nil {
		log.Printf("❌ Ошибка создания токена устройства: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to register device token",
		))
		return
	}

	log.Printf("✅ Токен устройства зарегистрирован для пользователя %s", currentUser.ID)
	c.JSON(http.StatusCreated, models.SuccessResponse(
		deviceToken,
		"Device token registered successfully",
	))
}

// UnregisterDeviceToken удаляет токен устройства
func (dtc *DeviceTokenController) UnregisterDeviceToken(c *gin.Context) {
	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not authenticated",
		))
		return
	}

	currentUser := user.(models.User)

	tokenParam := c.Param("token")
	if tokenParam == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Token parameter is required",
		))
		return
	}

	// Удаляем токен или деактивируем его
	result := database.DB.Where("token = ? AND user_id = ?", tokenParam, currentUser.ID).Delete(&models.DeviceToken{})
	if result.Error != nil {
		log.Printf("❌ Ошибка удаления токена устройства: %v", result.Error)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to unregister device token",
		))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
			models.ErrNotFound,
			"Device token not found",
		))
		return
	}

	log.Printf("✅ Токен устройства удален для пользователя %s", currentUser.ID)
	c.JSON(http.StatusOK, models.SuccessResponse(
		nil,
		"Device token unregistered successfully",
	))
}

// GetUserDeviceTokens возвращает все токены устройства пользователя
func (dtc *DeviceTokenController) GetUserDeviceTokens(c *gin.Context) {
	// Получаем текущего пользователя
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not authenticated",
		))
		return
	}

	currentUser := user.(models.User)

	var tokens []models.DeviceToken
	if err := database.DB.Where("user_id = ? AND is_active = ?", currentUser.ID, true).Find(&tokens).Error; err != nil {
		log.Printf("❌ Ошибка получения токенов устройств: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to get device tokens",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(tokens))
}

