package config

import "os"

type RedisConfig struct {
	Addr string
}

func DefaultRedisConfig() *RedisConfig {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}
	return &RedisConfig{
		Addr: addr,
	}
}
