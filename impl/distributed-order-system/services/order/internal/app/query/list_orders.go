package query

import "context"

// ListOrders is the query for fetching all orders.
type ListOrders struct{}

// ListOrdersHandler processes the ListOrders query.
type ListOrdersHandler struct {
	readModel OrderReadModel
}

// NewListOrdersHandler constructs a new handler with its dependencies.
func NewListOrdersHandler(readModel OrderReadModel) ListOrdersHandler {
	return ListOrdersHandler{readModel: readModel}
}

// Handle returns all orders as OrderView slice.
func (h ListOrdersHandler) Handle(ctx context.Context, q ListOrders) ([]OrderView, error) {
	return h.readModel.ListOrders(ctx)
}
