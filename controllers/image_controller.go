package controllers

import (
	"log"
	"net/http"
	"strings"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/gin-gonic/gin"
)

type ImageController struct{}

// FixImageURLs исправляет URL изображений в базе данных
func (ic *ImageController) FixImageURLs(c *gin.Context) {
	log.Printf("🔧 Начинаем исправление URL изображений...")

	// Получаем все товары с вариациями
	var products []models.Product
	if err := database.DB.Preload("Variations").Find(&products).Error; err != nil {
		log.Printf("❌ Ошибка загрузки товаров: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load products",
		})
		return
	}

	fixedCount := 0
	totalImages := 0

	// Проходим по всем товарам
	for _, product := range products {
		// Проверяем вариации товара
		for i, variation := range product.Variations {
			if len(variation.ImageURLs) > 0 {
				totalImages += len(variation.ImageURLs)

				// Проверяем и исправляем каждый URL
				for j, imageURL := range variation.ImageURLs {
					if imageURL != "" {
						// Если абсолютный URL на /images/ — делаем относительным
						if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
							if strings.Contains(imageURL, "/images/") {
								// Извлекаем путь /images/... из абсолютного URL
								parts := strings.Split(imageURL, "/images/")
								if len(parts) > 1 {
									newURL := "/images/" + parts[1]
									product.Variations[i].ImageURLs[j] = newURL
									fixedCount++
									log.Printf("🔧 Исправлен URL: %s -> %s", imageURL, newURL)
								}
							}
						}
					}
				}
			}
		}

		// Сохраняем изменения
		if err := database.DB.Save(&product).Error; err != nil {
			log.Printf("❌ Ошибка сохранения товара %s: %v", product.ID, err)
		}
	}

	log.Printf("✅ Исправление завершено. Исправлено %d из %d изображений", fixedCount, totalImages)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Image URLs fixed successfully",
		"fixedCount":  fixedCount,
		"totalImages": totalImages,
	})
}

// GetImageURL возвращает правильный URL для изображения
func (ic *ImageController) GetImageURL(c *gin.Context) {
	filename := c.Param("filename")
	folder := c.DefaultQuery("folder", "uploads")

	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Filename is required",
		})
		return
	}

	// Используем функцию из UploadController
	uploadController := &UploadController{}
	imageURL := uploadController.GetImageURL(filename, folder)

	c.JSON(http.StatusOK, gin.H{
		"url": imageURL,
	})
}
