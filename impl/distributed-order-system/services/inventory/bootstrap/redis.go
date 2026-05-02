package bootstrap

import (
	sharedRedis "shared/redis"

	"github.com/redis/go-redis/v9"
)

// InitRedis creates the Redis client.
// The returned client is shared across all components that need claim check storage.
func InitRedis() *redis.Client {
	cfg := sharedRedis.DefaultConfig()
	return sharedRedis.NewClient(cfg)
}
