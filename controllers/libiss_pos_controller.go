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

// LibissPosController обрабатывает загрузку и выдачу файлов программ libiss_pos
type LibissPosController struct{}

// UploadFile загружает файл программы libiss_pos (только для админов)
func (lc *LibissPosController) UploadFile(c *gin.Context) {
	// Получаем текущего пользователя из контекста
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	userModel := user.(models.User)
	userID := userModel.ID

	// Получаем параметры формы
	fileTypeStr := strings.TrimSpace(c.PostForm("type"))
	platformStr := strings.TrimSpace(c.PostForm("platform"))
	version := strings.TrimSpace(c.PostForm("version"))
	description := strings.TrimSpace(c.PostForm("description"))
	isPublicStr := strings.TrimSpace(c.PostForm("isPublic"))

	if fileTypeStr == "" || version == "" || platformStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "type, platform and version are required",
		})
		return
	}

	fileType := models.LibissPosType(fileTypeStr)
	if fileType != models.LibissPosTypeFull &&
		fileType != models.LibissPosTypeCassa2 &&
		fileType != models.LibissPosTypeServerOnly {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid type (allowed: full, cassa2, server_only)",
		})
		return
	}

	platform := models.LibissPosPlatform(platformStr)
	if platform != models.LibissPosPlatformWindows && platform != models.LibissPosPlatformAndroid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid platform (allowed: windows, android)",
		})
		return
	}

	// Получаем файл
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "file is required",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	// Проверяем расширение файла в зависимости от платформы
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if platform == models.LibissPosPlatformWindows && ext != ".exe" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "for Windows platform only .exe files are allowed",
		})
		return
	}
	if platform == models.LibissPosPlatformAndroid && ext != ".apk" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "for Android platform only .apk files are allowed",
		})
		return
	}

	// Проверяем размер файла (максимум 600MB для поддержки полного пакета ~595MB)
	maxSize := int64(600 * 1024 * 1024)
		if header.Size > maxSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"success": false,
			"error":   fmt.Sprintf("file size too large (max 600MB), got %d bytes", header.Size),
		})
		return
	}

	// Создаем директорию для хранения файлов
	dir := filepath.Join("libiss_pos", string(fileType))
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("❌ Failed to create directory %s: %v", dir, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to create upload directory",
			"details": err.Error(),
		})
		return
	}

	// Генерируем уникальное имя файла
	filename := fmt.Sprintf("%s_%s_%s%s", fileType, version, uuid.NewString()[:8], ext)
	filePath := filepath.Join(dir, filename)

	// Создаем файл
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("❌ Failed to create file %s: %v", filePath, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to create file",
			"details": err.Error(),
		})
		return
	}
	defer dst.Close()

	// Сохраняем файл и вычисляем SHA256
	hasher := sha256.New()
	size, err := io.Copy(io.MultiWriter(dst, hasher), file)
	if err != nil {
		log.Printf("❌ Failed to save file %s: %v", filePath, err)
		// Удаляем файл при ошибке
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to save file",
			"details": err.Error(),
		})
		return
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Формируем URL для доступа к файлу
	fileURL := fmt.Sprintf("/api/v1/libiss-pos/download/%s", filename)
	publicURL := fmt.Sprintf("/api/v1/libiss-pos/public/%s", filename)

	// Определяем, должен ли файл быть публичным
	isPublic := true
	if isPublicStr == "false" {
		isPublic = false
	}

	// Сохраняем информацию о файле в БД
	libissFile := models.LibissPosFile{
		Type:           fileType,
		Platform:       platform,
		Version:        version,
		FileName:       filename,
		OriginalName:   header.Filename,
		FilePath:       filePath,
		FileURL:        fileURL,
		PublicURL:      publicURL,
		FileSize:       size,
		ChecksumSHA256: checksum,
		Description:    description,
		IsActive:       true,
		IsPublic:       isPublic,
		DownloadCount:  0,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := database.DB.Create(&libissFile).Error; err != nil {
		log.Printf("❌ Failed to save file metadata: %v", err)
		// Удаляем файл при ошибке
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to save file metadata",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ File uploaded successfully: %s (type: %s, version: %s)", filename, fileType, version)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "File uploaded successfully",
		"data":    libissFile,
	})
}

// ListFiles возвращает список загруженных файлов (админ)
func (lc *LibissPosController) ListFiles(c *gin.Context) {
	fileType := c.Query("type")
	platform := c.Query("platform")

	query := database.DB.Model(&models.LibissPosFile{}).Order("created_at DESC")
	if fileType != "" {
		query = query.Where("type = ?", fileType)
	}
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	var files []models.LibissPosFile
	if err := query.Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch files",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    files,
	})
}

// DownloadFile скачивает файл (требует аутентификации)
func (lc *LibissPosController) DownloadFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "filename is required",
		})
		return
	}

	// Ищем файл в БД
	var libissFile models.LibissPosFile
	if err := database.DB.Where("file_name = ?", filename).First(&libissFile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file not found",
		})
		return
	}

	// Проверяем, активен ли файл
	if !libissFile.IsActive {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file is not active",
		})
		return
	}

	// Проверяем существование файла
	if _, err := os.Stat(libissFile.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file not found on disk",
		})
		return
	}

	// Увеличиваем счетчик скачиваний
	database.DB.Model(&libissFile).Update("download_count", libissFile.DownloadCount+1)

	// Отправляем файл
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", libissFile.OriginalName))
	c.Header("Content-Type", "application/octet-stream")
	c.File(libissFile.FilePath)
}

// PublicDownload скачивает файл публично (без аутентификации)
func (lc *LibissPosController) PublicDownload(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "filename is required",
		})
		return
	}

	// Ищем файл в БД
	var libissFile models.LibissPosFile
	if err := database.DB.Where("file_name = ? AND is_public = ?", filename, true).First(&libissFile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file not found or not available for public download",
		})
		return
	}

	// Проверяем, активен ли файл
	if !libissFile.IsActive {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file is not active",
		})
		return
	}

	// Проверяем существование файла
	if _, err := os.Stat(libissFile.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file not found on disk",
		})
		return
	}

	// Увеличиваем счетчик скачиваний
	database.DB.Model(&libissFile).Update("download_count", libissFile.DownloadCount+1)

	// Отправляем файл
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", libissFile.OriginalName))
	c.Header("Content-Type", "application/octet-stream")
	c.File(libissFile.FilePath)
}

// DeleteFile удаляет файл (только для админов)
func (lc *LibissPosController) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "file id is required",
		})
		return
	}

	// Ищем файл в БД
	var libissFile models.LibissPosFile
	if err := database.DB.Where("id = ?", fileID).First(&libissFile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file not found",
		})
		return
	}

	// Удаляем файл с диска
	if err := os.Remove(libissFile.FilePath); err != nil {
		log.Printf("⚠️ Failed to delete file from disk: %v", err)
		// Продолжаем удаление записи из БД даже если файл не найден
	}

	// Удаляем запись из БД
	if err := database.DB.Delete(&libissFile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to delete file record",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "File deleted successfully",
	})
}

// GetFileInfo возвращает информацию о файле
func (lc *LibissPosController) GetFileInfo(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "file id is required",
		})
		return
	}

	var libissFile models.LibissPosFile
	if err := database.DB.Where("id = ?", fileID).First(&libissFile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    libissFile,
	})
}

// GetLatestFile возвращает последний активный файл по типу и платформе
func (lc *LibissPosController) GetLatestFile(c *gin.Context) {
	fileTypeStr := c.Query("type")
	platformStr := c.Query("platform")
	
	if fileTypeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "type is required",
		})
		return
	}

	fileType := models.LibissPosType(fileTypeStr)
	if fileType != models.LibissPosTypeFull &&
		fileType != models.LibissPosTypeCassa2 &&
		fileType != models.LibissPosTypeServerOnly {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid type",
		})
		return
	}

	query := database.DB.Where("type = ? AND is_active = ?", fileType, true)
	
	if platformStr != "" {
		platform := models.LibissPosPlatform(platformStr)
		if platform != models.LibissPosPlatformWindows && platform != models.LibissPosPlatformAndroid {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid platform",
			})
			return
		}
		query = query.Where("platform = ?", platform)
	}

	var libissFile models.LibissPosFile
	if err := query.Order("created_at DESC").First(&libissFile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    libissFile,
	})
}

