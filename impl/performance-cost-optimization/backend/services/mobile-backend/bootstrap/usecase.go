package bootstrap

import (
	"log"

	"mobile-backend/usecases/auth"
	usecaseCart "mobile-backend/usecases/cart"
	"mobile-backend/usecases/checkout"
	usecaseProduct "mobile-backend/usecases/product"
	"mobile-backend/usecases/profile"
)

// useCases holds all use case instances.
type useCases struct {
	authUC     auth.AuthUseCase
	productUC  usecaseProduct.ProductUseCase
	cartUC     usecaseCart.CartUseCase
	checkoutUC checkout.CheckoutUseCase
	profileUC  profile.ProfileUseCase
}

// initUseCases creates all use case instances.
func initUseCases(infra *Infrastructure, repos *Repositories, svc *DomainServices) *useCases {
	authUC := auth.NewAuthUseCase(svc.UserSvc, infra.JwtSvc)
	productUC := usecaseProduct.NewProductUseCase(svc.ProductSvc, repos.ProductQueryRepo)
	cartUC := usecaseCart.NewCartUseCase(svc.CartSvc, svc.ProductSvc, repos.CartQueryRepo)
	checkoutUC := checkout.NewCheckoutUseCase(svc.CartSvc, svc.OrderSvc, repos.CartQueryRepo, repos.OrderQueryRepo)
	profileUC := profile.NewProfileUseCase(svc.UserSvc)
	log.Println("Use cases initialized")

	return &useCases{
		authUC:     authUC,
		productUC:  productUC,
		cartUC:     cartUC,
		checkoutUC: checkoutUC,
		profileUC:  profileUC,
	}
}
