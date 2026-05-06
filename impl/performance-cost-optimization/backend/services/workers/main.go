package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"shared/config"
	"shared/domain/cart"
	"shared/domain/product"
	"shared/workers"
	webRepository "shared/repository"
)

func main() {
	pgDB, err := config.ConnectPostgres(config.DefaultDatabaseConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("PostgreSQL connected")

	rdb := config.NewRedisClient(config.DefaultRedisConfig())
	log.Println("Redis client initialized")

	cartRepo := webRepository.NewCartRepository(pgDB)
	cartCacheRepo := webRepository.NewCartCacheRepository(rdb)
	productRepo := webRepository.NewProductRepository(pgDB)
	productCacheRepo := webRepository.NewProductCacheRepository(rdb)

	cartSvc := cart.NewCachedCartService(cartRepo, cartCacheRepo)
	productSvc := product.NewProductService(productRepo, productCacheRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go workers.RunCacheWarmer(ctx, pgDB, rdb)
	go workers.RunCartSync(ctx, cartSvc, cartCacheRepo)
	go workers.RunProductViewSync(ctx, productSvc)

	log.Println("All workers started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down workers...")
	cancel()
}
