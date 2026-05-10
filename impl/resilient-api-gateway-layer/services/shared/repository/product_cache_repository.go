package repository

import (
	"context"
	"fmt"
	"strconv"

	goredis "github.com/redis/go-redis/v9"

	"shared/config"
	"shared/domain/product"
)

// productCacheRepository implements product.ProductCacheRepository using Redis.
type productCacheRepository struct {
	rdb *goredis.Client
}

// NewProductCacheRepository creates a new Redis-backed ProductCacheRepository.
func NewProductCacheRepository(rdb *goredis.Client) product.ProductCacheRepository {
	return &productCacheRepository{rdb: rdb}
}

// productViewCountHashKey returns the Redis hash key used for storing all product view counters.
func productViewCountHashKey() string {
	return "product:view_counts"
}

// IncrementViewCount increments the view count for a product by 1 in cache.
func (r *productCacheRepository) IncrementViewCount(ctx context.Context, productID string) error {
	rdb := config.RedisFromContext(ctx, r.rdb)
	err := rdb.HIncrBy(ctx, productViewCountHashKey(), productID, 1).Err()
	if err != nil {
		return fmt.Errorf("failed to increment view count in cache: %w", err)
	}
	return nil
}

// GetViewCount retrieves the cached view count for a product.
// Returns 0 if not found in cache.
func (r *productCacheRepository) GetViewCount(ctx context.Context, productID string) (int64, error) {
	rdb := config.RedisFromContext(ctx, r.rdb)
	val, err := rdb.HGet(ctx, productViewCountHashKey(), productID).Result()
	if err != nil {
		if err == goredis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get view count from cache: %w", err)
	}
	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse view count from cache: %w", err)
	}
	return count, nil
}

// GetAllViewCounts retrieves all product view counts from cache.
// Returns a map of productID -> view count.
func (r *productCacheRepository) GetAllViewCounts(ctx context.Context) (map[string]int64, error) {
	rdb := config.RedisFromContext(ctx, r.rdb)
	result, err := rdb.HGetAll(ctx, productViewCountHashKey()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all view counts from cache: %w", err)
	}

	viewCounts := make(map[string]int64, len(result))
	for productID, val := range result {
		count, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse view count for product %s: %w", productID, err)
		}
		viewCounts[productID] = count
	}

	return viewCounts, nil
}

// ResetAllViewCounts deletes all product view counters from cache.
func (r *productCacheRepository) ResetAllViewCounts(ctx context.Context) error {
	rdb := config.RedisFromContext(ctx, r.rdb)
	err := rdb.Del(ctx, productViewCountHashKey()).Err()
	if err != nil {
		return fmt.Errorf("failed to reset view counts in cache: %w", err)
	}
	return nil
}
