package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mm-api/mm-api/database"
)

// DebugController предоставляет диагностические эндпоинты
type DebugController struct{}

// DBInfo возвращает информацию о подключении к БД и базовые счетчики
func (dc *DebugController) DBInfo(c *gin.Context) {
	var currentDB string
	var productsCount int64
	var variationsCount int64
	var joinedCount int64

	_ = database.DB.Raw("SELECT current_database()").Scan(&currentDB).Error
	_ = database.DB.Raw("SELECT COUNT(*) FROM public.products").Scan(&productsCount).Error
	_ = database.DB.Raw("SELECT COUNT(*) FROM public.product_variations").Scan(&variationsCount).Error
	_ = database.DB.Raw("SELECT COUNT(*) FROM public.products p INNER JOIN public.product_variations pv ON p.id = pv.product_id").Scan(&joinedCount).Error

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"currentDatabase": currentDB,
			"counts": gin.H{
				"products":             productsCount,
				"productVariations":    variationsCount,
				"productsJoinVariants": joinedCount,
			},
		},
	})
}
