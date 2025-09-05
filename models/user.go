package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Address  string    `json:"address"`
	Regency  string    `json:"regency"`
	District string    `json:"district"`
	Phone    string    `json:"phone" gorm:"unique"`
	PhotoUrl string    `json:"photo_url" `
	IsActive bool      `json:"is_active" gorm:"default:false"`
	Lat      float64   `json:"lat"`
	Lang     float64   `json:"lang"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Auto-generate UUID before insert
func (c *User) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}
