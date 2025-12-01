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

// SettingsController обрабатывает запросы настроек пользователя
type SettingsController struct{}

// GetSettings возвращает настройки пользователя
func (sc *SettingsController) GetSettings(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	var settings models.UserSettings
	err := database.DB.Preload("City").Where("user_id = ?", user.ID).First(&settings).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Создаем настройки по умолчанию, если их нет
			settings = models.UserSettings{
				UserID:               user.ID,
				Language:             "ru",
				Theme:                "system",
				NotificationsEnabled: true,
				EmailNotifications:   true,
				PushNotifications:    true,
			}

			if err := database.DB.Create(&settings).Error; err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
					models.ErrInternalError,
					"Failed to create default settings",
				))
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to fetch settings",
			))
			return
		}
	}

	// Загружаем City если настройки были только что созданы
	if settings.City == nil && settings.SelectedCityID != nil {
		database.DB.Preload("City").First(&settings, settings.ID)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(settings.ToResponse()))
}

// UpdateSettings обновляет настройки пользователя
func (sc *SettingsController) UpdateSettings(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	var req models.SettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	var settings models.UserSettings
	err := database.DB.Preload("City").Where("user_id = ?", user.ID).First(&settings).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Создаем новые настройки
			settings = models.UserSettings{
				UserID: user.ID,
			}
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
			return
		}
	}

	// Обновляем только переданные поля
	if req.Language != "" {
		settings.Language = req.Language
	}

	if req.Theme != "" {
		settings.Theme = req.Theme
	}

	if req.NotificationsEnabled != nil {
		settings.NotificationsEnabled = *req.NotificationsEnabled
	}

	if req.EmailNotifications != nil {
		settings.EmailNotifications = *req.EmailNotifications
	}

	if req.PushNotifications != nil {
		settings.PushNotifications = *req.PushNotifications
	}

	// Обновляем выбранный город
	if req.SelectedCityID != nil {
		if cityUUID, err := uuid.Parse(*req.SelectedCityID); err == nil {
			settings.SelectedCityID = &cityUUID
		}
	}

	// Сохраняем настройки
	if settings.ID == uuid.Nil {
		// Создаем новые настройки
		if err := database.DB.Create(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to create settings",
			))
			return
		}
	} else {
		// Обновляем существующие
		if err := database.DB.Save(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to update settings",
			))
			return
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		settings.ToResponse(),
		"Settings updated successfully",
	))
}

// ResetSettings сбрасывает настройки к значениям по умолчанию
func (sc *SettingsController) ResetSettings(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	// Устанавливаем настройки по умолчанию
	updates := map[string]interface{}{
		"language":              "ru",
		"theme":                 "system",
		"notifications_enabled": true,
		"email_notifications":   true,
		"push_notifications":    true,
	}

	// Обновляем или создаем настройки
	var settings models.UserSettings
	err := database.DB.Where("user_id = ?", user.ID).First(&settings).Error

	if err == gorm.ErrRecordNotFound {
		// Создаем новые настройки с умолчаниями
		settings = models.UserSettings{
			UserID:               user.ID,
			Language:             "ru",
			Theme:                "system",
			NotificationsEnabled: true,
			EmailNotifications:   true,
			PushNotifications:    true,
		}

		if err := database.DB.Create(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to reset settings",
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
		// Обновляем существующие настройки
		if err := database.DB.Model(&settings).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to reset settings",
			))
			return
		}

		// Перезагружаем обновленные настройки
		database.DB.First(&settings, settings.ID)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		settings.ToResponse(),
		"Settings reset to defaults",
	))
}
