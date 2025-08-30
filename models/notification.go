package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID      `json:"user_id"`
	Message   string         `json:"message"`
	OrderID   uuid.UUID      `json:"order_id"`
	Order     Order          `json:"order" gorm:"foreignKey:OrderID"`
	CreatedAt time.Time      `json:"created_at"`
	Read      bool           `json:"read"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	n.ID = uuid.New()
	n.CreatedAt = time.Now()
	return
}
