package dbrepository

import "time"

// paymentModel is the GORM persistence model for payments.
// It is private to the adapter — the domain layer never sees GORM tags.
type paymentModel struct {
	PaymentID  string    `gorm:"primaryKey;column:payment_id"`
	OrderID    string    `gorm:"column:order_id;not null;index"`
	TotalPrice float64   `gorm:"column:total_price;not null"`
	Status     string    `gorm:"column:status;not null;default:pending"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null"`
}

func (paymentModel) TableName() string { return "payments" }
