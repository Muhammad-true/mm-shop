package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UpdatePlatform перечисление платформ обновлений
type UpdatePlatform string

const (
	UpdatePlatformServer  UpdatePlatform = "server"  // Node.js (zip)
	UpdatePlatformWindows UpdatePlatform = "windows" // Flutter Windows (.exe)
	UpdatePlatformAndroid UpdatePlatform = "android" // Flutter Android (.apk)
)

// UpdateRelease хранит информацию об опубликованных обновлениях
type UpdateRelease struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	Platform       UpdatePlatform `json:"platform" gorm:"type:varchar(20);index;not null"`
	Version        string         `json:"version" gorm:"type:varchar(50);not null;index"`
	FileName       string         `json:"fileName" gorm:"type:varchar(255);not null"`
	FilePath       string         `json:"filePath" gorm:"type:varchar(255);not null"`
	FileURL        string         `json:"fileUrl" gorm:"type:varchar(255);not null"`
	FileSize       int64          `json:"fileSize" gorm:"not null"`
	ChecksumSHA256 string         `json:"checksumSha256" gorm:"type:varchar(128);not null"`
	ReleaseNotes   string         `json:"releaseNotes" gorm:"type:text"`
	IsActive       bool           `json:"isActive" gorm:"default:true"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// BeforeCreate устанавливает UUID перед созданием записи
func (u *UpdateRelease) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
