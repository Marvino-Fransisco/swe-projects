package cart

import (
	"context"
	"encoding/base64"

	"mobile-backend/apperror"

	cartDomain "shared/domain/cart"
	"shared/domain/product"
)

const defaultPageSize = 10

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

// encodeCursor base64-encodes a cart item ID into an opaque cursor string.
func encodeCursor(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}

// decodeCursor decodes a base64-encoded cursor back into the cart item ID.
// Returns empty string if the cursor is invalid or empty.
func decodeCursor(cursor string) string {
	if cursor == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return ""
	}
	return string(decoded)
}

// mapCartItems converts domain cart items to slim CartItem responses,
// enriching each with the product name and unit price.
func (uc *cartUseCase) mapCartItems(ctx context.Context, domainItems []cartDomain.Cart) ([]CartItem, error) {
	items := make([]CartItem, 0, len(domainItems))
	for _, ci := range domainItems {
		p, err := uc.productSvc.GetByID(ctx, ci.ProductID)
		if err != nil {
			return nil, apperror.NewNotFound("product not found")
		}

		items = append(items, CartItem{
			ProductID:   ci.ProductID,
			ProductName: p.Name,
			Price:       p.Price.Float64(),
			Quantity:    ci.Quantity,
			TotalPrice:  ci.TotalPrice,
		})
	}
	return items, nil
}

func (uc *cartUseCase) GetCart(ctx context.Context, req GetCartRequest) (*GetCartResponse, error) {
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = defaultPageSize
	}

	// Decode the cursor into the raw ID (empty string means first page).
	cursorID := decodeCursor(req.Cursor)

	// Fetch limit+1 items to determine if there is a next page.
	domainItems, err := uc.cartSvc.GetCartAfterCursor(ctx, req.UserID, cursorID, pageSize+1)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	hasMore := len(domainItems) > pageSize
	if hasMore {
		domainItems = domainItems[:pageSize]
	}

	items, err := uc.mapCartItems(ctx, domainItems)
	if err != nil {
		return nil, err
	}

	nextCursor := ""
	if hasMore && len(domainItems) > 0 {
		nextCursor = encodeCursor(domainItems[len(domainItems)-1].ID)
	}

	return &GetCartResponse{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}

func (uc *cartUseCase) AddItem(ctx context.Context, req AddCartItemRequest) (*CartResponse, error) {
	_, err := uc.productSvc.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound("product not found")
	}

	domainItems, err := uc.cartSvc.AddItem(ctx, req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	items, err := uc.mapCartItems(ctx, domainItems)
	if err != nil {
		return nil, err
	}

	return &CartResponse{Items: items}, nil
}

func (uc *cartUseCase) RemoveItem(ctx context.Context, req RemoveCartItemRequest) (*CartResponse, error) {
	domainItems, err := uc.cartSvc.RemoveItem(ctx, req.UserID, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound(err.Error())
	}

	items, err := uc.mapCartItems(ctx, domainItems)
	if err != nil {
		return nil, err
	}

	return &CartResponse{Items: items}, nil
}

func (uc *cartUseCase) UpdateItemQuantity(ctx context.Context, req UpdateCartItemRequest) (*CartResponse, error) {
	_, err := uc.productSvc.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound("product not found")
	}

	domainItems, err := uc.cartSvc.UpdateItemQuantity(ctx, req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	items, err := uc.mapCartItems(ctx, domainItems)
	if err != nil {
		return nil, err
	}

	return &CartResponse{Items: items}, nil
}
