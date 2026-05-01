package app

import (
	"inventory-service/internal/app/command"
	"inventory-service/internal/app/query"
)

// Application is the central entry point for all use cases.
// It groups all command handlers (write side) and query handlers (read side).
type Application struct {
	Commands Commands
	Queries  Queries
}

// Commands holds all write-side handlers.
type Commands struct {
	ReserveStock         command.ReserveStockHandler
	CompleteReservation  command.CompleteReservationHandler
	CancelReservation    command.CancelReservationHandler
}

// Queries holds all read-side handlers.
type Queries struct {
	ListInventories query.ListInventoriesHandler
}

// NewApplication wires all handlers together.
func NewApplication(
	reserveStock command.ReserveStockHandler,
	completeReservation command.CompleteReservationHandler,
	cancelReservation command.CancelReservationHandler,
	listInventories query.ListInventoriesHandler,
) *Application {
	return &Application{
		Commands: Commands{
			ReserveStock:        reserveStock,
			CompleteReservation: completeReservation,
			CancelReservation:   cancelReservation,
		},
		Queries: Queries{
			ListInventories: listInventories,
		},
	}
}
