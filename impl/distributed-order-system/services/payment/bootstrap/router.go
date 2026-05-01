package bootstrap

import (
	"payment-service/internal/adapters/dbrepository"
	"payment-service/internal/adapters/http"
	"payment-service/internal/app"
	"payment-service/internal/app/command"
	"payment-service/messaging/publisher"
	"payment-service/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitApp wires the entire application together: repositories,
// command handlers, and returns the central Application struct.
func InitApp(db *gorm.DB, pub *publisher.Publisher) *app.Application {
	repo := dbrepository.NewGormPaymentRepository(db)

	createPaymentHandler := command.NewCreatePaymentHandler(repo)
	processPaymentHandler := command.NewProcessPaymentHandler(repo, pub)

	return app.NewApplication(
		createPaymentHandler,
		processPaymentHandler,
	)
}

// InitRouter creates the Gin engine and sets up HTTP routes.
func InitRouter(application *app.Application) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	paymentHandler := http.NewPaymentHandler(application)
	routes.SetupRoutes(router, paymentHandler)

	return router
}
