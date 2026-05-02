package bootstrap

import (
	"inventory-service/internal/adapters/dbrepository"
	"inventory-service/internal/adapters/http"
	"inventory-service/internal/app"
	"inventory-service/internal/app/command"
	"inventory-service/internal/app/query"
	"inventory-service/messaging/publisher"
	"inventory-service/routes"

	sharedTx "shared/tx"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitApp wires the entire application together: repositories, read models,
// command/query handlers, and returns the central Application struct.
func InitApp(db *gorm.DB, pub *publisher.Publisher) *app.Application {
	repo := dbrepository.NewGormInventoryRepository(db)
	readModel := dbrepository.NewGormInventoryReadModel(db)
	dbTx := sharedTx.NewDBTransaction(db)

	reserveStockHandler := command.NewReserveStockHandler(repo, pub, dbTx)
	completeReservationHandler := command.NewCompleteReservationHandler(repo)
	cancelReservationHandler := command.NewCancelReservationHandler(repo)
	listInventoriesHandler := query.NewListInventoriesHandler(readModel)

	return app.NewApplication(
		reserveStockHandler,
		completeReservationHandler,
		cancelReservationHandler,
		listInventoriesHandler,
	)
}

// InitRouter creates the Gin engine and sets up HTTP routes.
func InitRouter(application *app.Application) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	inventoryHandler := http.NewInventoryHandler(application)
	routes.SetupRoutes(router, inventoryHandler)

	return router
}
