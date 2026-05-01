package dbrepository

import "time"

// orderModel is the GORM persistence model for orders.
// It is private to the adapter — the domain layer never sees GORM tags.
type orderModel struct {
	ID            string              `gorm:"primaryKey"`
	Products      []orderProductModel `gorm:"foreignKey:OrderID"`
	Status        string              `gorm:"column:status;not null;default:pending"`
	FailureReason string              `gorm:"column:failure_reason;not null;default:none"`
	CreatedAt     time.Time           `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time           `gorm:"column:updated_at;not null"`
}

func (orderModel) TableName() string { return "orders" }

// orderProductModel is the GORM persistence model for order products.
type orderProductModel struct {
	ID        string    `gorm:"primaryKey"`
	OrderID   string    `gorm:"column:order_id;not null;index"`
	ProductID string    `gorm:"column:product_id;not null"`
	Quantity  int       `gorm:"column:quantity;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (orderProductModel) TableName() string { return "order_products" }
