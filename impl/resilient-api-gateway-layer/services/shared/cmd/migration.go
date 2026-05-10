package main

import (
	"log"

	"shared/config"
	"shared/domain/cart"
	"shared/domain/order"
	"shared/domain/product"
	"shared/domain/user"
)

func main() {
	// Connect to PostgreSQL.
	db, err := config.ConnectPostgres(config.DefaultDatabaseConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("PostgreSQL connected")

	// Run auto-migration for all domain entities.
	if err := db.AutoMigrate(
		&user.User{},
		&user.UserPreferences{},
		&product.Product{},
		&cart.Cart{},
		&cart.CartItem{},
		&order.Order{},
		&order.OrderDetail{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migration completed successfully")
}
