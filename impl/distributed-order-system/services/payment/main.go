package main

import (
	"log"

	"payment-service/bootstrap"
)

func main() {
	db := bootstrap.InitDatabase()

	pub := bootstrap.InitPublisher()
	defer pub.Close()

	application := bootstrap.InitApp(db, pub)

	con := bootstrap.InitRabbitMQ(application)
	defer con.Close()

	router := bootstrap.InitRouter(application)

	log.Println("Payment service starting on :8003")
	if err := router.Run("0.0.0.0:8003"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
