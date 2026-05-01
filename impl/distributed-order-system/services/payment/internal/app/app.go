package app

import (
	"payment-service/internal/app/command"
)

// Application is the central entry point for all use cases.
// It groups all command handlers (write side).
type Application struct {
	Commands Commands
}

// Commands holds all write-side handlers.
type Commands struct {
	CreatePayment  command.CreatePaymentHandler
	ProcessPayment command.ProcessPaymentHandler
}

// NewApplication wires all handlers together.
func NewApplication(
	createPayment command.CreatePaymentHandler,
	processPayment command.ProcessPaymentHandler,
) *Application {
	return &Application{
		Commands: Commands{
			CreatePayment:  createPayment,
			ProcessPayment: processPayment,
		},
	}
}
