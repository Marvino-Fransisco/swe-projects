package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"shared/config"
	"shared/domain/cart"
)

// cartRepository implements cart.CartRepository using GORM.
type cartRepository struct {
	db *gorm.DB
}

// NewCartRepository creates a new GORM-backed CartRepository.
func NewCartRepository(db *gorm.DB) cart.CartRepository {
	return &cartRepository{db: db}
}

// FindByUserID retrieves the cart for a given user, preloading items and their products.
// Returns nil if not found.
func (r *cartRepository) FindByUserID(ctx context.Context, userID string) (*cart.Cart, error) {
	db := config.DBFromContext(ctx, r.db)
	var c cart.Cart
	err := db.WithContext(ctx).Preload("Items").Preload("Items.Product").Where("user_id = ?", userID).First(&c).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// Save persists a cart (create or update), including its item associations.
func (r *cartRepository) Save(ctx context.Context, c *cart.Cart) error {
	db := config.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Save(c).Error
}

// Delete removes a cart and its items.
func (r *cartRepository) Delete(ctx context.Context, c *cart.Cart) error {
	db := config.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Select("Items").Delete(c).Error
}

func (r *cartRepository) ReplaceCart(ctx context.Context, c *cart.Cart) error {
	db := config.DBFromContext(ctx, r.db)

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the cart row to prevent concurrent modifications.
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", c.UserID).
			Find(&cart.Cart{}).Error; err != nil {
			return err
		}

		// Delete existing items only — keep the parent carts row.
		if err := tx.
			Where("cart_id = ?", c.ID).
			Delete(&cart.CartItem{}).Error; err != nil {
			return err
		}

		// Insert new items.
		if len(c.Items) > 0 {
			for i := range c.Items {
				c.Items[i].CartID = c.ID
			}

			if err := tx.Create(&c.Items).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
