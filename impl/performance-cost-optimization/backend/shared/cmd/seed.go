package main

import (
	"log"

	"shared/config"
	"shared/domain/product"

	"gorm.io/gorm"
)

func main() {
	db, err := config.ConnectPostgres(config.DefaultDatabaseConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("PostgreSQL connected")

	seedProducts(db)
}

func seedProducts(db *gorm.DB) {
	var count int64
	db.Model(&product.Product{}).Count(&count)
	if count > 0 {
		log.Printf("Products table already has %d rows, skipping seed", count)
		return
	}

	products := buildSeedProducts()

	if err := db.Create(&products).Error; err != nil {
		log.Fatalf("Failed to seed products: %v", err)
	}

	log.Printf("Seeded %d products successfully", len(products))
}

func buildSeedProducts() []product.Product {
	type seed struct {
		name        string
		description string
		price       float64
		stock       int
		status      product.ProductStatus
	}

	seeds := []seed{
		{"Wireless Mouse", "Ergonomic wireless mouse with adjustable DPI", 29.99, 150, product.ProductStatusInsufficient},
		{"Mechanical Keyboard", "RGB mechanical keyboard with Cherry MX switches", 89.99, 75, product.ProductStatusInsufficient},
		{"USB-C Hub", "7-in-1 USB-C hub with HDMI, USB 3.0, and SD card reader", 49.99, 200, product.ProductStatusInsufficient},
		{"Monitor Stand", "Adjustable monitor stand with built-in cable management", 39.99, 50, product.ProductStatusDanger},
		{"Webcam HD", "1080p HD webcam with auto-focus and noise-canceling mic", 59.99, 120, product.ProductStatusInsufficient},
		{"Laptop Sleeve", "Neoprene laptop sleeve for 14-inch laptops", 19.99, 300, product.ProductStatusInsufficient},
		{"Desk Lamp", "LED desk lamp with adjustable brightness and color temperature", 34.99, 80, product.ProductStatusDanger},
		{"Noise-Canceling Headphones", "Over-ear headphones with active noise cancellation", 149.99, 30, product.ProductStatusDanger},
		{"Portable SSD 1TB", "Compact portable SSD with USB 3.2 interface", 109.99, 60, product.ProductStatusDanger},
		{"Ergonomic Chair", "Mesh-back ergonomic office chair with lumbar support", 299.99, 15, product.ProductStatusDanger},
		{"Drawing Tablet", "Pen display tablet with 8192 levels of pressure sensitivity", 199.99, 0, product.ProductStatusEmpty},
		{"Bluetooth Speaker", "Waterproof portable Bluetooth speaker with 12h battery", 44.99, 180, product.ProductStatusInsufficient},
		{"Cable Management Kit", "Under-desk cable tray with zip ties and clips", 14.99, 400, product.ProductStatusInsufficient},
		{"Second Monitor 27 inch", "27-inch IPS monitor with 144Hz refresh rate", 249.99, 0, product.ProductStatusEmpty},
		{"Desk Mat XXL", "Extra-large desk mat with stitched edges, 90x40cm", 24.99, 250, product.ProductStatusInsufficient},
	}

	products := make([]product.Product, 0, len(seeds))
	for i, s := range seeds {
		price, err := product.NewPrice(s.price)
		if err != nil {
			log.Fatalf("Invalid price for seed #%d (%s): %v", i+1, s.name, err)
		}

		stock, err := product.NewStock(s.stock)
		if err != nil {
			log.Fatalf("Invalid stock for seed #%d (%s): %v", i+1, s.name, err)
		}

		products = append(products, product.Product{
			Name:        s.name,
			Description: s.description,
			Price:       price,
			Stock:       stock,
			Status:      s.status,
		})
	}

	return products
}
