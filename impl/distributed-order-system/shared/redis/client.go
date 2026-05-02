package redis

import (
	"log"
	"os"

	goredis "github.com/redis/go-redis/v9"
)

// RedisConfig holds the Redis connection configuration.
type RedisConfig struct {
	Addr string
}

// DefaultConfig returns the default Redis configuration.
// It reads REDIS_ADDR from the environment, falling back to "redis:6379".
func DefaultConfig() *RedisConfig {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}
	return &RedisConfig{
		Addr: addr,
	}
}

// NewClient creates a new Redis client using the provided configuration.
func NewClient(cfg *RedisConfig) *goredis.Client {
	client := goredis.NewClient(&goredis.Options{
		Addr: cfg.Addr,
	})

	log.Printf("Redis client initialized for %s", cfg.Addr)
	return client
}
