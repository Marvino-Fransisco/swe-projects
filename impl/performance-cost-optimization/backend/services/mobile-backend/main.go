package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"shared/config"
	"shared/domain/cart"
	"shared/domain/order"
	"shared/domain/product"
	"shared/domain/user"
	"shared/util"

	webRepository "shared/repository"
	"mobile-backend/controller"
	mobileMiddleware "mobile-backend/middleware"
	queryRepository "mobile-backend/repository"
	"mobile-backend/usecases/auth"
	usecaseCart "mobile-backend/usecases/cart"
	"mobile-backend/usecases/checkout"
	usecaseProduct "mobile-backend/usecases/product"
	"mobile-backend/usecases/profile"
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

	// Initialize JWT service with reduced access token TTL for mobile clients.
	tokenConfig := util.DefaultTokenConfig()
	tokenConfig.AccessTokenTTL = 5 * time.Minute
	jwtSvc := util.NewJWTService(tokenConfig)
	log.Println("JWT service initialized (access TTL: 5m)")

	// --- Repositories (GORM-backed implementations) ---
	userRepo := webRepository.NewUserRepository(pgDB)
	productRepo := webRepository.NewProductRepository(pgDB, rdb)
	cartRepo := webRepository.NewCartRepository(pgDB)
	orderRepo := webRepository.NewOrderRepository(pgDB)
	log.Println("Repositories initialized")

	// --- Query Repositories (mobile-backend read-side) ---
	cartQueryRepo := queryRepository.NewCartQueryRepository(pgDB)
	productQueryRepo := queryRepository.NewProductQueryRepository(pgDB)
	orderQueryRepo := queryRepository.NewOrderQueryRepository(pgDB)
	log.Println("Query repositories initialized")

	// --- Domain Services (from shared module) ---
	userSvc := user.NewUserService(userRepo)
	productSvc := product.NewProductService(productRepo)
	cartSvc := cart.NewCartService(cartRepo)
	orderSvc := order.NewOrderService(orderRepo)
	log.Println("Domain services initialized")

	// --- Use Cases ---
	authUC := auth.NewAuthUseCase(userSvc, jwtSvc)
	productUC := usecaseProduct.NewProductUseCase(productSvc, productQueryRepo)
	cartUC := usecaseCart.NewCartUseCase(cartSvc, productSvc, cartQueryRepo)
	checkoutUC := checkout.NewCheckoutUseCase(cartSvc, orderSvc, cartQueryRepo, orderQueryRepo)
	profileUC := profile.NewProfileUseCase(userSvc)
	log.Println("Use cases initialized")

	// --- Controllers (HTTP handlers) ---
	authCtrl := controller.NewAuthController(authUC)
	productCtrl := controller.NewProductController(productUC)
	cartCtrl := controller.NewCartController(cartUC)
	checkoutCtrl := controller.NewCheckoutController(checkoutUC)
	profileCtrl := controller.NewProfileController(profileUC)
	log.Println("Controllers initialized")

	// --- Middleware ---
	authMiddleware := mobileMiddleware.NewAuthMiddleware(jwtSvc)

	// Create Gin engine.
	r := gin.Default()

	// Health check endpoint.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "mobile-backend"})
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
	log.Println("Mobile backend starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
