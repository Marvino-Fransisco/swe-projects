package cart

import (
	"context"
	"errors"
	"log"

	"shared/domain/shared"
)

// Sentinel errors for cart domain operations.
var (
	ErrCartNotFound    = errors.New("cart not found")
	ErrItemNotFound    = errors.New("item not found in cart")
	ErrInvalidQuantity = errors.New("quantity must be at least 1")
)

// CartService provides domain logic for shopping cart operations.
// It operates on the Cart aggregate root (one cart per user).
//
// When a cacheRepo is configured (via NewCachedCartService), reads follow a
// cache-first strategy (cache → DB on miss → populate cache) and writes
// target the cache only. When cacheRepo is nil (via NewCartService), all
// operations fall back to the database directly.
type CartService struct {
	repo      CartRepository
	cacheRepo CartCacheRepository
}

// NewCartService creates a new CartService with DB-only persistence (no cache).
// Use this when cache is not needed (e.g. mobile-backend).
func NewCartService(repo CartRepository) *CartService {
	return &CartService{repo: repo}
}

// NewCachedCartService creates a new CartService with cache-first reads and
// cache-only writes. The DB repository is used only as a read fallback on
// cache misses.
func NewCachedCartService(repo CartRepository, cacheRepo CartCacheRepository) *CartService {
	return &CartService{repo: repo, cacheRepo: cacheRepo}
}

// getCart is an internal helper that retrieves the user's cart using a
// cache-first strategy when cache is available, otherwise falls back to DB.
func (s *CartService) getCart(ctx context.Context, userID string) (*Cart, error) {
	if s.cacheRepo != nil {
		cached, err := s.cacheRepo.GetByUserID(ctx, userID)
		if err != nil {
			log.Printf("cache read error for user %s: %v", userID, err)
		}
		if cached != nil {
			return cached, nil
		}
	}

	// Cache miss or no cache — read from database.
	c, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Populate cache on miss.
	if c != nil && s.cacheRepo != nil {
		if cacheErr := s.cacheRepo.Set(ctx, userID, c); cacheErr != nil {
			log.Printf("cache population error for user %s: %v", userID, cacheErr)
		}
	}

	return c, nil
}

// saveCart persists the cart — to cache when available, otherwise to database.
// When writing to cache, it also marks the cart as dirty so the sync worker
// will persist it to the database.
func (s *CartService) saveCart(ctx context.Context, c *Cart) error {
	if s.cacheRepo != nil {
		if err := s.cacheRepo.Set(ctx, c.UserID, c); err != nil {
			return err
		}
		if dirtyErr := s.cacheRepo.SetDirty(ctx, c.UserID); dirtyErr != nil {
			log.Printf("failed to mark cart dirty for user %s: %v", c.UserID, dirtyErr)
		}
		return nil
	}
	return s.repo.Save(ctx, c)
}

// GetCart retrieves the user's cart.
// Returns nil if no cart exists for the user.
func (s *CartService) GetCart(ctx context.Context, userID string) (*Cart, error) {
	return s.getCart(ctx, userID)
}

// SyncToDB persists the user's cached cart state to the database.
// This is intended to be called by the background sync worker.
//   - If the cart exists in cache, it replaces the cart in the database.
//   - If the cart was deleted from cache (nil), it deletes the cart from the database.
//   - After a successful database sync, it removes the dirty flag from cache.
//
// No-op when cache is not configured.
func (s *CartService) SyncToDB(ctx context.Context, userID string) error {
	if s.cacheRepo == nil {
		return nil
	}

	// Read directly from cache — it is the source of truth for sync.
	cached, err := s.cacheRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if cached != nil {
		for i := range cached.Items {
			cached.Items[i].CartID = cached.ID
		}

		// Cart exists in cache → persist to database.
		// Check if the cart row already exists in DB.
		existing, err := s.repo.FindByUserID(ctx, userID)
		if err != nil {
			return err
		}

		if existing == nil {
			// Cart does not exist in DB yet (created in cache only).
			// GORM's Save processes associations before the parent, which
			// causes FK violations on cart_items. To avoid this, we save
			// the parent carts row first (without items), then insert items.
			items := cached.Items
			cached.Items = nil
			if err := s.repo.Save(ctx, cached); err != nil {
				return err
			}
			cached.Items = items
			if len(items) > 0 {
				if err := s.repo.ReplaceCart(ctx, cached); err != nil {
					return err
				}
			}
		} else {
			// Cart already exists in DB — only replace the items.
			if err := s.repo.ReplaceCart(ctx, cached); err != nil {
				return err
			}
		}
	} else {
		// Cart was deleted from cache → delete from database.
		existing, err := s.repo.FindByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if existing != nil {
			if err := s.repo.Delete(ctx, existing); err != nil {
				return err
			}
		}
	}

	// Remove dirty flag after successful sync.
	return s.cacheRepo.DeleteCartDirtyMember(ctx, userID)
}

// DeleteCart deletes the user's entire cart including all items.
// When cache is configured, deletes from cache and marks dirty for DB sync.
// When no cache, falls back to database deletion.
// Returns ErrCartNotFound if no cart exists for the user (DB mode only).
func (s *CartService) DeleteCart(ctx context.Context, userID string) error {
	if s.cacheRepo != nil {
		if err := s.cacheRepo.Delete(ctx, userID); err != nil {
			return err
		}
		if dirtyErr := s.cacheRepo.SetDirty(ctx, userID); dirtyErr != nil {
			log.Printf("failed to mark cart dirty for user %s: %v", userID, dirtyErr)
		}
		return nil
	}

	c, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCartNotFound
	}
	return s.repo.Delete(ctx, c)
}

// AddItem adds a product to the user's cart.
// If the user has no cart, one is created.
// If the product already exists in the cart, the quantity is incremented.
// Returns the updated cart.
func (s *CartService) AddItem(ctx context.Context, userID, productID string, quantity int) (*Cart, error) {
	if quantity < 1 {
		return nil, ErrInvalidQuantity
	}

	c, err := s.getCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create cart if it doesn't exist.
	if c == nil {
		id, err := shared.GenerateUUID()
		if err != nil {
			return nil, err
		}
		c = &Cart{
			ID:     id,
			UserID: userID,
			Items:  []CartItem{},
		}
	}

	// Check if product already exists in cart.
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			c.Items[i].Quantity += quantity
			if err := s.saveCart(ctx, c); err != nil {
				return nil, err
			}
			return c, nil
		}
	}

	// Product not in cart — add new item.
	c.Items = append(c.Items, CartItem{
		CartID:    c.ID,
		ProductID: productID,
		Quantity:  quantity,
	})

	if err := s.saveCart(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

// RemoveItem removes a product from the user's cart.
// Returns the updated cart.
func (s *CartService) RemoveItem(ctx context.Context, userID, productID string) (*Cart, error) {
	c, err := s.getCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCartNotFound
	}

	found := false
	updatedItems := make([]CartItem, 0, len(c.Items))
	for _, item := range c.Items {
		if item.ProductID == productID {
			found = true
			continue
		}
		updatedItems = append(updatedItems, item)
	}

	if !found {
		return nil, ErrItemNotFound
	}

	c.Items = updatedItems
	if err := s.saveCart(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

// UpdateItemQuantity updates the quantity of a specific item in the cart.
// Returns the updated cart.
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID, productID string, quantity int) (*Cart, error) {
	if quantity < 1 {
		return nil, ErrInvalidQuantity
	}

	c, err := s.getCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCartNotFound
	}

	found := false
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			c.Items[i].Quantity = quantity
			found = true
			break
		}
	}

	if !found {
		return nil, ErrItemNotFound
	}

	if err := s.saveCart(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}
