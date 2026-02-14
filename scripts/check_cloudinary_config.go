package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔍 Проверка конфигурации Cloudinary...")
	fmt.Println()

	// Загружаем env.development
	if err := godotenv.Load("env.development"); err != nil {
		log.Println("⚠️ env.development файл не найден, проверяем переменные окружения")
	}

	useCloudinary := os.Getenv("USE_CLOUDINARY") == "true"
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	uploadPreset := os.Getenv("CLOUDINARY_UPLOAD_PRESET")
	removeBackground := os.Getenv("CLOUDINARY_REMOVE_BACKGROUND") == "true"

	fmt.Println("📋 Текущая конфигурация:")
	fmt.Printf("   USE_CLOUDINARY: %v\n", useCloudinary)
	fmt.Printf("   CLOUDINARY_CLOUD_NAME: %s\n", maskString(cloudName))
	fmt.Printf("   CLOUDINARY_API_KEY: %s\n", maskString(apiKey))
	fmt.Printf("   CLOUDINARY_API_SECRET: %s\n", maskString(apiSecret))
	fmt.Printf("   CLOUDINARY_UPLOAD_PRESET: %s\n", uploadPreset)
	fmt.Printf("   CLOUDINARY_REMOVE_BACKGROUND: %v\n", removeBackground)
	fmt.Println()

	// Проверка
	allOK := true

	if !useCloudinary {
		fmt.Println("❌ Cloudinary отключен (USE_CLOUDINARY=false)")
		fmt.Println("   → Изображения будут обрабатываться локально")
		allOK = false
	} else {
		fmt.Println("✅ Cloudinary включен")
	}

	if cloudName == "" {
		fmt.Println("❌ CLOUDINARY_CLOUD_NAME не настроен")
		allOK = false
	} else {
		fmt.Printf("✅ Cloud Name: %s\n", cloudName)
	}

	if apiKey == "" {
		fmt.Println("❌ CLOUDINARY_API_KEY не настроен")
		allOK = false
	} else {
		fmt.Println("✅ API Key настроен")
	}

	if apiSecret == "" {
		fmt.Println("❌ CLOUDINARY_API_SECRET не настроен")
		allOK = false
	} else {
		fmt.Println("✅ API Secret настроен")
	}

	if uploadPreset == "" {
		fmt.Println("❌ CLOUDINARY_UPLOAD_PRESET не настроен")
		allOK = false
	} else {
		fmt.Printf("✅ Upload Preset: %s\n", uploadPreset)
	}

	if removeBackground {
		fmt.Println("✅ Удаление фона включено")
		fmt.Println("   ⚠️  ВАЖНО: Убедитесь, что в Upload Preset настроена трансформация:")
		fmt.Println("      e_background_removal:fineedges_y")
		fmt.Println("   → Проверьте настройки preset в Cloudinary Dashboard:")
		fmt.Printf("      https://console.cloudinary.com/settings/upload_presets/%s\n", uploadPreset)
	} else {
		fmt.Println("ℹ️  Удаление фона отключено")
		fmt.Println("   → Фон НЕ будет удаляться")
	}

	fmt.Println()
	fmt.Println("📝 Инструкции:")
	fmt.Println("   1. Проверьте, что все переменные настроены в env.development")
	fmt.Println("   2. Если удаление фона включено, проверьте настройки Upload Preset:")
	fmt.Println("      - Откройте Cloudinary Dashboard")
	fmt.Println("      - Settings → Upload → Upload Presets")
	fmt.Println("      - Найдите ваш preset и проверьте 'Incoming Transformation'")
	fmt.Println("      - Должна быть цепочка с e_background_removal:fineedges_y")
	fmt.Println("   3. Перезапустите сервер после изменения конфигурации")
	fmt.Println()

	if allOK && useCloudinary {
		fmt.Println("✅ Конфигурация Cloudinary корректна!")
		fmt.Println("   → Изображения будут загружаться в Cloudinary")
		if removeBackground {
			fmt.Println("   → Удаление фона включено (проверьте preset!)")
		}
	} else {
		fmt.Println("⚠️  Есть проблемы с конфигурацией")
		fmt.Println("   → Исправьте ошибки выше и перезапустите сервер")
	}
}

func maskString(s string) string {
	if s == "" {
		return "(не настроен)"
	}
	if len(s) <= 4 {
		return "***"
	}
	return s[:4] + "***" + s[len(s)-4:]
}

