package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mm-api/mm-api/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ServiceController обрабатывает служебные операции
type ServiceController struct {
	db *gorm.DB
}

// NewServiceController создает новый экземпляр контроллера
func NewServiceController(db *gorm.DB) *ServiceController {
	return &ServiceController{db: db}
}

// CreateLog создает новый служебный лог
func (sc *ServiceController) CreateLog(c *gin.Context) {
	var req models.ServiceLogCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log := models.ServiceLog{
		Level:   req.Level,
		Message: req.Message,
		Context: req.Context,
	}

	if err := sc.db.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create log"})
		return
	}

	c.JSON(http.StatusCreated, log.ToResponse())
}

// GetLogs получает список логов с пагинацией
func (sc *ServiceController) GetLogs(c *gin.Context) {
	var logs []models.ServiceLog
	var total int64

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "50")
	level := c.Query("level")

	query := sc.db.Model(&models.ServiceLog{})

	if level != "" {
		query = query.Where("level = ?", level)
	}

	query.Count(&total)

	offset := (parseInt(page) - 1) * parseInt(limit)
	if err := query.Order("created_at DESC").Offset(offset).Limit(parseInt(limit)).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	var responses []models.ServiceLogResponse
	for _, log := range logs {
		responses = append(responses, log.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"total": total,
		"page":  parseInt(page),
		"limit": parseInt(limit),
	})
}

// CreateMetric создает новую метрику
func (sc *ServiceController) CreateMetric(c *gin.Context) {
	var req models.ServiceMetricCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metric := models.ServiceMetric{
		MetricName:  req.MetricName,
		MetricValue: req.MetricValue,
		Tags:        req.Tags,
		RecordedAt:  time.Now(),
	}

	if err := sc.db.Create(&metric).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create metric"})
		return
	}

	c.JSON(http.StatusCreated, metric.ToResponse())
}

// GetMetrics получает список метрик с пагинацией
func (sc *ServiceController) GetMetrics(c *gin.Context) {
	var metrics []models.ServiceMetric
	var total int64

	// Параметры пагинации
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "50")
	metricName := c.Query("metric_name")

	query := sc.db.Model(&models.ServiceMetric{})

	// Фильтр по имени метрики
	if metricName != "" {
		query = query.Where("metric_name = ?", metricName)
	}

	// Подсчет общего количества
	query.Count(&total)

	// Получение данных с пагинацией
	offset := (parseInt(page) - 1) * parseInt(limit)
	if err := query.Order("recorded_at DESC").Offset(offset).Limit(parseInt(limit)).Find(&metrics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics"})
		return
	}

	// Формирование ответа
	var responses []models.ServiceMetricResponse
	for _, metric := range metrics {
		responses = append(responses, metric.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"total": total,
		"page":  parseInt(page),
		"limit": parseInt(limit),
	})
}

// GetSystemMetrics получает агрегированные метрики системы
func (sc *ServiceController) GetSystemMetrics(c *gin.Context) {
	var result []struct {
		MetricName string  `json:"metric_name"`
		AvgValue   float64 `json:"avg_value"`
		MaxValue   float64 `json:"max_value"`
		MinValue   float64 `json:"min_value"`
		Count      int64   `json:"count"`
	}

	// Получаем агрегированные метрики за последний час
	if err := sc.db.Raw(`
		SELECT 
			metric_name,
			AVG(metric_value) as avg_value,
			MAX(metric_value) as max_value,
			MIN(metric_value) as min_value,
			COUNT(*) as count
		FROM service_metrics 
		WHERE recorded_at > NOW() - INTERVAL '1 hour'
		GROUP BY metric_name
		ORDER BY metric_name
	`).Scan(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// SetCache устанавливает значение в кэш
func (sc *ServiceController) SetCache(c *gin.Context) {
	var req models.ServiceCacheCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cache := models.ServiceCache{
		CacheKey:   req.CacheKey,
		CacheValue: req.CacheValue,
		ExpiresAt:  req.ExpiresAt,
	}

	// Используем Upsert (создать или обновить)
	if err := sc.db.Save(&cache).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set cache"})
		return
	}

	c.JSON(http.StatusOK, cache.ToResponse())
}

// GetCache получает значение из кэша
func (sc *ServiceController) GetCache(c *gin.Context) {
	cacheKey := c.Param("key")
	if cacheKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cache key is required"})
		return
	}

	var cache models.ServiceCache
	if err := sc.db.Where("cache_key = ?", cacheKey).First(&cache).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cache key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cache"})
		return
	}

	// Проверяем, не истек ли срок действия
	if cache.ExpiresAt != nil && time.Now().After(*cache.ExpiresAt) {
		// Удаляем истекший кэш
		sc.db.Delete(&cache)
		c.JSON(http.StatusNotFound, gin.H{"error": "Cache expired"})
		return
	}

	c.JSON(http.StatusOK, cache.ToResponse())
}

// UpdateCache обновляет значение в кэше
func (sc *ServiceController) UpdateCache(c *gin.Context) {
	cacheKey := c.Param("key")
	if cacheKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cache key is required"})
		return
	}

	var req models.ServiceCacheUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cache models.ServiceCache
	if err := sc.db.Where("cache_key = ?", cacheKey).First(&cache).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cache key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cache"})
		return
	}

	// Обновляем значения
	cache.CacheValue = req.CacheValue
	if req.ExpiresAt != nil {
		cache.ExpiresAt = req.ExpiresAt
	}

	if err := sc.db.Save(&cache).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cache"})
		return
	}

	c.JSON(http.StatusOK, cache.ToResponse())
}

// DeleteCache удаляет значение из кэша
func (sc *ServiceController) DeleteCache(c *gin.Context) {
	cacheKey := c.Param("key")
	if cacheKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cache key is required"})
		return
	}

	if err := sc.db.Where("cache_key = ?", cacheKey).Delete(&models.ServiceCache{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cache deleted successfully"})
}

// ClearExpiredCache очищает истекший кэш
func (sc *ServiceController) ClearExpiredCache(c *gin.Context) {
	result := sc.db.Where("expires_at IS NOT NULL AND expires_at < NOW()").Delete(&models.ServiceCache{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear expired cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Expired cache cleared successfully",
		"deleted_count": result.RowsAffected,
	})
}

// Вспомогательная функция для парсинга int
func parseInt(s string) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 1
	}
	return result
}
