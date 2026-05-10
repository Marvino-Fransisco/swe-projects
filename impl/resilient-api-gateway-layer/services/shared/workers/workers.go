package workers

import (
	"context"
	"log"
	"shared/domain/cart"
	"shared/domain/product"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RunCacheWarmer(ctx context.Context, db *gorm.DB, rdb *redis.Client) {
	if err := cacheWarmer(ctx, db, rdb); err != nil {
		log.Printf("cache warm failed: %v", err)
	}
	log.Println("wrote product into cache")

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := cacheWarmer(ctx, db, rdb); err != nil {
				log.Printf("cache warm failed: %v", err)
			}
			log.Println("wrote product into cache")
		case <-ctx.Done():
			log.Println("worker CacheWarmer is done")
			return
		}
	}
}

func RunCartSync(ctx context.Context, cartSvc *cart.CartService, cartCacheRepo cart.CartCacheRepository) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := CartSync(ctx, cartSvc, cartCacheRepo); err != nil {
				log.Printf("Cart Sync failed: %v", err)
			}
			log.Println("Cart synced from cache to database")
		case <-ctx.Done():
			log.Println("worker CartSync is done")
			return
		}
	}
}

func RunProductViewSync(ctx context.Context, productSvc *product.ProductService) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := ProductViewSync(ctx, productSvc); err != nil {
				log.Printf("Product View Sync failed: %v", err)
			}
			log.Println("Product view counts synced from cache to database")
		case <-ctx.Done():
			log.Println("worker ProductViewSync is done")
			return
		}
	}
}
