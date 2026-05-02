package bootstrap

import (
	"shared/redis"
)

// InitRedis creates the Redis client using the shared module.
// The returned client is shared across all components that need claim check storage.
func InitRedis() *redis.Client {
	cfg := redis.DefaultConfig()
	return redis.NewClient(cfg)
}
