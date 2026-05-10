package cart

import "context"

type CartUseCase interface {
	GetCart(ctx context.Context, req GetCartRequest) (*GetCartResponse, error)
	AddItem(ctx context.Context, req AddCartItemRequest) (*CartResponse, error)
	RemoveItem(ctx context.Context, req RemoveCartItemRequest) (*CartResponse, error)
	UpdateItemQuantity(ctx context.Context, req UpdateCartItemRequest) (*CartResponse, error)
}
