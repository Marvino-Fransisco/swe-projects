package bootstrap

import (
	"order-service/internal/adapters/http"
	"order-service/internal/app"
	"order-service/internal/app/command"
	"order-service/internal/app/query"
	"order-service/internal/adapters/dbrepository"
	"order-service/messaging/publisher"
	"order-service/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitApp wires the entire application together: repositories, read models,
// command/query handlers, and returns the central Application struct.
func InitApp(db *gorm.DB, pub *publisher.Publisher) *app.Application {
	repo := dbrepository.NewGormOrderRepository(db)
	readModel := dbrepository.NewGormOrderReadModel(db)

	failOrderHandler := command.NewFailOrderHandler(repo)
	createOrderHandler := command.NewCreateOrderHandler(repo, pub, failOrderHandler)
	updateOrderStatusHandler := command.NewUpdateOrderStatusHandler(repo)
	getOrderHandler := query.NewGetOrderHandler(readModel)
	listOrdersHandler := query.NewListOrdersHandler(readModel)

	return app.NewApplication(
		createOrderHandler,
		updateOrderStatusHandler,
		failOrderHandler,
		getOrderHandler,
		listOrdersHandler,
	)
}

// InitRouter creates the Gin engine and sets up HTTP routes.
func InitRouter(application *app.Application) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	orderHandler := http.NewOrderHandler(application)
	routes.SetupRoutes(router, orderHandler)

	return router
}
