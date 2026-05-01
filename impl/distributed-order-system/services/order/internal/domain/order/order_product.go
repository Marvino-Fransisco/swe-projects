package order

import "time"

// OrderProduct represents a product line item within an order.
type OrderProduct struct {
	id        string
	orderID   string
	productID string
	quantity  int
	createdAt time.Time
	updatedAt time.Time
}

// NewOrderProduct creates a new OrderProduct with the given details.
func NewOrderProduct(id, productID string, quantity int, now time.Time) OrderProduct {
	return OrderProduct{
		id:        id,
		productID: productID,
		quantity:  quantity,
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructOrderProduct rebuilds an OrderProduct from persistence.
func ReconstructOrderProduct(id, orderID, productID string, quantity int, createdAt, updatedAt time.Time) OrderProduct {
	return OrderProduct{
		id:        id,
		orderID:   orderID,
		productID: productID,
		quantity:  quantity,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (p OrderProduct) ID() string        { return p.id }
func (p OrderProduct) OrderID() string    { return p.orderID }
func (p OrderProduct) ProductID() string  { return p.productID }
func (p OrderProduct) Quantity() int      { return p.quantity }
func (p OrderProduct) CreatedAt() time.Time { return p.createdAt }
func (p OrderProduct) UpdatedAt() time.Time { return p.updatedAt }
