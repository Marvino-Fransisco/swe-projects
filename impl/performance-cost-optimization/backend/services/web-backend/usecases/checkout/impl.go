package checkout

import (
	"context"

	"web-backend/apperror"

	"shared/domain/cart"
	"shared/domain/order"
)

type checkoutUseCase struct {
	cartSvc  *cart.CartService
	orderSvc *order.OrderService
}

func NewCheckoutUseCase(cartSvc *cart.CartService, orderSvc *order.OrderService) CheckoutUseCase {
	return &checkoutUseCase{
		cartSvc:  cartSvc,
		orderSvc: orderSvc,
	}
}

func (uc *checkoutUseCase) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*order.Order, error) {
	items, err := uc.cartSvc.GetCart(ctx, req.UserID)
	if err != nil {
		return nil, apperror.NewBadRequest("failed to retrieve cart")
	}
	if len(items) == 0 {
		return nil, apperror.NewBadRequest("cart is empty")
	}

	cartID := items[0].ID
	details := make([]order.OrderDetail, 0, len(items))
	for _, item := range items {
		details = append(details, order.OrderDetail{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	o, err := uc.orderSvc.PlaceOrder(ctx, cartID, details)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	for _, item := range items {
		_, _ = uc.cartSvc.RemoveItem(ctx, req.UserID, item.ProductID)
	}

	return o, nil
}

func (uc *checkoutUseCase) GetOrder(ctx context.Context, req GetOrderRequest) (*OrderWithDetailsResponse, error) {
	o, err := uc.orderSvc.GetByID(ctx, req.UserID, req.OrderID)
	if err != nil {
		return nil, apperror.NewNotFound(err.Error())
	}

	return &OrderWithDetailsResponse{
		Order: o,
	}, nil
}

func (uc *checkoutUseCase) GetOrderHistory(ctx context.Context, req GetOrderHistoryRequest) (*OrderHistoryResponse, error) {
	orders, total, err := uc.orderSvc.GetHistory(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	return &OrderHistoryResponse{
		Orders: orders,
		Total:  total,
	}, nil
}
