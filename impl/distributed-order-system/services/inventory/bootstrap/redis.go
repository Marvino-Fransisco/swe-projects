package bootstrap

import (
	"log"

	"inventory-service/config"

	"github.com/redis/go-redis/v9"
)

// InitRedis creates the Redis client.
// The returned client is shared across all components that need claim check storage.
func InitRedis() *redis.Client {
	cfg := config.DefaultRedisConfig()

	client := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
	})

	log.Printf("Redis client initialized for %s", cfg.Addr)
	return client
}
