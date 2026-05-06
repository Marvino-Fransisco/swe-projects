package bootstrap

import (
	"log"

	"shared/domain/cart"
	"shared/domain/order"
	"shared/domain/product"
	"shared/domain/user"

	webRepository "shared/repository"
	"web-backend/repository"
)

// Repositories holds all repository instances.
type Repositories struct {
	UserRepo         user.UserRepository
	ProductRepo      product.ProductRepository
	CartRepo         cart.CartRepository
	OrderRepo        order.OrderRepository
	UserCacheRepo    user.UserCacheRepository
	CartCacheRepo    cart.CartCacheRepository
	ProductCacheRepo product.ProductCacheRepository
	ProductQueryRepo repository.ProductQueryRepository
	OrderQueryRepo   repository.OrderQueryRepository
}

// initRepositories creates all repository instances.
func initRepositories(infra *Infrastructure) *Repositories {
	// GORM-backed repositories.
	userRepo := webRepository.NewUserRepository(infra.PgDB)
	productRepo := webRepository.NewProductRepository(infra.PgDB)
	cartRepo := webRepository.NewCartRepository(infra.PgDB)
	orderRepo := webRepository.NewOrderRepository(infra.PgDB)
	log.Println("Repositories initialized")

	// Cache repositories (Redis).
	userCacheRepo := webRepository.NewUserCacheRepository(infra.Rdb)
	cartCacheRepo := webRepository.NewCartCacheRepository(infra.Rdb)
	productCacheRepo := webRepository.NewProductCacheRepository(infra.Rdb)
	log.Println("Cache repositories initialized")

	// Query repositories (read-only, paginated).
	productQueryRepo := repository.NewProductQueryRepository(infra.PgDB)
	orderQueryRepo := repository.NewOrderQueryRepository(infra.PgDB)
	log.Println("Query repositories initialized")

	return &Repositories{
		UserRepo:         userRepo,
		ProductRepo:      productRepo,
		CartRepo:         cartRepo,
		OrderRepo:        orderRepo,
		UserCacheRepo:    userCacheRepo,
		CartCacheRepo:    cartCacheRepo,
		ProductCacheRepo: productCacheRepo,
		ProductQueryRepo: productQueryRepo,
		OrderQueryRepo:   orderQueryRepo,
	}
}
