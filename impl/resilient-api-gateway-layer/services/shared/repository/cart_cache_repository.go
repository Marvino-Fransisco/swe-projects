package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"shared/config"
	"shared/domain/cart"
)

// cartCacheRepository implements cart.CartCacheRepository using Redis.
type cartCacheRepository struct {
	rdb *goredis.Client
}

// NewCartCacheRepository creates a new Redis-backed CartCacheRepository.
func NewCartCacheRepository(rdb *goredis.Client) cart.CartCacheRepository {
	return &cartCacheRepository{rdb: rdb}
}

// cartCacheKey returns the Redis key for a user's cart cache entry.
func cartCacheKey(userID string) string {
	return "cart:" + userID
}

func cartCacheKeyDirty() string {
	return "cart:dirty"
}

// GetByUserID retrieves the user's cart from cache.
// Returns nil if not found in cache.
func (r *cartCacheRepository) GetByUserID(ctx context.Context, userID string) (*cart.Cart, error) {
	rdb := config.RedisFromContext(ctx, r.rdb)
	key := cartCacheKey(userID)

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cart from cache: %w", err)
	}

	var c cart.Cart
	if err := json.Unmarshal([]byte(val), &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart cache: %w", err)
	}

	return &c, nil
}

func (r *cartCacheRepository) GetCartDirtyMembers(ctx context.Context) ([]string, error) {
	rdb := config.RedisFromContext(ctx, r.rdb)
	dirtyKey := cartCacheKeyDirty()

	userIDs, err := rdb.SMembers(ctx, dirtyKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cart dirty members from cache: %w", err)
	}

	return userIDs, nil
}

// Set persists the user's cart to cache with a 1-day TTL.
func (r *cartCacheRepository) Set(ctx context.Context, userID string, c *cart.Cart) error {
	rdb := config.RedisFromContext(ctx, r.rdb)
	key := cartCacheKey(userID)

	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal cart cache: %w", err)
	}

	err = rdb.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache cart: %w", err)
	}

	return nil
}

func (r *cartCacheRepository) SetDirty(ctx context.Context, userID string) error {
	rdb := config.RedisFromContext(ctx, r.rdb)
	dirtyKey := cartCacheKeyDirty()

	err := rdb.SAdd(ctx, dirtyKey, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to cache cart dirty: %w", err)
	}

	return nil
}

// Delete removes the user's cart from cache.
func (r *cartCacheRepository) Delete(ctx context.Context, userID string) error {
	rdb := config.RedisFromContext(ctx, r.rdb)
	key := cartCacheKey(userID)

	err := rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cart from cache: %w", err)
	}

	return nil
}

func (r *cartCacheRepository) DeleteCartDirtyMember(ctx context.Context, userID string) error {
	rdb := config.RedisFromContext(ctx, r.rdb)
	dirtyKey := cartCacheKeyDirty()

	err := rdb.SRem(ctx, dirtyKey, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cart dirty member from cache: %w", err)
	}

	return nil
}
