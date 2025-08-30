package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Info struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Phone     string    `json:"phone"`
	Telephone string    `json:"telephone"`
	Address   string    `json:"address"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (i *Info) BeforeCreate(tx *gorm.DB) (err error) {
	i.ID = uuid.New()
	return
}
