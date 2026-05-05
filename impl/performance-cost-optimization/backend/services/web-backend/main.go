package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"shared/config"
	"shared/domain/cart"
	"shared/domain/order"
	"shared/domain/product"
	"shared/domain/user"
	"shared/util"

	webRepository "shared/repository"
	"web-backend/controller"
	webMiddleware "web-backend/middleware"
	"web-backend/usecases/auth"
	usecaseCart "web-backend/usecases/cart"
	"web-backend/usecases/checkout"
	usecaseProduct "web-backend/usecases/product"
	"web-backend/usecases/profile"
)

func main() {
	// Initialize PostgreSQL connection.
	pgDB, err := config.ConnectPostgres(config.DefaultDatabaseConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("PostgreSQL connected")

	// Initialize Redis client (reserved for future use).
	rdb := config.NewRedisClient(config.DefaultRedisConfig())
	log.Println("Redis client initialized")

	// Initialize JWT service.
	jwtSvc := util.NewJWTService(util.DefaultTokenConfig())
	log.Println("JWT service initialized")

	// --- Repositories (GORM-backed implementations) ---
	userRepo := webRepository.NewUserRepository(pgDB)
	productRepo := webRepository.NewProductRepository(pgDB, rdb)
	cartRepo := webRepository.NewCartRepository(pgDB)
	orderRepo := webRepository.NewOrderRepository(pgDB)
	log.Println("Repositories initialized")

	// --- Domain Services (from shared module) ---
	userSvc := user.NewUserService(userRepo)
	productSvc := product.NewProductService(productRepo)
	cartSvc := cart.NewCartService(cartRepo)
	orderSvc := order.NewOrderService(orderRepo)
	log.Println("Domain services initialized")

	// --- Use Cases ---
	authUC := auth.NewAuthUseCase(userSvc, jwtSvc)
	productUC := usecaseProduct.NewProductUseCase(productSvc)
	cartUC := usecaseCart.NewCartUseCase(cartSvc, productSvc)
	checkoutUC := checkout.NewCheckoutUseCase(cartSvc, orderSvc)
	profileUC := profile.NewProfileUseCase(userSvc)
	log.Println("Use cases initialized")

	// --- Controllers (HTTP handlers) ---
	tokenStrategy := controller.NewCookieTokenStrategy()
	authCtrl := controller.NewAuthController(authUC, tokenStrategy)
	productCtrl := controller.NewProductController(productUC)
	cartCtrl := controller.NewCartController(cartUC)
	checkoutCtrl := controller.NewCheckoutController(checkoutUC)
	profileCtrl := controller.NewProfileController(profileUC)
	log.Println("Controllers initialized")

	// --- Middleware ---
	authMiddleware := webMiddleware.NewAuthMiddleware(jwtSvc)

	// Create Gin engine.
	r := gin.Default()

	// Health check endpoint.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "web-backend"})
	})

	// Setup API routes.
	SetupRoutes(r, &Services{
		AuthCtrl:     authCtrl,
		ProductCtrl:  productCtrl,
		CartCtrl:     cartCtrl,
		CheckoutCtrl: checkoutCtrl,
		ProfileCtrl:  profileCtrl,
	}, authMiddleware)

	// Start server.
	log.Println("Web backend starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
