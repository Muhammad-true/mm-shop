package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
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
	
	// Используем стандартные методы Gin - они автоматически парсят multipart при первом обращении
	// При потоковой передаче (proxy_request_buffering off) Gin парсит форму по мере чтения
	platformStr := c.PostForm("platform")
	version := strings.TrimSpace(c.PostForm("version"))
	releaseNotes := c.PostForm("releaseNotes")

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

	log.Println("📁 [UploadUpdate] Получение файла из запроса...")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("❌ [UploadUpdate] Ошибка получения файла: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "file is required",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()
	
	log.Printf("✅ [UploadUpdate] Файл получен: %s, размер: %d байт", header.Filename, header.Size)

	ext := strings.ToLower(filepath.Ext(header.Filename))
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

	// Отправляем промежуточный ответ, чтобы браузер знал, что сервер получил файл
	// Это особенно важно для Cloudflare, чтобы он не закрыл соединение
	if flusher, ok := c.Writer.(http.Flusher); ok {
		c.Writer.WriteHeader(http.StatusProcessing) // 102 Processing
		flusher.Flush()
		log.Println("🔄 [UploadUpdate] Отправлен промежуточный ответ 102 Processing")
	}

	log.Println("📥 [UploadUpdate] Начало копирования файла и вычисления SHA256...")
	hasher := sha256.New()
	size, err := io.Copy(io.MultiWriter(dst, hasher), file)
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
