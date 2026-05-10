package bootstrap

import (
	"log"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"shared/config"
	"shared/util"
)

// Infrastructure holds the core infrastructure clients initialized during bootstrap.
type Infrastructure struct {
	PgDB    *gorm.DB
	Rdb     *goredis.Client
	DbTx    config.DBTransaction
	RedisTx config.RedisTransaction
	JwtSvc  *util.JWTService
}

// initInfrastructure sets up PostgreSQL, Redis, transactions, and JWT.
func initInfrastructure() *Infrastructure {
	// PostgreSQL.
	pgDB, err := config.ConnectPostgres(config.DefaultDatabaseConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("PostgreSQL connected")

	// Redis.
	rdb := config.NewRedisClient(config.DefaultRedisConfig())
	log.Println("Redis client initialized")

	// Transactions.
	dbTx := config.NewDBTransaction(pgDB)
	redisTx := config.NewRedisTransaction(rdb)
	log.Println("Database and Redis transaction initialized")

	// JWT.
	jwtSvc := util.NewJWTService(util.DefaultTokenConfig())
	log.Println("JWT service initialized")

	return &Infrastructure{
		PgDB:    pgDB,
		Rdb:     rdb,
		DbTx:    dbTx,
		RedisTx: redisTx,
		JwtSvc:  jwtSvc,
	}
}
