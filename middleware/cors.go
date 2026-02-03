package middleware

import (
	"github.com/mm-api/mm-api/config"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	// Получаем конфигурацию один раз при старте
	cfg := config.GetConfig()
	allowedOrigins := cfg.GetCORSOrigins() // Функция, которая вернет []string

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		isAllowed := false
		allowAll := false

		// Проверяем, есть ли origin в списке разрешенных
		for _, allowed := range allowedOrigins {
			if allowed == "*" {
				allowAll = true
				isAllowed = true
				break
			} else if allowed == origin {
				isAllowed = true
				break
			}
		}

		// Устанавливаем заголовок CORS
		if isAllowed {
			if allowAll {
				// Если разрешены все origins, используем конкретный origin из запроса или *
				if origin != "" {
					c.Header("Access-Control-Allow-Origin", origin)
				} else {
					c.Header("Access-Control-Allow-Origin", "*")
				}
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
