package models

import (
	"time"

	"gorm.io/gorm"
)

// ServiceLog - модель для служебных логов
type ServiceLog struct {
	ID        uint                   `json:"id" gorm:"primaryKey"`
	Level     string                 `json:"level" gorm:"type:varchar(10);not null"`
	Message   string                 `json:"message" gorm:"type:text;not null"`
	Context   map[string]interface{} `json:"context" gorm:"type:jsonb"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt gorm.DeletedAt         `json:"deleted_at,omitempty" gorm:"index"`
}

// ServiceMetric - модель для служебных метрик
type ServiceMetric struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	MetricName  string                 `json:"metric_name" gorm:"type:varchar(100);not null"`
	MetricValue float64                `json:"metric_value" gorm:"type:decimal(15,2);not null"`
	Tags        map[string]interface{} `json:"tags" gorm:"type:jsonb"`
	RecordedAt  time.Time              `json:"recorded_at"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `json:"deleted_at,omitempty" gorm:"index"`
}

// ServiceCache - модель для служебного кэша
type ServiceCache struct {
	CacheKey   string         `json:"cache_key" gorm:"type:varchar(255);primaryKey"`
	CacheValue string         `json:"cache_value" gorm:"type:text;not null"`
	ExpiresAt  *time.Time     `json:"expires_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// ServiceLogResponse - ответ для API
type ServiceLogResponse struct {
	ID        uint                   `json:"id"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context"`
	CreatedAt time.Time              `json:"created_at"`
}

// ServiceMetricResponse - ответ для API
type ServiceMetricResponse struct {
	ID          uint                   `json:"id"`
	MetricName  string                 `json:"metric_name"`
	MetricValue float64                `json:"metric_value"`
	Tags        map[string]interface{} `json:"tags"`
	RecordedAt  time.Time              `json:"recorded_at"`
}

// ServiceCacheResponse - ответ для API
type ServiceCacheResponse struct {
	CacheKey   string     `json:"cache_key"`
	CacheValue string     `json:"cache_value"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ToResponse - конвертация в ответ API
func (sl *ServiceLog) ToResponse() ServiceLogResponse {
	return ServiceLogResponse{
		ID:        sl.ID,
		Level:     sl.Level,
		Message:   sl.Message,
		Context:   sl.Context,
		CreatedAt: sl.CreatedAt,
	}
}

// ToResponse - конвертация в ответ API
func (sm *ServiceMetric) ToResponse() ServiceMetricResponse {
	return ServiceMetricResponse{
		ID:          sm.ID,
		MetricName:  sm.MetricName,
		MetricValue: sm.MetricValue,
		Tags:        sm.Tags,
		RecordedAt:  sm.RecordedAt,
	}
}

// ToResponse - конвертация в ответ API
func (sc *ServiceCache) ToResponse() ServiceCacheResponse {
	return ServiceCacheResponse{
		CacheKey:   sc.CacheKey,
		CacheValue: sc.CacheValue,
		ExpiresAt:  sc.ExpiresAt,
		CreatedAt:  sc.CreatedAt,
	}
}

// ServiceLogCreateRequest - запрос на создание лога
type ServiceLogCreateRequest struct {
	Level   string                 `json:"level" binding:"required,oneof=debug info warn error fatal"`
	Message string                 `json:"message" binding:"required"`
	Context map[string]interface{} `json:"context"`
}

// ServiceMetricCreateRequest - запрос на создание метрики
type ServiceMetricCreateRequest struct {
	MetricName  string                 `json:"metric_name" binding:"required"`
	MetricValue float64                `json:"metric_value" binding:"required"`
	Tags        map[string]interface{} `json:"tags"`
}

// ServiceCacheCreateRequest - запрос на создание кэша
type ServiceCacheCreateRequest struct {
	CacheKey   string     `json:"cache_key" binding:"required"`
	CacheValue string     `json:"cache_value" binding:"required"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// ServiceCacheUpdateRequest - запрос на обновление кэша
type ServiceCacheUpdateRequest struct {
	CacheValue string     `json:"cache_value" binding:"required"`
	ExpiresAt  *time.Time `json:"expires_at"`
}
