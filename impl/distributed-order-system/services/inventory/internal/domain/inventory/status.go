package inventory

// InventoryStatus represents the current state of an inventory item.
type InventoryStatus string

const (
	StatusAvailable  InventoryStatus = "available"
	StatusLowStock   InventoryStatus = "low_stock"
	StatusOutOfStock InventoryStatus = "out_of_stock"
)

func (s InventoryStatus) String() string {
	return string(s)
}

// ReservationStatus represents the current state of an inventory reservation.
type ReservationStatus string

const (
	ReservationStatusReserved  ReservationStatus = "reserved"
	ReservationStatusCompleted ReservationStatus = "completed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
)

func (s ReservationStatus) String() string {
	return string(s)
}
