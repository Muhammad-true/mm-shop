package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT секретный ключ (в продакшене должен быть в переменных окружения)
var jwtSecret = []byte("your-secret-key-here-make-it-stronger")

// Claims представляет структуру JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT создает новый JWT токен
func GenerateJWT(userID uuid.UUID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 часа
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mm-api",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT проверяет и парсит JWT токен
func ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshJWT создает новый токен на основе старого
func RefreshJWT(tokenString string) (string, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}

	// Проверяем, что токен не слишком старый для обновления (например, не старше 7 дней)
	if time.Since(claims.IssuedAt.Time) > 7*24*time.Hour {
		return "", errors.New("token too old to refresh")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", err
	}

	return GenerateJWT(userID, claims.Email, claims.Role)
}

// NormalizePhone нормализует номер телефона (убирает пробелы, дефисы и другие символы)
// Оставляет только цифры и знак + в начале
func NormalizePhone(phone string) string {
	if phone == "" {
		return ""
	}
	
	// Убираем все пробелы, дефисы, скобки и другие символы кроме цифр и +
	result := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			result += string(char)
		} else if char == '+' && len(result) == 0 {
			result += string(char)
		}
	}
	
	return result
}