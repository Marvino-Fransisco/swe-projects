package cart

type GetCartRequest struct {
	UserID   string
	Cursor   string // base64-encoded ID of the last item from previous page (empty for first page)
	PageSize int    // number of items to return (default 10)
}

type AddCartItemRequest struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type RemoveCartItemRequest struct {
	UserID    string
	ProductID string
}

type UpdateCartItemRequest struct {
	UserID    string
	ProductID string
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// CartItem is a slim representation of a cart item for mobile clients.
type CartItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	TotalPrice  float64 `json:"total_price"`
}

type GetCartResponse struct {
	Items      []CartItem `json:"items"`
	NextCursor string     `json:"next_cursor"` // base64-encoded ID of last item, empty string if no more items
}

type CartResponse struct {
	Items []CartItem `json:"items"`
}
