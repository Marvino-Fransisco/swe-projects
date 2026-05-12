package main

import (
	"log"
	"os"

	"order-service/bootstrap"
)

func main() {
	db := bootstrap.InitDatabase()

	redisClient := bootstrap.InitRedis()

	pub := bootstrap.InitPublisher(redisClient)
	defer pub.Close()

	application := bootstrap.InitApp(db, pub)

	con := bootstrap.InitRabbitMQ(application)
	defer con.Close()

	router := bootstrap.InitRouter(application)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	log.Println("Order service starting on :" + port)
	if err := router.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
