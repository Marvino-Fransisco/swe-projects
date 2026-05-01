package bootstrap

import (
	"log"

	"payment-service/config"
	"payment-service/internal/adapters/dbrepository"

	"gorm.io/gorm"
)

func InitDatabase() *gorm.DB {
	cfg := config.DefaultConfig()
	db, err := config.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	repo := dbrepository.NewGormPaymentRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	return db
}
