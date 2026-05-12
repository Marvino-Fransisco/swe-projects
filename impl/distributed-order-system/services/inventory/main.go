package main

import (
	"log"
	"os"

	"inventory-service/bootstrap"
)

func main() {
	db := bootstrap.InitDatabase()

	redisClient := bootstrap.InitRedis()

	pub := bootstrap.InitPublisher()
	defer pub.Close()

	application := bootstrap.InitApp(db, pub)

	con := bootstrap.InitRabbitMQ(application, redisClient)
	defer con.Close()

	router := bootstrap.InitRouter(application)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	log.Println("Inventory service starting on 0.0.0.0:" + port)
	if err := router.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
