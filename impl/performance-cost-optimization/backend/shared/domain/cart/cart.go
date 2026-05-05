// Package cart defines the Cart domain model.
package cart

import (
	"time"

	"gorm.io/gorm"

	"shared/domain/shared"
)

// Cart represents a single item in a user's shopping cart.
// Each row maps one product to one user with its quantity and total price.
type Cart struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"type:uuid;not null;index" json:"user_id"`
	ProductID  string    `gorm:"type:uuid;not null" json:"product_id"`
	Quantity   int       `gorm:"type:integer;not null;default:1" json:"quantity"`
	TotalPrice float64   `gorm:"type:decimal(10,2);not null;default:0" json:"total_price"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set.
func (c *Cart) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		id, err := shared.GenerateUUID()
		if err != nil {
			return err
		}
		c.ID = id
	}
	return nil
}
