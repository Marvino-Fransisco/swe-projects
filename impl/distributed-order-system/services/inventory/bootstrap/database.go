package bootstrap

import (
	"log"

	"inventory-service/config"
	"inventory-service/internal/adapters/dbrepository"

	"gorm.io/gorm"
)

func InitDatabase() *gorm.DB {
	cfg := config.DefaultConfig()
	db, err := config.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	repo := dbrepository.NewGormInventoryRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	return db
}
