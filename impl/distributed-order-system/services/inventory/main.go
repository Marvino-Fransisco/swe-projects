package main

import (
	"log"

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

	log.Println("Inventory service starting on 0.0.0.0:8001")
	if err := router.Run("0.0.0.0:8001"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
