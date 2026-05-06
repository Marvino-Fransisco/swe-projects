package bootstrap

import (
	"log"

	"mobile-backend/controller"
	mobileMiddleware "mobile-backend/middleware"

	sharedMiddleware "shared/middleware"
)

// controllers holds all controller and middleware instances.
type controllers struct {
	authCtrl     controller.AuthController
	productCtrl  controller.ProductController
	cartCtrl     controller.CartController
	checkoutCtrl controller.CheckoutController
	profileCtrl  controller.ProfileController
	authMW       sharedMiddleware.AuthMiddleware
}

// initControllers creates all controller and middleware instances.
func initControllers(infra *Infrastructure, ucs *useCases) *controllers {
	authCtrl := controller.NewAuthController(ucs.authUC)
	productCtrl := controller.NewProductController(ucs.productUC)
	cartCtrl := controller.NewCartController(ucs.cartUC)
	checkoutCtrl := controller.NewCheckoutController(ucs.checkoutUC)
	profileCtrl := controller.NewProfileController(ucs.profileUC)
	log.Println("Controllers initialized")

	authMW := mobileMiddleware.NewAuthMiddleware(infra.JwtSvc)

	return &controllers{
		authCtrl:     authCtrl,
		productCtrl:  productCtrl,
		cartCtrl:     cartCtrl,
		checkoutCtrl: checkoutCtrl,
		profileCtrl:  profileCtrl,
		authMW:       authMW,
	}
}
