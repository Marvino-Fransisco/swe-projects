// Package cart defines the Cart domain model.
package cart

import (
	"time"

	"gorm.io/gorm"

	"shared/domain/product"
	"shared/domain/shared"
)

// Cart represents a user's shopping cart (aggregate root).
// Each user has exactly one cart.
type Cart struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"type:uuid;not null;index" json:"user_id"`
	Items     []CartItem `gorm:"foreignKey:CartID" json:"items"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// CartItem represents a single product entry within a cart.
// It uses a composite primary key of CartID and ProductID.
type CartItem struct {
	CartID    string          `gorm:"type:uuid;primaryKey" json:"cart_id"`
	ProductID string          `gorm:"type:uuid;primaryKey" json:"product_id"`
	Product   product.Product `gorm:"foreignKey:ProductID" json:"product"`
	Quantity  int             `gorm:"type:integer;not null;default:1" json:"quantity"`
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
