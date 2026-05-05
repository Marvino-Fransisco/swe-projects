package order

import (
	"database/sql/driver"
	"errors"
)

// OrderStatus represents the current status of an order.
type OrderStatus string

const (
	// OrderStatusPending indicates the order has been created but not yet completed.
	OrderStatusPending OrderStatus = "PENDING"

	// OrderStatusCompleted indicates the order has been successfully fulfilled.
	OrderStatusCompleted OrderStatus = "COMPLETED"

	// OrderStatusFailed indicates the order could not be completed.
	OrderStatusFailed OrderStatus = "FAILED"
)

// validOrderStatuses contains all allowed order status values.
var validOrderStatuses = map[OrderStatus]bool{
	OrderStatusPending:   true,
	OrderStatusCompleted: true,
	OrderStatusFailed:    true,
}

// NewOrderStatus creates and validates an OrderStatus value.
func NewOrderStatus(status string) (OrderStatus, error) {
	os := OrderStatus(status)
	if !validOrderStatuses[os] {
		return "", errors.New("invalid order status: must be PENDING, COMPLETED, or FAILED")
	}

	return os, nil
}

// String returns the string representation of the OrderStatus.
func (s OrderStatus) String() string {
	return string(s)
}

// Value implements the driver.Valuer interface for database writes.
func (s OrderStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements the sql.Scanner interface for database reads.
func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*s = OrderStatus(v)
	case []byte:
		*s = OrderStatus(v)
	default:
		return errors.New("cannot scan OrderStatus from non-string type")
	}

	return nil
}
