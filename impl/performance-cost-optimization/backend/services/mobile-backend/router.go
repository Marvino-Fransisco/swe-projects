package main

import (
	"github.com/gin-gonic/gin"

	"mobile-backend/controller"

	sharedMiddleware "shared/middleware"
)

// Services holds all controllers needed by the router.
type Services struct {
	AuthCtrl     controller.AuthController
	ProductCtrl  controller.ProductController
	CartCtrl     controller.CartController
	CheckoutCtrl controller.CheckoutController
	ProfileCtrl  controller.ProfileController
}

// SetupRoutes registers all /api/v1 routes grouped by domain,
// wiring controllers with the appropriate middleware.
func SetupRoutes(r *gin.Engine, svc *Services, authMW sharedMiddleware.AuthMiddleware) {
	v1 := r.Group("/api/v1")

	// --- Auth routes (public) ---
	auth := v1.Group("/auth")
	{
		auth.POST("/register", svc.AuthCtrl.Register)
		auth.POST("/login", svc.AuthCtrl.Login)
		auth.POST("/refresh", svc.AuthCtrl.Refresh)
		auth.POST("/logout", svc.AuthCtrl.Logout)
	}

	// --- Product routes (public read, with optional auth) ---
	products := v1.Group("/products")
	{
		products.Use(authMW.OptionalAuth())
		products.GET("", svc.ProductCtrl.List)
		products.GET("/search", svc.ProductCtrl.Search)
		products.GET("/categories", svc.ProductCtrl.GetCategories)
		products.GET("/:id", svc.ProductCtrl.GetByID)
		products.POST("/:id/view", svc.ProductCtrl.TrackView)
	}

	// --- Protected routes (require auth) ---
	protected := v1.Group("")
	protected.Use(authMW.RequireAuth())
	{
		// --- Cart routes ---
		cart := protected.Group("/cart")
		{
			cart.GET("", svc.CartCtrl.GetCart)
			cart.POST("/items", svc.CartCtrl.AddItem)
			cart.PUT("/items/:productId", svc.CartCtrl.UpdateItemQuantity)
			cart.DELETE("/items/:productId", svc.CartCtrl.RemoveItem)
		}

		// --- Checkout routes ---
		checkout := protected.Group("/checkout")
		{
			checkout.POST("/orders", svc.CheckoutCtrl.PlaceOrder)
			checkout.GET("/orders", svc.CheckoutCtrl.GetOrderHistory)
			checkout.GET("/orders/:id", svc.CheckoutCtrl.GetOrder)
		}

		// --- Profile routes ---
		profile := protected.Group("/profile")
		{
			profile.GET("", svc.ProfileCtrl.GetProfile)
			profile.PUT("", svc.ProfileCtrl.UpdateProfile)
			profile.PUT("/password", svc.ProfileCtrl.ChangePassword)
		}
	}
}
