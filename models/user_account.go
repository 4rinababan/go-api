package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserAccount struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Phone        string         `json:"phone"`
	PasswordHash string         `json:"password_hash"`
	Role         string         `json:"role"`
	UserID       uuid.UUID      `json:"user_id" gorm:"type:uuid"`      // FK to Category
	User         User           `json:"user" gorm:"foreignKey:UserID"` // optional preload
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Auto-generate UUID before insert
func (c *UserAccount) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}
