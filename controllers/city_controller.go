package controllers

import (
	"math"
	"net/http"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"

	"github.com/gin-gonic/gin"
)

// CityController обрабатывает запросы городов
type CityController struct{}

// GetCities возвращает список всех активных городов
func (cc *CityController) GetCities(c *gin.Context) {
	var cities []models.City
	if err := database.DB.Where("is_active = ?", true).Order("name ASC").Find(&cities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch cities",
		})
		return
	}

	cityResponses := make([]models.CityResponse, len(cities))
	for i, city := range cities {
		cityResponses[i] = city.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"cities": cityResponses,
		},
	})
}

// GetCity возвращает информацию о городе по ID
func (cc *CityController) GetCity(c *gin.Context) {
	cityID := c.Param("id")
	
	var city models.City
	if err := database.DB.Where("id = ? AND is_active = ?", cityID, true).First(&city).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "City not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    city.ToResponse(),
	})
}

// FindCityByLocation находит ближайший город по координатам
func (cc *CityController) FindCityByLocation(c *gin.Context) {
	var req struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Получаем все активные города
	var cities []models.City
	if err := database.DB.Where("is_active = ?", true).Find(&cities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch cities",
		})
		return
	}

	if len(cities) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No cities found",
		})
		return
	}

	// Находим ближайший город
	var nearestCity models.City
	minDistance := math.MaxFloat64

	for _, city := range cities {
		distance := calculateDistance(req.Latitude, req.Longitude, city.Latitude, city.Longitude)
		if distance < minDistance {
			minDistance = distance
			nearestCity = city
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"city":     nearestCity.ToResponse(),
			"distance": minDistance, // Расстояние в километрах
		},
	})
}

// calculateDistance вычисляет расстояние между двумя точками по формуле гаверсинуса (в километрах)
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // Радиус Земли в километрах

	// Преобразуем градусы в радианы
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Разница координат
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Формула гаверсинуса
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

