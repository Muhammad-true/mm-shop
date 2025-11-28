package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// FCMService обрабатывает отправку push-уведомлений через Firebase Cloud Messaging
type FCMService struct {
	ServerKey string
	APIURL    string
}

// FCMRequest представляет запрос к FCM API
type FCMRequest struct {
	To           string                 `json:"to,omitempty"`
	RegistrationIDs []string            `json:"registration_ids,omitempty"`
	Notification FCMNotification        `json:"notification"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Priority     string                 `json:"priority,omitempty"`
}

// FCMNotification представляет уведомление FCM
type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
	Sound string `json:"sound,omitempty"`
}

// FCMResponse представляет ответ от FCM API
type FCMResponse struct {
	Success int `json:"success"`
	Failure int `json:"failure"`
	Results []struct {
		MessageID string `json:"message_id,omitempty"`
		Error     string `json:"error,omitempty"`
	} `json:"results,omitempty"`
}

var fcmService *FCMService

// InitFCMService инициализирует FCM сервис
func InitFCMService(serverKey string) {
	fcmService = &FCMService{
		ServerKey: serverKey,
		APIURL:    "https://fcm.googleapis.com/fcm/send",
	}
	log.Println("✅ FCM Service инициализирован")
}

// GetFCMService возвращает экземпляр FCM сервиса
func GetFCMService() *FCMService {
	return fcmService
}

// SendPushNotification отправляет push-уведомление на одно устройство
func (fcm *FCMService) SendPushNotification(token, title, body, actionURL string) error {
	if fcm.ServerKey == "" {
		log.Println("⚠️ FCM Server Key не настроен, пропускаем отправку push-уведомления")
		return nil
	}

	req := FCMRequest{
		To: token,
		Notification: FCMNotification{
			Title: title,
			Body:  body,
			Sound: "default",
		},
		Data: map[string]interface{}{
			"action_url": actionURL,
			"click_action": actionURL,
		},
		Priority: "high",
	}

	return fcm.sendRequest(req)
}

// SendPushNotificationToMultiple отправляет push-уведомление на несколько устройств
func (fcm *FCMService) SendPushNotificationToMultiple(tokens []string, title, body, actionURL string) error {
	if fcm.ServerKey == "" {
		log.Println("⚠️ FCM Server Key не настроен, пропускаем отправку push-уведомления")
		return nil
	}

	if len(tokens) == 0 {
		return nil
	}

	req := FCMRequest{
		RegistrationIDs: tokens,
		Notification: FCMNotification{
			Title: title,
			Body:  body,
			Sound: "default",
		},
		Data: map[string]interface{}{
			"action_url": actionURL,
			"click_action": actionURL,
		},
		Priority: "high",
	}

	return fcm.sendRequest(req)
}

// sendRequest отправляет запрос к FCM API
func (fcm *FCMService) sendRequest(req FCMRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга FCM запроса: %w", err)
	}

	httpReq, err := http.NewRequest("POST", fcm.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP запроса: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "key="+fcm.ServerKey)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса к FCM: %w", err)
	}
	defer resp.Body.Close()

	var fcmResp FCMResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmResp); err != nil {
		log.Printf("⚠️ Ошибка декодирования ответа FCM: %v", err)
		return nil // Не критично, продолжаем
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ FCM вернул статус %d: успешно %d, ошибок %d", resp.StatusCode, fcmResp.Success, fcmResp.Failure)
		return nil
	}

	log.Printf("✅ FCM: успешно отправлено %d, ошибок %d", fcmResp.Success, fcmResp.Failure)
	return nil
}

