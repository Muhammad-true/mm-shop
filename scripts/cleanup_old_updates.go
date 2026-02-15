//go:build cleanup
// +build cleanup

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Config конфигурация скрипта
type Config struct {
	DaysOld    int    // Количество дней
	UpdatesDir string // Папка с обновлениями
}

func main() {
	// Параметры по умолчанию
	daysOld := 10
	updatesDir := "./updates"

	// Можно передать параметры через переменные окружения
	if daysEnv := os.Getenv("DAYS_OLD"); daysEnv != "" {
		fmt.Sscanf(daysEnv, "%d", &daysOld)
	}
	if dirEnv := os.Getenv("UPDATES_DIR"); dirEnv != "" {
		updatesDir = dirEnv
	}

	// Или через аргументы командной строки
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &daysOld)
	}
	if len(os.Args) > 2 {
		updatesDir = os.Args[2]
	}

	config := Config{
		DaysOld:    daysOld,
		UpdatesDir: updatesDir,
	}

	log.Printf("🧹 Очистка старых обновлений")
	log.Printf("   Папка: %s", config.UpdatesDir)
	log.Printf("   Удаляем файлы старше: %d дней", config.DaysOld)
	log.Println("")

	// Проверка существования папки
	if _, err := os.Stat(config.UpdatesDir); os.IsNotExist(err) {
		log.Fatalf("❌ Папка не найдена: %s", config.UpdatesDir)
	}

	// Вычисляем время отсечки
	cutoffTime := time.Now().AddDate(0, 0, -config.DaysOld)

	var totalSize int64
	var deletedCount int

	// Обрабатываем все подпапки (android, windows, server)
	err := filepath.Walk(config.UpdatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории
		if info.IsDir() {
			return nil
		}

		// Проверяем, старше ли файл указанного времени
		if info.ModTime().Before(cutoffTime) {
			fileSize := info.Size()
			relativePath, _ := filepath.Rel(config.UpdatesDir, path)

			log.Printf("   🗑️  Удаление: %s (%s)", relativePath, formatSize(fileSize))

			if err := os.Remove(path); err != nil {
				log.Printf("   ❌ Ошибка удаления %s: %v", relativePath, err)
				return nil // Продолжаем обработку других файлов
			}

			totalSize += fileSize
			deletedCount++
		}

		return nil
	})

	if err != nil {
		log.Fatalf("❌ Ошибка при обходе папки: %v", err)
	}

	// Выводим результат
	log.Println("")
	if deletedCount == 0 {
		log.Println("✅ Старых файлов не найдено")
	} else {
		log.Printf("✅ Очистка завершена:")
		log.Printf("   Удалено файлов: %d", deletedCount)
		log.Printf("   Освобождено места: %s", formatSize(totalSize))
	}
}

// formatSize форматирует размер файла в читаемый вид
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

