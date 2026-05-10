package repository

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"shared/config"
	"shared/domain/user"
)

// userCacheRepository implements user.UserCacheRepository using Redis.
type userCacheRepository struct {
	rdb *goredis.Client
}

// NewUserCacheRepository creates a new Redis-backed UserCacheRepository.
func NewUserCacheRepository(rdb *goredis.Client) user.UserCacheRepository {
	return &userCacheRepository{rdb: rdb}
}

// userCacheKey returns the Redis key for a user cache entry.
func userCacheKey(id string) string {
	return "user:" + id
}

// GetByID retrieves a user from cache by their ID.
// Returns nil if not found in cache.
func (r *userCacheRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	rdb := config.RedisFromContext(ctx, r.rdb)
	key := userCacheKey(id)

	var u user.User
	err := rdb.HGetAll(ctx, key).Scan(&u)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from cache: %w", err)
	}
	if u.ID == "" {
		return nil, nil
	}

	return &u, nil
}

// Set persists a user to cache with a 7-day TTL.
func (r *userCacheRepository) Set(ctx context.Context, u *user.User) error {
	rdb := config.RedisFromContext(ctx, r.rdb)
	key := userCacheKey(u.ID)

	err := rdb.HSet(ctx, key, map[string]interface{}{
		"id":       u.ID,
		"fullName": u.FullName.String(),
		"email":    u.Email.String(),
		"address":  u.Address,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to cache user: %w", err)
	}

	err = rdb.Expire(ctx, key, 7*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache TTL: %w", err)
	}

	return nil
}

// Delete removes a user from cache by their ID.
func (r *userCacheRepository) Delete(ctx context.Context, id string) error {
	rdb := config.RedisFromContext(ctx, r.rdb)
	key := userCacheKey(id)

	err := rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user from cache: %w", err)
	}

	return nil
}
