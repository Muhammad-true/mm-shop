package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Config конфигурация скрипта
type Config struct {
	FTPWatchDir    string // Папка для мониторинга (например: /var/ftp/uploads)
	APIBaseURL     string // URL API (например: https://api.libiss.com/api/v1)
	APIToken       string // Токен для авторизации
	ProcessedDir   string // Папка для обработанных файлов
	CheckInterval  int    // Интервал проверки в секундах (по умолчанию 30)
}

// UpdateInfo информация об обновлении из имени файла
type UpdateInfo struct {
	Platform     string
	Version      string
	ReleaseNotes string
}

// Парсинг имени файла для определения платформы и версии
// Поддерживаемые форматы:
// 1. android_1.0.0.apk
// 2. windows_1.2.0.exe
// 3. server_2.0.0.zip
// 4. android-v1.0.0.apk
// 5. android_1.0.0_release.apk
// 6. app-android-1.0.0.apk
func parseFileName(filename string) (*UpdateInfo, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Определяем платформу по расширению
	var platform string
	switch ext {
	case ".apk":
		platform = "android"
	case ".exe":
		platform = "windows"
	case ".zip":
		platform = "server"
	default:
		return nil, fmt.Errorf("неподдерживаемое расширение: %s", ext)
	}

	// Пытаемся извлечь версию из имени файла
	// Паттерны для поиска версии: 1.0.0, 1.2.3, v1.0.0, версия_1.0.0
	versionPattern := regexp.MustCompile(`(?i)(?:v|version[_-]?)?(\d+\.\d+\.\d+(?:\.\d+)?)`)
	matches := versionPattern.FindStringSubmatch(nameWithoutExt)

	if len(matches) < 2 {
		// Если версия не найдена, пытаемся извлечь из стандартного формата
		// android_1.0.0 или android-1.0.0
		parts := regexp.MustCompile(`[_-]`).Split(nameWithoutExt, -1)
		for i, part := range parts {
			if part == platform && i+1 < len(parts) {
				// Проверяем, является ли следующий элемент версией
				if versionPattern.MatchString(parts[i+1]) {
					matches = versionPattern.FindStringSubmatch(parts[i+1])
					break
				}
			}
		}
	}

	if len(matches) < 2 {
		return nil, fmt.Errorf("не удалось определить версию из имени файла: %s", filename)
	}

	version := matches[1]

	return &UpdateInfo{
		Platform:     platform,
		Version:      version,
		ReleaseNotes: fmt.Sprintf("Автоматическая загрузка через FTP: %s", filename),
	}, nil
}

// Вычисляет SHA256 хеш файла
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Загружает файл через API
func uploadToAPI(config Config, filePath string, info *UpdateInfo) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	// Получаем размер файла
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("ошибка получения информации о файле: %v", err)
	}

	// Создаем multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Добавляем поля формы
	writer.WriteField("platform", info.Platform)
	writer.WriteField("version", info.Version)
	writer.WriteField("releaseNotes", info.ReleaseNotes)

	// Добавляем файл
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("ошибка создания поля файла: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("ошибка копирования файла: %v", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия writer: %v", err)
	}

	// Создаем HTTP запрос
	url := fmt.Sprintf("%s/admin/updates/upload", config.APIBaseURL)
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIToken))

	// Выполняем запрос
	client := &http.Client{
		Timeout: 30 * time.Minute, // Таймаут для больших файлов
	}

	log.Printf("📤 Загрузка файла %s (%.2f MB) на сервер...", filepath.Base(filePath), float64(fileInfo.Size())/1024/1024)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return fmt.Errorf("ошибка API (код %d): %v", resp.StatusCode, errorResp)
		}
		return fmt.Errorf("ошибка API (код %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	log.Printf("✅ Файл успешно загружен: %s", filepath.Base(filePath))
	if data, ok := result["data"].(map[string]interface{}); ok {
		if fileName, ok := data["fileName"].(string); ok {
			log.Printf("   Имя файла на сервере: %s", fileName)
		}
	}

	return nil
}

// Перемещает файл в папку обработанных
func moveToProcessed(config Config, filePath string) error {
	processedPath := filepath.Join(config.ProcessedDir, filepath.Base(filePath))
	
	// Создаем папку, если её нет
	if err := os.MkdirAll(config.ProcessedDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания папки обработанных файлов: %v", err)
	}

	// Переименовываем файл с timestamp для избежания конфликтов
	timestamp := time.Now().Format("20060102_150405")
	ext := filepath.Ext(processedPath)
	nameWithoutExt := strings.TrimSuffix(processedPath, ext)
	processedPath = fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)

	if err := os.Rename(filePath, processedPath); err != nil {
		return fmt.Errorf("ошибка перемещения файла: %v", err)
	}

	log.Printf("📦 Файл перемещен в обработанные: %s", filepath.Base(processedPath))
	return nil
}

// Обрабатывает один файл
func processFile(config Config, filePath string) error {
	filename := filepath.Base(filePath)
	log.Printf("🔍 Обработка файла: %s", filename)

	// Парсим имя файла
	info, err := parseFileName(filename)
	if err != nil {
		return fmt.Errorf("ошибка парсинга имени файла: %v", err)
	}

	log.Printf("   Платформа: %s, Версия: %s", info.Platform, info.Version)

	// Загружаем через API
	if err := uploadToAPI(config, filePath, info); err != nil {
		return fmt.Errorf("ошибка загрузки через API: %v", err)
	}

	// Перемещаем в папку обработанных
	if err := moveToProcessed(config, filePath); err != nil {
		return fmt.Errorf("ошибка перемещения файла: %v", err)
	}

	return nil
}

// Сканирует папку и обрабатывает новые файлы
func scanAndProcess(config Config) error {
	entries, err := os.ReadDir(config.FTPWatchDir)
	if err != nil {
		return fmt.Errorf("ошибка чтения папки: %v", err)
	}

	processed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		// Проверяем расширение
		if ext != ".apk" && ext != ".exe" && ext != ".zip" {
			continue
		}

		filePath := filepath.Join(config.FTPWatchDir, filename)

		// Обрабатываем файл
		if err := processFile(config, filePath); err != nil {
			log.Printf("❌ Ошибка обработки файла %s: %v", filename, err)
			continue
		}

		processed++
	}

	if processed > 0 {
		log.Printf("✅ Обработано файлов: %d", processed)
	}

	return nil
}

func main() {
	// Конфигурация из переменных окружения
	config := Config{
		FTPWatchDir:   getEnv("FTP_WATCH_DIR", "/var/ftp/uploads"),
		APIBaseURL:    getEnv("API_BASE_URL", "https://api.libiss.com/api/v1"),
		APIToken:      getEnv("API_TOKEN", ""),
		ProcessedDir:  getEnv("PROCESSED_DIR", "/var/ftp/processed"),
		CheckInterval: getEnvInt("CHECK_INTERVAL", 30),
	}

	// Проверяем обязательные параметры
	if config.APIToken == "" {
		log.Fatal("❌ API_TOKEN не установлен. Установите переменную окружения API_TOKEN")
	}

	// Создаем папки, если их нет
	if err := os.MkdirAll(config.FTPWatchDir, 0755); err != nil {
		log.Fatalf("❌ Ошибка создания папки для мониторинга: %v", err)
	}

	log.Printf("🚀 Запуск мониторинга папки: %s", config.FTPWatchDir)
	log.Printf("   API URL: %s", config.APIBaseURL)
	log.Printf("   Интервал проверки: %d секунд", config.CheckInterval)

	// Бесконечный цикл проверки
	ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Second)
	defer ticker.Stop()

	// Первая проверка сразу
	if err := scanAndProcess(config); err != nil {
		log.Printf("❌ Ошибка при первой проверке: %v", err)
	}

	// Периодическая проверка
	for range ticker.C {
		if err := scanAndProcess(config); err != nil {
			log.Printf("❌ Ошибка при проверке: %v", err)
		}
	}
}

// Вспомогательные функции
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

