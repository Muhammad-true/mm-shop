package controllers

import (
	"net/http"
	"strconv"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleController struct{}

// GetRoles - получить все роли
func (rc *RoleController) GetRoles(c *gin.Context) {
	var roles []models.Role
	var total int64

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Подсчет общего количества
	database.DB.Model(&models.Role{}).Count(&total)

	// Получение ролей с пагинацией
	result := database.DB.Preload("Users").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&roles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении ролей",
		))
		return
	}

	// Преобразование в ответы с подсчетом пользователей
	var roleResponses []models.RoleResponse
	for _, role := range roles {
		response := role.ToResponse()
		response.UserCount = len(role.Users)
		roleResponses = append(roleResponses, response)
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data: gin.H{
			"roles": roleResponses,
			"pagination": models.PaginationInfo{
				Page:       page,
				Limit:      limit,
				Total:      int(total),
				TotalPages: int((total + int64(limit) - 1) / int64(limit)),
			},
		},
		Message: "Роли получены успешно",
	})
}

// GetRole - получить роль по ID
func (rc *RoleController) GetRole(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	result := database.DB.Preload("Users").
		First(&role, "id = ?", roleID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Роль не найдена",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении роли",
		))
		return
	}

	response := role.ToResponse()
	response.UserCount = len(role.Users)

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    response,
		Message: "Роль получена успешно",
	})
}

// CreateRole - создать новую роль
func (rc *RoleController) CreateRole(c *gin.Context) {
	var req models.RoleCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверные данные запроса",
			err.Error(),
		))
		return
	}

	// Проверяем, существует ли роль с таким именем
	var existingRole models.Role
	if err := database.DB.Where("name = ?", req.Name).First(&existingRole).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
			models.ErrUserAlreadyExists,
			"Роль с таким именем уже существует",
		))
		return
	}

	// Создаем новую роль
	role := models.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Permissions: req.Permissions,
		IsActive:    true,
		IsSystem:    false, // Пользовательские роли не системные
	}

	if err := database.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при создании роли",
		))
		return
	}

	c.JSON(http.StatusCreated, models.StandardResponse{
		Success: true,
		Data:    role.ToResponse(),
		Message: "Роль создана успешно",
	})
}

// UpdateRole - обновить роль
func (rc *RoleController) UpdateRole(c *gin.Context) {
	roleID := c.Param("id")

	var req models.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			models.ErrValidationError,
			"Неверные данные запроса",
			err.Error(),
		))
		return
	}

	var role models.Role
	if err := database.DB.First(&role, "id = ?", roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Роль не найдена",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении роли",
		))
		return
	}

	// Проверяем, не пытаемся ли изменить системную роль
	if role.IsSystem {
		c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
			models.ErrForbidden,
			"Нельзя изменять системные роли",
		))
		return
	}

	// Обновляем поля
	updates := make(map[string]interface{})
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Permissions != nil {
		updates["permissions"] = *req.Permissions
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := database.DB.Model(&role).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при обновлении роли",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    role.ToResponse(),
		Message: "Роль обновлена успешно",
	})
}

// DeleteRole - удалить роль
func (rc *RoleController) DeleteRole(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	if err := database.DB.Preload("Users").First(&role, "id = ?", roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				models.ErrNotFound,
				"Роль не найдена",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при получении роли",
		))
		return
	}

	// Проверяем, не пытаемся ли удалить системную роль
	if role.IsSystem {
		c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
			models.ErrForbidden,
			"Нельзя удалять системные роли",
		))
		return
	}

	// Проверяем, есть ли пользователи с этой ролью
	if len(role.Users) > 0 {
		c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
			models.ErrConflict,
			"Нельзя удалить роль, которая назначена пользователям",
		))
		return
	}

	if err := database.DB.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			models.ErrInternalError,
			"Ошибка при удалении роли",
		))
		return
	}

	c.JSON(http.StatusOK, models.StandardResponse{
		Success: true,
		Message: "Роль удалена успешно",
	})
}
