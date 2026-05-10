package order

import (
	"time"

	"gorm.io/gorm"

	"shared/domain/shared"
)

// Order represents a placed order in the system.
type Order struct {
	ID            string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        string      `gorm:"type:uuid;not null;index" json:"user_id"`
	CartID        string      `gorm:"type:uuid;not null" json:"cart_id"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"created_at"`
	Status        OrderStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	FailureReason string      `gorm:"type:text" json:"failure_reason"`
}

// BeforeCreate hook to generate UUID if not set.
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		id, err := shared.GenerateUUID()
		if err != nil {
			return err
		}
		o.ID = id
	}
	return nil
}
