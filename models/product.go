package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// @schemes http
// @title Example API
// @version 1.0
// @description This is an example server.
// @host localhost:8080
// @BasePath /api

// @securityDefinitions.basic BasicAuth

// @x-definitions pq.StringArray
// @name pq.StringArray
// @type array
// @items type string
type Product struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name       string         `json:"name"`
	Detail     string         `json:"detail"`
	Images     pq.StringArray `json:"images" swaggertype:"array,string" gorm:"type:text[]"`
	CategoryID uuid.UUID      `json:"category_id" gorm:"type:uuid"`
	Category   Category       `json:"category" gorm:"foreignKey:CategoryID"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// Auto-generate UUID before insert
func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}
