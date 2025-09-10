package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Favorite представляет избранный товар пользователя
type Favorite struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	UserID    uuid.UUID `json:"userId" gorm:"type:uuid;not null"`
	ProductID uuid.UUID `json:"productId" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"createdAt"`

	// Связи
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (f *Favorite) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// FavoriteRequest представляет запрос на добавление в избранное
type FavoriteRequest struct {
	ProductID uuid.UUID `json:"productId" binding:"required"`
}

// FavoriteResponse представляет ответ с избранным товаром
type FavoriteResponse struct {
	ID        uuid.UUID       `json:"id"`
	Product   ProductResponse `json:"product"`
	CreatedAt time.Time       `json:"createdAt"`
}

// FavoriteSyncRequest представляет запрос на синхронизацию избранного
type FavoriteSyncRequest struct {
	ProductIDs []uuid.UUID `json:"productIds" binding:"required"`
}

// ToResponse преобразует Favorite в FavoriteResponse
func (f *Favorite) ToResponse() FavoriteResponse {
	return FavoriteResponse{
		ID:        f.ID,
		Product:   f.Product.ToResponse(),
		CreatedAt: f.CreatedAt,
	}
}
