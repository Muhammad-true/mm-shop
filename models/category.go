package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category представляет категорию товаров
type Category struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"`
	Name         string     `json:"name" gorm:"not null"`
	Description  string     `json:"description"`
	IconURL      string     `json:"iconUrl"`
	ParentID     *uuid.UUID `json:"parentId" gorm:"type:uuid"`
	ProductCount int        `json:"productCount" gorm:"default:0"`
	IsActive     bool       `json:"isActive" gorm:"default:true"`
	SortOrder    int        `json:"sortOrder" gorm:"default:0"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`

	// Связи
	Parent        *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Subcategories []Category `json:"subcategories,omitempty" gorm:"foreignKey:ParentID"`
	Products      []Product  `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
}

// BeforeCreate устанавливает UUID перед созданием
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CategoryRequest представляет запрос на создание/обновление категории
type CategoryRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	IconURL     string     `json:"iconUrl"`
	ParentID    *uuid.UUID `json:"parentId"`
	SortOrder   int        `json:"sortOrder"`
	IsActive    bool       `json:"isActive"`
}

// CategoryResponse представляет ответ с категорией
type CategoryResponse struct {
	ID            uuid.UUID          `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	IconURL       string             `json:"iconUrl"`
	ParentID      *uuid.UUID         `json:"parentId"`
	ProductCount  int                `json:"productCount"`
	IsActive      bool               `json:"isActive"`
	SortOrder     int                `json:"sortOrder"`
	Subcategories []CategoryResponse `json:"subcategories,omitempty"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}

// CategoryTree представляет дерево категорий
type CategoryTree struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	IconURL      string         `json:"iconUrl"`
	ProductCount int            `json:"productCount"`
	IsActive     bool           `json:"isActive"`
	SortOrder    int            `json:"sortOrder"`
	Children     []CategoryTree `json:"children,omitempty"`
}

// ToResponse преобразует Category в CategoryResponse
func (c *Category) ToResponse() CategoryResponse {
	subcategories := make([]CategoryResponse, len(c.Subcategories))
	for i, sub := range c.Subcategories {
		subcategories[i] = sub.ToResponse()
	}

	return CategoryResponse{
		ID:            c.ID,
		Name:          c.Name,
		Description:   c.Description,
		IconURL:       c.IconURL,
		ParentID:      c.ParentID,
		ProductCount:  c.ProductCount,
		IsActive:      c.IsActive,
		SortOrder:     c.SortOrder,
		Subcategories: subcategories,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// ToTree преобразует Category в CategoryTree
func (c *Category) ToTree() CategoryTree {
	children := make([]CategoryTree, len(c.Subcategories))
	for i, sub := range c.Subcategories {
		children[i] = sub.ToTree()
	}

	return CategoryTree{
		ID:           c.ID,
		Name:         c.Name,
		Description:  c.Description,
		IconURL:      c.IconURL,
		ProductCount: c.ProductCount,
		IsActive:     c.IsActive,
		SortOrder:    c.SortOrder,
		Children:     children,
	}
}

// IsRootCategory проверяет, является ли категория корневой
func (c *Category) IsRootCategory() bool {
	return c.ParentID == nil
}

// GetPath возвращает путь категории (например: "Одежда > Мужская > Футболки")
func (c *Category) GetPath(db *gorm.DB) string {
	if c.ParentID == nil {
		return c.Name
	}

	var parent Category
	if err := db.First(&parent, "id = ?", c.ParentID).Error; err != nil {
		return c.Name
	}

	return parent.GetPath(db) + " > " + c.Name
}
