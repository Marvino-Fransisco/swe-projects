package cart

import (
	"context"
	"errors"
)

// CartService provides domain logic for shopping cart operations.
type CartService struct {
	repo CartRepository
}

// NewCartService creates a new CartService with the given repository.
func NewCartService(repo CartRepository) *CartService {
	return &CartService{repo: repo}
}

// GetCart retrieves all cart items for the given user.
func (s *CartService) GetCart(ctx context.Context, userID string) ([]Cart, error) {
	items, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetCartAfterCursor retrieves cart items for a user using cursor-based pagination.
// cursor is the base64-decoded ID of the last item from the previous page.
// limit is the maximum number of items to return.
func (s *CartService) GetCartAfterCursor(ctx context.Context, userID, cursor string, limit int) ([]Cart, error) {
	return s.repo.FindByUserIDAfterCursor(ctx, userID, cursor, limit)
}

// AddItem adds a product to the user's cart.
// If the product already exists in the cart, the quantity is incremented.
func (s *CartService) AddItem(ctx context.Context, userID, productID string, quantity int) ([]Cart, error) {
	if quantity < 1 {
		return nil, errors.New("quantity must be at least 1")
	}

	existing, err := s.repo.FindByUserAndProduct(ctx, userID, productID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		existing.Quantity += quantity
		if err := s.repo.Save(ctx, existing); err != nil {
			return nil, err
		}
	} else {
		item := &Cart{
			UserID:    userID,
			ProductID: productID,
			Quantity:  quantity,
		}
		if err := s.repo.Save(ctx, item); err != nil {
			return nil, err
		}
	}

	return s.repo.FindByUserID(ctx, userID)
}

// RemoveItem removes a product from the user's cart.
func (s *CartService) RemoveItem(ctx context.Context, userID, productID string) ([]Cart, error) {
	item, err := s.repo.FindByUserAndProduct(ctx, userID, productID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found in cart")
	}

	if err := s.repo.Delete(ctx, item); err != nil {
		return nil, err
	}

	return s.repo.FindByUserID(ctx, userID)
}

// UpdateItemQuantity updates the quantity of a specific item in the cart.
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID, productID string, quantity int) ([]Cart, error) {
	if quantity < 1 {
		return nil, errors.New("quantity must be at least 1")
	}

	item, err := s.repo.FindByUserAndProduct(ctx, userID, productID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found in cart")
	}

	item.Quantity = quantity
	if err := s.repo.Save(ctx, item); err != nil {
		return nil, err
	}

	return s.repo.FindByUserID(ctx, userID)
}
