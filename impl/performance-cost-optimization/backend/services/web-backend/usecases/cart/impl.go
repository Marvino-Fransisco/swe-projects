package cart

import (
	"context"

	"web-backend/apperror"

	cartDomain "shared/domain/cart"
	"shared/domain/product"
)

type cartUseCase struct {
	cartSvc    *cartDomain.CartService
	productSvc *product.ProductService
}

func NewCartUseCase(cartSvc *cartDomain.CartService, productSvc *product.ProductService) CartUseCase {
	return &cartUseCase{
		cartSvc:    cartSvc,
		productSvc: productSvc,
	}
}

func (uc *cartUseCase) GetCart(ctx context.Context, req GetCartRequest) (*GetCartResponse, error) {
	fetchedItems, err := uc.cartSvc.GetCart(ctx, req.UserID)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	items := make([]cartDomain.Cart, 10)
	if len(fetchedItems) != 0 {
		for i := (req.Page - 1) * 10; i < req.Page*10; i++ {
			items = append(items, fetchedItems[i])
		}
	} else {
		items = []cartDomain.Cart{}
	}

	return &GetCartResponse{
		Items:      items,
		Page:       req.Page,
		TotalPages: len(items),
	}, nil
}

func (uc *cartUseCase) AddItem(ctx context.Context, req AddCartItemRequest) (*CartResponse, error) {
	_, err := uc.productSvc.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound("product not found")
	}

	items, err := uc.cartSvc.AddItem(ctx, req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	return &CartResponse{Items: items}, nil
}

func (uc *cartUseCase) RemoveItem(ctx context.Context, req RemoveCartItemRequest) (*CartResponse, error) {
	items, err := uc.cartSvc.RemoveItem(ctx, req.UserID, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound(err.Error())
	}

	return &CartResponse{Items: items}, nil
}

func (uc *cartUseCase) UpdateItemQuantity(ctx context.Context, req UpdateCartItemRequest) (*CartResponse, error) {
	_, err := uc.productSvc.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound("product not found")
	}

	items, err := uc.cartSvc.UpdateItemQuantity(ctx, req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	return &CartResponse{Items: items}, nil
}
