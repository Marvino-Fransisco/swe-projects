package main

import (
	"log"

	"mobile-backend/bootstrap"
)

func main() {
	app := bootstrap.Boot()
	r := app.SetupEngine()

	log.Println("Mobile backend starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
