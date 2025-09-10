package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config структура для хранения конфигурации приложения
type Config struct {
	// Сервер
	Port       string
	Host       string
	PublicHost string
	GinMode    string
	Debug      bool

	// База данных
	DatabaseURL string

	// JWT
	JWTSecret string
	JWTExpiry string

	// Логирование
	LogLevel  string
	LogFormat string

	// CORS
	CORSAllowOrigins     string
	CORSAllowMethods     string
	CORSAllowHeaders     string
	CORSExposeHeaders    string
	CORSAllowCredentials bool

	// Загрузка файлов
	UploadMaxSize      string
	UploadAllowedTypes string
	UploadPath         string

	// Email
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string

	// Redis
	RedisURL string

	// Мониторинг
	EnableMetrics bool
	MetricsPort   string

	// Безопасность
	BcryptCost        int
	RateLimitRequests int
	RateLimitWindow   string

	// Функциональность
	EnableSwagger bool
	EnableCORS    bool
}

var AppConfig *Config

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	// Загружаем env.development файл если он существует
	if err := godotenv.Load("env.development"); err != nil {
		log.Println("📝 env.development файл не найден, используем переменные окружения")
	} else {
		log.Println("✅ env.development файл загружен")
	}

	config := &Config{
		// Сервер
		Port:       getEnv("PORT", "8080"),
		Host:       getEnv("HOST", "0.0.0.0"),
		PublicHost: getEnv("PUBLIC_HOST", "localhost"),
		GinMode:    getEnv("GIN_MODE", "debug"),
		Debug:      getBoolEnv("DEBUG", true),

		// База данных
		DatabaseURL: getEnv("DATABASE_URL", "postgres://mm_user:muhammadjon@postgres:5432/mm_shop_prod?sslmode=disable"),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "moya-super-secretnaya-fraza-dlya-api-777"),
		JWTExpiry: getEnv("JWT_EXPIRY", "24h"),

		// Логирование
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		// CORS
		CORSAllowOrigins:     getEnv("CORS_ALLOW_ORIGINS", "*"),
		CORSAllowMethods:     getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
		CORSAllowHeaders:     getEnv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Authorization"),
		CORSExposeHeaders:    getEnv("CORS_EXPOSE_HEADERS", "Content-Length"),
		CORSAllowCredentials: getBoolEnv("CORS_ALLOW_CREDENTIALS", true),

		// Загрузка файлов
		UploadMaxSize:      getEnv("UPLOAD_MAX_SIZE", "50MB"),
		UploadAllowedTypes: getEnv("UPLOAD_ALLOWED_TYPES", "image/jpeg,image/png,image/webp"),
		UploadPath:         getEnv("UPLOAD_PATH", "./images"),

		// Email
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		// Мониторинг
		EnableMetrics: getBoolEnv("ENABLE_METRICS", true),
		MetricsPort:   getEnv("METRICS_PORT", "9090"),

		// Безопасность
		BcryptCost:        getIntEnv("BCRYPT_COST", 12),
		RateLimitRequests: getIntEnv("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnv("RATE_LIMIT_WINDOW", "1m"),

		// Функциональность
		EnableSwagger: getBoolEnv("ENABLE_SWAGGER", true),
		EnableCORS:    getBoolEnv("ENABLE_CORS", true),
	}

	// Сохраняем глобальную конфигурацию
	AppConfig = config

	// Выводим основную информацию о конфигурации
	log.Printf("🔧 Конфигурация загружена:")
	log.Printf("  📡 Сервер: %s:%s", config.Host, config.Port)
	log.Printf("  🌐 Публичный URL: %s:%s", config.PublicHost, config.Port)
	log.Printf("  🗄️ База данных: %s", maskDatabaseURL(config.DatabaseURL))
	log.Printf("  🔑 JWT секрет: %s", maskSecret(config.JWTSecret))
	log.Printf("  📝 Логирование: %s (%s)", config.LogLevel, config.LogFormat)
	log.Printf("  🔒 Gin режим: %s", config.GinMode)
	log.Printf("  🐛 Отладка: %t", config.Debug)

	return config
}

// GetConfig возвращает глобальную конфигурацию
func GetConfig() *Config {
	if AppConfig == nil {
		return Load()
	}
	return AppConfig
}

// getEnv получает переменную окружения со значением по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv получает булевую переменную окружения
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getIntEnv получает целочисленную переменную окружения
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// maskDatabaseURL маскирует пароль в URL базы данных
func maskDatabaseURL(url string) string {
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) >= 2 {
			beforeAt := parts[0]
			afterAt := strings.Join(parts[1:], "@")

			if strings.Contains(beforeAt, ":") {
				userParts := strings.Split(beforeAt, ":")
				if len(userParts) >= 3 {
					// ${DB_USER}://user:password -> ${DB_USER}://user:***
					userParts[len(userParts)-1] = "***"
					beforeAt = strings.Join(userParts, ":")
				}
			}
			return beforeAt + "@" + afterAt
		}
	}
	return url
}

// maskSecret маскирует секретный ключ
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}

// IsDevelopment проверяет, находится ли приложение в режиме разработки
func (c *Config) IsDevelopment() bool {
	return c.GinMode == "debug" || c.Debug
}

// IsProduction проверяет, находится ли приложение в режиме продакшена
func (c *Config) IsProduction() bool {
	return c.GinMode == "release" && !c.Debug
}

// GetCORSOrigins возвращает список разрешенных CORS origins
func (c *Config) GetCORSOrigins() []string {
	if c.CORSAllowOrigins == "*" {
		return []string{"*"}
	}
	return strings.Split(c.CORSAllowOrigins, ",")
}

// GetCORSMethods возвращает список разрешенных CORS методов
func (c *Config) GetCORSMethods() []string {
	return strings.Split(c.CORSAllowMethods, ",")
}

// GetCORSHeaders возвращает список разрешенных CORS заголовков
func (c *Config) GetCORSHeaders() []string {
	return strings.Split(c.CORSAllowHeaders, ",")
}

// GetUploadAllowedTypes возвращает список разрешенных типов файлов
func (c *Config) GetUploadAllowedTypes() []string {
	return strings.Split(c.UploadAllowedTypes, ",")
}

// Validate валидирует конфигурацию
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL не может быть пустым")
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET не может быть пустым")
	}

	if c.Port == "" {
		return fmt.Errorf("PORT не может быть пустым")
	}

	return nil
}
