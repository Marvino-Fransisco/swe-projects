package bootstrap

import (
	"log"

	"web-backend/controller"
	webMiddleware "web-backend/middleware"

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
	tokenStrategy := controller.NewCookieTokenStrategy()
	authCtrl := controller.NewAuthController(ucs.authUC, tokenStrategy)
	productCtrl := controller.NewProductController(ucs.productUC)
	cartCtrl := controller.NewCartController(ucs.cartUC)
	checkoutCtrl := controller.NewCheckoutController(ucs.checkoutUC)
	profileCtrl := controller.NewProfileController(ucs.profileUC)
	log.Println("Controllers initialized")

	authMW := webMiddleware.NewAuthMiddleware(infra.JwtSvc)

	return &controllers{
		authCtrl:     authCtrl,
		productCtrl:  productCtrl,
		cartCtrl:     cartCtrl,
		checkoutCtrl: checkoutCtrl,
		profileCtrl:  profileCtrl,
		authMW:       authMW,
	}
}
