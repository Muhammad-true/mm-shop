package controllers

import (
	"net/http"
	"strconv"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryController обрабатывает запросы категорий
type CategoryController struct{}

// GetCategories возвращает все категории
func (cc *CategoryController) GetCategories(c *gin.Context) {
	var categories []models.Category

	if err := database.DB.Where("parent_id IS NULL").Preload("Subcategories").Order("sort_order ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to fetch categories",
		))
		return
	}

	// Преобразуем в response
	categoryTrees := make([]models.CategoryTree, len(categories))
	for i, category := range categories {
		categoryTrees[i] = category.ToTree()
	}

	c.JSON(http.StatusOK, models.SuccessResponse(categoryTrees))
}

// GetCategory возвращает категорию по ID
func (cc *CategoryController) GetCategory(c *gin.Context) {
	id := c.Param("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid category ID",
		))
		return
	}

	var category models.Category
	if err := database.DB.Preload("Subcategories").First(&category, "id = ?", categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Category not found",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(category.ToResponse()))
}

// GetCategoryProducts возвращает товары в категории
func (cc *CategoryController) GetCategoryProducts(c *gin.Context) {
	id := c.Param("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid category ID",
		))
		return
	}

	// Проверяем существование категории
	var category models.Category
	if err := database.DB.First(&category, "id = ?", categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Category not found",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
		}
		return
	}

	var products []models.Product
	query := database.DB.Model(&models.Product{}).Where("category_id = ?", categoryID)

	// Фильтрация по доступности
	if inStock := c.Query("in_stock"); inStock == "true" {
		query = query.Where("is_available = ?", true)
	}

	// Поиск
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Сортировка
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	allowedSorts := map[string]bool{
		"name":       true,
		"price":      true,
		"rating":     true,
		"created_at": true,
	}

	if allowedSorts[sortBy] {
		if sortOrder == "asc" {
			query = query.Order(sortBy + " ASC")
		} else {
			query = query.Order(sortBy + " DESC")
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

	// Получаем продукты
	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to fetch products",
		))
		return
	}

	// Преобразуем в response
	productResponses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = product.ToResponse()
	}

	// Вычисляем пагинацию
	totalPages := (int(total) + limit - 1) / limit
	pagination := models.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.PaginatedSuccessResponse(productResponses, pagination))
}

// CreateCategory создает новую категорию (только для админов)
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	var req models.CategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	// Если указан родитель, проверяем его существование
	if req.ParentID != nil {
		var parent models.Category
		if err := database.DB.First(&parent, "id = ?", *req.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
				models.ErrValidationError,
				"Parent category not found",
			))
			return
		}
	}

	category := models.Category{
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}

	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to create category",
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(
		category.ToResponse(),
		"Category created successfully",
	))
}

// UpdateCategory обновляет категорию (только для админов)
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid category ID",
		))
		return
	}

	var req models.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid request data",
			err.Error(),
		))
		return
	}

	var category models.Category
	if err := database.DB.First(&category, "id = ?", categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Category not found",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				models.ErrInternalError,
				"Database error",
			))
		}
		return
	}

	// Обновляем поля
	category.Name = req.Name
	category.Description = req.Description
	category.IconURL = req.IconURL
	category.ParentID = req.ParentID
	category.SortOrder = req.SortOrder
	category.IsActive = req.IsActive

	if err := database.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to update category",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		category.ToResponse(),
		"Category updated successfully",
	))
}

// DeleteCategory удаляет категорию (только для админов)
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Invalid category ID",
		))
		return
	}

	// Проверяем, есть ли продукты в этой категории
	var productCount int64
	database.DB.Model(&models.Product{}).Where("category_id = ?", categoryID).Count(&productCount)
	if productCount > 0 {
		c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Cannot delete category with existing products",
		))
		return
	}

	// Проверяем, есть ли подкатегории
	var subcategoryCount int64
	database.DB.Model(&models.Category{}).Where("parent_id = ?", categoryID).Count(&subcategoryCount)
	if subcategoryCount > 0 {
		c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Cannot delete category with subcategories",
		))
		return
	}

	result := database.DB.Delete(&models.Category{}, "id = ?", categoryID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Failed to delete category",
		))
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
			models.ErrNotFound,
			"Category not found",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		nil,
		"Category deleted successfully",
	))
}
