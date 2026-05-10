package bootstrap

import (
	"log"

	"web-backend/usecases/auth"
	usecaseCart "web-backend/usecases/cart"
	"web-backend/usecases/checkout"
	usecaseProduct "web-backend/usecases/product"
	"web-backend/usecases/profile"
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
	cartUC := usecaseCart.NewCartUseCase(svc.CartSvc, svc.ProductSvc)
	checkoutUC := checkout.NewCheckoutUseCase(svc.CartSvc, svc.OrderSvc, repos.OrderQueryRepo)
	profileUC := profile.NewProfileUseCase(svc.UserSvc, repos.UserCacheRepo, infra.DbTx, infra.RedisTx)
	log.Println("Use cases initialized")

	return &useCases{
		authUC:     authUC,
		productUC:  productUC,
		cartUC:     cartUC,
		checkoutUC: checkoutUC,
		profileUC:  profileUC,
	}
}
