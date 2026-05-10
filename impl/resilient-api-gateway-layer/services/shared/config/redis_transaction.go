package config

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

// redisTxKey is the unexported context key type used to store the transaction-scoped Redis pipeline.
type redisTxKey struct{}

// RedisTxKey is the exported context key for accessing the Redis transaction from context.
var RedisTxKey redisTxKey

// RedisTransaction executes a function within a Redis transaction pipeline (MULTI/EXEC).
// All Redis commands within fn are queued and executed atomically when fn returns without error.
// The pipeline is stored in the context so that repository methods
// automatically participate in the transaction via RedisFromContext.
type RedisTransaction func(ctx context.Context, fn func(ctx context.Context) error) error

// NewRedisTransaction creates a RedisTransaction using a Redis client.
//
// Usage:
//
//	rtx := config.NewRedisTransaction(redisClient)
//	err := rtx(ctx, func(txCtx context.Context) error {
//	    // All Redis commands within this function will be queued in a TxPipeline
//	    // and executed atomically via MULTI/EXEC when the function returns without error.
//	    rdb := config.RedisFromContext(txCtx, defaultClient)
//	    return rdb.Set(txCtx, "key", "value", 0).Err()
//	})
func NewRedisTransaction(client *goredis.Client) RedisTransaction {
	return func(ctx context.Context, fn func(ctx context.Context) error) error {
		pipe := client.TxPipeline()
		txCtx := context.WithValue(ctx, RedisTxKey, pipe)

		if err := fn(txCtx); err != nil {
			return err
		}

		_, err := pipe.Exec(ctx)
		return err
	}
}

// RedisFromContext returns the transaction-scoped Redis pipeline from the context,
// or falls back to the provided default client if no transaction is active.
//
// Repository methods should use this to transparently participate in Redis transactions:
//
//	func (r *MyRepo) Increment(ctx context.Context, key string) error {
//	    rdb := config.RedisFromContext(ctx, r.redis)
//	    return rdb.Incr(ctx, key).Err()
//	}
func RedisFromContext(ctx context.Context, defaultClient *goredis.Client) goredis.Cmdable {
	if pipe, ok := ctx.Value(RedisTxKey).(goredis.Pipeliner); ok {
		return pipe
	}
	return defaultClient
}
