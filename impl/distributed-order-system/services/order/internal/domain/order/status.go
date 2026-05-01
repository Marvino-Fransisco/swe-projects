package order

// OrderStatus represents the current state of an order.
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusCancelled OrderStatus = "cancelled"
	StatusFailed    OrderStatus = "failed"
)

func (s OrderStatus) String() string {
	return string(s)
}
