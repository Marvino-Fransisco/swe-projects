package product

import (
	"time"

	"gorm.io/gorm"

	"shared/domain/shared"
)

// Product represents a product available for purchase.
type Product struct {
	ID          string        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string        `gorm:"type:varchar(255);not null" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	Price       Price         `gorm:"type:decimal(10,2);not null;default:0" json:"price"`
	Stock       Stock         `gorm:"type:integer;not null;default:0" json:"stock"`
	View        View          `gorm:"type:bigint;not null;default:0" json:"view"`
	Status      ProductStatus `gorm:"type:varchar(20);not null;default:'EMPTY'" json:"status"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set.
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		id, err := shared.GenerateUUID()
		if err != nil {
			return err
		}
		p.ID = id
	}
	return nil
}
