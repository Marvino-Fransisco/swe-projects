package app

import (
	"order-service/internal/app/command"
	"order-service/internal/app/query"
)

// Application is the central entry point for all use cases.
// It groups all command handlers (write side) and query handlers (read side).
type Application struct {
	Commands Commands
	Queries  Queries
}

// Commands holds all write-side handlers.
type Commands struct {
	CreateOrder       command.CreateOrderHandler
	UpdateOrderStatus command.UpdateOrderStatusHandler
	FailOrder         command.FailOrderHandler
}

// Queries holds all read-side handlers.
type Queries struct {
	GetOrder   query.GetOrderHandler
	ListOrders query.ListOrdersHandler
}

// NewApplication wires all handlers together.
func NewApplication(
	createOrder command.CreateOrderHandler,
	updateOrderStatus command.UpdateOrderStatusHandler,
	failOrder command.FailOrderHandler,
	getOrder query.GetOrderHandler,
	listOrders query.ListOrdersHandler,
) *Application {
	return &Application{
		Commands: Commands{
			CreateOrder:       createOrder,
			UpdateOrderStatus: updateOrderStatus,
			FailOrder:         failOrder,
		},
		Queries: Queries{
			GetOrder:   getOrder,
			ListOrders: listOrders,
		},
	}
}
