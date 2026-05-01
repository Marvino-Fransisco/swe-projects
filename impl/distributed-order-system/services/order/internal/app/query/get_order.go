package query

import "context"

// GetOrder is the query for fetching a single order by ID.
type GetOrder struct {
	ID string
}

// GetOrderHandler processes the GetOrder query.
type GetOrderHandler struct {
	readModel OrderReadModel
}

// NewGetOrderHandler constructs a new handler with its dependencies.
func NewGetOrderHandler(readModel OrderReadModel) GetOrderHandler {
	return GetOrderHandler{readModel: readModel}
}

// Handle returns the OrderView for the given order ID.
func (h GetOrderHandler) Handle(ctx context.Context, q GetOrder) (OrderView, error) {
	return h.readModel.GetOrderByID(ctx, q.ID)
}
