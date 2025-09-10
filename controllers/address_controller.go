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

// AddressController обрабатывает запросы адресов
type AddressController struct{}

// GetAddresses возвращает все адреса пользователя
func (ac *AddressController) GetAddresses(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	var addresses []models.Address
	if err := database.DB.Where("user_id = ?", user.ID).Order("is_default DESC, created_at DESC").Find(&addresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to fetch addresses",
		))
		return
	}

	// Преобразуем в response
	addressResponses := make([]models.AddressResponse, len(addresses))
	for i, address := range addresses {
		addressResponses[i] = address.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessResponse(addressResponses))
}

// CreateAddress создает новый адрес
func (ac *AddressController) CreateAddress(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	var req models.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	// Если это основной адрес, убираем флаг у других адресов
	if req.IsDefault {
		database.DB.Model(&models.Address{}).Where("user_id = ?", user.ID).Update("is_default", false)
	}

	address := models.Address{
		UserID:    user.ID,
		Street:    req.Street,
		City:      req.City,
		State:     req.State,
		ZipCode:   req.ZipCode,
		Country:   req.Country,
		Apartment: req.Apartment,
		Building:  req.Building,
		Entrance:  req.Entrance,
		Floor:     req.Floor,
		Intercom:  req.Intercom,
		IsDefault: req.IsDefault,
		Label:     req.Label,
	}

	if err := database.DB.Create(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to create address",
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(
		address.ToResponse(),
		"Address created successfully",
	))
}

// UpdateAddress обновляет адрес
func (ac *AddressController) UpdateAddress(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	id := c.Param("id")
	addressID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid address ID",
		))
		return
	}

	var req models.AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	var address models.Address
	if err := database.DB.Where("id = ? AND user_id = ?", addressID, user.ID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Address not found",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
		}
		return
	}

	// Если это основной адрес, убираем флаг у других адресов
	if req.IsDefault && !address.IsDefault {
		database.DB.Model(&models.Address{}).Where("user_id = ? AND id != ?", user.ID, addressID).Update("is_default", false)
	}

	// Обновляем поля
	address.Street = req.Street
	address.City = req.City
	address.State = req.State
	address.ZipCode = req.ZipCode
	address.Country = req.Country
	address.Apartment = req.Apartment
	address.Building = req.Building
	address.Entrance = req.Entrance
	address.Floor = req.Floor
	address.Intercom = req.Intercom
	address.IsDefault = req.IsDefault
	address.Label = req.Label

	if err := database.DB.Save(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to update address",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		address.ToResponse(),
		"Address updated successfully",
	))
}

// DeleteAddress удаляет адрес
func (ac *AddressController) DeleteAddress(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	id := c.Param("id")
	addressID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid address ID",
		))
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", addressID, user.ID).Delete(&models.Address{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to delete address",
		))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
			models.ErrNotFound,
			"Address not found",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		nil,
		"Address deleted successfully",
	))
}

// SetDefaultAddress устанавливает адрес как основной
func (ac *AddressController) SetDefaultAddress(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	id := c.Param("id")
	addressID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid address ID",
		))
		return
	}

	// Проверяем, принадлежит ли адрес пользователю
	var address models.Address
	if err := database.DB.Where("id = ? AND user_id = ?", addressID, user.ID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Address not found",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
		}
		return
	}

	// Выполняем операции в транзакции
	tx := database.DB.Begin()

	// Убираем флаг у всех адресов пользователя
	if err := tx.Model(&models.Address{}).Where("user_id = ?", user.ID).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to update default address",
		))
		return
	}

	// Устанавливаем флаг для выбранного адреса
	if err := tx.Model(&address).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to update default address",
		))
		return
	}

	tx.Commit()

	// Загружаем обновленный адрес
	database.DB.First(&address, addressID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		address.ToResponse(),
		"Default address updated successfully",
	))
}
