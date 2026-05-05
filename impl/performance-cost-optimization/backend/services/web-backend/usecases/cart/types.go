package cart

import cartDomain "shared/domain/cart"

type GetCartRequest struct {
	UserID string
	Page   int
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

type GetCartResponse struct {
	Items      []cartDomain.Cart `json:"carts"`
	Page       int               `json:"page"`
	TotalPages int               `json:"totalPages"`
}

type CartResponse struct {
	Items []cartDomain.Cart `json:"items"`
}
