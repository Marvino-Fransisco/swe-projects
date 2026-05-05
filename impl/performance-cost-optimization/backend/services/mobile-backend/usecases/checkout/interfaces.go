package checkout

import "context"

type CheckoutUseCase interface {
	PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*OrderResponse, error)
	GetOrder(ctx context.Context, req GetOrderRequest) (*OrderWithDetailsResponse, error)
	GetOrderHistory(ctx context.Context, req GetOrderHistoryRequest) (*OrderHistoryResponse, error)
}
