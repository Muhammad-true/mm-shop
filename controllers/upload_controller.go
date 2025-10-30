package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mm-api/mm-api/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadController struct{}

// GetImageURL возвращает правильный URL для изображения
func (uc *UploadController) GetImageURL(filename, folder string) string {
	// Всегда возвращаем относительный путь для same-origin запросов
	// nginx будет проксировать /images/ к API
	return fmt.Sprintf("/images/%s/%s", folder, filename)
}

// UploadImage загружает изображение
func (uc *UploadController) UploadImage(c *gin.Context) {
	log.Printf("📸 Начало загрузки изображения...")

	// Проверяем, что это POST запрос
	if c.Request.Method != "POST" {
		log.Printf("❌ Неверный метод: %s", c.Request.Method)
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Method not allowed",
		})
		return
	}

	// Получаем файл из формы
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		log.Printf("❌ Ошибка получения файла: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No image file provided",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	log.Printf("📁 Получен файл: %s, размер: %d байт", header.Filename, header.Size)

	// Проверяем тип файла
	contentType := header.Header.Get("Content-Type")
	log.Printf("📋 Content-Type: %s", contentType)

	if !strings.HasPrefix(contentType, "image/") {
		log.Printf("❌ Неверный тип файла: %s", contentType)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "File is not an image",
			"contentType": contentType,
		})
		return
	}

	// Получаем конфигурацию для максимального размера файла
	cfg := config.GetConfig()
	maxSizeStr := cfg.UploadMaxSize
	maxSize := int64(50 * 1024 * 1024) // По умолчанию 50MB

	// Парсим размер из конфигурации
	if strings.HasSuffix(maxSizeStr, "MB") {
		var mb int
		fmt.Sscanf(maxSizeStr, "%dMB", &mb)
		maxSize = int64(mb) * 1024 * 1024
	} else if strings.HasSuffix(maxSizeStr, "KB") {
		var kb int
		fmt.Sscanf(maxSizeStr, "%dKB", &kb)
		maxSize = int64(kb) * 1024
	}

	// Проверяем размер файла
	if header.Size > maxSize {
		log.Printf("❌ Файл слишком большой: %d байт (максимум %d байт)", header.Size, maxSize)
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":   fmt.Sprintf("File size too large (max %s)", maxSizeStr),
			"size":    header.Size,
			"maxSize": maxSize,
		})
		return
	}

	// Получаем расширение файла
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg" // По умолчанию
	}

	// Проверяем допустимые расширения
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".JPG", ".JPEG", ".PNG", ".GIF", ".WEBP"}
	isAllowed := false
	for _, allowedExt := range allowedExts {
		if strings.EqualFold(ext, allowedExt) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		log.Printf("❌ Неподдерживаемое расширение: %s", ext)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Unsupported file extension",
			"extension": ext,
			"allowed":   allowedExts,
		})
		return
	}

	// Генерируем уникальное имя файла
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	log.Printf("🆔 Сгенерировано имя файла: %s", filename)

	// Определяем папку для сохранения
	folder := c.DefaultQuery("folder", "uploads")
	uploadDir := fmt.Sprintf("images/%s", folder)
	log.Printf("📂 Папка для сохранения: %s", uploadDir)

	// Создаем папку, если её нет
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("❌ Ошибка создания папки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create upload directory",
			"details": err.Error(),
		})
		return
	}

	// Путь для сохранения файла
	filePath := filepath.Join(uploadDir, filename)
	log.Printf("💾 Путь сохранения: %s", filePath)

	// Создаем файл
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("❌ Ошибка создания файла: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create file",
			"details": err.Error(),
		})
		return
	}
	defer dst.Close()

	// Копируем содержимое файла
	bytesWritten, err := io.Copy(dst, file)
	if err != nil {
		log.Printf("❌ Ошибка сохранения файла: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save file",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Файл успешно сохранен: %d байт записано", bytesWritten)

	// Формируем URL для доступа к файлу
	fileURL := uc.GetImageURL(filename, folder)
	log.Printf("🔗 URL файла: %s", fileURL)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"url":      fileURL,
		"filename": filename,
		"size":     bytesWritten,
		"folder":   folder,
	})
}

// DeleteImage удаляет изображение
func (uc *UploadController) DeleteImage(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Filename is required",
		})
		return
	}

	folder := c.DefaultQuery("folder", "uploads")
	filePath := filepath.Join("images", folder, filename)

	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	// Удаляем файл
	if err := os.Remove(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "File deleted successfully",
	})
}
