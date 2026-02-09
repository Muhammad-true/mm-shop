package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
)

// UpdateController обрабатывает загрузку и выдачу обновлений
type UpdateController struct{}

// UploadUpdate загружает файл обновления (только для админов)
func (uc *UpdateController) UploadUpdate(c *gin.Context) {
	log.Println("📤 [UploadUpdate] Начало загрузки обновления")
	
	// Логируем информацию о запросе
	log.Printf("🔍 [UploadUpdate] Content-Type: %s", c.Request.Header.Get("Content-Type"))
	log.Printf("🔍 [UploadUpdate] Content-Length: %s", c.Request.Header.Get("Content-Length"))
	log.Printf("🔍 [UploadUpdate] Method: %s", c.Request.Method)
	log.Printf("🔍 [UploadUpdate] URL: %s", c.Request.URL.String())
	
	// Отправляем промежуточный ответ сразу, чтобы клиент знал, что сервер обрабатывает запрос
	if flusher, ok := c.Writer.(http.Flusher); ok {
		c.Writer.WriteHeader(http.StatusProcessing) // 102 Processing
		flusher.Flush()
		log.Println("✅ [UploadUpdate] Отправлен промежуточный ответ 102 Processing")
	}
	
	// Парсим multipart форму потоково через multipart.Reader
	// Это работает при proxy_request_buffering off
	contentType := c.Request.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		log.Printf("❌ [UploadUpdate] Неверный Content-Type: %s", contentType)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Content-Type must be multipart/form-data",
		})
		return
	}
	
	// Создаем multipart reader для потокового парсинга
	boundary := ""
	if parts := strings.Split(contentType, "boundary="); len(parts) > 1 {
		boundary = parts[1]
	}
	if boundary == "" {
		log.Println("❌ [UploadUpdate] Boundary не найден в Content-Type")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "multipart boundary not found",
		})
		return
	}
	
	reader := multipart.NewReader(c.Request.Body, boundary)
	
	// Переменные для хранения данных формы
	var platformStr, version, releaseNotes string
	var filePart *multipart.Part
	var fileName string
	
	// Читаем все части multipart формы
	log.Println("🔄 [UploadUpdate] Парсинг multipart формы потоково...")
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("❌ [UploadUpdate] Ошибка чтения части формы: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "failed to parse multipart form",
				"details": err.Error(),
			})
			return
		}
		
		formName := part.FormName()
		log.Printf("📋 [UploadUpdate] Обработка поля формы: %s", formName)
		
		if formName == "file" {
			// Это файл - сохраняем part для дальнейшего чтения
			filePart = part
			fileName = part.FileName()
			log.Printf("✅ [UploadUpdate] Файл найден: %s", fileName)
			// НЕ закрываем part здесь - будем читать из него дальше
		} else {
			// Это текстовое поле - читаем сразу
			data, err := io.ReadAll(part)
			part.Close()
			if err != nil {
				log.Printf("❌ [UploadUpdate] Ошибка чтения поля %s: %v", formName, err)
				continue
			}
			value := string(data)
			
			switch formName {
			case "platform":
				platformStr = value
				log.Printf("✅ [UploadUpdate] platform: %s", platformStr)
			case "version":
				version = strings.TrimSpace(value)
				log.Printf("✅ [UploadUpdate] version: %s", version)
			case "releaseNotes":
				releaseNotes = value
				log.Printf("✅ [UploadUpdate] releaseNotes: %s", releaseNotes)
			}
		}
	}
	
	log.Printf("📋 [UploadUpdate] Параметры: platform=%s, version=%s, releaseNotes=%s", platformStr, version, releaseNotes)

	if platformStr == "" || version == "" {
		log.Printf("❌ [UploadUpdate] Отсутствуют обязательные параметры: platform=%s, version=%s", platformStr, version)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "platform and version are required",
		})
		return
	}

	platform := models.UpdatePlatform(platformStr)
	if platform != models.UpdatePlatformServer &&
		platform != models.UpdatePlatformWindows &&
		platform != models.UpdatePlatformAndroid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid platform (allowed: server, windows, android)",
		})
		return
	}

	// Проверяем, что файл был найден
	if filePart == nil {
		log.Println("❌ [UploadUpdate] Файл не найден в форме")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "file is required",
		})
		return
	}
	defer filePart.Close()
	
	log.Printf("✅ [UploadUpdate] Файл получен: %s", fileName)

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "file extension is required",
		})
		return
	}

	allowedExts := []string{".zip", ".exe", ".apk"}
	isAllowed := false
	for _, e := range allowedExts {
		if ext == e {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("unsupported extension %s (allowed: %v)", ext, allowedExts),
		})
		return
	}

	dir := filepath.Join("updates", string(platform))
	log.Printf("📂 [UploadUpdate] Создание директории: %s", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("❌ [UploadUpdate] Ошибка создания директории: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to create updates directory",
			"details": err.Error(),
		})
		return
	}

	filename := fmt.Sprintf("%s_%s_%s%s", platform, version, uuid.NewString(), ext)
	filePath := filepath.Join(dir, filename)
	log.Printf("💾 [UploadUpdate] Сохранение файла: %s", filePath)

	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("❌ [UploadUpdate] Ошибка создания файла: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to create file",
			"details": err.Error(),
		})
		return
	}
	defer dst.Close()

	log.Println("📥 [UploadUpdate] Начало копирования файла и вычисления SHA256...")
	hasher := sha256.New()
	size, err := io.Copy(io.MultiWriter(dst, hasher), filePart)
	if err != nil {
		log.Printf("❌ [UploadUpdate] Ошибка копирования файла: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to save file",
			"details": err.Error(),
		})
		return
	}
	log.Printf("✅ [UploadUpdate] Файл скопирован: %d байт", size)

	log.Println("🔐 [UploadUpdate] Вычисление SHA256...")
	checksum := hex.EncodeToString(hasher.Sum(nil))
	log.Printf("✅ [UploadUpdate] SHA256 вычислен: %s", checksum[:16]+"...")
	
	fileURL := fmt.Sprintf("/updates/%s/%s", platform, filename)

	log.Println("💾 [UploadUpdate] Сохранение метаданных в БД...")
	update := models.UpdateRelease{
		Platform:       platform,
		Version:        version,
		FileName:       filename,
		FilePath:       filePath,
		FileURL:        fileURL,
		FileSize:       size,
		ChecksumSHA256: checksum,
		ReleaseNotes:   releaseNotes,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := database.DB.Create(&update).Error; err != nil {
		log.Printf("❌ [UploadUpdate] Ошибка сохранения в БД: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to save update metadata",
			"details": err.Error(),
		})
		return
	}
	log.Printf("✅ [UploadUpdate] Метаданные сохранены в БД, ID: %s", update.ID)

	log.Println("🎉 [UploadUpdate] Загрузка завершена успешно!")
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Update uploaded successfully",
		"data":    update,
	})
}

// ListUpdates возвращает список обновлений (админ)
func (uc *UpdateController) ListUpdates(c *gin.Context) {
	platform := c.Query("platform")

	query := database.DB.Model(&models.UpdateRelease{}).Order("created_at DESC")
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	var updates []models.UpdateRelease
	if err := query.Find(&updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch updates",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updates,
	})
}

// GetLatestUpdate возвращает последнее активное обновление по платформе
func (uc *UpdateController) GetLatestUpdate(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "platform is required",
		})
		return
	}

	var update models.UpdateRelease
	if err := database.DB.Where("platform = ? AND is_active = ?", platform, true).
		Order("created_at DESC").
		First(&update).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "update not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    update,
	})
}
