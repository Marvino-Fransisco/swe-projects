package cart

import (
	"context"
	"encoding/base64"

	"mobile-backend/apperror"
	"mobile-backend/repository"

	cartDomain "shared/domain/cart"
	"shared/domain/product"
)

const defaultPageSize = 10

type cartUseCase struct {
	cartSvc       *cartDomain.CartService
	productSvc    *product.ProductService
	cartQueryRepo repository.CartQueryRepository
}

func NewCartUseCase(
	cartSvc *cartDomain.CartService,
	productSvc *product.ProductService,
	cartQueryRepo repository.CartQueryRepository,
) CartUseCase {
	return &cartUseCase{
		cartSvc:       cartSvc,
		productSvc:    productSvc,
		cartQueryRepo: cartQueryRepo,
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

// mapCartItem converts a single domain cart item to a CartItem response,
// enriching it with the product name and unit price.
func (uc *cartUseCase) mapCartItem(ctx context.Context, ci cartDomain.Cart) (*CartItem, error) {
	p, err := uc.productSvc.GetByID(ctx, ci.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound("product not found")
	}

	return &CartItem{
		ProductID:   ci.ProductID,
		ProductName: p.Name,
		Price:       p.Price.Float64(),
		Quantity:    ci.Quantity,
		TotalPrice:  ci.TotalPrice,
	}, nil
}

// mapCartItems converts a slice of domain cart items to CartItem responses,
// enriching each with the product name and unit price.
func (uc *cartUseCase) mapCartItems(ctx context.Context, domainItems []cartDomain.Cart) ([]CartItem, error) {
	items := make([]CartItem, 0, len(domainItems))
	for _, ci := range domainItems {
		mapped, err := uc.mapCartItem(ctx, ci)
		if err != nil {
			return nil, err
		}
		items = append(items, *mapped)
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
	domainItems, err := uc.cartQueryRepo.FindByUserIDAfterCursor(ctx, req.UserID, cursorID, pageSize+1)
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

func (uc *cartUseCase) AddItem(ctx context.Context, req AddCartItemRequest) (*CartItem, error) {
	_, err := uc.productSvc.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound("product not found")
	}

	cartItem, err := uc.cartSvc.AddItem(ctx, req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	return uc.mapCartItem(ctx, *cartItem)
}

func (uc *cartUseCase) RemoveItem(ctx context.Context, req RemoveCartItemRequest) (*CartItem, error) {
	cartItem, err := uc.cartSvc.RemoveItem(ctx, req.UserID, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound(err.Error())
	}

	return uc.mapCartItem(ctx, *cartItem)
}

func (uc *cartUseCase) UpdateItemQuantity(ctx context.Context, req UpdateCartItemRequest) (*CartItem, error) {
	_, err := uc.productSvc.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, apperror.NewNotFound("product not found")
	}

	cartItem, err := uc.cartSvc.UpdateItemQuantity(ctx, req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	return uc.mapCartItem(ctx, *cartItem)
}
