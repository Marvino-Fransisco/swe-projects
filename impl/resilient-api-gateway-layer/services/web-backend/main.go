package main

import (
	"log"

	"web-backend/bootstrap"
)

func main() {
	app := bootstrap.Boot()
	r := app.SetupEngine()

	log.Println("Web backend starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
