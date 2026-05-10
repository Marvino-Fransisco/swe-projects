package bootstrap

import (
	"log"

	"shared/domain/cart"
	"shared/domain/order"
	"shared/domain/product"
	"shared/domain/user"
)

// DomainServices holds all domain service instances.
type DomainServices struct {
	UserSvc    *user.UserService
	ProductSvc *product.ProductService
	CartSvc    *cart.CartService
	OrderSvc   *order.OrderService
}

// initDomainServices creates all domain service instances.
func initDomainServices(repos *Repositories) *DomainServices {
	userSvc := user.NewUserService(repos.UserRepo)
	productSvc := product.NewProductService(repos.ProductRepo, repos.ProductCacheRepo)
	cartSvc := cart.NewCachedCartService(repos.CartRepo, repos.CartCacheRepo)
	orderSvc := order.NewOrderService(repos.OrderRepo)
	log.Println("Domain services initialized")

	return &DomainServices{
		UserSvc:    userSvc,
		ProductSvc: productSvc,
		CartSvc:    cartSvc,
		OrderSvc:   orderSvc,
	}
}
