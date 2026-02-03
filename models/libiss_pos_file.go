package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LibissPosType тип программы libiss_pos
type LibissPosType string

const (
	LibissPosTypeFull       LibissPosType = "full"        // Полный пакет (программа + сервер + mysql) для касса1
	LibissPosTypeCassa2     LibissPosType = "cassa2"      // Программа для касса2 - только Windows программа
	LibissPosTypeServerOnly LibissPosType = "server_only" // Программа + сервер без mysql (если у касса1 уже установлен mysql)
)

// LibissPosPlatform платформа программы
type LibissPosPlatform string

const (
	LibissPosPlatformWindows LibissPosPlatform = "windows" // Windows (.exe)
	LibissPosPlatformAndroid LibissPosPlatform = "android" // Android (.apk)
)

// LibissPosFile хранит информацию о загруженных файлах программ libiss_pos
type LibissPosFile struct {
	ID             uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;"`
	Type           LibissPosType      `json:"type" gorm:"type:varchar(20);index;not null"`
	Platform       LibissPosPlatform  `json:"platform" gorm:"type:varchar(20);index;not null"` // Windows или Android
	Version        string             `json:"version" gorm:"type:varchar(50);not null;index"`
	FileName       string             `json:"fileName" gorm:"type:varchar(255);not null"`
	OriginalName   string             `json:"originalName" gorm:"type:varchar(255);not null"`
	FilePath       string             `json:"filePath" gorm:"type:varchar(500);not null"`
	FileURL        string             `json:"fileUrl" gorm:"type:varchar(500);not null"`
	PublicURL      string             `json:"publicUrl" gorm:"type:varchar(500);not null"` // Публичный URL для скачивания
	FileSize       int64              `json:"fileSize" gorm:"not null"`
	ChecksumSHA256 string             `json:"checksumSha256" gorm:"type:varchar(128);not null"`
	Description    string             `json:"description" gorm:"type:text"`
	IsActive       bool               `json:"isActive" gorm:"default:true"`
	IsPublic       bool               `json:"isPublic" gorm:"default:true"` // Доступен для публичного скачивания
	DownloadCount  int64              `json:"downloadCount" gorm:"default:0"`
	CreatedBy      uuid.UUID          `json:"createdBy" gorm:"type:uuid;index"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

// BeforeCreate устанавливает UUID перед созданием записи
func (l *LibissPosFile) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// GetTypeDisplayName возвращает отображаемое имя типа
func (l *LibissPosFile) GetTypeDisplayName() string {
	switch l.Type {
	case LibissPosTypeFull:
		return "Полный пакет (Касса1: программа + сервер + MySQL)"
	case LibissPosTypeCassa2:
		return "Программа для Касса2 (только Windows)"
	case LibissPosTypeServerOnly:
		return "Программа + сервер без MySQL"
	default:
		return string(l.Type)
	}
}

