package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	OrderCode string    `json:"order_id" gorm:"uniqueIndex"` // contoh: ORD123
	// UserID and ProductID are foreign keys to User and Product models
	UserID      uuid.UUID           `json:"user_id" gorm:"type:uuid"`            // FK to User
	User        User                `json:"user" gorm:"foreignKey:UserID"`       // optional preload
	ProductID   uuid.UUID           `json:"product_id" gorm:"type:uuid"`         // FK to Product
	Product     Product             `json:"product" gorm:"foreignKey:ProductID"` // optional preload
	CompanyName string              `json:"company_name" gorm:"type:text"`       // name of the product
	Priority    string              `json:"priority" gorm:"default:'normal'"`    // e.g., "low", "normal", "high"
	Details     string              `json:"details" gorm:"type:text"`            // additional details about the order
	Address     string              `json:"address" gorm:"type:text"`            // delivery address
	Quantity    int                 `json:"quantity"`
	Updates     []OrderStatusUpdate `json:"updates" gorm:"foreignKey:OrderID"`
	Status      string              `json:"status"` // e.g., "pending", "completed",
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `gorm:"index" json:"-"`
}

// Auto-generate UUID before insert
func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = uuid.New()
	if o.OrderCode == "" {
		o.OrderCode = "ORD" + time.Now().Format("060102150405") // misalnya generate ORD250815xxxx
	}
	return
}

type OrderStatusUpdate struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	OrderID   uuid.UUID `json:"-" gorm:"type:uuid"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func (u *OrderStatusUpdate) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	if u.Timestamp.IsZero() {
		u.Timestamp = time.Now()
	}
	return
}
