package bootstrap

import (
	goredis "github.com/redis/go-redis/v9"
	"shared/redis"
)

// InitRedis creates the Redis client using the shared module.
// The returned client is shared across all components that need claim check storage.
func InitRedis() *goredis.Client {
	cfg := redis.DefaultConfig()
	return redis.NewClient(cfg)
}
