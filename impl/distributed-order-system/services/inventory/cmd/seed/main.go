package main

import (
	"fmt"
	"log"
	"math/rand"

	"inventory-service/config"
	"inventory-service/internal/adapters/dbrepository"
	"inventory-service/internal/domain/inventory"

	"github.com/go-faker/faker/v4"
)

func main() {
	// Connect to database
	cfg := config.DefaultConfig()
	db, err := config.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate to ensure the table exists
	repo := dbrepository.NewGormInventoryRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Define status options to randomly pick from
	statuses := []inventory.InventoryStatus{
		inventory.StatusAvailable,
		inventory.StatusOutOfStock,
		inventory.StatusLowStock,
	}

	// Define shirt name components
	adjectives := []string{
		"Classic", "Slim Fit", "Oversized", "Vintage", "Essential",
		"Premium", "Relaxed", "Graphic", "Striped", "Plain",
		"Cropped", "Fitted", "Heavyweight", "Lightweight", "Retro",
	}
	shirtTypes := []string{
		"Crew Neck Tee", "Polo Shirt", "V-Neck Tee", "Henley Shirt",
		"Oxford Shirt", "Flannel Shirt", "Hoodie", "Sweatshirt",
		"Tank Top", "Button-Down Shirt", "Pullover", "Long Sleeve Tee",
		"Pocket Tee", "Chambray Shirt", "Thermal Shirt",
	}

	// Generate 100 fake inventory records
	// We need to use the GORM database directly for seeding since the domain
	// uses auto-increment IDs and doesn't expose a bulk-create method.
	type seedModel struct {
		ProductID   string  `gorm:"column:product_id;not null"`
		ProductName string  `gorm:"column:product_name;not null"`
		Stock       int     `gorm:"column:stock;not null;default:0"`
		Price       float64 `gorm:"column:price;not null;default:0"`
		Status      string  `gorm:"column:status;not null;default:available"`
	}

	inventories := make([]seedModel, 0, 100)
	for i := 0; i < 100; i++ {
		productName := fmt.Sprintf("%s %s",
			adjectives[rand.Intn(len(adjectives))],
			shirtTypes[rand.Intn(len(shirtTypes))],
		)

		inventories = append(inventories, seedModel{
			ProductID:   faker.UUIDHyphenated(),
			ProductName: productName,
			Stock:       rand.Intn(500),
			Price:       float64(rand.Intn(100000)) / 100.0,
			Status:      statuses[rand.Intn(len(statuses))].String(),
		})
	}

	// Insert all records into the database
	if err := db.Table("inventories").Create(&inventories).Error; err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	fmt.Printf("Successfully seeded %d inventory records\n", len(inventories))
}
