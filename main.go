package main

import (
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/mm-api/mm-api/config"
	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/routes"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ PANIC: %v", r)
			log.Printf("Stack trace: %s", debug.Stack())
		}
	}()

	log.Println("🚀 Starting MM API Server...")

	// Загрузка конфигурации
	log.Println("⚙️ Loading configuration...")
	cfg := config.Load()

	// Валидация конфигурации
	if err := cfg.Validate(); err != nil {
		log.Fatal("❌ Configuration validation failed:", err)
	}

	// Настройка режима Gin
	gin.SetMode(cfg.GinMode)
	log.Printf("✅ Gin mode set to: %s", cfg.GinMode)

	// Подключение к базе данных
	log.Println("🔗 Connecting to database...")
	if err := database.Connect(); err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	log.Println("✅ Database connected successfully")

	// Настройка маршрутов
	log.Println("🛣️  Setting up routes...")
	r := routes.SetupRoutes()
	log.Println("✅ Routes configured")

	log.Printf("🚀 Server starting on %s:%s", cfg.Host, cfg.Port)
	log.Printf("📖 API Documentation: http://%s:%s/", cfg.PublicHost, cfg.Port)
	log.Printf("🔗 Health check: http://%s:%s/health", cfg.PublicHost, cfg.Port)
	log.Printf("🖥️ Admin panel: http://%s:%s/admin", cfg.PublicHost, cfg.Port)

	// Запуск сервера
	log.Println("🌐 Starting HTTP server...")
	if err := r.Run(cfg.Host + ":" + cfg.Port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
