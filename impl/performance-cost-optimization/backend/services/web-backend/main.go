package main

import (
	"context"
	"log"

	"web-backend/bootstrap"
	"web-backend/workers"
)

func main() {
	app := bootstrap.Boot()
	r := app.SetupEngine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go workers.RunCacheWarmer(ctx, app.Infrastructure.PgDB, app.Infrastructure.Rdb)
	go workers.RunCartSync(ctx, app.DomainServices.CartSvc, app.Repositories.CartCacheRepo)
	go workers.RunProductViewSync(ctx, app.DomainServices.ProductSvc)

	log.Println("Web backend starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
