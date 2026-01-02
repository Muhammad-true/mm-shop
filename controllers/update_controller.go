package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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
	platformStr := c.PostForm("platform")
	version := strings.TrimSpace(c.PostForm("version"))
	releaseNotes := c.PostForm("releaseNotes")

	if platformStr == "" || version == "" {
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
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to create updates directory",
			"details": err.Error(),
		})
		return
	}

	filename := fmt.Sprintf("%s_%s_%s%s", platform, version, uuid.NewString(), ext)
	filePath := filepath.Join(dir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to create file",
			"details": err.Error(),
		})
		return
	}
	defer dst.Close()

	hasher := sha256.New()
	size, err := io.Copy(io.MultiWriter(dst, hasher), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to save file",
			"details": err.Error(),
		})
		return
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	fileURL := fmt.Sprintf("/updates/%s/%s", platform, filename)

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to save update metadata",
			"details": err.Error(),
		})
		return
	}

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
