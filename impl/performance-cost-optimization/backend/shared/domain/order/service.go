package order

import (
	"context"
	"errors"
)

// OrderService provides domain logic for order and checkout operations.
type OrderService struct {
	repo OrderRepository
}

// NewOrderService creates a new OrderService with the given repository.
func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// PlaceOrder creates a new order from cart items.
// Returns the created order.
func (s *OrderService) PlaceOrder(ctx context.Context, cartID string, details []OrderDetail) (*Order, error) {
	if cartID == "" {
		return nil, errors.New("cart ID is required")
	}
	if len(details) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	order := &Order{
		CartID: cartID,
		Status: OrderStatusPending,
	}

	if err := s.repo.Save(ctx, order, details); err != nil {
		return nil, err
	}

	return order, nil
}

// GetByID retrieves a single order by ID, scoped to the given user.
func (s *OrderService) GetByID(ctx context.Context, userID, orderID string) (*Order, error) {
	order, err := s.repo.FindByIDAndUser(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	return order, nil
}

// GetHistory retrieves a paginated list of orders for a user.
// Returns the orders and the total count of matching records.
func (s *OrderService) GetHistory(ctx context.Context, userID string, page, pageSize int) ([]Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return s.repo.FindByUser(ctx, userID, page, pageSize)
}
