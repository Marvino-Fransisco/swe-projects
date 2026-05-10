package checkout

import (
	"context"

	"shared/domain/order"
)

type CheckoutUseCase interface {
	PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*order.Order, error)
	GetOrder(ctx context.Context, req GetOrderRequest) (*OrderWithDetailsResponse, error)
	GetOrderHistory(ctx context.Context, req GetOrderHistoryRequest) (*OrderHistoryResponse, error)
}
