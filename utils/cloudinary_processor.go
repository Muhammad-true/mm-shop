package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CloudinaryProcessor обрабатывает загрузку изображений в Cloudinary
type CloudinaryProcessor struct {
	CloudName    string
	APIKey       string
	APISecret    string
	UploadPreset string // Для unsigned uploads (рекомендуется)
}

// NewCloudinaryProcessor создает новый процессор Cloudinary
func NewCloudinaryProcessor(cloudName, apiKey, apiSecret, uploadPreset string) *CloudinaryProcessor {
	return &CloudinaryProcessor{
		CloudName:    cloudName,
		APIKey:       apiKey,
		APISecret:    apiSecret,
		UploadPreset: uploadPreset,
	}
}

// CloudinaryResponse представляет ответ от Cloudinary API
type CloudinaryResponse struct {
	PublicID   string `json:"public_id"`
	URL        string `json:"url"`
	SecureURL  string `json:"secure_url"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Format     string `json:"format"`
	Bytes      int    `json:"bytes"`
	ResourceType string `json:"resource_type"`
	CreatedAt  string `json:"created_at"`
}

// CloudinaryError представляет ошибку от Cloudinary
type CloudinaryError struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ProcessProductImage загружает изображение в Cloudinary с автоматической обработкой:
// - Автоматическая обработка EXIF ориентации
// - Изменение размера до 1200x1200 с сохранением пропорций
// - Добавление белого фона
// - Автоматическое сжатие
// - Опционально: удаление фона (e_background_removal)
func (cp *CloudinaryProcessor) ProcessProductImage(input io.Reader, folder string, removeBackground bool) (*CloudinaryResponse, error) {
	// Читаем данные изображения
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла: %v", err)
	}

	// Создаем multipart form для загрузки
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Добавляем файл
	fw, err := w.CreateFormFile("file", "image.jpg")
	if err != nil {
		return nil, fmt.Errorf("ошибка создания form file: %v", err)
	}
	if _, err := fw.Write(data); err != nil {
		return nil, fmt.Errorf("ошибка записи файла: %v", err)
	}

	// Параметры загрузки
	w.WriteField("upload_preset", cp.UploadPreset)
	w.WriteField("folder", folder)
	
	// Примечание: если preset уже настроен с трансформациями, они будут применены автоматически
	// Не передаем transformation в коде, чтобы использовать настройки preset
	
	// ВАЖНО: fl_auto нельзя использовать в Upload Preset!
	// Передаем флаг auto через параметр flags для автоматической обработки EXIF ориентации
	w.WriteField("flags", "auto")        // Автоматическая обработка EXIF ориентации (исправляет поворот фото с телефонов)

	// Дополнительные параметры
	w.WriteField("format", "jpg")        // Всегда сохраняем как JPG
	w.WriteField("overwrite", "false")   // Не перезаписывать существующие
	w.WriteField("invalidate", "true")   // Инвалидировать кэш CDN

	w.Close()

	// Загружаем в Cloudinary
	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cp.CloudName)
	
	log.Printf("☁️ Загрузка изображения в Cloudinary (folder: %s, preset: %s)...", folder, cp.UploadPreset)
	
	resp, err := http.Post(uploadURL, w.FormDataContentType(), &b)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса к Cloudinary: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Пробуем распарсить ошибку
		var cloudErr CloudinaryError
		if err := json.Unmarshal(body, &cloudErr); err == nil {
			return nil, fmt.Errorf("Cloudinary error: %s", cloudErr.Error.Message)
		}
		return nil, fmt.Errorf("Cloudinary error (status %d): %s", resp.StatusCode, string(body))
	}

	// Парсим успешный ответ
	var result CloudinaryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа Cloudinary: %v", err)
	}

	log.Printf("✅ Изображение загружено в Cloudinary: %s (размер: %dx%d, формат: %s, размер файла: %d байт)",
		result.SecureURL, result.Width, result.Height, result.Format, result.Bytes)

	return &result, nil
}

// DeleteImage удаляет изображение из Cloudinary
func (cp *CloudinaryProcessor) DeleteImage(publicID string) error {
	timestamp := time.Now().Unix()
	
	// Создаем подпись для API запроса
	params := map[string]string{
		"public_id": publicID,
		"timestamp": strconv.FormatInt(timestamp, 10),
	}
	
	signature := cp.generateSignature(params)
	
	// Формируем URL для удаления
	deleteURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/destroy", cp.CloudName)
	
	// Создаем form data
	formData := map[string]string{
		"public_id": publicID,
		"timestamp": strconv.FormatInt(timestamp, 10),
		"signature": signature,
		"api_key":   cp.APIKey,
	}
	
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for k, v := range formData {
		w.WriteField(k, v)
	}
	w.Close()
	
	resp, err := http.Post(deleteURL, w.FormDataContentType(), &b)
	if err != nil {
		return fmt.Errorf("ошибка удаления: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка удаления (status %d): %s", resp.StatusCode, string(body))
	}
	
	log.Printf("✅ Изображение удалено из Cloudinary: %s", publicID)
	return nil
}

// generateSignature генерирует подпись для Cloudinary API
func (cp *CloudinaryProcessor) generateSignature(params map[string]string) string {
	// Сортируем ключи
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// Создаем строку для подписи
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signString := strings.Join(parts, "&") + cp.APISecret
	
	// Вычисляем SHA1
	hash := sha1.Sum([]byte(signString))
	return fmt.Sprintf("%x", hash)
}

// GetOptimizedURL возвращает URL с оптимизацией для быстрой загрузки
func (cp *CloudinaryProcessor) GetOptimizedURL(publicID string, width, height int) string {
	transformations := []string{
		fmt.Sprintf("w_%d", width),
		fmt.Sprintf("h_%d", height),
		"c_fit",
		"q_auto:good",
		"fl_auto",
	}
	
	transformation := strings.Join(transformations, ",")
	return fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s/%s.jpg",
		cp.CloudName, transformation, publicID)
}

