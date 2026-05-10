package cart

import (
	"context"
	"errors"
	"math"

	"web-backend/apperror"

	cartDomain "shared/domain/cart"
	"shared/domain/product"
)

type cartUseCase struct {
	cartSvc    *cartDomain.CartService
	productSvc *product.ProductService
}

func NewCartUseCase(
	cartSvc *cartDomain.CartService,
	productSvc *product.ProductService,
) CartUseCase {
	return &cartUseCase{
		cartSvc:    cartSvc,
		productSvc: productSvc,
	}
}

// findCartItem is a helper to find a CartItem by productID within a Cart's Items.
func findCartItem(c *cartDomain.Cart, productID string) *cartDomain.CartItem {
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			return &c.Items[i]
		}
	}
	return nil
}

func (uc *cartUseCase) GetCart(ctx context.Context, req GetCartRequest) (*GetCartResponse, error) {
	c, err := uc.cartSvc.GetCart(ctx, req.UserID)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	if c == nil {
		c = &cartDomain.Cart{UserID: req.UserID, Items: []cartDomain.CartItem{}}
	}

	items := c.Items

	// Paginate in memory
	total := int64(len(items))
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	offset := (req.Page - 1) * req.PageSize
	end := offset + req.PageSize
	if end > len(items) {
		end = len(items)
	}
	if offset > len(items) {
		offset = len(items)
	}

	pageItems := items[offset:end]

	return &GetCartResponse{
		Items:      pageItems,
		Page:       req.Page,
		TotalPages: totalPages,
	}, nil
}

func (uc *cartUseCase) AddItem(ctx context.Context, req AddCartItemRequest) (*CartResponse, error) {
	c, err := uc.cartSvc.AddItem(ctx, req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	targetItem := findCartItem(c, req.ProductID)
	if targetItem == nil {
		return nil, apperror.NewBadRequest("failed to locate item in cart")
	}

	return &CartResponse{Item: targetItem}, nil
}

func (uc *cartUseCase) RemoveItem(ctx context.Context, req RemoveCartItemRequest) (*CartResponse, error) {
	// Get cart to find the item before removal.
	c, err := uc.cartSvc.GetCart(ctx, req.UserID)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}
	if c == nil {
		return nil, apperror.NewNotFound("cart not found")
	}

	removedItem := findCartItem(c, req.ProductID)
	if removedItem == nil {
		return nil, apperror.NewNotFound("item not found in cart")
	}

	// Make a copy before removal.
	itemCopy := *removedItem

	// Remove item via domain service.
	_, err = uc.cartSvc.RemoveItem(ctx, req.UserID, req.ProductID)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	return &CartResponse{Item: &itemCopy}, nil
}

func (uc *cartUseCase) UpdateItemQuantity(ctx context.Context, req UpdateCartItemRequest) (*CartResponse, error) {
	// Validate product exists
	_, err := uc.productSvc.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound("product not found")
	}

	c, err := uc.cartSvc.UpdateItemQuantity(ctx, req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		if errors.Is(err, cartDomain.ErrItemNotFound) {
			return nil, apperror.NewNotFound("item not found in cart")
		}
		return nil, apperror.NewBadRequest(err.Error())
	}

	targetItem := findCartItem(c, req.ProductID)
	if targetItem == nil {
		return nil, apperror.NewBadRequest("failed to locate updated item in cart")
	}

	return &CartResponse{Item: targetItem}, nil
}
