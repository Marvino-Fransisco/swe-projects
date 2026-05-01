package inventory

import "time"

// InventoryReservation represents a stock reservation for an order.
type InventoryReservation struct {
	id        uint
	orderID   string
	productID string
	quantity  int
	status    ReservationStatus
	createdAt time.Time
	updatedAt time.Time
}

// NewInventoryReservation creates a new reservation in "reserved" status.
func NewInventoryReservation(orderID, productID string, quantity int, now time.Time) InventoryReservation {
	return InventoryReservation{
		orderID:   orderID,
		productID: productID,
		quantity:  quantity,
		status:    ReservationStatusReserved,
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructInventoryReservation rebuilds a reservation from persistence.
func ReconstructInventoryReservation(id uint, orderID, productID string, quantity int, status ReservationStatus, createdAt, updatedAt time.Time) InventoryReservation {
	return InventoryReservation{
		id:        id,
		orderID:   orderID,
		productID: productID,
		quantity:  quantity,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Getters — external code cannot mutate state directly.

func (r InventoryReservation) ID() uint                   { return r.id }
func (r InventoryReservation) OrderID() string             { return r.orderID }
func (r InventoryReservation) ProductID() string           { return r.productID }
func (r InventoryReservation) Quantity() int               { return r.quantity }
func (r InventoryReservation) Status() ReservationStatus   { return r.status }
func (r InventoryReservation) CreatedAt() time.Time        { return r.createdAt }
func (r InventoryReservation) UpdatedAt() time.Time        { return r.updatedAt }

// Complete transitions the reservation to "completed" status.
func (r *InventoryReservation) Complete() {
	r.status = ReservationStatusCompleted
	r.updatedAt = time.Now()
}

// Cancel transitions the reservation to "cancelled" status.
func (r *InventoryReservation) Cancel() {
	r.status = ReservationStatusCancelled
	r.updatedAt = time.Now()
}
