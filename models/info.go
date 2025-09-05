package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Info struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string    `json:"name" form:"name"`
	Detail    string    `json:"detail" form:"detail"`
	Phone     string    `json:"phone" form:"phone"`
	Telephone string    `json:"telephone" form:"telephone"`
	Address   string    `json:"address" form:"address"`
	ImagePath string    `json:"image_path"`
	Latitude  float64   `json:"latitude" form:"latitude"`
	Longitude float64   `json:"longitude" form:"longitude"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (i *Info) BeforeCreate(tx *gorm.DB) (err error) {
	i.ID = uuid.New()
	return
}
