package inventory

import "time"

// Inventory is an aggregate root representing a product's stock level.
type Inventory struct {
	id          uint
	productID   string
	productName string
	stock       int
	price       float64
	status      InventoryStatus
	createdAt   time.Time
	updatedAt   time.Time
}

// NewInventory creates a new Inventory with the given details.
func NewInventory(productID, productName string, stock int, price float64, status InventoryStatus, now time.Time) *Inventory {
	return &Inventory{
		productID:   productID,
		productName: productName,
		stock:       stock,
		price:       price,
		status:      status,
		createdAt:   now,
		updatedAt:   now,
	}
}

// ReconstructInventory rebuilds an Inventory from persistence (used by repository).
func ReconstructInventory(id uint, productID, productName string, stock int, price float64, status InventoryStatus, createdAt, updatedAt time.Time) *Inventory {
	return &Inventory{
		id:          id,
		productID:   productID,
		productName: productName,
		stock:       stock,
		price:       price,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Getters — external code cannot mutate state directly.

func (i *Inventory) ID() uint                { return i.id }
func (i *Inventory) ProductID() string       { return i.productID }
func (i *Inventory) ProductName() string     { return i.productName }
func (i *Inventory) Stock() int              { return i.stock }
func (i *Inventory) Price() float64          { return i.price }
func (i *Inventory) Status() InventoryStatus { return i.status }
func (i *Inventory) CreatedAt() time.Time    { return i.createdAt }
func (i *Inventory) UpdatedAt() time.Time    { return i.updatedAt }

// HasSufficientStock checks if the inventory has enough stock for the requested quantity.
func (i *Inventory) HasSufficientStock(quantity int) bool {
	return i.stock >= quantity
}

// DeductStock reduces the stock by the given quantity and updates the status.
func (i *Inventory) DeductStock(quantity int) {
	i.stock -= quantity
	if i.stock <= 0 {
		i.stock = 0
		i.status = StatusOutOfStock
	}
	i.updatedAt = time.Now()
}

// RestoreStock adds the given quantity back to stock and updates the status.
func (i *Inventory) RestoreStock(quantity int) {
	i.stock += quantity
	if i.stock > 0 && i.status == StatusOutOfStock {
		i.status = StatusAvailable
	}
	i.updatedAt = time.Now()
}
