package bootstrap

import (
	"log"

	"payment-service/internal/adapters/dbrepository"

	sharedDB "shared/db"

	"gorm.io/gorm"
)

func InitDatabase() *gorm.DB {
	cfg := sharedDB.DefaultConfig()
	db, err := sharedDB.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	repo := dbrepository.NewGormPaymentRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	return db
}
