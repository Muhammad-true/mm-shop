package controllers

import (
	"net/http"

	"github.com/mm-api/mm-api/config"
	"github.com/gin-gonic/gin"
)

// GetConfig возвращает текущую конфигурацию (только для разработки)
func GetConfig(c *gin.Context) {
	cfg := config.GetConfig()

	// Проверяем, что мы в режиме разработки
	if !cfg.IsDevelopment() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Configuration endpoint available only in development mode",
		})
		return
	}

	// Безопасная версия конфигурации (без секретов)
	safeConfig := gin.H{
		"server": gin.H{
			"port":     cfg.Port,
			"host":     cfg.Host,
			"gin_mode": cfg.GinMode,
			"debug":    cfg.Debug,
		},
		"database": gin.H{
			"url_masked": maskDatabaseURL(cfg.DatabaseURL),
		},
		"jwt": gin.H{
			"secret_masked": maskSecret(cfg.JWTSecret),
			"expiry":        cfg.JWTExpiry,
		},
		"logging": gin.H{
			"level":  cfg.LogLevel,
			"format": cfg.LogFormat,
		},
		"cors": gin.H{
			"allow_origins":     cfg.CORSAllowOrigins,
			"allow_methods":     cfg.CORSAllowMethods,
			"allow_headers":     cfg.CORSAllowHeaders,
			"allow_credentials": cfg.CORSAllowCredentials,
		},
		"upload": gin.H{
			"max_size":      cfg.UploadMaxSize,
			"allowed_types": cfg.UploadAllowedTypes,
			"path":          cfg.UploadPath,
		},
		"features": gin.H{
			"swagger_enabled": cfg.EnableSwagger,
			"cors_enabled":    cfg.EnableCORS,
			"metrics_enabled": cfg.EnableMetrics,
		},
		"security": gin.H{
			"bcrypt_cost":         cfg.BcryptCost,
			"rate_limit_requests": cfg.RateLimitRequests,
			"rate_limit_window":   cfg.RateLimitWindow,
		},
		"environment": gin.H{
			"is_development": cfg.IsDevelopment(),
			"is_production":  cfg.IsProduction(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    safeConfig,
		"message": "Current configuration (development mode)",
	})
}

// maskDatabaseURL маскирует пароль в URL базы данных
func maskDatabaseURL(url string) string {
	if len(url) > 30 {
		return url[:15] + "***" + url[len(url)-10:]
	}
	return "***"
}

// maskSecret маскирует секретный ключ
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}
