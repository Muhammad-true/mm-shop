package controllers

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mm-api/mm-api/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/image/webp"
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
	
	// Нормализуем расширение к нижнему регистру для Linux
	ext = strings.ToLower(ext)

	// Проверяем допустимые расширения
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
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

	// Сжимаем и сохраняем изображение
	originalSize := header.Size
	finalFilename, bytesWritten, err := uc.compressAndSaveImage(file, filePath, ext, contentType)
	if err != nil {
		log.Printf("❌ Ошибка сохранения файла: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save file",
			"details": err.Error(),
		})
		return
	}

	// Обновляем filename, если формат изменился (PNG/WebP -> JPEG)
	if finalFilename != filename {
		filename = finalFilename
		log.Printf("🔄 Формат изменен, новое имя файла: %s", filename)
	}

	// Вычисляем процент сжатия
	compressionRatio := float64(bytesWritten) / float64(originalSize) * 100
	savedBytes := originalSize - bytesWritten
	log.Printf("✅ Файл успешно сохранен: %d байт записано (было %d байт, сжато на %.1f%%, сэкономлено %d байт)", 
		bytesWritten, originalSize, 100-compressionRatio, savedBytes)

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

// compressAndSaveImage сжимает и сохраняет изображение
// Поддерживает JPEG, PNG, WebP
// Для JPEG: качество 85% (хороший баланс между размером и качеством)
// Для PNG: конвертирует в JPEG для лучшего сжатия (если возможно)
// Возвращает: (finalFilename, bytesWritten, error)
func (uc *UploadController) compressAndSaveImage(file io.Reader, filePath string, ext string, contentType string) (string, int64, error) {
	// Декодируем изображение
	var img image.Image
	var err error
	finalExt := ext
	finalPath := filePath

	switch {
	case ext == ".jpg" || ext == ".jpeg" || strings.Contains(contentType, "jpeg"):
		img, err = jpeg.Decode(file)
		if err != nil {
			return "", 0, fmt.Errorf("ошибка декодирования JPEG: %v", err)
		}
	case ext == ".png" || strings.Contains(contentType, "png"):
		img, err = png.Decode(file)
		if err != nil {
			return "", 0, fmt.Errorf("ошибка декодирования PNG: %v", err)
		}
		// PNG конвертируем в JPEG для лучшего сжатия
		finalExt = ".jpg"
		finalPath = strings.TrimSuffix(filePath, ".png") + ".jpg"
	case ext == ".webp" || strings.Contains(contentType, "webp"):
		img, err = webp.Decode(file)
		if err != nil {
			// Если не удалось декодировать WebP, пробуем сохранить как есть
			log.Printf("⚠️ Не удалось декодировать WebP, сохраняем без сжатия: %v", err)
			dst, err := os.Create(filePath)
			if err != nil {
				return "", 0, err
			}
			defer dst.Close()
			bytesWritten, err := io.Copy(dst, file)
			return filepath.Base(filePath), bytesWritten, err
		}
		// WebP конвертируем в JPEG для лучшего сжатия
		finalExt = ".jpg"
		finalPath = strings.TrimSuffix(filePath, ".webp") + ".jpg"
	default:
		// Для других форматов (GIF и т.д.) просто копируем без сжатия
		dst, err := os.Create(filePath)
		if err != nil {
			return "", 0, err
		}
		defer dst.Close()
		bytesWritten, err := io.Copy(dst, file)
		return filepath.Base(filePath), bytesWritten, err
	}

	// Создаем файл для сохранения
	dst, err := os.Create(finalPath)
	if err != nil {
		return "", 0, err
	}
	defer dst.Close()

	// Сохраняем с сжатием
	// JPEG качество 85% - хороший баланс между размером и качеством
	// Можно уменьшить до 75% для большего сжатия, но качество будет хуже
	quality := 85

	err = jpeg.Encode(dst, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return "", 0, fmt.Errorf("ошибка кодирования JPEG: %v", err)
	}

	// Получаем размер файла
	info, err := dst.Stat()
	if err != nil {
		return "", 0, err
	}
	bytesWritten := info.Size()

	return filepath.Base(finalPath), bytesWritten, nil
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
