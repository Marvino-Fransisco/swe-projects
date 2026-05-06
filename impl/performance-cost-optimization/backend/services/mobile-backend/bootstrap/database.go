package bootstrap

import (
	"log"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"shared/config"
	"shared/util"
)

// Infrastructure holds the core infrastructure clients initialized during bootstrap.
type Infrastructure struct {
	PgDB   *gorm.DB
	Rdb    *goredis.Client
	JwtSvc *util.JWTService
}

// initInfrastructure sets up PostgreSQL, Redis, and JWT.
func initInfrastructure() *Infrastructure {
	// PostgreSQL.
	pgDB, err := config.ConnectPostgres(config.DefaultDatabaseConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("PostgreSQL connected")

	// Redis client (reserved for future use).
	rdb := config.NewRedisClient(config.DefaultRedisConfig())
	log.Println("Redis client initialized")

	// JWT service with reduced access token TTL for mobile clients.
	tokenConfig := util.DefaultTokenConfig()
	tokenConfig.AccessTokenTTL = 5 * time.Minute
	jwtSvc := util.NewJWTService(tokenConfig)
	log.Println("JWT service initialized (access TTL: 5m)")

	return &Infrastructure{
		PgDB:   pgDB,
		Rdb:    rdb,
		JwtSvc: jwtSvc,
	}
}
