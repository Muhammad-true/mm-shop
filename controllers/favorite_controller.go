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

// FavoriteController обрабатывает запросы избранного
type FavoriteController struct{}

// GetFavorites возвращает избранные товары пользователя
func (fc *FavoriteController) GetFavorites(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	var favorites []models.Favorite
	if err := database.DB.Preload("Product").Where("user_id = ?", user.ID).Find(&favorites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to fetch favorites",
		))
		return
	}

	// Преобразуем в response
	favoriteResponses := make([]models.FavoriteResponse, len(favorites))
	for i, favorite := range favorites {
		favoriteResponses[i] = favorite.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessResponse(favoriteResponses))
}

// AddToFavorites добавляет товар в избранное
func (fc *FavoriteController) AddToFavorites(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid product ID",
		))
		return
	}

	// Проверяем, существует ли продукт
	var product models.Product
	if err := database.DB.First(&product, "id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrProductNotFound,
				"Product not found",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
		}
		return
	}

	// Проверяем, не добавлен ли уже товар в избранное
	var existingFavorite models.Favorite
	err = database.DB.Where("user_id = ? AND product_id = ?", user.ID, productID).First(&existingFavorite).Error
	if err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Product already in favorites",
		))
		return
	}

	// Создаем новую запись в избранном
	favorite := models.Favorite{
		UserID:    user.ID,
		ProductID: productID,
	}

	if err := database.DB.Create(&favorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to add to favorites",
		))
		return
	}

	// Загружаем продукт для ответа
	database.DB.Preload("Product").First(&favorite, favorite.ID)

	c.JSON(http.StatusCreated, models.SuccessResponse(
		favorite.ToResponse(),
		"Product added to favorites",
	))
}

// RemoveFromFavorites удаляет товар из избранного
func (fc *FavoriteController) RemoveFromFavorites(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid product ID",
		))
		return
	}

	result := database.DB.Where("user_id = ? AND product_id = ?", user.ID, productID).Delete(&models.Favorite{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to remove from favorites",
		))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
			models.ErrNotFound,
			"Product not found in favorites",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		nil,
		"Product removed from favorites",
	))
}

// SyncFavorites синхронизирует избранное
func (fc *FavoriteController) SyncFavorites(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	var req models.FavoriteSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	// Получаем текущие избранные товары пользователя
	var currentFavorites []models.Favorite
	database.DB.Where("user_id = ?", user.ID).Find(&currentFavorites)

	// Создаем map для быстрого поиска
	currentFavoriteMap := make(map[uuid.UUID]bool)
	for _, fav := range currentFavorites {
		currentFavoriteMap[fav.ProductID] = true
	}

	// Товары для добавления
	var toAdd []uuid.UUID
	for _, productID := range req.ProductIDs {
		if !currentFavoriteMap[productID] {
			toAdd = append(toAdd, productID)
		}
	}

	// Товары для удаления
	requestMap := make(map[uuid.UUID]bool)
	for _, productID := range req.ProductIDs {
		requestMap[productID] = true
	}

	var toRemove []uuid.UUID
	for _, fav := range currentFavorites {
		if !requestMap[fav.ProductID] {
			toRemove = append(toRemove, fav.ProductID)
		}
	}

	// Выполняем операции в транзакции
	tx := database.DB.Begin()

	// Удаляем лишние
	if len(toRemove) > 0 {
		if err := tx.Where("user_id = ? AND product_id IN ?", user.ID, toRemove).Delete(&models.Favorite{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to sync favorites",
			))
			return
		}
	}

	// Добавляем новые
	for _, productID := range toAdd {
		favorite := models.Favorite{
			UserID:    user.ID,
			ProductID: productID,
		}
		if err := tx.Create(&favorite).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Failed to sync favorites",
			))
			return
		}
	}

	tx.Commit()

	// Получаем обновленный список избранного
	var updatedFavorites []models.Favorite
	database.DB.Preload("Product").Where("user_id = ?", user.ID).Find(&updatedFavorites)

	favoriteResponses := make([]models.FavoriteResponse, len(updatedFavorites))
	for i, favorite := range updatedFavorites {
		favoriteResponses[i] = favorite.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		favoriteResponses,
		"Favorites synchronized successfully",
	))
}

// CheckFavorite проверяет, добавлен ли товар в избранное
func (fc *FavoriteController) CheckFavorite(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			models.ErrAuthRequired,
			"User not found",
		))
		return
	}

	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid product ID",
		))
		return
	}

	var favorite models.Favorite
	err = database.DB.Where("user_id = ? AND product_id = ?", user.ID, productID).First(&favorite).Error

	isFavorite := err == nil

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]bool{
		"isFavorite": isFavorite,
	}))
}
