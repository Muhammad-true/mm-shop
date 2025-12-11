package controllers

import (
	"net/http"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionController обрабатывает запросы планов подписки
type SubscriptionController struct{}

// GetSubscriptionPlans возвращает список всех активных планов подписки (публичный)
func (sc *SubscriptionController) GetSubscriptionPlans(c *gin.Context) {
	var plans []models.SubscriptionPlan
	if err := database.DB.Where("is_active = ?", true).Order("sort_order ASC, price ASC").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch subscription plans",
		})
		return
	}

	responses := make([]models.SubscriptionPlanResponse, len(plans))
	for i, plan := range plans {
		responses[i] = plan.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"plans": responses,
		},
	})
}

// GetSubscriptionPlan возвращает информацию о плане подписки (публичный)
func (sc *SubscriptionController) GetSubscriptionPlan(c *gin.Context) {
	planIDParam := c.Param("id")
	planID, err := uuid.Parse(planIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid plan ID",
		})
		return
	}

	var plan models.SubscriptionPlan
	if err := database.DB.Where("id = ? AND is_active = ?", planID, true).First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Subscription plan not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Database error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    plan.ToResponse(),
	})
}

