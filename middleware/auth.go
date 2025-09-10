package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"
	"github.com/mm-api/mm-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthRequired проверяет JWT токен
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		log.Printf("🔐 AuthRequired: запрос к %s, заголовок Authorization: %s", c.Request.URL.Path, authHeader)

		if authHeader == "" {
			log.Printf("❌ AuthRequired: заголовок Authorization отсутствует для %s", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Извлекаем токен из заголовка "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Валидируем токен
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Получаем пользователя из базы данных
		var user models.User
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user ID in token",
			})
			c.Abort()
			return
		}

		if err := database.DB.Preload("Role").First(&user, "id = ? AND is_active = ?", userID, true).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found or inactive",
			})
			c.Abort()
			return
		}

		// Добавляем пользователя в контекст
		c.Set("user", user)
		c.Set("userID", user.ID)

		// Логируем информацию о пользователе для отладки
		log.Printf("🔐 Аутентификация: пользователь %s (ID: %s, роль: %s) успешно аутентифицирован",
			user.Email, user.ID, user.Role.Name)

		c.Next()
	}
}

// AdminRequired проверяет права администратора
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found in context",
			})
			c.Abort()
			return
		}

		userModel := user.(models.User)
		// Проверяем роль через связь с таблицей ролей
		if userModel.Role == nil || userModel.Role.Name != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ShopOwnerRequired проверяет права владельца магазина
func ShopOwnerRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("🏪 ShopOwnerRequired: проверка доступа для %s", c.Request.URL.Path)

		user, exists := c.Get("user")
		if !exists {
			log.Printf("❌ ShopOwnerRequired: пользователь не найден в контексте для %s", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found in context",
			})
			c.Abort()
			return
		}

		userModel := user.(models.User)
		if userModel.Role == nil || (userModel.Role.Name != "shop_owner" && userModel.Role.Name != "admin") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Shop owner access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuperAdminRequired проверяет права супер администратора
func SuperAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found in context",
			})
			c.Abort()
			return
		}

		userModel := user.(models.User)
		// Проверяем роль через связь с таблицей ролей
		if userModel.Role == nil || userModel.Role.Name != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Super admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOrShopOwnerRequired проверяет права админа или владельца магазина
func AdminOrShopOwnerRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found in context",
			})
			c.Abort()
			return
		}

		userModel := user.(models.User)
		if userModel.Role == nil || (userModel.Role.Name != "admin" && userModel.Role.Name != "shop_owner") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin or shop owner access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOrSuperAdminRequired проверяет права админа или супер админа
func AdminOrSuperAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found in context",
			})
			c.Abort()
			return
		}

		userModel := user.(models.User)
		// Проверяем роль через связь с таблицей ролей
		if userModel.Role == nil || (userModel.Role.Name != "admin" && userModel.Role.Name != "super_admin") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin or super admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser возвращает текущего пользователя из контекста
func GetCurrentUser(c *gin.Context) (models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return models.User{}, false
	}
	return user.(models.User), true
}
