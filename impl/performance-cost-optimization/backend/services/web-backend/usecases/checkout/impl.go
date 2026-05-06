package checkout

import (
	"context"

	"web-backend/apperror"
	"web-backend/repository"

	"shared/domain/cart"
	"shared/domain/order"
)

type checkoutUseCase struct {
	cartSvc        *cart.CartService
	orderSvc       *order.OrderService
	orderQueryRepo repository.OrderQueryRepository
}

func NewCheckoutUseCase(
	cartSvc *cart.CartService,
	orderSvc *order.OrderService,
	orderQueryRepo repository.OrderQueryRepository,
) CheckoutUseCase {
	return &checkoutUseCase{
		cartSvc:        cartSvc,
		orderSvc:       orderSvc,
		orderQueryRepo: orderQueryRepo,
	}
}

func (uc *checkoutUseCase) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*order.Order, error) {
	c, err := uc.cartSvc.GetCart(ctx, req.UserID)
	if err != nil {
		return nil, apperror.NewBadRequest("failed to retrieve cart")
	}
	if c == nil || len(c.Items) == 0 {
		return nil, apperror.NewBadRequest("cart is empty")
	}

	details := make([]order.OrderDetail, 0, len(c.Items))
	for _, item := range c.Items {
		details = append(details, order.OrderDetail{
			OrderID:   "", // assigned by PlaceOrder
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	o, err := uc.orderSvc.PlaceOrder(ctx, c.ID, details)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	// Clear the cart after successful order placement.
	if err := uc.cartSvc.DeleteCart(ctx, req.UserID); err != nil {
		// Order is already placed; log but don't fail the response.
		_ = err
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
	orders, total, err := uc.orderQueryRepo.FindByUserPaginated(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	return &OrderHistoryResponse{
		Orders: orders,
		Total:  total,
	}, nil
}
