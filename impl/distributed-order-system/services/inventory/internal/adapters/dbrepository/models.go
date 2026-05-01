package dbrepository

import "time"

// inventoryModel is the GORM persistence model for inventories.
// It is private to the adapter — the domain layer never sees GORM tags.
type inventoryModel struct {
	ID          uint      `gorm:"primaryKey"`
	ProductID   string    `gorm:"column:product_id;not null"`
	ProductName string    `gorm:"column:product_name;not null"`
	Stock       int       `gorm:"column:stock;not null;default:0"`
	Price       float64   `gorm:"column:price;not null;default:0"`
	Status      string    `gorm:"column:status;not null;default:available"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (inventoryModel) TableName() string { return "inventories" }

// reservationModel is the GORM persistence model for inventory reservations.
// It is private to the adapter — the domain layer never sees GORM tags.
type reservationModel struct {
	ID        uint      `gorm:"primaryKey"`
	OrderID   string    `gorm:"column:order_id;not null"`
	ProductID string    `gorm:"column:product_id;not null"`
	Quantity  int       `gorm:"column:quantity;not null"`
	Status    string    `gorm:"column:status;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (reservationModel) TableName() string { return "inventory_reservations" }
