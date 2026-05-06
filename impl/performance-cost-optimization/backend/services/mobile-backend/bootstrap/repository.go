package bootstrap

import (
	"log"

	"shared/domain/cart"
	"shared/domain/order"
	"shared/domain/product"
	"shared/domain/user"

	webRepository "shared/repository"
	"mobile-backend/repository"
)

// Repositories holds all repository instances.
type Repositories struct {
	UserRepo         user.UserRepository
	ProductRepo      product.ProductRepository
	CartRepo         cart.CartRepository
	OrderRepo        order.OrderRepository
	CartQueryRepo    repository.CartQueryRepository
	ProductQueryRepo repository.ProductQueryRepository
	OrderQueryRepo   repository.OrderQueryRepository
}

// initRepositories creates all repository instances.
func initRepositories(infra *Infrastructure) *Repositories {
	// GORM-backed repositories.
	userRepo := webRepository.NewUserRepository(infra.PgDB)
	productRepo := webRepository.NewProductRepository(infra.PgDB, infra.Rdb)
	cartRepo := webRepository.NewCartRepository(infra.PgDB)
	orderRepo := webRepository.NewOrderRepository(infra.PgDB)
	log.Println("Repositories initialized")

	// Query repositories (read-side).
	cartQueryRepo := repository.NewCartQueryRepository(infra.PgDB)
	productQueryRepo := repository.NewProductQueryRepository(infra.PgDB)
	orderQueryRepo := repository.NewOrderQueryRepository(infra.PgDB)
	log.Println("Query repositories initialized")

	return &Repositories{
		UserRepo:         userRepo,
		ProductRepo:      productRepo,
		CartRepo:         cartRepo,
		OrderRepo:        orderRepo,
		CartQueryRepo:    cartQueryRepo,
		ProductQueryRepo: productQueryRepo,
		OrderQueryRepo:   orderQueryRepo,
	}
}
