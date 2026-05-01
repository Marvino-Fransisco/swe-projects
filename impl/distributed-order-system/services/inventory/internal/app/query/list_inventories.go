package query

import "context"

// ListInventories is the query for fetching paginated inventories.
type ListInventories struct {
	Page  int
	Limit int
}

// ListInventoriesHandler processes the ListInventories query.
type ListInventoriesHandler struct {
	readModel InventoryReadModel
}

// NewListInventoriesHandler constructs a new handler with its dependencies.
func NewListInventoriesHandler(readModel InventoryReadModel) ListInventoriesHandler {
	return ListInventoriesHandler{readModel: readModel}
}

// Handle returns a paginated list of InventoryView.
func (h ListInventoriesHandler) Handle(ctx context.Context, q ListInventories) (PaginationResult, error) {
	return h.readModel.ListInventories(ctx, q.Page, q.Limit)
}
