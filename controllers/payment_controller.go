package controllers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/mm-api/mm-api/config"
	"github.com/mm-api/mm-api/database"
	"github.com/mm-api/mm-api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentController обрабатывает запросы платежей
type PaymentController struct{}

// LemonSqueezyCheckoutRequest запрос на создание checkout в Lemon Squeezy
type LemonSqueezyCheckoutRequest struct {
	ShopID            string `json:"shopId" binding:"required"`
	SubscriptionPlanID string `json:"subscriptionPlanId" binding:"required"`
}

// LemonSqueezyCheckoutResponse ответ с URL checkout
type LemonSqueezyCheckoutResponse struct {
	Success    bool   `json:"success"`
	CheckoutURL string `json:"checkoutUrl"`
	Message    string `json:"message,omitempty"`
}

// CreateLemonSqueezyCheckout создает checkout в Lemon Squeezy
func (pc *PaymentController) CreateLemonSqueezyCheckout(c *gin.Context) {
	var req LemonSqueezyCheckoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Парсим ShopID
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid shop ID",
		})
		return
	}

	// Проверяем существование магазина
	var shop models.Shop
	if err := database.DB.First(&shop, shopID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Shop not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Парсим SubscriptionPlanID
	planID, err := uuid.Parse(req.SubscriptionPlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid subscription plan ID",
		})
		return
	}

	// Получаем план подписки
	var plan models.SubscriptionPlan
	if err := database.DB.First(&plan, planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Subscription plan not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Проверяем наличие lemonsqueezyVariantId
	if plan.LemonSqueezyVariantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Subscription plan does not have Lemon Squeezy variant ID configured",
		})
		return
	}

	// Получаем конфигурацию
	cfg := config.GetConfig()
	if cfg.LemonSqueezyAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Lemon Squeezy API key not configured",
		})
		return
	}

	if cfg.LemonSqueezyStoreID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Lemon Squeezy Store ID not configured",
		})
		return
	}

	// Формируем запрос к Lemon Squeezy API
	checkoutData := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "checkouts",
			"attributes": map[string]interface{}{
				"custom_price": nil,
				"product_options": map[string]interface{}{
					"name":        plan.Name,
					"description": plan.Description,
				},
				"checkout_options": map[string]interface{}{
					"embed": false,
					"media": false,
					"logo":  false,
				},
				"checkout_data": map[string]interface{}{
					"custom": map[string]interface{}{
						"shop_id": shopID.String(),
					},
				},
				"expires_at": nil,
				"preview":    false,
			},
			"relationships": map[string]interface{}{
				"store": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "stores",
						"id":   cfg.LemonSqueezyStoreID,
					},
				},
				"variant": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "variants",
						"id":   plan.LemonSqueezyVariantID,
					},
				},
			},
		},
	}

	// Преобразуем в JSON
	jsonData, err := json.Marshal(checkoutData)
	if err != nil {
		log.Printf("❌ [LemonSqueezyCheckout] Ошибка маршалинга JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to prepare checkout request",
		})
		return
	}

	// Создаем HTTP запрос к Lemon Squeezy API
	apiURL := "https://api.lemonsqueezy.com/v1/checkouts"
	reqHTTP, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ [LemonSqueezyCheckout] Ошибка создания HTTP запроса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create checkout request",
		})
		return
	}

	// Устанавливаем заголовки
	reqHTTP.Header.Set("Authorization", "Bearer "+cfg.LemonSqueezyAPIKey)
	reqHTTP.Header.Set("Accept", "application/vnd.api+json")
	reqHTTP.Header.Set("Content-Type", "application/vnd.api+json")

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		log.Printf("❌ [LemonSqueezyCheckout] Ошибка выполнения запроса к Lemon Squeezy: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to communicate with Lemon Squeezy API",
		})
		return
	}
	defer resp.Body.Close()

	// Читаем ответ
	var lemonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&lemonResponse); err != nil {
		log.Printf("❌ [LemonSqueezyCheckout] Ошибка парсинга ответа от Lemon Squeezy: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to parse Lemon Squeezy response",
		})
		return
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Printf("❌ [LemonSqueezyCheckout] Lemon Squeezy вернул ошибку: %d, %+v", resp.StatusCode, lemonResponse)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Lemon Squeezy API error",
			"details": lemonResponse,
		})
		return
	}

	// Извлекаем checkout URL из ответа
	var checkoutURL string
	if data, ok := lemonResponse["data"].(map[string]interface{}); ok {
		if attributes, ok := data["attributes"].(map[string]interface{}); ok {
			if url, ok := attributes["url"].(string); ok {
				checkoutURL = url
			}
		}
	}

	if checkoutURL == "" {
		log.Printf("❌ [LemonSqueezyCheckout] Checkout URL не найден в ответе: %+v", lemonResponse)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Checkout URL not found in response",
		})
		return
	}

	log.Printf("✅ [LemonSqueezyCheckout] Checkout создан успешно для магазина %s: %s", shopID, checkoutURL)

	c.JSON(http.StatusOK, LemonSqueezyCheckoutResponse{
		Success:     true,
		CheckoutURL: checkoutURL,
		Message:     "Checkout created successfully",
	})
}

