package order

import "context"

// OrderRepository defines the interface for order persistence operations.
type OrderRepository interface {
	// Save persists a new order and its order details.
	Save(ctx context.Context, o *Order, details []OrderDetail) error

	// FindByID retrieves an order by its ID.
	// Returns nil if not found.
	FindByID(ctx context.Context, id string) (*Order, error)

	// FindByIDAndUser retrieves an order by its ID filtered by userID.
	// Returns nil if not found or does not belong to the user.
	FindByIDAndUser(ctx context.Context, id, userID string) (*Order, error)

	// FindDetailsByOrderID retrieves all order details for a given order.
	FindDetailsByOrderID(ctx context.Context, orderID string) ([]OrderDetail, error)

	// FindByUser retrieves a paginated list of orders for a user.
	// Returns the orders and the total count of matching records.
	FindByUser(ctx context.Context, userID string, page, pageSize int) ([]Order, int64, error)

	// Update persists changes to an existing order.
	Update(ctx context.Context, o *Order) error
}
