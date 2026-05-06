package cart

import "context"

type CartUseCase interface {
	GetCart(ctx context.Context, req GetCartRequest) (*GetCartResponse, error)
	AddItem(ctx context.Context, req AddCartItemRequest) (*CartItem, error)
	RemoveItem(ctx context.Context, req RemoveCartItemRequest) (*CartItem, error)
	UpdateItemQuantity(ctx context.Context, req UpdateCartItemRequest) (*CartItem, error)
}
