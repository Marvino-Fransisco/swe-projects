package main

import (
	"log"

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

	log.Println("Order service starting on :8002")
	if err := router.Run("0.0.0.0:8002"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
