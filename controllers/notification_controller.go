package controllers

import (
	"net/http"
	"strconv"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/middleware"
	"github.com/mm-api/mm-api/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationController обрабатывает запросы уведомлений
type NotificationController struct{}

// GetNotifications возвращает уведомления пользователя с пагинацией
func (nc *NotificationController) GetNotifications(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	query := database.DB.Model(&models.Notification{}).Where("user_id = ?", user.ID)

	// Фильтрация по типу
	if notificationType := c.Query("type"); notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}

	// Фильтрация по прочитанности
	if isReadStr := c.Query("isRead"); isReadStr != "" {
		if isRead, err := strconv.ParseBool(isReadStr); err == nil {
			query = query.Where("is_read = ?", isRead)
		}
	}

	// Пагинация
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Получаем общее количество
	var total int64
	query.Count(&total)

	// Получаем уведомления
	var notifications []models.Notification
	if err := query.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to fetch notifications",
		))
		return
	}

	// Преобразуем в response
	notificationResponses := make([]models.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		notificationResponses[i] = notification.ToResponse()
	}

	// Вычисляем пагинацию
	totalPages := (int(total) + limit - 1) / limit
	pagination := models.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.PaginatedSuccessResponse(notificationResponses, pagination))
}

// MarkAsRead отмечает уведомление как прочитанное
func (nc *NotificationController) MarkAsRead(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	id := c.Param("id")
	notificationID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid notification ID",
		))
		return
	}

	var notification models.Notification
	if err := database.DB.Where("id = ? AND user_id = ?", notificationID, user.ID).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Notification not found",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
		}
		return
	}

	if !notification.IsRead {
		notification.IsRead = true
		if err := database.DB.Save(&notification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to mark notification as read",
			))
			return
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		notification.ToResponse(),
		"Notification marked as read",
	))
}

// MarkAllAsRead отмечает все уведомления как прочитанные
func (nc *NotificationController) MarkAllAsRead(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	if err := database.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", user.ID, false).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to mark notifications as read",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		nil,
		"All notifications marked as read",
	))
}

// DeleteNotification удаляет уведомление
func (nc *NotificationController) DeleteNotification(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	id := c.Param("id")
	notificationID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid notification ID",
		))
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", notificationID, user.ID).Delete(&models.Notification{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to delete notification",
		))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
			models.ErrNotFound,
			"Notification not found",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		nil,
		"Notification deleted successfully",
	))
}

// CreateNotification создает новое уведомление (для системы/админов)
func (nc *NotificationController) CreateNotification(c *gin.Context) {
	var req models.NotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	// Проверяем, существует ли пользователь
	var user models.User
	if err := database.DB.First(&user, "id = ?", req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"User not found",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
		}
		return
	}

	notification := models.Notification{
		UserID:    req.UserID,
		Title:     req.Title,
		Body:      req.Body,
		Type:      req.Type,
		ImageURL:  req.ImageURL,
		ActionURL: req.ActionURL,
	}

	// Сериализуем дополнительные данные в JSON если есть
	if req.Data != nil {
		// Здесь можно добавить JSON маршаллинг, пока оставляем пустым
		notification.Data = ""
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to create notification",
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(
		notification.ToResponse(),
		"Notification created successfully",
	))
}

// GetUnreadCount возвращает количество непрочитанных уведомлений
func (nc *NotificationController) GetUnreadCount(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	var count int64
	if err := database.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", user.ID, false).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to get unread count",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]int64{
		"unreadCount": count,
	}))
}
