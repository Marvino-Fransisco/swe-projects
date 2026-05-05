package order

// OrderDetail represents the line items within an order.
// It uses a composite primary key of OrderID and ProductID.
type OrderDetail struct {
	OrderID   string `gorm:"type:uuid;primaryKey" json:"order_id"`
	ProductID string `gorm:"type:uuid;primaryKey" json:"product_id"`
	Quantity  int    `gorm:"type:integer;not null;default:1" json:"quantity"`
}
